package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eryajf/go-ldap-admin/config"
	_ "github.com/eryajf/go-ldap-admin/docs"
	"github.com/eryajf/go-ldap-admin/middleware"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/public/static"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRoutes initializes the Gin engine with all middleware and registers all route groups.
func InitRoutes() *gin.Engine {
	// Set Gin mode from configuration
	gin.SetMode(config.Conf.System.Mode)

	// Create router with default middleware (logger & recovery)
	r := gin.Default()
	// Create router without middleware:
	// r := gin.New()
	// r.Use(gin.Recovery())

	r.Use(middleware.Serve("/", middleware.EmbedFolder(static.Static, "dist")))
	r.NoRoute(func(c *gin.Context) {
		data, err := static.Static.ReadFile("dist/index.html")
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Enable rate-limiting middleware
	// Default: refill one token every 50ms, bucket capacity from config
	fillInterval := time.Duration(config.Conf.RateLimit.FillInterval)
	capacity := config.Conf.RateLimit.Capacity
	r.Use(middleware.RateLimitMiddleware(time.Millisecond*fillInterval, capacity))

	// Enable global CORS middleware
	r.Use(middleware.CORSMiddleware())

	// Enable operation logging middleware
	r.Use(middleware.OperationLogMiddleware())

	// Initialize JWT authentication middleware
	authMiddleware, err := middleware.InitAuth()
	if err != nil {
		common.Log.Panicf("初始化JWT中间件失败：%v", err)
		panic(fmt.Sprintf("初始化JWT中间件失败：%v", err))
	}
	// swag
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	apiGroup := r.Group("/" + config.Conf.System.UrlPathPrefix)
	// swag
	// Register routes
	InitBaseRoutes(apiGroup, authMiddleware)          // Register base routes (no JWT, no Casbin)
	InitUserRoutes(apiGroup, authMiddleware)          // Register user routes (JWT + Casbin)
	InitGroupRoutes(apiGroup, authMiddleware)         // Register group routes (JWT + Casbin)
	InitRoleRoutes(apiGroup, authMiddleware)          // Register role routes (JWT + Casbin)
	InitMenuRoutes(apiGroup, authMiddleware)          // Register menu routes (JWT + Casbin)
	InitApiRoutes(apiGroup, authMiddleware)           // Register API routes (JWT + Casbin)
	InitOperationLogRoutes(apiGroup, authMiddleware)  // Register operation log routes (JWT + Casbin)
	InitFieldRelationRoutes(apiGroup, authMiddleware) // Register field relation routes (JWT + Casbin)
	InitOAuthRoutes(apiGroup, authMiddleware)         // Register OAuth routes

	common.Log.Info("初始化路由完成！")
	return r
}
