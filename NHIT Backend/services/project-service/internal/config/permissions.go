package config

// GetPermissionMap returns the permission requirements for each RPC method
// Note: Method paths use proto package name 'project' (from project.proto line 3)
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Project CRUD operations
		"/project.ProjectService/GetProject":    {"view-projects"},
		"/project.ProjectService/ListProjects":  {"view-projects"},
		"/project.ProjectService/UpdateProject": {"edit-project"},
		"/project.ProjectService/DeleteProject": {"delete-project"},
		"/project.ProjectService/ListProjectsByOrganization": {"view-projects"},
	}
}

// GetPublicMethods returns methods that don't require authentication
// Note: CreateProject is public for initial project creation during parent org setup
func GetPublicMethods() []string {
	return []string{
		"/project.ProjectService/CreateProject",
	}
}

