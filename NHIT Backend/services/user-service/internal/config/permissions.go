package config

// GetPermissionMap returns the permission requirements for each RPC method
// Note: Method paths are /UserManagement/MethodName because proto has no package declaration
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Tenant
		"/UserManagement/GetTenant":    {"system-configuration"}, // Only super admin usually, or system-configuration

		// Role Management
		"/UserManagement/CreateRole":              {"create-role"},
		"/UserManagement/ListRoles":               {"view-role"},
		"/UserManagement/ListRolesByOrganization": {"view-role"},
		"/UserManagement/GetRole":                 {"view-role"},
		"/UserManagement/UpdateRole":              {"edit-role"},
		"/UserManagement/DeleteRole":              {"delete-role"},
		"/UserManagement/CloneRole":               {"create-role"},

		// User Management
		"/UserManagement/CreateUser":      {"create-user"},
		"/UserManagement/GetUser":         {"view-user"},
		"/UserManagement/ListUsers":       {"view-user"},
		"/UserManagement/ListUsersPaginated": {"view-user"},
		"/UserManagement/CountUsersByTenant": {"view-user"},
		"/UserManagement/UpdateUser":      {"edit-user"},
		"/UserManagement/DeleteUser":      {"delete-user"}, // Hard delete
		"/UserManagement/DeactivateUser":  {"delete-user"}, // Soft delete
		"/UserManagement/ReactivateUser":  {"edit-user"},   // Restore
		"/UserManagement/AssignRolesToUser": {"edit-user"},
		"/UserManagement/UploadUserSignature": {"edit-user"},

		// User-Organization
		"/UserManagement/AddUserToOrganization":      {"edit-organizations"}, // Matches organization-service
		"/UserManagement/RemoveUserFromOrganization": {"edit-organizations"}, // Matches organization-service
		
		// Activity Logs
		"/UserManagement/ListActivityLogs": {"view-activity-logs"},


	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		"/UserManagement/CreateTenant",           // Signup
		"/UserManagement/ListUserOrganizations",  // Login flow
		"/UserManagement/ListRolesOfUser",        // Login flow
		"/UserManagement/CreateUserLoginHistory", // Login flow
		"/UserManagement/CreateActivityLog",      // Internal/System
		"/UserManagement/GetDepartmentsDropdown", // Dropdowns often need to be accessible or just generic view
		"/UserManagement/GetDesignationsDropdown",
		"/UserManagement/GetRolesDropdown",
		"/UserManagement/GetUser",
	}
}
