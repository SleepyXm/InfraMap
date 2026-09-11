package structs

import "time"

// Generic infrastructure/service connection.
// This maps directly to the connections table.
type Connection struct {
	ID                string                 `json:"id"`
	AccountID         string                 `json:"account_id"`
	Name              string                 `json:"name"`
	Provider          string                 `json:"provider"`
	ConnectionType    string                 `json:"connection_type"`
	ExternalAccountID string                 `json:"external_account_id,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy         string                 `json:"created_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// Provider-specific representations can still exist.
// They are views of Connection, rather than separate ownership models.
type GitHubConnection struct {
	ID             string    `json:"id"`
	AccountID      string    `json:"account_id"`
	Name           string    `json:"name"`
	InstallationID string    `json:"installation_id"`
	Login          string    `json:"login"`
	AccountType    string    `json:"account_type"` // "User", "Organization"
	Status         string    `json:"status"`
	ConnectedAt    time.Time `json:"connected_at"`
	LastVerified   time.Time `json:"last_verified"`
}

type AWSConnection struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	AWSAccountID string    `json:"aws_account_id"`
	RoleARN      string    `json:"role_arn"`
	ExternalID   string    `json:"-"`
	Alias        string    `json:"alias"`
	Region       string    `json:"region"`
	Status       string    `json:"status"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastVerified time.Time `json:"last_verified"`
}

type GCPConnection struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	ProjectID    string    `json:"project_id"`
	ProjectName  string    `json:"project_name"`
	Alias        string    `json:"alias"`
	Credentials  string    `json:"-"`
	Status       string    `json:"status"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastVerified time.Time `json:"last_verified"`
}

// Generic connection request used by POST /accounts/:accountID/connections.
type CreateConnectionRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Provider          string                 `json:"provider" binding:"required"`
	ConnectionType    string                 `json:"connection_type" binding:"required"`
	ExternalAccountID string                 `json:"external_account_id"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// Provider-specific connection requests.
type ConnectGitHubRequest struct {
	Code        string `json:"code" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
	Name        string `json:"name"`
}

type ConnectAWSRequest struct {
	RoleARN string `json:"role_arn" binding:"required"`
	Alias   string `json:"alias"`
	Region  string `json:"region"`
}

type ConnectGCPRequest struct {
	Credentials string `json:"credentials" binding:"required"`
	Alias       string `json:"alias"`
}

type DisconnectConnectionRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
}
