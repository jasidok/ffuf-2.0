package parser

import (
	"encoding/json"
	"testing"
)

func TestSchemaDetection(t *testing.T) {
	// Test JSON data
	jsonData := []byte(`{
		"id": 123,
		"name": "Test Product",
		"price": 19.99,
		"is_available": true,
		"tags": ["electronics", "gadgets"],
		"dimensions": {
			"width": 10.5,
			"height": 15.2,
			"unit": "cm"
		},
		"created_at": "2023-01-15T14:30:45Z",
		"manufacturer": {
			"id": 456,
			"name": "ACME Corp",
			"website": "https://example.com",
			"contact_email": "info@example.com"
		}
	}`)

	// Create a schema detector
	detector := NewSchemaDetector()

	// Test DetectSchema
	schema, err := detector.DetectSchema(jsonData)
	if err != nil {
		t.Fatalf("Failed to detect schema: %v", err)
	}

	// Verify schema type
	if schema.Type != TypeObject {
		t.Errorf("Expected schema type to be object, got %s", schema.Type)
	}

	// Verify properties
	expectedProperties := []struct {
		name     string
		dataType SchemaType
		format   SchemaFormat
	}{
		{"id", TypeInteger, FormatNone},
		{"name", TypeString, FormatNone},
		{"price", TypeNumber, FormatNone},
		{"is_available", TypeBoolean, FormatNone},
		{"tags", TypeArray, FormatNone},
		{"dimensions", TypeObject, FormatNone},
		{"created_at", TypeString, FormatDateTime},
		{"manufacturer", TypeObject, FormatNone},
	}

	for _, prop := range expectedProperties {
		field, exists := schema.Properties[prop.name]
		if !exists {
			t.Errorf("Expected property %s not found in schema", prop.name)
			continue
		}

		if field.Type != prop.dataType {
			t.Errorf("Property %s: expected type %s, got %s", prop.name, prop.dataType, field.Type)
		}

		if field.Format != prop.format {
			t.Errorf("Property %s: expected format %s, got %s", prop.name, prop.format, field.Format)
		}
	}

	// Test nested object properties
	dimensionsField, exists := schema.Properties["dimensions"]
	if !exists {
		t.Error("Expected dimensions property not found in schema")
	} else {
		// Check dimensions properties
		expectedDimensionsProps := []struct {
			name     string
			dataType SchemaType
		}{
			{"width", TypeNumber},
			{"height", TypeNumber},
			{"unit", TypeString},
		}

		for _, prop := range expectedDimensionsProps {
			field, exists := dimensionsField.Properties[prop.name]
			if !exists {
				t.Errorf("Expected dimensions property %s not found", prop.name)
				continue
			}

			if field.Type != prop.dataType {
				t.Errorf("Dimensions property %s: expected type %s, got %s", prop.name, prop.dataType, field.Type)
			}
		}
	}

	// Test array items
	tagsField, exists := schema.Properties["tags"]
	if !exists {
		t.Error("Expected tags property not found in schema")
	} else if tagsField.Items == nil {
		t.Error("Expected tags.items to be non-nil")
	} else if tagsField.Items.Type != TypeString {
		t.Errorf("Expected tags.items.type to be string, got %s", tagsField.Items.Type)
	}

	// Test special format detection
	manufacturerField, exists := schema.Properties["manufacturer"]
	if !exists {
		t.Error("Expected manufacturer property not found in schema")
	} else {
		contactEmailField, exists := manufacturerField.Properties["contact_email"]
		if !exists {
			t.Error("Expected manufacturer.contact_email property not found")
		} else if contactEmailField.Format != FormatEmail {
			t.Errorf("Expected manufacturer.contact_email format to be email, got %s", contactEmailField.Format)
		}

		websiteField, exists := manufacturerField.Properties["website"]
		if !exists {
			t.Error("Expected manufacturer.website property not found")
		} else if websiteField.Format != FormatURI {
			t.Errorf("Expected manufacturer.website format to be uri, got %s", websiteField.Format)
		}
	}
}

func TestSchemaDetectionFromSamples(t *testing.T) {
	// Test JSON data samples
	sample1 := []byte(`{
		"id": 123,
		"name": "Product A",
		"price": 19.99,
		"in_stock": true
	}`)

	sample2 := []byte(`{
		"id": 456,
		"name": "Product B",
		"price": 29.99,
		"in_stock": false,
		"tags": ["new", "featured"]
	}`)

	// Create a schema detector
	detector := NewSchemaDetector()

	// Test DetectSchemaFromSamples
	schema, err := detector.DetectSchemaFromSamples([][]byte{sample1, sample2})
	if err != nil {
		t.Fatalf("Failed to detect schema from samples: %v", err)
	}

	// Verify schema type
	if schema.Type != TypeObject {
		t.Errorf("Expected schema type to be object, got %s", schema.Type)
	}

	// Verify properties from both samples are present
	expectedProperties := []struct {
		name     string
		dataType SchemaType
	}{
		{"id", TypeInteger},
		{"name", TypeString},
		{"price", TypeNumber},
		{"in_stock", TypeBoolean},
		{"tags", TypeArray},
	}

	for _, prop := range expectedProperties {
		field, exists := schema.Properties[prop.name]
		if !exists {
			t.Errorf("Expected property %s not found in merged schema", prop.name)
			continue
		}

		if field.Type != prop.dataType {
			t.Errorf("Property %s: expected type %s, got %s", prop.name, prop.dataType, field.Type)
		}
	}

	// Test range constraints
	priceField, exists := schema.Properties["price"]
	if !exists {
		t.Error("Expected price property not found in schema")
	} else {
		if priceField.Minimum == nil {
			t.Error("Expected price.minimum to be non-nil")
		} else if *priceField.Minimum != 19.99 {
			t.Errorf("Expected price.minimum to be 19.99, got %f", *priceField.Minimum)
		}

		if priceField.Maximum == nil {
			t.Error("Expected price.maximum to be non-nil")
		} else if *priceField.Maximum != 29.99 {
			t.Errorf("Expected price.maximum to be 29.99, got %f", *priceField.Maximum)
		}
	}
}

func TestConvertToJSONSchema(t *testing.T) {
	// Create a simple schema
	schema := &Schema{
		Title:       "Test Schema",
		Description: "A test schema",
		Type:        TypeObject,
		Properties: SchemaMap{
			"id": SchemaField{
				Type:        TypeInteger,
				Description: "The ID",
				Required:    true,
			},
			"name": SchemaField{
				Type:        TypeString,
				Description: "The name",
				Required:    true,
				MinLength:   intPtr(1),
				MaxLength:   intPtr(100),
			},
			"email": SchemaField{
				Type:        TypeString,
				Format:      FormatEmail,
				Description: "The email address",
			},
			"tags": SchemaField{
				Type:        TypeArray,
				Description: "The tags",
				Items: &SchemaField{
					Type: TypeString,
				},
			},
		},
	}

	// Create a schema detector
	detector := NewSchemaDetector()

	// Convert to JSON Schema
	jsonSchemaBytes, err := detector.ConvertToJSONSchema(schema)
	if err != nil {
		t.Fatalf("Failed to convert to JSON Schema: %v", err)
	}

	// Parse the JSON Schema
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(jsonSchemaBytes, &jsonSchema); err != nil {
		t.Fatalf("Failed to parse JSON Schema: %v", err)
	}

	// Verify JSON Schema
	if jsonSchema["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("Expected $schema to be http://json-schema.org/draft-07/schema#, got %s", jsonSchema["$schema"])
	}

	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type to be object, got %s", jsonSchema["type"])
	}

	if jsonSchema["title"] != "Test Schema" {
		t.Errorf("Expected title to be Test Schema, got %s", jsonSchema["title"])
	}

	if jsonSchema["description"] != "A test schema" {
		t.Errorf("Expected description to be A test schema, got %s", jsonSchema["description"])
	}

	// Verify properties
	properties, ok := jsonSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// Verify required properties
	required, ok := jsonSchema["required"].([]interface{})
	if !ok {
		t.Fatal("Expected required to be an array")
	}

	if len(required) != 2 {
		t.Errorf("Expected 2 required properties, got %d", len(required))
	}

	// Verify id property
	idProp, ok := properties["id"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected id property to be a map")
	}

	if idProp["type"] != "integer" {
		t.Errorf("Expected id.type to be integer, got %s", idProp["type"])
	}

	// Verify email property
	emailProp, ok := properties["email"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected email property to be a map")
	}

	if emailProp["format"] != "email" {
		t.Errorf("Expected email.format to be email, got %s", emailProp["format"])
	}

	// Verify tags property
	tagsProp, ok := properties["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tags property to be a map")
	}

	if tagsProp["type"] != "array" {
		t.Errorf("Expected tags.type to be array, got %s", tagsProp["type"])
	}

	tagsItems, ok := tagsProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tags.items to be a map")
	}

	if tagsItems["type"] != "string" {
		t.Errorf("Expected tags.items.type to be string, got %s", tagsItems["type"])
	}
}

func TestResponseParserSchemaDetection(t *testing.T) {
	// Test JSON data
	jsonData := []byte(`{
		"users": [
			{
				"id": 1,
				"name": "John Doe",
				"email": "john@example.com"
			},
			{
				"id": 2,
				"name": "Jane Smith",
				"email": "jane@example.com"
			}
		],
		"total": 2,
		"page": 1
	}`)

	// Create a response parser
	parser := NewResponseParser("application/json")

	// Test DetectSchema
	schema, err := parser.DetectSchema(jsonData)
	if err != nil {
		t.Fatalf("Failed to detect schema: %v", err)
	}

	// Verify schema type
	if schema.Type != TypeObject {
		t.Errorf("Expected schema type to be object, got %s", schema.Type)
	}

	// Verify properties
	expectedProperties := []string{"users", "total", "page"}
	for _, propName := range expectedProperties {
		if _, exists := schema.Properties[propName]; !exists {
			t.Errorf("Expected property %s not found in schema", propName)
		}
	}

	// Verify users array
	usersField, exists := schema.Properties["users"]
	if !exists {
		t.Error("Expected users property not found in schema")
	} else if usersField.Type != TypeArray {
		t.Errorf("Expected users.type to be array, got %s", usersField.Type)
	} else if usersField.Items == nil {
		t.Error("Expected users.items to be non-nil")
	} else if usersField.Items.Type != TypeObject {
		t.Errorf("Expected users.items.type to be object, got %s", usersField.Items.Type)
	}

	// Test ConvertToJSONSchema
	jsonSchemaBytes, err := parser.ConvertToJSONSchema(schema)
	if err != nil {
		t.Fatalf("Failed to convert to JSON Schema: %v", err)
	}

	// Verify JSON Schema is valid JSON
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(jsonSchemaBytes, &jsonSchema); err != nil {
		t.Fatalf("Failed to parse JSON Schema: %v", err)
	}

	// Verify JSON Schema has the expected structure
	if jsonSchema["type"] != "object" {
		t.Errorf("Expected type to be object, got %s", jsonSchema["type"])
	}

	properties, ok := jsonSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(properties))
	}
}

// Helper function to create an int pointer
func intPtr(i int) *int {
	return &i
}