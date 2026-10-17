package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Payment Note CRUD operations
		"/paymentnote.PaymentNoteService/CreatePaymentNote":         {"create-payment-note"},
		"/paymentnote.PaymentNoteService/UpdatePaymentNote":         {"edit-payment-note"},
		"/paymentnote.PaymentNoteService/DeletePaymentNote":         {"delete-payment-note"},
		"/paymentnote.PaymentNoteService/ListPaymentNotes":          {"view-payment-notes"},
		"/paymentnote.PaymentNoteService/ListDraftPaymentNotes":     {"view-payment-notes"},
		"/paymentnote.PaymentNoteService/GetPaymentNote":            {"view-payment-notes"},
		
		// Draft operations
		"/paymentnote.PaymentNoteService/DeleteDraftPaymentNote":    {"delete-payment-note"},
		"/paymentnote.PaymentNoteService/ConvertDraftToActive":      {"approve-payment-note"},
		
		// Hold operations
		"/paymentnote.PaymentNoteService/PutPaymentNoteOnHold":      {"hold-payment-note"},
		"/paymentnote.PaymentNoteService/RemovePaymentNoteFromHold": {"hold-payment-note"},
		
		// UTR operations
		"/paymentnote.PaymentNoteService/UpdatePaymentNoteUtr":      {"update-payment-utr"},
		
		// Admin operations
		"/paymentnote.PaymentNoteService/CreatePaymentNoteForSuperAdmin": {}, // No permission check (SUPER_ADMIN only)
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		"/paymentnote.PaymentNoteService/TestPaymentNoteAPI",
		"/paymentnote.PaymentNoteService/GeneratePaymentNoteOrderNumber",
	}
}
