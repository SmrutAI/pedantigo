package schemagen

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs
// SimpleStruct represents the data structure.
type SimpleStruct struct {
	Name  string `json:"name" pedantigo:"required,min=3,max=50"`
	Age   int    `json:"age" pedantigo:"min=18,max=100"`
	Email string `json:"email" pedantigo:"email"`
}

type NestedStruct struct {
	User    SimpleStruct `json:"user" pedantigo:"required"`
	Created time.Time    `json:"created"`
}

type SliceStruct struct {
	Tags []string `json:"tags" pedantigo:"min=1,max=10"`
	IDs  []int    `json:"ids" pedantigo:"min=1"`
}

type MapStruct struct {
	Metadata map[string]string `json:"metadata"`
}

type ConstraintsStruct struct {
	URL        string  `json:"url" pedantigo:"url"`
	UUID       string  `json:"uuid" pedantigo:"uuid"`
	IPv4       string  `json:"ipv4" pedantigo:"ipv4"`
	IPv6       string  `json:"ipv6" pedantigo:"ipv6"`
	Pattern    string  `json:"pattern" pedantigo:"regexp=^[A-Z]+$"`
	Status     string  `json:"status" pedantigo:"oneof=active inactive pending"`
	Score      float64 `json:"score" pedantigo:"gte=0,lte=100"`
	Count      int     `json:"count" pedantigo:"gt=0,lt=1000"`
	Enabled    bool    `json:"enabled" pedantigo:"default=true"`
	MaxRetries int     `json:"max_retries" pedantigo:"default=3"`
}

// Test containers for nested types
// SliceContainer represents the data structure.
type SliceContainer struct {
	Items []SimpleStruct `json:"items"`
}

type MapContainer struct {
	Data map[string]SimpleStruct `json:"data"`
}

func TestGenerateBaseSchema(t *testing.T) {
	tests := []struct {
		name         string
		generateFunc func() *jsonschema.Schema
		wantType     string
		checkProps   []string
		wantRequired int
		// Optional: extra validation for specific properties
		checkPropType map[string]string // property name -> expected type
	}{
		{
			name:         "simple struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[SimpleStruct]() },
			wantType:     "object",
			checkProps:   []string{"name", "age", "email"},
			wantRequired: 0, // Cleared by GenerateBaseSchema
		},
		{
			name:         "nested struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[NestedStruct]() },
			wantType:     "object",
			checkProps:   []string{"user", "created"},
			wantRequired: 0,
			checkPropType: map[string]string{
				"user": "object",
			},
		},
		{
			name:         "slice struct",
			generateFunc: func() *jsonschema.Schema { return GenerateBaseSchema[SliceStruct]() },
			wantType:     "object",
			checkProps:   []string{"tags", "ids"},
			wantRequired: 0,
			checkPropType: map[string]string{
				"tags": "array",
			},
		},
	}

	// Common validation logic - runs for EVERY test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.generateFunc()

			require.NotNil(t, schema, "schema is nil")
			assert.Equal(t, tt.wantType, schema.Type)
			require.NotNil(t, schema.Properties, "properties is nil")

			// Check all expected properties exist
			for _, prop := range tt.checkProps {
				_, ok := schema.Properties.Get(prop)
				assert.True(t, ok, "%s property not found", prop)
			}

			// Check required count
			assert.Len(t, schema.Required, tt.wantRequired)

			// Optional: check specific property types
			for propName, expectedType := range tt.checkPropType {
				prop, ok := schema.Properties.Get(propName)
				require.True(t, ok, "%s property not found", propName)
				assert.Equal(t, expectedType, prop.Type)
			}
		})
	}
}

func TestGenerateOpenAPIBaseSchema(t *testing.T) {
	schema := GenerateOpenAPIBaseSchema[SimpleStruct]()

	require.NotNil(t, schema, "schema is nil")
	assert.Equal(t, "object", schema.Type)
}

func TestEnhanceSchema(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		// Simple parser for testing
		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	tests := []struct {
		name         string
		structType   reflect.Type
		wantRequired []string
		// Optional: custom check (max 4-5 lines)
		checkFunc func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:         "enhance simple struct with constraints",
			structType:   reflect.TypeOf(SimpleStruct{}),
			wantRequired: []string{"name"},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				// Check name constraints
				nameProp, _ := schema.Properties.Get("name")
				require.NotNil(t, nameProp.MinLength)
				assert.Equal(t, uint64(3), *nameProp.MinLength)
				require.NotNil(t, nameProp.MaxLength)
				assert.Equal(t, uint64(50), *nameProp.MaxLength)

				// Check age constraints
				ageProp, _ := schema.Properties.Get("age")
				assert.Equal(t, json.Number("18"), ageProp.Minimum)
				assert.Equal(t, json.Number("100"), ageProp.Maximum)

				// Check email format
				emailProp, _ := schema.Properties.Get("email")
				assert.Equal(t, "email", emailProp.Format)
			},
		},
		{
			name:         "enhance nested struct",
			structType:   reflect.TypeOf(NestedStruct{}),
			wantRequired: []string{"user"},
		},
	}

	// Common validation logic
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateBaseSchema[SimpleStruct]()
			if tt.structType == reflect.TypeOf(NestedStruct{}) {
				schema = GenerateBaseSchema[NestedStruct]()
			}

			EnhanceSchema(schema, tt.structType, mockParseTagFunc)

			// Check required fields
			for _, reqField := range tt.wantRequired {
				assert.Contains(t, schema.Required, reqField)
			}

			// Optional custom validation
			if tt.checkFunc != nil {
				tt.checkFunc(t, schema)
			}
		})
	}
}

func TestApplyConstraints(t *testing.T) {
	tests := []struct {
		name        string
		fieldType   reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:      "min max for string",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"min": "5",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(100), *schema.MaxLength)
			},
		},
		{
			name:      "min max for int",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"min": "10",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("10"), schema.Minimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:      "gt gte lt lte",
			fieldType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gt":  "0",
				"lte": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:      "format constraints",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"email": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "email", schema.Format)
			},
		},
		{
			name:      "url format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"url": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uri", schema.Format)
			},
		},
		{
			name:      "uuid format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uuid": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uuid", schema.Format)
			},
		},
		{
			name:      "ipv4 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv4", schema.Format)
			},
		},
		{
			name:      "ipv6 format",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv6", schema.Format)
			},
		},
		{
			name:      "regexp pattern",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"regexp": "^[A-Z]+$",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[A-Z]+$", schema.Pattern)
			},
		},
		{
			name:      "oneof enum",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"oneof": "red green blue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Len(t, schema.Enum, 3)
			},
		},
		{
			name:      "default value",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"default": "hello",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "hello", schema.Default)
			},
		},
		{
			name:      "len constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"len": "8",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(8), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(8), *schema.MaxLength)
			},
		},
		{
			name:      "ascii constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ascii": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[\\x00-\\x7F]*$", schema.Pattern)
			},
		},
		{
			name:      "alpha constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alpha": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z]+$", schema.Pattern)
			},
		},
		{
			name:      "alphanum constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"alphanum": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-zA-Z0-9]+$", schema.Pattern)
			},
		},
		{
			name:      "contains constraint",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"contains": "test",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, ".*test.*", schema.Pattern)
			},
		},
		{
			name:      "contains with special chars",
			fieldType: reflect.TypeOf(""),
			constraints: map[string]string{
				"contains": "@example.com",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, `.*@example\.com.*`, schema.Pattern)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{}
			ApplyConstraints(schema, tt.constraints, tt.fieldType)
			tt.checkFunc(t, schema)
		})
	}
}

func TestParseDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		typ      reflect.Type
		expected any
	}{
		{
			name:     "string default",
			value:    "test",
			typ:      reflect.TypeOf(""),
			expected: "test",
		},
		{
			name:     "int default",
			value:    "42",
			typ:      reflect.TypeOf(0),
			expected: int64(42),
		},
		{
			name:     "uint default",
			value:    "42",
			typ:      reflect.TypeOf(uint(0)),
			expected: uint64(42),
		},
		{
			name:     "float default",
			value:    "3.14",
			typ:      reflect.TypeOf(0.0),
			expected: 3.14,
		},
		{
			name:     "bool default true",
			value:    "true",
			typ:      reflect.TypeOf(false),
			expected: true,
		},
		{
			name:     "bool default false",
			value:    "false",
			typ:      reflect.TypeOf(false),
			expected: false,
		},
		{
			name:     "invalid int returns string",
			value:    "invalid",
			typ:      reflect.TypeOf(0),
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDefaultValue(tt.value, tt.typ)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhanceNestedTypes(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	tests := []struct {
		name             string
		structType       reflect.Type
		checkProp        string // Property name to check
		expectItems      bool   // Expect Items field (for slices)
		expectAdditional bool   // Expect AdditionalProperties (for maps)
		skipCheck        bool   // Skip property checks (e.g., time.Time)
	}{
		{
			name:        "enhance slice of structs",
			structType:  reflect.TypeOf(SliceContainer{}),
			checkProp:   "items",
			expectItems: true,
		},
		{
			name:             "enhance map of structs",
			structType:       reflect.TypeOf(MapContainer{}),
			checkProp:        "data",
			expectAdditional: true,
		},
		{
			name:       "skip time.Time enhancement",
			structType: reflect.TypeOf(time.Time{}),
			skipCheck:  true,
		},
	}

	// Common validation logic
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCheck {
				// Should not panic
				schema := &jsonschema.Schema{}
				EnhanceNestedTypes(schema, tt.structType, mockParseTagFunc)
				return
			}

			// Generate and enhance schema
			var schema *jsonschema.Schema
			if tt.structType == reflect.TypeOf(SliceContainer{}) {
				schema = GenerateBaseSchema[SliceContainer]()
			} else if tt.structType == reflect.TypeOf(MapContainer{}) {
				schema = GenerateBaseSchema[MapContainer]()
			}

			EnhanceSchema(schema, tt.structType, mockParseTagFunc)

			// Check property exists
			prop, ok := schema.Properties.Get(tt.checkProp)
			require.True(t, ok, "%s property not found", tt.checkProp)

			// Check expected fields (max 4 lines)
			if tt.expectItems {
				require.NotNil(t, prop.Items, "items schema is nil")
			}
			if tt.expectAdditional {
				require.NotNil(t, prop.AdditionalProperties, "additionalProperties schema is nil")
			}
		})
	}
}

func TestApplyConstraintsToItems(t *testing.T) {
	tests := []struct {
		name        string
		elemType    reflect.Type
		constraints map[string]string
		checkFunc   func(t *testing.T, schema *jsonschema.Schema)
	}{
		{
			name:     "format constraints for items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"email": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "email", schema.Format)
			},
		},
		{
			name:     "min max for string items",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"min": "5",
				"max": "50",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				require.NotNil(t, schema.MinLength)
				assert.Equal(t, uint64(5), *schema.MinLength)
				require.NotNil(t, schema.MaxLength)
				assert.Equal(t, uint64(50), *schema.MaxLength)
			},
		},
		{
			name:     "min max for int items",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"min": "1",
				"max": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
				assert.Equal(t, json.Number("100"), schema.Maximum)
			},
		},
		{
			name:     "url format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"url": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uri", schema.Format)
			},
		},
		{
			name:     "uuid format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"uuid": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "uuid", schema.Format)
			},
		},
		{
			name:     "ipv4 format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv4": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv4", schema.Format)
			},
		},
		{
			name:     "ipv6 format constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"ipv6": "",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "ipv6", schema.Format)
			},
		},
		{
			name:     "regexp pattern constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"regexp": "^[a-z]+$",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, "^[a-z]+$", schema.Pattern)
			},
		},
		{
			name:     "oneof enum constraint",
			elemType: reflect.TypeOf(""),
			constraints: map[string]string{
				"oneof": "red green blue",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Len(t, schema.Enum, 3)
				expected := []string{"red", "green", "blue"}
				for i, val := range schema.Enum {
					assert.Equal(t, expected[i], val)
				}
			},
		},
		{
			name:     "gt exclusive minimum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gt": "0",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("0"), schema.ExclusiveMinimum)
			},
		},
		{
			name:     "gte inclusive minimum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"gte": "1",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("1"), schema.Minimum)
			},
		},
		{
			name:     "lt exclusive maximum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"lt": "100",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("100"), schema.ExclusiveMaximum)
			},
		},
		{
			name:     "lte inclusive maximum",
			elemType: reflect.TypeOf(0),
			constraints: map[string]string{
				"lte": "99",
			},
			checkFunc: func(t *testing.T, schema *jsonschema.Schema) {
				assert.Equal(t, json.Number("99"), schema.Maximum)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{}
			ApplyConstraintsToItems(schema, tt.constraints, tt.elemType)
			tt.checkFunc(t, schema)
		})
	}
}

func TestFullIntegration(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		pedantigoTag := tag.Get("pedantigo")
		if pedantigoTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(pedantigoTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateBaseSchema[ConstraintsStruct]()
	typ := reflect.TypeOf(ConstraintsStruct{})
	EnhanceSchema(schema, typ, mockParseTagFunc)

	// Marshal to JSON to verify structure
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err, "failed to marshal schema")

	// Basic sanity checks
	assert.Equal(t, "object", schema.Type)
	require.NotNil(t, schema.Properties, "properties is nil")

	// Verify at least one property was enhanced
	urlProp, ok := schema.Properties.Get("url")
	require.True(t, ok, "url property not found")
	assert.Equal(t, "uri", urlProp.Format)

	t.Logf("Generated schema:\n%s", string(jsonData))
}

// Helper functions for simple tag parsing.
func splitConstraints(tag string) []string {
	var parts []string
	current := ""
	for _, ch := range tag {
		if ch == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitKeyValue(part string) (key, value string, found bool) {
	for i, ch := range part {
		if ch == '=' {
			return part[:i], part[i+1:], true
		}
	}
	return "", "", false
}

// Test structs for discriminated union schema generation.

// CatUnion represents a cat variant with validation constraints.
type CatUnion struct {
	Name  string `json:"name" validate:"required"`
	Lives int    `json:"lives" validate:"min=1,max=9"`
}

// DogUnion represents a dog variant with validation constraints.
type DogUnion struct {
	Name  string `json:"name" validate:"required"`
	Breed string `json:"breed"`
}

// BirdUnion represents a bird variant with validation constraints.
type BirdUnion struct {
	Name string `json:"name" validate:"required"`
	Eggs int    `json:"eggs" validate:"min=0"`
}

// NestedVariant represents a variant with nested struct fields.
type NestedVariant struct {
	ID    string       `json:"id" validate:"required"`
	Owner SimpleStruct `json:"owner" validate:"required"`
}

// TestGenerateVariantSchema_BasicVariant tests GenerateVariantSchema
// generates schema for variant type with properties.
func TestGenerateVariantSchema_BasicVariant(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(CatUnion{}),
		"pet_type",
		"cat",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "variant schema should not be nil")
	assert.Equal(t, "object", schema.Type)
	require.NotNil(t, schema.Properties, "properties should not be nil")

	// Check discriminator field exists
	_, ok := schema.Properties.Get("pet_type")
	assert.True(t, ok, "pet_type property should exist")

	// Check variant properties exist
	_, ok = schema.Properties.Get("name")
	assert.True(t, ok, "name property should exist")
	_, ok = schema.Properties.Get("lives")
	assert.True(t, ok, "lives property should exist")
}

// TestGenerateVariantSchema_DiscriminatorConstConstraint tests
// GenerateVariantSchema adds const constraint on discriminator field.
func TestGenerateVariantSchema_DiscriminatorConstConstraint(t *testing.T) {
	tests := []struct {
		name               string
		variantType        reflect.Type
		discriminatorField string
		discriminatorValue string
	}{
		{
			name:               "cat variant",
			variantType:        reflect.TypeOf(CatUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "cat",
		},
		{
			name:               "dog variant",
			variantType:        reflect.TypeOf(DogUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "dog",
		},
		{
			name:               "bird variant",
			variantType:        reflect.TypeOf(BirdUnion{}),
			discriminatorField: "pet_type",
			discriminatorValue: "bird",
		},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateVariantSchema(
				tt.variantType,
				tt.discriminatorField,
				tt.discriminatorValue,
				mockParseTagFunc,
			)

			require.NotNil(t, schema)
			require.NotNil(t, schema.Properties)

			// Check discriminator field has const constraint
			discriminatorSchema, ok := schema.Properties.Get(tt.discriminatorField)
			require.True(t, ok, "discriminator field should exist")
			assert.Equal(t, tt.discriminatorValue, discriminatorSchema.Const,
				"discriminator field should have const %s", tt.discriminatorValue)
		})
	}
}

// TestGenerateVariantSchema_IncludesValidationConstraints tests
// GenerateVariantSchema includes validation constraints from pedantigo tags.
func TestGenerateVariantSchema_IncludesValidationConstraints(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(CatUnion{}),
		"pet_type",
		"cat",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.Properties)

	// Check required constraint on name
	_, ok := schema.Properties.Get("name")
	require.True(t, ok, "name property should exist")
	// Name should be in required array
	assert.Contains(t, schema.Required, "name", "name should be in required fields")

	// Check min/max constraints on lives
	livesSchema, ok := schema.Properties.Get("lives")
	require.True(t, ok, "lives property should exist")
	require.NotNil(t, livesSchema.Minimum, "lives should have minimum constraint")
	assert.Equal(t, json.Number("1"), livesSchema.Minimum)
	require.NotNil(t, livesSchema.Maximum, "lives should have maximum constraint")
	assert.Equal(t, json.Number("9"), livesSchema.Maximum)
}

// TestGenerateVariantSchema_HandlesNestedStructs tests
// GenerateVariantSchema handles nested struct fields.
func TestGenerateVariantSchema_HandlesNestedStructs(t *testing.T) {
	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateVariantSchema(
		reflect.TypeOf(NestedVariant{}),
		"type",
		"nested",
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "variant schema should not be nil")
	require.NotNil(t, schema.Properties)

	// Check nested struct property exists
	ownerSchema, ok := schema.Properties.Get("owner")
	require.True(t, ok, "owner property should exist")
	assert.Equal(t, "object", ownerSchema.Type, "owner should be an object type")
}

// TestGenerateVariantSchema_DifferentDiscriminatorFields tests
// GenerateVariantSchema works with different discriminator field names.
func TestGenerateVariantSchema_DifferentDiscriminatorFields(t *testing.T) {
	tests := []struct {
		name                  string
		discriminatorField    string
		expectedFieldInSchema string
	}{
		{name: "pet_type", discriminatorField: "pet_type", expectedFieldInSchema: "pet_type"},
		{name: "type", discriminatorField: "type", expectedFieldInSchema: "type"},
		{name: "kind", discriminatorField: "kind", expectedFieldInSchema: "kind"},
		{name: "variant_type", discriminatorField: "variant_type", expectedFieldInSchema: "variant_type"},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string { return nil }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateVariantSchema(
				reflect.TypeOf(CatUnion{}),
				tt.discriminatorField,
				"test_value",
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.Properties)
			_, ok := schema.Properties.Get(tt.expectedFieldInSchema)
			assert.True(t, ok, "discriminator field %s should exist", tt.expectedFieldInSchema)
		})
	}
}

// TestGenerateUnionSchema_ReturnsOneOfSchema tests
// GenerateUnionSchema returns schema with oneOf array.
func TestGenerateUnionSchema_ReturnsOneOfSchema(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat":  reflect.TypeOf(CatUnion{}),
		"dog":  reflect.TypeOf(DogUnion{}),
		"bird": reflect.TypeOf(BirdUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "union schema should not be nil")
	require.NotNil(t, schema.OneOf, "oneOf array should not be nil")
}

// TestGenerateUnionSchema_OneOfArrayLength tests
// GenerateUnionSchema oneOf array contains correct number of variants.
func TestGenerateUnionSchema_OneOfArrayLength(t *testing.T) {
	tests := []struct {
		name          string
		variants      map[string]reflect.Type
		expectedCount int
	}{
		{
			name: "two variants",
			variants: map[string]reflect.Type{
				"cat": reflect.TypeOf(CatUnion{}),
				"dog": reflect.TypeOf(DogUnion{}),
			},
			expectedCount: 2,
		},
		{
			name: "three variants",
			variants: map[string]reflect.Type{
				"cat":  reflect.TypeOf(CatUnion{}),
				"dog":  reflect.TypeOf(DogUnion{}),
				"bird": reflect.TypeOf(BirdUnion{}),
			},
			expectedCount: 3,
		},
		{
			name: "single variant",
			variants: map[string]reflect.Type{
				"cat": reflect.TypeOf(CatUnion{}),
			},
			expectedCount: 1,
		},
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateUnionSchema(
				"pet_type",
				tt.variants,
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.OneOf)
			assert.Len(t, schema.OneOf, tt.expectedCount,
				"oneOf array should contain %d variant schemas", tt.expectedCount)
		})
	}
}

// TestGenerateUnionSchema_EachVariantProperlyGenerated tests
// GenerateUnionSchema generates each variant schema properly.
func TestGenerateUnionSchema_EachVariantProperlyGenerated(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Verify each variant schema is properly formed
	for i, variantSchema := range schema.OneOf {
		assert.NotNil(t, variantSchema, "variant schema at index %d should not be nil", i)
		assert.Equal(t, "object", variantSchema.Type, "variant at index %d should be object type", i)
		assert.NotNil(t, variantSchema.Properties, "variant at index %d should have properties", i)
	}
}

// TestGenerateUnionSchema_DiscriminatorConstInEachVariant tests
// GenerateUnionSchema adds discriminator field with const in each variant.
func TestGenerateUnionSchema_DiscriminatorConstInEachVariant(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Check each variant has pet_type with const (order may vary due to map iteration)
	foundConsts := make(map[string]bool)
	for _, variantSchema := range schema.OneOf {
		require.NotNil(t, variantSchema.Properties)
		petTypeSchema, ok := variantSchema.Properties.Get("pet_type")
		require.True(t, ok, "variant should have pet_type property")
		if constVal, ok := petTypeSchema.Const.(string); ok {
			foundConsts[constVal] = true
		}
	}
	assert.True(t, foundConsts["cat"], "should have variant with const 'cat'")
	assert.True(t, foundConsts["dog"], "should have variant with const 'dog'")
}

// TestGenerateUnionSchema_VariantPropertiesIncluded tests
// GenerateUnionSchema includes variant-specific properties.
func TestGenerateUnionSchema_VariantPropertiesIncluded(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)
	require.Len(t, schema.OneOf, 2)

	// Find cat and dog schemas (order may vary in map)
	var catSchema, dogSchema *jsonschema.Schema
	for _, vs := range schema.OneOf {
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "cat" {
			catSchema = vs
		}
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "dog" {
			dogSchema = vs
		}
	}

	require.NotNil(t, catSchema, "cat variant schema should exist")
	require.NotNil(t, dogSchema, "dog variant schema should exist")

	// Check cat-specific properties
	_, ok := catSchema.Properties.Get("lives")
	assert.True(t, ok, "cat schema should have lives property")

	// Check dog-specific properties
	_, ok = dogSchema.Properties.Get("breed")
	assert.True(t, ok, "dog schema should have breed property")
}

// TestGenerateUnionSchema_DifferentDiscriminatorFields tests
// GenerateUnionSchema works with different discriminator field names.
func TestGenerateUnionSchema_DifferentDiscriminatorFields(t *testing.T) {
	tests := []struct {
		name               string
		discriminatorField string
	}{
		{name: "pet_type", discriminatorField: "pet_type"},
		{name: "type", discriminatorField: "type"},
		{name: "kind", discriminatorField: "kind"},
	}

	variants := map[string]reflect.Type{
		"cat": reflect.TypeOf(CatUnion{}),
		"dog": reflect.TypeOf(DogUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string { return nil }

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := GenerateUnionSchema(
				tt.discriminatorField,
				variants,
				mockParseTagFunc,
			)

			require.NotNil(t, schema, "schema should not be nil")
			require.NotNil(t, schema.OneOf)
			require.Len(t, schema.OneOf, 2)

			// Check discriminator field exists in each variant
			for i, variantSchema := range schema.OneOf {
				require.NotNil(t, variantSchema.Properties)
				_, ok := variantSchema.Properties.Get(tt.discriminatorField)
				assert.True(t, ok,
					"variant at index %d should have %s field", i, tt.discriminatorField)
			}
		})
	}
}

// TestGenerateUnionSchema_WithValidationConstraints tests
// GenerateUnionSchema respects validation constraints in variants.
func TestGenerateUnionSchema_WithValidationConstraints(t *testing.T) {
	variants := map[string]reflect.Type{
		"cat":  reflect.TypeOf(CatUnion{}),
		"bird": reflect.TypeOf(BirdUnion{}),
	}

	mockParseTagFunc := func(tag reflect.StructTag) map[string]string {
		validateTag := tag.Get("validate")
		if validateTag == "" {
			return nil
		}

		constraints := make(map[string]string)
		parts := splitConstraints(validateTag)
		for _, part := range parts {
			if key, value, found := splitKeyValue(part); found {
				constraints[key] = value
			} else {
				constraints[part] = ""
			}
		}
		return constraints
	}

	schema := GenerateUnionSchema(
		"pet_type",
		variants,
		mockParseTagFunc,
	)

	require.NotNil(t, schema, "schema should not be nil")
	require.NotNil(t, schema.OneOf)

	// Find cat schema and verify constraints
	for _, vs := range schema.OneOf {
		if pt, ok := vs.Properties.Get("pet_type"); ok && pt.Const == "cat" {
			// Check lives has min=1 and max=9
			livesSchema, ok := vs.Properties.Get("lives")
			require.True(t, ok, "cat should have lives property")
			assert.NotNil(t, livesSchema.Minimum, "lives should have minimum")
			assert.NotNil(t, livesSchema.Maximum, "lives should have maximum")
		}
	}
}
