package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters"
	grpcAdapter "github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/storage"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort = ":50058"
	httpPort = ":8084"
)

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()
	ctx := context.Background()

	// Database connection with optimized pool configuration
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	l.Info("Connecting to database", "url", dbURL)
	
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		l.Error("Failed to parse database config", "error", err)
		os.Exit(1)
	}

	// ✅ Optimized connection pool limits
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		l.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		l.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	l.Info("Connected to PostgreSQL database", "max_conns", 5, "min_conns", 2)

	// Initialize dependencies
	publisher := adapters.NewNoOpEventPublisher()
	serviceConfig := ports.VendorServiceConfig{
		EnableCodeGeneration: true,
		DefaultVendorType:    "EXTERNAL",
		MaxAccountsPerVendor: 10,
	}

	// Initialize repository
	vendorRepo := repository.NewVendorRepository(pool, l)
	l.Info("Repository initialized")

	// Initialize service
	vendorService := services.NewVendorService(vendorRepo, l, publisher, serviceConfig)
	l.Info("Vendor service initialized")

	// Initialize MinIO client for vendor documents
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "192.168.1.22:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "admin"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "admin123"
	}
	minioBucket := os.Getenv("MINIO_BUCKET_VENDORS")
	if minioBucket == "" {
		minioBucket = "nhit.vendors"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, useSSL, l)
	if err != nil {
		l.Warn("Failed to initialize MinIO client for vendors", "error", err)
	}

	// Initialize gRPC handler
	vendorHandler := grpcAdapter.NewVendorGRPCHandler(vendorService, minioClient, l)
	l.Info("gRPC handler initialized")

	// Start gRPC server
	go func() {
		if err := startGRPCServer(vendorHandler, l); err != nil {
			l.Error("Failed to start gRPC server", "error", err)
			os.Exit(1)
		}
	}()

	// Start HTTP gateway server
	go func() {
		if err := startHTTPGateway(ctx); err != nil {
			l.Error("Failed to start HTTP gateway", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down servers...")
}

func startGRPCServer(vendorHandler vendorpb.VendorServiceServer, l *slog.Logger) error {
	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	// Connect to auth-service for RBAC
	authConn, err := grpcLib.Dial("localhost:50052", grpcLib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Warn("Failed to connect to auth-service, RBAC disabled", "error", err)
		// Continue without RBAC if auth service is down
		server := grpcLib.NewServer()
		vendorpb.RegisterVendorServiceServer(server, vendorHandler)
		reflection.Register(server)
		l.Info("gRPC server listening", "port", grpcPort, "rbac", "disabled")
		return server.Serve(listener)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	l.Info("Connected to Auth Service")

	// Initialize RBAC interceptor
	rbac := middleware.NewRBACInterceptor(authClient)
	for method, perms := range config.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	for _, method := range config.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	l.Info("RBAC interceptor initialized")

	// Create server with RBAC
	server := grpcLib.NewServer(
		grpcLib.UnaryInterceptor(rbac.UnaryServerInterceptor()),
	)
	
	// Register vendor handler
	vendorpb.RegisterVendorServiceServer(server, vendorHandler)

	// Enable reflection for development
	reflection.Register(server)

	l.Info("gRPC server listening", "port", grpcPort, "rbac", "enabled")
	return server.Serve(listener)
}

func startHTTPGateway(ctx context.Context) error {
	mux := runtime.NewServeMux()

	opts := []grpcLib.DialOption{grpcLib.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost%s", grpcPort)

	err := vendorpb.RegisterVendorServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Add CORS middleware
	handler := corsMiddleware(mux)

	return http.ListenAndServe(httpPort, handler)
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
