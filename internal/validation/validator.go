package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

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

// OneOfValidator checks if a string is one of the allowed values
func OneOfValidator(fl validator.FieldLevel) bool {
	field := fl.Field()
	param := fl.Param() // Get the comma-separated list of valid values

	validValues := strings.Split(param, ",") // Split the parameter string by commas

	for _, validValue := range validValues {
		if field.String() == strings.TrimSpace(validValue) { // Trim whitespace from validValue.  Updated
			return true
		}
	}

	return false // Return false if no match is found. Updated
}

// ValidatorError struct to contain the error from the validator library
type ValidatorError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

// Errorf creates error message for validator error
func (v *ValidatorError) Errorf(str string, args ...interface{}) string {
	return fmt.Sprintf(str, args...)
}
