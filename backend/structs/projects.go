package structs

import "time"

type Project struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type ProjectResource struct {
	ID           string                 `json:"id"`
	ProjectID    string                 `json:"project_id"`
	ConnectionID string                 `json:"connection_id"`
	Provider     string                 `json:"provider"`
	ResourceType string                 `json:"resource_type"`
	ExternalID   string                 `json:"external_id"`
	Name         string                 `json:"name"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedBy    string                 `json:"created_by,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type AttachSupabaseProjectRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	ProjectRef   string `json:"project_ref" binding:"required"`
}

type AttachGitHubRepositoryRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	RepositoryID int64  `json:"repository_id" binding:"required"`
}

type AttachVercelProjectRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	ProjectID    string `json:"vercel_project_id" binding:"required"`
}

type CreateSupabaseProjectRequest struct {
	ConnectionID     string `json:"connection_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	OrganizationSlug string `json:"organization_slug" binding:"required"`
	RegionGroup      string `json:"region_group" binding:"required"`
	DatabasePassword string `json:"database_password" binding:"required"`
}
