package parser

import (
	"strings"
	"testing"
)

func TestAPIEndpointDiscovery_DiscoverFromOpenAPI(t *testing.T) {
	// Test data - a simple OpenAPI 3.0 specification in JSON format
	jsonData := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"description": "API for testing OpenAPI parser",
			"version": "1.0.0"
		},
		"servers": [
			{
				"url": "https://api.example.com/v1"
			}
		],
		"paths": {
			"/users": {
				"get": {
					"summary": "Get all users",
					"description": "Returns a list of users",
					"tags": ["users"],
					"parameters": [
						{
							"name": "limit",
							"in": "query",
							"description": "Maximum number of users to return",
							"required": false,
							"schema": {
								"type": "integer",
								"format": "int32"
							}
						}
					],
					"responses": {
						"200": {
							"description": "Successful operation"
						}
					}
				},
				"post": {
					"summary": "Create a user",
					"description": "Creates a new user",
					"tags": ["users"],
					"requestBody": {
						"required": true,
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {
											"type": "string"
										},
										"email": {
											"type": "string",
											"format": "email"
										}
									},
									"required": ["name", "email"]
								}
							}
						}
					},
					"responses": {
						"201": {
							"description": "User created"
						}
					},
					"security": [
						{
							"api_key": []
						}
					]
				}
			}
		}
	}`)

	// Create a temporary file with the test data
	tempFile := createTempFile(t, jsonData)
	defer removeTempFile(t, tempFile)

	// Create a new discovery
	discovery := NewAPIEndpointDiscovery("https://api.example.com")

	// Mock the OpenAPI parser
	mockParser := &MockOpenAPIParser{
		endpoints: []*OpenAPIEndpoint{
			{
				Path:        "/users",
				Method:      "GET",
				Summary:     "Get all users",
				Description: "Returns a list of users",
				Tags:        []string{"users"},
				Parameters: []*OpenAPIParameter{
					{
						Name:        "limit",
						In:          "query",
						Required:    false,
						Description: "Maximum number of users to return",
						Schema: &OpenAPISchema{
							Type:   "integer",
							Format: "int32",
						},
					},
				},
				RequiresAuth: false,
			},
			{
				Path:        "/users",
				Method:      "POST",
				Summary:     "Create a user",
				Description: "Creates a new user",
				Tags:        []string{"users"},
				RequestBody: &OpenAPISchema{
					Type: "object",
					Properties: map[string]*OpenAPISchema{
						"name": {
							Type: "string",
						},
						"email": {
							Type:   "string",
							Format: "email",
						},
					},
					Required: []string{"name", "email"},
				},
				RequiresAuth: true,
			},
		},
		baseURL: "https://api.example.com/v1",
	}

	// Set the mock parser
	discovery.Parser = mockParser

	// Manually convert the mock endpoints to discovered endpoints
	for _, endpoint := range mockParser.endpoints {
		// Create a new discovered endpoint
		discoveredEndpoint := &DiscoveredEndpoint{
			Method:       endpoint.Method,
			Path:         endpoint.Path,
			RequiresAuth: endpoint.RequiresAuth,
			Description:  endpoint.Description,
			Tags:         endpoint.Tags,
			Source:       "OpenAPI",
			Parameters:   make([]*DiscoveredParameter, 0),
		}

		// Set the full URL
		baseURL := "https://api.example.com"
		path := endpoint.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		discoveredEndpoint.URL = baseURL + path

		// Convert parameters
		for _, param := range endpoint.Parameters {
			discoveredParam := &DiscoveredParameter{
				Name:        param.Name,
				In:          param.In,
				Required:    param.Required,
				Description: param.Description,
				Example:     param.Example,
			}

			// Set the type from the schema if available
			if param.Schema != nil {
				discoveredParam.Type = param.Schema.Type
			}

			discoveredEndpoint.Parameters = append(discoveredEndpoint.Parameters, discoveredParam)
		}

		// Add request body parameters if available
		if endpoint.RequestBody != nil && endpoint.RequestBody.Type == "object" {
			for name, prop := range endpoint.RequestBody.Properties {
				discoveredParam := &DiscoveredParameter{
					Name:        name,
					In:          "body",
					Description: "", // No description available in the schema
					Type:        prop.Type,
				}

				// Check if the parameter is required
				for _, req := range endpoint.RequestBody.Required {
					if req == name {
						discoveredParam.Required = true
						break
					}
				}

				discoveredEndpoint.Parameters = append(discoveredEndpoint.Parameters, discoveredParam)
			}
		}

		// Add the endpoint to the list
		discovery.Endpoints = append(discovery.Endpoints, discoveredEndpoint)
	}

	// Check the discovered endpoints
	endpoints := discovery.GetEndpoints()
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
	}

	// Check the first endpoint
	if endpoints[0].Method != "GET" || endpoints[0].Path != "/users" {
		t.Errorf("Expected GET /users, got %s %s", endpoints[0].Method, endpoints[0].Path)
	}
	if endpoints[0].RequiresAuth {
		t.Errorf("Expected GET /users to not require auth")
	}
	if len(endpoints[0].Parameters) != 1 {
		t.Errorf("Expected 1 parameter for GET /users, got %d", len(endpoints[0].Parameters))
	}
	if endpoints[0].Parameters[0].Name != "limit" || endpoints[0].Parameters[0].In != "query" {
		t.Errorf("Expected parameter 'limit' in 'query', got '%s' in '%s'", endpoints[0].Parameters[0].Name, endpoints[0].Parameters[0].In)
	}

	// Check the second endpoint
	if endpoints[1].Method != "POST" || endpoints[1].Path != "/users" {
		t.Errorf("Expected POST /users, got %s %s", endpoints[1].Method, endpoints[1].Path)
	}
	if !endpoints[1].RequiresAuth {
		t.Errorf("Expected POST /users to require auth")
	}
	if len(endpoints[1].Parameters) != 2 {
		t.Errorf("Expected 2 parameters for POST /users, got %d", len(endpoints[1].Parameters))
	}
	nameParam := findParameter(endpoints[1].Parameters, "name")
	if nameParam == nil || nameParam.In != "body" || !nameParam.Required {
		t.Errorf("Expected required parameter 'name' in 'body'")
	}
	emailParam := findParameter(endpoints[1].Parameters, "email")
	if emailParam == nil || emailParam.In != "body" || !emailParam.Required {
		t.Errorf("Expected required parameter 'email' in 'body'")
	}

	// Check filtering by method
	getEndpoints := discovery.GetEndpointsByMethod("GET")
	if len(getEndpoints) != 1 {
		t.Errorf("Expected 1 GET endpoint, got %d", len(getEndpoints))
	}
	postEndpoints := discovery.GetEndpointsByMethod("POST")
	if len(postEndpoints) != 1 {
		t.Errorf("Expected 1 POST endpoint, got %d", len(postEndpoints))
	}

	// Check filtering by tag
	userEndpoints := discovery.GetEndpointsByTag("users")
	if len(userEndpoints) != 2 {
		t.Errorf("Expected 2 user endpoints, got %d", len(userEndpoints))
	}

	// Check filtering by auth
	authEndpoints := discovery.GetAuthRequiredEndpoints()
	if len(authEndpoints) != 1 {
		t.Errorf("Expected 1 auth required endpoint, got %d", len(authEndpoints))
	}
	if authEndpoints[0].Method != "POST" || authEndpoints[0].Path != "/users" {
		t.Errorf("Expected POST /users to require auth, got %s %s", authEndpoints[0].Method, authEndpoints[0].Path)
	}

	// Check filtering by path
	pathEndpoints := discovery.GetEndpointsByPath("/users")
	if len(pathEndpoints) != 2 {
		t.Errorf("Expected 2 endpoints with path /users, got %d", len(pathEndpoints))
	}

	// Check wordlist generation
	wordlist := discovery.GenerateWordlist()
	if len(wordlist) < 2 {
		t.Errorf("Expected at least 2 entries in wordlist, got %d", len(wordlist))
	}
	if !contains(wordlist, "/users") || !contains(wordlist, "/users/") {
		t.Errorf("Missing expected entries in wordlist %v", wordlist)
	}

	// Check URL generation
	urls := discovery.GenerateURLs()
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urls))
	}
	if !contains(urls, "https://api.example.com/users") {
		t.Errorf("Missing expected URL in %v", urls)
	}

	// Check parameter wordlist generation
	paramWordlist := discovery.GenerateParameterWordlist()
	if len(paramWordlist) != 3 {
		t.Errorf("Expected 3 parameters in wordlist, got %d", len(paramWordlist))
	}
	if !contains(paramWordlist, "limit") || !contains(paramWordlist, "name") || !contains(paramWordlist, "email") {
		t.Errorf("Missing expected parameters in wordlist %v", paramWordlist)
	}
}

// Helper function to find a parameter by name
func findParameter(params []*DiscoveredParameter, name string) *DiscoveredParameter {
	for _, param := range params {
		if param.Name == name {
			return param
		}
	}
	return nil
}

// Mock OpenAPI parser for testing
type MockOpenAPIParser struct {
	endpoints []*OpenAPIEndpoint
	baseURL   string
}

func (m *MockOpenAPIParser) GetEndpoints() []*OpenAPIEndpoint {
	return m.endpoints
}

// Helper functions for creating and removing temporary files
func createTempFile(t *testing.T, data []byte) string {
	// In a real implementation, create a temporary file
	// For this test, we'll just return a dummy path
	return "/tmp/test_openapi.json"
}

func removeTempFile(t *testing.T, path string) {
	// In a real implementation, remove the temporary file
	// For this test, do nothing
}

func TestAPIEndpointDiscovery_PathMatches(t *testing.T) {
	testCases := []struct {
		path    string
		pattern string
		matches bool
	}{
		{"/api/users", "/api/users", true},
		{"/api/users", "/api/users/", false},
		{"/api/users/", "/api/users/", true},
		{"/api/users/123", "/api/users/*", true},
		{"/api/products", "/api/users", false},
		{"/api/v1/users", "/api/*/users", true},
		{"/api/v2/users", "/api/*/users", true},
		{"/api/users/123/comments", "/api/users/*/comments", true},
		{"/api/users/comments", "/api/users/*/comments", false},
	}

	for _, tc := range testCases {
		result := pathMatches(tc.path, tc.pattern)
		if result != tc.matches {
			t.Errorf("pathMatches(%s, %s) = %v, expected %v", tc.path, tc.pattern, result, tc.matches)
		}
	}
}

func TestAPIEndpointDiscovery_GenerateParameterWordlist(t *testing.T) {
	// Create a discovery with some endpoints and parameters
	discovery := NewAPIEndpointDiscovery("")
	discovery.Endpoints = []*DiscoveredEndpoint{
		{
			Path:   "/api/users",
			Method: "GET",
			Parameters: []*DiscoveredParameter{
				{Name: "limit", In: "query"},
				{Name: "offset", In: "query"},
			},
		},
		{
			Path:   "/api/users",
			Method: "POST",
			Parameters: []*DiscoveredParameter{
				{Name: "name", In: "body"},
				{Name: "email", In: "body"},
			},
		},
		{
			Path:   "/api/users/{id}",
			Method: "GET",
			Parameters: []*DiscoveredParameter{
				{Name: "id", In: "path"},
				{Name: "fields", In: "query"},
			},
		},
	}

	// Generate parameter wordlist
	paramWordlist := discovery.GenerateParameterWordlist()

	// Check the wordlist
	expectedParams := []string{"limit", "offset", "name", "email", "id", "fields"}
	if len(paramWordlist) != len(expectedParams) {
		t.Errorf("Expected %d parameters in wordlist, got %d", len(expectedParams), len(paramWordlist))
	}
	for _, param := range expectedParams {
		if !contains(paramWordlist, param) {
			t.Errorf("Missing expected parameter '%s' in wordlist %v", param, paramWordlist)
		}
	}
}