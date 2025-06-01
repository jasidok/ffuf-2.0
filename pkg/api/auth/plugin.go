// Package auth provides authentication mechanisms for API testing.
package auth

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// AuthType for custom plugins
const (
	// AuthCustom represents a custom authentication scheme
	AuthCustom AuthType = 100 // Starting from 100 to avoid conflicts with built-in types
)

// CustomAuthProvider is a function that creates a new AuthProvider
type CustomAuthProvider func(config map[string]string) (AuthProvider, error)

// PluginRegistry holds registered custom authentication providers
type PluginRegistry struct {
	providers map[string]CustomAuthProvider
	mu        sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		providers: make(map[string]CustomAuthProvider),
	}
}

// Register adds a custom authentication provider to the registry
func (r *PluginRegistry) Register(name string, provider CustomAuthProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("auth provider with name '%s' already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a custom authentication provider from the registry
func (r *PluginRegistry) Get(name string) (CustomAuthProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("auth provider with name '%s' not found", name)
	}

	return provider, nil
}

// List returns all registered provider names
func (r *PluginRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Unregister removes a custom authentication provider from the registry
func (r *PluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("auth provider with name '%s' not found", name)
	}

	delete(r.providers, name)
	return nil
}

// Create instantiates a custom authentication provider with the given configuration
func (r *PluginRegistry) Create(name string, config map[string]string) (AuthProvider, error) {
	provider, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return provider(config)
}

// CustomAuth is a base implementation of a custom authentication provider
type CustomAuth struct {
	Name        string
	Description string
	Config      map[string]string
	AuthFunc    func(req *http.Request, config map[string]string) error
}

// NewCustomAuth creates a new CustomAuth provider
func NewCustomAuth(name, description string, config map[string]string, authFunc func(req *http.Request, config map[string]string) error) *CustomAuth {
	return &CustomAuth{
		Name:        name,
		Description: description,
		Config:      config,
		AuthFunc:    authFunc,
	}
}

// AddAuth adds custom authentication to the given HTTP request
func (a *CustomAuth) AddAuth(req *http.Request) error {
	if a.AuthFunc == nil {
		return api.NewAPIError("Custom auth function is not defined", 0)
	}
	return a.AuthFunc(req, a.Config)
}

// GetAuthType returns the type of authentication
func (a *CustomAuth) GetAuthType() AuthType {
	return AuthCustom
}

// GetDescription returns a human-readable description of the authentication
func (a *CustomAuth) GetDescription() string {
	if a.Description != "" {
		return fmt.Sprintf("Custom Auth (%s)", a.Description)
	}
	return fmt.Sprintf("Custom Auth (%s)", a.Name)
}

// DefaultRegistry is the global plugin registry
var DefaultRegistry = NewPluginRegistry()

// RegisterAuthProvider registers a custom authentication provider with the default registry
func RegisterAuthProvider(name string, provider CustomAuthProvider) error {
	return DefaultRegistry.Register(name, provider)
}

// GetAuthProvider retrieves a custom authentication provider from the default registry
func GetAuthProvider(name string) (CustomAuthProvider, error) {
	return DefaultRegistry.Get(name)
}

// ListAuthProviders returns all registered provider names from the default registry
func ListAuthProviders() []string {
	return DefaultRegistry.List()
}

// UnregisterAuthProvider removes a custom authentication provider from the default registry
func UnregisterAuthProvider(name string) error {
	return DefaultRegistry.Unregister(name)
}

// CreateAuthProvider instantiates a custom authentication provider with the given configuration
func CreateAuthProvider(name string, config map[string]string) (AuthProvider, error) {
	return DefaultRegistry.Create(name, config)
}