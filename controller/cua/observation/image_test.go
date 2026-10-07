package observation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func imageData(t *testing.T, width int) string {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, width, 12))
	im.Set(1, 1, color.RGBA{R: 128, B: 220, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func TestScreenshotRetainsExactBytesAndRejectsTampering(t *testing.T) {
	data := imageData(t, 16)
	raw, _ := json.Marshal(map[string]any{"content": []any{map[string]string{"type": "image", "mimeType": "image/png", "data": data}}})
	r := Project(Scope{}, raw)
	if r.Error != nil || r.Screenshot == nil || r.Screenshot.Data != data || r.Screenshot.Width != 16 {
		t.Fatalf("%+v", r)
	}
	saved, _ := json.Marshal(r)
	var replay Result
	json.Unmarshal(saved, &replay)
	if replay.Screenshot.Data != data || replay.Screenshot.Validate() != nil {
		t.Fatal("recorded image changed")
	}
	replay.Screenshot.SHA256 = "wrong"
	if replay.Screenshot.Validate() == nil {
		t.Fatal("accepted wrong hash")
	}
	for _, tc := range []struct{ data, mime string }{{"bad", "image/png"}, {data, "image/jpeg"}, {imageData(t, 1281), "image/png"}} {
		if _, err := ScreenshotFrom(tc.data, tc.mime); err == nil {
			t.Fatal("accepted invalid screenshot")
		}
	}
	duplicate, _ := json.Marshal(map[string]any{"content": []any{map[string]string{"type": "image", "mimeType": "image/png", "data": data}, map[string]string{"type": "image", "mimeType": "image/png", "data": data}}})
	if r := Project(Scope{}, duplicate); r.Error == nil || r.Screenshot != nil {
		t.Fatal("accepted duplicate images")
	}
}
