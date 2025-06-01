// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// MassAssignmentTester implements testing for Mass Assignment (API6:2019)
type MassAssignmentTester struct {
	// Configuration options
	SensitiveProperties []string
	AdminProperties     []string
	TestPayloads        map[string]interface{}
}

// NewMassAssignmentTester creates a new tester for Mass Assignment
func NewMassAssignmentTester() *MassAssignmentTester {
	// Define sensitive properties that should not be mass-assignable
	sensitiveProps := []string{
		"role", "isAdmin", "admin", "is_admin", "isadmin",
		"permissions", "permission", "access_level", "accessLevel",
		"group", "groups", "privilege", "privileges",
		"is_verified", "isVerified", "verified",
		"is_active", "isActive", "active",
		"is_deleted", "isDeleted", "deleted",
		"created_at", "createdAt", "created",
		"updated_at", "updatedAt", "updated",
		"password", "password_hash", "passwordHash",
		"api_key", "apiKey", "api_token", "apiToken",
		"secret", "secret_key", "secretKey",
		"credit_card", "creditCard", "credit_card_number", "creditCardNumber",
		"ssn", "social_security", "socialSecurity",
		"balance", "account_balance", "accountBalance",
		"points", "reward_points", "rewardPoints",
	}

	// Define admin-specific properties
	adminProps := []string{
		"role", "isAdmin", "admin", "is_admin", "isadmin",
		"permissions", "permission", "access_level", "accessLevel",
		"group", "groups", "privilege", "privileges",
	}

	// Define test payloads for different scenarios
	testPayloads := map[string]interface{}{
		"Role Elevation": map[string]interface{}{
			"role":       "admin",
			"isAdmin":    true,
			"is_admin":   true,
			"admin":      true,
			"permission": "admin",
			"privileges": []string{"admin", "superuser"},
		},
		"Status Manipulation": map[string]interface{}{
			"is_verified": true,
			"isVerified":  true,
			"verified":    true,
			"is_active":   true,
			"isActive":    true,
			"active":      true,
			"is_deleted":  false,
			"isDeleted":   false,
			"deleted":     false,
		},
		"Timestamp Manipulation": map[string]interface{}{
			"created_at": "2020-01-01T00:00:00Z",
			"createdAt":  "2020-01-01T00:00:00Z",
			"created":    "2020-01-01T00:00:00Z",
			"updated_at": "2020-01-01T00:00:00Z",
			"updatedAt":  "2020-01-01T00:00:00Z",
			"updated":    "2020-01-01T00:00:00Z",
		},
		"Financial Manipulation": map[string]interface{}{
			"balance":         999999.99,
			"account_balance": 999999.99,
			"accountBalance":  999999.99,
			"points":          999999,
			"reward_points":   999999,
			"rewardPoints":    999999,
		},
	}

	return &MassAssignmentTester{
		SensitiveProperties: sensitiveProps,
		AdminProperties:     adminProps,
		TestPayloads:        testPayloads,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *MassAssignmentTester) GetType() VulnerabilityType {
	return VulnMassAssignment
}

// GetName returns the name of the security test
func (t *MassAssignmentTester) GetName() string {
	return "Mass Assignment"
}

// GetDescription returns a description of the security test
func (t *MassAssignmentTester) GetDescription() string {
	return "Tests for API endpoints that automatically bind client-provided data to internal object properties without proper filtering, potentially allowing attackers to modify object properties they shouldn't have access to."
}

// Test runs the security test against the target
func (t *MassAssignmentTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for mass assignment vulnerabilities
	for _, endpoint := range endpoints {
		// Skip endpoints that don't look like they would accept POST/PUT/PATCH requests
		if !isWritableEndpoint(endpoint) {
			continue
		}

		// Test for mass assignment vulnerabilities with different payloads
		for payloadName, payload := range t.TestPayloads {
			t.testMassAssignment(endpoint, payloadName, payload, r, result)
		}

		// Test for mass assignment in nested objects
		t.testNestedMassAssignment(endpoint, r, result)

		// Test for mass assignment in arrays
		t.testArrayMassAssignment(endpoint, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testMassAssignment tests for mass assignment vulnerabilities
func (t *MassAssignmentTester) testMassAssignment(endpoint, payloadName string, payload interface{}, r ffuf.RunnerProvider, result *TestResult) {
	// Create a JSON payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Create a request with the payload
	req := &ffuf.Request{
		Method: "POST", // Try POST first
		Url:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Role":       "user", // Simulate a regular user
		},
		Data: jsonPayload,
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Check if the response indicates that the sensitive properties were accepted
		if t.responseIndicatesSuccess(resp, payload) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnMassAssignment,
				Name:        "Mass Assignment Vulnerability",
				Description: fmt.Sprintf("The API endpoint allows mass assignment of sensitive properties (%s).", payloadName),
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully assigned sensitive properties via %s payload", payloadName),
				Remediation: "Implement proper input validation and filtering. Use a whitelist approach to explicitly define which properties can be mass-assigned. Consider using DTOs (Data Transfer Objects) to separate API models from internal models.",
				CVSS:        8.0,
				CWE:         "CWE-915",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa6-mass-assignment/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	// Try with PUT and PATCH methods as well
	for _, method := range []string{"PUT", "PATCH"} {
		req.Method = method
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Check if the response indicates that the sensitive properties were accepted
			if t.responseIndicatesSuccess(resp, payload) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnMassAssignment,
					Name:        "Mass Assignment Vulnerability",
					Description: fmt.Sprintf("The API endpoint allows mass assignment of sensitive properties (%s) via %s method.", payloadName, method),
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully assigned sensitive properties via %s payload using %s method", payloadName, method),
					Remediation: "Implement proper input validation and filtering. Use a whitelist approach to explicitly define which properties can be mass-assigned. Consider using DTOs (Data Transfer Objects) to separate API models from internal models.",
					CVSS:        8.0,
					CWE:         "CWE-915",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa6-mass-assignment/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}
}

// testNestedMassAssignment tests for mass assignment vulnerabilities in nested objects
func (t *MassAssignmentTester) testNestedMassAssignment(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Create a nested payload with sensitive properties
	nestedPayload := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     "Test User",
			"email":    "test@example.com",
			"role":     "admin",
			"isAdmin":  true,
			"is_admin": true,
		},
		"profile": map[string]interface{}{
			"bio":         "Test bio",
			"permissions": []string{"admin", "superuser"},
			"is_verified": true,
		},
		"settings": map[string]interface{}{
			"theme":       "dark",
			"language":    "en",
			"accessLevel": "admin",
		},
	}

	// Create a JSON payload
	jsonPayload, err := json.Marshal(nestedPayload)
	if err != nil {
		return
	}

	// Create a request with the payload
	req := &ffuf.Request{
		Method: "POST",
		Url:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Role":       "user", // Simulate a regular user
		},
		Data: jsonPayload,
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Check if the response indicates that the sensitive properties were accepted
		if t.responseIndicatesSuccess(resp, nestedPayload) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnMassAssignment,
				Name:        "Nested Mass Assignment Vulnerability",
				Description: "The API endpoint allows mass assignment of sensitive properties in nested objects.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    "Successfully assigned sensitive properties in nested objects",
				Remediation: "Implement proper input validation and filtering for nested objects. Use a whitelist approach to explicitly define which properties can be mass-assigned at all levels of the object hierarchy.",
				CVSS:        8.0,
				CWE:         "CWE-915",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa6-mass-assignment/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testArrayMassAssignment tests for mass assignment vulnerabilities in arrays
func (t *MassAssignmentTester) testArrayMassAssignment(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Create an array payload with sensitive properties
	arrayPayload := map[string]interface{}{
		"users": []map[string]interface{}{
			{
				"name":     "User 1",
				"email":    "user1@example.com",
				"role":     "user",
				"isAdmin":  false,
				"is_admin": false,
			},
			{
				"name":     "Admin User",
				"email":    "admin@example.com",
				"role":     "admin",
				"isAdmin":  true,
				"is_admin": true,
			},
		},
	}

	// Create a JSON payload
	jsonPayload, err := json.Marshal(arrayPayload)
	if err != nil {
		return
	}

	// Create a request with the payload
	req := &ffuf.Request{
		Method: "POST",
		Url:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Role":       "user", // Simulate a regular user
		},
		Data: jsonPayload,
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Check if the response indicates that the sensitive properties were accepted
		if t.responseIndicatesSuccess(resp, arrayPayload) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnMassAssignment,
				Name:        "Array Mass Assignment Vulnerability",
				Description: "The API endpoint allows mass assignment of sensitive properties in arrays.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    "Successfully assigned sensitive properties in arrays",
				Remediation: "Implement proper input validation and filtering for arrays. Use a whitelist approach to explicitly define which properties can be mass-assigned for each element in the array.",
				CVSS:        8.0,
				CWE:         "CWE-915",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa6-mass-assignment/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// responseIndicatesSuccess checks if the response indicates that the sensitive properties were accepted
func (t *MassAssignmentTester) responseIndicatesSuccess(resp ffuf.Response, payload interface{}) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	// We're looking for indications that the server accepted our payload

	// Check if the response contains any of the sensitive properties we sent
	responseData := string(resp.Data)
	
	// Convert payload to JSON string for easier checking
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	
	payloadStr := string(payloadJSON)
	
	// Check for each sensitive property in the response
	for _, prop := range t.SensitiveProperties {
		// If the property is in our payload and also in the response, it might indicate success
		if strings.Contains(payloadStr, fmt.Sprintf(`"%s"`, prop)) && 
		   strings.Contains(responseData, fmt.Sprintf(`"%s"`, prop)) {
			return true
		}
	}
	
	// Check for success messages
	successIndicators := []string{
		"success", "created", "updated", "saved", "ok", "done",
		"200", "201", "202", "204",
	}
	
	for _, indicator := range successIndicators {
		if strings.Contains(strings.ToLower(responseData), strings.ToLower(indicator)) {
			return true
		}
	}
	
	return false
}

// isWritableEndpoint checks if an endpoint looks like it would accept POST/PUT/PATCH requests
func isWritableEndpoint(endpoint string) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	writablePatterns := []string{
		"create", "update", "edit", "save", "add", "new",
		"user", "profile", "account", "settings", "config",
		"register", "signup", "sign-up", "login", "signin", "sign-in",
		"post", "put", "patch",
	}

	for _, pattern := range writablePatterns {
		if strings.Contains(strings.ToLower(endpoint), pattern) {
			return true
		}
	}

	// Also check if the endpoint ends with a common API pattern
	apiPatterns := []string{
		"/api/", "/v1/", "/v2/", "/v3/", "/rest/", "/graphql",
	}

	for _, pattern := range apiPatterns {
		if strings.Contains(endpoint, pattern) {
			return true
		}
	}

	return false
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewMassAssignmentTester())
}