package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
	"strings"
	"testing"
	"time"
)

func buildCodexFixture(t *testing.T) codexConfig {
	t.Helper()
	dir := t.TempDir()
	c := codexConfig{Binary: filepath.Join(dir, "codex"), Echo: filepath.Join(dir, "echo"), Provider: filepath.Join(dir, "provider"), Descriptor: "../codex/profiles.json"}
	for _, b := range []struct{ path, pkg string }{{c.Binary, "../codex"}, {c.Echo, "./echo"}, {c.Provider, "../codex/testdata/fake-provider"}} {
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
		trace := &Trace{}
		r, err := bootCodex(ctx, c, trace)
		must(t, err)
		defer r.close()
		profiles, err := r.execution.Capabilities(p.Codex)
		must(t, err)
		if len(profiles) != 8 {
			t.Fatal("profiles not exposed", profiles)
		}
		profiles[0].Fields[0].Key = "tampered"
		again, _ := r.execution.Capabilities(p.Codex)
		if again[0].Fields[0].Key == "tampered" {
			t.Fatal("mutable capability leak")
		}
		parent, err := r.execution.mintParent(p.Codex)
		must(t, err)
		second, err := r.execution.mintParent(p.Codex)
		must(t, err)
		fork, err := r.execution.mintParent(p.Codex)
		must(t, err)
		id := uint64(1)
		call := func(args []p.Observation) p.Move { id++; m, err := r.call(ctx, id, args); must(t, err); return m }
		expectText(t, call(p.Args("health", "connection", r.connection, "deadline_ms", "1000")), "inference readiness untested")
		expectText(t, call(p.Args("start-thread", "connection", r.connection, "parent", parent, "text", "one", "tools", `["echo"]`)), "thread-1:one")
		expectText(t, call(p.Args("submit-input", "parent", parent, "text", "two")), "thread-1:two")
		expectText(t, call(p.Args("start-thread", "connection", r.connection, "parent", second, "text", "other", "tools", `[]`)), "thread-2:other")
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
		must(t, r.kernel.scheduler.waitIdle(ctx, "root"))
		r.execution.mu.Lock()
		for _, v := range r.execution.bookkeeper.parents {
			if v.live {
				t.Error("parent not retired")
			}
		}
		r.execution.mu.Unlock()
		r.close()
		assertCodexClean(t, r)
		if !strings.Contains(diagnostics.String(), "provider reaped") {
			t.Fatal("provider not reaped", diagnostics.String())
		}
	})
	t.Run("interleaved independent parents and echo progress", func(t *testing.T) {
		ctx := guard(t)
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		p1, _ := r.execution.mintParent(p.Codex)
		p2, _ := r.execution.mintParent(p.Codex)
		send := func(id uint64, args []p.Observation) {
			must(t, r.client.egress.Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}))
		}
		send(2, p.Args("start-thread", "parent", p1, "connection", r.connection, "text", "hold", "tools", `[]`))
		// An echo result is a deterministic barrier after the first start is committed.
		must(t, r.client.egress.Send(ctx, p.Move{Identity: 3, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", "independent")}}))
		echo, err := r.results.ingress.Receive(ctx)
		must(t, err)
		expectText(t, echo, "independent")
		send(4, p.Args("start-thread", "parent", p2, "connection", r.connection, "text", "release", "tools", `[]`))
		got := map[uint64]string{}
		for i := 0; i < 2; i++ {
			m, err := r.results.ingress.Receive(ctx)
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
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		replacement, peer := pair("root", 21, p.Echo, false)
		defer peer.Close()
		replacement.binding.generation = 2
		must(t, r.execution.rebind(1, replacement))
		select {
		case r.kernel.attachments <- replacement.ingress:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		parent, _ := r.execution.mintParent(p.Codex)
		must(t, r.client.egress.Send(ctx, p.Move{Identity: 2, Target: p.Codex, Invoke: &p.Invoke{Args: p.Args("start-thread", "parent", parent, "connection", r.connection, "text", "taint-child", "tools", `["echo"]`)}}))
		child, err := peer.ingress.Receive(ctx)
		must(t, err)
		if child.Invoke.ReturnTo == 0 || child.Invoke.Frame == child.Invoke.ReturnTo {
			t.Fatal("missing separate child identity")
		}
		fault, err := r.results.ingress.Receive(ctx)
		must(t, err)
		expectFault(t, fault, "contamination")
		must(t, peer.egress.Send(ctx, returned(child, "late clean child")))
		health, err := r.call(ctx, 3, p.Args("health", "connection", r.connection, "deadline_ms", "1000"))
		must(t, err)
		expectText(t, health, "responsive")
		if health.Identity != 3 {
			t.Fatal("late output became second result", health)
		}
		r.execution.mu.Lock()
		if r.execution.bookkeeper.frames[child.Invoke.Frame].phase != "faulted" {
			t.Error("descendant remained live")
		}
		r.execution.mu.Unlock()
	})
	t.Run("failed replacement retains original parent and thread", func(t *testing.T) {
		ctx := guard(t)
		c := c
		c.ProviderArgs = []string{"reject-replace"}
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		parent, _ := r.execution.mintParent(p.Codex)
		m, err := r.call(ctx, 2, p.Args("start-thread", "parent", parent, "connection", r.connection, "text", "before", "tools", `["echo"]`))
		must(t, err)
		expectText(t, m, "thread-1:before")
		m, err = r.call(ctx, 3, p.Args("tools-remove", "parent", parent, "tools", `["echo"]`))
		must(t, err)
		expectFault(t, m, "unsupported")
		m, err = r.call(ctx, 4, p.Args("submit-input", "parent", parent, "text", "tool:retained"))
		must(t, err)
		expectText(t, m, "tool:retained")
	})

	t.Run("cancel outstanding provider work releases all owners", func(t *testing.T) {
		ctx := guard(t)
		var diagnostics bytes.Buffer
		c := c
		c.Stderr = &diagnostics
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		parent, _ := r.execution.mintParent(p.Codex)
		must(t, r.client.egress.Send(ctx, p.Move{Identity: 2, Target: p.Codex, Invoke: &p.Invoke{Args: p.Args("start-thread", "parent", parent, "connection", r.connection, "text", "hold", "tools", `[]`)}}))
		// The health result is a correlated barrier, while the first turn stays open.
		health, err := r.call(ctx, 3, p.Args("health", "connection", r.connection, "deadline_ms", "1000"))
		must(t, err)
		expectText(t, health, "responsive")
		r.close()
		assertCodexClean(t, r)
		for _, c := range r.execution.bookkeeper.parents {
			if c.live || c.busy != 0 {
				t.Fatal("parent retained after cancellation")
			}
		}
		for _, f := range r.execution.bookkeeper.frames {
			if f.phase == "running" || f.phase == "waiting" {
				t.Fatal("frame retained after cancellation")
			}
		}
		if !strings.Contains(diagnostics.String(), "provider reaped") {
			t.Fatal("provider not reaped")
		}
	})
	t.Run("healthy port reports provider exit", func(t *testing.T) {
		ctx := guard(t)
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		parent, _ := r.execution.mintParent(p.Codex)
		m, err := r.call(ctx, 2, p.Args("start-thread", "parent", parent, "connection", r.connection, "text", "provider-exit", "tools", `[]`))
		must(t, err)
		expectFault(t, m, "provider:")
	})
	t.Run("health deadline does not stop echo", func(t *testing.T) {
		ctx := guard(t)
		c := c
		c.ProviderArgs = []string{"hang-health"}
		r, err := bootCodex(ctx, c, &Trace{})
		must(t, err)
		defer r.close()
		m, err := r.call(ctx, 2, p.Args("health", "connection", r.connection, "deadline_ms", "1"))
		must(t, err)
		expectFault(t, m, "timeout")
		must(t, r.client.egress.Send(ctx, p.Move{Identity: 3, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", "alive")}}))
		m, err = r.results.ingress.Receive(ctx)
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
			r, err := bootCodex(ctx, c, &Trace{})
			if err == nil {
				r.close()
				t.Fatal("startup unexpectedly succeeded")
			}
			assertCodexClean(t, r)
			if (mode == "bad-connect" || mode == "hang-connect") && !strings.Contains(stderr.String(), "provider reaped") {
				t.Fatal("provider not reaped", stderr.String())
			}
		})
	}
}
func assertCodexClean(t *testing.T, r *codexBoot) {
	t.Helper()
	if len(r.kernel.runtimeState.active.reservations) != 0 || len(r.kernel.runtimeState.active.contexts) != 0 || len(r.kernel.scheduler.realizing) != 0 {
		t.Fatal("live resources after close")
	}
	if _, ok := r.kernel.runtimeState.memory.executables[p.Codex]; ok {
		t.Fatal("registration leaked")
	}
	for _, h := range []*helper{r.actor, r.echo} {
		if h != nil {
			if h.endpoint.binding.live.Load() {
				t.Fatal("live endpoint")
			}
			if h.command.Process != nil && (!h.waited || h.command.ProcessState == nil) {
				t.Fatal("unreaped process")
			}
		}
	}
}
func TestWaitingFaultInvalidatesDescendants(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	parent := f.start(t, invoke(1, "parent"))
	request := invoke(2, "child")
	request.Source = parent.Invoke.Frame
	request.State = []byte("saved")
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, request))
	done := runAsync(f.ctx, f.kernel)
	child, err := f.peer.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	fault := p.Move{Identity: 3, Context: "A", Source: parent.Invoke.Frame, Fault: &p.Fault{Frame: parent.Invoke.Frame, Message: "contamination"}}
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, fault))
	done = runAsync(f.ctx, f.kernel)
	result, err := f.result.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	expectFault(t, result, "contamination")
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, returned(child, "late")))
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, returned(parent, "clean")))
}
func TestProfileAuthorityBeforeDispatch(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	raw, err := os.ReadFile("../codex/profiles.json")
	must(t, err)
	d, err := p.LoadDescriptor(raw)
	must(t, err)
	// Build on the fixture's authenticated ingress and reserved endpoint.
	f.execution.mu.Lock()
	for _, m := range f.execution.horizon {
		if m.executable.ID == p.Echo {
			m.executable = definition(d, map[string]bool{"health": true})
			m.endpoint.binding.target = p.Codex
		}
	}
	f.execution.mu.Unlock()
	for _, args := range [][]p.Observation{p.Args("unknown"), p.Args("connect", "configuration", "default", "descriptor", "x"), p.Args("tool-result", "parent", "{}", "pending", "fake", "text", "bad"), p.Args("health", "connection", "x", "deadline_ms", "1", "deadline_ms", "2"), p.Args("health", "connection", "x", "deadline_ms", "0")} {
		m := p.Move{Identity: 10, Context: "A", Target: p.Codex, Invoke: &p.Invoke{Args: args}}
		f.execution.mu.Lock()
		_, _, _, err := f.execution.commit(queuedMove{move: m, source: f.external.binding})
		f.execution.mu.Unlock()
		rejected(t, err)
	}
	visible, err := f.execution.Capabilities(p.Codex)
	must(t, err)
	if len(visible) != 1 || visible[0].Operation != "health" {
		t.Fatal(visible)
	}
}

func TestHostParentOwnershipAndBusyFence(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	raw, err := os.ReadFile("../codex/profiles.json")
	must(t, err)
	d, err := p.LoadDescriptor(raw)
	must(t, err)
	f.execution.mu.Lock()
	f.execution.bookkeeper.key = []byte("private host key")
	f.execution.bookkeeper.namespace = "host"
	var target *mapping
	for _, m := range f.execution.horizon {
		if m.executable.ID == p.Echo {
			m.executable = definition(d, nil)
			m.endpoint.binding.target = p.Codex
			target = m
		}
	}
	f.execution.mu.Unlock()
	parent, err := f.execution.mintParent(p.Codex)
	must(t, err)
	args := p.Args("start-thread", "parent", parent, "connection", "connection", "text", "hello", "tools", `[]`)
	commit := func(id uint64, args []p.Observation) (p.Move, error) {
		f.execution.mu.Lock()
		defer f.execution.mu.Unlock()
		m, _, _, err := f.execution.commit(queuedMove{move: p.Move{Identity: id, Context: "A", Target: p.Codex, Invoke: &p.Invoke{Args: args}}, source: f.external.binding})
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
	f.execution.mu.Lock()
	if len(f.execution.bookkeeper.frames) != 1 || f.execution.bookkeeper.parents[out.Invoke.Frame] != nil {
		t.Error("parent and command identity conflated")
	}
	target.endpoint.binding.generation++
	f.execution.mu.Unlock()
	// A replacement binding must not inherit a parent minted for an older incarnation.
	replacement, peer := pair("A", 10, p.Codex, false)
	defer peer.Close()
	replacement.binding.generation = 3
	must(t, f.execution.rebind(0, replacement))
	_, err = commit(4, args)
	rejected(t, err)
}
