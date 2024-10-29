package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMedicalHistoryRepository_CreateMedicalHistoryEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewMedicalHistoryRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entry := &domain.MedicalHistoryEntry{
			PatientID:     1,
			Condition:     "Hypertension",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Some details about the condition",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patient_medical_history (patient_id, condition, diagnosis_date, status, details)`)).
			WithArgs(sqlmock.AnyArg(), entry.Condition, entry.DiagnosisDate, entry.Status, entry.Details).
			WillReturnResult(sqlmock.NewResult(1, 1))

		createdEntry, err := repo.CreateMedicalHistoryEntry(context.Background(), entry)

		assert.NoError(t, err)
		assert.NotNil(t, createdEntry)
		assert.Equal(t, entry.Condition, createdEntry.Condition)
		assert.Equal(t, entry.Details, createdEntry.Details)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("duplicate_entry", func(t *testing.T) {
		entry := &domain.MedicalHistoryEntry{
			PatientID:     1,
			Condition:     "Hypertension",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Some details",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patient_medical_history (patient_id, condition, diagnosis_date, status, details)`)).
			WithArgs(sqlmock.AnyArg(), entry.Condition, entry.DiagnosisDate, entry.Status, entry.Details).
			WillReturnError(&pgconn.PgError{Code: "23505"}) // Unique violation error code

		_, err := repo.CreateMedicalHistoryEntry(context.Background(), entry)

		assert.Error(t, err) // Expect an error

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) { // Correct test case name. Updated.
		entry := &domain.MedicalHistoryEntry{
			PatientID:     1,
			Condition:     "Hypertension",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Some details about the condition",
		}
		mock.ExpectExec("INSERT INTO patient_medical_history").
			WillReturnError(errors.New("database error"))

		_, err := repo.CreateMedicalHistoryEntry(context.Background(), entry)

		assert.Error(t, err) // Correctly checks for an error. Updated.
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("foreign_key_violation", func(t *testing.T) {
		entry := &domain.MedicalHistoryEntry{
			PatientID:     999, // Non-existent patient ID
			Condition:     "Some condition",
			DiagnosisDate: time.Now(),
			Status:        "Active",
			Details:       "Some details",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patient_medical_history`)).
			WithArgs(sqlmock.AnyArg(), entry.Condition, entry.DiagnosisDate, entry.Status, entry.Details).
			WillReturnError(&pgconn.PgError{Code: "23503"}) // Foreign key violation error code

		_, err := repo.CreateMedicalHistoryEntry(context.Background(), entry)

		assert.ErrorIs(t, err, domain.ErrPatientNotFound) // Use assert.ErrorIs. Updated.

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetMedicalHistoryEntries(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewMedicalHistoryRepository(q, log)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		expectedEntries := []db.PatientMedicalHistory{
			{
				PatientMedicalHistoryID: 1,
				PatientID:               sql.NullInt32{Int32: 1, Valid: true},
				Condition:               "Condition 1",
				DiagnosisDate:           sql.NullTime{Time: time.Now(), Valid: true},
				Status:                  sql.NullString{String: "Active", Valid: true},
				Details:                 sql.NullString{String: "Details 1", Valid: true},
				CreatedAt:               sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt:               sql.NullTime{Time: time.Now(), Valid: true},
			},
			// Add more expected entries if needed
		}

		rows := sqlmock.NewRows([]string{"patient_medical_history_id", "patient_id", "condition", "diagnosis_date", "status", "details", "created_at", "updated_at"})
		for _, entry := range expectedEntries {
			rows.AddRow(entry.PatientMedicalHistoryID, entry.PatientID, entry.Condition, entry.DiagnosisDate, entry.Status, entry.Details, entry.CreatedAt, entry.UpdatedAt)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at FROM patient_medical_history WHERE patient_id = $1`)).
			WithArgs(int32(patientID)).
			WillReturnRows(rows)
		// Call the repository method
		entries, err := repo.GetMedicalHistoryEntries(context.Background(), patientID)

		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Equal(t, len(expectedEntries), len(entries)) // Check if the correct number of entries is returned. Updated

		for i := range entries {

			assert.Equal(t, int(expectedEntries[i].PatientMedicalHistoryID), entries[i].PatientMedicalHistoryID)
			assert.Equal(t, expectedEntries[i].Condition, entries[i].Condition)
			// Add assertions for other fields...
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("not_found", func(t *testing.T) {
		patientID := 1

		// Expect query to return no rows
		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).WillReturnError(sql.ErrNoRows) // Correct mock setup for no rows. Updated

		entries, err := repo.GetMedicalHistoryEntries(context.Background(), patientID)

		assert.Nil(t, entries)                                        // Expect nil entries
		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Expect not found error. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		patientID := 1

		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).WillReturnError(errors.New("database error"))

		_, err := repo.GetMedicalHistoryEntries(context.Background(), patientID)

		assert.Error(t, err)                              // Correct assertion to check for an error. Updated.
		assert.Contains(t, err.Error(), "database error") // Correctly checks for database error. Updated.

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

// ... (Add tests for GetMedicalHistoryEntry, UpdateMedicalHistoryEntry, and DeleteMedicalHistoryEntry)
func TestGetMedicalHistoryEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewMedicalHistoryRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1
		expectedEntry := db.PatientMedicalHistory{
			PatientMedicalHistoryID: 1,
			PatientID:               sql.NullInt32{Int32: 1, Valid: true},
			Condition:               "Condition 1",
			DiagnosisDate:           sql.NullTime{Time: time.Now(), Valid: true},
			Status:                  sql.NullString{String: "Active", Valid: true},
			Details:                 sql.NullString{String: "Details 1", Valid: true},
			CreatedAt:               sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:               sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"patient_medical_history_id", "patient_id", "condition", "diagnosis_date", "status", "details", "created_at", "updated_at"}).
			AddRow(expectedEntry.PatientMedicalHistoryID, expectedEntry.PatientID, expectedEntry.Condition, expectedEntry.DiagnosisDate, expectedEntry.Status, expectedEntry.Details, expectedEntry.CreatedAt, expectedEntry.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at FROM patient_medical_history WHERE patient_medical_history_id = $1`)).
			WithArgs(int32(entryID)).WillReturnRows(rows)

		entry, err := repo.GetMedicalHistoryEntry(context.Background(), entryID) // call repository method
		assert.NoError(t, err)
		assert.NotNil(t, entry)

		assert.Equal(t, int(expectedEntry.PatientMedicalHistoryID), entry.PatientMedicalHistoryID) // assertions
		// ... (assertions for other fields as needed)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("not_found", func(t *testing.T) {
		// Test case when the medical history entry is not found
		entryID := 999 // Non-existent entry ID
		mock.ExpectQuery("SELECT").WithArgs(int32(entryID)).WillReturnError(sql.ErrNoRows)

		entry, err := repo.GetMedicalHistoryEntry(context.Background(), entryID) // Updated repository method call

		assert.Nil(t, entry)                                     // Assert that the entry is nil
		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Assert correct error. Updated. Corrected error type.
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	t.Run("database_error", func(t *testing.T) { // Implement database error case
		entryID := 1
		mock.ExpectQuery("SELECT").WithArgs(int32(entryID)).WillReturnError(errors.New("database error"))

		_, err := repo.GetMedicalHistoryEntry(context.Background(), entryID)

		assert.Error(t, err)                                  // Correct assertion for an error. Updated
		assert.Contains(t, err.Error(), "database error")      // Correctly checks database error. Updated.
		assert.NotErrorIs(t, err, sql.ErrNoRows)                // Ensure error is not NoRows. Updated
		assert.NotErrorIs(t, err, domain.ErrPatientNotFound) // Ensure error is not PatientNotFound. Updated.
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}

func TestUpdateMedicalHistoryEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewMedicalHistoryRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1
		updatedEntry := &domain.MedicalHistoryEntry{
			Condition:     "Updated Condition",
			DiagnosisDate: time.Now(),
			Status:        "Resolved",
			Details:       "Updated details",
		}
		expectedUpdatedEntry := db.PatientMedicalHistory{
			PatientMedicalHistoryID: int32(entryID),
			PatientID:               sql.NullInt32{Int32: 1, Valid: true},
			Condition:               updatedEntry.Condition,
			DiagnosisDate:           sql.NullTime{Time: updatedEntry.DiagnosisDate, Valid: true},
			Status:                  sql.NullString{String: updatedEntry.Status, Valid: true}, // Updated
			Details:                 sql.NullString{String: updatedEntry.Details, Valid: true}, // Updated
			CreatedAt:               sql.NullTime{Time: time.Now(), Valid: true},             // Should not change
			UpdatedAt:               sql.NullTime{Time: time.Now(), Valid: true},             // Updated to now
		}

		rows := sqlmock.NewRows([]string{"patient_medical_history_id", "patient_id", "condition", "diagnosis_date", "status", "details", "created_at", "updated_at"}).
			AddRow(expectedUpdatedEntry.PatientMedicalHistoryID, expectedUpdatedEntry.PatientID, expectedUpdatedEntry.Condition, expectedUpdatedEntry.DiagnosisDate, expectedUpdatedEntry.Status, expectedUpdatedEntry.Details, expectedUpdatedEntry.CreatedAt, expectedUpdatedEntry.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE patient_medical_history SET condition = $2, diagnosis_date = $3, status = $4, details = $5, updated_at = NOW() WHERE patient_medical_history_id = $1 RETURNING patient_medical_history_id, patient_id, condition, diagnosis_date, status, details, created_at, updated_at`)).
			WithArgs(int32(entryID), updatedEntry.Condition, updatedEntry.DiagnosisDate, updatedEntry.Status, updatedEntry.Details).
			WillReturnRows(rows)

		entry, err := repo.UpdateMedicalHistoryEntry(context.Background(), entryID, updatedEntry)
		require.NoError(t, err)
		assert.NotNil(t, entry)

		assert.Equal(t, updatedEntry.Condition, entry.Condition)
		// Assert other fields as needed

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})
	t.Run("not_found", func(t *testing.T) {
		entryID := 999
		updatedEntry := &domain.MedicalHistoryEntry{
			Condition: "Some New Condition",
		}
		mock.ExpectQuery("UPDATE").WithArgs(int32(entryID), updatedEntry.Condition, updatedEntry.DiagnosisDate, updatedEntry.Status, updatedEntry.Details).WillReturnError(sql.ErrNoRows)
		_, err := repo.UpdateMedicalHistoryEntry(context.Background(), entryID, updatedEntry)

		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		entryID := 1
		updatedEntry := &domain.MedicalHistoryEntry{
			Condition: "Some New Condition",
		}

		mock.ExpectQuery("UPDATE").WithArgs(int32(entryID), updatedEntry.Condition, updatedEntry.DiagnosisDate, updatedEntry.Status, updatedEntry.Details).WillReturnError(errors.New("database error"))

		_, err := repo.UpdateMedicalHistoryEntry(context.Background(), entryID, updatedEntry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})
	// Test cases for invalid input, database errors, etc.

}

func TestDeleteMedicalHistoryEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewMedicalHistoryRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1

		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM patient_medical_history WHERE patient_medical_history_id = $1")).
			WithArgs(int32(entryID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteMedicalHistoryEntry(context.Background(), entryID)
		require.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		entryID := 999 // Non-existent ID

		// Expect an error when deleting a non-existent entry.
		mock.ExpectExec("DELETE").WithArgs(int32(entryID)).WillReturnError(sql.ErrNoRows)

		err := repo.DeleteMedicalHistoryEntry(context.Background(), entryID)

		assert.ErrorIs(t, err, domain.ErrMedicalHistoryEntryNotFound) // Check for specific not found error. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	t.Run("database_error", func(t *testing.T) { // Implement database error case
		entryID := 1

		mock.ExpectExec("DELETE").WithArgs(int32(entryID)).WillReturnError(errors.New("database error"))

		err := repo.DeleteMedicalHistoryEntry(context.Background(), entryID)

		assert.Error(t, err) // Check for an error. Updated
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}