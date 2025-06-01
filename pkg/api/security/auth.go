// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// BrokenAuthTester implements testing for Broken User Authentication (API2:2019)
type BrokenAuthTester struct {
	// Configuration options
	CommonPasswords []string
	WeakTokenTests  []string
	JWTTests        []string
}

// NewBrokenAuthTester creates a new tester for Broken User Authentication
func NewBrokenAuthTester() *BrokenAuthTester {
	return &BrokenAuthTester{
		CommonPasswords: []string{
			"password", "123456", "admin", "welcome", "password123",
			"12345678", "qwerty", "1234567890", "111111", "1234567",
		},
		WeakTokenTests: []string{
			"", // Empty token
			"null",
			"undefined",
			"guest",
			"demo",
			"test",
			"user",
			"admin",
		},
		JWTTests: []string{
			// Common JWT testing payloads would go here
			"eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.", // alg:none attack
		},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *BrokenAuthTester) GetType() VulnerabilityType {
	return VulnBrokenAuth
}

// GetName returns the name of the security test
func (t *BrokenAuthTester) GetName() string {
	return "Broken User Authentication"
}

// GetDescription returns a description of the security test
func (t *BrokenAuthTester) GetDescription() string {
	return "Tests for API endpoints with broken authentication mechanisms, including weak passwords, improper token validation, and insecure credential storage."
}

// Test runs the security test against the target
func (t *BrokenAuthTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential auth endpoints from the config
	endpoints := extractAuthEndpointsFromConfig(config)

	// Test each endpoint for authentication vulnerabilities
	for _, endpoint := range endpoints {
		// Test for weak password vulnerabilities
		t.testWeakPasswords(endpoint, r, result)

		// Test for weak token vulnerabilities
		t.testWeakTokens(endpoint, r, result)

		// Test for JWT vulnerabilities
		t.testJWTVulnerabilities(endpoint, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testWeakPasswords tests for weak password vulnerabilities
func (t *BrokenAuthTester) testWeakPasswords(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Skip endpoints that don't look like login endpoints
	if !isLoginEndpoint(endpoint) {
		return
	}

	// Common usernames to try
	usernames := []string{"admin", "user", "test", "demo", "guest"}

	for _, username := range usernames {
		for _, password := range t.CommonPasswords {
			// Create a login request with the username and password
			req := &ffuf.Request{
				Method: "POST",
				Url:    endpoint,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Data: []byte(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)),
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue // Skip this endpoint if there's an error
			}

			// Check if the login was successful
			if isSuccessfulLogin(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnBrokenAuth,
					Name:        "Weak Password Accepted",
					Description: "The API endpoint accepts weak or common passwords.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully logged in with username '%s' and common password '%s'", username, password),
					Remediation: "Implement strong password policies. Require complex passwords and check against common password lists. Consider implementing multi-factor authentication.",
					CVSS:        7.5,
					CWE:         "CWE-521",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa2-broken-authentication/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}
}

// testWeakTokens tests for weak token vulnerabilities
func (t *BrokenAuthTester) testWeakTokens(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Skip endpoints that don't look like they require authentication
	if !requiresAuthentication(endpoint) {
		return
	}

	for _, token := range t.WeakTokenTests {
		// Create a request with the token
		req := &ffuf.Request{
			Method: "GET",
			Url:    endpoint,
			Headers: map[string]string{
				"Authorization": "Bearer " + token,
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue // Skip this endpoint if there's an error
		}

		// Check if the request was successful despite using a weak token
		if isSuccessfulAccess(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnBrokenAuth,
				Name:        "Weak Token Accepted",
				Description: "The API endpoint accepts weak or predictable authentication tokens.",
				Severity:    "Critical",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed endpoint with weak token: '%s'", token),
				Remediation: "Implement proper token generation with sufficient entropy. Validate tokens properly on the server side. Consider using industry standard token formats like JWT with proper signing.",
				CVSS:        9.0,
				CWE:         "CWE-330",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa2-broken-authentication/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testJWTVulnerabilities tests for JWT-specific vulnerabilities
func (t *BrokenAuthTester) testJWTVulnerabilities(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Skip endpoints that don't look like they require authentication
	if !requiresAuthentication(endpoint) {
		return
	}

	for _, jwt := range t.JWTTests {
		// Create a request with the JWT
		req := &ffuf.Request{
			Method: "GET",
			Url:    endpoint,
			Headers: map[string]string{
				"Authorization": "Bearer " + jwt,
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue // Skip this endpoint if there's an error
		}

		// Check if the request was successful despite using a manipulated JWT
		if isSuccessfulAccess(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnBrokenAuth,
				Name:        "JWT Vulnerability",
				Description: "The API endpoint accepts manipulated JWT tokens.",
				Severity:    "Critical",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    "Successfully accessed endpoint with manipulated JWT token",
				Remediation: "Implement proper JWT validation. Use strong signing algorithms (RS256 instead of HS256). Validate all parts of the token including signature, expiration, and claims.",
				CVSS:        9.8,
				CWE:         "CWE-347",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa2-broken-authentication/",
					"https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// Helper functions

// extractAuthEndpointsFromConfig extracts potential authentication endpoints from the config
func extractAuthEndpointsFromConfig(config *ffuf.Config) []string {
	// In a real implementation, this would extract endpoints from the config
	// For now, we'll just use the URL from the config
	return []string{config.Url}
}

// isLoginEndpoint checks if an endpoint looks like a login endpoint
func isLoginEndpoint(endpoint string) bool {
	patterns := []string{
		"login",
		"auth",
		"authenticate",
		"signin",
		"sign-in",
		"token",
		"session",
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(endpoint), pattern) {
			return true
		}
	}
	return false
}

// requiresAuthentication checks if an endpoint likely requires authentication
func requiresAuthentication(endpoint string) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	patterns := []string{
		"api",
		"user",
		"account",
		"profile",
		"admin",
		"dashboard",
		"secure",
		"private",
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(endpoint), pattern) {
			return true
		}
	}
	return false
}

// isSuccessfulLogin checks if a response indicates a successful login
func isSuccessfulLogin(resp ffuf.Response) bool {
	// In a real implementation, this would be more sophisticated
	// Look for tokens in the response
	hasToken := strings.Contains(string(resp.Data), "token") ||
		strings.Contains(string(resp.Data), "jwt") ||
		strings.Contains(string(resp.Data), "access_token")

	// Check for success status code
	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	return isSuccess && hasToken
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewBrokenAuthTester())
}
