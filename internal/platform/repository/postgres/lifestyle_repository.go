package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stackvity/aidoc-server/internal/core/domain"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"go.uber.org/zap"
)

type LifestyleRepositoryImpl struct {
	q   *db.Queries
	log *zap.Logger
}

// NewLifestyleRepository creates a new LifestyleRepositoryImpl
func NewLifestyleRepository(q *db.Queries, log *zap.Logger) *LifestyleRepositoryImpl {
	return &LifestyleRepositoryImpl{q: q, log: log}
}

// CreateLifestyleEntry implements ports.LifestyleRepository
func (r *LifestyleRepositoryImpl) CreateLifestyleEntry(ctx context.Context, entry *domain.LifestyleEntry) (*domain.LifestyleEntry, error) {
	r.log.Info("CreateLifestyleEntry repository started")

	arg := db.CreateLifestyleEntryParams{
		PatientID:       int32(entry.PatientID),
		LifestyleFactor: entry.LifestyleFactor,
		Value:           sql.NullString{String: entry.Value, Valid: entry.Value != ""},
		StartDate:       sql.NullTime{Time: entry.StartDate, Valid: !entry.StartDate.IsZero()},
		EndDate:         sql.NullTime{Time: entry.EndDate, Valid: !entry.EndDate.IsZero()},
	}

	newEntry, err := r.q.CreateLifestyleEntry(ctx, arg)
	if err != nil {
		r.log.Error("failed create lifestyle entry", zap.Error(err))
		return nil, fmt.Errorf("create lifestyle entry error: %w", err)
	}

	r.log.Info("CreateLifestyleEntry repository completed successfully")
	return convertDbLifestyleEntryToDomain(newEntry), nil
}

// GetLifestyleEntries implements ports.LifestyleRepository
func (r *LifestyleRepositoryImpl) GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error) {
	r.log.Info("GetLifestyleEntries repository started", zap.Int("patient_id", patientID))

	entries, err := r.q.GetLifestyleEntries(ctx, int32(patientID))
	if err != nil {
		r.log.Error("failed get lifestyle entries", zap.Error(err), zap.Int("patient_id", patientID))
		return nil, fmt.Errorf("get lifestyle entries error: %w", err) // More descriptive error
	}

	if len(entries) == 0 {
		return nil, domain.ErrLifestyleEntryNotFound // Return not found error
	}

	domainEntries := make([]*domain.LifestyleEntry, len(entries))
	for i, entry := range entries {
		domainEntries[i] = convertDbLifestyleEntryToDomain(entry)
	}

	r.log.Info("GetLifestyleEntries repository completed successfully")
	return domainEntries, nil
}

// GetLifestyleEntry implements ports.LifestyleRepository. Retrieves a single lifestyle entry.
func (r *LifestyleRepositoryImpl) GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error) {
	r.log.Info("GetLifestyleEntry repository started", zap.Int("entry_id", entryID))

	dbEntry, err := r.q.GetLifestyleEntry(ctx, int32(entryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLifestyleEntryNotFound
		}
		r.log.Error("failed get lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("get lifestyle entry error: %w", err)
	}

	r.log.Info("GetLifestyleEntry repository completed successfully")
	return convertDbLifestyleEntryToDomain(dbEntry), nil
}

// UpdateLifestyleEntry implements ports.LifestyleRepository
func (r *LifestyleRepositoryImpl) UpdateLifestyleEntry(ctx context.Context, entryID int, updatedEntry *domain.LifestyleEntry) (*domain.LifestyleEntry, error) {
	r.log.Info("UpdateLifestyleEntry repository started")

	arg := db.UpdateLifestyleEntryParams{
		PatientLifestyleID: int32(entryID),
		LifestyleFactor:    updatedEntry.LifestyleFactor,
		Value:              sql.NullString{String: updatedEntry.Value, Valid: updatedEntry.Value != ""},
		StartDate:          sql.NullTime{Time: updatedEntry.StartDate, Valid: !updatedEntry.StartDate.IsZero()},
		EndDate:            sql.NullTime{Time: updatedEntry.EndDate, Valid: !updatedEntry.EndDate.IsZero()},
	}

	entry, err := r.q.UpdateLifestyleEntry(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Handle not found error during update
			return nil, domain.ErrLifestyleEntryNotFound
		}

		r.log.Error("failed update lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("update lifestyle entry error: %w", err)
	}

	r.log.Info("UpdateLifestyleEntry repository completed successfully")
	return convertDbLifestyleEntryToDomain(entry), nil

}

// DeleteLifestyleEntry implements ports.LifestyleRepository
func (r *LifestyleRepositoryImpl) DeleteLifestyleEntry(ctx context.Context, entryID int) error {
	r.log.Info("DeleteLifestyleEntry repository started")

	err := r.q.DeleteLifestyleEntry(ctx, int32(entryID))
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) { // Handle not found during delete
			return domain.ErrLifestyleEntryNotFound
		}
		r.log.Error("failed delete lifestyle entry", zap.Error(err), zap.Int("entry_id", entryID))
		return fmt.Errorf("delete lifestyle entry error: %w", err)
	}

	r.log.Info("DeleteLifestyleEntry repository completed successfully")
	return nil
}

func convertDbLifestyleEntryToDomain(dbEntry db.PatientLifestyle) *domain.LifestyleEntry {
	return &domain.LifestyleEntry{
		PatientLifestyleID: int(dbEntry.PatientLifestyleID),
		PatientID:          int(dbEntry.PatientID),
		LifestyleFactor:    dbEntry.LifestyleFactor,
		Value:              dbEntry.Value.String,
		StartDate:          dbEntry.StartDate.Time,
		EndDate:            dbEntry.EndDate.Time,
		CreatedAt:          dbEntry.CreatedAt.Time,
		UpdatedAt:          dbEntry.UpdatedAt.Time,
	}
}
