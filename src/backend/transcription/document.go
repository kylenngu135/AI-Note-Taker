package transcription

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/fumiama/go-docx"
	"github.com/ledongthuc/pdf"
)

func ExtractText(file multipart.File, fileType string) (string, error) {
	switch fileType {
	case "application/pdf":
		return extractPDFText(file)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return extractDOCXText(file)
	case "text/plain":
		return extractTXTText(file)
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func extractTXTText(file multipart.File) (string, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func extractDOCXText(file multipart.File) (string, error) {
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return "", err
	}

	// seek back to start before reading
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	doc, err := docx.Parse(file, size)
	if err != nil {
		return "", err
	}

	var text string
	for _, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table:
			text += fmt.Sprintf("%v\n", it)
		}
	}

	return text, nil
}

func extractPDFText(file multipart.File) (string, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(b)
	r, err := pdf.NewReader(reader, int64(len(b)))
	if err != nil {
		return "", err
	}

	var text string
	for pageIndex := 1; pageIndex <= r.NumPage(); pageIndex++ {
		page := r.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		text += content
	}

	return text, nil
}
