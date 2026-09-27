// Package demos owns the scripted scenarios that run a booted system and verify
// its results. Each demo boots through package boot, drives the root client
// endpoint, and records evidence into the trace.
package demos

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/boot"
	"skyra/history/dna19/root/trace"
)

func Entrypoint(ctx context.Context, args []string, output io.Writer) error {
	return baseline(ctx, args, trace.New(output))
}

// Run runs the baseline demo, then the extension demo, on one trace.
func Run(ctx context.Context, args []string, out io.Writer) error {
	t := trace.New(out)
	if err := baseline(ctx, args, t); err != nil {
		return err
	}
	return extension(ctx, args, t)
}

func baseline(ctx context.Context, args []string, t *trace.Trace) error {
	r, err := boot.Root(ctx, args, t)
	if err != nil {
		return err
	}
	loopCtx, cancel := context.WithCancel(ctx)
	loopDone := make(chan error, 1)
	go func() { loopDone <- r.Kernel.Loop(loopCtx, r.Ingress) }()
	defer func() { cancel(); <-loopDone; r.Shutdown() }()
	for i, text := range append(append([]string(nil), r.Config.Observations...), "child") {
		target, operation, expected := p.Echo, "echo", text
		if i == 2 {
			target, operation, expected = p.Parent, "parent", "parent:child"
		}
		request := p.Move{Identity: uint64(i + 1), Target: target, Invoke: &p.Invoke{Args: p.Arguments(operation, text)}}
		if err := r.Client.Egress().Send(ctx, request); err != nil {
			return err
		}
		result, err := r.Results.Ingress().Receive(ctx)
		if err != nil {
			return err
		}
		if result.Identity != request.Identity || result.Context != "root" || result.Return == nil || string(result.Return.Observation.Bytes) != expected {
			return fmt.Errorf("incorrect correlated result: %+v", result)
		}
		t.Record("root", "root", result.Identity, result.Return.Answering, "result verified: "+expected)
		if err := r.Kernel.Scheduler().WaitIdle(ctx, "root"); err != nil {
			return err
		}
	}
	return nil
}

func Codex(ctx context.Context, c boot.CodexConfig, out io.Writer) error {
	t := trace.New(out)
	r, err := boot.Codex(ctx, c, t)
	if err != nil {
		return err
	}
	defer r.Close()
	profiles, err := r.Execution.Capabilities(p.Codex)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		t.Record("planner", "root", 0, 0, "admitted profile: "+profile.Operation)
	}
	p1, err := r.Execution.MintParent(p.Codex)
	if err != nil {
		return err
	}
	p2, err := r.Execution.MintParent(p.Codex)
	if err != nil {
		return err
	}
	calls := [][]p.Observation{
		p.Args("start-thread", "connection", r.Connection, "parent", p1, "text", "hello", "tools", `["echo"]`),
		p.Args("submit-input", "parent", p1, "text", "again"),
		p.Args("start-thread", "connection", r.Connection, "parent", p2, "text", "other", "tools", `["echo"]`),
		p.Args("submit-input", "parent", p1, "text", "tool:child"),
		p.Args("submit-input", "parent", p2, "text", "taint"),
	}
	for i, args := range calls {
		result, err := r.Call(ctx, uint64(i+2), args)
		if err != nil {
			return err
		}
		if i == 4 {
			if result.Fault == nil {
				return fmt.Errorf("contamination returned success")
			}
			t.Record("demo", "root", result.Identity, result.Source, "contamination rejected: "+result.Fault.Message)
		} else {
			if result.Return == nil {
				return fmt.Errorf("unexpected command fault: %+v", result.Fault)
			}
			t.Record("demo", "root", result.Identity, result.Source, "command result: "+string(result.Return.Observation.Bytes))
		}
	}
	return nil
}

// User boots the user+echo stack, opens one interaction, and serves the
// browser until the interaction is closed or the process is interrupted. Point
// the frontend's NEXT_PUBLIC_WS_URL at the printed ws:// address.
func User(c boot.UserConfig, out io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	r, err := boot.User(ctx, c, trace.New(out))
	if err != nil {
		return err
	}
	defer r.Close()
	fmt.Fprintf(os.Stderr, "user interaction open — point the browser at ws://%s/ws (Ctrl-C to stop)\n", r.Addr)
	done := make(chan struct{})
	go func() { _, _ = r.AwaitClose(ctx); close(done) }()
	select {
	case <-ctx.Done():
	case <-done:
	}
	return nil
}

// UserCodex boots the user↔codex stack and serves the browser until the
// interaction is closed or the process is interrupted. Point the frontend's
// NEXT_PUBLIC_WS_URL at the printed ws:// address. Provider is whatever binary
// Codex should launch (the fake provider for a deterministic dev loop, or the
// real codex app-server).
func UserCodex(c boot.UserCodexConfig, out io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	r, err := boot.UserCodex(ctx, c, trace.New(out))
	if err != nil {
		return err
	}
	defer r.Close()
	fmt.Fprintf(os.Stderr, "user↔codex interaction open — point the browser at ws://%s/ws (Ctrl-C to stop)\n", r.Addr)
	done := make(chan struct{})
	go func() { _, _ = r.AwaitClose(ctx); close(done) }()
	select {
	case <-ctx.Done():
	case <-done:
	}
	return nil
}
