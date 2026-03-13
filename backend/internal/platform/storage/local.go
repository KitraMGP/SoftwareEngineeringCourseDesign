package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const ProviderLocalFS = "local_fs"

type LocalService struct {
	provider string
	bucket   string
	rootDir  string
}

func NewLocalService(provider, bucket, rootDir string) (*LocalService, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return nil, fmt.Errorf("local storage root is required")
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(absRoot, bucket), 0o755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}

	return &LocalService{
		provider: provider,
		bucket:   bucket,
		rootDir:  absRoot,
	}, nil
}

func (s *LocalService) PutObject(_ context.Context, objectKey string, reader io.Reader, size int64, _ string) (UploadedObject, error) {
	targetPath, err := s.resolvePath(s.bucket, objectKey)
	if err != nil {
		return UploadedObject{}, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return UploadedObject{}, fmt.Errorf("create object directory: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return UploadedObject{}, fmt.Errorf("create object file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, reader)
	if err != nil {
		return UploadedObject{}, fmt.Errorf("write object file: %w", err)
	}

	return UploadedObject{
		Provider:  s.provider,
		Bucket:    s.bucket,
		ObjectKey: objectKey,
		SizeBytes: max(size, written),
	}, nil
}

func (s *LocalService) OpenObject(_ context.Context, bucketName, objectKey string) (io.ReadCloser, error) {
	targetPath, err := s.resolvePath(bucketName, objectKey)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(targetPath)
	if err != nil {
		return nil, fmt.Errorf("open object file: %w", err)
	}
	return file, nil
}

func (s *LocalService) DeleteObject(_ context.Context, bucketName, objectKey string) error {
	targetPath, err := s.resolvePath(bucketName, objectKey)
	if err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete object file: %w", err)
	}
	return nil
}

func (s *LocalService) resolvePath(bucketName, objectKey string) (string, error) {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return "", fmt.Errorf("storage bucket is required")
	}

	cleanKey := path.Clean("/" + objectKey)
	cleanKey = strings.TrimPrefix(cleanKey, "/")
	if cleanKey == "." || cleanKey == "" {
		return "", fmt.Errorf("object key is required")
	}

	baseDir := filepath.Join(s.rootDir, bucketName)
	targetPath := filepath.Join(baseDir, filepath.FromSlash(cleanKey))
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve object path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("object key escapes storage root")
	}
	return targetPath, nil
}
