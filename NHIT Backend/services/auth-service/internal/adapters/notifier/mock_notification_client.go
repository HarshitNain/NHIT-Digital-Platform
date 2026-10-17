package notifier

import (
	"context"
	"log/slog"
)

// MockNotificationClient provides a mock implementation for testing
type MockNotificationClient struct {
	logger *slog.Logger
}

// NewMockNotificationClient creates a new mock notification client
func NewMockNotificationClient(logger *slog.Logger) *MockNotificationClient {
	return &MockNotificationClient{
		logger: logger,
	}
}

// SendPasswordResetEmail sends a password reset email through the notification service
func (c *MockNotificationClient) SendPasswordResetEmail(ctx context.Context, email, name, resetToken string) error {
	c.logger.Info("MOCK: Sending password reset email", 
		"name", name, 
		"email", email, 
		"token", resetToken,
		"link", "http://localhost:3000/reset-password?token="+resetToken)
	return nil
}

// SendEmailUpdateNotification sends a notification about email issues
func (c *MockNotificationClient) SendEmailUpdateNotification(ctx context.Context, email, name string) error {
	c.logger.Info("MOCK: Sending email update notification", "name", name, "email", email)
	return nil
}

// SendOtpPasswordResetEmail sends a password reset email with an OTP code
func (c *MockNotificationClient) SendOtpPasswordResetEmail(ctx context.Context, email, name, otp string) error {
	c.logger.Info("MOCK: Sending OTP password reset email", "name", name, "email", email, "otp", otp)
	return nil
}
