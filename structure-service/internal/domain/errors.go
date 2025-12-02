package domain

import (
	"structure-service/internal/infrastructure/errors"
)

var (
	ErrUniversityNotFound        = errors.NotFoundError("university")
	ErrBranchNotFound            = errors.NotFoundError("branch")
	ErrFacultyNotFound           = errors.NotFoundError("faculty")
	ErrGroupNotFound             = errors.NotFoundError("group")
	ErrInvalidStructure          = errors.ValidationError("invalid structure")
	ErrDuplicateEntry            = errors.AlreadyExistsError("entry", "")
	ErrDepartmentManagerNotFound = errors.NotFoundError("department manager")
	ErrEmployeeNotFound          = errors.NotFoundError("employee")
	ErrInvalidDepartment         = errors.ValidationError("invalid department: must specify branch_id or faculty_id")
	ErrInvalidFile               = errors.ValidationError("invalid file format")
	ErrMissingColumns            = errors.ValidationError("missing required columns")
)

