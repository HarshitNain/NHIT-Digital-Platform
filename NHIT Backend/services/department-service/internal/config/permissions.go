package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Department CRUD operations
		"/departments.DepartmentService/CreateDepartment": {"create-department"},
		"/departments.DepartmentService/GetDepartment":    {"view-department"},
		"/departments.DepartmentService/ListDepartments":  {"view-department"},
		"/departments.DepartmentService/UpdateDepartment": {"edit-department"},
		"/departments.DepartmentService/DeleteDepartment": {"delete-department"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Add public methods here if any
	}
}
