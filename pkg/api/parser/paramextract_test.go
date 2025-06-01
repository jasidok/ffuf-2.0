package parser

import (
	"strings"
	"testing"
)

func TestAPIParameterExtractor_ExtractParameters(t *testing.T) {
	// Create a discovery with some endpoints and parameters
	discovery := NewAPIEndpointDiscovery("")
	discovery.Endpoints = []*DiscoveredEndpoint{
		{
			Path:   "/api/users",
			Method: "GET",
			Parameters: []*DiscoveredParameter{
				{
					Name:        "limit",
					In:          "query",
					Required:    false,
					Type:        "integer",
					Description: "Maximum number of users to return",
					Example:     10,
				},
				{
					Name:        "offset",
					In:          "query",
					Required:    false,
					Type:        "integer",
					Description: "Number of users to skip",
					Example:     0,
				},
			},
		},
		{
			Path:   "/api/users",
			Method: "POST",
			Parameters: []*DiscoveredParameter{
				{
					Name:        "name",
					In:          "body",
					Required:    true,
					Type:        "string",
					Description: "User's name",
					Example:     "John Doe",
				},
				{
					Name:        "email",
					In:          "body",
					Required:    true,
					Type:        "string",
					Description: "User's email",
					Example:     "john@example.com",
				},
			},
		},
		{
			Path:   "/api/users/{id}",
			Method: "GET",
			Parameters: []*DiscoveredParameter{
				{
					Name:        "id",
					In:          "path",
					Required:    true,
					Type:        "integer",
					Description: "User ID",
					Example:     123,
				},
				{
					Name:        "fields",
					In:          "query",
					Required:    false,
					Type:        "string",
					Description: "Comma-separated list of fields to include",
					Example:     "name,email,role",
				},
			},
		},
		{
			Path:   "/api/users/{id}",
			Method: "PUT",
			Parameters: []*DiscoveredParameter{
				{
					Name:        "id",
					In:          "path",
					Required:    true,
					Type:        "integer",
					Description: "User ID",
					Example:     123,
				},
				{
					Name:        "name",
					In:          "body",
					Required:    false,
					Type:        "string",
					Description: "User's name",
					Example:     "John Doe",
				},
				{
					Name:        "email",
					In:          "body",
					Required:    false,
					Type:        "string",
					Description: "User's email",
					Example:     "john@example.com",
				},
			},
		},
	}

	// Create a parameter extractor
	extractor := NewAPIParameterExtractor(discovery)

	// Extract parameters
	err := extractor.ExtractParameters()
	if err != nil {
		t.Fatalf("Failed to extract parameters: %v", err)
	}

	// Check the extracted parameters
	params := extractor.GetParameters()
	if len(params) != 6 {
		t.Errorf("Expected 6 parameters, got %d", len(params))
	}

	// Check parameter frequencies
	idParam := extractor.GetParameterByName("id")
	if idParam == nil {
		t.Errorf("Expected to find parameter 'id'")
	} else if idParam.Frequency != 2 {
		t.Errorf("Expected parameter 'id' to have frequency 2, got %d", idParam.Frequency)
	}

	nameParam := extractor.GetParameterByName("name")
	if nameParam == nil {
		t.Errorf("Expected to find parameter 'name'")
	} else if nameParam.Frequency != 2 {
		t.Errorf("Expected parameter 'name' to have frequency 2, got %d", nameParam.Frequency)
	}

	// Check parameter locations
	queryParams := extractor.GetParametersByLocation("query")
	if len(queryParams) != 3 {
		t.Errorf("Expected 3 query parameters, got %d", len(queryParams))
	}

	bodyParams := extractor.GetParametersByLocation("body")
	if len(bodyParams) != 2 {
		t.Errorf("Expected 2 body parameters, got %d", len(bodyParams))
	}

	pathParams := extractor.GetParametersByLocation("path")
	if len(pathParams) != 1 {
		t.Errorf("Expected 1 path parameter, got %d", len(pathParams))
	}

	// Check parameter types
	stringParams := extractor.GetParametersByType("string")
	if len(stringParams) != 3 {
		t.Errorf("Expected 3 string parameters, got %d", len(stringParams))
	}

	intParams := extractor.GetParametersByType("integer")
	if len(intParams) != 3 {
		t.Errorf("Expected 3 integer parameters, got %d", len(intParams))
	}

	// Check required parameters
	requiredParams := extractor.GetRequiredParameters()
	if len(requiredParams) != 3 {
		t.Errorf("Expected 3 required parameters, got %d", len(requiredParams))
	}

	// Check parameters by frequency
	freqParams := extractor.GetParametersByFrequency()
	if len(freqParams) != 6 {
		t.Errorf("Expected 6 parameters, got %d", len(freqParams))
	}
	if freqParams[0].Frequency < freqParams[1].Frequency {
		t.Errorf("Expected parameters to be sorted by frequency (descending)")
	}

	// Check parameter wordlist
	wordlist := extractor.GenerateParameterWordlist()
	if len(wordlist) != 6 {
		t.Errorf("Expected 6 parameters in wordlist, got %d", len(wordlist))
	}
	expectedParams := []string{"limit", "offset", "name", "email", "id", "fields"}
	for _, param := range expectedParams {
		if !contains(wordlist, param) {
			t.Errorf("Expected wordlist to contain '%s'", param)
		}
	}

	// Check parameter wordlist by location
	queryWordlist := extractor.GenerateParameterWordlistByLocation("query")
	if len(queryWordlist) != 3 {
		t.Errorf("Expected 3 query parameters in wordlist, got %d", len(queryWordlist))
	}
	expectedQueryParams := []string{"limit", "offset", "fields"}
	for _, param := range expectedQueryParams {
		if !contains(queryWordlist, param) {
			t.Errorf("Expected query wordlist to contain '%s'", param)
		}
	}

	// Check parameter report
	report := extractor.GenerateParameterReport()
	if !strings.Contains(report, "Total parameters: 6") {
		t.Errorf("Expected report to contain 'Total parameters: 6'")
	}
	if !strings.Contains(report, "query: 3") {
		t.Errorf("Expected report to contain 'query: 3'")
	}
	if !strings.Contains(report, "body: 2") {
		t.Errorf("Expected report to contain 'body: 2'")
	}
	if !strings.Contains(report, "path: 1") {
		t.Errorf("Expected report to contain 'path: 1'")
	}
}

func TestAPIParameterExtractor_GenerateParameterFuzzingPayloads(t *testing.T) {
	// Create a discovery with some endpoints and parameters
	discovery := NewAPIEndpointDiscovery("")
	discovery.Endpoints = []*DiscoveredEndpoint{
		{
			Path:   "/api/test",
			Method: "POST",
			Parameters: []*DiscoveredParameter{
				{
					Name:     "string_param",
					In:       "body",
					Required: true,
					Type:     "string",
					Example:  "test",
				},
				{
					Name:     "int_param",
					In:       "body",
					Required: true,
					Type:     "integer",
					Example:  42,
				},
				{
					Name:     "bool_param",
					In:       "body",
					Required: true,
					Type:     "boolean",
					Example:  true,
				},
				{
					Name:     "array_param",
					In:       "body",
					Required: true,
					Type:     "array",
					Example:  []interface{}{1, 2, 3},
				},
				{
					Name:     "object_param",
					In:       "body",
					Required: true,
					Type:     "object",
					Example:  map[string]interface{}{"key": "value"},
				},
				{
					Name:     "unknown_param",
					In:       "body",
					Required: true,
					Type:     "unknown",
					Example:  "test",
				},
			},
		},
	}

	// Create a parameter extractor
	extractor := NewAPIParameterExtractor(discovery)

	// Extract parameters
	err := extractor.ExtractParameters()
	if err != nil {
		t.Fatalf("Failed to extract parameters: %v", err)
	}

	// Generate fuzzing payloads
	payloads := extractor.GenerateParameterFuzzingPayloads()

	// Check payloads for each parameter type
	if len(payloads["string_param"]) == 0 {
		t.Errorf("Expected payloads for string_param")
	}
	if !contains(payloads["string_param"], "test") {
		t.Errorf("Expected string_param payloads to contain example value 'test'")
	}
	if !contains(payloads["string_param"], "<script>alert(1)</script>") {
		t.Errorf("Expected string_param payloads to contain XSS payload")
	}

	if len(payloads["int_param"]) == 0 {
		t.Errorf("Expected payloads for int_param")
	}
	if !contains(payloads["int_param"], "42") {
		t.Errorf("Expected int_param payloads to contain example value '42'")
	}
	if !contains(payloads["int_param"], "0") {
		t.Errorf("Expected int_param payloads to contain '0'")
	}

	if len(payloads["bool_param"]) == 0 {
		t.Errorf("Expected payloads for bool_param")
	}
	if !contains(payloads["bool_param"], "true") {
		t.Errorf("Expected bool_param payloads to contain 'true'")
	}
	if !contains(payloads["bool_param"], "false") {
		t.Errorf("Expected bool_param payloads to contain 'false'")
	}

	if len(payloads["array_param"]) == 0 {
		t.Errorf("Expected payloads for array_param")
	}
	if !contains(payloads["array_param"], "[]") {
		t.Errorf("Expected array_param payloads to contain '[]'")
	}
	if !contains(payloads["array_param"], "[1,2,3]") {
		t.Errorf("Expected array_param payloads to contain '[1,2,3]'")
	}

	if len(payloads["object_param"]) == 0 {
		t.Errorf("Expected payloads for object_param")
	}
	if !contains(payloads["object_param"], "{}") {
		t.Errorf("Expected object_param payloads to contain '{}'")
	}
	if !contains(payloads["object_param"], "{\"key\":\"value\"}") {
		t.Errorf("Expected object_param payloads to contain '{\"key\":\"value\"}'")
	}

	if len(payloads["unknown_param"]) == 0 {
		t.Errorf("Expected payloads for unknown_param")
	}
	if !contains(payloads["unknown_param"], "test") {
		t.Errorf("Expected unknown_param payloads to contain example value 'test'")
	}
}

func TestAPIParameterExtractor_NoDiscovery(t *testing.T) {
	// Create a parameter extractor with no discovery
	extractor := NewAPIParameterExtractor(nil)

	// Extract parameters
	err := extractor.ExtractParameters()
	if err == nil {
		t.Errorf("Expected error when extracting parameters with no discovery")
	}
}

func TestAPIParameterExtractor_EmptyDiscovery(t *testing.T) {
	// Create a discovery with no endpoints
	discovery := NewAPIEndpointDiscovery("")

	// Create a parameter extractor
	extractor := NewAPIParameterExtractor(discovery)

	// Extract parameters
	err := extractor.ExtractParameters()
	if err == nil {
		t.Errorf("Expected error when extracting parameters with no endpoints")
	}
}

func TestGenerateStringPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "string",
		Example: "example",
	}

	payloads := generateStringPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for string parameter")
	}
	if !contains(payloads, "example") {
		t.Errorf("Expected payloads to contain example value 'example'")
	}
	if !contains(payloads, "") {
		t.Errorf("Expected payloads to contain empty string")
	}
	if !contains(payloads, "<script>alert(1)</script>") {
		t.Errorf("Expected payloads to contain XSS payload")
	}
}

func TestGenerateNumberPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "integer",
		Example: 42,
	}

	payloads := generateNumberPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for number parameter")
	}
	if !contains(payloads, "42") {
		t.Errorf("Expected payloads to contain example value '42'")
	}
	if !contains(payloads, "0") {
		t.Errorf("Expected payloads to contain '0'")
	}
	if !contains(payloads, "-1") {
		t.Errorf("Expected payloads to contain '-1'")
	}
}

func TestGenerateBooleanPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "boolean",
		Example: true,
	}

	payloads := generateBooleanPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for boolean parameter")
	}
	if !contains(payloads, "true") {
		t.Errorf("Expected payloads to contain 'true'")
	}
	if !contains(payloads, "false") {
		t.Errorf("Expected payloads to contain 'false'")
	}
	if !contains(payloads, "1") {
		t.Errorf("Expected payloads to contain '1'")
	}
	if !contains(payloads, "0") {
		t.Errorf("Expected payloads to contain '0'")
	}
}

func TestGenerateArrayPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "array",
		Example: []interface{}{1, 2, 3},
	}

	payloads := generateArrayPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for array parameter")
	}
	if !contains(payloads, "[]") {
		t.Errorf("Expected payloads to contain '[]'")
	}
	if !contains(payloads, "[1,2,3]") {
		t.Errorf("Expected payloads to contain '[1,2,3]'")
	}
	if !contains(payloads, "") {
		t.Errorf("Expected payloads to contain empty string")
	}
}

func TestGenerateObjectPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "object",
		Example: map[string]interface{}{"key": "value"},
	}

	payloads := generateObjectPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for object parameter")
	}
	if !contains(payloads, "{}") {
		t.Errorf("Expected payloads to contain '{}'")
	}
	if !contains(payloads, "{\"key\":\"value\"}") {
		t.Errorf("Expected payloads to contain '{\"key\":\"value\"}'")
	}
	if !contains(payloads, "") {
		t.Errorf("Expected payloads to contain empty string")
	}
}

func TestGenerateDefaultPayloads(t *testing.T) {
	param := &ExtractedParameter{
		Name:    "test",
		Type:    "unknown",
		Example: "example",
	}

	payloads := generateDefaultPayloads(param)
	if len(payloads) == 0 {
		t.Errorf("Expected payloads for unknown parameter")
	}
	if !contains(payloads, "example") {
		t.Errorf("Expected payloads to contain example value 'example'")
	}
	if !contains(payloads, "") {
		t.Errorf("Expected payloads to contain empty string")
	}
	if !contains(payloads, "null") {
		t.Errorf("Expected payloads to contain 'null'")
	}
}