// Package workflow provides a concurrency model for complex API workflows.
package workflow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSONPath represents a JSONPath expression.
type JSONPath struct {
	// Path is the parsed path components
	Path []PathComponent
}

// PathComponent represents a component of a JSONPath.
type PathComponent interface {
	// Evaluate evaluates the component against the given data
	Evaluate(data interface{}) (interface{}, error)
}

// RootComponent represents the root of a JSONPath.
type RootComponent struct{}

// Evaluate implements PathComponent.
func (c *RootComponent) Evaluate(data interface{}) (interface{}, error) {
	return data, nil
}

// PropertyComponent represents a property access in a JSONPath.
type PropertyComponent struct {
	// Name is the name of the property
	Name string
}

// Evaluate implements PathComponent.
func (c *PropertyComponent) Evaluate(data interface{}) (interface{}, error) {
	// Handle different data types
	switch d := data.(type) {
	case map[string]interface{}:
		// Check if the property exists
		if value, ok := d[c.Name]; ok {
			return value, nil
		}
		return nil, fmt.Errorf("property %s not found", c.Name)
	default:
		return nil, fmt.Errorf("cannot access property %s on non-object value", c.Name)
	}
}

// IndexComponent represents an array index access in a JSONPath.
type IndexComponent struct {
	// Index is the index to access
	Index int
}

// Evaluate implements PathComponent.
func (c *IndexComponent) Evaluate(data interface{}) (interface{}, error) {
	// Handle different data types
	switch d := data.(type) {
	case []interface{}:
		// Check if the index is valid
		if c.Index >= 0 && c.Index < len(d) {
			return d[c.Index], nil
		}
		return nil, fmt.Errorf("index %d out of bounds", c.Index)
	default:
		return nil, fmt.Errorf("cannot access index %d on non-array value", c.Index)
	}
}

// ParseJSONPath parses a JSONPath expression.
func ParseJSONPath(path string) (*JSONPath, error) {
	// Remove leading $ if present
	if strings.HasPrefix(path, "$") {
		path = path[1:]
	}

	// Split the path into components
	components := []PathComponent{&RootComponent{}}
	
	// If path is empty, return just the root component
	if path == "" {
		return &JSONPath{Path: components}, nil
	}

	// Split by dots and brackets
	parts := strings.Split(path, ".")
	for _, part := range parts {
		// Skip empty parts
		if part == "" {
			continue
		}

		// Handle array indices
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Split property name and indices
			propEnd := strings.Index(part, "[")
			propName := part[:propEnd]
			
			// Add property component if there's a property name
			if propName != "" {
				components = append(components, &PropertyComponent{Name: propName})
			}
			
			// Process indices
			indices := part[propEnd:]
			for indices != "" {
				// Extract index
				start := strings.Index(indices, "[")
				end := strings.Index(indices, "]")
				if start == -1 || end == -1 || end <= start {
					return nil, fmt.Errorf("invalid array index syntax: %s", indices)
				}
				
				// Parse index
				indexStr := indices[start+1:end]
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return nil, fmt.Errorf("invalid array index: %s", indexStr)
				}
				
				// Add index component
				components = append(components, &IndexComponent{Index: index})
				
				// Move to next index
				indices = indices[end+1:]
			}
		} else {
			// Simple property access
			components = append(components, &PropertyComponent{Name: part})
		}
	}

	return &JSONPath{Path: components}, nil
}

// EvaluateJSONPath evaluates a JSONPath expression against JSON data.
func EvaluateJSONPath(jsonData []byte, path string) (string, error) {
	// Parse the JSON data
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Parse the JSONPath
	jsonPath, err := ParseJSONPath(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSONPath: %v", err)
	}

	// Evaluate the JSONPath
	result := data
	for _, component := range jsonPath.Path {
		result, err = component.Evaluate(result)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate JSONPath: %v", err)
		}
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
		return "", nil
	default:
		// For complex types, convert back to JSON
		resultJSON, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to convert result to string: %v", err)
		}
		return string(resultJSON), nil
	}
}