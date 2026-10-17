package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/storage"
)

type GreenNoteHandler struct {
	client  greennotepb.GreenNoteServiceClient
	storage *storage.MinIOClient
}

func NewGreenNoteHandler(client greennotepb.GreenNoteServiceClient, storage *storage.MinIOClient) *GreenNoteHandler {
	return &GreenNoteHandler{
		client:  client,
		storage: storage,
	}
}

func (h *GreenNoteHandler) HandleCreateGreenNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var payload *greennotepb.GreenNotePayload

	if strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 1. Parse JSON Payload
		jsonPayload := r.FormValue("payload")
		if jsonPayload == "" {
			http.Error(w, "Missing 'payload' field in multipart form", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
			http.Error(w, "Invalid JSON in 'payload': "+err.Error(), http.StatusBadRequest)
			return
		}

		// 2. Extract Specialized Section Files
		payload.AdditionalInfoDocument = h.extractDoc(r, "additional_info_doc")
		payload.ContractExtensionDocument = h.extractDoc(r, "contract_extension_doc")
		payload.HrDepartmentDocument = h.extractDoc(r, "hr_department_doc")

	} else {
		// Standard JSON request
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	ctx := forwardMetadata(r)
	resp, err := h.client.CreateGreenNote(ctx, &greennotepb.CreateGreenNoteRequest{
		Note: payload,
	})
	if err != nil {
		log.Printf("Failed to create GreenNote: %v", err)
		http.Error(w, "Failed to create GreenNote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GreenNoteHandler) HandleUpdateGreenNote(w http.ResponseWriter, r *http.Request) {
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

	contentType := r.Header.Get("Content-Type")
	var payload *greennotepb.GreenNotePayload

	if strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		jsonPayload := r.FormValue("payload")
		if jsonPayload == "" {
			http.Error(w, "Missing 'payload' field", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
			http.Error(w, "Invalid JSON in 'payload': "+err.Error(), http.StatusBadRequest)
			return
		}

		payload.AdditionalInfoDocument = h.extractDoc(r, "additional_info_doc")
		payload.ContractExtensionDocument = h.extractDoc(r, "contract_extension_doc")
		payload.HrDepartmentDocument = h.extractDoc(r, "hr_department_doc")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	ctx := forwardMetadata(r)
	resp, err := h.client.UpdateGreenNote(ctx, &greennotepb.UpdateGreenNoteRequest{
		Id:   id,
		Note: payload,
	})
	if err != nil {
		log.Printf("Failed to update GreenNote: %v", err)
		http.Error(w, "Failed to update GreenNote: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GreenNoteHandler) HandleUploadDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := parts[4]

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	var docs []*greennotepb.SupportingDocumentUpload
	multiFiles := r.MultipartForm.File
	for fieldName, files := range multiFiles {
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				log.Printf("Error opening file %s: %v", fh.Filename, err)
				continue
			}
			content, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				log.Printf("Error reading file %s: %v", fh.Filename, err)
				continue
			}

			docs = append(docs, &greennotepb.SupportingDocumentUpload{
				Name:             fieldName,
				OriginalFilename: fh.Filename,
				MimeType:         fh.Header.Get("Content-Type"),
				FileContent:      content,
				DocumentType:     r.FormValue(fieldName + "_type"), // Assuming a fieldName_type for each file
			})
		}
	}

	ctx := forwardMetadata(r)
	resp, err := h.client.UploadGreenNoteDocuments(ctx, &greennotepb.UploadGreenNoteDocumentsRequest{
		NoteId:    id,
		Documents: docs,
	})
	if err != nil {
		log.Printf("Failed to upload documents: %v", err)
		http.Error(w, "Failed to upload documents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *GreenNoteHandler) extractDoc(r *http.Request, fieldName string) *greennotepb.SupportingDocumentUpload {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if err != http.ErrMissingFile { // Only log if it's not just a missing optional file
			log.Printf("Error getting form file '%s': %v", fieldName, err)
		}
		return nil
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file content for '%s': %v", fieldName, err)
		return nil
	}

	return &greennotepb.SupportingDocumentUpload{
		Name:             fieldName,
		OriginalFilename: header.Filename,
		MimeType:         header.Header.Get("Content-Type"),
		FileContent:      content,
	}
}
