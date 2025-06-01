// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// IDORTester implements testing for Insecure Direct Object Reference vulnerabilities
type IDORTester struct {
	// Configuration options
	MaxIDsToTest       int
	IDParameterNames   []string
	TestUserIDs        []string
	TestObjectIDs      []string
	TestPredictableIDs bool
	TestSequentialIDs  bool
	TestCommonIDs      bool
	TestDifferentHTTPMethods bool
}

// NewIDORTester creates a new tester for IDOR vulnerabilities
func NewIDORTester() *IDORTester {
	return &IDORTester{
		MaxIDsToTest: 20,
		IDParameterNames: []string{
			"id", "user_id", "user", "account", "account_id", "customer", "customer_id",
			"object", "object_id", "record", "record_id", "uuid", "guid", "file", "document",
			"order", "invoice", "payment", "transaction", "profile", "report",
		},
		TestUserIDs:        []string{},
		TestObjectIDs:      []string{},
		TestPredictableIDs: true,
		TestSequentialIDs:  true,
		TestCommonIDs:      true,
		TestDifferentHTTPMethods: true,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *IDORTester) GetType() VulnerabilityType {
	return VulnBrokenObjectLevelAuth // IDOR is a type of Broken Object Level Authorization
}

// GetName returns the name of the security test
func (t *IDORTester) GetName() string {
	return "Insecure Direct Object Reference (IDOR)"
}

// GetDescription returns a description of the security test
func (t *IDORTester) GetDescription() string {
	return "Tests for API endpoints that allow direct access to objects based on user-supplied input without proper authorization checks."
}

// Test runs the security test against the target
func (t *IDORTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Generate test IDs if none are provided
	if len(t.TestObjectIDs) == 0 {
		t.TestObjectIDs = t.generateTestIDs()
	}

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for IDOR vulnerabilities
	for _, endpoint := range endpoints {
		// Skip endpoints that don't look like they would have object IDs
		if !t.containsIDPattern(endpoint) {
			continue
		}

		// Test the endpoint with different object IDs
		t.testEndpointWithIDs(endpoint, r, config, result)

		// Test the endpoint with predictable IDs if enabled
		if t.TestPredictableIDs {
			t.testEndpointWithPredictableIDs(endpoint, r, config, result)
		}

		// Test the endpoint with sequential IDs if enabled
		if t.TestSequentialIDs {
			t.testEndpointWithSequentialIDs(endpoint, r, config, result)
		}

		// Test the endpoint with common IDs if enabled
		if t.TestCommonIDs {
			t.testEndpointWithCommonIDs(endpoint, r, config, result)
		}

		// Test the endpoint with different HTTP methods if enabled
		if t.TestDifferentHTTPMethods {
			t.testEndpointWithDifferentMethods(endpoint, r, config, result)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testEndpointWithIDs tests an endpoint with different object IDs
func (t *IDORTester) testEndpointWithIDs(endpoint string, r ffuf.RunnerProvider, config *ffuf.Config, result *TestResult) {
	for _, testID := range t.TestObjectIDs {
		// Create a modified endpoint with the test ID
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
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
				Name:        "Insecure Direct Object Reference (IDOR)",
				Description: "The API endpoint allows direct access to objects without proper authorization checks.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed resource with ID %s without proper authorization", testID),
				Remediation: "Implement proper authorization checks for all API endpoints that access resources. Verify that the requesting user has the necessary permissions to access the requested resource. Consider using indirect references that are mapped on the server side.",
				CVSS:        8.2,
				CWE:         "CWE-639",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testEndpointWithPredictableIDs tests an endpoint with predictable IDs
func (t *IDORTester) testEndpointWithPredictableIDs(endpoint string, r ffuf.RunnerProvider, config *ffuf.Config, result *TestResult) {
	// Common predictable ID patterns
	predictableIDs := []string{
		"admin", "administrator", "root", "superuser", "supervisor",
		"user1", "user2", "user3", "test", "demo", "sample",
		"default", "guest", "anonymous", "system",
	}

	for _, testID := range predictableIDs {
		// Create a modified endpoint with the test ID
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
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
				Name:        "IDOR with Predictable ID",
				Description: "The API endpoint allows access to objects with predictable IDs without proper authorization.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed resource with predictable ID '%s' without proper authorization", testID),
				Remediation: "Use unpredictable, random IDs for resources. Implement proper authorization checks. Consider using indirect references that are mapped on the server side.",
				CVSS:        8.5,
				CWE:         "CWE-639",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testEndpointWithSequentialIDs tests an endpoint with sequential IDs
func (t *IDORTester) testEndpointWithSequentialIDs(endpoint string, r ffuf.RunnerProvider, config *ffuf.Config, result *TestResult) {
	// First, try to find a valid ID by testing sequential IDs
	var validID string

	// Test IDs from 1 to 10 to find a valid one
	for i := 1; i <= 10; i++ {
		testID := strconv.Itoa(i)
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
		if modifiedEndpoint == endpoint {
			continue // No ID was replaced
		}

		req := &ffuf.Request{
			Method:  "GET",
			Url:     modifiedEndpoint,
			Headers: config.Headers,
		}

		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		if isSuccessfulAccess(resp) {
			validID = testID
			break
		}
	}

	// If we found a valid ID, try to access other sequential IDs
	if validID != "" {
		validIDInt, _ := strconv.Atoi(validID)

		// Try IDs before and after the valid ID
		for i := validIDInt - 5; i <= validIDInt + 5; i++ {
			if i <= 0 || i == validIDInt {
				continue
			}

			testID := strconv.Itoa(i)
			modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
			if modifiedEndpoint == endpoint {
				continue
			}

			req := &ffuf.Request{
				Method:  "GET",
				Url:     modifiedEndpoint,
				Headers: config.Headers,
			}

			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnBrokenObjectLevelAuth,
					Name:        "IDOR with Sequential IDs",
					Description: "The API endpoint uses sequential IDs and allows access without proper authorization.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed resources with sequential IDs (%s and %s) without proper authorization", validID, testID),
					Remediation: "Use unpredictable, random IDs for resources. Implement proper authorization checks. Consider using indirect references that are mapped on the server side.",
					CVSS:        8.5,
					CWE:         "CWE-639",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more IDs
			}
		}
	}
}

// testEndpointWithCommonIDs tests an endpoint with common IDs
func (t *IDORTester) testEndpointWithCommonIDs(endpoint string, r ffuf.RunnerProvider, config *ffuf.Config, result *TestResult) {
	// Common IDs that might be used in systems
	commonIDs := []string{
		"1", "2", "3", "10", "100", "1000",
		"a1b2c3", "test123", "demo123", "sample123",
		"00000000-0000-0000-0000-000000000000", // Null UUID
		"11111111-1111-1111-1111-111111111111", // Common test UUID
	}

	for _, testID := range commonIDs {
		// Create a modified endpoint with the test ID
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
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
				Name:        "IDOR with Common ID",
				Description: "The API endpoint allows access to objects with common IDs without proper authorization.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed resource with common ID '%s' without proper authorization", testID),
				Remediation: "Use unpredictable, random IDs for resources. Implement proper authorization checks. Consider using indirect references that are mapped on the server side.",
				CVSS:        8.2,
				CWE:         "CWE-639",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testEndpointWithDifferentMethods tests an endpoint with different HTTP methods
func (t *IDORTester) testEndpointWithDifferentMethods(endpoint string, r ffuf.RunnerProvider, config *ffuf.Config, result *TestResult) {
	// HTTP methods to test
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	// First, find a valid ID
	var validID string
	for i := 1; i <= 10; i++ {
		testID := strconv.Itoa(i)
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, testID)
		if modifiedEndpoint == endpoint {
			continue
		}

		req := &ffuf.Request{
			Method:  "GET",
			Url:     modifiedEndpoint,
			Headers: config.Headers,
		}

		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		if isSuccessfulAccess(resp) {
			validID = testID
			break
		}
	}

	// If we found a valid ID, try different HTTP methods
	if validID != "" {
		modifiedEndpoint := t.replaceIDInEndpoint(endpoint, validID)

		for _, method := range methods {
			req := &ffuf.Request{
				Method:  method,
				Url:     modifiedEndpoint,
				Headers: config.Headers,
				Data:    []byte(`{"test":"data"}`), // Add some data for POST/PUT/PATCH
			}

			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful access
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnBrokenObjectLevelAuth,
					Name:        "IDOR with Different HTTP Method",
					Description: fmt.Sprintf("The API endpoint allows %s access to objects without proper authorization.", method),
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed/modified resource with ID %s using %s method without proper authorization", validID, method),
					Remediation: "Implement proper authorization checks for all HTTP methods. Ensure that users can only access or modify resources they are authorized to.",
					CVSS:        8.5,
					CWE:         "CWE-639",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa1-broken-object-level-authorization/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}
}

// generateTestIDs generates a list of test IDs
func (t *IDORTester) generateTestIDs() []string {
	var ids []string

	// Generate numeric IDs
	for i := 1; i <= t.MaxIDsToTest/2; i++ {
		ids = append(ids, strconv.Itoa(i))
	}

	// Generate UUIDs
	for i := 1; i <= t.MaxIDsToTest/2; i++ {
		// This is a simplified UUID generation - in a real implementation, use a proper UUID library
		uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", 
			i, i%65536, i%65536, i%65536, i)
		ids = append(ids, uuid)
	}

	return ids
}

// containsIDPattern checks if an endpoint contains patterns that suggest it has an ID
func (t *IDORTester) containsIDPattern(endpoint string) bool {
	patterns := []string{
		"/[0-9]+",
		"/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}", // UUID
	}

	for _, paramName := range t.IDParameterNames {
		patterns = append(patterns, paramName+"=")
	}

	for _, pattern := range patterns {
		if strings.Contains(endpoint, pattern) {
			return true
		}
	}
	return false
}

// replaceIDInEndpoint replaces an ID in an endpoint with a test ID
func (t *IDORTester) replaceIDInEndpoint(endpoint, testID string) string {
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
	for _, paramName := range t.IDParameterNames {
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


func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewIDORTester())
}
