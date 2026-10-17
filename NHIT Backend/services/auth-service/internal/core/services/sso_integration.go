package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// SSOProvider defines the interface for different SSO providers
type SSOProvider interface {
	GetAuthURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*SSOUserInfo, error)
}

// SSOUserInfo holds normalized user info from providers
type SSOUserInfo struct {
	Email string
	Name  string
	ID    string
}

// GoogleProvider implements SSOProvider for Google
type GoogleProvider struct {
	config *oauth2.Config
	logger *slog.Logger
}

func NewGoogleProvider(cfg *config.SSOConfig, logger *slog.Logger) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
		logger: logger,
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*SSOUserInfo, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: status %s", resp.Status)
	}

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &SSOUserInfo{
		Email: googleUser.Email,
		Name:  googleUser.Name,
		ID:    googleUser.ID,
	}, nil
}

// MicrosoftProvider implements SSOProvider for Microsoft (Azure AD)
type MicrosoftProvider struct {
	config *oauth2.Config
	logger *slog.Logger
}

func NewMicrosoftProvider(cfg *config.SSOConfig, logger *slog.Logger) *MicrosoftProvider {
	endpoint := microsoft.AzureADEndpoint(cfg.AzureTenantID)
	// If tenant ID is "common" or not specified, use common endpoint
	if cfg.AzureTenantID == "" {
		endpoint = microsoft.AzureADEndpoint("common")
	}

	logger.Info("Initializing Microsoft Provider",
		"ClientID", cfg.AzureClientID,
		"Tenant", cfg.AzureTenantID,
		"Redirect", cfg.AzureRedirectURL)

	provider := &MicrosoftProvider{
		config: &oauth2.Config{
			ClientID:     cfg.AzureClientID,
			ClientSecret: cfg.AzureClientSecret,
			RedirectURL:  cfg.AzureRedirectURL,
			Scopes:       []string{"https://graph.microsoft.com/User.Read"},
			Endpoint:     endpoint,
		},
		logger: logger,
	}
	// Force AuthStyle to InParams (send client_id/secret in body) as required by Azure AD
	// oauth2.AuthStyleInParams
	provider.config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	
	return provider
}

func (p *MicrosoftProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *MicrosoftProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	// Explicitly pass client_id and client_secret as options to ensure they are in the body
	// This overrides any auto-detection issues
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("client_id", p.config.ClientID),
		oauth2.SetAuthURLParam("client_secret", p.config.ClientSecret),
	}
	return p.config.Exchange(ctx, code, opts...)
}

func (p *MicrosoftProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*SSOUserInfo, error) {
	client := p.config.Client(ctx, token)
	// Graph API endpoint for user profile
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Log body for debugging
		// bodyBytes, _ := io.ReadAll(resp.Body)
		// log.Printf("Microsoft Graph Error: %s", string(bodyBytes))
		return nil, fmt.Errorf("failed to fetch user info: status %s", resp.Status)
	}

	var msUser struct {
		ID    string `json:"id"`
		Email string `json:"mail"`
		UPN   string `json:"userPrincipalName"`
		Name  string `json:"displayName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&msUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Microsoft sometimes puts email in mail or userPrincipalName
	if msUser.Email == "" && msUser.UPN != "" {
		msUser.Email = msUser.UPN
	}

	if msUser.Email == "" {
		p.logger.Warn("Microsoft user info has no email", "ID", msUser.ID, "Name", msUser.Name)
	}

	return &SSOUserInfo{
		Email: msUser.Email,
		Name:  msUser.Name,
		ID:    msUser.ID,
	}, nil
}
