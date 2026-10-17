package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/google/uuid"
)

// DesignationRepository defines the interface for designation data access
type DesignationRepository interface {
	// Create creates a new designation
	Create(ctx context.Context, designation *domain.Designation) error

	// GetByID retrieves a designation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Designation, error)

	// Update updates an existing designation
	Update(ctx context.Context, designation *domain.Designation) error

	// Delete deletes a designation by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves designations with pagination and filters
	List(ctx context.Context, orgID *uuid.UUID, page, pageSize int32) ([]*domain.Designation, int64, error)
}
