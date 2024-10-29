package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stackvity/aidoc-server/internal/core/ports"
	"go.uber.org/zap"
)

// LifestyleService struct
type LifestyleService struct {
	lifestyleRepo ports.LifestyleRepository
	patientRepo   ports.PatientRepository
	log           *zap.Logger
	validator     *validator.Validate
	authorize     func(context.Context, int) bool
}

// NewLifestyleService creates a new LifestyleService. Inject repositories, logger, validator, and authorize function.
func NewLifestyleService(lifestyleRepo ports.LifestyleRepository, patientRepo ports.PatientRepository, log *zap.Logger, validator *validator.Validate, authorize func(context.Context, int) bool) *LifestyleService {
	return &LifestyleService{
		lifestyleRepo: lifestyleRepo,
		patientRepo:   patientRepo,
		log:           log,
		validator:     validator,
		authorize:     authorize,
	}
}

func (s *LifestyleService) CreateLifestyleEntry(ctx context.Context, patientID int, req domain.CreateLifestyleRequest) (*domain.LifestyleEntry, error) {
	s.log.Info("CreateLifestyleEntry service started", zap.Int("patient_id", patientID))

	if err := s.validator.Struct(req); err != nil {
		s.log.Error("Validation error", zap.Error(err))
		return nil, domain.ErrInvalidInput
	}

	_, err := s.patientRepo.GetPatient(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrPatientNotFound) {
			return nil, domain.ErrPatientNotFound
		}
		return nil, fmt.Errorf("failed to check patient existence: %w", err)
	}

	entry := &domain.LifestyleEntry{
		PatientID:       patientID,
		LifestyleFactor: req.LifestyleFactor,
		Value:           req.Value,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
	}

	newEntry, err := s.lifestyleRepo.CreateLifestyleEntry(ctx, entry)
	if err != nil {
		s.log.Error("failed to create lifestyle entry", zap.Error(err), zap.Int("patient_id", patientID), zap.String("lifestyle_factor", req.LifestyleFactor))
		return nil, fmt.Errorf("create lifestyle entry error: %w", err)
	}
	s.log.Info("Lifestyle entry created successfully", zap.Int("patient_lifestyle_id", newEntry.PatientLifestyleID))

	return newEntry, nil
}

func (s *LifestyleService) GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error) {
	s.log.Info("GetLifestyleEntries service started", zap.Int("patient_id", patientID))

	_, err := s.patientRepo.GetPatient(ctx, patientID)
	if err != nil {
		if errors.Is(err, domain.ErrPatientNotFound) {
			return nil, domain.ErrPatientNotFound
		}
		return nil, fmt.Errorf("failed to check patient existence: %w", err)
	}

	entries, err := s.lifestyleRepo.GetLifestyleEntries(ctx, patientID)
	if err != nil {
		s.log.Error("failed to get lifestyle entries", zap.Error(err), zap.Int("patient_id", patientID))
		return nil, fmt.Errorf("get lifestyle entries error: %w", err)
	}

	if len(entries) == 0 {
		return nil, domain.ErrLifestyleEntryNotFound // Return not found error if no entries
	}

	s.log.Info("GetLifestyleEntries service completed successfully", zap.Int("patient_id", patientID), zap.Int("count", len(entries)))
	return entries, nil
}

func (s *LifestyleService) GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error) {
	s.log.Info("GetLifestyleEntry service started", zap.Int("entry_id", entryID))

	entry, err := s.lifestyleRepo.GetLifestyleEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrLifestyleEntryNotFound) {
			return nil, domain.ErrLifestyleEntryNotFound // Return not found error if entry does not exist
		}
		s.log.Error("failed to get lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("get lifestyle entry error: %w", err)
	}

	s.log.Info("GetLifestyleEntry service completed successfully", zap.Int("entry_id", entryID))
	return entry, nil
}

func (s *LifestyleService) UpdateLifestyleEntry(ctx context.Context, entryID int, req domain.UpdateLifestyleRequest) (*domain.LifestyleEntry, error) {
	s.log.Info("UpdateLifestyleEntry service started", zap.Int("entry_id", entryID))

	if err := s.validator.Struct(req); err != nil {
		s.log.Error("Validation error", zap.Error(err))
		return nil, domain.ErrInvalidInput
	}

	existingEntry, err := s.lifestyleRepo.GetLifestyleEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrLifestyleEntryNotFound) { // Use errors.Is for more robust error checking
			return nil, domain.ErrLifestyleEntryNotFound
		}
		// Log the detailed error and wrap it for context
		s.log.Error("Failed to retrieve existing entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("failed to retrieve existing entry: %w", err)
	}

	// Authorization Check
	if !s.authorize(ctx, existingEntry.PatientID) {
		return nil, domain.ErrForbidden // Return appropriate error for unauthorized access
	}

	// Update only provided fields
	if req.LifestyleFactor != "" {
		existingEntry.LifestyleFactor = req.LifestyleFactor
	}
	if req.Value != "" {
		existingEntry.Value = req.Value
	}
	if !req.StartDate.IsZero() {
		existingEntry.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		existingEntry.EndDate = req.EndDate
	}

	entry, err := s.lifestyleRepo.UpdateLifestyleEntry(ctx, entryID, existingEntry)
	if err != nil {
		// Enhanced error handling
		if errors.Is(err, domain.ErrLifestyleEntryNotFound) { // Check for not found error from repository
			return nil, domain.ErrLifestyleEntryNotFound // Return not found error. Updated
		}

		s.log.Error("failed update lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID), zap.String("lifestyle_factor", existingEntry.LifestyleFactor)) // Include existingEntry fields in log
		return nil, fmt.Errorf("update lifestyle entry error: %w", err)                                                                                           // Wrap and return error
	}

	s.log.Info("Lifestyle entry updated successfully", zap.Int("entry_id", entryID))
	return entry, nil // Return updated entry
}

func (s *LifestyleService) DeleteLifestyleEntry(ctx context.Context, entryID int) error {
	s.log.Info("DeleteLifestyleEntry service started", zap.Int("entry_id", entryID))

	existingEntry, err := s.lifestyleRepo.GetLifestyleEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrLifestyleEntryNotFound) {
			return domain.ErrLifestyleEntryNotFound // Return not found error if the entry doesn't exist
		}
		s.log.Error("Failed to retrieve entry before deletion", zap.Error(err), zap.Int("entry_id", entryID))
		return fmt.Errorf("failed to retrieve entry before deleting: %w", err) // Wrap error for context
	}

	if !s.authorize(ctx, existingEntry.PatientID) { // Authorization check
		return domain.ErrForbidden // Return forbidden error if unauthorized
	}

	err = s.lifestyleRepo.DeleteLifestyleEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, domain.ErrLifestyleEntryNotFound) { // Check for not found error from the repository
			return domain.ErrLifestyleEntryNotFound // Return not found error
		}
		s.log.Error("Failed to delete lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
		return fmt.Errorf("delete lifestyle entry error: %w", err) // Wrap error for context
	}

	s.log.Info("Lifestyle entry deleted successfully", zap.Int("entry_id", entryID))
	return nil
}
