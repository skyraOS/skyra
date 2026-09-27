package execution

import (
	"context"
	"errors"
	"fmt"
	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/ports"
	"skyra/history/dna19/root/trace"
	"strings"
)

func New(id p.ContextID, t *trace.Trace, sink *ports.Egress) *ExecutionContext {
	return &ExecutionContext{identity: id, executable: Executable{ID: 100, Contracts: []ArgContract{{Operation: "echo"}, {Operation: "parent"}}},
		horizon: make(map[uint64]*mapping), bookkeeper: &Bookkeeper{frames: make(map[p.FrameID]*frame), requests: make(map[uint64]bool), rootSink: 0, bottom: &frame{id: 0, phase: "sink"}}, rootEgress: sink, trace: t}
}

// Identity is fixed at construction and needs no lock.
func (e *ExecutionContext) Identity() p.ContextID { return e.identity }

// Activation exposes a detached declaration; bound activation peers route this ID.
func (e *ExecutionContext) Activation() Executable {
	e.mu.Lock()
	defer e.mu.Unlock()
	descriptor := e.executable
	descriptor.Contracts = append([]ArgContract(nil), descriptor.Contracts...)
	return descriptor
}

// Declare replaces the activation declaration before publication.
func (e *ExecutionContext) Declare(executable Executable) {
	e.mu.Lock()
	defer e.mu.Unlock()
	executable.Contracts = append([]ArgContract(nil), executable.Contracts...)
	e.executable = executable
}

// Install prepares authority before publication. Live replacements use Rebind.
func (e *ExecutionContext) Install(virtual uint64, endpoint *ports.Endpoint, executable Executable) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.live {
		return fmt.Errorf("cannot install into published context")
	}
	if _, ok := e.horizon[virtual]; ok {
		return fmt.Errorf("virtual port already installed")
	}
	b := endpoint.Binding()
	if b.Owner() != e.identity || executable.ID != b.Target() {
		return fmt.Errorf("mapping authority mismatch")
	}
	executable.Contracts = append([]ArgContract(nil), executable.Contracts...)
	e.horizon[virtual] = &mapping{actual: b.Actual(), executable: executable, endpoint: endpoint}
	e.trace.Record("context", e.identity, 0, 0, fmt.Sprintf("local mapping installed: virtual=%d actual=%d executable=%d generation=%d", virtual, b.Actual(), executable.ID, b.Generation()))
	return nil
}

// Publish checks that reserved covers the Horizon exactly, then marks the
// context live. commit is Active's reservation write; it runs under this
// context's lock so both owners publish in one step.
func (e *ExecutionContext) Publish(reserved map[uint64]bool, commit func()) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(reserved) != len(e.horizon) {
		return fmt.Errorf("reservations must cover mappings exactly")
	}
	mapped := make(map[uint64]bool)
	for _, mapping := range e.horizon {
		b := mapping.endpoint.Binding()
		if !reserved[mapping.actual] || mapped[mapping.actual] || b.Actual() != mapping.actual || b.Owner() != e.identity || !b.Live() {
			return fmt.Errorf("invalid mapping reservation")
		}
		mapped[mapping.actual] = true
	}
	commit()
	e.live = true
	return nil
}

// Extend installs an address into a live context. check runs Active's authority
// and reservation tests, and commit its reservation write, under this context's
// lock so the mapping and the reservation change together.
func (e *ExecutionContext) Extend(virtual uint64, ep *ports.Endpoint, definition Executable, check func(Horizon) error, commit func()) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.live {
		return fmt.Errorf("inactive creator")
	}
	if err := check(heldHorizon{e}); err != nil {
		return err
	}
	if _, exists := e.horizon[virtual]; exists {
		return fmt.Errorf("virtual port installed")
	}
	b := ep.Binding()
	if b.Owner() != e.identity || b.Target() != definition.ID || !b.Live() {
		return fmt.Errorf("extension authority mismatch")
	}
	for _, m := range e.horizon {
		if definition.ID != 0 && m.executable.ID == definition.ID {
			return fmt.Errorf("address already installed")
		}
	}
	definition.Contracts = append([]ArgContract(nil), definition.Contracts...)
	commit()
	e.horizon[virtual] = &mapping{actual: b.Actual(), executable: definition, endpoint: ep}
	return nil
}

func (e *ExecutionContext) Rebind(virtual uint64, replacement *ports.Endpoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, ok := e.horizon[virtual]
	if !e.live || !ok {
		return fmt.Errorf("inactive mapping")
	}
	old, next := m.endpoint.Binding(), replacement.Binding()
	if next.Owner() != e.identity || next.Actual() != m.actual || next.Target() != m.executable.ID || next.Generation() <= old.Generation() || !next.Live() {
		return fmt.Errorf("invalid replacement")
	}
	_ = m.endpoint.Close()
	m.endpoint = replacement
	e.trace.Record("context", e.identity, 0, 0, "binding invalidated and replaced")
	return nil
}
func (e *ExecutionContext) Invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	alreadyClosed := !e.live
	e.live = false
	for _, c := range e.bookkeeper.parents {
		c.live = false
	}
	for _, f := range e.bookkeeper.frames {
		if f.phase == "running" || f.phase == "waiting" {
			f.phase = "faulted"
		}
	}
	for _, m := range e.horizon {
		if m.endpoint.Binding().Live() {
			alreadyClosed = false
		}
		_ = m.endpoint.Close()
	}
	if !alreadyClosed {
		e.trace.Record("context", e.identity, 0, 0, "bindings invalidated")
	}
}

// Superseded reports a binding intentionally replaced in a still-live context.
// Closing its reader is not loss of the context's current transport.
func (e *ExecutionContext) Superseded(b *ports.Binding) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.live || b.Live() {
		return false
	}
	for _, m := range e.horizon {
		if m.endpoint.Binding() == b {
			return false
		}
	}
	return true
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
	if result == nil || !result.endpoint.Binding().Live() {
		return nil, fmt.Errorf("executable outside live Horizon")
	}
	return result, nil
}

// Horizon is a context's address space, viewed while its lock is held.
type Horizon interface {
	Resolve(p.ExecutableID) (Executable, error)
}
type heldHorizon struct{ e *ExecutionContext }

func (h heldHorizon) Resolve(target p.ExecutableID) (Executable, error) {
	m, err := h.e.resolve(target)
	if err != nil {
		return Executable{}, err
	}
	x := m.executable
	x.Contracts = append([]ArgContract(nil), x.Contracts...)
	return x, nil
}

// Inspect runs fn against this context's Horizon under its lock.
func (e *ExecutionContext) Inspect(fn func(Horizon) error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return fn(heldHorizon{e})
}

// Admit validates work against this context's live bindings and frames. accept
// runs under the context lock, so no transition interleaves with the enqueue.
func (e *ExecutionContext) Admit(work Work, accept func() error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.validateSource(work); err != nil {
		return err
	}
	return accept()
}

func (e *ExecutionContext) validateSource(work Work) error {
	m, b := work.Move, work.Source
	if err := m.Validate(); err != nil {
		return err
	}
	if !e.live || b == nil || !b.Live() || b.Owner() != e.identity || m.Context != e.identity {
		return fmt.Errorf("inactive or foreign binding")
	}
	found := false
	for _, entry := range e.horizon {
		if entry.endpoint.Binding() == b {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("superseded binding")
	}
	if b.External() {
		if m.Invoke == nil || m.Source != 0 || m.Invoke.Frame != 0 || m.Invoke.ReturnTo != 0 || len(m.State) != 0 {
			return fmt.Errorf("external endpoint only admits initial Invoke")
		}
		return nil
	}
	f := e.bookkeeper.frames[m.Source]
	if f == nil || f.binding != b || (f.phase != "running" && !(m.Fault != nil && f.phase == "waiting")) {
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

func (b *Bookkeeper) traceDependencies(work Work) error {
	m := work.Move
	if m.Invoke != nil {
		if work.Source.External() {
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
		if f == nil || (f.phase != "running" && !(m.Fault != nil && f.phase == "waiting")) {
			return fmt.Errorf("duplicate or premature completion")
		}
		if f.parent != 0 {
			parent := b.frames[f.parent]
			if parent == nil || parent.phase != "waiting" || parent.waiting != f.id || !parent.binding.Live() {
				return fmt.Errorf("stale parent continuation")
			}
		}
	}
	return nil
}

func (b *Bookkeeper) record(m p.Move, binding *ports.Binding, external bool) *frame {
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
func (e *ExecutionContext) Realize(ctx context.Context, work Work) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	out, destination, frameID, err := e.commit(work)
	e.mu.Unlock()
	if err != nil {
		var fatal *runtimeFailure
		if errors.As(err, &fatal) {
			return err
		}
		return ports.Reject(err)
	}
	if err := destination.Send(ctx, out); err != nil {
		return err
	}
	transition := "invocation sent"
	if out.Return != nil || out.Fault != nil {
		transition = "parent resumed"
		if destination == e.rootEgress {
			transition = "root result sent"
		}
	}
	e.trace.Record("port", e.identity, out.Identity, frameID, transition)
	return nil
}

func (e *ExecutionContext) commit(work Work) (p.Move, *ports.Egress, p.FrameID, error) {
	fail := func(err error) (p.Move, *ports.Egress, p.FrameID, error) { return p.Move{}, nil, 0, err }
	if err := e.validateSource(work); err != nil {
		return fail(err)
	}
	if err := e.bookkeeper.traceDependencies(work); err != nil {
		return fail(err)
	}
	m := work.Move
	b := e.bookkeeper
	var out p.Move
	var destination *ports.Egress
	var frameID p.FrameID
	if m.Invoke != nil {
		entry, err := e.resolve(m.Target)
		if err != nil {
			return fail(err)
		}
		valid := false
		for _, contract := range entry.executable.Contracts {
			if contract.check(m.Invoke.Args) == nil {
				valid = true
				break
			}
		}
		if !valid {
			return fail(fmt.Errorf("argument contract rejected"))
		}
		var selected ArgContract
		for _, c := range entry.executable.Contracts {
			if c.check(m.Invoke.Args) == nil {
				selected = c
				break
			}
		}
		handles, err := e.authorizeHandles(selected.Profile, m.Invoke.Args, entry.endpoint.Binding())
		if err != nil {
			return fail(err)
		}
		f := b.record(m, entry.endpoint.Binding(), work.Source.External())
		f.profile, f.operation, f.handles = selected.Profile, selected.Operation, handles
		frameID = f.id
		out = m.Clone()
		out.Source, out.State = 0, nil
		out.Invoke.Frame, out.Invoke.ReturnTo = f.id, f.parent
		destination = entry.endpoint.Egress()
		e.trace.Record("bookkeeper", e.identity, m.Identity, f.id, "frame recorded")
		if f.parent != 0 {
			e.trace.Record("bookkeeper", e.identity, m.Identity, f.parent, fmt.Sprintf("parent yielded waiting for child %d", f.id))
		}
	} else {
		f := b.frames[m.Source]
		// Check the declared result before settling anything.
		if m.Fault != nil && f.profile != nil {
			valid := false
			for _, code := range f.profile.Failures {
				if strings.HasPrefix(m.Fault.Message, code+":") {
					valid = true
				}
			}
			if !valid {
				return fail(fmt.Errorf("fault contract rejected"))
			}
		}
		if m.Return != nil {
			for _, entry := range e.horizon {
				if entry.endpoint.Binding() == f.binding {
					key := entry.executable.OutputContract.Key
					if key == "" {
						key = "text"
					}
					allowed := m.Return.Observation.Key == key
					if f.profile != nil {
						allowed = false
						for _, k := range f.profile.Results {
							if k == m.Return.Observation.Key {
								allowed = true
							}
						}
					}
					if !allowed {
						return fail(fmt.Errorf("output contract rejected"))
					}
				}
			}
		}
		if f.parent == 0 && (b.bottom == nil || b.bottom.id != b.rootSink || b.bottom.phase != "sink" || e.rootEgress == nil) {
			return fail(&runtimeFailure{fmt.Errorf("missing root sink destination")})
		}
		frameID = f.id
		if m.Fault != nil {
			b.invalidateDescendants(f.id)
		}
		b.settleHandles(f, m.Fault)
		f.phase = "returned"
		if m.Return != nil {
			f.result = m.Clone().Return.Observation
		} else {
			f.phase = "faulted"
		}
		e.trace.Record("bookkeeper", e.identity, m.Identity, f.id, "frame settled")
		out = m.Clone()
		out.Target, out.State = 0, nil
		if f.parent == 0 {
			out.Identity = f.request
			destination = e.rootEgress
			e.trace.Record("bookkeeper", e.identity, m.Identity, b.rootSink, "root sink resolved")
		} else {
			parent := b.frames[f.parent]
			parent.phase, parent.waiting = "running", 0
			out.Source, out.Target, out.State = parent.id, parent.target, append([]byte(nil), parent.state...)
			for _, entry := range e.horizon {
				if entry.endpoint.Binding() == parent.binding {
					destination = entry.endpoint.Egress()
					break
				}
			}
		}
	}
	e.trace.Record("context", e.identity, m.Identity, frameID, "transition committed")
	return out, destination, frameID, nil
}
