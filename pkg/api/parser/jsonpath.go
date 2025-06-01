// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// JSONPathParser provides methods for parsing and evaluating JSONPath expressions
type JSONPathParser struct {
	// Root data to evaluate expressions against
	data interface{}
}

// NewJSONPathParser creates a new JSONPathParser with the given JSON data
func NewJSONPathParser(jsonData []byte) (*JSONPathParser, error) {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, api.NewAPIError("Failed to parse JSON data: "+err.Error(), 0)
	}

	return &JSONPathParser{
		data: data,
	}, nil
}

// NewJSONPathParserFromObject creates a new JSONPathParser with the given parsed JSON object
func NewJSONPathParserFromObject(data interface{}) *JSONPathParser {
	return &JSONPathParser{
		data: data,
	}
}

// Evaluate evaluates a JSONPath expression against the data and returns the result
func (p *JSONPathParser) Evaluate(expression string) (interface{}, error) {
	// Handle empty expression
	if expression == "" || expression == "$" {
		return p.data, nil
	}

	// Normalize expression
	expression = normalizeExpression(expression)

	// Split the expression into segments
	segments, err := parseExpression(expression)
	if err != nil {
		return nil, err
	}

	// Evaluate the expression
	return evaluateSegments(p.data, segments)
}

// EvaluateToString evaluates a JSONPath expression and returns the result as a string
func (p *JSONPathParser) EvaluateToString(expression string) (string, error) {
	result, err := p.Evaluate(expression)
	if err != nil {
		return "", err
	}

	// Convert result to string
	switch v := result.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	case nil:
		return "null", nil
	default:
		// For complex types, convert to JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", api.NewAPIError("Failed to convert result to string: "+err.Error(), 0)
		}
		return string(jsonBytes), nil
	}
}

// EvaluateToArray evaluates a JSONPath expression and returns the result as an array
func (p *JSONPathParser) EvaluateToArray(expression string) ([]interface{}, error) {
	result, err := p.Evaluate(expression)
	if err != nil {
		return nil, err
	}

	// If result is already an array, return it
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}

	// If result is not an array, wrap it in an array
	return []interface{}{result}, nil
}

// EvaluateToMap evaluates a JSONPath expression and returns the result as a map
func (p *JSONPathParser) EvaluateToMap(expression string) (map[string]interface{}, error) {
	result, err := p.Evaluate(expression)
	if err != nil {
		return nil, err
	}

	// If result is a map, return it
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, api.NewAPIError("Result is not an object", 0)
}

// Filter filters the data using a JSONPath expression and returns matching items
func (p *JSONPathParser) Filter(expression string, filterExpr string) ([]interface{}, error) {
	// Evaluate the base expression
	baseResult, err := p.Evaluate(expression)
	if err != nil {
		return nil, err
	}

	// If base result is not an array, return empty result
	baseArray, ok := baseResult.([]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	// Parse the filter expression
	filterFunc, err := parseFilterExpression(filterExpr)
	if err != nil {
		return nil, err
	}

	// Apply the filter
	var result []interface{}
	for _, item := range baseArray {
		// Create a parser for this item
		itemParser := NewJSONPathParserFromObject(item)

		// Apply the filter
		matches, err := filterFunc(itemParser)
		if err != nil {
			return nil, err
		}

		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

// normalizeExpression normalizes a JSONPath expression
func normalizeExpression(expression string) string {
	// Ensure expression starts with $
	if !strings.HasPrefix(expression, "$") {
		expression = "$" + expression
	}

	return expression
}

// parseExpression parses a JSONPath expression into segments
func parseExpression(expression string) ([]string, error) {
	// Remove the root symbol
	expression = strings.TrimPrefix(expression, "$")

	// Handle empty expression
	if expression == "" {
		return []string{}, nil
	}

	// Split by dots, but handle brackets
	var segments []string
	var currentSegment strings.Builder
	inBracket := false

	for _, char := range expression {
		switch char {
		case '.':
			if inBracket {
				currentSegment.WriteRune(char)
			} else {
				// End of segment
				if currentSegment.Len() > 0 {
					segments = append(segments, currentSegment.String())
					currentSegment.Reset()
				}
			}
		case '[':
			if inBracket {
				return nil, api.NewAPIError("Nested brackets are not supported", 0)
			}

			// End of segment
			if currentSegment.Len() > 0 {
				segments = append(segments, currentSegment.String())
				currentSegment.Reset()
			}

			inBracket = true
			currentSegment.WriteRune(char)
		case ']':
			if !inBracket {
				return nil, api.NewAPIError("Unmatched closing bracket", 0)
			}

			currentSegment.WriteRune(char)
			segments = append(segments, currentSegment.String())
			currentSegment.Reset()
			inBracket = false
		default:
			currentSegment.WriteRune(char)
		}
	}

	// Add the last segment
	if currentSegment.Len() > 0 {
		segments = append(segments, currentSegment.String())
	}

	// Validate segments
	for _, segment := range segments {
		if strings.HasPrefix(segment, "[") && !strings.HasSuffix(segment, "]") {
			return nil, api.NewAPIError("Invalid bracket notation: "+segment, 0)
		}
	}

	return segments, nil
}

// evaluateSegments evaluates JSONPath segments against data
func evaluateSegments(data interface{}, segments []string) (interface{}, error) {
	current := data

	for _, segment := range segments {
		// Handle bracket notation
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			// Extract the index or key
			indexOrKey := segment[1 : len(segment)-1]

			// Handle array index
			if index, err := strconv.Atoi(indexOrKey); err == nil {
				// Array access
				arr, ok := current.([]interface{})
				if !ok {
					return nil, api.NewAPIError("Cannot access index on non-array", 0)
				}

				// Reject negative indices and ensure the index is within the array bounds
				if index < 0 || index >= len(arr) {
					return nil, api.NewAPIError(fmt.Sprintf("Array index out of bounds: %d", index), 0)
				}

				current = arr[index]
			} else if strings.HasPrefix(indexOrKey, "'") && strings.HasSuffix(indexOrKey, "'") {
				// Object access with quoted key
				key := indexOrKey[1 : len(indexOrKey)-1]
				obj, ok := current.(map[string]interface{})
				if !ok {
					return nil, api.NewAPIError("Cannot access property on non-object", 0)
				}

				value, exists := obj[key]
				if !exists {
					return nil, api.NewAPIError(fmt.Sprintf("Property not found: %s", key), 0)
				}

				current = value
			} else if strings.HasPrefix(indexOrKey, "\"") && strings.HasSuffix(indexOrKey, "\"") {
				// Object access with double-quoted key
				key := indexOrKey[1 : len(indexOrKey)-1]
				obj, ok := current.(map[string]interface{})
				if !ok {
					return nil, api.NewAPIError("Cannot access property on non-object", 0)
				}

				value, exists := obj[key]
				if !exists {
					return nil, api.NewAPIError(fmt.Sprintf("Property not found: %s", key), 0)
				}

				current = value
			} else if indexOrKey == "*" {
				// Wildcard - return all elements/properties
				switch v := current.(type) {
				case []interface{}:
					return v, nil
				case map[string]interface{}:
					var values []interface{}
					for _, val := range v {
						values = append(values, val)
					}
					return values, nil
				default:
					return nil, api.NewAPIError("Cannot use wildcard on non-array/object", 0)
				}
			} else {
				// Treat as object key
				obj, ok := current.(map[string]interface{})
				if !ok {
					return nil, api.NewAPIError("Cannot access property on non-object", 0)
				}

				value, exists := obj[indexOrKey]
				if !exists {
					return nil, api.NewAPIError(fmt.Sprintf("Property not found: %s", indexOrKey), 0)
				}

				current = value
			}
		} else {
			// Dot notation
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, api.NewAPIError("Cannot access property on non-object", 0)
			}

			value, exists := obj[segment]
			if !exists {
				return nil, api.NewAPIError(fmt.Sprintf("Property not found: %s", segment), 0)
			}

			current = value
		}
	}

	return current, nil
}

// FilterFunc is a function that evaluates whether an item matches a filter
type FilterFunc func(parser *JSONPathParser) (bool, error)

// parseFilterExpression parses a filter expression and returns a FilterFunc
func parseFilterExpression(expression string) (FilterFunc, error) {
	// Simple equality filter: @.property == value
	eqRegex := regexp.MustCompile(`@\.([a-zA-Z0-9_]+)\s*==\s*(.+)`)
	if matches := eqRegex.FindStringSubmatch(expression); len(matches) == 3 {
		property := matches[1]
		valueStr := strings.TrimSpace(matches[2])

		// Parse the value
		var expectedValue interface{}

		// String value
		if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
			(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
			// Remove quotes
			expectedValue = valueStr[1 : len(valueStr)-1]
		} else if valueStr == "true" {
			expectedValue = true
		} else if valueStr == "false" {
			expectedValue = false
		} else if valueStr == "null" {
			expectedValue = nil
		} else if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
			expectedValue = num
		} else {
			return nil, api.NewAPIError("Invalid value in filter expression: "+valueStr, 0)
		}

		// Return the filter function
		return func(parser *JSONPathParser) (bool, error) {
			// Evaluate the property - convert @.property to just property
			propertyPath := property
			actualValue, err := parser.Evaluate(propertyPath)
			if err != nil {
				// Property not found, doesn't match
				return false, nil
			}

			// Compare values
			return compareValues(actualValue, expectedValue), nil
		}, nil
	}

	// Contains filter: @.property contains value
	containsRegex := regexp.MustCompile(`@\.([a-zA-Z0-9_]+)\s+contains\s+(.+)`)
	if matches := containsRegex.FindStringSubmatch(expression); len(matches) == 3 {
		property := matches[1]
		valueStr := strings.TrimSpace(matches[2])

		// Parse the value (must be a string for contains)
		var expectedValue string
		if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
			(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
			// Remove quotes
			expectedValue = valueStr[1 : len(valueStr)-1]
		} else {
			return nil, api.NewAPIError("Value for 'contains' must be a string: "+valueStr, 0)
		}

		// Return the filter function
		return func(parser *JSONPathParser) (bool, error) {
			// Evaluate the property - convert @.property to just property
			propertyPath := property
			actualValue, err := parser.Evaluate(propertyPath)
			if err != nil {
				// Property not found, doesn't match
				return false, nil
			}

			// Check if the property contains the value
			switch v := actualValue.(type) {
			case string:
				return strings.Contains(v, expectedValue), nil
			case []interface{}:
				// Check if array contains the value
				for _, item := range v {
					if str, ok := item.(string); ok && str == expectedValue {
						return true, nil
					}
				}
				return false, nil
			default:
				return false, nil
			}
		}, nil
	}

	// Greater than filter: @.property > value
	gtRegex := regexp.MustCompile(`@\.([a-zA-Z0-9_]+)\s*>\s*(.+)`)
	if matches := gtRegex.FindStringSubmatch(expression); len(matches) == 3 {
		property := matches[1]
		valueStr := strings.TrimSpace(matches[2])

		// Parse the value (must be a number for >)
		expectedValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, api.NewAPIError("Value for '>' must be a number: "+valueStr, 0)
		}

		// Return the filter function
		return func(parser *JSONPathParser) (bool, error) {
			// Evaluate the property - convert @.property to just property
			propertyPath := property
			actualValue, err := parser.Evaluate(propertyPath)
			if err != nil {
				// Property not found, doesn't match
				return false, nil
			}

			// Check if the property is greater than the value
			switch v := actualValue.(type) {
			case float64:
				return v > expectedValue, nil
			case int:
				return float64(v) > expectedValue, nil
			default:
				return false, nil
			}
		}, nil
	}

	// Less than filter: @.property < value
	ltRegex := regexp.MustCompile(`@\.([a-zA-Z0-9_]+)\s*<\s*(.+)`)
	if matches := ltRegex.FindStringSubmatch(expression); len(matches) == 3 {
		property := matches[1]
		valueStr := strings.TrimSpace(matches[2])

		// Parse the value (must be a number for <)
		expectedValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, api.NewAPIError("Value for '<' must be a number: "+valueStr, 0)
		}

		// Return the filter function
		return func(parser *JSONPathParser) (bool, error) {
			// Evaluate the property - convert @.property to just property
			propertyPath := property
			actualValue, err := parser.Evaluate(propertyPath)
			if err != nil {
				// Property not found, doesn't match
				return false, nil
			}

			// Check if the property is less than the value
			switch v := actualValue.(type) {
			case float64:
				return v < expectedValue, nil
			case int:
				return float64(v) < expectedValue, nil
			default:
				return false, nil
			}
		}, nil
	}

	return nil, api.NewAPIError("Unsupported filter expression: "+expression, 0)
}

// compareValues compares two values for equality
func compareValues(a, b interface{}) bool {
	// Handle nil values
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare based on type
	switch aVal := a.(type) {
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case float64:
		switch bVal := b.(type) {
		case float64:
			return aVal == bVal
		case int:
			return aVal == float64(bVal)
		}
	case int:
		switch bVal := b.(type) {
		case float64:
			return float64(aVal) == bVal
		case int:
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	case []interface{}:
		if bVal, ok := b.([]interface{}); ok {
			// Compare arrays
			if len(aVal) != len(bVal) {
				return false
			}
			for i := range aVal {
				if !compareValues(aVal[i], bVal[i]) {
					return false
				}
			}
			return true
		}
	case map[string]interface{}:
		if bVal, ok := b.(map[string]interface{}); ok {
			// Compare objects
			if len(aVal) != len(bVal) {
				return false
			}
			for k, v := range aVal {
				bv, exists := bVal[k]
				if !exists || !compareValues(v, bv) {
					return false
				}
			}
			return true
		}
	}

	return false
}

// Add JSONPath methods to ResponseParser

// ParseJSONWithPath parses a JSON response and evaluates a JSONPath expression
func (p *ResponseParser) ParseJSONWithPath(data []byte, path string) (interface{}, error) {
	if p.format != FormatJSON && p.format != FormatUnknown {
		return nil, api.NewAPIError("Response is not in JSON format", 0)
	}

	// Parse the JSON
	parser, err := NewJSONPathParser(data)
	if err != nil {
		return nil, err
	}

	// Evaluate the JSONPath expression
	return parser.Evaluate(path)
}

// FilterJSON filters a JSON response using a JSONPath expression and filter
func (p *ResponseParser) FilterJSON(data []byte, path string, filter string) ([]interface{}, error) {
	if p.format != FormatJSON && p.format != FormatUnknown {
		return nil, api.NewAPIError("Response is not in JSON format", 0)
	}

	// Parse the JSON
	parser, err := NewJSONPathParser(data)
	if err != nil {
		return nil, err
	}

	// Filter the data
	return parser.Filter(path, filter)
}
