package ports

import (
	"context"
)

// NotificationClient defines the interface for sending notifications
type NotificationClient interface {
	// SendPasswordResetEmail sends a password reset email with a token link
	SendPasswordResetEmail(ctx context.Context, email, name, resetToken string) error

	// SendEmailUpdateNotification sends a notification about email issues
	SendEmailUpdateNotification(ctx context.Context, email, name string) error

	// SendOtpPasswordResetEmail sends a password reset email with an OTP code
	SendOtpPasswordResetEmail(ctx context.Context, email, name, otp string) error
}
