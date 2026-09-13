package routes

import (
	"database/sql"

	"InfraMap/middleware"
	scan "InfraMap/services/scan"

	"github.com/gin-gonic/gin"
)

func RegisterScanRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/scans/:workspaceID", middleware.AuthMiddleware(db), scan.GetEZDeployAnalysis(db))
	rg.POST("/:accountID/scans/:workspaceID", middleware.AuthMiddleware(db), scan.RunEZDeployAnalysis(db))
	rg.PUT("/:accountID/scans/:workspaceID/plan", middleware.AuthMiddleware(db), scan.SaveEZDeployDecisions(db))
}
