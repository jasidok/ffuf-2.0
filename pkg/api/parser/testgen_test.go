package parser

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAPITestGenerator_GenerateTestCases(t *testing.T) {
	// Create a discovery with some endpoints
	discovery := NewAPIEndpointDiscovery("https://api.example.com")
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
			RequiresAuth: false,
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
			RequiresAuth: true,
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
			},
			RequiresAuth: false,
		},
	}

	// Create a parameter extractor
	extractor := NewAPIParameterExtractor(discovery)
	err := extractor.ExtractParameters()
	if err != nil {
		t.Fatalf("Failed to extract parameters: %v", err)
	}

	// Create a test generator
	generator := NewAPITestGenerator(discovery, extractor)

	// Generate test cases
	err = generator.GenerateTestCases()
	if err != nil {
		t.Fatalf("Failed to generate test cases: %v", err)
	}

	// Check the generated test cases
	testCases := generator.GetTestCases()
	if len(testCases) == 0 {
		t.Errorf("Expected test cases to be generated")
	}

	// Check test case categories
	positiveTestCases := generator.GetTestCasesByCategory("positive")
	if len(positiveTestCases) == 0 {
		t.Errorf("Expected positive test cases to be generated")
	}

	negativeTestCases := generator.GetTestCasesByCategory("negative")
	if len(negativeTestCases) == 0 {
		t.Errorf("Expected negative test cases to be generated")
	}

	securityTestCases := generator.GetTestCasesByCategory("security")
	if len(securityTestCases) == 0 {
		t.Errorf("Expected security test cases to be generated")
	}

	// Check test case priorities
	priority1TestCases := generator.GetTestCasesByPriority(1)
	if len(priority1TestCases) == 0 {
		t.Errorf("Expected priority 1 test cases to be generated")
	}

	// Check test cases by endpoint
	getUserTestCases := generator.GetTestCasesByEndpoint("/api/users", "GET")
	if len(getUserTestCases) == 0 {
		t.Errorf("Expected test cases for GET /api/users to be generated")
	}

	postUserTestCases := generator.GetTestCasesByEndpoint("/api/users", "POST")
	if len(postUserTestCases) == 0 {
		t.Errorf("Expected test cases for POST /api/users to be generated")
	}

	getUserByIdTestCases := generator.GetTestCasesByEndpoint("/api/users/{id}", "GET")
	if len(getUserByIdTestCases) == 0 {
		t.Errorf("Expected test cases for GET /api/users/{id} to be generated")
	}

	// Check test case details
	for _, testCase := range testCases {
		// Check that the test case has a name
		if testCase.Name == "" {
			t.Errorf("Test case has no name")
		}

		// Check that the test case has a description
		if testCase.Description == "" {
			t.Errorf("Test case has no description")
		}

		// Check that the test case has a method
		if testCase.Method == "" {
			t.Errorf("Test case has no method")
		}

		// Check that the test case has a path
		if testCase.Path == "" {
			t.Errorf("Test case has no path")
		}

		// Check that the test case has a category
		if testCase.Category == "" {
			t.Errorf("Test case has no category")
		}

		// Check that the test case has a priority
		if testCase.Priority == 0 {
			t.Errorf("Test case has no priority")
		}

		// Check that the test case has an expected status
		if testCase.ExpectedStatus == 0 {
			t.Errorf("Test case has no expected status")
		}

		// Check that POST test cases have a Content-Type header
		if testCase.Method == "POST" && testCase.Headers["Content-Type"] == "" {
			t.Errorf("POST test case has no Content-Type header")
		}

		// Check that test cases for endpoints with required parameters have those parameters
		if testCase.Path == "/api/users/{id}" && testCase.Category == "positive" {
			if testCase.PathParams["id"] == "" {
				t.Errorf("Test case for GET /api/users/{id} is missing required path parameter 'id'")
			}
		}

		// Check that test cases for endpoints with body parameters have a body
		if testCase.Method == "POST" && testCase.Category == "positive" {
			if testCase.Body == "" {
				t.Errorf("Test case for POST /api/users is missing body")
			}
		}
	}
}

func TestAPITestGenerator_ExportTestCasesToJSON(t *testing.T) {
	// Create a test generator with some test cases
	generator := createTestGenerator(t)

	// Export test cases to JSON
	jsonStr, err := generator.ExportTestCasesToJSON()
	if err != nil {
		t.Fatalf("Failed to export test cases to JSON: %v", err)
	}

	// Check that the JSON is valid
	var testCases []map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &testCases)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check that the JSON contains the expected number of test cases
	if len(testCases) != len(generator.TestCases) {
		t.Errorf("Expected %d test cases in JSON, got %d", len(generator.TestCases), len(testCases))
	}

	// Check that the JSON contains the expected fields
	for _, testCase := range testCases {
		if testCase["name"] == nil {
			t.Errorf("Test case in JSON is missing 'name' field")
		}
		if testCase["description"] == nil {
			t.Errorf("Test case in JSON is missing 'description' field")
		}
		if testCase["method"] == nil {
			t.Errorf("Test case in JSON is missing 'method' field")
		}
		if testCase["path"] == nil {
			t.Errorf("Test case in JSON is missing 'path' field")
		}
		if testCase["category"] == nil {
			t.Errorf("Test case in JSON is missing 'category' field")
		}
		if testCase["priority"] == nil {
			t.Errorf("Test case in JSON is missing 'priority' field")
		}
		if testCase["expected_status"] == nil {
			t.Errorf("Test case in JSON is missing 'expected_status' field")
		}
	}
}

func TestAPITestGenerator_ExportTestCasesToCurl(t *testing.T) {
	// Create a test generator with some test cases
	generator := createTestGenerator(t)

	// Export test cases to curl commands
	commands := generator.ExportTestCasesToCurl()

	// Check that the commands are generated
	if len(commands) != len(generator.TestCases) {
		t.Errorf("Expected %d curl commands, got %d", len(generator.TestCases), len(commands))
	}

	// Check that the commands contain the expected elements
	for _, command := range commands {
		if !strings.HasPrefix(command, "curl -X ") {
			t.Errorf("Curl command does not start with 'curl -X ': %s", command)
		}
		if !strings.Contains(command, " '") {
			t.Errorf("Curl command does not contain a URL: %s", command)
		}
	}
}

func TestAPITestGenerator_GenerateTestReport(t *testing.T) {
	// Create a test generator with some test cases
	generator := createTestGenerator(t)

	// Generate test report
	report := generator.GenerateTestReport()

	// Check that the report is generated
	if report == "" {
		t.Errorf("Expected test report to be generated")
	}

	// Check that the report contains the expected sections
	if !strings.Contains(report, "# API Test Case Report") {
		t.Errorf("Test report does not contain title")
	}
	if !strings.Contains(report, "Total test cases:") {
		t.Errorf("Test report does not contain total test cases")
	}
	if !strings.Contains(report, "## Test Case Categories") {
		t.Errorf("Test report does not contain categories section")
	}
	if !strings.Contains(report, "## Test Case Priorities") {
		t.Errorf("Test report does not contain priorities section")
	}
	if !strings.Contains(report, "## Endpoints") {
		t.Errorf("Test report does not contain endpoints section")
	}
	if !strings.Contains(report, "## Test Case Details") {
		t.Errorf("Test report does not contain details section")
	}
}

func TestAPITestGenerator_SetOptions(t *testing.T) {
	// Create a test generator
	generator := NewAPITestGenerator(nil, nil)

	// Set options
	options := &APITestGenerationOptions{
		GeneratePositive:        false,
		GenerateNegative:        true,
		GenerateSecurity:        false,
		GeneratePerformance:     true,
		IncludeAuthEndpoints:    false,
		MaxTestCasesPerEndpoint: 5,
		MinPriority:             2,
		MaxPriority:             4,
		BaseURL:                 "https://api.example.org",
		Auth: &APITestAuth{
			Type:     "bearer",
			Token:    "test-token",
		},
	}
	generator.SetOptions(options)

	// Check that the options are set
	if generator.Options.GeneratePositive != false {
		t.Errorf("Expected GeneratePositive to be false")
	}
	if generator.Options.GenerateNegative != true {
		t.Errorf("Expected GenerateNegative to be true")
	}
	if generator.Options.GenerateSecurity != false {
		t.Errorf("Expected GenerateSecurity to be false")
	}
	if generator.Options.GeneratePerformance != true {
		t.Errorf("Expected GeneratePerformance to be true")
	}
	if generator.Options.IncludeAuthEndpoints != false {
		t.Errorf("Expected IncludeAuthEndpoints to be false")
	}
	if generator.Options.MaxTestCasesPerEndpoint != 5 {
		t.Errorf("Expected MaxTestCasesPerEndpoint to be 5")
	}
	if generator.Options.MinPriority != 2 {
		t.Errorf("Expected MinPriority to be 2")
	}
	if generator.Options.MaxPriority != 4 {
		t.Errorf("Expected MaxPriority to be 4")
	}
	if generator.Options.BaseURL != "https://api.example.org" {
		t.Errorf("Expected BaseURL to be 'https://api.example.org'")
	}
	if generator.Options.Auth.Type != "bearer" {
		t.Errorf("Expected Auth.Type to be 'bearer'")
	}
	if generator.Options.Auth.Token != "test-token" {
		t.Errorf("Expected Auth.Token to be 'test-token'")
	}
}

func TestAPITestGenerator_AddTemplate(t *testing.T) {
	// Create a test generator
	generator := NewAPITestGenerator(nil, nil)

	// Add a template
	template := &APITestCaseTemplate{
		Name:           "Test Template",
		Description:    "Test template description",
		Category:       "test",
		Priority:       3,
		MethodPattern:  "GET",
		PathPattern:    "/api/test/*",
		ParamPatterns:  []string{"id", "name"},
		ExpectedStatus: 200,
		Generator:      func(endpoint *DiscoveredEndpoint, params []*ExtractedParameter) []*APITestCase { return nil },
	}
	generator.AddTemplate(template)

	// Check that the template is added
	if len(generator.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(generator.Templates))
	}
	if generator.Templates[0] != template {
		t.Errorf("Expected template to be added")
	}
}

func TestAPITestGenerator_AddDefaultTemplates(t *testing.T) {
	// Create a test generator
	generator := NewAPITestGenerator(nil, nil)

	// Add default templates
	generator.AddDefaultTemplates()

	// Check that the templates are added
	if len(generator.Templates) == 0 {
		t.Errorf("Expected templates to be added")
	}

	// Check that the templates include positive, negative, and security templates
	hasPositive := false
	hasNegative := false
	hasSecurity := false
	for _, template := range generator.Templates {
		if template.Category == "positive" {
			hasPositive = true
		}
		if template.Category == "negative" {
			hasNegative = true
		}
		if template.Category == "security" {
			hasSecurity = true
		}
	}
	if !hasPositive {
		t.Errorf("Expected positive templates to be added")
	}
	if !hasNegative {
		t.Errorf("Expected negative templates to be added")
	}
	if !hasSecurity {
		t.Errorf("Expected security templates to be added")
	}
}

func TestMatchesTemplate(t *testing.T) {
	// Create an endpoint
	endpoint := &DiscoveredEndpoint{
		Path:   "/api/users/{id}",
		Method: "GET",
		Parameters: []*DiscoveredParameter{
			{
				Name: "id",
				In:   "path",
			},
			{
				Name: "fields",
				In:   "query",
			},
		},
	}

	// Test cases
	testCases := []struct {
		template *APITestCaseTemplate
		matches  bool
	}{
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users/{id}",
			},
			matches: true,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "POST",
				PathPattern:   "/api/users/{id}",
			},
			matches: false,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users",
			},
			matches: false,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "*",
				PathPattern:   "*",
			},
			matches: true,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users/*",
			},
			matches: true,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/*/id}",
			},
			matches: true,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users/{id}",
				ParamPatterns: []string{"id"},
			},
			matches: true,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users/{id}",
				ParamPatterns: []string{"name"},
			},
			matches: false,
		},
		{
			template: &APITestCaseTemplate{
				MethodPattern: "GET",
				PathPattern:   "/api/users/{id}",
				ParamPatterns: []string{"*"},
			},
			matches: true,
		},
	}

	for i, tc := range testCases {
		result := matchesTemplate(endpoint, tc.template)
		if result != tc.matches {
			t.Errorf("Test case %d: expected %v, got %v", i, tc.matches, result)
		}
	}
}

func TestGenerateValidRequestTestCases(t *testing.T) {
	// Create an endpoint
	endpoint := &DiscoveredEndpoint{
		Path:   "/api/users",
		Method: "GET",
		Parameters: []*DiscoveredParameter{
			{
				Name:     "limit",
				In:       "query",
				Required: false,
				Type:     "integer",
				Example:  10,
			},
		},
	}

	// Create parameters
	params := []*ExtractedParameter{
		{
			Name:     "limit",
			In:       "query",
			Required: false,
			Type:     "integer",
			Example:  10,
		},
	}

	// Generate test cases
	testCases := generateValidRequestTestCases(endpoint, params)

	// Check that test cases are generated
	if len(testCases) == 0 {
		t.Errorf("Expected test cases to be generated")
	}

	// Check the test case details
	testCase := testCases[0]
	if testCase.Name == "" {
		t.Errorf("Test case has no name")
	}
	if testCase.Description == "" {
		t.Errorf("Test case has no description")
	}
	if testCase.Method != "GET" {
		t.Errorf("Expected method to be GET, got %s", testCase.Method)
	}
	if testCase.Path != "/api/users" {
		t.Errorf("Expected path to be /api/users, got %s", testCase.Path)
	}
	if testCase.Category != "positive" {
		t.Errorf("Expected category to be positive, got %s", testCase.Category)
	}
	if testCase.Priority != 1 {
		t.Errorf("Expected priority to be 1, got %d", testCase.Priority)
	}
	if testCase.ExpectedStatus != 200 {
		t.Errorf("Expected expected status to be 200, got %d", testCase.ExpectedStatus)
	}
	if testCase.QueryParams["limit"] != "10" {
		t.Errorf("Expected query parameter 'limit' to be '10', got '%s'", testCase.QueryParams["limit"])
	}
}

// Helper function to create a test generator with some test cases
func createTestGenerator(t *testing.T) *APITestGenerator {
	// Create a discovery with some endpoints
	discovery := NewAPIEndpointDiscovery("https://api.example.com")
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
			},
			RequiresAuth: false,
		},
	}

	// Create a parameter extractor
	extractor := NewAPIParameterExtractor(discovery)
	err := extractor.ExtractParameters()
	if err != nil {
		t.Fatalf("Failed to extract parameters: %v", err)
	}

	// Create a test generator
	generator := NewAPITestGenerator(discovery, extractor)

	// Generate test cases
	err = generator.GenerateTestCases()
	if err != nil {
		t.Fatalf("Failed to generate test cases: %v", err)
	}

	return generator
}

// TestAPITestGenerator_GenerateTestCasesFromOpenAPI tests generating test cases from an OpenAPI specification
func TestAPITestGenerator_GenerateTestCasesFromOpenAPI(t *testing.T) {
	// Skip this test if running in CI environment as it requires a file
	// This is just a placeholder - in a real implementation, you would use a test fixture
	t.Skip("Skipping test that requires an OpenAPI specification file")

	// Create a test generator
	generator := NewAPITestGenerator(nil, nil)

	// Set options
	generator.SetOptions(&APITestGenerationOptions{
		GeneratePositive:        true,
		GenerateNegative:        true,
		GenerateSecurity:        true,
		MaxTestCasesPerEndpoint: 10,
	})

	// Generate test cases from an OpenAPI specification
	// In a real test, you would use a test fixture file
	err := generator.GenerateTestCasesFromOpenAPI("./testdata/openapi.json")
	if err != nil {
		t.Fatalf("Failed to generate test cases from OpenAPI: %v", err)
	}

	// Check that test cases were generated
	testCases := generator.GetTestCases()
	if len(testCases) == 0 {
		t.Errorf("Expected test cases to be generated from OpenAPI specification")
	}

	// Check that the test cases have the expected properties
	for _, testCase := range testCases {
		// Check that the test case has a name
		if testCase.Name == "" {
			t.Errorf("Test case has no name")
		}

		// Check that the test case has a method
		if testCase.Method == "" {
			t.Errorf("Test case has no method")
		}

		// Check that the test case has a path
		if testCase.Path == "" {
			t.Errorf("Test case has no path")
		}

		// Check that the test case has a category
		if testCase.Category == "" {
			t.Errorf("Test case has no category")
		}

		// Check that the test case has an expected status
		if testCase.ExpectedStatus == 0 {
			t.Errorf("Test case has no expected status")
		}
	}
}
