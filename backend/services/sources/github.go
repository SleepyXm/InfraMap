package sources

import (
	"database/sql"
	"net/http"
	"strconv"

	connections "InfraMap/services/connections"
	workspaces "InfraMap/services/workspaces"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AttachGitHubRepository(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		_, role, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID)
		if err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot attach workspace sources"})
			return
		}

		var request structs.AttachGitHubRepositoryRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) || request.RepositoryID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose a valid GitHub repository"})
			return
		}
		token, _, _, err := connections.LoadGitHubConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub connection could not be opened"})
			return
		}
		repositories, err := connections.FetchGitHubRepositories(c, token)
		if err != nil {
			connections.WriteGitHubRepositoryError(c, err, "verified")
			return
		}

		var selected *connections.GitHubRepository
		for index := range repositories {
			if repositories[index].ID == request.RepositoryID {
				selected = &repositories[index]
				break
			}
		}
		if selected == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That GitHub repository is no longer accessible"})
			return
		}

		resource, err := workspaces.SaveResource(c, db, userID, workspaceID, workspaces.ResourceInput{
			ConnectionID: request.ConnectionID,
			Provider:     "github",
			ResourceType: "repository",
			ExternalID:   strconv.FormatInt(selected.ID, 10),
			Name:         selected.FullName,
			Status:       "active",
			Metadata: map[string]interface{}{
				"name":           selected.Name,
				"private":        selected.Private,
				"description":    selected.Description,
				"default_branch": selected.DefaultBranch,
				"language":       selected.Language,
				"html_url":       selected.HTMLURL,
				"clone_url":      selected.HTMLURL + ".git",
				"pushed_at":      selected.PushedAt,
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub repository could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"source": resource})
	}
}

func validUUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func GetSources(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		if _, _, err := workspaces.LoadWorkspace(c, db, userID, accountID, workspaceID); err != nil {
			workspaces.WriteWorkspaceAccessError(c, err)
			return
		}
		resources, err := workspaces.LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace sources could not be loaded"})
			return
		}
		sources := make([]structs.WorkspaceResource, 0)
		for _, resource := range resources {
			if resource.Provider == "github" {
				sources = append(sources, resource)
			}
		}
		c.JSON(http.StatusOK, gin.H{"sources": sources})
	}
}
