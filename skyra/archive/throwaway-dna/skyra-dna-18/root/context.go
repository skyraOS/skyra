package main

import (
	"context"
	"fmt"
	p "skyra-v1/src/skyra/skyra-dna-18/protocol"
)

func newExecution(id p.ContextID, trace *Trace, sink *Egress) *ExecutionContext {
	return &ExecutionContext{identity: id, executable: Executable{ID: 100, Contracts: []ArgContract{{Operation: "echo"}, {Operation: "parent"}}},
		horizon: make(map[uint64]*mapping), bookkeeper: &Bookkeeper{frames: make(map[p.FrameID]*frame), requests: make(map[uint64]bool), rootSink: sink}, trace: trace}
}

// Activation exposes a detached executable declaration, never local authority.
// The bound external ingress is the fixture's activation endpoint; full outward
// activation envelopes and routing by this descriptor's ID remain open.
func (e *ExecutionContext) Activation() Executable {
	e.mu.Lock()
	defer e.mu.Unlock()
	descriptor := e.executable
	descriptor.Contracts = append([]ArgContract(nil), descriptor.Contracts...)
	return descriptor
}

// install prepares authority before publication. Live replacements use rebind.
func (e *ExecutionContext) install(virtual uint64, endpoint *endpoint, executable Executable) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.live {
		return fmt.Errorf("cannot install into published context")
	}
	if _, ok := e.horizon[virtual]; ok {
		return fmt.Errorf("virtual port already installed")
	}
	if endpoint.binding.owner != e.identity || executable.ID != endpoint.binding.target {
		return fmt.Errorf("mapping authority mismatch")
	}
	executable.Contracts = append([]ArgContract(nil), executable.Contracts...)
	e.horizon[virtual] = &mapping{actual: endpoint.binding.actual, executable: executable, endpoint: endpoint}
	e.trace.record("context", e.identity, 0, 0, fmt.Sprintf("local mapping installed: virtual=%d actual=%d executable=%d generation=%d", virtual, endpoint.binding.actual, executable.ID, endpoint.binding.generation))
	return nil
}
func (e *ExecutionContext) rebind(virtual uint64, replacement *endpoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.horizon[virtual]
	if !e.live || !ok {
		return fmt.Errorf("inactive mapping")
	}
	old, next := m.endpoint.binding, replacement.binding
	if next.owner != e.identity || next.actual != m.actual || next.target != m.executable.ID || next.generation <= old.generation || !next.live.Load() {
		return fmt.Errorf("invalid replacement")
	}
	_ = m.endpoint.Close()
	m.endpoint = replacement
	e.trace.record("context", e.identity, 0, 0, "binding invalidated and replaced")
	return nil
}
func (e *ExecutionContext) invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	alreadyClosed := !e.live
	e.live = false
	for _, m := range e.horizon {
		if m.endpoint.binding.live.Load() {
			alreadyClosed = false
		}
		_ = m.endpoint.Close()
	}
	if !alreadyClosed {
		e.trace.record("context", e.identity, 0, 0, "bindings invalidated")
	}
}
func (e *ExecutionContext) resolve(target p.ExecutableID) (*mapping, error) {
	var result *mapping
	for _, entry := range e.horizon {
		if target != 0 && entry.executable.ID == target {
			if result != nil {
				return nil, fmt.Errorf("ambiguous executable")
			}
			result = entry
		}
	}
	if result == nil || !result.endpoint.binding.live.Load() {
		return nil, fmt.Errorf("executable outside live Horizon")
	}
	return result, nil
}
func (e *ExecutionContext) validateSource(work queuedMove) error {
	m, b := work.move, work.source
	if err := m.Validate(); err != nil {
		return err
	}
	if !e.live || b == nil || !b.live.Load() || b.owner != e.identity || m.Context != e.identity {
		return fmt.Errorf("inactive or foreign binding")
	}
	found := false
	for _, entry := range e.horizon {
		if entry.endpoint.binding == b {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("superseded binding")
	}
	if b.external {
		if m.Invoke == nil || m.Source != 0 || m.Invoke.Frame != 0 || m.Invoke.ReturnTo != 0 || len(m.State) != 0 {
			return fmt.Errorf("external endpoint only admits initial Invoke")
		}
		return nil
	}
	f := e.bookkeeper.frames[m.Source]
	if f == nil || f.binding != b || f.phase != "running" {
		return fmt.Errorf("output does not name a live bound frame")
	}
	if m.Return != nil && m.Return.Answering != f.id {
		return fmt.Errorf("forged return correlation")
	}
	if m.Fault != nil && m.Fault.Frame != f.id {
		return fmt.Errorf("forged fault correlation")
	}
	if m.Invoke != nil && (m.Invoke.Frame != 0 || m.Invoke.ReturnTo != 0) {
		return fmt.Errorf("child identities belong to Bookkeeper")
	}
	return nil
}

func (b *Bookkeeper) traceDependencies(work queuedMove) error {
	m := work.move
	if m.Invoke != nil {
		if work.source.external {
			if b.requests[m.Identity] {
				return fmt.Errorf("duplicate activation")
			}
		} else {
			parent := b.frames[m.Source]
			if parent == nil || parent.phase != "running" || parent.waiting != 0 {
				return fmt.Errorf("one pending child allowed")
			}
		}
	} else {
		f := b.frames[m.Source]
		if f == nil || f.phase != "running" {
			return fmt.Errorf("duplicate or premature completion")
		}
		if f.parent != 0 {
			parent := b.frames[f.parent]
			if parent == nil || parent.phase != "waiting" || parent.waiting != f.id || !parent.binding.live.Load() {
				return fmt.Errorf("stale parent continuation")
			}
		}
	}
	return nil
}

func (b *Bookkeeper) record(m p.Move, binding *binding, external bool) *frame {
	b.next++
	f := &frame{id: b.next, request: m.Identity, target: m.Target, binding: binding, phase: "running"}
	if external {
		b.requests[m.Identity] = true
	} else {
		parent := b.frames[m.Source]
		f.parent = parent.id
		parent.waiting, parent.phase, parent.state = f.id, "waiting", append([]byte(nil), m.State...)
	}
	b.frames[f.id] = f
	return f
}

// Realize commits under the context lock, then sends through the committed bound
// destination. It never waits for a provider's completion or calls foreign methods.
func (e *ExecutionContext) Realize(ctx context.Context, work queuedMove) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	out, destination, frameID, err := e.commit(work)
	e.mu.Unlock()
	if err != nil {
		return err
	}
	if err := destination.Send(ctx, out); err != nil {
		return err
	}
	transition := "invocation sent"
	if out.Return != nil || out.Fault != nil {
		transition = "parent resumed"
		if destination == e.bookkeeper.rootSink {
			transition = "root result sent"
		}
	}
	e.trace.record("port", e.identity, out.Identity, frameID, transition)
	return nil
}

func (e *ExecutionContext) commit(work queuedMove) (p.Move, *Egress, p.FrameID, error) {
	fail := func(err error) (p.Move, *Egress, p.FrameID, error) { return p.Move{}, nil, 0, err }
	if err := e.validateSource(work); err != nil {
		return fail(err)
	}
	if err := e.bookkeeper.traceDependencies(work); err != nil {
		return fail(err)
	}
	m := work.move
	b := e.bookkeeper
	var out p.Move
	var destination *Egress
	var frameID p.FrameID
	if m.Invoke != nil {
		entry, err := e.resolve(m.Target)
		if err != nil {
			return fail(err)
		}
		valid := false
		for _, contract := range entry.executable.Contracts {
			if p.CheckArguments(m.Invoke.Args, contract.Operation) == nil {
				valid = true
				break
			}
		}
		if !valid {
			return fail(fmt.Errorf("argument contract rejected"))
		}
		f := b.record(m, entry.endpoint.binding, work.source.external)
		frameID = f.id
		out = m.Clone()
		out.Source, out.State = 0, nil
		out.Invoke.Frame, out.Invoke.ReturnTo = f.id, f.parent
		destination = entry.endpoint.egress
		e.trace.record("bookkeeper", e.identity, m.Identity, f.id, "frame recorded")
		if f.parent != 0 {
			e.trace.record("bookkeeper", e.identity, m.Identity, f.parent, fmt.Sprintf("parent yielded waiting for child %d", f.id))
		}
	} else {
		f := b.frames[m.Source]
		frameID = f.id
		f.phase = "returned"
		if m.Return != nil {
			f.result = m.Clone().Return.Observation
		} else {
			f.phase = "faulted"
		}
		e.trace.record("bookkeeper", e.identity, m.Identity, f.id, "frame settled")
		out = m.Clone()
		out.Target, out.State = 0, nil
		if f.parent == 0 {
			out.Identity = f.request
			destination = b.rootSink
		} else {
			parent := b.frames[f.parent]
			parent.phase, parent.waiting = "running", 0
			out.Source, out.Target, out.State = parent.id, parent.target, append([]byte(nil), parent.state...)
			for _, entry := range e.horizon {
				if entry.endpoint.binding == parent.binding {
					destination = entry.endpoint.egress
					break
				}
			}
		}
	}
	e.trace.record("context", e.identity, m.Identity, frameID, "transition committed")
	return out, destination, frameID, nil
}
