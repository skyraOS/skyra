package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// pasted is an image the human attached to a message: a screenshot from the
// clipboard or a copied image file, as the client read it (base64 bytes).
type pasted struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// attachment is a pasted image as the runtime sees it: saved in the
// conversation's scratch under its own hash, so the lineage records exactly
// which image was sent while the bytes stay beside it. Codex reads it from
// path; the app-server's profile lets it read images there and nothing else.
type attachment struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

const (
	maxImages     = 4
	maxImageBytes = 10 << 20 // all of a message's images together, decoded
)

var imageTypes = map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/gif": "gif", "image/webp": "webp"}

// saveAttachments checks a message's images and keeps them in scratch. An
// image is what its bytes say, not what its name claims.
func saveAttachments(scratch string, images []pasted) ([]attachment, error) {
	if len(images) == 0 {
		return nil, nil
	}
	if len(images) > maxImages {
		return nil, fmt.Errorf("at most %d images in a message", maxImages)
	}
	var decoded [][]byte
	total := 0
	for _, im := range images {
		b, err := base64.StdEncoding.DecodeString(im.Data)
		if err != nil {
			return nil, errors.New("an image is not base64")
		}
		if total += len(b); total > maxImageBytes {
			return nil, fmt.Errorf("a message's images may be %d MB together", maxImageBytes>>20)
		}
		decoded = append(decoded, b)
	}
	dir := filepath.Join(scratch, "attachments")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	var out []attachment
	for i, b := range decoded {
		kind := http.DetectContentType(b)
		ext, ok := imageTypes[kind]
		if !ok {
			return nil, fmt.Errorf("%q is not a PNG, JPEG, GIF or WebP image", name(images[i].Name))
		}
		sum := sha256.Sum256(b)
		hash := hex.EncodeToString(sum[:])
		path := filepath.Join(dir, hash+"."+ext)
		if _, err := os.Stat(path); err != nil {
			tmp, err := os.CreateTemp(dir, ".tmp-")
			if err != nil {
				return nil, err
			}
			_, werr := tmp.Write(b)
			cerr := tmp.Close()
			if werr != nil || cerr != nil {
				os.Remove(tmp.Name())
				return nil, errors.Join(werr, cerr)
			}
			if err := os.Rename(tmp.Name(), path); err != nil {
				os.Remove(tmp.Name())
				return nil, err
			}
		}
		out = append(out, attachment{Path: path, SHA256: hash, Name: name(images[i].Name), Type: kind})
	}
	return out, nil
}

// name is an image's name for showing: its base name, short, or "image".
func name(n string) string {
	n = filepath.Base(strings.TrimSpace(n))
	if n == "" || n == "." || n == "/" {
		return "image"
	}
	if r := []rune(n); len(r) > 80 {
		return string(r[:80])
	}
	return n
}
