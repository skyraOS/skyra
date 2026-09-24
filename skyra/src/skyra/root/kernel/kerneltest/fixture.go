// Package kerneltest provides a published context wired to in-process echo
// and client endpoints, so tests can drive admission and dispatch Move by Move.
package kerneltest

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

type Fixture struct {
	Ctx       context.Context
	Kernel    *kernel.Kernel
	Execution *execution.ExecutionContext
	External  *ports.Endpoint
	Client    *ports.Endpoint
	Echo      *ports.Endpoint
	Peer      *ports.Endpoint
	Result    *ports.Endpoint
	Gate      *GateConnection
}

// GateConnection holds the first Write until Release closes or Ctx ends.
type GateConnection struct {
	io.ReadWriteCloser
	Entered chan struct{}
	Release chan struct{}
	Ctx     context.Context
	once    sync.Once
}

func (g *GateConnection) Write(data []byte) (int, error) {
	g.once.Do(func() { close(g.Entered) })
	select {
	case <-g.Release:
		return g.ReadWriteCloser.Write(data)
	case <-g.Ctx.Done():
		return 0, g.Ctx.Err()
	}
}

func Guard(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}
func Must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func Rejected(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected boundary rejection")
	}
	t.Log("rejected:", err)
}
func Await(t *testing.T, ctx context.Context, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		t.Fatal("completion guard:", ctx.Err())
		return ctx.Err()
	}
}

// New publishes context id with echo at actual port base, an external caller
// at base+1 and a root sink at base+2. A nil k starts a fresh Kernel.
func New(t *testing.T, k *kernel.Kernel, id p.ContextID, base uint64, gated bool) *Fixture {
	t.Helper()
	ctx := Guard(t)
	if k == nil {
		k = kernel.New(&trace.Trace{})
	}
	f := &Fixture{Ctx: ctx, Kernel: k}
	in, remote := net.Pipe()
	sink, result := net.Pipe()
	echo, peer := net.Pipe()
	f.External = ports.Bind(in, id, base+1, 0, 1, true)
	f.Client = ports.Bind(remote, id, 0, 0, 1, false)
	f.Result = ports.Bind(result, id, 0, 0, 1, false)
	f.Peer = ports.Bind(peer, id, 0, 0, 1, false)
	if gated {
		f.Gate = &GateConnection{ReadWriteCloser: echo, Entered: make(chan struct{}), Release: make(chan struct{}), Ctx: ctx}
		f.Echo = ports.Bind(f.Gate, id, base, p.Echo, 1, false)
	} else {
		f.Echo = ports.Bind(echo, id, base, p.Echo, 1, false)
	}
	rootSink := ports.Bind(sink, id, base+2, 0, 1, false)
	f.Execution = execution.New(id, k.Trace(), rootSink.Egress())
	Must(t, f.Execution.Install(0, f.Echo, execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}))
	Must(t, f.Execution.Install(1, f.External, execution.Executable{}))
	Must(t, f.Execution.Install(2, rootSink, execution.Executable{}))
	rt := k.State()
	rt.Memory.Retain(f.Execution)
	Must(t, rt.Active.Provision(ctx, f.Execution, []uint64{base, base + 1, base + 2}))
	t.Cleanup(func() { rt.Active.Retire(id); _ = f.Peer.Close(); _ = f.Result.Close(); _ = remote.Close() })
	return f
}
func Invoke(id uint64, text string) p.Move {
	return p.Move{Identity: id, Target: p.Echo, Invoke: &p.Invoke{Args: p.Arguments("echo", text)}}
}
func (f *Fixture) Enqueue(t *testing.T, m p.Move) {
	t.Helper()
	wire := Exchange(t, f.Ctx, f.Client.Egress(), f.External.Ingress(), m)
	Must(t, f.Kernel.Admit(f.Ctx, f.External.Ingress(), wire))
}

func Exchange(t *testing.T, ctx context.Context, out *ports.Egress, in *ports.Ingress, m p.Move) p.Move {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- out.Send(ctx, m) }()
	wire, err := in.Receive(ctx)
	Must(t, err)
	Must(t, Await(t, ctx, done))
	return wire
}

// RunAsync runs one Scheduler turn with the Kernel's dispatch.
func RunAsync(ctx context.Context, k *kernel.Kernel) <-chan error {
	done := make(chan error, 1)
	go func() { done <- k.Scheduler().Run(ctx, k.Dispatch) }()
	return done
}
func (f *Fixture) Start(t *testing.T, m p.Move) p.Move {
	t.Helper()
	f.Enqueue(t, m)
	done := RunAsync(f.Ctx, f.Kernel)
	wire, err := f.Peer.Ingress().Receive(f.Ctx)
	Must(t, err)
	Must(t, Await(t, f.Ctx, done))
	return wire
}
func Returned(wire p.Move, text string) p.Move {
	return p.Move{Identity: wire.Identity, Context: wire.Context, Source: wire.Invoke.Frame, Return: &p.Return{Answering: wire.Invoke.Frame, Observation: p.Observation{Key: "text", Bytes: []byte(text)}}}
}
func (f *Fixture) Finish(t *testing.T, m p.Move) p.Move {
	t.Helper()
	wire := Exchange(t, f.Ctx, f.Peer.Egress(), f.Echo.Ingress(), m)
	Must(t, f.Kernel.Admit(f.Ctx, f.Echo.Ingress(), wire))
	done := RunAsync(f.Ctx, f.Kernel)
	result, err := f.Result.Ingress().Receive(f.Ctx)
	Must(t, err)
	Must(t, Await(t, f.Ctx, done))
	return result
}

// Counts reports Bookkeeper frames and queued Moves.
func (f *Fixture) Counts() (int, int) {
	return len(f.Execution.Frames()), f.Kernel.Scheduler().Queued()
}

// TraceCount counts owner's transitions that start with prefix.
func TraceCount(k *kernel.Kernel, owner p.ContextID, prefix string) int {
	n := 0
	for _, e := range k.Trace().Snapshot() {
		if e.Context == owner && strings.HasPrefix(e.Transition, prefix) {
			n++
		}
	}
	return n
}

// LiveLoop runs the Kernel loop; the returned stop cancels it and asserts that
// shutdown released every claim, context and reservation.
func LiveLoop(t *testing.T, ctx context.Context, k *kernel.Kernel, inputs []*ports.Ingress) func() {
	t.Helper()
	shutdownGuard := Guard(t)
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- k.Loop(ctx, inputs) }()
	return func() {
		cancel()
		if err := Await(t, shutdownGuard, done); !errors.Is(err, context.Canceled) {
			t.Errorf("loop failed before explicit cancellation: %v", err)
		}
		if k.Scheduler().Claims() != 0 {
			t.Error("claims leaked")
		}
		if contexts, reservations := k.State().Active.Size(); reservations != 0 || contexts != 0 {
			t.Error("shutdown retained authority")
		}
	}
}
