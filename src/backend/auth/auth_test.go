package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"AI-Note-Taker/auth"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}

func newDB(t *testing.T) (*auth.Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &auth.Handler{DB: db}, mock
}

func jsonBody(v interface{}) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func validToken(t *testing.T) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "uid-1",
		"email":   "a@b.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	signed, err := tok.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

// --- RegisterUserHandler ---

func TestRegisterUserHandler_Success(t *testing.T) {
	h, mock := newDB(t)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), "new@example.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(map[string]string{"email": "new@example.com", "password": "password123"}))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.RegisterUserHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["token"] == "" {
		t.Error("expected token in response")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestRegisterUserHandler_InvalidEmail(t *testing.T) {
	h, _ := newDB(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(map[string]string{"email": "not-an-email", "password": "pass"}))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestRegisterUserHandler_EmailAlreadyExists(t *testing.T) {
	h, mock := newDB(t)

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("existing@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(map[string]string{"email": "existing@example.com", "password": "pass"}))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

// --- LoginHandler ---

func TestLoginHandler_Success(t *testing.T) {
	h, mock := newDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 4) // cost 4 for speed
	mock.ExpectQuery("SELECT password_hash").
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(string(hash)))
	mock.ExpectQuery("SELECT id").
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("uid-42"))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(map[string]string{"email": "user@example.com", "password": "correctpassword"}))
	rr := httptest.NewRecorder()

	h.LoginHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("expected token in response")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	h, mock := newDB(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), 4)
	mock.ExpectQuery("SELECT password_hash").
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(string(hash)))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(map[string]string{"email": "user@example.com", "password": "wrongpassword"}))
	rr := httptest.NewRecorder()

	h.LoginHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestLoginHandler_InvalidEmail(t *testing.T) {
	h, _ := newDB(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(map[string]string{"email": "bademail", "password": "pass"}))
	rr := httptest.NewRecorder()

	h.LoginHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// --- LogoutHandler ---

func TestLogoutHandler_ClearsAuthCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rr := httptest.NewRecorder()

	auth.LogoutHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	cookies := rr.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "auth_token" {
			found = true
			if c.MaxAge != -1 {
				t.Errorf("auth_token MaxAge = %d, want -1", c.MaxAge)
			}
		}
	}
	if !found {
		t.Error("expected auth_token cookie in response")
	}
}

// --- UserDataHandler ---

func TestUserDataHandler_ValidCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: validToken(t)})
	rr := httptest.NewRecorder()

	auth.UserDataHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["user_id"] != "uid-1" {
		t.Errorf("user_id = %q, want 'uid-1'", resp["user_id"])
	}
	if resp["email"] != "a@b.com" {
		t.Errorf("email = %q, want 'a@b.com'", resp["email"])
	}
}

func TestUserDataHandler_NoCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rr := httptest.NewRecorder()

	auth.UserDataHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestUserDataHandler_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "invalid.token.here"})
	rr := httptest.NewRecorder()

	auth.UserDataHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
