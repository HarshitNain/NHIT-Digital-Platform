package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
)

type designationRepository struct {
	queries *sqlc.Queries
}

// NewDesignationRepository creates a new designation repository
func NewDesignationRepository(queries *sqlc.Queries) ports.DesignationRepository {
	return &designationRepository{
		queries: queries,
	}
}

// Create creates a new designation
func (r *designationRepository) Create(ctx context.Context, designation *domain.Designation) error {
	var orgID pgtype.UUID
	if designation.OrgID != nil {
		orgID = pgtype.UUID{Bytes: *designation.OrgID, Valid: true}
	} else {
		orgID = pgtype.UUID{Valid: false}
	}

	_, err := r.queries.CreateDesignation(ctx, sqlc.CreateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		CreatedAt:   pgtype.Timestamptz{Time: designation.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: designation.UpdatedAt, Valid: true},
		OrgID:       orgID,
	})

	return err
}

// GetByID retrieves a designation by ID
func (r *designationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Designation, error) {
	dbDesignation, err := r.queries.GetDesignationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toDomainDesignation(dbDesignation), nil
}

// Update updates an existing designation
func (r *designationRepository) Update(ctx context.Context, designation *domain.Designation) error {
	_, err := r.queries.UpdateDesignation(ctx, sqlc.UpdateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		UpdatedAt:   pgtype.Timestamptz{Time: designation.UpdatedAt, Valid: true},
	})

	return err
}

// Delete deletes a designation by ID
func (r *designationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDesignation(ctx, id)
}

func (r *designationRepository) List(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error) {
	offset := (page - 1) * pageSize

	var orgIDParam pgtype.UUID
	if orgID != nil {
		orgIDParam = pgtype.UUID{Bytes: *orgID, Valid: true}
	} else {
		orgIDParam = pgtype.UUID{Valid: false}
	}

	// Get total count
	totalCount, err := r.queries.CountDesignations(ctx, orgIDParam)
	if err != nil {
		return nil, 0, err
	}

	dbDesignations, err := r.queries.ListDesignations(ctx, sqlc.ListDesignationsParams{
		Column1: orgIDParam,
		Limit:   pageSize,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	designations := make([]*domain.Designation, len(dbDesignations))
	for i, d := range dbDesignations {
		designations[i] = toDomainDesignation(d)
	}

	return designations, totalCount, nil
}

func (r *designationRepository) Exists(ctx context.Context, name string, orgID *uuid.UUID) (bool, error) {
	var orgIDParam pgtype.UUID
	if orgID != nil {
		orgIDParam = pgtype.UUID{Bytes: *orgID, Valid: true}
	} else {
		orgIDParam = pgtype.UUID{Valid: false}
	}

	return r.queries.DesignationExists(ctx, sqlc.DesignationExistsParams{
		Lower:   name,
		OrgID: orgIDParam,
	})
}

func (r *designationRepository) GetByName(ctx context.Context, name string, orgID *uuid.UUID) (*domain.Designation, error) {
	var orgIDParam pgtype.UUID
	if orgID != nil {
		orgIDParam = pgtype.UUID{Bytes: *orgID, Valid: true}
	} else {
		orgIDParam = pgtype.UUID{Valid: false}
	}

	dbDesignation, err := r.queries.GetDesignationByName(ctx, sqlc.GetDesignationByNameParams{
		Lower:   name,
		OrgID: orgIDParam,
	})
	if err != nil {
		return nil, err
	}

	return toDomainDesignation(dbDesignation), nil
}

// toDomainDesignation converts a database designation to a domain designation
func toDomainDesignation(dbDesignation *sqlc.Designation) *domain.Designation {
	var orgID *uuid.UUID
	if dbDesignation.OrgID.Valid {
		id := uuid.UUID(dbDesignation.OrgID.Bytes)
		orgID = &id
	}

	return &domain.Designation{
		ID:          dbDesignation.ID,
		OrgID:       orgID,
		Name:        dbDesignation.Name,
		Description: dbDesignation.Description,
		CreatedAt:   dbDesignation.CreatedAt.Time,
		UpdatedAt:   dbDesignation.UpdatedAt.Time,
	}
}

