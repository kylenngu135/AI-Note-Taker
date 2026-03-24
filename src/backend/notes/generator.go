package notes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// the request body we send to Groq
type GroqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// the response we get back from Groq
type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GenerateStudySheet(transcription string) (string, error) {
	// build the request body
	reqBody := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []Message{
			{
				Role: "system",
				Content: `You are an expert study assistant. When given a transcription or document, 
				you generate a comprehensive study sheet that includes:
				- A brief summary
				- Key concepts and definitions
				- Main topics covered
				- Important facts to remember
				- A to-do list of action items if any were mentioned
				Format your response in clear markdown.`,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Please create a study sheet from the following transcription:\n\n%s", transcription),
			},
		},
	}

	// convert request body to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// create the HTTP request
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// decode the response
	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from groq")
	}

	return groqResp.Choices[0].Message.Content, nil
}
