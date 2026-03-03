package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type GroupController struct{}

// List retrieves the list of group records
// @Summary Get group record list
// @Description Get group record list
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/list [get]
// @Security ApiKeyAuth
func (m *GroupController) List(c *gin.Context) {
	req := new(request.GroupListReq)
	Run(c, req, func() (any, any) {
		return logic.Group.List(c, req)
	})
}

// UserInGroup retrieves users within a group
// @Summary Get users in a group
// @Description Get users in a group
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param groupId query int true "Group ID"
// @Param nickname query string false "Nickname"
// @Success 200 {object} response.ResponseBody
// @Router /group/useringroup [get]
// @Security ApiKeyAuth
func (m *GroupController) UserInGroup(c *gin.Context) {
	req := new(request.UserInGroupReq)
	Run(c, req, func() (any, any) {
		return logic.Group.UserInGroup(c, req)
	})
}

// UserNoInGroup retrieves users not in a group
// @Summary Get users not in a group
// @Description Get users not in a group
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param groupId query int true "Group ID"
// @Param nickname query string false "Nickname"
// @Success 200 {object} response.ResponseBody
// @Router /group/usernoingroup [get]
// @Security ApiKeyAuth
func (m *GroupController) UserNoInGroup(c *gin.Context) {
	req := new(request.UserNoInGroupReq)
	Run(c, req, func() (any, any) {
		return logic.Group.UserNoInGroup(c, req)
	})
}

// GetTree retrieves the group tree
// @Summary Get group tree
// @Description Get group tree
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/tree [get]
// @Security ApiKeyAuth
func (m *GroupController) GetTree(c *gin.Context) {
	req := new(request.GroupListReq)
	Run(c, req, func() (any, any) {
		return logic.Group.GetTree(c, req)
	})
}

// Add creates a new group record
// @Summary Add a group record
// @Description Add a group record
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.GroupAddReq true "Add group record request body"
// @Success 200 {object} response.ResponseBody
// @Router /group/add [post]
// @Security ApiKeyAuth
func (m *GroupController) Add(c *gin.Context) {
	req := new(request.GroupAddReq)
	Run(c, req, func() (any, any) {
		return logic.Group.Add(c, req)
	})
}

// Update updates a group record
// @Summary Update a group record
// @Description Update a group record
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.GroupUpdateReq true "Update group record request body"
// @Success 200 {object} response.ResponseBody
// @Router /group/update [post]
// @Security ApiKeyAuth
func (m *GroupController) Update(c *gin.Context) {
	req := new(request.GroupUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.Group.Update(c, req)
	})
}

// Delete deletes a group record
// @Summary Delete a group record
// @Description Delete a group record
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.GroupDeleteReq true "Delete group record request body"
// @Success 200 {object} response.ResponseBody
// @Router /group/delete [post]
// @Security ApiKeyAuth
func (m *GroupController) Delete(c *gin.Context) {
	req := new(request.GroupDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.Group.Delete(c, req)
	})
}

// AddUser adds a user to a group
// @Summary Add a user to a group
// @Description Add a user to a group
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.GroupAddUserReq true "Add user to group request body"
// @Success 200 {object} response.ResponseBody
// @Router /group/adduser [post]
// @Security ApiKeyAuth
func (m *GroupController) AddUser(c *gin.Context) {
	req := new(request.GroupAddUserReq)
	Run(c, req, func() (any, any) {
		return logic.Group.AddUser(c, req)
	})
}

// RemoveUser removes a user from a group
// @Summary Remove a user from a group
// @Description Remove a user from a group
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.GroupRemoveUserReq true "Remove user from group request body"
// @Success 200 {object} response.ResponseBody
// @Router /group/removeuser [post]
// @Security ApiKeyAuth
func (m *GroupController) RemoveUser(c *gin.Context) {
	req := new(request.GroupRemoveUserReq)
	Run(c, req, func() (any, any) {
		return logic.Group.RemoveUser(c, req)
	})
}

// SyncDingTalkDepts synchronizes DingTalk department information
// @Summary Sync DingTalk department information
// @Description Sync DingTalk department information
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/syncDingTalkDepts [post]
// @Security ApiKeyAuth
func (m *GroupController) SyncDingTalkDepts(c *gin.Context) {
	req := new(request.SyncDingTalkDeptsReq)
	Run(c, req, func() (any, any) {
		return logic.DingTalk.SyncDingTalkDepts(c, req)
	})
}

// SyncWeComDepts synchronizes WeCom department information
// @Summary Sync WeCom department information
// @Description Sync WeCom department information
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/syncWeComDepts [post]
// @Security ApiKeyAuth
func (m *GroupController) SyncWeComDepts(c *gin.Context) {
	req := new(request.SyncWeComDeptsReq)
	Run(c, req, func() (any, any) {
		return logic.WeCom.SyncWeComDepts(c, req)
	})
}

// SyncFeiShuDepts synchronizes FeiShu (Lark) department information
// @Summary Sync FeiShu department information
// @Description Sync FeiShu department information
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/syncFeiShuDepts [post]
// @Security ApiKeyAuth
func (m *GroupController) SyncFeiShuDepts(c *gin.Context) {
	req := new(request.SyncFeiShuDeptsReq)
	Run(c, req, func() (any, any) {
		return logic.FeiShu.SyncFeiShuDepts(c, req)
	})
}

// SyncOpenLdapDepts synchronizes existing OpenLDAP department information
// @Summary Sync OpenLDAP department information
// @Description Sync OpenLDAP department information
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/syncOpenLdapDepts [post]
// @Security ApiKeyAuth
func (m *GroupController) SyncOpenLdapDepts(c *gin.Context) {
	req := new(request.SyncOpenLdapDeptsReq)
	Run(c, req, func() (any, any) {
		return logic.OpenLdap.SyncOpenLdapDepts(c, req)
	})
}

// SyncSqlGroups synchronizes group information from SQL to LDAP
// @Summary Sync group information from SQL to LDAP
// @Description Sync group information from SQL to LDAP
// @Tags Group Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /group/syncSqlGroups [post]
// @Security ApiKeyAuth
func (m *GroupController) SyncSqlGroups(c *gin.Context) {
	req := new(request.SyncSqlGrooupsReq)
	Run(c, req, func() (any, any) {
		return logic.Sql.SyncSqlGroups(c, req)
	})
}
