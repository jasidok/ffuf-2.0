// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// CorrelationType represents the type of correlation between API requests/responses
type CorrelationType string

const (
	// CorrelationTypeID represents a correlation based on ID values
	CorrelationTypeID CorrelationType = "id"
	// CorrelationTypeReference represents a correlation based on reference links
	CorrelationTypeReference CorrelationType = "reference"
	// CorrelationTypeParentChild represents a parent-child relationship
	CorrelationTypeParentChild CorrelationType = "parent_child"
	// CorrelationTypeSequence represents a sequential relationship
	CorrelationTypeSequence CorrelationType = "sequence"
	// CorrelationTypeUnknown represents an unknown correlation type
	CorrelationTypeUnknown CorrelationType = "unknown"
)

// Correlation represents a detected correlation between API requests/responses
type Correlation struct {
	// Type is the type of correlation
	Type CorrelationType `json:"type"`
	// SourcePath is the JSONPath in the source response
	SourcePath string `json:"source_path,omitempty"`
	// TargetPath is the JSONPath or parameter in the target request
	TargetPath string `json:"target_path,omitempty"`
	// SourceValue is the value from the source response
	SourceValue string `json:"source_value,omitempty"`
	// SourceResponse is a reference to the source response
	SourceResponse *ffuf.Response `json:"-"`
	// TargetRequest is a reference to the target request
	TargetRequest *ffuf.Request `json:"-"`
	// Confidence is a score from 0-100 indicating confidence in the correlation
	Confidence int `json:"confidence"`
	// Description is a human-readable description of the correlation
	Description string `json:"description,omitempty"`
}

// APISession represents a session of related API requests and responses
type APISession struct {
	// ID is a unique identifier for the session
	ID string `json:"id"`
	// StartTime is when the session started
	StartTime time.Time `json:"start_time"`
	// LastActivity is when the session was last active
	LastActivity time.Time `json:"last_activity"`
	// Requests is a map of request IDs to requests
	Requests map[string]*ffuf.Request `json:"-"`
	// Responses is a map of response IDs to responses
	Responses map[string]*ffuf.Response `json:"-"`
	// Correlations is a list of detected correlations
	Correlations []Correlation `json:"correlations"`
	// ExtractedValues is a map of extracted values from responses
	ExtractedValues map[string]string `json:"extracted_values"`
	// mutex for thread safety
	mutex sync.RWMutex
}

// CorrelationDetector provides methods for detecting correlations between API requests/responses
type CorrelationDetector struct {
	// Sessions is a map of session IDs to sessions
	Sessions map[string]*APISession
	// Common patterns for different correlation types
	patterns map[string]*regexp.Regexp
	// JSONPath parser for extracting values
	jsonPathParser *JSONPathParser
	// mutex for thread safety
	mutex sync.RWMutex
}

// NewCorrelationDetector creates a new CorrelationDetector
func NewCorrelationDetector() *CorrelationDetector {
	detector := &CorrelationDetector{
		Sessions: make(map[string]*APISession),
		patterns: make(map[string]*regexp.Regexp),
	}

	// Initialize patterns for different correlation types
	detector.patterns["id"] = regexp.MustCompile(`(?i)(id|uuid|key)`)
	detector.patterns["reference"] = regexp.MustCompile(`(?i)(ref|reference|link|url|href)`)
	detector.patterns["parentChild"] = regexp.MustCompile(`(?i)(parent|child|owner|owned)`)

	return detector
}

// NewAPISession creates a new API session
func NewAPISession(id string) *APISession {
	now := time.Now()
	return &APISession{
		ID:              id,
		StartTime:       now,
		LastActivity:    now,
		Requests:        make(map[string]*ffuf.Request),
		Responses:       make(map[string]*ffuf.Response),
		Correlations:    make([]Correlation, 0),
		ExtractedValues: make(map[string]string),
	}
}

// AddRequest adds a request to the session
func (s *APISession) AddRequest(req *ffuf.Request) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate a request ID if not already present
	requestID := fmt.Sprintf("req_%d", len(s.Requests)+1)
	s.Requests[requestID] = req
	s.LastActivity = time.Now()

	return requestID
}

// AddResponse adds a response to the session
func (s *APISession) AddResponse(resp *ffuf.Response, requestID string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate a response ID
	responseID := fmt.Sprintf("resp_%d", len(s.Responses)+1)
	s.Responses[responseID] = resp
	s.LastActivity = time.Now()

	return responseID
}

// ExtractValue extracts a value from a response using JSONPath
func (s *APISession) ExtractValue(responseID, jsonPath, name string) (string, error) {
	s.mutex.RLock()
	resp, ok := s.Responses[responseID]
	s.mutex.RUnlock()

	if !ok {
		return "", api.NewAPIError(fmt.Sprintf("Response with ID %s not found", responseID), 0)
	}

	// Parse the response body as JSON
	var jsonData interface{}
	if err := json.Unmarshal(resp.Data, &jsonData); err != nil {
		return "", api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
	}

	// Create a JSONPath parser
	parser := NewJSONPathParserFromObject(jsonData)

	// Evaluate the JSONPath expression
	value, err := parser.EvaluateToString(jsonPath)
	if err != nil {
		return "", err
	}

	// Store the extracted value
	s.mutex.Lock()
	s.ExtractedValues[name] = value
	s.mutex.Unlock()

	return value, nil
}

// GetExtractedValue gets a previously extracted value
func (s *APISession) GetExtractedValue(name string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.ExtractedValues[name]
	return value, ok
}

// AddCorrelation adds a correlation to the session
func (s *APISession) AddCorrelation(correlation Correlation) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Correlations = append(s.Correlations, correlation)
}

// CreateSession creates a new API session
func (d *CorrelationDetector) CreateSession(id string) *APISession {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	session := NewAPISession(id)
	d.Sessions[id] = session
	return session
}

// GetSession gets an API session by ID
func (d *CorrelationDetector) GetSession(id string) (*APISession, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	session, ok := d.Sessions[id]
	if !ok {
		return nil, api.NewAPIError(fmt.Sprintf("Session with ID %s not found", id), 0)
	}

	return session, nil
}

// DetectCorrelations detects correlations between requests and responses in a session
func (d *CorrelationDetector) DetectCorrelations(sessionID string) ([]Correlation, error) {
	session, err := d.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	correlations := make([]Correlation, 0)

	// Get all responses in the session
	session.mutex.RLock()
	responses := make([]*ffuf.Response, 0, len(session.Responses))
	for _, resp := range session.Responses {
		responses = append(responses, resp)
	}
	session.mutex.RUnlock()

	// Process each response
	for i, resp1 := range responses {
		// Skip if not JSON
		if !strings.Contains(resp1.ContentType, "application/json") {
			continue
		}

		// Parse the response body as JSON
		var jsonData1 interface{}
		if err := json.Unmarshal(resp1.Data, &jsonData1); err != nil {
			continue
		}

		// Create a JSONPath parser
		parser1 := NewJSONPathParserFromObject(jsonData1)

		// Compare with other responses
		for j, resp2 := range responses {
			// Skip self-comparison
			if i == j {
				continue
			}

			// Skip if not JSON
			if !strings.Contains(resp2.ContentType, "application/json") {
				continue
			}

			// Parse the response body as JSON
			var jsonData2 interface{}
			if err := json.Unmarshal(resp2.Data, &jsonData2); err != nil {
				continue
			}

			// Create a JSONPath parser
			parser2 := NewJSONPathParserFromObject(jsonData2)

			// Detect ID correlations
			idCorrelations := d.detectIDCorrelations(parser1, parser2, resp1, resp2)
			correlations = append(correlations, idCorrelations...)

			// Detect reference correlations
			refCorrelations := d.detectReferenceCorrelations(parser1, parser2, resp1, resp2)
			correlations = append(correlations, refCorrelations...)
		}
	}

	// Add the correlations to the session
	for _, correlation := range correlations {
		session.AddCorrelation(correlation)
	}

	return correlations, nil
}

// detectIDCorrelations detects ID-based correlations between two responses
func (d *CorrelationDetector) detectIDCorrelations(parser1, parser2 *JSONPathParser, resp1, resp2 *ffuf.Response) []Correlation {
	correlations := make([]Correlation, 0)

	// Extract all ID-like fields from both responses
	idFields1 := d.extractIDFields(parser1)
	idFields2 := d.extractIDFields(parser2)

	// Compare ID values
	for path1, value1 := range idFields1 {
		for path2, value2 := range idFields2 {
			if value1 == value2 && value1 != "" {
				// Found a matching ID
				correlation := Correlation{
					Type:           CorrelationTypeID,
					SourcePath:     path1,
					TargetPath:     path2,
					SourceValue:    value1,
					SourceResponse: resp1,
					TargetRequest:  resp2.Request,
					Confidence:     85,
					Description:    fmt.Sprintf("ID correlation: %s matches %s with value %s", path1, path2, value1),
				}
				correlations = append(correlations, correlation)
			}
		}
	}

	return correlations
}

// detectReferenceCorrelations detects reference-based correlations between two responses
func (d *CorrelationDetector) detectReferenceCorrelations(parser1, parser2 *JSONPathParser, resp1, resp2 *ffuf.Response) []Correlation {
	correlations := make([]Correlation, 0)

	// Extract all URL-like fields from both responses
	urlFields1 := d.extractURLFields(parser1)
	urlFields2 := d.extractURLFields(parser2)

	// Check if any URL in resp1 points to the endpoint in resp2
	for path1, url1 := range urlFields1 {
		if strings.Contains(url1, resp2.Request.Url) {
			// Found a reference
			correlation := Correlation{
				Type:           CorrelationTypeReference,
				SourcePath:     path1,
				TargetPath:     "request.url",
				SourceValue:    url1,
				SourceResponse: resp1,
				TargetRequest:  resp2.Request,
				Confidence:     90,
				Description:    fmt.Sprintf("Reference correlation: %s references %s", path1, resp2.Request.Url),
			}
			correlations = append(correlations, correlation)
		}
	}

	// Check if any URL in resp2 points to the endpoint in resp1
	for path2, url2 := range urlFields2 {
		if strings.Contains(url2, resp1.Request.Url) {
			// Found a reference
			correlation := Correlation{
				Type:           CorrelationTypeReference,
				SourcePath:     path2,
				TargetPath:     "request.url",
				SourceValue:    url2,
				SourceResponse: resp2,
				TargetRequest:  resp1.Request,
				Confidence:     90,
				Description:    fmt.Sprintf("Reference correlation: %s references %s", path2, resp1.Request.Url),
			}
			correlations = append(correlations, correlation)
		}
	}

	return correlations
}

// extractIDFields extracts all ID-like fields from a JSON object
func (d *CorrelationDetector) extractIDFields(parser *JSONPathParser) map[string]string {
	fields := make(map[string]string)

	// Use reflection to get the underlying data
	data := parser.data

	// Extract ID fields recursively
	d.extractIDFieldsRecursive(data, "$", fields)

	return fields
}

// extractIDFieldsRecursive recursively extracts ID-like fields from a JSON object
func (d *CorrelationDetector) extractIDFieldsRecursive(data interface{}, path string, fields map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair
		for key, value := range v {
			newPath := path
			if newPath == "$" {
				newPath = "$." + key
			} else {
				newPath = path + "." + key
			}

			// Check if key is ID-like
			if d.patterns["id"].MatchString(key) {
				// Extract the value as string
				if strValue, ok := value.(string); ok {
					fields[newPath] = strValue
				} else if numValue, ok := value.(float64); ok {
					fields[newPath] = fmt.Sprintf("%v", numValue)
				}
			}

			// Recursively process the value
			d.extractIDFieldsRecursive(value, newPath, fields)
		}
	case []interface{}:
		// Process each element in the array
		for i, elem := range v {
			newPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			d.extractIDFieldsRecursive(elem, newPath, fields)
		}
	}
}

// extractURLFields extracts all URL-like fields from a JSON object
func (d *CorrelationDetector) extractURLFields(parser *JSONPathParser) map[string]string {
	fields := make(map[string]string)

	// Use reflection to get the underlying data
	data := parser.data

	// Extract URL fields recursively
	d.extractURLFieldsRecursive(data, "$", fields)

	return fields
}

// extractURLFieldsRecursive recursively extracts URL-like fields from a JSON object
func (d *CorrelationDetector) extractURLFieldsRecursive(data interface{}, path string, fields map[string]string) {
	urlPattern := regexp.MustCompile(`^(https?://|/)[\w\-\./%&=\?]+$`)

	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair
		for key, value := range v {
			newPath := path
			if newPath == "$" {
				newPath = "$." + key
			} else {
				newPath = path + "." + key
			}

			// Check if key suggests this might be a URL
			if d.patterns["reference"].MatchString(key) {
				// Extract the value as string
				if strValue, ok := value.(string); ok {
					if urlPattern.MatchString(strValue) {
						fields[newPath] = strValue
					}
				}
			}

			// Recursively process the value
			d.extractURLFieldsRecursive(value, newPath, fields)
		}
	case []interface{}:
		// Process each element in the array
		for i, elem := range v {
			newPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			d.extractURLFieldsRecursive(elem, newPath, fields)
		}
	case string:
		// Check if the string looks like a URL
		if urlPattern.MatchString(v) {
			fields[path] = v
		}
	}
}

// GenerateCorrelatedRequest generates a new request based on correlations
func (d *CorrelationDetector) GenerateCorrelatedRequest(sessionID, baseURL, method string, correlationID int) (*ffuf.Request, error) {
	session, err := d.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Get the correlation
	session.mutex.RLock()
	if correlationID < 0 || correlationID >= len(session.Correlations) {
		session.mutex.RUnlock()
		return nil, api.NewAPIError(fmt.Sprintf("Correlation with ID %d not found", correlationID), 0)
	}
	correlation := session.Correlations[correlationID]
	session.mutex.RUnlock()

	// Create a new request
	req := &ffuf.Request{
		Method:  method,
		Url:     baseURL,
		Headers: make(map[string]string),
	}

	// Apply correlation to the request
	switch correlation.Type {
	case CorrelationTypeID:
		// Add the ID as a query parameter or path parameter
		if strings.Contains(baseURL, "{id}") {
			req.Url = strings.Replace(baseURL, "{id}", correlation.SourceValue, -1)
		} else {
			if strings.Contains(baseURL, "?") {
				req.Url = baseURL + "&id=" + correlation.SourceValue
			} else {
				req.Url = baseURL + "?id=" + correlation.SourceValue
			}
		}
	case CorrelationTypeReference:
		// Use the referenced URL
		req.Url = correlation.SourceValue
	}

	return req, nil
}

// CorrelateResponses is a convenience method on ResponseParser to correlate responses
func (p *ResponseParser) CorrelateResponses(resp1, resp2 *ffuf.Response) ([]Correlation, error) {
	detector := NewCorrelationDetector()
	sessionID := "temp_session"
	session := detector.CreateSession(sessionID)

	// Add the responses to the session
	req1ID := session.AddRequest(resp1.Request)
	session.AddResponse(resp1, req1ID)
	req2ID := session.AddRequest(resp2.Request)
	session.AddResponse(resp2, req2ID)

	// Detect correlations
	return detector.DetectCorrelations(sessionID)
}