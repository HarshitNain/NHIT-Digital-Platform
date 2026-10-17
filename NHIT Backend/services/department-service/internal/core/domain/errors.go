package domain

import "errors"

var (
	// Department validation errors
	ErrDepartmentNameRequired         = errors.New("department name is required")
	ErrDepartmentNameTooLong          = errors.New("department name must not exceed 255 characters")
	ErrDepartmentDescriptionRequired  = errors.New("department description is required")
	ErrDepartmentDescriptionTooLong   = errors.New("department description must not exceed 500 characters")
	
	// Department business errors
	ErrDepartmentNotFound             = errors.New("department not found")
	ErrDepartmentAlreadyExists        = errors.New("department with this name already exists")
	ErrDepartmentHasUsers             = errors.New("cannot delete department with assigned users")
	ErrInvalidDepartmentID            = errors.New("invalid department ID")
)
