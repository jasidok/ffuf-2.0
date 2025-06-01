package wordlist

import (
	"testing"
)

func TestCategorizeEndpoint(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		expectedTypes  []CategoryType
		expectedNames  []string
		unexpectedType CategoryType
	}{
		{
			name:          "REST API endpoint",
			path:          "/api/v1/users",
			expectedTypes: []CategoryType{APITypeCategory, FunctionalCategory},
			expectedNames: []string{"rest", "user"},
		},
		{
			name:          "GraphQL endpoint",
			path:          "/graphql",
			expectedTypes: []CategoryType{APITypeCategory},
			expectedNames: []string{"graphql"},
		},
		{
			name:          "Admin endpoint",
			path:          "/api/admin/settings",
			expectedTypes: []CategoryType{APITypeCategory, FunctionalCategory, SecurityCategory},
			expectedNames: []string{"rest", "admin", "sensitive", "privileged"},
		},
		{
			name:          "Payment endpoint",
			path:          "/api/v2/payment/process",
			expectedTypes: []CategoryType{APITypeCategory, FunctionalCategory, SecurityCategory},
			expectedNames: []string{"rest", "payment", "payment_data"},
		},
		{
			name:          "Mobile API endpoint",
			path:          "/mobile/api/v1/profile",
			expectedTypes: []CategoryType{APITypeCategory, FunctionalCategory},
			expectedNames: []string{"mobile", "user"},
		},
		{
			name:          "Internal API endpoint",
			path:          "/internal/api/metrics",
			expectedTypes: []CategoryType{APITypeCategory, FunctionalCategory, SecurityCategory},
			expectedNames: []string{"internal", "analytics", "sensitive"},
		},
		{
			name:          "AWS-related endpoint",
			path:          "/api/aws/s3/upload",
			expectedTypes: []CategoryType{APITypeCategory, TechnologyCategory, FunctionalCategory},
			expectedNames: []string{"rest", "aws", "file"},
		},
		{
			name:           "Non-API endpoint",
			path:           "/static/images/logo.png",
			unexpectedType: APITypeCategory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			categories := CategorizeEndpoint(tc.path)
			
			// Check that we have at least one category
			if len(categories) == 0 && len(tc.expectedTypes) > 0 {
				t.Errorf("Expected at least one category, got none")
			}
			
			// Check for expected category types and names
			for i, expectedType := range tc.expectedTypes {
				expectedName := tc.expectedNames[i]
				expectedCategory := string(expectedType) + ":" + expectedName
				
				found := false
				for _, category := range categories {
					if category == expectedCategory {
						found = true
						break
					}
				}
				
				if !found {
					t.Errorf("Expected category %s not found in %v", expectedCategory, categories)
				}
			}
			
			// Check that unexpected type is not present
			if tc.unexpectedType != "" {
				for _, category := range categories {
					if string(tc.unexpectedType)+":" == category[:len(string(tc.unexpectedType))+1] {
						t.Errorf("Unexpected category type %s found in %v", tc.unexpectedType, categories)
					}
				}
			}
		})
	}
}

func TestCategorizeEndpoints(t *testing.T) {
	paths := []string{
		"/api/v1/users",
		"/graphql",
		"/api/admin/settings",
	}
	
	result := CategorizeEndpoints(paths)
	
	// Check that all paths are in the result
	for _, path := range paths {
		if _, exists := result[path]; !exists {
			t.Errorf("Path %s not found in result", path)
		}
	}
	
	// Check that the first path has the expected categories
	categories := result["/api/v1/users"]
	restCategory := "api_type:rest"
	userCategory := "functional:user"
	
	foundRest := false
	foundUser := false
	for _, category := range categories {
		if category == restCategory {
			foundRest = true
		}
		if category == userCategory {
			foundUser = true
		}
	}
	
	if !foundRest {
		t.Errorf("Expected category %s not found in %v", restCategory, categories)
	}
	if !foundUser {
		t.Errorf("Expected category %s not found in %v", userCategory, categories)
	}
}

func TestGetCategoryDefinition(t *testing.T) {
	// Test getting an existing category
	category, exists := GetCategoryDefinition(APITypeCategory, "rest")
	if !exists {
		t.Error("Expected REST category to exist")
	}
	if category.Name != "rest" {
		t.Errorf("Expected name 'rest', got '%s'", category.Name)
	}
	if category.Type != APITypeCategory {
		t.Errorf("Expected type APITypeCategory, got %s", category.Type)
	}
	
	// Test getting a non-existent category
	_, exists = GetCategoryDefinition(APITypeCategory, "nonexistent")
	if exists {
		t.Error("Expected nonexistent category to not exist")
	}
}

func TestGetCategoryDefinitionsByType(t *testing.T) {
	// Test getting API type categories
	categories := GetCategoryDefinitionsByType(APITypeCategory)
	if len(categories) == 0 {
		t.Error("Expected at least one API type category")
	}
	
	// Check that all categories have the correct type
	for _, category := range categories {
		if category.Type != APITypeCategory {
			t.Errorf("Expected type APITypeCategory, got %s", category.Type)
		}
	}
	
	// Test getting functional categories
	categories = GetCategoryDefinitionsByType(FunctionalCategory)
	if len(categories) == 0 {
		t.Error("Expected at least one functional category")
	}
	
	// Check that all categories have the correct type
	for _, category := range categories {
		if category.Type != FunctionalCategory {
			t.Errorf("Expected type FunctionalCategory, got %s", category.Type)
		}
	}
}

func TestEnhanceAPIWordlist(t *testing.T) {
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
		},
		categories: make(map[string][]int),
	}
	
	// Initialize the categories index
	wl.categories["test"] = []int{0, 1}
	
	// Enhance the wordlist
	EnhanceAPIWordlist(wl)
	
	// Check that the first endpoint has the REST and user categories
	endpoint := wl.entries[0]
	restCategory := "api_type:rest"
	userCategory := "functional:user"
	
	foundRest := false
	foundUser := false
	for _, category := range endpoint.Categories {
		if category == restCategory {
			foundRest = true
		}
		if category == userCategory {
			foundUser = true
		}
	}
	
	if !foundRest {
		t.Errorf("Expected category %s not found in %v", restCategory, endpoint.Categories)
	}
	if !foundUser {
		t.Errorf("Expected category %s not found in %v", userCategory, endpoint.Categories)
	}
	
	// Check that the second endpoint has the GraphQL category
	endpoint = wl.entries[1]
	graphqlCategory := "api_type:graphql"
	
	foundGraphQL := false
	for _, category := range endpoint.Categories {
		if category == graphqlCategory {
			foundGraphQL = true
		}
	}
	
	if !foundGraphQL {
		t.Errorf("Expected category %s not found in %v", graphqlCategory, endpoint.Categories)
	}
	
	// Check that the categories index has been updated
	if _, exists := wl.categories[restCategory]; !exists {
		t.Errorf("Category %s not found in categories index", restCategory)
	}
	if _, exists := wl.categories[userCategory]; !exists {
		t.Errorf("Category %s not found in categories index", userCategory)
	}
	if _, exists := wl.categories[graphqlCategory]; !exists {
		t.Errorf("Category %s not found in categories index", graphqlCategory)
	}
}

func TestGetEndpointsByCategory(t *testing.T) {
	// Create a simple APIWordlist
	wl := &APIWordlist{
		entries: []APIEndpoint{
			{
				Path:       "/api/v1/users",
				Method:     "GET",
				Categories: []string{"api_type:rest", "functional:user"},
			},
			{
				Path:       "/graphql",
				Method:     "POST",
				Categories: []string{"api_type:graphql"},
			},
			{
				Path:       "/api/v1/admin",
				Method:     "GET",
				Categories: []string{"api_type:rest", "functional:admin", "security:privileged"},
			},
		},
		categories: make(map[string][]int),
	}
	
	// Initialize the categories index
	wl.categories["api_type:rest"] = []int{0, 2}
	wl.categories["functional:user"] = []int{0}
	wl.categories["api_type:graphql"] = []int{1}
	wl.categories["functional:admin"] = []int{2}
	wl.categories["security:privileged"] = []int{2}
	
	// Test getting REST endpoints
	endpoints := wl.GetEndpointsByCategory(APITypeCategory, "rest")
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 REST endpoints, got %d", len(endpoints))
	}
	
	// Test getting user endpoints
	endpoints = wl.GetEndpointsByCategory(FunctionalCategory, "user")
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 user endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/api/v1/users" {
		t.Errorf("Expected path '/api/v1/users', got '%s'", endpoints[0].Path)
	}
	
	// Test getting non-existent category
	endpoints = wl.GetEndpointsByCategory(FunctionalCategory, "nonexistent")
	if len(endpoints) != 0 {
		t.Errorf("Expected 0 endpoints for nonexistent category, got %d", len(endpoints))
	}
}

func TestGetEndpointsByCategoryType(t *testing.T) {
	// Create a simple APIWordlist
	wl := &APIWordlist{
		entries: []APIEndpoint{
			{
				Path:       "/api/v1/users",
				Method:     "GET",
				Categories: []string{"api_type:rest", "functional:user"},
			},
			{
				Path:       "/graphql",
				Method:     "POST",
				Categories: []string{"api_type:graphql"},
			},
			{
				Path:       "/api/v1/admin",
				Method:     "GET",
				Categories: []string{"api_type:rest", "functional:admin", "security:privileged"},
			},
		},
		categories: make(map[string][]int),
	}
	
	// Initialize the categories index
	wl.categories["api_type:rest"] = []int{0, 2}
	wl.categories["functional:user"] = []int{0}
	wl.categories["api_type:graphql"] = []int{1}
	wl.categories["functional:admin"] = []int{2}
	wl.categories["security:privileged"] = []int{2}
	
	// Test getting all API type endpoints
	endpoints := wl.GetEndpointsByCategoryType(APITypeCategory)
	if len(endpoints) != 3 {
		t.Errorf("Expected 3 API type endpoints, got %d", len(endpoints))
	}
	
	// Test getting all functional endpoints
	endpoints = wl.GetEndpointsByCategoryType(FunctionalCategory)
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 functional endpoints, got %d", len(endpoints))
	}
	
	// Test getting all security endpoints
	endpoints = wl.GetEndpointsByCategoryType(SecurityCategory)
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 security endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/api/v1/admin" {
		t.Errorf("Expected path '/api/v1/admin', got '%s'", endpoints[0].Path)
	}
}

func TestGetCategoryTypes(t *testing.T) {
	types := GetCategoryTypes()
	
	// Check that all expected types are present
	expectedTypes := []CategoryType{
		APITypeCategory,
		FunctionalCategory,
		SecurityCategory,
		ArchitecturalCategory,
		TechnologyCategory,
	}
	
	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d category types, got %d", len(expectedTypes), len(types))
	}
	
	for _, expectedType := range expectedTypes {
		found := false
		for _, actualType := range types {
			if actualType == expectedType {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Expected category type %s not found", expectedType)
		}
	}
}

func TestGetCategoriesByType(t *testing.T) {
	// Test getting API type categories
	categories := GetCategoriesByType(APITypeCategory)
	if len(categories) == 0 {
		t.Error("Expected at least one API type category")
	}
	
	// Check that "rest" is in the API type categories
	foundRest := false
	for _, category := range categories {
		if category == "rest" {
			foundRest = true
			break
		}
	}
	
	if !foundRest {
		t.Errorf("Expected 'rest' in API type categories, got %v", categories)
	}
	
	// Test getting functional categories
	categories = GetCategoriesByType(FunctionalCategory)
	if len(categories) == 0 {
		t.Error("Expected at least one functional category")
	}
	
	// Check that "user" is in the functional categories
	foundUser := false
	for _, category := range categories {
		if category == "user" {
			foundUser = true
			break
		}
	}
	
	if !foundUser {
		t.Errorf("Expected 'user' in functional categories, got %v", categories)
	}
}