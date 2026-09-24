package boot_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skyra-v1/src/skyra/root/boot"
	"skyra-v1/src/skyra/root/trace"
)

// End-to-end proof that a browser message drives a real Codex thread through
// the isolation fence, with no change to Codex or the runtime:
//
//	browser (WS) -> user -> root -> codex -> provider
//	                          \-> echo (codex's tool child)
//
// The host connects Codex and mints U's parent; the user relays each human
// message into the Codex conversation (start-thread, then submit-input) and
// resumes U with the model result. Uses the deterministic fake provider.
func TestUserCodexBrowserLoop(t *testing.T) {
	dir := t.TempDir()
	c := boot.UserCodexConfig{
		UserBinary:  filepath.Join(dir, "user"),
		CodexBinary: filepath.Join(dir, "codex"),
		Echo:        filepath.Join(dir, "echo"),
		Provider:    filepath.Join(dir, "provider"),
		Descriptor:  "../../codex/profiles.json",
	}
	for _, b := range []struct{ path, pkg string }{
		{c.UserBinary, "../../user"},
		{c.CodexBinary, "../../codex"},
		{c.Echo, "../echo"},
		{c.Provider, "../../codex/testdata/fake-provider"},
	} {
		cmd := exec.CommandContext(guard(t), "go", "build", "-race", "-o", b.path, b.pkg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", b.pkg, err, out)
		}
	}

	var diag bytes.Buffer
	c.Stderr = &diag
	ctx := guard(t)
	r, err := boot.UserCodex(ctx, c, &trace.Trace{})
	must(t, err)
	defer r.Close()

	ws := dialWS(t, r.Addr)
	defer ws.close()

	// Wait for U to open.
	var incarnation, conversation string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && incarnation == "" {
		m := ws.read(t)
		if m["type"] != "snapshot" {
			continue
		}
		if convs, _ := m["conversations"].([]any); len(convs) > 0 {
			incarnation = m["uiIncarnation"].(string)
			conversation = convs[0].(map[string]any)["conversation"].(string)
		}
	}
	if incarnation == "" {
		t.Fatalf("U never opened; stderr:\n%s", diag.String())
	}

	submit := func(id, text string) {
		ws.send(t, map[string]any{
			"type": "submit", "uiIncarnation": incarnation,
			"conversation": conversation, "requestId": id, "text": text,
		})
	}

	// First message starts the Codex thread; the fake provider labels replies
	// "thread-1:<text>". Second message submits input on the same thread.
	submit("r1", "hello")
	if !ws.waitAgent(t, "thread-1:hello") {
		t.Fatalf("start-thread result never routed back; stderr:\n%s", diag.String())
	}
	submit("r2", "again")
	if !ws.waitAgent(t, "thread-1:again") {
		t.Fatalf("submit-input result never routed back; stderr:\n%s", diag.String())
	}

	// Closing settles U's single terminal Return.
	ws.send(t, map[string]any{
		"type": "close-interaction", "uiIncarnation": incarnation,
		"conversation": conversation, "requestId": "close",
	})
	term, err := r.AwaitClose(ctx)
	must(t, err)
	if term.Identity != r.InteractID || term.Return == nil || string(term.Return.Observation.Bytes) != "interaction-closed" {
		t.Fatalf("bad terminal U return: %+v", term)
	}
}

// waitAgent reads snapshots until some agent entry contains want.
func (c *wsClient) waitAgent(t *testing.T, want string) bool {
	t.Helper()
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		m := c.read(t)
		if m["type"] != "snapshot" {
			continue
		}
		convs, _ := m["conversations"].([]any)
		if len(convs) == 0 {
			continue
		}
		entries, _ := convs[0].(map[string]any)["entries"].([]any)
		for _, e := range entries {
			em := e.(map[string]any)
			if em["role"] == "agent" && strings.Contains(em["text"].(string), want) {
				return true
			}
		}
	}
	return false
}
