package middleware

import (
	"context"
	"strings"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor validates JWT tokens and enforces RBAC
type AuthInterceptor struct {
	authService   ports.AuthService
	publicMethods map[string]bool
	permissionMap map[string][]string
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(authService ports.AuthService) *AuthInterceptor {
	// Methods that don't require authentication
	// We will load these from config in main, but keeping default here as fallback or initialization
	publicMethods := map[string]bool{
		"/AuthService/RegisterUser":           true,
		"/AuthService/Login":                  true,
		"/AuthService/ForgotPassword":         true,
		"/AuthService/ResetPasswordByToken":   true,
		"/AuthService/SendPasswordResetEmail": true,
		"/AuthService/ValidateToken":          true,
	}

	return &AuthInterceptor{
		authService:   authService,
		publicMethods: publicMethods,
		permissionMap: make(map[string][]string),
	}
}

// RegisterPermissions registers required permissions for a method
func (interceptor *AuthInterceptor) RegisterPermissions(method string, permissions []string) {
	interceptor.permissionMap[method] = permissions
}

// RegisterPublicMethod registers a method that doesn't require authentication
func (interceptor *AuthInterceptor) RegisterPublicMethod(method string) {
	interceptor.publicMethods[method] = true
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if method is public
		if interceptor.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := values[0]
		// Remove "Bearer " prefix if present
		if strings.HasPrefix(accessToken, "Bearer ") {
			accessToken = strings.TrimPrefix(accessToken, "Bearer ")
		}

		// Validate token
		validation, err := interceptor.authService.ValidateToken(ctx, accessToken)
		if err != nil || !validation.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
		}

		// Add user info to context
		ctx = context.WithValue(ctx, "user_id", validation.UserID.String())
		ctx = context.WithValue(ctx, "email", validation.Email)
		ctx = context.WithValue(ctx, "tenant_id", validation.TenantID.String())
		ctx = context.WithValue(ctx, "roles", validation.Roles)
		ctx = context.WithValue(ctx, "permissions", validation.Permissions)

		// Check RBAC permissions
		if !interceptor.publicMethods[info.FullMethod] {
			requiredPerms, hasEntry := interceptor.permissionMap[info.FullMethod]
			if hasEntry && len(requiredPerms) > 0 {
				// Check if user has any of the required permissions
				hasPermission := false
				
				// Bypass for SUPER_ADMIN
				for _, role := range validation.Roles {
					if role == "SUPER_ADMIN" {
						hasPermission = true
						break
					}
				}

				if !hasPermission {
					userPermsMap := make(map[string]bool)
					for _, p := range validation.Permissions {
						userPermsMap[p] = true
					}

					for _, required := range requiredPerms {
						if userPermsMap[required] {
							hasPermission = true
							break
						}
					}
				}

				if !hasPermission {
					return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
				}
			}
		}

		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Check if method is public
		if interceptor.publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := values[0]
		if strings.HasPrefix(accessToken, "Bearer ") {
			accessToken = strings.TrimPrefix(accessToken, "Bearer ")
		}

		// Validate token
		validation, err := interceptor.authService.ValidateToken(stream.Context(), accessToken)
		if err != nil || !validation.Valid {
			return status.Errorf(codes.Unauthenticated, "invalid or expired token")
		}

		return handler(srv, stream)
	}
}
