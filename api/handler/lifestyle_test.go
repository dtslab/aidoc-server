package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain" // Import ports
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLifestyleService mocks the LifestyleService
type MockLifestyleService struct {
	mock.Mock
}

// CreateLifestyleEntry mocks CreateLifestyleEntry
func (m *MockLifestyleService) CreateLifestyleEntry(ctx context.Context, patientID int, req domain.CreateLifestyleRequest) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, patientID, req)
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)
}

// GetLifestyleEntries mocks GetLifestyleEntries
func (m *MockLifestyleService) GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error) {
	args := m.Called(ctx, patientID)
	return args.Get(0).([]*domain.LifestyleEntry), args.Error(1)

}

// GetLifestyleEntry mocks GetLifestyleEntry
func (m *MockLifestyleService) GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)
}

// UpdateLifestyleEntry mocks UpdateLifestyleEntry
func (m *MockLifestyleService) UpdateLifestyleEntry(ctx context.Context, entryID int, req domain.UpdateLifestyleRequest) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, entryID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)
}

// DeleteLifestyleEntry mocks DeleteLifestyleEntry
func (m *MockLifestyleService) DeleteLifestyleEntry(ctx context.Context, entryID int) error {
	args := m.Called(ctx, entryID)
	return args.Error(0)
}

func TestCreateLifestyleEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	mockSvc := new(MockLifestyleService)
	handler := NewLifestyleHandler(mockSvc, log)

	t.Run("valid_input", func(t *testing.T) {
		patientID := 1
		reqBody := domain.CreateLifestyleRequest{
			LifestyleFactor: "Test Factor",
			Value:           "Test Value",
		}
		expectedEntry := &domain.LifestyleEntry{
			PatientLifestyleID: 1,
			PatientID:          patientID,
			LifestyleFactor:    reqBody.LifestyleFactor,
			Value:              reqBody.Value,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		mockSvc.On("CreateLifestyleEntry", mock.Anything, patientID, reqBody).Return(expectedEntry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/lifestyle", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateLifestyleEntry(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createdEntry domain.LifestyleEntry
		_ = json.Unmarshal(w.Body.Bytes(), &createdEntry)
		assert.Equal(t, expectedEntry, &createdEntry)

	})
	// ... other test cases (invalid input, patient not found, server error)
	t.Run("invalid_input", func(t *testing.T) {
		patientID := 1
		reqBody := domain.CreateLifestyleRequest{
			LifestyleFactor: "", // Invalid: Missing lifestyle factor
			Value:           "Test Value",
		}

		mockSvc.On("CreateLifestyleEntry", mock.Anything, patientID, reqBody).Return(nil, domain.ErrInvalidInput) // Return invalid input error

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/lifestyle", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateLifestyleEntry(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "invalid input", errResp.Error) // Expected error from handler

	})

	t.Run("patient_not_found", func(t *testing.T) {
		patientID := 999
		reqBody := domain.CreateLifestyleRequest{
			LifestyleFactor: "Test Factor",
			Value:           "Test Value",
		}

		mockSvc.On("CreateLifestyleEntry", mock.Anything, patientID, reqBody).Return(nil, domain.ErrPatientNotFound) // Return not found error

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/999/lifestyle", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateLifestyleEntry(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "patient not found", errResp.Error)

	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1
		reqBody := domain.CreateLifestyleRequest{
			LifestyleFactor: "Test Factor",
			Value:           "Test Value",
		}

		mockSvc.On("CreateLifestyleEntry", mock.Anything, patientID, reqBody).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients/1/lifestyle", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.CreateLifestyleEntry(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Check for correct status code

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to create lifestyle entry", errResp.Error)
	})
}

func TestGetLifestyleEntries(t *testing.T) {
	// ... (Implementation will be very similar to TestCreateLifestyleEntry)
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	mockSvc := new(MockLifestyleService)
	handler := NewLifestyleHandler(mockSvc, log)

	t.Run("valid_patient_id", func(t *testing.T) {
		patientID := 1
		expectedEntries := []*domain.LifestyleEntry{
			{PatientLifestyleID: 1, PatientID: patientID, LifestyleFactor: "Factor 1", Value: "Value 1"},
			{PatientLifestyleID: 2, PatientID: patientID, LifestyleFactor: "Factor 2", Value: "Value 2"},
		}
		mockSvc.On("GetLifestyleEntries", mock.Anything, patientID).Return(expectedEntries, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/lifestyle", nil) // Correct path
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.GetLifestyleEntries(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var entries []*domain.LifestyleEntry
		_ = json.Unmarshal(w.Body.Bytes(), &entries)
		assert.Equal(t, expectedEntries, entries)

	})
	// ... Test cases for invalid patient ID, not found, server error.
	t.Run("invalid_patient_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/invalid/lifestyle", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: "invalid"}}

		handler.GetLifestyleEntries(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Invalid patient ID", errResp.Error) // Correct error message
	})

	t.Run("patient_not_found", func(t *testing.T) {
		patientID := 999 // non-existent patient
		mockSvc.On("GetLifestyleEntries", mock.Anything, patientID).Return(nil, domain.ErrPatientNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/999/lifestyle", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.GetLifestyleEntries(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "patient not found", errResp.Error)

	})

	t.Run("no_lifestyle_entries_found", func(t *testing.T) {
		patientID := 1
		mockSvc.On("GetLifestyleEntries", mock.Anything, patientID).Return(nil, domain.ErrLifestyleEntryNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/lifestyle", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.GetLifestyleEntries(c)

		assert.Equal(t, http.StatusOK, w.Code) // Should return 200 OK for empty result
		assert.Equal(t, "[]", w.Body.String()) // Assert for empty JSON Array

	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1
		mockSvc.On("GetLifestyleEntries", mock.Anything, patientID).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1/lifestyle", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: strconv.Itoa(patientID)}}

		handler.GetLifestyleEntries(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to get lifestyle entries", errResp.Error)
	})
}

// ... TestUpdateLifestyleEntry, TestDeleteLifestyleEntry (Implement these similarly)
func TestUpdateLifestyleEntry(t *testing.T) {
	// ... similar structure to TestCreateLifestyleEntry and TestGetLifestyleEntries.
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	mockSvc := new(MockLifestyleService)
	handler := NewLifestyleHandler(mockSvc, log)

	t.Run("valid_input", func(t *testing.T) {
		patientID := 1
		entryID := 1
		reqBody := domain.UpdateLifestyleRequest{
			LifestyleFactor: "Updated Factor",
			Value:           "Updated Value",
		}
		expectedEntry := &domain.LifestyleEntry{
			PatientLifestyleID: entryID,
			PatientID:          patientID,
			LifestyleFactor:    reqBody.LifestyleFactor,
			Value:              reqBody.Value,
			UpdatedAt:          time.Now(),
		}

		mockSvc.On("UpdateLifestyleEntry", mock.Anything, entryID, reqBody).Return(expectedEntry, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/lifestyle/1", bytes.NewBuffer(requestBodyBytes)) // Correct path
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "lifestyle_id", Value: strconv.Itoa(entryID)}, // Add lifestyle_id param
		}

		handler.UpdateLifestyleEntry(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedEntry domain.LifestyleEntry
		_ = json.Unmarshal(w.Body.Bytes(), &updatedEntry)

		assert.Equal(t, expectedEntry, &updatedEntry)
	})
	t.Run("invalid_lifestyle_id", func(t *testing.T) {
		patientID := 1
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1/lifestyle/invalid", nil)
		c.Params = []gin.Param{
			{Key: "patient_id", Value: strconv.Itoa(patientID)},
			{Key: "lifestyle_id", Value: "invalid"},
		}

		handler.UpdateLifestyleEntry(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Invalid lifestyle ID", errResp.Error)
	})
	// ... other error cases
}

func TestDeleteLifestyleEntry(t *testing.T) {
	// ... similar to other handler test functions.
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	mockSvc := new(MockLifestyleService)
	handler := NewLifestyleHandler(mockSvc, log)

	t.Run("valid_id", func(t *testing.T) {
		entryID := 1
		mockSvc.On("DeleteLifestyleEntry", mock.Anything, entryID).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/lifestyle/1", nil)
		c.Params = []gin.Param{
			{Key: "patient_id", Value: "1"},
			{Key: "lifestyle_id", Value: "1"},
		}
		handler.DeleteLifestyleEntry(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)

	})

	// ... other test cases for DeleteLifestyleEntry:
	t.Run("invalid_lifestyle_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/lifestyle/invalid", nil)
		c.Params = []gin.Param{
			{Key: "patient_id", Value: "1"},
			{Key: "lifestyle_id", Value: "invalid"},
		}

		handler.DeleteLifestyleEntry(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)                                                        // Ensure no error during unmarshal
		assert.Equal(t, domain.ErrorResponse{Error: "Invalid lifestyle ID"}, errResp) // Check for the correct error response

	})

	t.Run("lifestyle_entry_not_found", func(t *testing.T) {
		entryID := 999 // Non-existent entry ID
		mockSvc.On("DeleteLifestyleEntry", mock.Anything, entryID).Return(domain.ErrLifestyleEntryNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/lifestyle/999", nil)
		c.Params = []gin.Param{
			{Key: "patient_id", Value: "1"},
			{Key: "lifestyle_id", Value: "999"}, // Use the non-existent ID
		}

		handler.DeleteLifestyleEntry(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "lifestyle entry not found", errResp.Error) // Correct not found message.

	})

	t.Run("internal_server_error", func(t *testing.T) {
		entryID := 1
		mockSvc.On("DeleteLifestyleEntry", mock.Anything, entryID).Return(errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/v1/patients/1/lifestyle/1", nil)
		c.Params = []gin.Param{
			{Key: "patient_id", Value: "1"},
			{Key: "lifestyle_id", Value: "1"},
		}

		handler.DeleteLifestyleEntry(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to delete lifestyle entry", errResp.Error)

	})
}
