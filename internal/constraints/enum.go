// Package constraints provides validation constraint types and builders for pedantigo.
package constraints

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Enum constraint types.
type (
	enumConstraint    struct{ values []string }
	defaultConstraint struct{ value string }
)

// enumConstraint validates that value is one of the allowed values.
func (c enumConstraint) Validate(value any) error {
	v, ok := derefValue(value)
	if !ok {
		return nil // Skip validation for invalid/nil values
	}

	// Convert value to string for comparison
	var str string
	switch v.Kind() {
	case reflect.String:
		str = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		str = strconv.FormatBool(v.Bool())
	default:
		return fmt.Errorf("enum constraint not supported for type %s", v.Kind())
	}

	// Check if value is in allowed list
	for _, allowed := range c.values {
		if str == allowed {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", strings.Join(c.values, ", "))
}

// defaultConstraint is not a validator - it's handled during unmarshaling.
func (c defaultConstraint) Validate(value any) error {
	return nil // No-op for validation
}

// buildEnumConstraint parses space-separated enum values.
func buildEnumConstraint(value string) Constraint {
	values := strings.Fields(value)
	return enumConstraint{values: values}
}
