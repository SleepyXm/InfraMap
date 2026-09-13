package workspaces

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"InfraMap/structs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResourceInput struct {
	ConnectionID string
	Provider     string
	ResourceType string
	ExternalID   string
	Name         string
	Status       string
	Metadata     map[string]interface{}
	SecretRef    string
}

func SaveResource(ctx context.Context, db *sql.DB, userID, workspaceID string, input ResourceInput) (structs.WorkspaceResource, error) {
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	defer tx.Rollback()

	var resource structs.WorkspaceResource
	var storedMetadata []byte
	err = tx.QueryRowContext(ctx, `
		INSERT INTO workspace_resources (
			id, workspace_id, connection_id, provider, resource_type, external_id,
			name, status, metadata, secret_ref, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, NULLIF($10, ''), $11, NOW(), NOW())
		ON CONFLICT (workspace_id, provider, resource_type) DO UPDATE SET
			connection_id = EXCLUDED.connection_id, external_id = EXCLUDED.external_id,
			name = EXCLUDED.name, status = EXCLUDED.status, metadata = EXCLUDED.metadata,
			secret_ref = EXCLUDED.secret_ref, updated_at = NOW()
		RETURNING id, workspace_id, connection_id, provider, resource_type, external_id,
		          name, status, metadata, COALESCE(created_by::text, ''), created_at, updated_at`,
		uuid.New(), workspaceID, input.ConnectionID, input.Provider, input.ResourceType, input.ExternalID,
		input.Name, input.Status, string(metadata), input.SecretRef, userID,
	).Scan(&resource.ID, &resource.WorkspaceID, &resource.ConnectionID, &resource.Provider, &resource.ResourceType, &resource.ExternalID, &resource.Name, &resource.Status, &storedMetadata, &resource.CreatedBy, &resource.CreatedAt, &resource.UpdatedAt)
	if err != nil {
		return structs.WorkspaceResource{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE workspaces SET status = 'active', updated_at = NOW() WHERE id = $1`, workspaceID); err != nil {
		return structs.WorkspaceResource{}, err
	}
	if err := tx.Commit(); err != nil {
		return structs.WorkspaceResource{}, err
	}
	resource.Metadata = make(map[string]interface{})
	if err := json.Unmarshal(storedMetadata, &resource.Metadata); err != nil {
		return structs.WorkspaceResource{}, err
	}
	return resource, nil
}

func LoadResources(ctx context.Context, db *sql.DB, workspaceID string) ([]structs.WorkspaceResource, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, workspace_id, connection_id, provider, resource_type, external_id,
		       name, status, metadata, COALESCE(created_by::text, ''), created_at, updated_at
		FROM workspace_resources
		WHERE workspace_id = $1
		ORDER BY created_at ASC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resources := make([]structs.WorkspaceResource, 0)
	for rows.Next() {
		var resource structs.WorkspaceResource
		var metadata []byte
		if err := rows.Scan(&resource.ID, &resource.WorkspaceID, &resource.ConnectionID, &resource.Provider, &resource.ResourceType, &resource.ExternalID, &resource.Name, &resource.Status, &metadata, &resource.CreatedBy, &resource.CreatedAt, &resource.UpdatedAt); err != nil {
			return nil, err
		}
		resource.Metadata = make(map[string]interface{})
		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &resource.Metadata); err != nil {
				return nil, err
			}
		}
		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

func FindResource(resources []structs.WorkspaceResource, provider string) *structs.WorkspaceResource {
	for index := range resources {
		if resources[index].Provider == provider {
			return &resources[index]
		}
	}
	return nil
}

func requireServiceWriteAccess(c *gin.Context, db *sql.DB, userID, accountID, workspaceID string, provision bool) bool {
	_, role, err := LoadWorkspace(c, db, userID, accountID, workspaceID)
	if err != nil {
		WriteWorkspaceAccessError(c, err)
		return false
	}
	if role == "viewer" || (provision && role != "owner" && role != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to modify workspace services"})
		return false
	}
	return true
}
