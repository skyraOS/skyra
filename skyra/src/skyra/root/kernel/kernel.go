// Package kernel owns the main loop: it reads Moves from bound ingress, admits
// them against runtime state, queues them in the Scheduler, and dispatches each
// to its execution context, one turn per context at a time. Boot constructs and
// wires a Kernel; the Kernel never creates processes or authority.
package kernel

import (
	"context"
	"errors"
	"fmt"
	"sync"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/execution"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/state"
	"skyra-v1/src/skyra/root/trace"
)

type Kernel struct {
	runtimeState *state.Runtime
	scheduler    *Scheduler
	trace        *trace.Trace
	attachments  chan *ports.Ingress
}
type DispatchMove func(context.Context, execution.Work) error
type Scheduler struct {
	mu        sync.Mutex
	queue     []execution.Work
	realizing map[p.ContextID]bool
	changed   chan struct{}
	trace     *trace.Trace
}

func New(t *trace.Trace) *Kernel {
	return &Kernel{runtimeState: state.New(t),
		scheduler: &Scheduler{realizing: make(map[p.ContextID]bool), changed: make(chan struct{}), trace: t}, trace: t, attachments: make(chan *ports.Ingress)}
}

func (k *Kernel) State() *state.Runtime { return k.runtimeState }
func (k *Kernel) Scheduler() *Scheduler { return k.scheduler }
func (k *Kernel) Trace() *trace.Trace   { return k.trace }

// Attach hands a newly bound ingress to the running loop.
func (k *Kernel) Attach(ctx context.Context, in *ports.Ingress) error {
	select {
	case k.attachments <- in:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) signal() { close(s.changed); s.changed = make(chan struct{}) }
func (s *Scheduler) Enqueue(m execution.Work) error {
	if m.Move.Context == "" {
		return fmt.Errorf("missing admitted owner")
	}
	m.Move = m.Move.Clone()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, m)
	s.trace.Record("scheduler", m.Move.Context, m.Move.Identity, 0, "move enqueued")
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
			id := work.Move.Context
			if s.realizing[id] {
				continue
			}
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			s.realizing[id] = true
			s.trace.Record("scheduler", id, work.Move.Identity, 0, "context claimed")
			s.mu.Unlock()
			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.realizing, id)
				s.trace.Record("scheduler", id, work.Move.Identity, 0, "claim released")
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

func (s *Scheduler) WaitIdle(ctx context.Context, id p.ContextID) error {
	for {
		s.mu.Lock()
		busy := s.realizing[id]
		for _, work := range s.queue {
			if work.Move.Context == id {
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

// Claims and Queued report contexts mid-turn and Moves awaiting a turn.
func (s *Scheduler) Claims() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.realizing)
}
func (s *Scheduler) Queued() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}

func (k *Kernel) Admit(ctx context.Context, source *ports.Ingress, move p.Move) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b := source.Binding()
	id, ok := k.runtimeState.Active.OwnerOf(b.Actual())
	if !ok || id != b.Owner() {
		return fmt.Errorf("unreserved ingress")
	}
	e, ok := k.runtimeState.Active.Lookup(id)
	if !ok {
		return fmt.Errorf("inactive ingress")
	}
	if move.Context != "" && move.Context != id {
		return fmt.Errorf("forged context claim")
	}
	move.Context = id
	work := execution.Work{Move: move.Clone(), Source: b}
	return e.Admit(work, func() error {
		k.trace.Record("kernel", id, move.Identity, move.Source, variant(move)+" admitted")
		return k.scheduler.Enqueue(work)
	})
}
func (k *Kernel) Dispatch(ctx context.Context, move execution.Work) error {
	e, ok := k.runtimeState.Active.Lookup(move.Move.Context)
	if !ok {
		return fmt.Errorf("missing/inactive context")
	}
	err := e.Realize(ctx, move)
	var rejection *ports.Rejection
	if errors.As(err, &rejection) {
		k.trace.Record("kernel", move.Move.Context, move.Move.Identity, move.Move.Source, "dispatch rejected: "+err.Error())
	}
	return err
}

// Loop's input endpoints are supplied by root wiring, never copied into Scheduler.
func (k *Kernel) Loop(ctx context.Context, inputs []*ports.Ingress) error {
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
	startReader := func(input *ports.Ingress) {
		workers.Add(1)
		go func(in *ports.Ingress) {
			defer workers.Done()
			for {
				m, err := in.Receive(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					var rejection *ports.Rejection
					if errors.As(err, &rejection) {
						k.trace.Record("kernel", in.Binding().Owner(), 0, 0, "ingress rejected: "+err.Error())
						continue
					}
					// An intentional replacement closes its old reader. That
					// is not loss of the context's current transport.
					if k.superseded(in.Binding()) {
						k.trace.Record("kernel", in.Binding().Owner(), 0, 0, "superseded ingress closed")
						return
					}
					report(err)
					return
				}
				k.trace.Record("port", in.Binding().Owner(), m.Identity, m.Source, variant(m)+" received")
				if err := k.Admit(ctx, in, m); err != nil {
					k.trace.Record("kernel", in.Binding().Owner(), m.Identity, m.Source, "rejected: "+err.Error())
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
				if err := k.scheduler.Run(ctx, k.Dispatch); err != nil {
					if ctx.Err() != nil {
						return
					}
					var rejected *ports.Rejection
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
	k.runtimeState.Active.RetireAll()
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

func (k *Kernel) superseded(b *ports.Binding) bool {
	e, ok := k.runtimeState.Active.Lookup(b.Owner())
	if !ok {
		return false
	}
	return e.Superseded(b)
}
