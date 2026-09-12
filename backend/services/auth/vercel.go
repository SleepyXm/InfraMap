package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"InfraMap/utils"

	"github.com/google/uuid"
)

const vercelAPIURL = "https://api.vercel.com"

func connectVercelIntegration(ctx context.Context, db *sql.DB, state oauthState, token oauthToken, details oauthCallbackDetails) error {
	if details.ConfigurationID == "" {
		return fmt.Errorf("vercel callback did not include a configuration id")
	}
	bundle, err := json.Marshal(utils.OAuthTokenBundle{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Scope:       token.Scope,
	})
	if err != nil {
		return err
	}
	encryptedToken, err := utils.Encrypt(string(bundle))
	if err != nil {
		return err
	}

	accountName := fetchVercelAccountName(ctx, token.AccessToken, details.TeamID)
	name := "Vercel"
	if accountName != "" {
		name += " · " + accountName
	}
	metadata, err := json.Marshal(map[string]any{
		"configuration_id": details.ConfigurationID,
		"team_id":          details.TeamID,
		"account_name":     accountName,
		"scope":            token.Scope,
	})
	if err != nil {
		return err
	}

	var connectionID string
	err = db.QueryRowContext(ctx, `SELECT id::text FROM connections
		WHERE account_id = $1 AND provider = 'vercel' AND external_account_id = $2
		ORDER BY created_at ASC LIMIT 1`, state.AccountID, details.ConfigurationID).Scan(&connectionID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
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
	) VALUES ($1, $2, $3, 'vercel', 'oauth', $4, $5, $6::jsonb, $7, NOW(), NOW())`,
		uuid.New(), state.AccountID, name, details.ConfigurationID, encryptedToken, string(metadata), state.UserID)
	return err
}

func fetchVercelAccountName(ctx context.Context, accessToken, teamID string) string {
	endpoint := vercelAPIURL + "/v2/user"
	if teamID != "" {
		endpoint = vercelAPIURL + "/v2/teams/" + url.PathEscape(teamID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return ""
	}
	var raw map[string]any
	if json.NewDecoder(response.Body).Decode(&raw) != nil {
		return ""
	}
	if user, ok := raw["user"].(map[string]any); ok {
		return firstValue(stringValue(user["name"]), stringValue(user["username"]))
	}
	return firstValue(stringValue(raw["name"]), stringValue(raw["slug"]))
}
