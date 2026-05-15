package notes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type StudySheetResult struct {
	Content string
	Tags    []string
}

// represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// the request body we send to OpenAI
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// the response we get back from OpenAI
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// OpenAIBaseURL is the base URL for the OpenAI API. Override in tests.
var OpenAIBaseURL = "https://api.openai.com"

const systemPrompt = `You are an expert study assistant. Your job is to transform raw transcriptions into a clean, structured, and highly useful notes sheet.

When given a transcription, produce a notes sheet with the following sections:

**Summary**
A concise 3-5 sentence overview of the entire content. Capture the core topic, purpose, and key takeaway.

**Key Concepts**
A list of the most important ideas, terms, definitions, or frameworks introduced in the transcription. For each concept, include a brief explanation.

**Detailed Notes**
A structured breakdown of the content in the order it was presented. Use headers for major topics and bullet points for supporting details. Preserve important context and explanations — do not over-compress.

**Important Quotes or Statements**
Any particularly significant statements, conclusions, or phrasing from the transcription worth remembering verbatim or near-verbatim.

**Action Items or Takeaways**
A list of practical takeaways, things to follow up on, questions raised, or actions implied by the content. If none apply, omit this section.

Be thorough. A student should be able to study entirely from this notes sheet without needing to refer back to the original transcription.`

func GenerateStudySheet(transcription string) (StudySheetResult, error) {
	reqBody := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Please create a notes sheet from the following transcription:\n\n%s", transcription)},
		},
	}

	raw, err := callOpenAI(reqBody)
	if err != nil {
		return StudySheetResult{}, err
	}

	return parseStudySheetResponse(raw), nil
}

// RegenerateStudySheetWithHistory regenerates a study sheet using the full conversation history.
// conversationHistory is a slice of Message structs containing all previous user prompts and assistant responses.
func RegenerateStudySheetWithHistory(conversationHistory []Message, newPrompt string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: `You are an expert study assistant. Generate a comprehensive notes sheet that includes:
- A brief summary
- Key concepts and definitions
- Main topics covered
- Important facts to remember
- Action items or takeaways if any were mentioned
Format your response in clear markdown.`,
		},
	}
	messages = append(messages, conversationHistory...)
	messages = append(messages, Message{Role: "user", Content: newPrompt})

	reqBody := ChatRequest{
		Model:    "gpt-4o-mini",
		Messages: messages,
	}

	return callOpenAI(reqBody)
}

func callOpenAI(reqBody ChatRequest) (string, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", OpenAIBaseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func parseStudySheetResponse(raw string) StudySheetResult {
	idx := strings.Index(raw, "\n---\n")
	if idx == -1 {
		if strings.HasPrefix(raw, "---\n") {
			return StudySheetResult{Content: strings.TrimSpace(raw[4:])}
		}
		return StudySheetResult{Content: strings.TrimSpace(raw)}
	}
	header := strings.TrimSpace(raw[:idx])
	content := strings.TrimSpace(raw[idx+5:])
	return StudySheetResult{Content: content, Tags: extractTagsFromLine(header)}
}

func extractTagsFromLine(line string) []string {
	for _, prefix := range []string{"TAGS:", "Tags:", "tags:"} {
		if strings.HasPrefix(line, prefix) {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			var tags []string
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				t = strings.TrimPrefix(t, "#")
				t = strings.ToLower(t)
				t = strings.ReplaceAll(t, " ", "-")
				if t != "" {
					tags = append(tags, t)
				}
			}
			return tags
		}
	}
	return nil
}

// RegenerateStudySheet is deprecated - use RegenerateStudySheetWithHistory instead.
func RegenerateStudySheet(existingNotes, prompt string) (string, error) {
	return RegenerateStudySheetWithHistory([]Message{
		{Role: "user", Content: fmt.Sprintf("Please create a notes sheet from the following transcription:\n\n%s", existingNotes)},
	}, prompt)
}
