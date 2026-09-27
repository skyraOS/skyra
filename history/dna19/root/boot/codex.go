package boot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/execution"
	"skyra/history/dna19/root/kernel"
	"skyra/history/dna19/root/ports"
	"skyra/history/dna19/root/trace"
	"time"
)

type CodexConfig struct {
	Binary, Descriptor, Provider, Echo string
	ProviderArgs                       []string
	Grants                             map[string]bool
	Stderr                             io.Writer
}
type CodexBoot struct {
	Kernel          *kernel.Kernel
	Execution       *execution.ExecutionContext
	Actor           *ports.Helper
	Echo            *ports.Helper
	Client, Results *ports.Endpoint
	inputs          []*ports.Ingress
	cancel          context.CancelFunc
	done            chan error
	// Connection is the opaque handle returned by the validated connect.
	Connection string
	ready      bool
	closed     bool
}

// Codex registers the Codex descriptor, launches Codex and echo in one root
// context, starts the Kernel loop, and returns once connect is validated.
func Codex(ctx context.Context, c CodexConfig, t *trace.Trace) (r *CodexBoot, err error) {
	r = &CodexBoot{Kernel: kernel.New(t)}
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
	rt.Memory.Register(execution.Definition(d, nil))
	t.Record("memory", "root", 0, 0, "Codex descriptor registered")
	keyText, namespace := execution.RandomToken(), execution.RandomToken()
	key, _ := hex.DecodeString(keyText)
	r.Actor, err = ports.PreparePeer(c.Binary, append([]string{c.Provider}, c.ProviderArgs...),
		[]string{"SKYRA_PARENT_KEY=" + keyText, "SKYRA_PARENT_NAMESPACE=" + namespace}, c.Stderr, "root", 20, p.Codex)
	if err != nil {
		return r, err
	}
	r.Echo, err = ports.PrepareHelper(ctx, c.Echo, "root", 21)
	if err != nil {
		return r, err
	}
	external, client := ports.Pair("root", 22, 0, true)
	sink, results := ports.Pair("root", 23, 0, false)
	r.Client, r.Results = client, results
	e := execution.New("root", t, sink.Egress())
	r.Execution = e
	e.ProvisionParentAuthority(key, namespace)
	for i, item := range []struct {
		ep  *ports.Endpoint
		def execution.Executable
	}{{r.Actor.Endpoint(), execution.Definition(d, c.Grants)}, {r.Echo.Endpoint(), execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}}, {external, execution.Executable{}}, {sink, execution.Executable{}}} {
		if err = e.Install(uint64(i), item.ep, item.def); err != nil {
			return r, err
		}
	}
	rt.Memory.Retain(e)
	if err = rt.Active.Provision(ctx, e, []uint64{20, 21, 22, 23}); err != nil {
		return r, err
	}
	if err = r.Echo.Launch(); err != nil {
		return r, err
	}
	if err = r.Actor.Launch(); err != nil {
		return r, err
	}
	t.Record("launcher", "root", 0, 0, "Codex and echo started")
	r.inputs = []*ports.Ingress{external.Ingress(), r.Actor.Endpoint().Ingress(), r.Echo.Endpoint().Ingress()}
	r.cancel, r.done = startLoop(ctx, r.Kernel, r.inputs)
	digest := sha256.Sum256(raw)
	connectCtx, stop := context.WithTimeout(ctx, 6*time.Second)
	defer stop()
	result, err := r.Call(connectCtx, 1, p.Args("connect", "configuration", "default", "descriptor", hex.EncodeToString(digest[:])))
	if err != nil {
		return r, err
	}
	if result.Return == nil || result.Return.Observation.Key != "connection" || len(result.Return.Observation.Bytes) != 64 {
		return r, fmt.Errorf("connect failed: %+v", result.Fault)
	}
	r.Connection = string(result.Return.Observation.Bytes)
	r.ready = true
	t.Record("root", "root", 1, result.Return.Answering, "connect validated; ready")
	return r, nil
}

// startLoop runs the Kernel loop until the returned cancel; done yields its error.
func startLoop(ctx context.Context, k *kernel.Kernel, inputs []*ports.Ingress) (context.CancelFunc, chan error) {
	loopCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- k.Loop(loopCtx, inputs) }()
	return cancel, done
}

func (r *CodexBoot) Call(ctx context.Context, id uint64, args []p.Observation) (p.Move, error) {
	if err := r.Client.Egress().Send(ctx, p.Move{Identity: id, Target: p.Codex, Invoke: &p.Invoke{Args: args}}); err != nil {
		return p.Move{}, err
	}
	return r.Results.Ingress().Receive(ctx)
}
func (r *CodexBoot) Close() {
	if r.closed {
		return
	}
	r.closed = true
	r.ready = false
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
	r.Actor.Reap(r.Kernel.Trace(), "Codex reaped (provider supervised by Codex)")
	if r.Echo != nil {
		_ = r.Echo.Close()
	}
	rt.Memory.Unregister(p.Codex)
	rt.Memory.Forget("root")
}
