package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/ports"
)

type departmentService struct {
	repo   ports.DepartmentRepository
	logger *slog.Logger
}

// NewDepartmentService creates a new department service
func NewDepartmentService(repo ports.DepartmentRepository, logger *slog.Logger) ports.DepartmentService {
	return &departmentService{
		repo:   repo,
		logger: logger,
	}
}

// CreateDepartment creates a new department with validation
func (s *departmentService) CreateDepartment(ctx context.Context, name, description string, orgID *uuid.UUID) (*domain.Department, error) {
	// Trim and validate input
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Create department domain object
	dept := domain.NewDepartment(name, description, orgID)

	// Validate
	if err := dept.Validate(); err != nil {
		s.logger.Warn("Department validation failed", "error", err)
		return nil, err
	}

	// Check if department already exists in this organization
	exists, err := s.repo.Exists(ctx, name, orgID)
	if err != nil {
		s.logger.Error("Error checking department existence", "error", err)
		return nil, err
	}
	if exists {
		s.logger.Warn("Department already exists", "name", name, "org_id", orgID)
		return nil, domain.ErrDepartmentAlreadyExists
	}

	// Create department
	created, err := s.repo.Create(ctx, dept)
	if err != nil {
		s.logger.Error("Error creating department", "error", err)
		return nil, err
	}

	s.logger.Info("Department created", "name", created.Name, "id", created.ID)
	return created, nil
}

// GetDepartment retrieves a department by ID
func (s *departmentService) GetDepartment(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidDepartmentID
	}

	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("Department not found", "id", id, "error", err)
		return nil, domain.ErrDepartmentNotFound
	}

	return dept, nil
}

// UpdateDepartment updates a department with validation
func (s *departmentService) UpdateDepartment(ctx context.Context, id uuid.UUID, name, description string) (*domain.Department, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidDepartmentID
	}

	// Trim input
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Check if department exists
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrDepartmentNotFound
	}

	// Get existing department
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrDepartmentNotFound
	}

	// Update fields
	dept.Name = name
	dept.Description = description

	// Validate
	if err := dept.Validate(); err != nil {
		s.logger.Warn("Department update validation failed", "id", id, "error", err)
		return nil, err
	}

	// Check if name is already taken by another department in THE SAME organization
	existingDept, err := s.repo.GetByName(ctx, name, dept.OrgID)
	if err == nil && existingDept.ID != id {
		return nil, domain.ErrDepartmentAlreadyExists
	}

	// Update department
	updated, err := s.repo.Update(ctx, dept)
	if err != nil {
		s.logger.Error("Error updating department", "id", id, "error", err)
		return nil, err
	}

	s.logger.Info("Department updated", "name", updated.Name, "id", updated.ID)
	return updated, nil
}

// DeleteDepartment deletes a department with validation
func (s *departmentService) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidDepartmentID
	}

	// Check if department exists
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrDepartmentNotFound
	}

	// Check if department has users
	userCount, err := s.repo.CountUsersByDepartment(ctx, id)
	if err != nil {
		s.logger.Error("Error checking department users", "id", id, "error", err)
		return err
	}
	if userCount > 0 {
		return fmt.Errorf("%w: %d users assigned", domain.ErrDepartmentHasUsers, userCount)
	}

	// Delete department
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Error deleting department", "id", id, "error", err)
		return err
	}

	s.logger.Info("Department deleted", "id", id)
	return nil
}

// ListDepartments retrieves departments with pagination
func (s *departmentService) ListDepartments(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Department, int32, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	departments, total, err := s.repo.List(ctx, orgID, page, pageSize)
	if err != nil {
		s.logger.Error("Error listing departments", "org_id", orgID, "error", err)
		return nil, 0, err
	}

	return departments, total, nil
}
