package constraints

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkConstraintError asserts validation errors based on expected outcome.
func checkConstraintError(t *testing.T, err error, wantErr bool) {
	t.Helper()

	if wantErr {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

// simpleTestCase is a test case structure for simple constraint tests.
type simpleTestCase struct {
	name    string
	value   any
	wantErr bool
}

// runSimpleConstraintTests runs table-driven tests for a simple constraint.
func runSimpleConstraintTests(t *testing.T, c Constraint, tests []simpleTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.Validate(tt.value)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestToFloat64_AllNumericTypes tests toFloat64 with all numeric type cases.
func TestToFloat64_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected float64
	}{
		// Signed integers
		{name: "int", value: int(42), expected: 42.0},
		{name: "int8", value: int8(42), expected: 42.0},
		{name: "int16", value: int16(42), expected: 42.0},
		{name: "int32", value: int32(42), expected: 42.0},
		{name: "int64", value: int64(42), expected: 42.0},
		// Unsigned integers
		{name: "uint", value: uint(42), expected: 42.0},
		{name: "uint8", value: uint8(42), expected: 42.0},
		{name: "uint16", value: uint16(42), expected: 42.0},
		{name: "uint32", value: uint32(42), expected: 42.0},
		{name: "uint64", value: uint64(42), expected: 42.0},
		// Floats
		{name: "float32", value: float32(42.5), expected: 42.5},
		{name: "float64", value: float64(42.5), expected: 42.5},
		// Non-numeric (returns 0)
		{name: "string", value: "test", expected: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			result := toFloat64(val)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

// TestCheckTypeCompatibility_BoolAndTime tests missing branches in CheckTypeCompatibility.
func TestCheckTypeCompatibility_BoolAndTime(t *testing.T) {
	tests := []struct {
		name    string
		a       any
		b       any
		wantErr bool
	}{
		// Bool types
		{name: "bool compatible", a: true, b: false, wantErr: false},
		{name: "bool vs int incompatible", a: true, b: 42, wantErr: true},
		// Time types
		{name: "time.Time compatible", a: time.Now(), b: time.Now(), wantErr: false},
		{name: "time vs string incompatible", a: time.Now(), b: "test", wantErr: true},
		// Nil cases
		{name: "both nil", a: nil, b: nil, wantErr: false},
		{name: "one nil non-pointer", a: nil, b: 42, wantErr: true},
		{name: "nil vs string", a: "test", b: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckTypeCompatibility(tt.a, tt.b)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDereference_PointerLevels tests Dereference with various pointer levels.
func TestDereference_PointerLevels(t *testing.T) {
	tests := []struct {
		name     string
		getType  func() reflect.Type
		expected reflect.Kind
	}{
		{
			name:     "non-pointer",
			getType:  func() reflect.Type { return reflect.TypeOf(42) },
			expected: reflect.Int,
		},
		{
			name: "single pointer",
			getType: func() reflect.Type {
				x := 42
				return reflect.TypeOf(&x)
			},
			expected: reflect.Int,
		},
		{
			name: "double pointer",
			getType: func() reflect.Type {
				x := 42
				p1 := &x
				return reflect.TypeOf(&p1)
			},
			expected: reflect.Int,
		},
		{
			name: "triple pointer",
			getType: func() reflect.Type {
				x := 42
				p1 := &x
				p2 := &p1
				return reflect.TypeOf(&p2)
			},
			expected: reflect.Int,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dereference(tt.getType())
			assert.Equal(t, tt.expected, result.Kind())
		})
	}
}

// TestCompareToString_BoolAndDefault tests missing branches in CompareToString.
func TestCompareToString_BoolAndDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		// Bool cases
		{name: "bool true", value: true, expected: "true"},
		{name: "bool false", value: false, expected: "false"},
		// Pointer to bool
		{name: "pointer to bool", value: func() *bool { b := true; return &b }(), expected: "true"},
		{name: "nil pointer", value: (*int)(nil), expected: ""},
		// Default case (non-standard types)
		{name: "struct default", value: struct{ X int }{X: 42}, expected: "{42}"},
		{name: "slice default", value: []int{1, 2, 3}, expected: "[1 2 3]"},
		// Already covered types (sanity check)
		{name: "string", value: "test", expected: "test"},
		{name: "int", value: 42, expected: "42"},
		{name: "uint", value: uint(42), expected: "42"},
		{name: "float", value: 42.5, expected: "42.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareToString(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildConstraints_MissingBranches tests uncovered constraint types in BuildConstraints.
func TestBuildConstraints_MissingBranches(t *testing.T) {
	tests := []struct {
		name          string
		constraints   map[string]string
		fieldType     reflect.Type
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "gt constraint",
			constraints:   map[string]string{"gt": "10.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"gtConstraint"},
		},
		{
			name:          "gte constraint",
			constraints:   map[string]string{"gte": "20.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"geConstraint"},
		},
		{
			name:          "lt constraint",
			constraints:   map[string]string{"lt": "30.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"ltConstraint"},
		},
		{
			name:          "lte constraint",
			constraints:   map[string]string{"lte": "40.5"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 1,
			expectedTypes: []string{"leConstraint"},
		},
		{
			name:          "ipv4 constraint",
			constraints:   map[string]string{"ipv4": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"ipv4Constraint"},
		},
		{
			name:          "ipv6 constraint",
			constraints:   map[string]string{"ipv6": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"ipv6Constraint"},
		},
		{
			name:          "default constraint",
			constraints:   map[string]string{"default": "test"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"defaultConstraint"},
		},
		{
			name:          "gt with invalid float",
			constraints:   map[string]string{"gt": "invalid"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "gte with invalid float",
			constraints:   map[string]string{"gte": "not-a-number"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "lt with invalid float",
			constraints:   map[string]string{"lt": "xyz"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "lte with invalid float",
			constraints:   map[string]string{"lte": "abc"},
			fieldType:     reflect.TypeOf(float64(0)),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "email constraint",
			constraints:   map[string]string{"email": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"emailConstraint"},
		},
		{
			name:          "url constraint",
			constraints:   map[string]string{"url": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"urlConstraint"},
		},
		{
			name:          "uuid constraint",
			constraints:   map[string]string{"uuid": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"uuidConstraint"},
		},
		{
			name:          "regexp constraint",
			constraints:   map[string]string{"regexp": "^[a-z]+$"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"regexConstraint"},
		},
		{
			name:          "oneof constraint",
			constraints:   map[string]string{"oneof": "red green blue"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 1,
			expectedTypes: []string{"enumConstraint"},
		},
		{
			name:          "required constraint (skipped)",
			constraints:   map[string]string{"required": ""},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "multiple constraints",
			constraints:   map[string]string{"gt": "5", "lte": "100", "ipv4": "", "default": "10"},
			fieldType:     reflect.TypeOf(""),
			expectedCount: 4,
			expectedTypes: []string{"gtConstraint", "leConstraint", "ipv4Constraint", "defaultConstraint"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildConstraints(tt.constraints, tt.fieldType)
			assert.Len(t, result, tt.expectedCount)

			// Verify constraint types (order may vary due to map iteration)
			if len(tt.expectedTypes) > 0 {
				foundTypes := make(map[string]bool)
				for _, c := range result {
					typeName := reflect.TypeOf(c).Name()
					foundTypes[typeName] = true
				}
				for _, expectedType := range tt.expectedTypes {
					assert.True(t, foundTypes[expectedType], "Expected constraint type %s not found", expectedType)
				}
			}
		})
	}
}

// TestParseConditionalConstraint_ErrorPath tests the false return branch.
func TestParseConditionalConstraint_ErrorPath(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		separator  string
		wantOk     bool
		wantFirst  string
		wantSecond string
	}{
		{
			name:       "valid two parts",
			value:      "field:value",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "field",
			wantSecond: "value",
		},
		{
			name:       "no separator",
			value:      "fieldvalue",
			separator:  ":",
			wantOk:     false,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "empty value",
			value:      "",
			separator:  ":",
			wantOk:     false,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "only separator",
			value:      ":",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "",
			wantSecond: "",
		},
		{
			name:       "multiple separators (splits on first)",
			value:      "field:value:extra",
			separator:  ":",
			wantOk:     true,
			wantFirst:  "field",
			wantSecond: "value:extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			first, second, ok := parseConditionalConstraint(tt.value, tt.separator)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantFirst, first)
			assert.Equal(t, tt.wantSecond, second)
		})
	}
}

// TestConstConstraint tests the const constraint directly.
func TestConstConstraint(t *testing.T) {
	tests := []struct {
		name       string
		constValue string
		input      any
		wantErr    bool
	}{
		// String tests
		{"string exact match", "hello", "hello", false},
		{"string mismatch", "hello", "world", true},
		{"string empty mismatch", "hello", "", true},
		// Integer tests
		{"int exact match", "42", 42, false},
		{"int mismatch", "42", 43, true},
		{"int8 exact match", "42", int8(42), false},
		{"int64 exact match", "42", int64(42), false},
		// Uint tests
		{"uint exact match", "100", uint(100), false},
		{"uint8 exact match", "100", uint8(100), false},
		// Float tests
		{"float exact match", "3.14", 3.14, false},
		{"float32 exact match", "3.140000104904175", float32(3.14), false}, // float32 to string has precision loss
		// Bool tests
		{"bool true match", "true", true, false},
		{"bool false match", "false", false, false},
		{"bool mismatch", "true", false, true},
		// Nil tests
		{"nil pointer skips", "test", (*string)(nil), false},
		// Pointer tests
		{"pointer string match", "hello", ptrTo("hello"), false},
		{"pointer string mismatch", "hello", ptrTo("world"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := constConstraint{value: tt.constValue}
			err := c.Validate(tt.input)
			checkConstraintError(t, err, tt.wantErr)
		})
	}
}

// TestConstConstraintUnsupportedType tests const with unsupported types.
func TestConstConstraintUnsupportedType(t *testing.T) {
	c := constConstraint{value: "test"}
	err := c.Validate(struct{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// TestBuildConstConstraint tests buildConstConstraint function.
func TestBuildConstConstraint(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantOk  bool
		wantVal string
	}{
		{"valid value", "hello", true, "hello"},
		{"numeric value", "42", true, "42"},
		{"empty value returns false", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, ok := buildConstConstraint(tt.value)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				cc := c.(constConstraint)
				assert.Equal(t, tt.wantVal, cc.value)
			}
		})
	}
}

// ptrTo returns a pointer to the given value.
func ptrTo[T any](v T) *T {
	return &v
}
