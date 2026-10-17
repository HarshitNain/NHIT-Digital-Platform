package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuthInterceptor extracts claims from the JWT in the Authorization header
// and populates the context metadata with user_id, tenant_id, and org_id.
func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization token format")
	}

	// Parse JWT Payload (Part 2) without verification
	// We assume the gateway or upstream service has already verified the signature.
	claims, err := parseJWTPayload(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token payload: %v", err)
	}

	// Create new metadata with extracted values
	newMD := metadata.Join(md, metadata.Pairs(
		"user_id", getClaimString(claims, "sub"), // 'sub' is standard for user_id
		"tenant_id", getClaimString(claims, "tenant_id"),
		"org_id", getClaimString(claims, "org_id"),
		"email", getClaimString(claims, "email"),
		"name", getClaimString(claims, "name"),
	))

	// Handle Roles (array)
	if roles, ok := claims["roles"].([]interface{}); ok {
		for _, r := range roles {
			if rStr, ok := r.(string); ok {
				newMD.Append("roles", rStr)
			}
		}
	}

	// Create new context with updated metadata
	newCtx := metadata.NewIncomingContext(ctx, newMD)

	return handler(newCtx, req)
}

func parseJWTPayload(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payloadPart := parts[1]
	// Add padding if needed
	if l := len(payloadPart) % 4; l > 0 {
		payloadPart += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(payloadPart)
	if err != nil {
		// Try standard encoding if URL encoding fails
		decoded, err = base64.StdEncoding.DecodeString(payloadPart)
		if err != nil {
			return nil, fmt.Errorf("failed to decode payload: %w", err)
		}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return claims, nil
}

func getClaimString(claims map[string]interface{}, key string) string {
	if val, ok := claims[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
