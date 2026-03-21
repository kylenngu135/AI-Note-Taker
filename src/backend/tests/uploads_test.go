package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"multipart/form-data"
	"ai-note-taker/api"
)

// create a minimal in-memory pdf instead of reading from disk
func mockpdf() []byte {
    return []byte("%pdf-1.4 mock pdf content")
}

func testuploadpdfhandlermock(t *testing.t) {
    var body bytes.buffer
    writer := multipart.newwriter(&body)

    part, _ := writer.createformfile("file", "test.pdf")
    part.write(mockpdf()) // write mock bytes instead of a real file

    writer.close()

    req, _ := http.newrequest("post", "/upload", &body)
    req.header.set("content-type", writer.formdatacontenttype())

    rr := httptest.newrecorder()
    http.handlerfunc(uploadpdfhandler).servehttp(rr, req)

    if rr.code != http.statusok {
        t.errorf("expected 200, got %v", rr.code)
    }
}

// note: this is a unit test, learn integration testing with mock functions and depedency injections for api testing

/*
func testuploaddochandler(t *testing.t) {
	file, err := os.open("tesetdata/sample.pdf")
	if err != nil {
		t.fatal(err)
	}
	defer file.close()

	var body bytes.buffer
	writer := multipart.newwriter(&body)

	part, err := writer.createformfile("file", filepath.base(file.name()))
	if err != nil {
		t.fatal(err)
	}
	writer.close()

	// todo: define a file
	req, err := http.newrequest("post", "/api/uploads/documents", &body)
	if err != nil {
		t.fatal(err)
	}

	// call the handler directly
	rr := httptest.newrecorder()
	handler := http.handlerfunc(createuserhandler)
	handler.servehttp(rr, req)

	if rr.code != http.statuscreated {
		t.errorf("expected status 201, got %v", rr.code)
	}
}
*/
