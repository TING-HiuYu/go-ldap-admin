package oauthprovider

import (
	"github.com/eryajf/go-ldap-admin/model"
	"github.com/eryajf/go-ldap-admin/model/response"
)

// OAuthUserProfile describes a normalized OAuth user profile.
type OAuthUserProfile struct {
	ID       string
	Username string
	Name     string
	Email    string
	Avatar   string
}

// Provider defines the OAuth provider contract.
type Provider interface {
	ID() string
	Name() string
	Schema() response.OAuthProviderSchema
	ValidateConfig(connector *model.OAuthConnector) error
	BuildAuthorizeURL(connector *model.OAuthConnector, state string) (string, error)
	FetchProfile(connector *model.OAuthConnector, code string) (*OAuthUserProfile, error)
}
