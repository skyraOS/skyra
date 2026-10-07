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
		out := handle(context.Background(), strings.NewReader(tc.input), unreachable(t), scope)
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

func TestAppInventoryListsOnlyPinnedWindowsAndDeniesOtherIDs(t *testing.T) {
	scopes := []observation.Scope{{BundleID: "com.example.One", PID: 7, WindowID: 9, AppName: "One"}, {BundleID: "com.example.Two", PID: 8, WindowID: 10, AppName: "Two"}}
	out := handleScopes(context.Background(), strings.NewReader("{}\n"), unreachable(t), scopes, true)
	if out.Error != nil || len(out.Windows) != 2 || out.Windows[1].BundleID != "com.example.Two" || out.Screenshot != nil {
		t.Fatalf("%+v", out)
	}
	out = handleScopes(context.Background(), strings.NewReader("{\"pid\":99,\"window_id\":99}\n"), unreachable(t), scopes, true)
	if out.Error == nil || out.Error.Code != "outside_scope" {
		t.Fatalf("%+v", out)
	}
}

func TestWindowInventoryMustMatchManifestApps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "windows.json")
	common := observation.Scope{DriverVersion: "0.34.0", ManifestHash: strings.Repeat("a", 64), BundleID: "com.example.One,com.example.Two"}
	os.WriteFile(path, []byte(`[{"bundle_id":"com.example.One","pid":7,"window_id":9},{"bundle_id":"com.example.Two","pid":8,"window_id":10}]`), 0600)
	scopes, err := readScopes(path, common)
	if err != nil || len(scopes) != 2 || scopes[1].ManifestHash != common.ManifestHash {
		t.Fatalf("%+v: %v", scopes, err)
	}
	os.WriteFile(path, []byte(`[{"bundle_id":"com.example.Other","pid":7,"window_id":9}]`), 0600)
	if _, err := readScopes(path, common); err == nil {
		t.Fatal("accepted an unapproved app")
	}
}
