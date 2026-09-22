package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	p "skyra-v1/src/skyra/protocol"
	"syscall"
	"time"
)

// userConfig launches the external `user` executable and an agent for it to
// drive. For this first proof the agent is `echo`, so the whole
// browser -> user -> root -> echo -> Return -> user loop runs with no provider.
type userConfig struct {
	Binary string // the user executable
	Echo   string // the agent executable (echo for this proof)
	Addr   string // browser listen address; "" picks an ephemeral loopback port
	Stderr io.Writer
}

// userBoot wires the user binary in exactly as codexBoot wires Codex: a bound
// stdio Move port for the external peer, echo installed as a host child in its
// Horizon, an external caller endpoint and a root sink. Root does not change
// any Move ownership boundary; user imports no root package and vice versa.
type userBoot struct {
	kernel          *Kernel
	execution       *ExecutionContext
	actor           *helper
	echo            *helper
	client, results *endpoint
	inputs          []*Ingress
	cancel          context.CancelFunc
	done            chan error
	interactID      uint64
	addr            string
	closed          bool
}

func bootUser(ctx context.Context, c userConfig, trace *Trace) (r *userBoot, err error) {
	r = &userBoot{kernel: newKernel(trace)}
	defer func() {
		if err != nil {
			r.close()
		}
	}()

	if c.Addr == "" {
		l, e := net.Listen("tcp", "127.0.0.1:0")
		if e != nil {
			return r, e
		}
		c.Addr = l.Addr().String()
		_ = l.Close()
	}
	r.addr = c.Addr

	cmd := exec.Command(c.Binary)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), "SKYRA_USER_ADDR="+c.Addr)
	cmd.Stderr = c.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return r, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return r, err
	}
	r.actor = &helper{command: cmd, endpoint: bind(&duplex{ReadCloser: stdout, writer: stdin}, "root", 20, p.User, 1, false)}
	r.echo, err = prepareHelper(ctx, c.Echo, "root", 21)
	if err != nil {
		return r, err
	}
	external, client := pair("root", 22, 0, true)
	sink, results := pair("root", 23, 0, false)
	r.client, r.results = client, results

	e := newExecution("root", trace, sink.egress)
	r.execution = e
	userDef := Executable{ID: p.User, Contracts: []ArgContract{{Operation: "interact"}}}
	echoDef := Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}
	r.kernel.runtimeState.memory.register(userDef)
	r.kernel.runtimeState.memory.register(echoDef)
	for i, item := range []struct {
		ep  *endpoint
		def Executable
	}{{r.actor.endpoint, userDef}, {r.echo.endpoint, echoDef}, {external, Executable{}}, {sink, Executable{}}} {
		if err = e.install(uint64(i), item.ep, item.def); err != nil {
			return r, err
		}
	}
	r.kernel.runtimeState.memory.retain(e)
	if err = r.kernel.runtimeState.active.provision(ctx, e, []uint64{20, 21, 22, 23}); err != nil {
		return r, err
	}
	if err = r.echo.launch(); err != nil {
		return r, err
	}
	if err = r.actor.launch(); err != nil {
		return r, err
	}
	trace.record("launcher", "root", 0, 0, "user and echo started")

	// Wait for the browser listener before returning, so a caller can connect
	// without racing the child process's startup.
	probeCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	defer stop()
	for {
		if conn, e := net.Dial("tcp", r.addr); e == nil {
			_ = conn.Close()
			break
		}
		if probeCtx.Err() != nil {
			return r, fmt.Errorf("user browser listener never came up on %s", r.addr)
		}
		time.Sleep(20 * time.Millisecond)
	}

	r.inputs = []*Ingress{external.ingress, r.actor.endpoint.ingress, r.echo.endpoint.ingress}
	loopCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.done = make(chan error, 1)
	go func() { r.done <- r.kernel.loop(loopCtx, r.inputs) }()

	// The `interact` activation opens the persistent interaction U. It is a
	// normal external Invoke; U stays alive until an explicit close returns it.
	r.interactID = 1
	if err = r.client.egress.Send(ctx, p.Move{Identity: r.interactID, Target: p.User, Invoke: &p.Invoke{Args: p.Arguments("interact", "")}}); err != nil {
		return r, err
	}
	trace.record("root", "root", r.interactID, 0, "interact sent; U opened")
	return r, nil
}

// awaitClose receives U's single terminal Return (interaction-closed) from the
// root sink, correlated to the original interact Identity.
func (r *userBoot) awaitClose(ctx context.Context) (p.Move, error) {
	return r.results.ingress.Receive(ctx)
}

// userDemo boots the user+echo stack, opens one interaction, and serves the
// browser until the interaction is closed or the process is interrupted. Point
// the frontend's NEXT_PUBLIC_WS_URL at the printed ws:// address.
func userDemo(c userConfig, out io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	trace := &Trace{output: out}
	r, err := bootUser(ctx, c, trace)
	if err != nil {
		return err
	}
	defer r.close()
	fmt.Fprintf(os.Stderr, "user interaction open — point the browser at ws://%s/ws (Ctrl-C to stop)\n", r.addr)
	done := make(chan struct{})
	go func() { _, _ = r.awaitClose(ctx); close(done) }()
	select {
	case <-ctx.Done():
	case <-done:
	}
	return nil
}

func (r *userBoot) close() {
	if r.closed {
		return
	}
	r.closed = true
	if r.cancel != nil {
		r.cancel()
		<-r.done
	}
	if r.execution != nil {
		r.kernel.runtimeState.active.retire(r.execution.identity)
		r.execution.invalidate()
	}
	if r.client != nil {
		r.client.Close()
	}
	if r.results != nil {
		r.results.Close()
	}
	if r.actor != nil {
		r.actor.endpoint.Close()
		if r.actor.command.Process != nil && !r.actor.waited {
			done := make(chan error, 1)
			go func() { done <- r.actor.command.Wait() }()
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				_ = syscall.Kill(-r.actor.command.Process.Pid, syscall.SIGKILL)
				<-done
			}
			r.actor.waited = true
			r.kernel.trace.record("launcher", "root", 0, 0, "user reaped")
		}
	}
	if r.echo != nil {
		_ = r.echo.close()
	}
	r.kernel.runtimeState.memory.mu.Lock()
	delete(r.kernel.runtimeState.memory.executables, p.User)
	delete(r.kernel.runtimeState.memory.contexts, "root")
	r.kernel.runtimeState.memory.mu.Unlock()
}
