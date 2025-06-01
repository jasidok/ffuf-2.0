// Package contenttype provides functionality for automatic content type detection and handling.
//
// This package includes utilities for detecting content types from request and response data,
// automatically setting appropriate content type headers, and handling different content types
// for API testing.
package contenttype

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// ContentType represents the type of content in a request or response
type ContentType int

const (
	// TypeUnknown represents an unknown content type
	TypeUnknown ContentType = iota
	// TypeJSON represents JSON content
	TypeJSON
	// TypeXML represents XML content
	TypeXML
	// TypeFormURLEncoded represents form URL-encoded content
	TypeFormURLEncoded
	// TypeMultipartForm represents multipart form data
	TypeMultipartForm
	// TypeText represents plain text content
	TypeText
	// TypeHTML represents HTML content
	TypeHTML
	// TypeGraphQL represents GraphQL content
	TypeGraphQL
)

// ContentTypeString returns the MIME type string for a ContentType
func ContentTypeString(ct ContentType) string {
	switch ct {
	case TypeJSON:
		return "application/json"
	case TypeXML:
		return "application/xml"
	case TypeFormURLEncoded:
		return "application/x-www-form-urlencoded"
	case TypeMultipartForm:
		return "multipart/form-data"
	case TypeText:
		return "text/plain"
	case TypeHTML:
		return "text/html"
	case TypeGraphQL:
		return "application/graphql"
	default:
		return "application/octet-stream"
	}
}

// ContentTypeFromString returns the ContentType for a MIME type string
func ContentTypeFromString(mimeType string) ContentType {
	mimeType = strings.ToLower(mimeType)

	// Extract the base MIME type without parameters
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err == nil {
		mimeType = mediaType
	}

	switch {
	case strings.Contains(mimeType, "application/json"):
		return TypeJSON
	case strings.Contains(mimeType, "application/xml") || strings.Contains(mimeType, "text/xml"):
		return TypeXML
	case strings.Contains(mimeType, "application/x-www-form-urlencoded"):
		return TypeFormURLEncoded
	case strings.Contains(mimeType, "multipart/form-data"):
		return TypeMultipartForm
	case strings.Contains(mimeType, "text/plain"):
		return TypeText
	case strings.Contains(mimeType, "text/html"):
		return TypeHTML
	case strings.Contains(mimeType, "application/graphql"):
		return TypeGraphQL
	default:
		return TypeUnknown
	}
}

// DetectContentType attempts to detect the content type from the data
func DetectContentType(data []byte) ContentType {
	// Empty data
	if len(data) == 0 {
		return TypeUnknown
	}

	// Try to detect JSON
	if isJSON(data) {
		return TypeJSON
	}

	// Try to detect HTML before XML since HTML can be parsed as XML
	if isHTML(data) {
		return TypeHTML
	}

	// Try to detect XML
	if isXML(data) {
		return TypeXML
	}

	// Try to detect form URL-encoded
	if isFormURLEncoded(data) {
		return TypeFormURLEncoded
	}

	// Try to detect GraphQL
	if isGraphQL(data) {
		return TypeGraphQL
	}

	// Default to text if it's printable
	if isPrintable(data) {
		return TypeText
	}

	return TypeUnknown
}

// isJSON checks if the data is valid JSON
func isJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// isXML checks if the data is valid XML
func isXML(data []byte) bool {
	return xml.Unmarshal(data, new(interface{})) == nil
}

// isFormURLEncoded checks if the data is form URL-encoded
func isFormURLEncoded(data []byte) bool {
	s := string(data)
	_, err := url.ParseQuery(s)
	return err == nil && strings.Contains(s, "=")
}

// isGraphQL checks if the data looks like a GraphQL query
func isGraphQL(data []byte) bool {
	s := string(data)
	return (strings.Contains(s, "query") || strings.Contains(s, "mutation")) &&
		(strings.Contains(s, "{") && strings.Contains(s, "}"))
}

// isHTML checks if the data looks like HTML
func isHTML(data []byte) bool {
	s := strings.ToLower(string(data))
	return strings.Contains(s, "<html") || strings.Contains(s, "<body") ||
		strings.Contains(s, "<!doctype html") || strings.Contains(s, "<head") ||
		strings.Contains(s, "</html>") || strings.Contains(s, "</body>")
}

// isPrintable checks if the data is printable text
func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 && !isAllowedControlChar(b) {
			return false
		}
	}
	return true
}

// isAllowedControlChar checks if a control character is allowed in printable text
func isAllowedControlChar(b byte) bool {
	// Allow tab, newline, and carriage return
	return b == 9 || b == 10 || b == 13
}

// ContentTypeHandler provides methods for handling content types in requests and responses
type ContentTypeHandler struct {
	DefaultType ContentType
}

// NewContentTypeHandler creates a new ContentTypeHandler with the specified default type
func NewContentTypeHandler(defaultType ContentType) *ContentTypeHandler {
	return &ContentTypeHandler{
		DefaultType: defaultType,
	}
}

// DetectRequestContentType detects the content type of a request
func (h *ContentTypeHandler) DetectRequestContentType(req *ffuf.Request) ContentType {
	// Check if Content-Type header is set
	if contentTypeHeader, ok := req.Headers["Content-Type"]; ok {
		return ContentTypeFromString(contentTypeHeader)
	}

	// If no Content-Type header, try to detect from the request data
	if len(req.Data) > 0 {
		return DetectContentType(req.Data)
	}

	// Default to the handler's default type
	return h.DefaultType
}

// SetRequestContentType sets the Content-Type header based on the detected content type
func (h *ContentTypeHandler) SetRequestContentType(req *ffuf.Request) error {
	contentType := h.DetectRequestContentType(req)
	if contentType == TypeUnknown {
		return api.NewAPIError("Could not detect content type for request", 0)
	}

	req.Headers["Content-Type"] = ContentTypeString(contentType)
	return nil
}

// ConvertRequestData converts request data between different content types
func (h *ContentTypeHandler) ConvertRequestData(req *ffuf.Request, targetType ContentType) error {
	sourceType := h.DetectRequestContentType(req)

	// If source and target are the same, no conversion needed
	if sourceType == targetType {
		return nil
	}

	// If no data, nothing to convert
	if len(req.Data) == 0 {
		req.Headers["Content-Type"] = ContentTypeString(targetType)
		return nil
	}

	// Handle conversions
	switch {
	case sourceType == TypeJSON && targetType == TypeFormURLEncoded:
		return h.convertJSONToForm(req)
	case sourceType == TypeFormURLEncoded && targetType == TypeJSON:
		return h.convertFormToJSON(req)
	default:
		return api.NewAPIError(fmt.Sprintf("Conversion from %s to %s is not supported",
			ContentTypeString(sourceType), ContentTypeString(targetType)), 0)
	}
}

// convertJSONToForm converts JSON data to form URL-encoded
func (h *ContentTypeHandler) convertJSONToForm(req *ffuf.Request) error {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(req.Data, &jsonData); err != nil {
		return api.NewAPIError("Failed to parse JSON data: "+err.Error(), 0)
	}

	values := url.Values{}
	for key, val := range jsonData {
		// Convert value to string
		switch v := val.(type) {
		case string:
			values.Add(key, v)
		case float64:
			values.Add(key, fmt.Sprintf("%g", v))
		case bool:
			values.Add(key, fmt.Sprintf("%t", v))
		default:
			// For complex types, convert back to JSON
			jsonVal, err := json.Marshal(v)
			if err != nil {
				return api.NewAPIError("Failed to convert JSON value: "+err.Error(), 0)
			}
			values.Add(key, string(jsonVal))
		}
	}

	req.Data = []byte(values.Encode())
	req.Headers["Content-Type"] = ContentTypeString(TypeFormURLEncoded)
	return nil
}

// convertFormToJSON converts form URL-encoded data to JSON
func (h *ContentTypeHandler) convertFormToJSON(req *ffuf.Request) error {
	values, err := url.ParseQuery(string(req.Data))
	if err != nil {
		return api.NewAPIError("Failed to parse form data: "+err.Error(), 0)
	}

	jsonData := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			jsonData[key] = vals[0]
		} else {
			jsonData[key] = vals
		}
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return api.NewAPIError("Failed to convert to JSON: "+err.Error(), 0)
	}

	req.Data = jsonBytes
	req.Headers["Content-Type"] = ContentTypeString(TypeJSON)
	return nil
}

// ProcessResponse processes a response and extracts content type information
func (h *ContentTypeHandler) ProcessResponse(resp *http.Response) (ContentType, []byte, error) {
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TypeUnknown, nil, api.NewAPIError("Failed to read response body: "+err.Error(), 0)
	}

	// Create a new reader with the same data for the next consumer
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// Get content type from header
	contentType := ContentTypeFromString(resp.Header.Get("Content-Type"))

	// If content type is unknown, try to detect from the body
	if contentType == TypeUnknown && len(body) > 0 {
		contentType = DetectContentType(body)
	}

	return contentType, body, nil
}

// GetAcceptHeader returns an appropriate Accept header value for the given content type
func (h *ContentTypeHandler) GetAcceptHeader(contentType ContentType) string {
	switch contentType {
	case TypeJSON:
		return "application/json"
	case TypeXML:
		return "application/xml, text/xml"
	case TypeHTML:
		return "text/html"
	case TypeText:
		return "text/plain"
	case TypeGraphQL:
		return "application/json" // GraphQL responses are typically JSON
	default:
		return "*/*"
	}
}

// SetAcceptHeader sets the Accept header based on the preferred content type
func (h *ContentTypeHandler) SetAcceptHeader(req *ffuf.Request, preferredType ContentType) {
	req.Headers["Accept"] = h.GetAcceptHeader(preferredType)
}
