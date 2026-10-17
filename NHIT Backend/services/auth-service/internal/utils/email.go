package utils

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// EmailService interface for sending emails
type EmailService interface {
	SendVerificationEmail(to, name, token string) error
	SendPasswordResetEmail(to, name, token string) error
	SendEmailUpdateNotification(to, name string) error
	SendWelcomeEmail(to, name string, data map[string]interface{}) error
}

// MockEmailService is a mock implementation for development/testing
type MockEmailService struct {
	SentEmails []SentEmail
	logger     *slog.Logger
}

// SentEmail represents an email that was sent
type SentEmail struct {
	To        string
	Subject   string
	Body      string
	Timestamp time.Time
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService(logger *slog.Logger) *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]SentEmail, 0),
		logger:     logger,
	}
}

// SendVerificationEmail sends a verification email (mock)
func (s *MockEmailService) SendVerificationEmail(to, name, token string) error {
	subject := "Verify Your Email - NHIT"
	body := fmt.Sprintf(`
Hello %s,

Thank you for registering with NHIT!

Please verify your email address by using the following verification code:

Verification Token: %s

This token will expire in 24 hours.

If you did not create an account, please ignore this email.

Best regards,
NHIT Team
`, name, token)

	s.SentEmails = append(s.SentEmails, SentEmail{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
	})

	s.logger.Info("MOCK EMAIL: Verification email sent", "to", to, "token", token)
	return nil
}

// SendPasswordResetEmail sends a password reset email (mock)
func (s *MockEmailService) SendPasswordResetEmail(to, name, token string) error {
	subject := "Password Reset Request - NHIT"
	body := fmt.Sprintf(`
Hello %s,

We received a request to reset your password for your NHIT account.

Please use the following token to reset your password:

Reset Token: %s

This token will expire in 1 hour.

If you did not request a password reset, please ignore this email and your password will remain unchanged.

Best regards,
NHIT Team
`, name, token)

	s.SentEmails = append(s.SentEmails, SentEmail{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
	})

	s.logger.Info("MOCK EMAIL: Password reset email sent", "to", to, "token", token)
	return nil
}

// SendEmailUpdateNotification sends a notification about email delivery failure
func (s *MockEmailService) SendEmailUpdateNotification(to, name string) error {
	subject := "Email Delivery Failed - Update Required"
	body := fmt.Sprintf(`
Hello %s,

We were unable to deliver important notifications to your registered email address: %s

This could be because:
- The email address is incorrect
- The email service is blocking our messages
- The mailbox is full

Please update your email address in your account settings to continue receiving important notifications.

To update your email:
1. Log in to your account
2. Go to Settings > Profile
3. Update your email address
4. Verify the new email address

Best regards,
NHIT Team
`, name, to)

	s.SentEmails = append(s.SentEmails, SentEmail{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
	})

	s.logger.Info("MOCK EMAIL: Email update notification sent", "to", to)
	return nil
}

// GenerateVerificationToken generates a verification token
func GenerateVerificationToken() string {
	return uuid.New().String()
}

// GeneratePasswordResetToken generates a password reset token
func GeneratePasswordResetToken() string {
	return uuid.New().String()
}

// SendWelcomeEmail sends a welcome email to new users (mock)
func (s *MockEmailService) SendWelcomeEmail(to, name string, data map[string]interface{}) error {
	subject := "Welcome to NHIT! 🎉"
	
	orgName := "your organization"
	isSuperAdmin := false
	
	if val, ok := data["organization"]; ok {
		orgName = val.(string)
	}
	if val, ok := data["is_super_admin"]; ok {
		isSuperAdmin = val.(bool)
	}
	
	roleInfo := ""
	if isSuperAdmin {
		roleInfo = "\n\nYou have been assigned as the Super Administrator of " + orgName + ". You have full access to manage all aspects of your organization including:\n- Adding and managing users\n- Creating departments and designations\n- Managing projects\n- Configuring organization settings"
	}
	
	body := fmt.Sprintf(`
Hello %s,

Welcome to NHIT! 🚀

Your account has been successfully created. We're excited to have you on board!

Organization: %s%s

You can now log in to your account and start exploring all the features.

If you have any questions or need assistance, feel free to reach out to our support team.

Best regards,
NHIT Team
`, name, orgName, roleInfo)
	
	s.SentEmails = append(s.SentEmails, SentEmail{
		To:        to,
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now(),
	})
	
	s.logger.Info("MOCK EMAIL: Welcome email sent", "to", to, "organization", orgName, "is_super_admin", isSuperAdmin)
	return nil
}
