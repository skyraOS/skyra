package graph

import "skyra-v1/src/skyra/spacetime/kernel"

// A Handle is one ephemeral position in the graph — the thing that actually implements
// kernel.Space.
//
// It is Unix's open file description, near exactly. `open()` does not hand you an inode;
// it hands you a per-process object coupling durable data to a way of using it, plus the
// transient state of this particular use:
//
//	inode   →  Node           durable, shared, agent-neutral
//	mode    →  Behavior       chosen by the caller at open time, NEVER by the file
//	offset  →  Accumulation   transient, per-use, dies with the handle
//
// The durable Node deliberately does NOT implement Space. Execution is a traversal OVER
// the graph, not a property of it (specs/2026-08-03/execution.md): the filesystem exists
// independently of any process, and a process does not become the filesystem. So a Handle
// REFERENCES a Node and never copies it — the same discipline dnav2's EphemeralRoot used
// ("reference + select, never copy the genome").
//
// # A Handle is not the workers' FRAME
//
// 2026-07-27/workers.md uses "frame" for what a worker can look at. That is a different
// thing, and the two are related by DERIVATION rather than identity:
//
//	HANDLE   a POSITION — where execution is standing
//	FRAME    the NEIGHBOURHOOD that position exposes — what is in view from here
//
// A handle's position determines the attention frame. It is not one.
// The Handle does NOT hold the graph. An open file description does not carry the
// filesystem, and the graph is part of the CONTEXT the agent supplies — so it travels
// with the behavior, which is the half chosen per agent. Keeping it here would put
// ambient reach over the whole durable store on every value that crosses the kernel.
type Handle struct {
	node     *Node
	behavior Behavior

	// Scratch is behavior-private per-frame state — a phase cursor, a walk position.
	// It lives here because a Behavior is shared across an agent and must hold no
	// per-execution state of its own. Ephemeral: it dies with the handle.
	Scratch any

	// Residual is what this frame could not resolve — the nodes it reached and could not
	// place. Collected during the wave, SPENT AT THE BOUNDARY afterward, because a wave
	// never calls the model.
	Residual []*Node

	// Frontier is how many candidates were in view — the denominator the gate reads.
	Frontier int

	// Gate is what the gap condition WOULD have said. Recorded, never enforced.
	Gate GateVerdict

	// WaveID is the execution this handle belongs to, carried for instrumentation.
	WaveID string

	// Rejects are writes that were refused, kept so a run can report them rather than
	// swallowing them.
	Rejects []string

	// Accumulation is this handle's derived arithmetic, gathered on ascent from what
	// its children realized. It is this position's locals: transient, and it dies here.
	//
	// Nothing derived is ever written back. T, G, |O|, P, and coherence are properties
	// of the support graph, computed on demand and discarded (2026-07-27/trust.md). A
	// derived magnitude landing on the durable side is the one thing that would give
	// back the property that dissolved concurrency.
	Accumulation Accumulation
}

// compile-time: a Handle is a Space. The durable Node is not, and must not become one.
var _ kernel.Space = (*Handle)(nil)

// A Behavior is the methods half of a Space.
//
//	SPACE CREATION IS THE COUPLING OF DATA TO METHODS, IN A CONTEXT SPECIFIC TO THE
//	AGENT BEING CREATED.
//
// Behavior is therefore NOT a property of the node and not a property of the handle type.
// It is supplied at coupling time, which is what lets the cognitive agent, a world, and
// a user agent walk the same durable graph and do entirely different things over it —
// one filesystem, many programs, not merely many processes.
//
// This is dnav2's Transcribe generalized: that selected which genes to express for one
// wave; this selects which methods are coupled to this data for this walk.
//
// # The rail
//
//	THE COUPLING IS CHOSEN BY THE AGENT. NEVER BY THE DATA.
//
// A node that could influence which behavior gets coupled to it would be selecting its
// own transformer, which is the guarantee 2026-07-26/definitions.md §0a exists to hold —
// and the epistemic form of the kernel's own discipline. Nothing on Node may reach this
// choice.
type Behavior interface {
	// ID identifies this behavior for the transcript. Two waves over the same node
	// with different behavior are different events, and the record has to say so.
	//
	// The kernel's Hop does not carry this yet — see specs/2026-08-03/questions.md Q7.
	ID() string

	Perturb(h *Handle, origin kernel.Space, impulse any) kernel.Space
	Observe(h *Handle, origin kernel.Space, impulse any, realized kernel.Space) kernel.Space
}

// Open couples one durable node to one behavior for one walk. Neither half is modified;
// the Handle is the coupling, and it dies with the wave.
//
// This is the ENTRY POINT into the graph. An agent opens a node with its own behavior and
// hands the result to the kernel; the graph itself never starts anything. Opening
// Graph.Root is the ordinary case — see Graph.Root for why that node is special and why
// nothing forces you to start there.
func Open(n *Node, b Behavior) *Handle {
	return &Handle{node: n, behavior: b}
}

// Node is the durable node this handle is standing at.
func (h *Handle) Node() *Node { return h.node }

// Behavior is the methods this handle was coupled with.
func (h *Handle) Behavior() Behavior { return h.behavior }

func (h *Handle) ID() string {
	if h == nil || h.node == nil {
		return ""
	}
	return string(h.node.ID)
}

// Perturb answers "which executable Space should the kernel run next," and delegates to
// the coupled behavior. The Handle owns the position; the behavior owns what to do from
// it.
//
// NO BEHAVIOR IS WRITTEN YET. What one has to do is the genuinely undecided piece: an
// expectation has many entries, but Perturb returns exactly ONE Space or nil, so a
// directory must return an ephemeral GROUPING handle that privately walks its own
// entries. That grouping — sequence, fan-out, join — lives inside the Space by
// construction (kernel.md §9, and 2026-07-27/kernel-alteration.md §13 Q3). Its shape is
// open.
//
// A file returns nil: an action is a leaf, and nil is the descent base case rather than a
// verdict about it.
//
// An uncoupled handle is inert rather than an error. It is a position with no program.
func (h *Handle) Perturb(origin kernel.Space, impulse any) kernel.Space {
	if h == nil || h.behavior == nil {
		return nil
	}
	return h.behavior.Perturb(h, origin, impulse)
}

// Observe answers "what concrete Space did this perturbed Space realize," and delegates
// to the coupled behavior. This is the ascent fold: take what the continuation realized,
// gather it into h.Accumulation, return the concrete counterpart. Post-order, on the
// stack — the return path IS the tree (kernel.md §3), so there is nothing to assemble.
func (h *Handle) Observe(origin kernel.Space, impulse any, realized kernel.Space) kernel.Space {
	if h == nil || h.behavior == nil {
		return nil
	}
	return h.behavior.Observe(h, origin, impulse, realized)
}

// An Accumulation is the derived arithmetic for one node, gathered during a wave and
// thrown away with it. STORED NOWHERE.
//
//	T = Σ|c|      MASS        ranks. a well-refuted claim is as worth attending to as a
//	                          well-supported one, and a contested one more than either.
//	G = Σ c       DIRECTION   which way it leans, net.
//	|O|           COUNT       distinct grounding observations.
//	P = G / |O|   POSITION    how much of the change has closed.
//
// c is the product of edge supports along a path. Distinct OBSERVATIONS, never distinct
// paths — an observation reaching a node by several routes counts once, at its strongest
// route, selected on |∏ s(e)| with the sign preserved. A signed max would pick −0.2 over
// −0.8 and take the weaker grounding.
type Accumulation struct {
	T     float64 // Σ|c|
	G     float64 // Σc
	Count int     // |O|
}

// P is the position — how much of the change has happened.
//
// TENTATIVE. The mean is not monotone in evidence: adding nine weak but POSITIVE
// observations to one strong one drops P from 0.900 to 0.135, so it punishes breadth.
// The recorded-not-adopted replacement is to squash the SIGNED sum — tanh(G) for a
// 0-centred scale — which changes only this function. See 2026-07-27/truth.md.
func (a Accumulation) P() float64 {
	if a.Count == 0 {
		return 0
	}
	return a.G / float64(a.Count)
}

// Coherence is |G|/T ∈ [0,1] — 1 when all evidence agrees, 0 when it perfectly cancels.
// It resolves an ambiguity P alone cannot: a claim at the midpoint is CONTESTED (T high,
// coherence low) or UNEXAMINED (T low).
//
// Undefined at T = 0, and that is a real hole rather than a rounding concern — dynamics
// keys on coherence and Toy 5 turns on it. Reported here as (value, ok) so a caller must
// decide what an ungrounded expectation means rather than silently reading 0, which would
// say "evidence perfectly cancels" about something nothing has borne on.
func (a Accumulation) Coherence() (float64, bool) {
	if a.T == 0 {
		return 0, false
	}
	g := a.G
	if g < 0 {
		g = -g
	}
	return g / a.T, true
}
