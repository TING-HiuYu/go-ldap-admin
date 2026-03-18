package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OAuthWebhook 描述挂载在 OAuth 连接器上的 Webhook 端点
// 每个连接器可以配置多个 webhook，每个 webhook 可以监听多种事件并选择推送的字段
type OAuthWebhook struct {
	gorm.Model
	OAuthConnectorID uint           `gorm:"column:oauth_connector_id;not null;index;comment:'所属连接器ID'" json:"oauthConnectorId"`
	URL              string         `gorm:"type:varchar(512);not null;comment:'Webhook目标URL'" json:"url"`
	Events           datatypes.JSON `gorm:"type:json;comment:'订阅的事件列表'" json:"events"`  // ["user.created","user.deleted"]
	Fields           datatypes.JSON `gorm:"type:json;comment:'附加推送字段列表'" json:"fields"` // ["username","email","nickname",...]
	Enabled          bool           `gorm:"default:true;comment:'是否启用'" json:"enabled"`
	Description      string         `gorm:"type:varchar(255);comment:'备注'" json:"description"`
}
