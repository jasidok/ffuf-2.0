// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// SecurityMisconfigTester implements testing for Security Misconfiguration (API7:2019)
type SecurityMisconfigTester struct {
	// Configuration options
	InsecureHeaders      map[string]string
	MissingHeaders       []string
	DangerousMethods     []string
	DefaultCredentials   []struct{ Username, Password string }
	CommonDebugEndpoints []string
}

// NewSecurityMisconfigTester creates a new tester for Security Misconfiguration
func NewSecurityMisconfigTester() *SecurityMisconfigTester {
	return &SecurityMisconfigTester{
		InsecureHeaders: map[string]string{
			"X-Frame-Options":        "",
			"X-Content-Type-Options": "",
			"Content-Security-Policy": "",
			"Strict-Transport-Security": "",
			"X-XSS-Protection":       "",
			"Cache-Control":          "no-store, no-cache",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS, HEAD, PATCH",
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Credentials": "true",
			"Server":                 "",
			"X-Powered-By":           "",
		},
		MissingHeaders: []string{
			"X-Frame-Options",
			"X-Content-Type-Options",
			"Content-Security-Policy",
			"Strict-Transport-Security",
			"X-XSS-Protection",
		},
		DangerousMethods: []string{
			"TRACE",
			"OPTIONS",
			"PUT",
			"DELETE",
			"CONNECT",
			"PATCH",
		},
		DefaultCredentials: []struct{ Username, Password string }{
			{"admin", "admin"},
			{"admin", "password"},
			{"admin", "123456"},
			{"root", "root"},
			{"root", "password"},
			{"user", "user"},
			{"user", "password"},
			{"test", "test"},
			{"guest", "guest"},
			{"demo", "demo"},
			{"default", "default"},
			{"superuser", "superuser"},
		},
		CommonDebugEndpoints: []string{
			"/debug",
			"/debug/vars",
			"/debug/pprof",
			"/status",
			"/health",
			"/metrics",
			"/actuator",
			"/actuator/health",
			"/actuator/info",
			"/actuator/metrics",
			"/actuator/env",
			"/actuator/trace",
			"/api/debug",
			"/api/status",
			"/api/health",
			"/api/metrics",
			"/admin",
			"/admin/console",
			"/admin/status",
			"/admin/metrics",
			"/console",
			"/swagger",
			"/swagger-ui",
			"/swagger-ui.html",
			"/api-docs",
			"/api/docs",
			"/graphiql",
			"/graphql",
			"/graphql-explorer",
			"/.git",
			"/.env",
			"/.config",
			"/config",
			"/configuration",
			"/settings",
			"/system",
			"/logs",
			"/log",
			"/trace",
			"/stats",
			"/server-status",
			"/server-info",
			"/phpinfo.php",
			"/info.php",
		},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *SecurityMisconfigTester) GetType() VulnerabilityType {
	return VulnSecurityMisconfig
}

// GetName returns the name of the security test
func (t *SecurityMisconfigTester) GetName() string {
	return "Security Misconfiguration"
}

// GetDescription returns a description of the security test
func (t *SecurityMisconfigTester) GetDescription() string {
	return "Tests for API security misconfigurations, such as insecure default configurations, incomplete or ad-hoc configurations, open cloud storage, misconfigured HTTP headers, unnecessary HTTP methods, permissive CORS, and verbose error messages."
}

// Test runs the security test against the target
func (t *SecurityMisconfigTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract the base URL from the config
	baseURL := extractBaseURL(config.Url)

	// Test for insecure HTTP headers
	t.testInsecureHeaders(baseURL, r, result)

	// Test for dangerous HTTP methods
	t.testDangerousMethods(baseURL, r, result)

	// Test for default credentials
	t.testDefaultCredentials(baseURL, r, result)

	// Test for common debug endpoints
	t.testDebugEndpoints(baseURL, r, result)

	// Test for CORS misconfiguration
	t.testCORSMisconfiguration(baseURL, r, result)

	// Test for TLS misconfiguration
	t.testTLSMisconfiguration(baseURL, r, result)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testInsecureHeaders tests for insecure HTTP headers
func (t *SecurityMisconfigTester) testInsecureHeaders(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Create a request to check headers
	req := &ffuf.Request{
		Method: "GET",
		Url:    baseURL,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		},
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check for missing security headers
	var missingHeaders []string
	for _, header := range t.MissingHeaders {
		if _, exists := resp.Headers[header]; !exists {
			missingHeaders = append(missingHeaders, header)
		}
	}

	if len(missingHeaders) > 0 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnSecurityMisconfig,
			Name:        "Missing Security Headers",
			Description: "The API is missing important security headers that help protect against common web vulnerabilities.",
			Severity:    "Medium",
			Request:     convertToHTTPRequest(req),
			Response:    convertToHTTPResponse(resp),
			Evidence:    fmt.Sprintf("Missing headers: %s", strings.Join(missingHeaders, ", ")),
			Remediation: "Configure the server to include all necessary security headers. Consider using a security header middleware or framework that automatically adds these headers.",
			CVSS:        5.0,
			CWE:         "CWE-16",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
				"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				"https://owasp.org/www-project-secure-headers/",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	// Check for insecure header values
	var insecureHeaders []string
	for header, insecureValue := range t.InsecureHeaders {
		if value, exists := resp.Headers[header]; exists {
			if insecureValue != "" && strings.Contains(value[0], insecureValue) {
				insecureHeaders = append(insecureHeaders, fmt.Sprintf("%s: %s", header, value[0]))
			}
		}
	}

	if len(insecureHeaders) > 0 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnSecurityMisconfig,
			Name:        "Insecure Header Values",
			Description: "The API is using insecure values for security headers.",
			Severity:    "Medium",
			Request:     convertToHTTPRequest(req),
			Response:    convertToHTTPResponse(resp),
			Evidence:    fmt.Sprintf("Insecure headers: %s", strings.Join(insecureHeaders, ", ")),
			Remediation: "Configure the server to use secure values for security headers. Avoid using wildcard values for CORS headers and ensure proper restrictions are in place.",
			CVSS:        5.0,
			CWE:         "CWE-16",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
				"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				"https://owasp.org/www-project-secure-headers/",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	// Check for information disclosure headers
	var infoHeaders []string
	for _, header := range []string{"Server", "X-Powered-By", "X-AspNet-Version", "X-AspNetMvc-Version"} {
		if value, exists := resp.Headers[header]; exists {
			infoHeaders = append(infoHeaders, fmt.Sprintf("%s: %s", header, value[0]))
		}
	}

	if len(infoHeaders) > 0 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnSecurityMisconfig,
			Name:        "Information Disclosure Headers",
			Description: "The API is disclosing potentially sensitive information through HTTP headers.",
			Severity:    "Low",
			Request:     convertToHTTPRequest(req),
			Response:    convertToHTTPResponse(resp),
			Evidence:    fmt.Sprintf("Information disclosure headers: %s", strings.Join(infoHeaders, ", ")),
			Remediation: "Configure the server to remove or obfuscate headers that reveal information about the technology stack. This helps prevent attackers from targeting known vulnerabilities in specific technologies.",
			CVSS:        3.0,
			CWE:         "CWE-200",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
				"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}
}

// testDangerousMethods tests for dangerous HTTP methods
func (t *SecurityMisconfigTester) testDangerousMethods(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, method := range t.DangerousMethods {
		// Create a request with the dangerous method
		req := &ffuf.Request{
			Method: method,
			Url:    baseURL,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the method is allowed
		if resp.StatusCode != 405 && resp.StatusCode != 501 && resp.StatusCode != 403 {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnSecurityMisconfig,
				Name:        "Dangerous HTTP Method Enabled",
				Description: fmt.Sprintf("The API allows the potentially dangerous HTTP method: %s", method),
				Severity:    "Medium",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("HTTP method %s returned status code %d", method, resp.StatusCode),
				Remediation: "Disable unnecessary HTTP methods. Configure the server to only allow the HTTP methods that are required for the API to function properly.",
				CVSS:        5.0,
				CWE:         "CWE-16",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testDefaultCredentials tests for default credentials
func (t *SecurityMisconfigTester) testDefaultCredentials(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Look for potential login endpoints
	loginEndpoints := []string{
		"/login",
		"/auth",
		"/authenticate",
		"/signin",
		"/sign-in",
		"/api/login",
		"/api/auth",
		"/api/authenticate",
		"/api/signin",
		"/api/sign-in",
		"/admin/login",
		"/admin/auth",
		"/admin",
		"/user/login",
		"/account/login",
	}

	for _, endpoint := range loginEndpoints {
		loginURL := baseURL
		if !strings.HasSuffix(loginURL, "/") && !strings.HasPrefix(endpoint, "/") {
			loginURL += "/"
		}
		loginURL += endpoint

		for _, cred := range t.DefaultCredentials {
			username := cred.Username
			password := cred.Password
			// Create a JSON login payload
			payload := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

			// Create a request with the default credentials
			req := &ffuf.Request{
				Method: "POST",
				Url:    loginURL,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
				Data: []byte(payload),
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				continue
			}

			// Check if the login was successful
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// Look for success indicators in the response
				responseData := string(resp.Data)
				successIndicators := []string{
					"token", "jwt", "auth", "success", "welcome", "logged in", "session",
					"user", "profile", "account", "dashboard", "admin",
				}

				for _, indicator := range successIndicators {
					if strings.Contains(strings.ToLower(responseData), indicator) {
						// Create a vulnerability info
						vuln := VulnerabilityInfo{
							Type:        VulnSecurityMisconfig,
							Name:        "Default Credentials",
							Description: "The API accepts default or commonly used credentials.",
							Severity:    "Critical",
							Request:     convertToHTTPRequest(req),
							Response:    convertToHTTPResponse(resp),
							Evidence:    fmt.Sprintf("Successfully authenticated with username '%s' and password '%s'", username, password),
							Remediation: "Ensure that all default credentials are changed. Implement strong password policies and consider using multi-factor authentication for sensitive accounts.",
							CVSS:        9.0,
							CWE:         "CWE-16",
							References: []string{
								"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
								"https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html",
							},
							DetectedAt: time.Now(),
						}
						result.Vulnerabilities = append(result.Vulnerabilities, vuln)
						break
					}
				}
			}
		}
	}
}

// testDebugEndpoints tests for common debug endpoints
func (t *SecurityMisconfigTester) testDebugEndpoints(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, endpoint := range t.CommonDebugEndpoints {
		debugURL := baseURL
		if !strings.HasSuffix(debugURL, "/") && !strings.HasPrefix(endpoint, "/") {
			debugURL += "/"
		}
		debugURL += endpoint

		// Create a request for the debug endpoint
		req := &ffuf.Request{
			Method: "GET",
			Url:    debugURL,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the endpoint is accessible
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnSecurityMisconfig,
				Name:        "Debug Endpoint Exposed",
				Description: fmt.Sprintf("The API exposes a debug endpoint: %s", endpoint),
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Debug endpoint %s is accessible and returned status code %d", endpoint, resp.StatusCode),
				Remediation: "Disable or properly secure debug endpoints in production environments. Consider using environment-specific configurations to ensure that debug features are only enabled in development environments.",
				CVSS:        7.0,
				CWE:         "CWE-16",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testCORSMisconfiguration tests for CORS misconfiguration
func (t *SecurityMisconfigTester) testCORSMisconfiguration(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Create a request with an Origin header
	req := &ffuf.Request{
		Method: "GET",
		Url:    baseURL,
		Headers: map[string]string{
			"Origin":     "https://evil.com",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		},
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check for permissive CORS headers
	if value, exists := resp.Headers["Access-Control-Allow-Origin"]; exists {
		if value[0] == "*" || value[0] == "https://evil.com" {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnSecurityMisconfig,
				Name:        "Permissive CORS Configuration",
				Description: "The API has a permissive CORS configuration that allows requests from any origin.",
				Severity:    "Medium",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Access-Control-Allow-Origin: %s", value[0]),
				Remediation: "Configure CORS to only allow requests from trusted origins. Avoid using wildcard (*) values for Access-Control-Allow-Origin. Consider implementing a whitelist of allowed origins.",
				CVSS:        5.0,
				CWE:         "CWE-16",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					"https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	// Check for permissive credentials configuration
	if value, exists := resp.Headers["Access-Control-Allow-Credentials"]; exists {
		if value[0] == "true" {
			// Check if Access-Control-Allow-Origin is also permissive
			if origin, exists := resp.Headers["Access-Control-Allow-Origin"]; exists {
				if origin[0] == "*" || origin[0] == "https://evil.com" {
					// Create a vulnerability info
					vuln := VulnerabilityInfo{
						Type:        VulnSecurityMisconfig,
						Name:        "Permissive CORS Credentials Configuration",
						Description: "The API allows credentials to be sent with cross-origin requests from any origin.",
						Severity:    "High",
						Request:     convertToHTTPRequest(req),
						Response:    convertToHTTPResponse(resp),
						Evidence:    fmt.Sprintf("Access-Control-Allow-Credentials: true, Access-Control-Allow-Origin: %s", origin[0]),
						Remediation: "Configure CORS to only allow credentials from trusted origins. Never use wildcard (*) values for Access-Control-Allow-Origin when Access-Control-Allow-Credentials is set to true.",
						CVSS:        7.0,
						CWE:         "CWE-16",
						References: []string{
							"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
							"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
							"https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html",
						},
						DetectedAt: time.Now(),
					}
					result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				}
			}
		}
	}
}

// testTLSMisconfiguration tests for TLS misconfiguration
func (t *SecurityMisconfigTester) testTLSMisconfiguration(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Check if the API is available over HTTP
	if strings.HasPrefix(baseURL, "https://") {
		httpURL := strings.Replace(baseURL, "https://", "http://", 1)

		// Create a request for the HTTP URL
		req := &ffuf.Request{
			Method: "GET",
			Url:    httpURL,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			return
		}

		// Check if the HTTP endpoint is accessible
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnSecurityMisconfig,
				Name:        "Insecure HTTP Access",
				Description: "The API is accessible over unencrypted HTTP.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("HTTP endpoint %s is accessible and returned status code %d", httpURL, resp.StatusCode),
				Remediation: "Configure the server to redirect all HTTP traffic to HTTPS. Consider implementing HTTP Strict Transport Security (HSTS) to ensure that clients always use HTTPS.",
				CVSS:        7.0,
				CWE:         "CWE-319",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa7-security-misconfiguration/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					"https://cheatsheetseries.owasp.org/cheatsheets/Transport_Layer_Protection_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// extractBaseURL extracts the base URL from a URL
func extractBaseURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// Return the scheme and host
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewSecurityMisconfigTester())
}
