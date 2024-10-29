// internal/core/ports/lifestyle_port.go
package ports

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
)

type LifestyleRepository interface {
	CreateLifestyleEntry(ctx context.Context, entry *domain.LifestyleEntry) (*domain.LifestyleEntry, error)
	GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error)
	GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error)
	UpdateLifestyleEntry(ctx context.Context, entryID int, updatedEntry *domain.LifestyleEntry) (*domain.LifestyleEntry, error)
	DeleteLifestyleEntry(ctx context.Context, entryID int) error
}

type LifestyleService interface {
	CreateLifestyleEntry(ctx context.Context, patientID int, req domain.CreateLifestyleRequest) (*domain.LifestyleEntry, error)
	GetLifestyleEntries(ctx context.Context, patientID int) ([]*domain.LifestyleEntry, error)
	GetLifestyleEntry(ctx context.Context, entryID int) (*domain.LifestyleEntry, error)
	UpdateLifestyleEntry(ctx context.Context, entryID int, req domain.UpdateLifestyleRequest) (*domain.LifestyleEntry, error)
	DeleteLifestyleEntry(ctx context.Context, entryID int) error
}
