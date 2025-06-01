// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// ExcessiveDataExposureTester implements testing for Excessive Data Exposure (API3:2019)
type ExcessiveDataExposureTester struct {
	// Configuration options
	SensitiveDataPatterns map[string]*regexp.Regexp
}

// NewExcessiveDataExposureTester creates a new tester for Excessive Data Exposure
func NewExcessiveDataExposureTester() *ExcessiveDataExposureTester {
	// Compile regular expressions for common sensitive data patterns
	patterns := map[string]*regexp.Regexp{
		"Credit Card":        regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
		"Social Security":    regexp.MustCompile(`\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`),
		"Email":              regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`),
		"Password":           regexp.MustCompile(`(?i)("password"\s*:\s*"[^"]*")`),
		"API Key":            regexp.MustCompile(`(?i)("api[_-]?key"\s*:\s*"[^"]*")`),
		"Secret Key":         regexp.MustCompile(`(?i)("secret[_-]?key"\s*:\s*"[^"]*")`),
		"Access Token":       regexp.MustCompile(`(?i)("access[_-]?token"\s*:\s*"[^"]*")`),
		"Private Key":        regexp.MustCompile(`(?i)(-----BEGIN PRIVATE KEY-----)`),
		"AWS Key":            regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`),
		"Internal IP":        regexp.MustCompile(`\b(?:10|172\.(?:1[6-9]|2[0-9]|3[01])|192\.168)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`),
		"Internal Hostname":  regexp.MustCompile(`\b(?:localhost|internal|dev|staging|test|prod|production)\b`),
		"Debug Information":  regexp.MustCompile(`(?i)("debug"\s*:\s*true)`),
		"Stack Trace":        regexp.MustCompile(`(?i)(at\s+[\w.]+\([\w./: ]+\))`),
		"Personal Data":      regexp.MustCompile(`(?i)("(?:first|last)[_-]?name"\s*:\s*"[^"]*")`),
		"Authentication":     regexp.MustCompile(`(?i)("auth(?:entication)?"\s*:\s*"[^"]*")`),
		"Authorization":      regexp.MustCompile(`(?i)("authorization"\s*:\s*"[^"]*")`),
		"Session":            regexp.MustCompile(`(?i)("session[_-]?(?:id|token)"\s*:\s*"[^"]*")`),
		"Internal Endpoint":  regexp.MustCompile(`(?i)("(?:endpoint|url|uri)"\s*:\s*"(?:https?://)?(?:localhost|internal|dev|staging|test|prod|production)[^"]*")`),
		"Database Details":   regexp.MustCompile(`(?i)("(?:db|database|connection)[_-]?(?:string|url|uri)"\s*:\s*"[^"]*")`),
		"Hidden Field":       regexp.MustCompile(`(?i)("hidden"\s*:\s*true)`),
		"Internal ID":        regexp.MustCompile(`(?i)("internal[_-]?id"\s*:\s*"[^"]*")`),
	}

	return &ExcessiveDataExposureTester{
		SensitiveDataPatterns: patterns,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *ExcessiveDataExposureTester) GetType() VulnerabilityType {
	return VulnExcessiveDataExposure
}

// GetName returns the name of the security test
func (t *ExcessiveDataExposureTester) GetName() string {
	return "Excessive Data Exposure"
}

// GetDescription returns a description of the security test
func (t *ExcessiveDataExposureTester) GetDescription() string {
	return "Tests for API endpoints that return excessive data, potentially exposing sensitive information."
}

// Test runs the security test against the target
func (t *ExcessiveDataExposureTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for excessive data exposure
	for _, endpoint := range endpoints {
		// Create a request for the endpoint
		req := &ffuf.Request{
			Method:  "GET",
			Url:     endpoint,
			Headers: config.Headers,
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue // Skip this endpoint if there's an error
		}

		// Check if the response contains sensitive data
		findings := t.checkForSensitiveData(resp)
		if len(findings) > 0 {
			// Create a vulnerability info
			evidence := strings.Join(findings, "\n")
			vuln := VulnerabilityInfo{
				Type:        VulnExcessiveDataExposure,
				Name:        "Excessive Data Exposure",
				Description: "The API endpoint returns excessive data that may contain sensitive information.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    evidence,
				Remediation: "Implement proper data filtering. Only return the data that is necessary for the client. Consider using response filtering, data masking, or implementing a proper authorization model that restricts access to sensitive data.",
				CVSS:        7.5,
				CWE:         "CWE-213",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa3-excessive-data-exposure/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}

		// Also check for verbose error messages
		if isVerboseError(resp) {
			vuln := VulnerabilityInfo{
				Type:        VulnExcessiveDataExposure,
				Name:        "Verbose Error Messages",
				Description: "The API endpoint returns verbose error messages that may expose sensitive information.",
				Severity:    "Medium",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    "Response contains verbose error messages",
				Remediation: "Implement proper error handling. Return generic error messages to clients and log detailed errors on the server side.",
				CVSS:        5.0,
				CWE:         "CWE-209",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa3-excessive-data-exposure/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// checkForSensitiveData checks if a response contains sensitive data
func (t *ExcessiveDataExposureTester) checkForSensitiveData(resp ffuf.Response) []string {
	var findings []string
	responseData := string(resp.Data)

	// Check each pattern
	for name, pattern := range t.SensitiveDataPatterns {
		matches := pattern.FindAllString(responseData, -1)
		if len(matches) > 0 {
			// Limit the number of matches to report
			maxMatches := 3
			if len(matches) > maxMatches {
				matches = matches[:maxMatches]
			}
			
			// Create a finding
			finding := fmt.Sprintf("Found potential %s: %s", name, strings.Join(matches, ", "))
			findings = append(findings, finding)
		}
	}

	// Also check for large JSON responses with many fields
	if isJSONResponse(resp) {
		var jsonData interface{}
		if err := json.Unmarshal(resp.Data, &jsonData); err == nil {
			// Check if the JSON response is excessively large
			if isExcessiveJSON(jsonData) {
				findings = append(findings, "Response contains excessive JSON data")
			}
		}
	}

	return findings
}

// isJSONResponse checks if a response is JSON
func isJSONResponse(resp ffuf.Response) bool {
	return strings.Contains(resp.ContentType, "application/json")
}

// isExcessiveJSON checks if a JSON response is excessively large
func isExcessiveJSON(data interface{}) bool {
	// This is a simplified check - in a real implementation, this would be more sophisticated
	switch v := data.(type) {
	case map[string]interface{}:
		// Check if the map has too many keys
		if len(v) > 20 {
			return true
		}
		
		// Check nested objects
		for _, value := range v {
			if isExcessiveJSON(value) {
				return true
			}
		}
	case []interface{}:
		// Check if the array is too large
		if len(v) > 100 {
			return true
		}
		
		// Check a sample of array elements
		sampleSize := 5
		if len(v) < sampleSize {
			sampleSize = len(v)
		}
		for i := 0; i < sampleSize; i++ {
			if isExcessiveJSON(v[i]) {
				return true
			}
		}
	}
	
	return false
}

// isVerboseError checks if a response contains verbose error messages
func isVerboseError(resp ffuf.Response) bool {
	// Check if the response is an error
	isError := resp.StatusCode >= 400 && resp.StatusCode < 600
	
	if !isError {
		return false
	}
	
	// Check for common verbose error patterns
	responseData := string(resp.Data)
	verbosePatterns := []string{
		"Exception",
		"Error:",
		"Stack trace",
		"at ",
		"line ",
		"file ",
		"syntax error",
		"unexpected",
		"undefined",
		"null reference",
		"NullPointerException",
		"SQLException",
		"ORA-",
		"MySQL",
		"PostgreSQL",
		"MongoDB",
	}
	
	for _, pattern := range verbosePatterns {
		if strings.Contains(responseData, pattern) {
			return true
		}
	}
	
	return false
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewExcessiveDataExposureTester())
}