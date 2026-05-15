package notes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"AI-Note-Taker/notes"
)

func mockOpenAIServer(t *testing.T, content string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]string{"content": content}},
				},
			})
		}
	}))
}

func TestGenerateStudySheet_Success(t *testing.T) {
	srv := mockOpenAIServer(t, "# Study Sheet\n- Key point", http.StatusOK)
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	result, err := notes.GenerateStudySheet("lecture transcription text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "# Study Sheet\n- Key point" {
		t.Errorf("result.Content = %q, want study sheet content", result.Content)
	}
}

func TestGenerateStudySheet_ParsesTags(t *testing.T) {
	body := "TAGS: biology,lecture,exam-prep\n---\n# Study Sheet\n- Key point"
	srv := mockOpenAIServer(t, body, http.StatusOK)
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	result, err := notes.GenerateStudySheet("transcription text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "# Study Sheet\n- Key point" {
		t.Errorf("result.Content = %q, want study sheet content", result.Content)
	}
	if len(result.Tags) != 3 {
		t.Errorf("len(Tags) = %d, want 3; tags = %v", len(result.Tags), result.Tags)
	}
	if result.Tags[0] != "biology" {
		t.Errorf("Tags[0] = %q, want biology", result.Tags[0])
	}
}

func TestGenerateStudySheet_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{}})
	}))
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	result, err := notes.GenerateStudySheet("text")
	if err == nil {
		t.Errorf("expected error for empty choices, got nil (result=%+v)", result)
	}
}

func TestGenerateStudySheet_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	result, err := notes.GenerateStudySheet("text")
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil (result=%+v)", result)
	}
}

func TestRegenerateStudySheetWithHistory_Success(t *testing.T) {
	srv := mockOpenAIServer(t, "Updated sheet", http.StatusOK)
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	history := []notes.Message{
		{Role: "user", Content: "original transcription"},
		{Role: "assistant", Content: "original sheet"},
	}
	result, err := notes.RegenerateStudySheetWithHistory(history, "add more detail")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Updated sheet" {
		t.Errorf("result = %q, want 'Updated sheet'", result)
	}
}

func TestRegenerateStudySheetWithHistory_SendsFullHistory(t *testing.T) {
	var received notes.ChatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "response"}},
			},
		})
	}))
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	history := []notes.Message{
		{Role: "user", Content: "turn 1"},
		{Role: "assistant", Content: "response 1"},
	}
	notes.RegenerateStudySheetWithHistory(history, "new prompt")

	// system prompt + 2 history messages + new user prompt = 4
	if len(received.Messages) != 4 {
		t.Errorf("message count = %d, want 4", len(received.Messages))
	}
	if received.Messages[0].Role != "system" {
		t.Errorf("first message role = %q, want system", received.Messages[0].Role)
	}
	last := received.Messages[len(received.Messages)-1]
	if last.Role != "user" || last.Content != "new prompt" {
		t.Errorf("last message = {%s %s}, want {user new prompt}", last.Role, last.Content)
	}
}

func TestRegenerateStudySheetWithHistory_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{}})
	}))
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	_, err := notes.RegenerateStudySheetWithHistory(nil, "prompt")
	if err == nil {
		t.Error("expected error for empty choices, got nil")
	}
}

func TestRegenerateStudySheet_Deprecated(t *testing.T) {
	srv := mockOpenAIServer(t, "Sheet from deprecated", http.StatusOK)
	defer srv.Close()
	notes.OpenAIBaseURL = srv.URL

	result, err := notes.RegenerateStudySheet("existing notes", "update it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Sheet from deprecated" {
		t.Errorf("result = %q", result)
	}
}
