package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"InfraMap/utils"

	"github.com/google/uuid"
)

const supabaseManagementAPI = "https://api.supabase.com"

type supabaseOrganization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func connectSupabaseIntegration(ctx context.Context, db *sql.DB, state oauthState, token oauthToken) error {
	organizations, err := fetchSupabaseOrganizations(ctx, token.AccessToken)
	if err != nil {
		return err
	}
	if len(organizations) == 0 {
		return fmt.Errorf("supabase returned no authorised organization")
	}

	expiresAt := time.Time{}
	if token.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	bundle, err := json.Marshal(utils.OAuthTokenBundle{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Scope:        token.Scope,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return err
	}
	encryptedToken, err := utils.Encrypt(string(bundle))
	if err != nil {
		return err
	}

	organization := organizations[0]
	metadata, err := json.Marshal(map[string]any{
		"organization":  organization,
		"organizations": organizations,
		"scope":         token.Scope,
	})
	if err != nil {
		return err
	}

	var connectionID string
	err = db.QueryRowContext(ctx, `SELECT id::text FROM connections
		WHERE account_id = $1 AND provider = 'supabase' AND external_account_id = $2
		ORDER BY created_at ASC LIMIT 1`, state.AccountID, organization.ID).Scan(&connectionID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	name := "Supabase · " + organization.Name
	if err == nil {
		_, err = db.ExecContext(ctx, `UPDATE connections
			SET name = $1,
				connection_type = 'oauth',
				secret_ref = $2,
				metadata = COALESCE(metadata, '{}'::jsonb) || $3::jsonb,
				updated_at = NOW()
			WHERE id = $4`, name, encryptedToken, string(metadata), connectionID)
		return err
	}

	_, err = db.ExecContext(ctx, `INSERT INTO connections (
		id, account_id, name, provider, connection_type, external_account_id,
		secret_ref, metadata, created_by, created_at, updated_at
	) VALUES ($1, $2, $3, 'supabase', 'oauth', $4, $5, $6::jsonb, $7, NOW(), NOW())`,
		uuid.New(), state.AccountID, name, organization.ID, encryptedToken, string(metadata), state.UserID)
	return err
}

func fetchSupabaseOrganizations(ctx context.Context, accessToken string) ([]supabaseOrganization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, supabaseManagementAPI+"/v1/organizations", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase organizations returned status %d", response.StatusCode)
	}

	var organizations []supabaseOrganization
	if err := json.NewDecoder(response.Body).Decode(&organizations); err != nil {
		return nil, err
	}
	return organizations, nil
}
