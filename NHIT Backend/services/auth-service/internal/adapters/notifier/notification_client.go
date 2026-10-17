package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// RealNotificationClient provides an HTTP-based implementation that calls the notification service
type RealNotificationClient struct {
	BaseURL         string
	FrontendBaseURL string
	HTTPClient      *http.Client
	logger          *slog.Logger
}

// NewRealNotificationClient creates a new notification client that communicates via HTTP
func NewRealNotificationClient(baseURL, frontendBaseURL string, logger *slog.Logger) *RealNotificationClient {
	return &RealNotificationClient{
		BaseURL:         baseURL,
		FrontendBaseURL: frontendBaseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	To           string                 `json:"to"`
	TemplateName string                 `json:"template_name"`
	Data         map[string]interface{} `json:"data"`
}

// SendPasswordResetEmail sends a password reset email through the notification service
func (c *RealNotificationClient) SendPasswordResetEmail(ctx context.Context, email, name, resetToken string) error {
	// Create reset link - uses configured frontend base URL
	// Note: Including full API path as default for backend-first testing
	resetLink := fmt.Sprintf("%s/api/v1/auth/reset-password?otp=%s", c.FrontendBaseURL, resetToken)

	// Create current year for template
	currentYear := fmt.Sprintf("%d", time.Now().Year())

	// Build notification data
	data := map[string]interface{}{
		"name":         name,
		"reset_link":   resetLink,
		"current_year": currentYear,
	}

	// Build request payload
	req := SendNotificationRequest{
		To:           email,
		TemplateName: "password-reset",
		Data:         data,
	}

	// Send the request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/notifications", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned non-OK status: %d", resp.StatusCode)
	}

	c.logger.Info("Password reset email sent", "email", email)
	return nil
}

// SendEmailUpdateNotification sends a notification about email issues
func (c *RealNotificationClient) SendEmailUpdateNotification(ctx context.Context, email, name string) error {
	// Build notification data
	data := map[string]interface{}{
		"name":  name,
		"email": email,
	}

	// Build request payload
	req := SendNotificationRequest{
		To:           email,
		TemplateName: "email-update-required", // This would be another template in the notification service
		Data:         data,
	}

	// Send the request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/notifications", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

// SendOtpPasswordResetEmail sends a password reset email with an OTP code
func (c *RealNotificationClient) SendOtpPasswordResetEmail(ctx context.Context, email, name, otp string) error {
	// Create current year for template
	currentYear := fmt.Sprintf("%d", time.Now().Year())

	// Build notification data
	data := map[string]interface{}{
		"name":         name,
		"otp":          otp,
		"current_year": currentYear,
	}

	// Build request payload
	req := SendNotificationRequest{
		To:           email,
		TemplateName: "otp-reset-password",
		Data:         data,
	}

	// Send the request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/notifications", c.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned non-OK status: %d", resp.StatusCode)
	}

	c.logger.Info("Password reset OTP email sent", "email", email)
	return nil
}
