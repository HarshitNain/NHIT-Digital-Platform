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

// NewOrganizationCreatedEvent creates a new organization created event
func NewOrganizationCreatedEvent(tenantID, orgID, orgName, createdBy string, projects []string) *OrganizationCreatedEvent {
	return &OrganizationCreatedEvent{
		EventID:   uuid.New().String(),
		EventType: "organization.created",
		Timestamp: time.Now().UTC(),
		TenantID:  tenantID,
		OrgID:     orgID,
		OrgName:   orgName,
		Projects:  projects,
		CreatedBy: createdBy,
	}
}
