package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCodexModels(t *testing.T) {
	dir := t.TempDir()
	cache := map[string]any{
		"client_version": "0.160.0",
		"models": []map[string]any{
			{"slug": "gpt-4o", "display_name": "GPT-4o", "description": "Fast and capable", "context_window": 128000, "max_context_window": 128000},
			{"slug": "o3", "display_name": "o3", "description": "Reasoning", "context_window": 200000, "max_context_window": 200000},
		},
	}
	b, _ := json.Marshal(cache)
	codexDir := filepath.Join(dir, ".codex")
	os.MkdirAll(codexDir, 0o700)
	os.WriteFile(filepath.Join(codexDir, "models_cache.json"), b, 0o644)

	t.Setenv("HOME", dir)
	r := codexModels(result{Provider: "codex"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if len(r.Models) != 2 {
		t.Fatalf("want 2 models, got %d", len(r.Models))
	}
	if r.Models[0].ID != "gpt-4o" || r.Models[0].Name != "GPT-4o" {
		t.Errorf("unexpected first model: %+v", r.Models[0])
	}
	if r.Models[0].Context != 128000 {
		t.Errorf("want context 128000, got %d", r.Models[0].Context)
	}
}

func TestCodexModelsMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	r := codexModels(result{Provider: "codex"})
	if r.Error == "" {
		t.Fatal("expected error for missing cache")
	}
}

func TestClaudeModels(t *testing.T) {
	dir := t.TempDir()
	catalog := map[string]any{
		"version": 2,
		"catalog": map[string]any{
			"config": map[string]any{
				"models": []map[string]any{
					{"id": "claude-opus-5-5", "name": "Opus 5.5", "description": "Complex work"},
					{"id": "claude-sonnet-5-5", "name": "Sonnet 5.5", "description": "Efficient"},
				},
			},
		},
	}
	b, _ := json.Marshal(catalog)
	catDir := filepath.Join(dir, ".claude", "cache", "model-catalog")
	os.MkdirAll(catDir, 0o700)
	os.WriteFile(filepath.Join(catDir, "test-uuid-cc.json"), b, 0o644)

	t.Setenv("HOME", dir)
	r := claudeModels(result{Provider: "claude"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if len(r.Models) != 2 {
		t.Fatalf("want 2 models, got %d", len(r.Models))
	}
	if r.Models[0].ID != "claude-opus-5-5" || r.Models[0].Name != "Opus 5.5" {
		t.Errorf("unexpected first model: %+v", r.Models[0])
	}
}

func TestClaudeModelsMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	r := claudeModels(result{Provider: "claude"})
	if r.Error == "" {
		t.Fatal("expected error for missing catalog")
	}
}

func TestCodexUsesMaxContext(t *testing.T) {
	dir := t.TempDir()
	cache := map[string]any{
		"client_version": "0.160.0",
		"models": []map[string]any{
			{"slug": "test", "display_name": "Test", "context_window": 128000, "max_context_window": 256000},
		},
	}
	b, _ := json.Marshal(cache)
	codexDir := filepath.Join(dir, ".codex")
	os.MkdirAll(codexDir, 0o700)
	os.WriteFile(filepath.Join(codexDir, "models_cache.json"), b, 0o644)

	t.Setenv("HOME", dir)
	r := codexModels(result{Provider: "codex"})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Models[0].Context != 256000 {
		t.Errorf("want max_context_window 256000, got %d", r.Models[0].Context)
	}
}
