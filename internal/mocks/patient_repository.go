// internal/mocks/patient_repository.go
package mocks

import (
	"context"
	"time"

	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type MockPatientRepository struct {
	mock.Mock
}

// CreatePatient mocks the CreatePatient method
func (m *MockPatientRepository) CreatePatient(ctx context.Context, patient *domain.Patient) (*domain.Patient, error) {
	args := m.Called(ctx, patient)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Set CreatedAt and UpdatedAt to current time in mock response
	now := time.Now()
	if createdPatient, ok := args.Get(0).(*domain.Patient); ok {
		createdPatient.CreatedAt = now
		createdPatient.UpdatedAt = now
	}

	return args.Get(0).(*domain.Patient), args.Error(1)
}

func (m *MockPatientRepository) GetPatient(ctx context.Context, patientID int) (*domain.Patient, error) {
	args := m.Called(ctx, patientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Patient), args.Error(1)

}

// UpdatePatient mocks the UpdatePatient method. Sets UpdatedAt to current time.
func (m *MockPatientRepository) UpdatePatient(ctx context.Context, patientID int, patient *domain.Patient) (*domain.Patient, error) {
	args := m.Called(ctx, patientID, patient)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Set UpdatedAt to the current time
	now := time.Now()
	if updatedPatient, ok := args.Get(0).(*domain.Patient); ok {
		updatedPatient.UpdatedAt = now
	}

	return args.Get(0).(*domain.Patient), args.Error(1)

}
