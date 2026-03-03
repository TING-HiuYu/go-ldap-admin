package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/service/isql"
	"github.com/gin-gonic/gin"
)

type OAuthController struct{}

// CreateConnector creates a new OAuth connector
// @Summary Create OAuth connector
// @Description Create a new OAuth connector configuration
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OAuthConnectorCreateReq true "OAuth connector creation payload"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/connectors [post]
// @Security ApiKeyAuth
func (m OAuthController) CreateConnector(c *gin.Context) {
	req := new(request.OAuthConnectorCreateReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Create(c, req)
	})
}

// UpdateConnector updates an existing OAuth connector
// @Summary Update OAuth connector
// @Description Update an existing OAuth connector configuration
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OAuthConnectorUpdateReq true "OAuth connector update payload"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/connectors/update [post]
// @Security ApiKeyAuth
func (m OAuthController) UpdateConnector(c *gin.Context) {
	req := new(request.OAuthConnectorUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Update(c, req)
	})
}

// DeleteConnector deletes OAuth connectors
// @Summary Delete OAuth connectors
// @Description Delete one or more OAuth connectors by IDs
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OAuthConnectorDeleteReq true "OAuth connector deletion payload"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/connectors/delete [post]
// @Security ApiKeyAuth
func (m OAuthController) DeleteConnector(c *gin.Context) {
	req := new(request.OAuthConnectorDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Delete(c, req)
	})
}

// ListConnectors lists all OAuth connectors for admin
// @Summary List OAuth connectors
// @Description List all OAuth connectors for the admin panel
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param provider query string false "Filter by provider"
// @Param enabled query bool false "Filter by enabled status"
// @Param keyword query string false "Search keyword"
// @Param pageNum query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/connectors [get]
// @Security ApiKeyAuth
func (m OAuthController) ListConnectors(c *gin.Context) {
	req := new(request.OAuthConnectorListReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.List(c, req)
	})
}

// PublicConnectors lists enabled OAuth connectors for the login page
// @Summary List public OAuth connectors
// @Description List enabled OAuth connectors visible on the login page
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param provider query string false "Filter by provider"
// @Param keyword query string false "Search keyword"
// @Param pageNum query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/connectors/public [get]
func (m OAuthController) PublicConnectors(c *gin.Context) {
	req := new(request.OAuthConnectorListReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.PublicList(c, req)
	})
}

// StartLogin starts the OAuth login flow for unauthenticated users
// @Summary Start OAuth login
// @Description Initiate an OAuth login flow by redirecting to the provider
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OAuthStartReq true "OAuth login start payload"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/start [post]
func (m OAuthController) StartLogin(c *gin.Context) {
	req := new(request.OAuthStartReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.StartAuth(c, req, 0)
	})
}

// StartPasswordVerify starts password verification via OAuth for logged-in users
// @Summary Start OAuth password verification
// @Description Initiate an OAuth flow to verify the current user's password
// @Tags OAuth Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OAuthStartReq true "OAuth password verification payload"
// @Success 200 {object} response.ResponseBody
// @Router /oauth/start/password [post]
// @Security ApiKeyAuth
func (m OAuthController) StartPasswordVerify(c *gin.Context) {
	req := new(request.OAuthStartReq)
	Run(c, req, func() (any, any) {
		user, err := isql.User.GetCurrentLoginUser(c)
		if err != nil {
			return nil, err
		}
		req.Intent = "password"
		return logic.OAuth.StartAuth(c, req, user.ID)
	})
}

// Callback handles the OAuth provider callback
// @Summary OAuth provider callback
// @Description Handle the callback from an OAuth provider after user authorization
// @Tags OAuth Management
// @Produce text/html
// @Param provider path string true "OAuth provider name"
// @Param code query string true "Authorization code"
// @Param state query string true "OAuth state parameter"
// @Success 200 {string} string "HTML page that posts result to opener window"
// @Router /oauth/callback/{provider} [get]
func (m OAuthController) Callback(auth *jwt.GinJWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		code := c.Query("code")
		state := c.Query("state")
		result, err := logic.OAuth.Callback(c, auth, provider, code, state)
		if err != nil {
			handleOAuthError(c, err)
			return
		}
		payload := response.OAuthLoginHTMLPayload{
			Intent:            result.Intent,
			State:             result.State,
			ConnectorID:       result.Connector.ID,
			Provider:          result.Provider,
			Token:             result.Token,
			Code:              result.VerificationCode,
			GeneratedPassword: result.GeneratedPassword,
		}
		if !result.ExpiresAt.IsZero() {
			payload.ExpiresAt = result.ExpiresAt.Format("2006-01-02 15:04:05")
		}
		b, _ := json.Marshal(payload)
		html := fmt.Sprintf(`<!doctype html><html><body><script>const payload=%s;if(window.opener){window.opener.postMessage(payload,"*");}setTimeout(function(){window.close();},500);</script><div style="margin:40px auto;text-align:center;font-family:sans-serif;"><p>授权完成，可以关闭此窗口。</p></div></body></html>`, string(b))
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}

func handleOAuthError(c *gin.Context, err any) {
	msg := fmt.Sprint(err)
	html := fmt.Sprintf(`<!doctype html><html><body><div style="margin:40px auto;text-align:center;font-family:sans-serif;color:#d9534f;">授权失败：%s</div></body></html>`, msg)
	c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(html))
}
