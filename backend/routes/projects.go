package routes

import (
	"database/sql"

	"InfraMap/middleware"
	projects "InfraMap/services/projects"

	"github.com/gin-gonic/gin"
)

func RegisterProjectRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/projects", middleware.AuthMiddleware(db), projects.GetProjects(db))
	rg.POST("/:accountID/projects", middleware.AuthMiddleware(db), projects.CreateProject(db))
	rg.GET("/:accountID/projects/:projectID", middleware.AuthMiddleware(db), projects.GetProject(db))
	rg.GET("/:accountID/projects/:projectID/control", middleware.AuthMiddleware(db), projects.GetProjectControl(db))
	rg.POST("/:accountID/projects/:projectID/resources/github/attach", middleware.AuthMiddleware(db), projects.AttachGitHubRepository(db))
	rg.POST("/:accountID/projects/:projectID/resources/vercel/attach", middleware.AuthMiddleware(db), projects.AttachVercelProject(db))
	rg.POST("/:accountID/projects/:projectID/resources/supabase/attach", middleware.AuthMiddleware(db), projects.AttachSupabaseProject(db))
	rg.POST("/:accountID/projects/:projectID/resources/supabase/create", middleware.AuthMiddleware(db), projects.CreateSupabaseProject(db))
}
