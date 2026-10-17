package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

// Service contains dependencies for organization business logic.
type Service struct {
	repo   ports.Repository
	kafka  ports.KafkaPublisher
	logger *slog.Logger
	// Optionally add metrics/tracing clients here.
}

// NewService creates a new organization service.
func NewService(repo ports.Repository, kafka ports.KafkaPublisher, logger *slog.Logger) *Service {
	return &Service{repo: repo, kafka: kafka, logger: logger}
}

// Constants and configuration used across service.
const (
	defaultPageSize = 20
	maxPageSize     = 200
	bcryptCost      = bcrypt.DefaultCost // can bump to 12 in prod if CPU allows
)

// ErrValidation is returned for user input validation errors.
var ErrValidation = errors.New("validation error")

// CreateOrganization implements production-grade business logic for creating an organization.
// Rules implemented (per your grpc comments and extra safety):
//   - name & code required
//   - if ParentOrgID is empty => parent org creation -> super admin MUST be present
//   - if ParentOrgID is present  => child org creation -> super admin MUST NOT be set (ignored if provided)
//   - code must be unique
//   - tenant id generated if empty
//   - super admin password hashed before saving
//   - db name generated if not provided
func (s *Service) CreateOrganization(ctx context.Context, in ports.OrganizationModel) (ports.OrganizationModel, error) {
	// basic validation
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Code) == "" {
		return ports.OrganizationModel{}, fmt.Errorf("%w: name and code required", ErrValidation)
	}

	// sanitize code, name
	in.Code = strings.TrimSpace(in.Code)
	in.Name = strings.TrimSpace(in.Name)

	// Ensure code uniqueness
	existing, err := s.tryGetByCode(ctx, in.Code)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		// unexpected repo error
		s.logger.Error("Error checking code uniqueness", "code", in.Code, "error", err)
		return ports.OrganizationModel{}, fmt.Errorf("failed to check code uniqueness: %w", err)
	}
	if existing != nil {
		return ports.OrganizationModel{}, fmt.Errorf("%w: organization code already exists", ErrValidation)
	}

	isParent := in.ParentOrgID == nil || strings.TrimSpace(*in.ParentOrgID) == ""
	if isParent {
		// Creating a root organization — must have super admin details
		if in.SuperAdminName == nil || in.SuperAdminEmail == nil || in.SuperAdminPass == nil {
			return ports.OrganizationModel{}, fmt.Errorf("%w: super admin required for parent organization", ErrValidation)
		}
	} else {
		// Creating a child org — ensure parent exists and ignore super admin
		parentID := strings.TrimSpace(*in.ParentOrgID)
		if parentID == "" {
			return ports.OrganizationModel{}, fmt.Errorf("%w: parent_org_id cannot be blank", ErrValidation)
		}
		if _, err := uuid.Parse(parentID); err != nil {
			return ports.OrganizationModel{}, fmt.Errorf("%w: invalid parent_org_id: %v", ErrValidation, err)
		}
		// confirm parent exists
		if _, err := s.repo.GetOrganizationByID(ctx, parentID); err != nil {
			s.logger.Warn("Parent organization not found", "parent_id", parentID, "error", err)
			return ports.OrganizationModel{}, fmt.Errorf("%w: parent organization not found", ErrValidation)
		}
		// ignore any provided super admin for child
		in.SuperAdminName = nil
		in.SuperAdminEmail = nil
		in.SuperAdminPass = nil
	}

	// set generated OrgID if not provided
	if strings.TrimSpace(in.OrgID) == "" {
		in.OrgID = uuid.New().String()
	} else {
		// validate provided OrgID
		if _, err := uuid.Parse(in.OrgID); err != nil {
			return ports.OrganizationModel{}, fmt.Errorf("%w: invalid org_id: %v", ErrValidation, err)
		}
	}

	// set or generate TenantID
	if strings.TrimSpace(in.TenantID) == "" {
		in.TenantID = uuid.New().String()
	} else {
		if _, err := uuid.Parse(in.TenantID); err != nil {
			return ports.OrganizationModel{}, fmt.Errorf("%w: invalid tenant_id: %v", ErrValidation, err)
		}
	}

	// database name generation if not set
	if strings.TrimSpace(in.DatabaseName) == "" {
		in.DatabaseName = generateDatabaseName(in.Code)
	}

	// prepare timestamps
	now := time.Now().UTC()
	in.CreatedAt = now
	in.UpdatedAt = now

	// hash super admin password if present (only for parent creation)
	if in.SuperAdminPass != nil && *in.SuperAdminPass != "" {
		hashed, err := hashPassword(*in.SuperAdminPass)
		if err != nil {
			s.logger.Error("Error hashing password", "error", err)
			return ports.OrganizationModel{}, fmt.Errorf("failed to secure super admin credentials: %w", err)
		}
		in.SuperAdminPass = &hashed
	}

	// call repository to persist
	created, err := s.repo.CreateOrganization(ctx, in)
	if err != nil {
		s.logger.Error("Repository error in CreateOrganization", "error", err)
		return ports.OrganizationModel{}, fmt.Errorf("failed to create organization: %w", err)
	}

	// zero-out sensitive fields before returning (best practice)
	if created.SuperAdminPass != nil {
		empty := ""
		created.SuperAdminPass = &empty
	}

	// Publish organization created event for project service to create initial projects
	if s.kafka != nil && len(in.InitialProjects) > 0 {
		createdBy := ""
		if in.SuperAdminName != nil {
			createdBy = *in.SuperAdminName
		}

		event := domain.NewOrganizationCreatedEvent(
			in.TenantID,
			in.OrgID,
			in.Name,
			createdBy,
			in.InitialProjects,
		)

		if err := s.kafka.Publish(ctx, "organization.events", event); err != nil {
			s.logger.Error("Failed to publish organization created event", "error", err)
			// Don't fail the operation if Kafka fails, just log it
		}
	}

	return created, nil
}

// GetOrganizationByID retrieves a single organization by id.
func (s *Service) GetOrganizationByID(ctx context.Context, orgID string) (ports.OrganizationModel, error) {
	if strings.TrimSpace(orgID) == "" {
		return ports.OrganizationModel{}, fmt.Errorf("%w: org_id required", ErrValidation)
	}
	if _, err := uuid.Parse(orgID); err != nil {
		return ports.OrganizationModel{}, fmt.Errorf("%w: invalid org_id: %v", ErrValidation, err)
	}

	o, err := s.repo.GetOrganizationByID(ctx, orgID)
	if err != nil {
		s.logger.Error("Repository error in GetOrganizationByID", "org_id", orgID, "error", err)
		return ports.OrganizationModel{}, err
	}
	// hide password in result
	if o.SuperAdminPass != nil {
		empty := ""
		o.SuperAdminPass = &empty
	}
	return o, nil
}

// GetOrganizationByCode returns an organization by code.
func (s *Service) GetOrganizationByCode(ctx context.Context, code string) (ports.OrganizationModel, error) {
	if strings.TrimSpace(code) == "" {
		return ports.OrganizationModel{}, fmt.Errorf("%w: code required", ErrValidation)
	}
	o, err := s.repo.GetOrganizationByCode(ctx, code)
	if err != nil {
		s.logger.Error("Repository error in GetOrganizationByCode", "code", code, "error", err)
		return ports.OrganizationModel{}, err
	}
	if o.SuperAdminPass != nil {
		empty := ""
		o.SuperAdminPass = &empty
	}
	return o, nil
}

// ListOrganizations provides pagination and sanitization.
func (s *Service) ListOrganizations(ctx context.Context, page, pageSize int) ([]ports.OrganizationModel, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	offset := (page - 1) * pageSize
	result, total, err := s.repo.ListOrganizations(ctx, offset, pageSize)
	if err != nil {
		s.logger.Error("Repository error in ListOrganizations", "error", err)
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}
	// redaction
	for i := range result {
		if result[i].SuperAdminPass != nil {
			empty := ""
			result[i].SuperAdminPass = &empty
		}
	}
	return result, total, nil
}

// ListOrganizationsByTenant returns orgs for a tenant with pagination.
func (s *Service) ListOrganizationsByTenant(ctx context.Context, tenantID string, page, pageSize int) ([]ports.OrganizationModel, int, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, 0, fmt.Errorf("%w: tenant_id required", ErrValidation)
	}
	if _, err := uuid.Parse(tenantID); err != nil {
		return nil, 0, fmt.Errorf("%w: invalid tenant_id: %v", ErrValidation, err)
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize

	list, total, err := s.repo.ListOrganizationsByTenant(ctx, tenantID, offset, pageSize)
	if err != nil {
		s.logger.Error("Repository error in ListOrganizationsByTenant", "tenant_id", tenantID, "error", err)
		return nil, 0, fmt.Errorf("failed to list organizations by tenant: %w", err)
	}
	for i := range list {
		if list[i].SuperAdminPass != nil {
			empty := ""
			list[i].SuperAdminPass = &empty
		}
	}
	return list, total, nil
}

// ListChildOrganizations returns children of a parent org.
func (s *Service) ListChildOrganizations(ctx context.Context, parentOrgID string, page, pageSize int) ([]ports.OrganizationModel, int, error) {
	if strings.TrimSpace(parentOrgID) == "" {
		return nil, 0, fmt.Errorf("%w: parent_org_id required", ErrValidation)
	}
	if _, err := uuid.Parse(parentOrgID); err != nil {
		return nil, 0, fmt.Errorf("%w: invalid parent_org_id: %v", ErrValidation, err)
	}
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	offset := (page - 1) * pageSize

	list, total, err := s.repo.ListChildOrganizations(ctx, parentOrgID, offset, pageSize)
	if err != nil {
		s.logger.Error("Repository error in ListChildOrganizations", "parent_org_id", parentOrgID, "error", err)
		return nil, 0, fmt.Errorf("failed to list child organizations: %w", err)
	}
	for i := range list {
		if list[i].SuperAdminPass != nil {
			empty := ""
			list[i].SuperAdminPass = &empty
		}
	}
	return list, total, nil
}

// UpdateOrganization updates mutable fields of an organization (name, code, description, logo, status).
// It validates uniqueness when code changes.
func (s *Service) UpdateOrganization(ctx context.Context, in ports.OrganizationModel) (ports.OrganizationModel, error) {
	if strings.TrimSpace(in.OrgID) == "" {
		return ports.OrganizationModel{}, fmt.Errorf("%w: org_id required", ErrValidation)
	}
	if _, err := uuid.Parse(in.OrgID); err != nil {
		return ports.OrganizationModel{}, fmt.Errorf("%w: invalid org_id: %v", ErrValidation, err)
	}

	// Fetch existing
	existing, err := s.repo.GetOrganizationByID(ctx, in.OrgID)
	if err != nil {
		s.logger.Error("Repository error in UpdateOrganization (Fetch)", "org_id", in.OrgID, "error", err)
		return ports.OrganizationModel{}, err
	}

	// If code changed, ensure unique
	if in.Code != "" && in.Code != existing.Code {
		existingWithCode, err := s.tryGetByCode(ctx, in.Code)
		if err != nil && !errors.Is(err, ports.ErrNotFound) {
			s.logger.Error("Error checking code uniqueness in update", "code", in.Code, "error", err)
			return ports.OrganizationModel{}, fmt.Errorf("failed to validate code uniqueness: %w", err)
		}
		if existingWithCode != nil {
			return ports.OrganizationModel{}, fmt.Errorf("%w: organization code already exists", ErrValidation)
		}
		existing.Code = in.Code
	}
	// update mutable fields
	if in.Name != "" {
		existing.Name = in.Name
	}
	if in.Description != nil {
		existing.Description = in.Description
	}
	if in.Logo != nil {
		existing.Logo = in.Logo
	}
	// status must be explicitly set by caller
	existing.Status = in.Status
	existing.UpdatedAt = time.Now().UTC()

	updated, err := s.repo.UpdateOrganization(ctx, existing)
	if err != nil {
		s.logger.Error("Repository error in UpdateOrganization", "org_id", in.OrgID, "error", err)
		return ports.OrganizationModel{}, fmt.Errorf("failed to update organization: %w", err)
	}

	// redact sensitive fields
	if updated.SuperAdminPass != nil {
		empty := ""
		updated.SuperAdminPass = &empty
	}
	return updated, nil
}

// DeleteOrganization implements deletion with validation.
func (s *Service) DeleteOrganization(ctx context.Context, orgID string) error {
	if strings.TrimSpace(orgID) == "" {
		return fmt.Errorf("%w: org_id required", ErrValidation)
	}
	if _, err := uuid.Parse(orgID); err != nil {
		return fmt.Errorf("%w: invalid org_id: %v", ErrValidation, err)
	}
	// Optionally: check for children or dependencies before deletion — business decision.
	// Example: block deletion if child orgs exist
	children, _, err := s.repo.ListChildOrganizations(ctx, orgID, 0, 1)
	if err != nil {
		s.logger.Error("Error checking child orgs before delete", "org_id", orgID, "error", err)
		return fmt.Errorf("failed to check children before delete: %w", err)
	}
	if len(children) > 0 {
		return fmt.Errorf("%w: cannot delete organization with child organizations", ErrValidation)
	}

	if err := s.repo.DeleteOrganization(ctx, orgID); err != nil {
		s.logger.Error("Repository error in DeleteOrganization", "org_id", orgID, "error", err)
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

// ------------------ helpers ------------------

// hashPassword uses bcrypt to hash the given password.
func hashPassword(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	bs, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// generateDatabaseName returns a safe db name based on code (lowercase + sanitized).
func generateDatabaseName(code string) string {
	clean := strings.ToLower(strings.TrimSpace(code))
	clean = strings.ReplaceAll(clean, "-", "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	if clean == "" {
		clean = "orgdb"
	}
	return fmt.Sprintf("%s_db", clean)
}

// tryGetByCode returns a pointer to OrganizationModel or (nil, ports.ErrNotFound).
// It hides the concrete repo NotFound behavior to the service.
func (s *Service) tryGetByCode(ctx context.Context, code string) (*ports.OrganizationModel, error) {
	o, err := s.repo.GetOrganizationByCode(ctx, code)
	if err != nil {
		// If your repository returns a sentinel error for not found, use that.
		// Otherwise, inspect as needed — here we assume ports.ErrNotFound is defined.
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}
