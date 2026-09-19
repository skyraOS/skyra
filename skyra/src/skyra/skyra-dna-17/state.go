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
// pure record operations only, and dispatch/root delivery happen to the caller
// after the lock is released.
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
// stale between them. decide is pure and reads only through the restricted view; it
// returns the branch-free Request the writer then applies. Nothing partial is ever
// observed, and no driver code runs under the lock.
func (s *StateExecutor) InterpretTurn(ctx context.Context, move Move, decide func(readView, Move) (Request, error)) (commitResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	request, err := decide(stateReader{s}, move)
	if err != nil {
		return commitResult{}, err
	}
	return s.applyLocked(request)
}

// applyLocked applies a Request's operations in order within one transaction. It
// assumes mu is held. It mints fds, settles records, and collects the authorized
// executions and root results for the caller to act on OUTSIDE the lock. Enqueues
// of re-entrant Moves go through the single door.
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
				phase:    PhaseRunning,
			}
		case opAuthorizeStart:
			if err := s.authorizeStartLocked(o.inv, &result); err != nil {
				return commitResult{}, err
			}
		case opSettleReturn:
			s.settleReturnLocked(o.inv, o.obs, &result)
		case opSettleFault:
			s.settleFaultLocked(o.inv, o.reason, &result)
		default:
			return commitResult{}, fmt.Errorf("unknown operation %T", op)
		}
	}
	s.revision++
	return result, nil
}

// authorizeStartLocked is the one-transaction authority gate (possibility invariant
// 6): check the target against the installed Horizon, resolve the Surface, and mint
// the fd. On a pre-execution trap (unauthorized / unresolvable / focus exceeds
// grants) it authorizes nothing and re-enters a corrective Fault Move — the trapped
// original never executes.
func (s *StateExecutor) authorizeStartLocked(inv InvocationID, result *commitResult) error {
	rec := s.invocations[inv]
	if rec == nil {
		return fmt.Errorf("authorize: unknown invocation %q", inv)
	}
	poss, ok := s.possibilities[rec.poss]
	if !ok {
		s.trapLocked(rec, "authorize: unknown possibility")
		return nil
	}
	if _, granted := poss.Horizon.Grants[rec.surface]; !granted {
		s.trapLocked(rec, fmt.Sprintf("authorize: surface %q not in horizon", rec.surface))
		return nil
	}
	for _, f := range rec.focus {
		if _, granted := poss.Horizon.Grants[f.Target]; !granted {
			s.trapLocked(rec, fmt.Sprintf("authorize: focus %q exceeds horizon", f.Target))
			return nil
		}
	}
	if _, ok := s.surfaces[rec.surface]; !ok {
		s.trapLocked(rec, fmt.Sprintf("authorize: surface %q not registered", rec.surface))
		return nil
	}
	id := s.nextFd
	s.nextFd++
	s.fds[id] = &fdEntry{id: id, inv: rec.inv, poss: rec.poss, surface: rec.surface, gen: poss.Horizon.Generation, live: true}
	rec.fd = id
	result.authorized = append(result.authorized, dispatchItem{
		id:      id,
		surface: rec.surface,
		input:   DriverInput{Move: s.buildInvokeMoveLocked(rec)},
	})
	return nil
}

func (s *StateExecutor) trapLocked(rec *invocationRecord, reason string) {
	rec.phase = PhaseFaulted
	s.enqueue(Move{
		Identity:      s.mintMoveIDLocked(),
		PossibilityID: rec.poss,
		Fault:         &Fault{Invocation: rec.inv, Reason: reason},
	})
}

// settleReturnLocked settles a returning invocation and resolves what its return
// unblocks — resume is inside interpret. If the invocation answers the root, it
// produces the single root result; otherwise it resumes the parent in the SAME
// commit by authorizing a dispatch that carries the child observation.
func (s *StateExecutor) settleReturnLocked(inv InvocationID, obs Observation, result *commitResult) {
	rec := s.invocations[inv]
	if rec == nil || rec.phase == PhaseReturned || rec.phase == PhaseFaulted {
		return // late/duplicate return; the settlement fence already closed this
	}
	rec.phase = PhaseReturned
	rec.result = cloneObservation(obs)
	s.killFdLocked(rec.fd)

	if rec.returnTo == "" {
		out := cloneObservation(obs)
		result.rootResults = append(result.rootResults, RootResult{Invocation: inv, Observation: &out})
		return
	}
	parent := s.invocations[rec.returnTo]
	if parent == nil || parent.phase == PhaseReturned || parent.phase == PhaseFaulted {
		return
	}
	// Resume the parent's live execution: reuse its fd, hand it the child result.
	child := cloneObservation(obs)
	parent.waitingOn = ""
	parent.phase = PhaseRunning
	result.authorized = append(result.authorized, dispatchItem{
		id:      parent.fd,
		surface: parent.surface,
		input: DriverInput{
			Move:        s.buildInvokeMoveLocked(parent),
			State:       cloneDriverState(parent.driverState),
			ChildResult: &child,
		},
	})
}

func (s *StateExecutor) settleFaultLocked(inv InvocationID, reason string, result *commitResult) {
	rec := s.invocations[inv]
	if rec == nil {
		return
	}
	if rec.phase != PhaseFaulted {
		rec.phase = PhaseFaulted
	}
	s.killFdLocked(rec.fd)
	if rec.returnTo == "" {
		result.rootResults = append(result.rootResults, RootResult{Invocation: inv, Err: errors.New(reason)})
		return
	}
	// Propagate to the parent as a corrective Fault Move (minimal; the two
	// acceptance paths fault at the root, so parent propagation is lightly tested).
	parent := s.invocations[rec.returnTo]
	if parent != nil && parent.phase != PhaseReturned && parent.phase != PhaseFaulted {
		s.enqueue(Move{
			Identity:      s.mintMoveIDLocked(),
			PossibilityID: parent.poss,
			Fault:         &Fault{Invocation: parent.inv, Reason: "child faulted: " + reason},
		})
	}
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
// cannot enqueue a Return).
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
// detached copies only; handlers hold no alias into canonical state.
// ---------------------------------------------------------------------------

type readView interface {
	invocation(InvocationID) (invSnapshot, bool)
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
