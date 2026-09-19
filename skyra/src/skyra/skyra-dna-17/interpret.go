package dna17

import "errors"

// decode is the entry, not a phase: the opcode (which payload pointer is set)
// selects the handler BEFORE any read, because the read is opcode-dependent. The
// interpreter does not decide WHETHER to branch — the branch is the opcode.
//
// Each returned decide function performs its own opcode-specific read over the
// locked snapshot (via readView) and returns the branch-free Request{Move,
// []Operation}. The read and the commit run in one uninterrupted turn of the
// single service (see StateExecutor.InterpretTurn), so nothing goes stale and
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

// interpretInvoke: an Invoke reads target authority + input availability. The
// authority check itself is deferred to the commit (opAuthorizeStart), where it
// happens in one transaction with Surface binding and fd minting. The handler's
// job is to record the invocation and request its authorization.
func interpretInvoke(_ readView, move Move) (Request, error) {
	inv := move.Invoke
	if inv.InvocationID == "" || inv.Action == "" {
		return Request{}, errors.New("invoke requires an invocation id and action")
	}
	return Request{
		Move: move,
		Operations: []Operation{
			opRecordInvocation{inv: *inv, target: move.Target, poss: move.PossibilityID, source: move.Source},
			opAuthorizeStart{inv: inv.InvocationID},
		},
	}, nil
}

// interpretReturn: a Return correlates against the recorded invocation it answers
// (never a payload claim) and requests its settlement. Settlement resolves what the
// return unblocks — root result or parent resume — inside this same turn.
func interpretReturn(view readView, move Move) (Request, error) {
	ret := move.Return
	if ret.Answering == "" {
		return Request{}, errors.New("return requires an answered invocation")
	}
	if _, ok := view.invocation(ret.Answering); !ok {
		// Unknown source: nothing to settle. A Return is admitted only against a
		// recorded invocation, so this is dropped rather than trusted.
		return Request{Move: move}, nil
	}
	return Request{
		Move:       move,
		Operations: []Operation{opSettleReturn{inv: ret.Answering, obs: ret.Observation}},
	}, nil
}

// interpretFault: a fault Move is a payload variant, interpreted normally. It
// settles the faulted invocation (root fault or parent propagation). Trap owns any
// corrective-work decision; here the correction is settlement.
func interpretFault(_ readView, move Move) (Request, error) {
	f := move.Fault
	if f.Invocation == "" {
		return Request{}, errors.New("fault requires an invocation")
	}
	return Request{
		Move:       move,
		Operations: []Operation{opSettleFault{inv: f.Invocation, reason: f.Reason}},
	}, nil
}
