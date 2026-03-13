package storage

import (
	"context"
	"io"
)

type UploadedObject struct {
	Provider  string
	Bucket    string
	ObjectKey string
	SizeBytes int64
}

type Service interface {
	PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (UploadedObject, error)
	OpenObject(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucketName, objectKey string) error
}
