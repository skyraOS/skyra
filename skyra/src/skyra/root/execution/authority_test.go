package execution

import (
	"net"
	"os"
	"strings"
	"testing"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func rejected(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected boundary rejection")
	}
	t.Log("rejected:", err)
}

// codexContext publishes context A with Codex, narrowed to grants, at virtual 0
// (actual 10), an external caller at 1 and a root sink at 2. It publishes
// without Active: these tests exercise commit, never admission or dispatch.
func codexContext(t *testing.T, grants map[string]bool) (*ExecutionContext, *ports.Endpoint) {
	t.Helper()
	raw, err := os.ReadFile("../../codex/profiles.json")
	must(t, err)
	d, err := p.LoadDescriptor(raw)
	must(t, err)
	var endpoints []*ports.Endpoint
	bind := func(actual uint64, target p.ExecutableID, external bool) *ports.Endpoint {
		a, b := net.Pipe()
		ep, peer := ports.Bind(a, "A", actual, target, 1, external), ports.Bind(b, "A", 0, 0, 1, false)
		endpoints = append(endpoints, ep, peer)
		return ep
	}
	codex, external, sink := bind(10, p.Codex, false), bind(11, 0, true), bind(12, 0, false)
	t.Cleanup(func() {
		for _, ep := range endpoints {
			_ = ep.Close()
		}
	})
	e := New("A", &trace.Trace{}, sink.Egress())
	must(t, e.Install(0, codex, Definition(d, grants)))
	must(t, e.Install(1, external, Executable{}))
	must(t, e.Install(2, sink, Executable{}))
	must(t, e.Publish(map[uint64]bool{10: true, 11: true, 12: true}, func() {}))
	return e, external
}

func TestProfileAuthorityBeforeDispatch(t *testing.T) {
	e, external := codexContext(t, map[string]bool{"health": true})
	for _, args := range [][]p.Observation{p.Args("unknown"), p.Args("connect", "configuration", "default", "descriptor", "x"), p.Args("tool-result", "parent", "{}", "pending", "fake", "text", "bad"), p.Args("health", "connection", "x", "deadline_ms", "1", "deadline_ms", "2"), p.Args("health", "connection", "x", "deadline_ms", "0")} {
		m := p.Move{Identity: 10, Context: "A", Target: p.Codex, Invoke: &p.Invoke{Args: args}}
		e.mu.Lock()
		_, _, _, err := e.commit(Work{Move: m, Source: external.Binding()})
		e.mu.Unlock()
		rejected(t, err)
	}
	visible, err := e.Capabilities(p.Codex)
	must(t, err)
	if len(visible) != 1 || visible[0].Operation != "health" {
		t.Fatal(visible)
	}
}

func TestHostParentOwnershipAndBusyFence(t *testing.T) {
	e, external := codexContext(t, nil)
	e.ProvisionParentAuthority([]byte("private host key"), "host")
	parent, err := e.MintParent(p.Codex)
	must(t, err)
	args := p.Args("start-thread", "parent", parent, "connection", "connection", "text", "hello", "tools", `[]`)
	commit := func(id uint64, args []p.Observation) (p.Move, error) {
		e.mu.Lock()
		defer e.mu.Unlock()
		m, _, _, err := e.commit(Work{Move: p.Move{Identity: id, Context: "A", Target: p.Codex, Invoke: &p.Invoke{Args: args}}, Source: external.Binding()})
		return m, err
	}
	for _, bad := range []string{strings.Replace(parent, `"owner":"A"`, `"owner":"B"`, 1), strings.Replace(parent, `"namespace":"host"`, `"namespace":"foreign"`, 1), strings.Replace(parent, `"generation":1`, `"generation":2`, 1)} {
		_, err := commit(1, p.Args("start-thread", "parent", bad, "connection", "connection", "text", "hello", "tools", `[]`))
		rejected(t, err)
	}
	out, err := commit(2, args)
	must(t, err)
	_, err = commit(3, p.Args("tools-add", "parent", parent, "tools", `["echo"]`))
	rejected(t, err)
	e.mu.Lock()
	if len(e.bookkeeper.frames) != 1 || e.bookkeeper.parents[out.Invoke.Frame] != nil {
		t.Error("parent and command identity conflated")
	}
	e.mu.Unlock()
	// A replacement binding must not inherit a parent minted for an older incarnation.
	a, b := net.Pipe()
	replacement, peer := ports.Bind(a, "A", 10, p.Codex, 3, false), ports.Bind(b, "A", 0, 0, 1, false)
	defer peer.Close()
	must(t, e.Rebind(0, replacement))
	_, err = commit(4, args)
	rejected(t, err)
}
