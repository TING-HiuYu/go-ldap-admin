package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/eryajf/go-ldap-admin/controller"
	"github.com/eryajf/go-ldap-admin/middleware"
	"github.com/gin-gonic/gin"
)

// InitOAuthRoutes registers OAuth-related routes.
func InitOAuthRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	oauth := r.Group("/oauth")
	{
		oauth.GET("/connectors/public", controller.OAuth.PublicConnectors)
		oauth.POST("/start", controller.OAuth.StartLogin)
		oauth.GET("/callback/:provider", controller.OAuth.Callback(authMiddleware))
	}

	secured := oauth.Group("")
	secured.Use(authMiddleware.MiddlewareFunc())
	secured.Use(middleware.CasbinMiddleware())
	{
		secured.GET("/connectors", controller.OAuth.ListConnectors)
		secured.POST("/connectors", controller.OAuth.CreateConnector)
		secured.POST("/connectors/update", controller.OAuth.UpdateConnector)
		secured.POST("/connectors/delete", controller.OAuth.DeleteConnector)
		secured.POST("/start/password", controller.OAuth.StartPasswordVerify)
	}
	return r
}
