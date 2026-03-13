package kb

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	MIMEApplicationPDF   = "application/pdf"
	MIMETextPlain        = "text/plain"
	MIMETextMarkdown     = "text/markdown"
	MIMEDocx             = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	DefaultChunkTokens   = 500
	DefaultOverlapTokens = 80
	embeddingDimension   = 1536
	charsPerTokenApprox  = 4
)

var (
	allowedUploadMIMEs = map[string]struct{}{
		MIMEApplicationPDF: {},
		MIMETextPlain:      {},
		MIMETextMarkdown:   {},
		MIMEDocx:           {},
	}
	allowedUploadExtensions = map[string]string{
		".pdf":      MIMEApplicationPDF,
		".txt":      MIMETextPlain,
		".md":       MIMETextMarkdown,
		".markdown": MIMETextMarkdown,
		".docx":     MIMEDocx,
	}
	zeroVectorOnce sync.Once
	zeroVector     string
)

type IngestError struct {
	Message   string
	Retryable bool
}

func (e *IngestError) Error() string {
	return e.Message
}

func NormalizeUploadMetadata(filename, contentType string) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	canonicalFromExt := allowedUploadExtensions[ext]

	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if _, ok := allowedUploadMIMEs[contentType]; ok {
		return contentType, nil
	}
	if canonicalFromExt != "" {
		return canonicalFromExt, nil
	}
	return "", fmt.Errorf("unsupported file type")
}

func ParseDocumentContent(mimeType, filename string, data []byte) (string, error) {
	switch mimeType {
	case MIMETextPlain, MIMETextMarkdown:
		return normalizeExtractedText(string(data)), nil
	case MIMEDocx:
		return parseDOCX(data)
	case MIMEApplicationPDF:
		return "", &IngestError{
			Message:   "pdf parser is not implemented yet",
			Retryable: false,
		}
	default:
		if detected, err := NormalizeUploadMetadata(filename, mimeType); err == nil {
			return ParseDocumentContent(detected, filename, data)
		}
		return "", &IngestError{
			Message:   "unsupported file type",
			Retryable: false,
		}
	}
}

func BuildDocumentChunks(text string) []DocumentChunkInput {
	text = normalizeExtractedText(text)
	if text == "" {
		return nil
	}

	targetChars := DefaultChunkTokens * charsPerTokenApprox
	overlapChars := DefaultOverlapTokens * charsPerTokenApprox
	paragraphs := splitParagraphs(text)

	var rawChunks []string
	var current strings.Builder

	flushCurrent := func() {
		chunk := strings.TrimSpace(current.String())
		if chunk != "" {
			rawChunks = append(rawChunks, chunk)
		}
		current.Reset()
	}

	for _, paragraph := range paragraphs {
		if paragraph == "" {
			continue
		}

		if utf8.RuneCountInString(paragraph) > targetChars {
			flushCurrent()
			rawChunks = append(rawChunks, windowSplit(paragraph, targetChars, overlapChars)...)
			continue
		}

		if current.Len() == 0 {
			current.WriteString(paragraph)
			continue
		}

		candidate := current.String() + "\n\n" + paragraph
		if utf8.RuneCountInString(candidate) <= targetChars {
			current.WriteString("\n\n")
			current.WriteString(paragraph)
			continue
		}

		flushCurrent()
		current.WriteString(paragraph)
	}
	flushCurrent()

	chunks := make([]DocumentChunkInput, 0, len(rawChunks))
	for idx, chunk := range rawChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}
		chunks = append(chunks, DocumentChunkInput{
			ChunkIndex: idx,
			Content:    chunk,
			TokenCount: estimateTokens(chunk),
			Embedding:  ZeroVector(),
		})
	}
	return chunks
}

func ZeroVector() string {
	zeroVectorOnce.Do(func() {
		parts := make([]string, embeddingDimension)
		for idx := range parts {
			parts[idx] = "0"
		}
		zeroVector = "[" + strings.Join(parts, ",") + "]"
	})
	return zeroVector
}

func parseDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", &IngestError{
			Message:   "invalid docx file",
			Retryable: false,
		}
	}

	var documentFile *zip.File
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			documentFile = file
			break
		}
	}
	if documentFile == nil {
		return "", &IngestError{
			Message:   "invalid docx file",
			Retryable: false,
		}
	}

	rc, err := documentFile.Open()
	if err != nil {
		return "", fmt.Errorf("open docx document.xml: %w", err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var (
		builder strings.Builder
		inText  bool
	)

	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("decode docx xml: %w", err)
		}

		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "t":
				inText = true
			case "br", "cr", "p":
				appendLineBreak(&builder)
			case "tab":
				builder.WriteByte('\t')
			}
		case xml.EndElement:
			if value.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				builder.Write([]byte(value))
			}
		}
	}

	content := normalizeExtractedText(builder.String())
	if content == "" {
		return "", &IngestError{
			Message:   "document content is empty",
			Retryable: false,
		}
	}
	return content, nil
}

func appendLineBreak(builder *strings.Builder) {
	current := builder.String()
	if current == "" || strings.HasSuffix(current, "\n") {
		return
	}
	builder.WriteByte('\n')
}

func normalizeExtractedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for idx := range lines {
		lines[idx] = strings.TrimSpace(lines[idx])
	}
	text = strings.Join(lines, "\n")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(text)
}

func splitParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")
	paragraphs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			paragraphs = append(paragraphs, part)
		}
	}
	return paragraphs
}

func windowSplit(text string, targetChars, overlapChars int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if len(runes) <= targetChars {
		return []string{string(runes)}
	}

	step := targetChars - overlapChars
	if step <= 0 {
		step = targetChars
	}

	chunks := make([]string, 0, (len(runes)/step)+1)
	for start := 0; start < len(runes); start += step {
		end := start + targetChars
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func estimateTokens(text string) int {
	count := utf8.RuneCountInString(strings.TrimSpace(text))
	if count == 0 {
		return 0
	}
	tokens := count / charsPerTokenApprox
	if count%charsPerTokenApprox != 0 {
		tokens++
	}
	if tokens == 0 {
		return 1
	}
	return tokens
}
