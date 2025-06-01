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

// BrokenFunctionLevelAuthTester implements testing for Broken Function Level Authorization (API5:2019)
type BrokenFunctionLevelAuthTester struct {
	// Configuration options
	AdminEndpoints     []string
	AdminMethods       []string
	UserRoles          []string
	RolePermissionMap  map[string][]string
	MethodPermissionMap map[string][]string
}

// NewBrokenFunctionLevelAuthTester creates a new tester for Broken Function Level Authorization
func NewBrokenFunctionLevelAuthTester() *BrokenFunctionLevelAuthTester {
	return &BrokenFunctionLevelAuthTester{
		AdminEndpoints: []string{
			"admin", "management", "console", "dashboard", "config", "settings",
			"users", "roles", "permissions", "accounts", "system", "internal",
		},
		AdminMethods: []string{
			"DELETE", "PUT", "PATCH", "POST",
		},
		UserRoles: []string{
			"user", "customer", "guest", "anonymous", "public",
		},
		RolePermissionMap: map[string][]string{
			"admin": {"read", "write", "delete", "manage"},
			"user":  {"read", "write_own"},
			"guest": {"read"},
		},
		MethodPermissionMap: map[string][]string{
			"GET":    {"read"},
			"POST":   {"write", "write_own"},
			"PUT":    {"write", "write_own"},
			"PATCH":  {"write", "write_own"},
			"DELETE": {"delete"},
		},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *BrokenFunctionLevelAuthTester) GetType() VulnerabilityType {
	return VulnBrokenFunctionLevelAuth
}

// GetName returns the name of the security test
func (t *BrokenFunctionLevelAuthTester) GetName() string {
	return "Broken Function Level Authorization"
}

// GetDescription returns a description of the security test
func (t *BrokenFunctionLevelAuthTester) GetDescription() string {
	return "Tests for API endpoints that don't properly verify that the requesting user has the necessary permissions to perform the requested function."
}

// Test runs the security test against the target
func (t *BrokenFunctionLevelAuthTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for function level authorization vulnerabilities
	for _, endpoint := range endpoints {
		// Test for admin endpoints accessible without admin privileges
		if t.isAdminEndpoint(endpoint) {
			t.testAdminEndpoint(endpoint, r, result)
		}

		// Test for sensitive methods accessible without proper authorization
		t.testSensitiveMethods(endpoint, r, result)

		// Test for horizontal privilege escalation
		t.testHorizontalPrivilegeEscalation(endpoint, r, result)

		// Test for vertical privilege escalation
		t.testVerticalPrivilegeEscalation(endpoint, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// isAdminEndpoint checks if an endpoint looks like an admin endpoint
func (t *BrokenFunctionLevelAuthTester) isAdminEndpoint(endpoint string) bool {
	for _, pattern := range t.AdminEndpoints {
		if strings.Contains(strings.ToLower(endpoint), pattern) {
			return true
		}
	}
	return false
}

// testAdminEndpoint tests if an admin endpoint is accessible without admin privileges
func (t *BrokenFunctionLevelAuthTester) testAdminEndpoint(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Try to access the endpoint without admin credentials
	req := &ffuf.Request{
		Method: "GET",
		Url:    endpoint,
		Headers: map[string]string{
			"X-Role": "user", // Simulate a non-admin user
		},
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnBrokenFunctionLevelAuth,
			Name:        "Admin Endpoint Accessible",
			Description: "An administrative endpoint is accessible without admin privileges.",
			Severity:    "Critical",
			Request:     convertToHTTPRequest(req),
			Response:    convertToHTTPResponse(resp),
			Evidence:    fmt.Sprintf("Successfully accessed admin endpoint %s with user role", endpoint),
			Remediation: "Implement proper function level authorization checks. Verify that the user has the necessary role or permissions to access administrative endpoints.",
			CVSS:        9.0,
			CWE:         "CWE-285",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa5-broken-function-level-authorization/",
				"https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}
}

// testSensitiveMethods tests if sensitive HTTP methods are accessible without proper authorization
func (t *BrokenFunctionLevelAuthTester) testSensitiveMethods(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	for _, method := range t.AdminMethods {
		// Try to access the endpoint with a sensitive method
		req := &ffuf.Request{
			Method: method,
			Url:    endpoint,
			Headers: map[string]string{
				"X-Role": "user", // Simulate a non-admin user
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnBrokenFunctionLevelAuth,
				Name:        "Sensitive Method Accessible",
				Description: "A sensitive HTTP method is accessible without proper authorization.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed endpoint %s with method %s using user role", endpoint, method),
				Remediation: "Implement proper function level authorization checks. Verify that the user has the necessary permissions to perform sensitive operations.",
				CVSS:        8.0,
				CWE:         "CWE-285",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa5-broken-function-level-authorization/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testHorizontalPrivilegeEscalation tests for horizontal privilege escalation vulnerabilities
func (t *BrokenFunctionLevelAuthTester) testHorizontalPrivilegeEscalation(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Skip endpoints that don't look like they would have user-specific resources
	if !containsIDPattern(endpoint) {
		return
	}

	// Try to access the endpoint with a different user ID
	originalEndpoint := endpoint
	modifiedEndpoint := replaceIDInEndpoint(endpoint, "different_user_id")
	
	if modifiedEndpoint == originalEndpoint {
		return // No ID was replaced
	}

	req := &ffuf.Request{
		Method: "GET",
		Url:    modifiedEndpoint,
		Headers: map[string]string{
			"X-User-ID": "original_user_id", // Simulate the original user
			"X-Role":    "user",
		},
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return
	}

	// Check if the request was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnBrokenFunctionLevelAuth,
			Name:        "Horizontal Privilege Escalation",
			Description: "A user can access resources belonging to another user of the same privilege level.",
			Severity:    "High",
			Request:     convertToHTTPRequest(req),
			Response:    convertToHTTPResponse(resp),
			Evidence:    fmt.Sprintf("Successfully accessed endpoint %s with user ID different_user_id while authenticated as original_user_id", modifiedEndpoint),
			Remediation: "Implement proper function level authorization checks. Verify that the user can only access their own resources.",
			CVSS:        8.0,
			CWE:         "CWE-639",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa5-broken-function-level-authorization/",
				"https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}
}

// testVerticalPrivilegeEscalation tests for vertical privilege escalation vulnerabilities
func (t *BrokenFunctionLevelAuthTester) testVerticalPrivilegeEscalation(endpoint string, r ffuf.RunnerProvider, result *TestResult) {
	// Try to access the endpoint with a higher privilege level operation
	for _, role := range t.UserRoles {
		// Get the permissions for this role
		rolePermissions, roleExists := t.RolePermissionMap[role]
		if !roleExists {
			continue
		}

		for method, methodPermissions := range t.MethodPermissionMap {
			// Check if this method requires permissions that the role doesn't have
			hasPermission := false
			for _, methodPerm := range methodPermissions {
				for _, rolePerm := range rolePermissions {
					if methodPerm == rolePerm {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}

			// If the role doesn't have permission for this method, test it
			if !hasPermission {
				req := &ffuf.Request{
					Method: method,
					Url:    endpoint,
					Headers: map[string]string{
						"X-Role": role,
					},
				}

				// Execute the request
				resp, err := r.Execute(req)
				if err != nil {
					continue
				}

				// Check if the request was successful
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					// Create a vulnerability info
					vuln := VulnerabilityInfo{
						Type:        VulnBrokenFunctionLevelAuth,
						Name:        "Vertical Privilege Escalation",
						Description: "A user can perform operations requiring higher privileges than they possess.",
						Severity:    "Critical",
						Request:     convertToHTTPRequest(req),
						Response:    convertToHTTPResponse(resp),
						Evidence:    fmt.Sprintf("Successfully performed %s operation on endpoint %s with role %s, which should not have the required permissions", method, endpoint, role),
						Remediation: "Implement proper function level authorization checks. Verify that the user has the necessary role or permissions to perform the requested operation.",
						CVSS:        9.0,
						CWE:         "CWE-269",
						References: []string{
							"https://owasp.org/API-Security/editions/2019/en/0xa5-broken-function-level-authorization/",
							"https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html",
						},
						DetectedAt: time.Now(),
					}
					result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				}
			}
		}
	}
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewBrokenFunctionLevelAuthTester())
}