// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: lifestyle.sql

package db

import (
	"context"
	"database/sql"
)

const createLifestyleEntry = `-- name: CreateLifestyleEntry :one
INSERT INTO patient_lifestyle (patient_id, lifestyle_factor, value, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at
`

type CreateLifestyleEntryParams struct {
	PatientID       int32          `json:"patient_id"`
	LifestyleFactor string         `json:"lifestyle_factor"`
	Value           sql.NullString `json:"value"`
	StartDate       sql.NullTime   `json:"start_date"`
	EndDate         sql.NullTime   `json:"end_date"`
}

func (q *Queries) CreateLifestyleEntry(ctx context.Context, arg CreateLifestyleEntryParams) (PatientLifestyle, error) {
	row := q.db.QueryRowContext(ctx, createLifestyleEntry,
		arg.PatientID,
		arg.LifestyleFactor,
		arg.Value,
		arg.StartDate,
		arg.EndDate,
	)
	var i PatientLifestyle
	err := row.Scan(
		&i.PatientLifestyleID,
		&i.PatientID,
		&i.LifestyleFactor,
		&i.Value,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteLifestyleEntry = `-- name: DeleteLifestyleEntry :exec
DELETE FROM patient_lifestyle
WHERE patient_lifestyle_id = $1
`

func (q *Queries) DeleteLifestyleEntry(ctx context.Context, patientLifestyleID int32) error {
	_, err := q.db.ExecContext(ctx, deleteLifestyleEntry, patientLifestyleID)
	return err
}

const getLifestyleEntries = `-- name: GetLifestyleEntries :many
SELECT patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at
FROM patient_lifestyle
WHERE patient_id = $1
`

func (q *Queries) GetLifestyleEntries(ctx context.Context, patientID int32) ([]PatientLifestyle, error) {
	rows, err := q.db.QueryContext(ctx, getLifestyleEntries, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PatientLifestyle{}
	for rows.Next() {
		var i PatientLifestyle
		if err := rows.Scan(
			&i.PatientLifestyleID,
			&i.PatientID,
			&i.LifestyleFactor,
			&i.Value,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLifestyleEntry = `-- name: GetLifestyleEntry :one
SELECT patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at 
FROM patient_lifestyle
WHERE patient_lifestyle_id = $1
`

func (q *Queries) GetLifestyleEntry(ctx context.Context, patientLifestyleID int32) (PatientLifestyle, error) {
	row := q.db.QueryRowContext(ctx, getLifestyleEntry, patientLifestyleID)
	var i PatientLifestyle
	err := row.Scan(
		&i.PatientLifestyleID,
		&i.PatientID,
		&i.LifestyleFactor,
		&i.Value,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateLifestyleEntry = `-- name: UpdateLifestyleEntry :one
UPDATE patient_lifestyle
SET lifestyle_factor = $2,
    value = $3,
    start_date = $4,
    end_date = $5
WHERE patient_lifestyle_id = $1
RETURNING patient_lifestyle_id, patient_id, lifestyle_factor, value, start_date, end_date, created_at, updated_at
`

type UpdateLifestyleEntryParams struct {
	PatientLifestyleID int32          `json:"patient_lifestyle_id"`
	LifestyleFactor    string         `json:"lifestyle_factor"`
	Value              sql.NullString `json:"value"`
	StartDate          sql.NullTime   `json:"start_date"`
	EndDate            sql.NullTime   `json:"end_date"`
}

func (q *Queries) UpdateLifestyleEntry(ctx context.Context, arg UpdateLifestyleEntryParams) (PatientLifestyle, error) {
	row := q.db.QueryRowContext(ctx, updateLifestyleEntry,
		arg.PatientLifestyleID,
		arg.LifestyleFactor,
		arg.Value,
		arg.StartDate,
		arg.EndDate,
	)
	var i PatientLifestyle
	err := row.Scan(
		&i.PatientLifestyleID,
		&i.PatientID,
		&i.LifestyleFactor,
		&i.Value,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}