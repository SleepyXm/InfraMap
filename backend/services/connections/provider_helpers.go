package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"InfraMap/utils"

	"github.com/google/uuid"
)

type OAuthConnection struct {
	AccountID, UserID, Provider, Name, ExternalID string
	Token                                         utils.OAuthTokenBundle
	Metadata                                      map[string]any
}

func SaveOAuthConnection(ctx context.Context, db *sql.DB, connection OAuthConnection) error {
	encrypted, err := encryptOAuthToken(connection.Token)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(connection.Metadata)
	if err != nil {
		return err
	}
	var connectionID string
	err = db.QueryRowContext(ctx, `SELECT id::text FROM connections
		WHERE account_id = $1 AND provider = $2 AND external_account_id = $3
		ORDER BY created_at ASC LIMIT 1`, connection.AccountID, connection.Provider, connection.ExternalID).Scan(&connectionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if err == nil {
		_, err = db.ExecContext(ctx, `UPDATE connections
			SET name = $1, connection_type = 'oauth', secret_ref = $2,
				metadata = COALESCE(metadata, '{}'::jsonb) || $3::jsonb, updated_at = NOW()
			WHERE id = $4 AND account_id = $5 AND provider = $6`,
			connection.Name, encrypted, string(metadata), connectionID, connection.AccountID, connection.Provider)
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO connections (
		id, account_id, name, provider, connection_type, external_account_id,
		secret_ref, metadata, created_by, created_at, updated_at
	) VALUES ($1, $2, $3, $4, 'oauth', $5, $6, $7::jsonb, $8, NOW(), NOW())`,
		uuid.New(), connection.AccountID, connection.Name, connection.Provider, connection.ExternalID,
		encrypted, string(metadata), connection.UserID)
	return err
}

func encryptOAuthToken(bundle utils.OAuthTokenBundle) (string, error) {
	payload, err := json.Marshal(bundle)
	if err != nil {
		return "", err
	}
	return utils.Encrypt(string(payload))
}

func loadProviderConnection(ctx context.Context, db *sql.DB, userID, accountID, connectionID, provider string) (utils.OAuthTokenBundle, []byte, string, error) {
	var encrypted string
	var metadata []byte
	var role string
	err := db.QueryRowContext(ctx, `SELECT c.secret_ref, c.metadata, am.role
		FROM connections c
		JOIN account_memberships am ON am.account_id = c.account_id
		WHERE c.id = $1 AND c.account_id = $2 AND c.provider = $3 AND am.user_id = $4`,
		connectionID, accountID, provider, userID).Scan(&encrypted, &metadata, &role)
	if err != nil {
		return utils.OAuthTokenBundle{}, nil, "", err
	}
	decrypted, err := utils.Decrypt(encrypted)
	if err != nil {
		return utils.OAuthTokenBundle{}, nil, "", err
	}
	var bundle utils.OAuthTokenBundle
	if err := json.Unmarshal([]byte(decrypted), &bundle); err != nil {
		return utils.OAuthTokenBundle{}, nil, "", err
	}
	return bundle, metadata, role, nil
}

type selectableService[T any] interface {
	serviceID() string
	withSelection() T
}

func selectServices[T selectableService[T]](available []T, ids []string) ([]T, error) {
	byID := make(map[string]T, len(available))
	for _, service := range available {
		byID[service.serviceID()] = service
	}
	selected := make([]T, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		service, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("selected service is no longer accessible")
		}
		if !seen[id] {
			selected = append(selected, service.withSelection())
			seen[id] = true
		}
	}
	return selected, nil
}

func selectedServiceIDs[T selectableService[T]](metadata []byte) (map[string]bool, error) {
	var stored map[string]json.RawMessage
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &stored); err != nil {
			return nil, err
		}
	}
	raw, present := stored["services"]
	if !present {
		raw = stored["projects"]
	}
	var services []T
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &services); err != nil {
			return nil, err
		}
	}
	selected := make(map[string]bool, len(services))
	for _, service := range services {
		selected[service.serviceID()] = true
	}
	return selected, nil
}

func saveServiceSelection(ctx context.Context, db *sql.DB, accountID, connectionID, provider string, selected any) error {
	payload, err := json.Marshal(selected)
	if err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `UPDATE connections
		SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('services', $1::jsonb), updated_at = NOW()
		WHERE id = $2 AND account_id = $3 AND provider = $4`, string(payload), connectionID, accountID, provider)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func providerRequest(ctx context.Context, method, endpoint, token string, input, output any, timeout time.Duration) error {
	var payload []byte
	var err error
	if input != nil {
		payload, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		var failure struct {
			Message string          `json:"message"`
			Error   json.RawMessage `json:"error"`
		}
		_ = json.NewDecoder(response.Body).Decode(&failure)
		message := failure.Message
		if message == "" {
			_ = json.Unmarshal(failure.Error, &message)
		}
		if message == "" {
			var nested struct {
				Message string `json:"message"`
			}
			_ = json.Unmarshal(failure.Error, &nested)
			message = nested.Message
		}
		if message != "" {
			return fmt.Errorf("provider returned status %d: %s", response.StatusCode, message)
		}
		return fmt.Errorf("provider returned status %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(output)
}
