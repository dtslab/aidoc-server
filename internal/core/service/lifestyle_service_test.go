package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestCreateLifestyleEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockLifestyleRepo := new(mocks.MockLifestyleRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)
	mockAuth := new(mocks.AuthorizeMock)                                                       // Create mock for auth.Authorize. Corrected.
	svc := NewLifestyleService(mockLifestyleRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject the mock. Corrected

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		patientID := 1

		req := domain.CreateLifestyleRequest{
			LifestyleFactor: "Tobacco Use",
			Value:           "Never Smoked",
		}

		expectedEntry := &domain.LifestyleEntry{
			PatientLifestyleID: 1,
			PatientID:          patientID,
			LifestyleFactor:    req.LifestyleFactor,
			Value:              req.Value,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil)
		mockLifestyleRepo.On("CreateLifestyleEntry", ctx, mock.AnythingOfType("*domain.LifestyleEntry")).Return(expectedEntry, nil)

		entry, err := svc.CreateLifestyleEntry(ctx, patientID, req)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry, entry)

		mockLifestyleRepo.AssertExpectations(t)
		mockPatientRepo.AssertExpectations(t)
	})

	t.Run("invalid_input", func(t *testing.T) {
		ctx := context.Background()
		patientID := 1
		req := domain.CreateLifestyleRequest{
			LifestyleFactor: "", // Missing required field
			Value:           "Some value",
		}
		_, err := svc.CreateLifestyleEntry(ctx, patientID, req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("patient_not_found", func(t *testing.T) {
		ctx := context.Background()
		patientID := 999 // Non-existent patient ID
		req := domain.CreateLifestyleRequest{
			LifestyleFactor: "Valid Factor",
			Value:           "Valid Value",
		}

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(nil, domain.ErrPatientNotFound)

		_, err := svc.CreateLifestyleEntry(ctx, patientID, req)

		assert.ErrorIs(t, err, domain.ErrPatientNotFound)

		mockLifestyleRepo.AssertNotCalled(t, "CreateLifestyleEntry")
	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		patientID := 1
		req := domain.CreateLifestyleRequest{
			LifestyleFactor: "Tobacco Use",
			Value:           "Never Smoked",
		}
		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil)
		mockLifestyleRepo.On("CreateLifestyleEntry", ctx, mock.AnythingOfType("*domain.LifestyleEntry")).Return(nil, errors.New("database error"))
		_, err := svc.CreateLifestyleEntry(ctx, patientID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

	})
}

func TestGetLifestyleEntries(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockLifestyleRepo := new(mocks.MockLifestyleRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)
	mockAuth := new(mocks.AuthorizeMock)                                                       // Initialize mockAuth. Corrected.
	svc := NewLifestyleService(mockLifestyleRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject mock. Corrected

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		patientID := 1
		expectedEntries := []*domain.LifestyleEntry{
			{PatientLifestyleID: 1, LifestyleFactor: "Factor 1"},
			{PatientLifestyleID: 2, LifestyleFactor: "Factor 2"},
		}

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil)
		mockLifestyleRepo.On("GetLifestyleEntries", ctx, patientID).Return(expectedEntries, nil)

		entries, err := svc.GetLifestyleEntries(ctx, patientID)

		assert.NoError(t, err)
		assert.Equal(t, expectedEntries, entries)

		mockLifestyleRepo.AssertExpectations(t)
		mockPatientRepo.AssertExpectations(t)
	})

	t.Run("invalid_patient_id", func(t *testing.T) {
		ctx := context.Background()
		patientID := -1 // Invalid patient ID.

		_, err := svc.GetLifestyleEntries(ctx, patientID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check patient existence")

		mockPatientRepo.AssertNotCalled(t, "GetPatient")            // Pastikan GetPatient tidak dipanggil. Updated
		mockLifestyleRepo.AssertNotCalled(t, "GetLifestyleEntries") // Pastikan GetLifestyleEntries tidak dipanggil. Updated

	})

	t.Run("patient_not_found", func(t *testing.T) {
		ctx := context.Background()
		patientID := 999 // Non-existent patient ID.

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(nil, domain.ErrPatientNotFound)

		_, err := svc.GetLifestyleEntries(ctx, patientID)

		assert.ErrorIs(t, err, domain.ErrPatientNotFound)

		mockLifestyleRepo.AssertNotCalled(t, "GetLifestyleEntries") // Verify repo method was not called.

	})

	t.Run("no_lifestyle_entries_found", func(t *testing.T) { // Test case for empty result
		ctx := context.Background()
		patientID := 1

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil) // Mock successful patient retrieval
		mockLifestyleRepo.On("GetLifestyleEntries", ctx, patientID).Return(nil, domain.ErrLifestyleEntryNotFound)

		entries, err := svc.GetLifestyleEntries(ctx, patientID)

		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Use assert.ErrorIs to check for specific error type

		assert.Nil(t, entries) // No entries expected

		mockLifestyleRepo.AssertExpectations(t)
		mockPatientRepo.AssertExpectations(t)

	})

	t.Run("internal_server_error", func(t *testing.T) {
		ctx := context.Background()
		patientID := 1

		mockPatientRepo.On("GetPatient", ctx, patientID).Return(&domain.Patient{}, nil)
		mockLifestyleRepo.On("GetLifestyleEntries", ctx, patientID).Return(nil, errors.New("database error"))

		_, err := svc.GetLifestyleEntries(ctx, patientID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

	})
}

func TestGetLifestyleEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockLifestyleRepo := new(mocks.MockLifestyleRepository) // Use mocks.MockLifestyleRepository from shared mocks
	mockPatientRepo := new(mocks.MockPatientRepository)     // Same here

	mockAuth := new(mocks.AuthorizeMock) // Initialize AuthorizeMock. Updated.

	svc := NewLifestyleService(mockLifestyleRepo, mockPatientRepo, log, v, mockAuth.Authorize) // Inject the mock

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		expectedEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, LifestyleFactor: "Factor 1"}

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(expectedEntry, nil)

		entry, err := svc.GetLifestyleEntry(ctx, entryID)

		assert.NoError(t, err)
		assert.Equal(t, expectedEntry, entry)
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999 // Non-existent ID

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(nil, domain.ErrLifestyleEntryNotFound)

		entry, err := svc.GetLifestyleEntry(ctx, entryID)

		assert.Nil(t, entry)                                     // Check entry is nil (not found). Updated.
		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Use assert.ErrorIs. Updated.

	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(nil, errors.New("database error"))

		_, err := svc.GetLifestyleEntry(ctx, entryID)

		assert.Error(t, err)                              // Correct assertion to check for an error. Updated.
		assert.Contains(t, err.Error(), "database error") // Correctly checks for repository error. Updated.

	})

}

func TestUpdateLifestyleEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockLifestyleRepo := new(mocks.MockLifestyleRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)
	mockAuth := new(mocks.AuthorizeMock)
	svc := NewLifestyleService(mockLifestyleRepo, mockPatientRepo, log, v, mockAuth.Authorize)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		req := domain.UpdateLifestyleRequest{
			LifestyleFactor: "Updated Factor",
		}
		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID, LifestyleFactor: "Original Factor"}
		expectedEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID, LifestyleFactor: req.LifestyleFactor, UpdatedAt: time.Now()}

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil)
		mockLifestyleRepo.On("UpdateLifestyleEntry", ctx, entryID, mock.AnythingOfType("*domain.LifestyleEntry")).Return(expectedEntry, nil)
		mockAuth.On("Authorize", ctx, patientID).Return(true) // Expect Authorize call

		entry, err := svc.UpdateLifestyleEntry(ctx, entryID, req)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry.LifestyleFactor, entry.LifestyleFactor)

		mockLifestyleRepo.AssertExpectations(t)
		mockAuth.AssertExpectations(t)
	})

	t.Run("invalid_input", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		req := domain.UpdateLifestyleRequest{
			LifestyleFactor: "", // Invalid, missing required field
		}

		_, err := svc.UpdateLifestyleEntry(ctx, entryID, req)

		assert.ErrorIs(t, err, domain.ErrInvalidInput)

		mockLifestyleRepo.AssertNotCalled(t, "GetLifestyleEntry")
		mockLifestyleRepo.AssertNotCalled(t, "UpdateLifestyleEntry")
	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999 // Non-existent entry ID.
		req := domain.UpdateLifestyleRequest{
			LifestyleFactor: "Updated Factor",
		}

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(nil, domain.ErrLifestyleEntryNotFound)

		_, err := svc.UpdateLifestyleEntry(ctx, entryID, req)

		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound)

		mockLifestyleRepo.AssertNotCalled(t, "UpdateLifestyleEntry")

	})

	t.Run("unauthorized", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 2 // Different patient ID for unauthorized case
		req := domain.UpdateLifestyleRequest{
			LifestyleFactor: "Some Factor",
		}

		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID} // Existing entry for auth check

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil) // Entry exists, but user is not authorized. Updated.

		mockAuth := new(mocks.AuthorizeMock) // Use a *new* mock instance for this test case!  This is very important.

		mockAuth.On("Authorize", ctx, patientID).Return(false) // Correct mock method name and arguments. Updated.

		svc.authorize = mockAuth.Authorize // Inject the mock. Updated

		_, err := svc.UpdateLifestyleEntry(ctx, entryID, req) // Corrected svc variable name and method call

		assert.ErrorIs(t, err, domain.ErrForbidden) // Correct error type assertion

		mockLifestyleRepo.AssertNotCalled(t, "UpdateLifestyleEntry") // Check that repository method was *not* called. Updated

		mockAuth.AssertExpectations(t)

	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		req := domain.UpdateLifestyleRequest{
			LifestyleFactor: "Updated Factor",
		}

		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID, LifestyleFactor: "Original Factor"}

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil)
		mockLifestyleRepo.On("UpdateLifestyleEntry", ctx, entryID, mock.AnythingOfType("*domain.LifestyleEntry")).Return(nil, errors.New("database error"))

		mockAuth.On("Authorize", ctx, patientID).Return(true)
		svc.authorize = mockAuth.Authorize

		_, err := svc.UpdateLifestyleEntry(ctx, entryID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		mockAuth.AssertExpectations(t)

	})
}

func TestDeleteLifestyleEntry(t *testing.T) {
	log := zap.NewNop()
	v := validator.New()
	mockLifestyleRepo := new(mocks.MockLifestyleRepository)
	mockPatientRepo := new(mocks.MockPatientRepository)
	mockAuth := new(mocks.AuthorizeMock) // Moved mockAuth initialization outside t.Run blocks. Updated

	svc := NewLifestyleService(mockLifestyleRepo, mockPatientRepo, log, v, mockAuth.Authorize)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID} // Declare existingEntry

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil)
		mockLifestyleRepo.On("DeleteLifestyleEntry", ctx, entryID).Return(nil)
		mockAuth.On("Authorize", ctx, patientID).Return(true) // Use mockAuth in all test cases. Updated
		svc.authorize = mockAuth.Authorize                    // Inject the mock authorize function

		err := svc.DeleteLifestyleEntry(ctx, entryID)

		assert.NoError(t, err)
		mockLifestyleRepo.AssertExpectations(t)
		mockAuth.AssertExpectations(t)

	})

	t.Run("not_found", func(t *testing.T) {
		ctx := context.Background()
		entryID := 999

		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(nil, domain.ErrLifestyleEntryNotFound)

		err := svc.DeleteLifestyleEntry(ctx, entryID)

		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound)

		mockLifestyleRepo.AssertNotCalled(t, "DeleteLifestyleEntry")
	})
	t.Run("unauthorized", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 2                                                                             // Use a different patientID to test unauthorized access
		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID} // Create existing entry. Updated
		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil)         // Mock to return existing entry
		mockAuth.On("Authorize", ctx, patientID).Return(false)                                     // Use mockAuth and correct arguments. Updated

		svc.authorize = mockAuth.Authorize // Inject the mock authorize function. Corrected

		err := svc.DeleteLifestyleEntry(ctx, entryID)

		assert.ErrorIs(t, err, domain.ErrForbidden)                  // Correctly assert forbidden error. Updated
		mockLifestyleRepo.AssertNotCalled(t, "DeleteLifestyleEntry") // Ensure DeleteLifestyleEntry is not called. Updated
		mockAuth.AssertExpectations(t)                               // Assert mock expectations
	})

	t.Run("repository_error", func(t *testing.T) {
		ctx := context.Background()
		entryID := 1
		patientID := 1
		existingEntry := &domain.LifestyleEntry{PatientLifestyleID: entryID, PatientID: patientID}      // Updated
		mockLifestyleRepo.On("GetLifestyleEntry", ctx, entryID).Return(existingEntry, nil)              // Corrected mock setup. Updated
		mockLifestyleRepo.On("DeleteLifestyleEntry", ctx, entryID).Return(errors.New("database error")) // Mock a repository error.
		mockAuth.On("Authorize", ctx, patientID).Return(true)

		svc.authorize = mockAuth.Authorize // Inject the mock authorize function for this test case

		err := svc.DeleteLifestyleEntry(ctx, entryID)

		assert.Error(t, err)                              // Correctly checks for an error. Updated.
		assert.Contains(t, err.Error(), "database error") // Correctly checks for the repository error. Updated

	})
}
