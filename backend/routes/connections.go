package routes

import (
	"database/sql"

	"InfraMap/middleware"
	auth "InfraMap/services/auth"
	connections "InfraMap/services/connections"

	"github.com/gin-gonic/gin"
)

func RegisterConnectionRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/connections", middleware.AuthMiddleware(db), connections.GetConnections(db))
	rg.POST("/:accountID/connections", middleware.AuthMiddleware(db), connections.CreateConnection(db))
	rg.GET("/:accountID/connections/oauth/:provider/start", middleware.AuthMiddleware(db), auth.OAuthConnectionStart(db))
	rg.GET("/:accountID/connections/:connectionID", middleware.AuthMiddleware(db), connections.GetConnection(db))
	rg.GET("/:accountID/connections/:connectionID/github/repositories", middleware.AuthMiddleware(db), connections.GetGitHubRepositories(db))
	rg.PUT("/:accountID/connections/:connectionID/github/repositories", middleware.AuthMiddleware(db), connections.UpdateGitHubRepositories(db))
	rg.GET("/:accountID/connections/:connectionID/supabase/services", middleware.AuthMiddleware(db), connections.GetSupabaseServices(db))
	rg.PUT("/:accountID/connections/:connectionID/supabase/services", middleware.AuthMiddleware(db), connections.UpdateSupabaseServices(db))
	rg.GET("/:accountID/connections/:connectionID/vercel/services", middleware.AuthMiddleware(db), connections.GetVercelServices(db))
	rg.PUT("/:accountID/connections/:connectionID/vercel/services", middleware.AuthMiddleware(db), connections.UpdateVercelServices(db))
	rg.DELETE("/:accountID/connections/:connectionID", middleware.AuthMiddleware(db), connections.DeleteConnection(db))
}
