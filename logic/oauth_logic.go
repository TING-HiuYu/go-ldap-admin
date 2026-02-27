package logic

import (
	crand "crypto/rand"
	"encoding/base64"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/logic/oauthprovider"
	"github.com/eryajf/go-ldap-admin/logic/webhook"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/request"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/public/tools"
	"github.com/eryajf/go-ldap-admin/service/isql"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OAuthLogic struct{}

var (
	OAuth           = new(OAuthLogic)
	oauthStateCache = cache.New(30*time.Minute, 60*time.Minute)
)

type oauthState struct {
	ConnectorID uint   `json:"connectorId"`
	Intent      string `json:"intent"`
	UserID      uint   `json:"userId"`
}

type oauthCallbackResult struct {
	Intent            string
	User              *model.User
	Connector         *model.OAuthConnector
	Provider          string
	VerificationCode  string
	State             string
	Token             string
	ExpiresAt         time.Time
	GeneratedPassword string // 仅新创建用户时有值
}

// Create 连接器
func (l OAuthLogic) Create(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.OAuthConnectorCreateReq)
	if !ok {
		return nil, ReqAssertErr
	}
	if oauthprovider.Get(r.Provider) == nil {
		return nil, tools.NewValidatorError(fmt.Errorf("不支持的OAuth提供方"))
	}
	if len(r.DefaultRoleIds) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("请至少选择一个默认角色"))
	}
	if r.DepartmentId == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("请选择所属部门"))
	}
	// 校验角色与部门存在
	if _, err := isql.Role.GetRolesByIds(r.DefaultRoleIds); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("角色无效: %v", err))
	}
	if !isql.Group.Exist(tools.H{"id": r.DepartmentId}) {
		return nil, tools.NewValidatorError(fmt.Errorf("部门不存在"))
	}

	configMap, err := normalizeConnectorConfig(r.Config)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析配置信息失败: %v", err))
	}
	cfgBytes, err := stdjson.Marshal(configMap)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("保存配置信息失败: %v", err))
	}
	roleBytes, _ := stdjson.Marshal(r.DefaultRoleIds)

	creator := "system"
	if u, err := isql.User.GetCurrentLoginUser(c); err == nil {
		creator = u.Username
	}

	conn := &model.OAuthConnector{
		Name:              r.Name,
		Provider:          strings.ToLower(r.Provider),
		Enabled:           r.Enabled,
		DefaultRoleIds:    datatypes.JSON(roleBytes),
		DefaultStatus:     r.DefaultStatus,
		DepartmentId:      r.DepartmentId,
		DefaultLoginShell: r.DefaultLoginShell,
		Config:            datatypes.JSON(cfgBytes),
		Description:       r.Description,
		Creator:           creator,
	}

	if err := validateConnectorConfig(conn); err != nil {
		return nil, err
	}

	if err := isql.OAuthConnector.Create(conn); err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("创建连接器失败: %v", err))
	}

	// 保存 webhooks
	if len(r.Webhooks) > 0 {
		if err := saveWebhooks(conn.ID, r.Webhooks); err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("保存Webhook失败: %v", err))
		}
	}

	return conn.ID, nil
}

// Update 更新连接器
func (l OAuthLogic) Update(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.OAuthConnectorUpdateReq)
	if !ok {
		return nil, ReqAssertErr
	}
	if r.ID == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("缺少连接器ID"))
	}
	if len(r.DefaultRoleIds) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("请至少选择一个默认角色"))
	}
	if r.DepartmentId == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("请选择所属部门"))
	}
	if oauthprovider.Get(r.Provider) == nil {
		return nil, tools.NewValidatorError(fmt.Errorf("不支持的OAuth提供方"))
	}
	if _, err := isql.Role.GetRolesByIds(r.DefaultRoleIds); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("角色无效: %v", err))
	}
	if !isql.Group.Exist(tools.H{"id": r.DepartmentId}) {
		return nil, tools.NewValidatorError(fmt.Errorf("部门不存在"))
	}

	configMap, err := normalizeConnectorConfig(r.Config)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析配置信息失败: %v", err))
	}
	cfgBytes, err := stdjson.Marshal(configMap)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("保存配置信息失败: %v", err))
	}
	roleBytes, _ := stdjson.Marshal(r.DefaultRoleIds)

	conn := &model.OAuthConnector{
		Model:             gorm.Model{ID: r.ID},
		Name:              r.Name,
		Provider:          strings.ToLower(r.Provider),
		Enabled:           r.Enabled,
		DefaultRoleIds:    datatypes.JSON(roleBytes),
		DefaultStatus:     r.DefaultStatus,
		DepartmentId:      r.DepartmentId,
		DefaultLoginShell: r.DefaultLoginShell,
		Config:            datatypes.JSON(cfgBytes),
		Description:       r.Description,
	}

	if err := validateConnectorConfig(conn); err != nil {
		return nil, err
	}

	if err := isql.OAuthConnector.Update(conn); err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("更新连接器失败: %v", err))
	}

	// 更新 webhooks（先删后建）
	if err := saveWebhooks(conn.ID, r.Webhooks); err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("保存Webhook失败: %v", err))
	}

	return conn.ID, nil
}

// Delete 删除连接器
func (l OAuthLogic) Delete(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.OAuthConnectorDeleteReq)
	if !ok {
		return nil, ReqAssertErr
	}
	if len(r.Ids) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("请选择要删除的连接器"))
	}
	// 先删除关联的 webhooks
	if err := common.DB.Where("oauth_connector_id IN (?)", r.Ids).Delete(&model.OAuthWebhook{}).Error; err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("删除关联Webhook失败: %v", err))
	}
	if err := isql.OAuthConnector.Delete(r.Ids); err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("删除连接器失败: %v", err))
	}
	return nil, nil
}

// List 管理端列表
func (l OAuthLogic) List(c *gin.Context, req any) (data any, rspError any) {
	r, ok := req.(*request.OAuthConnectorListReq)
	if !ok {
		return nil, ReqAssertErr
	}
	list, total, err := isql.OAuthConnector.List(r)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取连接器列表失败: %v", err))
	}
	items, err := l.buildConnectorItems(list, true)
	if err != nil {
		return nil, err
	}
	return gin.H{
		"total":         total,
		"list":          items,
		"providers":     oauthprovider.ListOptions(),
		"schemas":       oauthprovider.ListSchemas(),
		"webhookEvents": webhook.AllEvents(),
		"webhookFields": webhook.AllFields(),
	}, nil
}

// PublicList 登录页使用的连接器列表（去除敏感字段）
func (l OAuthLogic) PublicList(c *gin.Context, req any) (data any, rspError any) {
	_, ok := req.(*request.OAuthConnectorListReq)
	if !ok {
		return nil, ReqAssertErr
	}
	list, err := isql.OAuthConnector.ListEnabled()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取连接器失败: %v", err))
	}
	items, err := l.buildConnectorItemsPtr(list, false)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// StartAuth 登录或验证入口，userId 在密码验证场景传入
func (l OAuthLogic) StartAuth(c *gin.Context, req any, userId uint) (data any, rspError any) {
	r, ok := req.(*request.OAuthStartReq)
	if !ok {
		return nil, ReqAssertErr
	}
	intent := r.Intent
	if intent == "" {
		intent = "login"
	}
	if intent != "login" && intent != "password" {
		return nil, tools.NewValidatorError(fmt.Errorf("不支持的intent"))
	}
	connector := new(model.OAuthConnector)
	if err := isql.OAuthConnector.Find(tools.H{"id": r.ConnectorID}, connector); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("连接器不存在或已删除"))
	}
	if !connector.Enabled {
		return nil, tools.NewValidatorError(fmt.Errorf("连接器未启用"))
	}
	state := randomState()
	oauthStateCache.Set(state, oauthState{ConnectorID: connector.ID, Intent: intent, UserID: userId}, 20*time.Minute)
	provider := oauthprovider.Get(connector.Provider)
	if provider == nil {
		return nil, tools.NewValidatorError(fmt.Errorf("不支持的OAuth提供方"))
	}
	authUrl, err := provider.BuildAuthorizeURL(connector, state)
	if err != nil {
		return nil, err
	}
	return response.OAuthStartRsp{AuthUrl: authUrl, State: state, ConnectorID: connector.ID, Provider: connector.Provider}, nil
}

// Callback 统一回调入口
func (l OAuthLogic) Callback(c *gin.Context, auth *jwt.GinJWTMiddleware, provider, code, state string) (*oauthCallbackResult, any) {
	rawState, ok := oauthStateCache.Get(state)
	if !ok {
		return nil, tools.NewValidatorError(fmt.Errorf("state已失效或不存在"))
	}
	oauthStateCache.Delete(state)
	st := rawState.(oauthState)

	connector := new(model.OAuthConnector)
	if err := isql.OAuthConnector.Find(tools.H{"id": st.ConnectorID}, connector); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("连接器不存在"))
	}
	if strings.ToLower(provider) != connector.Provider {
		return nil, tools.NewValidatorError(fmt.Errorf("provider不匹配"))
	}
	providerImpl := oauthprovider.Get(connector.Provider)
	if providerImpl == nil {
		return nil, tools.NewValidatorError(fmt.Errorf("不支持的OAuth提供方"))
	}

	profile, err := providerImpl.FetchProfile(connector, code)
	if err != nil {
		return nil, err
	}
	// Dump incoming OAuth payload for debugging
	logOAuthProfile(profile, connector, nil)

	result := &oauthCallbackResult{Intent: st.Intent, Connector: connector, Provider: connector.Provider, State: state}
	if st.Intent == "password" {
		if st.UserID == 0 {
			return nil, tools.NewValidatorError(fmt.Errorf("缺少用户信息"))
		}
		user, err := isql.User.GetUserById(st.UserID)
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("获取用户失败: %v", err))
		}
		if user.Mail != "" && !strings.EqualFold(user.Mail, profile.Email) {
			return nil, tools.NewValidatorError(fmt.Errorf("邮箱不匹配，无法完成校验"))
		}
		codeVal, _, err := tools.GenerateVerificationCode(tools.PurposePasswordChange, user.Mail)
		if err != nil {
			return nil, tools.NewValidatorError(err)
		}
		result.User = &user
		result.VerificationCode = codeVal
		return result, nil
	}

	user, generatedPass, err := l.ensureUserFromOAuth(profile, connector)
	if err != nil {
		return nil, err
	}
	if user == nil || user.ID == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("创建用户失败，无法生成token"))
	}
	// Re-fetch user by email to ensure full profile (roles) before issuing JWT.
	userFull := new(model.User)
	if err := isql.User.Find(tools.H{"mail": profile.Email}, userFull); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取用户信息失败: %v", err))
	}
	if userFull.ID == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("用户信息为空，无法生成token"))
	}
	payload := tools.H{"user": tools.Struct2Json(userFull)}
	// Validate that PayloadFunc produces correct claims
	claims := auth.PayloadFunc(payload)
	userClaim, ok := claims["user"].(string)
	if !ok || strings.TrimSpace(userClaim) == "" {
		return nil, tools.NewValidatorError(fmt.Errorf("JWT payload 缺少 user 信息"))
	}
	if claims[jwt.IdentityKey] == nil {
		return nil, tools.NewValidatorError(fmt.Errorf("JWT payload 缺少身份信息"))
	}
	// IMPORTANT: pass original payload (tools.H), NOT claims (jwt.MapClaims).
	// TokenGenerator internally calls PayloadFunc(data) again; if data is
	// jwt.MapClaims the type assertion data.(tools.H) in payloadFunc fails
	// (different named types) and returns empty claims → token missing "user".
	token, expires, err := auth.TokenGenerator(payload)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("生成token失败: %v", err))
	}
	result.User = user
	result.Token = token
	result.ExpiresAt = expires
	result.GeneratedPassword = generatedPass
	return result, nil
}

func (l OAuthLogic) buildConnectorItems(list []*model.OAuthConnector, includeConfig bool) ([]response.OAuthConnectorItem, error) {
	items := make([]response.OAuthConnectorItem, 0, len(list))
	for _, conn := range list {
		item, err := l.connectorToItem(conn, includeConfig)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (l OAuthLogic) buildConnectorItemsPtr(list []model.OAuthConnector, includeConfig bool) ([]response.OAuthConnectorItem, error) {
	items := make([]response.OAuthConnectorItem, 0, len(list))
	for i := range list {
		item, err := l.connectorToItem(&list[i], includeConfig)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (l OAuthLogic) connectorToItem(conn *model.OAuthConnector, includeConfig bool) (response.OAuthConnectorItem, error) {
	var roleIds []uint
	if len(conn.DefaultRoleIds) > 0 {
		if err := stdjson.Unmarshal(conn.DefaultRoleIds, &roleIds); err != nil {
			return response.OAuthConnectorItem{}, tools.NewValidatorError(fmt.Errorf("解析默认角色失败: %v", err))
		}
	}
	item := response.OAuthConnectorItem{
		ID:                conn.ID,
		Name:              conn.Name,
		Provider:          conn.Provider,
		Enabled:           conn.Enabled,
		DefaultRoleIds:    roleIds,
		DefaultStatus:     conn.DefaultStatus,
		DepartmentId:      conn.DepartmentId,
		DefaultLoginShell: conn.GetDefaultShell(),
		Description:       conn.Description,
		Creator:           conn.Creator,
		CreatedAt:         conn.CreatedAt,
		UpdatedAt:         conn.UpdatedAt,
	}
	if includeConfig {
		var cfg map[string]any
		_ = stdjson.Unmarshal(conn.Config, &cfg)
		item.Config = cfg

		// 管理接口才返回 webhooks
		var webhookItems []response.OAuthWebhookItem
		for _, wh := range conn.Webhooks {
			var events, fields []string
			_ = stdjson.Unmarshal(wh.Events, &events)
			_ = stdjson.Unmarshal(wh.Fields, &fields)
			webhookItems = append(webhookItems, response.OAuthWebhookItem{
				ID:          wh.ID,
				URL:         wh.URL,
				Events:      events,
				Fields:      fields,
				Enabled:     wh.Enabled,
				Description: wh.Description,
			})
		}
		item.Webhooks = webhookItems
	}
	return item, nil
}

func validateConnectorConfig(conn *model.OAuthConnector) error {
	provider := oauthprovider.Get(conn.Provider)
	if provider == nil {
		return tools.NewValidatorError(fmt.Errorf("不支持的OAuth提供方"))
	}
	return provider.ValidateConfig(conn)
}

// normalizeConnectorConfig 支持前端传入 {encoded: base64(json)} 或直接 map
func normalizeConnectorConfig(raw map[string]any) (map[string]any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}
	if enc, ok := raw["encoded"]; ok {
		encStr, ok := enc.(string)
		if !ok {
			return nil, fmt.Errorf("encoded 字段格式错误")
		}
		bytes, err := base64.StdEncoding.DecodeString(encStr)
		if err != nil {
			return nil, fmt.Errorf("base64 解码失败: %v", err)
		}
		cfg := map[string]any{}
		if err := stdjson.Unmarshal(bytes, &cfg); err != nil {
			return nil, fmt.Errorf("JSON 解析失败: %v", err)
		}
		return cfg, nil
	}
	return raw, nil
}

func (l OAuthLogic) ensureUserFromOAuth(profile *oauthprovider.OAuthUserProfile, connector *model.OAuthConnector) (*model.User, string, error) {
	user := new(model.User)
	err := isql.User.Find(tools.H{"mail": profile.Email}, user)
	if err == nil {
		if user.Status != 1 {
			return nil, "", tools.NewValidatorError(fmt.Errorf("账号已被禁用"))
		}
		updates := map[string]any{
			"oauth_connector_id": connector.ID,
			"oauth_provider":     connector.Provider,
			"source":             fmt.Sprintf("oauth-%s", connector.Provider),
			"source_user_id":     profile.ID,
			"source_union_id":    profile.Email,
		}
		if profile.Avatar != "" && user.Avatar == "" {
			updates["avatar"] = profile.Avatar
		}
		if err := common.DB.Model(user).Updates(updates).Error; err != nil {
			return nil, "", tools.NewMySqlError(fmt.Errorf("更新用户来源失败: %v", err))
		}
		return user, "", nil // 已存在用户不返回密码
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", tools.NewMySqlError(fmt.Errorf("查询用户失败: %v", err))
	}

	roles, err := parseDefaultRoles(connector)
	if err != nil {
		return nil, "", err
	}
	roleObjs, err := isql.Role.GetRolesByIds(roles)
	if err != nil {
		return nil, "", tools.NewValidatorError(fmt.Errorf("获取默认角色失败: %v", err))
	}
	groups, err := isql.Group.GetGroupByIds([]uint{connector.DepartmentId})
	if err != nil {
		return nil, "", tools.NewValidatorError(fmt.Errorf("获取默认部门失败: %v", err))
	}

	// 拼接部门名称，防止 CommonAddUser 里 fallback 成 "默认:研发中心"
	var deptNames []string
	for _, g := range groups {
		if g.GroupName != "" {
			deptNames = append(deptNames, g.GroupName)
		}
	}
	departments := strings.Join(deptNames, ",")

	username := profile.Username
	suffix := 1
	for isql.User.Exist(tools.H{"username": username}) {
		username = fmt.Sprintf("%s-%d", profile.Username, suffix)
		suffix++
	}

	randomPass := tools.GenerateRandomPassword()
	newUser := &model.User{
		Username:         username,
		Password:         randomPass,
		Nickname:         profile.Name,
		GivenName:        profile.Name,
		Mail:             profile.Email,
		Avatar:           profile.Avatar,
		JobNumber:        "0000",
		Mobile:           "0000",
		Status:           connector.DefaultStatus,
		Creator:          connector.Creator,
		Departments:      departments,
		DepartmentId:     tools.SliceToString([]uint{connector.DepartmentId}, ","),
		Roles:            roleObjs,
		Source:           fmt.Sprintf("oauth-%s", connector.Provider),
		SourceUserId:     profile.ID,
		SourceUnionId:    profile.Email,
		OAuthConnectorID: connector.ID,
		OAuthProvider:    connector.Provider,
		UserDN:           fmt.Sprintf("uid=%s,%s", username, config.Conf.Ldap.UserDN),
		LoginShell:       connector.GetDefaultShell(),
	}

	if err := CommonAddUser(newUser, groups); err != nil {
		return nil, "", err
	}

	// 触发用户创建事件（异步推送 webhook）
	webhook.Fire(webhook.EventUserCreated, connector.ID, connector.Provider, newUser)

	return newUser, randomPass, nil // 新用户返回随机密码
}

func parseDefaultRoles(connector *model.OAuthConnector) ([]uint, error) {
	var roleIds []uint
	if len(connector.DefaultRoleIds) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("连接器缺少默认角色"))
	}
	if err := stdjson.Unmarshal(connector.DefaultRoleIds, &roleIds); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析默认角色失败: %v", err))
	}
	if len(roleIds) == 0 {
		return nil, tools.NewValidatorError(fmt.Errorf("连接器未配置默认角色"))
	}
	return roleIds, nil
}

func randomState() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	if _, err := crand.Read(b); err != nil {
		ts := time.Now().UnixNano()
		return fmt.Sprintf("state-%d", ts)
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

// logOAuthProfile dumps profile and connector info for debugging OAuth payloads
func logOAuthProfile(profile *oauthprovider.OAuthUserProfile, connector *model.OAuthConnector, err error) {
	if profile == nil || connector == nil {
		return
	}
	common.Log.Debugf("OAuth profile dump: provider=%s connectorID=%d profile=%+v err=%v", connector.Provider, connector.ID, profile, err)
}

// saveWebhooks 保存连接器的 webhook 列表（先删旧，再插新）
func saveWebhooks(connectorID uint, reqs []request.OAuthWebhookReq) error {
	// 删除旧的
	if err := common.DB.Where("oauth_connector_id = ?", connectorID).Delete(&model.OAuthWebhook{}).Error; err != nil {
		return err
	}
	// 插入新的
	for _, req := range reqs {
		eventsJSON, _ := stdjson.Marshal(req.Events)
		fieldsJSON, _ := stdjson.Marshal(req.Fields)
		wh := model.OAuthWebhook{
			OAuthConnectorID: connectorID,
			URL:              req.URL,
			Events:           datatypes.JSON(eventsJSON),
			Fields:           datatypes.JSON(fieldsJSON),
			Enabled:          req.Enabled,
			Description:      req.Description,
		}
		if err := common.DB.Create(&wh).Error; err != nil {
			return err
		}
	}
	return nil
}
