package api

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

// --- mock file helpers ---

func mockPDF() []byte {
	// DetectContentType returns "application/pdf"
	return []byte(
		"%PDF-1.4\n" +
			"1 0 obj\n" +
			"<< /Type /Catalog >>\n" +
			"endobj\n" +
			"%%EOF",
	)
}

func mockDOCX() []byte {
	// DetectContentType returns "application/zip" for DOCX
	// We set declared Content-Type to "application/zip" to match
	return []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00}
}

func mockTXT() []byte {
	content := "This is a plain text file for testing. "
	// repeat until we have 512+ bytes
	for len(content) < 512 {
		content += "Adding more plain text content to ensure correct MIME detection. "
	}
	return []byte(content)
}

func mockMP4() []byte {
	// DetectContentType returns "video/mp4"
	return []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32}
}

func mockMP3() []byte {
	// http.DetectContentType only detects audio/mpeg via ID3 header, not raw MPEG frame sync bytes
	// ID3 header: "ID3" + version + flags + size
	header := []byte{
		0x49, 0x44, 0x33, // "ID3"
		0x03, 0x00, // version 2.3.0
		0x00,                   // flags
		0x00, 0x00, 0x00, 0x00, // size
	}
	padded := make([]byte, 512)
	copy(padded, header)
	return padded
}

// --- shared upload test function ---
func runUploadTest(t *testing.T, endpoint string, cases []struct {
	name         string
	filename     string
	contentType  string
	fileBytes    []byte
	expectedCode int
}) {
	t.Helper()

	// select the correct handler based on the endpoint
	var handler http.HandlerFunc
	switch endpoint {
	case "/upload/documents":
		handler = http.HandlerFunc(DocumentUploadHandler)
	case "/upload/video":
		handler = http.HandlerFunc(VideoUploadHandler)
	case "/upload/audio":
		handler = http.HandlerFunc(AudioUploadHandler)
	default:
		t.Fatalf("unknown endpoint: %s", endpoint)
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)

			part, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Disposition": []string{
					fmt.Sprintf(`form-data; name="file"; filename="%s"`, tt.filename),
				},
				"Content-Type": []string{tt.contentType},
			})
			if err != nil {
				t.Fatal(err)
			}

			part.Write(tt.fileBytes)
			writer.Close()

			req, err := http.NewRequest("POST", endpoint, &body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("test %q: expected %v, got %v: %v", tt.name, tt.expectedCode, rr.Code, rr.Body.String())
			}
		})
	}
}

// --- individual endpoint tests ---

func TestDocumentUpload(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		contentType  string
		fileBytes    []byte
		expectedCode int
	}{
		{
			name:         "valid PDF",
			filename:     "test.pdf",
			contentType:  "application/pdf",
			fileBytes:    mockPDF(),
			expectedCode: http.StatusCreated,
		},
		{
			name:         "valid DOCX",
			filename:     "test.docx",
			contentType:  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			fileBytes:    mockDOCX(),
			expectedCode: http.StatusCreated,
		},
		{
			name:         "valid TXT",
			filename:     "test.txt",
			contentType:  "text/plain",
			fileBytes:    mockTXT(),
			expectedCode: http.StatusCreated,
		},
	}

	runUploadTest(t, "/upload/documents", cases)
}

func TestVideoUpload(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		contentType  string
		fileBytes    []byte
		expectedCode int
	}{
		{
			name:         "valid MP4",
			filename:     "test.mp4",
			contentType:  "video/mp4",
			fileBytes:    mockMP4(),
			expectedCode: http.StatusCreated,
		},
	}

	runUploadTest(t, "/upload/video", cases)
}

func TestAudioUpload(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		contentType  string
		fileBytes    []byte
		expectedCode int
	}{
		{
			name:         "valid MP3",
			filename:     "test.mp3",
			contentType:  "audio/mpeg",
			fileBytes:    mockMP3(),
			expectedCode: http.StatusCreated,
		},
	}

	runUploadTest(t, "/upload/audio", cases)
}
