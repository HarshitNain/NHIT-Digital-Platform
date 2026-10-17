package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	SessionID    uuid.UUID
	UserID       uuid.UUID
	SessionToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token     string
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        uuid.UUID // Primary ID for the record
	Token     uuid.UUID // For link-based resets
	OTP       string    // For OTP-based resets
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	Used      bool   // Whether this reset token/OTP has been used
	ResetType string // "token" or "otp"
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	Token     uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// LoginRequest represents a login attempt
type LoginRequest struct {
	Email    string
	Password string
	TenantID uuid.UUID
	OrgID    *uuid.UUID
}

// LoginResponse represents a successful login
type LoginResponse struct {
	Token            string
	RefreshToken     string
	UserID           uuid.UUID
	Email            string
	Name             string
	Roles            []string
	Permissions      []string
	LastLoginAt      time.Time
	LastLoginIP      string
	TenantID         uuid.UUID
	OrgID            *uuid.UUID
	TokenExpiresAt   int64
	RefreshExpiresAt int64
	SessionID        uuid.UUID // Added for tracking specific session
}

// TokenValidation represents a validated token
type TokenValidation struct {
	Valid       bool
	UserID      uuid.UUID
	Email       string
	Name        string
	TenantID    uuid.UUID
	OrgID       *uuid.UUID
	Roles       []string
	Permissions []string
	ExpiresAt   time.Time
}
