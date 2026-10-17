package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Designation CRUD operations
		"/designations.DesignationService/CreateDesignation": {"create-designation"},
		"/designations.DesignationService/GetDesignation":    {"view-designation"},
		"/designations.DesignationService/ListDesignations":  {"view-designation"},
		"/designations.DesignationService/UpdateDesignation": {"edit-designation"},
		"/designations.DesignationService/DeleteDesignation": {"delete-designation"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Add public methods here if any
	}
}
