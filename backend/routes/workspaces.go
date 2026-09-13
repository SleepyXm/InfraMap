package routes

import (
	"database/sql"

	"InfraMap/middleware"
	workspaces "InfraMap/services/workspaces"

	"github.com/gin-gonic/gin"
)

func RegisterWorkspaceRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/workspaces", middleware.AuthMiddleware(db), workspaces.GetWorkspaces(db))
	rg.POST("/:accountID/workspaces", middleware.AuthMiddleware(db), workspaces.CreateWorkspace(db))
	rg.GET("/:accountID/workspaces/:workspaceID", middleware.AuthMiddleware(db), workspaces.GetWorkspace(db))
	rg.DELETE("/:accountID/workspaces/:workspaceID", middleware.AuthMiddleware(db), workspaces.DeleteWorkspace(db))
}
