package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	TenantID    string   `json:"tenant_id"`
	OrgID       string   `json:"org_id,omitempty"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// GenerateAccessToken generates a new access token
func (m *JWTManager) GenerateAccessToken(userID, email, name, tenantID, orgID string, roles, permissions []string) (string, int64, error) {
	expiresAt := time.Now().UTC().Add(m.accessTokenDuration)
	
	claims := JWTClaims{
		UserID:      userID,
		Email:       email,
		Name:        name,
		TenantID:    tenantID,
		OrgID:       orgID,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "nhit-auth-service",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// GenerateRefreshToken generates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID, tenantID, orgID string) (string, int64, error) {
	fmt.Printf("DEBUG: GenerateRefreshToken called with UserID=%s, TenantID=%s, OrgID=%s\n", userID, tenantID, orgID)
	expiresAt := time.Now().UTC().Add(m.refreshTokenDuration)
	
	// Use the same claims structure as access token but focusing on identity
	claims := JWTClaims{
		UserID:   userID,
		TenantID: tenantID,
		OrgID:    orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "nhit-auth-service",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	// Check if token is expired (with 10s leeway for clock skew)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now().UTC().Add(-10*time.Second)) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is expired (with 10s leeway for clock skew)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now().UTC().Add(-10*time.Second)) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	return claims, nil
}

// ExtractTokenFromBearer extracts the token from "Bearer <token>" format
func ExtractTokenFromBearer(bearerToken string) (string, error) {
	if len(bearerToken) < 7 || bearerToken[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return bearerToken[7:], nil
}

// GenerateToken is a wrapper for GenerateAccessToken for backward compatibility
func (m *JWTManager) GenerateToken(userID uuid.UUID, email, name string, tenantID uuid.UUID, orgID *uuid.UUID, roles, permissions []string) (string, error) {
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}
	
	token, _, err := m.GenerateAccessToken(userID.String(), email, name, tenantID.String(), orgIDStr, roles, permissions)
	return token, err
}
