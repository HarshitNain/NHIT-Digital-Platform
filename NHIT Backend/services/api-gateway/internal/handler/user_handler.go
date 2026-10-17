package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/storage"
	"google.golang.org/grpc/metadata"
)

type UserHandler struct {
	client     userpb.UserManagementClient
	storage    *storage.MinIOClient
	bucketName string
}

func NewUserHandler(client userpb.UserManagementClient, storage *storage.MinIOClient, bucketName string) *UserHandler {
	return &UserHandler{
		client:     client,
		storage:    storage,
		bucketName: bucketName,
	}
}

// getUserName helps extract user name for folder creation
func getUserName(r *http.Request) string {
	if name := r.FormValue("name"); name != "" {
		return name
	}
	return "unknown_user"
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Multipart Form
	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Handle Signature Upload if present
	var signatureURL string
	file, header, err := r.FormFile("signature")
	if err == nil {
		defer file.Close()

		// Construct MinIO path: UserName/timestamp_filename
		userName := sanitizeFilename(getUserName(r))
		objectName := fmt.Sprintf("%s/%d_%s", userName, time.Now().Unix(), header.Filename)

		// Upload to MinIO
		url, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
		if err != nil {
			log.Printf("Failed to upload signature to MinIO: %v", err)
			// We can continue or fail. Let's fail for consistency.
			http.Error(w, "Failed to upload signature: "+err.Error(), http.StatusInternalServerError)
			return
		}
		signatureURL = url
	}

	// 3. Prepare gRPC Request
	req := &userpb.CreateUserRequest{
		TenantId:            r.FormValue("tenant_id"),
		OrgId:               r.FormValue("org_id"),
		Email:               r.FormValue("email"),
		Name:                r.FormValue("name"),
		Password:            r.FormValue("password"),
		RoleId:              r.FormValue("role_id"),
		DepartmentId:        r.FormValue("department_id"),
		DesignationId:       r.FormValue("designation_id"),
		CreatedBy:           r.FormValue("created_by"),
		AccountHolderName:   r.FormValue("account_holder_name"),
		BankName:            r.FormValue("bank_name"),
		BankAccountNumber:   r.FormValue("bank_account_number"),
		IfscCode:            r.FormValue("ifsc_code"),
	}

	// 4. Forward to User Service
	// Copy metadata (Authorization token etc.)
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))

	resp, err := h.client.CreateUser(ctx, req)
	if err != nil {
		if signatureURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(signatureURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned signature %s due to gRPC failure", signatureURL)
			}
		}
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. If signature was uploaded, update the user with the URL
	if signatureURL != "" {
		_, uploadErr := h.client.UploadUserSignature(ctx, &userpb.UploadSignatureRequest{
			UserId:       resp.UserId,
			SignatureUrl: signatureURL,
		})
		if uploadErr != nil {
			log.Printf("Warning: Failed to update user signature URL: %v", uploadErr)
			// User is created, so we don't return error here, but maybe return partial success info
		} else {
			// Update the response with the signature URL if possible (though UserResponse doesn't have it, User object does)
			// Since CreateUser returns UserResponse, and it doesn't have signature_url, we just proceed.
		}
	}

	// 6. Return Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) HandleUploadSignature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user_id from path /api/v1/users/{user_id}/signature
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	userID := parts[4]

	// 1. Parse Multipart Form
	log.Printf("📥 Signature upload request - User ID: %s", userID)
	err := r.ParseMultipartForm(5 << 20) // 5MB
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Handle File
	file, header, err := r.FormFile("signature")
	if err != nil {
		http.Error(w, "Signature file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Validate signature image
	log.Printf("🔍 Starting signature validation for file: %s (size: %d bytes)", header.Filename, header.Size)
	validationConfig := DefaultSignatureConfig()
	if err := ValidateSignatureImage(file, header, validationConfig); err != nil {
		http.Error(w, "Signature validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Reset file pointer after validation
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// 4. Upload to MinIO
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
	folderName := userID // fallback
	existingUser, err := h.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if err == nil && existingUser != nil {
		folderName = sanitizeFilename(existingUser.Name)
	}

	objectName := fmt.Sprintf("%s/%d_%s", folderName, time.Now().Unix(), header.Filename)
	signatureURL, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Failed to upload to MinIO: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Cleanup old signature if exists
	if existingUser != nil && existingUser.SignatureUrl != "" {
		bucket, obj, pErr := h.storage.ParseBucketAndObject(existingUser.SignatureUrl)
		if pErr == nil {
			_ = h.storage.DeleteObject(r.Context(), bucket, obj)
		}
	}

	// 4. Call gRPC to update URL in DB
	// ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r)) // Removed re-declaration

	resp, err := h.client.UploadUserSignature(ctx, &userpb.UploadSignatureRequest{
		UserId:       userID,
		SignatureUrl: signatureURL,
		Filename:     header.Filename,
	})
	if err != nil {
		if signatureURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(signatureURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned signature %s due to gRPC failure", signatureURL)
			}
		}
		http.Error(w, "Failed to update user-service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse user_id from path /api/v1/users/{user_id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	userID := parts[4]

	// 2. Parse Multipart
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Handle Signature
	var signatureURL string
	file, header, err := r.FormFile("signature")
	
	ctx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
	existingUser, getErr := h.client.GetUser(ctx, &userpb.GetUserRequest{UserId: userID})
	if getErr != nil {
		http.Error(w, "Failed to fetch user: "+getErr.Error(), http.StatusInternalServerError)
		return
	}

	if err == nil {
		defer file.Close()
		folderName := sanitizeFilename(existingUser.Name)
		if n := r.FormValue("name"); n != "" {
			folderName = sanitizeFilename(n)
		}
		
		// Validate signature image
		validationConfig := DefaultSignatureConfig()
		if err := ValidateSignatureImage(file, header, validationConfig); err != nil {
			http.Error(w, "Signature validation failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Reset file pointer after validation
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, 0)
		}
		
		objectName := fmt.Sprintf("%s/%d_%s", folderName, time.Now().Unix(), header.Filename)
		url, err := h.storage.UploadStream(r.Context(), h.bucketName, objectName, file, header.Size, header.Header.Get("Content-Type"))
		if err == nil {
			signatureURL = url
			
			// Cleanup old signature
			if existingUser.SignatureUrl != "" {
				bucket, obj, pErr := h.storage.ParseBucketAndObject(existingUser.SignatureUrl)
				if pErr == nil {
					_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				}
			}
		}
	}

	// 4. Update User info
	req := &userpb.UpdateUserRequest{
		UserId:        userID,
		Name:          r.FormValue("name"),
		Email:         r.FormValue("email"),
		Password:      r.FormValue("password"),
		DepartmentId:  r.FormValue("department_id"),
		DesignationId: r.FormValue("designation_id"),
		IsActive:      r.FormValue("is_active") == "true",
	}
	if r.FormValue("roles") != "" {
		req.Roles = strings.Split(r.FormValue("roles"), ",")
	}

	updateCtx := metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
	updateResp, updateErr := h.client.UpdateUser(updateCtx, req)
	if updateErr != nil {
		if signatureURL != "" {
			bucket, obj, pErr := h.storage.ParseBucketAndObject(signatureURL)
			if pErr == nil {
				_ = h.storage.DeleteObject(r.Context(), bucket, obj)
				log.Printf("Compensating Transaction: Deleted orphaned signature %s due to gRPC failure", signatureURL)
			}
		}
		http.Error(w, "Failed to update user: "+updateErr.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Update Signature if uploaded
	if signatureURL != "" {
		_, uploadErr := h.client.UploadUserSignature(updateCtx, &userpb.UploadSignatureRequest{
			UserId:       userID,
			SignatureUrl: signatureURL,
		})
		if uploadErr != nil {
			log.Printf("Warning: Failed to update user signature URL: %v", uploadErr)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updateResp)
}

// sanitizeFilename is expected to be available in the package (defined in vendor_handler.go or organization_handler.go)
// If you get redeclared error, remove from one of the files.
// Moving to a shared file is better.
