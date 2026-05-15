package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"AI-Note-Taker/middleware"
)

func makeToken(secret string, claims jwt.MapClaims) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

// --- EnableCORS ---

func TestEnableCORS_SetsHeaders(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")

	handler := middleware.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestEnableCORS_OptionsShortCircuits(t *testing.T) {
	reached := false
	handler := middleware.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if reached {
		t.Error("inner handler should not be called for OPTIONS")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestEnableCORS_AllowedMethodsHeader(t *testing.T) {
	handler := middleware.EnableCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	got := rr.Header().Get("Access-Control-Allow-Methods")
	if got == "" {
		t.Error("Access-Control-Allow-Methods should be set")
	}
}

// --- ValidateJWT ---

func TestValidateJWT_ValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token := makeToken("test-secret", jwt.MapClaims{
		"user_id": "uid-1",
		"email":   "a@b.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	claims, err := middleware.ValidateJWT(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*claims)["user_id"] != "uid-1" {
		t.Errorf("user_id = %v, want uid-1", (*claims)["user_id"])
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token := makeToken("wrong-secret", jwt.MapClaims{
		"user_id": "uid-1",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	_, err := middleware.ValidateJWT(token)
	if err == nil {
		t.Error("expected error for wrong signature, got nil")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token := makeToken("test-secret", jwt.MapClaims{
		"user_id": "uid-1",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})

	_, err := middleware.ValidateJWT(token)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := middleware.ValidateJWT("not.a.jwt")
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}
}

// --- AuthMiddleware ---

func TestAuthMiddleware_NoCookie(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "bad-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthMiddleware_ValidToken_PassesThrough(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	reached := false
	handler := middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))

	token := makeToken("test-secret", jwt.MapClaims{
		"user_id": "uid-1",
		"email":   "a@b.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !reached {
		t.Error("inner handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}
