package pedantigo

import (
	"testing"
)

// ==================================================
// enum constraint tests
// ==================================================

func TestEnum_ValidString(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	jsonData := []byte(`{"role":"admin"}`)

	user, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid enum value, got %v", err)
	}

	if user.Role != "admin" {
		t.Errorf("expected role 'admin', got %s", user.Role)
	}
}

func TestEnum_InvalidString(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	jsonData := []byte(`{"role":"superadmin"}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Error("expected validation error for invalid enum value")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Role" && fieldErr.Message == "must be one of: admin, user, guest" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be one of' error, got %v", ve.Errors)
	}
}

func TestEnum_ValidInteger(t *testing.T) {
	type Status struct {
		Code int `json:"code" pedantigo:"oneof=200 201 204"`
	}

	validator := New[Status]()
	jsonData := []byte(`{"code":200}`)

	status, err := validator.Unmarshal(jsonData)
	if err != nil {
		t.Errorf("expected no errors for valid enum value, got %v", err)
	}

	if status.Code != 200 {
		t.Errorf("expected code 200, got %d", status.Code)
	}
}

func TestEnum_InvalidInteger(t *testing.T) {
	type Status struct {
		Code int `json:"code" pedantigo:"oneof=200 201 204"`
	}

	validator := New[Status]()
	jsonData := []byte(`{"code":404}`)

	_, err := validator.Unmarshal(jsonData)
	if err == nil {
		t.Error("expected validation error for invalid enum value")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	foundError := false
	for _, fieldErr := range ve.Errors {
		if fieldErr.Field == "Code" && fieldErr.Message == "must be one of: 200, 201, 204" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected 'must be one of' error, got %v", ve.Errors)
	}
}

func TestEnum_InSlice(t *testing.T) {
	type Config struct {
		Roles []string `json:"roles" pedantigo:"oneof=admin user guest"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"roles":["admin","user","superadmin"]}`)

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
		if fieldErr.Field == "Roles[2]" && fieldErr.Message == "must be one of: admin, user, guest" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Roles[2]', got %v", ve.Errors)
	}
}

func TestEnum_InMap(t *testing.T) {
	type Config struct {
		Permissions map[string]string `json:"permissions" pedantigo:"oneof=read write execute"`
	}

	validator := New[Config]()
	jsonData := []byte(`{"permissions":{"file":"read","script":"delete"}}`)

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
		if fieldErr.Field == "Permissions[script]" && fieldErr.Message == "must be one of: read, write, execute" {
			foundError = true
		}
	}

	if !foundError {
		t.Errorf("expected error at 'Permissions[script]', got %v", ve.Errors)
	}
}

func TestEnum_Schema(t *testing.T) {
	type User struct {
		Role string `json:"role" pedantigo:"oneof=admin user guest"`
	}

	validator := New[User]()
	schema := validator.Schema()

	roleProp, ok := schema.Properties.Get("role")
	if !ok || roleProp == nil {
		t.Fatal("expected 'role' property to exist")
	}

	if len(roleProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(roleProp.Enum))
	}

	expectedValues := map[string]bool{"admin": false, "user": false, "guest": false}
	for _, val := range roleProp.Enum {
		strVal, ok := val.(string)
		if !ok {
			t.Errorf("expected enum value to be string, got %T", val)
			continue
		}
		if _, exists := expectedValues[strVal]; exists {
			expectedValues[strVal] = true
		}
	}

	for val, found := range expectedValues {
		if !found {
			t.Errorf("expected enum value '%s' not found", val)
		}
	}
}
