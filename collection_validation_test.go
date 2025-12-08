package pedantigo

import (
	"testing"
)

// ==================================================
// slice element validation tests
// ==================================================

func TestSlice_ValidEmails(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","bob@example.com"]}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid emails, got %v", err)
	}

	if len(config.Admins) != 2 {
		t.Errorf("expected 2 admins, got %d", len(config.Admins))
	}
}

func TestSlice_InvalidEmail_SingleElement(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["not-an-email"]}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Error("expected validation error for invalid email in slice")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[0]" && fieldErr.Message == "must be a valid email address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Admins[0]', got %v", ve.Errors)
	}
}

func TestSlice_InvalidEmail_MultipleElements(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":["alice@example.com","invalid","bob@example.com","also-invalid"]}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check first error at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[1]" && fieldErr.Message == "must be a valid email address" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Admins[1]', got %v", ve.Errors)
	}

	// Check second error at index 3
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Admins[3]" && fieldErr.Message == "must be a valid email address" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Admins[3]', got %v", ve.Errors)
	}
}

func TestSlice_MinLength(t *testing.T) {
	type User struct {
		Tags []string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[User]()
	jsonData := []byte(`{"tags":["abc","de","fgh"]}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 1 {
		t.Errorf("expected 1 validation error, got %d: %v", len(ve.Errors), ve.Errors)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Tags[1]" && fieldErr.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Tags[1]', got %v", ve.Errors)
	}
}

func TestSlice_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type User struct {
		Addresses []Address `json:"addresses"`
	}

	validator := New[User]()
	jsonData := []byte(`{"addresses":[{"city":"NYC","zip":"10001"},{"zip":"123"}]}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check for missing city at index 1
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].City" && fieldErr.Message == "is required" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Addresses[1].City', got %v", ve.Errors)
	}

	// Check for short zip at index 1
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Addresses[1].Zip" && fieldErr.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Addresses[1].Zip', got %v", ve.Errors)
	}
}

func TestSlice_EmptySlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":[]}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty slice, got %v", err)
	}

	if len(config.Admins) != 0 {
		t.Errorf("expected empty admins slice, got %d elements", len(config.Admins))
	}
}

func TestSlice_NilSlice(t *testing.T) {
	type Config struct {
		Admins []string `json:"admins" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"admins":null}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil slice, got %v", err)
	}

	if config.Admins != nil {
		t.Errorf("expected nil admins slice, got %v", config.Admins)
	}
}
