package routes

import (
	"github.com/eryajf/go-ldap-admin/controller"
	"github.com/eryajf/go-ldap-admin/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// InitUserRoutes registers user routes with JWT and Casbin middleware.
func InitUserRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	user := r.Group("/user")
	// Enable JWT authentication middleware
	user.Use(authMiddleware.MiddlewareFunc())
	// Enable Casbin authorization middleware
	user.Use(middleware.CasbinMiddleware())
	{
		user.GET("/info", controller.User.GetUserInfo)                             // Get user info (not yet complete)
		user.GET("/list", controller.User.List)                                    // List users
		user.POST("/add", controller.User.Add)                                     // Add user
		user.POST("/update", controller.User.Update)                               // Update user
		user.POST("/delete", controller.User.Delete)                               // Delete user
		user.POST("/changePwd", controller.User.ChangePwd)                         // Change user password
		user.POST("/password/code", controller.User.SendPasswordChangeCode)        // Send password change verification code
		user.POST("/issueSSHPubKey", controller.User.IssueSSHPubKey)               // Issue SSH certificate for current user
		user.POST("/revokeOriginalKeypair", controller.User.RevokeOriginalKeypair) // Revoke old SSH certificate/key (placeholder)
		user.POST("/resetPassword", controller.User.ResetPassword)                 // Reset user password
		user.POST("/changeUserStatus", controller.User.ChangeUserStatus)           // Change user status

		user.POST("/syncDingTalkUsers", controller.User.SyncDingTalkUsers) // Sync DingTalk users to platform
		user.POST("/syncWeComUsers", controller.User.SyncWeComUsers)       // Sync WeCom users to platform
		user.POST("/syncFeiShuUsers", controller.User.SyncFeiShuUsers)     // Sync FeiShu users to platform
		user.POST("/syncOpenLdapUsers", controller.User.SyncOpenLdapUsers) // Sync LDAP users to platform
		user.POST("/syncSqlUsers", controller.User.SyncSqlUsers)           // Sync SQL users to LDAP
	}
	return r
}
