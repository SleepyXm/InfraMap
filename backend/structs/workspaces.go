package structs

import "time"

type Workspace struct {
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

type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type WorkspaceResource struct {
	ID           string                 `json:"id"`
	WorkspaceID  string                 `json:"workspace_id"`
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

type AttachSupabaseServiceRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	ServiceRef   string `json:"service_ref" binding:"required"`
}

type AttachGitHubRepositoryRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	RepositoryID int64  `json:"repository_id" binding:"required"`
}

type AttachVercelServiceRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
	ServiceID    string `json:"vercel_service_id" binding:"required"`
}

type CreateSupabaseServiceRequest struct {
	ConnectionID     string `json:"connection_id" binding:"required"`
	Name             string `json:"name" binding:"required"`
	OrganizationSlug string `json:"organization_slug" binding:"required"`
	RegionGroup      string `json:"region_group" binding:"required"`
	DatabasePassword string `json:"database_password" binding:"required"`
}

type AttachAWSHostRequest struct {
	Name         string `json:"name" binding:"required"`
	RoleARN      string `json:"role_arn" binding:"required"`
	ExternalID   string `json:"external_id" binding:"required"`
	Region       string `json:"region" binding:"required"`
	InstanceID   string `json:"instance_id" binding:"required"`
	DocumentName string `json:"document_name" binding:"required"`
	StackID      string `json:"stack_id"`
	PublicIP     string `json:"public_ip"`
}

type CreateEZDeployDeploymentRequest struct {
	Domain  string `json:"domain" binding:"required"`
	Email   string `json:"email" binding:"required"`
	Runtime string `json:"runtime" binding:"required"`
}

type EZDeployDecision struct {
	ComponentID string `json:"component_id" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Provider    string `json:"provider"`
}

type SaveEZDeployDecisionsRequest struct {
	Decisions []EZDeployDecision `json:"decisions" binding:"required"`
}
