package errors

import "errors"

var (
	ErrDepartmentNotFound         = errors.New("department not found")
	ErrDepartmentAlreadyExists    = errors.New("department already exists")
	ErrParentDepartmentNotFound   = errors.New("parent department not found")
	ErrInvalidDepartmentName      = errors.New("invalid department name")
	ErrDepartmentParentSelf       = errors.New("department cannot be parent of itself")
	ErrDepartmentCycle            = errors.New("department cycle detected")
	ErrInvalidDeleteMode          = errors.New("invalid delete mode")
	ErrReassignDepartmentRequired = errors.New("reassign_to_department_id is required")
	ErrReassignDepartmentNotFound = errors.New("reassign department not found")
	ErrCannotReassignToSelf       = errors.New("cannot reassign employees to deleted department")
	ErrCannotReassignToChild      = errors.New("cannot reassign employees to child department")
)
