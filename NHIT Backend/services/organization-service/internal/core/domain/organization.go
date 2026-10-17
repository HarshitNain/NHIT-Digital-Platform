package domain

import (
	"time"
)

type OrganizationStatus int32

const (
	StatusActivated   OrganizationStatus = 0
	StatusDeactivated OrganizationStatus = 1
)

type SuperAdmin struct {
	Name     string
	Email    string
	Password string
}

type Organization struct {
	OrgID           string
	TenantID        string
	ParentOrgID     *string
	Name            string
	Code            string
	DatabaseName    string
	Description     *string
	Logo            *string
	SuperAdmin      *SuperAdmin
	InitialProjects []string
	Status          OrganizationStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
