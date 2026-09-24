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

func TestLiveLoopRejectionContainment(t *testing.T) {
	for _, stale := range []bool{false, true} {
		name := "duplicate"
		if stale {
			name = "binding invalidated after admission"
		}
		t.Run(name, func(t *testing.T) {
			a := newFixture(t, nil, "A", 10, false)
			b := newFixture(t, a.Kernel, "B", 20, false)
			wire := a.Start(t, invoke(1, "old"))
			terminal := returned(wire, "old")
			must(t, a.Kernel.Admit(a.Ctx, a.Echo.Ingress(), terminal))
			if stale {
				left, right := net.Pipe()
				defer right.Close()
				must(t, a.Execution.Rebind(0, ports.Bind(left, "A", 10, p.Echo, 2, false)))
			} else {
				must(t, a.Kernel.Admit(a.Ctx, a.Echo.Ingress(), terminal))
			}
			// Both decoded outputs are queued before the live loop is released. No
			// receive/admission race or sleep is needed to reach realization rejection.
			stop := liveLoop(t, a.Ctx, a.Kernel, []*ports.Ingress{a.Echo.Ingress()})
			defer stop()
			if !stale {
				_, err := a.Result.Ingress().Receive(a.Ctx)
				must(t, err)
			}
			must(t, a.Kernel.Scheduler().WaitIdle(a.Ctx, "A"))
			if traceCount(a.Kernel, "A", "dispatch rejected:") != 1 {
				t.Fatal("queued rejection not recorded")
			}
			expected := 1
			if stale {
				expected = 0
			}
			if traceCount(a.Kernel, "A", "root result sent") != expected {
				t.Fatal("extra root result")
			}
			if stale {
				phase := a.Execution.Frames()[wire.Invoke.Frame].Phase
				if phase != "running" {
					t.Fatal("stale output settled old frame")
				}
			}
			b.Enqueue(t, invoke(2, "still running"))
			other, err := b.Peer.Ingress().Receive(b.Ctx)
			must(t, err)
			must(t, b.Kernel.Admit(b.Ctx, b.Echo.Ingress(), returned(other, "still running")))
			result, err := b.Result.Ingress().Receive(b.Ctx)
			must(t, err)
			if result.Identity != 2 || string(result.Return.Observation.Bytes) != "still running" {
				t.Fatal("other context failed")
			}
			must(t, b.Kernel.Scheduler().WaitIdle(b.Ctx, "B"))
		})
	}
}

func TestRootSinkAndOutputContract(t *testing.T) {
	f := newFixture(t, nil, "A", 10, false)
	if !f.Execution.SinkWired() {
		t.Fatal("sink destination not wired")
	}
	wire := f.Start(t, invoke(1, "output"))
	bad := returned(wire, "output")
	bad.Return.Observation.Key = "executable"
	must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), bad))
	rejected(t, f.Kernel.Scheduler().Run(f.Ctx, f.Kernel.Dispatch))
	if f.Execution.Frames()[wire.Invoke.Frame].Phase != "running" {
		t.Fatal("invalid output settled frame")
	}
	_ = f.Finish(t, returned(wire, "output"))
	if traceCount(f.Kernel, "A", "root sink resolved") != 1 {
		t.Fatal("sink resolution missing")
	}
}

func TestLiveLoopFatalFailures(t *testing.T) {
	t.Run("committed send failure", func(t *testing.T) {
		f := newFixture(t, nil, "A", 10, false)
		f.Enqueue(t, invoke(1, "fail after commit"))
		must(t, f.Peer.Close())
		done := make(chan error, 1)
		go func() { done <- f.Kernel.Loop(f.Ctx, nil) }()
		err := await(t, f.Ctx, done)
		if err == nil {
			t.Fatal("transport loss silently ignored")
		}
		if traceCount(f.Kernel, "A", "transition committed") != 1 || traceCount(f.Kernel, "A", "frame recorded") != 1 {
			t.Fatal("send retried or never committed")
		}
		if f.Kernel.Scheduler().Claims() != 0 {
			t.Fatal("fatal failure leaked claim")
		}
	})
	t.Run("missing context", func(t *testing.T) {
		k := kernel.New(&trace.Trace{})
		ctx := guard(t)
		m := invoke(1, "missing")
		m.Context = "missing"
		must(t, k.Scheduler().Enqueue(execution.Work{Move: m}))
		done := make(chan error, 1)
		go func() { done <- k.Loop(ctx, nil) }()
		rejected(t, await(t, ctx, done))
	})
}

func TestLiveLoopCancellationClosesClaimedTurn(t *testing.T) {
	f := newFixture(t, nil, "A", 10, true)
	ctx, cancel := context.WithCancel(f.Ctx)
	f.Gate.Ctx = ctx
	f.Enqueue(t, invoke(1, "held send"))
	done := make(chan error, 1)
	go func() { done <- f.Kernel.Loop(ctx, []*ports.Ingress{f.External.Ingress()}) }()
	select {
	case <-f.Gate.Entered:
	case <-f.Ctx.Done():
		t.Fatal("send never claimed")
	}
	cancel()
	if !errors.Is(await(t, f.Ctx, done), context.Canceled) {
		t.Fatal("cancellation lost")
	}
	if f.Kernel.Scheduler().Claims() != 0 || f.External.Binding().Live() || f.Echo.Binding().Live() {
		t.Fatal("shutdown leaked claim or endpoint")
	}
	if _, ok := f.Kernel.State().Active.OwnerOf(10); ok {
		t.Fatal("shutdown leaked reservation")
	}
}
