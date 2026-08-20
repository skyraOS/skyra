package graph

import (
	"fmt"

	"skyra-v1/src/skyra/spacetime/kernel"
)

// The cognitive agent's behaviors, per specs/2026-08-03/behaviors.md.
//
//	PERTURB   READS.  builds a frontier, hands off a Walk. writes nothing.
//	OBSERVE   WRITES. commits, and folds what came back.
//
// A worker never traverses: one Perturb yields one Space, so it builds the frontier and a
// Walk visits it. See phases.md for why the read/write split is load-bearing rather than
// tidy — eviction depends on it.

// Workers is one agent's behavior set. The Turn behavior sequences the others, so this is
// what an agent is handed at construction.
type Workers struct {
	Turn      Behavior
	Interpret Behavior
	Expect    Behavior
	Pursue    Behavior

	// Visit is what a Walk opens its MEMBERS with. It is a leaf: it promotes nothing.
	//
	// This is not a detail. Opening members with the turn behavior makes every member
	// run a full nested turn — the walk recurses, the tree explodes, and EXPECT fires
	// once per nested turn. A walk VISITS its frontier; it does not re-enter it.
	Visit Behavior
}

// NewWorkers wires a standard cognitive agent. The namespace is supplied here — behaviors
// reach exactly what they were constructed with, and nothing more.
func NewWorkers(ns *Namespace) *Workers {
	w := &Workers{
		Interpret: &interpret{ns: ns},
		Expect:    &expect{ns: ns},
		Pursue:    &pursue{},
		Visit:     &visit{},
	}
	w.Turn = &turn{ns: ns, w: w}
	w.Interpret.(*interpret).w = w
	return w
}

// ── TURN ────────────────────────────────────────────────────────────────────────────
//
// The sequencer. It has no cognition of its own: it decides WHICH PHASE runs next, and
// each phase is a worker over the same node with different methods.
//
// Ordering lives here, in the structure, and not in the scheduler. Round-robin decides
// which execution steps next; it must never decide which cognition happens first.
type turn struct {
	ns *Namespace
	w  *Workers
}

func (t *turn) ID() string { return "turn/v1" }

// Perturb yields one phase per call. The frame is re-entered after each fold, so EXPECT
// runs against a graph INTERPRET has already written to — which is the whole reason the
// kernel's Perturb iterates.
//
// UNGATED (specs/2026-08-03/behaviors.md, 08-05): EXPECT runs on every impulse, gap or no
// gap. The gate is evaluated and recorded, never enforced, because a condition nobody has
// watched fire cannot be calibrated.
func (t *turn) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space {
	phase, _ := h.Scratch.(int)
	h.Scratch = phase + 1

	switch phase {
	case 0:
		return t.ns.OpenWith(h.Node(), t.w.Interpret)
	case 1:
		return t.ns.OpenWith(h.Node(), t.w.Expect)
	case 2:
		return t.ns.OpenWith(h.Node(), t.w.Pursue)
	}
	return nil // promotes nothing: the turn is over
}

// Observe folds a finished phase. When INTERPRET comes back it also RECORDS THE GATE —
// what the gap condition would have said, unused. Running ungated shows how EXPECT behaves;
// recording the verdict is the only way to learn whether the condition was any good, and it
// is not recoverable after the fact.
func (t *turn) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	r, ok := realized.(*Handle)
	if !ok || r == nil {
		return h
	}
	h.Accumulation.absorb(r.Accumulation)

	if r.Behavior() == t.w.Interpret {
		h.Gate = t.ns.recordGate(h.WaveID, h.Node().ID, r.Frontier, len(r.Residual))
		h.Residual = append(h.Residual, r.Residual...)
	}
	return h
}

// ── INTERPRET ───────────────────────────────────────────────────────────────────────
//
// impulse → support edges. It GRADES, and it writes no nodes.
//
// Revised 08-06 (specs/2026-08-03/families.md): the interpretation node is gone. A stored
// reading is a claim about the world, and therefore state — which would make the model a
// writer of state. EXPECT is the only writer of state.
//
// So this may write EDGES and never NODES. The rule is checkable: anything calling
// AddExpectation is the expectation worker, and nothing else may.
type interpret struct {
	ns *Namespace
	w  *Workers
}

func (i *interpret) ID() string { return "interpret/v1" }

// Perturb builds the frontier from where the handle is standing and hands off a Walk. It
// reads; it does not traverse and does not write.
func (i *interpret) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space {
	if _, done := h.Scratch.(*Walk); done {
		return nil // the frontier has been walked once; promote nothing further
	}
	frontier := i.ns.children(h.Node())
	w := NewWalk(h.ID()+"/interpret-walk", i.ns, i.w.Visit, frontier...)
	h.Scratch = w
	return w
}

// Observe receives the COMPLETE candidate set in one piece — the right shape for this job,
// since a reading should be decided against everything gathered rather than committed
// incrementally.
//
// NOT YET WRITING. And for a RAW payload there is nothing to write: matching is a comparison,
// a raw payload has no structure to compare, and giving it one is a model call. So the whole
// frontier coming back as residual is the correct answer for this input class — not a
// placeholder for missing logic.
//
// The exact pass only has work to do on STRUCTURED impulses — world returns and boundary
// answers, which arrive in a known vocabulary. That is the narrow remainder of questions.md Q10.
func (i *interpret) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	if w, ok := realized.(*Walk); ok && w != nil {
		h.Accumulation.absorb(w.Accumulation)
		h.Frontier = w.Len()
		h.Residual = append(h.Residual, w.Visited...)
	}
	return h
}

// ── EXPECT ─────────────────────────────────────────────────────────────────────────
//
// The expectation worker: CREATES, REFINES, BINDS, and AUTHORS the confidence that a gap
// will close. It alters the TREE and never the stack.
type expect struct{ ns *Namespace }

func (p *expect) ID() string { return "expect/v1" }

// Perturb finds the attachment point. It reads only — the growth happens in Observe, so an
// evicted turn has grown nothing.
func (p *expect) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space {
	return nil
}

// Observe grows the tree. AddExpectation creates the perceive with it, so an unfalsifiable
// expectation is not rejected here — it is unconstructible.
//
// UNGATED. With no boundary and no model, what it proposes is a placeholder named after the
// impulse. That is deliberately thin: the point of the first runs is to watch EXPECT fire,
// not to believe what it produces.
func (p *expect) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	p.ns.expectSeq++
	name := fmt.Sprintf("expect-%d", p.ns.expectSeq)
	_, err := p.ns.AddExpectation(h.Node().ID, name, fmt.Sprintf("%v", impulse),
		Act{Primitive: "perceive_" + name},
		Observation{ID: ObservationID(h.ID() + "#" + name)})
	if err != nil {
		h.Rejects = append(h.Rejects, err.Error())
	}
	return h
}

// ── PURSUE ──────────────────────────────────────────────────────────────────────────
//
// STUBBED — 08-05. Pursuit walks the tree and fires actions at the leaves, on ascent, which
// is deepest-first and matches expectations.md's ordering rule.
//
// It is stubbed because EXPECT is currently ungated: with both live the being would ACT on
// fabricated structure rather than merely storing it. Watching what gets proposed comes
// first.
type pursue struct{}

func (p *pursue) ID() string { return "pursue/stub" }

func (p *pursue) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space {
	return nil
}

func (p *pursue) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	return h
}

// children resolves a node's entries to nodes, skipping any the index cannot resolve.
func (n *Namespace) children(node *Node) []*Node {
	if n == nil || node == nil {
		return nil
	}
	out := make([]*Node, 0, len(node.Entries))
	for _, e := range node.Entries {
		if c, ok := n.nodes[e.Target]; ok {
			out = append(out, c)
		}
	}
	return out
}

// ── VISIT ───────────────────────────────────────────────────────────────────────────
//
// What a Walk opens its members with. A LEAF: it promotes nothing, so visiting a frontier
// never turns into traversing it.
//
// Without this, a walk opens members with whatever the namespace's default behavior is —
// the turn — and each member runs a nested turn of its own. The transcript shows it
// immediately as runaway depth.
type visit struct{}

func (v *visit) ID() string { return "visit/v1" }

func (v *visit) Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space { return nil }

func (v *visit) Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	return h
}
