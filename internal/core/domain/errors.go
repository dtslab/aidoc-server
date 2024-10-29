package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

// Define custom error types
var (
	ErrPatientNotFound             = errors.New("patient not found")
	ErrInvalidPatientData          = errors.New("invalid patient data")
	ErrDatabaseConnection          = errors.New("database connection error")
	ErrFailedToCreatePatient       = errors.New("failed to create patient")
	ErrMedicalHistoryEntryNotFound = errors.New("medical history entry not found")
	ErrInvalidMedicalHistoryData   = errors.New("invalid medical history data")
	ErrInvalidInput                = errors.New("invalid input")
	ErrLifestyleEntryNotFound      = errors.New("lifestyle entry not found")
	ErrForbidden                   = errors.New("forbidden") // unauthorized access
)

// ValidationError struct with details
type ValidationError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *ValidationError) Error() string {
	if e == nil { // Handle nil receiver.  Updated.
		return ""
	}

	if len(e.Details) > 0 {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Details) // Include details in error message if available
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error string `json:"error"`
}

// ClerkClaims for Clerk JWT
type ClerkClaims struct {
	UserID string `json:"user_id"`
}

// Custom Validator Functions

// PastDateValidator checks if a date is in the past
func PastDateValidator(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return date.Before(time.Now())
}

// DateFormatValidator checks date format (YYYY-MM-DD)
func DateFormatValidator(fl validator.FieldLevel) bool {
	dateString, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	_, err := time.Parse("2006-01-02", dateString)
	return err == nil
}

// MinimumAgeValidator checks if age is 18+
func MinimumAgeValidator(fl validator.FieldLevel) bool {
	age, ok := fl.Field().Interface().(int)
	if !ok {
		return false
	}
	return age >= 18
}

// PhoneNumberValidator checks phone number format
func PhoneNumberValidator(fl validator.FieldLevel) bool {
	phoneNumber, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	if phoneNumber == "" {
		return true // Allow empty
	}

	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phoneNumber)
}

// ValidateEmail validates an email address.
func ValidateEmail(fl validator.FieldLevel) bool {
	email, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if email == "" {
		return true // Allow empty
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
