package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()

	// Configuration
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	port := getEnv("PORT", "50055")

	l.Info("Starting Designation Service", "port", port)

	// Connect to database with optimized pool configuration
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		l.Error("Failed to parse database config", "error", err)
		os.Exit(1)
	}

	// Optimized connection pool limits
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

	if err := pool.Ping(ctx); err != nil {
		l.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	l.Info("Database connection established", "max_conns", 20, "min_conns", 5)

	// Initialize SQLC queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repository
	designationRepo := repository.NewDesignationRepository(queries)

	// Initialize service
	designationService := services.NewDesignationService(designationRepo, l)

	// Initialize gRPC handler
	designationHandler := grpcHandler.NewDesignationHandler(designationService, l)

	// Connect to auth-service for RBAC
	authURL := getEnv("AUTH_SERVICE_URL", "localhost:50052")
	authConn, err := grpc.Dial(authURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to auth-service", "url", authURL, "error", err)
		os.Exit(1)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	l.Info("Connected to Auth Service", "url", authURL)

	// Initialize RBAC interceptor
	rbac := middleware.NewRBACInterceptor(authClient)
	for method, perms := range config.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	for _, method := range config.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	l.Info("RBAC interceptor initialized")

	// Create gRPC server with RBAC
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rbac.UnaryServerInterceptor()),
	)
	designationpb.RegisterDesignationServiceServer(grpcServer, designationHandler)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		l.Error("Failed to listen", "port", port, "error", err)
		os.Exit(1)
	}

	l.Info("Designation Service listening", "port", port)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		l.Info("Shutting down gracefully")
		grpcServer.GracefulStop()
	}()

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		l.Error("Failed to serve gRPC", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
