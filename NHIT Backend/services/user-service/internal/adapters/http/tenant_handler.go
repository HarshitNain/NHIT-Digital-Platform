package http

import (
	"context"
	"encoding/json"
	"net/http"
	"log/slog"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TenantHTTPHandler struct {
	userService ports.UserService
	logger      *slog.Logger
}

func NewTenantHTTPHandler(userService ports.UserService, logger *slog.Logger) *TenantHTTPHandler {
	return &TenantHTTPHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateTenantRequest represents the HTTP request for creating a tenant
type CreateTenantRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// TenantResponse represents the HTTP response for tenant operations
type TenantResponse struct {
	TenantID         string `json:"tenant_id"`
	Name             string `json:"name"`
	SuperAdminUserID string `json:"super_admin_user_id,omitempty"`
}

// CreateTenant handles POST /api/v1/users/tenants
func (h *TenantHTTPHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	// Create tenant using the service
	tenant, err := h.userService.CreateTenant(context.Background(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		http.Error(w, "Failed to create tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	response := TenantResponse{
		TenantID:         tenant.TenantID.String(),
		Name:             tenant.Name,
		SuperAdminUserID: "", // Could be populated with the actual super admin user ID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTenant handles GET /api/v1/users/tenants/{tenant_id}
func (h *TenantHTTPHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// Get tenant using the service
	tenant, err := h.userService.GetTenant(context.Background(), tenantID)
	if err != nil {
		http.Error(w, "Tenant not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Return response
	response := TenantResponse{
		TenantID: tenant.TenantID.String(),
		Name:     tenant.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetupRoutes sets up HTTP routes for tenant endpoints
func (h *TenantHTTPHandler) SetupRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/tenants", h.CreateTenant).Methods("POST")
	router.HandleFunc("/api/v1/tenants/{tenant_id}", h.GetTenant).Methods("GET")
}
