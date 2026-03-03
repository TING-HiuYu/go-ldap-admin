package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// Add creates a new user record.
// @Summary Add user
// @Description Add a new user record
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserAddReq true "User creation request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/add [post]
// @Security ApiKeyAuth
func (m *UserController) Add(c *gin.Context) {
	req := new(request.UserAddReq)
	Run(c, req, func() (any, any) {
		return logic.User.Add(c, req)
	})
}

// Update modifies an existing user record.
// @Summary Update user
// @Description Update an existing user record
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserUpdateReq true "User update request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/update [post]
// @Security ApiKeyAuth
func (m *UserController) Update(c *gin.Context) {
	req := new(request.UserUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.User.Update(c, req)
	})
}

// List retrieves all user records.
// @Summary List users
// @Description Retrieve all user records
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /user/list [get]
// @Security ApiKeyAuth
func (m *UserController) List(c *gin.Context) {
	req := new(request.UserListReq)
	Run(c, req, func() (any, any) {
		return logic.User.List(c, req)
	})
}

// Delete removes a user record.
// @Summary Delete user
// @Description Delete a user record
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserDeleteReq true "User deletion request body with ID"
// @Success 200 {object} response.ResponseBody
// @Router /user/delete [post]
// @Security ApiKeyAuth
func (m UserController) Delete(c *gin.Context) {
	req := new(request.UserDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.User.Delete(c, req)
	})
}

// ChangePwd changes a user's password.
// @Summary Change password
// @Description Change user password
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserChangePwdReq true "Password change request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/changePwd [post]
// @Security ApiKeyAuth
func (m UserController) ChangePwd(c *gin.Context) {
	req := new(request.UserChangePwdReq)
	Run(c, req, func() (any, any) {
		return logic.User.ChangePwd(c, req)
	})
}

// SendPasswordChangeCode sends a verification code for password change.
// @Summary Send password change verification code
// @Description Send a verification code to the currently logged-in user for password change
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /user/password/code [post]
// @Security ApiKeyAuth
func (m UserController) SendPasswordChangeCode(c *gin.Context) {
	req := new(request.UserSendPasswordCodeReq)
	Run(c, req, func() (any, any) {
		return logic.User.SendPasswordChangeCode(c, req)
	})
}

// ResetPassword resets a user's password.
// @Summary Reset user password
// @Description Reset user password to a random password and send email notification
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserResetPasswordReq true "Password reset request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/resetPassword [post]
// @Security ApiKeyAuth
func (m UserController) ResetPassword(c *gin.Context) {
	req := new(request.UserResetPasswordReq)
	Run(c, req, func() (any, any) {
		return logic.User.ResetPassword(c, req)
	})
}

// ChangeUserStatus changes a user's status.
// @Summary Change user status
// @Description Change user status
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.UserChangeUserStatusReq true "User status change request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/changeUserStatus [post]
// @Security ApiKeyAuth
func (m UserController) ChangeUserStatus(c *gin.Context) {
	req := new(request.UserChangeUserStatusReq)
	Run(c, req, func() (any, any) {
		return logic.User.ChangeUserStatus(c, req)
	})
}

// GetUserInfo retrieves the current logged-in user's information.
// @Summary Get current user info
// @Description Retrieve the current logged-in user's information
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /user/info [get]
// @Security ApiKeyAuth
func (uc UserController) GetUserInfo(c *gin.Context) {
	req := new(request.UserGetUserInfoReq)
	Run(c, req, func() (any, any) {
		return logic.User.GetUserInfo(c, req)
	})
}

// IssueSSHPubKey issues an SSH certificate for the current user.
// @Summary Issue SSH certificate for current user
// @Description Generate a new SSH key pair and issue a user certificate signed by the configured CA
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /user/issueSSHPubKey [post]
// @Security ApiKeyAuth
func (m UserController) IssueSSHPubKey(c *gin.Context) {
	req := new(request.UserIssueSSHPubKeyReq)
	Run(c, req, func() (any, any) {
		return logic.User.IssueSSHPubKey(c, req)
	})
}

// RevokeOriginalKeypair revokes old SSH certificates/keys (placeholder, not yet implemented).
// @Summary Revoke old SSH certificates/keys
// @Description Reserved endpoint for broadcasting or webhook-based revocation of old certificates (not yet implemented)
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /user/revokeOriginalKeypair [post]
// @Security ApiKeyAuth
func (m UserController) RevokeOriginalKeypair(c *gin.Context) {
	req := new(request.UserRevokeOriginalKeypairReq)
	Run(c, req, func() (any, any) {
		return logic.User.RevokeOriginalKeypair(c, req)
	})
}

// SyncDingTalkUsers synchronizes DingTalk user information.
// @Summary Sync DingTalk users
// @Description Synchronize DingTalk user information
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.SyncDingUserReq true "DingTalk user sync request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/syncDingTalkUsers [post]
// @Security ApiKeyAuth
func (uc UserController) SyncDingTalkUsers(c *gin.Context) {
	req := new(request.SyncDingUserReq)
	Run(c, req, func() (any, any) {
		return logic.DingTalk.SyncDingTalkUsers(c, req)
	})
}

// SyncWeComUsers synchronizes WeCom (Enterprise WeChat) user information.
// @Summary Sync WeCom users
// @Description Synchronize WeCom user information
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.SyncWeComUserReq true "WeCom user sync request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/syncWeComUsers [post]
// @Security ApiKeyAuth
func (uc UserController) SyncWeComUsers(c *gin.Context) {
	req := new(request.SyncWeComUserReq)
	Run(c, req, func() (any, any) {
		return logic.WeCom.SyncWeComUsers(c, req)
	})
}

// SyncFeiShuUsers synchronizes FeiShu (Lark) user information.
// @Summary Sync FeiShu users
// @Description Synchronize FeiShu user information
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.SyncFeiShuUserReq true "FeiShu user sync request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/syncFeiShuUsers [post]
// @Security ApiKeyAuth
func (uc UserController) SyncFeiShuUsers(c *gin.Context) {
	req := new(request.SyncFeiShuUserReq)
	Run(c, req, func() (any, any) {
		return logic.FeiShu.SyncFeiShuUsers(c, req)
	})
}

// SyncOpenLdapUsers synchronizes OpenLDAP user information.
// @Summary Sync OpenLDAP users
// @Description Synchronize OpenLDAP user information
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.SyncOpenLdapUserReq true "OpenLDAP user sync request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/syncOpenLdapUsers [post]
// @Security ApiKeyAuth
func (uc UserController) SyncOpenLdapUsers(c *gin.Context) {
	req := new(request.SyncOpenLdapUserReq)
	Run(c, req, func() (any, any) {
		return logic.OpenLdap.SyncOpenLdapUsers(c, req)
	})
}

// SyncSqlUsers synchronizes SQL user information to LDAP.
// @Summary Sync SQL users to LDAP
// @Description Synchronize SQL user information to LDAP
// @Tags User Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.SyncSqlUserReq true "SQL user sync request body"
// @Success 200 {object} response.ResponseBody
// @Router /user/syncSqlUsers [post]
// @Security ApiKeyAuth
func (uc UserController) SyncSqlUsers(c *gin.Context) {
	req := new(request.SyncSqlUserReq)
	Run(c, req, func() (any, any) {
		return logic.Sql.SyncSqlUsers(c, req)
	})
}
