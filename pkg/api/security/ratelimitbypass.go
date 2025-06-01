// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// RateLimitBypassTester implements testing for rate limiting bypass techniques
type RateLimitBypassTester struct {
	// Configuration options
	RequestsPerTest     int
	TimeBetweenRequests time.Duration
	ConcurrentRequests  int
	BypassTechniques    []string
}

// NewRateLimitBypassTester creates a new tester for rate limiting bypass techniques
func NewRateLimitBypassTester() *RateLimitBypassTester {
	return &RateLimitBypassTester{
		RequestsPerTest:     30,
		TimeBetweenRequests: time.Millisecond * 100,
		ConcurrentRequests:  5,
		BypassTechniques: []string{
			"ip-rotation",
			"header-manipulation",
			"parameter-pollution",
			"http-method-switching",
			"distributed-attack",
			"cache-manipulation",
		},
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *RateLimitBypassTester) GetType() VulnerabilityType {
	return VulnLackOfResources
}

// GetName returns the name of the security test
func (t *RateLimitBypassTester) GetName() string {
	return "Rate Limiting Bypass"
}

// GetDescription returns a description of the security test
func (t *RateLimitBypassTester) GetDescription() string {
	return "Tests for API endpoints that have rate limiting mechanisms that can be bypassed using various techniques."
}

// Test runs the security test against the target
func (t *RateLimitBypassTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for rate limiting bypass vulnerabilities
	for _, endpoint := range endpoints {
		// First, check if the endpoint has rate limiting
		if !t.hasRateLimiting(endpoint, r) {
			// If there's no rate limiting, no need to test bypass techniques
			continue
		}

		// Test each bypass technique
		for _, technique := range t.BypassTechniques {
			if t.testBypassTechnique(endpoint, technique, r, result) {
				// If a bypass technique works, add a vulnerability
				vuln := VulnerabilityInfo{
					Type:        VulnLackOfResources,
					Name:        fmt.Sprintf("Rate Limiting Bypass (%s)", t.getTechniqueName(technique)),
					Description: fmt.Sprintf("The API endpoint's rate limiting can be bypassed using %s technique.", t.getTechniqueName(technique)),
					Severity:    "High",
					Evidence:    fmt.Sprintf("Successfully bypassed rate limiting using %s technique", t.getTechniqueName(technique)),
					Remediation: t.getRemediationForTechnique(technique),
					CVSS:        7.5,
					CWE:         "CWE-770",
					References: []string{
						"https://owasp.org/API-Security/editions/2019/en/0xa4-lack-of-resources-and-rate-limiting/",
						"https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html",
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

// hasRateLimiting checks if an endpoint has rate limiting
func (t *RateLimitBypassTester) hasRateLimiting(endpoint string, r ffuf.RunnerProvider) bool {
	// Send a burst of requests to trigger rate limiting
	responses := t.sendRequestBurst(endpoint, r, t.RequestsPerTest, t.ConcurrentRequests, nil)

	// Check if any response has a 429 status code (Too Many Requests)
	for _, resp := range responses {
		if resp.StatusCode == 429 {
			return true
		}
	}

	// Check for rate limiting headers
	for _, resp := range responses {
		if t.hasRateLimitHeaders(resp) {
			return true
		}
	}

	return false
}

// hasRateLimitHeaders checks if a response has rate limit headers
func (t *RateLimitBypassTester) hasRateLimitHeaders(resp ffuf.Response) bool {
	rateLimitHeaders := []string{
		"X-RateLimit-Limit",
		"X-RateLimit-Remaining",
		"X-RateLimit-Reset",
		"Retry-After",
		"RateLimit-Limit",
		"RateLimit-Remaining",
		"RateLimit-Reset",
	}

	for _, header := range rateLimitHeaders {
		if values, ok := resp.Headers[header]; ok && len(values) > 0 {
			return true
		}
	}

	return false
}

// testBypassTechnique tests a specific rate limiting bypass technique
func (t *RateLimitBypassTester) testBypassTechnique(endpoint, technique string, r ffuf.RunnerProvider, result *TestResult) bool {
	switch technique {
	case "ip-rotation":
		return t.testIPRotation(endpoint, r)
	case "header-manipulation":
		return t.testHeaderManipulation(endpoint, r)
	case "parameter-pollution":
		return t.testParameterPollution(endpoint, r)
	case "http-method-switching":
		return t.testHTTPMethodSwitching(endpoint, r)
	case "distributed-attack":
		return t.testDistributedAttack(endpoint, r)
	case "cache-manipulation":
		return t.testCacheManipulation(endpoint, r)
	default:
		return false
	}
}

// testIPRotation tests IP rotation bypass technique
func (t *RateLimitBypassTester) testIPRotation(endpoint string, r ffuf.RunnerProvider) bool {
	// Simulate IP rotation by changing X-Forwarded-For header
	successfulRequests := 0
	totalRequests := t.RequestsPerTest

	for i := 0; i < totalRequests; i++ {
		// Create a request with a different X-Forwarded-For header
		headers := map[string]string{
			"X-Forwarded-For": fmt.Sprintf("192.168.1.%d", i%255+1),
			"X-Real-IP":       fmt.Sprintf("192.168.1.%d", i%255+1),
		}

		req := &ffuf.Request{
			Method:  "GET",
			Url:     endpoint,
			Headers: headers,
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}

		// Wait between requests
		time.Sleep(t.TimeBetweenRequests)
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// testHeaderManipulation tests header manipulation bypass technique
func (t *RateLimitBypassTester) testHeaderManipulation(endpoint string, r ffuf.RunnerProvider) bool {
	// Test various header manipulations
	headerSets := []map[string]string{
		{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		{"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15"},
		{"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		{"X-Forwarded-For": "127.0.0.1"},
		{"X-Forwarded-For": "192.168.1.1"},
		{"X-Forwarded-For": "10.0.0.1"},
		{"X-Real-IP": "127.0.0.1"},
		{"X-Real-IP": "192.168.1.1"},
		{"X-Real-IP": "10.0.0.1"},
		{"X-Originating-IP": "127.0.0.1"},
		{"X-Client-IP": "127.0.0.1"},
		{"X-Remote-IP": "127.0.0.1"},
		{"X-Remote-Addr": "127.0.0.1"},
		{"X-Host": "localhost"},
		{"Host": "localhost"},
	}

	successfulRequests := 0
	totalRequests := len(headerSets)

	for _, headers := range headerSets {
		req := &ffuf.Request{
			Method:  "GET",
			Url:     endpoint,
			Headers: headers,
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}

		// Wait between requests
		time.Sleep(t.TimeBetweenRequests)
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// testParameterPollution tests parameter pollution bypass technique
func (t *RateLimitBypassTester) testParameterPollution(endpoint string, r ffuf.RunnerProvider) bool {
	// Add different parameters to each request
	successfulRequests := 0
	totalRequests := t.RequestsPerTest

	for i := 0; i < totalRequests; i++ {
		// Add a unique parameter to the URL
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		modifiedEndpoint := fmt.Sprintf("%s%sdummy%d=%d", endpoint, separator, i, i)

		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}

		// Wait between requests
		time.Sleep(t.TimeBetweenRequests)
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// testHTTPMethodSwitching tests HTTP method switching bypass technique
func (t *RateLimitBypassTester) testHTTPMethodSwitching(endpoint string, r ffuf.RunnerProvider) bool {
	// Test different HTTP methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	successfulRequests := 0
	totalRequests := len(methods) * (t.RequestsPerTest / len(methods))

	for i := 0; i < totalRequests; i++ {
		method := methods[i%len(methods)]

		req := &ffuf.Request{
			Method: method,
			Url:    endpoint,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		// Add a body for methods that support it
		if method == "POST" || method == "PUT" || method == "PATCH" {
			req.Data = []byte(`{"test":"data"}`)
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}

		// Wait between requests
		time.Sleep(t.TimeBetweenRequests)
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// testDistributedAttack tests distributed attack bypass technique
func (t *RateLimitBypassTester) testDistributedAttack(endpoint string, r ffuf.RunnerProvider) bool {
	// Simulate a distributed attack by using different headers for each request
	successfulRequests := 0
	totalRequests := t.RequestsPerTest

	// Send requests concurrently
	responses := t.sendRequestBurst(endpoint, r, totalRequests, t.ConcurrentRequests, func(i int) map[string]string {
		return map[string]string{
			"User-Agent":      fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.%d Safari/537.36", i),
			"X-Forwarded-For": fmt.Sprintf("192.168.%d.%d", (i/255)%255+1, i%255+1),
			"X-Real-IP":       fmt.Sprintf("10.%d.%d.%d", (i/65536)%255+1, (i/255)%255+1, i%255+1),
			"X-Request-ID":    fmt.Sprintf("req-%d", i),
		}
	})

	// Count successful responses
	for _, resp := range responses {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// testCacheManipulation tests cache manipulation bypass technique
func (t *RateLimitBypassTester) testCacheManipulation(endpoint string, r ffuf.RunnerProvider) bool {
	// Test cache manipulation by adding cache-busting parameters
	successfulRequests := 0
	totalRequests := t.RequestsPerTest

	for i := 0; i < totalRequests; i++ {
		// Add a cache-busting parameter to the URL
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		modifiedEndpoint := fmt.Sprintf("%s%s_=%d", endpoint, separator, time.Now().UnixNano())

		req := &ffuf.Request{
			Method: "GET",
			Url:    modifiedEndpoint,
			Headers: map[string]string{
				"Cache-Control": "no-cache",
				"Pragma":        "no-cache",
			},
		}

		// Execute the request
		resp, err := r.Execute(req)
		if err != nil {
			continue
		}

		// Check if the request was successful
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			successfulRequests++
		}

		// Wait between requests
		time.Sleep(t.TimeBetweenRequests)
	}

	// If most requests were successful, the bypass technique works
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// sendRequestBurst sends a burst of requests to an endpoint
func (t *RateLimitBypassTester) sendRequestBurst(endpoint string, r ffuf.RunnerProvider, count, concurrent int, headerFunc func(int) map[string]string) []ffuf.Response {
	var responses []ffuf.Response
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create a channel to limit concurrency
	semaphore := make(chan struct{}, concurrent)

	for i := 0; i < count; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire a slot

		go func(i int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the slot

			// Create headers
			headers := map[string]string{
				"X-Test-ID": fmt.Sprintf("rate-limit-bypass-test-%d", i),
			}

			// Add custom headers if provided
			if headerFunc != nil {
				customHeaders := headerFunc(i)
				for k, v := range customHeaders {
					headers[k] = v
				}
			}

			// Create a request
			req := &ffuf.Request{
				Method:  "GET",
				Url:     endpoint,
				Headers: headers,
			}

			// Execute the request
			resp, err := r.Execute(req)
			if err != nil {
				return
			}

			// Add the response to the list
			mutex.Lock()
			responses = append(responses, resp)
			mutex.Unlock()

			// Wait between requests
			time.Sleep(t.TimeBetweenRequests)
		}(i)
	}

	wg.Wait()
	return responses
}

// getTechniqueName returns a human-readable name for a bypass technique
func (t *RateLimitBypassTester) getTechniqueName(technique string) string {
	switch technique {
	case "ip-rotation":
		return "IP Rotation"
	case "header-manipulation":
		return "Header Manipulation"
	case "parameter-pollution":
		return "Parameter Pollution"
	case "http-method-switching":
		return "HTTP Method Switching"
	case "distributed-attack":
		return "Distributed Attack"
	case "cache-manipulation":
		return "Cache Manipulation"
	default:
		return technique
	}
}

// getRemediationForTechnique returns remediation advice for a bypass technique
func (t *RateLimitBypassTester) getRemediationForTechnique(technique string) string {
	switch technique {
	case "ip-rotation":
		return "Implement rate limiting based on a combination of IP address and other identifiers. Consider using API keys or tokens for authentication and rate limiting. Use a reputable IP reputation service to detect and block suspicious IP patterns."
	case "header-manipulation":
		return "Don't rely solely on headers for rate limiting decisions. Implement rate limiting based on authenticated user identity when possible. Validate and normalize headers before using them for rate limiting."
	case "parameter-pollution":
		return "Implement rate limiting at the API endpoint level, not just at the URL level. Normalize request parameters before applying rate limiting. Consider implementing a request signature mechanism."
	case "http-method-switching":
		return "Apply rate limiting consistently across all HTTP methods for the same resource. Implement proper method validation and restrict unused HTTP methods."
	case "distributed-attack":
		return "Implement global rate limiting across your infrastructure. Consider using a centralized rate limiting service. Implement progressive rate limiting that becomes more restrictive as traffic increases."
	case "cache-manipulation":
		return "Implement rate limiting at the application level, not just at the caching layer. Normalize URLs and parameters before applying rate limiting. Consider implementing token bucket or sliding window rate limiting algorithms."
	default:
		return "Implement proper rate limiting. Consider using token bucket, fixed window, or sliding window algorithms. Limit the number of requests a client can make in a given time period."
	}
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewRateLimitBypassTester())
}
