// internal/core/ports/medical_history_port.go
package ports

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
)

type MedicalHistoryRepository interface {
	CreateMedicalHistoryEntry(ctx context.Context, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error)
	GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error)
	GetMedicalHistoryEntry(ctx context.Context, entryID int) (*domain.MedicalHistoryEntry, error) // Add singular Get method. Updated
	UpdateMedicalHistoryEntry(ctx context.Context, entryID int, entry *domain.MedicalHistoryEntry) (*domain.MedicalHistoryEntry, error)
	DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error
}

type MedicalHistoryService interface {
	CreateMedicalHistoryEntry(ctx context.Context, patientID int, req domain.CreateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error)
	GetMedicalHistoryEntries(ctx context.Context, patientID int) ([]*domain.MedicalHistoryEntry, error)
	UpdateMedicalHistoryEntry(ctx context.Context, entryID int, req domain.UpdateMedicalHistoryRequest) (*domain.MedicalHistoryEntry, error)
	DeleteMedicalHistoryEntry(ctx context.Context, entryID int) error
}
