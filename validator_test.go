package pedantigo

import (
	"testing"
)

// NOTE: 'required' is only checked during Unmarshal (missing JSON keys), not Validate()
// Validate() only checks value constraints (min, max, email, etc.)

func TestValidator_Required_Present(t *testing.T) {
	type User struct {
		Email string `pedantigo:"required"`
	}

	validator := New[User]()
	user := &User{Email: "test@example.com"}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_Min_BelowMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 15}

	err := validator.Validate(user)
	if err == nil {
		t.Error("expected validation error for value below minimum")
	}
}

func TestValidator_Min_AtMinimum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{Age: 18}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_Max_AboveMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 150}

	err := validator.Validate(user)
	if err == nil {
		t.Error("expected validation error for value above maximum")
	}
}

func TestValidator_Max_AtMaximum(t *testing.T) {
	type User struct {
		Age int `pedantigo:"max=120"`
	}

	validator := New[User]()
	user := &User{Age: 120}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

func TestValidator_MinMax_InRange(t *testing.T) {
	type User struct {
		Age int `pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{Age: 25}

	err := validator.Validate(user)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}
}

// Test type for cross-field validation
type testPasswordChange struct {
	Password string `pedantigo:"required"`
	Confirm  string `pedantigo:"required"`
}

func (vpc *testPasswordChange) Validate() error {
	if vpc.Password != vpc.Confirm {
		return &ValidationError{
			Errors: []FieldError{{
				Field:   "Confirm",
				Message: "passwords do not match",
			}},
		}
	}
	return nil
}

func TestValidator_CrossField_PasswordConfirmation(t *testing.T) {
	validator := New[testPasswordChange]()
	pwd := &testPasswordChange{
		Password: "secret123",
		Confirm:  "different",
	}

	err := validator.Validate(pwd)
	if err == nil {
		t.Error("expected validation error for password mismatch")
	}

	// Should have cross-field error
	foundCrossFieldError := false
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Confirm" && fieldErr.Message == "passwords do not match" {
			foundCrossFieldError = true
		}
	}

	if !foundCrossFieldError {
		t.Error("expected cross-field validation error")
	}
}
