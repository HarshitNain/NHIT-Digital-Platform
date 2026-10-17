package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Authentication
	Register(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name, email, password string, roles []string) (*domain.LoginResponse, error)
	Login(ctx context.Context, email, password string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error)
	LoginGlobal(ctx context.Context, email, password string, orgID *uuid.UUID) (*domain.LoginResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*domain.TokenValidation, error)
	SwitchOrganization(ctx context.Context, userID uuid.UUID, newOrgID uuid.UUID, tenantID uuid.UUID) (*domain.LoginResponse, error)

	// SSO
	InitiateSSO(ctx context.Context, provider string) (string, error)
	CompleteSSO(ctx context.Context, provider, code string) (*domain.LoginResponse, error)

	// Email Verification
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID, token string) error

	// Password Reset - Token based
	ForgotPassword(ctx context.Context, email string) error
	ResetPasswordByToken(ctx context.Context, token, newPassword string) error

	// Password Reset - OTP based
	ForgotPasswordWithOTP(ctx context.Context, email string, tenantID uuid.UUID) error
	ForgotPasswordWithOTPByEmail(ctx context.Context, email string) error  // Fetches tenant_id from email
	VerifyOTPAndResetPassword(ctx context.Context, email, otp, newPassword string, tenantID uuid.UUID) error
	VerifyOTPAndResetPasswordByEmail(ctx context.Context, email, otp, newPassword string) error // Fetches tenant_id from email

	// Session Management
	InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
}
