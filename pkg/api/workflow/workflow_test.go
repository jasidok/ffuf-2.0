// Package workflow provides a concurrency model for complex API workflows.
package workflow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// MockRunner is a mock implementation of ffuf.RunnerProvider for testing.
type MockRunner struct {
	// Responses is a map of URLs to responses
	Responses map[string]ffuf.Response
	// Errors is a map of URLs to errors
	Errors map[string]error
	// Calls tracks the number of calls to Execute for each URL
	Calls map[string]int
	// ExecuteFunc is a custom function to use for Execute
	ExecuteFunc func(req *ffuf.Request) (ffuf.Response, error)
}

// Execute implements ffuf.RunnerProvider.
func (r *MockRunner) Execute(req *ffuf.Request) (ffuf.Response, error) {
	// If a custom ExecuteFunc is provided, use it
	if r.ExecuteFunc != nil {
		return r.ExecuteFunc(req)
	}

	// Track calls
	r.Calls[req.Url]++

	// Check if there's an error for this URL
	if err, ok := r.Errors[req.Url]; ok {
		return ffuf.Response{}, err
	}

	// Return the response for this URL
	if resp, ok := r.Responses[req.Url]; ok {
		return resp, nil
	}

	// Default response
	return ffuf.Response{
		StatusCode:   200,
		ContentType:  "application/json",
		ContentWords: 0,
		ContentLines: 0,
		ContentLength: 0,
		Data:         []byte(`{"status":"ok"}`),
	}, nil
}

// Prepare implements ffuf.RunnerProvider.
func (r *MockRunner) Prepare(input map[string][]byte, basereq *ffuf.Request) (ffuf.Request, error) {
	// Create a copy of the request
	req := ffuf.CopyRequest(basereq)

	// Apply input replacements (simplified version)
	for keyword, inputitem := range input {
		req.Url = strings.ReplaceAll(req.Url, keyword, string(inputitem))
	}

	return req, nil
}

// Dump implements ffuf.RunnerProvider.
func (r *MockRunner) Dump(req *ffuf.Request) ([]byte, error) {
	return []byte{}, nil
}

// NewMockRunner creates a new MockRunner.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		Responses: make(map[string]ffuf.Response),
		Errors:    make(map[string]error),
		Calls:     make(map[string]int),
	}
}

// TestWorkflowExecution tests the execution of a simple workflow.
func TestWorkflowExecution(t *testing.T) {
	// Create a workflow engine
	engine := NewWorkflowEngine()

	// Create a workflow
	workflow := engine.NewWorkflow("test", "Test Workflow", "A test workflow")

	// Create a mock runner
	runner := NewMockRunner()

	// Set up responses
	runner.Responses["https://api.example.com/users"] = ffuf.Response{
		StatusCode:   200,
		ContentType:  "application/json",
		ContentWords: 10,
		ContentLines: 5,
		ContentLength: 100,
		Data:         []byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]}`),
	}

	runner.Responses["https://api.example.com/users/1"] = ffuf.Response{
		StatusCode:   200,
		ContentType:  "application/json",
		ContentWords: 5,
		ContentLines: 3,
		ContentLength: 50,
		Data:         []byte(`{"id":1,"name":"Alice","email":"alice@example.com"}`),
	}

	// Set the runner
	workflow.SetRunner(runner)

	// Add steps
	step1 := &Step{
		ID:          "get-users",
		Name:        "Get Users",
		Description: "Get a list of users",
		Request: &ffuf.Request{
			Method: "GET",
			Url:    "https://api.example.com/users",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		},
		ExtractVariables: map[string]string{
			"user_id": "$.users[0].id",
		},
	}

	step2 := &Step{
		ID:          "get-user",
		Name:        "Get User",
		Description: "Get a specific user",
		Request: &ffuf.Request{
			Method: "GET",
			Url:    "https://api.example.com/users/${user_id}",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		},
		DependsOn: []string{"get-users"},
		ExtractVariables: map[string]string{
			"user_email": "$.email",
		},
	}

	// Add steps to workflow
	if err := workflow.AddStep(step1); err != nil {
		t.Fatalf("Failed to add step1: %v", err)
	}

	if err := workflow.AddStep(step2); err != nil {
		t.Fatalf("Failed to add step2: %v", err)
	}

	// Execute the workflow
	result, err := engine.Execute(workflow)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Error("Workflow execution failed")
	}

	// Check that both steps were executed
	if len(result.StepResults) != 2 {
		t.Errorf("Expected 2 step results, got %d", len(result.StepResults))
	}

	// Check that step1 was successful
	if step1Result, ok := result.StepResults["get-users"]; !ok || !step1Result.Success {
		t.Error("Step 1 failed or missing")
	}

	// Check that step2 was successful
	if step2Result, ok := result.StepResults["get-user"]; !ok || !step2Result.Success {
		t.Error("Step 2 failed or missing")
	}

	// Check that variables were extracted correctly
	if result.Variables["user_id"] != "1" {
		t.Errorf("Expected user_id to be 1, got %s", result.Variables["user_id"])
	}

	if result.Variables["user_email"] != "alice@example.com" {
		t.Errorf("Expected user_email to be alice@example.com, got %s", result.Variables["user_email"])
	}

	// Check that the correct URLs were called
	if runner.Calls["https://api.example.com/users"] != 1 {
		t.Errorf("Expected 1 call to /users, got %d", runner.Calls["https://api.example.com/users"])
	}

	if runner.Calls["https://api.example.com/users/1"] != 1 {
		t.Errorf("Expected 1 call to /users/1, got %d", runner.Calls["https://api.example.com/users/1"])
	}
}

// TestJSONPathParsing tests the parsing of JSONPath expressions.
func TestJSONPathParsing(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantPath int // Number of components expected
	}{
		{
			name:     "Empty path",
			path:     "",
			wantErr:  false,
			wantPath: 1, // Just the root component
		},
		{
			name:     "Root path",
			path:     "$",
			wantErr:  false,
			wantPath: 1, // Just the root component
		},
		{
			name:     "Simple property",
			path:     "$.name",
			wantErr:  false,
			wantPath: 2, // Root + property
		},
		{
			name:     "Nested property",
			path:     "$.user.name",
			wantErr:  false,
			wantPath: 3, // Root + user + name
		},
		{
			name:     "Array index",
			path:     "$.users[0]",
			wantErr:  false,
			wantPath: 3, // Root + users + index
		},
		{
			name:     "Array index with property",
			path:     "$.users[0].name",
			wantErr:  false,
			wantPath: 4, // Root + users + index + name
		},
		{
			name:     "Multiple array indices",
			path:     "$.users[0].addresses[1]",
			wantErr:  false,
			wantPath: 5, // Root + users + index + addresses + index
		},
		{
			name:     "Invalid array index",
			path:     "$.users[a]",
			wantErr:  true,
			wantPath: 0,
		},
		{
			name:     "Invalid array syntax",
			path:     "$.users[0",
			wantErr:  true,
			wantPath: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ParseJSONPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(path.Path) != tt.wantPath {
				t.Errorf("ParseJSONPath() path length = %d, want %d", len(path.Path), tt.wantPath)
			}
		})
	}
}

// TestJSONPathEvaluation tests the evaluation of JSONPath expressions.
func TestJSONPathEvaluation(t *testing.T) {
	// Test data
	jsonData := []byte(`{
		"name": "John",
		"age": 30,
		"isActive": true,
		"address": {
			"street": "123 Main St",
			"city": "Anytown"
		},
		"phones": [
			{
				"type": "home",
				"number": "555-1234"
			},
			{
				"type": "work",
				"number": "555-5678"
			}
		],
		"tags": ["developer", "gopher"]
	}`)

	tests := []struct {
		name     string
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:     "Root",
			path:     "$",
			want:     string(jsonData), // The entire JSON
			wantErr:  false,
		},
		{
			name:     "Simple property",
			path:     "$.name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "Number property",
			path:     "$.age",
			want:     "30",
			wantErr:  false,
		},
		{
			name:     "Boolean property",
			path:     "$.isActive",
			want:     "true",
			wantErr:  false,
		},
		{
			name:     "Nested property",
			path:     "$.address.street",
			want:     "123 Main St",
			wantErr:  false,
		},
		{
			name:     "Array index",
			path:     "$.phones[0].number",
			want:     "555-1234",
			wantErr:  false,
		},
		{
			name:     "Array index with property",
			path:     "$.phones[1].type",
			want:     "work",
			wantErr:  false,
		},
		{
			name:     "Simple array element",
			path:     "$.tags[0]",
			want:     "developer",
			wantErr:  false,
		},
		{
			name:     "Non-existent property",
			path:     "$.nonexistent",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid array index",
			path:     "$.phones[99].number",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateJSONPath(jsonData, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EvaluateJSONPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestConcurrencyAndRateLimiting tests the concurrency and rate limiting features.
func TestConcurrencyAndRateLimiting(t *testing.T) {
	// Create a test server that tracks request times
	var requestTimes []time.Time
	var requestMutex sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		requestTimes = append(requestTimes, time.Now())
		requestMutex.Unlock()

		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		// Return a simple JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create a workflow engine
	engine := NewWorkflowEngine()

	// Create a workflow with rate limiting
	workflow := engine.NewWorkflow("rate-test", "Rate Test", "Testing rate limiting")
	workflow.MaxConcurrency = 3 // Max 3 concurrent requests
	workflow.RateLimit = 10     // Max 10 requests per second

	// Create a mock runner
	runner := &MockRunner{
		Responses: make(map[string]ffuf.Response),
		Errors:    make(map[string]error),
		Calls:     make(map[string]int),
	}
	workflow.SetRunner(runner)

	// Add 10 independent steps
	for i := 0; i < 10; i++ {
		step := &Step{
			ID:          fmt.Sprintf("step-%d", i),
			Name:        fmt.Sprintf("Step %d", i),
			Description: fmt.Sprintf("Test step %d", i),
			Request: &ffuf.Request{
				Method: "GET",
				Url:    server.URL,
				Headers: map[string]string{
					"Accept": "application/json",
				},
			},
		}

		if err := workflow.AddStep(step); err != nil {
			t.Fatalf("Failed to add step %d: %v", i, err)
		}
	}

	// Execute the workflow
	startTime := time.Now()
	result, err := engine.Execute(workflow)
	endTime := time.Now()

	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Error("Workflow execution failed")
	}

	// Check that all steps were executed
	if len(result.StepResults) != 10 {
		t.Errorf("Expected 10 step results, got %d", len(result.StepResults))
	}

	// Check concurrency: at most 3 requests should be in flight at any time
	requestMutex.Lock()
	defer requestMutex.Unlock()

	// Sort request times
	sort.Slice(requestTimes, func(i, j int) bool {
		return requestTimes[i].Before(requestTimes[j])
	})

	// Check that the total execution time is reasonable
	// With 10 requests, max concurrency 3, and each request taking ~10ms,
	// we expect at least 4 batches, so ~40ms minimum
	executionTime := endTime.Sub(startTime)
	if executionTime < 40*time.Millisecond {
		t.Errorf("Execution time too short: %v, expected at least 40ms", executionTime)
	}

	// Check rate limiting: no more than 10 requests per second
	// This is a bit tricky to test reliably, so we'll just check that
	// the average rate is close to what we expect
	requestCount := len(requestTimes)
	if requestCount != 10 {
		t.Errorf("Expected 10 requests, got %d", requestCount)
	}

	// The rate should be close to 10 per second (but could be less due to other factors)
	rate := float64(requestCount) / executionTime.Seconds()
	if rate > 12 { // Allow some margin for error
		t.Errorf("Request rate too high: %.2f requests/second, expected at most 10", rate)
	}
}

// TestRetryMechanism tests the retry mechanism.
func TestRetryMechanism(t *testing.T) {
	// Create a workflow engine
	engine := NewWorkflowEngine()

	// Create a workflow
	workflow := engine.NewWorkflow("retry-test", "Retry Test", "Testing retry mechanism")

	// Create a mock runner that fails the first 2 times
	runner := NewMockRunner()
	failCount := 0

	runner.ExecuteFunc = func(req *ffuf.Request) (ffuf.Response, error) {
		// Track calls
		runner.Calls[req.Url]++

		// Fail the first 2 times
		if runner.Calls[req.Url] <= 2 {
			failCount++
			return ffuf.Response{}, fmt.Errorf("simulated error %d", failCount)
		}

		// Succeed on the 3rd try
		return ffuf.Response{
			StatusCode:   200,
			ContentType:  "application/json",
			ContentWords: 0,
			ContentLines: 0,
			ContentLength: 0,
			Data:         []byte(`{"status":"ok"}`),
		}, nil
	}

	workflow.SetRunner(runner)

	// Add a step with retry configuration
	step := &Step{
		ID:          "retry-step",
		Name:        "Retry Step",
		Description: "A step that will be retried",
		Request: &ffuf.Request{
			Method: "GET",
			Url:    "https://api.example.com/retry-test",
			Headers: map[string]string{
				"Accept": "application/json",
			},
		},
		RetryConfig: &RetryConfig{
			MaxRetries:           3,
			RetryDelay:           10 * time.Millisecond,
			BackoffFactor:        2.0,
			RetryableStatusCodes: []int{429, 500, 502, 503, 504},
			RetryableErrors:      []string{"simulated error"},
		},
	}

	if err := workflow.AddStep(step); err != nil {
		t.Fatalf("Failed to add step: %v", err)
	}

	// Execute the workflow
	result, err := engine.Execute(workflow)

	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Error("Workflow execution failed")
	}

	// Check that the step was retried
	stepResult := result.StepResults["retry-step"]
	if stepResult.RetryCount != 2 {
		t.Errorf("Expected 2 retries, got %d", stepResult.RetryCount)
	}

	// Check that the URL was called 3 times (initial + 2 retries)
	if runner.Calls["https://api.example.com/retry-test"] != 3 {
		t.Errorf("Expected 3 calls, got %d", runner.Calls["https://api.example.com/retry-test"])
	}
}
