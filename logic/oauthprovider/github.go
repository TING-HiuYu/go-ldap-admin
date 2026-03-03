package oauthprovider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/response"
	"github.com/eryajf/go-ldap-admin/public/tools"
)

type githubProvider struct{}

type githubConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURI  string `json:"redirectUri"`
	Scope        string `json:"scope"`
	BaseURL      string `json:"baseUrl"`
	APIBase      string `json:"apiBase"`
}

type githubProfile struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

var (
	githubHTTPClient = &http.Client{Timeout: 10 * time.Second}
	githubProxyBase  = "https://github.com"
	githubProxyAPI   = "https://api.github.com"
)

func init() {
	Register(githubProvider{})
}

func (p githubProvider) ID() string {
	return "github"
}

func (p githubProvider) Name() string {
	return "GitHub"
}

func (p githubProvider) Schema() response.OAuthProviderSchema {
	return response.OAuthProviderSchema{
		ID:    p.ID(),
		Name:  p.Name(),
		Title: "GitHub 配置",
		Fields: []response.OAuthProviderField{
			{Key: "clientId", Label: "Client ID", Required: true, Placeholder: "GitHub OAuth App Client ID"},
			{Key: "clientSecret", Label: "Client Secret", Required: true, Placeholder: "Client Secret", InputType: "password"},
			{Key: "redirectUri", Label: "Redirect URI", Required: true, Placeholder: "https://your.domain/api/oauth/callback/github", DefaultValue: githubRedirectDefault()},
			{Key: "scope", Label: "Scope", Required: false, Placeholder: "read:user user:email", DefaultValue: "read:user user:email"},
			{Key: "baseUrl", Label: "Base URL", Required: false, Placeholder: "自建 GitHub Enterprise 可填"},
			{Key: "apiBase", Label: "API Base", Required: false, Placeholder: "API 域名 (可选)"},
		},
	}
}

func (p githubProvider) ValidateConfig(connector *model.OAuthConnector) error {
	cfg, err := p.parseConfig(connector)
	if err != nil {
		return tools.NewValidatorError(fmt.Errorf("解析Github配置失败: %v", err))
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return tools.NewValidatorError(fmt.Errorf("Github配置缺少clientId/clientSecret/redirectUri"))
	}
	_, err = normalizeRedirectURI(cfg.RedirectURI)
	if err != nil {
		return err
	}
	return nil
}

func (p githubProvider) BuildAuthorizeURL(connector *model.OAuthConnector, state string) (string, error) {
	cfg, err := p.parseConfig(connector)
	if err != nil {
		return "", tools.NewValidatorError(fmt.Errorf("解析Github配置失败: %v", err))
	}
	redirectURI, err := normalizeRedirectURI(cfg.RedirectURI)
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(cfg.BaseURL, "/")
	if base == "" {
		base = "https://github.com"
	}
	scope := cfg.Scope
	if strings.TrimSpace(scope) == "" {
		scope = "read:user user:email"
	}
	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("scope", scope)
	values.Set("state", state)
	return fmt.Sprintf("%s/login/oauth/authorize?%s", base, values.Encode()), nil
}

func (p githubProvider) FetchProfile(connector *model.OAuthConnector, code string) (*OAuthUserProfile, error) {
	cfg, err := p.parseConfig(connector)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析Github配置失败: %v", err))
	}
	redirectURI, err := normalizeRedirectURI(cfg.RedirectURI)
	if err != nil {
		return nil, err
	}

	// ── 确定后端实际请求的目标地址 ──
	// 原则：只有后端向 GitHub 发起 HTTP 请求（取 token、取用户信息）时才走代理；
	//       面向客户端/浏览器的 URL（如 BuildAuthorizeURL）永远使用 connector 配置。
	//       如果 connector 配置了自建 GitHub Enterprise 地址，则不使用代理。
	configBase := strings.TrimSuffix(cfg.BaseURL, "/")
	if configBase == "" {
		configBase = "https://github.com"
	}
	configAPI := strings.TrimSuffix(cfg.APIBase, "/")
	if configAPI == "" {
		configAPI = "https://api.github.com"
	}

	// 仅当 connector 使用的是公共 github.com 且代理地址已配置时，后端走代理
	tokenBase := configBase
	apiBase := configAPI
	if configBase == "https://github.com" && githubProxyBase != "" {
		tokenBase = strings.TrimSuffix(githubProxyBase, "/")
	}
	if configAPI == "https://api.github.com" && githubProxyAPI != "" {
		apiBase = strings.TrimSuffix(githubProxyAPI, "/")
	}

	tokenURL := fmt.Sprintf("%s/login/oauth/access_token", tokenBase)
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)

	req, _ := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := githubHTTPClient.Do(req)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取GitHub token失败: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, tools.NewValidatorError(fmt.Errorf("获取GitHub token失败: %s", string(body)))
	}
	var tokenBody map[string]any
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &tokenBody); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析GitHub token响应失败: %v", err))
	}
	accessToken, _ := tokenBody["access_token"].(string)
	if accessToken == "" {
		return nil, tools.NewValidatorError(errors.New("GitHub未返回access_token"))
	}

	profileURL := fmt.Sprintf("%s/user", apiBase)
	profReq, _ := http.NewRequest(http.MethodGet, profileURL, nil)
	profReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	profReq.Header.Set("Accept", "application/json")
	profReq.Header.Set("User-Agent", "go-ldap-admin")

	profResp, err := githubHTTPClient.Do(profReq)
	if err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("获取GitHub用户信息失败: %v", err))
	}
	defer profResp.Body.Close()
	if profResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(profResp.Body)
		return nil, tools.NewValidatorError(fmt.Errorf("获取GitHub用户信息失败: %s", string(body)))
	}
	var prof githubProfile
	profBody, _ := io.ReadAll(profResp.Body)
	if err := json.Unmarshal(profBody, &prof); err != nil {
		return nil, tools.NewValidatorError(fmt.Errorf("解析GitHub用户信息失败: %v", err))
	}

	email := strings.TrimSpace(prof.Email)
	if email == "" {
		email, err = fetchGitHubPrimaryEmail(apiBase, accessToken)
		if err != nil {
			return nil, err
		}
	}
	if email == "" {
		return nil, tools.NewValidatorError(fmt.Errorf("GitHub账号未公开邮箱，无法登录"))
	}

	username := strings.ToLower(strings.TrimSpace(prof.Login))
	if username == "" {
		username = strings.Split(email, "@")[0]
	}
	name := strings.TrimSpace(prof.Name)
	if name == "" {
		name = username
	}

	return &OAuthUserProfile{ID: fmt.Sprintf("%d", prof.ID), Username: username, Name: name, Email: email, Avatar: prof.AvatarURL}, nil
}

func (p githubProvider) parseConfig(connector *model.OAuthConnector) (*githubConfig, error) {
	var cfg githubConfig
	if err := json.Unmarshal(connector.Config, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func fetchGitHubPrimaryEmail(apiBase, token string) (string, error) {
	urlStr := fmt.Sprintf("%s/user/emails", apiBase)
	req, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-ldap-admin")

	resp, err := githubHTTPClient.Do(req)
	if err != nil {
		return "", tools.NewValidatorError(fmt.Errorf("获取邮箱失败: %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", tools.NewValidatorError(fmt.Errorf("获取邮箱失败: %s", string(body)))
	}
	var emails []githubEmail
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", tools.NewValidatorError(fmt.Errorf("解析邮箱失败: %v", err))
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", nil
}

func normalizeRedirectURI(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", tools.NewValidatorError(fmt.Errorf("Github配置缺少redirectUri"))
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", tools.NewValidatorError(fmt.Errorf("redirectUri 无效: %v", err))
	}
	prefix := strings.Trim(strings.TrimSpace(config.Conf.System.UrlPathPrefix), "/")
	if prefix == "" {
		return parsed.String(), nil
	}
	prefixPath := "/" + prefix
	pathVal := parsed.Path
	if pathVal == "" {
		pathVal = "/"
	}
	if !strings.HasPrefix(pathVal, prefixPath+"/") && pathVal != prefixPath {
		cleanPath := "/" + strings.TrimPrefix(pathVal, "/")
		parsed.Path = prefixPath + cleanPath
	}
	return parsed.String(), nil
}

func githubRedirectDefault() string {
	prefix := strings.Trim(strings.TrimSpace(config.Conf.System.UrlPathPrefix), "/")
	if prefix == "" {
		return "/oauth/callback/github"
	}
	return "/" + prefix + "/oauth/callback/github"
}
