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

type employeeServiceMock struct {
	mock.Mock
}

func (m *employeeServiceMock) CreateEmployee(req model.CreateEmployeeRequest) (*model.Employee, error) {
	args := m.Called(req)
	emp, _ := args.Get(0).(*model.Employee)
	return emp, args.Error(1)
}

func TestCreateEmployeeSuccess(t *testing.T) {
	service := new(employeeServiceMock)
	handler := NewDepartmentsHandler(service)

	hiredAt := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	service.On("CreateEmployee", model.CreateEmployeeRequest{
		DepartmentID: 7,
		FullName:     "Ada Lovelace",
		Position:     "Engineer",
		HiredAt:      &hiredAt,
	}).Return(&model.Employee{
		ID:           10,
		DepartmentID: 7,
		FullName:     "Ada Lovelace",
		Position:     "Engineer",
		HiredAt:      &hiredAt,
		CreatedAt:    createdAt,
	}, nil).Once()

	body := `{"full_name":"Ada Lovelace","position":"Engineer","hired_at":"2026-05-25T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/departments/7/employees", bytes.NewBufferString(body))
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id":10,"department_id":7,"full_name":"Ada Lovelace","position":"Engineer","hired_at":"2026-05-25T00:00:00Z","created_at":"2026-05-25T12:00:00Z"}`, rec.Body.String())
	service.AssertExpectations(t)
}

func TestCreateEmployeeInvalidDepartmentID(t *testing.T) {
	service := new(employeeServiceMock)
	handler := NewDepartmentsHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/departments/not-a-number/employees", bytes.NewBufferString(`{"full_name":"Ada","position":"Engineer"}`))
	req.SetPathValue("id", "not-a-number")
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	service.AssertNotCalled(t, "CreateEmployee", mock.Anything)
}

func TestCreateEmployeeDepartmentNotFound(t *testing.T) {
	service := new(employeeServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("CreateEmployee", model.CreateEmployeeRequest{
		DepartmentID: 7,
		FullName:     "Ada Lovelace",
		Position:     "Engineer",
	}).Return((*model.Employee)(nil), customErr.ErrDepartmentNotFound).Once()

	req := httptest.NewRequest(http.MethodPost, "/departments/7/employees", bytes.NewBufferString(`{"full_name":"Ada Lovelace","position":"Engineer"}`))
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Department doesn't exists")
	service.AssertExpectations(t)
}

func TestCreateEmployeeValidationError(t *testing.T) {
	service := new(employeeServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("CreateEmployee", model.CreateEmployeeRequest{
		DepartmentID: 7,
		FullName:     "",
		Position:     "Engineer",
	}).Return((*model.Employee)(nil), customErr.ErrInvalidEmployeeFullName).Once()

	req := httptest.NewRequest(http.MethodPost, "/departments/7/employees", bytes.NewBufferString(`{"full_name":"","position":"Engineer"}`))
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), customErr.ErrInvalidEmployeeFullName.Error())
	service.AssertExpectations(t)
}

func TestCreateEmployeeUnexpectedError(t *testing.T) {
	service := new(employeeServiceMock)
	handler := NewDepartmentsHandler(service)

	service.On("CreateEmployee", mock.Anything).
		Return((*model.Employee)(nil), errors.New("db is down")).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/departments/7/employees", bytes.NewBufferString(`{"full_name":"Ada Lovelace","position":"Engineer"}`))
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	service.AssertExpectations(t)
}
