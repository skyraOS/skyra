package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"skyra-v1/src/keychain"
)

const (
	baseURL      = "https://openrouter.ai/api/v1/chat/completions"
	defaultModel = "anthropic/claude-sonnet-4-5"
	timeout      = 120 * time.Second
)

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func Call(system, present string) (string, error) {
	apiKey := keychain.Get("OpenRouter_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = defaultModel
	}

	payload := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: present},
		},
		Temperature: 0.2,
		MaxTokens:   4096,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("inference HTTP %d: %s", resp.StatusCode, raw)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(raw, &chatResp); err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
