package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// StructToJSONSchema converts a Go struct to JSON Schema format using reflection.
// Supports basic types (string, int, float, bool), nested structs, and arrays.
// Reference: research.md decision #2 (Go struct tags → JSON Schema conversion)
func StructToJSONSchema(v interface{}) (map[string]interface{}, error) {
	t := reflect.TypeOf(v)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %v", t.Kind())
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   []string{},
	}

	properties := schema["properties"].(map[string]interface{})
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag (handle omitempty)
		fieldName := strings.Split(jsonTag, ",")[0]

		// Get field schema
		fieldSchema, err := typeToJSONSchema(field.Type)
		if err != nil {
			return nil, fmt.Errorf("error converting field %s: %w", field.Name, err)
		}

		// Check for description in jsonschema tag
		if desc := field.Tag.Get("jsonschema"); desc != "" {
			parts := strings.Split(desc, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "description=") {
					fieldSchema["description"] = strings.TrimPrefix(part, "description=")
				}
			}
		}

		properties[fieldName] = fieldSchema

		// Check if field is required (no omitempty tag)
		if !strings.Contains(jsonTag, "omitempty") {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema, nil
}

// typeToJSONSchema converts a Go type to its JSON Schema representation
func typeToJSONSchema(t reflect.Type) (map[string]interface{}, error) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}, nil

	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}, nil

	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}, nil

	case reflect.Slice, reflect.Array:
		elemSchema, err := typeToJSONSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"type":  "array",
			"items": elemSchema,
		}, nil

	case reflect.Struct:
		// Recursively convert nested struct
		return StructToJSONSchema(reflect.New(t).Interface())

	default:
		return nil, fmt.Errorf("unsupported type: %v", t.Kind())
	}
}
