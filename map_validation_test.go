package pedantigo

import (
	"testing"
)

// ==================================================
// map value validation tests
// ==================================================

func TestMap_ValidEmails(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"bob@example.com"}}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid emails, got %v", err)
	}

	if len(config.Contacts) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(config.Contacts))
	}
}

func TestMap_InvalidEmail_SingleValue(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"not-an-email"}}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for invalid email in map")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Contacts[admin]" && fieldErr.Message == "must be a valid email address" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Contacts[admin]', got %v", ve.Errors)
	}
}

func TestMap_InvalidEmail_MultipleValues(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{"admin":"alice@example.com","support":"invalid","billing":"bob@example.com","sales":"also-invalid"}}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check that we have errors for the invalid keys (exact keys may vary due to map iteration order)
	invalidKeys := map[string]bool{"support": false, "sales": false}
	for _, fieldErr := range ve.Errors {
		if fieldErr.Message == "must be a valid email address" {
			switch fieldErr.Field {
			case "Contacts[support]":
				invalidKeys["support"] = true
			case "Contacts[sales]":
				invalidKeys["sales"] = true
			}
		}
	}

	if !invalidKeys["support"] {
		t.Errorf("expected error at 'Contacts[support]', got %v", ve.Errors)
	}
	if !invalidKeys["sales"] {
		t.Errorf("expected error at 'Contacts[sales]', got %v", ve.Errors)
	}
}

func TestMap_MinLength(t *testing.T) {
	type Config struct {
		Tags map[string]string `json:"tags" pedantigo:"min=3"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"tags":{"category":"abc","type":"de","status":"fgh"}}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error")
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
		if fieldErr.Field == "Tags[type]" && fieldErr.Message == "must be at least 3 characters" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Tags[type]', got %v", ve.Errors)
	}
}

func TestMap_NestedStructValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required"`
		Zip  string `json:"zip" pedantigo:"min=5"`
	}

	type Company struct {
		Offices map[string]Address `json:"offices"`
	}

	validator := New[Company]()
	jsonData := []byte(`{"offices":{"hq":{"city":"NYC","zip":"10001"},"branch":{"zip":"123"}}}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check for missing city at branch office
	foundError1 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].City" && fieldErr.Message == "is required" {
			foundError1 = true
		}
	}
	if !foundError1 {
		t.Errorf("expected error at 'Offices[branch].City', got %v", ve.Errors)
	}

	// Check for short zip at branch office
	foundError2 := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Offices[branch].Zip" && fieldErr.Message == "must be at least 5 characters" {
			foundError2 = true
		}
	}
	if !foundError2 {
		t.Errorf("expected error at 'Offices[branch].Zip', got %v", ve.Errors)
	}
}

func TestMap_EmptyMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":{}}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for empty map, got %v", err)
	}

	if len(config.Contacts) != 0 {
		t.Errorf("expected empty contacts map, got %d elements", len(config.Contacts))
	}
}

func TestMap_NilMap(t *testing.T) {
	type Config struct {
		Contacts map[string]string `json:"contacts" pedantigo:"email"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"contacts":null}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for nil map, got %v", err)
	}

	if config.Contacts != nil {
		t.Errorf("expected nil contacts map, got %v", config.Contacts)
	}
}
