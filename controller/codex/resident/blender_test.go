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

func TestBlenderPreviewReachesProviderAsExactImage(t *testing.T) {
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 12, 12))); err != nil {
		t.Fatal(err)
	}
	data := base64.StdEncoding.EncodeToString(b.Bytes())
	img, err := desktop.ScreenshotFrom(data, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"version": 1, "operation": "blender.tools", "scene_version": "scene-a", "candidate_status": "unreviewed", "screenshot": img})
	result := toolResult(json.RawMessage(`"call"`), step{To: "blender.tools", State: "return", Value: raw})["result"].(map[string]any)
	items := result["contentItems"].([]any)
	if result["success"] != true || len(items) != 2 {
		t.Fatalf("%+v", result)
	}
	text := items[0].(map[string]string)["text"]
	if strings.Contains(text, data) || !strings.Contains(text, "scene-a") || !strings.Contains(text, "unreviewed") {
		t.Fatal("preview metadata lost or image serialized as text")
	}
	if items[1].(map[string]string)["imageUrl"] != "data:image/png;base64,"+data {
		t.Fatal("preview bytes changed")
	}
}

func TestBlenderConflictDetailsReachProvider(t *testing.T) {
	raw := json.RawMessage(`{"version":1,"operation":"blender.tools","error":{"code":"scene_version_conflict","stage":"context","actual_scene_version":"new-version","message":"Inspect context before another edit."}}`)
	result := toolResult(json.RawMessage(`"call"`), step{To: "blender.tools", State: "return", Value: raw})["result"].(map[string]any)
	if result["success"] != false {
		t.Fatal("conflict reported as success")
	}
	text := result["contentItems"].([]any)[0].(map[string]string)["text"]
	if !strings.Contains(text, "scene_version_conflict") || !strings.Contains(text, "new-version") {
		t.Fatal("conflict details lost")
	}
}
