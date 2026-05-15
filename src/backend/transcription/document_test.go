package transcription_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"AI-Note-Taker/transcription"
)

// fakeFile wraps a bytes.Reader to satisfy the multipart.File interface.
type fakeFile struct{ *bytes.Reader }

func (f *fakeFile) Close() error                            { return nil }
func (f *fakeFile) ReadAt(p []byte, off int64) (int, error) { return f.Reader.ReadAt(p, off) }

func newFakeFile(data string) multipart.File {
	return &fakeFile{bytes.NewReader([]byte(data))}
}

func TestExtractText_TXTPlain(t *testing.T) {
	content := "Hello, world!\nThis is a test."
	file := newFakeFile(content)

	got, err := transcription.ExtractText(file, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestExtractText_TXTWithCharset(t *testing.T) {
	content := "UTF-8 content here."
	file := newFakeFile(content)

	// The handler may pass "text/plain; charset=utf-8" for detected MIME type
	// but ExtractText only switches on "text/plain" — test that branch explicitly.
	got, err := transcription.ExtractText(file, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestExtractText_UnsupportedType(t *testing.T) {
	file := newFakeFile("data")
	_, err := transcription.ExtractText(file, "image/png")
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error %q should mention 'unsupported'", err.Error())
	}
}

func TestExtractText_EmptyTXT(t *testing.T) {
	file := newFakeFile("")
	got, err := transcription.ExtractText(file, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestExtractText_LargeTXT(t *testing.T) {
	content := strings.Repeat("line of text\n", 1000)
	file := newFakeFile(content)

	got, err := transcription.ExtractText(file, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(content) {
		t.Errorf("length mismatch: got %d, want %d", len(got), len(content))
	}
}

// fakeFile also satisfies io.ReadSeeker so we can test Seek behaviour.
func TestExtractText_TXTSeekReset(t *testing.T) {
	content := "seek test content"
	f := &fakeFile{bytes.NewReader([]byte(content))}

	// Exhaust the reader first, then reset — ExtractText should still return full content.
	_, _ = io.ReadAll(f)
	_, _ = f.Seek(0, io.SeekStart)

	got, err := transcription.ExtractText(f, "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}
