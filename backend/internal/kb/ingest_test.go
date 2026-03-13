package kb

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeUploadMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filename    string
		contentType string
		want        string
		wantErr     bool
	}{
		{name: "content type", filename: "demo.txt", contentType: "text/plain", want: MIMETextPlain},
		{name: "fallback extension", filename: "demo.md", contentType: "application/octet-stream", want: MIMETextMarkdown},
		{name: "unsupported", filename: "demo.exe", contentType: "application/octet-stream", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeUploadMetadata(tc.filename, tc.contentType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeUploadMetadata() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeUploadMetadata() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeUploadMetadata() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseDocumentContentDOCX(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	xmlContent := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Hello</w:t></w:r></w:p>
    <w:p><w:r><w:t>World</w:t></w:r></w:p>
  </w:body>
</w:document>`

	if _, err := file.Write([]byte(xmlContent)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	content, err := ParseDocumentContent(MIMEDocx, "demo.docx", buffer.Bytes())
	if err != nil {
		t.Fatalf("ParseDocumentContent() error = %v", err)
	}
	if !strings.Contains(content, "Hello") || !strings.Contains(content, "World") {
		t.Fatalf("ParseDocumentContent() = %q", content)
	}
}

func TestBuildDocumentChunks(t *testing.T) {
	t.Parallel()

	input := strings.Repeat("paragraph ", 260)
	chunks := BuildDocumentChunks(input)
	if len(chunks) == 0 {
		t.Fatalf("BuildDocumentChunks() returned no chunks")
	}
	if chunks[0].Embedding != ZeroVector() {
		t.Fatalf("chunk embedding placeholder mismatch")
	}
}
