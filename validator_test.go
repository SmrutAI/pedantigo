package pedantigo

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ==================== Core Validation Tests ====================
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

// TestMarshal_Valid verifies that Marshal returns JSON for valid structs
func TestMarshal_Valid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18,max=120"`
	}

	validator := New[User]()
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	data, err := validator.Marshal(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify JSON is valid and contains expected fields
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Marshal returned invalid JSON: %v", err)
	}

	if result["name"] != "John Doe" {
		t.Errorf("expected name='John Doe', got %v", result["name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("expected email='john@example.com', got %v", result["email"])
	}
	if result["age"] != float64(25) {
		t.Errorf("expected age=25, got %v", result["age"])
	}
}

// TestMarshal_Invalid verifies that Marshal returns validation errors for invalid structs
func TestMarshal_Invalid(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"min=2"`
		Email string `json:"email" pedantigo:"email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	user := &User{
		Name:  "J",          // Too short (min=2)
		Email: "notanemail", // Invalid email
		Age:   15,           // Too young (min=18)
	}

	data, err := validator.Marshal(user)

	// Should return validation error, not JSON
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if data != nil {
		t.Errorf("expected nil data when validation fails, got %d bytes", len(data))
	}

	// Verify it's a ValidationError with multiple field errors
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(ve.Errors) != 3 {
		t.Errorf("expected 3 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}

	// Check that errors are for the expected fields
	errorFields := make(map[string]bool)
	for _, fieldErr := range ve.Errors {
		errorFields[fieldErr.Field] = true
	}

	if !errorFields["Name"] {
		t.Error("expected validation error for Name field")
	}
	if !errorFields["Email"] {
		t.Error("expected validation error for Email field")
	}
	if !errorFields["Age"] {
		t.Error("expected validation error for Age field")
	}
}

// TestMarshal_Nil verifies that Marshal handles nil pointer appropriately
func TestMarshal_Nil(t *testing.T) {
	type User struct {
		Name string `json:"name" pedantigo:"min=2"`
	}

	validator := New[User]()

	// Pass nil pointer
	data, err := validator.Marshal(nil)

	// Should handle nil gracefully (either return error or marshal "null")
	// Let's check what actually happens - if Validate accepts nil, json.Marshal will return "null"
	// If Validate rejects nil, we'll get a validation error
	if err != nil {
		// Validation error is acceptable for nil
		t.Logf("Marshal(nil) returned error: %v", err)
	} else if string(data) != "null" {
		t.Errorf("expected Marshal(nil) to return 'null', got %q", string(data))
	}
}

// ==================== Unmarshal Tests ====================

func TestUnmarshal_ValidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", user.Email)
	}

	if user.Age != 25 {
		t.Errorf("expected age 25, got %d", user.Age)
	}
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":}`) // Invalid JSON

	user, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Error("expected JSON decode error")
	}

	// Should return nil user on JSON decode errors
	if user != nil {
		t.Error("expected nil user on JSON decode error")
	}
}

func TestUnmarshal_ValidationError(t *testing.T) {
	type User struct {
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"min=18"`
	}

	validator := New[User]()
	// email is present but invalid (not an email), age is below min
	jsonData := []byte(`{"email":"notanemail","age":15}`)

	user, err := validator.Unmarshal(jsonData)

	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	t.Logf("Got %d errors:", len(ve.Errors))
	for _, fieldErr := range ve.Errors {
		t.Logf("  - %s: %s", fieldErr.Field, fieldErr.Message)
	}

	// Should still return the user struct even with validation errors
	if user == nil {
		t.Error("expected non-nil user even with validation errors")
	}

	// Check we have errors for both fields (use struct field names, not JSON names)
	foundEmailError := false
	foundAgeError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Email" {
			foundEmailError = true
		}
		if fieldErr.Field == "Age" {
			foundAgeError = true
		}
	}

	if !foundEmailError {
		t.Error("expected validation error for Email field")
	}

	if !foundAgeError {
		t.Error("expected validation error for Age field")
	}
}

func TestUnmarshal_DefaultValues(t *testing.T) {
	type User struct {
		Email  string `json:"email" pedantigo:"required"`
		Role   string `json:"role" pedantigo:"default=user"`
		Status string `json:"status" pedantigo:"default=active"`
	}

	validator := New[User]()
	jsonData := []byte(`{"email":"test@example.com"}`) // Missing role and status

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no validation errors, got %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	// Defaults should be applied
	if user.Role != "user" {
		t.Errorf("expected default role 'user', got %q", user.Role)
	}

	if user.Status != "active" {
		t.Errorf("expected default status 'active', got %q", user.Status)
	}
}

func TestUnmarshal_NestedValidation(t *testing.T) {
	type Address struct {
		City string `json:"city" pedantigo:"required,min=1"` // min=1 for non-empty string
	}

	type User struct {
		Email   string  `json:"email" pedantigo:"required"`
		Address Address `json:"address"`
	}

	validator := New[User]()
	// City is present but empty - should fail min=1 constraint
	jsonData := []byte(`{"email":"test@example.com","address":{"city":""}}`)

	user, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for empty city (min=1)")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Should have error for Address.City
	foundNestedError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Address.City" || fieldErr.Field == "City" {
			foundNestedError = true
		}
	}

	if !foundNestedError {
		t.Errorf("expected validation error for nested City field, got errors: %v", ve.Errors)
	}

	if user == nil {
		t.Error("expected non-nil user even with validation errors")
	}
}

// ==================== Pointer Tests ====================

// Test pointer field with explicit value
func TestPointer_ExplicitValue(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","age":25}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	if user.Name == nil {
		t.Fatal("expected non-nil Name pointer")
	}
	if *user.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", *user.Name)
	}

	if user.Age == nil {
		t.Fatal("expected non-nil Age pointer")
	}
	if *user.Age != 25 {
		t.Errorf("expected age 25, got %d", *user.Age)
	}
}

// Test pointer field with explicit zero value (should create pointer to zero)
func TestPointer_ExplicitZero(t *testing.T) {
	type Config struct {
		Port    *int    `json:"port"`
		Enabled *bool   `json:"enabled"`
		Name    *string `json:"name"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0,"enabled":false,"name":""}`)

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	// Explicit zeros should create pointers to zero values
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer for explicit 0")
	}
	if *config.Port != 0 {
		t.Errorf("expected port 0, got %d", *config.Port)
	}

	if config.Enabled == nil {
		t.Fatal("expected non-nil Enabled pointer for explicit false")
	}
	if *config.Enabled != false {
		t.Errorf("expected enabled false, got %v", *config.Enabled)
	}

	if config.Name == nil {
		t.Fatal("expected non-nil Name pointer for explicit empty string")
	}
	if *config.Name != "" {
		t.Errorf("expected name empty string, got %q", *config.Name)
	}
}

// Test pointer field with explicit null (should be nil pointer)
func TestPointer_ExplicitNull(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null,"age":null}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	// Explicit null should result in nil pointers
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for explicit null, got %v", *user.Name)
	}

	if user.Age != nil {
		t.Errorf("expected nil Age pointer for explicit null, got %v", *user.Age)
	}
}

// Test pointer field missing from JSON (should be nil pointer)
func TestPointer_Missing(t *testing.T) {
	type User struct {
		Name *string `json:"name"`
		Age  *int    `json:"age"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Both fields missing

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	// Missing fields should result in nil pointers
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for missing field, got %v", *user.Name)
	}

	if user.Age != nil {
		t.Errorf("expected nil Age pointer for missing field, got %v", *user.Age)
	}
}

// Test required pointer field with explicit value
func TestPointer_RequiredWithValue(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice"}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	if user.Name == nil {
		t.Fatal("expected non-nil Name pointer")
	}
	if *user.Name != "Alice" {
		t.Errorf("expected name 'Alice', got %q", *user.Name)
	}
}

// Test required pointer field missing (should fail)
func TestPointer_RequiredMissing(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{}`) // Missing required field

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	// Check for required field error
	foundRequiredError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "name" && fieldErr.Message == "is required" {
			foundRequiredError = true
		}
	}

	if !foundRequiredError {
		t.Errorf("expected 'is required' error for name field, got %v", ve.Errors)
	}
}

// Test required pointer field with explicit null (should pass - field is present)
func TestPointer_RequiredWithNull(t *testing.T) {
	type User struct {
		Name *string `json:"name" pedantigo:"required"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":null}`) // Field present but null

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors (field is present), got %v", err)
	}

	// Required means "field must be present", not "value can't be nil"
	if user.Name != nil {
		t.Errorf("expected nil Name pointer for explicit null, got %v", *user.Name)
	}
}

// Test pointer field with default value
func TestPointer_WithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{}`) // Missing port field

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	// Default should be applied to missing field
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer with default")
	}
	if *config.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", *config.Port)
	}
}

// Test pointer field with explicit zero and default (should keep zero)
func TestPointer_ExplicitZeroWithDefault(t *testing.T) {
	type Config struct {
		Port *int `json:"port" pedantigo:"default=8080"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"port":0}`) // Explicit zero

	config, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	// Explicit zero should be kept, not replaced with default
	if config.Port == nil {
		t.Fatal("expected non-nil Port pointer for explicit 0")
	}
	if *config.Port != 0 {
		t.Errorf("expected port 0 (not default), got %d", *config.Port)
	}
}

// Test nested struct with pointer fields
func TestPointer_NestedStruct(t *testing.T) {
	type Address struct {
		Street *string `json:"street"`
		City   *string `json:"city"`
	}

	type User struct {
		Name    string   `json:"name"`
		Address *Address `json:"address"`
	}

	validator := New[User]()
	jsonData := []byte(`{"name":"Alice","address":{"street":"123 Main St","city":null}}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors, got %v", err)
	}

	if user.Address == nil {
		t.Fatal("expected non-nil Address pointer")
	}

	if user.Address.Street == nil {
		t.Fatal("expected non-nil Street pointer")
	}
	if *user.Address.Street != "123 Main St" {
		t.Errorf("expected street '123 Main St', got %q", *user.Address.Street)
	}

	// City was explicitly null
	if user.Address.City != nil {
		t.Errorf("expected nil City pointer for explicit null, got %v", *user.Address.City)
	}
}

// ==================== Deserializer Tests ====================

// Test type for defaultUsingMethod
type UserWithTimestamp struct {
	Email     string    `json:"email" pedantigo:"required"`
	Role      string    `json:"role" pedantigo:"default=user"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=SetCreationTime"`
}

// Method that provides dynamic default value
func (u *UserWithTimestamp) SetCreationTime() (time.Time, error) {
	// Return a fixed time for testing (not time.Now())
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), nil
}

// Test type with invalid method signature (should panic at New() time)
type InvalidMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=WrongSignature"`
}

// Wrong signature: returns only value, no error
func (i *InvalidMethodType) WrongSignature() time.Time {
	return time.Now()
}

// Test type with non-existent method
type NonExistentMethodType struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" pedantigo:"defaultUsingMethod=DoesNotExist"`
}

// TestDeserializer_UnmarshalBehavior validates deserializer behavior across various scenarios:
// defaults, missing fields, explicit values, required fields, and validator options.
func TestDeserializer_UnmarshalBehavior(t *testing.T) {
	type Config struct {
		Name    string `json:"name" pedantigo:"required"`
		Port    int    `json:"port" pedantigo:"default=8080"`
		Timeout int    `json:"timeout" pedantigo:"default=30"`
	}

	type Settings struct {
		Name   string `json:"name" pedantigo:"required"`
		Active bool   `json:"active" pedantigo:"required"`
	}

	tests := []struct {
		name        string
		jsonData    []byte
		validatorFn func() (any, error)
		wantErr     bool
		assertions  func(*testing.T, any)
	}{
		{
			name:     "missing fields with defaults",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				if config.Port != 8080 {
					t.Errorf("expected default port 8080, got %d", config.Port)
				}
				if config.Timeout != 30 {
					t.Errorf("expected default timeout 30, got %d", config.Timeout)
				}
				if config.Name != "myapp" {
					t.Errorf("expected name 'myapp', got %q", config.Name)
				}
			},
		},
		{
			name:     "explicit zero values not replaced with defaults",
			jsonData: []byte(`{"name":"myapp","port":0,"timeout":0}`),
			validatorFn: func() (any, error) {
				v := New[Config]()
				return v.Unmarshal([]byte(`{"name":"myapp","port":0,"timeout":0}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				// Explicit zeros should be kept, NOT replaced with defaults
				if config.Port != 0 {
					t.Errorf("expected port 0 (not default), got %d", config.Port)
				}
				if config.Timeout != 0 {
					t.Errorf("expected timeout 0 (not default), got %d", config.Timeout)
				}
			},
		},
		{
			name:     "explicit false value passes required validation",
			jsonData: []byte(`{"name":"test","active":false}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test","active":false}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				settings := result.(*Settings)
				if settings.Active != false {
					t.Errorf("expected active=false, got %v", settings.Active)
				}
			},
		},
		{
			name:     "missing required field fails validation",
			jsonData: []byte(`{"name":"test"}`),
			validatorFn: func() (any, error) {
				v := New[Settings]()
				return v.Unmarshal([]byte(`{"name":"test"}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Error case - check error message through direct validation
				v := New[Settings]()
				_, err := v.Unmarshal([]byte(`{"name":"test"}`))
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}

				foundError := false
				for _, fieldErr := range ve.Errors {
					if fieldErr.Field == "active" && fieldErr.Message == "is required" {
						foundError = true
					}
				}
				if !foundError {
					t.Errorf("expected 'is required' error for field 'active', got errors: %+v", ve.Errors)
				}
			},
		},
		{
			name:     "defaultUsingMethod called for missing fields",
			jsonData: []byte(`{"email":"test@example.com"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				if !user.CreatedAt.Equal(expectedTime) {
					t.Errorf("expected created_at to be %v, got %v", expectedTime, user.CreatedAt)
				}
				if user.Role != "user" {
					t.Errorf("expected default role 'user', got %q", user.Role)
				}
			},
		},
		{
			name:     "defaultUsingMethod not called for explicit values",
			jsonData: []byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`),
			validatorFn: func() (any, error) {
				v := New[UserWithTimestamp]()
				return v.Unmarshal([]byte(`{"email":"test@example.com","created_at":"2023-06-15T12:30:00Z"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				user := result.(*UserWithTimestamp)
				explicitTime := time.Date(2023, 6, 15, 12, 30, 0, 0, time.UTC)
				if !user.CreatedAt.Equal(explicitTime) {
					t.Errorf("expected created_at to be %v (explicit), got %v", explicitTime, user.CreatedAt)
				}
			},
		},
		{
			name:     "strict mode applies defaults for missing fields",
			jsonData: []byte(`{"name":"myapp"}`),
			validatorFn: func() (any, error) {
				v := New[Config](ValidatorOptions{StrictMissingFields: true})
				return v.Unmarshal([]byte(`{"name":"myapp"}`))
			},
			wantErr: false,
			assertions: func(t *testing.T, result any) {
				config := result.(*Config)
				if config.Port != 8080 {
					t.Errorf("expected port 8080 (default applied), got %d", config.Port)
				}
				if config.Timeout != 30 {
					t.Errorf("expected timeout 30 (default applied), got %d", config.Timeout)
				}
			},
		},
		{
			name:     "relaxed mode skips constraints on zero values",
			jsonData: []byte(`{"name":"myapp","age":0}`),
			validatorFn: func() (any, error) {
				type Profile struct {
					Name string `json:"name" pedantigo:"required"`
					Age  int    `json:"age" pedantigo:"min=1"`
				}
				v := New[Profile](ValidatorOptions{StrictMissingFields: false})
				return v.Unmarshal([]byte(`{"name":"myapp","age":0}`))
			},
			wantErr: true,
			assertions: func(t *testing.T, result any) {
				// Age=0 is explicit, so validation should still run and fail
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.validatorFn()

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil || !tt.wantErr {
				tt.assertions(t, result)
			}
		})
	}
}

// TestDeserializer_ValidatorSetup validates fail-fast validation during New().
// Invalid method signatures or non-existent methods should panic at validator creation time.
func TestDeserializer_ValidatorSetup(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		expectPanic bool
	}{
		{
			name: "invalid method signature panics",
			setup: func() {
				_ = New[InvalidMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "non-existent method panics",
			setup: func() {
				_ = New[NonExistentMethodType]()
			},
			expectPanic: true,
		},
		{
			name: "valid method signature succeeds",
			setup: func() {
				_ = New[UserWithTimestamp]()
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic but none occurred")
					}
				}()
				tt.setup()
			} else {
				// Should not panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("unexpected panic: %v", r)
					}
				}()
				tt.setup()
			}
		})
	}
}

// ==================== Validator Options Tests ====================

// TestValidatorOptions_StrictMissingFields tests the StrictMissingFields behavior
// with various configuration combinations and JSON inputs.
func TestValidatorOptions_StrictMissingFields(t *testing.T) {
	type User struct {
		Name  string `json:"name" pedantigo:"required,min=2"`
		Email string `json:"email" pedantigo:"required,email"`
		Age   int    `json:"age" pedantigo:"required,min=18"`
	}

	tests := []struct {
		name            string
		strictMode      bool
		jsonInput       string
		expectErr       bool
		expectErrFields []string                       // Expected field names in ValidationError
		checkValues     func(t *testing.T, user *User) // Verify parsed values
	}{
		{
			name:       "StrictMissingFields_false_valid_values",
			strictMode: false,
			jsonInput:  `{"name":"John","email":"john@example.com","age":25}`,
			expectErr:  false,
			checkValues: func(t *testing.T, user *User) {
				if user.Name != "John" {
					t.Errorf("expected Name='John', got %q", user.Name)
				}
				if user.Email != "john@example.com" {
					t.Errorf("expected Email='john@example.com', got %q", user.Email)
				}
				if user.Age != 25 {
					t.Errorf("expected Age=25, got %d", user.Age)
				}
			},
		},
		{
			name:            "StrictMissingFields_false_zero_values_fail_min",
			strictMode:      false,
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"Name", "Age"},
			checkValues:     nil, // Error checking happens in test loop
		},
		{
			name:            "StrictMissingFields_false_invalid_email_and_age",
			strictMode:      false,
			jsonInput:       `{"email":"notanemail","age":15}`,
			expectErr:       true,
			expectErrFields: []string{"Name", "Email", "Age"}, // Name missing (zero value "") also fails min=2
		},
		{
			name:            "StrictMissingFields_true_required_field_missing",
			strictMode:      true,
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"name", "email", "age"}, // All required fields fail when missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[User](ValidatorOptions{
				StrictMissingFields: tt.strictMode,
			})
			user, err := validator.Unmarshal([]byte(tt.jsonInput))

			if (err != nil) != tt.expectErr {
				t.Errorf("expectErr=%v, got err=%v", tt.expectErr, err)
			}

			if err != nil && tt.expectErrFields != nil {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != len(tt.expectErrFields) {
					t.Errorf("expected %d errors, got %d: %v", len(tt.expectErrFields), len(ve.Errors), ve.Errors)
				}
			}

			if !tt.expectErr && tt.checkValues != nil {
				tt.checkValues(t, user)
			}
		})
	}
}

// TestValidatorOptions_PointerFields tests pointer field behavior with StrictMissingFields=false.
// Pointers to primitive types allow optional fields (nil when missing) while still validating when present.
func TestValidatorOptions_PointerFields(t *testing.T) {
	type Settings struct {
		Port    *int   `json:"port" pedantigo:"min=1024"`
		Enabled *bool  `json:"enabled"`
		Name    string `json:"name" pedantigo:"min=3"`
	}

	tests := []struct {
		name            string
		jsonInput       string
		expectErr       bool
		expectErrFields []string // Expected field names in ValidationError
		checkValues     func(t *testing.T, settings *Settings)
	}{
		{
			name:            "pointer_fields_all_missing",
			jsonInput:       `{}`,
			expectErr:       true,
			expectErrFields: []string{"Name"}, // Only Name should error (non-pointer zero value)
			checkValues: func(t *testing.T, settings *Settings) {
				// Port and Enabled should be nil (pointers skip validation when missing)
				if settings.Port != nil {
					t.Errorf("expected Port to be nil, got %v", *settings.Port)
				}
				if settings.Enabled != nil {
					t.Errorf("expected Enabled to be nil, got %v", *settings.Enabled)
				}
				// Name should have zero value ""
				if settings.Name != "" {
					t.Errorf("expected Name to be empty string, got %q", settings.Name)
				}
			},
		},
		{
			name:      "pointer_fields_with_valid_values",
			jsonInput: `{"port":8080,"enabled":true,"name":"John"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				if settings.Port == nil || *settings.Port != 8080 {
					t.Errorf("expected Port=8080, got %v", settings.Port)
				}
				if settings.Enabled == nil || *settings.Enabled != true {
					t.Errorf("expected Enabled=true, got %v", settings.Enabled)
				}
				if settings.Name != "John" {
					t.Errorf("expected Name='John', got %q", settings.Name)
				}
			},
		},
		{
			name:            "pointer_field_invalid_value",
			jsonInput:       `{"port":80}`,
			expectErr:       true,
			expectErrFields: []string{"Port", "Name"}, // Port too low, Name missing/empty
			checkValues: func(t *testing.T, settings *Settings) {
				// Pointer should still be set even with validation error
				if settings.Port == nil || *settings.Port != 80 {
					t.Errorf("expected Port=80 (even with error), got %v", settings.Port)
				}
			},
		},
		{
			name:      "pointer_fields_partial_missing",
			jsonInput: `{"port":2048,"name":"Alice"}`,
			expectErr: false,
			checkValues: func(t *testing.T, settings *Settings) {
				if settings.Port == nil || *settings.Port != 2048 {
					t.Errorf("expected Port=2048, got %v", settings.Port)
				}
				if settings.Enabled != nil {
					t.Errorf("expected Enabled to be nil, got %v", *settings.Enabled)
				}
				if settings.Name != "Alice" {
					t.Errorf("expected Name='Alice', got %q", settings.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New[Settings](ValidatorOptions{
				StrictMissingFields: false,
			})

			settings, err := validator.Unmarshal([]byte(tt.jsonInput))

			if (err != nil) != tt.expectErr {
				t.Errorf("expectErr=%v, got err=%v", tt.expectErr, err)
			}

			if err != nil && tt.expectErr {
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if len(ve.Errors) != len(tt.expectErrFields) {
					t.Errorf("expected %d errors, got %d: %v", len(tt.expectErrFields), len(ve.Errors), ve.Errors)
				}
			}

			if tt.checkValues != nil {
				tt.checkValues(t, settings)
			}
		})
	}
}

// TestValidatorOptions_PanicOnIncompatibleTags tests that creating a validator
// with StrictMissingFields=false and default/defaultUsingMethod tags panics.
// These combinations are incompatible because defaults only make sense when
// StrictMissingFields=true (missing field handling is disabled).
func TestValidatorOptions_PanicOnIncompatibleTags(t *testing.T) {
	tests := []struct {
		name              string
		testCase          string   // discriminator for which struct to use
		expectPanicFields []string // Expected field names in panic message
		expectPanicStrs   []string // Expected strings in panic message
	}{
		{
			name:              "panic_on_default_tag",
			testCase:          "single_default",
			expectPanicFields: []string{"Theme"},
			expectPanicStrs:   []string{"default=", "StrictMissingFields is false"},
		},
		{
			name:              "panic_on_multiple_default_tags",
			testCase:          "multiple_defaults",
			expectPanicFields: []string{"Name", "Port", "Enabled"},
			expectPanicStrs:   []string{"default=", "StrictMissingFields is false"},
		},
		{
			name:              "panic_on_defaultUsingMethod_tag",
			testCase:          "default_using_method",
			expectPanicFields: []string{"ID"},
			expectPanicStrs:   []string{"defaultUsingMethod=", "StrictMissingFields is false"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic but didn't panic")
				} else {
					panicMsg := r.(string)
					// Verify all expected strings are in panic message
					for _, expectedStr := range tt.expectPanicStrs {
						if !strings.Contains(panicMsg, expectedStr) {
							t.Errorf("panic message missing '%s', got: %s", expectedStr, panicMsg)
						}
					}
					// Verify at least one expected field is mentioned
					foundField := false
					for _, expectedField := range tt.expectPanicFields {
						if strings.Contains(panicMsg, expectedField) {
							foundField = true
							break
						}
					}
					if !foundField {
						t.Errorf("panic message should mention one of %v, got: %s", tt.expectPanicFields, panicMsg)
					}
				}
			}()

			// Use testCase discriminator to handle different struct types
			switch tt.testCase {
			case "single_default":
				type Settings struct {
					Theme    string `json:"theme" pedantigo:"default=dark"`
					Language string `json:"language"`
				}
				_ = New[Settings](ValidatorOptions{
					StrictMissingFields: false,
				})

			case "multiple_defaults":
				type Config struct {
					Name    string `json:"name" pedantigo:"default=unnamed"`
					Port    int    `json:"port" pedantigo:"default=8080"`
					Enabled bool   `json:"enabled" pedantigo:"default=true"`
				}
				_ = New[Config](ValidatorOptions{
					StrictMissingFields: false,
				})

			case "default_using_method":
				type Product struct {
					ID   string `json:"id" pedantigo:"defaultUsingMethod=GenerateID"`
					Name string `json:"name"`
				}
				_ = New[Product](ValidatorOptions{
					StrictMissingFields: false,
				})
			}
		})
	}
}
