package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockMedicalHistoryService mocks the MedicalHistoryService interface
type MockMedicalHistoryService struct {
	mock.Mock
}

// CreateMedicalHistoryEntry mocks the CreateMedicalHistoryEntry method
func (m *MockMedicalHistoryService) CreateMedicalHistoryEntry(ctx context.Context, patientID int, req domain.CreateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, patientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalHistoryEntry), args.Error(1)

}

// GetMedicalHistoryEntries mocks the GetMedicalHistoryEntries method
func (m *MockMedicalHistoryService) GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, patientID)
	return args.Get(0).([]*domain.MedicalHistoryEntry), args.Error(1)
}

// UpdateMedicalHistoryEntry mocks UpdateMedicalHistoryEntry
func (m *MockMedicalHistoryService) UpdateMedicalHistoryEntry(ctx context.Context, entryID int, req domain.UpdateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, entryID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalHistoryEntry), args.Error(1)
}

// DeleteMedicalHistoryEntry mocks the DeleteMedicalHistoryEntry method. Updated.
func (m *MockMedicalHistoryService) DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error {
	args := m.Called(ctx, entryID)
	return args.Error(0)
}

func TestCreateMedicalHistoryEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockMedicalHistoryService)
	handler := NewMedicalHistoryHandler(mockSvc, log)

	t.Run("valid_input", func(t *testing.T) {
		patientID := 1
		reqBody := domain.CreateMedicalHistoryRequest{
			Condition:     "Test Condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Test Details",
		}
		expectedEntry := &domain.MedicalHistoryEntry{
			PatientMedicalHistoryID: 1,
			PatientID:               patientID,
			Condition:               reqBody.Condition,
			DiagnosisDate:           reqBody.DiagnosisDate,
			Status:                  reqBody.Status,
			Details:                 reqBody.Details,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		}
		mockSvc.On("CreateMedicalHistoryEntry", mock.Anything, patientID, reqBody).Return(expectedEntry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/medical_history", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createdEntry domain.MedicalHistoryEntry
		_ = json.Unmarshal(w.Body.Bytes(), &createdEntry)
		assert.Equal(t, expectedEntry, &createdEntry)
	})

	t.Run("invalid_patient_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/invalid/medical_history", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: "invalid"}} // Updated

		handler.CreateMedicalHistoryEntry(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Invalid patient ID", errResp.Error)
	})
	t.Run("invalid_input", func(t *testing.T) {
		patientID := 1
		reqBody := domain.CreateMedicalHistoryRequest{
			Condition: "", // Invalid: Missing condition
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/medical_history", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		mockSvc.On("CreateMedicalHistoryEntry", mock.Anything, patientID, reqBody).Return(nil, domain.ErrInvalidMedicalHistoryData)

		handler.CreateMedicalHistoryEntry(c) // Actual call

		assert.Equal(t, http.StatusBadRequest, w.Code) // Assert correct status code
		var errResp domain.ErrorResponse

		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)                                                                // Ensure no error during unmarshaling
		assert.Equal(t, domain.ErrorResponse{Error: "invalid medical history data"}, errResp) // Assert expected error message.

	})

	t.Run("patient_not_found", func(t *testing.T) {
		patientID := 999
		reqBody := domain.CreateMedicalHistoryRequest{
			Condition:     "Valid Condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Valid Details",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/999/medical_history", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}} // Updated

		mockSvc.On("CreateMedicalHistoryEntry", mock.Anything, patientID, reqBody).
			Return(nil, domain.ErrPatientNotFound) // Correctly mocked service to return patient not found error. Updated.

		handler.CreateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "patient not found"}, errResp) // Assert expected error message.

	})

	t.Run("internal_server_error", func(t *testing.T) {
		// Test case setup
		patientID := 1
		reqBody := domain.CreateMedicalHistoryRequest{
			Condition:     "Test Condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Test Details",
		}

		mockSvc.On("CreateMedicalHistoryEntry", mock.Anything, patientID, reqBody).
			Return(nil, errors.New("database error")) // Mock the service to return an error. Updated.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		requestBodyBytes, _ := json.Marshal(reqBody)

		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/medical_history", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "Failed to create medical history entry"}, errResp) // Expected message from the handler. Updated.
	})
}

func TestGetMedicalHistoryEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockMedicalHistoryService)
	handler := NewMedicalHistoryHandler(mockSvc, log)
	t.Run("valid_patient_id", func(t *testing.T) { // Corrected naming
		patientID := 1
		expectedEntries := []*domain.MedicalHistoryEntry{
			{
				PatientMedicalHistoryID: 1,
				PatientID:               patientID,
				Condition:               "Condition 1",
				DiagnosisDate:           time.Now(),
				Status:                  "Active",
				Details:                 "Details 1",
			},
		}

		mockSvc.On("GetMedicalHistoryEntries", mock.Anything, patientID).Return(expectedEntries, nil) // Correct mock method and args. Updated.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/medical_history", nil) // Updated path. Updated.
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}           // Corrected param setup. Updated.

		handler.GetMedicalHistoryEntries(c)

		assert.Equal(t, http.StatusOK, w.Code) // Correct HTTP status. Updated.

		var actualEntries []*domain.MedicalHistoryEntry                           // Corrected variable name. Updated.
		err := json.Unmarshal(w.Body.Bytes(), &actualEntries)                     // Use json.Unmarshal on the body. Updated
		assert.NoError(t, err)                                                    // Check for unmarshal errors. Updated.
		assert.Equal(t, expectedEntries, actualEntries)                           // Correct assertion. Updated.
		assert.Equal(t, expectedEntries[0].Condition, actualEntries[0].Condition) // Access Condition correctly. Updated.
	})

	t.Run("invalid_patient_id", func(t *testing.T) { // Add test case for invalid patient ID
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/invalid/medical_history", nil) // Correct path. Updated.
		c.Params = []gin.Param{{Key: "patient_id", Value: "invalid"}}                               // Add parameter for invalid ID. Updated.

		handler.GetMedicalHistoryEntries(c) // Actual call

		assert.Equal(t, http.StatusBadRequest, w.Code) // Expecting 400. Updated.
		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp) // Unmarshal body to ErrorResponse. Updated.

		assert.NoError(t, err)                                                      // Ensure no unmarshaling error. Updated.
		assert.Equal(t, errResp, domain.ErrorResponse{Error: "Invalid patient ID"}) // Check error message. Updated.

	})

	t.Run("no_medical_history_found", func(t *testing.T) {
		patientID := 1

		mockSvc.On("GetMedicalHistoryEntries", mock.Anything, patientID).
			Return([]*domain.MedicalHistoryEntry{}, domain.ErrMedicalHistoryEntryNotFound) // Mock not found error case

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/medical_history", nil) // Correct path. Updated.

		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}} // Add params for a valid ID. Updated.

		handler.GetMedicalHistoryEntries(c)

		assert.Equal(t, http.StatusNotFound, w.Code) // Should be StatusNotFound (404) for consistency. Updated

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "medical history entry not found"}, errResp) // Check correct error. Updated.

	})

	t.Run("patient_not_found", func(t *testing.T) { // Changed test case name to be more descriptive. Updated.
		patientID := 999 // Non-existent patient ID
		mockSvc.On("GetMedicalHistoryEntries", mock.Anything, patientID).
			Return(nil, domain.ErrPatientNotFound) // Mock not found error case. Updated.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/999/medical_history", nil) // Correct path and non-existent ID. Updated.
		c.Params = []gin.Param{{Key: "patient_id", Value: "999"}}                               // Non-existent ID param. Updated.

		handler.GetMedicalHistoryEntries(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)                            // Use json.Unmarshal. Updated
		assert.NoError(t, err)                                                     // Assert no error during unmarshal. Updated
		assert.Equal(t, errResp, domain.ErrorResponse{Error: "patient not found"}) // Check the correct error response. Updated

	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1

		mockSvc.On("GetMedicalHistoryEntries", mock.Anything, patientID).
			Return(nil, errors.New("database error")) // Mock error. Updated

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/medical_history", nil) // Correct path
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}           // Add param for a valid ID. Updated

		handler.GetMedicalHistoryEntries(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Correct status code. Updated

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "Failed to get medical history entries"}, errResp) // Check correct error. Updated
	})
	// ... (Test cases for invalid input, server errors, etc.)
}

func TestUpdateMedicalHistoryEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockMedicalHistoryService)
	handler := NewMedicalHistoryHandler(mockSvc, log)

	t.Run("valid_input", func(t *testing.T) {
		patientID := 1
		entryID := 1
		reqBody := domain.UpdateMedicalHistoryRequest{
			Condition: "Updated Condition",
			Status:    "Resolved",
		}
		expectedEntry := &domain.MedicalHistoryEntry{
			PatientMedicalHistoryID: entryID,
			PatientID:               patientID,
			Condition:               reqBody.Condition,
			Status:                  reqBody.Status,
			UpdatedAt:               time.Now(),
		}

		mockSvc.On("UpdateMedicalHistoryEntry", mock.Anything, entryID, reqBody).Return(expectedEntry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/medical_history/1", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)},
		}

		handler.UpdateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedEntry domain.MedicalHistoryEntry
		_ = json.Unmarshal(w.Body.Bytes(), &updatedEntry)
		assert.Equal(t, expectedEntry, &updatedEntry)
	})

	t.Run("invalid_medical_history_id", func(t *testing.T) { // Updated test case name
		patientID := 1

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/medical_history/invalid", nil) // Updated path. Updated
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: "invalid"},
		}

		handler.UpdateMedicalHistoryEntry(c) // Actual call

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)                                                              // Ensure no unmarshaling error. Updated
		assert.Equal(t, domain.ErrorResponse{Error: "Invalid medical history ID"}, errResp) // Updated
	})

	t.Run("invalid_input", func(t *testing.T) { // Added test for invalid input.
		patientID := 1
		entryID := 1
		reqBody := domain.UpdateMedicalHistoryRequest{
			Status: "InvalidStatus", // Example: invalid status value
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/medical_history/1", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)},
		}

		mockSvc.On("UpdateMedicalHistoryEntry", mock.Anything, entryID, reqBody).Return(nil, domain.ErrInvalidMedicalHistoryData)

		handler.UpdateMedicalHistoryEntry(c) // Actual call

		assert.Equal(t, http.StatusBadRequest, w.Code) // Assert correct status code
		var errResp domain.ErrorResponse

		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)                                                                // Ensure no unmarshaling error. Updated
		assert.Equal(t, domain.ErrorResponse{Error: "invalid medical history data"}, errResp) // Assert expected error message.

	})

	t.Run("medical_history_entry_not_found", func(t *testing.T) {
		patientID := 1
		entryID := 999 // Non-existent entry ID
		reqBody := domain.UpdateMedicalHistoryRequest{Condition: "New Condition"}

		mockSvc.On("UpdateMedicalHistoryEntry", mock.Anything, entryID, reqBody).Return(nil, domain.ErrMedicalHistoryEntryNotFound) // Mock not found error. Updated.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)

		c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/v1/patients/%d/medical_history/%d", patientID, entryID), bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)},
		}

		handler.UpdateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "medical history entry not found"}, errResp) // Assert expected error message.

	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1
		entryID := 1
		reqBody := domain.UpdateMedicalHistoryRequest{Condition: "New Condition"}
		// Mock service to return an internal server error
		mockSvc.On("UpdateMedicalHistoryEntry", mock.Anything, entryID, reqBody).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		requestBodyBytes, _ := json.Marshal(reqBody)

		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/medical_history/1", bytes.NewBuffer(requestBodyBytes)) // Correct path and ID. Updated.
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)},
		}

		handler.UpdateMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to update medical history entry", errResp.Error) // Correct error message. Updated.
	})

	// ... other test cases for UpdateMedicalHistoryEntry (invalid input, authorization errors, etc.)
}

func TestDeleteMedicalHistoryEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockMedicalHistoryService)
	handler := NewMedicalHistoryHandler(mockSvc, log)

	t.Run("valid_id", func(t *testing.T) {
		patientID := 1
		entryID := 1
		mockSvc.On("DeleteMedicalHistoryEntry", mock.Anything, entryID).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/medical_history/1", nil) // Updated path
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)},
		}

		handler.DeleteMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusNoContent, w.Code) // Correct status code

	})

	t.Run("invalid_medical_history_id", func(t *testing.T) { // Added test case for invalid entry id.
		patientID := 1
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/medical_history/invalid", nil) // Updated path
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: "invalid"}, // Use "invalid" instead of -1
		}

		handler.DeleteMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)    // Unmarshal error response
		assert.NoError(t, err)                             // Ensure no error during unmarshal
		assert.Equal(t, "Invalid entry ID", errResp.Error) // Check error message

	})

	t.Run("medical_history_not_found", func(t *testing.T) { // Add test for not found
		patientID := 1
		entryID := 999

		mockSvc.On("DeleteMedicalHistoryEntry", mock.Anything, entryID).Return(domain.ErrMedicalHistoryEntryNotFound) // Updated

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/medical_history/999", nil) // Updated path
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)}, // Updated parameter
		}

		handler.DeleteMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusNotFound, w.Code) // Return 404. Updated.

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, domain.ErrorResponse{Error: "medical history entry not found"}, errResp) // Updated

	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1
		entryID := 1
		mockSvc.On("DeleteMedicalHistoryEntry", mock.Anything, entryID).Return(errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/medical_history/1", nil) // Updated path. Updated.
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "medical_history_id", Value: strconv.Itoa(entryID)}, // Use strconv.Itoa to convert entryID
		}
		handler.DeleteMedicalHistoryEntry(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Correct status code. Updated

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp) // Unmarshal response body. Updated.
		assert.NoError(t, err)                          // Ensure no unmarshaling error. Updated.

		assert.Equal(t, "Failed to delete medical history entry", errResp.Error) // Correct error message. Updated.
	})

	// ... other test cases for DeleteMedicalHistoryEntry (authorization errors, etc.)
}
