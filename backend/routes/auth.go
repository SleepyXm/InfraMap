package routes

import (
	"database/sql"

	"InfraMap/middleware"
	auth "InfraMap/services/auth"

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
	rg.GET("/hi", auth.Hi())
}
