package boot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
	"sync"
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

// UserCodexConfig wires the browser-facing `user` executable to the real
// `codex` agent in one execution context: user, codex, echo (codex's tool) and
// a provider. The host connects codex and mints the interaction's parent; the
// user relays a browser message into a Codex thread through the isolation fence.
type UserCodexConfig struct {
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

type UserCodexBoot struct {
	Kernel          *kernel.Kernel
	Execution       *execution.ExecutionContext
	User            *ports.Helper
	Codex           *ports.Helper
	Echo            *ports.Helper
	Client, Results *ports.Endpoint
	inputs          []*ports.Ingress
	cancel          context.CancelFunc
	done            chan error
	Connection      string
	InteractID      uint64
	Addr            string
	closed          bool
}

func UserCodex(ctx context.Context, c UserCodexConfig, t *trace.Trace) (r *UserCodexBoot, err error) {
	r = &UserCodexBoot{Kernel: kernel.New(t)}
	defer func() {
		if err != nil {
			r.Close()
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
	rt := r.Kernel.State()
	rt.Memory.Register(execution.Definition(d, c.Grants))

	if r.Addr, err = listenAddr(c.Addr); err != nil {
		return r, err
	}

	keyText, namespace := execution.RandomToken(), execution.RandomToken()
	key, _ := hex.DecodeString(keyText)

	// user and codex may share one diagnostic sink; serialize it.
	diag := c.Stderr
	if diag != nil {
		diag = &syncWriter{w: diag}
	}
	r.Codex, err = ports.PreparePeer(c.CodexBinary, append([]string{c.Provider}, c.ProviderArgs...),
		[]string{"SKYRA_PARENT_KEY=" + keyText, "SKYRA_PARENT_NAMESPACE=" + namespace}, diag, "root", 31, p.Codex)
	if err != nil {
		return r, err
	}
	r.User, err = ports.PreparePeer(c.UserBinary, nil, []string{"SKYRA_USER_ADDR=" + r.Addr}, diag, "root", 30, p.User)
	if err != nil {
		return r, err
	}
	r.Echo, err = ports.PrepareHelper(ctx, c.Echo, "root", 32)
	if err != nil {
		return r, err
	}
	external, client := ports.Pair("root", 33, 0, true)
	sink, results := ports.Pair("root", 34, 0, false)
	r.Client, r.Results = client, results

	e := execution.New("root", t, sink.Egress())
	r.Execution = e
	e.ProvisionParentAuthority(key, namespace)
	userDef := execution.Executable{ID: p.User, Contracts: []execution.ArgContract{{Operation: "interact"}}}
	rt.Memory.Register(userDef)
	for i, item := range []struct {
		ep  *ports.Endpoint
		def execution.Executable
	}{
		{r.User.Endpoint(), userDef},
		{r.Codex.Endpoint(), execution.Definition(d, c.Grants)},
		{r.Echo.Endpoint(), execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}},
		{external, execution.Executable{}},
		{sink, execution.Executable{}},
	} {
		if err = e.Install(uint64(i), item.ep, item.def); err != nil {
			return r, err
		}
	}
	rt.Memory.Retain(e)
	if err = rt.Active.Provision(ctx, e, []uint64{30, 31, 32, 33, 34}); err != nil {
		return r, err
	}
	if err = r.Echo.Launch(); err != nil {
		return r, err
	}
	if err = r.Codex.Launch(); err != nil {
		return r, err
	}
	if err = r.User.Launch(); err != nil {
		return r, err
	}
	t.Record("launcher", "root", 0, 0, "user, codex and echo started")

	r.inputs = []*ports.Ingress{external.Ingress(), r.User.Endpoint().Ingress(), r.Codex.Endpoint().Ingress(), r.Echo.Endpoint().Ingress()}
	r.cancel, r.done = startLoop(ctx, r.Kernel, r.inputs)

	// The host connects Codex and mints U's parent — the user never mints or
	// forges either; it relays them into the Codex thread.
	digest := sha256.Sum256(raw)
	connectCtx, stop := context.WithTimeout(ctx, 6*time.Second)
	defer stop()
	connect, err := r.Call(connectCtx, 1, p.Args("connect", "configuration", "default", "descriptor", hex.EncodeToString(digest[:])))
	if err != nil {
		return r, err
	}
	if connect.Return == nil || connect.Return.Observation.Key != "connection" || len(connect.Return.Observation.Bytes) != 64 {
		return r, fmt.Errorf("connect failed: %+v", connect.Fault)
	}
	r.Connection = string(connect.Return.Observation.Bytes)
	parent, err := r.Execution.MintParent(p.Codex)
	if err != nil {
		return r, err
	}
	cfg, _ := json.Marshal(map[string]string{
		"agent": "codex", "connection": r.Connection, "parent": parent, "tools": `["echo"]`,
	})

	// Wait for the browser listener, then open U with the provisioned config.
	if err = awaitListener(ctx, r.Addr); err != nil {
		return r, err
	}
	r.InteractID = 2
	if err = r.Client.Egress().Send(ctx, p.Move{Identity: r.InteractID, Target: p.User, Invoke: &p.Invoke{Args: p.Arguments("interact", string(cfg))}}); err != nil {
		return r, err
	}
	t.Record("root", "root", r.InteractID, 0, "interact sent; U opened over Codex")
	return r, nil
}

func (r *UserCodexBoot) Call(ctx context.Context, id uint64, args []p.Observation) (p.Move, error) {
	if err := r.Client.Egress().Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}); err != nil {
		return p.Move{}, err
	}
	return r.Results.Ingress().Receive(ctx)
}

func (r *UserCodexBoot) AwaitClose(ctx context.Context) (p.Move, error) {
	return r.Results.Ingress().Receive(ctx)
}

func (r *UserCodexBoot) Close() {
	if r.closed {
		return
	}
	r.closed = true
	if r.cancel != nil {
		r.cancel()
		<-r.done
	}
	rt := r.Kernel.State()
	if r.Execution != nil {
		rt.Active.Retire(r.Execution.Identity())
		r.Execution.Invalidate()
	}
	if r.Client != nil {
		r.Client.Close()
	}
	if r.Results != nil {
		r.Results.Close()
	}
	r.User.Reap(r.Kernel.Trace(), "user reaped")
	r.Codex.Reap(r.Kernel.Trace(), "codex reaped")
	if r.Echo != nil {
		_ = r.Echo.Close()
	}
	rt.Memory.Unregister(p.User, p.Codex)
	rt.Memory.Forget("root")
}
