package routes

import (
	"database/sql"

	"InfraMap/middleware"
	sources "InfraMap/services/sources"

	"github.com/gin-gonic/gin"
)

func RegisterSourceRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/sources/:workspaceID", middleware.AuthMiddleware(db), sources.GetSources(db))
	rg.POST("/:accountID/sources/:workspaceID/github", middleware.AuthMiddleware(db), sources.AttachGitHubRepository(db))
}
