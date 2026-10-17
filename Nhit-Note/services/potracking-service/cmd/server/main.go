package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	"github.com/ShristiRnr/Nhit-Note/api/pb/potrackingpb"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/adapters/events"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/adapters/grpc/handler"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/adapters/repository"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/services"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Load environment variables
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	grpcPort := getEnv("POTRACKING_GRPC_PORT", "50062")
	httpPort := getEnv("POTRACKING_HTTP_PORT", "8085")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Database connection pool
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("✅ Connected to database")

	// Initialize layers
	poRepo := repository.NewPOTrackingRepository(pool)
	poService := services.NewPOTrackingService(poRepo)
	poHandler := handler.NewPOTrackingHandler(poService)

	paymentRepo := repository.NewPaymentRepository(pool)
	paymentService := services.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	potrackingpb.RegisterPOTrackingServiceServer(grpcServer, poHandler)
	potrackingpb.RegisterPaymentServiceServer(grpcServer, paymentHandler)

	// Connect to Payment Note Service
	paymentNoteServiceEndpoint := getEnv("PAYMENTNOTE_SERVICE_ENDPOINT", "localhost:50056")
	log.Printf("🔌 Connecting to Payment Note Service at: %s", paymentNoteServiceEndpoint)
	pnConn, err := grpc.Dial(paymentNoteServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("⚠️ Warning: Failed to connect to Payment Note service: %v", err)
	} else {
		defer pnConn.Close()
	}
	var paymentNoteClient paymentnotepb.PaymentNoteServiceClient
	if pnConn != nil {
		paymentNoteClient = paymentnotepb.NewPaymentNoteServiceClient(pnConn)
	}

	// Start Kafka Consumer for GreenNote Approvals
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	// Consumer 1: GreenNote Approvals
	gnConsumer := events.NewKafkaConsumer(
		kafkaBrokers,
		getEnv("GREENNOTE_KAFKA_TOPIC_APPROVED", "greennote.approved"),
		"potracking-gn-group",
		poService,
		paymentNoteClient,
	)
	if gnConsumer != nil {
		go gnConsumer.Start(ctx)
		log.Println("✅ Kafka consumer started for GreenNote approvals")
	}

	// Consumer 2: Approval Status Changes (General)
	approvalConsumer := events.NewKafkaConsumer(
		kafkaBrokers,
		getEnv("KAFKA_TOPIC_APPROVAL_EVENTS", "approval_events"),
		"potracking-approval-group",
		poService,
		paymentNoteClient,
	)
	if approvalConsumer != nil {
		go approvalConsumer.StartApprovalTracking(ctx)
		log.Println("✅ Kafka consumer started for Approval tracking")
	}

	// Start servers
	grpcErrCh := make(chan error, 1)
	httpErrCh := make(chan error, 1)

	go func() {
		grpcErrCh <- runGRPCServer(ctx, grpcPort, grpcServer)
	}()

	go func() {
		httpErrCh <- runHTTPGatewayServer(ctx, httpPort, grpcPort)
	}()

	log.Printf("🚀 PO Tracking Service started (gRPC: %s, HTTP: %s)", grpcPort, httpPort)

	select {
	case <-ctx.Done():
		log.Println("Shutting down servers...")
	case err := <-grpcErrCh:
		log.Fatalf("gRPC server error: %v", err)
	case err := <-httpErrCh:
		log.Fatalf("HTTP gateway error: %v", err)
	}
}

func runGRPCServer(ctx context.Context, port string, server *grpc.Server) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	return server.Serve(lis)
}

func runHTTPGatewayServer(ctx context.Context, httpPort, grpcPort string) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	_ = opts // Keep for future use if needed or just remove it
	// Actually better just remove it or blank it

	// Gateway registration removed as potracking.pb.gw.go is not generated
	/*
	// Register PO Tracking Service Gateway
	err := potrackingpb.RegisterPOTrackingServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		return err
	}

	// Register Payment Service Gateway
	err = potrackingpb.RegisterPaymentServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		return err
	}
	*/

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
