// Package extension owns runtime growth: delegating a grant to a new context,
// bridging to it, and minting a new address through a bounded creation
// capability. It changes authority only through state.Active and the Kernel.
package extension

import (
	"context"
	"fmt"
	"net"
	"sync"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/creation"
	"skyra-v1/src/skyra/root/crosscontext"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/kernel"
	"skyra-v1/src/skyra/root/ports"
)

// GrantPolicy names authority that the launcher permits this context to grant.
// The bounded fixture permits only independent echo behavior, never a general
// host command or a copy of the creator's entire horizon.
type GrantPolicy struct{ Echo bool }

type Extension struct {
	kernel    *kernel.Kernel
	owner     *execution.ExecutionContext
	policy    GrantPolicy
	mu        sync.Mutex
	minted    bool
	peers     sync.WaitGroup
	endpoints []*ports.Endpoint
	// Inputs are the new ingress the caller's loop must read.
	Inputs []*ports.Ingress
	// Child is the delegated context B.
	Child *execution.ExecutionContext
}

func (policy GrantPolicy) authorize(h execution.Horizon, template p.ExecutableID) error {
	if !policy.Echo || template != p.Echo {
		return fmt.Errorf("authority not grantable")
	}
	entry, err := h.Resolve(template)
	if err != nil {
		return err
	}
	if len(entry.Contracts) != 1 || entry.Contracts[0].Operation != "echo" {
		return fmt.Errorf("incompatible grant template")
	}
	return nil
}

func (x *Extension) peer(fn func() error) {
	x.peers.Add(1)
	go func() { defer x.peers.Done(); _ = fn() }()
}

// Minted reports whether the one-address creation budget has been spent.
func (x *Extension) Minted() bool {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.minted
}
func (x *Extension) Close() {
	// Called only after cancellation and loop completion, hence no mint in flight.
	if x.Child != nil {
		x.kernel.State().Active.Retire(x.Child.Identity())
	}
	for _, ep := range x.endpoints {
		_ = ep.Close()
	}
	x.peers.Wait()
}

// Setup delegates echo to B, exposes B's activation executable to A,
// and installs one creation capability. It also works when A is the root context:
// there is no root identity special case in activation or delegation.
func Setup(ctx context.Context, k *kernel.Kernel, owner *execution.ExecutionContext, policy GrantPolicy) (_ *Extension, err error) {
	x := &Extension{kernel: k, owner: owner, policy: policy}
	defer func() {
		if err != nil {
			x.Close()
		}
	}()
	if err = owner.Inspect(func(h execution.Horizon) error { return policy.authorize(h, p.Echo) }); err != nil {
		return nil, err
	}
	rt := k.State()
	echo, echoPeer := ports.Pair("B", 30, p.Echo, false)
	external, activation := ports.Pair("B", 31, 0, true)
	sink, results := ports.Pair("B", 32, 0, false)
	b := execution.New("B", k.Trace(), sink.Egress())
	b.Declare(execution.Executable{ID: 101, Contracts: []execution.ArgContract{{Operation: "echo"}}, OutputContract: execution.OutputContract{Key: "text"}})
	x.Child = b
	x.endpoints = append(x.endpoints, echo, echoPeer, external, activation, sink, results)
	for i, entry := range []struct {
		ep  *ports.Endpoint
		def execution.Executable
	}{{echo, execution.Executable{ID: p.Echo, Contracts: []execution.ArgContract{{Operation: "echo"}}}}, {external, execution.Executable{}}, {sink, execution.Executable{}}} {
		if err = b.Install(uint64(i), entry.ep, entry.def); err != nil {
			return nil, err
		}
	}
	if err = rt.Active.Provision(ctx, b, []uint64{30, 31, 32}); err != nil {
		return nil, err
	}
	rt.Memory.Retain(b)
	rt.Memory.Register(b.Activation())
	bridge, bridgePeer := ports.Pair(owner.Identity(), 14, b.Activation().ID, false)
	cap, capPeer := ports.Pair(owner.Identity(), 15, creation.Capability, false)
	x.endpoints = append(x.endpoints, bridge, bridgePeer, cap, capPeer)
	if err = rt.Active.Extend(ctx, owner, 4, bridge, b.Activation()); err != nil {
		return nil, err
	}
	capability := execution.Executable{ID: creation.Capability, Contracts: []execution.ArgContract{{Operation: "create"}}, OutputContract: execution.OutputContract{Key: "executable"}}
	if err = rt.Active.Extend(ctx, owner, 5, cap, capability); err != nil {
		return nil, err
	}
	rt.Memory.Register(capability)
	x.Inputs = []*ports.Ingress{echo.Ingress(), external.Ingress(), bridge.Ingress(), cap.Ingress()}
	x.peer(func() error { return creation.ServeEcho(echoPeer.Conn(), p.Echo) })
	x.peer(func() error {
		return crosscontext.Serve(bridgePeer.Conn(), activation.Conn(), results.Conn(), b.Activation().ID)
	})
	x.peer(func() error {
		return creation.Serve(capPeer.Conn(), func(template p.ExecutableID) (p.ExecutableID, error) { return x.mint(ctx, template) })
	})
	return x, nil
}

func (x *Extension) mint(ctx context.Context, template p.ExecutableID) (p.ExecutableID, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.minted {
		return 0, fmt.Errorf("creation budget exhausted")
	}
	authorized := func(h execution.Horizon) error {
		if _, err := h.Resolve(creation.Capability); err != nil {
			return err
		}
		return x.policy.authorize(h, template)
	}
	if err := x.owner.Inspect(authorized); err != nil {
		return 0, err
	}
	// One newly allocated address in this bounded experiment. Memory arbitrates
	// global identity uniqueness; no caller chooses this address or reservation.
	var peerConn net.Conn
	var ep *ports.Endpoint
	id, err := x.kernel.State().Memory.Mint(1000, func(id p.ExecutableID) (execution.Executable, error) {
		definition := execution.Executable{ID: id, Contracts: []execution.ArgContract{{Operation: "echo"}}, OutputContract: execution.OutputContract{Key: "text"}}
		a, b := net.Pipe()
		ep = ports.Bind(a, x.owner.Identity(), 40, id, 1, false)
		if err := x.kernel.State().Active.Extend(ctx, x.owner, 6, ep, definition, authorized); err != nil {
			_ = ep.Close()
			_ = b.Close()
			return definition, err
		}
		peerConn = b
		return definition, nil
	})
	if err != nil {
		return 0, err
	}
	x.minted = true
	peer := ports.Bind(peerConn, x.owner.Identity(), 0, 0, 1, false)
	x.endpoints = append(x.endpoints, ep, peer)
	x.peer(func() error { return creation.ServeEcho(peer.Conn(), id) })
	if err := x.kernel.Attach(ctx, ep.Ingress()); err != nil {
		return 0, err
	}
	x.kernel.Trace().Record("creation", x.owner.Identity(), 0, 0, fmt.Sprintf("new executable minted: %d (echo only)", id))
	return id, nil
}
