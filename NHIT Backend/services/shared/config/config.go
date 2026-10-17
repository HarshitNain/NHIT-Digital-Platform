package config

import (
	"log"
	"os"
	"time"
)

// Config holds all configuration for microservices
type Config struct {
	ServiceName        string
	ServerPort         string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	
	// Service discovery
	UserServiceURL string
	AuthServiceURL string
	OrgServiceURL  string
}

// LoadConfig reads environment variables and returns a Config struct
func LoadConfig(serviceName string) *Config {
	return &Config{
		ServiceName:        serviceName,
		ServerPort:         getEnv("SERVER_PORT", "8082"),
		DatabaseURL:        getEnv("DB_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpiry:          getEnvAsDuration("JWT_EXPIRY", 15*time.Minute),
		RefreshTokenExpiry: getEnvAsDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		
		UserServiceURL: getEnv("USER_SERVICE_URL", "localhost:50051"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "localhost:50052"),
		OrgServiceURL:  getEnv("ORG_SERVICE_URL", "localhost:8082"),
	}
}

// getEnv reads an environment variable or returns a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsDuration reads an env variable as time.Duration or returns fallback
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return fallback
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("invalid duration for %s: %v, using fallback %v", key, err, fallback)
		return fallback
	}
	return value
}
