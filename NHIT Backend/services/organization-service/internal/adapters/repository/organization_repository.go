package repository

import (
	"context"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	db "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrganizationRepository struct {
	q *db.Queries
}

func NewOrganizationRepository(dbConn *pgxpool.Pool) ports.Repository {
	return &OrganizationRepository{
		q: db.New(dbConn),
	}
}

func toPgUUID(id *string) (pgtype.UUID, error) {
	if id == nil || *id == "" {
		return pgtype.UUID{Valid: false}, nil
	}

	parsed, err := uuid.Parse(*id)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: parsed,
		Valid: true,
	}, nil
}

func stringPtrFromPgUUID(id pgtype.UUID) *string {
	if !id.Valid {
		return nil
	}
	value := uuid.UUID(id.Bytes).String()
	return &value
}

// Create
func (r *OrganizationRepository) CreateOrganization(ctx context.Context, org ports.OrganizationModel) (ports.OrganizationModel, error) {
	orgID, err := uuid.Parse(org.OrgID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	tenantID, err := uuid.Parse(org.TenantID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	parentOrgID, err := toPgUUID(org.ParentOrgID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	row, err := r.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		OrgID:              orgID,
		TenantID:           tenantID,
		ParentOrgID:        parentOrgID,
		Name:               org.Name,
		Code:               org.Code,
		DatabaseName:       org.DatabaseName,
		Description:        org.Description,
		Logo:               org.Logo,
		SuperAdminName:     org.SuperAdminName,
		SuperAdminEmail:    org.SuperAdminEmail,
		SuperAdminPassword: org.SuperAdminPass,
		InitialProjects:    org.InitialProjects,
		Status:             int16(org.Status),
	})
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	return r.convertSQLC(row), nil
}

// Get By ID
func (r *OrganizationRepository) GetOrganizationByID(ctx context.Context, orgID string) (ports.OrganizationModel, error) {
	parsedID, err := uuid.Parse(orgID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	row, err := r.q.GetOrganizationByID(ctx, parsedID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}
	return r.convertSQLC(row), nil
}

// Get By Code
func (r *OrganizationRepository) GetOrganizationByCode(ctx context.Context, code string) (ports.OrganizationModel, error) {
	row, err := r.q.GetOrganizationByCode(ctx, code)
	if err != nil {
		return ports.OrganizationModel{}, err
	}
	return r.convertSQLC(row), nil
}

// List
func (r *OrganizationRepository) ListOrganizations(ctx context.Context, offset, limit int) ([]ports.OrganizationModel, int, error) {
	rows, err := r.q.ListOrganizations(ctx, db.ListOrganizationsParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.CountOrganizations(ctx)
	if err != nil {
		return nil, 0, err
	}

	return r.convertAll(rows), int(total), nil
}

// List By Tenant
func (r *OrganizationRepository) ListOrganizationsByTenant(ctx context.Context, tenantID string, offset, limit int) ([]ports.OrganizationModel, int, error) {
	parsedTenantID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.ListOrganizationsByTenant(ctx, db.ListOrganizationsByTenantParams{
		TenantID: parsedTenantID,
		Offset:   int32(offset),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.CountOrganizationsByTenant(ctx, parsedTenantID)
	if err != nil {
		return nil, 0, err
	}

	return r.convertAll(rows), int(total), nil
}

// List Children
func (r *OrganizationRepository) ListChildOrganizations(ctx context.Context, parentOrgID string, offset, limit int) ([]ports.OrganizationModel, int, error) {
	parentID := parentOrgID
	pgParentID, err := toPgUUID(&parentID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.ListChildOrganizations(ctx, db.ListChildOrganizationsParams{
		ParentOrgID: pgParentID,
		Offset:      int32(offset),
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.q.CountChildOrganizations(ctx, pgParentID)
	if err != nil {
		return nil, 0, err
	}

	return r.convertAll(rows), int(total), nil
}

// Update
func (r *OrganizationRepository) UpdateOrganization(ctx context.Context, org ports.OrganizationModel) (ports.OrganizationModel, error) {
	orgUUID, err := uuid.Parse(org.OrgID)
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	row, err := r.q.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		OrgID:       orgUUID,
		Name:        org.Name,
		Code:        org.Code,
		Description: org.Description,
		Logo:        org.Logo,
		Status:      int16(org.Status),
	})
	if err != nil {
		return ports.OrganizationModel{}, err
	}

	return r.convertSQLC(row), nil
}

// Delete
func (r *OrganizationRepository) DeleteOrganization(ctx context.Context, orgID string) error {
	parsedID, err := uuid.Parse(orgID)
	if err != nil {
		return err
	}
	return r.q.DeleteOrganization(ctx, parsedID)
}

// Helpers
func (r *OrganizationRepository) convertSQLC(row *db.Organization) ports.OrganizationModel {
	if row == nil {
		return ports.OrganizationModel{}
	}

	return ports.OrganizationModel{
		OrgID:           row.OrgID.String(),
		TenantID:        row.TenantID.String(),
		ParentOrgID:     stringPtrFromPgUUID(row.ParentOrgID),
		Name:            row.Name,
		Code:            row.Code,
		DatabaseName:    row.DatabaseName,
		Description:     row.Description,
		Logo:            row.Logo,
		SuperAdminName:  row.SuperAdminName,
		SuperAdminEmail: row.SuperAdminEmail,
		SuperAdminPass:  row.SuperAdminPassword,
		InitialProjects: row.InitialProjects,
		Status:          pb.OrganizationStatus(row.Status),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func (r *OrganizationRepository) convertAll(rows []*db.Organization) []ports.OrganizationModel {
	list := make([]ports.OrganizationModel, 0, len(rows))
	for _, row := range rows {
		list = append(list, r.convertSQLC(row))
	}
	return list
}
