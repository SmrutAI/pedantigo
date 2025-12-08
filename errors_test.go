package pedantigo

import (
	"testing"
)

func TestFieldError_Struct(t *testing.T) {
	err := FieldError{
		Field:   "Email",
		Message: "is required",
		Value:   "test@example.com",
	}

	if err.Field != "Email" {
		t.Errorf("expected field 'Email', got %q", err.Field)
	}

	if err.Message != "is required" {
		t.Errorf("expected message 'is required', got %q", err.Message)
	}

	if err.Value != "test@example.com" {
		t.Errorf("expected value 'test@example.com', got %v", err.Value)
	}
}

func TestValidationError_Error_Empty(t *testing.T) {
	err := &ValidationError{
		Errors: []FieldError{},
	}

	expected := "validation failed"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationError_Error_Single(t *testing.T) {
	err := &ValidationError{
		Errors: []FieldError{
			{Field: "Email", Message: "is required"},
		},
	}

	expected := "Email: is required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationError_Error_Multiple(t *testing.T) {
	err := &ValidationError{
		Errors: []FieldError{
			{Field: "Email", Message: "is required"},
			{Field: "Age", Message: "must be at least 18"},
		},
	}

	expected := "Email: is required (and 1 more errors)"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationError_Error_MultipleMany(t *testing.T) {
	err := &ValidationError{
		Errors: []FieldError{
			{Field: "Email", Message: "is required"},
			{Field: "Age", Message: "must be at least 18"},
			{Field: "Name", Message: "too short"},
		},
	}

	expected := "Email: is required (and 2 more errors)"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestValidationError_AccessErrors(t *testing.T) {
	err := &ValidationError{
		Errors: []FieldError{
			{Field: "Email", Message: "is required"},
			{Field: "Age", Message: "must be at least 18"},
		},
	}

	if len(err.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(err.Errors))
	}

	if err.Errors[0].Field != "Email" {
		t.Errorf("expected first error field 'Email', got %q", err.Errors[0].Field)
	}

	if err.Errors[0].Message != "is required" {
		t.Errorf("expected first error message 'is required', got %q", err.Errors[0].Message)
	}

	if err.Errors[1].Field != "Age" {
		t.Errorf("expected second error field 'Age', got %q", err.Errors[1].Field)
	}

	if err.Errors[1].Message != "must be at least 18" {
		t.Errorf("expected second error message 'must be at least 18', got %q", err.Errors[1].Message)
	}
}
