package parser

import (
	"testing"
)

func TestOpenAPIParser_ParseJSON(t *testing.T) {
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
							"description": "Successful operation",
							"content": {
								"application/json": {
									"schema": {
										"type": "array",
										"items": {
											"type": "object",
											"properties": {
												"id": {
													"type": "integer",
													"format": "int64"
												},
												"name": {
													"type": "string"
												},
												"email": {
													"type": "string",
													"format": "email"
												}
											}
										}
									}
								}
							}
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
			},
			"/users/{id}": {
				"get": {
					"summary": "Get user by ID",
					"description": "Returns a single user",
					"tags": ["users"],
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"description": "ID of the user to return",
							"required": true,
							"schema": {
								"type": "integer",
								"format": "int64"
							}
						}
					],
					"responses": {
						"200": {
							"description": "Successful operation",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"id": {
												"type": "integer",
												"format": "int64"
											},
											"name": {
												"type": "string"
											},
											"email": {
												"type": "string",
												"format": "email"
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)

	// Create a new parser
	parser := NewOpenAPIParser()

	// Parse the JSON data
	err := parser.ParseJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check basic information
	if parser.Spec.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", parser.Spec.Title)
	}
	if parser.Spec.Description != "API for testing OpenAPI parser" {
		t.Errorf("Expected description 'API for testing OpenAPI parser', got '%s'", parser.Spec.Description)
	}
	if parser.Spec.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", parser.Spec.Version)
	}
	if parser.Spec.BaseURL != "https://api.example.com/v1" {
		t.Errorf("Expected base URL 'https://api.example.com/v1', got '%s'", parser.Spec.BaseURL)
	}

	// Check OpenAPI version
	if parser.Version != OpenAPIV3 {
		t.Errorf("Expected OpenAPI version 3.0, got %s", parser.Version)
	}

	// Check endpoints
	endpoints := parser.GetEndpoints()
	if len(endpoints) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(endpoints))
	}

	// Check endpoint paths
	paths := parser.GetEndpointPaths()
	if len(paths) != 3 {
		t.Errorf("Expected 3 paths, got %d", len(paths))
	}
	if !contains(paths, "/users") || !contains(paths, "/users/{id}") {
		t.Errorf("Missing expected paths in %v", paths)
	}

	// Check endpoints by tag
	userEndpoints := parser.GetEndpointsByTag("users")
	if len(userEndpoints) != 3 {
		t.Errorf("Expected 3 user endpoints, got %d", len(userEndpoints))
	}

	// Check endpoints by method
	getEndpoints := parser.GetEndpointsByMethod("GET")
	if len(getEndpoints) != 2 {
		t.Errorf("Expected 2 GET endpoints, got %d", len(getEndpoints))
	}
	postEndpoints := parser.GetEndpointsByMethod("POST")
	if len(postEndpoints) != 1 {
		t.Errorf("Expected 1 POST endpoint, got %d", len(postEndpoints))
	}

	// Check auth required endpoints
	authEndpoints := parser.GetAuthRequiredEndpoints()
	if len(authEndpoints) != 1 {
		t.Errorf("Expected 1 auth required endpoint, got %d", len(authEndpoints))
	}
	if authEndpoints[0].Method != "POST" || authEndpoints[0].Path != "/users" {
		t.Errorf("Expected POST /users to require auth, got %s %s", authEndpoints[0].Method, authEndpoints[0].Path)
	}

	// Check tags
	tags := parser.GetTags()
	if len(tags) != 1 || tags[0] != "users" {
		t.Errorf("Expected tags [users], got %v", tags)
	}

	// Check wordlist generation
	wordlist := parser.GenerateWordlist()
	if len(wordlist) < 3 {
		t.Errorf("Expected at least 3 entries in wordlist, got %d", len(wordlist))
	}
	if !contains(wordlist, "/users") || !contains(wordlist, "/users/") {
		t.Errorf("Missing expected entries in wordlist %v", wordlist)
	}
}

func TestOpenAPIParser_ParseYAML(t *testing.T) {
	// Test data - a simple OpenAPI 3.0 specification in YAML-like format
	// Note: Since we're using a simplified YAML parser that treats YAML as JSON,
	// this test uses a YAML document that's also valid JSON
	yamlData := []byte(`{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API YAML",
			"description": "API for testing OpenAPI parser with YAML",
			"version": "1.0.0"
		},
		"paths": {
			"/products": {
				"get": {
					"summary": "Get all products",
					"tags": ["products"]
				}
			}
		}
	}`)

	// Create a new parser
	parser := NewOpenAPIParser()

	// Parse the YAML data
	err := parser.ParseYAML(yamlData)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Check basic information
	if parser.Spec.Title != "Test API YAML" {
		t.Errorf("Expected title 'Test API YAML', got '%s'", parser.Spec.Title)
	}

	// Check endpoints
	endpoints := parser.GetEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/products" || endpoints[0].Method != "GET" {
		t.Errorf("Expected GET /products, got %s %s", endpoints[0].Method, endpoints[0].Path)
	}
}

func TestOpenAPIParser_Swagger2(t *testing.T) {
	// Test data - a simple Swagger 2.0 specification
	jsonData := []byte(`{
		"swagger": "2.0",
		"info": {
			"title": "Swagger API",
			"description": "API for testing Swagger 2.0 parser",
			"version": "1.0.0"
		},
		"host": "api.example.com",
		"basePath": "/v2",
		"schemes": ["https"],
		"paths": {
			"/pets": {
				"get": {
					"summary": "Get all pets",
					"tags": ["pets"],
					"parameters": [
						{
							"name": "limit",
							"in": "query",
							"description": "Maximum number of pets to return",
							"required": false,
							"type": "integer",
							"format": "int32"
						}
					],
					"responses": {
						"200": {
							"description": "Successful operation"
						}
					}
				}
			}
		}
	}`)

	// Create a new parser
	parser := NewOpenAPIParser()

	// Parse the JSON data
	err := parser.ParseJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check OpenAPI version
	if parser.Version != OpenAPIV2 {
		t.Errorf("Expected OpenAPI version 2.0, got %s", parser.Version)
	}

	// Check base URL
	if parser.Spec.BaseURL != "https://api.example.com/v2" {
		t.Errorf("Expected base URL 'https://api.example.com/v2', got '%s'", parser.Spec.BaseURL)
	}

	// Check endpoints
	endpoints := parser.GetEndpoints()
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/pets" || endpoints[0].Method != "GET" {
		t.Errorf("Expected GET /pets, got %s %s", endpoints[0].Method, endpoints[0].Path)
	}
}

func TestOpenAPIParser_ExtractSchema(t *testing.T) {
	// Test data - a schema object
	schemaData := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":   "integer",
				"format": "int64",
			},
			"name": map[string]interface{}{
				"type": "string",
			},
			"tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []interface{}{"id", "name"},
		"example": map[string]interface{}{
			"id":   1,
			"name": "Example",
			"tags": []interface{}{"tag1", "tag2"},
		},
	}

	// Create a new parser
	parser := NewOpenAPIParser()

	// Extract the schema
	schema := parser.extractSchema(schemaData)

	// Check schema type
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// Check properties
	if len(schema.Properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(schema.Properties))
	}
	if schema.Properties["id"].Type != "integer" || schema.Properties["id"].Format != "int64" {
		t.Errorf("Expected id property to be integer with format int64, got %s with format %s", schema.Properties["id"].Type, schema.Properties["id"].Format)
	}
	if schema.Properties["name"].Type != "string" {
		t.Errorf("Expected name property to be string, got %s", schema.Properties["name"].Type)
	}
	if schema.Properties["tags"].Type != "array" || schema.Properties["tags"].Items.Type != "string" {
		t.Errorf("Expected tags property to be array of strings, got %s of %s", schema.Properties["tags"].Type, schema.Properties["tags"].Items.Type)
	}

	// Check required properties
	if len(schema.Required) != 2 || schema.Required[0] != "id" || schema.Required[1] != "name" {
		t.Errorf("Expected required properties [id, name], got %v", schema.Required)
	}

	// Check example
	if schema.Example == nil {
		t.Errorf("Expected example to be non-nil")
	}
}

func TestOpenAPIParser_GenerateWordlist(t *testing.T) {
	// Create a parser with some endpoints
	parser := NewOpenAPIParser()
	parser.Spec.Endpoints = []*OpenAPIEndpoint{
		{Path: "/api/users"},
		{Path: "/api/users/{id}"},
		{Path: "/api/products"},
		{Path: "/api/orders/"},
	}

	// Generate wordlist
	wordlist := parser.GenerateWordlist()

	// Check wordlist
	expectedEntries := []string{
		"/api/users",
		"/api/users/",
		"/api/users/{id}",
		"/api/products",
		"/api/products/",
		"/api/orders",
		"/api/orders/",
		"/api",
		"/api/",
	}
	for _, entry := range expectedEntries {
		if !contains(wordlist, entry) {
			t.Errorf("Expected wordlist to contain '%s', but it doesn't", entry)
		}
	}

	// Check that path components are included
	if !contains(wordlist, "/api") {
		t.Errorf("Expected wordlist to contain path component '/api', but it doesn't")
	}
}