package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/google/uuid"
)

// ProjectService defines the business logic interface for project operations
type ProjectService interface {
	// Project operations
	CreateProject(ctx context.Context, tenantID, orgID uuid.UUID, name, createdBy string) (*domain.Project, error)
	GetProject(ctx context.Context, projectID uuid.UUID) (*domain.Project, error)
	ListProjectsByOrganization(ctx context.Context, orgID uuid.UUID, page, pageSize int32) ([]*domain.Project, int, error)

	// Event handling
	StartEventConsumer(ctx context.Context) error
}
