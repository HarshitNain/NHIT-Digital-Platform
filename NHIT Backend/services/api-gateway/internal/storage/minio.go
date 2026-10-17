package storage

import (
	"context"
	"fmt"
	"log"
	"time"
	"path/filepath"
	"strings"
	"net/http"
	"mime"
	"bytes"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client        *minio.Client
	bucket        string
	baseURL       string
	publicBuckets map[string]bool
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("✅ Created MinIO bucket: %s", bucket)
	}
	
	// Determine base URL
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucket)


	client := &MinIOClient{
		client:        minioClient,
		bucket:        bucket,
		baseURL:       baseURL,
		publicBuckets: make(map[string]bool),
	}
	
	// Set public read policy for default bucket
	_ = client.EnsureBucketExists(ctx, bucket)
	
	return client, nil
}

// EnsureBucketExists checks if a bucket exists and creates it with public read policy if it doesn't.
func (m *MinIOClient) EnsureBucketExists(ctx context.Context, bucketName string) error {
	// Check if we've already set this bucket to public in this session
	if m.publicBuckets[bucketName] {
		return nil
	}

	exists, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("✅ Created MinIO bucket: %s", bucketName)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": ["s3:GetObject"],
				"Effect": "Allow",
				"Principal": "*",
				"Resource": ["arn:aws:s3:::%s/*"],
				"Sid": ""
			}
		]
	}`, bucketName)

	err = m.client.SetBucketPolicy(ctx, bucketName, policy)
	if err == nil {
		m.publicBuckets[bucketName] = true
	}
	return err
}

// UploadFile uploads a file to MinIO and returns the public URL
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) (string, error) {
	// Ensure default bucket is public
	_ = m.EnsureBucketExists(ctx, m.bucket)

	// If it's a file path on disk
	_, err := m.client.FPutObject(ctx, m.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", m.baseURL, objectName), nil
}

// UploadStream uploads a stream (e.g., from multipart form)
// bucketName is optional. If empty, uses the client's default bucket.
func (m *MinIOClient) UploadStream(ctx context.Context, bucketName, objectName string, reader interface{}, size int64, contentType string) (string, error) {
    targetBucket := m.bucket
    // If a specific bucket needed, use it
    if bucketName != "" {
        targetBucket = bucketName
    }
    
    // Ensure bucket exists and is public
    _ = m.EnsureBucketExists(ctx, targetBucket)

    // 4. Ensure reader implements io.Reader
    r, ok := reader.(io.Reader)
    if !ok {
        return "", fmt.Errorf("reader does not implement io.Reader")
    }

    // 5. Detect Content-Type robustly
    log.Printf("DEBUG: Uploading %s (size: %d, client-type: %s) to %s", objectName, size, contentType, targetBucket)

    sniffBuf := make([]byte, 512)
    n, readErr := r.Read(sniffBuf)
    if n > 0 {
        detectedType := http.DetectContentType(sniffBuf[:n])
        // Put the sniffed bytes back
        r = io.MultiReader(bytes.NewReader(sniffBuf[:n]), r)
        
        // If detected is solid, use it. application/octet-stream is generic.
        if detectedType != "application/octet-stream" {
            contentType = detectedType
            log.Printf("DEBUG: Detected type %s for %s", detectedType, objectName)
        } else {
            // Fallback to extension if sniffing is inconclusive
            ext := strings.ToLower(filepath.Ext(objectName))
            switch ext {
            case ".pdf": contentType = "application/pdf"
            case ".jpg", ".jpeg": contentType = "image/jpeg"
            case ".png": contentType = "image/png"
            case ".gif": contentType = "image/gif"
            case ".svg": contentType = "image/svg+xml"
            case ".csv": contentType = "text/csv"
            default:
                if mimeType := mime.TypeByExtension(ext); mimeType != "" {
                    contentType = mimeType
                }
            }
            log.Printf("DEBUG: Extension-based type %s for %s", contentType, objectName)
        }
    } else if readErr != nil && readErr != io.EOF {
        log.Printf("DEBUG: Read error for sniffing %s: %v", objectName, readErr)
    }

    log.Printf("DEBUG: Final content-type for %s: %s", objectName, contentType)

	info, err := m.client.PutObject(ctx, targetBucket, objectName, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	
	log.Printf("Successfully uploaded %s to bucket %s of size %d", objectName, targetBucket, info.Size)
	
	// Determine correct Base URL for this bucket
	// m.baseURL assumes default bucket. We need to construct for dynamic bucket.
	// HACK: Reconstruct URL or replace bucket in baseURL?
	// baseURL format: schema://endpoint/bucket
	// We can just construct it freshly.
	
	// If it's the default bucket, use optimized m.baseURL
	if targetBucket == m.bucket {
	    return fmt.Sprintf("%s/%s", m.baseURL, objectName), nil
	}
	
	// Construct for new bucket
	endpointURL := m.client.EndpointURL()
	// Scheme://Host/Bucket/Object
	return fmt.Sprintf("%s/%s/%s", endpointURL.String(), targetBucket, objectName), nil
}

func (m *MinIOClient) GeneratePresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// DeleteObject deletes an object from a specific bucket
func (m *MinIOClient) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	err := m.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove object: %w", err)
	}
	log.Printf("Successfully deleted %s from bucket %s", objectName, bucketName)
	return nil
}

// ParseBucketAndObject extracts bucket and object name from a MinIO URL
// Example: http://192.168.1.14:9000/nhit.vendors/VendorName/file.pdf -> nhit.vendors, VendorName/file.pdf
func (m *MinIOClient) ParseBucketAndObject(urlStr string) (string, string, error) {
    // Basic logic: find the part after the host
    // URL format: http://host:port/bucket/object/path...
    
    // We can use strings.Split and look for the base URL match or just index
    // Endpoint: http://192.168.1.14:9000
    endpoint := m.client.EndpointURL().String()
    if !strings.HasSuffix(endpoint, "/") {
        endpoint += "/"
    }
    
    if !strings.HasPrefix(urlStr, endpoint) {
        return "", "", fmt.Errorf("URL %s does not start with endpoint %s", urlStr, endpoint)
    }
    
    remaining := strings.TrimPrefix(urlStr, endpoint)
    parts := strings.SplitN(remaining, "/", 2)
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid MinIO URL format: %s", urlStr)
    }
    
    return parts[0], parts[1], nil
}

// GenerateUniqueFileName creates a key like "vendors/{vendor_id}/{timestamp}_{filename}"
func GenerateUniqueFileName(folder, id, filename string) string {
    ext := filepath.Ext(filename)
    return fmt.Sprintf("%s/%s/%d%s", folder, id, time.Now().Unix(), ext)
}
