package response

import "time"

// OAuthConnectorItem 返回给前端的连接器信息
// Config 字段仅在后台管理接口返回，公开接口会移除敏感字段
// DefaultRoleIds 提供已解析的角色ID数组，方便前端直接使用
// CreatedAt/UpdatedAt 采用 ISO8601 字符串，便于前端展示

type OAuthConnectorItem struct {
	ID                uint               `json:"id"`
	Name              string             `json:"name"`
	Provider          string             `json:"provider"`
	Enabled           bool               `json:"enabled"`
	DefaultRoleIds    []uint             `json:"defaultRoleIds"`
	DefaultStatus     uint               `json:"defaultStatus"`
	DepartmentId      uint               `json:"departmentId"`
	DefaultLoginShell string             `json:"defaultLoginShell"`
	Config            map[string]any     `json:"config,omitempty"`
	Description       string             `json:"description"`
	Creator           string             `json:"creator"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`
	Webhooks          []OAuthWebhookItem `json:"webhooks,omitempty"`
}

// OAuthWebhookItem 返回给前端的 Webhook 信息
type OAuthWebhookItem struct {
	ID          uint     `json:"id"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Fields      []string `json:"fields"`
	Enabled     bool     `json:"enabled"`
	Description string   `json:"description"`
}

// OAuthStartRsp 返回授权跳转信息
// AuthUrl 由后端拼装并包含state
// State 由前端保存用于回调校验

type OAuthStartRsp struct {
	AuthUrl     string `json:"authUrl"`
	State       string `json:"state"`
	ConnectorID uint   `json:"connectorId"`
	Provider    string `json:"provider"`
}

// OAuthLoginHTMLPayload 用于回调时注入页面脚本的负载
// Token 字段仅在 intent=login 时返回
// Code 字段在 intent=password 时返回验证码

type OAuthLoginHTMLPayload struct {
	Intent            string `json:"intent"`
	Token             string `json:"token,omitempty"`
	ExpiresAt         string `json:"expiresAt,omitempty"`
	State             string `json:"state"`
	ConnectorID       uint   `json:"connectorId"`
	Provider          string `json:"provider"`
	Code              string `json:"code,omitempty"`
	GeneratedPassword string `json:"generatedPassword,omitempty"`
}

// OAuthProviderOption providers for UI selector.
type OAuthProviderOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// OAuthProviderSchema describes provider config fields for UI.
type OAuthProviderSchema struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Title  string               `json:"title"`
	Fields []OAuthProviderField `json:"fields"`
}

// OAuthProviderField defines a single config field.
type OAuthProviderField struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Placeholder  string `json:"placeholder"`
	InputType    string `json:"inputType"`
	DefaultValue string `json:"defaultValue"`
	Required     bool   `json:"required"`
}
