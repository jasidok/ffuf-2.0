// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

// ParameterType represents the type of parameter
type ParameterType string

const (
	// ParamTypeQuery represents a query parameter
	ParamTypeQuery ParameterType = "query"
	// ParamTypePath represents a path parameter
	ParamTypePath ParameterType = "path"
	// ParamTypeBody represents a body parameter
	ParamTypeBody ParameterType = "body"
	// ParamTypeHeader represents a header parameter
	ParamTypeHeader ParameterType = "header"
	// ParamTypeUnknown represents an unknown parameter type
	ParamTypeUnknown ParameterType = "unknown"
)

// Parameter represents a discovered API parameter
type Parameter struct {
	// Name is the name of the parameter
	Name string `json:"name"`
	// Type is the type of parameter (query, path, body, header)
	Type ParameterType `json:"type"`
	// DataType is the data type of the parameter (string, number, boolean, etc.)
	DataType string `json:"data_type"`
	// Required indicates if the parameter is required
	Required bool `json:"required"`
	// Description is a description of the parameter
	Description string `json:"description,omitempty"`
	// Example is an example value for the parameter
	Example string `json:"example,omitempty"`
	// Path is the JSONPath where the parameter was found (for body parameters)
	Path string `json:"path,omitempty"`
	// Confidence is a score from 0-100 indicating confidence in the detection
	Confidence int `json:"confidence"`
}

// ParameterDiscovery provides methods for discovering API parameters from responses
type ParameterDiscovery struct {
	// Common patterns for different parameter types
	patterns map[string]*regexp.Regexp
	// Known parameter names and their types
	knownParams map[string]ParameterType
}

// NewParameterDiscovery creates a new ParameterDiscovery
func NewParameterDiscovery() *ParameterDiscovery {
	discovery := &ParameterDiscovery{
		patterns:    make(map[string]*regexp.Regexp),
		knownParams: make(map[string]ParameterType),
	}

	// Initialize patterns for different parameter types
	discovery.patterns["pathParam"] = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
	discovery.patterns["queryParam"] = regexp.MustCompile(`[\?&]([a-zA-Z0-9_]+)=`)
	discovery.patterns["jsonField"] = regexp.MustCompile(`"([a-zA-Z0-9_]+)":\s*`)

	// Initialize known parameter names and their types
	discovery.knownParams["id"] = ParamTypePath
	discovery.knownParams["uuid"] = ParamTypePath
	discovery.knownParams["page"] = ParamTypeQuery
	discovery.knownParams["limit"] = ParamTypeQuery
	discovery.knownParams["offset"] = ParamTypeQuery
	discovery.knownParams["sort"] = ParamTypeQuery
	discovery.knownParams["order"] = ParamTypeQuery
	discovery.knownParams["filter"] = ParamTypeQuery
	discovery.knownParams["q"] = ParamTypeQuery
	discovery.knownParams["query"] = ParamTypeQuery
	discovery.knownParams["search"] = ParamTypeQuery
	discovery.knownParams["token"] = ParamTypeHeader
	discovery.knownParams["api_key"] = ParamTypeHeader
	discovery.knownParams["apikey"] = ParamTypeHeader
	discovery.knownParams["authorization"] = ParamTypeHeader
	discovery.knownParams["content-type"] = ParamTypeHeader
	discovery.knownParams["accept"] = ParamTypeHeader

	return discovery
}

// DiscoverParameters discovers API parameters from a response
func (d *ParameterDiscovery) DiscoverParameters(resp *ffuf.Response) ([]Parameter, error) {
	params := make([]Parameter, 0)

	// Discover parameters from URL
	urlParams, err := d.discoverURLParameters(resp.Request.Url)
	if err != nil {
		return nil, err
	}
	params = append(params, urlParams...)

	// Discover parameters from response links
	linkParams, err := d.discoverLinksParameters(resp)
	if err != nil {
		return nil, err
	}
	params = append(params, linkParams...)

	// Discover parameters from JSON response body
	if strings.Contains(resp.ContentType, "application/json") {
		bodyParams, err := d.discoverJSONParameters(resp.Data)
		if err != nil {
			return nil, err
		}
		params = append(params, bodyParams...)
	}

	// Deduplicate parameters
	return d.deduplicateParameters(params), nil
}

// discoverURLParameters discovers parameters from a URL
func (d *ParameterDiscovery) discoverURLParameters(urlStr string) ([]Parameter, error) {
	params := make([]Parameter, 0)

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, api.NewAPIError("Failed to parse URL: "+err.Error(), 0)
	}

	// Extract path parameters
	pathParams := d.patterns["pathParam"].FindAllStringSubmatch(parsedURL.Path, -1)
	for _, match := range pathParams {
		if len(match) > 1 {
			paramName := match[1]
			params = append(params, Parameter{
				Name:       paramName,
				Type:       ParamTypePath,
				DataType:   "string",
				Required:   true,
				Confidence: 80,
			})
		}
	}

	// Extract query parameters
	queryParams := parsedURL.Query()
	for key := range queryParams {
		params = append(params, Parameter{
			Name:       key,
			Type:       ParamTypeQuery,
			DataType:   "string",
			Required:   false,
			Example:    queryParams.Get(key),
			Confidence: 90,
		})
	}

	return params, nil
}

// discoverLinksParameters discovers parameters from links in the response
func (d *ParameterDiscovery) discoverLinksParameters(resp *ffuf.Response) ([]Parameter, error) {
	params := make([]Parameter, 0)

	// Check for Link header
	if linkHeaders, ok := resp.Headers["Link"]; ok {
		for _, link := range linkHeaders {
			// Extract URL from link
			urlMatch := regexp.MustCompile(`<([^>]+)>`).FindStringSubmatch(link)
			if len(urlMatch) > 1 {
				linkURL := urlMatch[1]
				linkParams, err := d.discoverURLParameters(linkURL)
				if err != nil {
					continue
				}
				params = append(params, linkParams...)
			}
		}
	}

	// Check for links in JSON response
	if strings.Contains(resp.ContentType, "application/json") {
		var jsonData interface{}
		if err := json.Unmarshal(resp.Data, &jsonData); err == nil {
			// Extract URLs from JSON data
			urls := d.extractURLsFromJSON(jsonData)
			for _, urlStr := range urls {
				linkParams, err := d.discoverURLParameters(urlStr)
				if err != nil {
					continue
				}
				params = append(params, linkParams...)
			}
		}
	}

	return params, nil
}

// discoverJSONParameters discovers parameters from a JSON response body
func (d *ParameterDiscovery) discoverJSONParameters(data []byte) ([]Parameter, error) {
	params := make([]Parameter, 0)

	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, api.NewAPIError("Failed to parse JSON: "+err.Error(), 0)
	}

	// Use JSONPath parser to navigate the JSON structure
	parser := NewJSONPathParserFromObject(jsonData)

	// Extract parameters from the JSON data
	d.extractParametersFromJSON(jsonData, "$", parser, &params, 0, 10) // Add depth limit

	return params, nil
}

// extractParametersFromJSON recursively extracts parameters from JSON data
func (d *ParameterDiscovery) extractParametersFromJSON(data interface{}, path string, parser *JSONPathParser, params *[]Parameter, depth int, maxDepth int) {
	// Stop recursion when reaching the depth limit
	if depth > maxDepth {
		return
	}

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

			// Determine parameter type and data type
			paramType := ParamTypeBody
			dataType := "string"
			required := false
			confidence := 70

			// Check if key suggests this might be a specific parameter type
			if knownType, ok := d.knownParams[strings.ToLower(key)]; ok {
				paramType = knownType
				confidence = 85
			}

			// Determine data type
			switch value.(type) {
			case string:
				dataType = "string"
			case float64:
				dataType = "number"
			case bool:
				dataType = "boolean"
			case nil:
				dataType = "null"
			case []interface{}:
				dataType = "array"
			case map[string]interface{}:
				dataType = "object"
			}

			// Add parameter
			*params = append(*params, Parameter{
				Name:       key,
				Type:       paramType,
				DataType:   dataType,
				Required:   required,
				Path:       newPath,
				Example:    d.formatExampleValue(value),
				Confidence: confidence,
			})

			// Recursively process the value
			d.extractParametersFromJSON(value, newPath, parser, params, depth+1, maxDepth)
		}
	case []interface{}:
		// Process each element in the array
		for i, elem := range v {
			newPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			d.extractParametersFromJSON(elem, newPath, parser, params, depth+1, maxDepth)
		}
	}
}

// extractURLsFromJSON extracts URLs from JSON data
func (d *ParameterDiscovery) extractURLsFromJSON(data interface{}) []string {
	const maxURLs = 1000             // Limit to prevent memory issues
	urlsMap := make(map[string]bool) // Use a map to avoid duplicates
	urlPattern := regexp.MustCompile(`^(https?://|/)[\w\-\./%&=\?]+$`)

	d.extractURLsRecursive(data, urlPattern, urlsMap, 0, 10) // Add depth limit of 10

	// Convert map to slice
	urls := make([]string, 0, len(urlsMap))
	count := 0
	for url := range urlsMap {
		if count >= maxURLs {
			break
		}
		urls = append(urls, url)
		count++
	}

	return urls
}

// extractURLsRecursive recursively extracts URLs from JSON data
func (d *ParameterDiscovery) extractURLsRecursive(data interface{}, urlPattern *regexp.Regexp, urlsMap map[string]bool, depth int, maxDepth int) {
	// Stop recursion when reaching the depth limit
	if depth > maxDepth {
		return
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Process each key-value pair
		for key, value := range v {
			// Check if key suggests this might be a URL
			if strings.Contains(strings.ToLower(key), "url") ||
				strings.Contains(strings.ToLower(key), "link") ||
				strings.Contains(strings.ToLower(key), "href") ||
				key == "self" || key == "next" || key == "prev" {
				if strValue, ok := value.(string); ok {
					if urlPattern.MatchString(strValue) {
						urlsMap[strValue] = true
					}
				}
			}

			// Recursively process the value with increased depth
			d.extractURLsRecursive(value, urlPattern, urlsMap, depth+1, maxDepth)
		}
	case []interface{}:
		// Process each element in the array
		for _, elem := range v {
			d.extractURLsRecursive(elem, urlPattern, urlsMap, depth+1, maxDepth)
		}
	case string:
		// Only check strings that are directly in URL-related keys
		// Don't check all string values as that would be too broad
		// This case is handled by the map processing above
	}
}

// formatExampleValue formats a value for use as an example
func (d *ParameterDiscovery) formatExampleValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 50 {
			return v[:47] + "..."
		}
		return v
	case nil:
		return "null"
	default:
		// Convert to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		jsonStr := string(jsonBytes)
		if len(jsonStr) > 50 {
			return jsonStr[:47] + "..."
		}
		return jsonStr
	}
}

// deduplicateParameters removes duplicate parameters
func (d *ParameterDiscovery) deduplicateParameters(params []Parameter) []Parameter {
	uniqueParams := make([]Parameter, 0)
	seen := make(map[string]bool)

	for _, param := range params {
		key := param.Name + string(param.Type)
		if !seen[key] {
			seen[key] = true
			uniqueParams = append(uniqueParams, param)
		}
	}

	return uniqueParams
}

// DiscoverParameters is a convenience method on ResponseParser to discover parameters
func (p *ResponseParser) DiscoverParameters(resp *ffuf.Response) ([]Parameter, error) {
	discovery := NewParameterDiscovery()
	return discovery.DiscoverParameters(resp)
}
