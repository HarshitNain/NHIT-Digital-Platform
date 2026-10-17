package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"log/slog"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHTTPHandler struct {
	userService ports.UserService
	logger      *slog.Logger
}

func NewUserHTTPHandler(userService ports.UserService, logger *slog.Logger) *UserHTTPHandler {
	return &UserHTTPHandler{
		userService: userService,
		logger:      logger,
	}
}

type CreateUserRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AssignRolesRequest struct {
	Roles []string `json:"roles"`
}

type UserResponse struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type ListUsersResponse struct {
	Users      []*UserResponse `json:"users"`
	TotalCount int64           `json:"total_count"`
}

type ListRolesResponse struct {
	Roles []*RoleResponse `json:"roles"`
}

type RoleResponse struct {
	RoleID      string   `json:"role_id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func (h *UserHTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/users", h.CreateUser).Methods("POST")
	router.HandleFunc("/api/v1/users", h.ListUsers).Methods("GET")
	router.HandleFunc("/api/v1/users/{user_id}", h.GetUser).Methods("GET")
	router.HandleFunc("/api/v1/users/{user_id}", h.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/v1/users/{user_id}", h.DeleteUser).Methods("DELETE")
	router.HandleFunc("/api/v1/users/{user_id}/roles", h.AssignRolesToUser).Methods("POST")
	router.HandleFunc("/api/v1/users/{user_id}/roles", h.ListRolesOfUser).Methods("GET")
}

func (h *UserHTTPHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
		return
	}

	user := &domain.User{
		TenantID: tenantID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	createdUser, err := h.userService.CreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, toHTTPUser(createdUser, nil, nil))
}

func (h *UserHTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(mux.Vars(r), "user_id")
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUser(r.Context(), userID, nil)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, toHTTPUser(user, nil, nil))
}

func (h *UserHTTPHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(mux.Vars(r), "user_id")
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := &domain.User{
		UserID:   userID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, toHTTPUser(updatedUser, nil, nil))
}

func (h *UserHTTPHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(mux.Vars(r), "user_id")
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID, nil); err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHTTPHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantParam := r.URL.Query().Get("tenant_id")
	if tenantParam == "" {
		http.Error(w, "tenant_id query parameter is required", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantParam)
	if err != nil {
		http.Error(w, "Invalid tenant_id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	users, total, err := h.userService.ListUsersByTenant(r.Context(), tenantID, nil, limit, offset)
	if err != nil {
		http.Error(w, "Failed to list users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*UserResponse, len(users))
	for i, u := range users {
		responses[i] = toHTTPUser(u, nil, nil)
	}

	respondJSON(w, http.StatusOK, &ListUsersResponse{
		Users:      responses,
		TotalCount: total,
	})
}

func (h *UserHTTPHandler) AssignRolesToUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(mux.Vars(r), "user_id")
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var req AssignRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, roleIDStr := range req.Roles {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			http.Error(w, "Invalid role_id: "+roleIDStr, http.StatusBadRequest)
			return
		}

		if err := h.userService.AssignRoleToUser(r.Context(), userID, roleID); err != nil {
			http.Error(w, "Failed to assign role: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	user, err := h.userService.GetUser(r.Context(), userID, nil)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	roles, err := h.userService.GetUserRoles(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, toHTTPUser(user, roles, nil))
}

func (h *UserHTTPHandler) ListRolesOfUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(mux.Vars(r), "user_id")
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	roles, err := h.userService.GetUserRoles(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = &RoleResponse{
			RoleID:      role.RoleID.String(),
			TenantID:    role.TenantID.String(),
			Name:        role.Name,
			Permissions: role.Permissions,
		}
	}

	respondJSON(w, http.StatusOK, &ListRolesResponse{Roles: responses})
}

func parseUUIDParam(vars map[string]string, key string) (uuid.UUID, error) {
	value, ok := vars[key]
	if !ok {
		return uuid.Nil, errors.New("missing path parameter: " + key)
	}
	return uuid.Parse(value)
}

func parsePagination(r *http.Request) (int32, int32) {
	query := r.URL.Query()
	pageSize := int32(10)
	page := int32(1)

	if val := query.Get("page_size"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			pageSize = int32(parsed)
		}
	}

	if val := query.Get("page"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			page = int32(parsed)
		}
	}

	offset := (page - 1) * pageSize
	return pageSize, offset
}

func toHTTPUser(user *domain.User, roles []*domain.Role, permissions []string) *UserResponse {
	res := &UserResponse{
		UserID:   user.UserID.String(),
		TenantID: user.TenantID.String(),
		Name:     user.Name,
		Email:    user.Email,
	}

	if len(roles) > 0 {
		res.Roles = make([]string, len(roles))
		permSet := make(map[string]struct{})
		for i, role := range roles {
			res.Roles[i] = role.Name
			for _, perm := range role.Permissions {
				permSet[perm] = struct{}{}
			}
		}
		res.Permissions = flattenPermissions(permSet)
	} else if len(permissions) > 0 {
		res.Permissions = permissions
	}

	return res
}

func flattenPermissions(perms map[string]struct{}) []string {
	if len(perms) == 0 {
		return nil
	}
	out := make([]string, 0, len(perms))
	for perm := range perms {
		out = append(out, perm)
	}
	return out
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}
