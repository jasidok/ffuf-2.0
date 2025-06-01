package contenttype

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestContentTypeString(t *testing.T) {
	tests := []struct {
		name        string
		contentType ContentType
		expected    string
	}{
		{"JSON", TypeJSON, "application/json"},
		{"XML", TypeXML, "application/xml"},
		{"Form URL Encoded", TypeFormURLEncoded, "application/x-www-form-urlencoded"},
		{"Multipart Form", TypeMultipartForm, "multipart/form-data"},
		{"Text", TypeText, "text/plain"},
		{"HTML", TypeHTML, "text/html"},
		{"GraphQL", TypeGraphQL, "application/graphql"},
		{"Unknown", TypeUnknown, "application/octet-stream"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ContentTypeString(test.contentType)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestContentTypeFromString(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected ContentType
	}{
		{"JSON", "application/json", TypeJSON},
		{"JSON with charset", "application/json; charset=utf-8", TypeJSON},
		{"XML", "application/xml", TypeXML},
		{"Text XML", "text/xml", TypeXML},
		{"Form URL Encoded", "application/x-www-form-urlencoded", TypeFormURLEncoded},
		{"Multipart Form", "multipart/form-data", TypeMultipartForm},
		{"Text", "text/plain", TypeText},
		{"HTML", "text/html", TypeHTML},
		{"GraphQL", "application/graphql", TypeGraphQL},
		{"Unknown", "application/unknown", TypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ContentTypeFromString(test.mimeType)
			if result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ContentType
	}{
		{"Empty", []byte{}, TypeUnknown},
		{"JSON Object", []byte(`{"key":"value"}`), TypeJSON},
		{"JSON Array", []byte(`[1,2,3]`), TypeJSON},
		{"XML", []byte(`<root><item>value</item></root>`), TypeXML},
		{"Form URL Encoded", []byte(`key1=value1&key2=value2`), TypeFormURLEncoded},
		{"GraphQL Query", []byte(`query { user(id: 1) { name } }`), TypeGraphQL},
		{"HTML", []byte(`<!DOCTYPE html><html><body>Hello</body></html>`), TypeHTML},
		{"Plain Text", []byte(`Hello, world!`), TypeText},
		{"Binary", []byte{0x00, 0x01, 0x02, 0x03}, TypeUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := DetectContentType(test.data)
			if result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestDetectRequestContentType(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name     string
		request  *ffuf.Request
		expected ContentType
	}{
		{
			"With Content-Type Header",
			&ffuf.Request{
				Headers: map[string]string{"Content-Type": "application/json"},
				Data:    []byte{},
			},
			TypeJSON,
		},
		{
			"Without Header, With JSON Data",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte(`{"key":"value"}`),
			},
			TypeJSON,
		},
		{
			"Without Header, With XML Data",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte(`<root><item>value</item></root>`),
			},
			TypeXML,
		},
		{
			"Without Header or Data",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte{},
			},
			TypeJSON, // Default type
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := handler.DetectRequestContentType(test.request)
			if result != test.expected {
				t.Errorf("Expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestSetRequestContentType(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name           string
		request        *ffuf.Request
		expectedHeader string
		expectError    bool
	}{
		{
			"JSON Data",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte(`{"key":"value"}`),
			},
			"application/json",
			false,
		},
		{
			"XML Data",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte(`<root><item>value</item></root>`),
			},
			"application/xml",
			false,
		},
		{
			"No Data, Default Type",
			&ffuf.Request{
				Headers: map[string]string{},
				Data:    []byte{},
			},
			"application/json",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := handler.SetRequestContentType(test.request)

			if test.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !test.expectError {
				contentType, ok := test.request.Headers["Content-Type"]
				if !ok {
					t.Error("Content-Type header not set")
				} else if contentType != test.expectedHeader {
					t.Errorf("Expected Content-Type %s, got %s", test.expectedHeader, contentType)
				}
			}
		})
	}
}

func TestConvertRequestData(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name           string
		request        *ffuf.Request
		targetType     ContentType
		expectedData   string
		expectedHeader string
		expectError    bool
	}{
		{
			"JSON to Form",
			&ffuf.Request{
				Headers: map[string]string{"Content-Type": "application/json"},
				Data:    []byte(`{"key1":"value1","key2":"value2"}`),
			},
			TypeFormURLEncoded,
			"key1=value1&key2=value2",
			"application/x-www-form-urlencoded",
			false,
		},
		{
			"Form to JSON",
			&ffuf.Request{
				Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				Data:    []byte(`key1=value1&key2=value2`),
			},
			TypeJSON,
			`{"key1":"value1","key2":"value2"}`,
			"application/json",
			false,
		},
		{
			"Same Type",
			&ffuf.Request{
				Headers: map[string]string{"Content-Type": "application/json"},
				Data:    []byte(`{"key":"value"}`),
			},
			TypeJSON,
			`{"key":"value"}`,
			"application/json",
			false,
		},
		{
			"Unsupported Conversion",
			&ffuf.Request{
				Headers: map[string]string{"Content-Type": "application/json"},
				Data:    []byte(`{"key":"value"}`),
			},
			TypeXML,
			`{"key":"value"}`,
			"application/json",
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := handler.ConvertRequestData(test.request, test.targetType)

			if test.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !test.expectError {
				// For JSON, we need to normalize the data for comparison
				if test.targetType == TypeJSON {
					var jsonData map[string]interface{}
					if err := json.Unmarshal(test.request.Data, &jsonData); err != nil {
						t.Errorf("Failed to parse JSON data: %v", err)
					}

					var expectedData map[string]interface{}
					if err := json.Unmarshal([]byte(test.expectedData), &expectedData); err != nil {
						t.Errorf("Failed to parse expected JSON data: %v", err)
					}

					// Compare keys and values
					for key, expectedVal := range expectedData {
						actualVal, ok := jsonData[key]
						if !ok {
							t.Errorf("Expected key %s not found in result", key)
						} else if expectedVal != actualVal {
							t.Errorf("For key %s, expected %v, got %v", key, expectedVal, actualVal)
						}
					}
				} else {
					// For other formats, compare as strings
					if string(test.request.Data) != test.expectedData {
						t.Errorf("Expected data %s, got %s", test.expectedData, string(test.request.Data))
					}
				}

				contentType, ok := test.request.Headers["Content-Type"]
				if !ok {
					t.Error("Content-Type header not set")
				} else if contentType != test.expectedHeader {
					t.Errorf("Expected Content-Type %s, got %s", test.expectedHeader, contentType)
				}
			}
		})
	}
}

func TestProcessResponse(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name            string
		response        *http.Response
		expectedType    ContentType
		expectedBody    string
		expectError     bool
	}{
		{
			"JSON Response",
			&http.Response{
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   ioutil.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
			TypeJSON,
			`{"key":"value"}`,
			false,
		},
		{
			"XML Response",
			&http.Response{
				Header: http.Header{"Content-Type": []string{"application/xml"}},
				Body:   ioutil.NopCloser(bytes.NewReader([]byte(`<root><item>value</item></root>`))),
			},
			TypeXML,
			`<root><item>value</item></root>`,
			false,
		},
		{
			"Unknown Content Type, Detect from Body",
			&http.Response{
				Header: http.Header{},
				Body:   ioutil.NopCloser(bytes.NewReader([]byte(`{"key":"value"}`))),
			},
			TypeJSON,
			`{"key":"value"}`,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contentType, body, err := handler.ProcessResponse(test.response)

			if test.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !test.expectError {
				if contentType != test.expectedType {
					t.Errorf("Expected content type %d, got %d", test.expectedType, contentType)
				}

				if string(body) != test.expectedBody {
					t.Errorf("Expected body %s, got %s", test.expectedBody, string(body))
				}

				// Verify that the response body can still be read
				bodyBytes, err := ioutil.ReadAll(test.response.Body)
				if err != nil {
					t.Errorf("Failed to read response body: %v", err)
				}

				if string(bodyBytes) != test.expectedBody {
					t.Errorf("Expected body from response %s, got %s", test.expectedBody, string(bodyBytes))
				}
			}
		})
	}
}

func TestGetAcceptHeader(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name        string
		contentType ContentType
		expected    string
	}{
		{"JSON", TypeJSON, "application/json"},
		{"XML", TypeXML, "application/xml, text/xml"},
		{"HTML", TypeHTML, "text/html"},
		{"Text", TypeText, "text/plain"},
		{"GraphQL", TypeGraphQL, "application/json"},
		{"Unknown", TypeUnknown, "*/*"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := handler.GetAcceptHeader(test.contentType)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestSetAcceptHeader(t *testing.T) {
	handler := NewContentTypeHandler(TypeJSON)

	tests := []struct {
		name           string
		request        *ffuf.Request
		preferredType  ContentType
		expectedHeader string
	}{
		{
			"JSON",
			&ffuf.Request{Headers: map[string]string{}},
			TypeJSON,
			"application/json",
		},
		{
			"XML",
			&ffuf.Request{Headers: map[string]string{}},
			TypeXML,
			"application/xml, text/xml",
		},
		{
			"HTML",
			&ffuf.Request{Headers: map[string]string{}},
			TypeHTML,
			"text/html",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler.SetAcceptHeader(test.request, test.preferredType)

			acceptHeader, ok := test.request.Headers["Accept"]
			if !ok {
				t.Error("Accept header not set")
			} else if acceptHeader != test.expectedHeader {
				t.Errorf("Expected Accept header %s, got %s", test.expectedHeader, acceptHeader)
			}
		})
	}
}
