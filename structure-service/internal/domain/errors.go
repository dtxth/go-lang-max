package domain

import "errors"

var (
	ErrUniversityNotFound        = errors.New("university not found")
	ErrBranchNotFound            = errors.New("branch not found")
	ErrFacultyNotFound           = errors.New("faculty not found")
	ErrGroupNotFound             = errors.New("group not found")
	ErrInvalidStructure          = errors.New("invalid structure")
	ErrDuplicateEntry            = errors.New("duplicate entry")
	ErrDepartmentManagerNotFound = errors.New("department manager not found")
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrInvalidDepartment         = errors.New("invalid department: must specify branch_id or faculty_id")
)

