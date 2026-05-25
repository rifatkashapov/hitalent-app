package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	customErr "hitalent/internal/errors"
	"hitalent/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type departmentServiceMock struct {
	mock.Mock
}

func (m *departmentServiceMock) CreateDepartment(req model.CreateDepartmentRequest) (*model.Department, error) {
	args := m.Called(req)
	dep, _ := args.Get(0).(*model.Department)
	return dep, args.Error(1)
}

func (m *departmentServiceMock) IsDepartmentExist(id uint64) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *departmentServiceMock) UpdateDepartment(id uint64, req model.UpdateDepartmentRequest) (*model.Department, error) {
	args := m.Called(id, req)
	dep, _ := args.Get(0).(*model.Department)
	return dep, args.Error(1)
}

func (m *departmentServiceMock) DeleteDepartment(id uint64, req model.DeleteDepartmentRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func (m *departmentServiceMock) GetDepartment(id uint64, depth int, includeEmployees bool) (*model.GetDepartmentResponse, error) {
	args := m.Called(id, depth, includeEmployees)
	resp, _ := args.Get(0).(*model.GetDepartmentResponse)
	return resp, args.Error(1)
}

func TestCreateDepartmentSuccess(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	createdAt := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	service.On("CreateDepartment", model.CreateDepartmentRequest{Name: "Engineering"}).
		Return(&model.Department{ID: 1, Name: "Engineering", CreatedAt: createdAt}, nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(`{"name":"Engineering"}`))
	rec := httptest.NewRecorder()

	handler.CreateDepartment(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id":1,"name":"Engineering","parent_id":null,"created_at":"2026-05-25T12:00:00Z"}`, rec.Body.String())
	service.AssertExpectations(t)
}

func TestCreateDepartmentParentNotFound(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	parentID := uint64(42)
	service.On("CreateDepartment", model.CreateDepartmentRequest{Name: "Backend", ParentID: &parentID}).
		Return((*model.Department)(nil), customErr.ErrDepartmentNotFound).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(`{"name":"Backend","parent_id":42}`))
	rec := httptest.NewRecorder()

	handler.CreateDepartment(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Parent department doesn't exist")
	service.AssertExpectations(t)
}

func TestCreateDepartmentInvalidJSON(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(`{"name":`))
	rec := httptest.NewRecorder()

	handler.CreateDepartment(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	service.AssertNotCalled(t, "CreateDepartment", mock.Anything)
}

func TestCreateDepartmentInvalidName(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("CreateDepartment", model.CreateDepartmentRequest{Name: ""}).
		Return((*model.Department)(nil), customErr.ErrInvalidDepartmentName).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewBufferString(`{"name":""}`))
	rec := httptest.NewRecorder()

	handler.CreateDepartment(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), customErr.ErrInvalidDepartmentName.Error())
	service.AssertExpectations(t)
}

func TestGetDepartmentParsesQuery(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	resp := &model.GetDepartmentResponse{
		Department: model.Department{ID: 7, Name: "Engineering"},
	}
	service.On("GetDepartment", uint64(7), 2, false).Return(resp, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/departments/7?depth=2&include_employees=false", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.GetDepartment(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"department":{"id":7,"name":"Engineering","parent_id":null,"created_at":"0001-01-01T00:00:00Z"}}`, rec.Body.String())
	service.AssertExpectations(t)
}

func TestGetDepartmentRejectsDepthAboveLimit(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/departments/7?depth=6", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.GetDepartment(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "depth must be <= 5")
	service.AssertNotCalled(t, "GetDepartment", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetDepartmentNotFound(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("GetDepartment", uint64(404), 1, true).
		Return((*model.GetDepartmentResponse)(nil), customErr.ErrDepartmentNotFound).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/departments/404", nil)
	req.SetPathValue("id", "404")
	rec := httptest.NewRecorder()

	handler.GetDepartment(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	service.AssertExpectations(t)
}

func TestUpdateDepartmentParsesNullableParent(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("UpdateDepartment", uint64(3), mock.MatchedBy(func(req model.UpdateDepartmentRequest) bool {
		return req.Name != nil && *req.Name == "Platform" && req.ParentIDSet && req.ParentID == nil
	})).Return(&model.Department{ID: 3, Name: "Platform"}, nil).Once()

	req := httptest.NewRequest(http.MethodPatch, "/departments/3/update", bytes.NewBufferString(`{"name":"Platform","parent_id":null}`))
	req.SetPathValue("id", "3")
	rec := httptest.NewRecorder()

	handler.UpdateDepartment(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	service.AssertExpectations(t)
}

func TestUpdateDepartmentMapsValidationError(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("UpdateDepartment", uint64(3), mock.Anything).
		Return((*model.Department)(nil), customErr.ErrInvalidDepartmentName).
		Once()

	req := httptest.NewRequest(http.MethodPatch, "/departments/3/update", bytes.NewBufferString(`{"name":""}`))
	req.SetPathValue("id", "3")
	rec := httptest.NewRecorder()

	handler.UpdateDepartment(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	service.AssertExpectations(t)
}

func TestDeleteDepartmentReassign(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	reassignID := uint64(9)
	service.On("DeleteDepartment", uint64(3), model.DeleteDepartmentRequest{
		Mode:                   "reassign",
		ReassignToDepartmentID: &reassignID,
	}).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/departments/3/delete?mode=reassign&reassign_to_department_id=9", nil)
	req.SetPathValue("id", "3")
	rec := httptest.NewRecorder()

	handler.DeleteDepartment(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	service.AssertExpectations(t)
}

func TestDeleteDepartmentMapsUnexpectedError(t *testing.T) {
	service := new(departmentServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("DeleteDepartment", uint64(3), model.DeleteDepartmentRequest{Mode: "cascade"}).
		Return(errors.New("db is down")).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/departments/3/delete", nil)
	req.SetPathValue("id", "3")
	rec := httptest.NewRecorder()

	handler.DeleteDepartment(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	service.AssertExpectations(t)
}
