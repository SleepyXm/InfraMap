package projects

import (
	"database/sql"
	"net/http"

	connections "InfraMap/services/connections"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

func AttachVercelProject(db *sql.DB) gin.HandlerFunc {
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

		var request structs.AttachVercelProjectRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) || request.ProjectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose a valid Vercel project"})
			return
		}
		token, teamID, _, _, err := connections.LoadVercelConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Vercel connection could not be opened"})
			return
		}
		projects, err := connections.FetchVercelProjects(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel projects could not be verified"})
			return
		}

		var selected *connections.VercelProject
		for index := range projects {
			if projects[index].ID == request.ProjectID {
				selected = &projects[index]
				break
			}
		}
		if selected == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That Vercel project is no longer accessible"})
			return
		}

		resource, err := saveProjectResource(c, db, userID, projectID, projectResourceInput{
			ConnectionID: request.ConnectionID,
			Provider:     "vercel",
			ResourceType: "project",
			ExternalID:   selected.ID,
			Name:         selected.Name,
			Status:       "active",
			Metadata: map[string]interface{}{
				"framework":  selected.Framework,
				"account_id": selected.AccountID,
				"created_at": selected.CreatedAt,
				"updated_at": selected.UpdatedAt,
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Vercel project could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resource": resource})
	}
}
