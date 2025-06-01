// Package workflow provides a concurrency model for complex API workflows.
//
// This package implements a workflow engine that can handle dependencies between
// API requests, parallel execution of independent requests, rate limiting,
// backoff strategies, and error handling with retries.
package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// Step represents a single step in a workflow.
type Step struct {
	// ID is a unique identifier for the step
	ID string

	// Name is a human-readable name for the step
	Name string

	// Description provides details about what the step does
	Description string

	// Request is the request to be executed in this step
	Request *ffuf.Request

	// DependsOn is a list of step IDs that must complete before this step
	DependsOn []string

	// ExtractVariables defines variables to extract from the response
	// The key is the variable name, the value is a JSONPath expression
	ExtractVariables map[string]string

	// RetryConfig defines the retry behavior for this step
	RetryConfig *RetryConfig

	// OnSuccess is a function to call when the step succeeds
	OnSuccess func(resp ffuf.Response) error

	// OnError is a function to call when the step fails
	OnError func(err error) error
}

// RetryConfig defines the retry behavior for a step.
type RetryConfig struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration

	// BackoffFactor is the factor by which to increase the delay after each retry
	BackoffFactor float64

	// RetryableStatusCodes is a list of status codes that should trigger a retry
	RetryableStatusCodes []int

	// RetryableErrors is a list of error types that should trigger a retry
	RetryableErrors []string
}

// StepResult represents the result of executing a step.
type StepResult struct {
	// StepID is the ID of the step that was executed
	StepID string

	// Success indicates whether the step was successful
	Success bool

	// Response is the response from the request
	Response ffuf.Response

	// Error is the error that occurred, if any
	Error error

	// RetryCount is the number of retries that were performed
	RetryCount int

	// ExtractedVariables contains variables extracted from the response
	ExtractedVariables map[string]string

	// StartTime is when the step started execution
	StartTime time.Time

	// EndTime is when the step completed execution
	EndTime time.Time
}

// Workflow represents a sequence of API requests with dependencies.
type Workflow struct {
	// ID is a unique identifier for the workflow
	ID string

	// Name is a human-readable name for the workflow
	Name string

	// Description provides details about what the workflow does
	Description string

	// Steps is a map of step IDs to steps
	Steps map[string]*Step

	// Variables contains variables that can be used in requests
	Variables map[string]string

	// MaxConcurrency is the maximum number of steps to execute concurrently
	MaxConcurrency int

	// RateLimit is the maximum number of requests per second
	RateLimit int

	// Context is the context for the workflow execution
	Context context.Context

	// Runner is the runner to use for executing requests
	Runner ffuf.RunnerProvider
}

// WorkflowResult represents the result of executing a workflow.
type WorkflowResult struct {
	// WorkflowID is the ID of the workflow that was executed
	WorkflowID string

	// Success indicates whether the workflow was successful
	Success bool

	// StepResults is a map of step IDs to step results
	StepResults map[string]*StepResult

	// Variables contains the final state of variables
	Variables map[string]string

	// StartTime is when the workflow started execution
	StartTime time.Time

	// EndTime is when the workflow completed execution
	EndTime time.Time
}

// WorkflowEngine is responsible for executing workflows.
type WorkflowEngine struct {
	// DefaultMaxConcurrency is the default maximum number of steps to execute concurrently
	DefaultMaxConcurrency int

	// DefaultRateLimit is the default maximum number of requests per second
	DefaultRateLimit int

	// DefaultRetryConfig is the default retry configuration
	DefaultRetryConfig *RetryConfig
}

// NewWorkflowEngine creates a new workflow engine with default settings.
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		DefaultMaxConcurrency: 10,
		DefaultRateLimit:      10,
		DefaultRetryConfig: &RetryConfig{
			MaxRetries:           3,
			RetryDelay:           time.Second,
			BackoffFactor:        2.0,
			RetryableStatusCodes: []int{429, 500, 502, 503, 504},
			RetryableErrors:      []string{"timeout", "connection refused"},
		},
	}
}

// NewWorkflow creates a new workflow with the given ID, name, and description.
func (e *WorkflowEngine) NewWorkflow(id, name, description string) *Workflow {
	return &Workflow{
		ID:             id,
		Name:           name,
		Description:    description,
		Steps:          make(map[string]*Step),
		Variables:      make(map[string]string),
		MaxConcurrency: e.DefaultMaxConcurrency,
		RateLimit:      e.DefaultRateLimit,
		Context:        context.Background(),
	}
}

// AddStep adds a step to the workflow.
func (w *Workflow) AddStep(step *Step) error {
	if _, exists := w.Steps[step.ID]; exists {
		return fmt.Errorf("step with ID %s already exists", step.ID)
	}

	// Validate dependencies
	for _, depID := range step.DependsOn {
		if _, exists := w.Steps[depID]; !exists {
			return fmt.Errorf("dependency %s does not exist", depID)
		}
	}

	w.Steps[step.ID] = step
	return nil
}

// SetVariable sets a variable in the workflow.
func (w *Workflow) SetVariable(name, value string) {
	w.Variables[name] = value
}

// SetRunner sets the runner to use for executing requests.
func (w *Workflow) SetRunner(runner ffuf.RunnerProvider) {
	w.Runner = runner
}

// SetContext sets the context for the workflow execution.
func (w *Workflow) SetContext(ctx context.Context) {
	w.Context = ctx
}

// Execute executes the workflow and returns the result.
func (e *WorkflowEngine) Execute(w *Workflow) (*WorkflowResult, error) {
	if w.Runner == nil {
		return nil, fmt.Errorf("no runner provided")
	}

	result := &WorkflowResult{
		WorkflowID:  w.ID,
		StepResults: make(map[string]*StepResult),
		Variables:   make(map[string]string),
		StartTime:   time.Now(),
	}

	// Copy initial variables
	for k, v := range w.Variables {
		result.Variables[k] = v
	}

	// Build dependency graph
	dependencyGraph := buildDependencyGraph(w)

	// Execute steps in dependency order
	err := e.executeSteps(w, dependencyGraph, result)
	result.EndTime = time.Now()
	result.Success = err == nil

	return result, err
}

// buildDependencyGraph builds a graph of step dependencies.
// The returned map has step IDs as keys and a list of dependent step IDs as values.
func buildDependencyGraph(w *Workflow) map[string][]string {
	graph := make(map[string][]string)

	// Initialize graph with empty dependencies
	for id := range w.Steps {
		graph[id] = []string{}
	}

	// Add dependencies
	for id, step := range w.Steps {
		for _, depID := range step.DependsOn {
			graph[depID] = append(graph[depID], id)
		}
	}

	return graph
}

// executeSteps executes the steps in the workflow in dependency order.
func (e *WorkflowEngine) executeSteps(w *Workflow, dependencyGraph map[string][]string, result *WorkflowResult) error {
	// Create a map to track completed steps
	completed := make(map[string]bool)

	// Create a map to track steps that are ready to execute
	ready := make(map[string]bool)

	// Find steps with no dependencies
	for id, step := range w.Steps {
		if len(step.DependsOn) == 0 {
			ready[id] = true
		}
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, w.MaxConcurrency)

	// Create a rate limiter
	var rateLimiter <-chan time.Time
	if w.RateLimit > 0 {
		rateLimiter = time.Tick(time.Second / time.Duration(w.RateLimit))
	}

	// Execute steps until all are completed or an error occurs
	for len(completed) < len(w.Steps) {
		// Check if there are any steps ready to execute
		if len(ready) == 0 {
			// If no steps are ready but not all are completed, there might be a circular dependency
			if len(completed) < len(w.Steps) {
				return fmt.Errorf("circular dependency detected or some steps cannot be executed")
			}
			break
		}

		// Execute ready steps concurrently
		var wg sync.WaitGroup
		stepErrors := make(chan error, len(ready))
		stepResults := make(chan *StepResult, len(ready))

		for id := range ready {
			// Remove from ready list
			delete(ready, id)

			// Acquire a slot in the semaphore
			semaphore <- struct{}{}

			// Wait for rate limiter if enabled
			if rateLimiter != nil {
				<-rateLimiter
			}

			wg.Add(1)
			go func(stepID string) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release the semaphore slot

				// Execute the step
				stepResult, err := e.executeStep(w, stepID, result.Variables)
				if err != nil {
					stepErrors <- err
					return
				}

				// Add result to results map
				stepResults <- stepResult
			}(id)
		}

		// Wait for all steps to complete
		wg.Wait()
		close(stepErrors)
		close(stepResults)

		// Check for errors
		for err := range stepErrors {
			if err != nil {
				return err
			}
		}

		// Process results
		for stepResult := range stepResults {
			// Add result to results map
			result.StepResults[stepResult.StepID] = stepResult

			// Mark step as completed
			completed[stepResult.StepID] = true

			// Extract variables
			for k, v := range stepResult.ExtractedVariables {
				result.Variables[k] = v
			}

			// Find new ready steps
			for _, dependentID := range dependencyGraph[stepResult.StepID] {
				// Check if all dependencies of this step are completed
				allDepsCompleted := true
				for _, depID := range w.Steps[dependentID].DependsOn {
					if !completed[depID] {
						allDepsCompleted = false
						break
					}
				}

				// If all dependencies are completed, mark as ready
				if allDepsCompleted {
					ready[dependentID] = true
				}
			}
		}
	}

	return nil
}

// executeStep executes a single step in the workflow.
func (e *WorkflowEngine) executeStep(w *Workflow, stepID string, variables map[string]string) (*StepResult, error) {
	step := w.Steps[stepID]
	result := &StepResult{
		StepID:             stepID,
		ExtractedVariables: make(map[string]string),
		StartTime:          time.Now(),
	}

	// Apply variables to request
	req, err := applyVariables(step.Request, variables)
	if err != nil {
		result.Success = false
		result.Error = err
		result.EndTime = time.Now()
		return result, err
	}

	// Get retry config
	retryConfig := step.RetryConfig
	if retryConfig == nil {
		retryConfig = e.DefaultRetryConfig
	}

	// Execute with retries
	var resp ffuf.Response
	var execErr error
	retryCount := 0

	for retryCount <= retryConfig.MaxRetries {
		// Execute the request
		resp, execErr = w.Runner.Execute(req)

		// Check if successful
		if execErr == nil && !shouldRetry(resp.StatusCode, "", retryConfig) {
			break
		}

		// Increment retry count
		retryCount++

		// Check if we've reached max retries
		if retryCount > retryConfig.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := retryConfig.RetryDelay
		for i := 0; i < retryCount-1; i++ {
			delay = time.Duration(float64(delay) * retryConfig.BackoffFactor)
		}

		// Wait before retrying
		select {
		case <-w.Context.Done():
			return nil, w.Context.Err()
		case <-time.After(delay):
			// Continue with retry
		}
	}

	result.RetryCount = retryCount
	result.EndTime = time.Now()

	// Handle final result
	if execErr != nil {
		result.Success = false
		result.Error = execErr

		// Call OnError if provided
		if step.OnError != nil {
			if err := step.OnError(execErr); err != nil {
				return result, err
			}
		}

		return result, nil
	}

	result.Success = true
	result.Response = resp

	// Extract variables
	if len(step.ExtractVariables) > 0 {
		extractedVars, err := extractVariables(resp, step.ExtractVariables)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, nil
		}
		result.ExtractedVariables = extractedVars
	}

	// Call OnSuccess if provided
	if step.OnSuccess != nil {
		if err := step.OnSuccess(resp); err != nil {
			result.Success = false
			result.Error = err
			return result, nil
		}
	}

	return result, nil
}

// applyVariables applies variables to a request.
func applyVariables(req *ffuf.Request, variables map[string]string) (*ffuf.Request, error) {
	// Create a copy of the request
	newReq := ffuf.CopyRequest(req)

	// Apply variables to URL
	for name, value := range variables {
		placeholder := fmt.Sprintf("${%s}", name)
		newReq.Url = strings.ReplaceAll(newReq.Url, placeholder, value)
	}

	// Apply variables to headers
	for name, value := range variables {
		placeholder := fmt.Sprintf("${%s}", name)
		for headerName, headerValue := range newReq.Headers {
			newReq.Headers[headerName] = strings.ReplaceAll(headerValue, placeholder, value)
		}
	}

	// Apply variables to data
	if len(newReq.Data) > 0 {
		data := string(newReq.Data)
		for name, value := range variables {
			placeholder := fmt.Sprintf("${%s}", name)
			data = strings.ReplaceAll(data, placeholder, value)
		}
		newReq.Data = []byte(data)
	}

	return &newReq, nil
}

// extractVariables extracts variables from a response using JSONPath expressions.
func extractVariables(resp ffuf.Response, extractVars map[string]string) (map[string]string, error) {
	// Check if response has data
	if len(resp.Data) == 0 {
		return make(map[string]string), nil
	}

	// Check if response is JSON
	contentType := resp.ContentType
	if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "text/json") {
		// Not JSON, return empty map
		return make(map[string]string), nil
	}

	// Extract variables using JSONPath
	result := make(map[string]string)
	for varName, jsonPath := range extractVars {
		value, err := EvaluateJSONPath(resp.Data, jsonPath)
		if err != nil {
			// Log the error but continue with other variables
			continue
		}
		result[varName] = value
	}

	return result, nil
}

// shouldRetry determines if a request should be retried based on status code and error.
func shouldRetry(statusCode int64, errType string, config *RetryConfig) bool {
	// Check status code
	for _, code := range config.RetryableStatusCodes {
		if statusCode == int64(code) {
			return true
		}
	}

	// Check error type
	for _, errString := range config.RetryableErrors {
		if strings.Contains(errType, errString) {
			return true
		}
	}

	return false
}
