package routes

import (
	"database/sql"

	"InfraMap/middleware"
	auth "InfraMap/services/auth"
	connections "InfraMap/services/connections"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.POST("/signup", auth.Signup(db))
	rg.POST("/login", auth.Login(db))
	rg.POST("/refresh", auth.Refresh())
	rg.POST("/logout", auth.Logout())
	rg.GET("/oauth/:provider/start", auth.OAuthStart())
	rg.GET("/callback", auth.OAuthCallback(db))
	rg.GET("/me", middleware.AuthMiddleware(db), auth.Me(db))
	rg.GET("/:accountID/connections", middleware.AuthMiddleware(db), connections.GetConnections(db))
	rg.POST("/:accountID/connections", middleware.AuthMiddleware(db), connections.CreateConnection(db))
	rg.GET("/:accountID/connections/oauth/:provider/start", middleware.AuthMiddleware(db), auth.OAuthConnectionStart(db))
	rg.GET("/:accountID/connections/:connectionID", middleware.AuthMiddleware(db), connections.GetConnection(db))
	rg.GET("/:accountID/connections/:connectionID/github/repositories", middleware.AuthMiddleware(db), connections.GetGitHubRepositories(db))
	rg.PUT("/:accountID/connections/:connectionID/github/repositories", middleware.AuthMiddleware(db), connections.UpdateGitHubRepositories(db))
	rg.GET("/:accountID/connections/:connectionID/supabase/projects", middleware.AuthMiddleware(db), connections.GetSupabaseProjects(db))
	rg.PUT("/:accountID/connections/:connectionID/supabase/projects", middleware.AuthMiddleware(db), connections.UpdateSupabaseProjects(db))
	rg.DELETE("/:accountID/connections/:connectionID", middleware.AuthMiddleware(db), connections.DeleteConnection(db))
	rg.GET("/hi", auth.Hi())
}
