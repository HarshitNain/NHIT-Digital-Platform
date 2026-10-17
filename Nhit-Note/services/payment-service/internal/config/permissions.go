package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Payment CRUD operations
		"/payment.PaymentService/CreatePaymentRequests":        {"create-payment"},
		"/payment.PaymentService/UpdatePaymentGroup":           {"edit-payment"},
		"/payment.PaymentService/DeletePayment":                {"delete-payment"},
		"/payment.PaymentService/DeletePaymentItem":            {"delete-payment"},
		"/payment.PaymentService/ListPayments":                 {"view-payments"},
		"/payment.PaymentService/ListPaymentsDataTable":        {"view-payments"},
		"/payment.PaymentService/GetPaymentGroup":              {"view-payments"},
		
		// Search operations
		"/payment.PaymentService/SearchVendors":                {"view-vendors"},
		"/payment.PaymentService/SearchInternalVendors":        {"view-vendors"},
		"/payment.PaymentService/SearchProjects":               {"view-projects"},
		"/payment.PaymentService/GetAllVendors":                {"view-vendors"},
		"/payment.PaymentService/GetFromAccountOptions":        {"view-accounts"},
		
		// Shortcut operations
		"/payment.PaymentService/CreatePaymentShortcut":        {"create-payment"},
		"/payment.PaymentService/ExecutePaymentShortcut":       {"create-payment"},
		
		// Bank letter operations
		"/payment.PaymentService/CreateBankLetterFromNotes":    {"create-bank-letter"},
		"/payment.PaymentService/ProcessBankLetterLog":         {"approve-bank-letter"},
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		"/payment.PaymentService/TestPaymentAPI",
		"/payment.PaymentService/GeneratePaymentSerialNumber",
		"/payment.PaymentService/GetQueueSnapshot",
		"/payment.PaymentService/AddRequestToQueue",
		"/payment.PaymentService/RemoveRequestFromQueue",
	}
}
