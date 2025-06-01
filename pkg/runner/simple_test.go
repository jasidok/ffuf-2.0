package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestSimpleRunnerExecute(t *testing.T) {
	// Create a test server that returns a simple response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// Create a config for the runner
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}

	// Create a runner
	runner := NewSimpleRunner(config, false)

	// Create a request
	req := &ffuf.Request{
		Method:  "GET",
		Url:     ts.URL,
		Headers: make(map[string]string),
	}

	// Execute the request
	resp, err := runner.Execute(req)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}

	// Check the response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check the content type
	if resp.ContentType != "application/json" {
		t.Errorf("Expected content type application/json, got %s", resp.ContentType)
	}

	// Check the response body
	if string(resp.Data) != `{"status":"ok"}` {
		t.Errorf("Expected response body {\"status\":\"ok\"}, got %s", string(resp.Data))
	}
}