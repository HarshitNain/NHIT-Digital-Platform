package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	approvalpb "github.com/ShristiRnr/Nhit-Note/api/pb/approvalpb"
	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	potrackingpb "github.com/ShristiRnr/Nhit-Note/api/pb/potrackingpb"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/adapters/events"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/adapters/grpc/handler"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/adapters/repository"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/services"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "shristi")
	dbName := getEnv("DB_NAME", "nhit_db")
	grpcPort := getEnv("GRPC_PORT", "50056")

	// MinIO configuration
	minioEndpoint := getEnv("MINIO_ENDPOINT", "play.min.io:9443")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "")
	minioBucket := getEnv("MINIO_BUCKET", "payment-documents")
	minioUseSSL := getEnv("MINIO_USE_SSL", "true") == "true"

	// Database connection
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// ✅ Optimized connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)
	
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize MinIO client (optional - can be nil if not configured)
	var minioClient *storage.MinIOClient
	if minioAccessKey != "" && minioSecretKey != "" {
		minioClient, err = storage.NewMinIOClient(
			minioEndpoint,
			minioAccessKey,
			minioSecretKey,
			minioBucket,
			minioUseSSL,
		)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to initialize MinIO client: %v", err)
			log.Println("⚠️  Document upload will not be available")
		} else {
			log.Printf("✅ Connected to MinIO at %s (bucket: %s)", minioEndpoint, minioBucket)
		}
	} else {
		log.Println("⚠️  MinIO credentials not provided - document upload disabled")
	}

	// Initialize repository with MinIO client
	paymentNoteRepo := repository.NewPaymentNoteRepository(db, minioClient)
	log.Println("✅ Repository initialized")

	// Connect to Approval Service
	approvalServiceEndpoint := getEnv("APPROVAL_SERVICE_ENDPOINT", "localhost:50060")
	log.Printf("🔌 Connecting to Approval Service at: %s", approvalServiceEndpoint)
	approvalConn, err := grpc.Dial(approvalServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to approval service: %v", err)
	}
	defer approvalConn.Close()
	approvalClient := approvalpb.NewApprovalServiceClient(approvalConn)

	// Connect to PO Tracking Service
	potrackingServiceEndpoint := getEnv("POTRACKING_SERVICE_ENDPOINT", "localhost:50062")
	log.Printf("🔌 Connecting to PO Tracking Service at: %s", potrackingServiceEndpoint)
	poConn, err := grpc.Dial(potrackingServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️ Warning: Failed to connect to PO Tracking service: %v", err)
	} else {
		defer poConn.Close()
	}
	var potrackingClient potrackingpb.POTrackingServiceClient
	if poConn != nil {
		potrackingClient = potrackingpb.NewPOTrackingServiceClient(poConn)
	}

	// Connect to Green Note Service
	greennoteServiceEndpoint := getEnv("GREENNOTE_SERVICE_ENDPOINT", "localhost:50059")
	log.Printf("🔌 Connecting to Green Note Service at: %s", greennoteServiceEndpoint)
	gnConn, err := grpc.Dial(greennoteServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️ Warning: Failed to connect to Green Note service: %v", err)
	} else {
		defer gnConn.Close()
	}
	var greenNoteClient greennotepb.GreenNoteServiceClient
	if gnConn != nil {
		greenNoteClient = greennotepb.NewGreenNoteServiceClient(gnConn)
	}

	// Initialize service
	paymentNoteService := services.NewPaymentNoteService(paymentNoteRepo, approvalClient, potrackingClient, greenNoteClient)
	log.Println("✅ Service layer initialized")

	// Start Kafka Consumer for GreenNote Approvals
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	ctx := context.Background()
	service := paymentNoteService // Alias for clarity in consumer initialization

	// Consumer 1: GreenNote Approvals
	gnConsumer := events.NewKafkaConsumer(
		kafkaBrokers,
		getEnv("GREENNOTE_KAFKA_TOPIC_APPROVED", "greennote.approved"),
		"paymentnote-gn-group",
		service,
	)
	if gnConsumer != nil {
		go gnConsumer.Start(ctx, gnConsumer.HandleGreenNoteApproved)
	}

	// Consumer 2: Approval Status Changes
	approvalConsumer := events.NewKafkaConsumer(
		kafkaBrokers,
		getEnv("APPROVAL_KAFKA_TOPIC_STATUS_CHANGED", "approval_events"),
		"paymentnote-approval-group",
		service,
	)
	if approvalConsumer != nil {
		go approvalConsumer.Start(ctx, approvalConsumer.HandleApprovalStatusChanged)
	}

	// Initialize gRPC handler
	paymentNoteHandler := handler.NewPaymentNoteHandler(paymentNoteService)
	log.Println("✅ gRPC handler initialized")

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register payment note service
	paymentnotepb.RegisterPaymentNoteServiceServer(grpcServer, paymentNoteHandler)
	log.Println("✅ Payment Note Service registered")

	// Register reflection service for grpcurl
	reflection.Register(grpcServer)
	log.Println("✅ gRPC reflection registered")

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("\n🛑 Shutting down gracefully...")
		grpcServer.GracefulStop()
		db.Close()
		log.Println("✅ Server stopped")
		os.Exit(0)
	}()

	log.Printf("🚀 Payment Note Service listening on port %s", grpcPort)
	log.Println("=====================================")
	log.Println("Service Features:")
	log.Println("- Payment Note CRUD operations")
	log.Println("- Financial calculations (TDS, Net Payable, Number-to-Words)")
	if minioClient != nil {
		log.Println("- Document upload/download (MinIO)")
	}
	log.Println("- Draft workflow management")
	log.Println("- Hold/Unhold operations")
	log.Println("- UTR tracking")
	log.Println("- Approval logs & comments")
	log.Println("=====================================")
	log.Printf("Test with: grpcurl -plaintext localhost:%s paymentnote.PaymentNoteService/GeneratePaymentNoteOrderNumber", grpcPort)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
