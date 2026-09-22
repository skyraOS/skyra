package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	p "skyra-v1/src/skyra/protocol"
	"sync"
	"syscall"
	"time"
)

// syncWriter serializes concurrent writes so two child processes may share one
// diagnostic sink (e.g. a bytes.Buffer) without racing.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *syncWriter) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(b)
}

// userCodexConfig wires the browser-facing `user` executable to the real
// `codex` agent in one execution context: user, codex, echo (codex's tool) and
// a provider. The host connects codex and mints the interaction's parent; the
// user relays a browser message into a Codex thread through the isolation fence.
type userCodexConfig struct {
	UserBinary   string
	CodexBinary  string
	Descriptor   string
	Provider     string
	Echo         string
	Addr         string // browser listen address; "" picks an ephemeral loopback port
	ProviderArgs []string
	Grants       map[string]bool
	Stderr       io.Writer
}

type userCodexBoot struct {
	kernel          *Kernel
	execution       *ExecutionContext
	user            *helper
	codex           *helper
	echo            *helper
	client, results *endpoint
	inputs          []*Ingress
	cancel          context.CancelFunc
	done            chan error
	connection      string
	interactID      uint64
	addr            string
	closed          bool
}

func launchPeer(bin string, args []string, env []string, stderr io.Writer, owner p.ContextID, port uint64, target p.ExecutableID) (*helper, error) {
	cmd := exec.Command(bin, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = append(os.Environ(), env...)
	cmd.Stderr = stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	return &helper{command: cmd, endpoint: bind(&duplex{ReadCloser: stdout, writer: stdin}, owner, port, target, 1, false)}, nil
}

func bootUserCodex(ctx context.Context, c userCodexConfig, trace *Trace) (r *userCodexBoot, err error) {
	r = &userCodexBoot{kernel: newKernel(trace)}
	defer func() {
		if err != nil {
			r.close()
		}
	}()

	raw, err := os.ReadFile(c.Descriptor)
	if err != nil {
		return r, err
	}
	d, err := p.LoadDescriptor(raw)
	if err != nil {
		return r, err
	}
	r.kernel.runtimeState.memory.register(definition(d, c.Grants))

	if c.Addr == "" {
		l, e := net.Listen("tcp", "127.0.0.1:0")
		if e != nil {
			return r, e
		}
		c.Addr = l.Addr().String()
		_ = l.Close()
	}
	r.addr = c.Addr

	keyText, namespace := randomToken(), randomToken()
	key, _ := hex.DecodeString(keyText)

	// user and codex may share one diagnostic sink; serialize it.
	diag := c.Stderr
	if diag != nil {
		diag = &syncWriter{w: diag}
	}
	r.codex, err = launchPeer(c.CodexBinary, append([]string{c.Provider}, c.ProviderArgs...),
		[]string{"SKYRA_PARENT_KEY=" + keyText, "SKYRA_PARENT_NAMESPACE=" + namespace}, diag, "root", 31, p.Codex)
	if err != nil {
		return r, err
	}
	r.user, err = launchPeer(c.UserBinary, nil, []string{"SKYRA_USER_ADDR=" + c.Addr}, diag, "root", 30, p.User)
	if err != nil {
		return r, err
	}
	r.echo, err = prepareHelper(ctx, c.Echo, "root", 32)
	if err != nil {
		return r, err
	}
	external, client := pair("root", 33, 0, true)
	sink, results := pair("root", 34, 0, false)
	r.client, r.results = client, results

	e := newExecution("root", trace, sink.egress)
	r.execution = e
	e.bookkeeper.key = key
	e.bookkeeper.namespace = namespace
	userDef := Executable{ID: p.User, Contracts: []ArgContract{{Operation: "interact"}}}
	r.kernel.runtimeState.memory.register(userDef)
	for i, item := range []struct {
		ep  *endpoint
		def Executable
	}{
		{r.user.endpoint, userDef},
		{r.codex.endpoint, definition(d, c.Grants)},
		{r.echo.endpoint, Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}},
		{external, Executable{}},
		{sink, Executable{}},
	} {
		if err = e.install(uint64(i), item.ep, item.def); err != nil {
			return r, err
		}
	}
	r.kernel.runtimeState.memory.retain(e)
	if err = r.kernel.runtimeState.active.provision(ctx, e, []uint64{30, 31, 32, 33, 34}); err != nil {
		return r, err
	}
	if err = r.echo.launch(); err != nil {
		return r, err
	}
	if err = r.codex.launch(); err != nil {
		return r, err
	}
	if err = r.user.launch(); err != nil {
		return r, err
	}
	trace.record("launcher", "root", 0, 0, "user, codex and echo started")

	r.inputs = []*Ingress{external.ingress, r.user.endpoint.ingress, r.codex.endpoint.ingress, r.echo.endpoint.ingress}
	loopCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.done = make(chan error, 1)
	go func() { r.done <- r.kernel.loop(loopCtx, r.inputs) }()

	// The host connects Codex and mints U's parent — the user never mints or
	// forges either; it relays them into the Codex thread.
	digest := sha256.Sum256(raw)
	connectCtx, stop := context.WithTimeout(ctx, 6*time.Second)
	defer stop()
	connect, err := r.call(connectCtx, 1, p.Args("connect", "configuration", "default", "descriptor", hex.EncodeToString(digest[:])))
	if err != nil {
		return r, err
	}
	if connect.Return == nil || connect.Return.Observation.Key != "connection" || len(connect.Return.Observation.Bytes) != 64 {
		return r, fmt.Errorf("connect failed: %+v", connect.Fault)
	}
	r.connection = string(connect.Return.Observation.Bytes)
	parent, err := r.execution.mintParent(p.Codex)
	if err != nil {
		return r, err
	}
	cfg, _ := json.Marshal(map[string]string{
		"agent": "codex", "connection": r.connection, "parent": parent, "tools": `["echo"]`,
	})

	// Wait for the browser listener, then open U with the provisioned config.
	probeCtx, stopProbe := context.WithTimeout(ctx, 3*time.Second)
	defer stopProbe()
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
	r.interactID = 2
	if err = r.client.egress.Send(ctx, p.Move{Identity: r.interactID, Target: p.User, Invoke: &p.Invoke{Args: p.Arguments("interact", string(cfg))}}); err != nil {
		return r, err
	}
	trace.record("root", "root", r.interactID, 0, "interact sent; U opened over Codex")
	return r, nil
}

func (r *userCodexBoot) call(ctx context.Context, id uint64, args []p.Observation) (p.Move, error) {
	if err := r.client.egress.Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}); err != nil {
		return p.Move{}, err
	}
	return r.results.ingress.Receive(ctx)
}

func (r *userCodexBoot) awaitClose(ctx context.Context) (p.Move, error) {
	return r.results.ingress.Receive(ctx)
}

// userCodexDemo boots the user↔codex stack and serves the browser until the
// interaction is closed or the process is interrupted. Point the frontend's
// NEXT_PUBLIC_WS_URL at the printed ws:// address. Provider is whatever binary
// Codex should launch (the fake provider for a deterministic dev loop, or the
// real codex app-server).
func userCodexDemo(c userCodexConfig, out io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	trace := &Trace{output: out}
	r, err := bootUserCodex(ctx, c, trace)
	if err != nil {
		return err
	}
	defer r.close()
	fmt.Fprintf(os.Stderr, "user↔codex interaction open — point the browser at ws://%s/ws (Ctrl-C to stop)\n", r.addr)
	done := make(chan struct{})
	go func() { _, _ = r.awaitClose(ctx); close(done) }()
	select {
	case <-ctx.Done():
	case <-done:
	}
	return nil
}

func reap(h *helper, trace *Trace, label string) {
	if h == nil || h.command == nil || h.command.Process == nil || h.waited {
		return
	}
	h.endpoint.Close()
	done := make(chan error, 1)
	go func() { done <- h.command.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-h.command.Process.Pid, syscall.SIGKILL)
		<-done
	}
	h.waited = true
	trace.record("launcher", "root", 0, 0, label+" reaped")
}

func (r *userCodexBoot) close() {
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
	reap(r.user, r.kernel.trace, "user")
	reap(r.codex, r.kernel.trace, "codex")
	if r.echo != nil {
		_ = r.echo.close()
	}
	r.kernel.runtimeState.memory.mu.Lock()
	delete(r.kernel.runtimeState.memory.executables, p.User)
	delete(r.kernel.runtimeState.memory.executables, p.Codex)
	delete(r.kernel.runtimeState.memory.contexts, "root")
	r.kernel.runtimeState.memory.mu.Unlock()
}
