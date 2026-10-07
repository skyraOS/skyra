package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	desktop "skyra/controller/cua/observation"
	"strings"
	"testing"
)

func TestDesktopRefusalIsUnsuccessfulToolResult(t *testing.T) {
	for _, tc := range []struct {
		value   string
		success bool
	}{
		{`{"version":1,"operation":"desktop_observe","text":"42"}`, true},
		{`{"version":1,"operation":"desktop_observe","error":{"code":"outside_scope","message":"wrong window"}}`, false},
		{`{}`, false},
	} {
		result := toolResult(json.RawMessage(`"call"`), step{To: "desktop_observe", State: "return", Value: json.RawMessage(tc.value)})
		if got := result["result"].(map[string]any)["success"]; got != tc.success {
			t.Fatalf("%s: %v", tc.value, got)
		}
	}
}

func TestDesktopImageIsASeparateProviderContentItem(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 12, 12)))
	data := base64.StdEncoding.EncodeToString(b.Bytes())
	img, err := desktop.ScreenshotFrom(data, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	r := desktop.NewResult(desktop.Scope{})
	r.Text = "window text"
	r.Screenshot = img
	raw, _ := json.Marshal(r)
	result := toolResult(json.RawMessage(`"call"`), step{To: "desktop_observe", State: "return", Value: raw})["result"].(map[string]any)
	items := result["contentItems"].([]any)
	if result["success"] != true || len(items) != 2 {
		t.Fatalf("%+v", result)
	}
	text := items[0].(map[string]string)["text"]
	if strings.Contains(text, data) || !strings.Contains(text, "window text") {
		t.Fatal("image serialized as text")
	}
	if items[1].(map[string]string)["imageUrl"] != "data:image/png;base64,"+data {
		t.Fatal("provider image differs")
	}
}

func TestPenpotPreviewPreservesIdentityAndExactImage(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 12, 12)))
	data := base64.StdEncoding.EncodeToString(b.Bytes())
	img, err := desktop.ScreenshotFrom(data, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"version": 1, "operation": "design.tools", "file_id": "file1", "page_id": "page1", "screenshot": img})
	result := toolResult(json.RawMessage(`"call"`), step{To: "design.tools", State: "return", Value: raw})["result"].(map[string]any)
	items := result["contentItems"].([]any)
	if result["success"] != true || len(items) != 2 {
		t.Fatalf("%+v", result)
	}
	text := items[0].(map[string]string)["text"]
	if !strings.Contains(text, "file1") || strings.Contains(text, data) {
		t.Fatal("identity or image projection incorrect")
	}
	if items[1].(map[string]string)["imageUrl"] != "data:image/png;base64,"+data {
		t.Fatal("export bytes changed")
	}
}

func TestPenpotErrorDetailsReachTheModel(t *testing.T) {
	raw := json.RawMessage(`{"version":1,"operation":"design.tools","error":{"code":"upstream_tool_error","message":"TypeError: invalid resize argument","stage":"tool_call","tool":"execute_code","operation":"execute","edit_outcome":"unknown"}}`)
	result := toolResult(json.RawMessage(`"call"`), step{To: "design.tools", State: "return", Value: raw})["result"].(map[string]any)
	if result["success"] != false {
		t.Fatal("failure reported as success")
	}
	text := result["contentItems"].([]any)[0].(map[string]string)["text"]
	if !strings.Contains(text, "invalid resize argument") || !strings.Contains(text, `"edit_outcome":"unknown"`) || !strings.Contains(text, `"stage":"tool_call"`) {
		t.Fatalf("error details lost: %s", text)
	}
}
