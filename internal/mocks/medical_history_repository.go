package mocks

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// MockMedicalHistoryRepository struct
type MockMedicalHistoryRepository struct {
	mock.Mock
}

// CreateMedicalHistoryEntry mocks the CreateMedicalHistoryEntry method
func (m *MockMedicalHistoryRepository) CreateMedicalHistoryEntry(ctx context.Context, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalHistoryEntry), args.Error(1)

}

// GetMedicalHistoryEntries mocks the GetMedicalHistoryEntries method
func (m *MockMedicalHistoryRepository) GetMedicalHistoryEntry(ctx context.Context, entryID int) (*domain.MedicalHistoryEntry, error) { // Added singular mock
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalHistoryEntry), args.Error(1)
}

// GetMedicalHistoryEntries mocks the GetMedicalHistoryEntries method. This was missing, causing the error.
func (m *MockMedicalHistoryRepository) GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, patientID)

	if args.Get(0) == nil { // Check for nil to prevent panics during testing.  Updated.
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalHistoryEntry), args.Error(1)
}

// UpdateMedicalHistoryEntry mocks the UpdateMedicalHistoryEntry method
func (m *MockMedicalHistoryRepository) UpdateMedicalHistoryEntry(ctx context.Context, entryID int, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error) {
	args := m.Called(ctx, entryID, entry)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*domain.MedicalHistoryEntry), args.Error(1)
}

// DeleteMedicalHistoryEntry mocks the DeleteMedicalHistoryEntry method
func (m *MockMedicalHistoryRepository) DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error {

	args := m.Called(ctx, entryID)
	return args.Error(0)

}
