// Command design-tools provides a bounded draft writer and the private Penpot bridge.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

type request struct {
	Operation string `json:"operation"`
	Path      string `json:"path,omitempty"`
	Text      string `json:"text,omitempty"`
	Code      string `json:"code,omitempty"`
	Type      string `json:"type,omitempty"`
	Member    string `json:"member,omitempty"`
	ShapeID   string `json:"shape_id,omitempty"`
}

func draft(scratch string, r request) (map[string]any, error) {
	if scratch == "" || !filepath.IsLocal(r.Path) || r.Path == "." {
		return nil, fmt.Errorf("path must be a relative file inside ideas/design")
	}
	if !utf8.ValidString(r.Text) || strings.ContainsRune(r.Text, 0) || len(r.Text) > 1<<20 {
		return nil, fmt.Errorf("draft must be UTF-8 text, at most 1 MiB")
	}
	parent, err := os.OpenRoot(scratch)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	// Refuse even an internal symlink: design is a real, dedicated directory.
	info, err := parent.Lstat("design")
	if os.IsNotExist(err) {
		err = parent.Mkdir("design", 0700)
		info = nil
	}
	if err != nil {
		return nil, err
	}
	if info != nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return nil, fmt.Errorf("design must be a real directory")
	}
	root, err := parent.OpenRoot("design")
	if err != nil {
		return nil, err
	}
	defer root.Close()
	if err = root.MkdirAll(filepath.Dir(r.Path), 0700); err != nil {
		return nil, err
	}
	// Root rejects symlink escapes; reject final symlinks to avoid overwriting aliases.
	if info, err := root.Lstat(r.Path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("symlink destination refused")
	}
	f, err := root.OpenFile(r.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	_, err = io.WriteString(f, r.Text)
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	hash := sha256.Sum256([]byte(r.Text))
	return map[string]any{"version": 1, "operation": "design.tools", "path": "design/" + r.Path, "read_as": "scratch:design/" + r.Path, "size": len(r.Text), "sha256": hex.EncodeToString(hash[:])}, nil
}

type failure struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Stage       string `json:"stage"`
	Operation   string `json:"operation,omitempty"`
	EditOutcome string `json:"edit_outcome,omitempty"`
}

func (f *failure) Error() string { return f.Message }
func failed(code, message, stage string) error {
	return &failure{Code: code, Message: message, Stage: stage}
}
func errorResult(err error, operation string) map[string]any {
	var f *failure
	if !errors.As(err, &f) {
		f = &failure{Code: "invalid_request", Message: err.Error(), Stage: "request"}
	}
	f.Operation = operation
	return map[string]any{"version": 1, "operation": "design.tools", "error": f}
}

func dispatchedFailure(r request, code, message string) error {
	f := &failure{Code: code, Message: message, Stage: "bridge"}
	if r.Operation == "execute" {
		f.EditOutcome = "unknown"
	}
	return f
}

func run(r request) (any, error) {
	switch r.Operation {
	case "write":
		out, err := draft(os.Getenv("SKYRA_SCRATCH"), r)
		if err != nil {
			return nil, failed("draft_write_failed", err.Error(), "write")
		}
		return out, nil
	case "context", "overview", "api_info", "execute", "export":
	default:
		return nil, failed("invalid_request", "Unknown design operation.", "request")
	}
	var sockets []string
	raw := os.Getenv("SKYRA_CONNECTIONS")
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &sockets); err != nil {
			return nil, failed("invalid_bridge_grant", "The runtime supplied a malformed connection grant.", "bridge")
		}
	}
	selected := ""
	for _, s := range sockets {
		if filepath.IsAbs(s) && filepath.Base(s) == "penpot.sock" && strings.HasPrefix(filepath.Base(filepath.Dir(s)), "skyra-penpot-") {
			if selected != "" {
				return nil, failed("ambiguous_bridge_grant", "The conversation has multiple Penpot bridge grants.", "bridge")
			}
			selected = s
		}
	}
	if selected == "" {
		return nil, failed("no_bridge_grant", "This conversation has no Penpot bridge grant. Configure Penpot in Skyra, then start a new conversation; this does not establish that the browser plugin is disconnected.", "bridge")
	}
	conn, err := net.DialTimeout("unix", selected, 5*time.Second)
	if err != nil {
		return nil, failed("bridge_unavailable", "This conversation cannot reach its granted Penpot bridge. The backend may have restarted; start a new conversation with the current grant.", "bridge")
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(65 * time.Second))
	if err = json.NewEncoder(conn).Encode(r); err != nil {
		return nil, dispatchedFailure(r, "bridge_send_failed", "The local Penpot bridge failed while receiving this request.")
	}
	data, err := io.ReadAll(io.LimitReader(conn, 4<<20+1))
	if err != nil {
		code, message := "bridge_read_failed", "The local Penpot bridge closed before a complete response arrived."
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			code, message = "bridge_timeout", "Timed out waiting for the local Penpot bridge response."
		}
		f := &failure{Code: code, Message: message, Stage: "bridge"}
		if r.Operation == "execute" {
			f.EditOutcome = "unknown"
		}
		return nil, f
	}
	if len(data) > 4<<20 {
		return nil, dispatchedFailure(r, "response_too_large", "The local Penpot bridge response exceeds 4 MiB.")
	}
	var out map[string]any
	if json.Unmarshal(data, &out) != nil || out["version"] != float64(1) || out["operation"] != "design.tools" {
		return nil, dispatchedFailure(r, "invalid_bridge_response", "The local Penpot bridge returned a malformed result.")
	}
	return out, nil
}

func main() {
	data, err := io.ReadAll(io.LimitReader(os.Stdin, (1<<20)+8193))
	var r request
	if err == nil {
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		err = d.Decode(&r)
		if err == nil {
			var extra any
			if d.Decode(&extra) != io.EOF {
				err = fmt.Errorf("trailing request data")
			}
		}
	}
	var out any
	if err == nil {
		out, err = run(r)
	}
	if err != nil {
		out = errorResult(err, r.Operation)
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": out})
}
