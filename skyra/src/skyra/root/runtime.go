package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	p "skyra-v1/src/skyra/protocol"
)

func newKernel(trace *Trace) *Kernel {
	return &Kernel{runtimeState: &RuntimeState{
		memory: &Memory{executables: make(map[p.ExecutableID]Executable), contexts: make(map[p.ContextID]*ExecutionContext)},
		active: &Active{contexts: make(map[p.ContextID]*ExecutionContext), reservations: make(map[uint64]p.ContextID), trace: trace}},
		scheduler: &Scheduler{realizing: make(map[p.ContextID]bool), changed: make(chan struct{}), trace: trace}, trace: trace, attachments: make(chan *Ingress)}
}

func (m *Memory) register(executable Executable) {
	m.mu.Lock()
	defer m.mu.Unlock()
	executable.Contracts = append([]ArgContract(nil), executable.Contracts...)
	m.executables[executable.ID] = executable
}
func (m *Memory) retain(e *ExecutionContext) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.contexts[e.identity] = e
}

func (a *Active) provision(ctx context.Context, e *ExecutionContext, ports []uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := a.contexts[e.identity]; ok {
		return fmt.Errorf("context already active")
	}
	seen := make(map[uint64]bool)
	for _, port := range ports {
		if _, ok := a.reservations[port]; ok || seen[port] {
			return fmt.Errorf("actual port %d already reserved", port)
		}
		seen[port] = true
	}
	if len(seen) != len(e.horizon) {
		return fmt.Errorf("reservations must cover mappings exactly")
	}
	mapped := make(map[uint64]bool)
	for _, mapping := range e.horizon {
		b := mapping.endpoint.binding
		if !seen[mapping.actual] || mapped[mapping.actual] || b.actual != mapping.actual || b.owner != e.identity || !b.live.Load() {
			return fmt.Errorf("invalid mapping reservation")
		}
		mapped[mapping.actual] = true
	}
	for _, port := range ports {
		a.reservations[port] = e.identity
	}
	e.live = true
	a.contexts[e.identity] = e
	a.trace.record("active", e.identity, 0, 0, fmt.Sprintf("context published and ports reserved: %v", ports))
	return nil
}
func (a *Active) ownerOf(port uint64) (p.ContextID, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	id, ok := a.reservations[port]
	return id, ok
}
func (a *Active) lookup(id p.ContextID) (*ExecutionContext, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.contexts[id]
	return e, ok
}
func (a *Active) retire(id p.ContextID) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if e, ok := a.contexts[id]; ok {
		e.invalidate()
		delete(a.contexts, id)
		for port, owner := range a.reservations {
			if owner == id {
				delete(a.reservations, port)
			}
		}
		a.trace.record("active", id, 0, 0, "context retired and reservations released")
	}
}

func (s *Scheduler) signal() { close(s.changed); s.changed = make(chan struct{}) }
func (s *Scheduler) Enqueue(m queuedMove) error {
	if m.move.Context == "" {
		return fmt.Errorf("missing admitted owner")
	}
	m.move = m.move.Clone()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, m)
	s.trace.record("scheduler", m.move.Context, m.move.Identity, 0, "move enqueued")
	s.signal()
	return nil
}
func (s *Scheduler) Run(ctx context.Context, dispatch DispatchMove) error {
	for {
		s.mu.Lock()
		if err := ctx.Err(); err != nil {
			s.mu.Unlock()
			return err
		}
		for i, work := range s.queue {
			id := work.move.Context
			if s.realizing[id] {
				continue
			}
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			s.realizing[id] = true
			s.trace.record("scheduler", id, work.move.Identity, 0, "context claimed")
			s.mu.Unlock()
			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.realizing, id)
				s.trace.record("scheduler", id, work.move.Identity, 0, "claim released")
				s.signal()
			}()
			return dispatch(ctx, work)
		}
		changed := s.changed
		s.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

func (s *Scheduler) waitIdle(ctx context.Context, id p.ContextID) error {
	for {
		s.mu.Lock()
		busy := s.realizing[id]
		for _, work := range s.queue {
			if work.move.Context == id {
				busy = true
			}
		}
		changed := s.changed
		s.mu.Unlock()
		if !busy {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

func (k *Kernel) admit(ctx context.Context, source *Ingress, move p.Move) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b := source.endpoint.binding
	id, ok := k.runtimeState.active.ownerOf(b.actual)
	if !ok || id != b.owner {
		return fmt.Errorf("unreserved ingress")
	}
	e, ok := k.runtimeState.active.lookup(id)
	if !ok {
		return fmt.Errorf("inactive ingress")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if move.Context != "" && move.Context != id {
		return fmt.Errorf("forged context claim")
	}
	move.Context = id
	work := queuedMove{move: move.Clone(), source: b}
	if err := e.validateSource(work); err != nil {
		return err
	}
	k.trace.record("kernel", id, move.Identity, move.Source, variant(move)+" admitted")
	return k.scheduler.Enqueue(work)
}
func (k *Kernel) dispatch(ctx context.Context, move queuedMove) error {
	e, ok := k.runtimeState.active.lookup(move.move.Context)
	if !ok {
		return fmt.Errorf("missing/inactive context")
	}
	err := e.Realize(ctx, move)
	var rejection *moveRejection
	if errors.As(err, &rejection) {
		k.trace.record("kernel", move.move.Context, move.move.Identity, move.move.Source, "dispatch rejected: "+err.Error())
	}
	return err
}

type moveRejection struct{ error }

// loop's input endpoints are supplied by root wiring, never copied into Scheduler.
func (k *Kernel) loop(ctx context.Context, inputs []*Ingress) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var workers sync.WaitGroup
	failures := make(chan error, 1)
	report := func(err error) {
		select {
		case failures <- err:
		default:
		}
	}
	startReader := func(input *Ingress) {
		workers.Add(1)
		go func(in *Ingress) {
			defer workers.Done()
			for {
				m, err := in.Receive(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					var rejection *moveRejection
					if errors.As(err, &rejection) {
						k.trace.record("kernel", in.endpoint.binding.owner, 0, 0, "ingress rejected: "+err.Error())
						continue
					}
					// An intentional replacement closes its old reader. That
					// is not loss of the context's current transport.
					if k.superseded(in.endpoint.binding) {
						k.trace.record("kernel", in.endpoint.binding.owner, 0, 0, "superseded ingress closed")
						return
					}
					report(err)
					return
				}
				k.trace.record("port", in.endpoint.binding.owner, m.Identity, m.Source, variant(m)+" received")
				if err := k.admit(ctx, in, m); err != nil {
					k.trace.record("kernel", in.endpoint.binding.owner, m.Identity, m.Source, "rejected: "+err.Error())
				}
			}
		}(input)
	}
	for _, input := range inputs {
		startReader(input)
	}
	for i := 0; i < 2; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				if err := k.scheduler.Run(ctx, k.dispatch); err != nil {
					if ctx.Err() != nil {
						return
					}
					var rejected *moveRejection
					if errors.As(err, &rejected) {
						continue
					}
					report(err)
					return
				}
			}
		}()
	}
	var err error
running:
	for {
		select {
		case in := <-k.attachments:
			startReader(in)
		case <-ctx.Done():
			err = ctx.Err()
			break running
		case err = <-failures:
			break running
		}
	}
	cancel()
	workers.Wait()
	// Shutdown invalidates all remaining endpoints and releases Active reservations.
	k.runtimeState.active.mu.RLock()
	ids := make([]p.ContextID, 0, len(k.runtimeState.active.contexts))
	for id := range k.runtimeState.active.contexts {
		ids = append(ids, id)
	}
	k.runtimeState.active.mu.RUnlock()
	for _, id := range ids {
		k.runtimeState.active.retire(id)
	}
	return err
}

func variant(m p.Move) string {
	if m.Invoke != nil {
		return "invocation"
	}
	if m.Return != nil {
		return "return"
	}
	return "fault"
}

func (k *Kernel) superseded(b *binding) bool {
	e, ok := k.runtimeState.active.lookup(b.owner)
	if !ok {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.live || b.live.Load() {
		return false
	}
	for _, m := range e.horizon {
		if m.endpoint.binding == b {
			return false
		}
	}
	return true
}
