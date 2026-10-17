package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewDocumentStorage creates a MinIO-backed storage when configuration
// is present; otherwise it falls back to an in-memory storage suitable for
// development and tests.
func NewDocumentStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (ports.DocumentStorage, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		// No MinIO configuration provided: fall back to in-memory storage.
		return NewInMemoryDocumentStorage(), nil
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &minioDocumentStorage{client: client, bucket: bucket}, nil
}

// minioDocumentStorage uses MinIO/S3 for binary objects.
type minioDocumentStorage struct {
	client *minio.Client
	bucket string
}

func (s *minioDocumentStorage) Save(ctx context.Context, objectName string, content []byte, contentType string) error {
	reader := bytes.NewReader(content)
	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *minioDocumentStorage) Load(ctx context.Context, objectName string) ([]byte, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", err
	}

	stat, err := obj.Stat()
	if err != nil {
		return data, "", nil
	}

	return data, stat.ContentType, nil
}

// inMemoryDocumentStorage keeps documents in RAM, keyed by objectName.
type inMemoryDocumentStorage struct {
	mu      sync.RWMutex
	objects map[string]inMemoryObject
}

type inMemoryObject struct {
	data        []byte
	contentType string
}

// NewInMemoryDocumentStorage constructs an in-memory storage implementation.
func NewInMemoryDocumentStorage() ports.DocumentStorage {
	return &inMemoryDocumentStorage{
		objects: make(map[string]inMemoryObject),
	}
}

func (s *inMemoryDocumentStorage) Save(ctx context.Context, objectName string, content []byte, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[objectName] = inMemoryObject{data: append([]byte(nil), content...), contentType: contentType}
	return nil
}

func (s *inMemoryDocumentStorage) Load(ctx context.Context, objectName string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj, ok := s.objects[objectName]
	if !ok {
		return nil, "", ports.ErrNotFound
	}

	return append([]byte(nil), obj.data...), obj.contentType, nil
}
