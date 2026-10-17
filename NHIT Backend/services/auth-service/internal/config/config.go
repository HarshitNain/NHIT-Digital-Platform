package config

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

// SSOConfig holds the configuration for SSO providers
type SSOConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	AzureTenantID     string
	AzureClientID     string
	AzureClientSecret string
	AzureRedirectURL  string
}

// LoadSSOConfig loads SSO configuration from environment variables
func LoadSSOConfig() *SSOConfig {
	// Try loading from .env file directly since user is having trouble with terminal env vars
	loadEnvFile()

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	azureTenantID := os.Getenv("AZURE_TENANT_ID")
	azureClientID := os.Getenv("AZURE_CLIENT_ID")
	azureClientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	azureRedirectURL := os.Getenv("AZURE_REDIRECT_URL")

	slog.Info("Loading SSO Config", 
		"GoogleIDLen", len(googleClientID), 
		"AzureIDLen", len(azureClientID), 
		"AzureTenantLen", len(azureTenantID))
	if len(azureClientID) > 0 {
		slog.Debug("Azure Client ID preview", "preview", azureClientID[:min(4, len(azureClientID))]+"...")
	}
	if len(azureRedirectURL) > 0 {
		slog.Debug("Azure Redirect URL", "url", azureRedirectURL)
	}

	return &SSOConfig{
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		GoogleRedirectURL:  googleRedirectURL,

		AzureTenantID:     azureTenantID,
		AzureClientID:     azureClientID,
		AzureClientSecret: azureClientSecret,
		AzureRedirectURL:  azureRedirectURL,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		// .env file might not exist in production or if not created
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove potential quotes
			value = strings.Trim(value, `"'`)
			
			os.Setenv(key, value)
		}
	}
	slog.Info("Loaded configuration from .env file")
}
