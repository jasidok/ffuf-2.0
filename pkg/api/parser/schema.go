// Package parser provides functionality for parsing API responses and specifications.
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/ffuf/ffuf/v2/pkg/api"
)

// SchemaType represents the type of a schema field
type SchemaType string

const (
	// TypeString represents a string field
	TypeString SchemaType = "string"
	// TypeNumber represents a number field
	TypeNumber SchemaType = "number"
	// TypeInteger represents an integer field
	TypeInteger SchemaType = "integer"
	// TypeBoolean represents a boolean field
	TypeBoolean SchemaType = "boolean"
	// TypeArray represents an array field
	TypeArray SchemaType = "array"
	// TypeObject represents an object field
	TypeObject SchemaType = "object"
	// TypeNull represents a null field
	TypeNull SchemaType = "null"
)

// SchemaFormat represents the format of a schema field
type SchemaFormat string

const (
	// FormatNone represents no specific format
	FormatNone SchemaFormat = ""
	// FormatDate represents a date format
	FormatDate SchemaFormat = "date"
	// FormatDateTime represents a date-time format
	FormatDateTime SchemaFormat = "date-time"
	// FormatEmail represents an email format
	FormatEmail SchemaFormat = "email"
	// FormatUUID represents a UUID format
	FormatUUID SchemaFormat = "uuid"
	// FormatURI represents a URI format
	FormatURI SchemaFormat = "uri"
	// FormatIPv4 represents an IPv4 format
	FormatIPv4 SchemaFormat = "ipv4"
	// FormatIPv6 represents an IPv6 format
	FormatIPv6 SchemaFormat = "ipv6"
)

// SchemaField represents a field in a schema
type SchemaField struct {
	Type        SchemaType   `json:"type"`
	Format      SchemaFormat `json:"format,omitempty"`
	Description string       `json:"description,omitempty"`
	Required    bool         `json:"required,omitempty"`
	Enum        []string     `json:"enum,omitempty"`
	Pattern     string       `json:"pattern,omitempty"`
	Minimum     *float64     `json:"minimum,omitempty"`
	Maximum     *float64     `json:"maximum,omitempty"`
	MinLength   *int         `json:"minLength,omitempty"`
	MaxLength   *int         `json:"maxLength,omitempty"`
	Items       *SchemaField `json:"items,omitempty"`
	Properties  SchemaMap    `json:"properties,omitempty"`
}

// SchemaMap is a map of field names to schema fields
type SchemaMap map[string]SchemaField

// Schema represents a detected API schema
type Schema struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Type        SchemaType   `json:"type"`
	Format      SchemaFormat `json:"format,omitempty"`
	Properties  SchemaMap    `json:"properties,omitempty"`
	Items       *SchemaField `json:"items,omitempty"`
}

// SchemaDetector provides methods for detecting API schemas from responses
type SchemaDetector struct {
	// Configuration options
	DetectFormats     bool
	DetectPatterns    bool
	DetectEnums       bool
	DetectRanges      bool
	SampleSize        int
	MinEnumValues     int
	MaxEnumValues     int
	ConfidenceThreshold float64
}

// NewSchemaDetector creates a new SchemaDetector with default settings
func NewSchemaDetector() *SchemaDetector {
	return &SchemaDetector{
		DetectFormats:     true,
		DetectPatterns:    true,
		DetectEnums:       true,
		DetectRanges:      true,
		SampleSize:        10,
		MinEnumValues:     2,
		MaxEnumValues:     20,
		ConfidenceThreshold: 0.8,
	}
}

// DetectSchema detects the schema of a JSON response
func (d *SchemaDetector) DetectSchema(data []byte) (*Schema, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, api.NewAPIError("Failed to parse JSON data: "+err.Error(), 0)
	}

	schema, err := d.inferSchema(jsonData, "")
	if err != nil {
		return nil, err
	}

	return schema, nil
}

// DetectSchemaFromSamples detects the schema from multiple JSON response samples
func (d *SchemaDetector) DetectSchemaFromSamples(samples [][]byte) (*Schema, error) {
	if len(samples) == 0 {
		return nil, api.NewAPIError("No samples provided", 0)
	}

	// Parse all samples
	var parsedSamples []interface{}
	for _, sample := range samples {
		var jsonData interface{}
		if err := json.Unmarshal(sample, &jsonData); err != nil {
			return nil, api.NewAPIError("Failed to parse JSON sample: "+err.Error(), 0)
		}
		parsedSamples = append(parsedSamples, jsonData)
	}

	// Infer schema from the first sample
	schema, err := d.inferSchema(parsedSamples[0], "")
	if err != nil {
		return nil, err
	}

	// Refine schema with additional samples
	for i := 1; i < len(parsedSamples); i++ {
		refinedSchema, err := d.inferSchema(parsedSamples[i], "")
		if err != nil {
			return nil, err
		}

		schema = d.mergeSchemas(schema, refinedSchema)
	}

	return schema, nil
}

// inferSchema infers the schema of a JSON value
func (d *SchemaDetector) inferSchema(value interface{}, path string) (*Schema, error) {
	schema := &Schema{}

	switch v := value.(type) {
	case map[string]interface{}:
		// Object type
		schema.Type = TypeObject
		schema.Properties = make(SchemaMap)

		for key, val := range v {
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += key

			field, err := d.inferField(val, fieldPath)
			if err != nil {
				return nil, err
			}

			schema.Properties[key] = *field
		}

	case []interface{}:
		// Array type
		schema.Type = TypeArray

		if len(v) > 0 {
			// Infer schema of array items from the first item
			itemField, err := d.inferField(v[0], path+"[0]")
			if err != nil {
				return nil, err
			}

			// Refine with additional items
			for i := 1; i < len(v) && i < d.SampleSize; i++ {
				refinedField, err := d.inferField(v[i], fmt.Sprintf("%s[%d]", path, i))
				if err != nil {
					return nil, err
				}

				itemField = d.mergeFields(itemField, refinedField)
			}

			schema.Items = itemField
		}

	default:
		// Scalar type
		field, err := d.inferField(value, path)
		if err != nil {
			return nil, err
		}

		schema.Type = field.Type
		schema.Format = field.Format
	}

	return schema, nil
}

// inferField infers the schema field for a JSON value
func (d *SchemaDetector) inferField(value interface{}, path string) (*SchemaField, error) {
	field := &SchemaField{}

	switch v := value.(type) {
	case string:
		field.Type = TypeString

		// Detect format
		if d.DetectFormats {
			field.Format = d.detectStringFormat(v)
		}

		// Detect pattern
		if d.DetectPatterns && field.Format == FormatNone {
			pattern := d.detectStringPattern(v)
			if pattern != "" {
				field.Pattern = pattern
			}
		}

		// Set length constraints
		length := len(v)
		field.MinLength = &length
		field.MaxLength = &length

	case float64:
		// Check if it's an integer
		if v == float64(int(v)) {
			field.Type = TypeInteger
		} else {
			field.Type = TypeNumber
		}

		// Set range constraints
		field.Minimum = &v
		field.Maximum = &v

	case bool:
		field.Type = TypeBoolean

	case nil:
		field.Type = TypeNull

	case []interface{}:
		field.Type = TypeArray

		if len(v) > 0 {
			// Infer schema of array items from the first item
			itemField, err := d.inferField(v[0], path+"[0]")
			if err != nil {
				return nil, err
			}

			// Refine with additional items
			for i := 1; i < len(v) && i < d.SampleSize; i++ {
				refinedField, err := d.inferField(v[i], fmt.Sprintf("%s[%d]", path, i))
				if err != nil {
					return nil, err
				}

				itemField = d.mergeFields(itemField, refinedField)
			}

			field.Items = itemField
		}

	case map[string]interface{}:
		field.Type = TypeObject
		field.Properties = make(SchemaMap)

		for key, val := range v {
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += key

			subField, err := d.inferField(val, fieldPath)
			if err != nil {
				return nil, err
			}

			field.Properties[key] = *subField
		}

	default:
		// Unknown type
		return nil, api.NewAPIError(fmt.Sprintf("Unsupported type: %T", value), 0)
	}

	return field, nil
}

// detectStringFormat detects the format of a string value
func (d *SchemaDetector) detectStringFormat(value string) SchemaFormat {
	// Check for date-time format
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return FormatDateTime
	}

	// Check for date format
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return FormatDate
	}

	// Check for email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(value) {
		return FormatEmail
	}

	// Check for UUID format
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if uuidRegex.MatchString(value) {
		return FormatUUID
	}

	// Check for URI format
	uriRegex := regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)
	if uriRegex.MatchString(value) {
		return FormatURI
	}

	// Check for IPv4 format
	ipv4Regex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	if ipv4Regex.MatchString(value) {
		return FormatIPv4
	}

	// Check for IPv6 format
	ipv6Regex := regexp.MustCompile(`^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)
	if ipv6Regex.MatchString(value) {
		return FormatIPv6
	}

	return FormatNone
}

// detectStringPattern detects a pattern for a string value
func (d *SchemaDetector) detectStringPattern(value string) string {
	// Check for numeric pattern
	if regexp.MustCompile(`^\d+$`).MatchString(value) {
		return "^\\d+$"
	}

	// Check for alphanumeric pattern
	if regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value) {
		return "^[a-zA-Z0-9]+$"
	}

	// Check for hex pattern
	if regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(value) {
		return "^[0-9a-fA-F]+$"
	}

	// Check for slug pattern
	if regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value) {
		return "^[a-z0-9]+(?:-[a-z0-9]+)*$"
	}

	return ""
}

// mergeSchemas merges two schemas
func (d *SchemaDetector) mergeSchemas(schema1, schema2 *Schema) *Schema {
	if schema1 == nil {
		return schema2
	}
	if schema2 == nil {
		return schema1
	}

	result := &Schema{
		Title:       schema1.Title,
		Description: schema1.Description,
		Type:        schema1.Type,
	}

	// If types don't match, use a more general type
	if schema1.Type != schema2.Type {
		result.Type = d.generalizeType(schema1.Type, schema2.Type)
	}

	// Merge properties for object types
	if result.Type == TypeObject {
		result.Properties = make(SchemaMap)

		// Add properties from schema1
		for key, field := range schema1.Properties {
			result.Properties[key] = field
		}

		// Merge properties from schema2
		for key, field := range schema2.Properties {
			if existingField, ok := result.Properties[key]; ok {
				// Merge fields
				mergedField := d.mergeFields(&existingField, &field)
				result.Properties[key] = *mergedField
			} else {
				// Add new field
				result.Properties[key] = field
			}
		}
	}

	// Merge items for array types
	if result.Type == TypeArray {
		if schema1.Items != nil && schema2.Items != nil {
			result.Items = d.mergeFields(schema1.Items, schema2.Items)
		} else if schema1.Items != nil {
			result.Items = schema1.Items
		} else if schema2.Items != nil {
			result.Items = schema2.Items
		}
	}

	return result
}

// mergeFields merges two schema fields
func (d *SchemaDetector) mergeFields(field1, field2 *SchemaField) *SchemaField {
	if field1 == nil {
		return field2
	}
	if field2 == nil {
		return field1
	}

	result := &SchemaField{
		Description: field1.Description,
		Required:    field1.Required && field2.Required,
	}

	// If types don't match, use a more general type
	if field1.Type != field2.Type {
		result.Type = d.generalizeType(field1.Type, field2.Type)
	} else {
		result.Type = field1.Type
	}

	// Merge format
	if field1.Format == field2.Format {
		result.Format = field1.Format
	} else {
		// If formats don't match, don't specify a format
		result.Format = FormatNone
	}

	// Merge pattern
	if field1.Pattern == field2.Pattern {
		result.Pattern = field1.Pattern
	}

	// Merge range constraints
	if field1.Minimum != nil && field2.Minimum != nil {
		min := *field1.Minimum
		if *field2.Minimum < min {
			min = *field2.Minimum
		}
		result.Minimum = &min
	} else if field1.Minimum != nil {
		result.Minimum = field1.Minimum
	} else if field2.Minimum != nil {
		result.Minimum = field2.Minimum
	}

	if field1.Maximum != nil && field2.Maximum != nil {
		max := *field1.Maximum
		if *field2.Maximum > max {
			max = *field2.Maximum
		}
		result.Maximum = &max
	} else if field1.Maximum != nil {
		result.Maximum = field1.Maximum
	} else if field2.Maximum != nil {
		result.Maximum = field2.Maximum
	}

	// Merge length constraints
	if field1.MinLength != nil && field2.MinLength != nil {
		min := *field1.MinLength
		if *field2.MinLength < min {
			min = *field2.MinLength
		}
		result.MinLength = &min
	} else if field1.MinLength != nil {
		result.MinLength = field1.MinLength
	} else if field2.MinLength != nil {
		result.MinLength = field2.MinLength
	}

	if field1.MaxLength != nil && field2.MaxLength != nil {
		max := *field1.MaxLength
		if *field2.MaxLength > max {
			max = *field2.MaxLength
		}
		result.MaxLength = &max
	} else if field1.MaxLength != nil {
		result.MaxLength = field1.MaxLength
	} else if field2.MaxLength != nil {
		result.MaxLength = field2.MaxLength
	}

	// Merge properties for object types
	if result.Type == TypeObject {
		result.Properties = make(SchemaMap)

		// Add properties from field1
		for key, subField := range field1.Properties {
			result.Properties[key] = subField
		}

		// Merge properties from field2
		for key, subField := range field2.Properties {
			if existingField, ok := result.Properties[key]; ok {
				// Merge fields
				mergedField := d.mergeFields(&existingField, &subField)
				result.Properties[key] = *mergedField
			} else {
				// Add new field
				result.Properties[key] = subField
			}
		}
	}

	// Merge items for array types
	if result.Type == TypeArray {
		if field1.Items != nil && field2.Items != nil {
			result.Items = d.mergeFields(field1.Items, field2.Items)
		} else if field1.Items != nil {
			result.Items = field1.Items
		} else if field2.Items != nil {
			result.Items = field2.Items
		}
	}

	return result
}

// generalizeType returns a more general type that can represent both input types
func (d *SchemaDetector) generalizeType(type1, type2 SchemaType) SchemaType {
	// If either type is null, use the other type
	if type1 == TypeNull {
		return type2
	}
	if type2 == TypeNull {
		return type1
	}

	// If types are the same, return that type
	if type1 == type2 {
		return type1
	}

	// If one is integer and the other is number, use number
	if (type1 == TypeInteger && type2 == TypeNumber) || (type1 == TypeNumber && type2 == TypeInteger) {
		return TypeNumber
	}

	// For all other cases, we can't generalize, so return string as the most general type
	return TypeString
}

// ConvertToJSONSchema converts the detected schema to a JSON Schema document
func (d *SchemaDetector) ConvertToJSONSchema(schema *Schema) ([]byte, error) {
	jsonSchema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    string(schema.Type),
	}

	if schema.Title != "" {
		jsonSchema["title"] = schema.Title
	}

	if schema.Description != "" {
		jsonSchema["description"] = schema.Description
	}

	if schema.Type == TypeObject && len(schema.Properties) > 0 {
		properties := make(map[string]interface{})
		required := []string{}

		for name, field := range schema.Properties {
			properties[name] = d.fieldToJSONSchema(&field)

			if field.Required {
				required = append(required, name)
			}
		}

		jsonSchema["properties"] = properties

		if len(required) > 0 {
			jsonSchema["required"] = required
		}
	}

	if schema.Type == TypeArray && schema.Items != nil {
		jsonSchema["items"] = d.fieldToJSONSchema(schema.Items)
	}

	return json.MarshalIndent(jsonSchema, "", "  ")
}

// fieldToJSONSchema converts a schema field to a JSON Schema field
func (d *SchemaDetector) fieldToJSONSchema(field *SchemaField) map[string]interface{} {
	result := map[string]interface{}{
		"type": string(field.Type),
	}

	if field.Format != FormatNone {
		result["format"] = string(field.Format)
	}

	if field.Description != "" {
		result["description"] = field.Description
	}

	if field.Pattern != "" {
		result["pattern"] = field.Pattern
	}

	if field.Minimum != nil {
		result["minimum"] = *field.Minimum
	}

	if field.Maximum != nil {
		result["maximum"] = *field.Maximum
	}

	if field.MinLength != nil {
		result["minLength"] = *field.MinLength
	}

	if field.MaxLength != nil {
		result["maxLength"] = *field.MaxLength
	}

	if len(field.Enum) > 0 {
		result["enum"] = field.Enum
	}

	if field.Type == TypeObject && len(field.Properties) > 0 {
		properties := make(map[string]interface{})
		required := []string{}

		for name, subField := range field.Properties {
			properties[name] = d.fieldToJSONSchema(&subField)

			if subField.Required {
				required = append(required, name)
			}
		}

		result["properties"] = properties

		if len(required) > 0 {
			result["required"] = required
		}
	}

	if field.Type == TypeArray && field.Items != nil {
		result["items"] = d.fieldToJSONSchema(field.Items)
	}

	return result
}

// Add schema detection methods to ResponseParser

// DetectSchema detects the schema of a JSON response
func (p *ResponseParser) DetectSchema(data []byte) (*Schema, error) {
	if p.format != FormatJSON && p.format != FormatUnknown {
		return nil, api.NewAPIError("Response is not in JSON format", 0)
	}

	detector := NewSchemaDetector()
	return detector.DetectSchema(data)
}

// DetectSchemaFromSamples detects the schema from multiple JSON response samples
func (p *ResponseParser) DetectSchemaFromSamples(samples [][]byte) (*Schema, error) {
	if p.format != FormatJSON && p.format != FormatUnknown {
		return nil, api.NewAPIError("Response is not in JSON format", 0)
	}

	detector := NewSchemaDetector()
	return detector.DetectSchemaFromSamples(samples)
}

// ConvertToJSONSchema converts the detected schema to a JSON Schema document
func (p *ResponseParser) ConvertToJSONSchema(schema *Schema) ([]byte, error) {
	detector := NewSchemaDetector()
	return detector.ConvertToJSONSchema(schema)
}
