package model

import (
	"github.com/eryajf/go-ldap-admin/config"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OAuthConnector 描述可配置的 OAuth 提供方连接器
// 包含默认角色、部门和登录 shell 等默认入库行为
// Config 字段以 JSON 存储提供方所需的特定参数（如 clientId/clientSecret 等）
type OAuthConnector struct {
	gorm.Model
	Name              string         `gorm:"type:varchar(100);not null;comment:'连接器名称'" json:"name"`
	Provider          string         `gorm:"type:varchar(50);not null;comment:'OAuth提供方'" json:"provider"`
	Enabled           bool           `gorm:"default:true;comment:'是否启用'" json:"enabled"`
	DefaultRoleIds    datatypes.JSON `gorm:"type:json;comment:'默认角色ID集合'" json:"defaultRoleIds"`
	DefaultStatus     uint           `gorm:"type:tinyint(1);default:1;comment:'默认状态 1正常 2禁用'" json:"defaultStatus"`
	DepartmentId      uint           `gorm:"not null;comment:'默认部门ID'" json:"departmentId"`
	DefaultLoginShell string         `gorm:"type:varchar(64);comment:'默认登录shell'" json:"defaultLoginShell"`
	Config            datatypes.JSON `gorm:"type:json;comment:'提供方配置信息'" json:"config"`
	Description       string         `gorm:"type:varchar(255);comment:'备注'" json:"description"`
	Creator           string         `gorm:"type:varchar(50);comment:'创建人'" json:"creator"`
	Webhooks          []OAuthWebhook `gorm:"foreignKey:OAuthConnectorID;references:ID" json:"webhooks,omitempty"`
}

// GetDefaultShell 返回连接器定义的 shell 或系统默认值
func (o OAuthConnector) GetDefaultShell() string {
	if o.DefaultLoginShell != "" {
		return o.DefaultLoginShell
	}
	if config.Conf != nil && config.Conf.Ldap != nil && config.Conf.Ldap.DefaultLoginShell != "" {
		return config.Conf.Ldap.DefaultLoginShell
	}
	return "/bin/bash"
}
