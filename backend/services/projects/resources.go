package projects

import (
	"context"
	"database/sql"
	"encoding/json"

	"InfraMap/structs"

	"github.com/google/uuid"
)

type projectResourceInput struct {
	ConnectionID string
	Provider     string
	ResourceType string
	ExternalID   string
	Name         string
	Status       string
	Metadata     map[string]interface{}
	SecretRef    string
}

func saveProjectResource(ctx context.Context, db *sql.DB, userID, projectID string, input projectResourceInput) (structs.ProjectResource, error) {
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return structs.ProjectResource{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return structs.ProjectResource{}, err
	}
	defer tx.Rollback()

	var resource structs.ProjectResource
	var storedMetadata []byte
	err = tx.QueryRowContext(ctx, `
		INSERT INTO project_resources (
			id, project_id, connection_id, provider, resource_type, external_id,
			name, status, metadata, secret_ref, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, NULLIF($10, ''), $11, NOW(), NOW())
		ON CONFLICT (project_id, provider, resource_type) DO UPDATE SET
			connection_id = EXCLUDED.connection_id, external_id = EXCLUDED.external_id,
			name = EXCLUDED.name, status = EXCLUDED.status, metadata = EXCLUDED.metadata,
			secret_ref = EXCLUDED.secret_ref, updated_at = NOW()
		RETURNING id, project_id, connection_id, provider, resource_type, external_id,
		          name, status, metadata, COALESCE(created_by::text, ''), created_at, updated_at`,
		uuid.New(), projectID, input.ConnectionID, input.Provider, input.ResourceType, input.ExternalID,
		input.Name, input.Status, string(metadata), input.SecretRef, userID,
	).Scan(&resource.ID, &resource.ProjectID, &resource.ConnectionID, &resource.Provider, &resource.ResourceType, &resource.ExternalID, &resource.Name, &resource.Status, &storedMetadata, &resource.CreatedBy, &resource.CreatedAt, &resource.UpdatedAt)
	if err != nil {
		return structs.ProjectResource{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET status = 'active', updated_at = NOW() WHERE id = $1`, projectID); err != nil {
		return structs.ProjectResource{}, err
	}
	if err := tx.Commit(); err != nil {
		return structs.ProjectResource{}, err
	}
	resource.Metadata = make(map[string]interface{})
	if err := json.Unmarshal(storedMetadata, &resource.Metadata); err != nil {
		return structs.ProjectResource{}, err
	}
	return resource, nil
}
