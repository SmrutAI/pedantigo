package validation

import (
	"fmt"
	"reflect"

	"github.com/SmrutAI/Pedantigo/internal/tags"
)

// FieldError represents a validation error for a specific field
// This is a copy of the root package's FieldError to avoid circular imports
// FieldError represents an error condition.
type FieldError struct {
	Field   string
	Message string
	Value   any
}

// ConstraintValidator is the interface for validation constraints.
type ConstraintValidator interface {
	Validate(value any) error
}

// TagParser is a function type for parsing struct tags with dive support.
type TagParser func(tag reflect.StructTag) *tags.ParsedTag

// ConstraintBuilder is a function type for building constraint validators.
type ConstraintBuilder func(constraints map[string]string, fieldType reflect.Type) []ConstraintValidator

// ValidateValue recursively validates a reflected value with dive support.
// Uses ParseTagWithDive to handle collection-level and element-level constraints.
// NOTE: 'required' constraint is skipped (not built in BuildConstraints).
func ValidateValue(
	val reflect.Value,
	path string,
	strictMissingFields bool,
	parseTagFunc TagParser,
	buildConstraintsFunc ConstraintBuilder,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var errors []FieldError

	// Handle pointer indirection
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return errors // nil pointers are handled by required constraint
		}
		val = val.Elem()
	}

	// Only validate structs
	if val.Kind() != reflect.Struct {
		return errors
	}

	typ := val.Type()

	// Iterate through all fields
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Build field path
		fieldPath := field.Name
		if path != "" {
			fieldPath = path + "." + field.Name
		}

		// Parse validation tags with dive support
		parsedTag := parseTagFunc(field.Tag)
		if parsedTag == nil {
			// No validation tags, but still check nested structs, slices, and maps
			nestedErrors := validateNestedElements(fieldValue, recursiveValidateFunc, fieldPath)
			errors = append(errors, nestedErrors...)
			continue
		}

		// For nested structs (path != ""), check required fields
		if path != "" && strictMissingFields {
			// Check if "required" is in CollectionConstraints (before dive)
			if _, hasRequired := parsedTag.CollectionConstraints["required"]; hasRequired {
				// Check if field is zero value (indicates it was missing from JSON)
				if fieldValue.IsZero() {
					errors = append(errors, FieldError{
						Field:   fieldPath,
						Message: "is required",
						Value:   fieldValue.Interface(),
					})
					// Skip further validation for this field
					continue
				}
			}
		}

		// Validate based on field kind
		isCollection := fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Map
		isMap := fieldValue.Kind() == reflect.Map

		// Panic checks for invalid tag combinations
		if parsedTag.DivePresent && !isCollection {
			panic(fmt.Sprintf("field %s.%s: 'dive' can only be used on slice or map types, got %s",
				typ.Name(), field.Name, fieldValue.Kind()))
		}

		if len(parsedTag.KeyConstraints) > 0 && !isMap {
			panic(fmt.Sprintf("field %s.%s: 'keys' can only be used on map types, got %s",
				typ.Name(), field.Name, fieldValue.Kind()))
		}

		// Validate based on field type
		if isCollection {
			// Always apply collection-level constraints first (constraints before dive)
			if len(parsedTag.CollectionConstraints) > 0 {
				collectionValidators := buildConstraintsFunc(parsedTag.CollectionConstraints, field.Type)
				errors = append(errors, validateScalarField(fieldValue, fieldPath, collectionValidators, recursiveValidateFunc)...)
			}

			if parsedTag.DivePresent {
				// Dive into collection: validate elements with ElementConstraints
				elementValidators := buildConstraintsFunc(parsedTag.ElementConstraints, field.Type.Elem())
				if isMap {
					// Map with dive support
					errors = append(errors, validateMapElementsWithDive(
						fieldValue, fieldPath, elementValidators, parsedTag.KeyConstraints,
						buildConstraintsFunc, field.Type.Key(), recursiveValidateFunc)...)
				} else {
					// Slice with dive support
					errors = append(errors, validateSliceElements(fieldValue, fieldPath, elementValidators, recursiveValidateFunc)...)
				}
			} else {
				// No dive: still check nested structs in the collection
				nestedErrors := validateNestedElements(fieldValue, recursiveValidateFunc, fieldPath)
				errors = append(errors, nestedErrors...)
			}
		} else {
			// Non-collection field: apply constraints directly
			validators := buildConstraintsFunc(parsedTag.CollectionConstraints, field.Type)
			errors = append(errors, validateScalarField(fieldValue, fieldPath, validators, recursiveValidateFunc)...)
		}
	}

	return errors
}

func validateNestedElements(fieldValue reflect.Value,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
	fieldPath string,
) []FieldError {
	fieldErrors := make([]FieldError, 0)

	switch fieldValue.Kind() {
	case reflect.Struct:
		fieldErrors = append(fieldErrors, recursiveValidateFunc(fieldValue, fieldPath)...)
	case reflect.Slice:
		// Recursively validate struct elements in slices.
		for i := 0; i < fieldValue.Len(); i++ {
			elemValue := fieldValue.Index(i)
			elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
			if elemValue.Kind() == reflect.Struct {
				fieldErrors = append(fieldErrors, recursiveValidateFunc(elemValue, elemPath)...)
			}
		}
	case reflect.Map:
		// Recursively validate struct values in maps.
		iter := fieldValue.MapRange()
		for iter.Next() {
			mapKey := iter.Key()
			mapValue := iter.Value()
			mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())
			if mapValue.Kind() == reflect.Struct {
				fieldErrors = append(fieldErrors, recursiveValidateFunc(mapValue, mapPath)...)
			}
		}
	}

	return fieldErrors
}

func validateSliceElements(
	fieldValue reflect.Value,
	fieldPath string,
	validators []ConstraintValidator,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var errors []FieldError
	for i := 0; i < fieldValue.Len(); i++ {
		elemValue := fieldValue.Index(i)
		elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		// Apply constraints to each element
		for _, validator := range validators {
			if err := validator.Validate(elemValue.Interface()); err != nil {
				errors = append(errors, FieldError{
					Field:   elemPath,
					Message: err.Error(),
					Value:   elemValue.Interface(),
				})
			}
		}

		// Recursively validate nested structs in slice
		if elemValue.Kind() == reflect.Struct {
			errors = append(errors, recursiveValidateFunc(elemValue, elemPath)...)
		}
	}
	return errors
}

func validateMapElementsWithDive(
	fieldValue reflect.Value,
	fieldPath string,
	elementValidators []ConstraintValidator,
	keyConstraints map[string]string,
	buildConstraintsFunc ConstraintBuilder,
	keyType reflect.Type,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var errors []FieldError

	// Build key validators
	keyValidators := buildConstraintsFunc(keyConstraints, keyType)

	iter := fieldValue.MapRange()
	for iter.Next() {
		mapKey := iter.Key()
		mapValue := iter.Value()
		mapPath := fmt.Sprintf("%s[%v]", fieldPath, mapKey.Interface())

		// Validate keys if key constraints exist
		for _, validator := range keyValidators {
			if err := validator.Validate(mapKey.Interface()); err != nil {
				errors = append(errors, FieldError{
					Field:   mapPath,
					Message: err.Error(),
					Value:   mapKey.Interface(),
				})
			}
		}

		// Validate values
		for _, validator := range elementValidators {
			if err := validator.Validate(mapValue.Interface()); err != nil {
				errors = append(errors, FieldError{
					Field:   mapPath,
					Message: err.Error(),
					Value:   mapValue.Interface(),
				})
			}
		}

		// Recursively validate nested structs in map
		if mapValue.Kind() == reflect.Struct {
			errors = append(errors, recursiveValidateFunc(mapValue, mapPath)...)
		}
	}
	return errors
}

func validateScalarField(
	fieldValue reflect.Value,
	fieldPath string,
	validators []ConstraintValidator,
	recursiveValidateFunc func(val reflect.Value, path string) []FieldError,
) []FieldError {
	var errors []FieldError

	// Apply constraints directly
	for _, validator := range validators {
		if err := validator.Validate(fieldValue.Interface()); err != nil {
			errors = append(errors, FieldError{
				Field:   fieldPath,
				Message: err.Error(),
				Value:   fieldValue.Interface(),
			})
		}
	}

	// Recursively validate nested structs
	if fieldValue.Kind() == reflect.Struct {
		errors = append(errors, recursiveValidateFunc(fieldValue, fieldPath)...)
	}

	return errors
}
