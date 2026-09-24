package boot

import (
	"context"
	"fmt"
	"io"
	"net"
	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
	"time"
)

// UserConfig launches the external `user` executable and an agent for it to
// drive. For this first proof the agent is `echo`, so the whole
// browser -> user -> root -> echo -> Return -> user loop runs with no provider.
type UserConfig struct {
	Binary string // the user executable
	Echo   string // the agent executable (echo for this proof)
	Addr   string // browser listen address; "" picks an ephemeral loopback port
	Stderr io.Writer
}

// UserBoot wires the user binary in exactly as CodexBoot wires Codex: a bound
// stdio Move port for the external peer, echo installed as a host child in its
// Horizon, an external caller endpoint and a root sink. Root does not change
// any Move ownership boundary; user imports no root package and vice versa.
type UserBoot struct {
	Kernel          *kernel.Kernel
	Execution       *execution.ExecutionContext
	Actor           *ports.Helper
	Echo            *ports.Helper
	Client, Results *ports.Endpoint
	inputs          []*ports.Ingress
	cancel          context.CancelFunc
	done            chan error
	InteractID      uint64
	Addr            string
	closed          bool
}

func User(ctx context.Context, c UserConfig, t *trace.Trace) (r *UserBoot, err error) {
	r = &UserBoot{Kernel: kernel.New(t)}
	defer func() {
		if err != nil {
			r.Close()
		}
	}()

	if r.Addr, err = listenAddr(c.Addr); err != nil {
		return r, err
	}
	r.Actor, err = ports.PreparePeer(c.Binary, nil, []string{"SKYRA_USER_ADDR=" + r.Addr}, c.Stderr, "root", 20, p.User)
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
	userDef := execution.Executable{ID: p.User, Contracts: []execution.ArgContract{{Operation: "interact"}}}
	echoDef := execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}
	rt := r.Kernel.State()
	rt.Memory.Register(userDef)
	rt.Memory.Register(echoDef)
	for i, item := range []struct {
		ep  *ports.Endpoint
		def execution.Executable
	}{{r.Actor.Endpoint(), userDef}, {r.Echo.Endpoint(), echoDef}, {external, execution.Executable{}}, {sink, execution.Executable{}}} {
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
	t.Record("launcher", "root", 0, 0, "user and echo started")

	// Wait for the browser listener before returning, so a caller can connect
	// without racing the child process's startup.
	if err = awaitListener(ctx, r.Addr); err != nil {
		return r, err
	}

	r.inputs = []*ports.Ingress{external.Ingress(), r.Actor.Endpoint().Ingress(), r.Echo.Endpoint().Ingress()}
	r.cancel, r.done = startLoop(ctx, r.Kernel, r.inputs)

	// The `interact` activation opens the persistent interaction U. It is a
	// normal external Invoke; U stays alive until an explicit close returns it.
	r.InteractID = 1
	if err = r.Client.Egress().Send(ctx, p.Move{Identity: r.InteractID, Target: p.User, Invoke: &p.Invoke{Args: p.Arguments("interact", "")}}); err != nil {
		return r, err
	}
	t.Record("root", "root", r.InteractID, 0, "interact sent; U opened")
	return r, nil
}

// listenAddr returns addr, or an ephemeral loopback address when it is empty.
func listenAddr(addr string) (string, error) {
	if addr != "" {
		return addr, nil
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	addr = l.Addr().String()
	_ = l.Close()
	return addr, nil
}

// awaitListener waits up to three seconds for the user's browser listener.
func awaitListener(ctx context.Context, addr string) error {
	probeCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	defer stop()
	for {
		if conn, e := net.Dial("tcp", addr); e == nil {
			_ = conn.Close()
			return nil
		}
		if probeCtx.Err() != nil {
			return fmt.Errorf("user browser listener never came up on %s", addr)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// AwaitClose receives U's single terminal Return (interaction-closed) from the
// root sink, correlated to the original interact Identity.
func (r *UserBoot) AwaitClose(ctx context.Context) (p.Move, error) {
	return r.Results.Ingress().Receive(ctx)
}

func (r *UserBoot) Close() {
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
	r.Actor.Reap(r.Kernel.Trace(), "user reaped")
	if r.Echo != nil {
		_ = r.Echo.Close()
	}
	rt.Memory.Unregister(p.User)
	rt.Memory.Forget("root")
}
