package boot_test

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/boot"
	kt "skyra/history/dna19/root/kernel/kerneltest"
	"skyra/history/dna19/root/ports"
	"skyra/history/dna19/root/trace"
)

var (
	guard    = kt.Guard
	must     = kt.Must
	rejected = kt.Rejected
	returned = kt.Returned
)

func buildCodexFixture(t *testing.T) boot.CodexConfig {
	t.Helper()
	dir := t.TempDir()
	c := boot.CodexConfig{Binary: filepath.Join(dir, "codex"), Echo: filepath.Join(dir, "echo"), Provider: filepath.Join(dir, "provider"), Descriptor: "../../codex/profiles.json"}
	for _, b := range []struct{ path, pkg string }{{c.Binary, "../../codex"}, {c.Echo, "../echo"}, {c.Provider, "../../codex/testdata/fake-provider"}} {
		cmd := exec.CommandContext(guard(t), "go", "build", "-race", "-o", b.path, b.pkg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build: %v %s", err, out)
		}
	}
	return c
}
func expectText(t *testing.T, m p.Move, want string) {
	t.Helper()
	if m.Return == nil || !strings.Contains(string(m.Return.Observation.Bytes), want) {
		t.Fatalf("want %q: %+v", want, m)
	}
}
func expectFault(t *testing.T, m p.Move, want string) {
	t.Helper()
	if m.Fault == nil || !strings.Contains(m.Fault.Message, want) {
		t.Fatalf("want fault %q: %+v", want, m)
	}
}
func TestCodexProcessAcceptance(t *testing.T) {
	c := buildCodexFixture(t)
	t.Run("registration persistent threads fork tools continuation contamination", func(t *testing.T) {
		var diagnostics bytes.Buffer
		c := c
		c.Stderr = &diagnostics
		ctx := guard(t)
		trace := &trace.Trace{}
		r, err := boot.Codex(ctx, c, trace)
		must(t, err)
		defer r.Close()
		profiles, err := r.Execution.Capabilities(p.Codex)
		must(t, err)
		if len(profiles) != 8 {
			t.Fatal("profiles not exposed", profiles)
		}
		profiles[0].Fields[0].Key = "tampered"
		again, _ := r.Execution.Capabilities(p.Codex)
		if again[0].Fields[0].Key == "tampered" {
			t.Fatal("mutable capability leak")
		}
		parent, err := r.Execution.MintParent(p.Codex)
		must(t, err)
		second, err := r.Execution.MintParent(p.Codex)
		must(t, err)
		fork, err := r.Execution.MintParent(p.Codex)
		must(t, err)
		id := uint64(1)
		call := func(args []p.Observation) p.Move { id++; m, err := r.Call(ctx, id, args); must(t, err); return m }
		expectText(t, call(p.Args("health", "connection", r.Connection, "deadline_ms", "1000")), "inference readiness untested")
		expectText(t, call(p.Args("start-thread", "connection", r.Connection, "parent", parent, "text", "one", "tools", `["echo"]`)), "thread-1:one")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "two")), "thread-1:two")
		expectText(t, call(p.Args("start-thread", "connection", r.Connection, "parent", second, "text", "other", "tools", `[]`)), "thread-2:other")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "tool:value")), "tool:value")
		expectText(t, call(p.Args("fork-thread", "parent", fork, "source", parent)), "fork created")
		expectText(t, call(p.Args("submit-input", "parent", fork, "text", "branch")), "thread-3:branch")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "history")), "one|two|tool:value|history")
		expectText(t, call(p.Args("tools-remove", "parent", parent, "tools", `["echo"]`)), "tools replaced")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "after")), "thread-4:after")
		expectText(t, call(p.Args("tools-add", "parent", parent, "tools", `["echo"]`)), "tools replaced")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "tool:restored")), "tool:restored")
		expectFault(t, call(p.Args("submit-input", "parent", second, "text", "tool:denied")), "contamination")
		expectFault(t, call(p.Args("submit-input", "parent", parent, "text", "taint")), "contamination")
		expectText(t, call(p.Args("end-parent", "parent", fork)), "parent retired")
		must(t, r.Kernel.Scheduler().WaitIdle(ctx, "root"))
		for _, v := range r.Execution.Parents() {
			if v.Live {
				t.Error("parent not retired")
			}
		}
		r.Close()
		assertCodexClean(t, r)
		if !strings.Contains(diagnostics.String(), "provider reaped") {
			t.Fatal("provider not reaped", diagnostics.String())
		}
	})
	t.Run("interleaved independent parents and echo progress", func(t *testing.T) {
		ctx := guard(t)
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		p1, _ := r.Execution.MintParent(p.Codex)
		p2, _ := r.Execution.MintParent(p.Codex)
		send := func(id uint64, args []p.Observation) {
			must(t, r.Client.Egress().Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}))
		}
		send(2, p.Args("start-thread", "parent", p1, "connection", r.Connection, "text", "hold", "tools", `[]`))
		// An echo result is a deterministic barrier after the first start is committed.
		must(t, r.Client.Egress().Send(ctx, p.Move{Identity: 3, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", "independent")}}))
		echo, err := r.Results.Ingress().Receive(ctx)
		must(t, err)
		expectText(t, echo, "independent")
		send(4, p.Args("start-thread", "parent", p2, "connection", r.Connection, "text", "release", "tools", `[]`))
		got := map[uint64]string{}
		for i := 0; i < 2; i++ {
			m, err := r.Results.Ingress().Receive(ctx)
			must(t, err)
			if m.Return == nil {
				t.Fatal(m)
			}
			got[m.Identity] = string(m.Return.Observation.Bytes)
		}
		if got[2] != "held:thread-1" || got[4] != "isolated:thread-2" {
			t.Fatal("cross-thread routing", got)
		}
	})

	t.Run("contamination faults while real binary waits on child", func(t *testing.T) {
		ctx := guard(t)
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		left, right := net.Pipe()
		replacement, peer := ports.Bind(left, "root", 21, p.Echo, 2, false), ports.Bind(right, "root", 0, 0, 1, false)
		defer peer.Close()
		must(t, r.Execution.Rebind(1, replacement))
		must(t, r.Kernel.Attach(ctx, replacement.Ingress()))
		parent, _ := r.Execution.MintParent(p.Codex)
		must(t, r.Client.Egress().Send(ctx, p.Move{Identity: 2, Target: p.Codex, Invoke: &p.Invoke{Args: p.Args("start-thread", "parent", parent, "connection", r.Connection, "text", "taint-child", "tools", `["echo"]`)}}))
		child, err := peer.Ingress().Receive(ctx)
		must(t, err)
		if child.Invoke.ReturnTo == 0 || child.Invoke.Frame == child.Invoke.ReturnTo {
			t.Fatal("missing separate child identity")
		}
		fault, err := r.Results.Ingress().Receive(ctx)
		must(t, err)
		expectFault(t, fault, "contamination")
		must(t, peer.Egress().Send(ctx, returned(child, "late clean child")))
		health, err := r.Call(ctx, 3, p.Args("health", "connection", r.Connection, "deadline_ms", "1000"))
		must(t, err)
		expectText(t, health, "responsive")
		if health.Identity != 3 {
			t.Fatal("late output became second result", health)
		}
		if r.Execution.Frames()[child.Invoke.Frame].Phase != "faulted" {
			t.Error("descendant remained live")
		}
	})
	t.Run("failed replacement retains original parent and thread", func(t *testing.T) {
		ctx := guard(t)
		c := c
		c.ProviderArgs = []string{"reject-replace"}
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		parent, _ := r.Execution.MintParent(p.Codex)
		m, err := r.Call(ctx, 2, p.Args("start-thread", "parent", parent, "connection", r.Connection, "text", "before", "tools", `["echo"]`))
		must(t, err)
		expectText(t, m, "thread-1:before")
		m, err = r.Call(ctx, 3, p.Args("tools-remove", "parent", parent, "tools", `["echo"]`))
		must(t, err)
		expectFault(t, m, "unsupported")
		m, err = r.Call(ctx, 4, p.Args("submit-input", "parent", parent, "text", "tool:retained"))
		must(t, err)
		expectText(t, m, "tool:retained")
	})

	t.Run("cancel outstanding provider work releases all owners", func(t *testing.T) {
		ctx := guard(t)
		var diagnostics bytes.Buffer
		c := c
		c.Stderr = &diagnostics
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		parent, _ := r.Execution.MintParent(p.Codex)
		must(t, r.Client.Egress().Send(ctx, p.Move{Identity: 2, Target: p.Codex, Invoke: &p.Invoke{Args: p.Args("start-thread", "parent", parent, "connection", r.Connection, "text", "hold", "tools", `[]`)}}))
		// The health result is a correlated barrier, while the first turn stays open.
		health, err := r.Call(ctx, 3, p.Args("health", "connection", r.Connection, "deadline_ms", "1000"))
		must(t, err)
		expectText(t, health, "responsive")
		r.Close()
		assertCodexClean(t, r)
		for _, c := range r.Execution.Parents() {
			if c.Live || c.Busy != 0 {
				t.Fatal("parent retained after cancellation")
			}
		}
		for _, f := range r.Execution.Frames() {
			if f.Phase == "running" || f.Phase == "waiting" {
				t.Fatal("frame retained after cancellation")
			}
		}
		if !strings.Contains(diagnostics.String(), "provider reaped") {
			t.Fatal("provider not reaped")
		}
	})
	t.Run("healthy port reports provider exit", func(t *testing.T) {
		ctx := guard(t)
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		parent, _ := r.Execution.MintParent(p.Codex)
		m, err := r.Call(ctx, 2, p.Args("start-thread", "parent", parent, "connection", r.Connection, "text", "provider-exit", "tools", `[]`))
		must(t, err)
		expectFault(t, m, "provider:")
	})
	t.Run("health deadline does not stop echo", func(t *testing.T) {
		ctx := guard(t)
		c := c
		c.ProviderArgs = []string{"hang-health"}
		r, err := boot.Codex(ctx, c, &trace.Trace{})
		must(t, err)
		defer r.Close()
		m, err := r.Call(ctx, 2, p.Args("health", "connection", r.Connection, "deadline_ms", "1"))
		must(t, err)
		expectFault(t, m, "timeout")
		must(t, r.Client.Egress().Send(ctx, p.Move{Identity: 3, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", "alive")}}))
		m, err = r.Results.Ingress().Receive(ctx)
		must(t, err)
		expectText(t, m, "alive")
	})
	for _, mode := range []string{"bad-connect", "hang-connect", "missing-binary", "descriptor-mismatch", "bad-version"} {
		t.Run("rollback "+mode, func(t *testing.T) {
			c := c
			var stderr bytes.Buffer
			c.Stderr = &stderr
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			switch mode {
			case "missing-binary":
				c.Binary = filepath.Join(t.TempDir(), "missing")
			case "descriptor-mismatch", "bad-version":
				raw, err := os.ReadFile(c.Descriptor)
				must(t, err)
				if mode == "bad-version" {
					raw = bytes.Replace(raw, []byte(`"protocol": 1`), []byte(`"protocol": 9`), 1)
				} else {
					raw = append(raw, ' ')
				}
				c.Descriptor = filepath.Join(t.TempDir(), "profiles.json")
				must(t, os.WriteFile(c.Descriptor, raw, 0600))
			default:
				c.ProviderArgs = []string{mode}
			}
			r, err := boot.Codex(ctx, c, &trace.Trace{})
			if err == nil {
				r.Close()
				t.Fatal("startup unexpectedly succeeded")
			}
			assertCodexClean(t, r)
			if (mode == "bad-connect" || mode == "hang-connect") && !strings.Contains(stderr.String(), "provider reaped") {
				t.Fatal("provider not reaped", stderr.String())
			}
		})
	}
}
func assertCodexClean(t *testing.T, r *boot.CodexBoot) {
	t.Helper()
	if contexts, reservations := r.Kernel.State().Active.Size(); reservations != 0 || contexts != 0 || r.Kernel.Scheduler().Claims() != 0 {
		t.Fatal("live resources after close")
	}
	if r.Kernel.State().Memory.Registered(p.Codex) {
		t.Fatal("registration leaked")
	}
	for _, h := range []*ports.Helper{r.Actor, r.Echo} {
		if h != nil {
			if h.Endpoint().Binding().Live() {
				t.Fatal("live endpoint")
			}
			if h.Unreaped() {
				t.Fatal("unreaped process")
			}
		}
	}
}
