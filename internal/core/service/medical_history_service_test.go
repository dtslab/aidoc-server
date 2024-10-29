package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/mocks" // Correct import path
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMedicalHistoryService_CreateMedicalHistoryEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()

	// Initialize mock repository and service instances
	mockRepo := new(mocks.MockMedicalHistoryRepository)
	mockPatientRepo := new(mocks.MockPatientRepository) // Add mock for patientRepo.  Updated
	mockAuth := new(mocks.AuthorizeMock)
	svc := NewMedicalHistoryService(mockRepo, mockPatientRepo, log, v, mockAuth.Authorize)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		req := domain.CreateMedicalHistoryRequest{
			Condition:     "Hypertension",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Some details about the condition",
		}
		expectedEntry := &domain.MedicalHistoryEntry{
			PatientMedicalHistoryID: 1, // Set expected ID
			PatientID:               patientID,
			Condition:               req.Condition,
			DiagnosisDate:           req.DiagnosisDate,
			Status:                  req.Status,
			Details:                 req.Details,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}

		mockPatientRepo.On("GetPatient", mock.Anything, patientID).Return(&domain.Patient{PatientID: patientID}, nil) // Mock patientRepo.GetPatient for success. Updated.
		mockRepo.On("CreateMedicalHistoryEntry", mock.Anything, mock.AnythingOfType("*domain.MedicalHistoryEntry")).Return(expectedEntry, nil)

		entry, err := svc.CreateMedicalHistoryEntry(context.Background(), patientID, req)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry, entry)
		mockRepo.AssertExpectations(t)
		mockPatientRepo.AssertExpectations(t) // Assert patientRepo expectations as well. Updated

	})

	t.Run("validation_error", func(t *testing.T) {
		patientID := 1
		req := domain.CreateMedicalHistoryRequest{
			Condition:     "",         // Missing required field
			DiagnosisDate: time.Now(), // optional
			Status:        "Fake",     // Invalid status
			Details:       "Details",  // optional
		}
		mockPatientRepo.On("GetPatient", mock.Anything, patientID).Return(&domain.Patient{}, nil)

		_, err := svc.CreateMedicalHistoryEntry(context.Background(), patientID, req)

		assert.Error(t, err)
		var validationError *domain.ValidationError // Use type assertion
		assert.True(t, errors.As(err, &validationError))

		assert.Equal(t, "INVALID_MEDICAL_HISTORY_DATA", validationError.Code)
		assert.Contains(t, validationError.Details, "Condition is a required field") // Ensure correct field name
		assert.Contains(t, validationError.Details, "Status must be one of the following")
	})
	// Add test cases for patient not found, and repository error.
	t.Run("patient_not_found", func(t *testing.T) { // Test case when patient is not found
		patientID := 999 // Non-existent patient ID.
		req := domain.CreateMedicalHistoryRequest{
			Condition:     "Some condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Details",
		}

		mockPatientRepo.On("GetPatient", mock.Anything, patientID).Return(nil, domain.ErrPatientNotFound) // Mock patientRepo to return ErrPatientNotFound

		_, err := svc.CreateMedicalHistoryEntry(context.Background(), patientID, req)

		assert.ErrorIs(t, err, domain.ErrPatientNotFound) // Use assert.ErrorIs

		mockRepo.AssertNotCalled(t, "CreateMedicalHistoryEntry") // Ensure medicalHistoryRepo is not called.

	})

	t.Run("repository_error", func(t *testing.T) {
		patientID := 1
		req := domain.CreateMedicalHistoryRequest{
			Condition:     "Some condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Details",
		}
		mockPatientRepo.On("GetPatient", mock.Anything, patientID).Return(&domain.Patient{}, nil)
		mockRepo.On("CreateMedicalHistoryEntry", mock.Anything, mock.AnythingOfType("*domain.MedicalHistoryEntry")).Return(nil, errors.New("database error"))

		_, err := svc.CreateMedicalHistoryEntry(context.Background(), patientID, req)

		assert.Error(t, err)                                 // Use assert.Error. Updated.
		assert.Contains(t, err.Error(), "database error")    // Correctly checks database error. Updated.
		assert.NotErrorIs(t, err, sql.ErrNoRows)             // Ensure it is not NoRows error. Updated
		assert.NotErrorIs(t, err, domain.ErrPatientNotFound) // Ensure error is not PatientNotFound. Updated.

	})
}

func TestMedicalHistoryService_GetMedicalHistoryEntries(t *testing.T) { // Correct function name
	log := zap.NewNop()
	v := validator.New()

	mockMedicalHistoryRepo := new(mocks.MockMedicalHistoryRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)                                                  // Mocked PatientRepository. Updated.
	mockAuth := new(mocks.AuthorizeMock)                                                                 // Corrected the type.
	svc := NewMedicalHistoryService(mockMedicalHistoryRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject dependencies. Corrected.

	t.Run("success", func(t *testing.T) {
		ctx := context.Background() // Add context for GetPatient call
		patientID := 1
		expectedEntries := []*domain.MedicalHistoryEntry{
			{PatientMedicalHistoryID: 1, Condition: "Condition 1"},
			{PatientMedicalHistoryID: 2, Condition: "Condition 2"},
		}
		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil)                    // Mock GetPatient call. Updated.
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntries", ctx, patientID).Return(expectedEntries, nil) // Updated to mockMedicalHistoryRepo

		entries, err := svc.GetMedicalHistoryEntries(ctx, patientID)

		assert.NoError(t, err)
		assert.Equal(t, expectedEntries, entries)
		mockMedicalHistoryRepo.AssertExpectations(t) // Assert expectations on the correct mock repo
	})

	t.Run("patient_not_found", func(t *testing.T) { // Updated to match the service's error handling
		ctx := context.Background()
		patientID := 999                                                                        // non-existent patient
		mockPatientRepo.On("GetPatient", ctx, patientID).Return(nil, domain.ErrPatientNotFound) // Setup mock to return ErrPatientNotFound

		_, err := svc.GetMedicalHistoryEntries(context.Background(), patientID)

		assert.ErrorIs(t, err, domain.ErrPatientNotFound) // Updated assertion

		mockMedicalHistoryRepo.AssertNotCalled(t, "GetMedicalHistoryEntries") // Ensure GetMedicalHistoryEntries was not called
	})

	t.Run("no_medical_history_found", func(t *testing.T) { // Added test case for no medical history
		ctx := context.Background()
		patientID := 1
		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{PatientID: patientID}, nil)                      // Updated
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntries", ctx, patientID).Return(nil, domain.ErrMedicalHistoryEntryNotFound) // Return empty slice for not found

		entries, err := svc.GetMedicalHistoryEntries(ctx, patientID) // Updated

		assert.Nil(t, entries)                                        // Expecting nil entries. Updated.
		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Expecting not found error. Updated.

	})

	t.Run("repository_error", func(t *testing.T) {
		patientID := 1
		ctx := context.Background() // Add context for GetPatient call

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil) // Mock patientRepo for success. Updated
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntries", ctx, patientID).
			Return(nil, errors.New("database error")) // Correct mock method. Updated.

		_, err := svc.GetMedicalHistoryEntries(context.Background(), patientID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

func TestGetMedicalHistoryEntry(t *testing.T) { // Correct test function name to align with service method. Updated.
	log := zap.NewNop()
	v := validator.New()
	mockMedicalHistoryRepo := new(mocks.MockMedicalHistoryRepository) // Use mocks.MockMedicalHistoryRepository
	mockPatientRepo := new(mocks.MockPatientRepository)               // Add mockPatientRepo. Updated
	mockAuth := new(mocks.AuthorizeMock)

	svc := NewMedicalHistoryService(mockMedicalHistoryRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject mock

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1

		expectedEntry := &domain.MedicalHistoryEntry{PatientMedicalHistoryID: entryID} // Initialize expectedEntry. Updated.
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(expectedEntry, nil)

		entry, err := svc.GetMedicalHistoryEntry(ctx, entryID) // Call service method. Updated.

		assert.NoError(t, err)
		assert.Equal(t, expectedEntry, entry) // Correct assertion. Updated.

	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999 // Non-existent entryID
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).
			Return(nil, domain.ErrMedicalHistoryEntryNotFound) // Correctly mocked repository to return entry not found error. Updated.

		entry, err := svc.GetMedicalHistoryEntry(ctx, entryID) // Call service method. Updated.

		assert.Nil(t, entry)                                          // Assert that the entry is nil. Updated.
		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Correctly use assert.ErrorIs. Updated. Corrected expected error type.

	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(nil, errors.New("database error"))
		_, err := svc.GetMedicalHistoryEntry(ctx, entryID)

		assert.Error(t, err)                                 // Use assert.Error. Updated
		assert.Contains(t, err.Error(), "database error")    // Correctly checks database error. Updated
		assert.NotErrorIs(t, err, sql.ErrNoRows)             // Ensure it's not sql.ErrNoRows. Updated.
		assert.NotErrorIs(t, err, domain.ErrPatientNotFound) // Ensure error is not domain.ErrPatientNotFound. Updated.

	})
}

func TestUpdateMedicalHistoryEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()

	mockMedicalHistoryRepo := new(mocks.MockMedicalHistoryRepository) // Use shared mock. Updated
	mockPatientRepo := new(mocks.MockPatientRepository)               // Add mockPatientRepo. Updated.
	mockAuth := new(mocks.AuthorizeMock)                              // Correct mock type. Updated

	svc := NewMedicalHistoryService(mockMedicalHistoryRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Correct svc instantiation. Updated.

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		req := domain.UpdateMedicalHistoryRequest{Condition: "Updated Condition"}

		// Create an existing entry for the mock to return
		existingEntry := &domain.MedicalHistoryEntry{PatientMedicalHistoryID: entryID, PatientID: patientID, Condition: "Original Condition"}
		// Expected updated entry
		expectedEntry := &domain.MedicalHistoryEntry{PatientMedicalHistoryID: entryID, PatientID: patientID, Condition: "Updated Condition", UpdatedAt: time.Now()}

		// Mock GetMedicalHistoryEntry to return the existingEntry
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(existingEntry, nil)
		// Mock UpdateMedicalHistoryEntry to return the expectedEntry
		mockMedicalHistoryRepo.On("UpdateMedicalHistoryEntry", ctx, entryID, mock.AnythingOfType("*domain.MedicalHistoryEntry")).Return(expectedEntry, nil)

		// Correctly mock the authorization check for the associated PatientID
		mockAuth.On("Authorize", ctx, patientID).Return(true)

		// Call the service method
		entry, err := svc.UpdateMedicalHistoryEntry(ctx, entryID, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry, entry)
		mockMedicalHistoryRepo.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("invalid_input", func(t *testing.T) { // Add invalid input test case
		ctx := context.Background()
		entryID := 1
		req := domain.UpdateMedicalHistoryRequest{Status: "Invalid"} // Invalid status

		_, err := svc.UpdateMedicalHistoryEntry(ctx, entryID, req)

		assert.Error(t, err) // Expect an error
		var validationErr *domain.ValidationError
		assert.True(t, errors.As(err, &validationErr)) // Assert it's a validation error. Updated
		assert.Equal(t, "INVALID_MEDICAL_HISTORY_DATA", validationErr.Code)
		assert.Contains(t, validationErr.Details, "Status must be one of the following")

		mockMedicalHistoryRepo.AssertNotCalled(t, "GetMedicalHistoryEntry")
		mockMedicalHistoryRepo.AssertNotCalled(t, "UpdateMedicalHistoryEntry")
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999 // Non-existent Entry ID.
		patientID := 1 // Declare patientID
		req := domain.UpdateMedicalHistoryRequest{Condition: "New Condition"}

		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).
			Return(nil, domain.ErrMedicalHistoryEntryNotFound) // Correctly mock repo's return error

		//Simulate an authorization check using the provided patientID.
		//Even if the entry is not found, the authorization would be checked first.
		mockAuth.On("Authorize", ctx, patientID).Return(true) // Assuming authorization succeeds.

		_, err := svc.UpdateMedicalHistoryEntry(ctx, entryID, req)

		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Assert the correct error. Updated. Corrected error type.

		mockMedicalHistoryRepo.AssertNotCalled(t, "UpdateMedicalHistoryEntry")
		mockAuth.AssertExpectations(t) // Assert that Authorize was called.
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 2                                                             // Patient ID for access check. Update for authorization logic.
		req := domain.UpdateMedicalHistoryRequest{Condition: "New Condition Test"} // Updated.

		mockMedicalHistoryRepo := new(mocks.MockMedicalHistoryRepository) // Initialize mockMedicalHistoryRepo

		existingEntry := &domain.MedicalHistoryEntry{
			PatientMedicalHistoryID: entryID,
			PatientID:               patientID, // Important: set the PatientID in existingEntry. Updated.

			Condition: "Some condition",
		}

		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(existingEntry, nil) // Setup mock for existing entry

		mockAuth := new(mocks.AuthorizeMock)                   // Use new mock instance. Updated. Very important!
		mockAuth.On("Authorize", ctx, patientID).Return(false) // Use mockAuth. Updated

		svc.authorize = mockAuth.Authorize                         // Inject the authorization mock for this test. Corrected.
		_, err := svc.UpdateMedicalHistoryEntry(ctx, entryID, req) // Updated.

		assert.ErrorIs(t, err, domain.ErrForbidden) // Assert forbidden error. Updated. Corrected error type.

		mockMedicalHistoryRepo.AssertNotCalled(t, "UpdateMedicalHistoryEntry") // Ensure update is not called when unauthorized. Updated.
		mockAuth.AssertExpectations(t)                                         // Ensure Authorize was called and expectations met. Updated

	})

	t.Run("repository_error", func(t *testing.T) { // Added test case for repository error.
		ctx := context.Background()
		entryID := 1
		patientID := 1
		req := domain.UpdateMedicalHistoryRequest{Condition: "New Condition Test"} // Updated
		existingEntry := &domain.MedicalHistoryEntry{PatientMedicalHistoryID: entryID, PatientID: patientID, Condition: "Original"}
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(existingEntry, nil)                                                                       // Setup mock to retrieve an existing entry. Updated
		mockMedicalHistoryRepo.On("UpdateMedicalHistoryEntry", ctx, entryID, mock.AnythingOfType("*domain.MedicalHistoryEntry")).Return(nil, errors.New("database error")) // Mock an error from update. Updated
		mockAuth.On("Authorize", ctx, patientID).Return(true)

		_, err := svc.UpdateMedicalHistoryEntry(ctx, entryID, req) // Updated

		assert.Error(t, err) // Correctly checks for a generic error. Updated
		assert.Contains(t, err.Error(), "database error")
		mockAuth.AssertExpectations(t) // Ensure Authorize was called and expectations met. Updated

	})
}

func TestDeleteMedicalHistoryEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockMedicalHistoryRepo := new(mocks.MockMedicalHistoryRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)                                                  // Mock patient repository for authorization checks.
	mockAuth := new(mocks.AuthorizeMock)                                                                 // Initialize mockAuth.
	svc := NewMedicalHistoryService(mockMedicalHistoryRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject AuthorizeMock

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(&domain.MedicalHistoryEntry{PatientID: patientID}, nil) // Corrected to *domain.MedicalHistoryEntry. Updated
		mockAuth.On("Authorize", ctx, patientID).Return(true)                                                                            // Expect authorization to succeed
		mockMedicalHistoryRepo.On("DeleteMedicalHistoryEntry", ctx, entryID).Return(nil)                                                 // Mock successful deletion. Updated.

		err := svc.DeleteMedicalHistoryEntry(ctx, entryID)

		assert.NoError(t, err) // Expect no error
		mockAuth.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999 // Non-existent entry
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).
			Return(nil, domain.ErrMedicalHistoryEntryNotFound) // Correctly mock repo's return error

		err := svc.DeleteMedicalHistoryEntry(ctx, entryID)

		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Assert the correct error. Updated. Corrected error type.

		mockMedicalHistoryRepo.AssertNotCalled(t, "DeleteMedicalHistoryEntry")
		mockAuth.AssertNotCalled(t, "Authorize") // Assert authorize not called
	})

	t.Run("unauthorized", func(t *testing.T) { // Added test for authorization check.
		ctx := context.Background()
		entryID := 1
		patientID := 2                                                                                                                   // User not authorized to access this entry. Updated
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(&domain.MedicalHistoryEntry{PatientID: patientID}, nil) // Added return existingEntry. Updated
		mockAuth.On("Authorize", ctx, patientID).Return(false)                                                                           // Correct arguments. Updated.

		err := svc.DeleteMedicalHistoryEntry(ctx, entryID) // Correct method name

		assert.ErrorIs(t, err, domain.ErrForbidden)                            // Correct error assertion. Updated. Corrected error type.
		mockMedicalHistoryRepo.AssertNotCalled(t, "DeleteMedicalHistoryEntry") // Correct mock method name. Updated.
		mockAuth.AssertExpectations(t)                                         // Assert the expectations on mockAuth are met. Updated
	})

	t.Run("repository_error", func(t *testing.T) { // Add database error test case
		ctx := context.Background()
		entryID := 1
		patientID := 1
		mockMedicalHistoryRepo.On("GetMedicalHistoryEntry", ctx, entryID).Return(&domain.MedicalHistoryEntry{PatientID: patientID}, nil) // Updated.
		mockAuth.On("Authorize", ctx, patientID).Return(true)                                                                            // Expect authorization to succeed. Updated
		mockMedicalHistoryRepo.On("DeleteMedicalHistoryEntry", ctx, entryID).Return(errors.New("database error"))                        // Correct mock method name. Updated.

		err := svc.DeleteMedicalHistoryEntry(ctx, entryID)

		assert.Error(t, err) // Correct error check. Updated
		assert.Contains(t, err.Error(), "database error")
		mockAuth.AssertExpectations(t) // Ensure Authorize was called and expectations met. Updated

	})
}
