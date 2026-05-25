package api

import (
	"encoding/json"
	"errors"
	"hitalent/internal/employee"
	customErr "hitalent/internal/errors"
	model "hitalent/internal/model"
	"net/http"
	"strconv"
)

type EmployeeHandler struct {
	service employee.EmployeeService
}

func NewEmployeeHandler(service employee.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeRequest

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	depID := r.PathValue("id")

	i, err := strconv.ParseUint(depID, 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	req.DepartmentID = i

	resp, err := h.service.CreateEmployee(req)

	if errors.Is(err, customErr.ErrDepartmentNotFound) {
		http.Error(w, "Department doesn't exists", http.StatusNotFound)
		return
	}
	if errors.Is(err, customErr.ErrInvalidEmployeeFullName) ||
		errors.Is(err, customErr.ErrInvalidEmployeePosition) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
