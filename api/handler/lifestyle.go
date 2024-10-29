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

type LifestyleHandler struct {
	lifestyleSvc ports.LifestyleService // Changed to interface
	log          *zap.Logger
}

// NewLifestyleHandler returns a new LifestyleHandler
func NewLifestyleHandler(lifestyleSvc ports.LifestyleService, log *zap.Logger) *LifestyleHandler {
	return &LifestyleHandler{
		lifestyleSvc: lifestyleSvc,
		log:          log,
	}
}

// CreateLifestyleEntry handles the creation of a new lifestyle entry
func (h *LifestyleHandler) CreateLifestyleEntry(c *gin.Context) {
	h.log.Info("CreateLifestyleEntry handler started")

	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil {
		h.log.Error("Invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"})
		return
	}

	var req domain.CreateLifestyleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	entry, err := h.lifestyleSvc.CreateLifestyleEntry(c, patientID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		default: // For all other errors

			h.log.Error("Failed to create lifestyle entry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to create lifestyle entry"})
		}
		return
	}

	h.log.Info("Lifestyle entry created successfully", zap.Int("patient_id", patientID), zap.String("lifestyle_factor", req.LifestyleFactor))
	c.JSON(http.StatusCreated, entry)
}

// GetLifestyleEntries handles retrieving lifestyle entries for a patient
func (h *LifestyleHandler) GetLifestyleEntries(c *gin.Context) {
	h.log.Info("GetLifestyleEntries handler started")

	patientID, err := strconv.Atoi(c.Param("patient_id"))
	if err != nil {
		h.log.Error("Invalid patient ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid patient ID"})
		return
	}

	entries, err := h.lifestyleSvc.GetLifestyleEntries(c, patientID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPatientNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrLifestyleEntryNotFound):
			c.JSON(http.StatusOK, []*domain.LifestyleEntry{}) // Return empty slice with 200 OK
		default:
			h.log.Error("Failed to get lifestyle entries", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get lifestyle entries"})
		}
		return
	}

	h.log.Info("Successfully retrieved lifestyle entries", zap.Int("patient_id", patientID), zap.Int("count", len(entries)))
	c.JSON(http.StatusOK, entries)
}

// UpdateLifestyleEntry handles updating an existing lifestyle entry
func (h *LifestyleHandler) UpdateLifestyleEntry(c *gin.Context) {
	h.log.Info("UpdateLifestyleEntry handler started")

	entryID, err := strconv.Atoi(c.Param("lifestyle_id"))
	if err != nil {
		h.log.Error("Invalid lifestyle ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid lifestyle ID"})
		return
	}

	var req domain.UpdateLifestyleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	entry, err := h.lifestyleSvc.UpdateLifestyleEntry(c, entryID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrLifestyleEntryNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})

		case errors.Is(err, domain.ErrForbidden): // Handle authorization error
			c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to update lifestyle entry", zap.Error(err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to update lifestyle entry"})
		}
		return

	}

	h.log.Info("Successfully updated lifestyle entry", zap.Int("entry_id", entryID))
	c.JSON(http.StatusOK, entry)
}

// DeleteLifestyleEntry handles deleting a lifestyle entry
func (h *LifestyleHandler) DeleteLifestyleEntry(c *gin.Context) {
	h.log.Info("DeleteLifestyleEntry handler started")

	entryID, err := strconv.Atoi(c.Param("lifestyle_id"))
	if err != nil {
		h.log.Error("Invalid lifestyle ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid lifestyle ID"})
		return
	}

	err = h.lifestyleSvc.DeleteLifestyleEntry(c, entryID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLifestyleEntryNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: err.Error()})
		case errors.Is(err, domain.ErrForbidden):
			c.JSON(http.StatusForbidden, domain.ErrorResponse{Error: err.Error()})
		default:
			h.log.Error("Failed to delete lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to delete lifestyle entry"})
		}
		return
	}

	h.log.Info("Lifestyle entry deleted successfully", zap.Int("entry_id", entryID))
	c.Status(http.StatusNoContent)
}
