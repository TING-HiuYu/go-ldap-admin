package oauthprovider

import (
	"strings"

	"github.com/eryajf/go-ldap-admin/model/response"
)

// providerRegistry maps lowercase provider IDs to their implementations.
var providerRegistry = map[string]Provider{}

// Register adds a provider implementation to the global registry.
// Nil providers are silently ignored.
func Register(p Provider) {
	if p == nil {
		return
	}
	providerRegistry[strings.ToLower(p.ID())] = p
}

// Get returns the registered provider for the given ID, or nil if not found.
func Get(id string) Provider {
	return providerRegistry[strings.ToLower(strings.TrimSpace(id))]
}

// ListOptions returns provider label/value pairs for UI dropdowns.
func ListOptions() []response.OAuthProviderOption {
	options := make([]response.OAuthProviderOption, 0, len(providerRegistry))
	for _, p := range providerRegistry {
		options = append(options, response.OAuthProviderOption{Label: p.Name(), Value: p.ID()})
	}
	return options
}

// ListSchemas returns the configuration schemas of all registered providers for UI rendering.
func ListSchemas() []response.OAuthProviderSchema {
	schemas := make([]response.OAuthProviderSchema, 0, len(providerRegistry))
	for _, p := range providerRegistry {
		schemas = append(schemas, p.Schema())
	}
	return schemas
}
