package dna17

import (
	"errors"
	"fmt"
)

// decode is the entry, not a phase: the opcode (which payload pointer is set)
// selects the handler BEFORE any read, because the read is opcode-dependent. The
// interpreter does not decide WHETHER to branch — the branch is the opcode.
//
// Each returned handler performs EVERY read over the locked snapshot (via readView)
// and makes EVERY decision — authority, correlation, resume, root-vs-propagate — and
// builds every output (the dispatch Moves). It returns a branch-free Request of
// DECIDED operations that carry all their data. The commit (StateExecutor.applyLocked)
// reads no canonical state: it only allocates ids and applies keyed writes. Read and
// commit run in one uninterrupted turn of the single service (InterpretTurn), so
// nothing goes stale and there is no revision/OCC and no per-move status.
func decode(move Move) (func(readView, Move) (Request, error), error) {
	switch {
	case move.Invoke != nil:
		return interpretInvoke, nil
	case move.Return != nil:
		return interpretReturn, nil
	case move.Fault != nil:
		return interpretFault, nil
	default:
		return nil, errors.New("undecodable move: no opcode payload")
	}
}

// interpretInvoke records the invocation unconditionally (an invoke is a real wire
// event) and decides authority HERE, in the read phase: it reads the Possibility's
// Horizon and the Surface registry and either authorizes the start — building the
// dispatch Move and carrying the Horizon generation — or traps. The commit re-reads
// nothing; it mints the fd and stamps the Move. Deferring the check to commit was the
// leak: the single locked turn already makes check+mint atomic, so the read belongs
// here, ahead of the writes.
func interpretInvoke(view readView, move Move) (Request, error) {
	inv := move.Invoke
	if inv.InvocationID == "" || inv.Action == "" {
		return Request{}, errors.New("invoke requires an invocation id and action")
	}
	record := opRecordInvocation{inv: *inv, target: move.Target, poss: move.PossibilityID, source: move.Source}

	gen, reason := authorize(view, move.PossibilityID, move.Target, inv.Focus)
	if reason != "" {
		record.phase = PhaseFaulted
		return Request{
			Move:       move,
			Operations: []Operation{record, opEmitFault{poss: move.PossibilityID, inv: inv.InvocationID, reason: reason}},
		}, nil
	}
	record.phase = PhaseRunning
	dispatch := invokeTemplate(inv.InvocationID, move.PossibilityID, move.Target, inv.Action, inv.Params, inv.Focus, inv.ReturnTo)
	return Request{
		Move:       move,
		Operations: []Operation{record, opAuthorizeStart{gen: gen, dispatch: dispatch}},
	}, nil
}

// authorize is the authority READ: it reads only the Horizon (grants) and the Surface
// registry and returns the Horizon, its generation, and "" when authorized (or a trap
// reason). All in the interpret phase — never a write. Running it before the writes
// (not inside the commit) is what keeps read and write separated while staying in one
// atomic turn.
func authorize(view readView, poss PossibilityID, target SurfaceID, focus []Capability) (uint64, string) {
	h, ok := view.horizon(poss)
	if !ok {
		return 0, "authorize: unknown possibility"
	}
	if _, granted := h.Grants[target]; !granted {
		return h.Generation, fmt.Sprintf("authorize: surface %q not in horizon", target)
	}
	for _, f := range focus {
		if _, granted := h.Grants[f.Target]; !granted {
			return h.Generation, fmt.Sprintf("authorize: focus %q exceeds horizon", f.Target)
		}
	}
	if !view.surfaceRegistered(target) {
		return h.Generation, fmt.Sprintf("authorize: surface %q not registered", target)
	}
	return h.Generation, ""
}

// interpretReturn correlates the return against the recorded invocation (never a
// payload claim) and decides settlement HERE: an unknown or already-settled return is
// dropped (the fence, decided in the read phase); a root return yields a root result;
// otherwise the live parent is resumed in this same turn — interpret reading the
// parent's fields and driver state, and building its resume dispatch. The commit only
// applies the decided, fully-carried writes.
func interpretReturn(view readView, move Move) (Request, error) {
	ret := move.Return
	if ret.Answering == "" {
		return Request{}, errors.New("return requires an answered invocation")
	}
	rec, ok := view.invocation(ret.Answering)
	if !ok {
		return Request{Move: move}, nil // unknown source: nothing to settle
	}
	if rec.phase == PhaseReturned || rec.phase == PhaseFaulted {
		return Request{Move: move}, nil // late/duplicate: already settled
	}

	ops := []Operation{opMarkReturned{inv: ret.Answering, fd: rec.fd, obs: ret.Observation}}

	if rec.returnTo == "" {
		ops = append(ops, opRootResult{inv: ret.Answering, obs: ret.Observation})
		return Request{Move: move, Operations: ops}, nil
	}

	parent, ok := view.invocation(rec.returnTo)
	if !ok || parent.phase == PhaseReturned || parent.phase == PhaseFaulted {
		return Request{Move: move, Operations: ops}, nil // parent gone: settle the child only
	}
	dispatch := invokeTemplate(parent.inv, parent.poss, parent.surface, parent.action, parent.params, parent.focus, parent.returnTo)
	ops = append(ops, opResumeParent{parentFd: parent.fd, dispatch: dispatch, driverState: parent.driverState, child: ret.Observation})
	return Request{Move: move, Operations: ops}, nil
}

// interpretFault settles a faulted invocation and decides propagation HERE: a root
// fault yields a root result; otherwise it emits a corrective Fault Move to the live
// parent. It carries the fd to kill. Trap owns the corrective-work decision; here the
// correction is settlement.
func interpretFault(view readView, move Move) (Request, error) {
	f := move.Fault
	if f.Invocation == "" {
		return Request{}, errors.New("fault requires an invocation")
	}
	rec, ok := view.invocation(f.Invocation)
	if !ok {
		return Request{Move: move}, nil
	}

	ops := []Operation{opMarkFaulted{inv: f.Invocation, fd: rec.fd}}

	if rec.returnTo == "" {
		ops = append(ops, opRootFault{inv: f.Invocation, reason: f.Reason})
		return Request{Move: move, Operations: ops}, nil
	}
	parent, ok := view.invocation(rec.returnTo)
	if ok && parent.phase != PhaseReturned && parent.phase != PhaseFaulted {
		ops = append(ops, opEmitFault{poss: parent.poss, inv: rec.returnTo, reason: "child faulted: " + f.Reason})
	}
	return Request{Move: move, Operations: ops}, nil
}

// invokeTemplate builds the dispatch Move handed to a bound Surface, WITHOUT an
// Identity (the writer stamps it at commit — the only writer-owned bit). Interpret
// builds it from data it read, so the commit never reads a record to construct it.
func invokeTemplate(inv InvocationID, poss PossibilityID, surface SurfaceID, action string, params []Observation, focus []Capability, returnTo InvocationID) Move {
	return Move{
		PossibilityID: poss,
		Target:        surface,
		Source:        MoveSource{SurfaceID: surface, InvocationID: inv},
		Invoke: &Invoke{
			InvocationID: inv,
			Action:       action,
			Params:       cloneObservations(params),
			Focus:        cloneCapabilities(focus),
			ReturnTo:     returnTo,
		},
	}
}
