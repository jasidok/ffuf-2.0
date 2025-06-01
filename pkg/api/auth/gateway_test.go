package auth

import (
	"net/http"
	"strings"
	"testing"
)

func TestAWSGatewayAuth(t *testing.T) {
	// Create a new AWSGatewayAuth provider
	auth := NewAWSGatewayAuth("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "us-west-2")

	// Check the gateway type
	if auth.GetGatewayType() != GatewayAWS {
		t.Errorf("Expected gateway type GatewayAWS, got %v", auth.GetGatewayType())
	}

	// Check the auth type
	if auth.GetAuthType() != AuthAPIKey {
		t.Errorf("Expected auth type AuthAPIKey, got %v", auth.GetAuthType())
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

	// Check that the necessary headers were added
	if req.Header.Get("X-Amz-Date") == "" {
		t.Error("X-Amz-Date header not set")
	}
	if req.Header.Get("Authorization") == "" {
		t.Error("Authorization header not set")
	}
	if !containsSubstring(req.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
		t.Error("Authorization header does not contain AWS4-HMAC-SHA256")
	}
	if !containsSubstring(req.Header.Get("Authorization"), "Credential=AKIAIOSFODNN7EXAMPLE") {
		t.Error("Authorization header does not contain the access key")
	}
}

func TestAzureGatewayAuth(t *testing.T) {
	// Create a new AzureGatewayAuth provider
	auth := NewAzureGatewayAuth("subscription-key-example", "")

	// Check the gateway type
	if auth.GetGatewayType() != GatewayAzure {
		t.Errorf("Expected gateway type GatewayAzure, got %v", auth.GetGatewayType())
	}

	// Check the auth type
	if auth.GetAuthType() != AuthAPIKey {
		t.Errorf("Expected auth type AuthAPIKey, got %v", auth.GetAuthType())
	}

	// Check the default key name
	if auth.KeyName != "Ocp-Apim-Subscription-Key" {
		t.Errorf("Expected default key name Ocp-Apim-Subscription-Key, got %s", auth.KeyName)
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

	// Check that the subscription key header was added
	if req.Header.Get("Ocp-Apim-Subscription-Key") != "subscription-key-example" {
		t.Errorf("Expected Ocp-Apim-Subscription-Key header to be subscription-key-example, got %s", req.Header.Get("Ocp-Apim-Subscription-Key"))
	}

	// Test with custom key name
	authCustom := NewAzureGatewayAuth("subscription-key-example", "Custom-Key-Name")
	if authCustom.KeyName != "Custom-Key-Name" {
		t.Errorf("Expected custom key name Custom-Key-Name, got %s", authCustom.KeyName)
	}

	req, _ = http.NewRequest("GET", "https://api.example.com/path", nil)
	authCustom.AddAuth(req)
	if req.Header.Get("Custom-Key-Name") != "subscription-key-example" {
		t.Errorf("Expected Custom-Key-Name header to be subscription-key-example, got %s", req.Header.Get("Custom-Key-Name"))
	}
}

func TestGoogleGatewayAuth(t *testing.T) {
	// Create a new GoogleGatewayAuth provider
	auth := NewGoogleGatewayAuth("google-api-key-example")

	// Check the gateway type
	if auth.GetGatewayType() != GatewayGoogle {
		t.Errorf("Expected gateway type GatewayGoogle, got %v", auth.GetGatewayType())
	}

	// Check the auth type
	if auth.GetAuthType() != AuthAPIKey {
		t.Errorf("Expected auth type AuthAPIKey, got %v", auth.GetAuthType())
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

	// Check that the API key was added as a query parameter
	if req.URL.Query().Get("key") != "google-api-key-example" {
		t.Errorf("Expected key query parameter to be google-api-key-example, got %s", req.URL.Query().Get("key"))
	}
}

func TestKongGatewayAuth(t *testing.T) {
	// Test with header-based authentication
	authHeader := NewKongGatewayAuth("kong-api-key-example", "apikey", true)

	// Check the gateway type
	if authHeader.GetGatewayType() != GatewayKong {
		t.Errorf("Expected gateway type GatewayKong, got %v", authHeader.GetGatewayType())
	}

	// Check the auth type
	if authHeader.GetAuthType() != AuthAPIKey {
		t.Errorf("Expected auth type AuthAPIKey, got %v", authHeader.GetAuthType())
	}

	// Create a test request
	req, err := http.NewRequest("GET", "https://api.example.com/path", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add authentication to the request
	err = authHeader.AddAuth(req)
	if err != nil {
		t.Fatalf("AddAuth failed: %v", err)
	}

	// Check that the API key was added as a header
	if req.Header.Get("apikey") != "kong-api-key-example" {
		t.Errorf("Expected apikey header to be kong-api-key-example, got %s", req.Header.Get("apikey"))
	}

	// Test with query parameter-based authentication
	authQuery := NewKongGatewayAuth("kong-api-key-example", "apikey", false)
	req, _ = http.NewRequest("GET", "https://api.example.com/path", nil)
	authQuery.AddAuth(req)

	// Check that the API key was added as a query parameter
	if req.URL.Query().Get("apikey") != "kong-api-key-example" {
		t.Errorf("Expected apikey query parameter to be kong-api-key-example, got %s", req.URL.Query().Get("apikey"))
	}

	// Test with default key name
	authDefault := NewKongGatewayAuth("kong-api-key-example", "", true)
	if authDefault.KeyName != "apikey" {
		t.Errorf("Expected default key name apikey, got %s", authDefault.KeyName)
	}
}

func TestTykGatewayAuth(t *testing.T) {
	// Create a new TykGatewayAuth provider
	auth := NewTykGatewayAuth("tyk-api-key-example")

	// Check the gateway type
	if auth.GetGatewayType() != GatewayTyk {
		t.Errorf("Expected gateway type GatewayTyk, got %v", auth.GetGatewayType())
	}

	// Check the auth type
	if auth.GetAuthType() != AuthAPIKey {
		t.Errorf("Expected auth type AuthAPIKey, got %v", auth.GetAuthType())
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

	// Check that the API key was added as the Authorization header
	if req.Header.Get("Authorization") != "tyk-api-key-example" {
		t.Errorf("Expected Authorization header to be tyk-api-key-example, got %s", req.Header.Get("Authorization"))
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
