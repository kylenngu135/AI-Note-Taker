package transcription

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func TranscribeAudio(file multipart.File, filename string) (string, error) {
	// write to temp file since whisper requires a file path
	tmp, err := os.CreateTemp("", "upload-*"+filepath.Ext(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, file); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tmp.Close()

	// run whisper CLI
	cmd := exec.Command("whisper", tmp.Name(), "--model", "small", "--output_format", "txt", "--output_dir", os.TempDir())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to transcribe: %w", err)
	}

	// whisper saves output as filename.txt in the output dir
	txtPath := filepath.Join(os.TempDir(), strings.TrimSuffix(filepath.Base(tmp.Name()), filepath.Ext(tmp.Name()))+".txt")
	defer os.Remove(txtPath)

	bytes, err := os.ReadFile(txtPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}

	return string(bytes), nil
}
