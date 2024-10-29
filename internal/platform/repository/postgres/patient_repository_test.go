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

func TestPatientRepository_CreatePatient(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewPatientRepository(q, log)

	t.Run("success", func(t *testing.T) {
		patient := &domain.Patient{
			UserID:                 1,
			FullName:               "John Doe",
			Age:                    30,
			DateOfBirth:            time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC),
			Sex:                    "Male",
			PhoneNumber:            "123-456-7890",
			EmailAddress:           "john.doe@example.com",
			PreferredCommunication: "Email",
			SocioeconomicStatus:    "Middle",
			GeographicLocation:     "Testville",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patients (user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location)`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		createdPatient, err := repo.CreatePatient(context.Background(), patient)
		require.NoError(t, err)
		assert.NotNil(t, createdPatient)

		assert.Equal(t, patient.FullName, createdPatient.FullName)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("duplicate_email", func(t *testing.T) {
		// ... (same as before)
		patient := &domain.Patient{
			EmailAddress: "john.doe@example.com", // Example duplicate email
		}
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO patients (user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location)`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(&pgconn.PgError{Code: "23505"})

		_, err := repo.CreatePatient(context.Background(), patient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "patient with that email already exists")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		// ... (same as before)
		patient := &domain.Patient{
			EmailAddress: "test@example.net",
		}
		mock.ExpectExec("INSERT INTO patients").
			WillReturnError(errors.New("database error"))

		_, err := repo.CreatePatient(context.Background(), patient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetPatient(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewPatientRepository(q, log)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		expectedPatient := db.Patient{
			PatientID:              1,
			FullName:               "John Doe",
			UserID:                 sql.NullInt32{Int32: int32(1), Valid: true},
			Age:                    sql.NullInt32{Int32: int32(35), Valid: true},
			DateOfBirth:            time.Now(),
			Sex:                    "Male",
			PhoneNumber:            sql.NullString{String: "123-456-7890", Valid: true},
			EmailAddress:           sql.NullString{String: "john.doe@example.com", Valid: true},
			PreferredCommunication: db.NullPreferredCommunicationEnum{PreferredCommunicationEnum: "Email", Valid: true},
			SocioeconomicStatus:    db.NullSocioeconomicStatusEnum{SocioeconomicStatusEnum: "Middle", Valid: true},
			GeographicLocation:     sql.NullString{String: "Anytown", Valid: true},
			CreatedAt:              sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:              sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"patient_id", "user_id", "full_name", "age", "date_of_birth", "sex", "phone_number", "email_address", "preferred_communication", "socioeconomic_status", "geographic_location", "created_at", "updated_at"}).
			AddRow(expectedPatient.PatientID, expectedPatient.UserID, expectedPatient.FullName, expectedPatient.Age, expectedPatient.DateOfBirth, expectedPatient.Sex, expectedPatient.PhoneNumber, expectedPatient.EmailAddress, expectedPatient.PreferredCommunication, expectedPatient.SocioeconomicStatus, expectedPatient.GeographicLocation, expectedPatient.CreatedAt, expectedPatient.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT patient_id, user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location, created_at, updated_at FROM patients WHERE patient_id = $1")).
			WithArgs(int32(patientID)).
			WillReturnRows(rows)

		patient, err := repo.GetPatient(context.Background(), patientID)
		require.NoError(t, err)
		assert.NotNil(t, patient)

		assert.Equal(t, int(expectedPatient.PatientID), patient.PatientID)
		assert.Equal(t, expectedPatient.FullName, patient.FullName)
		// ... assertions for other fields as needed

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		patientID := 1
		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).WillReturnError(sql.ErrNoRows)
		patient, err := repo.GetPatient(context.Background(), patientID)

		assert.Nil(t, patient)
		assert.ErrorIs(t, err, domain.ErrPatientNotFound)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		patientID := 999
		mock.ExpectQuery("SELECT").WithArgs(int32(patientID)).
			WillReturnError(errors.New("database error"))
		patient, err := repo.GetPatient(context.Background(), patientID)
		assert.Nil(t, patient)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestUpdatePatient(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	log := zap.NewNop()
	q := db.New(mockDB)
	repo := NewPatientRepository(q, log)

	t.Run("success", func(t *testing.T) {
		patientID := 1
		patient := &domain.Patient{
			FullName: "John Doe Updated",
			Age:      40,
		}

		updatedPatient := db.Patient{
			PatientID:              1,
			FullName:               "John Doe Updated",
			UserID:                 sql.NullInt32{Int32: int32(1), Valid: true},
			Age:                    sql.NullInt32{Int32: int32(40), Valid: true},
			DateOfBirth:            time.Time{}, // values are zero value for not updated field
			Sex:                    "",
			PhoneNumber:            sql.NullString{},
			EmailAddress:           sql.NullString{},
			PreferredCommunication: db.NullPreferredCommunicationEnum{},
			SocioeconomicStatus:    db.NullSocioeconomicStatusEnum{},
			GeographicLocation:     sql.NullString{},
			CreatedAt:              sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:              sql.NullTime{Time: time.Now(), Valid: true},
		}

		rows := sqlmock.NewRows([]string{"patient_id", "user_id", "full_name", "age", "date_of_birth", "sex", "phone_number", "email_address", "preferred_communication", "socioeconomic_status", "geographic_location", "created_at", "updated_at"}).
			AddRow(updatedPatient.PatientID, updatedPatient.UserID, updatedPatient.FullName, updatedPatient.Age, updatedPatient.DateOfBirth, updatedPatient.Sex, updatedPatient.PhoneNumber, updatedPatient.EmailAddress, updatedPatient.PreferredCommunication, updatedPatient.SocioeconomicStatus, updatedPatient.GeographicLocation, updatedPatient.CreatedAt, updatedPatient.UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE patients SET full_name = $2, age = $3, date_of_birth = $4, sex = $5, phone_number = $6, email_address = $7, preferred_communication = $8, socioeconomic_status = $9, geographic_location = $10, updated_at = NOW() WHERE patient_id = $1 RETURNING patient_id, user_id, full_name, age, date_of_birth, sex, phone_number, email_address, preferred_communication, socioeconomic_status, geographic_location, created_at, updated_at`)).
			WithArgs(int32(patientID), patient.FullName, patient.Age, patient.DateOfBirth, patient.Sex, patient.PhoneNumber, patient.EmailAddress, patient.PreferredCommunication, patient.SocioeconomicStatus, patient.GeographicLocation).
			WillReturnRows(rows)

		p, err := repo.UpdatePatient(context.Background(), patientID, patient)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "John Doe Updated", p.FullName)
		assert.Equal(t, 40, p.Age)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		patientID := 1
		patient := &domain.Patient{
			FullName: "John Doe Updated",
			Age:      40,
		}
		mock.ExpectQuery("UPDATE patients").
			WithArgs(int32(patientID), patient.FullName, patient.Age, patient.DateOfBirth, patient.Sex, patient.PhoneNumber, patient.EmailAddress, patient.PreferredCommunication, patient.SocioeconomicStatus, patient.GeographicLocation).
			WillReturnError(sql.ErrNoRows)

		p, err := repo.UpdatePatient(context.Background(), patientID, patient)

		assert.Nil(t, p)
		assert.ErrorIs(t, err, domain.ErrPatientNotFound)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database_error", func(t *testing.T) {
		patientID := 1
		patient := &domain.Patient{
			FullName: "John Doe Updated",
			Age:      40,
		}
		mock.ExpectQuery("UPDATE patients").
			WithArgs(int32(patientID), patient.FullName, patient.Age, patient.DateOfBirth, patient.Sex, patient.PhoneNumber, patient.EmailAddress, patient.PreferredCommunication, patient.SocioeconomicStatus, patient.GeographicLocation).
			WillReturnError(errors.New("database error"))

		p, err := repo.UpdatePatient(context.Background(), patientID, patient)
		assert.Nil(t, p)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
