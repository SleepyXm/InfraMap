package projects

import (
	"database/sql"
	"net/http"
	"strconv"

	connections "InfraMap/services/connections"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

func AttachGitHubRepository(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		projectID := c.Param("projectID")
		_, role, err := loadProject(c, db, userID, accountID, projectID)
		if err != nil {
			writeProjectAccessError(c, err)
			return
		}
		if role == "viewer" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Viewers cannot attach project resources"})
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
			c.JSON(http.StatusBadGateway, gin.H{"error": "GitHub repositories could not be verified"})
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

		resource, err := saveProjectResource(c, db, userID, projectID, projectResourceInput{
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
				"pushed_at":      selected.PushedAt,
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub repository could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resource": resource})
	}
}
