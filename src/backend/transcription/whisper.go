package transcription

import (
	"fmt"
	"io"
	"mime/multipart"
	"bytes"
	"net/http"
	"encoding/json"
)

func TranscribeAudio(file multipart.File, filename string) (string, error) {
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return "", fmt.Errorf("failed to create form file: %w", err)
    }

    if _, err := io.Copy(part, file); err != nil {
        return "", fmt.Errorf("failed to copy file: %w", err)
    }
    writer.Close()

    resp, err := http.Post(
        "http://localhost:8081/transcribe",
        writer.FormDataContentType(),
        body,
    )
    if err != nil {
        return "", fmt.Errorf("failed to call whisper service: %w", err)
    }
    defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

    var result struct {
        Text string `json:"text"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return result.Text, nil
}
