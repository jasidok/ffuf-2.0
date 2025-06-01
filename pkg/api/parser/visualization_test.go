package parser

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ffuf/ffuf/v2/pkg/ffuf"
)

func TestVisualizerCreation(t *testing.T) {
	// Test with nil options
	v := NewVisualizer(nil)
	if v == nil {
		t.Error("Failed to create visualizer with nil options")
	}
	if v.options == nil {
		t.Error("Visualizer options should not be nil when created with nil options")
	}
	if v.options.Format != VisFormatHTML {
		t.Errorf("Expected default format to be HTML, got %s", v.options.Format)
	}

	// Test with custom options
	options := &VisOptions{
		Format:        VisFormatJSON,
		Type:          VisTypeGraph,
		MaxDepth:      5,
		IncludeValues: false,
		ColorScheme:   "custom",
		Title:         "Custom Title",
	}
	v = NewVisualizer(options)
	if v == nil {
		t.Error("Failed to create visualizer with custom options")
	}
	if v.options.Format != VisFormatJSON {
		t.Errorf("Expected format to be JSON, got %s", v.options.Format)
	}
	if v.options.Type != VisTypeGraph {
		t.Errorf("Expected type to be graph, got %s", v.options.Type)
	}
	if v.options.MaxDepth != 5 {
		t.Errorf("Expected max depth to be 5, got %d", v.options.MaxDepth)
	}
	if v.options.IncludeValues != false {
		t.Errorf("Expected include values to be false, got %t", v.options.IncludeValues)
	}
	if v.options.ColorScheme != "custom" {
		t.Errorf("Expected color scheme to be custom, got %s", v.options.ColorScheme)
	}
	if v.options.Title != "Custom Title" {
		t.Errorf("Expected title to be Custom Title, got %s", v.options.Title)
	}
}

func TestVisualizeResponse(t *testing.T) {
	// Create a test response
	jsonData := `{"name":"test","value":123,"nested":{"key":"value"},"array":[1,2,3]}`
	resp := &ffuf.Response{
		ContentType: "application/json",
		Data:        []byte(jsonData),
		Request: &ffuf.Request{
			Url: "https://example.com/api",
		},
	}

	// Test JSON visualization
	v := NewVisualizer(&VisOptions{Format: VisFormatJSON})
	result, err := v.VisualizeResponse(resp)
	if err != nil {
		t.Errorf("Failed to visualize response as JSON: %v", err)
	}
	// Verify the result is valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Test HTML visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatHTML, MaxDepth: 2})
	result, err = v.VisualizeResponse(resp)
	if err != nil {
		t.Errorf("Failed to visualize response as HTML: %v", err)
	}
	// Verify the result contains HTML elements
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("HTML visualization does not contain DOCTYPE")
	}
	if !strings.Contains(result, "<html>") {
		t.Error("HTML visualization does not contain html tag")
	}
	if !strings.Contains(result, "name") || !strings.Contains(result, "test") {
		t.Errorf("HTML visualization does not contain expected content: %s", result)
	}

	// Test DOT visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatDOT})
	result, err = v.VisualizeResponse(resp)
	if err != nil {
		t.Errorf("Failed to visualize response as DOT: %v", err)
	}
	// Verify the result contains DOT elements
	if !strings.Contains(result, "digraph") {
		t.Error("DOT visualization does not contain digraph")
	}
	if !strings.Contains(result, "->") {
		t.Error("DOT visualization does not contain edges")
	}

	// Test Mermaid visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatMermaid})
	result, err = v.VisualizeResponse(resp)
	if err != nil {
		t.Errorf("Failed to visualize response as Mermaid: %v", err)
	}
	// Verify the result contains Mermaid elements
	if !strings.Contains(result, "graph TD") {
		t.Error("Mermaid visualization does not contain graph TD")
	}
	if !strings.Contains(result, "-->") {
		t.Error("Mermaid visualization does not contain edges")
	}

	// Test with non-JSON response
	resp.ContentType = "text/plain"
	v = NewVisualizer(nil)
	_, err = v.VisualizeResponse(resp)
	if err == nil {
		t.Error("Expected error for non-JSON response, got nil")
	}

	// Test with invalid JSON
	resp.ContentType = "application/json"
	resp.Data = []byte("{invalid json}")
	_, err = v.VisualizeResponse(resp)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestVisualizeSchema(t *testing.T) {
	// Create a test schema
	schema := &Schema{
		Properties: SchemaMap{
			"name": SchemaField{
				Type:     TypeString,
				Format:   "text",
				Required: true,
			},
			"age": SchemaField{
				Type: TypeNumber,
			},
			"address": SchemaField{
				Type: TypeObject,
				Properties: SchemaMap{
					"street": SchemaField{
						Type: TypeString,
					},
					"city": SchemaField{
						Type: TypeString,
					},
				},
			},
		},
	}

	// Test JSON visualization
	v := NewVisualizer(&VisOptions{Format: VisFormatJSON})
	result, err := v.VisualizeSchema(schema)
	if err != nil {
		t.Errorf("Failed to visualize schema as JSON: %v", err)
	}
	// Verify the result is valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	// Test HTML visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatHTML})
	result, err = v.VisualizeSchema(schema)
	if err != nil {
		t.Errorf("Failed to visualize schema as HTML: %v", err)
	}
	// Verify the result contains HTML elements
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("HTML visualization does not contain DOCTYPE")
	}
	if !strings.Contains(result, "<html>") {
		t.Error("HTML visualization does not contain html tag")
	}
	if !strings.Contains(result, "name") || !strings.Contains(result, "string") {
		t.Error("HTML visualization does not contain expected content")
	}

	// Test DOT visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatDOT})
	result, err = v.VisualizeSchema(schema)
	if err != nil {
		t.Errorf("Failed to visualize schema as DOT: %v", err)
	}
	// Verify the result contains DOT elements
	if !strings.Contains(result, "digraph") {
		t.Error("DOT visualization does not contain digraph")
	}
	if !strings.Contains(result, "->") {
		t.Error("DOT visualization does not contain edges")
	}

	// Test Mermaid visualization
	v = NewVisualizer(&VisOptions{Format: VisFormatMermaid})
	result, err = v.VisualizeSchema(schema)
	if err != nil {
		t.Errorf("Failed to visualize schema as Mermaid: %v", err)
	}
	// Verify the result contains Mermaid elements
	if !strings.Contains(result, "classDiagram") {
		t.Error("Mermaid visualization does not contain classDiagram")
	}
	if !strings.Contains(result, "class Schema") {
		t.Error("Mermaid visualization does not contain class Schema")
	}
}

func TestTreeConversion(t *testing.T) {
	// Create a test data structure
	data := map[string]interface{}{
		"string": "value",
		"number": 123.45,
		"bool":   true,
		"null":   nil,
		"object": map[string]interface{}{
			"nested": "value",
		},
		"array": []interface{}{1, "two", true},
	}

	// Test tree conversion
	v := NewVisualizer(nil)
	tree := v.convertToTree(data, 10)

	// Verify the tree structure
	if tree.Type != "object" {
		t.Errorf("Expected root type to be object, got %s", tree.Type)
	}
	if len(tree.Children) != 6 {
		t.Errorf("Expected 6 children, got %d", len(tree.Children))
	}

	// Check for specific children
	var stringNode, arrayNode, objectNode *TreeNode
	for _, child := range tree.Children {
		switch child.Key {
		case "string":
			stringNode = child
		case "array":
			arrayNode = child
		case "object":
			objectNode = child
		}
	}

	if stringNode == nil || stringNode.Type != "string" || stringNode.Value != "value" {
		t.Error("String node not found or has incorrect properties")
	}

	if arrayNode == nil || arrayNode.Type != "array" || len(arrayNode.Children) != 3 {
		t.Error("Array node not found or has incorrect properties")
	}

	if objectNode == nil || objectNode.Type != "object" || len(objectNode.Children) != 1 {
		t.Error("Object node not found or has incorrect properties")
	}

	// Test max depth limitation
	tree = v.convertToTree(data, 1)
	for _, child := range tree.Children {
		if child.Key == "object" && len(child.Children) > 0 {
			t.Error("Max depth not respected for object")
		}
		if child.Key == "array" && len(child.Children) > 0 {
			t.Error("Max depth not respected for array")
		}
	}
}

func TestSanitizeFunctions(t *testing.T) {
	// Test sanitizeID
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with space", "with_space"},
		{"with-dash", "with_dash"},
		{"with.dot", "with_dot"},
		{"with/slash", "with_slash"},
		{"with\\backslash", "with_backslash"},
		{"with\"quote", "with_quote"},
	}

	for _, tc := range testCases {
		result := sanitizeID(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeID(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}

	// Test sanitizeLabel
	testCases = []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with space", "with space"},
		{"with\"quote", "with\\\"quote"},
		{"with\\backslash", "with\\\\backslash"},
	}

	for _, tc := range testCases {
		result := sanitizeLabel(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeLabel(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}
