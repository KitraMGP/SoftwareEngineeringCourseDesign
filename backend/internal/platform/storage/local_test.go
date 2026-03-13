package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalServicePutOpenDelete(t *testing.T) {
	t.Parallel()

	service, err := NewLocalService(ProviderLocalFS, "test-bucket", t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalService() error = %v", err)
	}

	uploaded, err := service.PutObject(context.Background(), "knowledge-bases/demo/sample.txt", strings.NewReader("hello"), 5, "text/plain")
	if err != nil {
		t.Fatalf("PutObject() error = %v", err)
	}

	reader, err := service.OpenObject(context.Background(), uploaded.Bucket, uploaded.ObjectKey)
	if err != nil {
		t.Fatalf("OpenObject() error = %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if got := string(data); got != "hello" {
		t.Fatalf("stored content mismatch: got %q want %q", got, "hello")
	}

	if err := service.DeleteObject(context.Background(), uploaded.Bucket, uploaded.ObjectKey); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}

	if _, err := service.OpenObject(context.Background(), uploaded.Bucket, uploaded.ObjectKey); err == nil {
		t.Fatalf("OpenObject() after delete expected error")
	}
}
