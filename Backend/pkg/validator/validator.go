package validator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the go-playground validator with domain-specific rules.
type CustomValidator struct {
	validate *validator.Validate
}

// New creates and initializes a CustomValidator.
func New() *CustomValidator {
	v := validator.New()

	// Register custom password strength validator: "password_strength"
	_ = v.RegisterValidation("password_strength", func(fl validator.FieldLevel) bool {
		pwd := fl.Field().String()
		if len(pwd) < 8 {
			return false
		}
		var hasUpper, hasLower, hasNumberOrSpecial bool
		for _, c := range pwd {
			switch {
			case unicode.IsUpper(c):
				hasUpper = true
			case unicode.IsLower(c):
				hasLower = true
			case unicode.IsDigit(c) || unicode.IsPunct(c) || unicode.IsSymbol(c):
				hasNumberOrSpecial = true
			}
		}
		return hasUpper && hasLower && hasNumberOrSpecial
	})

	return &CustomValidator{validate: v}
}

// Validate validates a struct against validator tags.
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validate.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, ve := range validationErrors {
				errMsgs = append(errMsgs, fmt.Sprintf("field '%s' failed validation tag '%s'", ve.Field(), ve.Tag()))
			}
			return fmt.Errorf("validation error: %s", strings.Join(errMsgs, ", "))
		}
		return err
	}
	return nil
}
