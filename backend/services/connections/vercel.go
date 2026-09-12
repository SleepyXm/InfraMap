package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"InfraMap/utils"

	"github.com/gin-gonic/gin"
)

const vercelAPIURL = "https://api.vercel.com"

var errVercelConnectionNotFound = errors.New("vercel connection not found")

type VercelProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework"`
	AccountID string `json:"accountId"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	Selected  bool   `json:"selected"`
}

type VercelDeployment struct {
	UID          string                 `json:"uid"`
	Name         string                 `json:"name"`
	URL          string                 `json:"url"`
	State        string                 `json:"state"`
	ReadyState   string                 `json:"readyState"`
	Target       string                 `json:"target"`
	Created      int64                  `json:"created"`
	BuildingAt   int64                  `json:"buildingAt"`
	Ready        int64                  `json:"ready"`
	Creator      map[string]interface{} `json:"creator"`
	Meta         map[string]interface{} `json:"meta"`
	InspectorURL string                 `json:"inspectorUrl"`
}

type updateVercelProjectsRequest struct {
	ProjectIDs []string `json:"project_ids"`
}

func GetVercelProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")
		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		token, teamID, selected, _, err := LoadVercelConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errVercelConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Vercel connection"})
			return
		}
		projects, err := FetchVercelProjects(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel projects could not be loaded"})
			return
		}
		for index := range projects {
			projects[index].Selected = selected[projects[index].ID]
		}
		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

func UpdateVercelProjects(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")
		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		var request updateVercelProjectsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose the Vercel projects to attach"})
			return
		}
		if len(request.ProjectIDs) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many projects selected"})
			return
		}

		token, teamID, _, role, err := LoadVercelConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errVercelConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the Vercel connection"})
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can change Vercel project access"})
			return
		}

		projects, err := FetchVercelProjects(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel projects could not be verified"})
			return
		}
		available := make(map[string]VercelProject, len(projects))
		for _, project := range projects {
			available[project.ID] = project
		}
		selected := make([]VercelProject, 0, len(request.ProjectIDs))
		seen := make(map[string]bool, len(request.ProjectIDs))
		for _, projectID := range request.ProjectIDs {
			project, ok := available[projectID]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "A selected Vercel project is no longer accessible"})
				return
			}
			if seen[projectID] {
				continue
			}
			seen[projectID] = true
			project.Selected = true
			selected = append(selected, project)
		}

		payload, err := json.Marshal(selected)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Vercel project selection"})
			return
		}
		result, err := db.ExecContext(c, `UPDATE connections
			SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('projects', $1::jsonb),
				updated_at = NOW()
			WHERE id = $2 AND account_id = $3 AND provider = 'vercel'`, string(payload), connectionID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save Vercel project selection"})
			return
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != 1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vercel connection not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": selected})
	}
}

func LoadVercelConnection(ctx context.Context, db *sql.DB, userID, accountID, connectionID string) (string, string, map[string]bool, string, error) {
	var encryptedToken string
	var metadata []byte
	var role string
	err := db.QueryRowContext(ctx, `SELECT c.secret_ref, c.metadata, am.role
		FROM connections c
		JOIN account_memberships am ON am.account_id = c.account_id
		WHERE c.id = $1 AND c.account_id = $2 AND c.provider = 'vercel' AND am.user_id = $3`,
		connectionID, accountID, userID).Scan(&encryptedToken, &metadata, &role)
	if err == sql.ErrNoRows {
		return "", "", nil, "", errVercelConnectionNotFound
	}
	if err != nil {
		return "", "", nil, "", err
	}
	decrypted, err := utils.Decrypt(encryptedToken)
	if err != nil {
		return "", "", nil, "", err
	}
	var bundle utils.OAuthTokenBundle
	if err := json.Unmarshal([]byte(decrypted), &bundle); err != nil {
		return "", "", nil, "", err
	}
	var stored struct {
		TeamID   string          `json:"team_id"`
		Projects []VercelProject `json:"projects"`
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &stored); err != nil {
			return "", "", nil, "", err
		}
	}
	selected := make(map[string]bool, len(stored.Projects))
	for _, project := range stored.Projects {
		selected[project.ID] = true
	}
	return bundle.AccessToken, stored.TeamID, selected, role, nil
}

func FetchVercelProjects(ctx context.Context, accessToken, teamID string) ([]VercelProject, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	projects := make([]VercelProject, 0)
	var until int64
	for page := 0; page < 10; page++ {
		query := url.Values{"limit": {"100"}}
		if teamID != "" {
			query.Set("teamId", teamID)
		}
		if until > 0 {
			query.Set("until", strconv.FormatInt(until, 10))
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, vercelAPIURL+"/v9/projects?"+query.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if response.StatusCode >= 300 {
			response.Body.Close()
			return nil, fmt.Errorf("vercel projects returned status %d", response.StatusCode)
		}
		var result struct {
			Projects   []VercelProject `json:"projects"`
			Pagination struct {
				Next *int64 `json:"next"`
			} `json:"pagination"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&result)
		response.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		projects = append(projects, result.Projects...)
		if result.Pagination.Next == nil || *result.Pagination.Next == 0 {
			break
		}
		until = *result.Pagination.Next
	}
	return projects, nil
}

func FetchVercelDeployments(ctx context.Context, accessToken, teamID, projectID string) ([]VercelDeployment, error) {
	query := url.Values{"limit": {"20"}, "projectId": {projectID}}
	if teamID != "" {
		query.Set("teamId", teamID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vercelAPIURL+"/v6/deployments?"+query.Encode(), nil)
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
		return nil, fmt.Errorf("vercel deployments returned status %d", response.StatusCode)
	}
	var result struct {
		Deployments []VercelDeployment `json:"deployments"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Deployments, nil
}
