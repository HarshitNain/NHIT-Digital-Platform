package domain

import (
	"time"

	"github.com/google/uuid"
)

// Department represents a department entity
type Department struct {
	ID          uuid.UUID
	OrgID       *uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDepartment creates a new department
func NewDepartment(name, description string, orgID *uuid.UUID) *Department {
	return &Department{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate validates department fields
func (d *Department) Validate() error {
	if d.Name == "" {
		return ErrDepartmentNameRequired
	}
	if len(d.Name) > 255 {
		return ErrDepartmentNameTooLong
	}
	if d.Description == "" {
		return ErrDepartmentDescriptionRequired
	}
	if len(d.Description) > 500 {
		return ErrDepartmentDescriptionTooLong
	}
	return nil
}
