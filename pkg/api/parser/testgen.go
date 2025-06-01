// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// APITestGenerator provides methods for generating test cases from API specifications
type APITestGenerator struct {
	// Discovery used for generating test cases
	Discovery *APIEndpointDiscovery
	// Parameter extractor used for generating test cases
	Extractor *APIParameterExtractor
	// Generated test cases
	TestCases []*APITestCase
	// Test case templates
	Templates []*APITestCaseTemplate
	// Test case generation options
	Options *APITestGenerationOptions
}

// APITestCase represents a test case for an API endpoint
type APITestCase struct {
	// Name of the test case
	Name string
	// Description of the test case
	Description string
	// HTTP method (GET, POST, etc.)
	Method string
	// URL of the endpoint
	URL string
	// Path of the endpoint
	Path string
	// Headers for the request
	Headers map[string]string
	// Query parameters for the request
	QueryParams map[string]string
	// Path parameters for the request
	PathParams map[string]string
	// Body for the request
	Body string
	// Expected status code
	ExpectedStatus int
	// Expected content type
	ExpectedContentType string
	// Expected response body (partial match)
	ExpectedResponseBody string
	// Test case category (e.g., "positive", "negative", "security")
	Category string
	// Test case priority (1-5, where 1 is highest)
	Priority int
	// Whether the test case requires authentication
	RequiresAuth bool
	// Authentication details
	Auth *APITestAuth
	// Dependencies on other test cases
	Dependencies []*APITestCase
	// Test case template used to generate this test case
	Template *APITestCaseTemplate
}

// APITestAuth represents authentication details for a test case
type APITestAuth struct {
	// Type of authentication (e.g., "basic", "bearer", "oauth")
	Type string
	// Username for basic authentication
	Username string
	// Password for basic authentication
	Password string
	// Token for bearer authentication
	Token string
	// OAuth client ID
	ClientID string
	// OAuth client secret
	ClientSecret string
	// OAuth scope
	Scope string
	// OAuth token URL
	TokenURL string
}

// APITestCaseTemplate represents a template for generating test cases
type APITestCaseTemplate struct {
	// Name of the template
	Name string
	// Description of the template
	Description string
	// Category of the template
	Category string
	// Priority of the template
	Priority int
	// HTTP method pattern (e.g., "GET", "POST", "*")
	MethodPattern string
	// Path pattern (e.g., "/api/users/*", "/api/*/items")
	PathPattern string
	// Parameter patterns (e.g., "id", "name", "*")
	ParamPatterns []string
	// Expected status code
	ExpectedStatus int
	// Test case generator function
	Generator func(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase
}

// APITestGenerationOptions represents options for test case generation
type APITestGenerationOptions struct {
	// Whether to generate positive test cases
	GeneratePositive bool
	// Whether to generate negative test cases
	GenerateNegative bool
	// Whether to generate security test cases
	GenerateSecurity bool
	// Whether to generate performance test cases
	GeneratePerformance bool
	// Whether to generate test cases for endpoints that require authentication
	IncludeAuthEndpoints bool
	// Base URL for the API
	BaseURL string
	// Authentication details for test cases
	Auth *APITestAuth
	// Maximum number of test cases to generate per endpoint
	MaxTestCasesPerEndpoint int
	// Minimum priority of test cases to generate (1-5, where 1 is highest)
	MinPriority int
	// Maximum priority of test cases to generate (1-5, where 1 is highest)
	MaxPriority int
}

// NewAPITestGenerator creates a new APITestGenerator
func NewAPITestGenerator(discovery *APIEndpointDiscovery, extractor *APIParameterExtractor) *APITestGenerator {
	return &APITestGenerator{
		Discovery:  discovery,
		Extractor:  extractor,
		TestCases:  make([]*APITestCase, 0),
		Templates:  make([]*APITestCaseTemplate, 0),
		Options:    getDefaultTestGenerationOptions(),
	}
}

// getDefaultTestGenerationOptions returns default test generation options
func getDefaultTestGenerationOptions() *APITestGenerationOptions {
	return &APITestGenerationOptions{
		GeneratePositive:        true,
		GenerateNegative:        true,
		GenerateSecurity:        true,
		GeneratePerformance:     false,
		IncludeAuthEndpoints:    true,
		MaxTestCasesPerEndpoint: 10,
		MinPriority:             1,
		MaxPriority:             3,
	}
}

// SetOptions sets the test generation options
func (g *APITestGenerator) SetOptions(options *APITestGenerationOptions) {
	g.Options = options
}

// AddTemplate adds a test case template
func (g *APITestGenerator) AddTemplate(template *APITestCaseTemplate) {
	g.Templates = append(g.Templates, template)
}

// AddDefaultTemplates adds default test case templates
func (g *APITestGenerator) AddDefaultTemplates() {
	// Positive test case templates
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "Valid Request",
		Description:   "Test with valid parameters",
		Category:      "positive",
		Priority:      1,
		MethodPattern: "*",
		PathPattern:   "*",
		ExpectedStatus: 200,
		Generator:     generateValidRequestTestCases,
	})

	// Negative test case templates
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "Missing Required Parameters",
		Description:   "Test with missing required parameters",
		Category:      "negative",
		Priority:      2,
		MethodPattern: "*",
		PathPattern:   "*",
		ExpectedStatus: 400,
		Generator:     generateMissingRequiredParamsTestCases,
	})
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "Invalid Parameter Types",
		Description:   "Test with invalid parameter types",
		Category:      "negative",
		Priority:      2,
		MethodPattern: "*",
		PathPattern:   "*",
		ExpectedStatus: 400,
		Generator:     generateInvalidParamTypesTestCases,
	})

	// Security test case templates
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "SQL Injection",
		Description:   "Test for SQL injection vulnerabilities",
		Category:      "security",
		Priority:      1,
		MethodPattern: "*",
		PathPattern:   "*",
		ParamPatterns: []string{"id", "name", "email", "username", "password", "query", "search", "filter"},
		ExpectedStatus: 400,
		Generator:     generateSQLInjectionTestCases,
	})
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "XSS",
		Description:   "Test for Cross-Site Scripting vulnerabilities",
		Category:      "security",
		Priority:      1,
		MethodPattern: "*",
		PathPattern:   "*",
		ParamPatterns: []string{"name", "description", "title", "comment", "message", "content", "text"},
		ExpectedStatus: 400,
		Generator:     generateXSSTestCases,
	})
	g.AddTemplate(&APITestCaseTemplate{
		Name:          "Authentication Bypass",
		Description:   "Test for authentication bypass vulnerabilities",
		Category:      "security",
		Priority:      1,
		MethodPattern: "*",
		PathPattern:   "*",
		ExpectedStatus: 401,
		Generator:     generateAuthBypassTestCases,
	})
}

// GenerateTestCases generates test cases from the discovered endpoints
func (g *APITestGenerator) GenerateTestCases() error {
	// Check if discovery is available
	if g.Discovery == nil {
		return api.NewAPIError("No discovery provided", 0)
	}

	// Check if endpoints are available
	endpoints := g.Discovery.GetEndpoints()
	if len(endpoints) == 0 {
		return api.NewAPIError("No endpoints discovered", 0)
	}

	// Check if extractor is available
	if g.Extractor == nil {
		// Create a new extractor if not provided
		g.Extractor = NewAPIParameterExtractor(g.Discovery)
		if err := g.Extractor.ExtractParameters(); err != nil {
			return err
		}
	}

	// Check if templates are available
	if len(g.Templates) == 0 {
		g.AddDefaultTemplates()
	}

	// Generate test cases for each endpoint
	for _, endpoint := range endpoints {
		// Skip endpoints that require authentication if not included
		if endpoint.RequiresAuth && !g.Options.IncludeAuthEndpoints {
			continue
		}

		// Get parameters for the endpoint
		params := make([]*ExtractedParameter, 0)
		for _, param := range endpoint.Parameters {
			extractedParam := g.Extractor.GetParameterByName(param.Name)
			if extractedParam != nil {
				params = append(params, extractedParam)
			}
		}

		// Generate test cases for each template
		for _, template := range g.Templates {
			// Skip templates that don't match the endpoint
			if !matchesTemplate(endpoint, template) {
				continue
			}

			// Skip templates based on category options
			if (template.Category == "positive" && !g.Options.GeneratePositive) ||
				(template.Category == "negative" && !g.Options.GenerateNegative) ||
				(template.Category == "security" && !g.Options.GenerateSecurity) ||
				(template.Category == "performance" && !g.Options.GeneratePerformance) {
				continue
			}

			// Skip templates based on priority options
			if template.Priority < g.Options.MinPriority || template.Priority > g.Options.MaxPriority {
				continue
			}

			// Generate test cases using the template
			testCases := template.Generator(endpoint, params)

			// Add the test cases to the list
			for _, testCase := range testCases {
				// Set the template
				testCase.Template = template

				// Set the base URL if not set
				if testCase.URL == "" && g.Options.BaseURL != "" {
					baseURL := strings.TrimSuffix(g.Options.BaseURL, "/")
					path := testCase.Path
					if !strings.HasPrefix(path, "/") {
						path = "/" + path
					}
					testCase.URL = baseURL + path
				}

				// Set authentication details if required
				if testCase.RequiresAuth && testCase.Auth == nil && g.Options.Auth != nil {
					testCase.Auth = g.Options.Auth
				}

				// Add the test case to the list
				g.TestCases = append(g.TestCases, testCase)
			}

			// Limit the number of test cases per endpoint
			if len(g.TestCases) >= g.Options.MaxTestCasesPerEndpoint {
				break
			}
		}
	}

	return nil
}

// GetTestCases returns all generated test cases
func (g *APITestGenerator) GetTestCases() []*APITestCase {
	return g.TestCases
}

// GetTestCasesByCategory returns all test cases with the specified category
func (g *APITestGenerator) GetTestCasesByCategory(category string) []*APITestCase {
	category = strings.ToLower(category)
	testCases := make([]*APITestCase, 0)
	for _, testCase := range g.TestCases {
		if strings.ToLower(testCase.Category) == category {
			testCases = append(testCases, testCase)
		}
	}
	return testCases
}

// GetTestCasesByPriority returns all test cases with the specified priority
func (g *APITestGenerator) GetTestCasesByPriority(priority int) []*APITestCase {
	testCases := make([]*APITestCase, 0)
	for _, testCase := range g.TestCases {
		if testCase.Priority == priority {
			testCases = append(testCases, testCase)
		}
	}
	return testCases
}

// GetTestCasesByEndpoint returns all test cases for the specified endpoint
func (g *APITestGenerator) GetTestCasesByEndpoint(path string, method string) []*APITestCase {
	method = strings.ToUpper(method)
	testCases := make([]*APITestCase, 0)
	for _, testCase := range g.TestCases {
		if testCase.Path == path && testCase.Method == method {
			testCases = append(testCases, testCase)
		}
	}
	return testCases
}

// GenerateTestCasesFromOpenAPI generates test cases from an OpenAPI/Swagger specification
func (g *APITestGenerator) GenerateTestCasesFromOpenAPI(specPath string) error {
	// Create a new discovery if not provided
	if g.Discovery == nil {
		g.Discovery = NewAPIEndpointDiscovery("")
	}

	// Discover endpoints from the OpenAPI specification
	if err := g.Discovery.DiscoverFromOpenAPI(specPath); err != nil {
		return err
	}

	// Create a new extractor if not provided
	if g.Extractor == nil {
		g.Extractor = NewAPIParameterExtractor(g.Discovery)
	}

	// Extract parameters from the discovered endpoints
	if err := g.Extractor.ExtractParameters(); err != nil {
		return err
	}

	// Set base URL from the discovery if not already set
	if g.Options.BaseURL == "" && g.Discovery.BaseURL != "" {
		g.Options.BaseURL = g.Discovery.BaseURL
	}

	// Add default templates if none are provided
	if len(g.Templates) == 0 {
		g.AddDefaultTemplates()
	}

	// Generate test cases based on the OpenAPI specification
	return g.GenerateTestCases()
}

// ExportTestCasesToJSON exports the test cases to a JSON string
func (g *APITestGenerator) ExportTestCasesToJSON() (string, error) {
	// Create a simplified representation of the test cases for export
	type ExportedTestCase struct {
		Name                string            `json:"name"`
		Description         string            `json:"description"`
		Method              string            `json:"method"`
		URL                 string            `json:"url"`
		Path                string            `json:"path"`
		Headers             map[string]string `json:"headers,omitempty"`
		QueryParams         map[string]string `json:"query_params,omitempty"`
		PathParams          map[string]string `json:"path_params,omitempty"`
		Body                string            `json:"body,omitempty"`
		ExpectedStatus      int               `json:"expected_status"`
		ExpectedContentType string            `json:"expected_content_type,omitempty"`
		Category            string            `json:"category"`
		Priority            int               `json:"priority"`
		RequiresAuth        bool              `json:"requires_auth"`
	}

	exportedTestCases := make([]ExportedTestCase, 0, len(g.TestCases))
	for _, testCase := range g.TestCases {
		exportedTestCase := ExportedTestCase{
			Name:                testCase.Name,
			Description:         testCase.Description,
			Method:              testCase.Method,
			URL:                 testCase.URL,
			Path:                testCase.Path,
			Headers:             testCase.Headers,
			QueryParams:         testCase.QueryParams,
			PathParams:          testCase.PathParams,
			Body:                testCase.Body,
			ExpectedStatus:      testCase.ExpectedStatus,
			ExpectedContentType: testCase.ExpectedContentType,
			Category:            testCase.Category,
			Priority:            testCase.Priority,
			RequiresAuth:        testCase.RequiresAuth,
		}
		exportedTestCases = append(exportedTestCases, exportedTestCase)
	}

	// Marshal the test cases to JSON
	jsonData, err := json.MarshalIndent(exportedTestCases, "", "  ")
	if err != nil {
		return "", api.NewAPIError(fmt.Sprintf("Failed to marshal test cases to JSON: %s", err.Error()), 0)
	}

	return string(jsonData), nil
}

// ExportTestCasesToCurl exports the test cases to curl commands
func (g *APITestGenerator) ExportTestCasesToCurl() []string {
	commands := make([]string, 0, len(g.TestCases))
	for _, testCase := range g.TestCases {
		// Build the curl command
		command := fmt.Sprintf("curl -X %s", testCase.Method)

		// Add headers
		for key, value := range testCase.Headers {
			command += fmt.Sprintf(" -H '%s: %s'", key, value)
		}

		// Add authentication
		if testCase.RequiresAuth && testCase.Auth != nil {
			switch testCase.Auth.Type {
			case "basic":
				command += fmt.Sprintf(" -u '%s:%s'", testCase.Auth.Username, testCase.Auth.Password)
			case "bearer":
				command += fmt.Sprintf(" -H 'Authorization: Bearer %s'", testCase.Auth.Token)
			}
		}

		// Add body
		if testCase.Body != "" {
			command += fmt.Sprintf(" -d '%s'", testCase.Body)
		}

		// Add URL
		url := testCase.URL
		if url == "" {
			// Build URL from path and query parameters
			url = testCase.Path
			if len(testCase.QueryParams) > 0 {
				url += "?"
				queryParts := make([]string, 0, len(testCase.QueryParams))
				for key, value := range testCase.QueryParams {
					queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, value))
				}
				url += strings.Join(queryParts, "&")
			}
		}
		command += fmt.Sprintf(" '%s'", url)

		commands = append(commands, command)
	}

	return commands
}

// GenerateTestReport generates a report of the test cases
func (g *APITestGenerator) GenerateTestReport() string {
	if len(g.TestCases) == 0 {
		return "No test cases generated"
	}

	report := "# API Test Case Report\n\n"
	report += fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC1123))
	report += fmt.Sprintf("Total test cases: %d\n\n", len(g.TestCases))

	// Add test case categories
	categories := make(map[string]int)
	for _, testCase := range g.TestCases {
		categories[testCase.Category]++
	}
	report += "## Test Case Categories\n\n"
	for category, count := range categories {
		report += fmt.Sprintf("- %s: %d\n", category, count)
	}
	report += "\n"

	// Add test case priorities
	priorities := make(map[int]int)
	for _, testCase := range g.TestCases {
		priorities[testCase.Priority]++
	}
	report += "## Test Case Priorities\n\n"
	for i := 1; i <= 5; i++ {
		if count, ok := priorities[i]; ok {
			report += fmt.Sprintf("- Priority %d: %d\n", i, count)
		}
	}
	report += "\n"

	// Add endpoints
	endpoints := make(map[string]map[string]int)
	for _, testCase := range g.TestCases {
		if _, ok := endpoints[testCase.Path]; !ok {
			endpoints[testCase.Path] = make(map[string]int)
		}
		endpoints[testCase.Path][testCase.Method]++
	}
	report += "## Endpoints\n\n"
	for path, methods := range endpoints {
		report += fmt.Sprintf("### %s\n\n", path)
		for method, count := range methods {
			report += fmt.Sprintf("- %s: %d test cases\n", method, count)
		}
		report += "\n"
	}

	// Add test case details
	report += "## Test Case Details\n\n"
	for i, testCase := range g.TestCases {
		report += fmt.Sprintf("### %d. %s\n\n", i+1, testCase.Name)
		report += fmt.Sprintf("- Description: %s\n", testCase.Description)
		report += fmt.Sprintf("- Method: %s\n", testCase.Method)
		report += fmt.Sprintf("- Path: %s\n", testCase.Path)
		report += fmt.Sprintf("- Category: %s\n", testCase.Category)
		report += fmt.Sprintf("- Priority: %d\n", testCase.Priority)
		report += fmt.Sprintf("- Expected Status: %d\n", testCase.ExpectedStatus)
		if testCase.RequiresAuth {
			report += "- Requires Authentication: Yes\n"
		}
		report += "\n"
	}

	return report
}

// Helper function to check if an endpoint matches a template
func matchesTemplate(endpoint *DiscoveredEndpoint, template *APITestCaseTemplate) bool {
	// Check method pattern
	if template.MethodPattern != "*" && template.MethodPattern != endpoint.Method {
		return false
	}

	// Check path pattern
	if template.PathPattern != "*" && !pathMatches(endpoint.Path, template.PathPattern) {
		return false
	}

	// Check parameter patterns
	if len(template.ParamPatterns) > 0 {
		paramMatch := false
		for _, param := range endpoint.Parameters {
			for _, pattern := range template.ParamPatterns {
				if pattern == "*" || pattern == param.Name {
					paramMatch = true
					break
				}
			}
			if paramMatch {
				break
			}
		}
		if !paramMatch {
			return false
		}
	}

	return true
}

// Template generator functions

// generateValidRequestTestCases generates test cases for valid requests
func generateValidRequestTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// Create a test case with valid parameters
	testCase := &APITestCase{
		Name:           fmt.Sprintf("Valid %s request to %s", endpoint.Method, endpoint.Path),
		Description:    fmt.Sprintf("Test with valid parameters for %s %s", endpoint.Method, endpoint.Path),
		Method:         endpoint.Method,
		Path:           endpoint.Path,
		Headers:        make(map[string]string),
		QueryParams:    make(map[string]string),
		PathParams:     make(map[string]string),
		ExpectedStatus: 200,
		Category:       "positive",
		Priority:       1,
		RequiresAuth:   endpoint.RequiresAuth,
	}

	// Add content type header for POST, PUT, PATCH
	if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
		testCase.Headers["Content-Type"] = "application/json"
	}

	// Add parameters
	bodyParams := make(map[string]interface{})
	for _, param := range params {
		switch param.In {
		case "query":
			testCase.QueryParams[param.Name] = getExampleValue(param)
		case "path":
			testCase.PathParams[param.Name] = getExampleValue(param)
		case "header":
			testCase.Headers[param.Name] = getExampleValue(param)
		case "body":
			bodyParams[param.Name] = getExampleValueAsInterface(param)
		}
	}

	// Add body if there are body parameters
	if len(bodyParams) > 0 {
		bodyJSON, err := json.Marshal(bodyParams)
		if err == nil {
			testCase.Body = string(bodyJSON)
		}
	}

	testCases = append(testCases, testCase)
	return testCases
}

// generateMissingRequiredParamsTestCases generates test cases for missing required parameters
func generateMissingRequiredParamsTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// Find required parameters
	requiredParams := make([]*ExtractedParameter, 0)
	for _, param := range params {
		if param.Required {
			requiredParams = append(requiredParams, param)
		}
	}

	// Generate a test case for each required parameter
	for _, requiredParam := range requiredParams {
		testCase := &APITestCase{
			Name:           fmt.Sprintf("Missing required parameter '%s' for %s %s", requiredParam.Name, endpoint.Method, endpoint.Path),
			Description:    fmt.Sprintf("Test with missing required parameter '%s'", requiredParam.Name),
			Method:         endpoint.Method,
			Path:           endpoint.Path,
			Headers:        make(map[string]string),
			QueryParams:    make(map[string]string),
			PathParams:     make(map[string]string),
			ExpectedStatus: 400,
			Category:       "negative",
			Priority:       2,
			RequiresAuth:   endpoint.RequiresAuth,
		}

		// Add content type header for POST, PUT, PATCH
		if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
			testCase.Headers["Content-Type"] = "application/json"
		}

		// Add all parameters except the required one
		bodyParams := make(map[string]interface{})
		for _, param := range params {
			if param.Name == requiredParam.Name {
				continue
			}

			switch param.In {
			case "query":
				testCase.QueryParams[param.Name] = getExampleValue(param)
			case "path":
				testCase.PathParams[param.Name] = getExampleValue(param)
			case "header":
				testCase.Headers[param.Name] = getExampleValue(param)
			case "body":
				bodyParams[param.Name] = getExampleValueAsInterface(param)
			}
		}

		// Add body if there are body parameters
		if len(bodyParams) > 0 {
			bodyJSON, err := json.Marshal(bodyParams)
			if err == nil {
				testCase.Body = string(bodyJSON)
			}
		}

		testCases = append(testCases, testCase)
	}

	return testCases
}

// generateInvalidParamTypesTestCases generates test cases for invalid parameter types
func generateInvalidParamTypesTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// Generate a test case for each parameter with an invalid type
	for _, param := range params {
		// Skip parameters without a type
		if param.Type == "" {
			continue
		}

		testCase := &APITestCase{
			Name:           fmt.Sprintf("Invalid type for parameter '%s' in %s %s", param.Name, endpoint.Method, endpoint.Path),
			Description:    fmt.Sprintf("Test with invalid type for parameter '%s'", param.Name),
			Method:         endpoint.Method,
			Path:           endpoint.Path,
			Headers:        make(map[string]string),
			QueryParams:    make(map[string]string),
			PathParams:     make(map[string]string),
			ExpectedStatus: 400,
			Category:       "negative",
			Priority:       2,
			RequiresAuth:   endpoint.RequiresAuth,
		}

		// Add content type header for POST, PUT, PATCH
		if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
			testCase.Headers["Content-Type"] = "application/json"
		}

		// Add all parameters with valid values except the one with invalid type
		bodyParams := make(map[string]interface{})
		for _, p := range params {
			if p.Name == param.Name {
				// Add the parameter with an invalid type
				switch p.In {
				case "query":
					testCase.QueryParams[p.Name] = getInvalidTypeValue(p)
				case "path":
					testCase.PathParams[p.Name] = getInvalidTypeValue(p)
				case "header":
					testCase.Headers[p.Name] = getInvalidTypeValue(p)
				case "body":
					bodyParams[p.Name] = getInvalidTypeValueAsInterface(p)
				}
			} else {
				// Add other parameters with valid values
				switch p.In {
				case "query":
					testCase.QueryParams[p.Name] = getExampleValue(p)
				case "path":
					testCase.PathParams[p.Name] = getExampleValue(p)
				case "header":
					testCase.Headers[p.Name] = getExampleValue(p)
				case "body":
					bodyParams[p.Name] = getExampleValueAsInterface(p)
				}
			}
		}

		// Add body if there are body parameters
		if len(bodyParams) > 0 {
			bodyJSON, err := json.Marshal(bodyParams)
			if err == nil {
				testCase.Body = string(bodyJSON)
			}
		}

		testCases = append(testCases, testCase)
	}

	return testCases
}

// generateSQLInjectionTestCases generates test cases for SQL injection
func generateSQLInjectionTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// SQL injection payloads
	sqlInjectionPayloads := []string{
		"' OR '1'='1",
		"' OR '1'='1' --",
		"' OR '1'='1' /*",
		"' UNION SELECT 1,2,3 --",
		"' UNION SELECT table_name,2,3 FROM information_schema.tables --",
	}

	// Generate a test case for each parameter with SQL injection payloads
	for _, param := range params {
		// Skip parameters that are not likely to be vulnerable to SQL injection
		if !isLikelyVulnerableToSQLInjection(param) {
			continue
		}

		for _, payload := range sqlInjectionPayloads {
			testCase := &APITestCase{
				Name:           fmt.Sprintf("SQL Injection in parameter '%s' for %s %s", param.Name, endpoint.Method, endpoint.Path),
				Description:    fmt.Sprintf("Test for SQL injection vulnerability in parameter '%s' with payload: %s", param.Name, payload),
				Method:         endpoint.Method,
				Path:           endpoint.Path,
				Headers:        make(map[string]string),
				QueryParams:    make(map[string]string),
				PathParams:     make(map[string]string),
				ExpectedStatus: 400,
				Category:       "security",
				Priority:       1,
				RequiresAuth:   endpoint.RequiresAuth,
			}

			// Add content type header for POST, PUT, PATCH
			if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
				testCase.Headers["Content-Type"] = "application/json"
			}

			// Add all parameters with valid values except the one with SQL injection payload
			bodyParams := make(map[string]interface{})
			for _, p := range params {
				if p.Name == param.Name {
					// Add the parameter with SQL injection payload
					switch p.In {
					case "query":
						testCase.QueryParams[p.Name] = payload
					case "path":
						testCase.PathParams[p.Name] = payload
					case "header":
						testCase.Headers[p.Name] = payload
					case "body":
						bodyParams[p.Name] = payload
					}
				} else {
					// Add other parameters with valid values
					switch p.In {
					case "query":
						testCase.QueryParams[p.Name] = getExampleValue(p)
					case "path":
						testCase.PathParams[p.Name] = getExampleValue(p)
					case "header":
						testCase.Headers[p.Name] = getExampleValue(p)
					case "body":
						bodyParams[p.Name] = getExampleValueAsInterface(p)
					}
				}
			}

			// Add body if there are body parameters
			if len(bodyParams) > 0 {
				bodyJSON, err := json.Marshal(bodyParams)
				if err == nil {
					testCase.Body = string(bodyJSON)
				}
			}

			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

// generateXSSTestCases generates test cases for Cross-Site Scripting
func generateXSSTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// XSS payloads
	xssPayloads := []string{
		"<script>alert(1)</script>",
		"<img src=x onerror=alert(1)>",
		"<svg onload=alert(1)>",
		"javascript:alert(1)",
		"\"><script>alert(1)</script>",
	}

	// Generate a test case for each parameter with XSS payloads
	for _, param := range params {
		// Skip parameters that are not likely to be vulnerable to XSS
		if !isLikelyVulnerableToXSS(param) {
			continue
		}

		for _, payload := range xssPayloads {
			testCase := &APITestCase{
				Name:           fmt.Sprintf("XSS in parameter '%s' for %s %s", param.Name, endpoint.Method, endpoint.Path),
				Description:    fmt.Sprintf("Test for Cross-Site Scripting vulnerability in parameter '%s' with payload: %s", param.Name, payload),
				Method:         endpoint.Method,
				Path:           endpoint.Path,
				Headers:        make(map[string]string),
				QueryParams:    make(map[string]string),
				PathParams:     make(map[string]string),
				ExpectedStatus: 400,
				Category:       "security",
				Priority:       1,
				RequiresAuth:   endpoint.RequiresAuth,
			}

			// Add content type header for POST, PUT, PATCH
			if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
				testCase.Headers["Content-Type"] = "application/json"
			}

			// Add all parameters with valid values except the one with XSS payload
			bodyParams := make(map[string]interface{})
			for _, p := range params {
				if p.Name == param.Name {
					// Add the parameter with XSS payload
					switch p.In {
					case "query":
						testCase.QueryParams[p.Name] = payload
					case "path":
						testCase.PathParams[p.Name] = payload
					case "header":
						testCase.Headers[p.Name] = payload
					case "body":
						bodyParams[p.Name] = payload
					}
				} else {
					// Add other parameters with valid values
					switch p.In {
					case "query":
						testCase.QueryParams[p.Name] = getExampleValue(p)
					case "path":
						testCase.PathParams[p.Name] = getExampleValue(p)
					case "header":
						testCase.Headers[p.Name] = getExampleValue(p)
					case "body":
						bodyParams[p.Name] = getExampleValueAsInterface(p)
					}
				}
			}

			// Add body if there are body parameters
			if len(bodyParams) > 0 {
				bodyJSON, err := json.Marshal(bodyParams)
				if err == nil {
					testCase.Body = string(bodyJSON)
				}
			}

			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

// generateAuthBypassTestCases generates test cases for authentication bypass
func generateAuthBypassTestCases(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase {
	testCases := make([]*APITestCase, 0)

	// Skip endpoints that don't require authentication
	if !endpoint.RequiresAuth {
		return testCases
	}

	// Create a test case for authentication bypass
	testCase := &APITestCase{
		Name:           fmt.Sprintf("Authentication bypass for %s %s", endpoint.Method, endpoint.Path),
		Description:    "Test for authentication bypass by omitting authentication",
		Method:         endpoint.Method,
		Path:           endpoint.Path,
		Headers:        make(map[string]string),
		QueryParams:    make(map[string]string),
		PathParams:     make(map[string]string),
		ExpectedStatus: 401,
		Category:       "security",
		Priority:       1,
		RequiresAuth:   false,
	}

	// Add content type header for POST, PUT, PATCH
	if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
		testCase.Headers["Content-Type"] = "application/json"
	}

	// Add all parameters with valid values
	bodyParams := make(map[string]interface{})
	for _, p := range params {
		switch p.In {
		case "query":
			testCase.QueryParams[p.Name] = getExampleValue(p)
		case "path":
			testCase.PathParams[p.Name] = getExampleValue(p)
		case "header":
			// Skip authentication headers
			if !isAuthHeader(p.Name) {
				testCase.Headers[p.Name] = getExampleValue(p)
			}
		case "body":
			bodyParams[p.Name] = getExampleValueAsInterface(p)
		}
	}

	// Add body if there are body parameters
	if len(bodyParams) > 0 {
		bodyJSON, err := json.Marshal(bodyParams)
		if err == nil {
			testCase.Body = string(bodyJSON)
		}
	}

	testCases = append(testCases, testCase)
	return testCases
}

// Helper functions

// getExampleValue returns a string example value for a parameter
func getExampleValue(param *ExtractedParameter) string {
	if param.Example != nil {
		return fmt.Sprintf("%v", param.Example)
	}

	// Generate a default value based on the parameter type
	switch strings.ToLower(param.Type) {
	case "string":
		return "test"
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	case "array":
		return "[]"
	case "object":
		return "{}"
	default:
		return "test"
	}
}

// getExampleValueAsInterface returns an interface{} example value for a parameter
func getExampleValueAsInterface(param *ExtractedParameter) interface{} {
	if param.Example != nil {
		return param.Example
	}

	// Generate a default value based on the parameter type
	switch strings.ToLower(param.Type) {
	case "string":
		return "test"
	case "integer":
		return 1
	case "number":
		return 1.0
	case "boolean":
		return true
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	default:
		return "test"
	}
}

// getInvalidTypeValue returns a string value with an invalid type for a parameter
func getInvalidTypeValue(param *ExtractedParameter) string {
	// Generate an invalid value based on the parameter type
	switch strings.ToLower(param.Type) {
	case "string":
		return "123" // Not invalid, but could be if string has a specific format
	case "integer", "number":
		return "not_a_number"
	case "boolean":
		return "not_a_boolean"
	case "array":
		return "not_an_array"
	case "object":
		return "not_an_object"
	default:
		return "invalid"
	}
}

// getInvalidTypeValueAsInterface returns an interface{} value with an invalid type for a parameter
func getInvalidTypeValueAsInterface(param *ExtractedParameter) interface{} {
	// Generate an invalid value based on the parameter type
	switch strings.ToLower(param.Type) {
	case "string":
		return 123 // Number instead of string
	case "integer", "number":
		return "not_a_number" // String instead of number
	case "boolean":
		return "not_a_boolean" // String instead of boolean
	case "array":
		return "not_an_array" // String instead of array
	case "object":
		return "not_an_object" // String instead of object
	default:
		return "invalid"
	}
}

// isLikelyVulnerableToSQLInjection checks if a parameter is likely to be vulnerable to SQL injection
func isLikelyVulnerableToSQLInjection(param *ExtractedParameter) bool {
	// Check parameter name
	name := strings.ToLower(param.Name)
	vulnerableNames := []string{"id", "user", "name", "email", "username", "password", "query", "search", "filter", "where", "order", "sort", "limit", "offset"}
	for _, vulnName := range vulnerableNames {
		if strings.Contains(name, vulnName) {
			return true
		}
	}

	// Check parameter type
	if strings.ToLower(param.Type) == "string" {
		return true
	}

	return false
}

// isLikelyVulnerableToXSS checks if a parameter is likely to be vulnerable to XSS
func isLikelyVulnerableToXSS(param *ExtractedParameter) bool {
	// Check parameter name
	name := strings.ToLower(param.Name)
	vulnerableNames := []string{"name", "description", "title", "comment", "message", "content", "text", "html", "body", "data"}
	for _, vulnName := range vulnerableNames {
		if strings.Contains(name, vulnName) {
			return true
		}
	}

	// Check parameter type
	if strings.ToLower(param.Type) == "string" {
		return true
	}

	return false
}

// isAuthHeader checks if a header name is related to authentication
func isAuthHeader(name string) bool {
	name = strings.ToLower(name)
	authHeaders := []string{"authorization", "auth", "token", "api-key", "apikey", "x-api-key", "x-auth", "x-token"}
	for _, authHeader := range authHeaders {
		if name == authHeader {
			return true
		}
	}
	return false
}
