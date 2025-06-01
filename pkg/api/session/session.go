// Package session provides functionality for handling sessions in stateful API testing.
package session

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// SessionManager is responsible for managing API sessions
type SessionManager struct {
	sessions     map[string]*Session
	mu           sync.RWMutex
	defaultStore SessionStore
}

// Session represents a stateful API session
type Session struct {
	ID            string
	Name          string
	CookieJar     http.CookieJar
	Headers       map[string]string
	Variables     map[string]string
	Store         SessionStore
	LastAccessed  time.Time
	CreatedAt     time.Time
	RequestCount  int
	ResponseCount int
}

// SessionStore is an interface for storing session data
type SessionStore interface {
	// Get retrieves a value from the session store
	Get(key string) (interface{}, bool)
	// Set stores a value in the session store
	Set(key string, value interface{})
	// Delete removes a value from the session store
	Delete(key string)
	// Clear removes all values from the session store
	Clear()
	// GetAll returns all values in the session store
	GetAll() map[string]interface{}
}

// MemoryStore is an in-memory implementation of SessionStore
type MemoryStore struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewMemoryStore creates a new in-memory session store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the memory store
func (s *MemoryStore) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Set stores a value in the memory store
func (s *MemoryStore) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a value from the memory store
func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Clear removes all values from the memory store
func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
}

// GetAll returns all values in the memory store
func (s *MemoryStore) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{}, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*Session),
		defaultStore: NewMemoryStore(),
	}
}

// CreateSession creates a new session with the given name
func (m *SessionManager) CreateSession(name string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new cookie jar for the session
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, api.NewAPIError(fmt.Sprintf("Failed to create cookie jar: %v", err), 0)
	}

	// Generate a unique ID for the session
	id := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())

	// Create a new session
	session := &Session{
		ID:           id,
		Name:         name,
		CookieJar:    jar,
		Headers:      make(map[string]string),
		Variables:    make(map[string]string),
		Store:        NewMemoryStore(),
		LastAccessed: time.Now(),
		CreatedAt:    time.Now(),
	}

	// Store the session
	m.sessions[id] = session

	return session, nil
}

// GetSession retrieves a session by ID
func (m *SessionManager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil, api.NewAPIError(fmt.Sprintf("Session not found: %s", id), 0)
	}

	// Update last accessed time
	session.LastAccessed = time.Now()

	return session, nil
}

// DeleteSession removes a session by ID
func (m *SessionManager) DeleteSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[id]; !ok {
		return api.NewAPIError(fmt.Sprintf("Session not found: %s", id), 0)
	}

	delete(m.sessions, id)
	return nil
}

// GetAllSessions returns all sessions
func (m *SessionManager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// ClearAllSessions removes all sessions
func (m *SessionManager) ClearAllSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions = make(map[string]*Session)
}

// ApplySession applies session data to an HTTP request
func (m *SessionManager) ApplySession(sessionID string, req *http.Request) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Apply cookies from the session's cookie jar
	if session.CookieJar != nil {
		for _, cookie := range session.CookieJar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	// Apply headers from the session
	for key, value := range session.Headers {
		req.Header.Set(key, value)
	}

	return nil
}

// UpdateSessionFromResponse updates a session with data from an HTTP response
func (m *SessionManager) UpdateSessionFromResponse(sessionID string, resp *http.Response) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Update cookies in the session's cookie jar
	if session.CookieJar != nil && resp.Cookies() != nil {
		session.CookieJar.SetCookies(resp.Request.URL, resp.Cookies())
	}

	// Update response count
	session.ResponseCount++

	return nil
}

// ExtractVariablesFromResponse extracts variables from an HTTP response based on the provided extractors
func (m *SessionManager) ExtractVariablesFromResponse(sessionID string, resp *http.Response, extractors map[string]string) error {
	session, err := m.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Read the response body
	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return api.NewAPIError(fmt.Sprintf("Failed to decode response body: %v", err), 0)
	}

	// Extract variables based on the provided extractors
	for varName, path := range extractors {
		// Simple dot notation path extraction (in a real implementation, you would use a more robust solution)
		value, err := extractValueFromPath(respBody, path)
		if err != nil {
			return err
		}

		// Store the extracted variable
		if value != nil {
			session.Variables[varName] = fmt.Sprintf("%v", value)
		}
	}

	return nil
}

// extractValueFromPath extracts a value from a nested map using a dot notation path
func extractValueFromPath(data map[string]interface{}, path string) (interface{}, error) {
	// Split the path by dots
	parts := splitPath(path)

	// Navigate through the nested structure
	var current interface{} = data
	for _, part := range parts {
		// If current is a map, get the value for the current part
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
		} else {
			return nil, api.NewAPIError(fmt.Sprintf("Invalid path: %s", path), 0)
		}
	}

	return current, nil
}

// splitPath splits a path by dots, respecting quoted sections
func splitPath(path string) []string {
	// This is a simplified implementation
	// In a real implementation, you would handle quoted sections properly
	return splitDotPath(path)
}

// splitDotPath splits a path by dots
func splitDotPath(path string) []string {
	// This is a simplified implementation
	// In a real implementation, you would handle escaped dots and other special cases
	return strings.Split(path, ".")
}

// SessionRunner is a wrapper around ffuf.RunnerProvider that maintains session state
type SessionRunner struct {
	runner         ffuf.RunnerProvider
	sessionManager *SessionManager
	sessionID      string
	extractors     map[string]string
}

// NewSessionRunner creates a new session runner
func NewSessionRunner(runner ffuf.RunnerProvider, sessionManager *SessionManager, sessionID string) *SessionRunner {
	return &SessionRunner{
		runner:         runner,
		sessionManager: sessionManager,
		sessionID:      sessionID,
		extractors:     make(map[string]string),
	}
}

// AddExtractor adds a variable extractor
func (r *SessionRunner) AddExtractor(varName, path string) {
	r.extractors[varName] = path
}

// Execute executes a request with session handling
func (r *SessionRunner) Execute(req *ffuf.Request) (ffuf.Response, error) {
	// Get the session
	session, err := r.sessionManager.GetSession(r.sessionID)
	if err != nil {
		return ffuf.Response{}, err
	}

	// Update request count
	session.RequestCount++

	// Execute the request using the underlying runner
	resp, err := r.runner.Execute(req)
	if err != nil {
		return resp, err
	}

	// For session handling, we need to make a separate HTTP request to capture cookies and other session data
	// Create an HTTP client with a cookie jar
	client := &http.Client{
		Jar: session.CookieJar,
	}

	// Create an HTTP request from the ffuf request
	httpReq, err := http.NewRequest(req.Method, req.Url, nil)
	if err != nil {
		return resp, api.NewAPIError(fmt.Sprintf("Failed to create HTTP request: %v", err), 0)
	}

	// Apply session data to the request
	if err := r.sessionManager.ApplySession(r.sessionID, httpReq); err != nil {
		return resp, err
	}

	// Add headers from the ffuf request
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add request body if present
	if len(req.Data) > 0 {
		httpReq.Body = &bodyReader{data: req.Data}
		httpReq.ContentLength = int64(len(req.Data))
	}

	// Execute the HTTP request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return resp, api.NewAPIError(fmt.Sprintf("Failed to execute HTTP request: %v", err), 0)
	}
	defer httpResp.Body.Close()

	// Update session with response data
	if err := r.sessionManager.UpdateSessionFromResponse(r.sessionID, httpResp); err != nil {
		return resp, err
	}

	// Extract variables from the response if extractors are defined
	if len(r.extractors) > 0 {
		if err := r.sessionManager.ExtractVariablesFromResponse(r.sessionID, httpResp, r.extractors); err != nil {
			return resp, err
		}
	}

	return resp, nil
}

// bodyReader is a simple io.ReadCloser implementation for request bodies
type bodyReader struct {
	data []byte
	pos  int
}

// Read implements io.Reader
func (b *bodyReader) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return
}

// Close implements io.Closer
func (b *bodyReader) Close() error {
	return nil
}
