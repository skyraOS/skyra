// Package observation defines Skyra's first Cua tool: read one approved window.
// Desktop input and arbitrary driver dispatch are deliberately not represented.
package observation

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	MaxRequest   = 4096
	MaxResponse  = 4 << 20
	MaxImage     = 2 << 20
	MaxDimension = 1280
	MaxText      = 64 << 10
)

type Request struct {
	PID               int    `json:"pid"`
	WindowID          int    `json:"window_id"`
	Query             string `json:"query,omitempty"`
	MaxElements       int    `json:"max_elements,omitempty"`
	MaxDepth          int    `json:"max_depth,omitempty"`
	IncludeScreenshot *bool  `json:"include_screenshot,omitempty"`
}

func (r Request) Validate() error {
	if r.PID < 0 || r.WindowID < 0 || (r.PID == 0) != (r.WindowID == 0) {
		return errors.New("pid and window_id must both be positive integers, or both omitted")
	}
	if len(r.Query) > 512 || !utf8.ValidString(r.Query) {
		return errors.New("query must be UTF-8 and at most 512 bytes")
	}
	if r.MaxElements < 0 || r.MaxElements > 250 {
		return errors.New("max_elements must be between 1 and 250, or omitted")
	}
	if r.MaxDepth < 0 || r.MaxDepth > 25 {
		return errors.New("max_depth must be between 1 and 25, or omitted")
	}
	return nil
}

func DecodeRequest(b []byte) (Request, error) {
	var r Request
	if len(b) > MaxRequest {
		return r, errors.New("request exceeds 4 KiB")
	}
	if !utf8.Valid(b) || len(bytes.TrimSpace(b)) == 0 || bytes.TrimSpace(b)[0] != '{' {
		return r, errors.New("request must be a UTF-8 JSON object")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&r); err != nil {
		return r, err
	}
	if d.Decode(new(any)) != io.EOF {
		return r, errors.New("trailing request data")
	}
	return r, r.Validate()
}

// Scope is fixed by the operator. ManifestHash is Cua's domain-separated hash,
// not the plain hash of the manifest file. It identifies the exact bytes loaded.
type Scope struct {
	DriverVersion string `json:"driver_version"`
	ManifestHash  string `json:"manifest_sha256"`
	BundleID      string `json:"bundle_id"`
	PID           int    `json:"pid"`
	WindowID      int    `json:"window_id"`
}

func (s Scope) Validate() error {
	if s.DriverVersion == "" || len(s.DriverVersion) > 128 || s.BundleID == "" || len(s.BundleID) > 256 {
		return errors.New("driver version and app bundle ID are required")
	}
	b, err := hex.DecodeString(s.ManifestHash)
	if err != nil || len(b) != sha256.Size {
		return errors.New("manifest hash must be SHA-256")
	}
	if s.PID <= 0 || s.WindowID <= 0 {
		return errors.New("scope requires a positive pid and window_id")
	}
	return nil
}

// Manifest creates the only driver manifest this first cut accepts. Return the
// exact bytes so bootstrap can retain them; no YAML parser or freeform tool list.
func Manifest(bundleID string) ([]byte, error) {
	if bundleID == "" || len(bundleID) > 256 || strings.ContainsAny(bundleID, " \t\r\n\x00") {
		return nil, errors.New("a macOS app bundle ID is required")
	}
	return json.MarshalIndent(map[string]any{
		"version": 3, "expires_after": "8h", "idle_timeout": "30m",
		"allow": map[string]any{"tools": []string{"get_window_state"}}, "resources": map[string]any{
			"apps":    []any{map[string]any{"bundle_id": bundleID, "windows": "all", "launch": false}},
			"desktop": map[string]any{"display": false},
		},
	}, "", "  ")
}

func ManifestHash(b []byte) string {
	h := sha256.New()
	h.Write([]byte("cua-driver-capability-manifest-v3\x00"))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

type Failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Screenshot struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
	SHA256   string `json:"sha256"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// Validate binds a bounded decoded image to its recorded identity and dimensions.
func (s Screenshot) Validate() error {
	if len(s.Data) > base64.StdEncoding.EncodedLen(MaxImage) {
		return errors.New("screenshot exceeds 2 MiB")
	}
	b, err := base64.StdEncoding.Strict().DecodeString(s.Data)
	if err != nil || len(b) == 0 || len(b) > MaxImage {
		return errors.New("invalid screenshot encoding")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return err
	}
	mime := map[string]string{"png": "image/png", "jpeg": "image/jpeg"}[format]
	h := sha256.Sum256(b)
	if mime == "" || mime != s.MIMEType || s.SHA256 != hex.EncodeToString(h[:]) || cfg.Width != s.Width || cfg.Height != s.Height || s.Width <= 0 || s.Height <= 0 || s.Width > MaxDimension || s.Height > MaxDimension {
		return errors.New("screenshot format, dimensions or hash differ")
	}
	if _, _, err := image.Decode(bytes.NewReader(b)); err != nil {
		return err
	}
	return nil
}

func ScreenshotFrom(data, mime string) (*Screenshot, error) {
	if len(data) > base64.StdEncoding.EncodedLen(MaxImage) {
		return nil, errors.New("screenshot exceeds 2 MiB")
	}
	b, err := base64.StdEncoding.Strict().DecodeString(data)
	if err != nil {
		return nil, err
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(b)
	s := &Screenshot{MIMEType: mime, Data: data, SHA256: hex.EncodeToString(h[:]), Width: cfg.Width, Height: cfg.Height}
	return s, s.Validate()
}

type Result struct {
	Screenshot *Screenshot `json:"screenshot,omitempty"`
	Version    int         `json:"version"`
	Operation  string      `json:"operation"`
	Scope      Scope       `json:"scope"`
	SnapshotID string      `json:"snapshot_id,omitempty"`
	Text       string      `json:"text"`
	Truncated  bool        `json:"truncated"`
	Degraded   bool        `json:"degraded"`
	Error      *Failure    `json:"error,omitempty"`
}

func NewResult(s Scope) Result { return Result{Version: 1, Operation: "desktop_observe", Scope: s} }

func (r Result) Failed(code string, err error) Result {
	r.Screenshot = nil
	message, _ := bounded(err.Error(), 2048)
	r.Error = &Failure{Code: code, Message: message}
	return r
}

// Project retains bounded text and exact image bytes for recording and replay.
// Unknown structured fields and output paths are never forwarded.
func Project(s Scope, b []byte) Result {
	r := NewResult(s)
	if len(b) > MaxResponse || !utf8.Valid(b) {
		return r.Failed("invalid_response", errors.New("driver response exceeds the limit or is not UTF-8"))
	}
	var out struct {
		IsError bool `json:"isError"`
		Content []struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			Data     string `json:"data"`
			MIMEType string `json:"mimeType"`
		} `json:"content"`
		Structured struct {
			SnapshotID   string   `json:"snapshot_id"`
			TreeMarkdown string   `json:"tree_markdown"`
			Truncated    bool     `json:"truncated"`
			Degraded     bool     `json:"degraded"`
			Code         string   `json:"code"`
			Message      string   `json:"message"`
			Refusal      *Failure `json:"refusal"`
		} `json:"structuredContent"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return r.Failed("invalid_response", err)
	}
	if len(out.Structured.SnapshotID) > 512 {
		return r.Failed("invalid_response", errors.New("snapshot identity exceeds 512 bytes"))
	}
	if len(out.Content) == 0 && out.Structured.TreeMarkdown == "" && !out.IsError {
		return r.Failed("invalid_response", errors.New("driver omitted observation content"))
	}
	var texts []string
	for _, c := range out.Content {
		if c.Type == "image" {
			if r.Screenshot != nil {
				return r.Failed("invalid_response", errors.New("multiple screenshots"))
			}
			img, err := ScreenshotFrom(c.Data, c.MIMEType)
			if err != nil {
				return r.Failed("invalid_response", err)
			}
			r.Screenshot = img
			continue
		}
		if c.Type != "text" {
			return r.Failed("unexpected_content", errors.New("unsupported observation content"))
		}
		texts = append(texts, c.Text)
	}
	text := strings.Join(texts, "\n")
	if text == "" {
		text = out.Structured.TreeMarkdown
	}
	if !utf8.ValidString(text) {
		return r.Failed("invalid_response", errors.New("observation is not UTF-8"))
	}
	r.Text, r.Truncated = bounded(text, MaxText)
	r.Truncated = r.Truncated || out.Structured.Truncated
	r.Degraded, r.SnapshotID = out.Structured.Degraded, out.Structured.SnapshotID
	if out.IsError {
		code, message := out.Structured.Code, out.Structured.Message
		if out.Structured.Refusal != nil {
			code, message = out.Structured.Refusal.Code, out.Structured.Refusal.Message
		}
		if code == "" {
			code = "driver_refused"
		}
		if message == "" {
			message = text
		}
		if message == "" {
			message = "driver refused the observation"
		}
		return r.Failed(code, errors.New(message))
	}
	return r
}

func bounded(s string, n int) (string, bool) {
	if len(s) <= n {
		return s, false
	}
	s = s[:n]
	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s, true
}
