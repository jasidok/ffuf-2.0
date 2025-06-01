// Package security provides testing modules for API security vulnerabilities.
package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
	"github.com/ffuf/ffuf/v2/pkg/runner"
)

// LackOfResourcesTester implements testing for Lack of Resources & Rate Limiting (API4:2019)
type LackOfResourcesTester struct {
	// Configuration options
	RequestsPerBurst     int
	BurstCount           int
	TimeBetweenBursts    time.Duration
	ConcurrentRequests   int
	LargePayloadSize     int
	LargeParameterCount  int
	LargeParameterLength int
}

// NewLackOfResourcesTester creates a new tester for Lack of Resources & Rate Limiting
func NewLackOfResourcesTester() *LackOfResourcesTester {
	return &LackOfResourcesTester{
		RequestsPerBurst:     20,
		BurstCount:           3,
		TimeBetweenBursts:    time.Second * 2,
		ConcurrentRequests:   10,
		LargePayloadSize:     1024 * 100, // 100KB
		LargeParameterCount:  100,
		LargeParameterLength: 1000,
	}
}

// GetType returns the type of vulnerability this tester checks for
func (t *LackOfResourcesTester) GetType() VulnerabilityType {
	return VulnLackOfResources
}

// GetName returns the name of the security test
func (t *LackOfResourcesTester) GetName() string {
	return "Lack of Resources & Rate Limiting"
}

// GetDescription returns a description of the security test
func (t *LackOfResourcesTester) GetDescription() string {
	return "Tests for API endpoints that don't properly limit the amount of resources a client can request, potentially leading to denial of service."
}

// Test runs the security test against the target
func (t *LackOfResourcesTester) Test(ctx context.Context, config *ffuf.Config) (*TestResult, error) {
	result := &TestResult{
		TestName:  t.GetName(),
		StartTime: time.Now(),
	}

	// Create a runner for making HTTP requests
	r := runner.NewSimpleRunner(config, false)

	// Extract potential endpoints from the config
	endpoints := extractEndpointsFromConfig(config)

	// Test each endpoint for rate limiting vulnerabilities
	for _, endpoint := range endpoints {
		// Test for lack of rate limiting
		if t.testRateLimiting(endpoint, r, result) {
			// If rate limiting is missing, add a vulnerability
			vuln := VulnerabilityInfo{
				Type:        VulnLackOfResources,
				Name:        "Missing Rate Limiting",
				Description: "The API endpoint does not implement proper rate limiting.",
				Severity:    "High",
				Evidence:    fmt.Sprintf("Successfully sent %d requests in %d bursts without being rate limited", t.RequestsPerBurst*t.BurstCount, t.BurstCount),
				Remediation: "Implement proper rate limiting. Consider using token bucket, fixed window, or sliding window algorithms. Limit the number of requests a client can make in a given time period.",
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

		// Test for lack of resource limiting (large payload)
		if t.testLargePayload(endpoint, r, result) {
			// If resource limiting is missing, add a vulnerability
			vuln := VulnerabilityInfo{
				Type:        VulnLackOfResources,
				Name:        "Missing Resource Limiting (Large Payload)",
				Description: "The API endpoint accepts unusually large request payloads.",
				Severity:    "Medium",
				Evidence:    fmt.Sprintf("Successfully sent a request with a %d KB payload", t.LargePayloadSize/1024),
				Remediation: "Implement proper request size limiting. Set maximum allowed request body size. Consider implementing payload validation and sanitization.",
				CVSS:        5.0,
				CWE:         "CWE-400",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa4-lack-of-resources-and-rate-limiting/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html",
				},
				DetectedAt: time.Now(),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}

		// Test for lack of resource limiting (many parameters)
		if t.testManyParameters(endpoint, r, result) {
			// If resource limiting is missing, add a vulnerability
			vuln := VulnerabilityInfo{
				Type:        VulnLackOfResources,
				Name:        "Missing Resource Limiting (Many Parameters)",
				Description: "The API endpoint accepts requests with an unusually large number of parameters.",
				Severity:    "Medium",
				Evidence:    fmt.Sprintf("Successfully sent a request with %d parameters", t.LargeParameterCount),
				Remediation: "Implement proper parameter limiting. Set maximum allowed number of parameters. Consider implementing parameter validation and sanitization.",
				CVSS:        5.0,
				CWE:         "CWE-400",
				References: []string{
					"https://owasp.org/API-Security/editions/2019/en/0xa4-lack-of-resources-and-rate-limiting/",
					"https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html",
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

// testRateLimiting tests if an endpoint implements rate limiting
func (t *LackOfResourcesTester) testRateLimiting(endpoint string, r ffuf.RunnerProvider, result *TestResult) bool {
	// Track successful requests
	successfulRequests := 0
	totalRequests := t.RequestsPerBurst * t.BurstCount

	// Send requests in bursts
	for burst := 0; burst < t.BurstCount; burst++ {
		// Send a burst of requests
		responses := t.sendRequestBurst(endpoint, r, t.RequestsPerBurst, t.ConcurrentRequests)

		// Count successful responses
		for _, resp := range responses {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				successfulRequests++
			} else if resp.StatusCode == 429 {
				// If we get a 429 Too Many Requests, rate limiting is implemented
				return false
			}
		}

		// Wait between bursts
		if burst < t.BurstCount-1 {
			time.Sleep(t.TimeBetweenBursts)
		}
	}

	// If most requests were successful, rate limiting is likely missing
	return float64(successfulRequests) / float64(totalRequests) > 0.8
}

// sendRequestBurst sends a burst of requests to an endpoint
func (t *LackOfResourcesTester) sendRequestBurst(endpoint string, r ffuf.RunnerProvider, count, concurrent int) []ffuf.Response {
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

			// Create a request
			req := &ffuf.Request{
				Method: "GET",
				Url:    endpoint,
				Headers: map[string]string{
					"X-Test-ID": fmt.Sprintf("rate-limit-test-%d", i),
				},
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
		}(i)
	}

	wg.Wait()
	return responses
}

// testLargePayload tests if an endpoint accepts unusually large payloads
func (t *LackOfResourcesTester) testLargePayload(endpoint string, r ffuf.RunnerProvider, result *TestResult) bool {
	// Create a large payload
	payload := make([]byte, t.LargePayloadSize)
	for i := range payload {
		payload[i] = 'A'
	}

	// Create a request with the large payload
	req := &ffuf.Request{
		Method: "POST",
		Url:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Data: payload,
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return false
	}

	// If the request was successful, the endpoint accepts large payloads
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// testManyParameters tests if an endpoint accepts requests with many parameters
func (t *LackOfResourcesTester) testManyParameters(endpoint string, r ffuf.RunnerProvider, result *TestResult) bool {
	// Create a JSON object with many parameters
	jsonStart := "{"
	jsonEnd := "}"
	jsonMiddle := ""

	for i := 0; i < t.LargeParameterCount; i++ {
		// Create a parameter with a long value
		paramValue := make([]byte, t.LargeParameterLength)
		for j := range paramValue {
			paramValue[j] = 'A'
		}

		if i > 0 {
			jsonMiddle += ","
		}
		jsonMiddle += fmt.Sprintf(`"param%d":"%s"`, i, string(paramValue))
	}

	// Create a request with many parameters
	req := &ffuf.Request{
		Method: "POST",
		Url:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Data: []byte(jsonStart + jsonMiddle + jsonEnd),
	}

	// Execute the request
	resp, err := r.Execute(req)
	if err != nil {
		return false
	}

	// If the request was successful, the endpoint accepts many parameters
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func init() {
	// Register the tester with the default registry
	RegisterSecurityTester(NewLackOfResourcesTester())
}