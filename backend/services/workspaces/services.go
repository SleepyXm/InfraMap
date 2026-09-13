package workspaces

import (
	"database/sql"
	"net/http"

	"InfraMap/structs"

	"github.com/gin-gonic/gin"
)

func GetServices(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, accountID, workspaceID := c.MustGet("userID").(string), c.Param("accountID"), c.Param("workspaceID")
		if _, _, err := LoadWorkspace(c, db, userID, accountID, workspaceID); err != nil {
			WriteWorkspaceAccessError(c, err)
			return
		}
		resources, err := LoadResources(c, db, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Workspace services could not be loaded"})
			return
		}
		services := make([]structs.WorkspaceResource, 0)
		for _, resource := range resources {
			if resource.Provider != "github" {
				services = append(services, resource)
			}
		}
		c.JSON(http.StatusOK, gin.H{"services": services})
	}
}
