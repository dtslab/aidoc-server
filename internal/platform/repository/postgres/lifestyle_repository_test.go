package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	db "github.com/stackvity/aidoc-server/internal/platform/repository/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLifestyleRepository_CreateLifestyleEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop() // Use No-op logger in tests
	q := db.New(mockDB)
	repo := NewLifestyleRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entry := &domain.LifestyleEntry{
			PatientID:       1,
			LifestyleFactor: "Test Factor",
			Value:           "Test Value",
			StartDate:       time.Now(),
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patient_lifestyle (patient_id, lifestyle_factor, value, start_date, end_date)`)).
			WithArgs(int32(entry.PatientID), entry.LifestyleFactor, entry.Value, entry.StartDate, entry.EndDate).
			WillReturnResult(sqlmock.NewResult(1, 1))

		createdEntry, err := repo.CreateLifestyleEntry(context.Background(), entry)

		assert.NoError(t, err)
		assert.NotNil(t, createdEntry)

		assert.Equal(t, entry.LifestyleFactor, createdEntry.LifestyleFactor)
		assert.Equal(t, entry.Value, createdEntry.Value)
		assert.Equal(t, entry.StartDate, createdEntry.StartDate)
		assert.Equal(t, entry.EndDate, createdEntry.EndDate)  // Asserting end date, even if it's a zero value/nil
		assert.Greater(t, createdEntry.PatientLifestyleID, 0) // Check auto-generated ID

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		entry := &domain.LifestyleEntry{
			PatientID:       1,
			LifestyleFactor: "Test Factor",
			Value:           "Test Value",
		}

		mock.ExpectExec("INSERT INTO patient_lifestyle").WillReturnError(errors.New("database error"))

		_, err := repo.CreateLifestyleEntry(context.Background(), entry)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})
	// Add tests for other error scenarios (e.g., invalid input, foreign key violations) as needed.
}

func TestGetLifestyleEntries(t *testing.T) {
	// ... (setup - similar to TestCreateLifestyleEntry)
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop() // Use No-op logger in tests
	q := db.New(mockDB)
	repo := NewLifestyleRepository(q, log)

	t.Run("success", func(t *testing.T) {
		patientID := 1

		expectedDBEntries := []db.PatientLifestyle{
			{
				PatientLifestyleID: 1,
				PatientID:          int32(patientID),
				LifestyleFactor:    "Factor 1",
				Value:              sql.NullString{String: "Value 1", Valid: true},
				CreatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
			},
			{
				PatientLifestyleID: 2,
				PatientID:          int32(patientID),
				LifestyleFactor:    "Factor 2",
				Value:              sql.NullString{String: "Value 2", Valid: true},
				CreatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
				UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
			},
		}

		rows := sqlmock.NewRows([]string{"patient_lifestyle_id", "patient_id", "lifestyle_factor", "value", "start_date", "end_date", "created_at", "updated_at"})
		for _, entry := range expectedDBEntries {
			rows.AddRow(entry.PatientLifestyleID, entry.PatientID, entry.LifestyleFactor, entry.Value, entry.StartDate, entry.EndDate, entry.CreatedAt, entry.UpdatedAt)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at FROM patient_lifestyle WHERE patient_id = $1`)).
			WithArgs(int32(patientID)).
			WillReturnRows(rows)

		entries, err := repo.GetLifestyleEntries(context.Background(), patientID)

		assert.NoError(t, err)
		assert.NotNil(t, entries)
		assert.Equal(t, len(expectedDBEntries), len(entries))

		for i := range entries {

			assert.Equal(t, int(expectedDBEntries[i].PatientLifestyleID), entries[i].PatientLifestyleID)
			assert.Equal(t, expectedDBEntries[i].LifestyleFactor, entries[i].LifestyleFactor)

		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("not_found", func(t *testing.T) {
		patientID := 1

		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).WillReturnError(sql.ErrNoRows) // Correct mock setup for no rows

		entries, err := repo.GetLifestyleEntries(context.Background(), patientID)

		assert.Nil(t, entries)                                   // or assert.Empty(t, entries)
		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Correct error type. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) { // Implement the database error case
		patientID := 1

		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).WillReturnError(errors.New("database error"))

		_, err := repo.GetLifestyleEntries(context.Background(), patientID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetLifestyleEntry(t *testing.T) {
	// Initialize mock database and repository
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewLifestyleRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1
		expectedEntry := db.PatientLifestyle{
			PatientLifestyleID: int32(entryID),
			PatientID:          1,
			LifestyleFactor:    "Test Factor",
			Value:              sql.NullString{String: "Test Value", Valid: true},
			CreatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		}

		// Define expected query and result rows
		rows := sqlmock.NewRows([]string{"patient_lifestyle_id", "patient_id", "lifestyle_factor", "value", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(expectedEntry.PatientLifestyleID, expectedEntry.PatientID, expectedEntry.LifestyleFactor, expectedEntry.Value, expectedEntry.StartDate, expectedEntry.EndDate, expectedEntry.CreatedAt, expectedEntry.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at FROM patient_lifestyle WHERE patient_lifestyle_id = $1`)).
			WithArgs(int32(entryID)).
			WillReturnRows(rows)

		entry, err := repo.GetLifestyleEntry(context.Background(), entryID)
		assert.NoError(t, err)
		assert.NotNil(t, entry)

		// Assert the retrieved entry matches the expected entry
		assert.Equal(t, int(expectedEntry.PatientLifestyleID), entry.PatientLifestyleID)
		assert.Equal(t, expectedEntry.LifestyleFactor, entry.LifestyleFactor)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	t.Run("not_found", func(t *testing.T) {
		entryID := 999 // Entry that does not exist

		mock.ExpectQuery("SELECT").WithArgs(int32(entryID)).WillReturnError(sql.ErrNoRows)

		entry, err := repo.GetLifestyleEntry(context.Background(), entryID)

		assert.Nil(t, entry)                                     // Expect nil entry
		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Expect not found error. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		entryID := 1 // Valid entry ID

		mock.ExpectQuery("SELECT").WithArgs(int32(entryID)).WillReturnError(errors.New("database error"))

		_, err := repo.GetLifestyleEntry(context.Background(), entryID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error") // Check for expected error message

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUpdateLifestyleEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop() // No-op logger for testing
	q := db.New(mockDB)
	repo := NewLifestyleRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1
		updatedEntry := &domain.LifestyleEntry{
			LifestyleFactor: "Updated Factor",
			Value:           "Updated Value",
		}

		expectedDBEntry := db.PatientLifestyle{ // Expected entry after the update.
			PatientLifestyleID: int32(entryID),
			PatientID:          1, // Example
			LifestyleFactor:    updatedEntry.LifestyleFactor,
			Value:              sql.NullString{String: updatedEntry.Value, Valid: true}, // Assuming value is provided
			CreatedAt:          sql.NullTime{Time: time.Now(), Valid: true},             // CreatedAt should not change.
			UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"patient_lifestyle_id", "patient_id", "lifestyle_factor", "value", "start_date", "end_date", "created_at", "updated_at"}).
			AddRow(expectedDBEntry.PatientLifestyleID, expectedDBEntry.PatientID, expectedDBEntry.LifestyleFactor, expectedDBEntry.Value, expectedDBEntry.StartDate, expectedDBEntry.EndDate, expectedDBEntry.CreatedAt, expectedDBEntry.UpdatedAt)

		// Expect an update query with specific arguments and return the updated row.
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE patient_lifestyle SET lifestyle_factor = $2, value = $3, start_date = $4, end_date = $5 WHERE patient_lifestyle_id = $1 RETURNING patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at`)).
			WithArgs(int32(entryID), updatedEntry.LifestyleFactor, updatedEntry.Value, updatedEntry.StartDate, updatedEntry.EndDate).
			WillReturnRows(rows)

		entry, err := repo.UpdateLifestyleEntry(context.Background(), entryID, updatedEntry)
		require.NoError(t, err)

		assert.NotNil(t, entry)
		assert.Equal(t, updatedEntry.LifestyleFactor, entry.LifestyleFactor)
		assert.Equal(t, updatedEntry.Value, entry.Value)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		entryID := 999 // Entry does not exist
		updatedEntry := &domain.LifestyleEntry{
			LifestyleFactor: "Test Factor Updated",
			Value:           "Test Value Updated",
		}

		mock.ExpectQuery("UPDATE").WithArgs(int32(entryID), updatedEntry.LifestyleFactor, updatedEntry.Value, updatedEntry.StartDate, updatedEntry.EndDate).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.UpdateLifestyleEntry(context.Background(), entryID, updatedEntry)
		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Correct error check. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	// Add more test for UpdateLifestyleEntry
	t.Run("database_error", func(t *testing.T) {
		entryID := 1
		updatedEntry := &domain.LifestyleEntry{
			LifestyleFactor: "Test Factor Updated",
			Value:           "Test Value Updated",
		}

		mock.ExpectQuery("UPDATE").WithArgs(int32(entryID), updatedEntry.LifestyleFactor, updatedEntry.Value, updatedEntry.StartDate, updatedEntry.EndDate).
			WillReturnError(errors.New("database error"))

		_, err := repo.UpdateLifestyleEntry(context.Background(), entryID, updatedEntry)

		assert.Error(t, err) // Correctly checks for a generic error
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}

func TestLifestyleRepository_DeleteLifestyleEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewLifestyleRepository(q, log)

	t.Run("success", func(t *testing.T) {
		entryID := 1

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM patient_lifestyle WHERE patient_lifestyle_id = $1`)).WithArgs(int32(entryID)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteLifestyleEntry(context.Background(), entryID)

		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("not_found", func(t *testing.T) { // Correct test case name
		entryID := 999 // Non-existent entry

		mock.ExpectExec("DELETE").WithArgs(int32(entryID)).WillReturnError(sql.ErrNoRows) // Expect NoRows error

		err := repo.DeleteLifestyleEntry(context.Background(), entryID)

		assert.ErrorIs(t, err, domain.ErrLifestyleEntryNotFound) // Check for correct error. Updated

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		entryID := 1

		mock.ExpectExec("DELETE").WithArgs(int32(entryID)).WillReturnError(errors.New("database error")) // Mock a database error

		err := repo.DeleteLifestyleEntry(context.Background(), entryID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}
