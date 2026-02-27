package isql

import (
	"fmt"
	"strings"

	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/public/tools"
)

type OAuthConnectorService struct{}

func (s OAuthConnectorService) Create(conn *model.OAuthConnector) error {
	return common.DB.Create(conn).Error
}

func (s OAuthConnectorService) Update(conn *model.OAuthConnector) error {
	return common.DB.Model(&model.OAuthConnector{}).Where("id = ?", conn.ID).Updates(conn).Error
}

func (s OAuthConnectorService) Delete(ids []uint) error {
	return common.DB.Where("id IN (?)", ids).Delete(&model.OAuthConnector{}).Error
}

func (s OAuthConnectorService) Find(filter map[string]any, conn *model.OAuthConnector) error {
	return common.DB.Where(filter).First(conn).Error
}

func (s OAuthConnectorService) List(req *request.OAuthConnectorListReq) (list []*model.OAuthConnector, total int64, err error) {
	db := common.DB.Model(&model.OAuthConnector{})
	if req.Provider != "" {
		db = db.Where("provider = ?", req.Provider)
	}
	if req.Enabled != nil {
		db = db.Where("enabled = ?", *req.Enabled)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		db = db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", keyword))
	}

	if err = db.Count(&total).Error; err != nil {
		return
	}
	page := tools.NewPageOption(req.PageNum, req.PageSize)
	err = db.Order("id DESC").Offset(page.PageNum).Limit(page.PageSize).Preload("Webhooks").Find(&list).Error
	return
}

func (s OAuthConnectorService) ListEnabled() (list []model.OAuthConnector, err error) {
	err = common.DB.Where("enabled = ?", true).Order("id DESC").Find(&list).Error
	return
}
