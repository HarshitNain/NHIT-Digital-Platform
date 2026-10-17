package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
)

// DepartmentService defines the interface for department business logic
type DepartmentService interface {
	CreateDepartment(ctx context.Context, name, description string, orgID *uuid.UUID) (*domain.Department, error)
	GetDepartment(ctx context.Context, id uuid.UUID) (*domain.Department, error)
	UpdateDepartment(ctx context.Context, id uuid.UUID, name, description string) (*domain.Department, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID) error
	ListDepartments(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Department, int32, error)
}
