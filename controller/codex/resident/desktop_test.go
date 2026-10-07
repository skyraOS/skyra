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
