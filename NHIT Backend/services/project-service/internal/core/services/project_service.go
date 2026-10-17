package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
	"github.com/google/uuid"
)

type projectService struct {
	repo   ports.ProjectRepository
	kafka  ports.KafkaConsumer
	logger *log.Logger
}

// NewProjectService creates a new project service instance
func NewProjectService(repo ports.ProjectRepository, kafka ports.KafkaConsumer, logger *log.Logger) ports.ProjectService {
	if logger == nil {
		logger = log.Default()
	}
	return &projectService{
		repo:   repo,
		kafka:  kafka,
		logger: logger,
	}
}

// CreateProject creates a new project with validation
func (s *projectService) CreateProject(ctx context.Context, tenantID, orgID uuid.UUID, name, createdBy string) (*domain.Project, error) {

	// Create new project
	project, err := domain.NewProject(tenantID, orgID, name, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Save to repository
	createdProject, err := s.repo.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// TODO: Publish Kafka event - ProjectCreated
	// TODO: Log activity

	return createdProject, nil
}

// GetProject retrieves a project by ID
func (s *projectService) GetProject(ctx context.Context, projectID uuid.UUID) (*domain.Project, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// ListProjectsByOrganization lists all projects for an organization with pagination
func (s *projectService) ListProjectsByOrganization(ctx context.Context, orgID uuid.UUID, page, pageSize int32) ([]*domain.Project, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	limit := int(pageSize)
	offset := int((page - 1) * pageSize)

	projects, totalCount, err := s.repo.ListByOrganization(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, totalCount, nil
}

// HandleOrganizationCreatedEvent processes organization created events
func (s *projectService) HandleOrganizationCreatedEvent(ctx context.Context, event *domain.OrganizationCreatedEvent) error {
	if len(event.Projects) == 0 {
		s.logger.Printf("No projects to create for organization %s", event.OrgID)
		return nil
	}

	projects, err := domain.NewProjectFromEvent(event)
	if err != nil {
		return fmt.Errorf("failed to create projects from event: %w", err)
	}

	for _, project := range projects {
		_, err := s.repo.Create(ctx, project)
		if err != nil {
			s.logger.Printf("Failed to create project '%s' for org %s: %v", project.ProjectName, event.OrgID, err)
			continue // Continue with other projects even if one fails
		}
		s.logger.Printf("Successfully created project '%s' for org %s", project.ProjectName, event.OrgID)
	}

	return nil
}

// StartEventConsumer starts consuming Kafka events
func (s *projectService) StartEventConsumer(ctx context.Context) error {
	if s.kafka == nil {
		s.logger.Println("Kafka consumer not configured, skipping event consumption")
		return nil
	}

	return s.kafka.Subscribe(ctx, "organization.events", func(message interface{}) error {
		event, ok := message.(*domain.OrganizationCreatedEvent)
		if !ok {
			s.logger.Printf("Received unexpected message type: %T", message)
			return nil
		}

		if event.EventType != "organization.created" {
			s.logger.Printf("Ignoring event type: %s", event.EventType)
			return nil
		}

		s.logger.Printf("Processing organization created event for org: %s", event.OrgID)
		return s.HandleOrganizationCreatedEvent(ctx, event)
	})
}
