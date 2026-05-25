package model

import "time"

type Employee struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	DepartmentID uint64     `gorm:"not null;index" json:"department_id"`
	FullName     string     `gorm:"size:200;not null" json:"full_name"`
	Position     string     `gorm:"size:200;not null" json:"position"`
	HiredAt      *time.Time `json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateEmployeeRequest struct {
	DepartmentID uint64     `json:"department_id" validate:"required"`
	FullName     string     `json:"full_name" validate:"required"`
	Position     string     `json:"position" validate:"required"`
	HiredAt      *time.Time `json:"hired_at,omitempty"`
}
