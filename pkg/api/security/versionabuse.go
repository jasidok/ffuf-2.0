// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// APIVersionAbuseTester implements testing for API versioning abuse
type APIVersionAbuseTester struct {
	// Configuration options
	VersionPatterns     []string
	TestVersions        []string
	TestDeprecatedAPIs  bool
	TestBetaAPIs        bool
	TestVersionDowngrade bool
	TestVersionBypass   bool
}

// NewAPIVersionAbuseTester creates a new tester for API versioning abuse
func NewAPIVersionAbuseTester() *APIVersionAbuseTester {
	return &APIVersionAbuseTester{
		VersionPatterns: []string{
			"/v[0-9]+/",
			"/v[0-9]+\\.[0-9]+/",
			"/api/v[0-9]+/",
			"/api/v[0-9]+\\.[0-9]+/",
			"api-version=[0-9]+",
			"api-version=[0-9]+\\.[0-9]+",
			"version=[0-9]+",
			"version=[0-9]+\\.[0-9]+",
		},
		TestVersions: []string{
			"v1", "v2", "v3", "v4", "v5",
			"v0.1", "v0.5", "v0.9",
			"v1.0", "v1.1", "v1.5", "v1.9",
			"v2.0", "v2.1", "v2.5",
			"v3.0", "v3.1",
			"v4.0",
			"v5.0",
			"latest", "current", "stable", "beta", "alpha", "dev", "test", "old", "legacy",
		},
		TestDeprecatedAPIs:  true,
		TestBetaAPIs:        true,
		TestVersionDowngrade: true,
		TestVersionBypass:   true,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *APIVersionAbuseTester) GetType() VulnerabilityType {
	return VulnImproperAssetsMgmt
}

// GetName returns the name of the security test
func (t *APIVersionAbuseTester) GetName() string {
	return "API Versioning Abuse"
}

// GetDescription returns a description of the security test
func (t *APIVersionAbuseTester) GetDescription() string {
	return "Tests for API endpoints that are vulnerable to versioning abuse, such as accessing deprecated API versions or bypassing security controls by switching versions."
}

// Test runs the security test against the target
func (t *APIVersionAbuseTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract the base URL from the config
	baseURL := extractBaseURL(config.Url)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Find endpoints with version patterns
	versionedEndpoints := t.findVersionedEndpoints(endpoints)

	// If no versioned endpoints found, try to generate some based on common patterns
	if len(versionedEndpoints) == 0 {
		versionedEndpoints = t.generateVersionedEndpoints(baseURL)
	}

	// Test each versioned endpoint
	for _, endpoint := range versionedEndpoints {
		// Extract the current version from the endpoint
		currentVersion := t.extractVersionFromEndpoint(endpoint)
		if currentVersion == "" {
			continue
		}

		// Test for deprecated API versions
		if t.TestDeprecatedAPIs {
			t.testDeprecatedVersions(endpoint, currentVersion, r, result)
		}

		// Test for beta/alpha API versions
		if t.TestBetaAPIs {
			t.testBetaVersions(endpoint, currentVersion, r, result)
		}

		// Test for version downgrade vulnerabilities
		if t.TestVersionDowngrade {
			t.testVersionDowngrade(endpoint, currentVersion, r, result)
		}

		// Test for version bypass vulnerabilities
		if t.TestVersionBypass {
			t.testVersionBypass(endpoint, currentVersion, r, result)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result, nil
}

// findVersionedEndpoints finds endpoints that contain version patterns
func (t *APIVersionAbuseTester) findVersionedEndpoints(endpoints []string) []string {
	var versionedEndpoints []string

	for _, endpoint := range endpoints {
		for _, pattern := range t.VersionPatterns {
			matched, err := regexp.MatchString(pattern, endpoint)
			if err == nil && matched {
				versionedEndpoints = append(versionedEndpoints, endpoint)
				break
			}
		}
	}

	return versionedEndpoints
}

// generateVersionedEndpoints generates versioned endpoints based on common patterns
func (t *APIVersionAbuseTester) generateVersionedEndpoints(baseURL string) []string {
	var endpoints []string

	// Common API path patterns
	pathPatterns := []string{
		"/api/%s/resource",
		"/api/%s/users",
		"/api/%s/data",
		"/api/%s/items",
		"/api/%s/products",
		"/%s/api/resource",
		"/%s/api/users",
		"/%s/api/data",
		"/%s/api/items",
		"/%s/api/products",
		"/api-%s/resource",
		"/api-%s/users",
		"/api-%s/data",
		"/api-%s/items",
		"/api-%s/products",
	}

	// Generate endpoints with different versions
	for _, pattern := range pathPatterns {
		for _, version := range t.TestVersions {
			endpoint := fmt.Sprintf("%s%s", baseURL, fmt.Sprintf(pattern, version))
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// extractVersionFromEndpoint extracts the version from an endpoint
func (t *APIVersionAbuseTester) extractVersionFromEndpoint(endpoint string) string {
	// Try to extract version using regex patterns
	for _, pattern := range t.VersionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(endpoint)
		if len(matches) > 0 {
			// Extract the version from the match
			version := matches[0]
			version = strings.Trim(version, "/")
			version = strings.Replace(version, "api/", "", -1)
			version = strings.Replace(version, "api-", "", -1)
			version = strings.Replace(version, "api-version=", "", -1)
			version = strings.Replace(version, "version=", "", -1)
			return version
		}
	}

	return ""
}

// testDeprecatedVersions tests for deprecated API versions
func (t *APIVersionAbuseTester) testDeprecatedVersions(endpoint, currentVersion string, r ffuf.RunnerProvider, result *TestResult) {
	// Deprecated versions are typically older versions
	deprecatedVersions := t.getDeprecatedVersions(currentVersion)

	for _, version := range deprecatedVersions {
		// Create a modified endpoint with the deprecated version
		modifiedEndpoint := t.replaceVersionInEndpoint(endpoint, currentVersion, version)
		if modifiedEndpoint == endpoint {
			continue
		}

		// Create a request for the modified endpoint
		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
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
		}
	}
}

// testBetaVersions tests for beta/alpha API versions
func (t *APIVersionAbuseTester) testBetaVersions(endpoint, currentVersion string, r ffuf.RunnerProvider, result *TestResult) {
	// Beta/alpha versions
	betaVersions := []string{"beta", "alpha", "dev", "test", "nightly", "preview", "rc", "snapshot"}

	for _, version := range betaVersions {
		// Create a modified endpoint with the beta version
		modifiedEndpoint := t.replaceVersionInEndpoint(endpoint, currentVersion, version)
		if modifiedEndpoint == endpoint {
			continue
		}

		// Create a request for the modified endpoint
		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful access to a beta version
		if isSuccessfulAccess(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnImproperAssetsMgmt,
				Name:        "Beta/Alpha API Version Accessible",
				Description: "A beta or alpha API version is publicly accessible.",
				Severity:    "Medium",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed beta/alpha API version: %s", version),
				Remediation: "Restrict access to non-production API versions. Implement proper environment separation. Use different domains or authentication for development environments.",
				CVSS:        6.0,
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
}

// testVersionDowngrade tests for version downgrade vulnerabilities
func (t *APIVersionAbuseTester) testVersionDowngrade(endpoint, currentVersion string, r ffuf.RunnerProvider, result *TestResult) {
	// Get older versions for downgrade testing
	olderVersions := t.getOlderVersions(currentVersion)

	// First, make a request to the current version to establish a baseline
	baselineReq := &ffuf.Request{
		Method: "GET",
		Url:    endpoint,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		},
	}

	baselineResp, err := r.Execute(baselineReq)
	if err != nil {
		return
	}

	// Test each older version
	for _, version := range olderVersions {
		// Create a modified endpoint with the older version
		modifiedEndpoint := t.replaceVersionInEndpoint(endpoint, currentVersion, version)
		if modifiedEndpoint == endpoint {
			continue
		}

		// Create a request for the modified endpoint
		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful access to an older version
		// and if the response is different from the baseline (indicating different behavior)
		if isSuccessfulAccess(resp) && t.isResponseDifferent(baselineResp, resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnImproperAssetsMgmt,
				Name:        "API Version Downgrade",
				Description: "The API allows downgrading to an older version that may have different security controls.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully downgraded from API version %s to %s with different behavior", currentVersion, version),
				Remediation: "Implement consistent security controls across all API versions. Consider deprecating and eventually removing older API versions with weaker security controls.",
				CVSS:        7.5,
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
}

// testVersionBypass tests for version bypass vulnerabilities
func (t *APIVersionAbuseTester) testVersionBypass(endpoint, currentVersion string, r ffuf.RunnerProvider, result *TestResult) {
	// Version bypass techniques
	bypassVersions := []string{
		"v999", "v999.999", // Extremely high version
		"v0", "v0.0", // Extremely low version
		"latest", "current", "stable", // Special version names
		"null", "undefined", "none", // Special values
		"../api", "../../api", // Path traversal
		"", // Empty version
	}

	// First, make a request to the current version to establish a baseline
	baselineReq := &ffuf.Request{
		Method: "GET",
		Url:    endpoint,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		},
	}

	// We don't need to check the baseline response for version bypass testing
	_, err := r.Execute(baselineReq)
	if err != nil {
		return
	}

	// Test each bypass version
	for _, version := range bypassVersions {
		// Create a modified endpoint with the bypass version
		modifiedEndpoint := t.replaceVersionInEndpoint(endpoint, currentVersion, version)
		if modifiedEndpoint == endpoint {
			continue
		}

		// Create a request for the modified endpoint
		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the response indicates a successful access with a bypass version
		if isSuccessfulAccess(resp) {
			// Create a vulnerability info
			vuln := VulnerabilityInfo{
				Type:        VulnImproperAssetsMgmt,
				Name:        "API Version Bypass",
				Description: "The API allows accessing resources using a non-standard version identifier.",
				Severity:    "High",
				Request:     convertToHTTPRequest(req),
				Response:    convertToHTTPResponse(resp),
				Evidence:    fmt.Sprintf("Successfully accessed API using bypass version: %s", version),
				Remediation: "Implement strict version validation. Reject requests with invalid or unexpected version identifiers. Use a whitelist of allowed versions.",
				CVSS:        7.5,
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
}

// getDeprecatedVersions returns a list of deprecated versions based on the current version
func (t *APIVersionAbuseTester) getDeprecatedVersions(currentVersion string) []string {
	var deprecatedVersions []string

	// Extract major version number
	re := regexp.MustCompile(`v?(\d+)`)
	matches := re.FindStringSubmatch(currentVersion)
	if len(matches) < 2 {
		return deprecatedVersions
	}

	// Convert to integer
	majorVersion := 0
	fmt.Sscanf(matches[1], "%d", &majorVersion)

	// Add older major versions
	for i := 1; i < majorVersion; i++ {
		deprecatedVersions = append(deprecatedVersions, fmt.Sprintf("v%d", i))
		deprecatedVersions = append(deprecatedVersions, fmt.Sprintf("v%d.0", i))
	}

	// Add older minor versions for the same major version
	re = regexp.MustCompile(`v?(\d+)\.(\d+)`)
	matches = re.FindStringSubmatch(currentVersion)
	if len(matches) >= 3 {
		majorVersion := 0
		minorVersion := 0
		fmt.Sscanf(matches[1], "%d", &majorVersion)
		fmt.Sscanf(matches[2], "%d", &minorVersion)

		for i := 0; i < minorVersion; i++ {
			deprecatedVersions = append(deprecatedVersions, fmt.Sprintf("v%d.%d", majorVersion, i))
		}
	}

	// Add special deprecated versions
	deprecatedVersions = append(deprecatedVersions, "v0", "v0.1", "v0.5", "v0.9", "old", "legacy")

	return deprecatedVersions
}

// getOlderVersions returns a list of older versions based on the current version
func (t *APIVersionAbuseTester) getOlderVersions(currentVersion string) []string {
	var olderVersions []string

	// Extract major version number
	re := regexp.MustCompile(`v?(\d+)`)
	matches := re.FindStringSubmatch(currentVersion)
	if len(matches) < 2 {
		return olderVersions
	}

	// Convert to integer
	majorVersion := 0
	fmt.Sscanf(matches[1], "%d", &majorVersion)

	// Add older major versions
	for i := 1; i < majorVersion; i++ {
		olderVersions = append(olderVersions, fmt.Sprintf("v%d", i))
		olderVersions = append(olderVersions, fmt.Sprintf("v%d.0", i))
	}

	// Add older minor versions for the same major version
	re = regexp.MustCompile(`v?(\d+)\.(\d+)`)
	matches = re.FindStringSubmatch(currentVersion)
	if len(matches) >= 3 {
		majorVersion := 0
		minorVersion := 0
		fmt.Sscanf(matches[1], "%d", &majorVersion)
		fmt.Sscanf(matches[2], "%d", &minorVersion)

		for i := 0; i < minorVersion; i++ {
			olderVersions = append(olderVersions, fmt.Sprintf("v%d.%d", majorVersion, i))
		}
	}

	return olderVersions
}

// replaceVersionInEndpoint replaces the version in an endpoint with a new version
func (t *APIVersionAbuseTester) replaceVersionInEndpoint(endpoint, oldVersion, newVersion string) string {
	// Replace version in URL path
	patterns := []string{
		fmt.Sprintf("/v%s/", strings.TrimPrefix(oldVersion, "v")),
		fmt.Sprintf("/%s/", oldVersion),
		fmt.Sprintf("/api/v%s/", strings.TrimPrefix(oldVersion, "v")),
		fmt.Sprintf("/api/%s/", oldVersion),
		fmt.Sprintf("/api-v%s/", strings.TrimPrefix(oldVersion, "v")),
		fmt.Sprintf("/api-%s/", oldVersion),
	}

	replacements := []string{
		fmt.Sprintf("/v%s/", strings.TrimPrefix(newVersion, "v")),
		fmt.Sprintf("/%s/", newVersion),
		fmt.Sprintf("/api/v%s/", strings.TrimPrefix(newVersion, "v")),
		fmt.Sprintf("/api/%s/", newVersion),
		fmt.Sprintf("/api-v%s/", strings.TrimPrefix(newVersion, "v")),
		fmt.Sprintf("/api-%s/", newVersion),
	}

	modifiedEndpoint := endpoint
	for i := range patterns {
		modifiedEndpoint = strings.Replace(modifiedEndpoint, patterns[i], replacements[i], -1)
	}

	// Replace version in query parameters
	queryParams := []string{
		fmt.Sprintf("api-version=%s", oldVersion),
		fmt.Sprintf("version=%s", oldVersion),
	}

	queryReplacements := []string{
		fmt.Sprintf("api-version=%s", newVersion),
		fmt.Sprintf("version=%s", newVersion),
	}

	for i := range queryParams {
		modifiedEndpoint = strings.Replace(modifiedEndpoint, queryParams[i], queryReplacements[i], -1)
	}

	return modifiedEndpoint
}

// isResponseDifferent checks if two responses are significantly different
func (t *APIVersionAbuseTester) isResponseDifferent(resp1, resp2 ffuf.Response) bool {
	// Check if status codes are different
	if resp1.StatusCode != resp2.StatusCode {
		return true
	}

	// Check if content lengths are significantly different (more than 10% difference)
	if resp1.ContentLength > 0 && resp2.ContentLength > 0 {
		ratio := float64(resp1.ContentLength) / float64(resp2.ContentLength)
		if ratio < 0.9 || ratio > 1.1 {
			return true
		}
	}

	// Check if content types are different
	if resp1.ContentType != resp2.ContentType {
		return true
	}

	// Check if response bodies are different (simple check)
	if len(resp1.Data) > 0 && len(resp2.Data) > 0 {
		// Compare first 100 characters
		maxLen := 100
		if len(resp1.Data) < maxLen {
			maxLen = len(resp1.Data)
		}
		if len(resp2.Data) < maxLen {
			maxLen = len(resp2.Data)
		}

		if string(resp1.Data[:maxLen]) != string(resp2.Data[:maxLen]) {
			return true
		}
	}

	return false
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewAPIVersionAbuseTester())
}
