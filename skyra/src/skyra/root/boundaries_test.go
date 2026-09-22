package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	p "skyra-v1/src/skyra/protocol"
)

func guard(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
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
func await(t *testing.T, ctx context.Context, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		t.Fatal("completion guard:", ctx.Err())
		return ctx.Err()
	}
}
func (g *gateConnection) Write(data []byte) (int, error) {
	g.once.Do(func() { close(g.entered) })
	select {
	case <-g.release:
		return g.ReadWriteCloser.Write(data)
	case <-g.ctx.Done():
		return 0, g.ctx.Err()
	}
}

func (r *readSignalConnection) Read(data []byte) (int, error) {
	r.once.Do(func() { close(r.entered) })
	return r.ReadWriteCloser.Read(data)
}
func (c *waitSignalContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.entered) })
	return c.Context.Done()
}

func newFixture(t *testing.T, k *Kernel, id p.ContextID, base uint64, gated bool) *fixture {
	t.Helper()
	ctx := guard(t)
	if k == nil {
		k = newKernel(&Trace{})
	}
	f := &fixture{ctx: ctx, kernel: k}
	in, remote := net.Pipe()
	sink, result := net.Pipe()
	echo, peer := net.Pipe()
	f.external = bind(in, id, base+1, 0, 1, true)
	f.client = bind(remote, id, 0, 0, 1, false)
	f.result = bind(result, id, 0, 0, 1, false)
	f.peer = bind(peer, id, 0, 0, 1, false)
	if gated {
		f.gate = &gateConnection{ReadWriteCloser: echo, entered: make(chan struct{}), release: make(chan struct{}), ctx: ctx}
		f.echo = bind(f.gate, id, base, p.Echo, 1, false)
	} else {
		f.echo = bind(echo, id, base, p.Echo, 1, false)
	}
	rootSink := bind(sink, id, base+2, 0, 1, false)
	f.execution = newExecution(id, k.trace, rootSink.egress)
	must(t, f.execution.install(0, f.echo, Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}))
	must(t, f.execution.install(1, f.external, Executable{}))
	must(t, f.execution.install(2, rootSink, Executable{}))
	k.runtimeState.memory.retain(f.execution)
	must(t, k.runtimeState.active.provision(ctx, f.execution, []uint64{base, base + 1, base + 2}))
	t.Cleanup(func() { k.runtimeState.active.retire(id); _ = f.peer.Close(); _ = f.result.Close(); _ = remote.Close() })
	return f
}
func invoke(id uint64, text string) p.Move {
	return p.Move{Identity: id, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", text)}}
}
func (f *fixture) enqueue(t *testing.T, m p.Move) {
	t.Helper()
	wire := exchange(t, f.ctx, f.client.egress, f.external.ingress, m)
	must(t, f.kernel.admit(f.ctx, f.external.ingress, wire))
}

func exchange(t *testing.T, ctx context.Context, out *Egress, in *Ingress, m p.Move) p.Move {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- out.Send(ctx, m) }()
	wire, err := in.Receive(ctx)
	must(t, err)
	must(t, await(t, ctx, done))
	return wire
}
func runAsync(ctx context.Context, k *Kernel) <-chan error {
	done := make(chan error, 1)
	go func() { done <- k.scheduler.Run(ctx, k.dispatch) }()
	return done
}
func (f *fixture) start(t *testing.T, m p.Move) p.Move {
	t.Helper()
	f.enqueue(t, m)
	done := runAsync(f.ctx, f.kernel)
	wire, err := f.peer.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	return wire
}
func returned(wire p.Move, text string) p.Move {
	return p.Move{Identity: wire.Identity, Context: wire.Context, Source: wire.Invoke.Frame, Return: &p.Return{Answering: wire.Invoke.Frame, Observation: p.Observation{Key: "text", Bytes: []byte(text)}}}
}
func (f *fixture) finish(t *testing.T, m p.Move) p.Move {
	t.Helper()
	wire := exchange(t, f.ctx, f.peer.egress, f.echo.ingress, m)
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, wire))
	done := runAsync(f.ctx, f.kernel)
	result, err := f.result.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	return result
}
func counts(f *fixture) (int, int) {
	f.execution.mu.Lock()
	frames := len(f.execution.bookkeeper.frames)
	f.execution.mu.Unlock()
	f.kernel.scheduler.mu.Lock()
	queued := len(f.kernel.scheduler.queue)
	f.kernel.scheduler.mu.Unlock()
	return frames, queued
}

func TestReservationConflictAtomic(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	x, y := net.Pipe()
	defer x.Close()
	defer y.Close()
	c := newExecution("C", a.kernel.trace, nil)
	must(t, c.install(0, bind(x, "C", 99, p.Echo, 1, false), Executable{ID: p.Echo}))
	rejected(t, a.kernel.runtimeState.active.provision(a.ctx, c, []uint64{99, 10}))
	if _, ok := a.kernel.runtimeState.active.lookup("C"); ok {
		t.Fatal("partial context publication")
	}
	if _, ok := a.kernel.runtimeState.active.ownerOf(99); ok {
		t.Fatal("partial reservation")
	}
	if id, ok := a.kernel.runtimeState.active.ownerOf(10); !ok || id != "A" {
		t.Fatal("existing reservation changed")
	}
}

func TestForgedOwnershipAndVirtualPortIsolation(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	b := newFixture(t, a.kernel, "B", 20, false)
	wa := a.start(t, invoke(1, "A text"))
	wb := b.start(t, invoke(1, "B text"))
	bad := returned(wa, "forged")
	bad.Context = "B"
	bad = exchange(t, a.ctx, a.peer.egress, a.echo.ingress, bad)
	rejected(t, a.kernel.admit(a.ctx, a.echo.ingress, bad))
	if frames, queued := counts(b); frames != 1 || queued != 0 {
		t.Fatal("forgery changed B")
	}
	ra := a.finish(t, returned(wa, "A text"))
	rb := b.finish(t, returned(wb, "B text"))
	if ra.Context != "A" || rb.Context != "B" || string(ra.Return.Observation.Bytes) != "A text" || string(rb.Return.Observation.Bytes) != "B text" {
		t.Fatal("cross-context result")
	}
	if a.execution.horizon[0].actual == b.execution.horizon[0].actual {
		t.Fatal("virtual 0 aliases actual connection")
	}
	t.Log("virtual port 0 delivered distinct A and B results")
}

func TestFixedDestinationAndContractRejection(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	bad := invoke(1, "wrong")
	bad.Target = p.Parent
	rejected(t, f.echo.egress.Send(f.ctx, bad))
	bad = invoke(1, "wrong")
	bad.Invoke.Args = p.Arguments("unknown", "wrong")
	f.enqueue(t, bad)
	rejected(t, f.kernel.scheduler.Run(f.ctx, f.kernel.dispatch))
	if frames, queued := counts(f); frames != 0 || queued != 0 {
		t.Fatal("malformed call changed frames")
	}
	// With the peer unread, any erroneous helper send would have blocked the Run.
	wire := f.start(t, invoke(2, "accepted"))
	if wire.Identity != 2 {
		t.Fatal("helper received rejected call")
	}
}

func TestSameContextSerializationAndDifferentContextProgress(t *testing.T) {
	a := newFixture(t, nil, "A", 10, true)
	b := newFixture(t, a.kernel, "B", 20, false)
	a.enqueue(t, invoke(1, "first"))
	first := runAsync(a.ctx, a.kernel)
	select {
	case <-a.gate.entered:
	case <-a.ctx.Done():
		t.Fatal("first Realize never entered Send")
	}
	a.enqueue(t, invoke(2, "second"))
	b.enqueue(t, invoke(3, "B advances"))
	other := runAsync(b.ctx, b.kernel)
	bwire, err := b.peer.ingress.Receive(b.ctx)
	must(t, err)
	must(t, await(t, b.ctx, other))
	if bwire.Identity != 3 {
		t.Fatal("B did not progress")
	}
	a.execution.mu.Lock()
	n := len(a.execution.bookkeeper.frames)
	a.execution.mu.Unlock()
	if n != 1 {
		t.Fatal("overlapping A state turn")
	}
	close(a.gate.release)
	awire, err := a.peer.ingress.Receive(a.ctx)
	must(t, err)
	must(t, await(t, a.ctx, first))
	next := runAsync(a.ctx, a.kernel)
	second, err := a.peer.ingress.Receive(a.ctx)
	must(t, err)
	must(t, await(t, a.ctx, next))
	if awire.Identity != 1 || second.Identity != 2 {
		t.Fatal("admission order not retained")
	}
	t.Log("B advanced while A Realize was held; A turns retained order")
}

func TestDispatchFailureReleasesClaim(t *testing.T) {
	k := newKernel(&Trace{})
	ctx := guard(t)
	for i := uint64(1); i <= 2; i++ {
		m := invoke(i, "")
		m.Context = "missing"
		must(t, k.scheduler.Enqueue(queuedMove{move: m}))
	}
	rejected(t, k.scheduler.Run(ctx, k.dispatch))
	called := false
	rejected(t, k.scheduler.Run(ctx, func(context.Context, queuedMove) error { called = true; return errors.New("dispatch fixture failure") }))
	if !called || len(k.scheduler.realizing) != 0 {
		t.Fatal("claim leaked")
	}
	if _, ok := k.runtimeState.active.lookup("missing"); ok {
		t.Fatal("implicit provision")
	}
	f := newFixture(t, k, "live", 10, false)
	_ = f.start(t, invoke(3, "progress"))
}

func TestStaleOutputAndDuplicateCompletion(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	wire := f.start(t, invoke(1, "value"))
	terminal := returned(wire, "value")
	// Two copies can enter before settlement; the second must fail at realization.
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, terminal))
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, terminal))
	done := runAsync(f.ctx, f.kernel)
	_, err := f.result.ingress.Receive(f.ctx)
	must(t, err)
	must(t, await(t, f.ctx, done))
	rejected(t, f.kernel.scheduler.Run(f.ctx, f.kernel.dispatch))
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, terminal))
	oldWire := f.start(t, invoke(2, "old"))
	oldOutput := returned(oldWire, "old")
	must(t, f.kernel.admit(f.ctx, f.echo.ingress, oldOutput))
	x, y := net.Pipe()
	defer y.Close()
	replacement := bind(x, "A", 10, p.Echo, 2, false)
	must(t, f.execution.rebind(0, replacement))
	rejected(t, f.kernel.admit(f.ctx, f.echo.ingress, oldOutput))
	rejected(t, f.kernel.scheduler.Run(f.ctx, f.kernel.dispatch))
	f.execution.mu.Lock()
	phase := f.execution.bookkeeper.frames[oldWire.Invoke.Frame].phase
	f.execution.mu.Unlock()
	if phase != "running" {
		t.Fatal("stale queued return settled frame")
	}
	n := 0
	for _, e := range f.kernel.trace.snapshot() {
		if e.Transition == "root result sent" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("got %d results", n)
	}
}

func TestCancellationAndShutdown(t *testing.T) {
	t.Run("receive and waiting scheduler", func(t *testing.T) {
		f := newFixture(t, nil, "A", 10, false)
		ctx, cancel := context.WithCancel(f.ctx)
		defer cancel()
		received := make(chan error, 1)
		x, y := net.Pipe()
		defer y.Close()
		reader := &readSignalConnection{ReadWriteCloser: x, entered: make(chan struct{})}
		endpoint := bind(reader, "A", 0, 0, 1, false)
		defer endpoint.Close()
		go func() { _, err := endpoint.ingress.Receive(ctx); received <- err }()
		select {
		case <-reader.entered:
		case <-f.ctx.Done():
			t.Fatal("receiver never blocked")
		}
		waitingContext := &waitSignalContext{Context: ctx, entered: make(chan struct{})}
		waiting := runAsync(waitingContext, f.kernel)
		select {
		case <-waitingContext.entered:
		case <-f.ctx.Done():
			t.Fatal("scheduler never waited")
		}
		cancel()
		if !errors.Is(await(t, f.ctx, received), context.Canceled) || !errors.Is(await(t, f.ctx, waiting), context.Canceled) {
			t.Fatal("cancellation not propagated")
		}
	})
	t.Run("claimed turn", func(t *testing.T) {
		f := newFixture(t, nil, "A", 10, true)
		ctx, cancel := context.WithCancel(f.ctx)
		defer cancel()
		f.gate.ctx = ctx
		f.enqueue(t, invoke(1, "blocked"))
		done := runAsync(ctx, f.kernel)
		select {
		case <-f.gate.entered:
		case <-f.ctx.Done():
			t.Fatal("no claim")
		}
		cancel()
		if !errors.Is(await(t, f.ctx, done), context.Canceled) {
			t.Fatal("claimed dispatch not canceled")
		}
		if len(f.kernel.scheduler.realizing) != 0 {
			t.Fatal("claim leaked")
		}
		f.kernel.runtimeState.active.retire("A")
		if f.echo.binding.live.Load() {
			t.Fatal("binding survived retirement")
		}
		if _, ok := f.kernel.runtimeState.active.ownerOf(10); ok {
			t.Fatal("reservation survived retirement")
		}
	})
}

func TestMalformedStartup(t *testing.T) {
	for _, args := range [][]string{nil, {"{"}, {`{"echo":"unused","observations":["one"]}`}, {`{"echo":"unused","observations":["a","b"],"authority":"B"}`}, {`{"echo":"unused","observations":["a","b"]} {}`}} {
		trace := &Trace{}
		_, err := boot(guard(t), args, trace)
		rejected(t, err)
		if len(trace.snapshot()) != 0 {
			t.Fatal("startup performed work before validation")
		}
	}
}

func Example_startupEncoding() {
	b, err := decodeStartup([]string{`{"echo":"./artifacts/echo","observations":["hello","again"]}`})
	fmt.Println(b.Observations, err)
	// Output: [hello again] <nil>
}
