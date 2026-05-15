package api

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
)

// --- escapeXML ---

func TestEscapeXML_Ampersand(t *testing.T) {
	got := escapeXML("a & b")
	want := "a &amp; b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEscapeXML_LessThan(t *testing.T) {
	got := escapeXML("a < b")
	want := "a &lt; b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEscapeXML_GreaterThan(t *testing.T) {
	got := escapeXML("a > b")
	want := "a &gt; b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEscapeXML_AllSpecial(t *testing.T) {
	got := escapeXML("<tag attr=\"x\" & y='z'>")
	if !strings.Contains(got, "&lt;") || !strings.Contains(got, "&gt;") || !strings.Contains(got, "&amp;") {
		t.Errorf("escapeXML(%q) = %q; missing expected escape sequences", "<tag & y>", got)
	}
}

func TestEscapeXML_NoSpecial(t *testing.T) {
	input := "hello world"
	got := escapeXML(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

// --- contentToWordXML ---

func TestContentToWordXML_EmptyLine(t *testing.T) {
	got := contentToWordXML("line1\n\nline3")
	if !strings.Contains(got, "<w:p/>") {
		t.Errorf("expected empty paragraph tag for blank line, got: %s", got)
	}
}

func TestContentToWordXML_ContentLine(t *testing.T) {
	got := contentToWordXML("hello")
	if !strings.Contains(got, "<w:t>hello</w:t>") {
		t.Errorf("expected <w:t>hello</w:t>, got: %s", got)
	}
}

func TestContentToWordXML_EscapesSpecialChars(t *testing.T) {
	got := contentToWordXML("a & b")
	if !strings.Contains(got, "&amp;") {
		t.Errorf("expected &amp; in output, got: %s", got)
	}
}

// --- generatePDF ---

func TestGeneratePDF_WritesContent(t *testing.T) {
	var buf bytes.Buffer
	err := generatePDF(&buf, "# Notes\n\nKey point 1\nKey point 2")
	if err != nil {
		t.Fatalf("generatePDF returned error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
	// PDF files start with "%PDF"
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("output does not start with PDF magic bytes")
	}
}

func TestGeneratePDF_EmptyContent(t *testing.T) {
	var buf bytes.Buffer
	err := generatePDF(&buf, "")
	if err != nil {
		t.Fatalf("generatePDF returned error for empty content: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output even for empty content")
	}
}

// --- generateDOCX ---

func TestGenerateDOCX_ValidZip(t *testing.T) {
	var buf bytes.Buffer
	err := generateDOCX(&buf, "Line one\nLine two")
	if err != nil {
		t.Fatalf("generateDOCX returned error: %v", err)
	}
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("output is not valid ZIP: %v", err)
	}
	files := make(map[string]bool)
	for _, f := range r.File {
		files[f.Name] = true
	}
	for _, required := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml"} {
		if !files[required] {
			t.Errorf("DOCX missing required entry: %s", required)
		}
	}
}

func TestGenerateDOCX_ContentAppearsInDocument(t *testing.T) {
	var buf bytes.Buffer
	_ = generateDOCX(&buf, "my unique content")

	r, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			data, _ := io.ReadAll(rc)
			_ = rc.Close()
			if !bytes.Contains(data, []byte("my unique content")) {
				t.Error("document.xml does not contain the input content")
			}
		}
	}
}

// --- isDocument ---

func TestIsDocument_PDF(t *testing.T) {
	if !isDocument("application/pdf") {
		t.Error("expected application/pdf to be a document")
	}
}

func TestIsDocument_DOCX(t *testing.T) {
	if !isDocument("application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
		t.Error("expected DOCX to be a document")
	}
}

func TestIsDocument_TXT(t *testing.T) {
	if !isDocument("text/plain") {
		t.Error("expected text/plain to be a document")
	}
}

func TestIsDocument_TXTWithCharset(t *testing.T) {
	if !isDocument("text/plain; charset=utf-8") {
		t.Error("expected text/plain; charset=utf-8 to be a document")
	}
}

func TestIsDocument_Video(t *testing.T) {
	if isDocument("video/mp4") {
		t.Error("expected video/mp4 to NOT be a document")
	}
}

func TestIsDocument_Audio(t *testing.T) {
	if isDocument("audio/mpeg") {
		t.Error("expected audio/mpeg to NOT be a document")
	}
}

// --- writeSuccessResp ---

func TestWriteSuccessResp_StatusAndBody(t *testing.T) {
	rr := httptest.NewRecorder()
	writeSuccessResp(rr)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "upload successful") {
		t.Errorf("body %q does not contain 'upload successful'", body)
	}
}

// --- validateUploadRequest ---

func TestValidateUploadRequest_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/uploads", nil)
	rr := httptest.NewRecorder()
	ok := validateUploadRequest(rr, req)
	if ok {
		t.Error("expected false for GET request, got true")
	}
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rr.Code)
	}
}

func TestValidateUploadRequest_ValidMultipart(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.txt"`)
	h.Set("Content-Type", "text/plain")
	part, _ := writer.CreatePart(h)
	_, _ = part.Write([]byte("file content"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	ok := validateUploadRequest(rr, req)
	if !ok {
		t.Errorf("expected true for valid multipart POST, got false; body: %s", rr.Body.String())
	}
}

// --- buildExportContent ---

func TestBuildExportContent_FallbackWhenNoHistory(t *testing.T) {
	got := buildExportContent(nil, "fallback content")
	if got != "fallback content" {
		t.Errorf("expected fallback, got %q", got)
	}
}

func TestBuildExportContent_InitialSheetOnly(t *testing.T) {
	history := []NoteHistory{
		{Role: "user", Prompt: "transcription text", Content: ""},
		{Role: "assistant", Prompt: "", Content: "initial study sheet"},
	}
	got := buildExportContent(history, "fallback")
	if !strings.Contains(got, "Study Sheet") {
		t.Error("missing Study Sheet header")
	}
	if !strings.Contains(got, "initial study sheet") {
		t.Error("missing initial study sheet content")
	}
	if strings.Contains(got, "transcription text") {
		t.Error("transcription should be excluded from export")
	}
}

func TestBuildExportContent_IncludesAllFollowUpResponses(t *testing.T) {
	history := []NoteHistory{
		{Role: "user", Prompt: "transcription text", Content: ""},
		{Role: "assistant", Prompt: "", Content: "initial sheet"},
		{Role: "user", Prompt: "follow-up question 1", Content: ""},
		{Role: "assistant", Prompt: "", Content: "response 1"},
		{Role: "user", Prompt: "follow-up question 2", Content: ""},
		{Role: "assistant", Prompt: "", Content: "response 2"},
	}
	got := buildExportContent(history, "fallback")

	for _, want := range []string{"initial sheet", "response 1", "response 2", "follow-up question 1", "follow-up question 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("export missing %q", want)
		}
	}
}

func TestBuildExportContent_OldFormatSingleRow(t *testing.T) {
	// Old format: user row holds both prompt (question) and content (AI response).
	history := []NoteHistory{
		{Role: "user", Prompt: "old question 1", Content: "old response 1"},
		{Role: "user", Prompt: "old question 2", Content: "old response 2"},
	}
	got := buildExportContent(history, "fallback")

	for _, want := range []string{"old question 1", "old response 1", "old question 2", "old response 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("export missing %q in old-format output:\n%s", want, got)
		}
	}
}

func TestBuildExportContent_FallbackWhenHistoryEmpty(t *testing.T) {
	got := buildExportContent([]NoteHistory{}, "my fallback")
	if got != "my fallback" {
		t.Errorf("expected fallback for empty history, got %q", got)
	}
}
