package parser

import (
	"testing"
)

func TestJSONPathParser(t *testing.T) {
	// Test JSON data
	jsonData := []byte(`{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95
				},
				{
					"category": "fiction",
					"author": "Evelyn Waugh",
					"title": "Sword of Honour",
					"price": 12.99
				},
				{
					"category": "fiction",
					"author": "Herman Melville",
					"title": "Moby Dick",
					"isbn": "0-553-21311-3",
					"price": 8.99
				},
				{
					"category": "fiction",
					"author": "J. R. R. Tolkien",
					"title": "The Lord of the Rings",
					"isbn": "0-395-19395-8",
					"price": 22.99
				}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		},
		"expensive": 10
	}`)

	// Create a parser
	parser, err := NewJSONPathParser(jsonData)
	if err != nil {
		t.Fatalf("Failed to create JSONPathParser: %v", err)
	}

	// Test cases for Evaluate
	evaluateTests := []struct {
		name       string
		expression string
		expected   interface{}
		expectErr  bool
	}{
		{
			name:       "Root",
			expression: "$",
			expected:   parser.data,
			expectErr:  false,
		},
		{
			name:       "Simple property",
			expression: "$.expensive",
			expected:   float64(10),
			expectErr:  false,
		},
		{
			name:       "Nested property",
			expression: "$.store.bicycle.color",
			expected:   "red",
			expectErr:  false,
		},
		{
			name:       "Array element",
			expression: "$.store.book[0].title",
			expected:   "Sayings of the Century",
			expectErr:  false,
		},
		{
			name:       "Array element with bracket notation",
			expression: "$['store']['book'][0]['title']",
			expected:   "Sayings of the Century",
			expectErr:  false,
		},
		{
			name:       "Array element with mixed notation",
			expression: "$.store.book[0]['title']",
			expected:   "Sayings of the Century",
			expectErr:  false,
		},
		{
			name:       "Non-existent property",
			expression: "$.nonexistent",
			expected:   nil,
			expectErr:  true,
		},
		{
			name:       "Invalid array index",
			expression: "$.store.book[10]",
			expected:   nil,
			expectErr:  true,
		},
		{
			name:       "Invalid expression",
			expression: "$.store.book[a]",
			expected:   nil,
			expectErr:  true,
		},
	}

	for _, test := range evaluateTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parser.Evaluate(test.expression)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// For the root test, we can't directly compare the objects
					if test.name == "Root" {
						// Just check that we got a non-nil result
						if result == nil {
							t.Error("Expected non-nil result for root expression")
						}
					} else {
						// For other tests, compare the results
						if !compareValues(result, test.expected) {
							t.Errorf("Expected %v, got %v", test.expected, result)
						}
					}
				}
			}
		})
	}

	// Test EvaluateToString
	stringTests := []struct {
		name       string
		expression string
		expected   string
		expectErr  bool
	}{
		{
			name:       "String property",
			expression: "$.store.bicycle.color",
			expected:   "red",
			expectErr:  false,
		},
		{
			name:       "Number property",
			expression: "$.expensive",
			expected:   "10",
			expectErr:  false,
		},
		{
			name:       "Object property",
			expression: "$.store.bicycle",
			expected:   `{"color":"red","price":19.95}`,
			expectErr:  false,
		},
		{
			name:       "Array property",
			expression: "$.store.book[0]",
			expected:   `{"author":"Nigel Rees","category":"reference","price":8.95,"title":"Sayings of the Century"}`,
			expectErr:  false,
		},
	}

	for _, test := range stringTests {
		t.Run("ToString_"+test.name, func(t *testing.T) {
			result, err := parser.EvaluateToString(test.expression)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if result != test.expected {
					t.Errorf("Expected %s, got %s", test.expected, result)
				}
			}
		})
	}

	// Test EvaluateToArray
	arrayTests := []struct {
		name       string
		expression string
		length     int
		expectErr  bool
	}{
		{
			name:       "Array property",
			expression: "$.store.book",
			length:     4,
			expectErr:  false,
		},
		{
			name:       "Non-array property wrapped in array",
			expression: "$.expensive",
			length:     1,
			expectErr:  false,
		},
	}

	for _, test := range arrayTests {
		t.Run("ToArray_"+test.name, func(t *testing.T) {
			result, err := parser.EvaluateToArray(test.expression)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(result) != test.length {
					t.Errorf("Expected array of length %d, got %d", test.length, len(result))
				}
			}
		})
	}

	// Test EvaluateToMap
	mapTests := []struct {
		name       string
		expression string
		keys       []string
		expectErr  bool
	}{
		{
			name:       "Object property",
			expression: "$.store.bicycle",
			keys:       []string{"color", "price"},
			expectErr:  false,
		},
		{
			name:       "Non-object property",
			expression: "$.expensive",
			keys:       nil,
			expectErr:  true,
		},
	}

	for _, test := range mapTests {
		t.Run("ToMap_"+test.name, func(t *testing.T) {
			result, err := parser.EvaluateToMap(test.expression)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// Check that all expected keys are present
					for _, key := range test.keys {
						if _, exists := result[key]; !exists {
							t.Errorf("Expected key %s not found in result", key)
						}
					}
				}
			}
		})
	}

	// Test Filter
	filterTests := []struct {
		name       string
		expression string
		filter     string
		length     int
		expectErr  bool
	}{
		{
			name:       "Equality filter",
			expression: "$.store.book",
			filter:     "@.category == 'fiction'",
			length:     3,
			expectErr:  false,
		},
		{
			name:       "Contains filter",
			expression: "$.store.book",
			filter:     "@.author contains 'Tolkien'",
			length:     1,
			expectErr:  false,
		},
		{
			name:       "Greater than filter",
			expression: "$.store.book",
			filter:     "@.price > 10",
			length:     2,
			expectErr:  false,
		},
		{
			name:       "Less than filter",
			expression: "$.store.book",
			filter:     "@.price < 9",
			length:     2,
			expectErr:  false,
		},
		{
			name:       "Invalid filter",
			expression: "$.store.book",
			filter:     "@.price invalid 10",
			length:     0,
			expectErr:  true,
		},
	}

	for _, test := range filterTests {
		t.Run("Filter_"+test.name, func(t *testing.T) {
			result, err := parser.Filter(test.expression, test.filter)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(result) != test.length {
					t.Errorf("Expected %d items, got %d", test.length, len(result))
				}
			}
		})
	}
}

func TestResponseParserJSONPath(t *testing.T) {
	// Test JSON data
	jsonData := []byte(`{
		"users": [
			{
				"id": 1,
				"name": "John Doe",
				"email": "john@example.com",
				"active": true
			},
			{
				"id": 2,
				"name": "Jane Smith",
				"email": "jane@example.com",
				"active": false
			},
			{
				"id": 3,
				"name": "Bob Johnson",
				"email": "bob@example.com",
				"active": true
			}
		],
		"total": 3,
		"page": 1
	}`)

	// Create a response parser
	parser := NewResponseParser("application/json")

	// Test ParseJSONWithPath
	pathTests := []struct {
		name       string
		path       string
		checkFunc  func(interface{}) bool
		expectErr  bool
	}{
		{
			name:      "Get total",
			path:      "$.total",
			checkFunc: func(v interface{}) bool { return v == float64(3) },
			expectErr: false,
		},
		{
			name:      "Get first user",
			path:      "$.users[0]",
			checkFunc: func(v interface{}) bool {
				if m, ok := v.(map[string]interface{}); ok {
					return m["name"] == "John Doe"
				}
				return false
			},
			expectErr: false,
		},
		{
			name:      "Get all active users",
			path:      "$.users[*]",
			checkFunc: func(v interface{}) bool {
				if arr, ok := v.([]interface{}); ok {
					return len(arr) == 3
				}
				return false
			},
			expectErr: false,
		},
		{
			name:      "Invalid path",
			path:      "$.nonexistent",
			checkFunc: nil,
			expectErr: true,
		},
	}

	for _, test := range pathTests {
		t.Run("ParseJSONWithPath_"+test.name, func(t *testing.T) {
			result, err := parser.ParseJSONWithPath(jsonData, test.path)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if !test.checkFunc(result) {
					t.Errorf("Result did not match expected value: %v", result)
				}
			}
		})
	}

	// Test FilterJSON
	filterTests := []struct {
		name       string
		path       string
		filter     string
		length     int
		expectErr  bool
	}{
		{
			name:      "Filter active users",
			path:      "$.users",
			filter:    "@.active == true",
			length:    2,
			expectErr: false,
		},
		{
			name:      "Filter by name",
			path:      "$.users",
			filter:    "@.name contains 'John'",
			length:    1,
			expectErr: false,
		},
		{
			name:      "Filter by ID",
			path:      "$.users",
			filter:    "@.id > 1",
			length:    2,
			expectErr: false,
		},
		{
			name:      "Invalid filter",
			path:      "$.users",
			filter:    "@.id invalid 1",
			length:    0,
			expectErr: true,
		},
	}

	for _, test := range filterTests {
		t.Run("FilterJSON_"+test.name, func(t *testing.T) {
			result, err := parser.FilterJSON(jsonData, test.path, test.filter)
			
			if test.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !test.expectErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(result) != test.length {
					t.Errorf("Expected %d items, got %d", test.length, len(result))
				}
			}
		})
	}
}