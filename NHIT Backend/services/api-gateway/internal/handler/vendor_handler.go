package handler

import (
    "encoding/json"
    "log"
	"fmt"
	"time"
    "net/http"
    "strings"

    "github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/storage"
    vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
)

type VendorHandler struct {
    client     vendorpb.VendorServiceClient
    storage    *storage.MinIOClient
    bucketName string
}

func NewVendorHandler(client vendorpb.VendorServiceClient, storage *storage.MinIOClient, bucketName string) *VendorHandler {
	return &VendorHandler{
		client:     client,
		storage:    storage,
		bucketName: bucketName,
	}
}

// getVendorName helps extract vendor name for folder creation
func getVendorName(r *http.Request) string {
	if vendorName := r.FormValue("vendor_name"); vendorName != "" {
		return vendorName
	}
	return "unknown_vendor"
}


func (h *VendorHandler) HandleCreateVendor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32MB limit
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	req := &vendorpb.CreateVendorRequest{
		TenantId:               getString(r, "tenant_id"),
		OrgId:                  getString(r, "org_id"),
		VendorName:             getString(r, "vendor_name"),
		VendorEmail:            getString(r, "vendor_email"),
		VendorMobile:           getStringPtr(r, "vendor_mobile"),
		AccountType:            getString(r, "account_type"),
		VendorNickName:         getStringPtr(r, "vendor_nick_name"),
		ActivityType:           getStringPtr(r, "activity_type"),
		Gstin:                  getStringPtr(r, "gstin"),
		Pan:                    getString(r, "pan"),
		Pin:                    getStringPtr(r, "pin"),
		CountryName:            getStringPtr(r, "country_name"),
		StateName:              getStringPtr(r, "state_name"),
		CityName:               getStringPtr(r, "city_name"),
		MsmeClassification:     getString(r, "msme_classification"),
		Msme:                   getStringPtr(r, "msme"),
		MsmeRegistrationNumber: getStringPtr(r, "msme_registration_number"),
		MaterialNature:         getStringPtr(r, "material_nature"),
		GstDefaulted:           getStringPtr(r, "gst_defaulted"),
		Section_206AbVerified:  getStringPtr(r, "section_206ab_verified"),
		BeneficiaryName:        getString(r, "beneficiary_name"),
		RemarksAddress:         getStringPtr(r, "remarks_address"),
		CommonBankDetails:      getStringPtr(r, "common_bank_details"),
		IncomeTaxType:          getStringPtr(r, "income_tax_type"),
		Project:                getStringPtr(r, "project"),
		FromAccountType:        getStringPtr(r, "from_account_type"),
		AccountName:            getStringPtr(r, "account_name"),
		ShortName:              getStringPtr(r, "short_name"),
		Parent:                 getStringPtr(r, "parent"),
		Address:                getStringPtr(r, "address"),
		AccountNumber:          getStringPtr(r, "account_number"),
		NameOfBank:             getStringPtr(r, "name_of_bank"),
		IfscCode:               getStringPtr(r, "ifsc_code"),
		// BranchName and Street are not in the CreateVendorRequest proto
		// BranchName:             getStringPtr(r, "branch_name"), 
		// Street:                 getStringPtr(r, "street"),
	}

	// Explicitly map is_active to Status
	if r.FormValue("is_active") == "true" {
		req.Status = "ACTIVE"
	} else {
		req.Status = "INACTIVE"
	}

	// Dates
	if v := r.FormValue("msme_start_date"); v != "" {
		req.MsmeStartDate = &v
	}
	if v := r.FormValue("msme_end_date"); v != "" {
		req.MsmeEndDate = &v
	}

	// File Upload (supporting multiple files if needed, but here assuming one named 'document')
	// Loop over all files to be safe
	fileHeaders := r.MultipartForm.File
	var uploadedFilePaths []string

	// Sanitize vendor name for path
	vendorName := sanitizeFilename(getVendorName(r))
	if vendorName == "" {
		vendorName = "unknown_vendor" // Fallback if sanitized name is empty
	}

	for _, headers := range fileHeaders {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				log.Printf("Failed to open uploaded file: %v", err)
				continue
			}
			defer file.Close()

			// Path: VendorName/timestamp_filename
			filename := fmt.Sprintf("%s/%d_%s", vendorName, time.Now().Unix(), header.Filename)

			// Upload to target bucket
			url, err := h.storage.UploadStream(r.Context(), h.bucketName, filename, file, header.Size, header.Header.Get("Content-Type"))
			if err != nil {
				log.Printf("Failed to upload file to MinIO: %v", err)
				http.Error(w, "Failed to upload file", http.StatusInternalServerError)
				return
			}
			uploadedFilePaths = append(uploadedFilePaths, url)
		}
	}

	if len(uploadedFilePaths) > 0 {
		req.DocumentUrl = &uploadedFilePaths[0]
	}

	ctx := forwardMetadata(r)
	resp, err := h.client.CreateVendor(ctx, req)
	if err != nil {
		log.Printf("gRPC CreateVendor failed: %v", err)
		http.Error(w, "Failed to create vendor: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *VendorHandler) HandleUpdateVendor(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 5 {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }
    id := parts[4]

    err := r.ParseMultipartForm(32 << 20)
    if err != nil {
        http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
        return
    }

    req := &vendorpb.UpdateVendorRequest{
        VendorId: id,
        TenantId: r.FormValue("tenant_id"),
        VendorName: getStringPtr(r, "vendor_name"),
        VendorEmail: getStringPtr(r, "vendor_email"),
        VendorMobile: getStringPtr(r, "vendor_mobile"),
        AccountType: getStringPtr(r, "account_type"),
        VendorNickName: getStringPtr(r, "vendor_nick_name"),
        ActivityType: getStringPtr(r, "activity_type"),
        Gstin: getStringPtr(r, "gstin"),
        Pan: getStringPtr(r, "pan"),
        Pin: getStringPtr(r, "pin"),
        CountryName: getStringPtr(r, "country_name"),
        StateName: getStringPtr(r, "state_name"),
        CityName: getStringPtr(r, "city_name"),
        MsmeClassification: getStringPtr(r, "msme_classification"),
        Msme: getStringPtr(r, "msme"),
        MsmeRegistrationNumber: getStringPtr(r, "msme_registration_number"),
        MaterialNature: getStringPtr(r, "material_nature"),
        GstDefaulted: getStringPtr(r, "gst_defaulted"),
        Section_206AbVerified: getStringPtr(r, "section_206ab_verified"),
        BeneficiaryName: getStringPtr(r, "beneficiary_name"),
        RemarksAddress: getStringPtr(r, "remarks_address"),
        CommonBankDetails: getStringPtr(r, "common_bank_details"),
        IncomeTaxType: getStringPtr(r, "income_tax_type"),
        Project: getStringPtr(r, "project"),
        FromAccountType: getStringPtr(r, "from_account_type"),
        AccountName: getStringPtr(r, "account_name"),
        ShortName: getStringPtr(r, "short_name"),
        Parent: getStringPtr(r, "parent"),
        Address: getStringPtr(r, "address"), 
        // These fields are not in the UpdateVendorRequest proto
        // AccountNumber: getStringPtr(r, "account_number"),
        // NameOfBank: getStringPtr(r, "name_of_bank"),
        // IfscCode: getStringPtr(r, "ifsc_code"),
        // BranchName: getStringPtr(r, "branch_name"),
        // Street: getStringPtr(r, "street"),
    }
    
    // Is Active
    if v := r.FormValue("is_active"); v != "" {
        isActive := v == "true"
        req.IsActive = &isActive
    }
    
    // Dates
    if v := r.FormValue("msme_start_date"); v != "" { req.MsmeStartDate = &v }
    if v := r.FormValue("msme_end_date"); v != "" { req.MsmeEndDate = &v }

    // File Upload
    fileHeaders := r.MultipartForm.File
    if len(fileHeaders) == 0 {
        goto skipUpload
    }

    // 1. Get current vendor for folder name fallback and cleanup
    {
        ctxMetadata := forwardMetadata(r)
        currentVendor, err := h.client.GetVendor(ctxMetadata, &vendorpb.GetVendorRequest{
            VendorId: id,
            TenantId: r.FormValue("tenant_id"),
        })
        if err != nil {
            log.Printf("Failed to fetch current vendor for update: %v", err)
            http.Error(w, "Failed to fetch vendor details: "+err.Error(), http.StatusInternalServerError)
            return
        }

        folderName := "updated_vendors"
        if v := r.FormValue("vendor_name"); v != "" {
            folderName = sanitizeFilename(v)
        } else if currentVendor != nil && currentVendor.Vendor != nil {
            folderName = sanitizeFilename(currentVendor.Vendor.VendorName)
        }

        var uploadedFilePaths []string
        var oldDocUrl string
        if currentVendor != nil && currentVendor.Vendor != nil && currentVendor.Vendor.DocumentUrl != nil {
            oldDocUrl = *currentVendor.Vendor.DocumentUrl
        }

        for _, headers := range fileHeaders {
            for _, header := range headers {
                file, err := header.Open()
                if err != nil {
                    log.Printf("Failed to open uploaded file: %v", err)
                    continue
                }
                defer file.Close()

                filename := fmt.Sprintf("%s/%d_%s", folderName, time.Now().Unix(), header.Filename)
                
                // 2. Upload new file
                url, err := h.storage.UploadStream(r.Context(), h.bucketName, filename, file, header.Size, header.Header.Get("Content-Type"))
                if err != nil {
                    log.Printf("Failed to upload file: %v", err)
                    http.Error(w, "Failed to upload document: "+err.Error(), http.StatusInternalServerError)
                    return
                }
                uploadedFilePaths = append(uploadedFilePaths, url)
            }
        }

        if len(uploadedFilePaths) > 0 {
            req.DocumentUrl = &uploadedFilePaths[0]
            
            // 3. Perform Update
            ctx := forwardMetadata(r)
            resp, err := h.client.UpdateVendor(ctx, req)
            if err != nil {
                log.Printf("gRPC UpdateVendor failed: %v", err)
                http.Error(w, "Failed to update vendor: "+err.Error(), http.StatusInternalServerError)
                return
            }

            // 4. If update successful, cleanup old document
            if oldDocUrl != "" {
                bucket, obj, pErr := h.storage.ParseBucketAndObject(oldDocUrl)
                if pErr == nil {
                    _ = h.storage.DeleteObject(r.Context(), bucket, obj)
                }
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(resp)
            return
        }
    }

skipUpload:
    ctx := forwardMetadata(r)
    resp, err := h.client.UpdateVendor(ctx, req)
    if err != nil {
        log.Printf("gRPC UpdateVendor failed: %v", err)
        http.Error(w, "Failed to update vendor: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func getString(r *http.Request, key string) string {
    return r.FormValue(key)
}

func getStringPtr(r *http.Request, key string) *string {
    v := r.FormValue(key)
    if v == "" {
        return nil
    }
    return &v
}

