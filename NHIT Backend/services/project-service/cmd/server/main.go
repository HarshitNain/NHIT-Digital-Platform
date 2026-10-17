package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/grpc/handler"
	kafkaAdapter "github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/kafka"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/adapters/repository"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/services"
)

func main() {
	// Load configuration from environment variables
	grpcPort := getEnv("GRPC_PORT", "50057")
	httpPort := getEnv("HTTP_PORT", "8057")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")

	// Connect to database with connection pooling
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse database config: %v", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = 5               // Max connections for project-service
	poolConfig.MinConns = 2               // Warm connections
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL with pgxpool (Max: 20, Min: 5)")

	// Initialize repository
	projectRepo := repository.NewProjectRepository(pool)

	// Initialize Kafka consumer (dynamic configuration)
	kafkaBrokersStr := getEnv("KAFKA_BROKERS", "127.0.0.1:9092")
	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")
	kafkaTopic := getEnv("KAFKA_TOPIC", "organization.events")
	kafkaGroupID := getEnv("KAFKA_GROUP_ID", "project-service-group")

	log.Printf("📡 Kafka Config: Brokers=%v, Topic=%s, GroupID=%s", kafkaBrokers, kafkaTopic, kafkaGroupID)

	kafkaConsumer, err := kafkaAdapter.NewRealKafkaConsumer(kafkaBrokers, kafkaTopic, kafkaGroupID, nil)
	if err != nil {
		log.Printf("⚠️ Failed to initialize real Kafka consumer, falling back to mock: %v", err)
		kafkaConsumer = kafkaAdapter.NewMockKafkaConsumer(nil)
	} else {
		log.Println("✅ Real Kafka consumer initialized")
	}
	defer kafkaConsumer.Close()

	// Initialize service
	projectService := services.NewProjectService(projectRepo, kafkaConsumer, nil)

	// Initialize gRPC handler
	projectHandler := handler.NewProjectHandler(projectService)

	// Start event consumer in background
	go func() {
		if err := projectService.StartEventConsumer(context.Background()); err != nil {
			log.Printf("⚠️ Event consumer error: %v", err)
		}
	}()

	log.Println("✅ Service components initialized")

	// Start gRPC server
	go func() {
		if err := startGRPCServer(grpcPort, projectHandler); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP gateway
	go func() {
		if err := startHTTPGateway(httpPort, grpcPort); err != nil {
			log.Fatalf("Failed to start HTTP gateway: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Project Service...")
}

func startGRPCServer(port string, handler pb.ProjectServiceServer) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Connect to auth-service for RBAC
	authConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️  Failed to connect to auth-service: %v (RBAC disabled)", err)
		// Continue without RBAC if auth service down
		grpcServer := grpc.NewServer()
		pb.RegisterProjectServiceServer(grpcServer, handler)
		reflection.Register(grpcServer)
		log.Printf("🚀 gRPC server listening on port %s", port)
		return grpcServer.Serve(lis)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)
	log.Println("✅ Connected to Auth Service")

	// Initialize RBAC
	rbac := middleware.NewRBACInterceptor(authClient)
	for method, perms := range config.GetPermissionMap() {
		rbac.RegisterPermissions(method, perms)
	}
	for _, method := range config.GetPublicMethods() {
		rbac.RegisterPublicMethod(method)
	}
	log.Println("✅ RBAC interceptor initialized")

	// Create server with RBAC
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(rbac.UnaryServerInterceptor()),
	)
	pb.RegisterProjectServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	log.Printf("🚀 gRPC server listening on port %s with RBAC", port)
	return grpcServer.Serve(lis)
}

func startHTTPGateway(httpPort, grpcPort string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb.RegisterProjectServiceHandlerFromEndpoint(
		ctx,
		mux,
		"localhost:"+grpcPort,
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	server := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🌐 HTTP gateway listening on port %s", httpPort)
	return server.ListenAndServe()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
