package services

import (
	"fmt"
)

// sendPasswordResetEmailLegacy is a legacy method that uses the email service
func (s *authService) sendPasswordResetEmailLegacy(email, name, resetToken string) error {
	if err := s.emailService.SendPasswordResetEmail(email, name, resetToken); err != nil {
		s.logger.Warn("Failed to send password reset email", "error", err)
		// Send notification about email failure
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			s.logger.Warn("Failed to send email update notification", "error", err)
		}
		return fmt.Errorf("failed to send password reset email. Please update your email address")
	}
	return nil
}
