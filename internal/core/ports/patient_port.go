// internal/core/ports/patient_port.go
package ports

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
)

type PatientRepository interface {
	CreatePatient(ctx context.Context, patient *domain.Patient) (*domain.Patient, error)
	GetPatient(ctx context.Context, patientID int) (*domain.Patient, error)
	UpdatePatient(ctx context.Context, patientID int, patient *domain.Patient) (*domain.Patient, error)
}

type PatientService interface {
	CreatePatient(ctx context.Context, req domain.CreatePatientRequest) (*domain.Patient, error)
	GetPatient(ctx context.Context, patientID int) (*domain.Patient, error)
	UpdatePatient(ctx context.Context, patientID int, req domain.UpdatePatientRequest) (*domain.Patient, error)
}
