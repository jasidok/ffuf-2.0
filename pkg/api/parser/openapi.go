// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

import (
	"github.com/ffuf/ffuf/v2/pkg/api"
)

// OpenAPIVersion represents the version of an OpenAPI specification
type OpenAPIVersion string

const (
	// OpenAPIV2 represents OpenAPI/Swagger 2.0
	OpenAPIV2 OpenAPIVersion = "2.0"
	// OpenAPIV3 represents OpenAPI 3.0.x
	OpenAPIV3 OpenAPIVersion = "3.0"
	// OpenAPIV31 represents OpenAPI 3.1.x
	OpenAPIV31 OpenAPIVersion = "3.1"
)

// OpenAPIParser provides methods for parsing OpenAPI/Swagger specifications
type OpenAPIParser struct {
	// The parsed specification
	Spec *OpenAPISpec
	// The version of the specification
	Version OpenAPIVersion
}

// OpenAPISpec represents a parsed OpenAPI/Swagger specification
type OpenAPISpec struct {
	// Raw data of the specification
	Raw map[string]interface{}
	// Extracted endpoints from the specification
	Endpoints []*OpenAPIEndpoint
	// Base URL for the API
	BaseURL string
	// Title of the API
	Title string
	// Description of the API
	Description string
	// Version of the API (not the OpenAPI version)
	Version string
}

// OpenAPIEndpoint represents an API endpoint extracted from an OpenAPI specification
type OpenAPIEndpoint struct {
	// Path of the endpoint
	Path string
	// HTTP method (GET, POST, etc.)
	Method string
	// Summary of the endpoint
	Summary string
	// Description of the endpoint
	Description string
	// Parameters for the endpoint
	Parameters []*OpenAPIParameter
	// Request body schema
	RequestBody *OpenAPISchema
	// Response schemas
	Responses map[string]*OpenAPISchema
	// Tags associated with the endpoint
	Tags []string
	// Whether the endpoint requires authentication
	RequiresAuth bool
}

// OpenAPIParameter represents a parameter for an API endpoint
type OpenAPIParameter struct {
	// Name of the parameter
	Name string
	// Location of the parameter (path, query, header, cookie)
	In string
	// Whether the parameter is required
	Required bool
	// Schema of the parameter
	Schema *OpenAPISchema
	// Description of the parameter
	Description string
	// Example value for the parameter
	Example interface{}
}

// OpenAPISchema represents a schema for a parameter, request body, or response
type OpenAPISchema struct {
	// Type of the schema (string, number, object, array, etc.)
	Type string
	// Format of the schema (date-time, email, etc.)
	Format string
	// Properties of the schema (for object types)
	Properties map[string]*OpenAPISchema
	// Items in the schema (for array types)
	Items *OpenAPISchema
	// Whether the schema is required
	Required []string
	// Enum values for the schema
	Enum []interface{}
	// Example value for the schema
	Example interface{}
}

// NewOpenAPIParser creates a new OpenAPIParser
func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{
		Spec: &OpenAPISpec{
			Endpoints: make([]*OpenAPIEndpoint, 0),
			Raw:       make(map[string]interface{}),
		},
	}
}

// ParseFromFile parses an OpenAPI/Swagger specification from a file
func (p *OpenAPIParser) ParseFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to read OpenAPI file: %s", err.Error()), 0)
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".json" {
		return p.ParseJSON(data)
	} else if ext == ".yaml" || ext == ".yml" {
		return p.ParseYAML(data)
	}

	// Try to parse as JSON first, then YAML if that fails
	if err := p.ParseJSON(data); err != nil {
		return p.ParseYAML(data)
	}

	return nil
}

// ParseFromURL parses an OpenAPI/Swagger specification from a URL
func (p *OpenAPIParser) ParseFromURL(specURL string) error {
	// Parse the URL
	parsedURL, err := url.Parse(specURL)
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Invalid URL: %s", err.Error()), 0)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Get(specURL)
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to fetch OpenAPI spec: %s", err.Error()), 0)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return api.NewAPIError(fmt.Sprintf("HTTP error %d: %s", resp.StatusCode, resp.Status), 0)
	}

	// Read the response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to read response body: %s", err.Error()), 0)
	}

	// Determine format based on Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := p.ParseJSON(data); err != nil {
			return api.NewAPIError(fmt.Sprintf("Failed to parse JSON spec: %s", err.Error()), 0)
		}
	} else if strings.Contains(contentType, "application/yaml") || strings.Contains(contentType, "text/yaml") {
		if err := p.ParseYAML(data); err != nil {
			return api.NewAPIError(fmt.Sprintf("Failed to parse YAML spec: %s", err.Error()), 0)
		}
	} else {
		// If Content-Type is not set or not recognized, try to parse as JSON first, then YAML
		jsonErr := p.ParseJSON(data)
		if jsonErr != nil {
			yamlErr := p.ParseYAML(data)
			if yamlErr != nil {
				return api.NewAPIError(fmt.Sprintf("Failed to parse spec as JSON (%s) or YAML (%s)", jsonErr.Error(), yamlErr.Error()), 0)
			}
		}
	}

	// Set the base URL from the request URL
	p.Spec.BaseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	return nil
}

// ParseJSON parses an OpenAPI/Swagger specification from JSON data
func (p *OpenAPIParser) ParseJSON(data []byte) error {
	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to parse JSON: %s", err.Error()), 0)
	}

	return p.parseSpec(spec)
}

// ParseYAML parses an OpenAPI/Swagger specification from YAML data
// Note: This is a simplified implementation that treats YAML as JSON
// For proper YAML parsing, a YAML library would be required
func (p *OpenAPIParser) ParseYAML(data []byte) error {
	// This is a simplified approach - in a real implementation, use a proper YAML parser
	// For now, we'll try to parse it as JSON and hope for the best
	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to parse YAML as JSON: %s", err.Error()), 0)
	}

	return p.parseSpec(spec)
}

// parseSpec parses the OpenAPI/Swagger specification
func (p *OpenAPIParser) parseSpec(spec map[string]interface{}) error {
	// Store the raw specification
	p.Spec.Raw = spec

	// Determine the OpenAPI version
	if _, ok := spec["swagger"].(string); ok {
		// Swagger 2.0
		p.Version = OpenAPIV2
	} else if openapi, ok := spec["openapi"].(string); ok {
		// OpenAPI 3.0.x or 3.1.x
		if strings.HasPrefix(openapi, "3.0") {
			p.Version = OpenAPIV3
		} else if strings.HasPrefix(openapi, "3.1") {
			p.Version = OpenAPIV31
		}
	} else {
		return api.NewAPIError("Invalid OpenAPI specification: missing version", 0)
	}

	// Extract basic information
	if info, ok := spec["info"].(map[string]interface{}); ok {
		if title, ok := info["title"].(string); ok {
			p.Spec.Title = title
		}
		if description, ok := info["description"].(string); ok {
			p.Spec.Description = description
		}
		if version, ok := info["version"].(string); ok {
			p.Spec.Version = version
		}
	}

	// Extract base URL
	if p.Version == OpenAPIV2 {
		// Swagger 2.0 uses host, basePath, and schemes
		if host, ok := spec["host"].(string); ok {
			scheme := "https"
			if schemes, ok := spec["schemes"].([]interface{}); ok && len(schemes) > 0 {
				if s, ok := schemes[0].(string); ok {
					scheme = s
				}
			}
			p.Spec.BaseURL = fmt.Sprintf("%s://%s", scheme, host)
			if basePath, ok := spec["basePath"].(string); ok {
				p.Spec.BaseURL = p.Spec.BaseURL + basePath
			}
		}
	} else {
		// OpenAPI 3.0.x and 3.1.x use servers
		if servers, ok := spec["servers"].([]interface{}); ok && len(servers) > 0 {
			if server, ok := servers[0].(map[string]interface{}); ok {
				if url, ok := server["url"].(string); ok {
					p.Spec.BaseURL = url
				}
			}
		}
	}

	// Extract endpoints
	if p.Version == OpenAPIV2 {
		// Swagger 2.0 uses paths
		if paths, ok := spec["paths"].(map[string]interface{}); ok {
			for path, pathItem := range paths {
				if methods, ok := pathItem.(map[string]interface{}); ok {
					for method, operation := range methods {
						// Skip non-HTTP method keys
						if !isHTTPMethod(method) {
							continue
						}

						if op, ok := operation.(map[string]interface{}); ok {
							endpoint := &OpenAPIEndpoint{
								Path:       path,
								Method:     strings.ToUpper(method),
								Parameters: make([]*OpenAPIParameter, 0),
								Responses:  make(map[string]*OpenAPISchema),
								Tags:       make([]string, 0),
							}

							// Extract operation details
							if summary, ok := op["summary"].(string); ok {
								endpoint.Summary = summary
							}
							if description, ok := op["description"].(string); ok {
								endpoint.Description = description
							}

							// Extract tags
							if tags, ok := op["tags"].([]interface{}); ok {
								for _, tag := range tags {
									if t, ok := tag.(string); ok {
										endpoint.Tags = append(endpoint.Tags, t)
									}
								}
							}

							// Extract parameters
							if params, ok := op["parameters"].([]interface{}); ok {
								for _, param := range params {
									if paramMap, ok := param.(map[string]interface{}); ok {
										parameter := &OpenAPIParameter{}
										if name, ok := paramMap["name"].(string); ok {
											parameter.Name = name
										}
										if in, ok := paramMap["in"].(string); ok {
											parameter.In = in
										}
										if required, ok := paramMap["required"].(bool); ok {
											parameter.Required = required
										}
										if desc, ok := paramMap["description"].(string); ok {
											parameter.Description = desc
										}
										if example, ok := paramMap["example"]; ok {
											parameter.Example = example
										}

										// Extract schema
										if schema, ok := paramMap["schema"].(map[string]interface{}); ok {
											parameter.Schema = p.extractSchema(schema)
										}

										endpoint.Parameters = append(endpoint.Parameters, parameter)
									}
								}
							}

							// Check if authentication is required
							if security, ok := op["security"].([]interface{}); ok && len(security) > 0 {
								endpoint.RequiresAuth = true
							}

							// Add the endpoint to the list
							p.Spec.Endpoints = append(p.Spec.Endpoints, endpoint)
						}
					}
				}
			}
		}
	} else {
		// OpenAPI 3.0.x and 3.1.x also use paths but with different structure
		if paths, ok := spec["paths"].(map[string]interface{}); ok {
			for path, pathItem := range paths {
				if methods, ok := pathItem.(map[string]interface{}); ok {
					for method, operation := range methods {
						// Skip non-HTTP method keys
						if !isHTTPMethod(method) {
							continue
						}

						if op, ok := operation.(map[string]interface{}); ok {
							endpoint := &OpenAPIEndpoint{
								Path:       path,
								Method:     strings.ToUpper(method),
								Parameters: make([]*OpenAPIParameter, 0),
								Responses:  make(map[string]*OpenAPISchema),
								Tags:       make([]string, 0),
							}

							// Extract operation details
							if summary, ok := op["summary"].(string); ok {
								endpoint.Summary = summary
							}
							if description, ok := op["description"].(string); ok {
								endpoint.Description = description
							}

							// Extract tags
							if tags, ok := op["tags"].([]interface{}); ok {
								for _, tag := range tags {
									if t, ok := tag.(string); ok {
										endpoint.Tags = append(endpoint.Tags, t)
									}
								}
							}

							// Extract parameters
							if params, ok := op["parameters"].([]interface{}); ok {
								for _, param := range params {
									if paramMap, ok := param.(map[string]interface{}); ok {
										parameter := &OpenAPIParameter{}
										if name, ok := paramMap["name"].(string); ok {
											parameter.Name = name
										}
										if in, ok := paramMap["in"].(string); ok {
											parameter.In = in
										}
										if required, ok := paramMap["required"].(bool); ok {
											parameter.Required = required
										}
										if desc, ok := paramMap["description"].(string); ok {
											parameter.Description = desc
										}
										if example, ok := paramMap["example"]; ok {
											parameter.Example = example
										}

										// Extract schema
										if schema, ok := paramMap["schema"].(map[string]interface{}); ok {
											parameter.Schema = p.extractSchema(schema)
										}

										endpoint.Parameters = append(endpoint.Parameters, parameter)
									}
								}
							}

							// Extract request body
							if requestBody, ok := op["requestBody"].(map[string]interface{}); ok {
								if content, ok := requestBody["content"].(map[string]interface{}); ok {
									// Try to get JSON schema first, then any other content type
									var schema map[string]interface{}
									if jsonContent, ok := content["application/json"].(map[string]interface{}); ok {
										if s, ok := jsonContent["schema"].(map[string]interface{}); ok {
											schema = s
										}
									} else {
										// Get the first content type
										for _, v := range content {
											if contentType, ok := v.(map[string]interface{}); ok {
												if s, ok := contentType["schema"].(map[string]interface{}); ok {
													schema = s
													break
												}
											}
										}
									}

									if schema != nil {
										endpoint.RequestBody = p.extractSchema(schema)
									}
								}
							}

							// Extract responses
							if responses, ok := op["responses"].(map[string]interface{}); ok {
								for code, response := range responses {
									if resp, ok := response.(map[string]interface{}); ok {
										if content, ok := resp["content"].(map[string]interface{}); ok {
											// Try to get JSON schema first, then any other content type
											var schema map[string]interface{}
											if jsonContent, ok := content["application/json"].(map[string]interface{}); ok {
												if s, ok := jsonContent["schema"].(map[string]interface{}); ok {
													schema = s
												}
											} else {
												// Get the first content type
												for _, v := range content {
													if contentType, ok := v.(map[string]interface{}); ok {
														if s, ok := contentType["schema"].(map[string]interface{}); ok {
															schema = s
															break
														}
													}
												}
											}

											if schema != nil {
												endpoint.Responses[code] = p.extractSchema(schema)
											}
										}
									}
								}
							}

							// Check if authentication is required
							if security, ok := op["security"].([]interface{}); ok && len(security) > 0 {
								endpoint.RequiresAuth = true
							}

							// Add the endpoint to the list
							p.Spec.Endpoints = append(p.Spec.Endpoints, endpoint)
						}
					}
				}
			}
		}
	}

	return nil
}

// extractSchema extracts a schema from an OpenAPI/Swagger specification
func (p *OpenAPIParser) extractSchema(schema map[string]interface{}) *OpenAPISchema {
	result := &OpenAPISchema{
		Properties: make(map[string]*OpenAPISchema),
		Required:   make([]string, 0),
		Enum:       make([]interface{}, 0),
	}

	// Extract type
	if t, ok := schema["type"].(string); ok {
		result.Type = t
	}

	// Extract format
	if format, ok := schema["format"].(string); ok {
		result.Format = format
	}

	// Extract properties for object types
	if result.Type == "object" {
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			for name, prop := range properties {
				if propMap, ok := prop.(map[string]interface{}); ok {
					result.Properties[name] = p.extractSchema(propMap)
				}
			}
		}
	}

	// Extract items for array types
	if result.Type == "array" {
		if items, ok := schema["items"].(map[string]interface{}); ok {
			result.Items = p.extractSchema(items)
		}
	}

	// Extract required properties
	if required, ok := schema["required"].([]interface{}); ok {
		for _, req := range required {
			if r, ok := req.(string); ok {
				result.Required = append(result.Required, r)
			}
		}
	}

	// Extract enum values
	if enum, ok := schema["enum"].([]interface{}); ok {
		result.Enum = enum
	}

	// Extract example
	if example, ok := schema["example"]; ok {
		result.Example = example
	}

	return result
}

// GetEndpoints returns all endpoints from the specification
func (p *OpenAPIParser) GetEndpoints() []*OpenAPIEndpoint {
	return p.Spec.Endpoints
}

// GetEndpointPaths returns all endpoint paths from the specification
func (p *OpenAPIParser) GetEndpointPaths() []string {
	paths := make([]string, 0, len(p.Spec.Endpoints))
	for _, endpoint := range p.Spec.Endpoints {
		paths = append(paths, endpoint.Path)
	}
	return paths
}

// GetEndpointsByTag returns all endpoints with the specified tag
func (p *OpenAPIParser) GetEndpointsByTag(tag string) []*OpenAPIEndpoint {
	endpoints := make([]*OpenAPIEndpoint, 0)
	for _, endpoint := range p.Spec.Endpoints {
		for _, t := range endpoint.Tags {
			if t == tag {
				endpoints = append(endpoints, endpoint)
				break
			}
		}
	}
	return endpoints
}

// GetEndpointsByMethod returns all endpoints with the specified HTTP method
func (p *OpenAPIParser) GetEndpointsByMethod(method string) []*OpenAPIEndpoint {
	method = strings.ToUpper(method)
	endpoints := make([]*OpenAPIEndpoint, 0)
	for _, endpoint := range p.Spec.Endpoints {
		if endpoint.Method == method {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// GetAuthRequiredEndpoints returns all endpoints that require authentication
func (p *OpenAPIParser) GetAuthRequiredEndpoints() []*OpenAPIEndpoint {
	endpoints := make([]*OpenAPIEndpoint, 0)
	for _, endpoint := range p.Spec.Endpoints {
		if endpoint.RequiresAuth {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// GetTags returns all tags from the specification
func (p *OpenAPIParser) GetTags() []string {
	tagMap := make(map[string]bool)
	for _, endpoint := range p.Spec.Endpoints {
		for _, tag := range endpoint.Tags {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	return tags
}

// GenerateWordlist generates a wordlist of API endpoints from the specification
func (p *OpenAPIParser) GenerateWordlist() []string {
	wordlist := make([]string, 0, len(p.Spec.Endpoints)*3) // Pre-allocate for better performance
	seen := make(map[string]bool)                          // Track already added entries to avoid duplicates

	for _, endpoint := range p.Spec.Endpoints {
		// Add the path as is
		if !seen[endpoint.Path] {
			wordlist = append(wordlist, endpoint.Path)
			seen[endpoint.Path] = true
		}

		// Add path with trailing slash if it doesn't have one
		pathWithSlash := endpoint.Path
		if !strings.HasSuffix(pathWithSlash, "/") {
			pathWithSlash += "/"
		}
		if !seen[pathWithSlash] {
			wordlist = append(wordlist, pathWithSlash)
			seen[pathWithSlash] = true
		}

		// Add path without trailing slash if it has one
		pathWithoutSlash := endpoint.Path
		if strings.HasSuffix(pathWithoutSlash, "/") && len(pathWithoutSlash) > 1 {
			pathWithoutSlash = pathWithoutSlash[:len(pathWithoutSlash)-1]
		}
		if !seen[pathWithoutSlash] {
			wordlist = append(wordlist, pathWithoutSlash)
			seen[pathWithoutSlash] = true
		}

		// Add path components
		components := strings.Split(endpoint.Path, "/")
		currentPath := ""
		for _, component := range components {
			if component == "" {
				continue
			}
			currentPath += "/" + component
			if !seen[currentPath] {
				wordlist = append(wordlist, currentPath)
				seen[currentPath] = true
			}

			// Also add this component with a trailing slash
			currentPathWithSlash := currentPath + "/"
			if !seen[currentPathWithSlash] {
				wordlist = append(wordlist, currentPathWithSlash)
				seen[currentPathWithSlash] = true
			}
		}
	}
	return wordlist
}

// Helper function to check if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to check if a string is an HTTP method
func isHTTPMethod(method string) bool {
	method = strings.ToUpper(method)
	return method == "GET" || method == "POST" || method == "PUT" || method == "DELETE" ||
		method == "PATCH" || method == "HEAD" || method == "OPTIONS" || method == "TRACE"
}
