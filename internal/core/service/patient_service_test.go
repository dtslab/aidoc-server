package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/mocks" // Import the mocks package

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCreatePatient(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockRepo := new(mocks.MockPatientRepository) // Use mocks.MockPatientRepository
	svc := NewPatientService(mockRepo, log, v)

	t.Run("success", func(t *testing.T) {
		req := domain.CreatePatientRequest{
			UserID:       1,
			FullName:     "John Doe",
			Age:          30,
			DateOfBirth:  time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC),
			Sex:          "Male",
			EmailAddress: "john.doe@example.com",
		}

		expectedPatient := &domain.Patient{
			UserID:       1,
			FullName:     "John Doe",
			Age:          30,
			DateOfBirth:  time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC),
			Sex:          "Male",
			EmailAddress: "john.doe@example.com",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			PatientID:    1,
		}
		mockRepo.On("CreatePatient", mock.Anything, mock.AnythingOfType("*domain.Patient")).Return(expectedPatient, nil)

		patient, err := svc.CreatePatient(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, patient)

		assert.Equal(t, expectedPatient, patient) // Check if patient data is correct. Updated
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		req := domain.CreatePatientRequest{
			UserID: 1,
			// Missing FullName, invalid Sex
			Age:          30,
			DateOfBirth:  time.Now(), // Invalid DOB (not in the past)
			Sex:          "Invalid",
			EmailAddress: "john.doe", // Invalid email
			PhoneNumber:  "invalid",  // Invalid phone
		}

		_, err := svc.CreatePatient(context.Background(), req)
		assert.Error(t, err)

		var validationError *domain.ValidationError
		assert.True(t, errors.As(err, &validationError)) // Assert it is ValidationError. Updated

		// Example assertions on the ValidationError details.
		assert.Equal(t, "INVALID_PATIENT_DATA", validationError.Code)
		assert.Contains(t, validationError.Details, "FullName is a required field")
		assert.Contains(t, validationError.Details, "Sex must be one of the following")
		assert.Contains(t, validationError.Details, "DateOfBirth must be in the past")            // Assert invalid date of birth
		assert.Contains(t, validationError.Details, "EmailAddress must be a valid email address") // Invalid email
		assert.Contains(t, validationError.Details, "PhoneNumber must be a valid phone number")   // Invalid phone number

		mockRepo.AssertNotCalled(t, "CreatePatient") // Ensure the repo is not called during validation error

	})

	t.Run("inconsistent_age_dob", func(t *testing.T) {
		req := domain.CreatePatientRequest{
			UserID:                 1,
			FullName:               "John Doe",
			Age:                    25,                                          // Inconsistent age
			DateOfBirth:            time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC), // DOB corresponds to a different age
			Sex:                    "Male",
			EmailAddress:           "john.doe@example.com",
			PreferredCommunication: "Email",
			SocioeconomicStatus:    "Middle",
			GeographicLocation:     "Testville",
		}

		_, err := svc.CreatePatient(context.Background(), req)

		assert.Error(t, err)

		var validationError *domain.ValidationError // Using type assertion with errors.As
		assert.True(t, errors.As(err, &validationError))
		assert.Equal(t, "INCONSISTENT_DATA", validationError.Code)

	})

	t.Run("repository_error", func(t *testing.T) {
		req := domain.CreatePatientRequest{
			UserID:       1,
			FullName:     "John Doe",
			Age:          30,
			DateOfBirth:  time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC),
			Sex:          "Male",
			EmailAddress: "john.doe@example.com",
		}

		mockRepo.On("CreatePatient", mock.Anything, mock.AnythingOfType("*domain.Patient")).Return(nil, errors.New("database error"))
		_, err := svc.CreatePatient(context.Background(), req)
		assert.Error(t, err)
		// Add more assertions to check the specific error returned
		// if your service layer logic distinguishes between error types.
		assert.Contains(t, err.Error(), "database error") // Example: Assert the error message

	})
	// ... Add more test cases as needed (e.g., for email already exists)

}

func TestGetPatient(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockRepo := new(mocks.MockPatientRepository) // Use mocks.MockPatientRepository
	svc := NewPatientService(mockRepo, log, v)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		expectedPatient := &domain.Patient{PatientID: 1, FullName: "John Doe"}
		mockRepo.On("GetPatient", mock.Anything, patientID).Return(expectedPatient, nil)

		patient, err := svc.GetPatient(context.Background(), patientID)

		assert.Nil(t, err)
		assert.Equal(t, expectedPatient, patient)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found_error", func(t *testing.T) {
		patientID := 1
		mockRepo.On("GetPatient", mock.Anything, patientID).Return(nil, domain.ErrPatientNotFound)
		patient, err := svc.GetPatient(context.Background(), patientID)

		assert.Nil(t, patient)
		assert.Equal(t, domain.ErrPatientNotFound, err) // Check for correct error type. Updated
	})

	t.Run("repository_error", func(t *testing.T) {
		patientID := 1
		mockRepo.On("GetPatient", mock.Anything, patientID).Return(nil, errors.New("database error"))

		_, err := svc.GetPatient(context.Background(), patientID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error") // Assert the specific error message
	})
}

func TestUpdatePatient(t *testing.T) {
	// ... (Implementation will be very similar to TestCreatePatient, using UpdatePatient mock and request)
	log := zap.NewNop()
	v := validator.New()
	mockRepo := new(mocks.MockPatientRepository)
	svc := NewPatientService(mockRepo, log, v)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		req := domain.UpdatePatientRequest{FullName: "Updated Name"}
		expectedPatient := &domain.Patient{PatientID: patientID, FullName: "Updated Name", UpdatedAt: time.Now()}

		mockRepo.On("GetPatient", mock.Anything, patientID).Return(&domain.Patient{PatientID: patientID}, nil) // Existing patient. Updated
		mockRepo.On("UpdatePatient", mock.Anything, patientID, mock.AnythingOfType("*domain.Patient")).Return(expectedPatient, nil)

		patient, err := svc.UpdatePatient(context.Background(), patientID, req)
		assert.NoError(t, err)
		assert.NotNil(t, patient)
		assert.Equal(t, "Updated Name", patient.FullName) // Check updated name. Updated
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation_error", func(t *testing.T) {
		// ... (Implementation similar to validation_error in TestCreatePatient)
		patientID := 1
		req := domain.UpdatePatientRequest{DateOfBirth: time.Now()} // Example invalid input

		_, err := svc.UpdatePatient(context.Background(), patientID, req)
		assert.Error(t, err)
		var validationError *domain.ValidationError
		assert.True(t, errors.As(err, &validationError)) // Assert using errors.As
		// Check validationError.Details for specific field error messages...
		assert.Contains(t, validationError.Details, "DateOfBirth must be in the past")

		mockRepo.AssertNotCalled(t, "UpdatePatient")
	})

	// ... (Add more test cases for other error scenarios: not found, repository error, etc.)
	t.Run("not_found_error", func(t *testing.T) {
		// ... (Implementation similar to not_found_error in TestGetPatient, but within UpdatePatient context.)
		patientID := 1
		req := domain.UpdatePatientRequest{FullName: "Updated Name Test"}

		mockRepo.On("GetPatient", mock.Anything, patientID).Return(nil, domain.ErrPatientNotFound) // Updated not found error in mock setup
		_, err := svc.UpdatePatient(context.Background(), patientID, req)                          // Updated call with req

		assert.ErrorIs(t, err, domain.ErrPatientNotFound)
		mockRepo.AssertNotCalled(t, "UpdatePatient") // Ensure update is not called if not found. Updated

	})

	t.Run("repository_error", func(t *testing.T) {
		patientID := 1
		req := domain.UpdatePatientRequest{FullName: "Updated Name Test"}
		mockRepo.On("GetPatient", mock.Anything, patientID).Return(&domain.Patient{PatientID: patientID}, nil)
		mockRepo.On("UpdatePatient", mock.Anything, patientID, mock.AnythingOfType("*domain.Patient")).Return(nil, errors.New("database error"))

		_, err := svc.UpdatePatient(context.Background(), patientID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

	})
}
