package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HealthCheckHandler provides health check functionality
type HealthCheckHandler struct {
	pool              *pgxpool.Pool
	userServiceConn   *grpc.ClientConn
	orgServiceConn    *grpc.ClientConn
	authService       ports.AuthService
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(
	pool *pgxpool.Pool,
	userServiceConn *grpc.ClientConn,
	orgServiceConn *grpc.ClientConn,
	authService ports.AuthService,
) *HealthCheckHandler {
	return &HealthCheckHandler{
		pool:            pool,
		userServiceConn: userServiceConn,
		orgServiceConn:  orgServiceConn,
		authService:     authService,
	}
}

// checkDatabase checks database connectivity
func (h *HealthCheckHandler) checkDatabase(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

// checkGRPCConnection checks if a gRPC connection is ready
func (h *HealthCheckHandler) checkGRPCConnection(conn *grpc.ClientConn) string {
	if conn == nil {
		return "error: connection is nil"
	}

	state := conn.GetState()
	if state.String() == "READY" {
		return "ok"
	}

	return fmt.Sprintf("not ready: %s", state.String())
}

// PerformQuickHealthCheck performs a lightweight health check of all critical components
func (h *HealthCheckHandler) PerformQuickHealthCheck(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	results := make(map[string]string)
	var errs []error

	// 1. Check Database
	if err := h.pool.Ping(ctx); err != nil {
		results["database"] = fmt.Sprintf("DOWN: %v", err)
		errs = append(errs, fmt.Errorf("database unhealthy: %v", err))
	} else {
		results["database"] = "OK"
	}

	// 2. Check User Service Connection
	userState := h.checkGRPCConnection(h.userServiceConn)
	results["user_service"] = userState
	if userState != "ok" {
		errs = append(errs, fmt.Errorf("user_service unhealthy: %s", userState))
	}

	// 3. Check Organization Service Connection
	orgState := h.checkGRPCConnection(h.orgServiceConn)
	results["organization_service"] = orgState
	if orgState != "ok" {
		errs = append(errs, fmt.Errorf("organization_service unhealthy: %s", orgState))
	}

	if len(errs) > 0 {
		return results, status.Errorf(codes.Unavailable, "service partially unhealthy: %v", errs)
	}

	return results, nil
}
