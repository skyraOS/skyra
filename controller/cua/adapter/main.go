// Command desktop-observe is a one-shot sandboxed Skyra leaf.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"skyra/controller/cua/observation"
)

func main() {
	if os.Getenv("SKYRA_FRAME") == "" {
		fmt.Fprintln(os.Stderr, "desktop-observe: no frame")
		os.Exit(1)
	}
	b, err := io.ReadAll(io.LimitReader(os.Stdin, observation.MaxRequest+1))
	var r observation.Request
	if err == nil {
		r, err = observation.DecodeRequest(b)
	}
	out := observation.NewResult(observation.Scope{})
	if err != nil {
		out = out.Failed("invalid_request", err)
	} else if socket := bridgeSocket(); !filepath.IsAbs(socket) {
		out = out.Failed("unavailable", fmt.Errorf("no absolute bridge socket grant"))
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, err = observation.CallBridge(ctx, socket, r)
		if err != nil {
			out = observation.NewResult(observation.Scope{}).Failed("transport_error", err)
		}
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": out})
}

// SKYRA_CONNECT remains the legacy single-socket view. The complete grant
// list avoids selecting the browser/ideas socket in a shared Boundary.
func bridgeSocket() string {
	raw := os.Getenv("SKYRA_CONNECTIONS")
	if raw == "" {
		return os.Getenv("SKYRA_CONNECT")
	}
	var sockets []string
	if json.Unmarshal([]byte(raw), &sockets) != nil {
		return ""
	}
	selected := ""
	for _, s := range sockets {
		if filepath.IsAbs(s) && filepath.Base(s) == "bridge.sock" && strings.HasPrefix(filepath.Base(filepath.Dir(s)), "skyra-cua-") {
			if selected != "" {
				return ""
			}
			selected = s
		}
	}
	return selected
}
