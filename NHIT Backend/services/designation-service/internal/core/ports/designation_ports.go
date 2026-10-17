package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
)

// DesignationRepository defines the interface for designation data persistence
type DesignationRepository interface {
	Create(ctx context.Context, designation *domain.Designation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Designation, error)
	Update(ctx context.Context, designation *domain.Designation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error)
	Exists(ctx context.Context, name string, orgID *uuid.UUID) (bool, error)
	GetByName(ctx context.Context, name string, orgID *uuid.UUID) (*domain.Designation, error)
}

// DesignationService defines the interface for designation business logic
type DesignationService interface {
	CreateDesignation(ctx context.Context, name, description string, orgID *uuid.UUID) (*domain.Designation, error)
	GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error)
	UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string) (*domain.Designation, error)
	DeleteDesignation(ctx context.Context, id uuid.UUID) error
	ListDesignations(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error)
}
