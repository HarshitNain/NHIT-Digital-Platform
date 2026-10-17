package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		"/greennote.GreenNoteService/CreateGreenNote": {"create-note"},
		"/greennote.GreenNoteService/GetGreenNote":    {"view-note"},
		"/greennote.GreenNoteService/ListGreenNotes":  {"view-all-notes"}, // Admin/Manager view
		"/greennote.GreenNoteService/UpdateGreenNote": {"edit-note"},
		"/greennote.GreenNoteService/CancelGreenNote": {"delete-note"}, // Mapping cancel to delete
		
		// Organization Context Helpers (Dropdowns/Dependencies)
		// These require basic note viewing or creating permissions
		"/greennote.GreenNoteService/GetOrganizationProjects":    {"create-note"}, // Need to create note to pick project
		"/greennote.GreenNoteService/GetOrganizationVendors":     {"create-note"},
		"/greennote.GreenNoteService/GetOrganizationDepartments": {"create-note"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// No public methods in GreenNote service
	}
}
