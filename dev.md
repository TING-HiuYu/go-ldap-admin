# OAuth Connector Development Guide

This document describes how to add a new OAuth provider connector to **go-ldap-admin**. The existing GitHub connector (`logic/oauthprovider/github.go`) serves as the reference implementation.

---

## Architecture Overview

The OAuth subsystem follows a provider-registry pattern:

```
controller/oauth_controller.go   ← HTTP handlers (Swagger-annotated)
       │
routes/oauth_routes.go           ← Route registration
       │
logic/oauth_logic.go             ← Business logic (create/update/delete/callback)
       │
logic/oauthprovider/
  ├── types.go                   ← Provider interface + OAuthUserProfile
  ├── registry.go                ← Global provider registry
  └── github.go                  ← GitHub provider implementation (reference)
       │
model/
  ├── oauth_connector.go         ← OAuthConnector DB model
  └── oauth_webhook.go           ← OAuthWebhook DB model
       │
logic/webhook/
  ├── types.go                   ← Webhook event types + payload builder
  └── dispatcher.go              ← Async webhook dispatcher
```

### Data Flow

1. **Admin** creates a connector via `POST /api/oauth/connectors` with provider-specific config (client ID, secret, scopes, etc.).
2. **Login page** fetches enabled connectors via `GET /api/oauth/connectors/public`.
3. **User** clicks a provider button → frontend calls `POST /api/oauth/start` → backend returns an authorization URL.
4. **User** is redirected to the OAuth provider, authorizes, then is redirected back to `GET /api/oauth/callback/:provider`.
5. **Backend** exchanges the authorization code for an access token, fetches the user profile, creates or links the user account, issues a JWT, and returns an HTML page that posts the result to the opener window.

---

## Step-by-Step: Adding a New Provider

### 1. Create the Provider File

Create a new file under `logic/oauthprovider/`, e.g. `gitlab.go`:

```go
package oauthprovider

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/eryajf/go-ldap-admin/model"
    "github.com/eryajf/go-ldap-admin/model/response"
    "github.com/eryajf/go-ldap-admin/public/tools"
)
```

### 2. Define Provider-Specific Types

Define a config struct matching the JSON fields stored in `OAuthConnector.Config`, and a profile struct for the provider's API response:

```go
// gitlabConfig holds the OAuth configuration for a GitLab connector.
type gitlabConfig struct {
    ClientID     string `json:"clientId"`
    ClientSecret string `json:"clientSecret"`
    RedirectURI  string `json:"redirectUri"`
    Scope        string `json:"scope"`
    BaseURL      string `json:"baseUrl"` // e.g. https://gitlab.example.com
}

// gitlabProfile represents the user profile from the GitLab API.
type gitlabProfile struct {
    ID        int64  `json:"id"`
    Username  string `json:"username"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    AvatarURL string `json:"avatar_url"`
}
```

### 3. Implement the `Provider` Interface

The `Provider` interface (defined in `types.go`) requires six methods:

```go
type Provider interface {
    ID() string
    Name() string
    Schema() response.OAuthProviderSchema
    ValidateConfig(connector *model.OAuthConnector) error
    BuildAuthorizeURL(connector *model.OAuthConnector, state string) (string, error)
    FetchProfile(connector *model.OAuthConnector, code string) (*OAuthUserProfile, error)
}
```

#### 3a. `ID()` and `Name()`

`ID()` returns the lowercase identifier stored in the database. `Name()` returns the human-readable display name.

```go
type gitlabProvider struct{}

func (p gitlabProvider) ID() string   { return "gitlab" }
func (p gitlabProvider) Name() string { return "GitLab" }
```

#### 3b. `Schema()`

Returns the configuration schema used by the admin UI to render the connector creation form:

```go
func (p gitlabProvider) Schema() response.OAuthProviderSchema {
    return response.OAuthProviderSchema{
        ID:    p.ID(),
        Name:  p.Name(),
        Title: "GitLab Configuration",
        Fields: []response.OAuthProviderField{
            {Key: "clientId", Label: "Application ID", Required: true, Placeholder: "GitLab Application ID"},
            {Key: "clientSecret", Label: "Secret", Required: true, Placeholder: "Application secret", InputType: "password"},
            {Key: "redirectUri", Label: "Redirect URI", Required: true, Placeholder: "https://your.domain/api/oauth/callback/gitlab"},
            {Key: "scope", Label: "Scope", Required: false, Placeholder: "read_user", DefaultValue: "read_user"},
            {Key: "baseUrl", Label: "Base URL", Required: false, Placeholder: "https://gitlab.com"},
        },
    }
}
```

#### 3c. `ValidateConfig()`

Parses and validates the connector's JSON config. Return a `tools.NewValidatorError` for user-facing errors:

```go
func (p gitlabProvider) ValidateConfig(connector *model.OAuthConnector) error {
    cfg, err := p.parseConfig(connector)
    if err != nil {
        return tools.NewValidatorError(fmt.Errorf("failed to parse GitLab config: %v", err))
    }
    if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
        return tools.NewValidatorError(fmt.Errorf("GitLab config requires clientId, clientSecret, and redirectUri"))
    }
    return nil
}

func (p gitlabProvider) parseConfig(connector *model.OAuthConnector) (*gitlabConfig, error) {
    var cfg gitlabConfig
    if err := json.Unmarshal(connector.Config, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

#### 3d. `BuildAuthorizeURL()`

Constructs the OAuth authorization URL that the user's browser will be redirected to:

```go
func (p gitlabProvider) BuildAuthorizeURL(connector *model.OAuthConnector, state string) (string, error) {
    cfg, err := p.parseConfig(connector)
    if err != nil {
        return "", tools.NewValidatorError(fmt.Errorf("failed to parse GitLab config: %v", err))
    }
    base := strings.TrimSuffix(cfg.BaseURL, "/")
    if base == "" {
        base = "https://gitlab.com"
    }
    scope := cfg.Scope
    if scope == "" {
        scope = "read_user"
    }
    v := url.Values{}
    v.Set("client_id", cfg.ClientID)
    v.Set("redirect_uri", cfg.RedirectURI)
    v.Set("response_type", "code")
    v.Set("scope", scope)
    v.Set("state", state)
    return fmt.Sprintf("%s/oauth/authorize?%s", base, v.Encode()), nil
}
```

#### 3e. `FetchProfile()`

The core method. It exchanges the authorization code for an access token, fetches the user profile, and returns a normalized `OAuthUserProfile`:

```go
var gitlabHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (p gitlabProvider) FetchProfile(connector *model.OAuthConnector, code string) (*OAuthUserProfile, error) {
    cfg, err := p.parseConfig(connector)
    if err != nil {
        return nil, tools.NewValidatorError(fmt.Errorf("failed to parse GitLab config: %v", err))
    }
    base := strings.TrimSuffix(cfg.BaseURL, "/")
    if base == "" {
        base = "https://gitlab.com"
    }

    // Step 1: Exchange code for access token
    form := url.Values{}
    form.Set("client_id", cfg.ClientID)
    form.Set("client_secret", cfg.ClientSecret)
    form.Set("code", code)
    form.Set("redirect_uri", cfg.RedirectURI)
    form.Set("grant_type", "authorization_code")

    tokenReq, _ := http.NewRequest(http.MethodPost, base+"/oauth/token", strings.NewReader(form.Encode()))
    tokenReq.Header.Set("Accept", "application/json")
    tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    tokenResp, err := gitlabHTTPClient.Do(tokenReq)
    if err != nil {
        return nil, tools.NewValidatorError(fmt.Errorf("failed to get GitLab token: %v", err))
    }
    defer tokenResp.Body.Close()

    var tokenBody map[string]any
    body, _ := io.ReadAll(tokenResp.Body)
    if err := json.Unmarshal(body, &tokenBody); err != nil {
        return nil, tools.NewValidatorError(fmt.Errorf("failed to parse GitLab token response: %v", err))
    }
    accessToken, _ := tokenBody["access_token"].(string)
    if accessToken == "" {
        return nil, tools.NewValidatorError(fmt.Errorf("GitLab did not return an access_token"))
    }

    // Step 2: Fetch user profile
    profReq, _ := http.NewRequest(http.MethodGet, base+"/api/v4/user", nil)
    profReq.Header.Set("Authorization", "Bearer "+accessToken)
    profReq.Header.Set("Accept", "application/json")

    profResp, err := gitlabHTTPClient.Do(profReq)
    if err != nil {
        return nil, tools.NewValidatorError(fmt.Errorf("failed to get GitLab user profile: %v", err))
    }
    defer profResp.Body.Close()

    var prof gitlabProfile
    profBody, _ := io.ReadAll(profResp.Body)
    if err := json.Unmarshal(profBody, &prof); err != nil {
        return nil, tools.NewValidatorError(fmt.Errorf("failed to parse GitLab user profile: %v", err))
    }

    if prof.Email == "" {
        return nil, tools.NewValidatorError(fmt.Errorf("GitLab account has no email; cannot proceed"))
    }

    // Step 3: Return normalized profile
    username := strings.ToLower(strings.TrimSpace(prof.Username))
    if username == "" {
        username = strings.Split(prof.Email, "@")[0]
    }
    name := strings.TrimSpace(prof.Name)
    if name == "" {
        name = username
    }

    return &OAuthUserProfile{
        ID:       fmt.Sprintf("%d", prof.ID),
        Username: username,
        Name:     name,
        Email:    prof.Email,
        Avatar:   prof.AvatarURL,
    }, nil
}
```

### 4. Register the Provider

Use an `init()` function to register the provider at package load time:

```go
func init() {
    Register(gitlabProvider{})
}
```

The `Register` function in `registry.go` adds the provider to the global map. No other registration or wiring is needed—the system discovers the provider automatically.

### 5. Verify

After adding the file:

1. **Build**: `go build ./...` — the new provider is loaded via `init()`.
2. **Check the admin UI**: The provider should appear in the "Provider" dropdown when creating a new connector.
3. **Create a connector**: Fill in the configuration fields from your `Schema()`.
4. **Test the login flow**: Click the provider button on the login page and complete the OAuth flow.

---

## Key Design Decisions

| Concern | Approach |
|---|---|
| **Config storage** | Provider-specific config is stored as JSON in `OAuthConnector.Config`. Each provider parses its own schema. |
| **User creation** | `oauth_logic.go → ensureUserFromOAuth()` creates a new user if the email does not exist, otherwise links the existing user. A random password is generated for new users. |
| **JWT issuance** | After successful callback, a JWT is issued using the same `auth.TokenGenerator` pipeline as normal login. |
| **Webhooks** | `logic/webhook/dispatcher.go` fires `user.created` events asynchronously after OAuth-based user provisioning. |
| **Redirect URI normalization** | `normalizeRedirectURI()` in `github.go` automatically prepends the system URL path prefix. Reuse or adapt this for your provider. |
| **Error handling** | Return `tools.NewValidatorError(err)` for user-facing errors. The controller layer renders them as HTML in the callback or JSON in API responses. |

---

## File Checklist for a New Provider

| # | File | Action |
|---|---|---|
| 1 | `logic/oauthprovider/<provider>.go` | **Create** — implement `Provider` interface + `init()` registration |
| 2 | `model/response/oauth_rsp.go` | No change needed (schemas are dynamic) |
| 3 | `model/request/oauth_req.go` | No change needed |
| 4 | `controller/oauth_controller.go` | No change needed |
| 5 | `routes/oauth_routes.go` | No change needed |
| 6 | `logic/oauth_logic.go` | No change needed |
| 7 | `logic/webhook/types.go` | Add new event types if needed |

In most cases, **only one file needs to be created**. The rest of the system discovers and uses the new provider automatically through the registry.

---

## Reference: GitHub Provider

See `logic/oauthprovider/github.go` for the complete, production-ready implementation. Key aspects to note:

- **Proxy support**: The GitHub provider optionally routes backend HTTP requests through a proxy for environments where `github.com` is not directly accessible. You may omit this for providers that don't need it.
- **Email fallback**: If the primary profile endpoint doesn't return an email, the GitHub provider calls `/user/emails` as a fallback. Consider whether your provider needs similar handling.
- **Redirect URI normalization**: The `normalizeRedirectURI` helper automatically applies the system's URL path prefix. Reuse this function or implement similar logic for your provider.

---

## Testing

There is currently no automated test infrastructure for OAuth providers. To test a new provider:

1. Set up an OAuth application with the provider (e.g., create a GitLab OAuth app).
2. Create a connector via the admin UI with the provider's credentials.
3. Initiate the OAuth login flow from the login page.
4. Verify that:
   - The authorization URL redirects correctly.
   - The callback exchanges the code for a token.
   - The user profile is fetched and a user account is created/linked.
   - A JWT is issued and the user is logged in.
   - Webhooks fire for `user.created` events (if configured).
