package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	paymentpb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentpb"
	grpcadapter "github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/adapters/grpc"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/adapters/repository"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/config"
	"github.com/ShristiRnr/Nhit-Note/services/payment-service/internal/core/services"
)

func main() {
	// Load configuration from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "shristi")
	dbName := getEnv("DB_NAME", "nhit_db")
	grpcPort := getEnv("GRPC_PORT", "50064")

	// Connect to database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

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
	log.Println("✅ Connected to database")

	// Initialize layers
	repo := repository.NewPaymentRepository(db)
	service := services.NewPaymentService(repo)
	handler := grpcadapter.NewPaymentHandler(service)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register service
	paymentpb.RegisterPaymentServiceServer(grpcServer, handler)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("🚀 Payment Service listening on port %s", grpcPort)
	log.Printf("📝 Permissions configured for %d endpoints", len(config.GetPermissionMap()))
	log.Printf("🔓 Public endpoints: %d", len(config.GetPublicMethods()))


	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
