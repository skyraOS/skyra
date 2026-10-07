package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/cua/observation"
)

func TestBridgeRejectsInputAndOtherWindowsBeforeConnecting(t *testing.T) {
	scope := observation.Scope{PID: 7, WindowID: 9}
	for _, tc := range []struct{ input, code string }{
		{`{"pid":7,"window_id":9,"tool":"click"}` + "\n", "invalid_request"},
		{`{"pid":7,"window_id":10}` + "\n", "outside_scope"},
		{`{"pid":8,"window_id":9}` + "\n", "outside_scope"},
		{`{"pid":7,"window_id":9,"screenshot_out_file":"/tmp/x"}` + "\n", "invalid_request"},
		{`{"pid":7,"window_id":9}`, "invalid_request"},
	} {
		out := handle(context.Background(), strings.NewReader(tc.input), "/does-not-exist", scope)
		if out.Error == nil || out.Error.Code != tc.code {
			t.Fatalf("%s: %+v", tc.input, out)
		}
	}
}

func TestBridgeCannotAttachWithWidenedManifest(t *testing.T) {
	expected, err := observation.Manifest("com.example.Fixture")
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(expected, &value); err != nil {
		t.Fatal(err)
	}
	value["allow"].(map[string]any)["tools"] = []string{"get_window_state", "click"}
	widened, _ := json.Marshal(value)
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, widened, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readManifest(path, expected); err == nil {
		t.Fatal("accepted click grant")
	}
	if err := os.WriteFile(path, append(expected, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := readManifest(path, expected)
	if err != nil || string(b) != string(append(expected, '\n')) {
		t.Fatalf("exact bytes lost: %s: %v", b, err)
	}
}
