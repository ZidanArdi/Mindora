package validator_test

import (
	"testing"

	"mindora/backend/pkg/validator"
)

type TestDTO struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,password_strength"`
}

func TestValidator(t *testing.T) {
	v := validator.New()

	// Valid struct
	valid := TestDTO{
		Email:    "valid.user@mindora.app",
		Password: "StrongPassword123!",
	}
	if err := v.Validate(valid); err != nil {
		t.Errorf("expected valid struct, got error: %v", err)
	}

	// Invalid email
	invalidEmail := TestDTO{
		Email:    "not-an-email",
		Password: "StrongPassword123!",
	}
	if err := v.Validate(invalidEmail); err == nil {
		t.Errorf("expected validation error for invalid email, got nil")
	}

	// Weak password (too short)
	weakShort := TestDTO{
		Email:    "user@mindora.app",
		Password: "Sh0rt!",
	}
	if err := v.Validate(weakShort); err == nil {
		t.Errorf("expected validation error for short password, got nil")
	}

	// Weak password (no uppercase)
	weakNoUpper := TestDTO{
		Email:    "user@mindora.app",
		Password: "lowercase123!",
	}
	if err := v.Validate(weakNoUpper); err == nil {
		t.Errorf("expected validation error for password without uppercase, got nil")
	}
}
