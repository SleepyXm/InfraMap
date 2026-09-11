package structs

import "time"

type UserCreate struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// OAuth login/signup identity.
// This belongs to the USER, not an account.
type OAuthCallback struct {
	Provider    string `json:"provider" binding:"required"`
	Code        string `json:"code" binding:"required"`
	RedirectURI string `json:"redirect_uri"`
}

type ConnectedIdentity struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	ProviderID   string    `json:"provider_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"token_expiry"`
	ConnectedAt  time.Time `json:"connected_at"`
	Primary      bool      `json:"primary"`
}

// Account
type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Type      string    `json:"type"` // "personal", "organization"
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountMembership struct {
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // "owner", "admin", "member", "viewer"
	CreatedAt time.Time `json:"created_at"`
}

type AccountWithRole struct {
	Account
	Role string `json:"role"`
}

// User profile should describe the user.
// Connections no longer belong directly here.
type UserProfile struct {
	ID          string              `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	Verified    bool                `json:"verified"`
	Identities  []ConnectedIdentity `json:"identities"`
	Accounts    []AccountWithRole   `json:"accounts"`
	CreatedAt   time.Time           `json:"created_at"`
	LastLoginAt time.Time           `json:"last_login_at"`
}

// Account profile is where connections live.
type AccountProfile struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Type        string              `json:"type"`
	Role        string              `json:"role"`
	Members     []AccountMembership `json:"members,omitempty"`
	Connections []Connection        `json:"connections"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// Account requests
type CreateAccountRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
	Type string `json:"type"`
}

type UpdateAccountRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AddAccountMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

type UpdateAccountMemberRequest struct {
	Role string `json:"role" binding:"required"`
}
