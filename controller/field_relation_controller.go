package controller

import (
	"github.com/eryajf/go-ldap-admin/logic"
	"github.com/eryajf/go-ldap-admin/model/request"

	"github.com/gin-gonic/gin"
)

type FieldRelationController struct{}

// List retrieves a list of field relation records.
// @Summary List field relations
// @Description Retrieve a list of field relation management records
// @Tags Field Relation Management
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.ResponseBody
// @Router /fieldrelation/list [get]
// @Security ApiKeyAuth
func (m *FieldRelationController) List(c *gin.Context) {
	req := new(request.FieldRelationListReq)
	Run(c, req, func() (any, any) {
		return logic.FieldRelation.List(c, req)
	})
}

// Add creates a new field relation record.
// @Summary Add a field relation
// @Description Create a new field relation management record
// @Tags Field Relation Management
// @Accept application/json
// @Produce application/json
// @Param data body request.FieldRelationAddReq true "Field relation creation request"
// @Success 200 {object} response.ResponseBody
// @Router /fieldrelation/add [post]
// @Security ApiKeyAuth
func (m *FieldRelationController) Add(c *gin.Context) {
	req := new(request.FieldRelationAddReq)
	Run(c, req, func() (any, any) {
		return logic.FieldRelation.Add(c, req)
	})
}

// Update modifies an existing field relation record.
// @Summary Update a field relation
// @Description Update an existing field relation management record
// @Tags Field Relation Management
// @Accept application/json
// @Produce application/json
// @Param data body request.FieldRelationUpdateReq true "Field relation update request"
// @Success 200 {object} response.ResponseBody
// @Router /fieldrelation/update [post]
// @Security ApiKeyAuth
func (m *FieldRelationController) Update(c *gin.Context) {
	req := new(request.FieldRelationUpdateReq)
	Run(c, req, func() (any, any) {
		return logic.FieldRelation.Update(c, req)
	})
}

// Delete removes a field relation record.
// @Summary Delete a field relation
// @Description Delete an existing field relation management record
// @Tags Field Relation Management
// @Accept application/json
// @Produce application/json
// @Param data body request.FieldRelationDeleteReq true "Field relation deletion request"
// @Success 200 {object} response.ResponseBody
// @Router /fieldrelation/delete [post]
// @Security ApiKeyAuth
func (m *FieldRelationController) Delete(c *gin.Context) {
	req := new(request.FieldRelationDeleteReq)
	Run(c, req, func() (any, any) {
		return logic.FieldRelation.Delete(c, req)
	})
}
