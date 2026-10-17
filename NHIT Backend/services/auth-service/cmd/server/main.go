package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/notifier"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/organization"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	grpcMiddleware "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func parseDuration(s string, defaultDuration time.Duration, l *slog.Logger) time.Duration {
	if s == "" {
		return defaultDuration
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		if l != nil {
			l.Warn("Invalid duration, using default", "value", s, "default", defaultDuration)
		}
		return defaultDuration
	}
	return d
}

func parseInt(s string, defaultValue int32, l *slog.Logger) int32 {
	if s == "" {
		return defaultValue
	}
	var val int32
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		if l != nil {
			l.Warn("Invalid integer, using default", "value", s, "default", defaultValue)
		}
		return defaultValue
	}
	return val
}

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()

	// Load configuration from environment
	serviceName := "auth-service"
	serverPort := os.Getenv("AUTH_SERVICE_PORT")
	if serverPort == "" {
		serverPort = "50052"
	}

	ssoConfig := config.LoadSSOConfig()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	}

	l.Info("Starting service", "service", serviceName, "port", serverPort)

	// ============================================================
	// 1. VALIDATE JWT SECRET (Mandatory in Production)
	// ============================================================
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		l.Error("FATAL: JWT_SECRET environment variable must be set. Generate with: openssl rand -base64 32")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		l.Error("FATAL: JWT_SECRET must be at least 32 characters long", "length", len(jwtSecret))
		os.Exit(1)
	}
	l.Info("JWT secret loaded from environment")

	// ============================================================
	// 2. CONFIGURE DATABASE CONNECTION POOL (From Environment)
	// ============================================================
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		l.Error("Failed to parse database config", "error", err)
		os.Exit(1)
	}

	// Read pool configuration from environment with defaults
	maxConns := parseInt(os.Getenv("DB_MAX_CONNS"), 25, l)
	minConns := parseInt(os.Getenv("DB_MIN_CONNS"), 5, l)
	maxConnLifetime := parseDuration(os.Getenv("DB_MAX_CONN_LIFETIME"), time.Hour, l)
	healthCheckPeriod := parseDuration(os.Getenv("DB_HEALTH_CHECK_PERIOD"), 30*time.Second, l)

	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = maxConnLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = healthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		l.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		l.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	l.Info("Database connected", "max_conns", maxConns, "min_conns", minConns, "health_check", healthCheckPeriod)

	// ============================================================
	// 3. CONNECT TO DEPENDENT SERVICES
	// ============================================================
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50051"
	}

	userConn, err := grpcMiddleware.Dial(userServiceAddr, grpcMiddleware.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to User Service", "address", userServiceAddr, "error", err)
		os.Exit(1)
	}

	userServiceClient := userpb.NewUserManagementClient(userConn)
	l.Info("Connected to User Service", "address", userServiceAddr)

	orgServiceAddr := os.Getenv("ORGANIZATION_SERVICE_ADDR")
	if orgServiceAddr == "" {
		orgServiceAddr = "localhost:8082"
	}

	orgConn, err := grpcMiddleware.Dial(orgServiceAddr, grpcMiddleware.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to Organization Service", "address", orgServiceAddr, "error", err)
		os.Exit(1)
	}

	orgServiceClient := organizationpb.NewOrganizationServiceClient(orgConn)
	orgClientAdapter := organization.NewOrganizationClient(orgServiceClient)
	l.Info("Connected to Organization Service", "address", orgServiceAddr)

	// ============================================================
	// 4. INITIALIZE REPOSITORIES
	// ============================================================
	sessionRepo := repository.NewSessionRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	passwordResetRepo := repository.NewPasswordResetRepository(pool)
	emailVerificationRepo := repository.NewEmailVerificationRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	// ============================================================
	// 5. INITIALIZE JWT MANAGER
	// ============================================================
	accessTokenDuration := 60 * time.Minute
	refreshTokenDuration := 7 * 24 * time.Hour

	jwtManager := utils.NewJWTManager(jwtSecret, accessTokenDuration, refreshTokenDuration)

	// ============================================================
	// 6. INITIALIZE SERVICES
	// ============================================================
	emailService := utils.NewMockEmailService(l)

	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:50060"
	}
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:3000"
	}
	notificationClient := notifier.NewRealNotificationClient(notificationServiceURL, frontendBaseURL, l)
	l.Info("Notification client initialized", "url", notificationServiceURL, "frontend_url", frontendBaseURL)

	authService := services.NewAuthService(
		userRepo,
		sessionRepo,
		refreshTokenRepo,
		passwordResetRepo,
		emailVerificationRepo,
		jwtManager,
		emailService,
		nil,
		notificationClient,
		userServiceClient,
		orgClientAdapter,
		ssoConfig,
		l,
	)

	// ============================================================
	// 7. START SESSION CLEANUP SERVICE (Background Job)
	// ============================================================
	cleanupInterval := parseDuration(os.Getenv("SESSION_CLEANUP_INTERVAL"), time.Hour, l)
	sessionCleanupService := services.NewSessionCleanupService(pool, cleanupInterval, l)

	// Start cleanup service in background
	go sessionCleanupService.Start(ctx)

	// ============================================================
	// 8. INITIALIZE GRPC SERVER WITH MIDDLEWARE
	// ============================================================
	authInterceptor := middleware.NewAuthInterceptor(authService)

	for _, method := range config.GetPublicMethods() {
		authInterceptor.RegisterPublicMethod(method)
	}

	for method, perms := range config.GetPermissionMap() {
		authInterceptor.RegisterPermissions(method, perms)
	}

	grpcServer := grpcMiddleware.NewServer(
		grpcMiddleware.UnaryInterceptor(authInterceptor.Unary()),
		grpcMiddleware.StreamInterceptor(authInterceptor.Stream()),
	)

	// Register auth handler
	authHandler := grpc.NewAuthHandler(authService, orgClientAdapter, l)
	authpb.RegisterAuthServiceServer(grpcServer, authHandler)

	// Register health check handler
	healthHandler := grpc.NewHealthCheckHandler(pool, userConn, orgConn, authService)
	
	// Link health handler to auth handler for background health checks in refresh endpoint
	authHandler.SetHealthHandler(healthHandler)
	l.Info("Health check integrated into refresh endpoint")

	// ============================================================
	// 9. START GRPC SERVER
	// ============================================================
	lis, err := net.Listen("tcp", ":"+serverPort)
	if err != nil {
		l.Error("Failed to listen", "port", serverPort, "error", err)
		os.Exit(1)
	}

	l.Info("Auth Service is ready", "port", serverPort, "access_token_duration", accessTokenDuration, "refresh_token_duration", refreshTokenDuration, "cleanup_interval", cleanupInterval)

	// ============================================================
	// 10. GRACEFUL SHUTDOWN HANDLING
	// ============================================================
	// Create channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			l.Error("gRPC server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	l.Info("Received shutdown signal, initiating graceful shutdown")

	// Graceful shutdown sequence
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 1. Stop accepting new requests (graceful stop with timeout)
	l.Info("Stopping gRPC server gracefully")
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		l.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		l.Error("Graceful shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	// 2. Cancel context to stop background services (session cleanup)
	cancel()
	l.Info("Stopped session cleanup service")

	// 3. Close database connection pool
	pool.Close()
	l.Info("Closed database connections")

	// 4. Close gRPC client connections
	userConn.Close()
	orgConn.Close()
	l.Info("Closed gRPC client connections")

	l.Info("Server shutdown complete")
}
