package request

// OAuthConnectorCreateReq 创建 OAuth 连接器
type OAuthConnectorCreateReq struct {
	Name              string            `json:"name" validate:"required,min=1,max=100"`
	Provider          string            `json:"provider" validate:"required"`
	Enabled           bool              `json:"enabled"`
	DefaultRoleIds    []uint            `json:"defaultRoleIds" validate:"required"`
	DefaultStatus     uint              `json:"defaultStatus" validate:"oneof=1 2"`
	DepartmentId      uint              `json:"departmentId" validate:"required"`
	DefaultLoginShell string            `json:"defaultLoginShell" validate:"max=64"`
	Config            map[string]any    `json:"config" validate:"required"`
	Description       string            `json:"description" validate:"max=255"`
	Webhooks          []OAuthWebhookReq `json:"webhooks"`
}

// OAuthConnectorUpdateReq 更新 OAuth 连接器
type OAuthConnectorUpdateReq struct {
	ID                uint              `json:"id" validate:"required"`
	Name              string            `json:"name" validate:"required,min=1,max=100"`
	Provider          string            `json:"provider" validate:"required"`
	Enabled           bool              `json:"enabled"`
	DefaultRoleIds    []uint            `json:"defaultRoleIds" validate:"required"`
	DefaultStatus     uint              `json:"defaultStatus" validate:"oneof=1 2"`
	DepartmentId      uint              `json:"departmentId" validate:"required"`
	DefaultLoginShell string            `json:"defaultLoginShell" validate:"max=64"`
	Config            map[string]any    `json:"config" validate:"required"`
	Description       string            `json:"description" validate:"max=255"`
	Webhooks          []OAuthWebhookReq `json:"webhooks"`
}

// OAuthWebhookReq 单个 Webhook 配置（创建或更新时提交）
type OAuthWebhookReq struct {
	ID          uint     `json:"id"`
	URL         string   `json:"url" validate:"required,url"`
	Events      []string `json:"events" validate:"required,min=1"`
	Fields      []string `json:"fields"`
	Enabled     bool     `json:"enabled"`
	Description string   `json:"description" validate:"max=255"`
}

// OAuthConnectorDeleteReq 删除连接器
type OAuthConnectorDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

// OAuthConnectorListReq 查询连接器
type OAuthConnectorListReq struct {
	Provider string `json:"provider" form:"provider"`
	Enabled  *bool  `json:"enabled" form:"enabled"`
	Keyword  string `json:"keyword" form:"keyword"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// OAuthStartReq 启动 OAuth 授权
type OAuthStartReq struct {
	ConnectorID uint   `json:"connectorId" validate:"required"`
	Intent      string `json:"intent" validate:"omitempty,oneof=login password"`
}

// BaseSendLoginCodeReq 发送登录验证码
type BaseSendLoginCodeReq struct {
	Mail string `json:"mail" validate:"required,email"`
}

// BaseOtpLoginReq 邮箱验证码登录
type BaseOtpLoginReq struct {
	Mail  string `json:"mail" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
	Reset bool   `json:"reset"` // true 表示重置密码并登录
}
