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

// CreateConnector 创建连接器
func (m OAuthController) CreateConnector(c *gin.Context) {
	req := new(request.OAuthConnectorCreateReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Create(c, req)
	})
}

// UpdateConnector 更新连接器
func (m OAuthController) UpdateConnector(c *gin.Context) {
	req := new(request.OAuthConnectorUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Update(c, req)
	})
}

// DeleteConnector 删除连接器
func (m OAuthController) DeleteConnector(c *gin.Context) {
	req := new(request.OAuthConnectorDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.Delete(c, req)
	})
}

// ListConnectors 管理端列表
func (m OAuthController) ListConnectors(c *gin.Context) {
	req := new(request.OAuthConnectorListReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.List(c, req)
	})
}

// PublicConnectors 登录页列表
func (m OAuthController) PublicConnectors(c *gin.Context) {
	req := new(request.OAuthConnectorListReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.PublicList(c, req)
	})
}

// StartLogin 启动OAuth登录流程（无需登录）
func (m OAuthController) StartLogin(c *gin.Context) {
	req := new(request.OAuthStartReq)
	Run(c, req, func() (any, any) {
		return logic.OAuth.StartAuth(c, req, 0)
	})
}

// StartPasswordVerify 已登录用户发起密码校验
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

// Callback OAuth 回调
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
