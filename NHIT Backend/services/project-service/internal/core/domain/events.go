package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationCreatedEvent represents an event when an organization is created
type OrganizationCreatedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	TenantID  string    `json:"tenant_id"`
	OrgID     string    `json:"org_id"`
	OrgName   string    `json:"org_name"`
	Projects  []string  `json:"projects"` // Initial projects list
	CreatedBy string    `json:"created_by"`
}

// NewProjectFromEvent creates projects from organization created event
func NewProjectFromEvent(event *OrganizationCreatedEvent) ([]*Project, error) {
	var projects []*Project

	tenantUUID, err := uuid.Parse(event.TenantID)
	if err != nil {
		return nil, err
	}

	orgUUID, err := uuid.Parse(event.OrgID)
	if err != nil {
		return nil, err
	}

	for _, projectName := range event.Projects {
		if projectName == "" {
			continue // skip empty project names
		}

		project := &Project{
			ProjectID:   uuid.New(),
			TenantID:    tenantUUID,
			OrgID:       orgUUID,
			ProjectName: projectName,
			CreatedBy:   event.CreatedBy,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		projects = append(projects, project)
	}

	return projects, nil
}
