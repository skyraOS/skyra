package main

import (
	"context"
	"fmt"
	"net"
	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/creation"
	"skyra-v1/src/skyra/root/crosscontext"
)

// grantPolicy names authority that the launcher permits this context to grant.
// The bounded fixture permits only independent echo behavior, never a general
// host command or a copy of the creator's entire horizon.

func (policy grantPolicy) authorize(e *ExecutionContext, template p.ExecutableID) error {
	if !policy.echo || template != p.Echo {
		return fmt.Errorf("authority not grantable")
	}
	entry, err := e.resolve(template)
	if err != nil {
		return err
	}
	if len(entry.executable.Contracts) != 1 || entry.executable.Contracts[0].Operation != "echo" {
		return fmt.Errorf("incompatible grant template")
	}
	return nil
}

// extend coordinates Active's reservation write and the context's mapping write.
// Callers are trusted provisioning/creation plumbing, never wire programs.
func (a *Active) extend(ctx context.Context, e *ExecutionContext, virtual uint64, ep *endpoint, definition Executable, authorize ...func() error) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if a.contexts[e.identity] != e || !e.live {
		return fmt.Errorf("inactive creator")
	}
	for _, check := range authorize {
		if err := check(); err != nil {
			return err
		}
	}
	if _, exists := a.reservations[ep.binding.actual]; exists {
		return fmt.Errorf("actual port reserved")
	}
	if _, exists := e.horizon[virtual]; exists {
		return fmt.Errorf("virtual port installed")
	}
	if ep.binding.owner != e.identity || ep.binding.target != definition.ID || !ep.binding.live.Load() {
		return fmt.Errorf("extension authority mismatch")
	}
	for _, m := range e.horizon {
		if definition.ID != 0 && m.executable.ID == definition.ID {
			return fmt.Errorf("address already installed")
		}
	}
	definition.Contracts = append([]ArgContract(nil), definition.Contracts...)
	a.reservations[ep.binding.actual] = e.identity
	e.horizon[virtual] = &mapping{actual: ep.binding.actual, executable: definition, endpoint: ep}
	a.trace.record("active", e.identity, 0, 0, fmt.Sprintf("address installed: executable=%d virtual=%d actual=%d", definition.ID, virtual, ep.binding.actual))
	return nil
}

func (x *extension) peer(fn func() error) {
	x.peers.Add(1)
	go func() { defer x.peers.Done(); _ = fn() }()
}
func (x *extension) close() {
	// Called only after cancellation and loop completion, hence no mint in flight.
	if x.child != nil {
		x.kernel.runtimeState.active.retire(x.child.identity)
	}
	for _, ep := range x.endpoints {
		_ = ep.Close()
	}
	x.peers.Wait()
}

// setupExtension delegates echo to B, exposes B's activation executable to A,
// and installs one creation capability. It also works when A is the root context:
// there is no root identity special case in activation or delegation.
func setupExtension(ctx context.Context, k *Kernel, owner *ExecutionContext, policy grantPolicy) (_ *extension, err error) {
	x := &extension{kernel: k, owner: owner, policy: policy}
	defer func() {
		if err != nil {
			x.close()
		}
	}()
	owner.mu.Lock()
	err = policy.authorize(owner, p.Echo)
	owner.mu.Unlock()
	if err != nil {
		return nil, err
	}
	echo, echoPeer := pair("B", 30, p.Echo, false)
	external, activation := pair("B", 31, 0, true)
	sink, results := pair("B", 32, 0, false)
	b := newExecution("B", k.trace, sink.egress)
	b.executable = Executable{ID: 101, Contracts: []ArgContract{{Operation: "echo"}}, OutputContract: OutputContract{Key: "text"}}
	x.child = b
	x.endpoints = append(x.endpoints, echo, echoPeer, external, activation, sink, results)
	for i, entry := range []struct {
		ep  *endpoint
		def Executable
	}{{echo, Executable{ID: p.Echo, Contracts: []ArgContract{{Operation: "echo"}}}}, {external, Executable{}}, {sink, Executable{}}} {
		if err = b.install(uint64(i), entry.ep, entry.def); err != nil {
			return nil, err
		}
	}
	if err = k.runtimeState.active.provision(ctx, b, []uint64{30, 31, 32}); err != nil {
		return nil, err
	}
	k.runtimeState.memory.retain(b)
	k.runtimeState.memory.register(b.Activation())
	bridge, bridgePeer := pair(owner.identity, 14, b.Activation().ID, false)
	cap, capPeer := pair(owner.identity, 15, creation.Capability, false)
	x.endpoints = append(x.endpoints, bridge, bridgePeer, cap, capPeer)
	if err = k.runtimeState.active.extend(ctx, owner, 4, bridge, b.Activation()); err != nil {
		return nil, err
	}
	capability := Executable{ID: creation.Capability, Contracts: []ArgContract{{Operation: "create"}}, OutputContract: OutputContract{Key: "executable"}}
	if err = k.runtimeState.active.extend(ctx, owner, 5, cap, capability); err != nil {
		return nil, err
	}
	k.runtimeState.memory.register(capability)
	x.inputs = []*Ingress{echo.ingress, external.ingress, bridge.ingress, cap.ingress}
	x.peer(func() error { return creation.ServeEcho(echoPeer.connection, p.Echo) })
	x.peer(func() error {
		return crosscontext.Serve(bridgePeer.connection, activation.connection, results.connection, b.Activation().ID)
	})
	x.peer(func() error {
		return creation.Serve(capPeer.connection, func(template p.ExecutableID) (p.ExecutableID, error) { return x.mint(ctx, template) })
	})
	return x, nil
}

func (x *extension) mint(ctx context.Context, template p.ExecutableID) (p.ExecutableID, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.minted {
		return 0, fmt.Errorf("creation budget exhausted")
	}
	x.owner.mu.Lock()
	_, err := x.owner.resolve(creation.Capability)
	if err == nil {
		err = x.policy.authorize(x.owner, template)
	}
	x.owner.mu.Unlock()
	if err != nil {
		return 0, err
	}
	// One newly allocated address in this bounded experiment. Memory arbitrates
	// global identity uniqueness; no caller chooses this address or reservation.
	memory := x.kernel.runtimeState.memory
	memory.mu.Lock()
	var id p.ExecutableID = 1000
	for {
		if _, exists := memory.executables[id]; !exists {
			break
		}
		id++
	}
	definition := Executable{ID: id, Contracts: []ArgContract{{Operation: "echo"}}, OutputContract: OutputContract{Key: "text"}}
	a, b := net.Pipe()
	ep := bind(a, x.owner.identity, 40, id, 1, false)
	if err = x.kernel.runtimeState.active.extend(ctx, x.owner, 6, ep, definition, func() error {
		if _, err := x.owner.resolve(creation.Capability); err != nil {
			return err
		}
		return x.policy.authorize(x.owner, template)
	}); err != nil {
		memory.mu.Unlock()
		_ = ep.Close()
		_ = b.Close()
		return 0, err
	}
	memory.executables[id] = definition
	memory.mu.Unlock()
	x.minted = true
	peer := bind(b, x.owner.identity, 0, 0, 1, false)
	x.endpoints = append(x.endpoints, ep, peer)
	x.peer(func() error { return creation.ServeEcho(peer.connection, id) })
	select {
	case x.kernel.attachments <- ep.ingress:
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	x.kernel.trace.record("creation", x.owner.identity, 0, 0, fmt.Sprintf("new executable minted: %d (echo only)", id))
	return id, nil
}
