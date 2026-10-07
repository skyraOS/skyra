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

func TestDesktopPreviewUsesRecordedBytesAndPreservesFailures(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 16, 16)))
	data := base64.StdEncoding.EncodeToString(b.Bytes())
	shot, err := desktop.ScreenshotFrom(data, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	r := desktop.NewResult(desktop.Scope{})
	r.Screenshot = shot
	r.Degraded = true
	r.Truncated = true
	raw, _ := json.Marshal(r)
	url, text := desktopPreview(raw)
	if url != "data:image/png;base64,"+data || !strings.Contains(text, "degraded") || !strings.Contains(text, "truncated") {
		t.Fatalf("%s: %s", url, text)
	}
	r.Error = &desktop.Failure{Code: "permissions_pending", Message: "permissions missing"}
	raw, _ = json.Marshal(r)
	url, text = desktopPreview(raw)
	if url != "" || text != "error: permissions missing" {
		t.Fatalf("%s: %s", url, text)
	}
}
