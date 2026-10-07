// Command claude-start is a resident that runs one claude -p turn with an MCP
// bridge for recorded tool delivery. It reads a turn configuration from stdin,
// starts the bridge and CLI, forwards stream events on stdout, and routes tool
// calls through the bridge. When claude exits, the process exits and the holder
// reports EOF; the adapter starts a fresh incarnation for each turn.
//
// Its stdio is the holder's socketpair. It has no filesystem or process
// authority beyond launching the claude CLI.
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"skyra/controller/claude/bridge"
	"skyra/controller/claude/provider"
)

type turnConfig struct {
	Prompt       string        `json:"prompt"`
	Session      string        `json:"session"`
	Frame        string        `json:"frame"`
	SystemPrompt string        `json:"system_prompt"`
	Tools        []bridge.Tool `json:"tools"`
}

func main() {
	stdin := bufio.NewScanner(os.Stdin)
	stdin.Buffer(make([]byte, 4096), 4<<20)
	if !stdin.Scan() {
		fatal("no turn configuration on stdin")
	}
	var cfg turnConfig
	if json.Unmarshal(stdin.Bytes(), &cfg) != nil || cfg.Prompt == "" {
		fatal("invalid turn configuration")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		fatal("random: " + err.Error())
	}
	token := hex.EncodeToString(tokenBytes)

	var mu sync.Mutex
	pending := map[string]chan callResult{}
	var outMu sync.Mutex

	write := func(v any) {
		outMu.Lock()
		defer outMu.Unlock()
		b, _ := json.Marshal(v)
		os.Stdout.Write(append(b, '\n'))
	}

	call := func(ctx context.Context, id, name string, args json.RawMessage) (json.RawMessage, error) {
		ch := make(chan callResult, 1)
		mu.Lock()
		pending[id] = ch
		mu.Unlock()
		write(map[string]any{"_bridge": "call", "id": id, "name": name, "args": args})
		select {
		case <-ctx.Done():
			mu.Lock()
			delete(pending, id)
			mu.Unlock()
			return nil, ctx.Err()
		case r := <-ch:
			return r.value, r.err
		}
	}

	frame := cfg.Frame
	if frame == "" {
		frame = "turn"
	}
	srv, err := bridge.New(frame, token, cfg.Tools, bridge.Call(call))
	if err != nil {
		fatal("bridge: " + err.Error())
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fatal("listen: " + err.Error())
	}
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/mcp", listener.Addr().(*net.TCPAddr).Port)
	httpSrv := &http.Server{Handler: srv}
	go httpSrv.Serve(listener)
	defer func() { httpSrv.Close(); listener.Close() }()

	toolNames := make([]string, len(cfg.Tools))
	for i, t := range cfg.Tools {
		toolNames[i] = t.Name
	}
	args, err := provider.BridgeArguments(cfg.Session, cfg.Prompt, cfg.SystemPrompt, endpoint, token, toolNames)
	if err != nil {
		fatal("arguments: " + err.Error())
	}

	env, err := provider.Environment(os.Environ())
	if err != nil {
		fatal("environment: " + err.Error())
	}

	cmd := exec.Command("claude", args...)
	cmd.Env = env
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fatal("pipe: " + err.Error())
	}
	if err := cmd.Start(); err != nil {
		fatal("start: " + err.Error())
	}

	// Route bridge results from stdin to pending calls. This goroutine exits
	// when the process exits; it does not need a clean shutdown because each
	// resident runs exactly one turn.
	go func() {
		for stdin.Scan() {
			var msg struct {
				Bridge string          `json:"_bridge"`
				ID     string          `json:"id"`
				Value  json.RawMessage `json:"value"`
				Error  string          `json:"error"`
			}
			if json.Unmarshal(stdin.Bytes(), &msg) != nil || msg.Bridge == "" {
				continue
			}
			mu.Lock()
			ch := pending[msg.ID]
			mu.Unlock()
			if ch == nil {
				continue
			}
			if msg.Bridge == "result" {
				ch <- callResult{value: msg.Value}
			} else if msg.Bridge == "error" {
				ch <- callResult{err: errors.New(msg.Error)}
			}
		}
	}()

	// Forward claude's stdout to our stdout (the holder).
	out := bufio.NewScanner(stdout)
	out.Buffer(make([]byte, 4096), 8<<20)
	for out.Scan() {
		outMu.Lock()
		os.Stdout.Write(append(out.Bytes(), '\n'))
		outMu.Unlock()
	}
	cmd.Wait()
}

type callResult struct {
	value json.RawMessage
	err   error
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "claude-start:", msg)
	os.Exit(1)
}
