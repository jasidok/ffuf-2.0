// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// APIEndpointDiscovery provides methods for discovering API endpoints from documentation
type APIEndpointDiscovery struct {
	// Base URL for the API
	BaseURL string
	// Discovered endpoints
	Endpoints []*DiscoveredEndpoint
	// Parser used for discovery
	Parser interface{}
}

// DiscoveredEndpoint represents an API endpoint discovered from documentation
type DiscoveredEndpoint struct {
	// Full URL of the endpoint
	URL string
	// HTTP method (GET, POST, etc.)
	Method string
	// Path of the endpoint
	Path string
	// Parameters for the endpoint
	Parameters []*DiscoveredParameter
	// Whether the endpoint requires authentication
	RequiresAuth bool
	// Description of the endpoint
	Description string
	// Tags associated with the endpoint
	Tags []string
	// Source of the endpoint (e.g., "OpenAPI", "Swagger")
	Source string
}

// DiscoveredParameter represents a parameter for an API endpoint
type DiscoveredParameter struct {
	// Name of the parameter
	Name string
	// Location of the parameter (path, query, header, cookie, body)
	In string
	// Whether the parameter is required
	Required bool
	// Type of the parameter (string, integer, etc.)
	Type string
	// Description of the parameter
	Description string
	// Example value for the parameter
	Example interface{}
}

// NewAPIEndpointDiscovery creates a new APIEndpointDiscovery
func NewAPIEndpointDiscovery(baseURL string) *APIEndpointDiscovery {
	return &APIEndpointDiscovery{
		BaseURL:   baseURL,
		Endpoints: make([]*DiscoveredEndpoint, 0),
	}
}

// DiscoverFromOpenAPI discovers API endpoints from an OpenAPI/Swagger specification
func (d *APIEndpointDiscovery) DiscoverFromOpenAPI(specPath string) error {
	// Create a new OpenAPI parser
	parser := NewOpenAPIParser()
	d.Parser = parser

	// Determine if the spec path is a URL or a file path
	if strings.HasPrefix(specPath, "http://") || strings.HasPrefix(specPath, "https://") {
		// Parse from URL
		if err := parser.ParseFromURL(specPath); err != nil {
			return err
		}
	} else {
		// Parse from file
		if err := parser.ParseFromFile(specPath); err != nil {
			return err
		}
	}

	// If base URL is not set, use the one from the spec
	if d.BaseURL == "" && parser.Spec.BaseURL != "" {
		d.BaseURL = parser.Spec.BaseURL
	}

	// Convert OpenAPI endpoints to discovered endpoints
	for _, endpoint := range parser.GetEndpoints() {
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
		if d.BaseURL != "" {
			// Ensure the base URL doesn't end with a slash if the path starts with one
			baseURL := strings.TrimSuffix(d.BaseURL, "/")
			path := endpoint.Path
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			discoveredEndpoint.URL = baseURL + path
		}

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
					Example:     prop.Example,
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
		d.Endpoints = append(d.Endpoints, discoveredEndpoint)
	}

	return nil
}

// GetEndpoints returns all discovered endpoints
func (d *APIEndpointDiscovery) GetEndpoints() []*DiscoveredEndpoint {
	return d.Endpoints
}

// GetEndpointsByMethod returns all endpoints with the specified HTTP method
func (d *APIEndpointDiscovery) GetEndpointsByMethod(method string) []*DiscoveredEndpoint {
	method = strings.ToUpper(method)
	endpoints := make([]*DiscoveredEndpoint, 0)
	for _, endpoint := range d.Endpoints {
		if endpoint.Method == method {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// GetEndpointsByTag returns all endpoints with the specified tag
func (d *APIEndpointDiscovery) GetEndpointsByTag(tag string) []*DiscoveredEndpoint {
	endpoints := make([]*DiscoveredEndpoint, 0)
	for _, endpoint := range d.Endpoints {
		for _, t := range endpoint.Tags {
			if t == tag {
				endpoints = append(endpoints, endpoint)
				break
			}
		}
	}
	return endpoints
}

// GetAuthRequiredEndpoints returns all endpoints that require authentication
func (d *APIEndpointDiscovery) GetAuthRequiredEndpoints() []*DiscoveredEndpoint {
	endpoints := make([]*DiscoveredEndpoint, 0)
	for _, endpoint := range d.Endpoints {
		if endpoint.RequiresAuth {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// GetEndpointsByPath returns all endpoints that match the specified path pattern
func (d *APIEndpointDiscovery) GetEndpointsByPath(pathPattern string) []*DiscoveredEndpoint {
	endpoints := make([]*DiscoveredEndpoint, 0)
	for _, endpoint := range d.Endpoints {
		if pathMatches(endpoint.Path, pathPattern) {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// GenerateWordlist generates a wordlist of API endpoints
func (d *APIEndpointDiscovery) GenerateWordlist() []string {
	wordlist := make([]string, 0, len(d.Endpoints))
	for _, endpoint := range d.Endpoints {
		// Add the path as is
		wordlist = append(wordlist, endpoint.Path)

		// Add path with trailing slash if it doesn't have one
		if !strings.HasSuffix(endpoint.Path, "/") {
			wordlist = append(wordlist, endpoint.Path+"/")
		}

		// Add path without trailing slash if it has one
		if strings.HasSuffix(endpoint.Path, "/") && len(endpoint.Path) > 1 {
			wordlist = append(wordlist, endpoint.Path[:len(endpoint.Path)-1])
		}

		// Add path components
		components := strings.Split(endpoint.Path, "/")
		currentPath := ""
		for _, component := range components {
			if component == "" {
				continue
			}
			currentPath += "/" + component
			if !contains(wordlist, currentPath) {
				wordlist = append(wordlist, currentPath)
			}
		}
	}
	return wordlist
}

// GenerateURLs generates a list of full URLs for the discovered endpoints
func (d *APIEndpointDiscovery) GenerateURLs() []string {
	urls := make([]string, 0, len(d.Endpoints))
	for _, endpoint := range d.Endpoints {
		if endpoint.URL != "" {
			urls = append(urls, endpoint.URL)
		} else if d.BaseURL != "" {
			// Ensure the base URL doesn't end with a slash if the path starts with one
			baseURL := strings.TrimSuffix(d.BaseURL, "/")
			path := endpoint.Path
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			urls = append(urls, baseURL+path)
		}
	}
	return urls
}

// GenerateParameterWordlist generates a wordlist of parameter names
func (d *APIEndpointDiscovery) GenerateParameterWordlist() []string {
	paramMap := make(map[string]bool)
	for _, endpoint := range d.Endpoints {
		for _, param := range endpoint.Parameters {
			paramMap[param.Name] = true
		}
	}

	params := make([]string, 0, len(paramMap))
	for param := range paramMap {
		params = append(params, param)
	}
	return params
}

// DiscoverFromDirectory discovers API endpoints from all OpenAPI/Swagger specifications in a directory
func (d *APIEndpointDiscovery) DiscoverFromDirectory(dirPath string) error {
	// Get all JSON and YAML files in the directory
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to glob JSON files: %s", err.Error()), 0)
	}

	yamlFiles, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to glob YAML files: %s", err.Error()), 0)
	}
	files = append(files, yamlFiles...)

	ymlFiles, err := filepath.Glob(filepath.Join(dirPath, "*.yml"))
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to glob YML files: %s", err.Error()), 0)
	}
	files = append(files, ymlFiles...)

	// Try to parse each file as an OpenAPI/Swagger specification
	for _, file := range files {
		// Create a new discovery for each file to avoid mixing endpoints
		fileDiscovery := NewAPIEndpointDiscovery(d.BaseURL)
		if err := fileDiscovery.DiscoverFromOpenAPI(file); err != nil {
			// Skip files that can't be parsed as OpenAPI/Swagger
			continue
		}

		// Add the discovered endpoints to the main discovery
		d.Endpoints = append(d.Endpoints, fileDiscovery.Endpoints...)
	}

	return nil
}

// DiscoverFromURL discovers API endpoints from a URL that might be an API documentation
func (d *APIEndpointDiscovery) DiscoverFromURL(targetURL string) error {
	// Parse the URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return api.NewAPIError(fmt.Sprintf("Invalid URL: %s", err.Error()), 0)
	}

	// Set the base URL if not already set
	if d.BaseURL == "" {
		d.BaseURL = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	}

	// Try common paths for API documentation
	commonPaths := []string{
		"/swagger.json",
		"/api-docs.json",
		"/openapi.json",
		"/swagger/v1/swagger.json",
		"/api/swagger.json",
		"/api/v1/swagger.json",
		"/api/v2/swagger.json",
		"/api/v3/swagger.json",
		"/api/docs/swagger.json",
		"/docs/swagger.json",
		"/swagger-ui/swagger.json",
		"/swagger-resources",
	}

	for _, path := range commonPaths {
		// Create the full URL
		docURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path)

		// Try to discover from this URL
		fileDiscovery := NewAPIEndpointDiscovery(d.BaseURL)
		if err := fileDiscovery.DiscoverFromOpenAPI(docURL); err != nil {
			// Skip URLs that can't be parsed as OpenAPI/Swagger
			continue
		}

		// Add the discovered endpoints to the main discovery
		d.Endpoints = append(d.Endpoints, fileDiscovery.Endpoints...)
	}

	return nil
}

// Helper function to check if a path matches a pattern
func pathMatches(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Simple wildcard matching
	if strings.Contains(pattern, "*") {
		// Split both path and pattern by "/"
		pathParts := strings.Split(path, "/")
		patternParts := strings.Split(pattern, "/")

		// If the number of parts doesn't match, they can't match
		if len(pathParts) != len(patternParts) {
			return false
		}

		// Check each part
		for i, patternPart := range patternParts {
			if patternPart == "*" {
				// Wildcard matches any single path segment
				continue
			}
			if pathParts[i] != patternPart {
				return false
			}
		}

		return true
	}

	return false
}
