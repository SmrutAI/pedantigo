package pedantigo

import (
	"strings"
	"testing"
)

// TestStrictMissingFields_RequiredIgnored_HappyPath verifies that when StrictMissingFields is false,
// missing required fields do NOT cause "is required" errors, but validators still run on provided values
func TestStrictMissingFields_RequiredIgnored_HappyPath(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"required,min=2"`
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"required,min=18"`
	}

	// Create validator with StrictMissingFields disabled
	validator := New[User](ValidatorOptions{
		StrictMissingFields: false,
	})

	// Provide valid values (even though fields are marked required)
	jsonData := []byte(`{"name":"John","email":"john@example.com","age":25}`)

	user, err := validator.Unmarshal(jsonData)

	// Should have NO errors
	if err != nil {
		t.Errorf("expected no errors with valid values, got %v", err)
	}

	// Verify values are correctly parsed
	if user.Name != "John" {
		t.Errorf("expected Name='John', got %q", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got %q", user.Email)
	}
	if user.Age != 25 {
		t.Errorf("expected Age=25, got %d", user.Age)
	}
}

// TestStrictMissingFields_RequiredIgnored_ZeroValuesValidated verifies that when StrictMissingFields is false,
// missing fields get zero values and validators STILL RUN on those zero values (may cause validation errors)
func TestStrictMissingFields_RequiredIgnored_ZeroValuesValidated(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required,min=2"`
		Age  int    `json:"age" pedantigo:"required,min=18"`
	}

	// Create validator with StrictMissingFields disabled
	validator := New[User](ValidatorOptions{
		StrictMissingFields: false,
	})

	// JSON with all fields missing
	jsonData := []byte(`{}`)

	_, err := validator.Unmarshal(jsonData)

	// Should have validation errors because zero values fail min constraints
	// Name="" fails min=2, Age=0 fails min=18
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 2 {
		t.Errorf("expected 2 validation errors (Name min, Age min), got %d: %v", len(ve.Errors), ve.Errors)
	}

	hasNameError := false
	hasAgeError := false
	for _, fieldErr := range ve.Errors {
		// Should NOT have "is required" errors
		if strings.Contains(strings.ToLower(fieldErr.Message), "required") {
			t.Errorf("should not have 'required' error, got: %s", fieldErr.Message)
		}
		// Should have min constraint errors
		if fieldErr.Field == "Name" && strings.Contains(fieldErr.Message, "at least") {
			hasNameError = true
		}
		if fieldErr.Field == "Age" && strings.Contains(fieldErr.Message, "at least") {
			hasAgeError = true
		}
	}

	if !hasNameError {
		t.Error("expected min constraint error for Name")
	}
	if !hasAgeError {
		t.Error("expected min constraint error for Age")
	}
}

// TestStrictMissingFields_DefaultIgnored verifies that when StrictMissingFields is false,
// default values are NOT applied (stays as zero values)
func TestStrictMissingFields_DefaultIgnored(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"default=unnamed"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Enabled bool   `json:"enabled" pedantigo:"default=true"`
	}

	// This should panic because we're using default tags with StrictMissingFields=false
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when using default tags with StrictMissingFields=false, but didn't panic")
		} else {
			panicMsg := r.(string)
			if !strings.Contains(panicMsg, "default=") || !strings.Contains(panicMsg, "StrictMissingFields is false") {
				t.Errorf("panic message should mention default tag and StrictMissingFields, got: %s", panicMsg)
			}
		}
	}()

	// This should panic during New() due to safety check
	_ = New[Config](ValidatorOptions{
		StrictMissingFields: false,
	})
}

// TestStrictMissingFields_PanicOnDefaultTag verifies that creating a validator
// with default= tags panics when StrictMissingFields is false
func TestStrictMissingFields_PanicOnDefaultTag(t *testing.T) {
	type Settings struct {
		Theme    string `json:"theme" pedantigo:"default=dark"`
		Language string `json:"language"`
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when using default= tag with StrictMissingFields=false")
		} else {
			panicMsg := r.(string)
			// Verify panic message mentions the specific field and issue
			if !strings.Contains(panicMsg, "Theme") {
				t.Errorf("panic message should mention field 'Theme', got: %s", panicMsg)
			}
			if !strings.Contains(panicMsg, "default=") {
				t.Errorf("panic message should mention 'default=' tag, got: %s", panicMsg)
			}
			if !strings.Contains(panicMsg, "StrictMissingFields is false") {
				t.Errorf("panic message should mention 'StrictMissingFields is false', got: %s", panicMsg)
			}
		}
	}()

	_ = New[Settings](ValidatorOptions{
		StrictMissingFields: false,
	})
}

// TestStrictMissingFields_PanicOnDefaultUsingMethod verifies that creating a validator
// with defaultUsingMethod= tags panics when StrictMissingFields is false
func TestStrictMissingFields_PanicOnDefaultUsingMethod(t *testing.T) {
	type Product struct {
		ID   string `json:"id" pedantigo:"defaultUsingMethod=GenerateID"`
		Name string `json:"name"`
	}

	// Add the method to satisfy method validation (if we get that far)
	// Note: The panic should happen before method validation

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when using defaultUsingMethod= tag with StrictMissingFields=false")
		} else {
			panicMsg := r.(string)
			// Verify panic message mentions the specific field and issue
			if !strings.Contains(panicMsg, "ID") {
				t.Errorf("panic message should mention field 'ID', got: %s", panicMsg)
			}
			if !strings.Contains(panicMsg, "defaultUsingMethod=") {
				t.Errorf("panic message should mention 'defaultUsingMethod=' tag, got: %s", panicMsg)
			}
			if !strings.Contains(panicMsg, "StrictMissingFields is false") {
				t.Errorf("panic message should mention 'StrictMissingFields is false', got: %s", panicMsg)
			}
		}
	}()

	_ = New[Product](ValidatorOptions{
		StrictMissingFields: false,
	})
}

// TestStrictMissingFields_ValidatorsStillRun verifies that even with StrictMissingFields=false,
// validation constraints (min, max, email, etc.) still run on provided values
func TestStrictMissingFields_ValidatorsStillRun(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User](ValidatorOptions{
		StrictMissingFields: false,
	})

	// Provide invalid values
	jsonData := []byte(`{"email":"notanemail","age":15}`)

	_, err := validator.Unmarshal(jsonData)

	// Should have validation errors for invalid email and age
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Check that we have email and age errors
	hasEmailError := false
	hasAgeError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Email" {
			hasEmailError = true
		}
		if fieldErr.Field == "Age" {
			hasAgeError = true
		}
	}

	if !hasEmailError {
		t.Errorf("expected email validation error, got: %v", ve.Errors)
	}
	if !hasAgeError {
		t.Errorf("expected age validation error, got: %v", ve.Errors)
	}
}

// TestStrictMissingFields_PointerApproach verifies that using pointers (*int, *bool) works correctly
// with StrictMissingFields=false - nil pointers skip validation
func TestStrictMissingFields_PointerApproach(t *testing.T) {
	type Settings struct {
		Port    *int   `json:"port" pedantigo:"min=1024"` // Optional with pointer
		Enabled *bool  `json:"enabled"`                   // Optional with pointer
		Name    string `json:"name" pedantigo:"min=3"`    // Non-optional, will validate zero value
	}

	validator := New[Settings](ValidatorOptions{
		StrictMissingFields: false,
	})

	// JSON with all fields missing
	jsonData := []byte(`{}`)

	settings, err := validator.Unmarshal(jsonData)

	// Should only have error for Name (zero value "" fails min=3)
	// Port and Enabled are nil pointers, so validation is skipped
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 1 {
		t.Errorf("expected 1 error (Name min), got %d errors: %v", len(ve.Errors), ve.Errors)
	}

	hasNameError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Name" && strings.Contains(fieldErr.Message, "at least") {
			hasNameError = true
		}
	}

	if !hasNameError {
		t.Errorf("expected Name min error, got: %v", ve.Errors)
	}

	// Verify pointer fields are nil
	if settings.Port != nil {
		t.Errorf("expected Port to be nil, got %v", *settings.Port)
	}
	if settings.Enabled != nil {
		t.Errorf("expected Enabled to be nil, got %v", *settings.Enabled)
	}
	if settings.Name != "" {
		t.Errorf("expected Name to be empty string, got %q", settings.Name)
	}
}

// TestStrictMissingFields_PointerApproach_WithValues verifies pointers work correctly with provided values
func TestStrictMissingFields_PointerApproach_WithValues(t *testing.T) {
	type Settings struct {
		Port    *int  `json:"port" pedantigo:"min=1024"`
		Enabled *bool `json:"enabled"`
	}

	validator := New[Settings](ValidatorOptions{
		StrictMissingFields: false,
	})

	// Provide valid values
	jsonData := []byte(`{"port":8080,"enabled":true}`)

	settings, err := validator.Unmarshal(jsonData)

	if err != nil {
		t.Errorf("expected no errors with valid values, got: %v", err)
	}

	// Verify values are correctly parsed
	if settings.Port == nil || *settings.Port != 8080 {
		t.Errorf("expected Port=8080, got %v", settings.Port)
	}
	if settings.Enabled == nil || *settings.Enabled != true {
		t.Errorf("expected Enabled=true, got %v", settings.Enabled)
	}
}

// TestStrictMissingFields_PointerApproach_InvalidValues verifies pointers with constraints still validate
func TestStrictMissingFields_PointerApproach_InvalidValues(t *testing.T) {
	type Settings struct {
		Port *int `json:"port" pedantigo:"min=1024"`
	}

	validator := New[Settings](ValidatorOptions{
		StrictMissingFields: false,
	})

	// Provide invalid value (port too low)
	jsonData := []byte(`{"port":80}`)

	_, err := validator.Unmarshal(jsonData)

	// Should have validation error for port
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) != 1 {
		t.Errorf("expected 1 error (Port min), got %d: %v", len(ve.Errors), ve.Errors)
	}

	hasPortError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Port" && strings.Contains(fieldErr.Message, "at least") {
			hasPortError = true
		}
	}

	if !hasPortError {
		t.Errorf("expected Port min error, got: %v", ve.Errors)
	}
}

// TestStrictMissingFields_DefaultBehavior verifies that the default behavior
// (StrictMissingFields=true) still works correctly
func TestStrictMissingFields_DefaultBehavior(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"required"`
		Age  int    `json:"age" pedantigo:"default=25"`
	}

	// Create validator with default options (StrictMissingFields=true)
	validator := New[User]()

	// JSON with missing required field
	jsonData := []byte(`{}`)

	user, err := validator.Unmarshal(jsonData)

	// Should have error for missing required field
	hasRequiredError := false
	if err != nil {
		ve, ok := err.(*ValidationError)
		if ok {
			for _, fieldErr := range ve.Errors {
				if fieldErr.Field == "name" && strings.Contains(strings.ToLower(fieldErr.Message), "required") {
					hasRequiredError = true
				}
			}
		}
	}

	if !hasRequiredError {
		t.Error("expected required field error when StrictMissingFields=true")
	}

	// Age should have default value applied
	if user.Age != 25 {
		t.Errorf("expected default value 25 for Age, got %d", user.Age)
	}
}
