package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"AI-Note-Taker/api"
	"AI-Note-Taker/middleware"
)

// serveWithMux runs a handler through a real ServeMux so path values ({id}) are parsed.
func serveWithMux(pattern string, handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.Handle(pattern, handler)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func newHandler(t *testing.T) (*api.Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &api.Handler{DB: db}, mock
}

func authedRequest(t *testing.T, method, target string, body *bytes.Buffer) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, body)
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	req.AddCookie(authCookie(t))
	return req
}

// --- DocsHandler / OpenAPISpecHandler ---

func TestDocsHandler_ReturnsHTML(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api-docs", nil)
	api.DocsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rr.Body.String(), "<redoc") {
		t.Error("response body should contain <redoc> element")
	}
}

func TestOpenAPISpecHandler_ReturnsYAML(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api-docs/openapi.yaml", nil)
	api.OpenAPISpecHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("Content-Type = %q, want application/yaml", ct)
	}
	if !strings.Contains(rr.Body.String(), "openapi:") {
		t.Error("response body should contain OpenAPI spec content")
	}
}

// --- GetUploadsHandler ---

func TestGetUploadsHandler_Success(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()

	mock.ExpectQuery(`json_agg`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "filename", "created_at", "tags"}).
			AddRow("uid-1", "lecture.pdf", now, []byte("[]")).
			AddRow("uid-2", "notes.txt", now, []byte("[]")))

	req := authedRequest(t, http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.GetUploadsHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("got %d uploads, want 2", len(result))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetUploadsHandler_EmptyList(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery(`json_agg`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "filename", "created_at", "tags"}))

	req := authedRequest(t, http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.GetUploadsHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestGetUploadsHandler_DBError(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery(`json_agg`).
		WillReturnError(sql.ErrConnDone)

	req := authedRequest(t, http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.GetUploadsHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestGetUploadsHandler_NoAuth(t *testing.T) {
	h, _ := newHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.GetUploadsHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

// --- DeleteUploadHandler ---

func deleteUploadRows(t *testing.T) (*sqlmock.Rows, *sqlmock.Rows, *sqlmock.Rows) {
	t.Helper()
	now := time.Now()
	uploadRows := sqlmock.NewRows([]string{
		"id", "filename", "file_type", "file_size", "storage_key", "status", "created_at", "last_updated_at",
	}).AddRow("del-id", "file.txt", "text/plain", int64(10), "transcriptions/del-id.txt", "complete", now, now)

	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "del-id", "content", "transcriptions/note-id.txt", now, now)

	historyRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	})
	return uploadRows, noteRows, historyRows
}

func TestDeleteUploadHandler_Success(t *testing.T) {
	h, mock := newHandler(t)
	uploadRows, noteRows, histRows := deleteUploadRows(t)

	mock.ExpectQuery("SELECT id, filename").WithArgs("del-id").WillReturnRows(uploadRows)
	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("del-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id").WithArgs("del-id").WillReturnRows(histRows)
	mock.ExpectExec("DELETE FROM note_history").WithArgs("del-id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM notes").WithArgs("del-id").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM uploads").WithArgs("del-id").WillReturnResult(sqlmock.NewResult(0, 1))

	req := authedRequest(t, http.MethodDelete, "/api/uploads/del-id", nil)
	rr := serveWithMux(
		"DELETE /api/uploads/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(h.DeleteUploadHandler)),
		req,
	)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204; body: %s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteUploadHandler_UploadNotFound(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery("SELECT id, filename").
		WithArgs("missing-id").
		WillReturnError(sql.ErrNoRows)

	req := authedRequest(t, http.MethodDelete, "/api/uploads/missing-id", nil)
	rr := serveWithMux(
		"DELETE /api/uploads/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(h.DeleteUploadHandler)),
		req,
	)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

// --- GetNoteByUploadIDHandler ---

func noteWithHistoryRows(t *testing.T) (*sqlmock.Rows, *sqlmock.Rows) {
	t.Helper()
	now := time.Now()
	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "# Study Sheet", "key", now, now)

	histRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	}).
		AddRow("h1", "note-id", "upload-id", "user", "transcript", "", "k1", now).
		AddRow("h2", "note-id", "upload-id", "assistant", "", "# Study Sheet", "k2", now)

	return noteRows, histRows
}

func TestGetNoteByUploadIDHandler_JSONResponse(t *testing.T) {
	h, mock := newHandler(t)
	noteRows, histRows := noteWithHistoryRows(t)

	// GetNoteByUploadIDHandler calls GetNoteWithHistoryByUploadID once (notes query + history query).
	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("upload-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id").WithArgs("upload-id").WillReturnRows(histRows)

	req := authedRequest(t, http.MethodGet, "/api/uploads/upload-id/notes", nil)
	rr := serveWithMux(
		"GET /api/uploads/{id}/notes",
		middleware.AuthMiddleware(http.HandlerFunc(h.GetNoteByUploadIDHandler)),
		req,
	)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["note"] == nil {
		t.Error("response missing 'note' field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetNoteByUploadIDHandler_TXTDownload(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()
	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "Plain text content", "key", now, now)
	histRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	})
	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("upload-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id").WithArgs("upload-id").WillReturnRows(histRows)

	req := authedRequest(t, http.MethodGet, "/api/uploads/upload-id/notes?format=txt", nil)
	rr := serveWithMux(
		"GET /api/uploads/{id}/notes",
		middleware.AuthMiddleware(http.HandlerFunc(h.GetNoteByUploadIDHandler)),
		req,
	)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	// Empty history → buildExportContent falls back to note.Content.
	if rr.Body.String() != "Plain text content" {
		t.Errorf("body = %q, want 'Plain text content'", rr.Body.String())
	}
}

func TestGetNoteByUploadIDHandler_PDFDownload(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()
	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "PDF content here", "key", now, now)
	histRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	})
	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("upload-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id").WithArgs("upload-id").WillReturnRows(histRows)

	req := authedRequest(t, http.MethodGet, "/api/uploads/upload-id/notes?format=pdf", nil)
	rr := serveWithMux(
		"GET /api/uploads/{id}/notes",
		middleware.AuthMiddleware(http.HandlerFunc(h.GetNoteByUploadIDHandler)),
		req,
	)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
}

func TestGetNoteByUploadIDHandler_DOCXDownload(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()
	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "DOCX content", "key", now, now)
	histRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	})
	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("upload-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id").WithArgs("upload-id").WillReturnRows(histRows)

	req := authedRequest(t, http.MethodGet, "/api/uploads/upload-id/notes?format=docx", nil)
	rr := serveWithMux(
		"GET /api/uploads/{id}/notes",
		middleware.AuthMiddleware(http.HandlerFunc(h.GetNoteByUploadIDHandler)),
		req,
	)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestGetNoteByUploadIDHandler_DBError(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery("SELECT id, upload_id, content").
		WithArgs("upload-id").
		WillReturnError(sql.ErrConnDone)

	req := authedRequest(t, http.MethodGet, "/api/uploads/upload-id/notes", nil)
	rr := serveWithMux(
		"GET /api/uploads/{id}/notes",
		middleware.AuthMiddleware(http.HandlerFunc(h.GetNoteByUploadIDHandler)),
		req,
	)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

// --- RegenerateNoteHandler ---

func TestRegenerateNoteHandler_Success(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()

	noteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "Original sheet", "old-key", now, now)

	histRows := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	})

	histInsertRow := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	}).AddRow("h1", "note-id", "upload-id", "user", "expand this", "", "", now)

	histInsertRow2 := sqlmock.NewRows([]string{
		"id", "note_id", "upload_id", "role", "prompt", "content", "storage_key", "created_at",
	}).AddRow("h2", "note-id", "upload-id", "assistant", "", "Generated study sheet", "new-key", now)

	updatedNoteRows := sqlmock.NewRows([]string{
		"id", "upload_id", "content", "storage_key", "created_at", "last_updated_at",
	}).AddRow("note-id", "upload-id", "Generated study sheet", "new-key", now, now)

	mock.ExpectQuery("SELECT id, upload_id, content").WithArgs("upload-id").WillReturnRows(noteRows)
	mock.ExpectQuery("SELECT id, note_id, upload_id, role").WithArgs("upload-id").WillReturnRows(histRows)
	mock.ExpectQuery("INSERT INTO note_history").WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		"user", "expand this", "", "",
	).WillReturnRows(histInsertRow)
	mock.ExpectQuery("INSERT INTO note_history").WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		"assistant", "", sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnRows(histInsertRow2)
	mock.ExpectQuery("UPDATE notes").WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnRows(updatedNoteRows)

	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(map[string]string{"prompt": "expand this"})

	req := authedRequest(t, http.MethodPost, "/api/uploads/upload-id/notes/regenerate", body)
	req.Header.Set("Content-Type", "application/json")
	rr := serveWithMux(
		"POST /api/uploads/{id}/notes/regenerate",
		middleware.AuthMiddleware(http.HandlerFunc(h.RegenerateNoteHandler)),
		req,
	)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["content"] == nil {
		t.Error("response missing 'content' field")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRegenerateNoteHandler_InvalidBody(t *testing.T) {
	h, _ := newHandler(t)

	req := authedRequest(t, http.MethodPost, "/api/uploads/upload-id/notes/regenerate",
		bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := serveWithMux(
		"POST /api/uploads/{id}/notes/regenerate",
		middleware.AuthMiddleware(http.HandlerFunc(h.RegenerateNoteHandler)),
		req,
	)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestRegenerateNoteHandler_NoteNotFound(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery("SELECT id, upload_id, content").
		WithArgs("upload-id").
		WillReturnError(sql.ErrNoRows)

	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(map[string]string{"prompt": "expand"})

	req := authedRequest(t, http.MethodPost, "/api/uploads/upload-id/notes/regenerate", body)
	req.Header.Set("Content-Type", "application/json")
	rr := serveWithMux(
		"POST /api/uploads/{id}/notes/regenerate",
		middleware.AuthMiddleware(http.HandlerFunc(h.RegenerateNoteHandler)),
		req,
	)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

// --- UploadHandler ---

func txtMultipart(t *testing.T, content string) (*bytes.Buffer, string) {
	t.Helper()
	// http.DetectContentType reads up to 512 bytes; pad short content with spaces so
	// the sniffed MIME type is "text/plain; charset=utf-8" rather than octet-stream.
	for len(content) < 520 {
		content += " "
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.txt"`)
	h.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(content))
	_ = writer.Close()
	return &body, writer.FormDataContentType()
}

func TestUploadHandler_NoFile(t *testing.T) {
	h, _ := newHandler(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("other", "value")
	_ = writer.Close()

	req := authedRequest(t, http.MethodPost, "/api/uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestUploadHandler_UnsupportedFileType(t *testing.T) {
	h, _ := newHandler(t)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Disposition", `form-data; name="file"; filename="image.png"`)
	mh.Set("Content-Type", "image/png")
	part, _ := writer.CreatePart(mh)
	_, _ = part.Write([]byte("PNG data"))
	_ = writer.Close()

	req := authedRequest(t, http.MethodPost, "/api/uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", rr.Code)
	}
}

func TestUploadHandler_TXTSuccess(t *testing.T) {
	h, mock := newHandler(t)
	now := time.Now()

	// Fake Upstash Redis: any POST returns a successful LPUSH result.
	fakeRedis := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": 1})
	}))
	defer fakeRedis.Close()
	t.Setenv("REDIS_URL", fakeRedis.URL)

	// insertUploadPending: 5 args — status 'pending' is hardcoded in the SQL, not a parameter.
	mock.ExpectQuery("INSERT INTO uploads").
		WithArgs(sqlmock.AnyArg(), "test.txt", "text/plain", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "filename", "file_type", "file_size", "storage_key", "status", "created_at",
		}).AddRow("upload-uuid", "test.txt", "text/plain", int64(520), "raw/upload-uuid", "pending", now))

	// InsertJobRecord: INSERT INTO jobs (id, status, file_key) VALUES ($1, 'pending', $2)
	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, ct := txtMultipart(t, "Study this content please.")
	req := authedRequest(t, http.MethodPost, "/api/uploads", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["job_id"] == "" {
		t.Errorf("response missing 'job_id'")
	}
	if resp["upload_id"] == "" {
		t.Errorf("response missing 'upload_id'")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUploadHandler_NoAuth(t *testing.T) {
	h, _ := newHandler(t)
	body, ct := txtMultipart(t, "content")
	req := httptest.NewRequest(http.MethodPost, "/api/uploads", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestUploadHandler_DBError(t *testing.T) {
	h, mock := newHandler(t)

	mock.ExpectQuery("INSERT INTO uploads").
		WillReturnError(sql.ErrConnDone)

	body, ct := txtMultipart(t, "Study material.")
	req := authedRequest(t, http.MethodPost, "/api/uploads", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

// TestUploadHandler_WrongMethod verifies the method guard inside validateUploadRequest.
func TestUploadHandler_WrongMethod(t *testing.T) {
	h, _ := newHandler(t)
	req := authedRequest(t, http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	middleware.AuthMiddleware(http.HandlerFunc(h.UploadHandler)).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rr.Code)
	}
}
