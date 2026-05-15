package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"AI-Note-Taker/notes"
	"AI-Note-Taker/storage"
	"AI-Note-Taker/transcription"
)

var (
	fakeOpenAIServer  *httptest.Server
	fakeWhisperServer *httptest.Server
	fakeS3Server      *httptest.Server
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("R2_BUCKET_NAME", "test-bucket")

	fakeOpenAIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "Generated study sheet"}},
			},
		})
	}))
	notes.OpenAIBaseURL = fakeOpenAIServer.URL

	fakeWhisperServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"text": "Transcribed audio text"})
	}))
	transcription.OpenAIBaseURL = fakeWhisperServer.URL

	fakeS3Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			w.Header().Set("ETag", `"test-etag"`)
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	storage.SetTestEndpoint(fakeS3Server.URL)
	if err := storage.InitR2("test", "test-key", "test-secret-key"); err != nil {
		panic("failed to init test R2: " + err.Error())
	}

	code := m.Run()

	fakeOpenAIServer.Close()
	fakeWhisperServer.Close()
	fakeS3Server.Close()

	os.Exit(code)
}

// authCookie returns a valid signed JWT cookie for protected endpoint tests.
func authCookie(t *testing.T) *http.Cookie {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test-user-id",
		"email":   "test@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return &http.Cookie{Name: "auth_token", Value: signed}
}
