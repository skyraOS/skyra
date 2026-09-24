package kernel_test

import (
	"context"
	"errors"
	"net"
	"testing"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

func TestReservationConflictAtomic(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	x, y := net.Pipe()
	defer x.Close()
	defer y.Close()
	c := execution.New("C", a.Kernel.Trace(), nil)
	must(t, c.Install(0, ports.Bind(x, "C", 99, p.Echo, 1, false), execution.Executable{ID: p.Echo}))
	rejected(t, a.Kernel.State().Active.Provision(a.Ctx, c, []uint64{99, 10}))
	if _, ok := a.Kernel.State().Active.Lookup("C"); ok {
		t.Fatal("partial context publication")
	}
	if _, ok := a.Kernel.State().Active.OwnerOf(99); ok {
		t.Fatal("partial reservation")
	}
	if id, ok := a.Kernel.State().Active.OwnerOf(10); !ok || id != "A" {
		t.Fatal("existing reservation changed")
	}
}

func TestForgedOwnershipAndVirtualPortIsolation(t *testing.T) {
	a := newFixture(t, nil, "A", 10, false)
	b := newFixture(t, a.Kernel, "B", 20, false)
	wa := a.Start(t, invoke(1, "A text"))
	wb := b.Start(t, invoke(1, "B text"))
	bad := returned(wa, "forged")
	bad.Context = "B"
	bad = exchange(t, a.Ctx, a.Peer.Egress(), a.Echo.Ingress(), bad)
	rejected(t, a.Kernel.Admit(a.Ctx, a.Echo.Ingress(), bad))
	if frames, queued := b.Counts(); frames != 1 || queued != 0 {
		t.Fatal("forgery changed B")
	}
	ra := a.Finish(t, returned(wa, "A text"))
	rb := b.Finish(t, returned(wb, "B text"))
	if ra.Context != "A" || rb.Context != "B" || string(ra.Return.Observation.Bytes) != "A text" || string(rb.Return.Observation.Bytes) != "B text" {
		t.Fatal("cross-context result")
	}
	if a.Execution.Mappings()[0].Actual == b.Execution.Mappings()[0].Actual {
		t.Fatal("virtual 0 aliases actual connection")
	}
	t.Log("virtual port 0 delivered distinct A and B results")
}

func TestFixedDestinationAndContractRejection(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	bad := invoke(1, "wrong")
	bad.Target = p.Parent
	rejected(t, f.Echo.Egress().Send(f.Ctx, bad))
	bad = invoke(1, "wrong")
	bad.Invoke.Args = p.Arguments("unknown", "wrong")
	f.Enqueue(t, bad)
	rejected(t, f.Kernel.Scheduler().Run(f.Ctx, f.Kernel.Dispatch))
	if frames, queued := f.Counts(); frames != 0 || queued != 0 {
		t.Fatal("malformed call changed frames")
	}
	// With the peer unread, any erroneous helper send would have blocked the Run.
	wire := f.Start(t, invoke(2, "accepted"))
	if wire.Identity != 2 {
		t.Fatal("helper received rejected call")
	}
}

func TestSameContextSerializationAndDifferentContextProgress(t *testing.T) {
	a := newFixture(t, nil, "A", 10, true)
	b := newFixture(t, a.Kernel, "B", 20, false)
	a.Enqueue(t, invoke(1, "first"))
	first := runAsync(a.Ctx, a.Kernel)
	select {
	case <-a.Gate.Entered:
	case <-a.Ctx.Done():
		t.Fatal("first Realize never entered Send")
	}
	a.Enqueue(t, invoke(2, "second"))
	b.Enqueue(t, invoke(3, "B advances"))
	other := runAsync(b.Ctx, b.Kernel)
	bwire, err := b.Peer.Ingress().Receive(b.Ctx)
	must(t, err)
	must(t, await(t, b.Ctx, other))
	if bwire.Identity != 3 {
		t.Fatal("B did not progress")
	}
	n := len(a.Execution.Frames())
	if n != 1 {
		t.Fatal("overlapping A state turn")
	}
	close(a.Gate.Release)
	awire, err := a.Peer.Ingress().Receive(a.Ctx)
	must(t, err)
	must(t, await(t, a.Ctx, first))
	next := runAsync(a.Ctx, a.Kernel)
	second, err := a.Peer.Ingress().Receive(a.Ctx)
	must(t, err)
	must(t, await(t, a.Ctx, next))
	if awire.Identity != 1 || second.Identity != 2 {
		t.Fatal("admission order not retained")
	}
	t.Log("B advanced while A Realize was held; A turns retained order")
}

func TestDispatchFailureReleasesClaim(t *testing.T) {
	k := kernel.New(&trace.Trace{})
	ctx := guard(t)
	for i := uint64(1); i <= 2; i++ {
		m := invoke(i, "")
		m.Context = "missing"
		must(t, k.Scheduler().Enqueue(execution.Work{Move: m}))
	}
	rejected(t, k.Scheduler().Run(ctx, k.Dispatch))
	called := false
	rejected(t, k.Scheduler().Run(ctx, func(context.Context, execution.Work) error {
		called = true
		return errors.New("dispatch fixture failure")
	}))
	if !called || k.Scheduler().Claims() != 0 {
		t.Fatal("claim leaked")
	}
	if _, ok := k.State().Active.Lookup("missing"); ok {
		t.Fatal("implicit provision")
	}
	f := newFixture(t, k, "live", 10, false)
	_ = f.Start(t, invoke(3, "progress"))
}

func TestStaleOutputAndDuplicateCompletion(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	wire := f.Start(t, invoke(1, "value"))
	terminal := returned(wire, "value")
	// Two copies can enter before settlement; the second must fail at realization.
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), terminal))
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), terminal))
	done := runAsync(f.Ctx, f.Kernel)
	_, err := f.Result.Ingress().Receive(f.Ctx)
	must(t, err)
	must(t, await(t, f.Ctx, done))
	rejected(t, f.Kernel.Scheduler().Run(f.Ctx, f.Kernel.Dispatch))
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), terminal))
	oldWire := f.Start(t, invoke(2, "old"))
	oldOutput := returned(oldWire, "old")
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), oldOutput))
	x, y := net.Pipe()
	defer y.Close()
	replacement := ports.Bind(x, "A", 10, p.Echo, 2, false)
	must(t, f.Execution.Rebind(0, replacement))
	rejected(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), oldOutput))
	rejected(t, f.Kernel.Scheduler().Run(f.Ctx, f.Kernel.Dispatch))
	phase := f.Execution.Frames()[oldWire.Invoke.Frame].Phase
	if phase != "running" {
		t.Fatal("stale queued return settled frame")
	}
	n := 0
	for _, e := range f.Kernel.Trace().Snapshot() {
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
		ctx, cancel := context.WithCancel(f.Ctx)
		defer cancel()
		received := make(chan error, 1)
		x, y := net.Pipe()
		defer y.Close()
		reader := &readSignalConnection{ReadWriteCloser: x, entered: make(chan struct{})}
		endpoint := ports.Bind(reader, "A", 0, 0, 1, false)
		defer endpoint.Close()
		go func() { _, err := endpoint.Ingress().Receive(ctx); received <- err }()
		select {
		case <-reader.entered:
		case <-f.Ctx.Done():
			t.Fatal("receiver never blocked")
		}
		waitingContext := &waitSignalContext{Context: ctx, entered: make(chan struct{})}
		waiting := runAsync(waitingContext, f.Kernel)
		select {
		case <-waitingContext.entered:
		case <-f.Ctx.Done():
			t.Fatal("scheduler never waited")
		}
		cancel()
		if !errors.Is(await(t, f.Ctx, received), context.Canceled) || !errors.Is(await(t, f.Ctx, waiting), context.Canceled) {
			t.Fatal("cancellation not propagated")
		}
	})
	t.Run("claimed turn", func(t *testing.T) {
		f := newFixture(t, nil, "A", 10, true)
		ctx, cancel := context.WithCancel(f.Ctx)
		defer cancel()
		f.Gate.Ctx = ctx
		f.Enqueue(t, invoke(1, "blocked"))
		done := runAsync(ctx, f.Kernel)
		select {
		case <-f.Gate.Entered:
		case <-f.Ctx.Done():
			t.Fatal("no claim")
		}
		cancel()
		if !errors.Is(await(t, f.Ctx, done), context.Canceled) {
			t.Fatal("claimed dispatch not canceled")
		}
		if f.Kernel.Scheduler().Claims() != 0 {
			t.Fatal("claim leaked")
		}
		f.Kernel.State().Active.Retire("A")
		if f.Echo.Binding().Live() {
			t.Fatal("binding survived retirement")
		}
		if _, ok := f.Kernel.State().Active.OwnerOf(10); ok {
			t.Fatal("reservation survived retirement")
		}
	})
}
