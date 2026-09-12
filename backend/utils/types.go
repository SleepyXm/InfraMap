package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	UserInfoURL  string
	Scope        string
}

type OAuthTokenBundle struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}

type Config struct {
	DatabaseURL              string
	SecretKey                string
	Algorithm                string
	AccessTokenExpireMinutes int
	RefreshTokenExpireDays   int
	EncryptionKey            string
	DevServer                string
	DevServerBackend         string
	FrontendProd             string
	ResendAPIKey             string
	RedisAddr                string
	KnowledgeDir             string
	OAuthCallbackURL         string
	GitHubOAuth              OAuthProviderConfig
	GoogleOAuth              OAuthProviderConfig
	SupabaseOAuth            OAuthProviderConfig
	VercelOAuth              OAuthProviderConfig
	AWSOAuth                 OAuthProviderConfig
}

type Claims struct {
	Sub  string `json:"sub"`
	Type string `json:"type,omitempty"`
	jwt.RegisteredClaims
}
