package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProjectName        = errors.New("project name must be between 3 and 255 characters")
)

// Project represents a project entity in the system
type Project struct {
	ProjectID     uuid.UUID       `json:"project_id" db:"project_id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	OrgID         uuid.UUID       `json:"org_id" db:"org_id"`
	ProjectName   string          `json:"project_name" db:"project_name"`
	CreatedBy     string      `json:"created_by" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}


// NewProject creates a new project with validation
func NewProject(tenantID, orgID uuid.UUID, name, createdBy string) (*Project, error) {
	
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID cannot be empty")
	}
	
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID cannot be empty")
	}
	
	
	project := &Project{
		ProjectID:   uuid.New(),
		TenantID:    tenantID,
		OrgID:       orgID,
		ProjectName: name,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := project.Validate(); err != nil {
		return nil, err
	}
	
	return project, nil
}

// Validate validates the project fields
func (p *Project) Validate() error {
	if len(p.ProjectName) < 3 || len(p.ProjectName) > 255 {
		return ErrInvalidProjectName
	}
	return nil
}
