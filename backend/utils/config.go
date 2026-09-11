package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Cfg Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	backendURL := requireEnv("DEV_SERVER_BACKEND")
	Cfg = Config{
		DatabaseURL:              requireEnv("DATABASE"),
		SecretKey:                requireEnv("SECRET_KEY"),
		Algorithm:                requireEnv("ALGORITHM"),
		AccessTokenExpireMinutes: requireEnvInt("ACCESS_TOKEN_EXPIRE_MINUTES"),
		RefreshTokenExpireDays:   requireEnvInt("REFRESH_TOKEN_EXPIRE_DAYS"),
		EncryptionKey:            requireEnv("ENCRYPTION_KEY"),
		DevServer:                requireEnv("DEV_SERVER"),
		DevServerBackend:         backendURL,
		FrontendProd:             requireEnv("FRONTEND_PROD"),
		ResendAPIKey:             requireEnv("RESEND_API_KEY"),
		RedisAddr:                requireEnv("REDIS_URL"),
		KnowledgeDir:             envOrDefault("KNOWLEDGE_DIR", "knowledge"),
		OAuthCallbackURL:         envOrDefault("OAUTH_CALLBACK_URL", backendURL+"/api/auth/callback"),
		GitHubOAuth:              OAuthProviderConfig{ClientID: os.Getenv("GITHUB_CLIENT_ID"), ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"), AuthorizeURL: "https://github.com/login/oauth/authorize", TokenURL: "https://github.com/login/oauth/access_token", UserInfoURL: "https://api.github.com/user", Scope: "read:user user:email"},
		GoogleOAuth:              OAuthProviderConfig{ClientID: os.Getenv("GOOGLE_CLIENT_ID"), ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), AuthorizeURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token", UserInfoURL: "https://openidconnect.googleapis.com/v1/userinfo", Scope: "openid email profile"},
		AWSOAuth:                 OAuthProviderConfig{ClientID: os.Getenv("AWS_OAUTH_CLIENT_ID"), ClientSecret: os.Getenv("AWS_OAUTH_CLIENT_SECRET"), AuthorizeURL: os.Getenv("AWS_OAUTH_AUTHORIZE_URL"), TokenURL: os.Getenv("AWS_OAUTH_TOKEN_URL"), UserInfoURL: os.Getenv("AWS_OAUTH_USERINFO_URL"), Scope: envOrDefault("AWS_OAUTH_SCOPE", "openid email profile")},
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		log.Fatalf("Env var %s must be a positive integer, got: %s", key, value)
	}
	return n
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required env var %s is not set", key)
	}
	return val
}

func requireEnvInt(key string) int {
	val := requireEnv(key)
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Env var %s must be an integer, got: %s", key, val)
	}
	return n
}
