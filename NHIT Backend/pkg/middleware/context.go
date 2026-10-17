package middleware

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserIDFromContext extracts user_id from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user_id not found in context")
	}
	return userID, nil
}

// GetTenantIDFromContext extracts tenant_id from context
func GetTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		return "", status.Error(codes.Unauthenticated, "tenant_id not found in context")
	}
	return tenantID, nil
}

// GetOrgIDFromContext extracts org_id from context (optional)
func GetOrgIDFromContext(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value("org_id").(string)
	return orgID, ok && orgID != ""
}

// GetUserEmailFromContext extracts email from context
func GetUserEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value("email").(string)
	if !ok || email == "" {
		return "", status.Error(codes.Unauthenticated, "email not found in context")
	}
	return email, nil
}

// GetUserNameFromContext extracts name from context
func GetUserNameFromContext(ctx context.Context) (string, error) {
	name, ok := ctx.Value("name").(string)
	if !ok || name == "" {
		return "", status.Error(codes.Unauthenticated, "name not found in context")
	}
	return name, nil
}
