package transcription

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
    "strings"
	"github.com/fumiama/go-docx"
	"github.com/gen2brain/go-fitz"
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
		log.Printf(err.Error())
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
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	doc, err := fitz.NewFromMemory(b)
	if err != nil {
		return "", fmt.Errorf("failed to open pdf: %w", err)
	}
	defer doc.Close()

	var sb strings.Builder
	for i := 0; i < doc.NumPage(); i++ {
		text, err := doc.Text(i)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", i, err)
		}
		sb.WriteString(text)
	}

	return strings.TrimSpace(sb.String()), nil
}
