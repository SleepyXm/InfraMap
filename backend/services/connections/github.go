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
	"github.com/google/uuid"
)

const githubAPIURL = "https://api.github.com"

var errGitHubConnectionNotFound = errors.New("github connection not found")

type GitHubRepository struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	FullName      string     `json:"full_name"`
	Private       bool       `json:"private"`
	Description   string     `json:"description"`
	DefaultBranch string     `json:"default_branch"`
	Language      string     `json:"language"`
	HTMLURL       string     `json:"html_url"`
	PushedAt      *time.Time `json:"pushed_at"`
	Selected      bool       `json:"selected"`
}

type updateGitHubRepositoriesRequest struct {
	RepositoryIDs []int64 `json:"repository_ids"`
}

func GetGitHubRepositories(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		token, selected, _, err := LoadGitHubConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errGitHubConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the GitHub connection"})
			return
		}

		repositories, err := FetchGitHubRepositories(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "GitHub repositories could not be loaded"})
			return
		}
		for index := range repositories {
			repositories[index].Selected = selected[repositories[index].ID]
		}

		c.JSON(http.StatusOK, gin.H{"repositories": repositories})
	}
}

func UpdateGitHubRepositories(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		connectionID := c.Param("connectionID")

		if !validConnectionIDs(accountID, connectionID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace or connection ID"})
			return
		}

		var request updateGitHubRepositoriesRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose the repositories to attach"})
			return
		}
		if len(request.RepositoryIDs) > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Too many repositories selected"})
			return
		}

		token, _, role, err := LoadGitHubConnection(c, db, userID, accountID, connectionID)
		if errors.Is(err, errGitHubConnectionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub connection not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open the GitHub connection"})
			return
		}
		if role != "owner" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only workspace owners and admins can change repository access"})
			return
		}

		repositories, err := FetchGitHubRepositories(c, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "GitHub repositories could not be verified"})
			return
		}
		available := make(map[int64]GitHubRepository, len(repositories))
		for _, repository := range repositories {
			available[repository.ID] = repository
		}

		selected := make([]GitHubRepository, 0, len(request.RepositoryIDs))
		seen := make(map[int64]bool, len(request.RepositoryIDs))
		for _, repositoryID := range request.RepositoryIDs {
			repository, ok := available[repositoryID]
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "A selected repository is no longer accessible"})
				return
			}
			if seen[repositoryID] {
				continue
			}
			seen[repositoryID] = true
			repository.Selected = true
			selected = append(selected, repository)
		}

		payload, err := json.Marshal(selected)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save repository selection"})
			return
		}
		result, err := db.ExecContext(c, `UPDATE connections
			SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('repositories', $1::jsonb),
				updated_at = NOW()
			WHERE id = $2 AND account_id = $3 AND provider = 'github'`, string(payload), connectionID, accountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save repository selection"})
			return
		}
		rows, err := result.RowsAffected()
		if err != nil || rows != 1 {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub connection not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"repositories": selected})
	}
}

func validConnectionIDs(accountID, connectionID string) bool {
	_, accountErr := uuid.Parse(accountID)
	_, connectionErr := uuid.Parse(connectionID)
	return accountErr == nil && connectionErr == nil
}

func LoadGitHubConnection(
	ctx context.Context,
	db *sql.DB,
	userID string,
	accountID string,
	connectionID string,
) (string, map[int64]bool, string, error) {
	var encryptedToken string
	var metadata []byte
	var role string

	err := db.QueryRowContext(ctx, `SELECT c.secret_ref, c.metadata, am.role
		FROM connections c
		JOIN account_memberships am ON am.account_id = c.account_id
		WHERE c.id = $1 AND c.account_id = $2 AND c.provider = 'github' AND am.user_id = $3`,
		connectionID, accountID, userID).Scan(&encryptedToken, &metadata, &role)
	if err == sql.ErrNoRows {
		return "", nil, "", errGitHubConnectionNotFound
	}
	if err != nil {
		return "", nil, "", err
	}

	token, err := utils.Decrypt(encryptedToken)
	if err != nil {
		return "", nil, "", err
	}

	var stored struct {
		Repositories []GitHubRepository `json:"repositories"`
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &stored); err != nil {
			return "", nil, "", err
		}
	}
	selected := make(map[int64]bool, len(stored.Repositories))
	for _, repository := range stored.Repositories {
		selected[repository.ID] = true
	}

	return token, selected, role, nil
}

func FetchGitHubRepositories(ctx context.Context, token string) ([]GitHubRepository, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	repositories := make([]GitHubRepository, 0)

	for page := 1; page <= 10; page++ {
		query := url.Values{
			"affiliation": {"owner,collaborator,organization_member"},
			"direction":   {"desc"},
			"page":        {strconv.Itoa(page)},
			"per_page":    {"100"},
			"sort":        {"pushed"},
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIURL+"/user/repos?"+query.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if response.StatusCode >= 300 {
			response.Body.Close()
			return nil, fmt.Errorf("github returned status %d", response.StatusCode)
		}

		var pageRepositories []GitHubRepository
		decodeErr := json.NewDecoder(response.Body).Decode(&pageRepositories)
		response.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		repositories = append(repositories, pageRepositories...)
		if len(pageRepositories) < 100 {
			break
		}
	}

	return repositories, nil
}
