package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()

	// Get configuration from environment
	port := os.Getenv("DEPARTMENT_SERVICE_PORT")
	if port == "" {
		port = "50054"
	}
	
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	}

	l.Info("Starting Department Service", "port", port)

	// Connect to database with optimized pool configuration
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
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

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		l.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	l.Info("Database connection established", "max_conns", 20, "min_conns", 5)

	// Initialize sqlc queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repositories
	departmentRepo := repository.NewDepartmentRepository(queries)

	// Initialize services
	departmentService := services.NewDepartmentService(departmentRepo, l)

	// Initialize gRPC handlers
	departmentHandler := grpcHandler.NewDepartmentHandler(departmentService, l)

	// Connect to auth-service for RBAC
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "localhost:50052"
	}
	authConn, err := grpc.Dial(authServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to auth service", "url", authServiceURL, "error", err)
		os.Exit(1)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	l.Info("Connected to Auth Service", "url", authServiceURL)

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
	departmentpb.RegisterDepartmentServiceServer(grpcServer, departmentHandler)

	// Start listening
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		l.Error("Failed to listen", "port", port, "error", err)
		os.Exit(1)
	}

	l.Info("Department Service listening", "port", port)
	if err := grpcServer.Serve(lis); err != nil {
		l.Error("Failed to serve gRPC", "error", err)
		os.Exit(1)
	}
}
