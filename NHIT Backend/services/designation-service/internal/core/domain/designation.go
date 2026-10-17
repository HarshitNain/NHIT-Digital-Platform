package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// Designation represents a job designation/position in the organization
type Designation struct {
	ID          uuid.UUID
	OrgID       *uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDesignation creates a new designation with validation
func NewDesignation(name, description string, orgID *uuid.UUID) (*Designation, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Validate name
	if err := ValidateDesignationName(name); err != nil {
		return nil, err
	}

	// Validate description
	if err := ValidateDesignationDescription(description); err != nil {
		return nil, err
	}

	now := time.Now()

	return &Designation{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update updates existing designation with validation
func (d *Designation) Update(name, description string) error {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Validate name
	if err := ValidateDesignationName(name); err != nil {
		return err
	}

	// Validate description
	if err := ValidateDesignationDescription(description); err != nil {
		return err
	}

	// Apply updates
	d.Name = name
	d.Description = description
	d.UpdatedAt = time.Now()

	return nil
}

// ValidateDesignationName validates the designation name
func ValidateDesignationName(name string) error {
	if name == "" {
		return ErrDesignationNameRequired
	}

	if len(name) < 2 {
		return ErrDesignationNameTooShort
	}

	if len(name) > 250 {
		return ErrDesignationNameTooLong
	}

	// Check allowed characters
	if !isValidDesignationName(name) {
		return ErrDesignationNameInvalidChars
	}

	// Reserved names not allowed
	if isReservedName(name) {
		return ErrDesignationNameReserved
	}

	return nil
}

// ValidateDesignationDescription validates the designation description
func ValidateDesignationDescription(description string) error {
	if description == "" {
		return ErrDesignationDescriptionRequired
	}

	if len(description) < 5 {
		return ErrDesignationDescriptionTooShort
	}

	if len(description) > 500 {
		return ErrDesignationDescriptionTooLong
	}

	return nil
}

// isValidDesignationName checks allowed characters
func isValidDesignationName(name string) bool {
	for _, r := range name {
		if !unicode.IsLetter(r) &&
			!unicode.IsNumber(r) &&
			r != ' ' &&
			r != '-' &&
			r != '_' &&
			r != '/' &&
			r != '&' &&
			r != '.' &&
			r != '(' &&
			r != ')' {
			return false
		}
	}
	return true
}

// isReservedName checks if name is reserved
func isReservedName(name string) bool {
	reserved := []string{
		"admin", "administrator", "root", "system", "superuser",
		"null", "undefined", "none", "default",
	}

	lowerName := strings.ToLower(name)
	for _, r := range reserved {
		if lowerName == r {
			return true
		}
	}
	return false
}

// NormalizeDesignationName normalizes the name for comparison
func NormalizeDesignationName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
