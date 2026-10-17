package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Helper for pgtype.UUID
func toPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func toPgUUIDFromVal(u uuid.UUID) pgtype.UUID {
	if u == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

type userRepository struct {
	queries *sqlc.Queries
	db      *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(queries *sqlc.Queries, db *pgxpool.Pool) ports.UserRepository {
	return &userRepository{queries: queries, db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	q := getQueries(ctx, r.queries)
	if user.EmailVerifiedAt != nil {
		params := sqlc.CreateUserWithVerificationParams{
			TenantID:          toPgUUIDFromVal(user.TenantID),
			OrgID:             toPgUUID(user.OrgID),
			Name:              user.Name,
			Email:             user.Email,
			Password:          user.Password,
			EmailVerifiedAt:   pgtype.Timestamptz{Time: *user.EmailVerifiedAt, Valid: true},
			DepartmentID:      toPgUUID(user.DepartmentID),
			DesignationID:     toPgUUID(user.DesignationID),
			AccountHolderName: user.AccountHolderName,
			BankName:          user.BankName,
			BankAccountNumber: user.BankAccountNumber,
			IfscCode:          user.IFSCCode,
			SignatureUrl:      user.SignatureURL,
			IsActive:          user.IsActive,
		}

		dbUser, err := q.CreateUserWithVerification(ctx, params)
		if err != nil {
			return nil, err
		}

		return toDomainUser(dbUser), nil
	}

	params := sqlc.CreateUserParams{
		TenantID:          toPgUUIDFromVal(user.TenantID),
		OrgID:             toPgUUID(user.OrgID),
		Name:              user.Name,
		Email:             user.Email,
		Password:          user.Password,
		DepartmentID:      toPgUUID(user.DepartmentID),
		DesignationID:     toPgUUID(user.DesignationID),
		AccountHolderName: user.AccountHolderName,
		BankName:          user.BankName,
		BankAccountNumber: user.BankAccountNumber,
		IfscCode:          user.IFSCCode,
		SignatureUrl:      user.SignatureURL,
		IsActive:          user.IsActive,
	}

	dbUser, err := q.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*domain.User, error) {
	params := sqlc.GetUserParams{
		UserID: toPgUUIDFromVal(userID),
		OrgID:  toPgUUID(orgID),
	}
	dbUser, err := getQueries(ctx, r.queries).GetUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string, orgID *uuid.UUID) (*domain.User, error) {
	params := sqlc.GetUserByEmailAndTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		Email:    email,
		OrgID:    toPgUUID(orgID),
	}

	dbUser, err := getQueries(ctx, r.queries).GetUserByEmailAndTenant(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {

	var deactivatedAt pgtype.Timestamptz
	if user.DeactivatedAt != nil {
		deactivatedAt = pgtype.Timestamptz{Time: *user.DeactivatedAt, Valid: true}
	}

	var deactivatedBy pgtype.UUID
	if user.DeactivatedBy != nil {
		deactivatedBy = pgtype.UUID{Bytes: *user.DeactivatedBy, Valid: true}
	}

	params := sqlc.UpdateUserParams{
		UserID:            toPgUUIDFromVal(user.UserID),
		Name:              user.Name,
		Email:             user.Email,
		Password:          user.Password,
		DepartmentID:      toPgUUID(user.DepartmentID),
		DesignationID:     toPgUUID(user.DesignationID),
		AccountHolderName: user.AccountHolderName,
		BankName:          user.BankName,
		BankAccountNumber: user.BankAccountNumber,
		IfscCode:          user.IFSCCode,
		SignatureUrl:      user.SignatureURL,
		IsActive:          user.IsActive,
		DeactivatedAt:     deactivatedAt,
		DeactivatedBy:     deactivatedBy,
		DeactivatedByName: user.DeactivatedByName,
		OrgID:             toPgUUID(user.OrgID),
	}

	dbUser, err := getQueries(ctx, r.queries).UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error {
	params := sqlc.DeleteUserParams{
		UserID: toPgUUIDFromVal(userID),
		OrgID:  toPgUUID(orgID),
	}
	return getQueries(ctx, r.queries).DeleteUser(ctx, params)
}

func (r *userRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	q := getQueries(ctx, r.queries)
	// Get total count
	total, err := q.CountUsersByTenant(ctx, sqlc.CountUsersByTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
	})
	if err != nil {
		return nil, 0, err
	}

	params := sqlc.ListUsersByTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
		Limit:    limit,
		Offset:   offset,
	}

	dbUsers, err := q.ListUsersByTenant(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, total, nil
}

func (r *userRepository) ListAll(ctx context.Context, limit, offset int32) ([]*domain.User, int64, error) {
	q := getQueries(ctx, r.queries)
	total, err := q.CountAllUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	dbUsers, err := q.ListAllUsers(ctx, sqlc.ListAllUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, total, nil
}

func (r *userRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) (int64, error) {
	return getQueries(ctx, r.queries).CountUsersByTenant(ctx, sqlc.CountUsersByTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
	})
}

func (r *userRepository) ListByDepartment(ctx context.Context, departmentID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	dbUsers, err := getQueries(ctx, r.queries).ListUsersByDepartment(ctx, sqlc.ListUsersByDepartmentParams{
		DepartmentID: toPgUUIDFromVal(departmentID),
		OrgID:        toPgUUID(orgID),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, int64(len(users)), nil // Note: We might need a separate count query for department if pagination total is needed
}

func (r *userRepository) ListByDesignation(ctx context.Context, designationID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	dbUsers, err := getQueries(ctx, r.queries).ListUsersByDesignation(ctx, sqlc.ListUsersByDesignationParams{
		DesignationID: toPgUUIDFromVal(designationID),
		OrgID:         toPgUUID(orgID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, int64(len(users)), nil
}

func (r *userRepository) SearchByName(ctx context.Context, name string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	dbUsers, err := getQueries(ctx, r.queries).SearchUsersByName(ctx, sqlc.SearchUsersByNameParams{
		Name:   name,
		OrgID:  toPgUUID(orgID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, int64(len(users)), nil
}

func (r *userRepository) SearchByEmail(ctx context.Context, email string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	dbUsers, err := getQueries(ctx, r.queries).SearchUsersByEmail(ctx, sqlc.SearchUsersByEmailParams{
		Email:  email,
		OrgID:  toPgUUID(orgID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, int64(len(users)), nil
}

func (r *userRepository) ListByOrganization(ctx context.Context, tenantID, orgID uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	q := getQueries(ctx, r.queries)
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM users u
		WHERE u.tenant_id = $1 AND u.org_id = $2`
	
	var total int64
	err := r.db.QueryRow(ctx, countQuery, tenantID, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	params := sqlc.ListUsersByTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUIDFromVal(orgID),
		Limit:    limit,
		Offset:   offset,
	}

	dbUsers, err := q.ListUsersByTenant(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, total, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLoginIP, userAgent string) error {
	var ip *string
	if lastLoginIP != "" {
		ip = &lastLoginIP
	}
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}
	
	params := sqlc.UpdateUserLastLoginParams{
		UserID:      toPgUUIDFromVal(userID),
		LastLoginAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		LastLoginIp: ip,
		UserAgent:   ua,
		OrgID:       toPgUUID(nil), // Optionally add orgID here if needed, passing nil for now
	}

	_, err := getQueries(ctx, r.queries).UpdateUserLastLogin(ctx, params)
	return err
}

func (r *userRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(withTxContext(ctx, tx)); err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// Helper functions to convert between db and domain models
func toDomainUser(dbUser *sqlc.User) *domain.User {
	var deptID *uuid.UUID
	if dbUser.DepartmentID.Valid {
		id := dbUser.DepartmentID.UUID
		deptID = &id
	}

	var desigID *uuid.UUID
	if dbUser.DesignationID.Valid {
		id := dbUser.DesignationID.UUID
		desigID = &id
	}
	
	var deactivatedBy *uuid.UUID
	if dbUser.DeactivatedBy.Valid {
		id := uuid.UUID(dbUser.DeactivatedBy.Bytes)
		deactivatedBy = &id
	}

	var orgID *uuid.UUID
	if dbUser.OrgID.Valid {
		id := uuid.UUID(dbUser.OrgID.Bytes)
		orgID = &id
	}

	return &domain.User{
		UserID:            dbUser.UserID,
		TenantID:          dbUser.TenantID,
		OrgID:             orgID,
		Name:              dbUser.Name,
		Email:             dbUser.Email,
		Password:          dbUser.Password,
		DepartmentID:      deptID,
		DesignationID:     desigID,
		AccountHolderName: dbUser.AccountHolderName,
		BankName:          dbUser.BankName,
		BankAccountNumber: dbUser.BankAccountNumber,
		IFSCCode:          dbUser.IfscCode,
		SignatureURL:      dbUser.SignatureUrl,
		IsActive:          dbUser.IsActive,
		DeactivatedAt:     fromPgTimestamptz(dbUser.DeactivatedAt),
		DeactivatedBy:     deactivatedBy,
		EmailVerifiedAt:   fromPgTimestamptz(dbUser.EmailVerifiedAt),
		LastLoginAt:       fromPgTimestamptz(dbUser.LastLoginAt),
		LastLogoutAt:      fromPgTimestamptz(dbUser.LastLogoutAt),
		LastLoginIP:       fromStringPtr(dbUser.LastLoginIp),
		UserAgent:         fromStringPtr(dbUser.UserAgent),
		CreatedAt:         dbUser.CreatedAt.Time,
		UpdatedAt:         dbUser.UpdatedAt.Time,
	}
}

func fromPgTimestamptz(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func fromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
