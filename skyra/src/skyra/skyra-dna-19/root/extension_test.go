package main

import (
	"context"
	"errors"
	"net"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
	"skyra-v1/src/skyra/skyra-dna-19/root/creation"
	"strconv"
	"strings"
	"testing"
)

func liveLoop(t *testing.T, ctx context.Context, k *Kernel, inputs []*Ingress) func() {
	t.Helper()
	shutdownGuard := guard(t)
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- k.loop(ctx, inputs) }()
	return func() {
		cancel()
		if err := await(t, shutdownGuard, done); !errors.Is(err, context.Canceled) {
			t.Errorf("loop failed before explicit cancellation: %v", err)
		}
		k.scheduler.mu.Lock()
		defer k.scheduler.mu.Unlock()
		if len(k.scheduler.realizing) != 0 {
			t.Error("claims leaked")
		}
		k.runtimeState.active.mu.RLock()
		defer k.runtimeState.active.mu.RUnlock()
		if len(k.runtimeState.active.reservations) != 0 || len(k.runtimeState.active.contexts) != 0 {
			t.Error("shutdown retained authority")
		}
	}
}

func traceCount(k *Kernel, owner p.ContextID, prefix string) int {
	n := 0
	for _, e := range k.trace.snapshot() {
		if e.Context == owner && strings.HasPrefix(e.Transition, prefix) {
			n++
		}
	}
	return n
}

func TestLiveLoopRejectionContainment(t *testing.T) {
	for _, stale := range []bool{false, true} {
		name := "duplicate"
		if stale {
			name = "binding invalidated after admission"
		}
		t.Run(name, func(t *testing.T) {
			a := newFixture(t, nil, "A", 10, false)
			b := newFixture(t, a.kernel, "B", 20, false)
			wire := a.start(t, invoke(1, "old"))
			terminal := returned(wire, "old")
			must(t, a.kernel.admit(a.ctx, a.echo.ingress, terminal))
			if stale {
				left, right := net.Pipe()
				defer right.Close()
				must(t, a.execution.rebind(0, bind(left, "A", 10, p.Echo, 2, false)))
			} else {
				must(t, a.kernel.admit(a.ctx, a.echo.ingress, terminal))
			}
			// Both decoded outputs are queued before the live loop is released. No
			// receive/admission race or sleep is needed to reach realization rejection.
			stop := liveLoop(t, a.ctx, a.kernel, []*Ingress{a.echo.ingress})
			defer stop()
			if !stale {
				_, err := a.result.ingress.Receive(a.ctx)
				must(t, err)
			}
			must(t, a.kernel.scheduler.waitIdle(a.ctx, "A"))
			if traceCount(a.kernel, "A", "dispatch rejected:") != 1 {
				t.Fatal("queued rejection not recorded")
			}
			expected := 1
			if stale {
				expected = 0
			}
			if traceCount(a.kernel, "A", "root result sent") != expected {
				t.Fatal("extra root result")
			}
			if stale {
				a.execution.mu.Lock()
				phase := a.execution.bookkeeper.frames[wire.Invoke.Frame].phase
				a.execution.mu.Unlock()
				if phase != "running" {
					t.Fatal("stale output settled old frame")
				}
			}
			b.enqueue(t, invoke(2, "still running"))
			other, err := b.peer.ingress.Receive(b.ctx)
			must(t, err)
			must(t, b.kernel.admit(b.ctx, b.echo.ingress, returned(other, "still running")))
			result, err := b.result.ingress.Receive(b.ctx)
			must(t, err)
			if result.Identity != 2 || string(result.Return.Observation.Bytes) != "still running" {
				t.Fatal("other context failed")
			}
			must(t, b.kernel.scheduler.waitIdle(b.ctx, "B"))
		})
	}
}

func TestCrossContextAndCreation(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	ctx, cancel := context.WithCancel(a.ctx)
	defer cancel()
	x, err := setupExtension(ctx, a.kernel, a.execution, grantPolicy{echo: true})
	must(t, err)
	defer x.close()
	stop := liveLoop(t, ctx, a.kernel, append([]*Ingress{a.external.ingress}, x.inputs...))
	defer stop()
	call := func(id uint64, target p.ExecutableID, op, text string) p.Move {
		t.Helper()
		must(t, a.client.egress.Send(ctx, p.Move{Identity: id, Target: target, Invoke: &p.Invoke{Args: p.Arguments(op, text)}}))
		result, err := a.result.ingress.Receive(ctx)
		must(t, err)
		if result.Context != "A" || result.Identity != id {
			t.Fatal("lost caller correlation")
		}
		must(t, a.kernel.scheduler.waitIdle(ctx, "A"))
		return result
	}
	// B cannot invoke A's capability, even when it knows the numeric address.
	bad := p.Move{Identity: 900, Target: creation.Capability, Invoke: &p.Invoke{Args: p.Arguments("create", "1")}}
	must(t, a.kernel.admit(ctx, x.inputs[1], bad))
	must(t, a.kernel.scheduler.waitIdle(ctx, "B"))
	if traceCount(a.kernel, "B", "dispatch rejected:") != 1 {
		t.Fatal("B authority violation accepted")
	}
	before := len(a.kernel.runtimeState.memory.executables)
	if x.minted {
		t.Fatal("unauthorized creation minted")
	}
	for _, id := range []uint64{21, 22} {
		result := call(id, x.child.Activation().ID, "echo", "round trip")
		if result.Return == nil || string(result.Return.Observation.Bytes) != "round trip" {
			t.Fatal("wrong cross-context result")
		}
	}
	if traceCount(a.kernel, "B", "root sink resolved") != 2 || traceCount(a.kernel, "A", "root result sent") != 2 {
		t.Fatal("result bypassed own sink")
	}
	if a.execution.horizon[0].actual == x.child.horizon[0].actual {
		t.Fatal("virtual port spaces alias")
	}
	for _, m := range x.child.horizon {
		if m.executable.ID != 0 && m.executable.ID != p.Echo {
			t.Fatal("B received excess delegation")
		}
	}
	// The invocation has capability access but requests a non-grantable template.
	denied := call(23, creation.Capability, "create", "2")
	if denied.Fault == nil || x.minted {
		t.Fatal("creation exceeded grant policy")
	}
	result := call(24, creation.Capability, "create", "1")
	if result.Return == nil || result.Return.Observation.Key != "executable" {
		t.Fatal("missing new address")
	}
	n, err := strconv.ParseUint(string(result.Return.Observation.Bytes), 10, 64)
	must(t, err)
	if n < 1000 || len(a.kernel.runtimeState.memory.executables) != before+1 {
		t.Fatal("address was not freshly minted")
	}
	used := call(25, p.ExecutableID(n), "echo", "minted")
	if used.Return == nil || string(used.Return.Observation.Bytes) != "minted" {
		t.Fatal("new executable unusable")
	}
	exhausted := call(26, creation.Capability, "create", "1")
	if exhausted.Fault == nil {
		t.Fatal("one-address creation budget exceeded")
	}
	a.execution.mu.Lock()
	minted, err := a.execution.resolve(p.ExecutableID(n))
	must(t, err)
	if len(minted.executable.Contracts) != 1 || minted.executable.Contracts[0].Operation != "echo" {
		t.Fatal("created authority widened")
	}
	a.execution.mu.Unlock()
}

func TestRootSinkAndOutputContract(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	if f.execution.bookkeeper.bottom.id != f.execution.bookkeeper.rootSink || f.execution.rootEgress == nil {
		t.Fatal("sink destination not wired")
	}
	wire := f.start(t, invoke(1, "output"))
	bad := returned(wire, "output")
	bad.Return.Observation.Key = "executable"
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, bad))
	rejected(t, f.kernel.scheduler.Run(f.ctx, f.kernel.dispatch))
	if f.execution.bookkeeper.frames[wire.Invoke.Frame].phase != "running" {
		t.Fatal("invalid output settled frame")
	}
	_ = f.finish(t, returned(wire, "output"))
	if traceCount(f.kernel, "A", "root sink resolved") != 1 {
		t.Fatal("sink resolution missing")
	}
}

func TestLiveLoopFatalFailures(t *testing.T) {
	t.Run("committed send failure", func(t *testing.T) {
		f := newFixture(t, nil, "A", 10, false)
		f.enqueue(t, invoke(1, "fail after commit"))
		must(t, f.peer.Close())
		done := make(chan error, 1)
		go func() { done <- f.kernel.loop(f.ctx, nil) }()
		err := await(t, f.ctx, done)
		if err == nil {
			t.Fatal("transport loss silently ignored")
		}
		if traceCount(f.kernel, "A", "transition committed") != 1 || traceCount(f.kernel, "A", "frame recorded") != 1 {
			t.Fatal("send retried or never committed")
		}
		if len(f.kernel.scheduler.realizing) != 0 {
			t.Fatal("fatal failure leaked claim")
		}
	})
	t.Run("missing context", func(t *testing.T) {
		k := newKernel(&Trace{})
		ctx := guard(t)
		m := invoke(1, "missing")
		m.Context = "missing"
		must(t, k.scheduler.Enqueue(queuedMove{move: m}))
		done := make(chan error, 1)
		go func() { done <- k.loop(ctx, nil) }()
		rejected(t, await(t, ctx, done))
	})
}

func TestLiveLoopCancellationClosesClaimedTurn(t *testing.T) {
	f := newFixture(t, nil, "A", 10, true)
	ctx, cancel := context.WithCancel(f.ctx)
	f.gate.ctx = ctx
	f.enqueue(t, invoke(1, "held send"))
	done := make(chan error, 1)
	go func() { done <- f.kernel.loop(ctx, []*Ingress{f.external.ingress}) }()
	select {
	case <-f.gate.entered:
	case <-f.ctx.Done():
		t.Fatal("send never claimed")
	}
	cancel()
	if !errors.Is(await(t, f.ctx, done), context.Canceled) {
		t.Fatal("cancellation lost")
	}
	if len(f.kernel.scheduler.realizing) != 0 || f.external.binding.live.Load() || f.echo.binding.live.Load() {
		t.Fatal("shutdown leaked claim or endpoint")
	}
	if _, ok := f.kernel.runtimeState.active.ownerOf(10); ok {
		t.Fatal("shutdown leaked reservation")
	}
}

func TestDelegationDeniedWithoutGrant(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	_, err := setupExtension(f.ctx, f.kernel, f.execution, grantPolicy{})
	rejected(t, err)
	if _, ok := f.kernel.runtimeState.active.lookup("B"); ok {
		t.Fatal("ungrantable child published")
	}
	if len(f.execution.horizon) != 3 {
		t.Fatal("denied delegation changed creator horizon")
	}
}
