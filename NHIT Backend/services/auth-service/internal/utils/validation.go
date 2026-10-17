package utils

import (
	"regexp"
	"strings"
)

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// SanitizeString removes potentially harmful characters
func SanitizeString(input string) string {
	// Remove leading/trailing whitespace
	sanitized := strings.TrimSpace(input)
	
	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	
	return sanitized
}
