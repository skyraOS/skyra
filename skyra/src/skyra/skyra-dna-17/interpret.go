package dna17

import (
	"errors"
	"fmt"
)

// decode is the entry, not a phase: the opcode (which payload pointer is set)
// selects the handler BEFORE any read, because the read is opcode-dependent. The
// interpreter does not decide WHETHER to branch — the branch is the opcode.
//
// Each returned handler performs its own opcode-specific reads over the locked
// snapshot (via readView) AND makes every decision there — authority, correlation,
// resume, root-vs-propagate. It returns a branch-free Request{Move, []Operation}
// whose operations are decided WRITES; the commit (StateExecutor.applyLocked) applies
// them without reading canonical state to decide anything. Read and commit run in one
// uninterrupted turn of the single service (InterpretTurn), so nothing goes stale and
// there is no revision/OCC and no per-move status.
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
// Horizon and the Surface registry and either authorizes the start or traps. The
// commit does not re-check — opAuthorizeStart mints the fd for an ALREADY-authorized
// invocation, and a trap records the invocation faulted plus a corrective Fault Move.
// Deferring the check to commit was the leak: the single locked turn already makes
// check+mint atomic, so the read belongs here, ahead of the writes.
func interpretInvoke(view readView, move Move) (Request, error) {
	inv := move.Invoke
	if inv.InvocationID == "" || inv.Action == "" {
		return Request{}, errors.New("invoke requires an invocation id and action")
	}
	record := opRecordInvocation{inv: *inv, target: move.Target, poss: move.PossibilityID, source: move.Source}

	if reason := authorityReason(view, move.PossibilityID, move.Target, inv.Focus); reason != "" {
		record.phase = PhaseFaulted
		return Request{
			Move:       move,
			Operations: []Operation{record, opEmitFault{poss: move.PossibilityID, inv: inv.InvocationID, reason: reason}},
		}, nil
	}
	record.phase = PhaseRunning
	return Request{
		Move:       move,
		Operations: []Operation{record, opAuthorizeStart{inv: inv.InvocationID}},
	}, nil
}

// authorityReason performs the authority READ and returns "" when the target is
// authorized, or a trap reason otherwise. It reads only the Horizon (grants) and the
// Surface registry — canonical reads made in the interpret phase, never a write. This
// is the possibility authority gate; running it before the writes (not inside the
// commit) is what keeps read and write separated while staying in one atomic turn.
func authorityReason(view readView, poss PossibilityID, target SurfaceID, focus []Capability) string {
	h, ok := view.horizon(poss)
	if !ok {
		return "authorize: unknown possibility"
	}
	if _, granted := h.Grants[target]; !granted {
		return fmt.Sprintf("authorize: surface %q not in horizon", target)
	}
	for _, f := range focus {
		if _, granted := h.Grants[f.Target]; !granted {
			return fmt.Sprintf("authorize: focus %q exceeds horizon", f.Target)
		}
	}
	if !view.surfaceRegistered(target) {
		return fmt.Sprintf("authorize: surface %q not registered", target)
	}
	return ""
}

// interpretReturn correlates the return against the recorded invocation (never a
// payload claim) and decides settlement HERE: an unknown or already-settled return is
// dropped (the fence, decided in the read phase); a root return yields a root result;
// otherwise the live parent is resumed in this same turn. The commit only applies the
// decided writes.
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

	ops := []Operation{opMarkReturned{inv: ret.Answering, obs: ret.Observation}}

	if rec.returnTo == "" {
		ops = append(ops, opRootResult{inv: ret.Answering, obs: ret.Observation})
		return Request{Move: move, Operations: ops}, nil
	}

	parent, ok := view.invocation(rec.returnTo)
	if !ok || parent.phase == PhaseReturned || parent.phase == PhaseFaulted {
		return Request{Move: move, Operations: ops}, nil // parent gone: settle the child only
	}
	ops = append(ops, opResumeParent{parent: rec.returnTo, child: ret.Observation})
	return Request{Move: move, Operations: ops}, nil
}

// interpretFault settles a faulted invocation and decides propagation HERE: a root
// fault yields a root result; otherwise it emits a corrective Fault Move to the live
// parent. Trap owns the corrective-work decision; here the correction is settlement.
func interpretFault(view readView, move Move) (Request, error) {
	f := move.Fault
	if f.Invocation == "" {
		return Request{}, errors.New("fault requires an invocation")
	}
	rec, ok := view.invocation(f.Invocation)
	if !ok {
		return Request{Move: move}, nil
	}

	ops := []Operation{opMarkFaulted{inv: f.Invocation}}

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
