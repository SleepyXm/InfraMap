package routes

import (
	"database/sql"

	"InfraMap/middleware"
	workspaces "InfraMap/services/workspaces"

	"github.com/gin-gonic/gin"
)

func RegisterServiceRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/services/:workspaceID", middleware.AuthMiddleware(db), workspaces.GetServices(db))
	rg.POST("/:accountID/services/:workspaceID/vercel", middleware.AuthMiddleware(db), workspaces.AttachVercelService(db))
	rg.POST("/:accountID/services/:workspaceID/supabase", middleware.AuthMiddleware(db), workspaces.AttachSupabaseService(db))
	rg.POST("/:accountID/services/:workspaceID/supabase/provision", middleware.AuthMiddleware(db), workspaces.CreateSupabaseService(db))
	rg.GET("/:accountID/services/:workspaceID/aws/cloudformation", middleware.AuthMiddleware(db), workspaces.GetAWSCloudFormation(db))
	rg.POST("/:accountID/services/:workspaceID/aws", middleware.AuthMiddleware(db), workspaces.AttachAWSHost(db))
}
