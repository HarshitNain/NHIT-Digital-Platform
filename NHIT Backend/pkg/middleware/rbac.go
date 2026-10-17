package middleware

import (
	"context"
	"fmt"
	"strings"
	"os"
	"time"


	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RBACInterceptor provides role-based access control for gRPC services
type RBACInterceptor struct {
	authClient    authpb.AuthServiceClient
	permissionMap map[string][]string
	publicMethods map[string]bool
}

// NewRBACInterceptor creates a new RBAC interceptor
func NewRBACInterceptor(authClient authpb.AuthServiceClient) *RBACInterceptor {
	return &RBACInterceptor{
		authClient:    authClient,
		permissionMap: make(map[string][]string),
		publicMethods: make(map[string]bool),
	}
}

// RegisterPermissions registers required permissions for a method
func (i *RBACInterceptor) RegisterPermissions(method string, permissions []string) {
	i.permissionMap[method] = permissions
}

// RegisterPublicMethod registers a method that doesn't require authentication
func (i *RBACInterceptor) RegisterPublicMethod(method string) {
	i.publicMethods[method] = true
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for RBAC
func (i *RBACInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Debug logging to see actual method path
		fmt.Printf("DEBUG RBAC: Method=%s, IsPublic=%v\n", info.FullMethod, i.publicMethods[info.FullMethod])
		
		// Check if method is public
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		
		// Temporary file logging for Windows debugging
		f, _ := os.OpenFile("rbac_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			var keys []string
			for k := range md {
				keys = append(keys, k)
			}
			fmt.Fprintf(f, "[%s] RBAC: Method=%s, MetadataPresent=%v, Keys=%v\n", time.Now().Format(time.RFC3339), info.FullMethod, ok, keys)
			f.Close()
		}

		if !ok {
			fmt.Printf("DEBUG RBAC: Metadata NOT provided\n")
			return nil, status.Error(codes.Unauthenticated, "metadata not provided")
		}

		// LOG ALL METADATA KEYS FOR DEBUGGING
		var mdKeys []string
		for k := range md {
			mdKeys = append(mdKeys, k)
		}
		fmt.Printf("DEBUG RBAC: Received Metadata Keys: %v\n", mdKeys)

		if auth := md.Get("authorization"); len(auth) > 0 {
			fmt.Printf("DEBUG RBAC: Authorization header found (len: %d)\n", len(auth[0]))
		} else {
			fmt.Printf("DEBUG RBAC: Authorization header MISSING. FULL MD: %v\n", md)
		}

		token := extractToken(md)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "authorization token not provided")
		}

		// Validate token via auth-service
		validation, err := i.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{
			Token: token,
		})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
		}
		if !validation.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Check if user is SUPER_ADMIN (bypass permission checks)
		isSuperAdmin := false
		for _, role := range validation.Roles {
			if role == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}

		// Check permissions if not super admin
		if !isSuperAdmin {
			requiredPerms, hasPermissions := i.permissionMap[info.FullMethod]
			if hasPermissions && len(requiredPerms) > 0 {
				if !i.hasRequiredPermissions(validation.Permissions, requiredPerms) {
					return nil, status.Errorf(codes.PermissionDenied, 
						"insufficient permissions. Required: %v", requiredPerms)
				}
			}
		}

		// Add user context
		ctx = i.addUserContext(ctx, validation)

		// Call handler
		return handler(ctx, req)
	}
}

// extractToken extracts the bearer token from metadata
func extractToken(md metadata.MD) string {
	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}

	token := values[0]
	if strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}
	return token
}

// hasRequiredPermissions checks if user has at least one of the required permissions
func (i *RBACInterceptor) hasRequiredPermissions(userPerms, requiredPerms []string) bool {
	permSet := make(map[string]bool)
	for _, p := range userPerms {
		permSet[p] = true
	}

	for _, req := range requiredPerms {
		if permSet[req] {
			return true // User has this permission
		}
	}

	return false // User doesn't have any required permission
}

// addUserContext adds user information to context
func (i *RBACInterceptor) addUserContext(ctx context.Context, validation *authpb.ValidateTokenResponse) context.Context {
	ctx = context.WithValue(ctx, "user_id", validation.UserId)
	ctx = context.WithValue(ctx, "email", validation.Email)
	ctx = context.WithValue(ctx, "name", validation.Name)
	ctx = context.WithValue(ctx, "tenant_id", validation.TenantId)
	
	if validation.OrgId != "" {
		ctx = context.WithValue(ctx, "org_id", validation.OrgId)
	}
	
	// Store roles and permissions in context for additional checks
	ctx = context.WithValue(ctx, "roles", validation.Roles)
	ctx = context.WithValue(ctx, "permissions", validation.Permissions)
	
	return ctx
}

// LogPermissionDenial logs permission denial for audit purposes
func (i *RBACInterceptor) LogPermissionDenial(method, userID string, requiredPerms []string) {
	fmt.Printf("⚠️  Permission denied: User %s attempted %s, required: %v\n", 
		userID, method, requiredPerms)
}
