package utils

import "github.com/golang-jwt/jwt/v5"

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	UserInfoURL  string
	Scope        string
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
	AWSOAuth                 OAuthProviderConfig
}

type Claims struct {
	Sub  string `json:"sub"`
	Type string `json:"type,omitempty"`
	jwt.RegisteredClaims
}
