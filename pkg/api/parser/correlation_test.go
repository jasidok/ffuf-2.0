package parser

import (
	"encoding/json"
	"testing"
	"time"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestNewAPISession(t *testing.T) {
	session := NewAPISession("test_session")

	if session.ID != "test_session" {
		t.Errorf("Expected session ID to be 'test_session', got '%s'", session.ID)
	}

	if time.Since(session.StartTime) > time.Second {
		t.Errorf("Expected StartTime to be recent")
	}

	if time.Since(session.LastActivity) > time.Second {
		t.Errorf("Expected LastActivity to be recent")
	}

	if len(session.Requests) != 0 {
		t.Errorf("Expected Requests to be empty, got %d items", len(session.Requests))
	}

	if len(session.Responses) != 0 {
		t.Errorf("Expected Responses to be empty, got %d items", len(session.Responses))
	}

	if len(session.Correlations) != 0 {
		t.Errorf("Expected Correlations to be empty, got %d items", len(session.Correlations))
	}

	if len(session.ExtractedValues) != 0 {
		t.Errorf("Expected ExtractedValues to be empty, got %d items", len(session.ExtractedValues))
	}
}

func TestAddRequestAndResponse(t *testing.T) {
	session := NewAPISession("test_session")

	// Create a test request
	req := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}

	// Add the request to the session
	requestID := session.AddRequest(req)

	if len(session.Requests) != 1 {
		t.Errorf("Expected Requests to have 1 item, got %d", len(session.Requests))
	}

	if session.Requests[requestID] != req {
		t.Errorf("Expected request to be stored with ID %s", requestID)
	}

	// Create a test response
	resp := &ffuf.Response{
		Request:     req,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"id": 123, "name": "John Doe"}`),
	}

	// Add the response to the session
	responseID := session.AddResponse(resp, requestID)

	if len(session.Responses) != 1 {
		t.Errorf("Expected Responses to have 1 item, got %d", len(session.Responses))
	}

	if session.Responses[responseID] != resp {
		t.Errorf("Expected response to be stored with ID %s", responseID)
	}
}

func TestExtractValue(t *testing.T) {
	session := NewAPISession("test_session")

	// Create a test request
	req := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}

	// Create a test response
	resp := &ffuf.Response{
		Request:     req,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"id": 123, "name": "John Doe", "email": "john@example.com"}`),
	}

	// Add the request and response to the session
	requestID := session.AddRequest(req)
	responseID := session.AddResponse(resp, requestID)

	// Extract a value from the response
	value, err := session.ExtractValue(responseID, "$.name", "username")
	if err != nil {
		t.Fatalf("ExtractValue() error = %v", err)
	}

	if value != "John Doe" {
		t.Errorf("ExtractValue() got = %v, want %v", value, "John Doe")
	}

	// Check that the value was stored
	storedValue, ok := session.GetExtractedValue("username")
	if !ok {
		t.Errorf("GetExtractedValue() got ok = %v, want %v", ok, true)
	}

	if storedValue != "John Doe" {
		t.Errorf("GetExtractedValue() got = %v, want %v", storedValue, "John Doe")
	}
}

func TestDetectIDCorrelations(t *testing.T) {
	detector := NewCorrelationDetector()

	// Create test responses with matching IDs
	req1 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}
	resp1 := &ffuf.Response{
		Request:     req1,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"id": 123, "name": "John Doe"}`),
	}

	req2 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/posts?user_id=123",
	}
	resp2 := &ffuf.Response{
		Request:     req2,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"posts": [{"id": 1, "user_id": 123, "title": "First Post"}]}`),
	}

	// Parse the response bodies as JSON
	var jsonData1 interface{}
	if err := json.Unmarshal(resp1.Data, &jsonData1); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	parser1 := NewJSONPathParserFromObject(jsonData1)

	var jsonData2 interface{}
	if err := json.Unmarshal(resp2.Data, &jsonData2); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	parser2 := NewJSONPathParserFromObject(jsonData2)

	// Detect ID correlations
	correlations := detector.detectIDCorrelations(parser1, parser2, resp1, resp2)

	if len(correlations) == 0 {
		t.Errorf("Expected to find at least one correlation")
	}

	// Check that we found the ID correlation
	found := false
	for _, correlation := range correlations {
		if correlation.Type == CorrelationTypeID && 
		   correlation.SourcePath == "$.id" && 
		   correlation.SourceValue == "123" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find ID correlation between $.id and $.posts[0].user_id")
	}
}

func TestDetectReferenceCorrelations(t *testing.T) {
	detector := NewCorrelationDetector()

	// Create test responses with reference links
	req1 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}
	resp1 := &ffuf.Response{
		Request:     req1,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{
			"id": 123, 
			"name": "John Doe",
			"links": {
				"posts": "https://api.example.com/users/123/posts",
				"profile": "https://api.example.com/profiles/123"
			}
		}`),
	}

	req2 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123/posts",
	}
	resp2 := &ffuf.Response{
		Request:     req2,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"posts": [{"id": 1, "title": "First Post"}]}`),
	}

	// Parse the response bodies as JSON
	var jsonData1 interface{}
	if err := json.Unmarshal(resp1.Data, &jsonData1); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	parser1 := NewJSONPathParserFromObject(jsonData1)

	var jsonData2 interface{}
	if err := json.Unmarshal(resp2.Data, &jsonData2); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	parser2 := NewJSONPathParserFromObject(jsonData2)

	// Detect reference correlations
	correlations := detector.detectReferenceCorrelations(parser1, parser2, resp1, resp2)

	if len(correlations) == 0 {
		t.Errorf("Expected to find at least one correlation")
	}

	// Check that we found the reference correlation
	found := false
	for _, correlation := range correlations {
		if correlation.Type == CorrelationTypeReference && 
		   correlation.SourcePath == "$.links.posts" && 
		   correlation.SourceValue == "https://api.example.com/users/123/posts" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find reference correlation for posts link")
	}
}

func TestDetectCorrelations(t *testing.T) {
	detector := NewCorrelationDetector()
	sessionID := "test_session"
	session := detector.CreateSession(sessionID)

	// Create test requests and responses
	req1 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}
	resp1 := &ffuf.Response{
		Request:     req1,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{
			"id": 123, 
			"name": "John Doe",
			"links": {
				"posts": "https://api.example.com/users/123/posts"
			}
		}`),
	}

	req2 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123/posts",
	}
	resp2 := &ffuf.Response{
		Request:     req2,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{
			"posts": [
				{
					"id": 1, 
					"user_id": 123, 
					"title": "First Post"
				}
			]
		}`),
	}

	// Add the requests and responses to the session
	req1ID := session.AddRequest(req1)
	session.AddResponse(resp1, req1ID)
	req2ID := session.AddRequest(req2)
	session.AddResponse(resp2, req2ID)

	// Detect correlations
	correlations, err := detector.DetectCorrelations(sessionID)
	if err != nil {
		t.Fatalf("DetectCorrelations() error = %v", err)
	}

	if len(correlations) == 0 {
		t.Errorf("Expected to find at least one correlation")
	}

	// Check that the correlations were added to the session
	if len(session.Correlations) != len(correlations) {
		t.Errorf("Expected session to have %d correlations, got %d", len(correlations), len(session.Correlations))
	}

	// Check that we found both ID and reference correlations
	foundID := false
	foundRef := false

	for _, correlation := range correlations {
		if correlation.Type == CorrelationTypeID {
			foundID = true
		}
		if correlation.Type == CorrelationTypeReference {
			foundRef = true
		}
	}

	if !foundID {
		t.Errorf("Expected to find ID correlation")
	}

	if !foundRef {
		t.Errorf("Expected to find reference correlation")
	}
}

func TestGenerateCorrelatedRequest(t *testing.T) {
	detector := NewCorrelationDetector()
	sessionID := "test_session"
	session := detector.CreateSession(sessionID)

	// Create a correlation
	correlation := Correlation{
		Type:        CorrelationTypeID,
		SourcePath:  "$.id",
		TargetPath:  "$.user_id",
		SourceValue: "123",
		Confidence:  85,
	}

	// Add the correlation to the session
	session.AddCorrelation(correlation)

	// Generate a correlated request
	req, err := detector.GenerateCorrelatedRequest(sessionID, "https://api.example.com/users/{id}/posts", "GET", 0)
	if err != nil {
		t.Fatalf("GenerateCorrelatedRequest() error = %v", err)
	}

	// Check that the ID was substituted in the URL
	expectedURL := "https://api.example.com/users/123/posts"
	if req.Url != expectedURL {
		t.Errorf("GenerateCorrelatedRequest() got URL = %v, want %v", req.Url, expectedURL)
	}

	// Check that the method was set correctly
	if req.Method != "GET" {
		t.Errorf("GenerateCorrelatedRequest() got Method = %v, want %v", req.Method, "GET")
	}
}

func TestResponseParserCorrelateResponses(t *testing.T) {
	parser := NewResponseParser("application/json")

	// Create test requests and responses
	req1 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/users/123",
	}
	resp1 := &ffuf.Response{
		Request:     req1,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"id": 123, "name": "John Doe"}`),
	}

	req2 := &ffuf.Request{
		Method: "GET",
		Url:    "https://api.example.com/posts?user_id=123",
	}
	resp2 := &ffuf.Response{
		Request:     req2,
		StatusCode:  200,
		ContentType: "application/json",
		Data:        []byte(`{"posts": [{"id": 1, "user_id": 123, "title": "First Post"}]}`),
	}

	// Correlate the responses
	correlations, err := parser.CorrelateResponses(resp1, resp2)
	if err != nil {
		t.Fatalf("CorrelateResponses() error = %v", err)
	}

	if len(correlations) == 0 {
		t.Errorf("Expected to find at least one correlation")
	}

	// Check that we found the ID correlation
	found := false
	for _, correlation := range correlations {
		if correlation.Type == CorrelationTypeID && 
		   correlation.SourceValue == "123" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find ID correlation with value 123")
	}
}
