package dna17

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Phase is the coarse lifecycle of an invocation record. It is NOT a per-move
// processing status (there is no `claimed`/`ready` flag): it records whether the
// execution is live, parked on a child, or settled.
type Phase string

const (
	PhaseRunning  Phase = "running"
	PhaseWaiting  Phase = "waiting"
	PhaseReturned Phase = "returned"
	PhaseFaulted  Phase = "faulted"
)

// invocationRecord is canonical continuation state, owned solely by StateExecutor.
// The waiting record (waitingOn) IS the dependency graph — legitimate canonical
// state, not a gate.
type invocationRecord struct {
	inv         InvocationID
	poss        PossibilityID
	surface     SurfaceID
	action      string
	params      []Observation
	focus       []Capability
	returnTo    InvocationID
	driverState DriverState
	phase       Phase
	waitingOn   InvocationID
	result      Observation
	fd          Identifier
}

// fdEntry is the writer's file-descriptor table row. The adapter holds only the
// Identifier; scope (inv/poss/surface), the authority generation it was minted
// under, and liveness live here. A dead entry rejects every request (the fence).
type fdEntry struct {
	id      Identifier
	inv     InvocationID
	poss    PossibilityID
	surface SurfaceID
	gen     uint64
	live    bool
}

// StateExecutor is the SOLE canonical writer. All access is serialized by mu; no
// component receives a map alias. It never calls a driver/transport/Surface/host
// callback while holding mu — the interpret turn and each emission commit apply
// pure record operations only, and dispatch/root delivery/re-enqueue happen to the
// caller after the lock is released.
type StateExecutor struct {
	mu sync.Mutex

	// revision is present for honesty and future concurrency; it is deliberately
	// NOT consulted on the interpret path (no OCC — read and commit are one turn).
	revision uint64

	possibilities map[PossibilityID]Possibility
	surfaces      map[SurfaceID]Surface
	invocations   map[InvocationID]*invocationRecord
	fds           map[Identifier]*fdEntry

	nextFd  Identifier
	nextInv uint64
	nextMv  uint64

	// enqueue is the single inbound door. Even the writer re-enters emitted Moves
	// through it. Wired by the CU at construction.
	enqueue func(Move)
}

func newStateExecutor(surfaces map[SurfaceID]Surface) *StateExecutor {
	copied := make(map[SurfaceID]Surface, len(surfaces))
	for id, s := range surfaces {
		copied[id] = s
	}
	return &StateExecutor{
		possibilities: make(map[PossibilityID]Possibility),
		surfaces:      copied,
		invocations:   make(map[InvocationID]*invocationRecord),
		fds:           make(map[Identifier]*fdEntry),
		nextFd:        1,
	}
}

// Open records a Possibility and enqueues its first Move atomically (the collapsed
// world's "possibility record + first move" opening). The one queue is the single
// mailbox, so opening records the Possibility and posts the first Move together.
func (s *StateExecutor) Open(ctx context.Context, request OpenPossibility) error {
	if request.Identity == "" || request.First.Invoke == nil {
		return errors.New("open requires an identity and a first Invoke")
	}
	s.mu.Lock()
	if _, exists := s.possibilities[request.Identity]; exists {
		s.mu.Unlock()
		return fmt.Errorf("possibility %q already open", request.Identity)
	}
	first := request.First
	first.PossibilityID = request.Identity
	if first.Identity == "" {
		first.Identity = s.mintMoveIDLocked()
	}
	s.possibilities[request.Identity] = Possibility{
		Identity:    request.Identity,
		FirstMoveID: first.Identity,
		Horizon:     cloneHorizon(request.Horizon),
	}
	s.revision++
	s.mu.Unlock()
	s.enqueue(first)
	return nil
}

// InterpretTurn is one turn of the single service: it takes the lock ONCE and runs
// the handler's read (decide) and the commit inside it, so the snapshot cannot go
// stale between them. decide is pure: it reads only through the restricted view and
// makes every decision, returning the branch-free Request of DECIDED operations the
// writer then applies. Nothing partial is ever observed, and no driver code runs
// under the lock.
func (s *StateExecutor) InterpretTurn(ctx context.Context, move Move, decide func(readView, Move) (Request, error)) (commitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, err := decide(stateReader{s}, move)
	if err != nil {
		return commitResult{}, err
	}
	return s.applyLocked(request)
}

// applyLocked applies a Request's DECIDED operations. It assumes mu is held. Every
// operation is a pure WRITE: it does not read canonical state to decide an outcome
// (interpret already decided) and performs no side effects — re-entrant Moves are
// collected into result.enqueue and enqueued/dispatched by the caller AFTER the lock
// releases. Because the write phase makes no fallible decision, it cannot partially
// fail, so no rollback or private working set is needed. The only reads here are
// mechanical: locating a record just written in this same Request to attach the fd
// or the settlement.
func (s *StateExecutor) applyLocked(request Request) (commitResult, error) {
	var result commitResult
	for _, op := range request.Operations {
		switch o := op.(type) {
		case opRecordInvocation:
			s.invocations[o.inv.InvocationID] = &invocationRecord{
				inv:      o.inv.InvocationID,
				poss:     o.poss,
				surface:  o.target,
				action:   o.inv.Action,
				params:   cloneObservations(o.inv.Params),
				focus:    cloneCapabilities(o.inv.Focus),
				returnTo: o.inv.ReturnTo,
				phase:    o.phase, // decided by interpret: Running (authorized) or Faulted (trapped)
			}
		case opAuthorizeStart:
			// Already authorized in interpret. Mint the fd and the dispatch — no read
			// of the Horizon, no authority branch.
			rec := s.invocations[o.inv]
			id := s.nextFd
			s.nextFd++
			s.fds[id] = &fdEntry{id: id, inv: rec.inv, poss: rec.poss, surface: rec.surface, gen: s.possibilities[rec.poss].Horizon.Generation, live: true}
			rec.fd = id
			result.authorized = append(result.authorized, dispatchItem{
				id:      id,
				surface: rec.surface,
				input:   DriverInput{Move: s.buildInvokeMoveLocked(rec)},
			})
		case opEmitFault:
			result.enqueue = append(result.enqueue, Move{
				Identity:      s.mintMoveIDLocked(),
				PossibilityID: o.poss,
				Fault:         &Fault{Invocation: o.inv, Reason: o.reason},
			})
		case opMarkReturned:
			rec := s.invocations[o.inv]
			rec.phase = PhaseReturned
			rec.result = cloneObservation(o.obs)
			s.killFdLocked(rec.fd)
		case opRootResult:
			out := cloneObservation(o.obs)
			result.rootResults = append(result.rootResults, RootResult{Invocation: o.inv, Observation: &out})
		case opResumeParent:
			parent := s.invocations[o.parent]
			parent.waitingOn = ""
			parent.phase = PhaseRunning
			child := cloneObservation(o.child)
			result.authorized = append(result.authorized, dispatchItem{
				id:      parent.fd,
				surface: parent.surface,
				input: DriverInput{
					Move:        s.buildInvokeMoveLocked(parent),
					State:       cloneDriverState(parent.driverState),
					ChildResult: &child,
				},
			})
		case opMarkFaulted:
			if rec := s.invocations[o.inv]; rec != nil {
				rec.phase = PhaseFaulted
				s.killFdLocked(rec.fd)
			}
		case opRootFault:
			result.rootResults = append(result.rootResults, RootResult{Invocation: o.inv, Err: errors.New(o.reason)})
		default:
			return commitResult{}, fmt.Errorf("unknown operation %T", op)
		}
	}
	s.revision++
	return result, nil
}

// stateVia returns the execution-plane profile the Identifier grants. Liveness is
// checked live at each request, so a settled/dead fd rejects everything.
func (s *StateExecutor) stateVia(id Identifier) ExecutionState {
	return &executionProfile{state: s, id: id}
}

type executionProfile struct {
	state *StateExecutor
	id    Identifier
}

// Commit is one execution-plane transaction: replace the fd invocation's driver
// state, and either enqueue emitted Moves or record a failure — together. A dead
// fd rejects the whole thing (the settlement fence: a late emission after a fault
// cannot enqueue a Return). The liveness check IS a legitimate commit-time read: it
// is the fence, evaluated when the worker calls, and cannot be pre-decided.
func (e *executionProfile) Commit(ctx context.Context, ec EmissionCommit) (EmissionResult, error) {
	s := e.state
	s.mu.Lock()
	entry := s.fds[e.id]
	if entry == nil || !entry.live {
		s.mu.Unlock()
		return EmissionResult{}, fmt.Errorf("fd %d is not live", e.id)
	}
	rec := s.invocations[entry.inv]
	if rec == nil {
		s.mu.Unlock()
		return EmissionResult{}, fmt.Errorf("fd %d names an unknown invocation", e.id)
	}
	rec.driverState = cloneDriverState(ec.State)

	if ec.Failure != nil {
		// The execution declared itself failed. Kill the fd immediately (fence),
		// then re-enter a Fault Move so the interpreter settles it uniformly.
		entry.live = false
		rec.phase = PhaseFaulted
		fault := Move{Identity: s.mintMoveIDLocked(), PossibilityID: entry.poss, Fault: &Fault{Invocation: entry.inv, Reason: ec.Failure.Error()}}
		s.revision++
		s.mu.Unlock()
		s.enqueue(fault)
		return EmissionResult{}, nil
	}

	// Prepare emitted Moves: the writer mints child ids and binds source/parent.
	prepared := make([]Move, 0, len(ec.Moves))
	for _, m := range ec.Moves {
		m.PossibilityID = entry.poss
		m.Source = MoveSource{SurfaceID: entry.surface, InvocationID: entry.inv}
		if m.Identity == "" {
			m.Identity = s.mintMoveIDLocked()
		}
		switch {
		case m.Invoke != nil:
			m.Invoke.InvocationID = s.mintInvIDLocked()
			m.Invoke.ReturnTo = entry.inv
			rec.waitingOn = m.Invoke.InvocationID
			rec.phase = PhaseWaiting
		case m.Return != nil:
			m.Return.Answering = entry.inv
		}
		prepared = append(prepared, m)
	}
	s.revision++
	s.mu.Unlock()

	for _, m := range prepared {
		s.enqueue(m)
	}
	return EmissionResult{Enqueue: prepared}, nil
}

func (s *StateExecutor) killFdLocked(id Identifier) {
	if entry := s.fds[id]; entry != nil {
		entry.live = false
	}
}

func (s *StateExecutor) buildInvokeMoveLocked(rec *invocationRecord) Move {
	return Move{
		Identity:      s.mintMoveIDLocked(),
		PossibilityID: rec.poss,
		Target:        rec.surface,
		Source:        MoveSource{SurfaceID: rec.surface, InvocationID: rec.inv},
		Invoke: &Invoke{
			InvocationID: rec.inv,
			Action:       rec.action,
			Params:       cloneObservations(rec.params),
			Focus:        cloneCapabilities(rec.focus),
			ReturnTo:     rec.returnTo,
		},
	}
}

func (s *StateExecutor) mintInvIDLocked() InvocationID {
	s.nextInv++
	return InvocationID(fmt.Sprintf("inv-%d", s.nextInv))
}

func (s *StateExecutor) mintMoveIDLocked() MoveID {
	s.nextMv++
	return MoveID(fmt.Sprintf("mv-%d", s.nextMv))
}

// surfaceOf returns the bound Surface for a dispatch. It is a read door only.
func (s *StateExecutor) surfaceOf(id SurfaceID) (Surface, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	surface, ok := s.surfaces[id]
	return surface, ok
}

// snapshot returns a detached view of an invocation record for tests/inspection.
func (s *StateExecutor) snapshot(inv InvocationID) (InvocationView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.invocations[inv]
	if !ok {
		return InvocationView{}, false
	}
	return InvocationView{Invocation: inv, Phase: rec.phase, DriverState: cloneDriverState(rec.driverState)}, true
}

// InvocationView is a detached inspection copy.
type InvocationView struct {
	Invocation  InvocationID
	Phase       Phase
	DriverState DriverState
}

// ---------------------------------------------------------------------------
// readView — the restricted read profile handed to interpret handlers. It returns
// detached copies only; handlers hold no alias into canonical state. Interpret does
// ALL reads here (invocation correlation, Horizon authority, surface resolution) so
// the commit can be pure writes.
// ---------------------------------------------------------------------------

type readView interface {
	invocation(InvocationID) (invSnapshot, bool)
	horizon(PossibilityID) (Horizon, bool)
	surfaceRegistered(SurfaceID) bool
}

type invSnapshot struct {
	inv      InvocationID
	poss     PossibilityID
	surface  SurfaceID
	returnTo InvocationID
	phase    Phase
}

type stateReader struct{ s *StateExecutor }

func (r stateReader) invocation(inv InvocationID) (invSnapshot, bool) {
	rec, ok := r.s.invocations[inv] // mu already held by InterpretTurn
	if !ok {
		return invSnapshot{}, false
	}
	return invSnapshot{inv: rec.inv, poss: rec.poss, surface: rec.surface, returnTo: rec.returnTo, phase: rec.phase}, true
}

func (r stateReader) horizon(poss PossibilityID) (Horizon, bool) {
	p, ok := r.s.possibilities[poss] // mu already held
	if !ok {
		return Horizon{}, false
	}
	return cloneHorizon(p.Horizon), true // detached: the handler holds no alias into canonical state
}

func (r stateReader) surfaceRegistered(id SurfaceID) bool {
	_, ok := r.s.surfaces[id] // mu already held
	return ok
}

// ---------------------------------------------------------------------------
// small clone helpers (detachment).
// ---------------------------------------------------------------------------

func cloneObservation(o Observation) Observation {
	o.Bytes = append([]byte(nil), o.Bytes...)
	return o
}

func cloneObservations(in []Observation) []Observation {
	if in == nil {
		return nil
	}
	out := make([]Observation, len(in))
	for i, o := range in {
		out[i] = cloneObservation(o)
	}
	return out
}

func cloneCapabilities(in []Capability) []Capability {
	if in == nil {
		return nil
	}
	return append([]Capability(nil), in...)
}

func cloneDriverState(d DriverState) DriverState {
	d.Payload = append([]byte(nil), d.Payload...)
	return d
}

func cloneHorizon(h Horizon) Horizon {
	grants := make(map[SurfaceID]Grant, len(h.Grants))
	for k, v := range h.Grants {
		grants[k] = v
	}
	return Horizon{Generation: h.Generation, Grants: grants}
}
