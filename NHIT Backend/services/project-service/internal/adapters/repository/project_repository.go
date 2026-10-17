package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type projectRepository struct {
	db *pgxpool.Pool
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *pgxpool.Pool) ports.ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) (*domain.Project, error) {
	// Validate UUIDs are not nil
	if project.ProjectID == uuid.Nil {
		return nil, fmt.Errorf("project ID is nil")
	}
	if project.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant ID is nil")
	}
	if project.OrgID == uuid.Nil {
		return nil, fmt.Errorf("org ID is nil")
	}
	
	// Debug logging
	fmt.Printf("DEBUG REPO: Creating project - ID=%s, TenantID=%s, OrgID=%s, Name=%s, CreatedBy=%s\n",
		project.ProjectID.String(), project.TenantID.String(), project.OrgID.String(),
		project.ProjectName, project.CreatedBy)
	
	// Handle created_by - pass as string directly (database column should be VARCHAR)
	// Default to "superadmin" if empty
	createdByValue := project.CreatedBy
	if createdByValue == "" {
		createdByValue = "superadmin"
	}
	fmt.Printf("DEBUG REPO: Using created_by value: '%s'\n", createdByValue)
	
	query := `
		INSERT INTO projects (
			id, tenant_id, org_id, project_name, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	
	var returnedID uuid.UUID
	err := r.db.QueryRow(ctx, query,
		project.ProjectID.String(), project.TenantID.String(), project.OrgID.String(),
		project.ProjectName, createdByValue, project.CreatedAt, project.UpdatedAt,
	).Scan(&returnedID, &project.CreatedAt, &project.UpdatedAt)
	
	if err != nil {
		fmt.Printf("DEBUG REPO: Error creating project: %v\n", err)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	project.ProjectID = returnedID
	fmt.Printf("DEBUG REPO: Project created successfully - ID=%s\n", project.ProjectID.String())
	return project, nil
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, projectID uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, tenant_id, org_id, project_name, created_by,
			created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	
	project := &domain.Project{}
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&project.ProjectID, &project.TenantID, &project.OrgID, &project.ProjectName, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return project, nil
}

// ListByOrganization lists all projects for an organization with pagination
func (r *projectRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.Project, int, error) {
	// Get total count first
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM projects WHERE org_id = $1`
	err := r.db.QueryRow(ctx, countQuery, orgID.String()).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get project count: %w", err)
	}

	query := `
		SELECT id, tenant_id, org_id, project_name, created_by,
			created_at, updated_at
		FROM projects
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, orgID.String(), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		err := rows.Scan(
			&p.ProjectID, &p.TenantID, &p.OrgID, &p.ProjectName, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	
	return projects, totalCount, nil
}