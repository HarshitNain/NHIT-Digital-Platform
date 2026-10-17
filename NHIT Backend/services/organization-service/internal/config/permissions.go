package config

// GetPermissionMap returns the permission requirements for each RPC method
// Note: Method paths use proto package name 'organizations' (from organization.proto line 3)
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Organization CRUD operations
		"/organizations.OrganizationService/GetOrganization":          {"view-organizations"},
		"/organizations.OrganizationService/ListOrganizations":        {"view-organizations"},
		"/organizations.OrganizationService/ListOrganizationsByTenant": {"view-organizations"},
		"/organizations.OrganizationService/UpdateOrganization":       {"edit-organizations"},
		"/organizations.OrganizationService/DeleteOrganization":       {"delete-organizations"},
		
		// User-Organization operations
		"/organizations.OrganizationService/AddUserToOrganization":    {"edit-organizations"},
		"/organizations.OrganizationService/RemoveUserFromOrganization": {"edit-organizations"},
		"/organizations.OrganizationService/ListOrganizationUsers":    {"view-organizations"},
		"/organizations.OrganizationService/GetUserOrganizations":     {"view-organizations"},
	}
}

// GetPublicMethods returns methods that don't require authentication
// Note: CreateOrganization is public because parent org creation (before login) 
// doesn't require auth. The handler itself validates auth for child org creation.
// Note: ListUserOrganizations is public to allow auth-service to fetch user orgs during login
func GetPublicMethods() []string {
	return []string{
		"/organizations.OrganizationService/CreateOrganization",
		"/organizations.OrganizationService/ListUserOrganizations",
	}
}


