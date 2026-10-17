package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient handles file storage operations with MinIO for organizations
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
	err = ensureBucketPublic(client, bucket)
	if err != nil {
		logger.Warn("Failed to set public policy for bucket", "bucket", bucket, "error", err)
	}

	logger.Info("MinIO client initialized for Organization Service", "endpoint", endpoint, "bucket", bucket)

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

// UploadFile uploads a file to MinIO
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error) {
	// Upload the file
	_, err := m.client.PutObject(ctx, m.bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType, // Dynamic content type
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the object path
	fileURL := fmt.Sprintf("/%s/%s", m.bucket, objectName)
	m.logger.Info("Uploaded file to MinIO", "url", fileURL)

	return fileURL, nil
}

// GetLogoURL returns the presigned URL for a logo file
func (m *MinIOClient) GetLogoURL(ctx context.Context, objectPath string) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, 24*60*60, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteLogo deletes a logo file from MinIO
func (m *MinIOClient) DeleteLogo(ctx context.Context, objectPath string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete logo: %w", err)
	}

	m.logger.Info("Deleted logo from MinIO", "path", objectPath)
	return nil
}
