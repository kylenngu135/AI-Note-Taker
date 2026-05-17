package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

var OpenAIBaseURL = "https://api.openai.com"

// RetryBackoff controls sleep durations between Whisper API retry attempts.
// Overridable in tests to avoid slow runs.
var RetryBackoff = []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}

func TranscribeAudio(file multipart.File, filename string) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek file: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	bodyBytes := body.Bytes()
	contentType := writer.FormDataContentType()
	maxAttempts := len(RetryBackoff) + 1

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodPost, OpenAIBaseURL+"/v1/audio/transcriptions", bytes.NewReader(bodyBytes))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
		req.Header.Set("Content-Type", contentType)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to call whisper service: %w", err)
			if attempt < maxAttempts {
				log.Printf("chunk %s: TLS error, retrying (attempt %d of %d)...", filename, attempt+1, maxAttempts)
				time.Sleep(RetryBackoff[attempt-1])
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var result struct {
			Text string `json:"text"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		_ = resp.Body.Close()
		if decodeErr != nil {
			return "", fmt.Errorf("failed to decode response: %w", decodeErr)
		}

		return result.Text, nil
	}

	return "", lastErr
}
