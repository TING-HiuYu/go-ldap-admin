package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type ApiController struct{}

// List retrieves the paginated list of API endpoint records.
// @Summary Get API endpoint list
// @Description Retrieves the paginated list of API endpoint records
// @Tags API Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /api/list [get]
// @Security ApiKeyAuth
func (m *ApiController) List(c *gin.Context) {
	req := new(request.ApiListReq)
	Run(c, req, func() (any, any) {
		return logic.Api.List(c, req)
	})
}

// GetTree retrieves the API endpoint tree structure.
// @Summary Get API endpoint tree
// @Description Retrieves the tree structure of API endpoints
// @Tags API Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /api/tree [get]
// @Security ApiKeyAuth
func (m *ApiController) GetTree(c *gin.Context) {
	req := new(request.ApiGetTreeReq)
	Run(c, req, func() (any, any) {
		return logic.Api.GetTree(c, req)
	})
}

// Add creates a new API endpoint record.
// @Summary Create a new API endpoint
// @Description Creates a new API endpoint record
// @Tags API Management
// @Accept application/json
// @Produce application/json
// @Param data body request.ApiAddReq true "Create API endpoint"
// @Success 200 {object} response.ResponseBody
// @Router /api/add [post]
// @Security ApiKeyAuth
func (m *ApiController) Add(c *gin.Context) {
	req := new(request.ApiAddReq)
	Run(c, req, func() (any, any) {
		return logic.Api.Add(c, req)
	})
}

// Update modifies an existing API endpoint record.
// @Summary Update an API endpoint
// @Description Updates an existing API endpoint record
// @Tags API Management
// @Accept application/json
// @Produce application/json
// @Param data body request.ApiUpdateReq true "Update API endpoint"
// @Success 200 {object} response.ResponseBody
// @Router /api/update [post]
// @Security ApiKeyAuth
func (m *ApiController) Update(c *gin.Context) {
	req := new(request.ApiUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.Api.Update(c, req)
	})
}

// Delete removes an API endpoint record.
// @Summary Delete an API endpoint
// @Description Deletes an existing API endpoint record
// @Tags API Management
// @Accept application/json
// @Produce application/json
// @Param data body request.ApiDeleteReq true "Delete API endpoint"
// @Success 200 {object} response.ResponseBody
// @Router /api/delete [post]
// @Security ApiKeyAuth
func (m *ApiController) Delete(c *gin.Context) {
	req := new(request.ApiDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.Api.Delete(c, req)
	})
}
