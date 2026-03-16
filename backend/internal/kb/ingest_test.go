package kb

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ledongthuc/pdf"
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

func TestParseDocumentContentPDF(t *testing.T) {
	t.Parallel()

	content, err := ParseDocumentContent(MIMEApplicationPDF, "demo.pdf", buildMinimalPDF("Hello PDF World"))
	if err != nil {
		t.Fatalf("ParseDocumentContent() error = %v", err)
	}
	if !strings.Contains(content, "Hello PDF World") {
		t.Fatalf("ParseDocumentContent() = %q", content)
	}
}

func TestNormalizePDFHeader(t *testing.T) {
	t.Parallel()

	original := []byte("%PDF-1.4 Sharp Scanned ImagePDF\n%Sharp Non-Encryption\n")

	normalized, changed := normalizePDFHeader(original)
	if !changed {
		t.Fatalf("normalizePDFHeader() expected change")
	}
	if normalized[8] != '\n' {
		t.Fatalf("normalizePDFHeader() expected newline at header boundary, got %q", normalized[8])
	}
	if normalized[9] != '%' {
		t.Fatalf("normalizePDFHeader() expected comment marker, got %q", normalized[9])
	}
	if len(normalized) != len(original) {
		t.Fatalf("normalizePDFHeader() changed byte length")
	}
}

func TestParseDocumentContentPDFSample(t *testing.T) {
	t.Parallel()

	path := os.Getenv("BACKEND_PDF_SAMPLE_PATH")
	if path == "" {
		t.Skip("BACKEND_PDF_SAMPLE_PATH is not set")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	expected := os.Getenv("BACKEND_PDF_SAMPLE_EXPECT")
	content, err := ParseDocumentContent(MIMEApplicationPDF, path, data)
	if err == nil {
		if expected != "" {
			if !strings.Contains(content, expected) {
				t.Fatalf("ParseDocumentContent() missing expected text %q, got prefix %q", expected, truncateForTest(content, 240))
			}
			return
		}

		if len(content) < 200 || len(strings.Fields(content)) < 20 {
			t.Fatalf("ParseDocumentContent() extracted content looks too short: prefix %q", truncateForTest(content, 240))
		}
		return
	}

	var ingestErr *IngestError
	if !errors.As(err, &ingestErr) {
		t.Fatalf("ParseDocumentContent() unexpected error type = %v", err)
	}
	if !strings.Contains(ingestErr.Message, "OCR is not implemented") {
		t.Fatalf("ParseDocumentContent() unexpected error = %v", err)
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

func TestReplacePDFLigatures(t *testing.T) {
	t.Parallel()

	got := replacePDFLigatures("Efﬁcient ﬂow oﬀers ﬃ and ﬄ forms")
	want := "Efficient flow offers ffi and ffl forms"

	if got != want {
		t.Fatalf("replacePDFLigatures() = %q, want %q", got, want)
	}
}

func TestRebuildPDFPageText(t *testing.T) {
	t.Parallel()

	items := []pdf.Text{
		{X: 10, Y: 720, W: 8, FontSize: 12, S: "H"},
		{X: 18.2, Y: 720, W: 6, FontSize: 12, S: "e"},
		{X: 24.1, Y: 720, W: 3, FontSize: 12, S: "l"},
		{X: 27.0, Y: 720, W: 3, FontSize: 12, S: "l"},
		{X: 30.0, Y: 720, W: 7, FontSize: 12, S: "o"},
		{X: 42.5, Y: 720, W: 10, FontSize: 12, S: "W"},
		{X: 52.7, Y: 720, W: 7, FontSize: 12, S: "o"},
		{X: 59.8, Y: 720, W: 4, FontSize: 12, S: "r"},
		{X: 63.7, Y: 720, W: 3, FontSize: 12, S: "l"},
		{X: 66.6, Y: 720, W: 7, FontSize: 12, S: "d"},
		{X: 10, Y: 705, W: 8, FontSize: 12, S: "N"},
		{X: 18.2, Y: 705, W: 7, FontSize: 12, S: "e"},
		{X: 25.3, Y: 705, W: 6, FontSize: 12, S: "x"},
		{X: 31.2, Y: 705, W: 4, FontSize: 12, S: "t"},
		{X: 40.5, Y: 705, W: 3, FontSize: 12, S: "L"},
		{X: 43.6, Y: 705, W: 7, FontSize: 12, S: "i"},
		{X: 50.6, Y: 705, W: 7, FontSize: 12, S: "n"},
		{X: 57.7, Y: 705, W: 7, FontSize: 12, S: "e"},
	}

	got := rebuildPDFPageText(items)
	want := "Hello World\nNext Line"

	if got != want {
		t.Fatalf("rebuildPDFPageText() = %q, want %q", got, want)
	}
}

func buildMinimalPDF(text string) []byte {
	escapedText := strings.NewReplacer(`\`, `\\`, "(", `\(`, ")", `\)`).Replace(text)
	contentStream := fmt.Sprintf("BT\n/F1 24 Tf\n72 720 Td\n(%s) Tj\nET\n", escapedText)
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n",
		"4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
		fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(contentStream), contentStream),
	}

	var builder strings.Builder
	builder.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = builder.Len()
		builder.WriteString(object)
	}

	xrefOffset := builder.Len()
	builder.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	builder.WriteString("0000000000 65535 f \n")
	for index := 1; index <= len(objects); index++ {
		builder.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[index]))
	}
	builder.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", len(objects)+1))
	builder.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return []byte(builder.String())
}

func truncateForTest(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}
