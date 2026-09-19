// Package scheduler owns continuation state, selection, and dispatch reservations.
package scheduler

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"skyra-v1/src/skyra/skyra-dna-v14/controller"
	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

// Scheduler owns ALL live continuation state. The mutex protects admission,
// register updates, result recording, and selection. Controller is injected at
// startup. Complete is a host callback, called outside mu, not another state store.
type Scheduler struct {
	mu            sync.Mutex
	controller    controller.Controller
	continuations map[protocol.InvocationID]protocol.Continuation
	complete      func(protocol.RootResult) error
}

func NewScheduler(cc controller.Controller, complete func(protocol.RootResult) error) (*Scheduler, error) {
	if support.Missing(cc) || complete == nil {
		return nil, errors.New("controller and completion callback required")
	}
	return &Scheduler{controller: cc, complete: complete, continuations: make(map[protocol.InvocationID]protocol.Continuation)}, nil
}

// interpret is Scheduler's private, read-only contract assessment. The caller
// holds mu while reading live state and applying the assessment through the write
// path; interpret itself neither mutates state nor calls drivers. Assessment is
// distinct from dispatch lifecycle State.
// Compiled matching remains stubbed. Missing observations mean Waiting, not
// Mismatch. Nil compiled means no compiled expectations, not automatic Satisfied;
// general declaration-path policy remains open. For the echo probe, the accepted
// completeness stub below permits one parent resumption when all declared child
// observations arrive; it is not proof that a compiled contract was satisfied.
// No separate interpreter dependency exists.
func (s *Scheduler) interpret(live protocol.Register, compiled *protocol.CompiledRegister) protocol.Assessment {
	return protocol.Waiting
}

// Snapshot is observation-only for the host/tests. It cannot be used to mutate
// Scheduler or as an alternate execution path. All nested mutable data is detached.
func (s *Scheduler) Snapshot(id protocol.InvocationID) (protocol.Continuation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.continuations[id]
	return support.CloneContinuation(c), ok
}

// rootLocked follows only scheduler-established parent links. Caller holds mu.
func (s *Scheduler) rootLocked(id protocol.InvocationID) (protocol.InvocationID, protocol.Continuation, bool) {
	for {
		c, ok := s.continuations[id]
		if !ok {
			return id, c, false
		}
		parent := c.Move.Op.(protocol.Invoke).ReturnFd
		if parent.Value == "" {
			return id, c, true
		}
		id = parent
	}
}

// faultLocked marks the source and its root faulted and reserves one host
// notification. Caller holds mu; the returned notification is delivered after unlock.
func (s *Scheduler) faultLocked(id protocol.InvocationID, err error) (*protocol.RootResult, error) {
	c, ok := s.continuations[id]
	if !ok {
		return nil, errors.New("unknown failure source")
	}
	rootID, root, _ := s.rootLocked(id)
	if root.Result != nil || root.State == protocol.Faulted {
		return nil, nil
	}
	c.State = protocol.Faulted
	s.continuations[id] = c
	root = s.continuations[rootID]
	root.State = protocol.Faulted
	s.continuations[rootID] = root
	return &protocol.RootResult{Invocation: rootID, Err: err}, nil
}

// Write uses ctx for NEW admission only; existing work keeps its saved Context.
// Returns describe event-processing/reporting failure. The protocol carries actual
// execution failures. Admission, affected selection, and handoff reservation are
// synchronized; all CC/driver/host calls happen after unlocking.
func (s *Scheduler) Write(ctx context.Context, event protocol.OpcodeEvent) error {
	var selected []protocol.Continuation
	var notification *protocol.RootResult
	s.mu.Lock()
	err := func() error {
		switch e := event.(type) {
		case protocol.AdmissionFailure:
			if e.Err == nil {
				return errors.New("admission failure requires error")
			}
			return e.Err
		case protocol.ExecutionFailure:
			if e.Err == nil {
				return errors.New("execution failure requires error")
			}
			var err error
			notification, err = s.faultLocked(e.Source, e.Err)
			return err
		case protocol.DispatchResult:
			c, ok := s.continuations[e.Invocation]
			if !ok {
				return errors.New("unknown dispatch source")
			}
			if c.Result != nil || c.State == protocol.Faulted {
				return nil
			}
			if e.Err != nil || !e.Authorized {
				err := e.Err
				if err == nil {
					err = errors.New("dispatch denied")
				}
				var failure error
				notification, failure = s.faultLocked(e.Invocation, err)
				return failure
			}
			// An initial acknowledgement cannot describe a resumed handoff.
			if !c.Resumed {
				c.State = protocol.InFlight
				s.continuations[e.Invocation] = c
			}
			return nil
		case protocol.MoveBatch:
			if len(e.Moves) == 0 {
				return errors.New("empty admission batch")
			}
			var source protocol.Continuation
			var producer protocol.Invoke
			if e.Source != nil {
				var ok bool
				source, ok = s.continuations[*e.Source]
				if !ok {
					return errors.New("unknown output source")
				}
				producer = source.Move.Op.(protocol.Invoke)
			}
			if ret, ok := e.Moves[0].Op.(protocol.Return); ok {
				m := e.Moves[0]
				if len(e.Moves) != 1 || e.Source == nil || m.Target.Invocation != *e.Source || m.Root != source.Move.Root || m.Target.Entity != source.Move.Target.Entity || m.Target.Address != source.Move.Target.Address || m.Target.Host != source.Move.Target.Host {
					return errors.New("unbound return")
				}
				if source.Result != nil {
					if reflect.DeepEqual(*source.Result, ret.Observation) {
						return nil
					}
					return errors.New("conflicting return")
				}
				_, root, _ := s.rootLocked(*e.Source)
				if root.State == protocol.Faulted || source.State == protocol.Faulted {
					return errors.New("workflow faulted")
				}
				if len(producer.Register.Children) > 0 && !source.Resumed {
					return errors.New("parent returned before resumption")
				}
				result := support.CloneObservation(ret.Observation)
				source.Result = &result
				s.continuations[*e.Source] = source
				if producer.ReturnFd.Value == "" {
					o := support.CloneObservation(result)
					notification = &protocol.RootResult{Invocation: *e.Source, Observation: &o}
					return nil
				}
				parent := s.continuations[producer.ReturnFd]
				inv := parent.Move.Op.(protocol.Invoke)
				entry, exists := inv.Register.Children[*e.Source]
				if !exists {
					return errors.New("missing parent child entry")
				}
				entry.Observation = &protocol.ChildObservation{InvocationID: *e.Source, Observation: support.CloneObservation(result)}
				inv.Register.Children[*e.Source] = entry
				parent.Move.Op = inv
				s.continuations[producer.ReturnFd] = parent
				selected = append(selected, s.selectMoves(parent.Move)...)
				return nil
			}
			if e.Source == nil {
				if ctx == nil {
					return errors.New("context required")
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				if len(e.Moves) != 1 {
					return errors.New("submit one root")
				}
			} else {
				_, root, _ := s.rootLocked(*e.Source)
				if source.Result != nil || root.State == protocol.Faulted || source.State == protocol.Faulted || source.Resumed || len(producer.Register.Children) > 0 {
					return errors.New("source cannot declare another batch")
				}
				ctx = source.Context
				if err := ctx.Err(); err != nil {
					return err
				}
			}
			seen := make(map[protocol.InvocationID]bool)
			for _, m := range e.Moves {
				inv, ok := m.Op.(protocol.Invoke)
				if !ok || inv.InvocationID.Value == "" || !support.EmptyRegister(inv.Register) {
					return errors.New("invalid invoke admission")
				}
				if _, exists := s.continuations[inv.InvocationID]; exists || seen[inv.InvocationID] {
					return errors.New("invocation identity reused")
				}
				seen[inv.InvocationID] = true
				if e.Source == nil {
					if inv.ReturnFd.Value != "" {
						return errors.New("root cannot name parent")
					}
				} else {
					if inv.ReturnFd != *e.Source || m.Root != source.Move.Root || !reflect.DeepEqual(inv.Horizon, producer.Horizon) {
						return errors.New("invalid child linkage or horizon")
					}
				}
			}
			if e.Source != nil {
				producer.Register.Children = make(map[protocol.InvocationID]protocol.RegisterEntry)
			}
			for _, m := range e.Moves {
				m = support.CloneMove(m)
				inv := m.Op.(protocol.Invoke)
				s.continuations[inv.InvocationID] = protocol.Continuation{Move: m, Context: ctx}
				if e.Source != nil {
					producer.Register.Order = append(producer.Register.Order, inv.InvocationID)
					producer.Register.Children[inv.InvocationID] = protocol.RegisterEntry{Invocation: protocol.ChildInvocation{InvocationID: inv.InvocationID, Move: support.CloneMove(m)}}
				}
			}
			if e.Source != nil {
				source.Move.Op = producer
				s.continuations[*e.Source] = source
			}
			for _, m := range e.Moves {
				selected = append(selected, s.selectMoves(m)...)
			}
			return nil
		default:
			return errors.New("unknown scheduler event")
		}
	}()
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if notification != nil {
		return s.complete(*notification)
	}
	return s.run(selected)
}

// DECIDED FOR THE ECHO LOOP: a newly admitted Invoke is eligible for one initial
// dispatch. Record its reservation under mu before handing it off. A source-bound
// Return fills the matching parent child entry; missing observations keep that
// parent waiting. Once every child in its declared batch has an observation,
// reserve Resumed and select the parent's existing Invoke once with its updated
// Register and saved Context. An empty child set alone does not trigger resumption.
// Admit the entire child batch before selecting any child. Return is never
// dispatched; duplicate returns/late acknowledgements cannot repeat selection.
// This is a completeness stub, not compiled matching or a general retry policy.
// Compiled expectations do not need to be implemented for this local proof.
//
// Caller holds mu. Returned Continuations are detached dispatch snapshots,
// including Invoke.Register maps, order, observation bytes, and grants. They are
// not new authoritative stores. Only Invoke is dispatched. Completing the child
// set makes the existing parent Invoke selectable once with its updated Register
// and saved Context; Return itself is applied by Write, not sent to a driver.
func (s *Scheduler) selectMoves(m protocol.Move) []protocol.Continuation {
	inv, ok := m.Op.(protocol.Invoke)
	if !ok {
		return nil
	}
	c, ok := s.continuations[inv.InvocationID]
	if !ok || c.Result != nil || c.State == protocol.Faulted {
		return nil
	}
	live := c.Move.Op.(protocol.Invoke)
	if c.State != 0 {
		if c.Resumed || !support.CompleteRegister(live.Register) {
			return nil
		}
		_ = s.interpret(live.Register, c.CompiledRegister)
		c.Resumed = true
	}
	c.State = protocol.Dispatched
	s.continuations[inv.InvocationID] = c
	return []protocol.Continuation{support.CloneContinuation(c)}
}

// run dispatches snapshots already selected and reserved by Write. It does not
// select again. The root-status check holds mu only until the external handoff;
// it suppresses remaining work after an observed fault or completion.
func (s *Scheduler) run(selected []protocol.Continuation) error {
	var errs []error
	for _, c := range selected {
		s.mu.Lock()
		_, root, _ := s.rootLocked(c.Move.Op.(protocol.Invoke).InvocationID)
		stopped := root.State == protocol.Faulted || root.Result != nil
		s.mu.Unlock()
		if stopped {
			continue
		}
		saved := c.Context
		if err := s.controller.Dispatch(saved, c.Move, func(e protocol.OpcodeEvent) error { return s.Write(saved, e) }); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
