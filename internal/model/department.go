package model

import "time"

type Department struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:200;not null;uniqueIndex:idx_parent_name" json:"name"`
	ParentID  *uint64   `gorm:"index" json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateDepartmentRequest struct {
	Name     string  `json:"name" validate:"required"`
	ParentID *uint64 `json:"parent_id,omitempty"`
}

type DeleteDepartmentRequest struct {
	Mode                   string
	ReassignToDepartmentID *uint64
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name"`
	ParentID    *uint64 `json:"parent_id"`
	ParentIDSet bool    `json:"-"`
}

type GetDepartmentResponse struct {
	Department Department              `json:"department"`
	Employees  []Employee              `json:"employees,omitempty"`
	Children   []GetDepartmentResponse `json:"children,omitempty"`
}
