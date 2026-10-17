package domain

import "errors"

// Designation validation errors
var (
	// Name validation errors
	ErrDesignationNameRequired      = errors.New("designation name is required")
	ErrDesignationNameTooShort      = errors.New("designation name must be at least 2 characters")
	ErrDesignationNameTooLong       = errors.New("designation name must not exceed 250 characters")
	ErrDesignationNameInvalidChars  = errors.New("designation name contains invalid characters")
	ErrDesignationNameReserved      = errors.New("designation name is reserved and cannot be used")

	// Description validation errors
	ErrDesignationDescriptionRequired = errors.New("designation description is required")
	ErrDesignationDescriptionTooShort = errors.New("designation description must be at least 5 characters")
	ErrDesignationDescriptionTooLong  = errors.New("designation description must not exceed 500 characters")

	// Business logic errors
	ErrDesignationNotFound      = errors.New("designation not found")
	ErrDesignationAlreadyExists = errors.New("designation with this name already exists")
	ErrDesignationHasUsers      = errors.New("cannot delete designation: users are assigned to this designation")
	ErrCircularReference        = errors.New("designation cannot be its own parent")
	ErrInvalidParent            = errors.New("invalid parent designation")
	ErrParentNotFound           = errors.New("parent designation not found")
	ErrMaxHierarchyDepth        = errors.New("maximum hierarchy depth exceeded (max 5 levels)")
	ErrCannotDeactivateWithUsers = errors.New("cannot deactivate designation with assigned users")
	ErrSlugAlreadyExists        = errors.New("designation with this slug already exists")
)
