package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps MinIO client for document storage
type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadDocument uploads a document to MinIO
func (m *MinIOClient) UploadDocument(ctx context.Context, objectKey string, data []byte, contentType string) (string, int64, error) {
	reader := bytes.NewReader(data)
	size := int64(len(data))

	_, err := m.client.PutObject(ctx, m.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload document: %w", err)
	}

	return objectKey, size, nil
}

// DownloadDocument downloads a document from MinIO
func (m *MinIOClient) DownloadDocument(ctx context.Context, objectKey string) ([]byte, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	return data, nil
}

// DeleteDocument deletes a document from MinIO
func (m *MinIOClient) DeleteDocument(ctx context.Context, objectKey string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// GetDocumentURL generates a presigned URL for document access
func (m *MinIOClient) GetDocumentURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GenerateObjectKey generates a unique object key for a document
func GenerateObjectKey(paymentNoteID int64, filename string) string {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	return fmt.Sprintf("payment-notes/%d/%d%s", paymentNoteID, timestamp, ext)
}
