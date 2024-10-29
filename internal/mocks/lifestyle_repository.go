// internal/mocks/lifestyle_repository.go
package mocks

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type MockLifestyleRepository struct {
	mock.Mock
}

func (m *MockLifestyleRepository) CreateLifestyleEntry(ctx context.Context, entry *domain.LifestyleEntry) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, entry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)
}

func (m *MockLifestyleRepository) GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error) {
	args := m.Called(ctx, patientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LifestyleEntry), args.Error(1)
}

func (m *MockLifestyleRepository) GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, entryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)

}

func (m *MockLifestyleRepository) UpdateLifestyleEntry(ctx context.Context, entryID int, updatedEntry *domain.LifestyleEntry) (*domain.LifestyleEntry, error) {
	args := m.Called(ctx, entryID, updatedEntry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifestyleEntry), args.Error(1)
}

func (m *MockLifestyleRepository) DeleteLifestyleEntry(ctx context.Context, entryID int) error {
	args := m.Called(ctx, entryID)
	return args.Error(0)
}
