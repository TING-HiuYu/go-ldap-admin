package routes

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/eryajf/go-ldap-admin/controller"
	"github.com/gin-gonic/gin"
)

// LoginHandler
// @Summary Login (manually add: Bearer + token (password encryption endpoint))
// @Description User login
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RegisterAndLoginReq true "User login credentials (username and password)"
// @Success 200 {object} response.ResponseBody
// @Router /base/login [post]
func LoginHandler() {}

// LogoutHandler
// @Summary Logout
// @Description User logout
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /base/logout [post]
func LogoutHandler() {
}

// RefreshHandler
// @Summary Refresh Token
// @Description Use an old Token to obtain a new Token
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer old Token"
// @Success 200 {object} response.ResponseBody
// @Router /base/refreshToken [post]
func RefreshHandler() {

}

// InitBaseRoutes registers base routes that do not require JWT or Casbin middleware.
func InitBaseRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	base := r.Group("/base")
	{
		base.GET("ping", controller.Demo)
		base.GET("encryptpwd", controller.Base.EncryptPasswd) // Generate encrypted password
		base.GET("decryptpwd", controller.Base.DecryptPasswd) // Decrypt password to plaintext
		base.GET("config", controller.Base.GetConfig)         // Get system configuration
		base.GET("version", controller.Base.GetVersion)       // Get version info
		// Login, logout, and token refresh do not require authentication
		base.POST("/login", authMiddleware.LoginHandler)
		base.POST("/otp/send", controller.Base.SendLoginCode) // Send login verification code
		base.POST("/otp/login", controller.Base.OtpLogin(authMiddleware))
		base.POST("/logout", authMiddleware.LogoutHandler)
		base.POST("/refreshToken", authMiddleware.RefreshHandler)
		base.POST("/sendcode", controller.Base.SendCode)   // Send verification code to user email
		base.POST("/changePwd", controller.Base.ChangePwd) // Change user password
		base.GET("/dashboard", controller.Base.Dashboard)  // Dashboard data for system homepage
	}
	return r
}
