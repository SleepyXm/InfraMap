package routes

import (
	"database/sql"

	"InfraMap/middleware"
	metrics "InfraMap/services/metrics"

	"github.com/gin-gonic/gin"
)

func RegisterMetricRoutes(rg *gin.RouterGroup, db *sql.DB) {
	rg.GET("/:accountID/metrics/:workspaceID", middleware.AuthMiddleware(db), metrics.GetMetrics(db))
}
