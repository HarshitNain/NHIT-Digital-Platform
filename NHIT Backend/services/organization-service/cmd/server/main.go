package main

import (
	"context"
	"net"
	"os"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/logger"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	grpcHandler "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/grpc"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/kafka"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/adapters/repository"
	orgConfig "github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/config"
	"github.com/ShristiRnr/NHIT_Backend/services/shared/database"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// healthServer implements the gRPC health check protocol
type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

func main() {
	// Initialize enterprise logger
	l := logger.NewLogger()

	// Load configuration
	cfg := config.LoadConfig("organization-service")
	l.Info("Starting service", "service", cfg.ServiceName, "port", cfg.ServerPort)

	// Connect to database
	conn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		l.Error("Failed to connect to database", "url", cfg.DatabaseURL, "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	l.Info("Database connection established")

	// Initialize repositories (Adapters Layer)
	orgRepo := repository.NewOrganizationRepository(conn)
	l.Info("Repositories initialized")

	// Connect to Auth Service for token validation
	authConn, err := grpc.Dial(cfg.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error("Failed to connect to auth service", "url", cfg.AuthServiceURL, "error", err)
		os.Exit(1)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	l.Info("Connected to Auth Service", "url", cfg.AuthServiceURL)

	// Initialize Kafka publisher (real implementation)
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "organization.events"

	kafkaPublisher, err := kafka.NewRealKafkaPublisher(kafkaBrokers, kafkaTopic, l)
	if err != nil {
		l.Warn("Failed to initialize real Kafka publisher, falling back to mock", "error", err)
		kafkaPublisher = kafka.NewMockKafkaPublisher(l)
	} else {
		l.Info("Real Kafka publisher initialized")
	}
	defer kafkaPublisher.Close()

	// Initialize MinIO client for logos
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
	minioBucket := os.Getenv("MINIO_BUCKET_LOGOS")
	if minioBucket == "" {
		minioBucket = "nhit"
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, useSSL, l)
	if err != nil {
		l.Warn("Failed to initialize MinIO client for organizations", "error", err)
	}

	// Initialize gRPC handlers (Adapters Layer) and pass DB pool, auth client, kafka publisher and minio client
	orgHandler := grpcHandler.NewOrganizationHandler(orgRepo, conn, authClient, kafkaPublisher, minioClient, l)
	l.Info("gRPC handlers initialized")

	// Initialize RBAC interceptor
	rbac := middleware.NewRBACInterceptor(authClient)
	
	// Register permissions
	for method, perms := range orgConfig.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	
	// Register public methods
	for _, method := range orgConfig.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	l.Info("RBAC interceptor initialized with permissions")

	// Create gRPC server with RBAC interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rbac.UnaryServerInterceptor()),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Register organization service
	organizationpb.RegisterOrganizationServiceServer(grpcServer, orgHandler)
	l.Info("Organization service registered")

	// Register health check service
	healthServer := &healthServer{}
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	l.Info("Health check service registered")

	// Register reflection service (for tools like grpcurl)
	reflection.Register(grpcServer)
	l.Info("gRPC reflection registered")

	// Start listening
	lis, err := net.Listen("tcp", ":"+cfg.ServerPort)
	if err != nil {
		l.Error("Failed to listen on port", "port", cfg.ServerPort, "error", err)
		os.Exit(1)
	}

	l.Info("Organization Service listening", "port", cfg.ServerPort)
	l.Info("Service Architecture: Hexagonal (Ports & Adapters)")

	if err := grpcServer.Serve(lis); err != nil {
		l.Error("Failed to serve gRPC", "error", err)
		os.Exit(1)
	}
}
