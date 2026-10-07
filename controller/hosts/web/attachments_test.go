package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// A pasted image is kept in scratch under its own hash; it is what its bytes
// say, never what its name claims, and its name cannot place it.
func TestPastedImagesAreKeptByHash(t *testing.T) {
	scratch := t.TempDir()
	data := base64.StdEncoding.EncodeToString(pngBytes(t))
	got, err := saveAttachments(scratch, []pasted{{Name: "../../etc/Screenshot.png", Data: data}})
	if err != nil || len(got) != 1 {
		t.Fatalf("%+v %v", got, err)
	}
	a := got[0]
	if a.Type != "image/png" || a.Name != "Screenshot.png" || a.Path != filepath.Join(scratch, "attachments", a.SHA256+".png") {
		t.Fatalf("attachment: %+v", a)
	}
	if b, _ := os.ReadFile(a.Path); !bytes.Equal(b, pngBytes(t)) {
		t.Fatal("the bytes kept are not the bytes pasted")
	}
	again, _ := saveAttachments(scratch, []pasted{{Name: "other.png", Data: data}})
	if again[0].Path != a.Path {
		t.Fatal("the same image was kept twice")
	}
	for why, images := range map[string][]pasted{
		"text named .png": {{Name: "x.png", Data: base64.StdEncoding.EncodeToString([]byte("not an image"))}},
		"not base64":      {{Name: "x.png", Data: "%%%"}},
		"too many":        {{Data: data}, {Data: data}, {Data: data}, {Data: data}, {Data: data}},
		"too large":       {{Name: "big.png", Data: base64.StdEncoding.EncodeToString(append(pngBytes(t), make([]byte, maxImageBytes)...))}},
	} {
		if _, err := saveAttachments(scratch, images); err == nil {
			t.Fatalf("%s was accepted", why)
		}
	}
	if !strings.Contains(withImages("look", got), "🖼 Screenshot.png") {
		t.Fatal("the transcript does not name the image")
	}
}
