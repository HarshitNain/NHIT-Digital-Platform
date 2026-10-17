package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/storage"
	"google.golang.org/grpc/metadata"
)

type OrganizationHandler struct {
	client     pb.OrganizationServiceClient
	storage    *storage.MinIOClient
	bucketName string
}

func NewOrganizationHandler(client pb.OrganizationServiceClient, storage *storage.MinIOClient, bucketName string) *OrganizationHandler {
	return &OrganizationHandler{
		client:     client,
		storage:    storage,
		bucketName: bucketName,
	}
}

// getOrgName helps extract org name for folder creation
func getOrgName(r *http.Request) string {
	if name := r.FormValue("name"); name != "" {
		return name
	}
	return "unknown_org"
}

func (h *OrganizationHandler) HandleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Multipart Form
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Handle Logo Upload if present
	var logoURL string
	file, header, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()

		orgName := sanitizeFilename(getOrgName(r))
		objectName := fmt.Sprintf("%s/%d_%s", orgName, time.Now().Unix(), header.Filename)

		url, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			log.Printf("Failed to upload logo: %v", err)
			http.Error(w, "Failed to upload logo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logoURL = url
	}

	// 3. Prepare gRPC Request
	req := &pb.CreateOrganizationRequest{
		Name:            r.FormValue("name"),
		Code:            r.FormValue("code"),
		Description:     r.FormValue("description"),
		Logo:            logoURL,
		ParentOrgId:     r.FormValue("parent_org_id"),
		InitialProjects: r.PostForm["initial_projects[]"], // Assume array from form
	}

	// Handle Super Admin if provided (JSON string in form field or flattened?)
	// For simplicity, assume flattened fields for now if provided
	if r.FormValue("super_admin.email") != "" {
		req.SuperAdmin = &pb.SuperAdminDetails{
			Name:     r.FormValue("super_admin.name"),
			Email:    r.FormValue("super_admin.email"),
			Password: r.FormValue("super_admin.password"),
		}
	}

	// 4. Forward to Organization Service
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))

	resp, err := h.client.CreateOrganization(ctx, req)
	if err != nil {
		if logoURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(logoURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned logo %s due to gRPC failure", logoURL)
			}
		}
		http.Error(w, "Failed to create organization: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *OrganizationHandler) HandleUploadLogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Path: /api/v1/organizations/{org_id}/logo
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	orgID := parts[4]

	// 1. Parse Multipart
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Handle File
	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "Logo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Upload to MinIO
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
	folderName := orgID // default
	existing, err := h.client.GetOrganization(ctx, &pb.GetOrganizationRequest{OrgId: orgID})
	if err == nil && existing.Organization != nil {
		folderName = sanitizeFilename(existing.Organization.Name)
	}

	objectName := fmt.Sprintf("%s/logo_%d_%s", folderName, time.Now().Unix(), header.Filename)
	logoURL, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Failed to upload to MinIO: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Cleanup old logo if exists
	if existing != nil && existing.Organization != nil && existing.Organization.Logo != "" {
		bucket, obj, pErr := h.storage.ParseBucketAndObject(existing.Organization.Logo)
		if pErr == nil {
			_ = h.storage.DeleteObject(r.Context(), bucket, obj)
		}
	}

	// 5. Update via Service
	updateReq := &pb.UpdateOrganizationRequest{
		OrgId:       orgID,
		Name:        existing.Organization.Name,
		Code:        existing.Organization.Code,
		Description: existing.Organization.Description,
		Logo:        logoURL,
		Status:      existing.Organization.Status,
	}

	resp, err := h.client.UpdateOrganization(ctx, updateReq)
	if err != nil {
		if logoURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(logoURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned logo %s due to gRPC failure", logoURL)
			}
		}
		http.Error(w, "Failed to update organization: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (h *OrganizationHandler) HandleUpdateOrganization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse org_id from path /api/v1/organizations/{org_id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	orgID := parts[4]

	// 2. Parse Multipart Form
	log.Printf("DEBUG: HandleUpdateOrganization - Path: %s, Content-Type: %s", r.URL.Path, r.Header.Get("Content-Type"))
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		log.Printf("DEBUG: Failed to parse multipart form: %v", err)
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Handle Logo Upload if present
	var logoURL string
	file, header, err := r.FormFile("logo")
	
	// Fetch existing once if needed for name or cleanup
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
	existing, getErr := h.client.GetOrganization(ctx, &pb.GetOrganizationRequest{OrgId: orgID})
	if getErr != nil {
		http.Error(w, "Failed to fetch organization: "+getErr.Error(), http.StatusInternalServerError)
		return
	}

	if err == nil {
		defer file.Close()
		log.Printf("DEBUG: Logo file detected: %s (%d bytes)", header.Filename, header.Size)
		
		folderName := sanitizeFilename(existing.Organization.Name)
		if n := r.FormValue("name"); n != "" {
			folderName = sanitizeFilename(n)
		}
		
		objectName := fmt.Sprintf("%s/logo_%d_%s", folderName, time.Now().Unix(), header.Filename)

		url, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			log.Printf("Failed to upload logo: %v", err)
			http.Error(w, "Failed to upload logo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logoURL = url

		// Cleanup old logo if it exists
		if existing.Organization.Logo != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(existing.Organization.Logo)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
			}
		}
	}

	// 4. Update via Service

	updateReq := &pb.UpdateOrganizationRequest{
		OrgId:       orgID,
		Name:        r.FormValue("name"),
		Code:        r.FormValue("code"),
		Description: r.FormValue("description"),
		Logo:        logoURL,
		Status:      existing.Organization.Status, // Keep existing status unless also in form
	}

	// Fallback to existing values if not provided in form
	if updateReq.Name == "" {
		updateReq.Name = existing.Organization.Name
	}
	if updateReq.Code == "" {
		updateReq.Code = existing.Organization.Code
	}
	if updateReq.Description == "" {
		updateReq.Description = existing.Organization.Description
	}
	if updateReq.Logo == "" {
		updateReq.Logo = existing.Organization.Logo
	}
	if formStatus := r.FormValue("status"); formStatus != "" {
		// Map status if needed, for now keep existing or assume it's correct
	}

	resp, err := h.client.UpdateOrganization(ctx, updateReq)
	if err != nil {
		if logoURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(logoURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned logo %s due to gRPC failure", logoURL)
			}
		}
		http.Error(w, "Failed to update organization: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

