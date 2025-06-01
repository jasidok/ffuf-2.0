package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestAPIClientCreation(t *testing.T) {
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}

	client := NewAPIClient(config)
	if client == nil {
		t.Error("Failed to create API client")
	}

	// Check default headers
	if client.commonHeaders["Accept"] != "application/json, application/xml, */*" {
		t.Errorf("Expected Accept header to be 'application/json, application/xml, */*', got '%s'", client.commonHeaders["Accept"])
	}

	if client.commonHeaders["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", client.commonHeaders["Content-Type"])
	}
}

func TestAPIClientAuthTokens(t *testing.T) {
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}

	client := NewAPIClient(config)
	
	// Test setting and getting auth tokens
	client.SetAuthToken("bearer", "test-token")
	if client.GetAuthToken("bearer") != "test-token" {
		t.Errorf("Expected bearer token to be 'test-token', got '%s'", client.GetAuthToken("bearer"))
	}

	client.SetAuthToken("apikey", "api-key-value")
	if client.GetAuthToken("apikey") != "api-key-value" {
		t.Errorf("Expected API key to be 'api-key-value', got '%s'", client.GetAuthToken("apikey"))
	}
}

func TestAPIClientSetCommonHeader(t *testing.T) {
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}

	client := NewAPIClient(config)
	
	// Test setting common headers
	client.SetCommonHeader("X-Custom-Header", "custom-value")
	if client.commonHeaders["X-Custom-Header"] != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be 'custom-value', got '%s'", client.commonHeaders["X-Custom-Header"])
	}
}

func TestAPIClientPrepare(t *testing.T) {
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}

	client := NewAPIClient(config)
	
	// Set auth token
	client.SetAuthToken("bearer", "test-token")
	
	// Create base request
	baseReq := &ffuf.Request{
		Method:  "GET",
		Url:     "https://example.com/api/FUZZ",
		Headers: make(map[string]string),
	}
	
	// Prepare request with input
	input := map[string][]byte{
		"FUZZ": []byte("test"),
	}
	
	req, err := client.Prepare(input, baseReq)
	if err != nil {
		t.Errorf("Error preparing request: %v", err)
	}
	
	// Check if URL was properly replaced
	if req.Url != "https://example.com/api/test" {
		t.Errorf("Expected URL to be 'https://example.com/api/test', got '%s'", req.Url)
	}
	
	// Check if auth token was applied
	if req.Headers["Authorization"] != "Bearer test-token" {
		t.Errorf("Expected Authorization header to be 'Bearer test-token', got '%s'", req.Headers["Authorization"])
	}
	
	// Check if common headers were applied
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", req.Headers["Content-Type"])
	}
}

func TestAPIClientExecute(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request has the expected headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}
		
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		
		// Return a JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()
	
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}
	
	client := NewAPIClient(config)
	
	// Set auth token
	client.SetAuthToken("bearer", "test-token")
	
	// Create request
	req := &ffuf.Request{
		Method: "GET",
		Url:    ts.URL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	
	// Execute request
	resp, err := client.Execute(req)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	
	// Check response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	
	if resp.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", resp.ContentType)
	}
	
	if string(resp.Data) != `{"status":"ok"}` {
		t.Errorf("Expected response body '{\"status\":\"ok\"}', got '%s'", string(resp.Data))
	}
}

func TestAPIClientWithDifferentContentTypes(t *testing.T) {
	// Create a test server that returns different content types
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.URL.Query().Get("content-type")
		w.Header().Set("Content-Type", contentType)
		
		switch contentType {
		case "application/json":
			w.Write([]byte(`{"status":"ok"}`))
		case "application/xml":
			w.Write([]byte(`<response><status>ok</status></response>`))
		case "text/plain":
			w.Write([]byte(`status: ok`))
		default:
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer ts.Close()
	
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         10,
		FollowRedirects: false,
	}
	
	client := NewAPIClient(config)
	
	// Test JSON content type
	req := &ffuf.Request{
		Method: "GET",
		Url:    ts.URL + "?content-type=application/json",
		Headers: map[string]string{},
	}
	
	resp, err := client.Execute(req)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	
	if resp.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", resp.ContentType)
	}
	
	// Test XML content type
	req.Url = ts.URL + "?content-type=application/xml"
	resp, err = client.Execute(req)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	
	if resp.ContentType != "application/xml" {
		t.Errorf("Expected content type 'application/xml', got '%s'", resp.ContentType)
	}
	
	// Test plain text content type
	req.Url = ts.URL + "?content-type=text/plain"
	resp, err = client.Execute(req)
	if err != nil {
		t.Errorf("Error executing request: %v", err)
	}
	
	if resp.ContentType != "text/plain" {
		t.Errorf("Expected content type 'text/plain', got '%s'", resp.ContentType)
	}
}

func TestAPIClientTimeout(t *testing.T) {
	// Create a test server that delays response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()
	
	// Create config with short timeout
	config := &ffuf.Config{
		Context:         context.Background(),
		Timeout:         1, // 1 second timeout
		FollowRedirects: false,
	}
	
	client := NewAPIClient(config)
	
	// Create request
	req := &ffuf.Request{
		Method:  "GET",
		Url:     ts.URL,
		Headers: map[string]string{},
	}
	
	// Execute request - should timeout
	_, err := client.Execute(req)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}