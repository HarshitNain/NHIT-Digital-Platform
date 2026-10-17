package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient handles file storage operations with MinIO for vendors
type MinIOClient struct {
	client *minio.Client
	bucket string
	logger *slog.Logger
}

// NewMinIOClient creates a new MinIO client instance
func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool, logger *slog.Logger) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Check if bucket exists, create if it doesn't
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created MinIO bucket", "bucket", bucket)
	}

	// Set public read policy
	_ = ensureBucketPublic(client, bucket)

	logger.Info("MinIO client initialized for Vendor Service", "endpoint", endpoint, "bucket", bucket)

	return &MinIOClient{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

func ensureBucketPublic(client *minio.Client, bucketName string) error {
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

	return client.SetBucketPolicy(context.Background(), bucketName, policy)
}

// UploadDocument uploads a vendor-related document to MinIO
func (m *MinIOClient) UploadDocument(ctx context.Context, vendorID, filename string, file io.Reader, size int64, documentType string) (string, error) {
	// Create object name with vendor ID and document type prefix
	objectName := fmt.Sprintf("documents/%s/%s/%s", vendorID, documentType, filename)

	// Upload the file
	_, err := m.client.PutObject(ctx, m.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream", // Generic, ideally dynamic
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	// Return the object path
	documentURL := fmt.Sprintf("/%s/%s", m.bucket, objectName)
	m.logger.Info("Uploaded document", "url", documentURL)

	return documentURL, nil
}

// UploadSignature uploads a vendor signature to MinIO
func (m *MinIOClient) UploadSignature(ctx context.Context, vendorID, filename string, file io.Reader, size int64) (string, error) {
	return m.UploadDocument(ctx, vendorID, filename, file, size, "signatures")
}

// GetDocumentURL returns a presigned URL for a document/signature
func (m *MinIOClient) GetDocumentURL(ctx context.Context, objectPath string) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, 24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteDocument deletes a document/signature from MinIO
func (m *MinIOClient) DeleteDocument(ctx context.Context, objectPath string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	m.logger.Info("Deleted document", "path", objectPath)
	return nil
}
