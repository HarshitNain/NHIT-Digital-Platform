package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"log"
	"path/filepath"


	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	approvalpb "github.com/ShristiRnr/Nhit-Note/api/pb/approvalpb"

	"github.com/google/uuid"

	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GreenNoteService encapsulates business rules around GreenNotes.
type GreenNoteService struct {
	repo           ports.GreenNoteRepository
	events         ports.EventPublisher
	clock          func() time.Time
	projectClient  projectpb.ProjectServiceClient
	vendorClient   vendorpb.VendorServiceClient
	deptClient     departmentpb.DepartmentServiceClient
	approvalClient approvalpb.ApprovalServiceClient
	storage        ports.DocumentStorage
}

const (
	statusPending  = "pending"
	statusApproved = "approved"
	statusRejected = "rejected"
	statusDraft    = "draft"
	statusCancelled = "cancelled"
)

// NewGreenNoteService constructs a GreenNoteService from its required ports.
func NewGreenNoteService(repo ports.GreenNoteRepository, events ports.EventPublisher, projectClient projectpb.ProjectServiceClient, vendorClient vendorpb.VendorServiceClient, deptClient departmentpb.DepartmentServiceClient, approvalClient approvalpb.ApprovalServiceClient, storage ports.DocumentStorage) *GreenNoteService {
	if events == nil {
		// Guard with internal no-op publisher when none is provided.
		return &GreenNoteService{
			repo:           repo,
			events:         noopEventPublisher{},
			clock:          time.Now,
			projectClient:  projectClient,
			vendorClient:   vendorClient,
			deptClient:     deptClient,
			approvalClient: approvalClient,
			storage:        storage,
		}
	}
	return &GreenNoteService{
		repo:           repo,
		events:         events,
		clock:          time.Now,
		projectClient:  projectClient,
		vendorClient:   vendorClient,
		deptClient:     deptClient,
		approvalClient: approvalClient,
		storage:        storage,
	}
}

// UserContext represents authenticated user information
type UserContext struct {
	UserID   string
	TenantID string
	OrgID    string
	Email    string
	Name     string
	Roles    []string
	UserType string // "USER" or "VENDOR"
}

// validateJWT extracts user context from metadata (simplified approach)
func (s *GreenNoteService) validateJWT(ctx context.Context) (*UserContext, error) {
	// Extract authorization header from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Check authorization header
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// For now, we'll assume the JWT is validated by API Gateway
	// and extract user context from metadata set by gateway
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user_id in context")
	}

	tenantIDs := md.Get("tenant_id")
	if len(tenantIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing tenant_id in context")
	}

	orgIDs := md.Get("org_id")
	if len(orgIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing org_id in context")
	}

	emails := md.Get("email")
	names := md.Get("name")
	roles := md.Get("roles")

	userType := "USER"
	for _, role := range roles {
		if role == "VENDOR" {
			userType = "VENDOR"
			break
		}
	}

	return &UserContext{
		UserID:   userIDs[0],
		TenantID: tenantIDs[0],
		OrgID:    orgIDs[0],
		Email:    getStringValue(emails, ""),
		Name:     getStringValue(names, ""),
		Roles:    roles,
		UserType: userType,
	}, nil
}

// ensureOutgoingContext propagates incoming metadata to outgoing context
func (s *GreenNoteService) ensureOutgoingContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// getStringValue safely returns string from slice or default
func getStringValue(slice []string, defaultValue string) string {
	if len(slice) > 0 {
		return slice[0]
	}
	return defaultValue
}

// ListGreenNotes applies default filters and delegates to the repository.
func (s *GreenNoteService) ListGreenNotes(ctx context.Context, req *greennotepb.ListGreenNotesRequest) (*greennotepb.ListGreenNotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Log user context for debugging
	fmt.Println("\n========== ListGreenNotes SERVICE LAYER DEBUG ==========")
	fmt.Printf("🔐 Authenticated User:\n")
	fmt.Printf("  - User ID: %s\n", userCtx.UserID)
	fmt.Printf("  - User Type: %s\n", userCtx.UserType)
	fmt.Printf("  - Tenant ID: %s\n", userCtx.TenantID)
	fmt.Printf("  - Org ID: %s\n", userCtx.OrgID)
	fmt.Printf("  - Email: %s\n", userCtx.Email)
	fmt.Printf("  - Roles: %v\n", userCtx.Roles)
	fmt.Printf("\n📋 Request Details:\n")
	fmt.Printf("  - Status Filter: %v\n", req.GetStatus())
	fmt.Printf("  - Include All: %v\n", req.GetIncludeAll())
	fmt.Printf("  - Page: %d\n", req.GetPage())
	fmt.Printf("  - PerPage: %d\n", req.GetPerPage())
	fmt.Printf("========== CALLING REPOSITORY ==========\n")

	// TODO: Add business logic based on user type
	// For example: VENDORS can only see their own notes, USERS can see notes within their org

	return s.repo.List(ctx, req, userCtx.OrgID, userCtx.TenantID)
}

func (s *GreenNoteService) GetGreenNote(ctx context.Context, req *greennotepb.GetGreenNoteRequest) (*greennotepb.GreenNoteDetailResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	payload, orgID, tenantID, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}

	currentStatus := normalizeStatus(payload.GetDetailedStatus())
	log.Printf("📖 GetGreenNote called for ID: %s, Current Status: '%s' (normalized: '%s')", 
		req.GetId(), payload.GetDetailedStatus(), currentStatus)

	// If the note is in draft status, change it to pending as per user request (view transitions to pending)
	if currentStatus == statusDraft {
		log.Printf("🔄 Status is DRAFT, transitioning to PENDING for %s", req.GetId())
		payload.DetailedStatus = statusPending
		// Update status in DB
		if err := s.repo.Update(ctx, req.GetId(), payload, orgID, tenantID); err != nil {
			log.Printf("❌ Failed to update status from draft to pending for %s on view: %v", req.GetId(), err)
		} else {
			log.Printf("✅ Status transitioned from draft to pending for GreenNote %s", req.GetId())
			// Trigger approval flow if it transitions to pending
			userCtx, err := s.validateJWT(ctx)
			if err == nil {
				log.Printf("👤 User context validated, initiating approval for %s", req.GetId())
				s.initiateApproval(ctx, req.GetId(), payload, orgID, tenantID, userCtx.UserID)
			} else {
				log.Printf("⚠️ Failed to validate user context, skipping approval initiation: %v", err)
			}
		}
	} else {
		log.Printf("ℹ️  Status is '%s' (not draft), skipping approval initiation", currentStatus)
	}

	return &greennotepb.GreenNoteDetailResponse{
		Success: true,
		Message: "ok",
		Data:    payload,
	}, nil
}

func (s *GreenNoteService) CreateGreenNote(ctx context.Context, req *greennotepb.CreateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || req.Note == nil {
		return nil, status.Error(codes.InvalidArgument, "note payload is required")
	}

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Log user context for debugging
	fmt.Printf("🔐 CreateGreenNote - User ID: %s, Type: %s, Tenant: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.TenantID, userCtx.OrgID)

	// TODO: Add business logic based on user type
	// For example: VENDORS can only create notes for themselves, USERS can create notes for their org

	note := req.Note

	// Resolve SupplierName if it's a UUID (Vendor ID)
	fmt.Printf("🔍 CreateGreenNote - Checking SupplierName: '%s'\n", note.SupplierName)
	if note.SupplierName != "" {
		if _, err := uuid.Parse(note.SupplierName); err == nil {
			// It's a UUID, try to fetch vendor name
			fmt.Printf("🔍 It is a UUID! Resolving Vendor ID to Name: %s\n", note.SupplierName)
			
			// Ensure we propagate auth context
			outCtx := s.ensureOutgoingContext(ctx)
			vendorResp, err := s.vendorClient.GetVendor(outCtx, &vendorpb.GetVendorRequest{VendorId: note.SupplierName})
			if err == nil && vendorResp != nil && vendorResp.Vendor != nil {
				fmt.Printf("✅ Resolved Vendor Name: %s\n", vendorResp.Vendor.VendorName)
				note.SupplierName = vendorResp.Vendor.VendorName
			} else {
				fmt.Printf("⚠️ Failed to resolve vendor: %v\n", err)
			}
		}
	}

	applyDerivedFields(note)
	normalizeStatusOnCreate(note)

	if err := validateGreenNotePayload(note, true); err != nil {
		return nil, err
	}
	id, err := s.repo.Create(ctx, note, userCtx.OrgID, userCtx.TenantID)
	if err != nil {
		return nil, err
	}

	// Process categorized documents if present in the payload
	if err := s.processCategorizedDocuments(ctx, id, note, userCtx.OrgID, userCtx.TenantID); err != nil {
		log.Printf("⚠️ Failed to process categorized documents for %s: %v", id, err)
		// We don't fail the whole creation for this, but log it.
	}

	// If the note is a draft, we skip approval initiation
	if note.DetailedStatus == statusDraft {
		return &greennotepb.GreenNoteResponse{
			Success: true,
			Message: "green note created as draft",
			Id:      id,
		}, nil
	}

	// Call Approval Service to initiate approval flow
	s.initiateApproval(ctx, id, note, userCtx.OrgID, userCtx.TenantID, userCtx.UserID)

	s.publishApprovedIfNeeded(ctx, id, nil, note)
	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note created and approval initiated",
		Id:      id,
	}, nil
}

func (s *GreenNoteService) initiateApproval(ctx context.Context, id string, note *greennotepb.GreenNotePayload, orgID string, tenantID string, userID string) {
	if s.approvalClient == nil {
		log.Printf("⚠️ Approval client is nil, cannot initiate approval for %s", id)
		return
	}

	log.Printf("🚀 Initiating approval for GreenNote %s", id)

	// Ensure Auth Token is propagated
	ctx = s.ensureOutgoingContext(ctx)

	// Lookup IDs from Names
	// 1. Department ID
	var departmentID string
	deptResp, err := s.deptClient.ListDepartments(ctx, &departmentpb.ListDepartmentsRequest{
		Page:     1,
		PageSize: 1000,
	})
	if err == nil {
		for _, d := range deptResp.Departments {
			if strings.EqualFold(strings.TrimSpace(d.Name), strings.TrimSpace(note.DepartmentName)) {
				departmentID = d.Id
				log.Printf("✅ Resolved Department '%s' to ID: %s", note.DepartmentName, departmentID)
				break
			}
		}
	} else {
		log.Printf("⚠️ Failed to fetch departments: %v", err)
	}

	// 2. Project ID
	var projectID string
	projResp, err := s.projectClient.ListProjectsByOrganization(ctx, &projectpb.ListProjectsByOrganizationRequest{
		OrgId: orgID,
	})
	if err == nil {
		for _, p := range projResp.Projects {
			if strings.EqualFold(strings.TrimSpace(p.ProjectName), strings.TrimSpace(note.ProjectName)) {
				projectID = p.ProjectId
				log.Printf("✅ Resolved Project '%s' to ID: %s", note.ProjectName, projectID)
				break
			}
		}
	} else {
		log.Printf("⚠️ Failed to fetch projects: %v", err)
	}

	if departmentID == "" || projectID == "" {
		log.Printf("❌ Cannot initiate approval for %s: dept='%s' (ID: %s), project='%s' (ID: %s)", 
			id, note.DepartmentName, departmentID, note.ProjectName, projectID)
		return
	}

	// 3. Initiate Approval via gRPC
	approvalReq := &approvalpb.InitiateApprovalRequest{
		SlNo:            id,
		Amount:          note.TotalAmount,
		DepartmentId:    departmentID,
		ProjectId:       projectID,
		CreatedByUserId: userID,
		SourceType:      "GREENNOTE",
	}

	log.Printf("📤 Sending approval request to approval service: SlNo=%s, Amount=%.2f, Dept=%s, Project=%s, User=%s", 
		id, note.TotalAmount, departmentID, projectID, userID)

	approvalResp, err := s.approvalClient.InitiateApproval(ctx, approvalReq)
	if err != nil {
		log.Printf("❌ Failed to initiate approval for %s: %v", id, err)
		return
	}

	log.Printf("✅ Approval initiated successfully for %s, Response Status: %s", id, approvalResp.Status)

	// Determine final status and check for reviewer name
	finalStatus := statusPending
	nextReviewer := ""
	if approvalResp.Status != "" {
		normalized := normalizeStatus(approvalResp.Status)
		
		validStatuses := map[string]bool{
			statusPending:   true,
			statusApproved:  true,
			statusRejected:  true,
			statusDraft:     true,
			statusCancelled: true,
		}

		if validStatuses[normalized] {
			finalStatus = normalized
		} else if strings.Contains(normalized, "-") && len(normalized) > 30 {
			// It's a UUID, default to pending
			finalStatus = statusPending
		} else {
			// Likely a reviewer name
			nextReviewer = approvalResp.Status
			finalStatus = statusPending
			log.Printf("👤 Detected Next Reviewer: %s", nextReviewer)
		}
	}

	// Standardized log entry format
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	actorDisplay := userID
	if userCtx, err := s.validateJWT(ctx); err == nil && userCtx.Name != "" {
		actorDisplay = fmt.Sprintf("%s (%s)", userCtx.Name, userID)
	}

	// Build log entry with clear separation of submitter and reviewer
	logEntry := fmt.Sprintf("\n--- [Approval Log | %s] ---\nEvent: Approval Initiated\nAction: SUBMITTED\nSubmitted By: %s",
		timestamp, actorDisplay)
	
	if nextReviewer != "" {
		logEntry += fmt.Sprintf("\nAssigned To (Reviewer): %s", nextReviewer)
	}
	
	logEntry += fmt.Sprintf("\nRemarks: Approval workflow started\nMain Status: %s\n------------------------", finalStatus)

	// Append to Remarks
	if note.Remarks != "" {
		note.Remarks += logEntry
	} else {
		note.Remarks = strings.TrimSpace(logEntry)
	}

	// Update GreenNote payload
	note.DetailedStatus = finalStatus
	log.Printf("📝 Updated GreenNote %s status to: %s", id, finalStatus)
	
	// Update status in DB
	if err := s.repo.Update(ctx, id, note, orgID, tenantID); err != nil {
		log.Printf("❌ Failed to update status after approval initiation for %s: %v", id, err)
	} else {
		log.Printf("✅ Saved approval initiation log and status update for %s", id)
		log.Printf("📋 Approval log stored in Remarks field (%d characters)", len(note.Remarks))
	}
}

func (s *GreenNoteService) UpdateGreenNote(ctx context.Context, req *greennotepb.UpdateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Note == nil {
		return nil, status.Error(codes.InvalidArgument, "note payload is required")
	}
	existing, _, _, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}

	// Strict requirement: Only draft notes can be updated.
	// Once a note is pending, approved, or rejected, it is locked.
	currentStatus := normalizeStatus(existing.GetDetailedStatus())
	if currentStatus != statusDraft {
		log.Printf("🚫 Update rejected for GreenNote %s: status is '%s' (only '%s' notes can be updated)", 
			req.GetId(), currentStatus, statusDraft)
		return nil, status.Errorf(codes.FailedPrecondition, "only %s notes can be updated; current status is '%s'", statusDraft, currentStatus)
	}

	note := req.Note
	applyDerivedFields(note)
	normalizeStatusOnUpdate(existing, note)

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateGreenNotePayload(note, false); err != nil {
		return nil, err
	}
	// If it was a draft, transition it to pending on update as per user request
	if existing.GetDetailedStatus() == statusDraft {
		note.DetailedStatus = statusPending
		// Sync with enum if needed
		note.Status = greennotepb.Status_STATUS_PENDING
		fmt.Printf("🔄 Transitioning GreenNote %s from Draft to Pending on update\n", req.GetId())
	}

	if err := s.repo.Update(ctx, req.GetId(), note, userCtx.OrgID, userCtx.TenantID); err != nil {
		return nil, err
	}

	// Process categorized documents if present in the update payload
	if err := s.processCategorizedDocuments(ctx, req.GetId(), note, userCtx.OrgID, userCtx.TenantID); err != nil {
		log.Printf("⚠️ Failed to process categorized documents for %s on update: %v", req.GetId(), err)
	}

	// If it was a draft (or now pending), initiate approval
	if note.DetailedStatus == statusPending {
		s.initiateApproval(ctx, req.GetId(), note, userCtx.OrgID, userCtx.TenantID, userCtx.UserID)
	}

	s.publishApprovedIfNeeded(ctx, req.GetId(), existing, note)
	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note updated",
	}, nil
}

func (s *GreenNoteService) CancelGreenNote(ctx context.Context, req *greennotepb.CancelGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if strings.TrimSpace(req.GetCancelReason()) == "" {
		return nil, status.Error(codes.InvalidArgument, "cancel_reason is required")
	}
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Cancel(ctx, req.GetId(), req.GetCancelReason(), userCtx.OrgID, userCtx.TenantID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}

	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note cancelled",
	}, nil
}

// GetOrganizationProjects fetches projects for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationProjects(ctx context.Context, req *greennotepb.GetOrganizationProjectsRequest) (*greennotepb.GetOrganizationProjectsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationProjects - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Project Service
	resp, err := s.projectClient.ListProjectsByOrganization(ctx, &projectpb.ListProjectsByOrganizationRequest{
		OrgId: userCtx.OrgID,
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch projects from external service: %v\n", err)
		// Fallback to empty list instead of failing hard, or return error depending on requirements
		return nil, status.Errorf(codes.Internal, "failed to fetch projects: %v", err)
	}

	// Map external projects to local proto
	projects := make([]*greennotepb.Project, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = &greennotepb.Project{
			Id:          p.ProjectId,
			Name:        p.ProjectName,
			Code:        "", // Not available in project-service
			Status:      "Active", // Default
			Description: "", // Not available in project-service
		}
	}

	return &greennotepb.GetOrganizationProjectsResponse{
		Projects: projects,
		Message:  fmt.Sprintf("Found %d projects for organization %s", len(projects), userCtx.OrgID),
	}, nil
}

// GetOrganizationVendors fetches vendors for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationVendors(ctx context.Context, req *greennotepb.GetOrganizationVendorsRequest) (*greennotepb.GetOrganizationVendorsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationVendors - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Vendor Service
	// vendor-service ListVendors supports filtering by IDs in proto, but we need OrgID filter.
	// As per previous context, vendor service's ListVendors checks metadata for org_id.
	// So we just need to pass the context through.
	// Wait, we need to map the response format.
	resp, err := s.vendorClient.ListVendors(ctx, &vendorpb.ListVendorsRequest{
		TenantId: userCtx.TenantID,
		// Filters might be needed here if proto requires them
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch vendors from external service: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch vendors: %v", err)
	}

	// Map external vendors to local proto
	vendors := make([]*greennotepb.Vendor, len(resp.Vendors))
	for i, v := range resp.Vendors {
		msmeClass := v.MsmeClassification
		if msmeClass == "" && v.GetMsme() != "" {
			msmeClass = v.GetMsme()
		}

		vendors[i] = &greennotepb.Vendor{
			Id:                 v.Id,
			Name:               v.VendorName,
			Code:               v.VendorCode,
			Email:              v.VendorEmail,
			Phone:              v.GetVendorMobile(), // Using getter for optional string
			ContactPerson:      v.BeneficiaryName,
			Status:             v.Status,
			MsmeClassification: msmeClass,
			ActivityType:       v.GetActivityType(),
		}
		fmt.Printf("Mapping Vendor: %s, MSME: %s -> %s, Activity: %s\n", v.VendorName, v.MsmeClassification, msmeClass, v.GetActivityType())
	}

	return &greennotepb.GetOrganizationVendorsResponse{
		Vendors: vendors,
		Message: fmt.Sprintf("Found %d vendors for organization %s", len(vendors), userCtx.OrgID),
	}, nil
}

func (s *GreenNoteService) UploadGreenNoteDocuments(ctx context.Context, req *greennotepb.UploadGreenNoteDocumentsRequest) (*greennotepb.UploadGreenNoteDocumentsResponse, error) {
	if req == nil || req.NoteId == "" {
		return nil, status.Error(codes.InvalidArgument, "note_id is required")
	}

	if len(req.Documents) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one document is required")
	}

	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	var docsToSave []*greennotepb.SupportingDocument
	for _, d := range req.Documents {
		// Strict 10MB Limit
		fileSize := int64(len(d.FileContent))
		if fileSize == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "file '%s' is empty", d.OriginalFilename)
		}
		if fileSize > 10*1024*1024 {
			return nil, status.Errorf(codes.InvalidArgument, "file '%s' exceeds 10MB limit (size: %d bytes)", d.OriginalFilename, fileSize)
		}

		// Allowed Types: PDF, JPG, PNG, DOCX
		ext := strings.ToLower(filepath.Ext(d.OriginalFilename))
		allowed := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".docx": true}
		if !allowed[ext] {
			return nil, status.Errorf(codes.InvalidArgument, "file type '%s' not allowed for '%s'. Supported: PDF, JPG, PNG, DOCX", ext, d.OriginalFilename)
		}

		// Generate Object Key: greennote-docs/{note_id}/{type}/{filename}
		objectKey := fmt.Sprintf("greennote-docs/%s/%s/%s", req.NoteId, d.DocumentType, d.OriginalFilename)

		// Upload to MinIO
		if err := s.storage.Save(ctx, objectKey, d.FileContent, d.MimeType); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to upload document '%s' to storage: %v", d.OriginalFilename, err)
		}

		docsToSave = append(docsToSave, &greennotepb.SupportingDocument{
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			FileSize:         fileSize,
			ObjectKey:        objectKey,
			DocumentType:     d.DocumentType,
		})
	}

	if err := s.repo.AddSupportingDocuments(ctx, req.NoteId, docsToSave, userCtx.OrgID, userCtx.TenantID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save document metadata: %v", err)
	}

	return &greennotepb.UploadGreenNoteDocumentsResponse{
		Success: true,
		Message: fmt.Sprintf("successfully uploaded and linked %d documents", len(docsToSave)),
	}, nil
}

func (s *GreenNoteService) processCategorizedDocuments(ctx context.Context, noteID string, payload *greennotepb.GreenNotePayload, orgID, tenantID string) error {
	var docs []*greennotepb.SupportingDocument

	handleDoc := func(d *greennotepb.SupportingDocumentUpload, section string) error {
		if d == nil || len(d.FileContent) == 0 {
			return nil
		}

		// Validation
		fileSize := int64(len(d.FileContent))
		if fileSize > 10*1024*1024 {
			return fmt.Errorf("file '%s' in section '%s' exceeds 10MB limit", d.OriginalFilename, section)
		}

		ext := strings.ToLower(filepath.Ext(d.OriginalFilename))
		allowed := map[string]bool{".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".docx": true}
		if !allowed[ext] {
			return fmt.Errorf("file type '%s' not allowed for '%s' in section '%s'", ext, d.OriginalFilename, section)
		}

		// Generate Object Key
		objectKey := fmt.Sprintf("greennote-docs/%s/%s/%s", noteID, section, d.OriginalFilename)
		
		// Upload
		if err := s.storage.Save(ctx, objectKey, d.FileContent, d.MimeType); err != nil {
			return fmt.Errorf("failed to upload '%s' to storage: %w", d.OriginalFilename, err)
		}

		docs = append(docs, &greennotepb.SupportingDocument{
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			FileSize:         fileSize,
			ObjectKey:        objectKey,
			DocumentType:     section,
		})
		return nil
	}

	if err := handleDoc(payload.AdditionalInfoDocument, "additional"); err != nil {
		return err
	}
	if err := handleDoc(payload.ContractExtensionDocument, "contract_extension"); err != nil {
		return err
	}
	if err := handleDoc(payload.HrDepartmentDocument, "hr"); err != nil {
		return err
	}

	if len(docs) == 0 {
		return nil
	}

	return s.repo.AddSupportingDocuments(ctx, noteID, docs, orgID, tenantID)
}

// GetOrganizationDepartments fetches departments for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationDepartments(ctx context.Context, req *greennotepb.GetOrganizationDepartmentsRequest) (*greennotepb.GetOrganizationDepartmentsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationDepartments - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Department Service
	resp, err := s.deptClient.ListDepartments(ctx, &departmentpb.ListDepartmentsRequest{
		Page:     1,    // Default to page 1
		PageSize: 1000, // Fetch logical all for dropdown
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch departments from external service: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch departments: %v", err)
	}

	// Map external departments to local proto
	departments := make([]*greennotepb.Department, len(resp.Departments))
	for i, d := range resp.Departments {
		departments[i] = &greennotepb.Department{
			Id:          d.Id,
			Name:        d.Name,
			Code:        "", // Not available in department-service
			Description: d.Description,
			HeadName:    "", // Not available in department-service
			Status:      "Active",
		}
	}

	return &greennotepb.GetOrganizationDepartmentsResponse{
		Departments: departments,
		Message:     fmt.Sprintf("Found %d departments for organization %s", len(departments), userCtx.OrgID),
	}, nil
}

func validateGreenNotePayload(p *greennotepb.GreenNotePayload, isCreate bool) error {
	if p == nil {
		return status.Error(codes.InvalidArgument, "payload is required")
	}
	if strings.TrimSpace(p.GetProjectName()) == "" {
		return status.Error(codes.InvalidArgument, "project_name is required")
	}
	if strings.TrimSpace(p.GetSupplierName()) == "" {
		return status.Error(codes.InvalidArgument, "supplier_name is required")
	}
	if strings.TrimSpace(p.GetExpenseCategory()) == "" {
		return status.Error(codes.InvalidArgument, "expense_category is required")
	}
	_ = isCreate
	return nil
}

// noopEventPublisher is used internally when no publisher is provided.
type noopEventPublisher struct{}

func (n noopEventPublisher) PublishGreenNoteApproved(ctx context.Context, event ports.GreenNoteApprovedEvent) error {
	return nil
}

// normalizeStatus converts various textual representations into canonical
// internal status values.
// normalizeStatus converts various textual representations into canonical
// internal status values (raw strings).
func normalizeStatus(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return ""
	}
	// Check for standard ones
	switch s {
	case "pending", "status_pending":
		return statusPending
	case "approved", "status_approved":
		return statusApproved
	case "rejected", "reject", "status_rejected":
		return statusRejected
	case "draft", "status_draft":
		return statusDraft
	case "cancelled", "status_cancelled":
		return statusCancelled
	}
	// For dynamic status, return as is (lowercase)
	return s
}

// normalizeStatusOnCreate sets the default status for a newly created
// GreenNote. It respects the Status enum if DetailedStatus is not provided.
func normalizeStatusOnCreate(p *greennotepb.GreenNotePayload) {
	if p == nil {
		return
	}
	// 1. Check DetailedStatus (string)
	detailed := normalizeStatus(p.GetDetailedStatus())
	if detailed != "" {
		p.DetailedStatus = detailed
		return
	}

	// 2. Check Status (enum)
	enumStatus := mapStatusEnumToString(p.GetStatus())
	if enumStatus != "" {
		// Prevent manual creation as "approved"
		if enumStatus == statusApproved {
			enumStatus = statusPending
		}
		p.DetailedStatus = enumStatus
		return
	}

	// 3. Default to draft (will transition to pending on view/edit)
	p.DetailedStatus = statusDraft
}

func mapStatusEnumToString(s greennotepb.Status) string {
	switch s {
	case greennotepb.Status_STATUS_PENDING:
		return statusPending
	case greennotepb.Status_STATUS_APPROVED:
		return statusApproved
	case greennotepb.Status_STATUS_REJECTED:
		return statusRejected
	case greennotepb.Status_STATUS_DRAFT:
		return statusDraft
	case greennotepb.Status_STATUS_CANCELLED:
		return statusCancelled
	default:
		return ""
	}
}

// normalizeStatusOnUpdate preserves or updates the status string.
// It respects the Status enum if DetailedStatus is not provided in the update.
func normalizeStatusOnUpdate(existing, updated *greennotepb.GreenNotePayload) {
	if existing == nil || updated == nil {
		return
	}
	// 1. Check DetailedStatus (string)
	detailed := normalizeStatus(updated.GetDetailedStatus())
	if detailed != "" {
		updated.DetailedStatus = detailed
		return
	}

	// 2. Check Status (enum)
	enumStatus := mapStatusEnumToString(updated.GetStatus())
	if enumStatus != "" {
		// Prevent manual update to "approved" if not already there (which it shouldn't be)
		if enumStatus == statusApproved && existing.GetDetailedStatus() != statusApproved {
			enumStatus = existing.GetDetailedStatus()
		}
		updated.DetailedStatus = enumStatus
		return
	}

	// 3. Fallback to existing
	updated.DetailedStatus = existing.GetDetailedStatus()
}

// applyDerivedFields computes financial and budget-related derived fields on
// the payload so that the repository always persists consistent values.
func applyDerivedFields(p *greennotepb.GreenNotePayload) {
	if p == nil {
		return
	}

	// 1. De-duplicate invoices
	// Ensure the primary invoice number doesn't appear in the multiple invoices list to avoid double counting.
	if p.Invoice != nil && len(p.Invoices) > 0 {
		priNum := strings.TrimSpace(p.Invoice.InvoiceNumber)
		var cleaned []*greennotepb.InvoiceInput
		for _, in := range p.Invoices {
			if in != nil && strings.TrimSpace(in.InvoiceNumber) != priNum {
				cleaned = append(cleaned, in)
			}
		}
		p.Invoices = cleaned
	}

	// 2. Aggregate Totals
	if p.EnableMultipleInvoices {
		// When multiple invoices are enabled, the order totals MUST be the sum of all invoices.
		var baseTotal, gstTotal, otherTotal, valueTotal float64

		// Sum primary invoice
		if inv := p.Invoice; inv != nil {
			if inv.InvoiceValue == 0 {
				inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
			}
			baseTotal += inv.TaxableValue
			gstTotal += inv.Gst
			otherTotal += inv.OtherCharges
			valueTotal += inv.InvoiceValue
		}

		// Sum multiple invoices list
		for _, inv := range p.Invoices {
			if inv == nil {
				continue
			}
			if inv.InvoiceValue == 0 {
				inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
			}
			baseTotal += inv.TaxableValue
			gstTotal += inv.Gst
			otherTotal += inv.OtherCharges
			valueTotal += inv.InvoiceValue
		}

		// Override top-level fields with aggregated sums
		p.BaseValue = baseTotal
		p.Gst = gstTotal
		p.OtherCharges = otherTotal
		p.TotalAmount = valueTotal
	} else {
		// For single invoice: derivation only happens if top-level fields are missing or zero.
		if inv := p.Invoice; inv != nil {
			if inv.InvoiceValue == 0 {
				inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
			}
			// If top-level fields are zero, inherit from the single invoice provided.
			if p.BaseValue == 0 && p.Gst == 0 && p.OtherCharges == 0 {
				p.BaseValue = inv.TaxableValue
				p.Gst = inv.Gst
				p.OtherCharges = inv.OtherCharges
			}
		}
		// Final calculation for single invoice mode: always derive TotalAmount from components.
		p.TotalAmount = p.BaseValue + p.OtherCharges + p.Gst
	}

	// 3. Budget & Expenditure calculations
	p.ExpenditureOverBudget = p.ActualExpenditure - p.BudgetExpenditure
}

// publishApprovedIfNeeded emits a GreenNoteApprovedEvent when a note
// transitions into the approved status. This is used to trigger payment note
// creation in downstream services via Kafka.
func (s *GreenNoteService) publishApprovedIfNeeded(ctx context.Context, id string, before, after *greennotepb.GreenNotePayload) {
	if s == nil || s.events == nil || after == nil {
		return
	}
	newStatus := normalizeStatus(after.GetDetailedStatus())
	
	// If after is draft and before was not draft, it MIGHT be a transition from approved.
	// But the user specifically wants APPROVED -> draft.
	// So we trigger if it WAS approved (transiently) or if it's currently draft but came from a non-draft state?
	// Actually, the most reliable way is if after is draft and oldStatus was pending/etc.
	// BUT a user can create a draft directly.
	
	// Let's stick to the explicit transition if possible, but here we only have the states.
	// If after is draft and before was NOT draft, and it's not a new creation (before != nil),
	// it's likely an approval transition.
	
	if newStatus != statusApproved {
		return
	}
	
	oldStatus := ""
	if before != nil {
		oldStatus = normalizeStatus(before.GetDetailedStatus())
	}
	
	if oldStatus == statusDraft || oldStatus == "" {
		return
	}

	event := ports.GreenNoteApprovedEvent{
		GreenNoteID: id,
		OrderNo:     fmt.Sprintf("GN-%s", id),
		PoNumber:    after.GetPoNumber(),
		NetAmount:   after.GetTotalAmount(),
		Status:      statusApproved,
		Comments:    after.GetRemarks(),
		ApprovedAt:  s.clock().UTC().Format(time.RFC3339),
	}
	_ = s.events.PublishGreenNoteApproved(ctx, event)
}

func (s *GreenNoteService) LinkPaymentNote(ctx context.Context, req *greennotepb.LinkPaymentNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.repo.LinkPaymentNote(ctx, req.GetId(), req.GetPaymentNoteId()); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to link payment note: %v", err)
	}

	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "payment note linked successfully",
	}, nil
}
