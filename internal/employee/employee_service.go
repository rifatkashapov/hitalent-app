package employee

import (
	"errors"
	customErr "hitalent/internal/errors"
	model "hitalent/internal/model"
	"strings"

	"gorm.io/gorm"
)

type employeeService struct {
	db *gorm.DB
}

func NewEmployeeService(db *gorm.DB) EmployeeService {
	return &employeeService{db: db}
}

type EmployeeService interface {
	CreateEmployee(req model.CreateEmployeeRequest) (*model.Employee, error)
}

func (s *employeeService) CreateEmployee(req model.CreateEmployeeRequest) (*model.Employee, error) {
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" || len(fullName) > 200 {
		return nil, customErr.ErrInvalidEmployeeFullName
	}

	position := strings.TrimSpace(req.Position)
	if position == "" || len(position) > 200 {
		return nil, customErr.ErrInvalidEmployeePosition
	}

	emp := model.Employee{
		DepartmentID: req.DepartmentID,
		FullName:     fullName,
		Position:     position,
		HiredAt:      req.HiredAt,
	}

	var dep model.Department

	if err := s.db.First(&dep, req.DepartmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.ErrDepartmentNotFound
		}

		return nil, err
	}

	if err := s.db.Create(&emp).Error; err != nil {
		return nil, err
	}

	return &emp, nil
}
