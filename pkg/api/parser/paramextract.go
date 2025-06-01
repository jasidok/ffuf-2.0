// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"fmt"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// APIParameterExtractor provides methods for extracting parameters from API documentation
type APIParameterExtractor struct {
	// Discovery used for extracting parameters
	Discovery *APIEndpointDiscovery
	// Extracted parameters
	Parameters []*ExtractedParameter
	// Parameter types by name
	ParameterTypes map[string]string
	// Parameter locations by name
	ParameterLocations map[string]string
	// Required parameters
	RequiredParameters map[string]bool
	// Parameter descriptions
	ParameterDescriptions map[string]string
	// Parameter examples
	ParameterExamples map[string]interface{}
}

// ExtractedParameter represents a parameter extracted from API documentation
type ExtractedParameter struct {
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
	// Endpoints that use this parameter
	Endpoints []*DiscoveredEndpoint
	// Frequency of the parameter across all endpoints
	Frequency int
}

// NewAPIParameterExtractor creates a new APIParameterExtractor
func NewAPIParameterExtractor(discovery *APIEndpointDiscovery) *APIParameterExtractor {
	return &APIParameterExtractor{
		Discovery:             discovery,
		Parameters:            make([]*ExtractedParameter, 0),
		ParameterTypes:        make(map[string]string),
		ParameterLocations:    make(map[string]string),
		RequiredParameters:    make(map[string]bool),
		ParameterDescriptions: make(map[string]string),
		ParameterExamples:     make(map[string]interface{}),
	}
}

// ExtractParameters extracts parameters from the discovered endpoints
func (e *APIParameterExtractor) ExtractParameters() error {
	if e.Discovery == nil {
		return api.NewAPIError("No discovery provided", 0)
	}

	// Get all endpoints
	endpoints := e.Discovery.GetEndpoints()
	if len(endpoints) == 0 {
		return api.NewAPIError("No endpoints discovered", 0)
	}

	// Create a map to track parameter frequency
	paramFrequency := make(map[string]int)
	// Create a map to track endpoints for each parameter
	paramEndpoints := make(map[string][]*DiscoveredEndpoint)

	// Extract parameters from each endpoint
	for _, endpoint := range endpoints {
		for _, param := range endpoint.Parameters {
			// Update frequency
			paramFrequency[param.Name]++
			// Add endpoint to parameter's endpoints
			paramEndpoints[param.Name] = append(paramEndpoints[param.Name], endpoint)

			// Update parameter information
			e.ParameterTypes[param.Name] = param.Type
			e.ParameterLocations[param.Name] = param.In
			e.RequiredParameters[param.Name] = param.Required
			if param.Description != "" {
				e.ParameterDescriptions[param.Name] = param.Description
			}
			if param.Example != nil {
				e.ParameterExamples[param.Name] = param.Example
			}
		}
	}

	// Create extracted parameters
	for name, frequency := range paramFrequency {
		param := &ExtractedParameter{
			Name:        name,
			In:          e.ParameterLocations[name],
			Required:    e.RequiredParameters[name],
			Type:        e.ParameterTypes[name],
			Description: e.ParameterDescriptions[name],
			Example:     e.ParameterExamples[name],
			Endpoints:   paramEndpoints[name],
			Frequency:   frequency,
		}
		e.Parameters = append(e.Parameters, param)
	}

	return nil
}

// GetParameters returns all extracted parameters
func (e *APIParameterExtractor) GetParameters() []*ExtractedParameter {
	return e.Parameters
}

// GetParametersByLocation returns all parameters with the specified location
func (e *APIParameterExtractor) GetParametersByLocation(location string) []*ExtractedParameter {
	location = strings.ToLower(location)
	params := make([]*ExtractedParameter, 0)
	for _, param := range e.Parameters {
		if strings.ToLower(param.In) == location {
			params = append(params, param)
		}
	}
	return params
}

// GetParametersByType returns all parameters with the specified type
func (e *APIParameterExtractor) GetParametersByType(paramType string) []*ExtractedParameter {
	paramType = strings.ToLower(paramType)
	params := make([]*ExtractedParameter, 0)
	for _, param := range e.Parameters {
		if strings.ToLower(param.Type) == paramType {
			params = append(params, param)
		}
	}
	return params
}

// GetRequiredParameters returns all required parameters
func (e *APIParameterExtractor) GetRequiredParameters() []*ExtractedParameter {
	params := make([]*ExtractedParameter, 0)
	for _, param := range e.Parameters {
		if param.Required {
			params = append(params, param)
		}
	}
	return params
}

// GetParametersByFrequency returns all parameters sorted by frequency (descending)
func (e *APIParameterExtractor) GetParametersByFrequency() []*ExtractedParameter {
	// Create a copy of the parameters
	params := make([]*ExtractedParameter, len(e.Parameters))
	copy(params, e.Parameters)

	// Sort by frequency (descending)
	for i := 0; i < len(params); i++ {
		for j := i + 1; j < len(params); j++ {
			if params[i].Frequency < params[j].Frequency {
				params[i], params[j] = params[j], params[i]
			}
		}
	}

	return params
}

// GetParameterByName returns a parameter by name
func (e *APIParameterExtractor) GetParameterByName(name string) *ExtractedParameter {
	for _, param := range e.Parameters {
		if param.Name == name {
			return param
		}
	}
	return nil
}

// GenerateParameterWordlist generates a wordlist of parameter names
func (e *APIParameterExtractor) GenerateParameterWordlist() []string {
	params := make([]string, 0, len(e.Parameters))
	for _, param := range e.Parameters {
		params = append(params, param.Name)
	}
	return params
}

// GenerateParameterWordlistByLocation generates a wordlist of parameter names for a specific location
func (e *APIParameterExtractor) GenerateParameterWordlistByLocation(location string) []string {
	location = strings.ToLower(location)
	params := make([]string, 0)
	for _, param := range e.Parameters {
		if strings.ToLower(param.In) == location {
			params = append(params, param.Name)
		}
	}
	return params
}

// GenerateParameterReport generates a report of extracted parameters
func (e *APIParameterExtractor) GenerateParameterReport() string {
	if len(e.Parameters) == 0 {
		return "No parameters extracted"
	}

	report := "# API Parameter Report\n\n"
	report += fmt.Sprintf("Total parameters: %d\n\n", len(e.Parameters))

	// Add parameter locations
	locations := make(map[string]int)
	for _, param := range e.Parameters {
		locations[param.In]++
	}
	report += "## Parameter Locations\n\n"
	for location, count := range locations {
		report += fmt.Sprintf("- %s: %d\n", location, count)
	}
	report += "\n"

	// Add parameter types
	types := make(map[string]int)
	for _, param := range e.Parameters {
		types[param.Type]++
	}
	report += "## Parameter Types\n\n"
	for paramType, count := range types {
		report += fmt.Sprintf("- %s: %d\n", paramType, count)
	}
	report += "\n"

	// Add required parameters
	requiredParams := e.GetRequiredParameters()
	report += "## Required Parameters\n\n"
	for _, param := range requiredParams {
		report += fmt.Sprintf("- %s (%s): %s\n", param.Name, param.In, param.Description)
	}
	report += "\n"

	// Add most common parameters
	commonParams := e.GetParametersByFrequency()
	report += "## Most Common Parameters\n\n"
	for i, param := range commonParams {
		if i >= 10 {
			break
		}
		report += fmt.Sprintf("- %s (%s): %d occurrences\n", param.Name, param.In, param.Frequency)
	}
	report += "\n"

	// Add parameter details
	report += "## Parameter Details\n\n"
	for _, param := range e.Parameters {
		report += fmt.Sprintf("### %s\n\n", param.Name)
		report += fmt.Sprintf("- Location: %s\n", param.In)
		report += fmt.Sprintf("- Type: %s\n", param.Type)
		report += fmt.Sprintf("- Required: %t\n", param.Required)
		report += fmt.Sprintf("- Frequency: %d\n", param.Frequency)
		if param.Description != "" {
			report += fmt.Sprintf("- Description: %s\n", param.Description)
		}
		if param.Example != nil {
			report += fmt.Sprintf("- Example: %v\n", param.Example)
		}
		report += "\n"
	}

	return report
}

// ExtractParametersFromOpenAPI extracts parameters from an OpenAPI/Swagger specification
func (e *APIParameterExtractor) ExtractParametersFromOpenAPI(specPath string) error {
	// Create a new discovery if not provided
	if e.Discovery == nil {
		e.Discovery = NewAPIEndpointDiscovery("")
	}

	// Discover endpoints from the OpenAPI specification
	if err := e.Discovery.DiscoverFromOpenAPI(specPath); err != nil {
		return err
	}

	// Extract parameters from the discovered endpoints
	return e.ExtractParameters()
}

// ExtractParametersFromDirectory extracts parameters from all OpenAPI/Swagger specifications in a directory
func (e *APIParameterExtractor) ExtractParametersFromDirectory(dirPath string) error {
	// Create a new discovery if not provided
	if e.Discovery == nil {
		e.Discovery = NewAPIEndpointDiscovery("")
	}

	// Discover endpoints from the directory
	if err := e.Discovery.DiscoverFromDirectory(dirPath); err != nil {
		return err
	}

	// Extract parameters from the discovered endpoints
	return e.ExtractParameters()
}

// ExtractParametersFromURL extracts parameters from a URL that might be an API documentation
func (e *APIParameterExtractor) ExtractParametersFromURL(targetURL string) error {
	// Create a new discovery if not provided
	if e.Discovery == nil {
		e.Discovery = NewAPIEndpointDiscovery("")
	}

	// Discover endpoints from the URL
	if err := e.Discovery.DiscoverFromURL(targetURL); err != nil {
		return err
	}

	// Extract parameters from the discovered endpoints
	return e.ExtractParameters()
}

// GenerateParameterFuzzingPayloads generates fuzzing payloads for parameters
func (e *APIParameterExtractor) GenerateParameterFuzzingPayloads() map[string][]string {
	payloads := make(map[string][]string)

	for _, param := range e.Parameters {
		// Generate payloads based on parameter type
		switch strings.ToLower(param.Type) {
		case "string":
			payloads[param.Name] = generateStringPayloads(param)
		case "integer", "number":
			payloads[param.Name] = generateNumberPayloads(param)
		case "boolean":
			payloads[param.Name] = generateBooleanPayloads(param)
		case "array":
			payloads[param.Name] = generateArrayPayloads(param)
		case "object":
			payloads[param.Name] = generateObjectPayloads(param)
		default:
			// Default payloads for unknown types
			payloads[param.Name] = generateDefaultPayloads(param)
		}
	}

	return payloads
}

// Helper functions to generate payloads based on parameter type

func generateStringPayloads(param *ExtractedParameter) []string {
	payloads := []string{
		"",                    // Empty string
		"test",                // Simple string
		"12345",               // Numeric string
		"!@#$%^&*()",          // Special characters
		"<script>alert(1)</script>", // XSS payload
		"' OR '1'='1",         // SQL injection payload
		strings.Repeat("A", 1000), // Long string
	}

	// Add example value if available
	if param.Example != nil {
		if exampleStr, ok := param.Example.(string); ok {
			payloads = append(payloads, exampleStr)
		}
	}

	return payloads
}

func generateNumberPayloads(param *ExtractedParameter) []string {
	payloads := []string{
		"0",
		"1",
		"-1",
		"999999",
		"-999999",
		"3.14159",
		"0.0",
		"1e10",
		"NaN",
		"Infinity",
		"-Infinity",
		"not_a_number",
	}

	// Add example value if available
	if param.Example != nil {
		payloads = append(payloads, fmt.Sprintf("%v", param.Example))
	}

	return payloads
}

func generateBooleanPayloads(param *ExtractedParameter) []string {
	return []string{
		"true",
		"false",
		"0",
		"1",
		"yes",
		"no",
		"",
		"null",
		"undefined",
	}
}

func generateArrayPayloads(param *ExtractedParameter) []string {
	return []string{
		"[]",
		"[1,2,3]",
		"[\"a\",\"b\",\"c\"]",
		"[true,false]",
		"[null]",
		"[1,\"a\",true,null]",
		"",
		"not_an_array",
	}
}

func generateObjectPayloads(param *ExtractedParameter) []string {
	return []string{
		"{}",
		"{\"key\":\"value\"}",
		"{\"key\":1}",
		"{\"key\":true}",
		"{\"key\":null}",
		"{\"key\":[1,2,3]}",
		"{\"key\":{\"nested\":\"value\"}}",
		"",
		"not_an_object",
	}
}

func generateDefaultPayloads(param *ExtractedParameter) []string {
	// Generic payloads for any parameter type
	payloads := []string{
		"",
		"null",
		"undefined",
		"test",
		"12345",
		"true",
		"false",
		"[]",
		"{}",
		"!@#$%^&*()",
	}

	// Add example value if available
	if param.Example != nil {
		payloads = append(payloads, fmt.Sprintf("%v", param.Example))
	}

	return payloads
}