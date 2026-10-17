package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/grpc"
	httpHandler "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/http"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository"
	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/services"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/storage"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()

	// Get configuration from environment
	grpcPort := os.Getenv("USER_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("USER_SERVICE_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	}

	l.Info("Starting User Service", "grpcPort", grpcPort, "httpPort", httpPort)

	// Connect to database using pgxpool with optimized configuration
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		l.Error("Failed to parse database config", "error", err)
		os.Exit(1)
	}

	// Optimized connection pool limits for horizontal scaling
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
	l.Info("Database connection established", "max_conns", 30, "min_conns", 10)

	// Initialize sqlc queries with local SQLC
	queries := sqlc.New(pool)

	// Initialize repositories
	userRepo := repository.NewUserRepository(queries, pool)
	tenantRepo := repository.NewTenantRepository(queries)
	userRoleRepo := repository.NewUserRoleRepository(queries)
	roleRepo := repository.NewRoleRepository(queries)
	permissionRepo := repository.NewPermissionRepository(queries)
	loginHistoryRepo := repository.NewLoginHistoryRepository(queries)
	activityLogRepo := repository.NewActivityLogRepository(queries)

	// Initialize services
	userService := services.NewUserService(userRepo, tenantRepo, userRoleRepo, roleRepo, permissionRepo, loginHistoryRepo, activityLogRepo, l)

	// Connect to Auth Service for token validation / context
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

	// Connect to Department Service
	departmentServiceURL := os.Getenv("DEPARTMENT_SERVICE_URL")
	if departmentServiceURL == "" {
		departmentServiceURL = "localhost:50054"
	}
	deptConn, err := grpc.Dial(departmentServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to department service", "url", departmentServiceURL, "error", err)
		os.Exit(1)
	}
	defer deptConn.Close()
	l.Info("Connected to Department Service", "url", departmentServiceURL)

	// Connect to Designation Service
	designationServiceURL := os.Getenv("DESIGNATION_SERVICE_URL")
	if designationServiceURL == "" {
		designationServiceURL = "localhost:50055"
	}
	desigConn, err := grpc.Dial(designationServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to designation service", "url", designationServiceURL, "error", err)
		os.Exit(1)
	}
	defer desigConn.Close()
	l.Info("Connected to Designation Service", "url", designationServiceURL)
	
	// Initialize MinIO client for signatures
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
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "nhit.users"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	
	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, useSSL, l)
	if err != nil {
		l.Warn("Failed to initialize MinIO client", "error", err)
	}

	// Initialize handlers
	userGrpcHandler := grpcHandler.NewUserHandler(userService, pool, authClient, deptConn, desigConn, minioClient, l)
	tenantHttpHandler := httpHandler.NewTenantHTTPHandler(userService, l)

	// Initialize RBAC interceptor for gRPC
	rbacInterceptor := middleware.NewRBACInterceptor(authClient)
	
	// Register public methods
	for _, method := range config.GetPublicMethods() {
		rbacInterceptor.RegisterPublicMethod(method)
	}

	// Register permissions
	for method, perms := range config.GetPermissionMap() {
		rbacInterceptor.RegisterPermissions(method, perms)
	}
	
	// Start gRPC server in goroutine
	go func() {
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(rbacInterceptor.UnaryServerInterceptor()),
		)
		userpb.RegisterUserManagementServer(grpcServer, userGrpcHandler)

		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			l.Error("Failed to listen on gRPC port", "port", grpcPort, "error", err)
			os.Exit(1)
		}

		l.Info("gRPC Server listening", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			l.Error("Failed to serve gRPC", "error", err)
			os.Exit(1)
		}
	}()

	// Setup HTTP server with tenant endpoints
	router := mux.NewRouter()
	tenantHttpHandler.SetupRoutes(router)

	userHttpHandler := httpHandler.NewUserHTTPHandler(userService, l)
	userHttpHandler.RegisterRoutes(router)

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	l.Info("HTTP Server listening", "port", httpPort)
	if err := http.ListenAndServe(":"+httpPort, router); err != nil {
		l.Error("Failed to serve HTTP", "error", err)
		os.Exit(1)
	}
}
