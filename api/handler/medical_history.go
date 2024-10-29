package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/core/ports" // Import the ports package
	"go.uber.org/zap"
)

// MedicalHistoryHandler struct
type MedicalHistoryHandler struct {
	medicalHistorySvc ports.MedicalHistoryService // Use the interface
	log               *zap.Logger
}

// NewMedicalHistoryHandler returns a new MedicalHistoryHandler
func NewMedicalHistoryHandler(medicalHistorySvc ports.MedicalHistoryService, log *zap.Logger) *MedicalHistoryHandler {
	return &MedicalHistoryHandler{
		medicalHistorySvc: medicalHistorySvc,
		log:               log,
	}
}

func (h *MedicalHistoryHandler) CreateMedicalHistoryEntry(c *gin.Context) {
	h.log.Info("CreateMedicalHistoryEntry handler started")

	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil {
		h.log.Error("Invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"})
		return
	}

	var req domain.CreateMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()}) // Consistent error response
		return
	}

	entry, err := h.medicalHistorySvc.CreateMedicalHistoryEntry(c, patientID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidMedicalHistoryData):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()}) // Use ErrorResponse
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()}) // Use ErrorResponse
		default:
			h.log.Error("Failed to create medical history entry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to create medical history entry"}) // Use ErrorResponse
		}
		return
	}

	h.log.Info("Medical History entry created successfully", zap.Int("patientID", patientID), zap.Int("entryID", entry.PatientMedicalHistoryID))
	c.JSON(http.StatusCreated, entry)
}

func (h *MedicalHistoryHandler) GetMedicalHistoryEntries(c *gin.Context) {
	h.log.Info("GetMedicalHistoryEntries handler started")

	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil {
		h.log.Error("Invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"}) // Consistent error handling
		return
	}

	entries, err := h.medicalHistorySvc.GetMedicalHistoryEntries(c, patientID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMedicalHistoryEntryNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()}) // Return Not Found for no entries, using ErrorResponse
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()}) // Correct status code and error response
		default:
			h.log.Error("Failed to get medical history entries", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get medical history entries"})
		}
		return

	}

	h.log.Info("GetMedicalHistoryEntries handler completed successfully")
	c.JSON(http.StatusOK, entries)
}

func (h *MedicalHistoryHandler) UpdateMedicalHistoryEntry(c *gin.Context) {
	h.log.Info("UpdateMedicalHistoryEntry handler started")

	entryID, err := strconv.Atoi(c.Param("medical_history_id"))
	if err != nil {
		h.log.Error("Invalid medical history ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid medical history ID"}) // Consistent error handling
		return
	}

	var req domain.UpdateMedicalHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()}) // Correct error response
		return
	}

	entry, err := h.medicalHistorySvc.UpdateMedicalHistoryEntry(c, entryID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMedicalHistoryEntryNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidMedicalHistoryData):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})

		case errors.Is(err, domain.ErrForbidden):
			c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to update medical history entry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to update medical history entry"})
		}
		return
	}

	h.log.Info("UpdateMedicalHistoryEntry handler completed successfully")
	c.JSON(http.StatusOK, entry)
}

func (h *MedicalHistoryHandler) DeleteMedicalHistoryEntry(c *gin.Context) {

	h.log.Info("DeleteMedicalHistoryEntry handler started")

	entryID, err := strconv.Atoi(c.Param("medical_history_id"))
	if err != nil {
		h.log.Error("Invalid entry ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid entry ID"}) // Consistent error response
		return
	}

	err = h.medicalHistorySvc.DeleteMedicalHistoryEntry(c, entryID)
	if err != nil {
		switch {

		case errors.Is(err, domain.ErrMedicalHistoryEntryNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrForbidden):
			c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to delete medical history entry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to delete medical history entry"}) // Correct error response
		}
		return

	}

	h.log.Info("DeleteMedicalHistoryEntry handler completed successfully")
	c.Status(http.StatusNoContent) // Correct status code
}
