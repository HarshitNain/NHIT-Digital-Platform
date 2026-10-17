package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
)

// DepartmentRepository defines the interface for department data operations
type DepartmentRepository interface {
	Create(ctx context.Context, department *domain.Department) (*domain.Department, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Department, error)
	GetByName(ctx context.Context, name string, orgID *uuid.UUID) (*domain.Department, error)
	Update(ctx context.Context, department *domain.Department) (*domain.Department, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Department, int32, error)
	Exists(ctx context.Context, name string, orgID *uuid.UUID) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	CountUsersByDepartment(ctx context.Context, departmentID uuid.UUID) (int32, error)
}
