package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Vendor CRUD operations
		"/vendor.v1.VendorService/CreateVendor":       {"create-vendors"}, // Plural
		"/vendor.v1.VendorService/GetVendor":          {"view-vendors"},
		"/vendor.v1.VendorService/GetVendorByCode":    {"view-vendors"},
		"/vendor.v1.VendorService/ListVendors":        {"view-vendors"},
		"/vendor.v1.VendorService/UpdateVendor":       {"edit-vendors"}, // Plural
		"/vendor.v1.VendorService/DeleteVendor":       {"delete-vendors"}, // Plural
		
		// Vendor Code operations
		"/vendor.v1.VendorService/GenerateVendorCode":   {"create-vendors"}, // Plural
		"/vendor.v1.VendorService/UpdateVendorCode":     {"edit-vendors"}, // Plural
		"/vendor.v1.VendorService/RegenerateVendorCode": {"edit-vendors"}, // Plural
		
		// Vendor Account operations
		"/vendor.v1.VendorService/CreateVendorAccount":    {"edit-vendors"}, // Aligned to edit-vendors or create-vendors? User list has create-vendors. Accounts are sub-resource. I'll use edit-vendors for updating, create-vendors for creating? User list: "edit-vendors", "create-vendors". Vendor accounts are part of vendor mgmt. I'll map to edit-vendors generally or verify if "manage-vendor-accounts" is in list. It is NOT. I will use `edit-vendors` for account management to be safe and consistent.
		"/vendor.v1.VendorService/GetVendorAccounts":       {"view-vendors"},
		"/vendor.v1.VendorService/GetVendorBankingDetails": {"view-vendors"},
		"/vendor.v1.VendorService/UpdateVendorAccount":     {"edit-vendors"},
		"/vendor.v1.VendorService/DeleteVendorAccount":     {"edit-vendors"},
		"/vendor.v1.VendorService/ToggleAccountStatus":     {"edit-vendors"},
		
		// Dropdown operations
		"/vendor.v1.VendorService/GetProjectsDropdown":     {"view-vendors"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		// Vendor code generation can be public for UI
		"/vendor.v1.VendorService/GenerateVendorCode",
	}
}
