// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// BrokenObjectLevelAuthTester implements testing for Broken Object Level Authorization (API1:2019)
type BrokenObjectLevelAuthTester struct {
	// Configuration options
	MaxIDsToTest     int
	IDParameterNames []string
	TestUserIDs      []string
	TestObjectIDs    []string
}

// NewBrokenObjectLevelAuthTester creates a new tester for Broken Object Level Authorization
func NewBrokenObjectLevelAuthTester() *BrokenObjectLevelAuthTester {
	return &BrokenObjectLevelAuthTester{
		MaxIDsToTest: 10,
		IDParameterNames: []string{
			"id", "user_id", "user", "account", "account_id", "customer", "customer_id",
			"object", "object_id", "record", "record_id", "uuid", "guid",
		},
		TestUserIDs:   []string{},
		TestObjectIDs: []string{},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *BrokenObjectLevelAuthTester) GetType() VulnerabilityType {
	return VulnBrokenObjectLevelAuth
}

// GetName returns the name of the security test
func (t *BrokenObjectLevelAuthTester) GetName() string {
	return "Broken Object Level Authorization"
}

// GetDescription returns a description of the security test
func (t *BrokenObjectLevelAuthTester) GetDescription() string {
	return "Tests for API endpoints that don't properly verify that the requesting user has the necessary permissions to access the requested resource."
}

// Test runs the security test against the target
func (t *BrokenObjectLevelAuthTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// If no test IDs are provided, generate some
	if len(t.TestObjectIDs) == 0 {
		t.TestObjectIDs = generateTestIDs(t.MaxIDsToTest)
	}

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for BOLA vulnerabilities
	for _, endpoint := range endpoints {
		// Skip endpoints that don't look like they would have object IDs
		if !containsIDPattern(endpoint) {
			continue
		}

		// Test the endpoint with different object IDs
		for _, testID := range t.TestObjectIDs {
			// Create a modified endpoint with the test ID
			modifiedEndpoint := replaceIDInEndpoint(endpoint, testID)
			if modifiedEndpoint == endpoint {
				continue // No ID was replaced
			}

			// Create a request for the modified endpoint
			req := &ffuf.Request{
				Method:  "GET",
				Url:     modifiedEndpoint,
				Headers: config.Headers,
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue // Skip this endpoint if there's an error
			}

			// Check if the response indicates a successful access
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnBrokenObjectLevelAuth,
					Name:        "Broken Object Level Authorization",
					Description: "The API endpoint allows access to resources that should be protected.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed resource with ID %s without proper authorization", testID),
					Remediation: "Implement proper authorization checks for all API endpoints that access resources. Verify that the requesting user has the necessary permissions to access the requested resource.",
					CVSS:        8.2,
					CWE:         "CWE-285",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// Helper functions

// generateTestIDs generates a list of test IDs
func generateTestIDs(count int) []string {
	var ids []string
	for i := 1; i <= count; i++ {
		ids = append(ids, strconv.Itoa(i))
	}
	return ids
}

// extractEndpointsFromConfig extracts potential endpoints from the config
func extractEndpointsFromConfig(config *ffuf.Config) []string {
	// In a real implementation, this would extract endpoints from the config
	// For now, we'll just use the URL from the config
	return []string{config.Url}
}

// containsIDPattern checks if an endpoint contains patterns that suggest it has an ID
func containsIDPattern(endpoint string) bool {
	patterns := []string{
		"/[0-9]+",
		"/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}", // UUID
		"id=",
		"user=",
		"account=",
		"customer=",
		"object=",
		"record=",
	}

	for _, pattern := range patterns {
		if strings.Contains(endpoint, pattern) {
			return true
		}
	}
	return false
}

// replaceIDInEndpoint replaces an ID in an endpoint with a test ID
func replaceIDInEndpoint(endpoint, testID string) string {
	// Try to replace ID in path
	parts := strings.Split(endpoint, "/")
	for i, part := range parts {
		if isNumeric(part) || isUUID(part) {
			parts[i] = testID
			return strings.Join(parts, "/")
		}
	}

	// Try to replace ID in query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}

	q := u.Query()
	replaced := false
	for _, paramName := range []string{"id", "user_id", "user", "account", "account_id", "customer", "customer_id", "object", "object_id"} {
		if q.Get(paramName) != "" {
			q.Set(paramName, testID)
			replaced = true
		}
	}

	if replaced {
		u.RawQuery = q.Encode()
		return u.String()
	}

	return endpoint
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	return strings.Count(s, "-") == 4 && len(s) == 36
}

// isSuccessfulAccess checks if a response indicates successful access to a resource
func isSuccessfulAccess(resp ffuf.Response) bool {
	// In a real implementation, this would be more sophisticated
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// convertToHTTPRequest converts an ffuf.Request to an http.Request
func convertToHTTPRequest(req *ffuf.Request) *http.Request {
	httpReq, _ := http.NewRequest(req.Method, req.Url, strings.NewReader(string(req.Data)))
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	return httpReq
}

// convertToHTTPResponse converts an ffuf.Response to an http.Response
func convertToHTTPResponse(resp ffuf.Response) *http.Response {
	// This is a simplified conversion
	return &http.Response{
		StatusCode: int(resp.StatusCode),
		Body:       nil, // We don't need the body for now
		Header:     make(http.Header),
	}
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewBrokenObjectLevelAuthTester())
}
