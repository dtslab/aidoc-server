package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"go.uber.org/zap"
)

type MedicalHistoryRepositoryImpl struct {
	q   *db.Queries
	log *zap.Logger // Add logger field
}

func NewMedicalHistoryRepository(q *db.Queries, log *zap.Logger) *MedicalHistoryRepositoryImpl { // Inject logger
	return &MedicalHistoryRepositoryImpl{q: q, log: log}
}

func (r *MedicalHistoryRepositoryImpl) CreateMedicalHistoryEntry(ctx context.Context, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error) {
	r.log.Info("CreateMedicalHistoryEntry repository started") // Log the start of the repository function. Updated.

	arg := db.CreateMedicalHistoryEntryParams{
		PatientID:     sql.NullInt32{Int32: int32(entry.PatientID), Valid: true},
		Condition:     entry.Condition,
		DiagnosisDate: sql.NullTime{Time: entry.DiagnosisDate, Valid: !entry.DiagnosisDate.IsZero()},
		Status:        sql.NullString{String: entry.Status, Valid: entry.Status != ""},
		Details:       sql.NullString{String: entry.Details, Valid: entry.Details != ""},
	}

	createdEntry, err := r.q.CreateMedicalHistoryEntry(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" { // Example foreign key violation
			return nil, domain.ErrPatientNotFound // Or a more specific FK error type
		}
		r.log.Error("Failed to create medical history entry", zap.Error(err)) // Log the error. Updated.
		return nil, fmt.Errorf("failed to create medical history entry: %w", err)
	}

	r.log.Info("CreateMedicalHistoryEntry repository completed successfully") // Log successful completion. Updated.
	return convertDbMedicalHistoryEntryToDomain(createdEntry), nil
}

func (r *MedicalHistoryRepositoryImpl) GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error) {
	r.log.Info("GetMedicalHistoryEntries repository started", zap.Int("patientID", patientID)) // Log start, include patientID

	dbEntries, err := r.q.GetMedicalHistoryEntries(ctx, sql.NullInt32{Int32: int32(patientID), Valid: true})
	if err != nil {
		r.log.Error("failed to get medical history entries", zap.Error(err), zap.Int("patient_id", patientID)) // Log the error and patientID. Updated
		return nil, fmt.Errorf("failed to get medical history entries: %w", err)                               // Wrap and return the error for better context. Updated
	}

	if len(dbEntries) == 0 {
		return nil, domain.ErrMedicalHistoryEntryNotFound // Return not found error if no entries. Updated
	}

	domainEntries := make([]*domain.MedicalHistoryEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		domainEntries[i] = convertDbMedicalHistoryEntryToDomain(dbEntry)
	}
	r.log.Info("GetMedicalHistoryEntries repository completed successfully", zap.Int("patient_id", patientID)) // Logging with patient info
	return domainEntries, nil
}

func (r *MedicalHistoryRepositoryImpl) GetMedicalHistoryEntry(ctx context.Context, entryID int) (*domain.MedicalHistoryEntry, error) {
	r.log.Info("GetMedicalHistoryEntry repository started", zap.Int("entryID", entryID)) // Logging with entryID

	dbEntry, err := r.q.GetMedicalHistoryEntry(ctx, int32(entryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMedicalHistoryEntryNotFound
		}

		r.log.Error("Failed to get medical history entry", zap.Error(err), zap.Int("entry_id", entryID))
		return nil, fmt.Errorf("failed to get medical history entry: %w", err)
	}

	r.log.Info("GetMedicalHistoryEntry repository completed successfully", zap.Int("entryID", entryID)) // Logging with entryID
	return convertDbMedicalHistoryEntryToDomain(dbEntry), nil

}

func (r *MedicalHistoryRepositoryImpl) UpdateMedicalHistoryEntry(ctx context.Context, entryID int, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error) {
	r.log.Info("UpdateMedicalHistoryEntry repository started", zap.Int("entryID", entryID))

	arg := db.UpdateMedicalHistoryEntryParams{
		PatientMedicalHistoryID: int32(entryID),
		Condition:               entry.Condition,
		DiagnosisDate:           sql.NullTime{Time: entry.DiagnosisDate, Valid: !entry.DiagnosisDate.IsZero()},
		Status:                  sql.NullString{String: entry.Status, Valid: entry.Status != ""},
		Details:                 sql.NullString{String: entry.Details, Valid: entry.Details != ""},
	}

	updatedEntry, err := r.q.UpdateMedicalHistoryEntry(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Check for not found error before generic database error
			return nil, domain.ErrMedicalHistoryEntryNotFound // Return not found error if entry does not exist. Updated.
		}

		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" { // Example foreign key violation
			return nil, domain.ErrPatientNotFound // Or a more specific FK error type.
		}
		r.log.Error("Failed to update medical history entry", zap.Error(err), zap.Int("entryID", entryID)) // Log the error and entryID. Updated
		return nil, fmt.Errorf("failed to update medical history entry: %w", err)                          // Wrap and return the error for context. Updated.

	}

	r.log.Info("UpdateMedicalHistoryEntry repository completed successfully", zap.Int("entryID", entryID)) // Log success with entryID. Updated
	return convertDbMedicalHistoryEntryToDomain(updatedEntry), nil
}

func (r *MedicalHistoryRepositoryImpl) DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error {
	r.log.Info("DeleteMedicalHistoryEntry repository started", zap.Int("entryID", entryID)) // Logging with entryID

	err := r.q.DeleteMedicalHistoryEntry(ctx, int32(entryID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Handle not found error during delete
			return domain.ErrMedicalHistoryEntryNotFound // Return NotFound error
		}
		r.log.Error("Failed to delete medical history entry", zap.Error(err), zap.Int("entryID", entryID)) // Logging with entryID
		return fmt.Errorf("failed to delete medical history entry: %w", err)                               // Return wrapped error for better context

	}
	r.log.Info("DeleteMedicalHistoryEntry repository completed successfully", zap.Int("entryID", entryID)) // Logging with entryID
	return nil
}

func convertDbMedicalHistoryEntryToDomain(dbEntry db.PatientMedicalHistory) *domain.MedicalHistoryEntry {
	return &domain.MedicalHistoryEntry{
		PatientMedicalHistoryID: int(dbEntry.PatientMedicalHistoryID),
		PatientID:               int(dbEntry.PatientID.Int32),
		Condition:               dbEntry.Condition,
		DiagnosisDate:           dbEntry.DiagnosisDate.Time,
		Status:                  dbEntry.Status.String,
		Details:                 dbEntry.Details.String,
		CreatedAt:               dbEntry.CreatedAt.Time,
		UpdatedAt:               dbEntry.UpdatedAt.Time,
	}
}
