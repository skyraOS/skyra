// Package state owns what exists: Memory is the catalog of registered
// executables and retained contexts; Active is the set of published contexts and
// the actual ports they reserve. Admission reads it; nothing here runs Moves.
package state

import (
	"context"
	"fmt"
	"sync"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

type Runtime struct {
	Memory *Memory
	Active *Active
}
type Memory struct {
	mu          sync.Mutex
	executables map[p.ExecutableID]execution.Executable
	contexts    map[p.ContextID]*execution.ExecutionContext
}
type Active struct {
	mu           sync.RWMutex
	contexts     map[p.ContextID]*execution.ExecutionContext
	reservations map[uint64]p.ContextID
	trace        *trace.Trace
}

func New(t *trace.Trace) *Runtime {
	return &Runtime{
		Memory: &Memory{executables: make(map[p.ExecutableID]execution.Executable), contexts: make(map[p.ContextID]*execution.ExecutionContext)},
		Active: &Active{contexts: make(map[p.ContextID]*execution.ExecutionContext), reservations: make(map[uint64]p.ContextID), trace: t}}
}

func (m *Memory) Register(executable execution.Executable) {
	m.mu.Lock()
	defer m.mu.Unlock()
	executable.Contracts = append([]execution.ArgContract(nil), executable.Contracts...)
	m.executables[executable.ID] = executable
}
func (m *Memory) Retain(e *execution.ExecutionContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contexts[e.Identity()] = e
}

// Unregister removes a registration and Forget a retained context, at shutdown.
func (m *Memory) Unregister(ids ...p.ExecutableID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		delete(m.executables, id)
	}
}
func (m *Memory) Forget(id p.ContextID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.contexts, id)
}
func (m *Memory) Registered(id p.ExecutableID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.executables[id]
	return ok
}
func (m *Memory) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.executables)
}

// Mint arbitrates global identity uniqueness: it picks the lowest unregistered
// address at or above start and registers what install returns for it. Memory
// stays locked throughout, so no caller chooses or races the address.
func (m *Memory) Mint(start p.ExecutableID, install func(p.ExecutableID) (execution.Executable, error)) (p.ExecutableID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := start
	for {
		if _, exists := m.executables[id]; !exists {
			break
		}
		id++
	}
	definition, err := install(id)
	if err != nil {
		return 0, err
	}
	m.executables[id] = definition
	return id, nil
}

func (a *Active) Provision(ctx context.Context, e *execution.ExecutionContext, ports []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	id := e.Identity()
	if _, ok := a.contexts[id]; ok {
		return fmt.Errorf("context already active")
	}
	seen := make(map[uint64]bool)
	for _, port := range ports {
		if _, ok := a.reservations[port]; ok || seen[port] {
			return fmt.Errorf("actual port %d already reserved", port)
		}
		seen[port] = true
	}
	if err := e.Publish(seen, func() {
		for _, port := range ports {
			a.reservations[port] = id
		}
		a.contexts[id] = e
	}); err != nil {
		return err
	}
	a.trace.Record("active", id, 0, 0, fmt.Sprintf("context published and ports reserved: %v", ports))
	return nil
}

// Extend coordinates Active's reservation write and the context's mapping write.
// Callers are trusted provisioning/creation plumbing, never wire programs.
func (a *Active) Extend(ctx context.Context, e *execution.ExecutionContext, virtual uint64, ep *ports.Endpoint, definition execution.Executable, authorize ...func(execution.Horizon) error) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	id := e.Identity()
	if a.contexts[id] != e {
		return fmt.Errorf("inactive creator")
	}
	actual := ep.Binding().Actual()
	if err := e.Extend(virtual, ep, definition, func(h execution.Horizon) error {
		for _, check := range authorize {
			if err := check(h); err != nil {
				return err
			}
		}
		if _, exists := a.reservations[actual]; exists {
			return fmt.Errorf("actual port reserved")
		}
		return nil
	}, func() { a.reservations[actual] = id }); err != nil {
		return err
	}
	a.trace.Record("active", id, 0, 0, fmt.Sprintf("address installed: executable=%d virtual=%d actual=%d", definition.ID, virtual, actual))
	return nil
}

func (a *Active) OwnerOf(port uint64) (p.ContextID, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	id, ok := a.reservations[port]
	return id, ok
}
func (a *Active) Lookup(id p.ContextID) (*execution.ExecutionContext, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.contexts[id]
	return e, ok
}
func (a *Active) Retire(id p.ContextID) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if e, ok := a.contexts[id]; ok {
		e.Invalidate()
		delete(a.contexts, id)
		for port, owner := range a.reservations {
			if owner == id {
				delete(a.reservations, port)
			}
		}
		a.trace.Record("active", id, 0, 0, "context retired and reservations released")
	}
}

// RetireAll invalidates every remaining context and releases its reservations.
func (a *Active) RetireAll() {
	a.mu.RLock()
	ids := make([]p.ContextID, 0, len(a.contexts))
	for id := range a.contexts {
		ids = append(ids, id)
	}
	a.mu.RUnlock()
	for _, id := range ids {
		a.Retire(id)
	}
}

// Size reports published contexts and reserved ports.
func (a *Active) Size() (contexts, reservations int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.contexts), len(a.reservations)
}
