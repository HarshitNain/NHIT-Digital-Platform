package grpc

import (
	"context"
	"fmt"
	"bytes"

	"strings"
	"time"

	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
)

// VendorGRPCHandler implements the gRPC handler for vendor service
type VendorGRPCHandler struct {
	vendorpb.UnimplementedVendorServiceServer
	vendorService ports.VendorService
	minioClient   *storage.MinIOClient
	logger        ports.Logger
}

// NewVendorGRPCHandler creates a new gRPC handler
func NewVendorGRPCHandler(vendorService ports.VendorService, minioClient *storage.MinIOClient, logger ports.Logger) vendorpb.VendorServiceServer {
	return &VendorGRPCHandler{
		vendorService: vendorService,
		minioClient:   minioClient,
		logger:        logger,
	}
}

// CreateVendor creates a new vendor
func (h *VendorGRPCHandler) CreateVendor(ctx context.Context, req *vendorpb.CreateVendorRequest) (*vendorpb.VendorResponse, error) {
	tenantIDStr, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	createdByUUID, _ := uuid.Parse(userIDStr)

	var orgUUID uuid.UUID
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if ok {
		orgUUID, _ = uuid.Parse(orgID)
	} else if req.OrgId != "" {
		orgUUID, _ = uuid.Parse(req.OrgId)
	}

	if orgUUID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}

	// Build create params
	params := domain.CreateVendorParams{
		TenantID:               tenantUUID,
		OrgID:                  orgUUID,
		VendorName:             req.VendorName,
		VendorEmail:            req.VendorEmail,
		VendorMobile:           req.VendorMobile,
		AccountType:            convertAccountType(req.AccountType),
		VendorNickName:         req.VendorNickName,
		ActivityType:           req.ActivityType,
		GSTIN:                  req.Gstin,
		PAN:                    req.Pan,
		PIN:                    req.Pin,
		CountryName:            req.CountryName,
		StateName:              req.StateName,
		CityName:               req.CityName,
		MSMEClassification:     convertMSMEClassification(req.MsmeClassification),
		MSME:                   req.Msme,
		MSMERegistrationNumber: req.MsmeRegistrationNumber,
		MSMEStartDate:          nil, // TODO: convert if needed
		MSMEEndDate:            nil, // TODO: convert if needed
		MaterialNature:         req.MaterialNature,
		GSTDefaulted:           req.GstDefaulted,
		Section206ABVerified:   req.Section_206AbVerified,
		BeneficiaryName:        req.BeneficiaryName,
		RemarksAddress:         req.RemarksAddress,
		CommonBankDetails:      req.CommonBankDetails,
		IncomeTaxType:          req.IncomeTaxType,
		ProjectID:              req.Project,
		Status:                 convertVendorStatus(req.Status),
		FromAccountType:        req.FromAccountType,
		AccountName:            req.AccountName,
		ShortName:              req.ShortName,
		Parent:                 req.Parent,
		CreatedBy:              createdByUUID,
		AccountNumber:          req.AccountNumber,
		NameOfBank:             req.NameOfBank,
		IFSCCode:               req.IfscCode,
		Address:                req.Address,
		DocumentURL:            req.DocumentUrl,
	}

	// Parse MSME dates if provided
	if req.MsmeStartDate != nil && *req.MsmeStartDate != "" {
		if t, err := time.Parse("2006-01-02", *req.MsmeStartDate); err == nil {
			params.MSMEStartDate = &t
		}
	}
	if req.MsmeEndDate != nil && *req.MsmeEndDate != "" {
		if t, err := time.Parse("2006-01-02", *req.MsmeEndDate); err == nil {
			params.MSMEEndDate = &t
		}
	}

	vendor, err := h.vendorService.CreateVendor(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

// ListVendors lists vendors by tenant (from JWT)
func (h *VendorGRPCHandler) ListVendors(ctx context.Context, req *vendorpb.ListVendorsRequest) (*vendorpb.ListVendorsResponse, error) {
	tenantIDStr, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantUUID, _ := uuid.Parse(tenantIDStr)

	var orgIDPtr *uuid.UUID
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if ok {
		if parsed, err := uuid.Parse(orgID); err == nil {
			orgIDPtr = &parsed
		}
	}

	// Build filters with org_id from JWT
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Offset)

	// Handle Page/PageSize/PerPage if provided
	if req.Page > 0 {
		// Priority: PageSize > PerPage > Limit > DEFAULT(10)
		pSize := req.PageSize
		if pSize <= 0 {
			pSize = req.PerPage
		}
		if pSize > 0 {
			limit = int(pSize)
		} else if limit <= 0 {
			limit = 10
		}
		offset = int((req.Page - 1) * int32(limit))
	} else {
		// If page not provided, check if PageSize or PerPage should override limit
		if req.PageSize > 0 {
			limit = int(req.PageSize)
		} else if req.PerPage > 0 {
			limit = int(req.PerPage)
		}
	}

	filters := ports.VendorListFilters{
		IsActive: req.IsActive,
		OrgID:    orgIDPtr, // Filter by logged-in user's org (optional)
		Project:  req.Project,
		Search:   req.Search,
		Limit:    limit,
		Offset:   offset,
	}

	if orgID != "" {
		h.logger.Info("Filtering vendors by Organization ID (ignoring project-based filtering)", "org_id", orgID)
	}

	vendors, total, err := h.vendorService.ListVendors(ctx, tenantUUID, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoVendors := make([]*vendorpb.Vendor, len(vendors))
	for i, v := range vendors {
		protoVendors[i] = toProtoVendor(v)
	}

	// Calculate pagination if page/pageSize are used
	page := req.Page
	if page <= 0 {
		// Calculate page from offset if possible
		if limit > 0 {
			page = int32(offset / limit) + 1
		} else {
			page = 1
		}
	}
	
	pageSize := int32(limit)
	if req.PageSize > 0 {
		pageSize = req.PageSize
	} else if req.PerPage > 0 {
		pageSize = req.PerPage
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &vendorpb.ListVendorsResponse{
		Vendors:    protoVendors,
		TotalCount: total,
		Pagination: &vendorpb.PaginationMetadata{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}, nil
}

// GenerateVendorCode generates a vendor code
func (h *VendorGRPCHandler) GenerateVendorCode(ctx context.Context, req *vendorpb.GenerateVendorCodeRequest) (*vendorpb.GenerateVendorCodeResponse, error) {
	if req.VendorName == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_name is required")
	}

	accountType := convertAccountType(req.AccountType)

	code, err := h.vendorService.GenerateVendorCode(ctx, req.VendorName, &accountType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vendorpb.GenerateVendorCodeResponse{
		VendorCode: code,
	}, nil
}

// GetProjectsDropdown fetches projects from project-service
func (h *VendorGRPCHandler) GetProjectsDropdown(ctx context.Context, req *vendorpb.GetProjectsDropdownRequest) (*vendorpb.GetProjectsDropdownResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		h.logger.Warn("No org_id found in metadata for GetProjectsDropdown")
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	
	// Connect to project-service at port 50057
	conn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		h.logger.Error("Failed to connect to project service", "error", err)
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	defer conn.Close()
	
	// Create project service client
	projectClient := projectpb.NewProjectServiceClient(conn)
	
	// Forward metadata
	md, _ := metadata.FromIncomingContext(ctx)
	outCtx := metadata.NewOutgoingContext(ctx, md)
	
	// Fetch projects by organization ID
	projectResp, err := projectClient.ListProjectsByOrganization(outCtx, &projectpb.ListProjectsByOrganizationRequest{
		OrgId: orgID,
	})
	if err != nil {
		h.logger.Error("Failed to fetch projects from project-service", "error", err)
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	
	// Convert projects to dropdown items
	projects := make([]*vendorpb.ProjectDropdownItem, 0)
	if projectResp != nil && len(projectResp.Projects) > 0 {
		h.logger.Info("Found projects from project-service", "count", len(projectResp.Projects), "org_id", orgID)
		for _, project := range projectResp.Projects {
			projects = append(projects, &vendorpb.ProjectDropdownItem{
				Id:   project.ProjectId,
				Name: project.ProjectName,
			})
		}
	} else {
		h.logger.Warn("No projects found in project-service", "org_id", orgID)
	}

	return &vendorpb.GetProjectsDropdownResponse{
		Projects: projects,
	}, nil
}

// GetVendor retrieves a vendor by ID
func (h *VendorGRPCHandler) GetVendor(ctx context.Context, req *vendorpb.GetVendorRequest) (*vendorpb.VendorResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	vendor, err := h.vendorService.GetVendorByID(ctx, orgUUID, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) GetVendorByCode(ctx context.Context, req *vendorpb.GetVendorByCodeRequest) (*vendorpb.VendorResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	if req.VendorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_code is required")
	}
	
	vendor, err := h.vendorService.GetVendorByCode(ctx, orgUUID, req.VendorCode)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) UpdateVendor(ctx context.Context, req *vendorpb.UpdateVendorRequest) (*vendorpb.VendorResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	params := domain.UpdateVendorParams{
		VendorName:             req.VendorName,
		VendorEmail:            req.VendorEmail,
		VendorMobile:           req.VendorMobile,
		PAN:                    req.Pan,
		BeneficiaryName:        req.BeneficiaryName,
		// Add new fields
		MSME:                   req.Msme,
		MSMERegistrationNumber: req.MsmeRegistrationNumber,
		MaterialNature:         req.MaterialNature,
		GSTDefaulted:           req.GstDefaulted,
		Section206ABVerified:   req.Section_206AbVerified,
		IncomeTaxType:          req.IncomeTaxType,
		ProjectID:              req.Project,
		AccountName:            req.AccountName,
		RemarksAddress:         req.RemarksAddress,
		Address:                req.Address,
		PIN:                    req.Pin,
		ActivityType:           req.ActivityType,
		VendorNickName:         req.VendorNickName,
		CountryName:            req.CountryName,
		StateName:              req.StateName,
		CityName:               req.CityName,
		FromAccountType:        req.FromAccountType,
		ShortName:              req.ShortName,
		Parent:                 req.Parent,
		DocumentURL:            req.DocumentUrl,
	}

	// Parse MSME dates if provided
	if req.MsmeStartDate != nil && *req.MsmeStartDate != "" {
		if t, err := time.Parse("2006-01-02", *req.MsmeStartDate); err == nil {
			params.MSMEStartDate = &t
		}
	}
	if req.MsmeEndDate != nil && *req.MsmeEndDate != "" {
		if t, err := time.Parse("2006-01-02", *req.MsmeEndDate); err == nil {
			params.MSMEEndDate = &t
		}
	}

	// Handle MSME Classification
	if req.MsmeClassification != nil {
		val := convertMSMEClassification(*req.MsmeClassification)
		params.MSMEClassification = &val
	}
	// Handle Account Type from request if needed
	if req.FromAccountType != nil {
		val := convertAccountType(*req.FromAccountType)
		params.AccountType = &val
	}
	// Handle Status
	if req.Status != nil {
		val := convertVendorStatus(*req.Status)
		params.Status = &val
	}
	
	vendor, err := h.vendorService.UpdateVendor(ctx, orgUUID, vendorUUID, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) DeleteVendor(ctx context.Context, req *vendorpb.DeleteVendorRequest) (*emptypb.Empty, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	if err := h.vendorService.DeleteVendor(ctx, orgUUID, vendorUUID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &emptypb. Empty{}, nil
}

func (h *VendorGRPCHandler) UpdateVendorCode(ctx context.Context, req *vendorpb.UpdateVendorCodeRequest) (*vendorpb.VendorResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	if req.VendorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_code is required")
	}
	
	vendor, err := h.vendorService.UpdateVendorCode(ctx, orgUUID, vendorUUID, req.VendorCode)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) RegenerateVendorCode(ctx context.Context, req *vendorpb.RegenerateVendorCodeRequest) (*vendorpb.VendorResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	vendor, err := h.vendorService.RegenerateVendorCode(ctx, orgUUID, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

// Vendor Account methods
func (h *VendorGRPCHandler) CreateVendorAccount(ctx context.Context, req *vendorpb.CreateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	createdByUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id in JWT")
	}
	
	params := domain.CreateVendorAccountParams{
		VendorID:      vendorUUID,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		NameOfBank:    req.NameOfBank,
		IFSCCode:      req.IfscCode,
		IsPrimary:     req.IsPrimary,
		CreatedBy:     createdByUUID,
	}
	
	// Add optional fields
	if req.AccountType != nil {
		params.AccountType = req.AccountType
	}
	if req.BranchName != nil {
		params.BranchName = req.BranchName
	}
	if req.SwiftCode != nil {
		params.SwiftCode = req.SwiftCode
	}
	if req.Remarks != nil {
		params.Remarks = req.Remarks
	}
	
	account, err := h.vendorService.CreateVendorAccount(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

func (h *VendorGRPCHandler) GetVendorAccounts(ctx context.Context, req *vendorpb.GetVendorAccountsRequest) (*vendorpb.GetVendorAccountsResponse, error) {
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	accounts, err := h.vendorService.GetVendorAccounts(ctx, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	protoAccounts := make([]*vendorpb.VendorAccount, len(accounts))
	for i, acc := range accounts {
		protoAccounts[i] = toProtoVendorAccount(acc)
	}
	
	return &vendorpb.GetVendorAccountsResponse{
		Accounts: protoAccounts,
	}, nil
}

func (h *VendorGRPCHandler) GetVendorBankingDetails(ctx context.Context, req *vendorpb.GetVendorBankingDetailsRequest) (*vendorpb.BankingDetailsResponse, error) {
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	var accountUUID *uuid.UUID
	if req.AccountId != nil && *req.AccountId != "" {
		accUUID, err := uuid.Parse(*req.AccountId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid account_id")
		}
		accountUUID = &accUUID
	}
	
	bankingDetails, err := h.vendorService.GetVendorBankingDetails(ctx, vendorUUID, accountUUID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "banking details not found")
	}
	
	return &vendorpb.BankingDetailsResponse{
		BankingDetails: &vendorpb.BankingDetails{
			AccountName:   bankingDetails.AccountName,
			AccountNumber: bankingDetails.AccountNumber,
			AccountType:   bankingDetails.AccountType,
			NameOfBank:    bankingDetails.NameOfBank,
			BranchName:    bankingDetails.BranchName,
			IfscCode:      bankingDetails.IFSCCode,
			SwiftCode:     bankingDetails.SwiftCode,
			Remarks:       bankingDetails.Remarks,
		},
	}, nil
}

func (h *VendorGRPCHandler) UpdateVendorAccount(ctx context.Context, req *vendorpb.UpdateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	params := domain.UpdateVendorAccountParams{
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		NameOfBank:    req.NameOfBank,
		BranchName:    req.BranchName,
		IFSCCode:      req.IfscCode,
		SwiftCode:     req.SwiftCode,
		IsPrimary:     req.IsPrimary,
		IsActive:      req.IsActive,
		Remarks:       req.Remarks,
	}
	
	account, err := h.vendorService.UpdateVendorAccount(ctx, accountUUID, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

func (h *VendorGRPCHandler) DeleteVendorAccount(ctx context.Context, req *vendorpb.DeleteVendorAccountRequest) (*emptypb.Empty, error) {
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	if err := h.vendorService.DeleteVendorAccount(ctx, accountUUID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &emptypb.Empty{}, nil
}

func (h *VendorGRPCHandler) ToggleAccountStatus(ctx context.Context, req *vendorpb.ToggleAccountStatusRequest) (*vendorpb.VendorAccountResponse, error) {
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	account, err := h.vendorService.ToggleAccountStatus(ctx, accountUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

// UploadVendorDocument uploads a vendor document (PDF, XLSX, DOC, etc.)
func (h *VendorGRPCHandler) UploadVendorDocument(ctx context.Context, req *vendorpb.UploadVendorDocumentRequest) (*vendorpb.UploadVendorDocumentResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "org_id not found in context")
	}
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid org_id")
	}

	if req.VendorId == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_id is required")
	}
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}

	if len(req.FileContent) == 0 {
		return nil, status.Error(codes.InvalidArgument, "file_content is required")
	}

	// Use provided filename or generate one
	filename := req.FileName
	if filename == "" {
		filename = fmt.Sprintf("document_%s_%d.dat", req.VendorId, time.Now().Unix())
	} else {
		// Basic sanitization or prefixing
		filename = fmt.Sprintf("%d_%s", time.Now().Unix(), filename)
	}

	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileSize := int64(len(req.FileContent))
	fileReader := bytes.NewReader(req.FileContent)

	// Use MinIO client to upload (still using UploadSignature for now but it's just a name for a generic upload internally)
	// Actually if it's generic, we might want to check if there's a more generic upload method in MinIO client.
	documentURL, err := h.minioClient.UploadSignature(ctx, req.VendorId, filename, fileReader, fileSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload document: %v", err)
	}

	// Update vendor record with new document URL
	params := domain.UpdateVendorParams{
		DocumentURL: &documentURL,
	}

	_, err = h.vendorService.UpdateVendor(ctx, orgUUID, vendorUUID, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update vendor document url: %v", err)
	}

	return &vendorpb.UploadVendorDocumentResponse{
		Success:      true,
		Message:      "Document uploaded successfully",
		VendorId:     req.VendorId,
		DocumentUrl:  documentURL,
		FileName:     filename,
		MimeType:     mimeType,
		FileSize:     fileSize,
		UploadedAt:   timestamppb.Now(),
	}, nil
}

// Helper functions
func toProtoVendor(v *domain.Vendor) *vendorpb.Vendor {
	protoVendor := &vendorpb.Vendor{
		Id:                    v.ID.String(),
		TenantId:              v.TenantID.String(),
		OrgId:                 v.OrgID.String(),
		VendorCode:            v.VendorCode,
		VendorName:            v.VendorName,
		VendorEmail:           v.VendorEmail,
		VendorMobile:          v.VendorMobile,
		AccountType:           convertAccountType(v.AccountType),
		VendorNickName:        v.VendorNickName,
		ActivityType:          v.ActivityType,
		Gstin:                 v.GSTIN,
		Pan:                   v.PAN,
		Pin:                   v.PIN,
		CountryName:           v.CountryName,
		StateName:             v.StateName,
		CityName:              v.CityName,
		MsmeClassification:    v.MSMEClassification, // Already string in domain, just assign directly if normalized. Or use helper if needing normalization.
		Msme:                  v.MSME,
		MsmeRegistrationNumber: v.MSMERegistrationNumber,
		MaterialNature:        v.MaterialNature,
		GstDefaulted:          v.GSTDefaulted,
		Section_206AbVerified: v.Section206ABVerified,
		BeneficiaryName:       v.BeneficiaryName,
		RemarksAddress:        v.RemarksAddress,
		CommonBankDetails:     v.CommonBankDetails,
		IncomeTaxType:         v.IncomeTaxType,
		Project:               v.ProjectID,
		FromAccountType:       v.FromAccountType,
		AccountName:           v.AccountName,
		ShortName:             v.ShortName,
		Parent:                v.Parent,
		CodeAutoGenerated:     v.CodeAutoGenerated,
		IsActive:              v.Status == "ACTIVE",
		CreatedBy:             v.CreatedBy.String(),
		CreatedAt:             timestamppb.New(v.CreatedAt),
		UpdatedAt:             timestamppb.New(v.UpdatedAt),
		NameOfBank:            v.NameOfBank,
		IfscCode:              v.IFSCCode,
		Address:               v.Address,
		DocumentUrl:          v.DocumentURL,
	}

	// Map MSME dates
	if v.MSMEStartDate != nil {
		protoVendor.MsmeStartDate = timestamppb.New(*v.MSMEStartDate)
	}
	if v.MSMEEndDate != nil {
		protoVendor.MsmeEndDate = timestamppb.New(*v.MSMEEndDate)
	}
	
	// Map Status
	if v.Status == "" {
		protoVendor.Status = "ACTIVE"
	} else {
		protoVendor.Status = v.Status
	}

	return protoVendor
}

func convertAccountType(at string) string {
	// Normalize to uppercase
	at = strings.ToUpper(at)
	
	switch at {
	case "INTERNAL":
		return "INTERNAL"
	case "EXTERNAL":
		return "EXTERNAL"
	default:
		// Default to EXTERNAL if unspecified or invalid, matching basic default logic
		// Or return empty if strict validation is needed. 
		// Given proto default is 0 (UNSPECIFIED), but DB usually needs a value.
		// If the user sends "Internal" we want "INTERNAL".
		// If they send "Foo", we probably default to "EXTERNAL" or just return "EXTERNAL" as fallthrough
		// Check against proto map if we want to be strict, but for now simple normalization:
		return "EXTERNAL"
	}
}

func convertMSMEClassification(mc string) string {
	mc = strings.ToUpper(mc)
	switch mc {
	case "MICRO":
		return "MICRO"
	case "SMALL":
		return "SMALL"
	case "MEDIUM":
		return "MEDIUM"
	default:
		// If empty or invalid, return implicit default or empty string
		return "MSME_CLASSIFICATION_UNSPECIFIED"
	}
}

func convertVendorStatus(vs string) string {
	vs = strings.ToUpper(vs)
	switch vs {
	case "ACTIVE":
		return "ACTIVE"
	case "INACTIVE":
		return "INACTIVE"
	default:
		return "VENDOR_STATUS_UNSPECIFIED"
	}
}

func convertStringToMSMEClassification(s string) string {
	s = strings.ToUpper(s)
	switch s {
	case "MICRO":
		return "MICRO"
	case "SMALL":
		return "SMALL"
	case "MEDIUM":
		return "MEDIUM"
	default:
		return "MSME_CLASSIFICATION_UNSPECIFIED"
	}
}
