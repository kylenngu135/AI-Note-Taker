package transcription_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"AI-Note-Taker/transcription"
)

func TestTranscribeAudio_Success(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Authorization header with Bearer token, got %q", auth)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		if r.FormValue("model") != "whisper-1" {
			t.Errorf("expected model=whisper-1, got %q", r.FormValue("model"))
		}
		if r.MultipartForm == nil {
			t.Error("expected multipart form with 'file' field")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": "hello world"})
	}))
	defer srv.Close()
	transcription.OpenAIBaseURL = srv.URL

	file := &fakeFile{bytes.NewReader([]byte("fake audio data"))}
	got, err := transcription.TranscribeAudio(file, "audio.mp3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want 'hello world'", got)
	}
}

func TestTranscribeAudio_ServiceReturnsNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	transcription.OpenAIBaseURL = srv.URL

	file := &fakeFile{bytes.NewReader([]byte("audio"))}
	_, err := transcription.TranscribeAudio(file, "audio.mp3")
	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestTranscribeAudio_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	transcription.OpenAIBaseURL = srv.URL

	file := &fakeFile{bytes.NewReader([]byte("audio"))}
	_, err := transcription.TranscribeAudio(file, "audio.mp3")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestTranscribeAudio_ServiceUnavailable(t *testing.T) {
	transcription.OpenAIBaseURL = "http://127.0.0.1:19999" // nothing listening
	orig := transcription.RetryBackoff
	transcription.RetryBackoff = nil
	t.Cleanup(func() { transcription.RetryBackoff = orig })

	file := &fakeFile{bytes.NewReader([]byte("audio"))}
	_, err := transcription.TranscribeAudio(file, "audio.mp3")
	if err == nil {
		t.Error("expected error when service is unavailable, got nil")
	}
}
