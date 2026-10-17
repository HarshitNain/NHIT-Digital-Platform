package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"fmt"
	"time"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/handler"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/middleware"
	"github.com/ShristiRnr/NHIT_Backend/services/api-gateway/internal/storage"
	"golang.org/x/time/rate"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// metadataAnnotator extracts JWT claims and adds them to gRPC metadata
func metadataAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	
	// Extract Authorization header
	authHeader := req.Header.Get("Authorization")
	
	// Temporary file logging for Windows debugging
	f, _ := os.OpenFile("gateway_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] ANNOTATOR: Path=%s, AuthPresent=%v, Len=%d\n", time.Now().Format(time.RFC3339), req.URL.Path, authHeader != "", len(authHeader))
		f.Close()
	}

	log.Printf("ANNOTATOR: Received Authorization header: %v (len: %d)", authHeader != "", len(authHeader))
	
	// Always try to extract tenant and org info from individual headers first
	// (common for refresh requests or when frontend doesn't send JWT)
	if tid := req.Header.Get("X-Tenant-ID"); tid != "" {
		md.Set("tenant_id", tid)
	} else if tid := req.Header.Get("Tenant-ID"); tid != "" {
		md.Set("tenant_id", tid)
	} else if tid := req.Header.Get("tenant_id"); tid != "" {
		md.Set("tenant_id", tid)
	}

	if oid := req.Header.Get("X-Org-ID"); oid != "" {
		md.Set("org_id", oid)
	} else if oid := req.Header.Get("Org-ID"); oid != "" {
		md.Set("org_id", oid)
	} else if oid := req.Header.Get("org_id"); oid != "" {
		md.Set("org_id", oid)
	}

	if authHeader == "" {
		// Log missing auth for monitoring
		log.Printf("ANNOTATOR: Authorization header MISSING for path %s. Meta set: %v", req.URL.Path, md)
		return md
	}
	
	// Explicitly set authorization in metadata so downstream services' RBAC interceptors can find it
	// NOTE: gRPC metadata keys are case-insensitive but usually normalized to lowercase
	md.Set("authorization", authHeader)
	log.Printf("ANNOTATOR: Set 'authorization' metadata")

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		log.Printf("ANNOTATOR: Warning: No 'Bearer ' prefix in token")
		return md 
	}

	// Parse JWT without verification (just to extract claims)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		log.Printf("ANNOTATOR: Failed to parse JWT for claims: %v", err)
		return md
	}

	// Extract User-Agent and add to metadata
	if ua := req.Header.Get("User-Agent"); ua != "" {
		md.Set("grpcgateway-user-agent", ua)
		md.Set("x-user-agent", ua)
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		log.Printf("ANNOTATOR: Extracted claims: user_id=%v, tenant_id=%v, org_id=%v", 
			claims["user_id"], claims["tenant_id"], claims["org_id"])
			
		// Add user_id to metadata
		if userID, exists := claims["user_id"]; exists {
			if userIDStr, ok := userID.(string); ok {
				md.Set("user_id", userIDStr)
			}
		}
		
		// Add tenant_id to metadata
		if tenantID, exists := claims["tenant_id"]; exists {
			if tenantIDStr, ok := tenantID.(string); ok {
				md.Set("tenant_id", tenantIDStr)
			}
		}
		
		// Add org_id to metadata
		if orgID, exists := claims["org_id"]; exists {
			if orgIDStr, ok := orgID.(string); ok {
				md.Set("org_id", orgIDStr)
			}
		}
	}

	log.Printf("ANNOTATOR: Returning metadata keys: %v", getKeys(map[string]interface{}(map[string]interface{}{}))) // Just to see keys
	// Print all keys actually in md
	var keys []string
	for k := range md {
		keys = append(keys, k)
	}
	log.Printf("ANNOTATOR: Final MD keys: %v", keys)

	return md
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// gRPC service endpoints
	userServiceEndpoint := getEnv("USER_SERVICE_URL", "localhost:50051")
	authServiceEndpoint := getEnv("AUTH_SERVICE_URL", "localhost:50052")
	organizationServiceEndpoint := getEnv("ORG_SERVICE_URL", "localhost:8082")
	departmentServiceEndpoint := getEnv("DEPT_SERVICE_URL", "localhost:50054")
	designationServiceEndpoint := getEnv("DESIGNATION_SERVICE_URL", "localhost:50055")
	vendorServiceEndpoint := getEnv("VENDOR_SERVICE_URL", "localhost:50058")
	projectServiceEndpoint := getEnv("PROJECT_SERVICE_URL", "localhost:50057")
	greennoteServiceEndpoint := getEnv("GREENNOTE_SERVICE_URL", "localhost:50059")
	paymentnoteServiceEndpoint := getEnv("PAYMENTNOTE_SERVICE_URL", "localhost:50056")

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	
	// Create gRPC Gateway mux
	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(metadataAnnotator),
	)

	// Register gRPC Services
	mustRegister(ctx, gwmux, userpb.RegisterUserManagementHandlerFromEndpoint, userServiceEndpoint, opts)
	mustRegister(ctx, gwmux, authpb.RegisterAuthServiceHandlerFromEndpoint, authServiceEndpoint, opts)
	mustRegister(ctx, gwmux, organizationpb.RegisterOrganizationServiceHandlerFromEndpoint, organizationServiceEndpoint, opts)
	mustRegister(ctx, gwmux, departmentpb.RegisterDepartmentServiceHandlerFromEndpoint, departmentServiceEndpoint, opts)
	mustRegister(ctx, gwmux, designationpb.RegisterDesignationServiceHandlerFromEndpoint, designationServiceEndpoint, opts)
	mustRegister(ctx, gwmux, vendorpb.RegisterVendorServiceHandlerFromEndpoint, vendorServiceEndpoint, opts)
	mustRegister(ctx, gwmux, projectpb.RegisterProjectServiceHandlerFromEndpoint, projectServiceEndpoint, opts)
	mustRegister(ctx, gwmux, greennotepb.RegisterGreenNoteServiceHandlerFromEndpoint, greennoteServiceEndpoint, opts)
	mustRegister(ctx, gwmux, paymentnotepb.RegisterPaymentNoteServiceHandlerFromEndpoint, paymentnoteServiceEndpoint, opts)

	// --- 1. Initialize MinIO Client ---
	minioEndpoint := getEnv("MINIO_ENDPOINT", "192.168.1.22:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "admin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "admin123")
	minioUseSSL := getEnv("MINIO_USE_SSL", "false") == "true"

	// Distinct bucket names
	orgBucket := getEnv("MINIO_BUCKET_ORG", "nhit")
	userBucket := getEnv("MINIO_BUCKET_USER", "nhit.users")
	vendorBucket := getEnv("MINIO_BUCKET_VENDOR", "nhit.vendors")
	gnBucket := getEnv("MINIO_BUCKET_GREENNOTE", "nhit.greennote")

	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, orgBucket, minioUseSSL)
	if err != nil {
		log.Printf("Warning: Failed to initialize MinIO client: %v", err)
	} else {
		log.Println("✅ MinIO client initialized")
		// Ensure all required buckets exist
		_ = minioClient.EnsureBucketExists(ctx, userBucket)
		_ = minioClient.EnsureBucketExists(ctx, vendorBucket)
		_ = minioClient.EnsureBucketExists(ctx, gnBucket)
	}
	// --- 3. Initialize Vendor Handler (Direct gRPC) ---
	vendorConn, err := grpc.NewClient(vendorServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial vendor service: %v", err)
	}
	defer vendorConn.Close()
	vendorClient := vendorpb.NewVendorServiceClient(vendorConn)
	vendorHandler := handler.NewVendorHandler(vendorClient, minioClient, vendorBucket)

	// --- 4. Initialize User Handler ---
	userConn, err := grpc.NewClient(userServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial user service: %v", err)
	}
	defer userConn.Close()
	userClient := userpb.NewUserManagementClient(userConn)
	userHandler := handler.NewUserHandler(userClient, minioClient, userBucket)

	// --- 5. Initialize Organization Handler ---
	orgConn, err := grpc.NewClient(organizationServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial organization service: %v", err)
	}
	defer orgConn.Close()
	orgClient := organizationpb.NewOrganizationServiceClient(orgConn)
	orgHandler := handler.NewOrganizationHandler(orgClient, minioClient, orgBucket)

	// --- 6. Initialize GreenNote Handler ---
	gnConn, err := grpc.NewClient(greennoteServiceEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial greennote service: %v", err)
	}
	defer gnConn.Close()
	gnClient := greennotepb.NewGreenNoteServiceClient(gnConn)
	gnHandler := handler.NewGreenNoteHandler(gnClient, minioClient)

	// --- 7. Initialize Rate Limiter ---
	// Default: 100 requests per second, burst of 200
	limiter := middleware.NewIPRateLimiter(rate.Limit(100), 200)

	// --- 8. Setup Root Mux for Custom Routes ---
	rootMux := http.NewServeMux()

	// POST /api/v1/vendors -> Custom Handler if Multipart
	rootMux.HandleFunc("/api/v1/vendors", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/vendors - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		if r.Method == http.MethodPost && isMultipart {
			vendorHandler.HandleCreateVendor(w, r)
			return
		}
		gwmux.ServeHTTP(w, r)
	})

	// PUT/PATCH /api/v1/vendors/{id} -> Custom Handler if Multipart
	rootMux.HandleFunc("/api/v1/vendors/", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/vendors/ - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		if (r.Method == http.MethodPut || r.Method == http.MethodPatch) && isMultipart {
			vendorHandler.HandleUpdateVendor(w, r)
			return
		}
		gwmux.ServeHTTP(w, r)
	})

	// POST /api/v1/users -> Custom Handler if Multipart (Creation)
	rootMux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/users - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		if r.Method == http.MethodPost && isMultipart {
			userHandler.HandleCreateUser(w, r)
			return
		}
		gwmux.ServeHTTP(w, r)
	})

	// PUT/PATCH /api/v1/users/{id} or POST /api/v1/users/{id}/signature
	rootMux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/users/ - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		// Detect if it's a signature upload vs general user update
		if strings.HasSuffix(r.URL.Path, "/signature") {
			if r.Method == http.MethodPost && isMultipart {
				userHandler.HandleUploadSignature(w, r)
				return
			}
		} else {
			// Regular PUT/PATCH update
			if (r.Method == http.MethodPut || r.Method == http.MethodPatch) && isMultipart {
				userHandler.HandleUpdateUser(w, r)
				return
			}
		}
		gwmux.ServeHTTP(w, r)
	})

	// POST /api/v1/organizations -> Custom Handler if Multipart (Creation)
	rootMux.HandleFunc("/api/v1/organizations", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/organizations - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		if r.Method == http.MethodPost && isMultipart {
			orgHandler.HandleCreateOrganization(w, r)
			return
		}
		gwmux.ServeHTTP(w, r)
	})

	// PUT /api/v1/organizations/{id} or POST /api/v1/organizations/{id}/logo
	rootMux.HandleFunc("/api/v1/organizations/", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/organizations/ - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		// Detect if it's a logo upload vs general org update
		if strings.HasSuffix(r.URL.Path, "/logo") {
			if r.Method == http.MethodPost && isMultipart {
				orgHandler.HandleUploadLogo(w, r)
				return
			}
		} else {
			// Regular PUT/PATCH update
			if (r.Method == http.MethodPut || r.Method == http.MethodPatch) && isMultipart {
				orgHandler.HandleUpdateOrganization(w, r)
				return
			}
		}
		gwmux.ServeHTTP(w, r)
	})

	// POST /api/v1/green-notes -> Custom Handler if Multipart
	rootMux.HandleFunc("/api/v1/green-notes", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/green-notes - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		if r.Method == http.MethodPost && isMultipart {
			gnHandler.HandleCreateGreenNote(w, r)
			return
		}
		gwmux.ServeHTTP(w, r)
	})

	// PUT /api/v1/green-notes/{id} or POST /api/v1/green-notes/{id}/documents
	rootMux.HandleFunc("/api/v1/green-notes/", func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		isMultipart := strings.Contains(contentType, "multipart/form-data")
		log.Printf("DEBUG: Request to /api/v1/green-notes/ - Method: %s, Multipart: %v, Content-Type: %s", r.Method, isMultipart, contentType)
		
		if strings.HasSuffix(r.URL.Path, "/documents") {
			if r.Method == http.MethodPost && isMultipart {
				gnHandler.HandleUploadDocuments(w, r)
				return
			}
		} else {
			if (r.Method == http.MethodPut || r.Method == http.MethodPatch) && isMultipart {
				gnHandler.HandleUpdateGreenNote(w, r)
				return
			}
		}
		gwmux.ServeHTTP(w, r)
	})

	// Default to gRPC Gateway (with query param normalization)
	rootMux.Handle("/", normalizeParams(gwmux))

	// Add Middleware chain
	handler := middleware.RateLimit(limiter)(cors(rootMux))

	// Start HTTP server
	port := ":8083"
	log.Printf("API Gateway listening on %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Helper to register simple handlers
func mustRegister(ctx context.Context, mux *runtime.ServeMux, registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error, endpoint string, opts []grpc.DialOption) {
	err := registerFunc(ctx, mux, endpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register gateway for %s: %v", endpoint, err)
	}
	log.Printf("✅ Registered gateway -> %s", endpoint)
}

// cors adds CORS headers to allow browser requests and handles authentication forwarding
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		requestHeaders := r.Header.Get("Access-Control-Request-Headers")
		if requestHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Tenant-ID, X-Org-ID, tenant_id, org_id")
		}
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log incoming requests for debugging
		log.Printf("📝 %s %s - Auth header present: %v", r.Method, r.URL.Path, r.Header.Get("Authorization") != "")

		// Forward request to handler
		h.ServeHTTP(w, r)
	})
}


func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getKeys returns the keys of a map for debugging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// normalizeParams converts camelCase query params to snake_case for gRPC Gateway compatibility
// and also normalizes JSON body fields from snake_case to camelCase
func normalizeParams(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		originalQuery := r.URL.RawQuery
		changed := false
		
		// Map 'pageSize' -> 'page_size'
		if val := q.Get("pageSize"); val != "" {
			if q.Get("page_size") == "" {
				q.Set("page_size", val)
				changed = true
			}
		}
		
		// Map 'pageNumber' -> 'page' (optional, but helpful)
		if val := q.Get("pageNumber"); val != "" && q.Get("page") == "" {
			q.Set("page", val)
			changed = true
		}

		// Map 'token' -> 'otp' for password reset links
		if val := q.Get("token"); val != "" && q.Get("otp") == "" {
			q.Set("otp", val)
			changed = true
		}

		if changed {
			newQuery := q.Encode()
			log.Printf("DEBUG Middleware: Normalized params. Original='%s', New='%s'", originalQuery, newQuery)
			r.URL.RawQuery = newQuery
		}
		
		// Normalize JSON body: convert snake_case to camelCase for common fields
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			contentType := strings.ToLower(r.Header.Get("Content-Type"))
			log.Printf("DEBUG normalizeParams: Method=%s, ContentType=%s, Path=%s", r.Method, contentType, r.URL.Path)
			
			if strings.Contains(contentType, "application/json") {
				body, err := io.ReadAll(r.Body)
				log.Printf("DEBUG normalizeParams: Read body, len=%d, err=%v", len(body), err)
				
				if err == nil && len(body) > 0 {
					r.Body.Close()
					
					// Parse JSON
					var data map[string]interface{}
					if err := json.Unmarshal(body, &data); err == nil {
						log.Printf("DEBUG normalizeParams: Parsed JSON, keys=%v", getKeys(data))
						modified := false
						
						// refresh_token -> refreshToken
						if val, exists := data["refresh_token"]; exists {
							log.Printf("DEBUG normalizeParams: Found refresh_token=%v", val != nil)
							if _, hasRefreshToken := data["refreshToken"]; !hasRefreshToken {
								data["refreshToken"] = val
								delete(data, "refresh_token")
								modified = true
								log.Printf("DEBUG normalizeParams: Converted refresh_token -> refreshToken")
							}
						}
						
						// tenant_id -> tenantId (if needed)
						if val, exists := data["tenant_id"]; exists {
							if _, hasTenantId := data["tenantId"]; !hasTenantId {
								data["tenantId"] = val
								delete(data, "tenant_id")
								modified = true
							}
						}
						
						// org_id -> orgId (if needed)
						if val, exists := data["org_id"]; exists {
							if _, hasOrgId := data["orgId"]; !hasOrgId {
								data["orgId"] = val
								delete(data, "org_id")
								modified = true
							}
						}

						// otp -> move from query to body if missing (for reset-password links)
						if _, hasOtp := data["otp"]; !hasOtp {
							if otp := q.Get("otp"); otp != "" {
								data["otp"] = otp
								modified = true
								log.Printf("DEBUG: Moved otp from query to JSON body: %s", otp)
							}
						}
						
						if modified {
							newBody, _ := json.Marshal(data)
							r.Body = io.NopCloser(bytes.NewReader(newBody))
							r.ContentLength = int64(len(newBody))
							log.Printf("✅ Normalized JSON body fields (snake_case -> camelCase)")
						} else {
							r.Body = io.NopCloser(bytes.NewReader(body))
							log.Printf("DEBUG normalizeParams: No changes needed")
						}
					} else {
						r.Body = io.NopCloser(bytes.NewReader(body))
						log.Printf("DEBUG normalizeParams: JSON unmarshal failed: %v", err)
					}
				} else if len(body) == 0 {
					log.Printf("DEBUG normalizeParams: Body is empty")
				}
			}
		}
		
		h.ServeHTTP(w, r)
	})
}
