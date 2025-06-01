package session

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// MockRunner is a mock implementation of ffuf.RunnerProvider for testing
type MockRunner struct{}

// Prepare implements ffuf.RunnerProvider
func (r *MockRunner) Prepare(input map[string][]byte, basereq *ffuf.Request) (ffuf.Request, error) {
	return *basereq, nil
}

// Execute implements ffuf.RunnerProvider
func (r *MockRunner) Execute(req *ffuf.Request) (ffuf.Response, error) {
	resp := ffuf.Response{
		StatusCode:    200,
		ContentLength: int64(len(req.Data)),
		Data:          req.Data,
		Headers:       make(map[string][]string),
		Request:       req,
	}
	return resp, nil
}

// Dump implements ffuf.RunnerProvider
func (r *MockRunner) Dump(req *ffuf.Request) ([]byte, error) {
	return req.Data, nil
}

func TestSessionManager(t *testing.T) {
	// Create a new session manager
	manager := NewSessionManager()

	// Create a new session
	session, err := manager.CreateSession("test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Check that the session was created with the correct name
	if session.Name != "test-session" {
		t.Errorf("Expected session name to be test-session, got %s", session.Name)
	}

	// Get the session by ID
	retrievedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	// Check that the retrieved session is the same as the created session
	if retrievedSession.ID != session.ID {
		t.Errorf("Expected retrieved session ID to be %s, got %s", session.ID, retrievedSession.ID)
	}

	// Delete the session
	err = manager.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Check that the session was deleted
	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("Expected error when getting deleted session, got nil")
	}
}

func TestSessionStore(t *testing.T) {
	// Create a new memory store
	store := NewMemoryStore()

	// Set a value
	store.Set("key1", "value1")

	// Get the value
	value, ok := store.Get("key1")
	if !ok {
		t.Error("Expected to find key1 in store, but it was not found")
	}
	if value != "value1" {
		t.Errorf("Expected value for key1 to be value1, got %v", value)
	}

	// Delete the value
	store.Delete("key1")

	// Check that the value was deleted
	_, ok = store.Get("key1")
	if ok {
		t.Error("Expected key1 to be deleted, but it was found")
	}

	// Set multiple values
	store.Set("key2", "value2")
	store.Set("key3", "value3")

	// Get all values
	allValues := store.GetAll()
	if len(allValues) != 2 {
		t.Errorf("Expected 2 values in store, got %d", len(allValues))
	}
	if allValues["key2"] != "value2" {
		t.Errorf("Expected value for key2 to be value2, got %v", allValues["key2"])
	}
	if allValues["key3"] != "value3" {
		t.Errorf("Expected value for key3 to be value3, got %v", allValues["key3"])
	}

	// Clear the store
	store.Clear()

	// Check that the store is empty
	allValues = store.GetAll()
	if len(allValues) != 0 {
		t.Errorf("Expected store to be empty, got %d values", len(allValues))
	}
}

func TestSessionRunner(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set a cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: "test-value",
		})
		// Return a JSON response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key":"value"}`))
	}))
	defer server.Close()

	// Create a new session manager
	manager := NewSessionManager()

	// Create a new session
	session, err := manager.CreateSession("test-session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create a mock runner
	runner := &MockRunner{}

	// Create a session runner
	sessionRunner := NewSessionRunner(runner, manager, session.ID)

	// Create a request
	req := &ffuf.Request{
		Method: "GET",
		Url:    server.URL,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	// Execute the request
	resp, err := sessionRunner.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Check the response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Check that the session has cookies
	reqURL, err := url.Parse(req.Url)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	cookies := session.CookieJar.Cookies(reqURL)
	if len(cookies) == 0 {
		t.Error("Expected session to have cookies, but it has none")
	}

	// Check that the session cookie was set
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Error("Expected session cookie to be set, but it was not found")
	} else if sessionCookie.Value != "test-value" {
		t.Errorf("Expected session cookie value to be test-value, got %s", sessionCookie.Value)
	}

	// Add an extractor
	sessionRunner.AddExtractor("key", "key")

	// Execute the request again
	resp, err = sessionRunner.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Check that the variable was extracted
	value, ok := session.Variables["key"]
	if !ok {
		t.Error("Expected key variable to be extracted, but it was not found")
	} else if value != "value" {
		t.Errorf("Expected key variable value to be value, got %s", value)
	}
}
