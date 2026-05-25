package department

import (
	"errors"
	customErr "hitalent/internal/errors"
	model "hitalent/internal/model"
	"strings"

	"gorm.io/gorm"
)

type departmentService struct {
	db *gorm.DB
}

func NewDepartmentService(db *gorm.DB) DepartmentService {
	return &departmentService{db: db}
}

type DepartmentService interface {
	CreateDepartment(req model.CreateDepartmentRequest) (*model.Department, error)
	IsDepartmentExist(id uint64) (bool, error)
	UpdateDepartment(id uint64, req model.UpdateDepartmentRequest) (*model.Department, error)
	DeleteDepartment(id uint64, req model.DeleteDepartmentRequest) error
	GetDepartment(id uint64, depth int, includeEmployees bool) (*model.GetDepartmentResponse, error)
}

func (s *departmentService) CreateDepartment(req model.CreateDepartmentRequest) (*model.Department, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 200 {
		return nil, customErr.ErrInvalidDepartmentName
	}

	dept := model.Department{
		Name:     name,
		ParentID: req.ParentID,
	}

	if req.ParentID != nil {
		var dep model.Department
		if err := s.db.First(&dep, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, customErr.ErrDepartmentNotFound
			}

			return nil, err
		}
	}

	if err := s.db.Create(&dept).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, customErr.ErrDepartmentAlreadyExists
		}
		return nil, err
	}

	return &dept, nil
}

func (s *departmentService) IsDepartmentExist(id uint64) (bool, error) {
	dept := model.Department{
		ID: id,
	}

	if err := s.db.First(&dept).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *departmentService) UpdateDepartment(id uint64, req model.UpdateDepartmentRequest) (*model.Department, error) {
	var dep model.Department

	if err := s.db.First(&dep, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.ErrDepartmentNotFound
		}

		return nil, err
	}

	updates := make(map[string]any)

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" || len(name) > 200 {
			return nil, customErr.ErrInvalidDepartmentName
		}

		updates["name"] = name
	}

	if req.ParentIDSet {
		if req.ParentID != nil {
			if *req.ParentID == id {
				return nil, customErr.ErrDepartmentParentSelf
			}

			var parent model.Department
			if err := s.db.First(&parent, *req.ParentID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, customErr.ErrParentDepartmentNotFound
				}

				return nil, err
			}

			isDescendant, err := s.isDescendant(*req.ParentID, id)
			if err != nil {
				return nil, err
			}

			if isDescendant {
				return nil, customErr.ErrDepartmentCycle
			}

			updates["parent_id"] = *req.ParentID
		} else {
			updates["parent_id"] = nil
		}
	}

	if len(updates) == 0 {
		return &dep, nil
	}

	if err := s.db.Model(&dep).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&dep, id).Error; err != nil {
		return nil, err
	}

	return &dep, nil
}

func (s *departmentService) isDescendant(parentID uint64, childID uint64) (bool, error) {
	var count int64

	err := s.db.Raw(`
		WITH RECURSIVE subtree AS (
			SELECT id
			FROM departments
			WHERE parent_id = ?

			UNION ALL

			SELECT d.id
			FROM departments d
			JOIN subtree s ON d.parent_id = s.id
		)
		SELECT COUNT(*)
		FROM subtree
		WHERE id = ?
	`, childID, parentID).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *departmentService) DeleteDepartment(id uint64, req model.DeleteDepartmentRequest) error {
	var dep model.Department

	if err := s.db.First(&dep, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customErr.ErrDepartmentNotFound
		}

		return err
	}

	switch req.Mode {
	case "cascade":
		return s.db.Delete(&dep).Error

	case "reassign":
		if req.ReassignToDepartmentID == nil {
			return customErr.ErrReassignDepartmentRequired
		}

		reassignID := *req.ReassignToDepartmentID

		if reassignID == id {
			return customErr.ErrCannotReassignToSelf
		}

		var target model.Department
		if err := s.db.First(&target, reassignID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return customErr.ErrReassignDepartmentNotFound
			}

			return err
		}

		isChild, err := s.isDescendant(reassignID, id)
		if err != nil {
			return err
		}

		if isChild {
			return customErr.ErrCannotReassignToChild
		}

		return s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.Employee{}).
				Where("department_id = ?", id).
				Update("department_id", reassignID).
				Error; err != nil {
				return err
			}

			if err := tx.Model(&model.Department{}).
				Where("parent_id = ?", id).
				Update("parent_id", reassignID).
				Error; err != nil {
				return err
			}

			if err := tx.Delete(&dep).Error; err != nil {
				return err
			}

			return nil
		})

	default:
		return customErr.ErrInvalidDeleteMode
	}
}

func (s *departmentService) GetDepartment(id uint64, depth int, includeEmployees bool) (*model.GetDepartmentResponse, error) {
	if depth < 0 {
		depth = 0
	}

	if depth > 5 {
		depth = 5
	}

	return s.buildDepartmentTree(id, depth, includeEmployees)
}

func (s *departmentService) buildDepartmentTree(id uint64, depth int, includeEmployees bool) (*model.GetDepartmentResponse, error) {
	var dep model.Department

	if err := s.db.First(&dep, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErr.ErrDepartmentNotFound
		}

		return nil, err
	}

	resp := &model.GetDepartmentResponse{
		Department: dep,
	}

	if includeEmployees {
		var employees []model.Employee

		if err := s.db.
			Where("department_id = ?", id).
			Order("full_name ASC").
			Find(&employees).Error; err != nil {
			return nil, err
		}

		resp.Employees = employees
	}

	if depth <= 0 {
		return resp, nil
	}

	var children []model.Department

	if err := s.db.
		Where("parent_id = ?", id).
		Order("name ASC").
		Find(&children).Error; err != nil {
		return nil, err
	}

	for _, child := range children {
		childResp, err := s.buildDepartmentTree(child.ID, depth-1, includeEmployees)
		if err != nil {
			return nil, err
		}

		resp.Children = append(resp.Children, *childResp)
	}

	return resp, nil
}
