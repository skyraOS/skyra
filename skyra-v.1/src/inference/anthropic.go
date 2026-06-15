package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"skyra-v1/src/keychain"
)

const (
	anthropicURL     = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	defaultAnthModel = "claude-sonnet-4-5-20241022"
)

type anthropicRequest struct {
	Model     string             `json:"model"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMsg     `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func CallAnthropic(system, present string) (string, error) {
	apiKey := keychain.Get("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = defaultAnthModel
	}

	payload := anthropicRequest{
		Model:  model,
		System: system,
		Messages: []anthropicMsg{
			{Role: "user", Content: present},
		},
		MaxTokens: 4096,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

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
		return "", fmt.Errorf("anthropic HTTP %d: %s", resp.StatusCode, raw)
	}

	var anthResp anthropicResponse
	if err := json.Unmarshal(raw, &anthResp); err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}

	if anthResp.Error != nil {
		return "", fmt.Errorf("anthropic: %s", anthResp.Error.Message)
	}

	var texts []string
	for _, block := range anthResp.Content {
		if block.Type == "text" {
			texts = append(texts, block.Text)
		}
	}

	if len(texts) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return strings.TrimSpace(strings.Join(texts, "")), nil
}
