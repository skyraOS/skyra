package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

var EmbedEndpoint = "http://10.0.0.50:11434/api/embeddings"
var EmbedModel = "nomic-embed-text"

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func Embed(text string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{Model: EmbedModel, Prompt: text})
	if err != nil {
		return nil, fmt.Errorf("embed marshal: %w", err)
	}

	resp, err := http.Post(EmbedEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embed read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed HTTP %d: %s", resp.StatusCode, raw)
	}

	var result embedResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("embed parse: %w", err)
	}

	return result.Embedding, nil
}

func Cosine(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	d := math.Sqrt(normA) * math.Sqrt(normB)
	if d == 0 {
		return 0
	}
	return dot / d
}
