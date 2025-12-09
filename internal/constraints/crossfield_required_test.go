package constraints_test

import (
	"testing"

	. "github.com/SmrutAI/Pedantigo"
)

// ============================================================================
// required_if Tests
// ============================================================================

func TestRequiredIf(t *testing.T) {
	type FormStringCondition struct {
		Country string `json:"country"`
		State   string `json:"state" pedantigo:"required_if=Country:US"`
	}

	type FormBoolCondition struct {
		IsPremium      bool   `json:"is_premium"`
		PremiumFeature string `json:"premium_feature" pedantigo:"required_if=IsPremium:true"`
	}

	type FormIntCondition struct {
		Status       int    `json:"status"`
		TrackingCode string `json:"tracking_code" pedantigo:"required_if=Status:2"`
	}

	type FormMultiple struct {
		Country  string `json:"country"`
		Domestic bool   `json:"domestic"`
		State    string `json:"state" pedantigo:"required_if=Country:US"`
		TaxID    string `json:"tax_id" pedantigo:"required_if=Domestic:true"`
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "string condition - met with field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Country: "US", State: "CA"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string condition - met without field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Country: "US", State: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "State" {
					t.Errorf("expected error for State field")
				}
			},
		},
		{
			name: "string condition - not met without field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Country: "CA", State: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string condition - not met with field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Country: "Canada", State: "Ontario"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean condition - true with field",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{IsPremium: true, PremiumFeature: "advanced_analytics"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean condition - true without field",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{IsPremium: true, PremiumFeature: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "PremiumFeature" {
					t.Errorf("expected error for PremiumFeature field")
				}
			},
		},
		{
			name: "boolean condition - false",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{IsPremium: false, PremiumFeature: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer condition - match with field",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Status: 2, TrackingCode: "TRACK123"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer condition - match without field",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Status: 2, TrackingCode: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "TrackingCode" {
					t.Errorf("expected error for TrackingCode field")
				}
			},
		},
		{
			name: "integer condition - no match",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Status: 0, TrackingCode: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple conditions - all satisfied",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{Country: "US", Domestic: true, State: "CA", TaxID: "123456"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple conditions - State missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{Country: "US", Domestic: true, State: "", TaxID: "123456"})
				if err == nil {
					t.Error("expected validation error for State")
				}
			},
		},
		{
			name: "multiple conditions - TaxID missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{Country: "US", Domestic: true, State: "CA", TaxID: ""})
				if err == nil {
					t.Error("expected validation error for TaxID")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// ============================================================================
// required_unless Tests
// ============================================================================

func TestRequiredUnless(t *testing.T) {
	type FormStringCondition struct {
		Status   string `json:"status"`
		Password string `json:"password" pedantigo:"required_unless=Status:guest"`
	}

	type FormBoolCondition struct {
		Automated   bool   `json:"automated"`
		CaptchaCode string `json:"captcha_code" pedantigo:"required_unless=Automated:true"`
	}

	type FormMultiple struct {
		UserType     string `json:"user_type"`
		IsBot        bool   `json:"is_bot"`
		Email        string `json:"email" pedantigo:"required_unless=UserType:anonymous"`
		Verification string `json:"verification" pedantigo:"required_unless=IsBot:true"`
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "string condition - not met with field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Status: "active", Password: "securepass123"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string condition - not met without field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Status: "active", Password: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "Password" {
					t.Errorf("expected error for Password field")
				}
			},
		},
		{
			name: "string condition - met without field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Status: "guest", Password: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string condition - met with field",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Status: "guest", Password: "anypassword"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean condition - false with field",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Automated: false, CaptchaCode: "abc123xyz"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean condition - false without field",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Automated: false, CaptchaCode: ""})
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
		{
			name: "boolean condition - true",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Automated: true, CaptchaCode: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple conditions - all satisfied",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{UserType: "user", IsBot: false, Email: "user@example.com", Verification: "code123"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple conditions - Email missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{UserType: "user", IsBot: false, Email: "", Verification: "code123"})
				if err == nil {
					t.Error("expected validation error for Email")
				}
			},
		},
		{
			name: "multiple conditions - Verification missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{UserType: "user", IsBot: false, Email: "user@example.com", Verification: ""})
				if err == nil {
					t.Error("expected validation error for Verification")
				}
			},
		},
		{
			name: "multiple conditions - both exceptions",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{UserType: "anonymous", IsBot: true, Email: "", Verification: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// ============================================================================
// required_with Tests
// ============================================================================

func TestRequiredWith(t *testing.T) {
	type FormStringCondition struct {
		Method string `json:"method"`
		Token  string `json:"token" pedantigo:"required_with=Method"`
	}

	type FormIntCondition struct {
		Quantity  int    `json:"quantity"`
		Warehouse string `json:"warehouse" pedantigo:"required_with=Quantity"`
	}

	type FormBoolCondition struct {
		Enabled bool   `json:"enabled"`
		Config  string `json:"config" pedantigo:"required_with=Enabled"`
	}

	type FormMultiple struct {
		GuestName     string `json:"guest_name"`
		Phone         string `json:"phone"`
		Address       string `json:"address" pedantigo:"required_with=GuestName"`
		EmergencyName string `json:"emergency_name" pedantigo:"required_with=Phone"`
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "string field - present with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Method: "credit_card", Token: "tok_123"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string field - present without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Method: "credit_card", Token: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "Token" {
					t.Errorf("expected error for Token field")
				}
			},
		},
		{
			name: "string field - absent without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Method: "", Token: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string field - absent with optional dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{Method: "", Token: "tok_123"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer field - nonzero with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Quantity: 5, Warehouse: "main"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer field - nonzero without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Quantity: 5, Warehouse: ""})
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
		{
			name: "integer field - zero",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{Quantity: 0, Warehouse: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean field - true with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Enabled: true, Config: "settings"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean field - true without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Enabled: true, Config: ""})
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
		{
			name: "boolean field - false",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{Enabled: false, Config: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple fields - all satisfied",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					GuestName:     "John Doe",
					Phone:         "555-1234",
					Address:       "123 Main St",
					EmergencyName: "Jane Doe",
				})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple fields - Address missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					GuestName:     "John Doe",
					Phone:         "555-1234",
					Address:       "",
					EmergencyName: "Jane Doe",
				})
				if err == nil {
					t.Error("expected validation error for Address")
				}
			},
		},
		{
			name: "multiple fields - EmergencyName missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					GuestName:     "John Doe",
					Phone:         "555-1234",
					Address:       "123 Main St",
					EmergencyName: "",
				})
				if err == nil {
					t.Error("expected validation error for EmergencyName")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// ============================================================================
// required_without Tests
// ============================================================================

func TestRequiredWithout(t *testing.T) {
	type FormStringCondition struct {
		DefaultAddress string `json:"default_address"`
		CustomAddress  string `json:"custom_address" pedantigo:"required_without=DefaultAddress"`
	}

	type FormIntCondition struct {
		FixedAmount    int    `json:"fixed_amount"`
		PercentageCode string `json:"percentage_code" pedantigo:"required_without=FixedAmount"`
	}

	type FormBoolCondition struct {
		UseDefault bool   `json:"use_default"`
		CustomRule string `json:"custom_rule" pedantigo:"required_without=UseDefault"`
	}

	type FormMultiple struct {
		WarehouseLocation string `json:"warehouse_location"`
		ShippingLabel     string `json:"shipping_label"`
		StorageBox        string `json:"storage_box" pedantigo:"required_without=WarehouseLocation"`
		ShippingTracking  string `json:"shipping_tracking" pedantigo:"required_without=ShippingLabel"`
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "string field - absent with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{DefaultAddress: "", CustomAddress: "123 Oak St"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string field - absent without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{DefaultAddress: "", CustomAddress: ""})
				if err == nil {
					t.Error("expected validation error")
				}
				ve, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected ValidationError, got %T", err)
				}
				if len(ve.Errors) == 0 || ve.Errors[0].Field != "CustomAddress" {
					t.Errorf("expected error for CustomAddress field")
				}
			},
		},
		{
			name: "string field - present without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{DefaultAddress: "456 Elm St", CustomAddress: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "string field - present with optional dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormStringCondition]()
				err := validator.Validate(&FormStringCondition{DefaultAddress: "456 Elm St", CustomAddress: "123 Oak St"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer field - zero with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{FixedAmount: 0, PercentageCode: "SAVE10"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "integer field - zero without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{FixedAmount: 0, PercentageCode: ""})
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
		{
			name: "integer field - nonzero",
			testFunc: func(t *testing.T) {
				validator := New[FormIntCondition]()
				err := validator.Validate(&FormIntCondition{FixedAmount: 50, PercentageCode: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean field - false with dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{UseDefault: false, CustomRule: "notify_all"})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "boolean field - false without dependency",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{UseDefault: false, CustomRule: ""})
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
		{
			name: "boolean field - true",
			testFunc: func(t *testing.T) {
				validator := New[FormBoolCondition]()
				err := validator.Validate(&FormBoolCondition{UseDefault: true, CustomRule: ""})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple fields - all satisfied",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					WarehouseLocation: "WH-A1",
					ShippingLabel:     "LABEL-123",
					StorageBox:        "",
					ShippingTracking:  "",
				})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "multiple fields - StorageBox missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					WarehouseLocation: "",
					ShippingLabel:     "LABEL-123",
					StorageBox:        "",
					ShippingTracking:  "TRACK-456",
				})
				if err == nil {
					t.Error("expected validation error for StorageBox")
				}
			},
		},
		{
			name: "multiple fields - ShippingTracking missing",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					WarehouseLocation: "WH-A1",
					ShippingLabel:     "",
					StorageBox:        "BOX-789",
					ShippingTracking:  "",
				})
				if err == nil {
					t.Error("expected validation error for ShippingTracking")
				}
			},
		},
		{
			name: "multiple fields - alternates provided",
			testFunc: func(t *testing.T) {
				validator := New[FormMultiple]()
				err := validator.Validate(&FormMultiple{
					WarehouseLocation: "",
					ShippingLabel:     "",
					StorageBox:        "BOX-789",
					ShippingTracking:  "TRACK-456",
				})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

// ============================================================================
// Cross-Constraint Integration Tests
// ============================================================================

func TestCrossFieldConstraints(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "complex scenario - valid business account",
			testFunc: func(t *testing.T) {
				type UserProfile struct {
					AccountType      string `json:"account_type"`
					IsVerified       bool   `json:"is_verified"`
					BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
					TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
					VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
					BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
					NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
				}
				validator := New[UserProfile]()
				err := validator.Validate(&UserProfile{
					AccountType:      "business",
					IsVerified:       true,
					BusinessName:     "Acme Corp",
					TaxID:            "12-3456789",
					VerificationDoc:  "doc_123",
					BackupEmail:      "backup@acme.com",
					NotificationPref: "email",
				})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "complex scenario - business without BusinessName",
			testFunc: func(t *testing.T) {
				type UserProfile struct {
					AccountType      string `json:"account_type"`
					IsVerified       bool   `json:"is_verified"`
					BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
					TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
					VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
					BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
					NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
				}
				validator := New[UserProfile]()
				err := validator.Validate(&UserProfile{
					AccountType:      "business",
					IsVerified:       false,
					BusinessName:     "",
					TaxID:            "12-3456789",
					VerificationDoc:  "",
					BackupEmail:      "backup@example.com",
					NotificationPref: "email",
				})
				if err == nil {
					t.Error("expected validation error for BusinessName")
				}
			},
		},
		{
			name: "complex scenario - valid government account",
			testFunc: func(t *testing.T) {
				type UserProfile struct {
					AccountType      string `json:"account_type"`
					IsVerified       bool   `json:"is_verified"`
					BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
					TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
					VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
					BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
					NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
				}
				validator := New[UserProfile]()
				err := validator.Validate(&UserProfile{
					AccountType:      "government",
					IsVerified:       true,
					BusinessName:     "",
					TaxID:            "",
					VerificationDoc:  "gov_doc_456",
					BackupEmail:      "",
					NotificationPref: "",
				})
				if err != nil {
					t.Errorf("expected no errors, got: %v", err)
				}
			},
		},
		{
			name: "complex scenario - missing NotificationPref for BackupEmail",
			testFunc: func(t *testing.T) {
				type UserProfile struct {
					AccountType      string `json:"account_type"`
					IsVerified       bool   `json:"is_verified"`
					BusinessName     string `json:"business_name" pedantigo:"required_if=AccountType:business"`
					TaxID            string `json:"tax_id" pedantigo:"required_if=AccountType:business"`
					VerificationDoc  string `json:"verification_doc" pedantigo:"required_if=IsVerified:true"`
					BackupEmail      string `json:"backup_email" pedantigo:"required_unless=AccountType:government"`
					NotificationPref string `json:"notification_pref" pedantigo:"required_with=BackupEmail"`
				}
				validator := New[UserProfile]()
				err := validator.Validate(&UserProfile{
					AccountType:      "personal",
					IsVerified:       false,
					BusinessName:     "",
					TaxID:            "",
					VerificationDoc:  "",
					BackupEmail:      "backup@example.com",
					NotificationPref: "",
				})
				if err == nil {
					t.Error("expected validation error for NotificationPref")
				}
			},
		},
		{
			name: "field index resolution",
			testFunc: func(t *testing.T) {
				type Form struct {
					Field1 string `json:"field1"`
					Field2 string `json:"field2"`
					Field3 string `json:"field3" pedantigo:"required_if=Field1:trigger"`
					Field4 string `json:"field4" pedantigo:"required_unless=Field2:skip"`
				}
				validator := New[Form]()
				if validator == nil {
					t.Fatal("validator creation failed")
				}
				_ = validator.Validate(&Form{
					Field1: "trigger",
					Field2: "active",
					Field3: "value",
					Field4: "value",
				})
			},
		},
		{
			name: "zero value distinction",
			testFunc: func(t *testing.T) {
				type Form struct {
					TriggerField string `json:"trigger_field"`
					TargetField  string `json:"target_field" pedantigo:"required_with=TriggerField"`
				}
				validator := New[Form]()
				err := validator.Validate(&Form{
					TriggerField: "value",
					TargetField:  "",
				})
				if err == nil {
					t.Error("expected validation error for zero value with required_with")
				}
			},
		},
		{
			name: "unexported fields handling - valid",
			testFunc: func(t *testing.T) {
				type Form struct {
					privateField string
					PublicField  string `json:"public_field"`
					Conditional  string `json:"conditional" pedantigo:"required_if=PublicField:trigger"`
				}
				validator := New[Form]()
				err := validator.Validate(&Form{
					privateField: "ignored",
					PublicField:  "trigger",
					Conditional:  "value",
				})
				if err != nil {
					t.Errorf("expected validation to work, got: %v", err)
				}
			},
		},
		{
			name: "unexported fields handling - invalid",
			testFunc: func(t *testing.T) {
				type Form struct {
					privateField string
					PublicField  string `json:"public_field"`
					Conditional  string `json:"conditional" pedantigo:"required_if=PublicField:trigger"`
				}
				validator := New[Form]()
				err := validator.Validate(&Form{
					privateField: "ignored",
					PublicField:  "trigger",
					Conditional:  "",
				})
				if err == nil {
					t.Error("expected validation error when Conditional missing")
				}
			},
		},
		{
			name: "reflect value handling",
			testFunc: func(t *testing.T) {
				type Form struct {
					Status string `json:"status"`
					Detail string `json:"detail" pedantigo:"required_if=Status:complete"`
				}
				validator := New[Form]()
				form := Form{
					Status: "complete",
					Detail: "",
				}
				err := validator.Validate(&form)
				if err == nil {
					t.Error("expected validation error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
