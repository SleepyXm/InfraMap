package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"InfraMap/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const oauthStateTTL = 10 * time.Minute

type oauthState struct {
	Provider     string `json:"provider"`
	Intent       string `json:"intent"`
	UserID       string `json:"user_id,omitempty"`
	AccountID    string `json:"account_id,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
}
type oauthToken struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	TeamID           string `json:"team_id"`
	UserID           string `json:"user_id"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
type oauthIdentity struct{ ProviderID, Email, Username string }
type oauthCallbackDetails struct{ TeamID, ConfigurationID, NextURL string }

func OAuthStart() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := strings.ToLower(c.Param("provider"))
		intent := c.DefaultQuery("intent", "login")
		if provider != "github" && provider != "google" && provider != "aws" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported sign-in provider"})
			return
		}
		if intent != "login" && intent != "connect" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sign-in intent"})
			return
		}
		config, ok := oauthConfig(provider)
		if !ok || config.ClientID == "" || config.ClientSecret == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("%s sign-in is not configured", provider)})
			return
		}
		stateData := oauthState{Provider: provider, Intent: intent}
		if intent == "connect" {
			accessToken, err := c.Cookie("access_token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Sign in before connecting a provider"})
				return
			}
			stateData.UserID, err = utils.DecodeAccessToken(accessToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Sign in before connecting a provider"})
				return
			}
		}
		state, err := createOAuthState(c, stateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start sign-in"})
			return
		}
		query := url.Values{"client_id": {config.ClientID}, "redirect_uri": {utils.Cfg.OAuthCallbackURL}, "response_type": {"code"}, "scope": {config.Scope}, "state": {state}}
		c.Redirect(http.StatusFound, config.AuthorizeURL+"?"+query.Encode())
	}
}

func OAuthConnectionStart(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := strings.ToLower(c.Param("provider"))
		accountID := c.Param("accountID")
		userID := c.MustGet("userID").(string)

		if provider != "github" && provider != "supabase" && provider != "vercel" {
			c.JSON(http.StatusNotImplemented, gin.H{"error": fmt.Sprintf("%s connections are not available yet", provider)})
			return
		}
		if _, err := uuid.Parse(accountID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
			return
		}

		var allowed bool
		err := db.QueryRowContext(c, `SELECT EXISTS(
			SELECT 1 FROM account_memberships
			WHERE account_id = $1 AND user_id = $2 AND role IN ('owner', 'admin')
		)`, accountID, userID).Scan(&allowed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify workspace access"})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can add connections"})
			return
		}

		config, ok := oauthConfig(provider)
		if !ok || config.ClientID == "" || config.ClientSecret == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("%s OAuth is not configured", provider)})
			return
		}

		stateData := oauthState{
			Provider:  provider,
			Intent:    "integration",
			UserID:    userID,
			AccountID: accountID,
		}
		query := url.Values{}
		if provider != "vercel" {
			query.Set("client_id", config.ClientID)
			query.Set("redirect_uri", utils.Cfg.OAuthCallbackURL)
			query.Set("response_type", "code")
		}
		if provider == "github" {
			query.Set("scope", "read:user user:email repo read:org")
		} else if provider == "supabase" {
			stateData.CodeVerifier = createCodeVerifier()
			challenge := sha256.Sum256([]byte(stateData.CodeVerifier))
			query.Set("code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:]))
			query.Set("code_challenge_method", "S256")
		}

		state, err := createOAuthState(c, stateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start the provider connection"})
			return
		}

		query.Set("state", state)
		c.Redirect(http.StatusFound, config.AuthorizeURL+"?"+query.Encode())
	}
}

func OAuthCallback(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if providerError := c.Query("error"); providerError != "" {
			state, err := consumeOAuthState(c, c.Query("state"))
			if err != nil {
				oauthRedirect(c, "", providerError, "login")
				return
			}
			oauthRedirect(c, state.Provider, providerError, state.Intent)
			return
		}
		stateValue, code := c.Query("state"), c.Query("code")
		if stateValue == "" || code == "" {
			oauthRedirect(c, "", "invalid_callback", "login")
			return
		}
		state, err := consumeOAuthState(c, stateValue)
		if err != nil {
			oauthRedirect(c, "", "invalid_state", "login")
			return
		}
		config, ok := oauthConfig(state.Provider)
		if !ok {
			oauthRedirect(c, state.Provider, "unsupported_provider", state.Intent)
			return
		}
		token, err := exchangeOAuthCode(c, state.Provider, config, code, state.CodeVerifier)
		if err != nil {
			oauthRedirect(c, state.Provider, "token_exchange_failed", state.Intent)
			return
		}
		if state.Intent == "integration" {
			details := oauthCallbackDetails{
				TeamID:          firstValue(c.Query("teamId"), token.TeamID),
				ConfigurationID: c.Query("configurationId"),
				NextURL:         c.Query("next"),
			}
			if err := connectOAuthIntegration(c, db, state, token, details); err != nil {
				log.Printf("OAuth integration callback failed for %s: %v", state.Provider, err)
				oauthRedirect(c, state.Provider, "connection_failed", state.Intent)
				return
			}
			if state.Provider == "vercel" {
				if completionURL, ok := vercelCompletionURL(details.NextURL); ok {
					c.Redirect(http.StatusFound, completionURL)
					return
				}
			}
			oauthRedirect(c, state.Provider, "", state.Intent)
			return
		}
		identity, err := fetchOAuthIdentity(c, state.Provider, config, token.AccessToken)
		if err != nil {
			oauthRedirect(c, state.Provider, "identity_lookup_failed", state.Intent)
			return
		}
		if state.Intent == "connect" {
			if err := connectOAuthIdentity(c, db, state.UserID, state.Provider, identity); err != nil {
				oauthRedirect(c, state.Provider, "account_link_failed", state.Intent)
				return
			}
			oauthRedirect(c, state.Provider, "", state.Intent)
			return
		}
		userID, err := findOrCreateOAuthUser(c, db, state.Provider, identity)
		if err != nil {
			oauthRedirect(c, state.Provider, "account_link_failed", state.Intent)
			return
		}
		if err := issueOAuthSession(c, userID); err != nil {
			oauthRedirect(c, state.Provider, "session_failed", state.Intent)
			return
		}
		oauthRedirect(c, state.Provider, "", state.Intent)
	}
}

func connectOAuthIntegration(ctx context.Context, db *sql.DB, state oauthState, token oauthToken, details oauthCallbackDetails) error {
	if state.AccountID == "" || state.UserID == "" {
		return fmt.Errorf("invalid integration state")
	}

	var allowed bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM account_memberships
		WHERE account_id = $1 AND user_id = $2 AND role IN ('owner', 'admin')
	)`, state.AccountID, state.UserID).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("workspace access was revoked")
	}

	switch state.Provider {
	case "github":
		config, _ := oauthConfig("github")
		identity, err := fetchOAuthIdentity(ctx, "github", config, token.AccessToken)
		if err != nil {
			return err
		}
		return connectGitHubIntegration(ctx, db, state, token, identity)
	case "supabase":
		return connectSupabaseIntegration(ctx, db, state, token)
	case "vercel":
		return connectVercelIntegration(ctx, db, state, token, details)
	default:
		return fmt.Errorf("unsupported integration provider")
	}
}

func connectGitHubIntegration(ctx context.Context, db *sql.DB, state oauthState, token oauthToken, identity oauthIdentity) error {

	encryptedToken, err := utils.Encrypt(token.AccessToken)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(map[string]any{
		"email": identity.Email,
		"login": identity.Username,
		"scope": "read:user user:email repo read:org",
	})
	if err != nil {
		return err
	}

	var connectionID string
	err = db.QueryRowContext(ctx, `SELECT id::text FROM connections
		WHERE account_id = $1 AND provider = 'github' AND external_account_id = $2
		ORDER BY created_at ASC LIMIT 1`, state.AccountID, identity.ProviderID).Scan(&connectionID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	name := "GitHub · " + identity.Username
	if err == nil {
		_, err = db.ExecContext(ctx, `UPDATE connections
			SET name = $1,
				connection_type = 'oauth',
				secret_ref = $2,
				metadata = COALESCE(metadata, '{}'::jsonb) || $3::jsonb,
				updated_at = NOW()
			WHERE id = $4`, name, encryptedToken, string(metadata), connectionID)
		return err
	}

	_, err = db.ExecContext(ctx, `INSERT INTO connections (
		id, account_id, name, provider, connection_type, external_account_id,
		secret_ref, metadata, created_by, created_at, updated_at
	) VALUES ($1, $2, $3, 'github', 'oauth', $4, $5, $6::jsonb, $7, NOW(), NOW())`,
		uuid.New(), state.AccountID, name, identity.ProviderID, encryptedToken, string(metadata), state.UserID)
	return err
}

func oauthConfig(provider string) (utils.OAuthProviderConfig, bool) {
	switch provider {
	case "github":
		return utils.Cfg.GitHubOAuth, true
	case "google":
		return utils.Cfg.GoogleOAuth, true
	case "supabase":
		return utils.Cfg.SupabaseOAuth, true
	case "vercel":
		return utils.Cfg.VercelOAuth, true
	case "aws":
		return utils.Cfg.AWSOAuth, utils.Cfg.AWSOAuth.AuthorizeURL != "" && utils.Cfg.AWSOAuth.TokenURL != "" && utils.Cfg.AWSOAuth.UserInfoURL != ""
	default:
		return utils.OAuthProviderConfig{}, false
	}
}

func createOAuthState(ctx context.Context, state oauthState) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(bytes)
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return value, utils.RDB.SetEx(ctx, "oauth-state:"+value, payload, oauthStateTTL).Err()
}

func consumeOAuthState(ctx context.Context, value string) (oauthState, error) {
	payload, err := utils.RDB.Get(ctx, "oauth-state:"+value).Bytes()
	if err != nil {
		return oauthState{}, err
	}
	if err := utils.RDB.Del(ctx, "oauth-state:"+value).Err(); err != nil {
		return oauthState{}, err
	}
	var state oauthState
	return state, json.Unmarshal(payload, &state)
}

func exchangeOAuthCode(ctx context.Context, provider string, config utils.OAuthProviderConfig, code, codeVerifier string) (oauthToken, error) {
	form := url.Values{"code": {code}, "redirect_uri": {utils.Cfg.OAuthCallbackURL}}
	if provider != "vercel" {
		form.Set("grant_type", "authorization_code")
	}
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}
	if provider != "supabase" {
		form.Set("client_id", config.ClientID)
		form.Set("client_secret", config.ClientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return oauthToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if provider == "supabase" {
		req.SetBasicAuth(config.ClientID, config.ClientSecret)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return oauthToken{}, err
	}
	defer response.Body.Close()
	var token oauthToken
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return oauthToken{}, err
	}
	if response.StatusCode >= 300 || token.AccessToken == "" {
		return oauthToken{}, fmt.Errorf("oauth token exchange failed: %s", token.Error)
	}
	return token, nil
}

func createCodeVerifier() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return uuid.NewString() + uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func fetchOAuthIdentity(ctx context.Context, provider string, config utils.OAuthProviderConfig, accessToken string) (oauthIdentity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.UserInfoURL, nil)
	if err != nil {
		return oauthIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return oauthIdentity{}, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return oauthIdentity{}, fmt.Errorf("oauth identity request failed")
	}
	var raw map[string]any
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return oauthIdentity{}, err
	}
	identity := oauthIdentity{ProviderID: stringValue(raw["sub"]), Email: stringValue(raw["email"]), Username: firstValue(stringValue(raw["login"]), stringValue(raw["preferred_username"]), stringValue(raw["name"]))}
	if identity.ProviderID == "" {
		identity.ProviderID = stringValue(raw["id"])
	}
	if provider == "github" && identity.Email == "" {
		identity.Email, err = fetchGitHubEmail(ctx, accessToken)
		if err != nil {
			return oauthIdentity{}, err
		}
	}
	if identity.ProviderID == "" || identity.Email == "" {
		return oauthIdentity{}, fmt.Errorf("provider did not return an account id and email")
	}
	if identity.Username == "" {
		identity.Username = strings.Split(identity.Email, "@")[0]
	}
	return identity, nil
}

func fetchGitHubEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return "", fmt.Errorf("github email lookup failed")
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(response.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, entry := range emails {
		if entry.Primary && entry.Verified {
			return entry.Email, nil
		}
	}
	return "", fmt.Errorf("github did not return a verified primary email")
}

func findOrCreateOAuthUser(ctx context.Context, db *sql.DB, provider string, identity oauthIdentity) (string, error) {
	var existingID string
	err := db.QueryRowContext(ctx, `SELECT user_id FROM user_identities WHERE provider = $1 AND provider_id = $2`, provider, identity.ProviderID).Scan(&existingID)
	if err == nil {
		return existingID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	var emailExists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, strings.ToLower(identity.Email)).Scan(&emailExists); err != nil {
		return "", err
	}
	if emailExists {
		return "", fmt.Errorf("an account with this email already exists")
	}
	username, err := availableOAuthUsername(ctx, db, identity.Username)
	if err != nil {
		return "", err
	}
	password, err := utils.HashPassword(uuid.NewString())
	if err != nil {
		return "", err
	}
	userID, accountID := uuid.New(), uuid.New()
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `INSERT INTO users (id, username, email, password, verified, created_at, updated_at) VALUES ($1, $2, $3, $4, true, NOW(), NOW())`, userID, username, strings.ToLower(identity.Email), password); err != nil {
		return "", err
	}
	if _, err = transaction.ExecContext(ctx, `INSERT INTO accounts (id, name, slug, type, created_by, created_at, updated_at) VALUES ($1, $2, $3, 'personal', $4, NOW(), NOW())`, accountID, username, username+"-"+userID.String()); err != nil {
		return "", err
	}
	if _, err = transaction.ExecContext(ctx, `INSERT INTO account_memberships (account_id, user_id, role, created_at) VALUES ($1, $2, 'owner', NOW())`, accountID, userID); err != nil {
		return "", err
	}
	if _, err = transaction.ExecContext(ctx, `INSERT INTO user_identities (id, user_id, provider, provider_id, email, username, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())`, uuid.New(), userID, provider, identity.ProviderID, strings.ToLower(identity.Email), username); err != nil {
		return "", err
	}
	return userID.String(), transaction.Commit()
}

func connectOAuthIdentity(ctx context.Context, db *sql.DB, userID, provider string, identity oauthIdentity) error {
	if userID == "" {
		return fmt.Errorf("missing account")
	}
	var ownerID string
	err := db.QueryRowContext(ctx, `SELECT user_id FROM user_identities WHERE provider = $1 AND provider_id = $2`, provider, identity.ProviderID).Scan(&ownerID)
	if err == nil {
		if ownerID != userID {
			return fmt.Errorf("provider identity belongs to another account")
		}
		_, err = db.ExecContext(ctx, `UPDATE user_identities SET email = $1, username = $2 WHERE provider = $3 AND provider_id = $4`, strings.ToLower(identity.Email), identity.Username, provider, identity.ProviderID)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("account not found")
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_identities (id, user_id, provider, provider_id, email, username, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())`, uuid.New(), userID, provider, identity.ProviderID, strings.ToLower(identity.Email), identity.Username)
	return err
}

func availableOAuthUsername(ctx context.Context, db *sql.DB, value string) (string, error) {
	base := strings.ToLower(value)
	base = strings.Map(func(char rune) rune {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '_' {
			return char
		}
		return '_'
	}, base)
	base = strings.Trim(base, "_")
	if len(base) < 3 {
		base = "inframap_user"
	}
	if len(base) > 24 {
		base = base[:24]
	}
	for index := 0; index < 100; index++ {
		candidate := base
		if index > 0 {
			candidate = fmt.Sprintf("%s_%d", base, index)
		}
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate username")
}
func issueOAuthSession(ctx *gin.Context, userID string) error {
	access, err := utils.CreateAccessToken(userID)
	if err != nil {
		return err
	}
	refresh, err := utils.CreateRefreshToken(userID)
	if err != nil {
		return err
	}
	if err = utils.StoreRefreshToken(ctx, userID, refresh); err != nil {
		return err
	}
	utils.SetAuthCookies(ctx, access, refresh)
	return nil
}
func oauthRedirect(c *gin.Context, provider, problem, intent string) {
	target := utils.Cfg.DevServer + "/login/callback"
	query := url.Values{}
	if intent == "integration" {
		target = utils.Cfg.DevServer + "/workspaces"
		query.Set("tab", "connections")
		query.Set("oauth", "success")
	}
	if provider != "" {
		query.Set("provider", provider)
	}
	if problem != "" {
		query.Set("error", problem)
		query.Del("oauth")
	}
	if intent == "connect" {
		query.Set("intent", intent)
	}
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	c.Redirect(http.StatusFound, target)
}
func vercelCompletionURL(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	target, err := url.Parse(value)
	if err != nil || target.Scheme != "https" {
		return "", false
	}
	host := strings.ToLower(target.Hostname())
	if host != "vercel.com" && !strings.HasSuffix(host, ".vercel.com") {
		return "", false
	}
	return target.String(), true
}
func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		return ""
	}
}
func firstValue(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
