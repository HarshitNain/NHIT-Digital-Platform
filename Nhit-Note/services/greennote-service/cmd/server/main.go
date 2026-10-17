package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	approvalpb "github.com/ShristiRnr/Nhit-Note/api/pb/approvalpb"
	eventsadapter "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/events"
	grpcadapter "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/grpcadapter"
	memoryrepo "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/repository/memory"
	sqlcrepo "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/repository/sqlc"
	storageadapter "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/storage"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/services"

	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"

	sharedmiddleware "github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/config"
	gnmiddleware "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds runtime configuration for the GreenNote service.
type Config struct {
	GRPCPort string
	HTTPPort string

	KafkaBrokers       []string
	KafkaTopicApproved string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	DatabaseURL         string
	AuthServiceEndpoint string
	ApprovalServiceEndpoint string
}

func loadConfigFromEnv() Config {
	cfg := Config{
		GRPCPort:            getenvWithDefault("GREENNOTE_GRPC_PORT", "50059"),
		HTTPPort:            getenvWithDefault("GREENNOTE_HTTP_PORT", "8086"),
		KafkaTopicApproved:  getenvWithDefault("GREENNOTE_KAFKA_TOPIC_APPROVED", "greennote.approved"),
		MinIOEndpoint:       os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey:      os.Getenv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:      os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucket:         getenvWithDefault("MINIO_BUCKET_GREENNOTE", "nhit.greennote"),
		DatabaseURL:         strings.TrimSpace(getenvWithDefault("GREENNOTE_DATABASE_URL", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")),
		AuthServiceEndpoint: getenvWithDefault("AUTH_SERVICE_ENDPOINT", "localhost:50052"),
		ApprovalServiceEndpoint: getenvWithDefault("APPROVAL_SERVICE_ENDPOINT", "localhost:50060"),
	}

	if brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if useSSL := strings.TrimSpace(os.Getenv("MINIO_USE_SSL")); useSSL != "" {
		b, err := strconv.ParseBool(useSSL)
		if err == nil {
			cfg.MinIOUseSSL = b
		}
	}

	return cfg
}

func getenvWithDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func main() {
	// Load .env file (try local first, then shared fallback)
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	log.Println("Loaded environment variables (local preferred)")

	cfg := loadConfigFromEnv()
	log.Printf("CONFIG: ApprovalEndpoint=%s", cfg.ApprovalServiceEndpoint)
	log.Printf("CONFIG: AuthEndpoint=%s", cfg.AuthServiceEndpoint)
	log.Printf("CONFIG: DB=%s", cfg.DatabaseURL)
	log.Printf("CONFIG: KafkaBrokers=%v (len=%d)", cfg.KafkaBrokers, len(cfg.KafkaBrokers))
	log.Printf("CONFIG: MinIOEndpoint=%s, Bucket=%s, UseSSL=%v", cfg.MinIOEndpoint, cfg.MinIOBucket, cfg.MinIOUseSSL)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Set up document storage (MinIO or in-memory fallback).
	docStorage, err := storageadapter.NewDocumentStorage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatalf("failed to initialize document storage: %v", err)
	}

	// Choose repository implementation: Postgres/sqlc when GREENNOTE_DATABASE_URL is set,
	// otherwise fall back to in-memory repository (useful for local development/tests).
	var repo ports.GreenNoteRepository
	if cfg.DatabaseURL != "" {
		db, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to open Postgres connection: %v", err)
		}
		// Improved connection pooling configuration
		db.SetMaxIdleConns(2)           // Increased from 5
		db.SetMaxOpenConns(5)           // Increased from 20
		db.SetConnMaxLifetime(time.Hour)
		db.SetConnMaxIdleTime(30 * time.Minute)

		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("failed to ping Postgres: %v", err)
		}

		repo = sqlcrepo.NewPostgresGreenNoteRepository(db, docStorage)
		log.Println("using Postgres/sqlc GreenNote repository (Manual Pool: Max=30, Min=10)")
	} else {
		repo = memoryrepo.NewInMemoryGreenNoteRepository(docStorage)
		log.Println("using in-memory GreenNote repository")
	}

	// Connect to User Service
	userConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())) // default user port
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserManagementClient(userConn)

	// Set up Kafka event publisher (or noop if brokers not configured).
	events := eventsadapter.NewEventPublisher(cfg.KafkaBrokers, cfg.KafkaTopicApproved)

	// Connect to Project Service
	projectConn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials())) // default project port
	if err != nil {
		log.Fatalf("failed to connect to project service: %v", err)
	}
	defer projectConn.Close()
	projectClient := projectpb.NewProjectServiceClient(projectConn)

	// Connect to Vendor Service
	vendorConn, err := grpc.Dial("localhost:50058", grpc.WithTransportCredentials(insecure.NewCredentials())) // default vendor port
	if err != nil {
		log.Fatalf("failed to connect to vendor service: %v", err)
	}
	defer vendorConn.Close()
	vendorClient := vendorpb.NewVendorServiceClient(vendorConn)

	// Connect to Department Service
	deptConn, err := grpc.Dial("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials())) // default department port
	if err != nil {
		log.Fatalf("failed to connect to department service: %v", err)
	}
	defer deptConn.Close()
	deptClient := departmentpb.NewDepartmentServiceClient(deptConn)

	// Connect to Approval Service
	log.Printf("🔌 connecting to Approval Service at: %s", cfg.ApprovalServiceEndpoint)
	approvalConn, err := grpc.Dial(cfg.ApprovalServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to approval service: %v", err)
	}
	defer approvalConn.Close()
	approvalClient := approvalpb.NewApprovalServiceClient(approvalConn)

	// Core application service.
	appService := services.NewGreenNoteService(repo, events, projectClient, vendorClient, deptClient, approvalClient, docStorage)

	// gRPC service adapter.
	grpcSvc := grpcadapter.NewGreenNoteGRPCServer(appService)

	grpcErrCh := make(chan error, 1)
	httpErrCh := make(chan error, 1)

	go func() {
		grpcErrCh <- runGRPCServer(ctx, cfg, grpcSvc)
	}()

	// Start Kafka Consumer for Approval Events
	// Ideally, use a separate topic for approval events if different from 'approved' topic?
	// The user request said: "listening to the approval_events topic".
	// We didn't add that to config, let's use a hardcoded default or env var if needed.
	// For now, I'll assume a default "approval_events".
	approvalTopic := getenvWithDefault("KAFKA_TOPIC_APPROVAL_EVENTS", "approval_events")
	consumer := eventsadapter.NewKafkaConsumer(cfg.KafkaBrokers, approvalTopic, "greennote-service-approval-group", repo, events, userClient)
	if consumer != nil {
		go consumer.Start(ctx)
	} else {
		log.Println("Kafka Consumer not started (missing brokers or topic)")
	}

	go func() {
		httpErrCh <- runHTTPGatewayServer(ctx, cfg, grpcSvc, docStorage)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received, waiting for servers to stop...")
		select {
		case err := <-grpcErrCh:
			if err != nil {
				log.Printf("gRPC server error during shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			log.Println("gRPC server shutdown timeout")
		}
		select {
		case err := <-httpErrCh:
			if err != nil {
				log.Printf("HTTP gateway error during shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			log.Println("HTTP gateway shutdown timeout")
		}
	case err := <-grpcErrCh:
		log.Fatalf("gRPC server exited with error: %v", err)
	case err := <-httpErrCh:
		log.Fatalf("HTTP gateway exited with error: %v", err)
	}
}

func runGRPCServer(ctx context.Context, cfg Config, svc greennotepb.GreenNoteServiceServer) error {
	addr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Connect to Auth Service for RBAC
	authConn, err := grpc.Dial(cfg.AuthServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	// Initialize Shared RBAC Interceptor
	// Note: We use the shared middleware package from NHIT Backend
	rbacInterceptor := sharedmiddleware.NewRBACInterceptor(authClient)

	// Register Permissions
	for method, perms := range config.GetPermissionMap() {
		rbacInterceptor.RegisterPermissions(method, perms)
	}

	// Register Auth and RBAC Interceptor
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			gnmiddleware.UnaryAuthInterceptor,
			rbacInterceptor.UnaryServerInterceptor(),
		),
	)
	greennotepb.RegisterGreenNoteServiceServer(server, svc)

	go func() {
		<-ctx.Done()
		log.Println("stopping gRPC server...")
		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			log.Println("force-stopping gRPC server")
			server.Stop()
		}
	}()

	log.Printf("gRPC server listening on %s", addr)
	return server.Serve(lis)
}

func runHTTPGatewayServer(ctx context.Context, cfg Config, svc greennotepb.GreenNoteServiceServer, docs ports.DocumentStorage) error {
	gatewayMux := runtime.NewServeMux()

	if err := greennotepb.RegisterGreenNoteServiceHandlerServer(ctx, gatewayMux, svc); err != nil {
		return err
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/", gatewayMux)
	rootMux.Handle("/api/v1/green-notes/documents/upload", makeDocumentUploadHandler(docs))
	rootMux.Handle("/additional-document/upload", makeFixedSectionUploadHandler(docs, "additional", 5<<20))
	rootMux.Handle("/extension-document/upload", makeFixedSectionUploadHandler(docs, "extension", 10<<20))
	rootMux.Handle("/deviation-document/upload", makeFixedSectionUploadHandler(docs, "deviation", 10<<20))

	addr := ":" + cfg.HTTPPort
	httpServer := &http.Server{
		Addr:    addr,
		Handler: rootMux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP gateway shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP gateway listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func makeDocumentUploadHandler(docs ports.DocumentStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form; limit total memory usage.
		if err := r.ParseMultipartForm(12 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		noteIDStr := strings.TrimSpace(r.FormValue("note_id"))
		section := strings.ToLower(strings.TrimSpace(r.FormValue("section")))
		if noteIDStr == "" || section == "" {
			http.Error(w, "note_id and section are required", http.StatusBadRequest)
			return
		}
		noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
		if err != nil || noteID <= 0 {
			http.Error(w, "note_id must be a positive integer", http.StatusBadRequest)
			return
		}

		var maxBytes int64
		switch section {
		case "additional":
			maxBytes = 5 << 20 // 5 MB
		case "extension", "deviation":
			maxBytes = 10 << 20 // 10 MB
		default:
			http.Error(w, "invalid section", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limited := io.LimitReader(file, maxBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		if int64(len(data)) > maxBytes {
			http.Error(w, fmt.Sprintf("file too large; limit is %d bytes", maxBytes), http.StatusRequestEntityTooLarge)
			return
		}

		filename := header.Filename
		if filename == "" {
			filename = "upload.bin"
		}
		filename = sanitizeFilename(filename)
		objectName := fmt.Sprintf("greennotes/%d/%s/%d-%s", noteID, section, time.Now().UnixNano(), filename)
		contentType := header.Header.Get("Content-Type")

		if err := docs.Save(r.Context(), objectName, data, contentType); err != nil {
			http.Error(w, "failed to store document", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"success":     true,
			"message":     "document uploaded",
			"object_name": objectName,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func makeFixedSectionUploadHandler(docs ports.DocumentStorage, section string, maxBytes int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(12 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		noteIDStr := strings.TrimSpace(r.FormValue("note_id"))
		if noteIDStr == "" {
			http.Error(w, "note_id is required", http.StatusBadRequest)
			return
		}
		noteID, err := strconv.ParseInt(noteIDStr, 10, 64)
		if err != nil || noteID <= 0 {
			http.Error(w, "note_id must be a positive integer", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		limited := io.LimitReader(file, maxBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		if int64(len(data)) > maxBytes {
			http.Error(w, fmt.Sprintf("file too large; limit is %d bytes", maxBytes), http.StatusRequestEntityTooLarge)
			return
		}

		filename := header.Filename
		if filename == "" {
			filename = "upload.bin"
		}
		filename = sanitizeFilename(filename)
		objectName := fmt.Sprintf("greennotes/%d/%s/%d-%s", noteID, section, time.Now().UnixNano(), filename)
		contentType := header.Header.Get("Content-Type")

		if err := docs.Save(r.Context(), objectName, data, contentType); err != nil {
			http.Error(w, "failed to store document", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"success":     true,
			"message":     "document uploaded",
			"object_name": objectName,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}
