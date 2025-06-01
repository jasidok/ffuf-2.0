package wordlist

import (
	"regexp"
	"testing"
)

func TestMatchSignatures(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expectedCount  int
		expectedSigs   []string
		unexpectedSigs []string
	}{
		{
			name:          "Standard REST API",
			path:          "/api/v1/users",
			expectedCount: 1,
			expectedSigs:  []string{"standard_rest"},
		},
		{
			name:          "REST API with resource ID",
			path:          "/api/users/123",
			expectedCount: 1,
			expectedSigs:  []string{"resource_rest"},
		},
		{
			name:          "Nested REST API",
			path:          "/api/users/123/posts",
			expectedCount: 1,
			expectedSigs:  []string{"nested_rest"},
		},
		{
			name:          "GraphQL endpoint",
			path:          "/graphql",
			expectedCount: 1,
			expectedSigs:  []string{"graphql_endpoint"},
		},
		{
			name:          "Mobile API endpoint",
			path:          "/mobile/api/v1/profile",
			expectedCount: 1,
			expectedSigs:  []string{"mobile_api"},
		},
		{
			name:          "OAuth token endpoint",
			path:          "/oauth/token",
			expectedCount: 1,
			expectedSigs:  []string{"oauth_token"},
		},
		{
			name:          "Create operation endpoint",
			path:          "/api/v1/create/user",
			expectedCount: 1,
			expectedSigs:  []string{"create_operation"},
		},
		{
			name:           "Non-API endpoint",
			path:           "/static/images/logo.png",
			expectedCount:  0,
			unexpectedSigs: []string{"standard_rest", "graphql_endpoint"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := MatchSignatures(tc.path)

			// Check the number of matches
			if len(matches) != tc.expectedCount {
				t.Errorf("Expected %d matches, got %d", tc.expectedCount, len(matches))
			}

			// Check for expected signatures
			for _, expectedSig := range tc.expectedSigs {
				found := false
				for _, match := range matches {
					if match.Signature.Name == expectedSig {
						found = true
						break
					}
				}

				if !found && tc.expectedCount > 0 {
					t.Errorf("Expected signature %s not found in matches", expectedSig)
				}
			}

			// Check that unexpected signatures are not present
			for _, unexpectedSig := range tc.unexpectedSigs {
				for _, match := range matches {
					if match.Signature.Name == unexpectedSig {
						t.Errorf("Unexpected signature %s found in matches", unexpectedSig)
					}
				}
			}
		})
	}
}

func TestMatchSignaturesWithThreshold(t *testing.T) {
	// Test with high threshold
	matches := MatchSignaturesWithThreshold("/graphql", 95)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match with threshold 95, got %d", len(matches))
	}

	// Test with threshold that excludes some matches
	matches = MatchSignaturesWithThreshold("/api/v1/users", 90)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match with threshold 90, got %d", len(matches))
	}

	// Test with threshold that includes all matches
	matches = MatchSignaturesWithThreshold("/api/v1/users", 80)
	if len(matches) != 1 {
		t.Errorf("Expected 1 match with threshold 80, got %d", len(matches))
	}

	// Test with threshold that excludes all matches
	matches = MatchSignaturesWithThreshold("/api/v1/users", 100)
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches with threshold 100, got %d", len(matches))
	}
}

func TestGetSignaturesByCategory(t *testing.T) {
	// Test getting REST API signatures
	signatures := GetSignaturesByCategory("api_type:rest")
	if len(signatures) != 3 {
		t.Errorf("Expected 3 REST API signatures, got %d", len(signatures))
	}

	// Test getting GraphQL API signatures
	signatures = GetSignaturesByCategory("api_type:graphql")
	if len(signatures) != 3 {
		t.Errorf("Expected 3 GraphQL API signatures, got %d", len(signatures))
	}

	// Test getting auth signatures
	signatures = GetSignaturesByCategory("functional:auth")
	if len(signatures) != 4 {
		t.Errorf("Expected 4 auth signatures, got %d", len(signatures))
	}

	// Test getting non-existent category
	signatures = GetSignaturesByCategory("nonexistent")
	if len(signatures) != 0 {
		t.Errorf("Expected 0 signatures for nonexistent category, got %d", len(signatures))
	}
}

func TestEnhanceAPIWordlistWithSignatures(t *testing.T) {
	// Create a simple APIWordlist
	wl := &APIWordlist{
		entries: []APIEndpoint{
			{
				Path:       "/api/v1/users",
				Method:     "GET",
				Categories: []string{"test"},
			},
			{
				Path:       "/graphql",
				Method:     "POST",
				Categories: []string{"test"},
			},
			{
				Path:       "/oauth/token",
				Method:     "POST",
				Categories: []string{"test"},
			},
		},
		categories: make(map[string][]int),
	}

	// Initialize the categories index
	wl.categories["test"] = []int{0, 1, 2}

	// Enhance the wordlist with signatures
	EnhanceAPIWordlistWithSignatures(wl, 80)

	// Check that the first endpoint has the REST API category
	endpoint := wl.entries[0]
	restCategory := "api_type:rest"

	foundRest := false
	for _, category := range endpoint.Categories {
		if category == restCategory {
			foundRest = true
			break
		}
	}

	if !foundRest {
		t.Errorf("Expected category %s not found in %v", restCategory, endpoint.Categories)
	}

	// Check that the second endpoint has the GraphQL category
	endpoint = wl.entries[1]
	graphqlCategory := "api_type:graphql"

	foundGraphQL := false
	for _, category := range endpoint.Categories {
		if category == graphqlCategory {
			foundGraphQL = true
			break
		}
	}

	if !foundGraphQL {
		t.Errorf("Expected category %s not found in %v", graphqlCategory, endpoint.Categories)
	}

	// Check that the third endpoint has the auth category
	endpoint = wl.entries[2]
	authCategory := "functional:auth"

	foundAuth := false
	for _, category := range endpoint.Categories {
		if category == authCategory {
			foundAuth = true
			break
		}
	}

	if !foundAuth {
		t.Errorf("Expected category %s not found in %v", authCategory, endpoint.Categories)
	}

	// Check that the categories index has been updated
	if _, exists := wl.categories[restCategory]; !exists {
		t.Errorf("Category %s not found in categories index", restCategory)
	}
	if _, exists := wl.categories[graphqlCategory]; !exists {
		t.Errorf("Category %s not found in categories index", graphqlCategory)
	}
	if _, exists := wl.categories[authCategory]; !exists {
		t.Errorf("Category %s not found in categories index", authCategory)
	}
}

func TestDetectAPIType(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedType string
	}{
		{
			name:         "REST API",
			path:         "/api/v1/users",
			expectedType: "rest",
		},
		{
			name:         "GraphQL API",
			path:         "/graphql",
			expectedType: "graphql",
		},
		{
			name:         "Mobile API",
			path:         "/mobile/api/v1/profile",
			expectedType: "mobile",
		},
		{
			name:         "Default to REST",
			path:         "/v1/users",
			expectedType: "rest",
		},
		{
			name:         "Unknown API type",
			path:         "/static/images/logo.png",
			expectedType: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiType := DetectAPIType(tc.path)
			if apiType != tc.expectedType {
				t.Errorf("Expected API type %s, got %s", tc.expectedType, apiType)
			}
		})
	}
}

func TestDetectEndpointFunction(t *testing.T) {
	testCases := []struct {
		name             string
		path             string
		expectedFunction string
	}{
		{
			name:             "Auth endpoint",
			path:             "/oauth/token",
			expectedFunction: "auth",
		},
		{
			name:             "CRUD endpoint - create",
			path:             "/api/v1/create/user",
			expectedFunction: "crud",
		},
		{
			name:             "User endpoint",
			path:             "/api/v1/users",
			expectedFunction: "user",
		},
		{
			name:             "Admin endpoint",
			path:             "/api/v1/admin/dashboard",
			expectedFunction: "admin",
		},
		{
			name:             "File endpoint",
			path:             "/api/v1/files/upload",
			expectedFunction: "file",
		},
		{
			name:             "Payment endpoint",
			path:             "/api/v1/payment/process",
			expectedFunction: "payment",
		},
		{
			name:             "Config endpoint",
			path:             "/api/v1/settings",
			expectedFunction: "config",
		},
		{
			name:             "Unknown function",
			path:             "/api/v1/misc",
			expectedFunction: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			function := DetectEndpointFunction(tc.path)
			if function != tc.expectedFunction {
				t.Errorf("Expected function %s, got %s", tc.expectedFunction, function)
			}
		})
	}
}

func TestGenerateSignatureReport(t *testing.T) {
	paths := []string{
		"/api/v1/users",
		"/graphql",
		"/oauth/token",
		"/static/images/logo.png",
	}

	// Generate report with default confidence
	report := GenerateSignatureReport(paths, 0)

	// Check that the report contains the expected paths
	if len(report) != 3 {
		t.Errorf("Expected 3 paths in report, got %d", len(report))
	}

	// Check that each path has the expected number of matches
	if matches, exists := report["/api/v1/users"]; !exists || len(matches) != 1 {
		t.Errorf("Expected 1 match for /api/v1/users, got %d", len(matches))
	}

	if matches, exists := report["/graphql"]; !exists || len(matches) != 1 {
		t.Errorf("Expected 1 match for /graphql, got %d", len(matches))
	}

	if matches, exists := report["/oauth/token"]; !exists || len(matches) != 1 {
		t.Errorf("Expected 1 match for /oauth/token, got %d", len(matches))
	}

	// Check that the non-API path is not in the report
	if _, exists := report["/static/images/logo.png"]; exists {
		t.Error("Expected /static/images/logo.png to not be in the report")
	}

	// Generate report with high confidence threshold
	report = GenerateSignatureReport(paths, 95)

	// Check that only the high-confidence matches are in the report
	if len(report) != 2 {
		t.Errorf("Expected 2 paths in report with high confidence, got %d", len(report))
	}

	if _, exists := report["/graphql"]; !exists {
		t.Error("Expected /graphql to be in the high-confidence report")
	}

	if _, exists := report["/oauth/token"]; !exists {
		t.Error("Expected /oauth/token to be in the high-confidence report")
	}
}

func TestAddCustomSignature(t *testing.T) {
	// Create a custom signature
	customSig := PatternSignature{
		Name:        "custom_sig",
		Pattern:     regexp.MustCompile(`(?i)/custom/`),
		Description: "Custom signature for testing",
		Category:    "test:custom",
		Confidence:  80,
	}

	// Get the initial count of signatures
	initialCount := len(AllSignatures)

	// Add the custom signature
	AddCustomSignature(customSig)

	// Check that the signature was added
	if len(AllSignatures) != initialCount+1 {
		t.Errorf("Expected %d signatures after adding custom signature, got %d", initialCount+1, len(AllSignatures))
	}

	// Check that the custom signature works
	matches := MatchSignatures("/custom/endpoint")

	foundCustom := false
	for _, match := range matches {
		if match.Signature.Name == "custom_sig" {
			foundCustom = true
			break
		}
	}

	if !foundCustom {
		t.Error("Custom signature not found in matches")
	}
}

func TestCreateCustomSignature(t *testing.T) {
	// Create a custom signature
	sig, err := CreateCustomSignature(
		"test_sig",
		"/test/",
		"Test signature",
		"test:category",
		85,
	)

	// Check that there was no error
	if err != nil {
		t.Errorf("Error creating custom signature: %v", err)
	}

	// Check that the signature has the expected values
	if sig.Name != "test_sig" {
		t.Errorf("Expected name 'test_sig', got '%s'", sig.Name)
	}

	if sig.Description != "Test signature" {
		t.Errorf("Expected description 'Test signature', got '%s'", sig.Description)
	}

	if sig.Category != "test:category" {
		t.Errorf("Expected category 'test:category', got '%s'", sig.Category)
	}

	if sig.Confidence != 85 {
		t.Errorf("Expected confidence 85, got %d", sig.Confidence)
	}

	// Test that the pattern works
	if !sig.Pattern.MatchString("/test/endpoint") {
		t.Error("Pattern does not match expected string")
	}

	// Test with invalid pattern
	_, err = CreateCustomSignature(
		"invalid_sig",
		"[invalid",
		"Invalid signature",
		"test:invalid",
		80,
	)

	// Check that there was an error
	if err == nil {
		t.Error("Expected error for invalid pattern, got nil")
	}
}
