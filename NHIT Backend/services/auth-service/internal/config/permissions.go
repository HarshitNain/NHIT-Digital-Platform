package config

// GetPermissionMap returns the permission requirements for each RPC method
func GetPermissionMap() map[string][]string {
	return map[string][]string{
		// Organization Switching
		"/AuthService/SwitchOrganization": {"switch-organizations"},
		
		// User Management (if exposed via Auth Service, though usually in User Service)
		// "/AuthService/Logout": {"logout"}, // Logout usually just requires being logged in, no specific permission
	}
}

// GetPublicMethods returns methods that don't require authentication
func GetPublicMethods() []string {
	return []string{
		"/AuthService/RegisterUser",
		"/AuthService/Login",
		"/AuthService/ForgotPassword",
		"/AuthService/ResetPasswordByToken",
		"/AuthService/SendPasswordResetEmail",
		"/AuthService/ValidateToken",
		"/AuthService/VerifyEmail",
		"/AuthService/InitiateSSO",
		"/AuthService/CompleteSSO",
		"/AuthService/InitiateSSOLogout",
		"/AuthService/CompleteSSOLogout",
		"/AuthService/ForgotPasswordWithOTP",
		"/AuthService/VerifyOTPAndResetPassword",
		"/AuthService/SendVerificationEmail",
		"/AuthService/RefreshToken",
	}
}
