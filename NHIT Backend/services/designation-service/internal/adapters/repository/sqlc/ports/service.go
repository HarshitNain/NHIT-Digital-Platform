package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/google/uuid"
)

// DesignationService defines the interface for designation business logic
type DesignationService interface {
	// CreateDesignation creates a new designation with validation
	CreateDesignation(ctx context.Context, name, description string, orgID *uuid.UUID) (*domain.Designation, error)

	// GetDesignation retrieves a designation by ID
	GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error)

	// UpdateDesignation updates an existing designation
	UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string) (*domain.Designation, error)

	// DeleteDesignation deletes a designation
	DeleteDesignation(ctx context.Context, id uuid.UUID) error

	// ListDesignations retrieves designations with pagination and filters
	ListDesignations(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error)

}
