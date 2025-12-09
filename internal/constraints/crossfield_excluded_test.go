package constraints_test

import (
	"testing"

	. "github.com/SmrutAI/Pedantigo"
)

// ==================================================
// excluded_if constraint tests
// ==================================================
// Field must be absent (zero value) if another field equals specific value

func TestExcludedIf(t *testing.T) {
	type Payment struct {
		Method     string `json:"method" pedantigo:"required"`
		CashAmount int    `json:"cash_amount" pedantigo:"excluded_if=Method card"`
	}

	type Order struct {
		PaymentType string `json:"payment_type" pedantigo:"required"`
		CheckNumber string `json:"check_number" pedantigo:"excluded_if=PaymentType credit_card"`
	}

	type UserPreferences struct {
		OptIn       bool   `json:"opt_in" pedantigo:"required"`
		PhoneNumber string `json:"phone_number" pedantigo:"excluded_if=OptIn false"`
	}

	type Vehicle struct {
		Type         string `json:"type" pedantigo:"required"`
		LicensePlate string `json:"license_plate" pedantigo:"excluded_if=Type bicycle"`
		ParkingSpot  int    `json:"parking_spot" pedantigo:"excluded_if=Type bicycle"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "condition met field absent - valid",
			validator: New[Payment](),
			data: &Payment{
				Method:     "card",
				CashAmount: 0,
			},
			expectErr: false,
		},
		{
			name:      "condition met field present - invalid",
			validator: New[Payment](),
			data: &Payment{
				Method:     "card",
				CashAmount: 100,
			},
			expectErr: true,
			errField:  "CashAmount",
		},
		{
			name:      "condition not met field present - valid",
			validator: New[Payment](),
			data: &Payment{
				Method:     "cash",
				CashAmount: 100,
			},
			expectErr: false,
		},
		{
			name:      "condition not met field absent - valid",
			validator: New[Payment](),
			data: &Payment{
				Method:     "cash",
				CashAmount: 0,
			},
			expectErr: false,
		},
		{
			name:      "string comparison valid - empty check number",
			validator: New[Order](),
			data: &Order{
				PaymentType: "credit_card",
				CheckNumber: "",
			},
			expectErr: false,
		},
		{
			name:      "string comparison invalid - check number present",
			validator: New[Order](),
			data: &Order{
				PaymentType: "credit_card",
				CheckNumber: "CHK123456",
			},
			expectErr: true,
			errField:  "CheckNumber",
		},
		{
			name:      "boolean condition valid - false with absent",
			validator: New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       false,
				PhoneNumber: "",
			},
			expectErr: false,
		},
		{
			name:      "boolean condition invalid - false with present",
			validator: New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       false,
				PhoneNumber: "+1234567890",
			},
			expectErr: true,
			errField:  "PhoneNumber",
		},
		{
			name:      "boolean condition valid - true with present",
			validator: New[UserPreferences](),
			data: &UserPreferences{
				OptIn:       true,
				PhoneNumber: "+1234567890",
			},
			expectErr: false,
		},
		{
			name:      "multiple conditions - bicycle without license plate",
			validator: New[Vehicle](),
			data: &Vehicle{
				Type:         "bicycle",
				LicensePlate: "",
				ParkingSpot:  0,
			},
			expectErr: false,
		},
		{
			name:      "multiple conditions - bicycle with license plate invalid",
			validator: New[Vehicle](),
			data: &Vehicle{
				Type:         "bicycle",
				LicensePlate: "ABC123",
				ParkingSpot:  0,
			},
			expectErr: true,
			errField:  "LicensePlate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *Validator[Payment]:
				err := v.Validate(tt.data.(*Payment))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Order]:
				err := v.Validate(tt.data.(*Order))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[UserPreferences]:
				err := v.Validate(tt.data.(*UserPreferences))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Vehicle]:
				err := v.Validate(tt.data.(*Vehicle))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			}
		})
	}
}

// ==================================================
// excluded_unless constraint tests
// ==================================================
// Field must be absent (zero value) unless another field equals specific value

func TestExcludedUnless(t *testing.T) {
	type Document struct {
		Status        string `json:"status" pedantigo:"required"`
		ApprovalNotes string `json:"approval_notes" pedantigo:"excluded_unless=Status approved"`
	}

	type Permission struct {
		Role            string `json:"role" pedantigo:"required"`
		SecretKeyAccess string `json:"secret_key_access" pedantigo:"excluded_unless=Role admin"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "condition met field present - valid",
			validator: New[Document](),
			data: &Document{
				Status:        "approved",
				ApprovalNotes: "Looks good to me",
			},
			expectErr: false,
		},
		{
			name:      "condition met field absent - valid",
			validator: New[Document](),
			data: &Document{
				Status:        "approved",
				ApprovalNotes: "",
			},
			expectErr: false,
		},
		{
			name:      "condition not met field absent - valid",
			validator: New[Document](),
			data: &Document{
				Status:        "pending",
				ApprovalNotes: "",
			},
			expectErr: false,
		},
		{
			name:      "condition not met field present - invalid",
			validator: New[Document](),
			data: &Document{
				Status:        "pending",
				ApprovalNotes: "Some notes",
			},
			expectErr: true,
			errField:  "ApprovalNotes",
		},
		{
			name:      "multiple values admin with access - valid",
			validator: New[Permission](),
			data: &Permission{
				Role:            "admin",
				SecretKeyAccess: "secret123",
			},
			expectErr: false,
		},
		{
			name:      "multiple values user with access - invalid",
			validator: New[Permission](),
			data: &Permission{
				Role:            "user",
				SecretKeyAccess: "secret123",
			},
			expectErr: true,
			errField:  "SecretKeyAccess",
		},
		{
			name:      "multiple values user without access - valid",
			validator: New[Permission](),
			data: &Permission{
				Role:            "user",
				SecretKeyAccess: "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *Validator[Document]:
				err := v.Validate(tt.data.(*Document))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Permission]:
				err := v.Validate(tt.data.(*Permission))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			}
		})
	}
}

// ==================================================
// excluded_with constraint tests
// ==================================================
// Field must be absent (zero value) if another field is present (non-zero)

func TestExcludedWith(t *testing.T) {
	type User struct {
		HomePhone string `json:"home_phone" pedantigo:"required"`
		WorkPhone string `json:"work_phone" pedantigo:"excluded_with=HomePhone"`
	}

	type Account struct {
		BankBalance    int `json:"bank_balance" pedantigo:"min=0"`
		CreditLineUsed int `json:"credit_line_used" pedantigo:"excluded_with=BankBalance"`
	}

	type Feature struct {
		EnabledGlobally bool   `json:"enabled_globally" pedantigo:"required"`
		OverrideReason  string `json:"override_reason" pedantigo:"excluded_with=EnabledGlobally"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "other field present field absent - valid",
			validator: New[User](),
			data: &User{
				HomePhone: "+1234567890",
				WorkPhone: "",
			},
			expectErr: false,
		},
		{
			name:      "other field present field present - invalid",
			validator: New[User](),
			data: &User{
				HomePhone: "+1234567890",
				WorkPhone: "+0987654321",
			},
			expectErr: true,
			errField:  "WorkPhone",
		},
		{
			name:      "other field absent field present - valid",
			validator: New[User](),
			data: &User{
				HomePhone: "",
				WorkPhone: "+0987654321",
			},
			expectErr: false,
		},
		{
			name:      "other field absent field absent - valid",
			validator: New[User](),
			data: &User{
				HomePhone: "",
				WorkPhone: "",
			},
			expectErr: false,
		},
		{
			name:      "integer field present absent - valid",
			validator: New[Account](),
			data: &Account{
				BankBalance:    5000,
				CreditLineUsed: 0,
			},
			expectErr: false,
		},
		{
			name:      "integer field both present - invalid",
			validator: New[Account](),
			data: &Account{
				BankBalance:    5000,
				CreditLineUsed: 2000,
			},
			expectErr: true,
			errField:  "CreditLineUsed",
		},
		{
			name:      "integer field absent can be present - valid",
			validator: New[Account](),
			data: &Account{
				BankBalance:    0,
				CreditLineUsed: 2000,
			},
			expectErr: false,
		},
		{
			name:      "boolean field true reason absent - valid",
			validator: New[Feature](),
			data: &Feature{
				EnabledGlobally: true,
				OverrideReason:  "",
			},
			expectErr: false,
		},
		{
			name:      "boolean field true reason present - invalid",
			validator: New[Feature](),
			data: &Feature{
				EnabledGlobally: true,
				OverrideReason:  "Special case",
			},
			expectErr: true,
			errField:  "OverrideReason",
		},
		{
			name:      "boolean field false reason present - valid",
			validator: New[Feature](),
			data: &Feature{
				EnabledGlobally: false,
				OverrideReason:  "Special case",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *Validator[User]:
				err := v.Validate(tt.data.(*User))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Account]:
				err := v.Validate(tt.data.(*Account))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Feature]:
				err := v.Validate(tt.data.(*Feature))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			}
		})
	}
}

// ==================================================
// excluded_without constraint tests
// ==================================================
// Field must be absent (zero value) if another field is absent (zero)

func TestExcludedWithout(t *testing.T) {
	type Address struct {
		Country string `json:"country" pedantigo:"required"`
		ZipCode string `json:"zip_code" pedantigo:"excluded_without=Country"`
	}

	type Notification struct {
		IsEnabled   bool   `json:"is_enabled" pedantigo:"required"`
		RetryPolicy string `json:"retry_policy" pedantigo:"excluded_without=IsEnabled"`
	}

	tests := []struct {
		name      string
		validator interface{}
		data      interface{}
		expectErr bool
		errField  string
	}{
		{
			name:      "other field absent field absent - valid",
			validator: New[Address](),
			data: &Address{
				Country: "",
				ZipCode: "",
			},
			expectErr: false,
		},
		{
			name:      "other field absent field present - invalid",
			validator: New[Address](),
			data: &Address{
				Country: "",
				ZipCode: "12345",
			},
			expectErr: true,
			errField:  "ZipCode",
		},
		{
			name:      "other field present field present - valid",
			validator: New[Address](),
			data: &Address{
				Country: "USA",
				ZipCode: "12345",
			},
			expectErr: false,
		},
		{
			name:      "other field present field absent - valid",
			validator: New[Address](),
			data: &Address{
				Country: "USA",
				ZipCode: "",
			},
			expectErr: false,
		},
		{
			name:      "boolean field true policy present - valid",
			validator: New[Notification](),
			data: &Notification{
				IsEnabled:   true,
				RetryPolicy: "exponential",
			},
			expectErr: false,
		},
		{
			name:      "boolean field false policy present - invalid",
			validator: New[Notification](),
			data: &Notification{
				IsEnabled:   false,
				RetryPolicy: "exponential",
			},
			expectErr: true,
			errField:  "RetryPolicy",
		},
		{
			name:      "boolean field false policy absent - valid",
			validator: New[Notification](),
			data: &Notification{
				IsEnabled:   false,
				RetryPolicy: "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.validator.(type) {
			case *Validator[Address]:
				err := v.Validate(tt.data.(*Address))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			case *Validator[Notification]:
				err := v.Validate(tt.data.(*Notification))
				if tt.expectErr && err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if !tt.expectErr && err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if tt.expectErr && err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						t.Fatalf("expected *ValidationError, got %T", err)
					}
					foundError := false
					for _, fieldErr := range ve.Errors {
						if fieldErr.Field == tt.errField {
							foundError = true
							break
						}
					}
					if !foundError {
						t.Errorf("expected error for field %s, got %v", tt.errField, ve.Errors)
					}
				}
			}
		})
	}
}

// ==================================================
// excluded_without unmarshal integration tests
// ==================================================

func TestExcludedWithoutUnmarshal(t *testing.T) {
	type Shipping struct {
		Weight      int `json:"weight"`
		TrackingNum int `json:"tracking_num" pedantigo:"excluded_without=Weight"`
	}

	tests := []struct {
		name      string
		jsonData  string
		expectErr bool
		checkFn   func(*Shipping) bool
	}{
		{
			name:      "both fields absent from json - valid",
			jsonData:  `{}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 0 && s.TrackingNum == 0
			},
		},
		{
			name:      "weight absent tracking num present - invalid",
			jsonData:  `{"tracking_num": 123456789}`,
			expectErr: true,
		},
		{
			name:      "weight present tracking num present - valid",
			jsonData:  `{"weight": 500, "tracking_num": 123456789}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 500 && s.TrackingNum == 123456789
			},
		},
		{
			name:      "weight present tracking num absent - valid",
			jsonData:  `{"weight": 500}`,
			expectErr: false,
			checkFn: func(s *Shipping) bool {
				return s.Weight == 500 && s.TrackingNum == 0
			},
		},
	}

	validator := New[Shipping]()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.Unmarshal([]byte(tt.jsonData))
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}
			if !tt.expectErr && tt.checkFn != nil {
				if !tt.checkFn(result) {
					t.Errorf("data validation failed for %+v", result)
				}
			}
		})
	}
}

// ==================================================
// Integration tests combining multiple constraints
// ==================================================

func TestMultipleExclusionConstraints_Complex(t *testing.T) {
	type Subscription struct {
		Status             string `json:"status" pedantigo:"required"`
		CancellationReason string `json:"cancellation_reason" pedantigo:"excluded_unless=Status cancelled"`
		DowngradeReason    string `json:"downgrade_reason" pedantigo:"excluded_unless=Status downgraded"`
		SuspendedUntilDate string `json:"suspended_until_date" pedantigo:"excluded_without=Status"`
	}

	validator := New[Subscription]()

	tests := []struct {
		name      string
		data      *Subscription
		expectErr bool
	}{
		{
			name: "active subscription - valid",
			data: &Subscription{
				Status:             "active",
				CancellationReason: "",
				DowngradeReason:    "",
				SuspendedUntilDate: "",
			},
			expectErr: false,
		},
		{
			name: "cancelled subscription with reason - valid",
			data: &Subscription{
				Status:             "cancelled",
				CancellationReason: "Not needed",
				DowngradeReason:    "",
				SuspendedUntilDate: "",
			},
			expectErr: false,
		},
		{
			name: "active subscription with cancellation reason - invalid",
			data: &Subscription{
				Status:             "active",
				CancellationReason: "Not needed",
				DowngradeReason:    "",
				SuspendedUntilDate: "2025-01-01",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestConditionalExclusion_RealWorldPaymentExample(t *testing.T) {
	type PaymentMethod struct {
		Type           string `json:"type" pedantigo:"required"`
		CardNumber     string `json:"card_number" pedantigo:"excluded_unless=Type card"`
		BankAccount    string `json:"bank_account" pedantigo:"excluded_unless=Type bank_transfer"`
		CryptoCurrency string `json:"crypto_currency" pedantigo:"excluded_unless=Type crypto"`
		CardExpiryDate string `json:"card_expiry_date" pedantigo:"excluded_with=BankAccount,excluded_with=CryptoCurrency"`
		RoutingNumber  string `json:"routing_number" pedantigo:"excluded_without=Type"`
	}

	validator := New[PaymentMethod]()

	tests := []struct {
		name      string
		data      *PaymentMethod
		expectErr bool
	}{
		{
			name: "credit card payment - valid",
			data: &PaymentMethod{
				Type:           "card",
				CardNumber:     "4111111111111111",
				CardExpiryDate: "12/25",
				BankAccount:    "",
				CryptoCurrency: "",
				RoutingNumber:  "",
			},
			expectErr: false,
		},
		{
			name: "card payment with bank account - invalid",
			data: &PaymentMethod{
				Type:           "card",
				CardNumber:     "4111111111111111",
				CardExpiryDate: "12/25",
				BankAccount:    "123456789",
				CryptoCurrency: "",
				RoutingNumber:  "",
			},
			expectErr: true,
		},
		{
			name: "bank transfer payment - valid",
			data: &PaymentMethod{
				Type:           "bank_transfer",
				CardNumber:     "",
				CardExpiryDate: "",
				BankAccount:    "123456789",
				CryptoCurrency: "",
				RoutingNumber:  "021000021",
			},
			expectErr: false,
		},
		{
			name: "crypto payment - valid",
			data: &PaymentMethod{
				Type:           "crypto",
				CardNumber:     "",
				CardExpiryDate: "",
				BankAccount:    "",
				CryptoCurrency: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
				RoutingNumber:  "",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.expectErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
