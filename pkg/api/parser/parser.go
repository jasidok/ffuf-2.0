// Package parser provides functionality for parsing API responses and specifications.
//
// This package includes parsers for various API formats including JSON, XML,
// GraphQL, and OpenAPI/Swagger specifications. It enables extraction of
// meaningful data from API responses and automatic discovery of API endpoints
// from documentation.
package parser

import (
	"encoding/json"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// ResponseFormat represents the format of an API response
type ResponseFormat int

const (
	// FormatUnknown represents an unknown response format
	FormatUnknown ResponseFormat = iota
	// FormatJSON represents a JSON response
	FormatJSON
	// FormatXML represents an XML response
	FormatXML
	// FormatGraphQL represents a GraphQL response
	FormatGraphQL
)

// ResponseParser provides methods for parsing API responses
type ResponseParser struct {
	format ResponseFormat
}

// NewResponseParser creates a new ResponseParser with auto-detected format
func NewResponseParser(contentType string) *ResponseParser {
	format := FormatUnknown
	
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "application/json") {
		format = FormatJSON
	} else if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		format = FormatXML
	} else if strings.Contains(contentType, "application/graphql") {
		format = FormatGraphQL
	}
	
	return &ResponseParser{
		format: format,
	}
}

// ParseJSON parses a JSON response and returns the parsed data
func (p *ResponseParser) ParseJSON(data []byte) (map[string]interface{}, error) {
	if p.format != FormatJSON && p.format != FormatUnknown {
		return nil, api.NewAPIError("Response is not in JSON format", 0)
	}
	
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
	}
	
	return result, nil
}

// ExtractAPIEndpoints attempts to extract API endpoints from a response
func (p *ResponseParser) ExtractAPIEndpoints(data []byte) ([]string, error) {
	endpoints := make([]string, 0)
	
	// For JSON responses, look for URL-like strings
	if p.format == FormatJSON {
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
		}
		
		// Extract URL-like strings from the JSON data
		extractURLsFromJSON(jsonData, &endpoints)
	}
	
	// For other formats, implement specific extraction logic
	
	return endpoints, nil
}

// Helper function to extract URL-like strings from JSON data
func extractURLsFromJSON(data interface{}, endpoints *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair
		for key, value := range v {
			// Check if key looks like an endpoint
			if isEndpointLike(key) {
				*endpoints = append(*endpoints, key)
			}
			
			// Recursively process the value
			extractURLsFromJSON(value, endpoints)
		}
	case []interface{}:
		// Process each element in the array
		for _, elem := range v {
			extractURLsFromJSON(elem, endpoints)
		}
	case string:
		// Check if the string looks like an endpoint
		if isEndpointLike(v) {
			*endpoints = append(*endpoints, v)
		}
	}
}

// Helper function to check if a string looks like an API endpoint
func isEndpointLike(s string) bool {
	// Check for common API endpoint patterns
	if strings.HasPrefix(s, "/api/") || 
	   strings.HasPrefix(s, "/v1/") || 
	   strings.HasPrefix(s, "/v2/") || 
	   strings.HasPrefix(s, "/rest/") || 
	   strings.Contains(s, "/graphql") {
		return true
	}
	
	return false
}