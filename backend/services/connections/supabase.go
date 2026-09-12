package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"InfraMap/utils"

	"github.com/gin-gonic/gin"
)

const supabaseManagementAPI = "https://api.supabase.com"

var errSupabaseConnectionNotFound = errors.New("supabase connection not found")

type SupabaseProject struct {
	ID               string                   `json:"id"`
	Ref              string                   `json:"ref"`
	OrganizationID   string                   `json:"organization_id"`
	OrganizationSlug string                   `json:"organization_slug"`
	Name             string                   `json:"name"`
	Region           string                   `json:"region"`
	CreatedAt        string                   `json:"created_at"`
	Status           string                   `json:"status"`
	Database         *SupabaseProjectDatabase `json:"database,omitempty"`
	Selected         bool                     `json:"selected"`
}

type SupabaseProjectDatabase struct {
	Host           string `json:"host"`
	Version        string `json:"version"`
	PostgresEngine string `json:"postgres_engine"`
	ReleaseChannel string `json:"release_channel"`
}

type updateSupabaseProjectsRequest struct {
	ProjectRefs []string `json:"project_refs"`
}

func GetSupabaseProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		token, selected, _, err := loadSupabaseConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errSupabaseConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Supabase connection"})
			return
		}

		projects, err := fetchSupabaseProjects(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase projects could not be loaded"})
			return
		}
		for index := range projects {
			projects[index].Selected = selected[projects[index].Ref]
		}
		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

func UpdateSupabaseProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		var request updateSupabaseProjectsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose the Supabase projects to attach"})
			return
		}
		if len(request.ProjectRefs) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many projects selected"})
			return
		}

		token, _, role, err := loadSupabaseConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errSupabaseConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Supabase connection"})
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can change project access"})
			return
		}

		projects, err := fetchSupabaseProjects(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Supabase projects could not be verified"})
			return
		}
		available := make(map[string]SupabaseProject, len(projects))
		for _, project := range projects {
			available[project.Ref] = project
		}

		selected := make([]SupabaseProject, 0, len(request.ProjectRefs))
		seen := make(map[string]bool, len(request.ProjectRefs))
		for _, projectRef := range request.ProjectRefs {
			project, ok := available[projectRef]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "A selected Supabase project is no longer accessible"})
				return
			}
			if seen[projectRef] {
				continue
			}
			seen[projectRef] = true
			project.Selected = true
			selected = append(selected, project)
		}

		payload, err := json.Marshal(selected)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Supabase project selection"})
			return
		}
		result, err := db.ExecContext(c, `UPDATE connections
			SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('projects', $1::jsonb),
				updated_at = NOW()
			WHERE id = $2 AND account_id = $3 AND provider = 'supabase'`, string(payload), connectionID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Supabase project selection"})
			return
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != 1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Supabase connection not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"projects": selected})
	}
}

func loadSupabaseConnection(
	ctx context.Context,
	db *sql.DB,
	userID string,
	accountID string,
	connectionID string,
) (string, map[string]bool, string, error) {
	var encryptedToken string
	var metadata []byte
	var role string

	err := db.QueryRowContext(ctx, `SELECT c.secret_ref, c.metadata, am.role
		FROM connections c
		JOIN account_memberships am ON am.account_id = c.account_id
		WHERE c.id = $1 AND c.account_id = $2 AND c.provider = 'supabase' AND am.user_id = $3`,
		connectionID, accountID, userID).Scan(&encryptedToken, &metadata, &role)
	if err == sql.ErrNoRows {
		return "", nil, "", errSupabaseConnectionNotFound
	}
	if err != nil {
		return "", nil, "", err
	}

	decrypted, err := utils.Decrypt(encryptedToken)
	if err != nil {
		return "", nil, "", err
	}
	var bundle utils.OAuthTokenBundle
	if err := json.Unmarshal([]byte(decrypted), &bundle); err != nil {
		return "", nil, "", err
	}
	if !bundle.ExpiresAt.IsZero() && time.Now().Add(time.Minute).After(bundle.ExpiresAt) {
		bundle, err = refreshSupabaseToken(ctx, bundle)
		if err != nil {
			return "", nil, "", err
		}
		if err := saveSupabaseToken(ctx, db, connectionID, bundle); err != nil {
			return "", nil, "", err
		}
	}

	var stored struct {
		Projects []SupabaseProject `json:"projects"`
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &stored); err != nil {
			return "", nil, "", err
		}
	}
	selected := make(map[string]bool, len(stored.Projects))
	for _, project := range stored.Projects {
		selected[project.Ref] = true
	}
	return bundle.AccessToken, selected, role, nil
}

func fetchSupabaseProjects(ctx context.Context, accessToken string) ([]SupabaseProject, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, supabaseManagementAPI+"/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase projects returned status %d", response.StatusCode)
	}

	var projects []SupabaseProject
	if err := json.NewDecoder(response.Body).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func refreshSupabaseToken(ctx context.Context, current utils.OAuthTokenBundle) (utils.OAuthTokenBundle, error) {
	if current.RefreshToken == "" {
		return utils.OAuthTokenBundle{}, fmt.Errorf("supabase refresh token is missing")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {current.RefreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, utils.Cfg.SupabaseOAuth.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(utils.Cfg.SupabaseOAuth.ClientID, utils.Cfg.SupabaseOAuth.ClientSecret)

	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	defer response.Body.Close()

	var token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return utils.OAuthTokenBundle{}, err
	}
	if response.StatusCode >= 300 || token.AccessToken == "" {
		return utils.OAuthTokenBundle{}, fmt.Errorf("supabase token refresh failed: %s", token.Error)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = current.RefreshToken
	}
	return utils.OAuthTokenBundle{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Scope:        token.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}, nil
}

func saveSupabaseToken(ctx context.Context, db *sql.DB, connectionID string, bundle utils.OAuthTokenBundle) error {
	payload, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	encrypted, err := utils.Encrypt(string(payload))
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `UPDATE connections SET secret_ref = $1, updated_at = NOW() WHERE id = $2`, encrypted, connectionID)
	return err
}
