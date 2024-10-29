package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/core/ports"
	"go.uber.org/zap"
)

// PatientHandler struct
type PatientHandler struct {
	patientSvc ports.PatientService
	log        *zap.Logger
}

// NewPatientHandler returns a new PatientHandler
func NewPatientHandler(patientSvc ports.PatientService, log *zap.Logger) *PatientHandler {
	return &PatientHandler{
		patientSvc: patientSvc,
		log:        log,
	}
}

// CreatePatient handles creating a new patient
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	h.log.Info("CreatePatient handler started")

	var req domain.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("CreatePatient: Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	patient, err := h.patientSvc.CreatePatient(c, req)
	if err != nil {
		var validationErr *domain.ValidationError // Declare a variable of the concrete error type. Updated
		switch {
		case errors.As(err, &validationErr): // Check if err wraps a ValidationError. Updated
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to create patient", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to create patient"})
		}
		return
	}

	h.log.Info("CreatePatient handler completed successfully")
	c.JSON(http.StatusCreated, patient)
}

// GetPatient handles retrieving a patient by ID
func (h *PatientHandler) GetPatient(c *gin.Context) {
	h.log.Info("GetPatient handler started")

	patientIDStr := c.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		h.log.Error("invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"})
		return
	}

	patient, err := h.patientSvc.GetPatient(c, patientID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "Patient not found"})
		default:
			h.log.Error("Failed to get patient", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get patient"})
		}
		return
	}

	h.log.Info("GetPatient handler completed successfully")
	c.JSON(http.StatusOK, patient)
}

// UpdatePatient handles updating a patient
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	h.log.Info("UpdatePatient handler started")

	patientIDStr := c.Param("patient_id")
	patientID, err := strconv.Atoi(patientIDStr)
	if err != nil {
		h.log.Error("invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"})
		return
	}

	var req domain.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	patient, err := h.patientSvc.UpdatePatient(c, patientID, req)
	if err != nil {
		var validationErr *domain.ValidationError // Declare variable of concrete type. Updated
		switch {
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "Patient not found"})
		case errors.As(err, &validationErr): // Correctly check for ValidationError. Updated
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to update patient", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to update patient"})
		}
		return
	}

	h.log.Info("UpdatePatient handler completed successfully")
	c.JSON(http.StatusOK, patient)
}
