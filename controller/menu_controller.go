package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type MenuController struct{}

// GetTree retrieves the menu tree.
// @Summary Get menu tree
// @Description Retrieve the full menu tree
// @Tags Menu Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /menu/tree [get]
// @Security ApiKeyAuth
func (m *MenuController) GetTree(c *gin.Context) {
	req := new(request.MenuGetTreeReq)
	Run(c, req, func() (any, any) {
		return logic.Menu.GetTree(c, req)
	})
}

// GetAccessTree retrieves the user menu tree by user ID.
// @Summary Get user menu tree
// @Description Retrieve the menu tree accessible by a specific user
// @Tags Menu Management
// @Accept application/json
// @Produce application/json
// @Param id query int true "User ID"
// @Success 200 {object} response.ResponseBody
// @Router /menu/access/tree [get]
// @Security ApiKeyAuth
func (m *MenuController) GetAccessTree(c *gin.Context) {
	req := new(request.MenuGetAccessTreeReq)
	Run(c, req, func() (any, any) {
		return logic.Menu.GetAccessTree(c, req)
	})
}

// Add creates a new menu.
// @Summary Create a menu
// @Description Create a new menu entry
// @Tags Menu Management
// @Accept application/json
// @Produce application/json
// @Param data body request.MenuAddReq true "Menu creation request"
// @Success 200 {object} response.ResponseBody
// @Router /menu/add [post]
// @Security ApiKeyAuth
func (m *MenuController) Add(c *gin.Context) {
	req := new(request.MenuAddReq)
	Run(c, req, func() (any, any) {
		return logic.Menu.Add(c, req)
	})
}

// Update updates an existing menu.
// @Summary Update a menu
// @Description Update an existing menu entry
// @Tags Menu Management
// @Accept application/json
// @Produce application/json
// @Param data body request.MenuUpdateReq true "Menu update request"
// @Success 200 {object} response.ResponseBody
// @Router /menu/update [post]
// @Security ApiKeyAuth
func (m *MenuController) Update(c *gin.Context) {
	req := new(request.MenuUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.Menu.Update(c, req)
	})
}

// Delete deletes an existing menu.
// @Summary Delete a menu
// @Description Delete an existing menu entry
// @Tags Menu Management
// @Accept application/json
// @Produce application/json
// @Param data body request.MenuDeleteReq true "Menu deletion request"
// @Success 200 {object} response.ResponseBody
// @Router /menu/delete [post]
// @Security ApiKeyAuth
func (m *MenuController) Delete(c *gin.Context) {
	req := new(request.MenuDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.Menu.Delete(c, req)
	})
}
