// Command blender-tools invokes only typed operations through a private operator-owned bridge.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type request struct {
	Operation            string          `json:"operation"`
	ExpectedSceneVersion string          `json:"expected_scene_version,omitempty"`
	ObjectID             string          `json:"object_id,omitempty"`
	Translation          []float64       `json:"translation,omitempty"`
	Rotation             []float64       `json:"rotation,omitempty"`
	Scale                []float64       `json:"scale,omitempty"`
	MaterialID           string          `json:"material_id,omitempty"`
	Parameter            string          `json:"parameter,omitempty"`
	Value                json.RawMessage `json:"value,omitempty"`
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
	return map[string]any{"version": 1, "operation": "blender.tools", "error": f}
}

func dispatchedFailure(r request, code, message string) error {
	f := &failure{Code: code, Message: message, Stage: "bridge"}
	if r.Operation == "set_transform" || r.Operation == "set_material_parameter" {
		f.EditOutcome = "unknown"
	}
	return f
}

func run(r request) (any, error) {
	switch r.Operation {
	case "context", "inspect_scene", "inspect_object", "set_transform", "set_material_parameter", "export_asset", "render_preview":
	default:
		return nil, failed("invalid_request", "Unknown typed Blender operation.", "request")
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
		if filepath.IsAbs(s) && filepath.Base(s) == "blender.sock" && strings.HasPrefix(filepath.Base(filepath.Dir(s)), "skyra-blender-") {
			if selected != "" {
				return nil, failed("ambiguous_bridge_grant", "The conversation has multiple Blender bridge grants.", "bridge")
			}
			selected = s
		}
	}
	if selected == "" {
		return nil, failed("no_bridge_grant", "This conversation has no Blender bridge grant. Restart the backend and start a new conversation with the Blender bridge grant.", "bridge")
	}
	conn, err := net.DialTimeout("unix", selected, 5*time.Second)
	if err != nil {
		return nil, failed("bridge_unavailable", "This conversation cannot reach its granted Blender bridge. The backend may have restarted; start a new conversation with the current grant.", "bridge")
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(100 * time.Second))
	if err = json.NewEncoder(conn).Encode(r); err != nil {
		return nil, dispatchedFailure(r, "bridge_send_failed", "The local Blender bridge failed while receiving this request.")
	}
	data, err := io.ReadAll(io.LimitReader(conn, 4<<20+1))
	if err != nil {
		code, message := "bridge_read_failed", "The local Blender bridge closed before a complete response arrived."
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			code, message = "bridge_timeout", "Timed out waiting for the local Blender bridge response."
		}
		f := &failure{Code: code, Message: message, Stage: "bridge"}
		if r.Operation == "set_transform" || r.Operation == "set_material_parameter" {
			f.EditOutcome = "unknown"
		}
		return nil, f
	}
	if len(data) > 4<<20 {
		return nil, dispatchedFailure(r, "response_too_large", "The local Blender bridge response exceeds 4 MiB.")
	}
	var out map[string]any
	if json.Unmarshal(data, &out) != nil || out["version"] != float64(1) || out["operation"] != "blender.tools" {
		return nil, dispatchedFailure(r, "invalid_bridge_response", "The local Blender bridge returned a malformed result.")
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
