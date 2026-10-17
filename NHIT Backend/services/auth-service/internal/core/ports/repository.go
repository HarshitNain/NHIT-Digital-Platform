package ports

import (
	"context"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/google/uuid"
)

// SessionRepository defines the interface for session operations
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (*domain.Session, error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
	UpdateExpiresAt(ctx context.Context, token string, expiresAt time.Time) error
}

// PasswordResetRepository defines the interface for password reset operations
type PasswordResetRepository interface {
	// Token-based reset methods
	Create(ctx context.Context, userID uuid.UUID, token uuid.UUID, expiresAt time.Time) (*domain.PasswordReset, error)
	GetByToken(ctx context.Context, token uuid.UUID) (*domain.PasswordReset, error)
	Delete(ctx context.Context, token uuid.UUID) error

	// OTP-based reset methods
	CreateWithOTP(ctx context.Context, userID uuid.UUID, otp string, expiresAt time.Time) (*domain.PasswordReset, error)
	GetByUserIDAndOTP(ctx context.Context, userID uuid.UUID, otp string) (*domain.PasswordReset, error)
}

// EmailVerificationRepository defines the interface for email verification operations
type EmailVerificationRepository interface {
	Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*domain.EmailVerificationToken, error)
	Verify(ctx context.Context, userID uuid.UUID, token uuid.UUID) (bool, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *UserData) (*UserData, error)
	// GetByEmail gets user by email (tenant-specific)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*UserData, error)

	// GetByEmailGlobal gets user by email across all tenants
	GetByEmailGlobal(ctx context.Context, email string) (*UserData, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*UserData, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error
	UpdateLastLogout(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
	UpdateTenantPassword(ctx context.Context, email string, hashedPassword string) error
	UpdateOrganizationSuperAdminPassword(ctx context.Context, email string, hashedPassword string) error
}

// UserData represents user data needed for authentication
type UserData struct {
	UserID          uuid.UUID
	TenantID        uuid.UUID
	Email           string
	Name            string
	Password        string
	EmailVerifiedAt *time.Time
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
