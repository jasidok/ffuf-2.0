package reporting

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api/parser"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestNewCoverageAnalyzer(t *testing.T) {
	// Test with nil options
	analyzer := NewCoverageAnalyzer(nil)
	if analyzer == nil {
		t.Error("Expected non-nil analyzer with nil options")
	}
	if analyzer.options == nil {
		t.Error("Expected non-nil options in analyzer")
	}
	if analyzer.options.Format != FormatHTML {
		t.Errorf("Expected default format HTML, got %s", analyzer.options.Format)
	}

	// Test with custom options
	options := &CoverageOptions{
		IncludeUntested: false,
		DetailLevel:     3,
		GroupByTags:     false,
		Format:          FormatJSON,
		OutputFile:      "test.json",
	}
	analyzer = NewCoverageAnalyzer(options)
	if analyzer.options != options {
		t.Error("Expected analyzer to use provided options")
	}
}

func TestRecordTest(t *testing.T) {
	analyzer := NewCoverageAnalyzer(nil)
	
	// Record a test for a non-existent endpoint
	analyzer.RecordTest("GET", "/api/test", nil, []string{})
	
	// Check that the endpoint was created
	key := "GET /api/test"
	endpoint, exists := analyzer.endpoints[key]
	if !exists {
		t.Error("Expected endpoint to be created")
	}
	if endpoint.Path != "/api/test" {
		t.Errorf("Expected path /api/test, got %s", endpoint.Path)
	}
	if endpoint.Method != "GET" {
		t.Errorf("Expected method GET, got %s", endpoint.Method)
	}
	if endpoint.TestCount != 1 {
		t.Errorf("Expected test count 1, got %d", endpoint.TestCount)
	}
	
	// Record another test for the same endpoint
	resp := &ffuf.Response{
		StatusCode: 200,
	}
	analyzer.RecordTest("GET", "/api/test", resp, []string{})
	
	// Check that the test count was incremented
	endpoint = analyzer.endpoints[key]
	if endpoint.TestCount != 2 {
		t.Errorf("Expected test count 2, got %d", endpoint.TestCount)
	}
	if endpoint.ResponseStatus != 200 {
		t.Errorf("Expected response status 200, got %d", endpoint.ResponseStatus)
	}
	
	// Record a test with an error response
	resp = &ffuf.Response{
		StatusCode: 404,
	}
	analyzer.RecordTest("GET", "/api/test", resp, []string{})
	
	// Check that the error count was incremented
	endpoint = analyzer.endpoints[key]
	if endpoint.ErrorCount != 1 {
		t.Errorf("Expected error count 1, got %d", endpoint.ErrorCount)
	}
	if endpoint.Status != StatusError {
		t.Errorf("Expected status error, got %s", endpoint.Status)
	}
}

func TestImportFromDiscovery(t *testing.T) {
	// Create a mock discovery with some endpoints
	discovery := &parser.APIEndpointDiscovery{}
	
	// Create a coverage analyzer
	analyzer := NewCoverageAnalyzer(nil)
	
	// Import endpoints from discovery
	analyzer.ImportFromDiscovery(discovery)
	
	// Check that the discovery was set
	if analyzer.discovery != discovery {
		t.Error("Expected discovery to be set")
	}
}

func TestGetCoverageStats(t *testing.T) {
	analyzer := NewCoverageAnalyzer(nil)
	
	// Add some endpoints with different statuses
	analyzer.endpoints = map[string]*EndpointCoverage{
		"GET /api/test1": {
			Path:       "/api/test1",
			Method:     "GET",
			Status:     StatusTested,
			Parameters: []ParameterCoverage{},
			TestCount:  1,
		},
		"POST /api/test2": {
			Path:       "/api/test2",
			Method:     "POST",
			Status:     StatusPartial,
			Parameters: []ParameterCoverage{
				{Name: "param1", Tested: true},
				{Name: "param2", Tested: false},
			},
			TestCount:  1,
		},
		"GET /api/test3": {
			Path:       "/api/test3",
			Method:     "GET",
			Status:     StatusError,
			Parameters: []ParameterCoverage{},
			TestCount:  1,
			ErrorCount: 1,
		},
		"DELETE /api/test4": {
			Path:       "/api/test4",
			Method:     "DELETE",
			Status:     StatusUntested,
			Parameters: []ParameterCoverage{},
			TestCount:  0,
		},
	}
	
	// Get coverage stats
	stats := analyzer.GetCoverageStats()
	
	// Check the stats
	if stats["total_endpoints"] != 4 {
		t.Errorf("Expected total_endpoints 4, got %v", stats["total_endpoints"])
	}
	if stats["tested_endpoints"] != 1 {
		t.Errorf("Expected tested_endpoints 1, got %v", stats["tested_endpoints"])
	}
	if stats["partial_endpoints"] != 1 {
		t.Errorf("Expected partial_endpoints 1, got %v", stats["partial_endpoints"])
	}
	if stats["error_endpoints"] != 1 {
		t.Errorf("Expected error_endpoints 1, got %v", stats["error_endpoints"])
	}
	if stats["untested_endpoints"] != 1 {
		t.Errorf("Expected untested_endpoints 1, got %v", stats["untested_endpoints"])
	}
	
	// Check coverage percentages
	endpointCoverage := stats["endpoint_coverage"].(float64)
	if endpointCoverage != 50.0 {
		t.Errorf("Expected endpoint_coverage 50.0, got %v", endpointCoverage)
	}
	
	paramCoverage := stats["parameter_coverage"].(float64)
	if paramCoverage != 50.0 {
		t.Errorf("Expected parameter_coverage 50.0, got %v", paramCoverage)
	}
}

func TestGenerateReport(t *testing.T) {
	// Create a coverage analyzer with some test data
	analyzer := NewCoverageAnalyzer(&CoverageOptions{
		Format: FormatJSON,
	})
	
	// Add some endpoints
	analyzer.endpoints = map[string]*EndpointCoverage{
		"GET /api/test1": {
			Path:       "/api/test1",
			Method:     "GET",
			Status:     StatusTested,
			Parameters: []ParameterCoverage{},
			TestCount:  1,
			LastTested: time.Now(),
		},
	}
	
	// Generate a JSON report
	report, err := analyzer.GenerateReport()
	if err != nil {
		t.Errorf("Error generating report: %v", err)
	}
	
	// Check that the report is valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(report), &jsonData); err != nil {
		t.Errorf("Generated report is not valid JSON: %v", err)
	}
	
	// Check that the report contains the expected sections
	if _, exists := jsonData["stats"]; !exists {
		t.Error("Report does not contain stats section")
	}
	if _, exists := jsonData["endpoints"]; !exists {
		t.Error("Report does not contain endpoints section")
	}
	
	// Test HTML format
	analyzer.options.Format = FormatHTML
	report, err = analyzer.GenerateReport()
	if err != nil {
		t.Errorf("Error generating HTML report: %v", err)
	}
	if !strings.Contains(report, "<!DOCTYPE html>") {
		t.Error("HTML report does not contain DOCTYPE declaration")
	}
	
	// Test Markdown format
	analyzer.options.Format = FormatMarkdown
	report, err = analyzer.GenerateReport()
	if err != nil {
		t.Errorf("Error generating Markdown report: %v", err)
	}
	if !strings.Contains(report, "# API Coverage Report") {
		t.Error("Markdown report does not contain expected header")
	}
	
	// Test Text format
	analyzer.options.Format = FormatText
	report, err = analyzer.GenerateReport()
	if err != nil {
		t.Errorf("Error generating Text report: %v", err)
	}
	if !strings.Contains(report, "API COVERAGE REPORT") {
		t.Error("Text report does not contain expected header")
	}
	
	// Test invalid format
	analyzer.options.Format = "invalid"
	_, err = analyzer.GenerateReport()
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}

func TestGetEndpointsForReport(t *testing.T) {
	// Create a coverage analyzer with some test data
	analyzer := NewCoverageAnalyzer(&CoverageOptions{
		IncludeUntested: true,
	})
	
	// Add some endpoints with different statuses
	analyzer.endpoints = map[string]*EndpointCoverage{
		"GET /api/test1": {
			Path:       "/api/test1",
			Method:     "GET",
			Status:     StatusTested,
			Parameters: []ParameterCoverage{},
			TestCount:  1,
		},
		"POST /api/test2": {
			Path:       "/api/test2",
			Method:     "POST",
			Status:     StatusUntested,
			Parameters: []ParameterCoverage{},
			TestCount:  0,
		},
	}
	
	// Get endpoints for report
	endpoints := analyzer.getEndpointsForReport()
	
	// Check that all endpoints are included
	if len(endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
	}
	
	// Test with IncludeUntested = false
	analyzer.options.IncludeUntested = false
	endpoints = analyzer.getEndpointsForReport()
	
	// Check that only tested endpoints are included
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/api/test1" {
		t.Errorf("Expected path /api/test1, got %s", endpoints[0].Path)
	}
}