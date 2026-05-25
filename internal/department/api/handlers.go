package api

import (
	"encoding/json"
	"errors"
	"hitalent/internal/department"
	customErr "hitalent/internal/errors"
	model "hitalent/internal/model"
	"net/http"
	"strconv"
)

type DepartmentsHandler struct {
	service department.DepartmentService
}

func NewDepartmentsHandler(service department.DepartmentService) *DepartmentsHandler {
	return &DepartmentsHandler{service: service}
}

func (h *DepartmentsHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDepartmentRequest

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	resp, err := h.service.CreateDepartment(req)

	if errors.Is(err, customErr.ErrDepartmentNotFound) {
		http.Error(w, "Parent department doesn't exist", http.StatusNotFound)
		return
	}

	if errors.Is(err, customErr.ErrDepartmentAlreadyExists) {
		http.Error(w, "Department already exists", http.StatusConflict)
		return
	}
	if errors.Is(err, customErr.ErrInvalidDepartmentName) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Failed to create department", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *DepartmentsHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	depth := 1

	rawDepth := r.URL.Query().Get("depth")
	if rawDepth != "" {
		parsedDepth, err := strconv.Atoi(rawDepth)
		if err != nil {
			http.Error(w, "invalid depth", http.StatusBadRequest)
			return
		}

		depth = parsedDepth
	}

	if depth < 0 {
		http.Error(w, "depth must be >= 0", http.StatusBadRequest)
		return
	}

	if depth > 5 {
		http.Error(w, "depth must be <= 5", http.StatusBadRequest)
		return
	}

	includeEmployees := true

	rawIncludeEmployees := r.URL.Query().Get("include_employees")
	if rawIncludeEmployees != "" {
		parsedIncludeEmployees, err := strconv.ParseBool(rawIncludeEmployees)
		if err != nil {
			http.Error(w, "invalid include_employees", http.StatusBadRequest)
			return
		}

		includeEmployees = parsedIncludeEmployees
	}

	resp, err := h.service.GetDepartment(id, depth, includeEmployees)
	if err != nil {
		switch {
		case errors.Is(err, customErr.ErrDepartmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *DepartmentsHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var req model.UpdateDepartmentRequest

	if rawName, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(rawName, &name); err != nil {
			http.Error(w, "invalid name", http.StatusBadRequest)
			return
		}

		req.Name = &name
	}

	if rawParentID, ok := raw["parent_id"]; ok {
		req.ParentIDSet = true

		if string(rawParentID) == "null" {
			req.ParentID = nil
		} else {
			var parentID uint64
			if err := json.Unmarshal(rawParentID, &parentID); err != nil {
				http.Error(w, "invalid parent_id", http.StatusBadRequest)
				return
			}

			req.ParentID = &parentID
		}
	}

	dep, err := h.service.UpdateDepartment(id, req)

	if err != nil {
		switch {
		case errors.Is(err, customErr.ErrDepartmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, customErr.ErrParentDepartmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, customErr.ErrDepartmentCycle):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, customErr.ErrDepartmentParentSelf):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, customErr.ErrInvalidDepartmentName):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(dep); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *DepartmentsHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid department id", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "cascade"
	}

	req := model.DeleteDepartmentRequest{
		Mode: mode,
	}

	if mode == "reassign" {
		rawReassignID := r.URL.Query().Get("reassign_to_department_id")
		if rawReassignID == "" {
			http.Error(w, "reassign_to_department_id is required", http.StatusBadRequest)
			return
		}

		reassignID, err := strconv.ParseUint(rawReassignID, 10, 64)
		if err != nil {
			http.Error(w, "invalid reassign_to_department_id", http.StatusBadRequest)
			return
		}

		req.ReassignToDepartmentID = &reassignID
	}

	err = h.service.DeleteDepartment(id, req)
	if err != nil {
		switch {
		case errors.Is(err, customErr.ErrDepartmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)

		case errors.Is(err, customErr.ErrInvalidDeleteMode),
			errors.Is(err, customErr.ErrReassignDepartmentRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)

		case errors.Is(err, customErr.ErrReassignDepartmentNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)

		case errors.Is(err, customErr.ErrCannotReassignToSelf),
			errors.Is(err, customErr.ErrCannotReassignToChild):
			http.Error(w, err.Error(), http.StatusConflict)

		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
