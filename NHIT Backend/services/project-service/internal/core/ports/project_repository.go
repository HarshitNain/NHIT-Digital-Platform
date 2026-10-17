package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
)

// ProjectRepository defines the interface for project data operations
type ProjectRepository interface {
	// Project CRUD operations
	Create(ctx context.Context, project *domain.Project) (*domain.Project, error)
	GetByID(ctx context.Context, projectID uuid.UUID) (*domain.Project, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.Project, int, error)
}
