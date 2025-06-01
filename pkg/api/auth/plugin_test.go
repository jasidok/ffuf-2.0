package auth

import (
	"net/http"
	"testing"
)

func TestPluginRegistry(t *testing.T) {
	// Create a new registry
	registry := NewPluginRegistry()

	// Register a custom auth provider
	err := registry.Register("test-provider", func(config map[string]string) (AuthProvider, error) {
		return NewCustomAuth("test-provider", "Test Provider", config, func(req *http.Request, config map[string]string) error {
			req.Header.Set("X-Test-Auth", config["token"])
			return nil
		}), nil
	})
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Try to register the same provider again (should fail)
	err = registry.Register("test-provider", nil)
	if err == nil {
		t.Error("Expected error when registering duplicate provider, got nil")
	}

	// Get the provider
	provider, err := registry.Get("test-provider")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider is nil")
	}

	// Get a non-existent provider (should fail)
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent provider, got nil")
	}

	// List providers
	providers := registry.List()
	if len(providers) != 1 || providers[0] != "test-provider" {
		t.Errorf("Expected [test-provider], got %v", providers)
	}

	// Create an auth provider instance
	config := map[string]string{"token": "test-token"}
	auth, err := registry.Create("test-provider", config)
	if err != nil {
		t.Fatalf("Failed to create auth provider: %v", err)
	}

	// Check the auth type
	if auth.GetAuthType() != AuthCustom {
		t.Errorf("Expected auth type AuthCustom, got %v", auth.GetAuthType())
	}

	// Create a test request
	req, err := http.NewRequest("GET", "https://api.example.com/path", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add authentication to the request
	err = auth.AddAuth(req)
	if err != nil {
		t.Fatalf("AddAuth failed: %v", err)
	}

	// Check that the custom header was added
	if req.Header.Get("X-Test-Auth") != "test-token" {
		t.Errorf("Expected X-Test-Auth header to be test-token, got %s", req.Header.Get("X-Test-Auth"))
	}

	// Unregister the provider
	err = registry.Unregister("test-provider")
	if err != nil {
		t.Fatalf("Failed to unregister provider: %v", err)
	}

	// List providers again (should be empty)
	providers = registry.List()
	if len(providers) != 0 {
		t.Errorf("Expected empty list, got %v", providers)
	}

	// Unregister a non-existent provider (should fail)
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent provider, got nil")
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Clean up any existing registrations
	for _, name := range ListAuthProviders() {
		UnregisterAuthProvider(name)
	}

	// Register a custom auth provider
	err := RegisterAuthProvider("default-test", func(config map[string]string) (AuthProvider, error) {
		return NewCustomAuth("default-test", "Default Test", config, func(req *http.Request, config map[string]string) error {
			req.Header.Set("X-Default-Test", config["value"])
			return nil
		}), nil
	})
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Get the provider
	provider, err := GetAuthProvider("default-test")
	if err != nil {
		t.Fatalf("Failed to get provider: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider is nil")
	}

	// List providers
	providers := ListAuthProviders()
	if len(providers) != 1 || providers[0] != "default-test" {
		t.Errorf("Expected [default-test], got %v", providers)
	}

	// Create an auth provider instance
	config := map[string]string{"value": "test-value"}
	auth, err := CreateAuthProvider("default-test", config)
	if err != nil {
		t.Fatalf("Failed to create auth provider: %v", err)
	}

	// Check the description
	description := auth.GetDescription()
	if description != "Custom Auth (Default Test)" {
		t.Errorf("Expected description 'Custom Auth (Default Test)', got '%s'", description)
	}

	// Create a test request
	req, err := http.NewRequest("GET", "https://api.example.com/path", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add authentication to the request
	err = auth.AddAuth(req)
	if err != nil {
		t.Fatalf("AddAuth failed: %v", err)
	}

	// Check that the custom header was added
	if req.Header.Get("X-Default-Test") != "test-value" {
		t.Errorf("Expected X-Default-Test header to be test-value, got %s", req.Header.Get("X-Default-Test"))
	}

	// Unregister the provider
	err = UnregisterAuthProvider("default-test")
	if err != nil {
		t.Fatalf("Failed to unregister provider: %v", err)
	}
}

func TestCustomAuth(t *testing.T) {
	// Test with nil auth function
	auth := NewCustomAuth("nil-func", "Nil Function Test", nil, nil)
	req, _ := http.NewRequest("GET", "https://api.example.com/path", nil)
	err := auth.AddAuth(req)
	if err == nil {
		t.Error("Expected error with nil auth function, got nil")
	}

	// Test with valid auth function
	auth = NewCustomAuth("valid-func", "Valid Function Test", map[string]string{"header": "X-Custom", "value": "custom-value"}, func(req *http.Request, config map[string]string) error {
		req.Header.Set(config["header"], config["value"])
		return nil
	})

	req, _ = http.NewRequest("GET", "https://api.example.com/path", nil)
	err = auth.AddAuth(req)
	if err != nil {
		t.Fatalf("AddAuth failed: %v", err)
	}

	// Check that the custom header was added
	if req.Header.Get("X-Custom") != "custom-value" {
		t.Errorf("Expected X-Custom header to be custom-value, got %s", req.Header.Get("X-Custom"))
	}

	// Test description without custom description
	auth = NewCustomAuth("no-desc", "", nil, nil)
	description := auth.GetDescription()
	if description != "Custom Auth (no-desc)" {
		t.Errorf("Expected description 'Custom Auth (no-desc)', got '%s'", description)
	}
}