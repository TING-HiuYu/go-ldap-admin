package webhook

import "github.com/eryajf/go-ldap-admin/model"

// ────────────────────── 事件类型 ──────────────────────

// EventType 定义可触发 Webhook 的事件类型
type EventType string

const (
	EventUserCreated EventType = "user.created"
	EventUserDeleted EventType = "user.deleted"
	// 以后可扩展，例如：
	// EventUserUpdated EventType = "user.updated"
	// EventUserDisabled EventType = "user.disabled"
)

// AllEvents 返回当前支持的全部事件类型（前端用于展示可选列表）
func AllEvents() []EventOption {
	return []EventOption{
		{Value: string(EventUserCreated), Label: "用户创建"},
		{Value: string(EventUserDeleted), Label: "用户删除"},
	}
}

// EventOption 用于前端展示的事件选项
type EventOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ────────────────────── 可推送字段 ──────────────────────

// FieldOption 描述可选的附加推送字段
type FieldOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// AllFields 返回当前支持的全部可推送字段（前端用于多选）
func AllFields() []FieldOption {
	return []FieldOption{
		{Key: "username", Label: "用户名"},
		{Key: "email", Label: "邮箱"},
		{Key: "nickname", Label: "昵称"},
		{Key: "givenName", Label: "花名"},
		{Key: "avatar", Label: "头像"},
		{Key: "jobNumber", Label: "工号"},
		{Key: "mobile", Label: "手机号"},
		{Key: "department", Label: "部门"},
		{Key: "position", Label: "职位"},
		{Key: "source", Label: "用户来源"},
	}
}

// ────────────────────── Webhook 负载 ──────────────────────

// Payload 是发送给目标 URL 的 JSON 请求体
type Payload struct {
	Event       string         `json:"event"`
	ConnectorID uint           `json:"connectorId"`
	Provider    string         `json:"provider"`
	Timestamp   int64          `json:"timestamp"`
	User        map[string]any `json:"user"`
}

// BuildUserFields 根据配置的 fields 列表从 User 中提取指定字段
func BuildUserFields(user *model.User, fields []string) map[string]any {
	if user == nil {
		return map[string]any{}
	}
	// 完整映射表
	all := map[string]any{
		"username":   user.Username,
		"email":      user.Mail,
		"nickname":   user.Nickname,
		"givenName":  user.GivenName,
		"avatar":     user.Avatar,
		"jobNumber":  user.JobNumber,
		"mobile":     user.Mobile,
		"department": user.DepartmentId,
		"position":   user.Position,
		"source":     user.Source,
	}

	// 始终包含的基础字段
	result := map[string]any{
		"id":     user.ID,
		"userId": user.ID,
	}

	for _, key := range fields {
		if v, ok := all[key]; ok {
			result[key] = v
		}
	}
	return result
}
