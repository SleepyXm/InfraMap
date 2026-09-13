package workspaces

import (
	"database/sql"
	"net/http"

	connections "InfraMap/services/connections"
	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

func AttachVercelService(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		accountID := c.Param("accountID")
		workspaceID := c.Param("workspaceID")
		if !requireServiceWriteAccess(c, db, userID, accountID, workspaceID, false) {
			return
		}

		var request structs.AttachVercelServiceRequest
		if err := c.ShouldBindJSON(&request); err != nil || !validUUID(request.ConnectionID) || request.ServiceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Choose a valid Vercel service"})
			return
		}
		token, teamID, _, _, err := connections.LoadVercelConnection(c, db, userID, accountID, request.ConnectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Vercel connection could not be opened"})
			return
		}
		available, err := connections.FetchVercelServices(c, token, teamID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Vercel services could not be verified"})
			return
		}

		var selected *connections.VercelService
		for index := range available {
			if available[index].ID == request.ServiceID {
				selected = &available[index]
				break
			}
		}
		if selected == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "That Vercel service is no longer accessible"})
			return
		}

		resource, err := SaveResource(c, db, userID, workspaceID, ResourceInput{
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Vercel service could not be attached"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"service": resource})
	}
}
