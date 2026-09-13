package routes

import (
	"database/sql"

	"InfraMap/middleware"
	deployments "InfraMap/services/deployments"

	"github.com/gin-gonic/gin"
)

func RegisterDeploymentRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.POST("/:accountID/deployments/:workspaceID", middleware.AuthMiddleware(db), deployments.Create(db))
	rg.GET("/:accountID/deployments/:workspaceID/:commandID", middleware.AuthMiddleware(db), deployments.Get(db))
}
