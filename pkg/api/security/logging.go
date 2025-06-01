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

// InsufficientLoggingTester implements testing for Insufficient Logging & Monitoring (API10:2019)
type InsufficientLoggingTester struct {
	// Configuration options
	SecurityEvents         []string
	LoggingEndpoints       []string
	MonitoringEndpoints    []string
	TestFailedLogins       bool
	TestAccessViolations   bool
	TestDataManipulation   bool
	TestRateLimitViolations bool
}

// NewInsufficientLoggingTester creates a new tester for Insufficient Logging & Monitoring
func NewInsufficientLoggingTester() *InsufficientLoggingTester {
	return &InsufficientLoggingTester{
		SecurityEvents: []string{
			"login", "logout", "password_change", "password_reset", "account_creation",
			"account_lockout", "account_deletion", "data_export", "data_import",
			"permission_change", "role_change", "admin_action", "payment", "transaction",
		},
		LoggingEndpoints: []string{
			"logs", "audit", "events", "activity", "history", "logging",
			"log-viewer", "audit-log", "event-log", "security-log",
		},
		MonitoringEndpoints: []string{
			"monitor", "monitoring", "status", "health", "metrics", "stats",
			"dashboard", "analytics", "telemetry", "performance",
		},
		TestFailedLogins:       true,
		TestAccessViolations:   true,
		TestDataManipulation:   true,
		TestRateLimitViolations: true,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *InsufficientLoggingTester) GetType() VulnerabilityType {
	return VulnInsufficientLogging
}

// GetName returns the name of the security test
func (t *InsufficientLoggingTester) GetName() string {
	return "Insufficient Logging & Monitoring"
}

// GetDescription returns a description of the security test
func (t *InsufficientLoggingTester) GetDescription() string {
	return "Tests for API endpoints that lack proper logging and monitoring of security events, which can lead to undetected security breaches."
}

// Test runs the security test against the target
func (t *InsufficientLoggingTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract the base URL from the config
	baseURL := extractBaseURL(config.Url)

	// Test for access to logging endpoints
	t.testLoggingEndpointAccess(baseURL, r, result)

	// Test for access to monitoring endpoints
	t.testMonitoringEndpointAccess(baseURL, r, result)

	// Test for failed login logging
	if t.TestFailedLogins {
		t.testFailedLoginLogging(baseURL, r, result)
	}

	// Test for access violation logging
	if t.TestAccessViolations {
		t.testAccessViolationLogging(baseURL, r, result)
	}

	// Test for data manipulation logging
	if t.TestDataManipulation {
		t.testDataManipulationLogging(baseURL, r, result)
	}

	// Test for rate limit violation logging
	if t.TestRateLimitViolations {
		t.testRateLimitViolationLogging(baseURL, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testLoggingEndpointAccess tests for unauthorized access to logging endpoints
func (t *InsufficientLoggingTester) testLoggingEndpointAccess(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, endpoint := range t.LoggingEndpoints {
		// Create URLs with different patterns
		testURLs := []string{
			fmt.Sprintf("%s/%s", baseURL, endpoint),
			fmt.Sprintf("%s/api/%s", baseURL, endpoint),
			fmt.Sprintf("%s/%s.php", baseURL, endpoint),
			fmt.Sprintf("%s/%s.json", baseURL, endpoint),
		}

		for _, testURL := range testURLs {
			req := &ffuf.Request{
				Method: "GET",
				Url:    testURL,
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful access to a logging endpoint
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInsufficientLogging,
					Name:        "Logging Endpoint Accessible Without Authentication",
					Description: "A logging endpoint is accessible without proper authentication.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed logging endpoint: %s", endpoint),
					Remediation: "Restrict access to logging endpoints. Implement proper authentication and authorization. Consider using a separate logging system not directly accessible from the public internet.",
					CVSS:        7.5,
					CWE:         "CWE-532",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more URLs for this endpoint
			}
		}
	}
}

// testMonitoringEndpointAccess tests for unauthorized access to monitoring endpoints
func (t *InsufficientLoggingTester) testMonitoringEndpointAccess(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, endpoint := range t.MonitoringEndpoints {
		// Create URLs with different patterns
		testURLs := []string{
			fmt.Sprintf("%s/%s", baseURL, endpoint),
			fmt.Sprintf("%s/api/%s", baseURL, endpoint),
			fmt.Sprintf("%s/%s.php", baseURL, endpoint),
			fmt.Sprintf("%s/%s.json", baseURL, endpoint),
		}

		for _, testURL := range testURLs {
			req := &ffuf.Request{
				Method: "GET",
				Url:    testURL,
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the response indicates a successful access to a monitoring endpoint
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInsufficientLogging,
					Name:        "Monitoring Endpoint Accessible Without Authentication",
					Description: "A monitoring endpoint is accessible without proper authentication.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed monitoring endpoint: %s", endpoint),
					Remediation: "Restrict access to monitoring endpoints. Implement proper authentication and authorization. Consider using a separate monitoring system not directly accessible from the public internet.",
					CVSS:        7.5,
					CWE:         "CWE-532",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more URLs for this endpoint
			}
		}
	}
}

// testFailedLoginLogging tests for insufficient logging of failed login attempts
func (t *InsufficientLoggingTester) testFailedLoginLogging(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Common login endpoints
	loginEndpoints := []string{
		"/login", "/signin", "/auth", "/api/login", "/api/auth", "/api/v1/login",
		"/user/login", "/account/login", "/auth/login", "/api/auth/login",
	}

	// Test each login endpoint
	for _, endpoint := range loginEndpoints {
		testURL := fmt.Sprintf("%s%s", baseURL, endpoint)

		// Create a request with invalid credentials
		req := &ffuf.Request{
			Method: "POST",
			Url:    testURL,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
			Data: []byte(`{"username":"invalid_user","password":"invalid_password"}`),
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// If the login endpoint exists (returns 401 or 403), test for insufficient logging
		if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 400 {
			// Make multiple failed login attempts to test for account lockout or rate limiting
			for i := 0; i < 5; i++ {
				_, err := r.Execute(req)
				if err != nil {
					break
				}
				time.Sleep(100 * time.Millisecond) // Small delay between requests
			}

			// Check if there's any indication of logging or monitoring
			// This is a heuristic test - in a real implementation, we would need to check server logs
			if !hasLoggingIndication(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInsufficientLogging,
					Name:        "Insufficient Logging of Failed Login Attempts",
					Description: "The API does not appear to properly log failed login attempts.",
					Severity:    "Medium",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    "Multiple failed login attempts did not trigger account lockout or rate limiting, suggesting insufficient logging and monitoring.",
					Remediation: "Implement proper logging of all authentication events, especially failed login attempts. Set up monitoring and alerting for multiple failed login attempts. Implement account lockout or rate limiting after a certain number of failed attempts.",
					CVSS:        6.0,
					CWE:         "CWE-778",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more endpoints
			}
		}
	}
}

// testAccessViolationLogging tests for insufficient logging of access violations
func (t *InsufficientLoggingTester) testAccessViolationLogging(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Common protected resource endpoints
	protectedEndpoints := []string{
		"/admin", "/dashboard", "/settings", "/profile", "/account",
		"/api/admin", "/api/users", "/api/settings", "/api/config",
	}

	// Test each protected endpoint
	for _, endpoint := range protectedEndpoints {
		testURL := fmt.Sprintf("%s%s", baseURL, endpoint)

		// Create a request without authentication
		req := &ffuf.Request{
			Method: "GET",
			Url:    testURL,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// If the endpoint exists and requires authentication (returns 401 or 403), test for insufficient logging
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			// Make multiple access attempts to test for rate limiting or blocking
			for i := 0; i < 5; i++ {
				_, err := r.Execute(req)
				if err != nil {
					break
				}
				time.Sleep(100 * time.Millisecond) // Small delay between requests
			}

			// Check if there's any indication of logging or monitoring
			// This is a heuristic test - in a real implementation, we would need to check server logs
			if !hasLoggingIndication(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInsufficientLogging,
					Name:        "Insufficient Logging of Access Violations",
					Description: "The API does not appear to properly log access violations.",
					Severity:    "Medium",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    "Multiple unauthorized access attempts did not trigger rate limiting or IP blocking, suggesting insufficient logging and monitoring.",
					Remediation: "Implement proper logging of all access control decisions, especially denied access attempts. Set up monitoring and alerting for multiple unauthorized access attempts. Consider implementing temporary IP blocking after a certain number of violations.",
					CVSS:        6.0,
					CWE:         "CWE-778",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Access_Control_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more endpoints
			}
		}
	}
}

// testDataManipulationLogging tests for insufficient logging of data manipulation
func (t *InsufficientLoggingTester) testDataManipulationLogging(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Common data manipulation endpoints
	dataEndpoints := []string{
		"/api/data", "/api/records", "/api/users", "/api/items", "/api/products",
		"/api/v1/data", "/api/v1/records", "/api/v1/users", "/api/v1/items",
	}

	// Test each data endpoint
	for _, endpoint := range dataEndpoints {
		testURL := fmt.Sprintf("%s%s", baseURL, endpoint)

		// Create a request to modify data
		req := &ffuf.Request{
			Method: "PUT",
			Url:    testURL,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
			Data: []byte(`{"id":1,"name":"test","value":"modified"}`),
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// If the endpoint exists (returns any response), test for insufficient logging
		// Make multiple data modification attempts
		for i := 0; i < 3; i++ {
			_, err := r.Execute(req)
			if err != nil {
				break
			}
			time.Sleep(100 * time.Millisecond) // Small delay between requests
		}

		// Check if there's any indication of logging or monitoring
		// This is a heuristic test - in a real implementation, we would need to check server logs
		if !hasLoggingIndication(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnInsufficientLogging,
				Name:        "Insufficient Logging of Data Manipulation",
				Description: "The API does not appear to properly log data manipulation operations.",
				Severity:    "Medium",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    "Multiple data modification attempts did not show any indication of being logged or monitored.",
				Remediation: "Implement proper logging of all data modification operations. Log the user, timestamp, operation type, and affected data. Set up monitoring and alerting for suspicious data manipulation patterns.",
				CVSS:        5.5,
				CWE:         "CWE-778",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			break // Found a vulnerability, no need to test more endpoints
		}
	}
}

// testRateLimitViolationLogging tests for insufficient logging of rate limit violations
func (t *InsufficientLoggingTester) testRateLimitViolationLogging(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Common API endpoints that might have rate limiting
	apiEndpoints := []string{
		"/api", "/api/v1", "/api/data", "/api/search", "/api/query",
		"/api/v1/data", "/api/v1/search", "/api/v1/query",
	}

	// Test each API endpoint
	for _, endpoint := range apiEndpoints {
		testURL := fmt.Sprintf("%s%s", baseURL, endpoint)

		// Create a request
		req := &ffuf.Request{
			Method: "GET",
			Url:    testURL,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// If the endpoint exists, test for rate limiting
		if resp.StatusCode < 500 {
			// Make many requests to trigger rate limiting
			rateLimited := false
			for i := 0; i < 20; i++ {
				resp, err := r.Execute(req)
				if err != nil {
					break
				}

				// Check if rate limiting has been triggered
				if resp.StatusCode == 429 || hasRateLimitHeaders(resp) {
					rateLimited = true
					break
				}

				time.Sleep(50 * time.Millisecond) // Small delay between requests
			}

			// If rate limiting was triggered but there's no indication of logging
			if rateLimited && !hasLoggingIndication(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnInsufficientLogging,
					Name:        "Insufficient Logging of Rate Limit Violations",
					Description: "The API implements rate limiting but does not appear to properly log rate limit violations.",
					Severity:    "Low",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    "Rate limiting was triggered but there's no indication of logging or monitoring of these violations.",
					Remediation: "Implement proper logging of all rate limit violations. Log the client IP, timestamp, endpoint, and request count. Set up monitoring and alerting for repeated rate limit violations from the same client.",
					CVSS:        4.0,
					CWE:         "CWE-778",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xaa-insufficient-logging-monitoring/",
						"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more endpoints
			}
		}
	}
}

// hasLoggingIndication checks if there's any indication of logging in the response
func hasLoggingIndication(resp ffuf.Response) bool {
	// This is a heuristic check - in a real implementation, we would need more sophisticated methods
	// Check for headers that might indicate logging
	loggingHeaders := []string{
		"X-Log-ID", "X-Request-ID", "X-Trace-ID", "X-Transaction-ID",
		"X-Correlation-ID", "X-Debug-ID", "X-Activity-ID",
	}

	for _, header := range loggingHeaders {
		if values, ok := resp.Headers[header]; ok && len(values) > 0 {
			return true
		}
	}

	// Check response body for logging indications
	responseData := string(resp.Data)
	loggingPatterns := []string{
		"log", "audit", "track", "monitor", "record", "event",
		"activity", "security", "warning", "alert", "notification",
	}

	for _, pattern := range loggingPatterns {
		if strings.Contains(strings.ToLower(responseData), pattern) {
			return true
		}
	}

	return false
}

// hasRateLimitHeaders checks if the response has rate limit headers
func hasRateLimitHeaders(resp ffuf.Response) bool {
	rateLimitHeaders := []string{
		"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset",
		"Retry-After", "RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset",
		"X-Rate-Limit-Limit", "X-Rate-Limit-Remaining", "X-Rate-Limit-Reset",
	}

	for _, header := range rateLimitHeaders {
		if values, ok := resp.Headers[header]; ok && len(values) > 0 {
			return true
		}
	}

	return false
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewInsufficientLoggingTester())
}
