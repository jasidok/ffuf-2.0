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

// ImproperAssetsMgmtTester implements testing for Improper Assets Management (API9:2019)
type ImproperAssetsMgmtTester struct {
	// Configuration options
	DeprecatedVersions     []string
	BetaEndpoints          []string
	DebugEndpoints         []string
	BackupFiles            []string
	CommonVulnerablePaths  []string
	APIVersionsToTest      []string
	CheckMultipleVersions  bool
	CheckUnpublishedAPIs   bool
	CheckDeprecatedFeatures bool
}

// NewImproperAssetsMgmtTester creates a new tester for Improper Assets Management
func NewImproperAssetsMgmtTester() *ImproperAssetsMgmtTester {
	return &ImproperAssetsMgmtTester{
		DeprecatedVersions: []string{
			"v1", "v1.0", "v1.0.0", "v0", "v0.1", "v0.0.1", "beta", "alpha", "legacy", "old",
		},
		BetaEndpoints: []string{
			"beta", "dev", "development", "staging", "test", "testing", "uat", "sandbox",
		},
		DebugEndpoints: []string{
			"debug", "trace", "status", "health", "ping", "metrics", "stats", "admin",
		},
		BackupFiles: []string{
			".bak", ".backup", ".old", ".save", ".swp", ".copy", ".tmp", ".temp", "~",
		},
		CommonVulnerablePaths: []string{
			".git", ".svn", ".env", ".htaccess", "config", "settings", "backup", "admin",
			"console", "dashboard", "manage", "management", "phpinfo.php", "info.php",
		},
		APIVersionsToTest: []string{
			"v1", "v2", "v3", "v1.0", "v1.1", "v2.0", "v2.1", "v3.0", "latest",
		},
		CheckMultipleVersions:  true,
		CheckUnpublishedAPIs:   true,
		CheckDeprecatedFeatures: true,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *ImproperAssetsMgmtTester) GetType() VulnerabilityType {
	return VulnImproperAssetsMgmt
}

// GetName returns the name of the security test
func (t *ImproperAssetsMgmtTester) GetName() string {
	return "Improper Assets Management"
}

// GetDescription returns a description of the security test
func (t *ImproperAssetsMgmtTester) GetDescription() string {
	return "Tests for API endpoints that expose deprecated API versions, debug endpoints, or other assets that should not be publicly accessible."
}

// Test runs the security test against the target
func (t *ImproperAssetsMgmtTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract the base URL from the config
	baseURL := extractBaseURL(config.Url)

	// Test for deprecated API versions
	if t.CheckDeprecatedFeatures {
		t.testDeprecatedVersions(baseURL, r, result)
	}

	// Test for beta/development endpoints
	if t.CheckUnpublishedAPIs {
		t.testBetaEndpoints(baseURL, r, result)
	}

	// Test for debug endpoints
	t.testDebugEndpoints(baseURL, r, result)

	// Test for backup files
	t.testBackupFiles(baseURL, r, result)

	// Test for common vulnerable paths
	t.testCommonVulnerablePaths(baseURL, r, result)

	// Test for multiple API versions
	if t.CheckMultipleVersions {
		t.testMultipleAPIVersions(baseURL, r, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// testDeprecatedVersions tests for deprecated API versions
func (t *ImproperAssetsMgmtTester) testDeprecatedVersions(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, version := range t.DeprecatedVersions {
		// Create URLs with different version patterns
		testURLs := []string{
			fmt.Sprintf("%s/api/%s/", baseURL, version),
			fmt.Sprintf("%s/%s/api/", baseURL, version),
			fmt.Sprintf("%s/api-%s/", baseURL, version),
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

			// Check if the response indicates a successful access to a deprecated version
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnImproperAssetsMgmt,
					Name:        "Deprecated API Version Accessible",
					Description: "A deprecated API version is publicly accessible.",
					Severity:    "Medium",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed deprecated API version: %s", version),
					Remediation: "Properly retire and decommission old API versions. Implement API lifecycle management. Redirect clients to newer API versions.",
					CVSS:        6.5,
					CWE:         "CWE-1059",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
						"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more URLs for this version
			}
		}
	}
}

// testBetaEndpoints tests for beta/development endpoints
func (t *ImproperAssetsMgmtTester) testBetaEndpoints(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, endpoint := range t.BetaEndpoints {
		// Create URLs with different patterns
		testURLs := []string{
			fmt.Sprintf("%s/%s/", baseURL, endpoint),
			fmt.Sprintf("%s/api/%s/", baseURL, endpoint),
			fmt.Sprintf("%s/%s-api/", baseURL, endpoint),
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

			// Check if the response indicates a successful access to a beta endpoint
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnImproperAssetsMgmt,
					Name:        "Beta/Development API Endpoint Accessible",
					Description: "A beta or development API endpoint is publicly accessible.",
					Severity:    "Medium",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed beta/development endpoint: %s", endpoint),
					Remediation: "Restrict access to non-production API endpoints. Implement proper environment separation. Use different domains or authentication for development environments.",
					CVSS:        6.0,
					CWE:         "CWE-1059",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
						"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more URLs for this endpoint
			}
		}
	}
}

// testDebugEndpoints tests for debug endpoints
func (t *ImproperAssetsMgmtTester) testDebugEndpoints(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, endpoint := range t.DebugEndpoints {
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

			// Check if the response indicates a successful access to a debug endpoint
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnImproperAssetsMgmt,
					Name:        "Debug Endpoint Accessible",
					Description: "A debug or administrative endpoint is publicly accessible.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed debug endpoint: %s", endpoint),
					Remediation: "Restrict access to debug and administrative endpoints. Implement proper authentication and authorization. Consider removing debug endpoints in production.",
					CVSS:        7.5,
					CWE:         "CWE-215",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
						"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break // Found a vulnerability, no need to test more URLs for this endpoint
			}
		}
	}
}

// testBackupFiles tests for backup files
func (t *ImproperAssetsMgmtTester) testBackupFiles(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	// Common files to check for backups
	filesToCheck := []string{
		"config.js", "config.php", "config.xml", "config.json",
		"settings.js", "settings.php", "settings.xml", "settings.json",
		"app.js", "app.php", "app.config.js", "web.config",
		"api.js", "api.php", "api.config.js", "api.json",
	}

	for _, file := range filesToCheck {
		for _, ext := range t.BackupFiles {
			testURL := fmt.Sprintf("%s/%s%s", baseURL, file, ext)
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

			// Check if the response indicates a successful access to a backup file
			if isSuccessfulAccess(resp) {
				// Create a vulnerability info
				vuln := VulnerabilityInfo{
					Type:        VulnImproperAssetsMgmt,
					Name:        "Backup File Accessible",
					Description: "A backup file is publicly accessible.",
					Severity:    "High",
					Request:     convertToHTTPRequest(req),
					Response:    convertToHTTPResponse(resp),
					Evidence:    fmt.Sprintf("Successfully accessed backup file: %s%s", file, ext),
					Remediation: "Remove backup files from production servers. Implement proper file permissions. Use a web application firewall to block access to backup files.",
					CVSS:        7.5,
					CWE:         "CWE-530",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
						"https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html",
					},
					DetectedAt: time.Now(),
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}
}

// testCommonVulnerablePaths tests for common vulnerable paths
func (t *ImproperAssetsMgmtTester) testCommonVulnerablePaths(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	for _, path := range t.CommonVulnerablePaths {
		testURL := fmt.Sprintf("%s/%s", baseURL, path)
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

		// Check if the response indicates a successful access to a vulnerable path
		if isSuccessfulAccess(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnImproperAssetsMgmt,
				Name:        "Vulnerable Path Accessible",
				Description: "A potentially vulnerable path is publicly accessible.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed vulnerable path: %s", path),
				Remediation: "Restrict access to sensitive paths. Implement proper authentication and authorization. Remove unnecessary files and directories from production servers.",
				CVSS:        7.5,
				CWE:         "CWE-284",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
					"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}
}

// testMultipleAPIVersions tests for multiple API versions
func (t *ImproperAssetsMgmtTester) testMultipleAPIVersions(baseURL string, r ffuf.RunnerProvider, result *TestResult) {
	accessibleVersions := []string{}

	for _, version := range t.APIVersionsToTest {
		// Create URLs with different version patterns
		testURLs := []string{
			fmt.Sprintf("%s/api/%s/", baseURL, version),
			fmt.Sprintf("%s/%s/api/", baseURL, version),
			fmt.Sprintf("%s/api-%s/", baseURL, version),
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

			// Check if the response indicates a successful access to an API version
			if isSuccessfulAccess(resp) {
				accessibleVersions = append(accessibleVersions, version)
				break // Found an accessible version, no need to test more URLs for this version
			}
		}
	}

	// If multiple API versions are accessible, report a vulnerability
	if len(accessibleVersions) > 1 {
		// Create a vulnerability info
		vuln := VulnerabilityInfo{
			Type:        VulnImproperAssetsMgmt,
			Name:        "Multiple API Versions Accessible",
			Description: "Multiple API versions are publicly accessible.",
			Severity:    "Medium",
			Request:     nil, // No specific request to include
			Response:    nil, // No specific response to include
			Evidence:    fmt.Sprintf("Successfully accessed multiple API versions: %s", strings.Join(accessibleVersions, ", ")),
			Remediation: "Implement proper API lifecycle management. Deprecate and eventually retire old API versions. Redirect clients to newer API versions.",
			CVSS:        5.5,
			CWE:         "CWE-1059",
			References: []string{
				"https://owasp.org/API-Security/editions/2019/en/0xa9-improper-assets-management/",
				"https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html",
			},
			DetectedAt: time.Now(),
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}
}


func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewImproperAssetsMgmtTester())
}
