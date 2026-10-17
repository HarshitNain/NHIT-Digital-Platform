package handler

import(
	"context"
	"net/http"
	"strings"
	"log"

	"google.golang.org/grpc/metadata"
)

// sanitizeFilename sanitizes a string to be safe for use in a file path or object name.
func sanitizeFilename(name string) string {
	// Replace any character that is not alphanumeric, hyphen, or underscore with an underscore.
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	return safe
}

// ToMetadata converts HTTP headers to gRPC metadata, filtering out prohibited headers.
func ToMetadata(r *http.Request) metadata.MD {
	md := metadata.MD{}
	for k, vv := range r.Header {
		lowKey := strings.ToLower(k)
		// Skip prohibited HTTP/2 headers that cause protocol errors in gRPC
		if lowKey == "connection" ||
			lowKey == "transfer-encoding" ||
			lowKey == "te" ||
			lowKey == "upgrade" ||
			lowKey == "keep-alive" ||
			lowKey == "proxy-connection" ||
			lowKey == "content-length" {
			continue
		}
		md.Set(lowKey, vv[0])
	}
	
	// Always try to extract tenant and org info from individual headers first
	// (common for refresh requests or when frontend doesn't send JWT)
	if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
		md.Set("tenant_id", tid)
	} else if tid := r.Header.Get("Tenant-ID"); tid != "" {
		md.Set("tenant_id", tid)
	} else if tid := r.Header.Get("tenant_id"); tid != "" {
		md.Set("tenant_id", tid)
	}

	if oid := r.Header.Get("X-Org-ID"); oid != "" {
		md.Set("org_id", oid)
	} else if oid := r.Header.Get("Org-ID"); oid != "" {
		md.Set("org_id", oid)
	} else if oid := r.Header.Get("org_id"); oid != "" {
		md.Set("org_id", oid)
	}
	
	if auth := md.Get("authorization"); len(auth) > 0 {
		log.Printf("UTILS: ToMetadata included authorization header (len: %d)", len(auth[0]))
	} else {
		// Log all headers to see what happened
		allHeaders := make(map[string][]string)
		for k, v := range r.Header {
			allHeaders[k] = v
		}
		log.Printf("UTILS: [MISSED] Authorization header MISSING. Path: %s. All headers: %v", r.URL.Path, allHeaders)
	}

	return md
}

// forwardMetadata extracts relevant headers from the HTTP request and forwards them as gRPC metadata.
func forwardMetadata(r *http.Request) context.Context {
	return metadata.NewOutgoingContext(r.Context(), ToMetadata(r))
}
