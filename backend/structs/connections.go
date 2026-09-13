package structs

import "time"

// ConnectionCategory identifies the functional category a connection belongs to.
type ConnectionCategory string

const (
	ConnectionCategoryAccount    ConnectionCategory = "account"    // External account or organisation connection.
	ConnectionCategorySource     ConnectionCategory = "source"     // Source control provider connection.
	ConnectionCategoryCloud      ConnectionCategory = "cloud"      // Cloud infrastructure provider connection.
	ConnectionCategoryDeployment ConnectionCategory = "deployment" // Deployment platform connection.
	ConnectionCategoryData       ConnectionCategory = "data"       // Database or backend platform connection.
	ConnectionCategoryServer     ConnectionCategory = "server"     // Direct server or infrastructure connection.
)

// ConnectionStatus identifies the current health or usability state of a connection.
type ConnectionStatus string

const (
	ConnectionStatusPending  ConnectionStatus = "pending"  // Connection exists but has not been fully verified.
	ConnectionStatusActive   ConnectionStatus = "active"   // Connection is verified and available for use.
	ConnectionStatusInvalid  ConnectionStatus = "invalid"  // Connection credentials or configuration are invalid.
	ConnectionStatusDisabled ConnectionStatus = "disabled" // Connection exists but is intentionally disabled.
)

// Connection is the base connection model shared by every provider and category.
type Connection struct {
	ID                string                 `json:"id"`
	AccountID         string                 `json:"account_id"`
	Name              string                 `json:"name"`
	Category          ConnectionCategory     `json:"category"`
	Provider          string                 `json:"provider"`
	ConnectionType    string                 `json:"connection_type"`
	ExternalAccountID string                 `json:"external_account_id,omitempty"`
	Status            ConnectionStatus       `json:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy         string                 `json:"created_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	LastVerifiedAt    *time.Time             `json:"last_verified_at,omitempty"`
}

// AccountConnection represents an external provider account or organisation linked to an InfraMap account.
type AccountConnection struct {
	Connection
	ExternalName string `json:"external_name,omitempty"`
	ExternalType string `json:"external_type,omitempty"`
}

// SourceConnection represents a source control account or installation such as GitHub, GitLab, or Bitbucket.
type SourceConnection struct {
	Connection
	Owner          string `json:"owner,omitempty"`
	OwnerType      string `json:"owner_type,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`
}

// CloudConnection represents a cloud infrastructure provider account such as AWS, GCP, or Azure.
type CloudConnection struct {
	Connection
	CloudAccountID string `json:"cloud_account_id,omitempty"`
	Region         string `json:"region,omitempty"`
}

// DeploymentConnection represents a deployment platform account such as Vercel, Railway, Render, or Netlify.
type DeploymentConnection struct {
	Connection
	TeamID   string `json:"team_id,omitempty"`
	TeamName string `json:"team_name,omitempty"`
}

// DataConnection represents a database or backend platform such as Supabase, Neon, or PlanetScale.
type DataConnection struct {
	Connection
	ServiceID   string `json:"service_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Region      string `json:"region,omitempty"`
}

// ServerConnection represents direct infrastructure such as an SSH server, Docker host, or machine.
type ServerConnection struct {
	Connection
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

// CreateConnectionRequest contains the generic fields required to create a connection.
type CreateConnectionRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Category          ConnectionCategory     `json:"category" binding:"required"`
	Provider          string                 `json:"provider" binding:"required"`
	ConnectionType    string                 `json:"connection_type" binding:"required"`
	ExternalAccountID string                 `json:"external_account_id"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ConnectOAuthRequest contains the fields required for an OAuth-based provider connection.
type ConnectOAuthRequest struct {
	Name        string `json:"name"`
	Provider    string `json:"provider" binding:"required"`
	Code        string `json:"code" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
}

// ConnectGitHubRequest contains the fields required to connect a GitHub App installation.
type ConnectGitHubRequest struct {
	Name           string `json:"name"`
	Code           string `json:"code,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`
	RedirectURI    string `json:"redirect_uri,omitempty"`
}

// ConnectAWSRequest contains the fields required to connect an AWS account using an IAM role.
type ConnectAWSRequest struct {
	Name       string `json:"name"`
	RoleARN    string `json:"role_arn" binding:"required"`
	Region     string `json:"region"`
	ExternalID string `json:"external_id,omitempty"`
}

// ConnectGCPRequest contains the fields required to connect a Google Cloud project.
type ConnectGCPRequest struct {
	Name        string `json:"name"`
	ProjectID   string `json:"project_id"`
	Credentials string `json:"credentials" binding:"required"`
}

// ConnectAzureRequest contains the fields required to connect an Azure subscription.
type ConnectAzureRequest struct {
	Name           string `json:"name"`
	TenantID       string `json:"tenant_id" binding:"required"`
	SubscriptionID string `json:"subscription_id" binding:"required"`
	ClientID       string `json:"client_id" binding:"required"`
	ClientSecret   string `json:"client_secret" binding:"required"`
}

// ConnectServerRequest contains the fields required to connect directly to a server over SSH.
type ConnectServerRequest struct {
	Name       string `json:"name" binding:"required"`
	Host       string `json:"host" binding:"required"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	PrivateKey string `json:"private_key,omitempty"`
}

// ConnectKubernetesRequest contains the fields required to connect an existing Kubernetes cluster.
type ConnectKubernetesRequest struct {
	Name       string `json:"name" binding:"required"`
	Kubeconfig string `json:"kubeconfig" binding:"required"`
	Context    string `json:"context"`
}

// DisconnectConnectionRequest identifies the connection that should be removed.
type DisconnectConnectionRequest struct {
	ConnectionID string `json:"connection_id" binding:"required"`
}
