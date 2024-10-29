package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockPatientService is a mock implementation of the PatientService interface
type MockPatientService struct {
	mock.Mock
}

// CreatePatient mocks the CreatePatient method
func (m *MockPatientService) CreatePatient(ctx context.Context, req domain.CreatePatientRequest) (*domain.Patient, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.Patient), args.Error(1)
}

// GetPatient mocks the GetPatient method
func (m *MockPatientService) GetPatient(ctx context.Context, patientID int) (*domain.Patient, error) {
	args := m.Called(ctx, patientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Patient), args.Error(1)
}

// UpdatePatient mocks the UpdatePatient method
func (m *MockPatientService) UpdatePatient(ctx context.Context, patientID int, req domain.UpdatePatientRequest) (*domain.Patient, error) {
	args := m.Called(ctx, patientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Patient), args.Error(1)
}

func TestCreatePatient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop() // Use a no-op logger for testing

	mockSvc := new(MockPatientService)
	handler := NewPatientHandler(mockSvc, log)

	t.Run("valid_input", func(t *testing.T) {
		reqBody := domain.CreatePatientRequest{
			UserID:                 1,
			FullName:               "John Doe",
			Age:                    30,
			DateOfBirth:            time.Date(1994, time.January, 1, 0, 0, 0, 0, time.UTC),
			Sex:                    "Male",
			PhoneNumber:            "123-456-7890",
			EmailAddress:           "john.doe@example.com",
			PreferredCommunication: "Email",
			SocioeconomicStatus:    "Middle",
			GeographicLocation:     "Testville",
		}

		expectedPatient := &domain.Patient{
			PatientID:              1,
			UserID:                 reqBody.UserID,
			FullName:               reqBody.FullName,
			Age:                    reqBody.Age,
			DateOfBirth:            reqBody.DateOfBirth,
			Sex:                    reqBody.Sex,
			PhoneNumber:            reqBody.PhoneNumber,
			EmailAddress:           reqBody.EmailAddress,
			PreferredCommunication: reqBody.PreferredCommunication,
			SocioeconomicStatus:    reqBody.SocioeconomicStatus,
			GeographicLocation:     reqBody.GeographicLocation,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}

		mockSvc.On("CreatePatient", mock.Anything, reqBody).Return(expectedPatient, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreatePatient(c)
		assert.Equal(t, http.StatusCreated, w.Code)

		var createdPatient domain.Patient
		_ = json.Unmarshal(w.Body.Bytes(), &createdPatient)
		assert.Equal(t, expectedPatient.FullName, createdPatient.FullName) // Example assertion
		// Add more assertions for other fields as needed.
	})

	t.Run("invalid_input", func(t *testing.T) {
		reqBody := domain.CreatePatientRequest{
			UserID:      1,
			FullName:    "",         // missing full name
			DateOfBirth: time.Now(), // Invalid date of birth
			Sex:         "Invalid",  // Invalid sex value
		}

		mockSvc.On("CreatePatient", mock.Anything, reqBody).Return(nil, errors.New("validation error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreatePatient(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Contains(t, errResp.Error, "validation error")
	})

	t.Run("internal_server_error", func(t *testing.T) {
		reqBody := domain.CreatePatientRequest{
			UserID:                 1,
			FullName:               "John Doe",
			Age:                    30,
			DateOfBirth:            time.Date(1994, time.January, 1, 0, 0, 0, 0, time.UTC),
			Sex:                    "Male",
			PhoneNumber:            "123-456-7890",
			EmailAddress:           "john.doe@example.com",
			PreferredCommunication: "Email",
			SocioeconomicStatus:    "Middle",
			GeographicLocation:     "Testville",
		}
		mockSvc.On("CreatePatient", mock.Anything, reqBody).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1/patients", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.CreatePatient(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to create patient", errResp.Error) // from handler code
	})
}

func TestGetPatient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockPatientService)
	handler := NewPatientHandler(mockSvc, log)

	t.Run("valid_patient_id", func(t *testing.T) {
		patientID := 1
		expectedPatient := &domain.Patient{PatientID: 1, FullName: "John Doe"}

		mockSvc.On("GetPatient", mock.Anything, patientID).Return(expectedPatient, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1", nil)

		c.Params = []gin.Param{{Key: "patient_id", Value: "1"}}
		handler.GetPatient(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var patient domain.Patient
		_ = json.Unmarshal(w.Body.Bytes(), &patient)
		assert.Equal(t, *expectedPatient, patient)
	})

	t.Run("invalid_patient_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/invalid", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: "invalid"}}

		handler.GetPatient(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Invalid patient ID", errResp.Error)
	})

	t.Run("patient_not_found", func(t *testing.T) {
		patientID := 1
		mockSvc.On("GetPatient", mock.Anything, patientID).Return(nil, domain.ErrPatientNotFound)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: "1"}}

		handler.GetPatient(c)
		assert.Equal(t, http.StatusNotFound, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Patient not found", errResp.Error)
	})

	t.Run("internal_server_error", func(t *testing.T) {
		patientID := 1
		mockSvc.On("GetPatient", mock.Anything, patientID).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/v1/patients/1", nil)
		c.Params = []gin.Param{{Key: "patient_id", Value: "1"}}

		handler.GetPatient(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to get patient", errResp.Error)
	})
}

func TestUpdatePatient(t *testing.T) {
	// ... (Implementation similar to TestGetPatient, but using UpdatePatient service method and PUT request.)
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	mockSvc := new(MockPatientService)
	handler := NewPatientHandler(mockSvc, log)
	t.Run("valid_input", func(t *testing.T) {
		patientID := 1
		reqBody := domain.UpdatePatientRequest{
			FullName: "John Doe Updated",
		}

		expectedPatient := &domain.Patient{
			PatientID: 1,
			FullName:  "John Doe Updated",
			UpdatedAt: time.Now(),
		}
		mockSvc.On("UpdatePatient", mock.Anything, patientID, reqBody).Return(expectedPatient, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBodyBytes, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/v1/patients/1", bytes.NewBuffer(requestBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "patient_id", Value: "1"}}

		handler.UpdatePatient(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedPatient domain.Patient
		_ = json.Unmarshal(w.Body.Bytes(), &updatedPatient)

		assert.Equal(t, expectedPatient.FullName, updatedPatient.FullName)
	})
	// ... other test cases for UpdatePatient (invalid input, not found, server error)
}
