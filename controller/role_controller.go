package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type RoleController struct{}

// List retrieves the list of role records.
// @Summary Get role record list
// @Description Get role record list
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /role/list [get]
// @Security ApiKeyAuth
func (m *RoleController) List(c *gin.Context) {
	req := new(request.RoleListReq)
	Run(c, req, func() (any, any) {
		return logic.Role.List(c, req)
	})
}

// Add creates a new role record.
// @Summary Create a new role record
// @Description Create a new role record
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RoleAddReq true "Role creation request body"
// @Success 200 {object} response.ResponseBody
// @Router /role/add [post]
// @Security ApiKeyAuth
func (m *RoleController) Add(c *gin.Context) {
	req := new(request.RoleAddReq)
	Run(c, req, func() (any, any) {
		return logic.Role.Add(c, req)
	})
}

// Update modifies an existing role record.
// @Summary Update a role record
// @Description Update a role record
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RoleUpdateReq true "Role update request body"
// @Success 200 {object} response.ResponseBody
// @Router /role/update [post]
// @Security ApiKeyAuth
func (m *RoleController) Update(c *gin.Context) {
	req := new(request.RoleUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.Role.Update(c, req)
	})
}

// Delete removes a role record.
// @Summary Delete a role record
// @Description Delete a role record
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RoleDeleteReq true "Role deletion request body"
// @Success 200 {object} response.ResponseBody
// @Router /role/delete [post]
// @Security ApiKeyAuth
func (m *RoleController) Delete(c *gin.Context) {
	req := new(request.RoleDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.Role.Delete(c, req)
	})
}

// GetMenuList retrieves the menu list for a role.
// @Summary Get menu list for a role
// @Description Get menu list for a role
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param roleId query int true "Role ID"
// @Success 200 {object} response.ResponseBody
// @Router /role/getmenulist [get]
// @Security ApiKeyAuth
func (m *RoleController) GetMenuList(c *gin.Context) {
	req := new(request.RoleGetMenuListReq)
	Run(c, req, func() (any, any) {
		return logic.Role.GetMenuList(c, req)
	})
}

// GetApiList retrieves the API list for a role.
// @Summary Get API list for a role
// @Description Get API list for a role
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param roleId query int true "Role ID"
// @Success 200 {object} response.ResponseBody
// @Router /role/getapilist [get]
// @Security ApiKeyAuth
func (m *RoleController) GetApiList(c *gin.Context) {
	req := new(request.RoleGetApiListReq)
	Run(c, req, func() (any, any) {
		return logic.Role.GetApiList(c, req)
	})
}

// UpdateMenus updates the menus assigned to a role.
// @Summary Update role menus
// @Description Update role menus
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RoleUpdateMenusReq true "Role menu update request body"
// @Success 200 {object} response.ResponseBody
// @Router /role/updatemenus [post]
// @Security ApiKeyAuth
func (m *RoleController) UpdateMenus(c *gin.Context) {
	req := new(request.RoleUpdateMenusReq)
	Run(c, req, func() (any, any) {
		return logic.Role.UpdateMenus(c, req)
	})
}

// UpdateApis updates the APIs assigned to a role.
// @Summary Update role APIs
// @Description Update role APIs
// @Tags Role Management
// @Accept application/json
// @Produce application/json
// @Param  data body request.RoleUpdateApisReq true "Role API update request body"
// @Success 200 {object} response.ResponseBody
// @Router /role/updateapis [post]
// @Security ApiKeyAuth
func (m *RoleController) UpdateApis(c *gin.Context) {
	req := new(request.RoleUpdateApisReq)
	Run(c, req, func() (any, any) {
		return logic.Role.UpdateApis(c, req)
	})
}
