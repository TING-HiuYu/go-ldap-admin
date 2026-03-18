package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct{}

// List retrieves a paginated list of operation log records.
// @Summary Get operation log record list
// @Description Get a paginated list of operation log records
// @Tags Operation Log Management
// @Accept application/json
// @Produce application/json
// @Param username query string false "Username"
// @Param ip query string false "IP address"
// @Param path query string false "Request path"
// @Param method query string false "HTTP method"
// @Param status query int false "Status code"
// @Param pageNum query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} response.ResponseBody
// @Router /log/operation/list [get]
// @Security ApiKeyAuth
func (m *OperationLogController) List(c *gin.Context) {
	req := new(request.OperationLogListReq)
	Run(c, req, func() (any, any) {
		return logic.OperationLog.List(c, req)
	})
}

// Delete removes specified operation log records.
// @Summary Delete operation log records
// @Description Delete operation log records by ID
// @Tags Operation Log Management
// @Accept application/json
// @Produce application/json
// @Param data body request.OperationLogDeleteReq true "IDs of log records to delete"
// @Success 200 {object} response.ResponseBody
// @Router /log/operation/delete [post]
// @Security ApiKeyAuth
func (m *OperationLogController) Delete(c *gin.Context) {
	req := new(request.OperationLogDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.OperationLog.Delete(c, req)
	})
}

// Clean removes all operation log records.
// @Summary Clean all operation log records
// @Description Remove all operation log records
// @Tags Operation Log Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /log/operation/clean [delete]
// @Security ApiKeyAuth
func (m *OperationLogController) Clean(c *gin.Context) {
	req := new(request.OperationLogListReq)
	Run(c, req, func() (any, any) {
		return logic.OperationLog.Clean(c, req)
	})
}
