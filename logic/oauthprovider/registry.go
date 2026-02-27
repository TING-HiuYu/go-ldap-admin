package oauthprovider

import (
	"strings"

	"github.com/eryajf/go-ldap-admin/model/response"
)

var providerRegistry = map[string]Provider{}

// Register adds a provider to the registry.
func Register(p Provider) {
	if p == nil {
		return
	}
	providerRegistry[strings.ToLower(p.ID())] = p
}

// Get returns the provider by ID.
func Get(id string) Provider {
	return providerRegistry[strings.ToLower(strings.TrimSpace(id))]
}

// ListOptions returns provider options for UI.
func ListOptions() []response.OAuthProviderOption {
	options := make([]response.OAuthProviderOption, 0, len(providerRegistry))
	for _, p := range providerRegistry {
		options = append(options, response.OAuthProviderOption{Label: p.Name(), Value: p.ID()})
	}
	return options
}

// ListSchemas returns provider schemas for UI.
func ListSchemas() []response.OAuthProviderSchema {
	schemas := make([]response.OAuthProviderSchema, 0, len(providerRegistry))
	for _, p := range providerRegistry {
		schemas = append(schemas, p.Schema())
	}
	return schemas
}
