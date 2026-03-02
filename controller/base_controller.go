package controller

import (
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/public/tools"
	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

type BaseController struct{}

// SendCode sends a verification code to the user's email
// @Summary Send verification code
// @Description Send a verification code to the specified email address
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param data body request.BaseSendCodeReq true "Send verification code request data"
// @Success 200 {object} response.ResponseBody
// @Router /base/sendcode [post]
func (m *BaseController) SendCode(c *gin.Context) {
	req := new(request.BaseSendCodeReq)
	Run(c, req, func() (any, any) {
		return logic.Base.SendCode(c, req)
	})
}

// ChangePwd allows a user to change their password via email
// @Summary Change password via email
// @Description Change password using an email verification code
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.BaseChangePwdReq true "Change password request data"
// @Success 200 {object} response.ResponseBody
// @Router /base/changePwd [post]
func (m *BaseController) ChangePwd(c *gin.Context) {
	req := new(request.BaseChangePwdReq)
	Run(c, req, func() (any, any) {
		return logic.Base.ChangePwd(c, req)
	})
}

// SendLoginCode sends an OTP login verification code to the user's email
// @Summary Send OTP login verification code
// @Description Send an OTP login verification code to the user's email address
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param data body request.BaseSendLoginCodeReq true "Send login code request data"
// @Success 200 {object} response.ResponseBody
// @Router /base/otp/send [post]
func (m *BaseController) SendLoginCode(c *gin.Context) {
	req := new(request.BaseSendLoginCodeReq)
	Run(c, req, func() (any, any) {
		return logic.Base.SendLoginCode(c, req)
	})
}

// OtpLogin handles login using an OTP verification code
// @Summary Login using OTP verification code
// @Description Authenticate a user via an OTP verification code sent to their email
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param data body request.BaseOtpLoginReq true "OTP login request data"
// @Success 200 {object} response.ResponseBody
// @Router /base/otp/login [post]
func (m *BaseController) OtpLogin(auth *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(request.BaseOtpLoginReq)
		if err := c.Bind(req); err != nil {
			tools.Err(c, tools.NewValidatorError(err), nil)
			return
		}
		if err := validate.Struct(req); err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				tools.Err(c, tools.NewValidatorError(fmt.Errorf("%s", err.Translate(trans))), nil)
				return
			}
		}
		user, err := logic.Base.VerifyOtpLogin(c, req)
		if err != nil {
			tools.Err(c, tools.ReloadErr(err), nil)
			return
		}

		// If this is a password reset flow, generate a new password and update
		var newPassword string
		if req.Reset {
			newPass, resetErr := logic.Base.ResetUserPassword(user)
			if resetErr != nil {
				if rspErr, ok := resetErr.(*tools.RspError); ok {
					tools.Err(c, rspErr, nil)
				} else {
					tools.Err(c, tools.NewValidatorError(fmt.Errorf("%v", resetErr)), nil)
				}
				return
			}
			newPassword = newPass
		}

		payload := tools.H{"user": tools.Struct2Json(user)}
		claims := auth.PayloadFunc(payload)
		token, expires, tokenErr := auth.TokenGenerator(claims)
		if tokenErr != nil {
			tools.Err(c, tools.NewValidatorError(tokenErr), nil)
			return
		}

		if newPassword != "" {
			// Password reset flow: return token + newPassword
			response.Response(c, http.StatusOK, http.StatusOK,
				gin.H{
					"token":       token,
					"expires":     expires.Format("2006-01-02 15:04:05"),
					"newPassword": newPassword,
				},
				"重置密码成功")
		} else {
			auth.LoginResponse(c, http.StatusOK, token, expires)
		}
	}
}

// Dashboard returns the system homepage display data
// @Summary Get dashboard data
// @Description Get system dashboard overview data
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /base/dashboard [get]
func (m *BaseController) Dashboard(c *gin.Context) {
	req := new(request.BaseDashboardReq)
	Run(c, req, func() (any, any) {
		return logic.Base.Dashboard(c, req)
	})
}

// EncryptPasswd encrypts a plaintext password
// @Summary Encrypt password
// @Description Encrypt a plaintext password
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param passwd query string true "Plaintext password to encrypt"
// @Success 200 {object} response.ResponseBody
// @Router /base/encryptpwd [get]
func (m *BaseController) EncryptPasswd(c *gin.Context) {
	req := new(request.EncryptPasswdReq)
	Run(c, req, func() (any, any) {
		return logic.Base.EncryptPasswd(c, req)
	})
}

// DecryptPasswd decrypts an encrypted password to plaintext
// @Summary Decrypt password
// @Description Decrypt an encrypted password to plaintext
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Param passwd query string true "Encrypted password to decrypt"
// @Success 200 {object} response.ResponseBody
// @Router /base/decryptpwd [get]
func (m *BaseController) DecryptPasswd(c *gin.Context) {
	req := new(request.DecryptPasswdReq)
	Run(c, req, func() (any, any) {
		return logic.Base.DecryptPasswd(c, req)
	})
}

// GetConfig retrieves the system configuration
// @Summary Get system configuration
// @Description Get system configuration, used by the frontend to determine whether to show the sync button
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /base/config [get]
func (m *BaseController) GetConfig(c *gin.Context) {
	req := new(request.BaseConfigReq)
	Run(c, req, func() (any, any) {
		return logic.Base.GetConfig(c, req)
	})
}

// GetVersion retrieves the version information
// @Summary Get version information
// @Description Get the system version number, Git commit hash, and build time
// @Tags Base Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /base/version [get]
func (m *BaseController) GetVersion(c *gin.Context) {
	req := new(request.BaseVersionReq)
	Run(c, req, func() (any, any) {
		return logic.Base.GetVersion(c, req)
	})
}
