package kb

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
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
		return parsePDF(data)
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

func parsePDF(data []byte) (string, error) {
	reader, err := newPDFReader(data)
	if err != nil {
		return "", &IngestError{
			Message:   "invalid pdf file",
			Retryable: false,
		}
	}

	text, err := extractPDFText(reader)
	if err != nil {
		return "", &IngestError{
			Message:   fmt.Sprintf("failed to extract pdf text: %v", err),
			Retryable: false,
		}
	}

	content := normalizeExtractedText(text)
	if content == "" {
		return "", &IngestError{
			Message:   "pdf contains no extractable text; OCR is not implemented",
			Retryable: false,
		}
	}

	return content, nil
}

func newPDFReader(data []byte) (*pdf.Reader, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		return reader, nil
	}

	normalized, changed := normalizePDFHeader(data)
	if !changed {
		return nil, err
	}

	return pdf.NewReader(bytes.NewReader(normalized), int64(len(normalized)))
}

func normalizePDFHeader(data []byte) ([]byte, bool) {
	if len(data) < 10 || !bytes.HasPrefix(data, []byte("%PDF-1.")) {
		return nil, false
	}

	if data[8] == '\n' || data[8] == '\r' {
		return nil, false
	}

	newlineIndex := bytes.IndexByte(data, '\n')
	if newlineIndex < 0 || newlineIndex <= 8 {
		return nil, false
	}

	normalized := append([]byte(nil), data...)
	normalized[8] = '\n'

	if len(normalized) > 9 && normalized[9] != '%' && normalized[9] != '\n' && normalized[9] != '\r' {
		normalized[9] = '%'
	}

	return normalized, true
}

func extractPDFText(reader *pdf.Reader) (string, error) {
	var builder strings.Builder

	for pageIndex := 1; pageIndex <= reader.NumPage(); pageIndex++ {
		pageText, err := extractPDFPageText(reader.Page(pageIndex))
		if err != nil {
			return "", err
		}
		pageText = strings.TrimSpace(pageText)
		if pageText == "" {
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(pageText)
	}

	return builder.String(), nil
}

func extractPDFPageText(page pdf.Page) (result string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = ""
			err = fmt.Errorf("%v", recovered)
		}
	}()

	content := page.Content()
	return rebuildPDFPageText(content.Text), nil
}

func rebuildPDFPageText(items []pdf.Text) string {
	if len(items) == 0 {
		return ""
	}

	var (
		pageBuilder     strings.Builder
		lineBuilder     strings.Builder
		previous        pdf.Text
		hasPrevious     bool
		pendingSpace    bool
		lastLineY       float64
		hasLastLineY    bool
		lineHasContent  bool
		lineEndsWithGap bool
	)

	flushLine := func(next *pdf.Text) {
		line := strings.TrimSpace(lineBuilder.String())
		lineBuilder.Reset()
		pendingSpace = false
		lineHasContent = false
		lineEndsWithGap = false

		if line == "" {
			hasPrevious = false
			return
		}

		if pageBuilder.Len() > 0 {
			if next != nil && hasLastLineY && shouldInsertPDFParagraphBreak(lastLineY, next.Y, next.FontSize) {
				pageBuilder.WriteString("\n\n")
			} else {
				pageBuilder.WriteByte('\n')
			}
		}
		pageBuilder.WriteString(line)

		if hasPrevious {
			lastLineY = previous.Y
			hasLastLineY = true
		}
		hasPrevious = false
	}

	for index := range items {
		current := items[index]
		segment := replacePDFLigatures(current.S)
		if segment == "" {
			continue
		}

		if isPDFLineBreakSegment(segment) {
			flushLine(nil)
			continue
		}

		if isPDFWhitespaceSegment(segment) {
			if lineHasContent {
				pendingSpace = true
			}
			continue
		}

		if hasPrevious && shouldBreakPDFLine(previous, current) {
			flushLine(&current)
		}

		if lineHasContent && !beginsWithClosingPunctuation(segment) {
			if pendingSpace || (hasPrevious && shouldInsertPDFSpace(previous, current, segment)) {
				if !lineEndsWithGap {
					lineBuilder.WriteByte(' ')
					lineEndsWithGap = true
				}
			}
		}

		lineBuilder.WriteString(segment)
		lineHasContent = true
		lineEndsWithGap = false
		previous = current
		hasPrevious = true
		pendingSpace = false
	}

	flushLine(nil)
	return pageBuilder.String()
}

func isPDFLineBreakSegment(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\f' {
			return true
		}
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return false
}

func isPDFWhitespaceSegment(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if !unicode.IsSpace(r) {
			return false
		}
	}

	return true
}

func shouldBreakPDFLine(previous, current pdf.Text) bool {
	lineBreakThreshold := math.Max(math.Min(nonZeroPDFMetric(previous.FontSize, 12), nonZeroPDFMetric(current.FontSize, 12))*0.8, 3)
	xResetThreshold := math.Max(math.Min(nonZeroPDFMetric(previous.FontSize, 12), nonZeroPDFMetric(current.FontSize, 12))*0.5, 2)

	if current.Y > previous.Y+lineBreakThreshold {
		return true
	}
	if current.X+xResetThreshold < previous.X {
		return true
	}

	return math.Abs(current.Y-previous.Y) > lineBreakThreshold && current.X <= previous.X+previous.W+xResetThreshold
}

func shouldInsertPDFParagraphBreak(previousLineY, nextLineY, nextFontSize float64) bool {
	if nextLineY >= previousLineY {
		return false
	}

	return previousLineY-nextLineY > math.Max(nonZeroPDFMetric(nextFontSize, 12)*1.6, 14)
}

func shouldInsertPDFSpace(previous, current pdf.Text, currentSegment string) bool {
	if currentSegment == "" {
		return false
	}

	if beginsWithClosingPunctuation(currentSegment) {
		return false
	}

	gap := current.X - (previous.X + previous.W)
	if gap <= 0.05 {
		return false
	}

	spaceThreshold := math.Max(math.Max(nonZeroPDFMetric(previous.FontSize, 12), nonZeroPDFMetric(current.FontSize, 12))*0.18, 1.1)

	if gap >= spaceThreshold {
		return true
	}

	prevLast, okPrev := lastRune(replacePDFLigatures(previous.S))
	currFirst, okCurr := firstRune(currentSegment)
	if !okPrev || !okCurr {
		return false
	}

	if isPDFWordRune(prevLast) && isPDFWordRune(currFirst) && gap > 0.55 {
		return true
	}

	return prefersPDFSpaceAfter(prevLast, currFirst) && gap > 0.4
}

func beginsWithClosingPunctuation(value string) bool {
	first, ok := firstRune(value)
	if !ok {
		return false
	}

	return strings.ContainsRune(",.;:!?%)]}>", first)
}

func firstRune(value string) (rune, bool) {
	for _, r := range value {
		return r, true
	}
	return 0, false
}

func lastRune(value string) (rune, bool) {
	for index := len(value); index > 0; {
		r, size := utf8.DecodeLastRuneInString(value[:index])
		if r == utf8.RuneError && size == 1 {
			index -= size
			continue
		}
		return r, true
	}
	return 0, false
}

func isPDFWordRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value)
}

func prefersPDFSpaceAfter(previous, current rune) bool {
	return strings.ContainsRune(":;,!?%)]}", previous) && (isPDFWordRune(current) || strings.ContainsRune(`"'`, current))
}

func nonZeroPDFMetric(value, fallback float64) float64 {
	if value <= 0 {
		return fallback
	}
	return value
}

func replacePDFLigatures(value string) string {
	return strings.NewReplacer(
		"ﬁ", "fi",
		"ﬂ", "fl",
		"ﬀ", "ff",
		"ﬃ", "ffi",
		"ﬄ", "ffl",
		"ﬅ", "ft",
		"ﬆ", "st",
	).Replace(value)
}

func appendLineBreak(builder *strings.Builder) {
	current := builder.String()
	if current == "" || strings.HasSuffix(current, "\n") {
		return
	}
	builder.WriteByte('\n')
}

func normalizeExtractedText(text string) string {
	text = replacePDFLigatures(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\f", "\n")
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
