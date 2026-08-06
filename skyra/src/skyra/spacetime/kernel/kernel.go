package kernel

// Relation is the kernel. One method, and it does five things:
//
//  1. calls Perturb
//  2. records the returned ephemeral Space
//  3. executes that Space when it is non-nil
//  4. calls Observe with the realized Space that execution returned
//  5. returns the resulting concrete Space upward
//
// It is STRUCT-AGNOSTIC and stays that way. Nothing in this file mentions an
// expectation, an action, a support value, or a grade — it moves Spaces and writes down
// IDs. That is §3 ("the kernel does zero cognitive math") held by construction rather
// than by discipline: there is no method on Space that would let it enumerate children,
// interpret grouping, compare magnitudes, read a cost, or order a frontier, so it cannot
// do those things even by accident.
//
// What it does NOT do, which the eight-method kernel in ../relation did:
//
//	build the result tree      the return path IS the tree. Observe returns a Space,
//	                           so the fold happens on the stack (§6). Node,
//	                           TraversalPair, Survivors, and the tree double-buffer
//	                           have no counterpart here.
//	score the frontier         scoreChildren divided amplitude across candidates and
//	                           thresholded them. That is cognition and it moved into
//	                           Perturb, where the Space owns it (§5).
//	write a slice              the transcript is RETURNED. see Transcript.
type Relation struct {
	id string
}

func NewRelation(id string) *Relation { return &Relation{id: id} }

func (r *Relation) ID() string { return r.id }

// Realize runs one wave and returns the mechanism record of what ran.
//
// The concrete Space that ascends out of the root is deliberately DROPPED here rather
// than returned: handing it to the caller would be the chosen results channel the
// Perception Invariant forbids. Routing it — to the payload log, or onward as an
// involuntary perception wave — is the caller's job at the assembly boundary, not the
// kernel's.
func (r *Relation) Realize(origin, traversed Space, impulse any) Transcript {
	if traversed == nil {
		return Transcript{}
	}
	s := NewStack("", origin, traversed, impulse)
	s.Run()
	return s.Transcript()
}

// RealizeConcrete is Realize for callers that ARE the assembly boundary — the one place
// entitled to the root's concrete return, because it is what writes that return to the
// payload log as the exact grounding of a call (§6).
//
// It is a separate method on purpose. Realize is what a cell calls, and it cannot reach
// the result; this is what the boundary calls. Keeping them apart makes "who may see the
// outcome" a property of which method you can name, rather than a rule to remember.
func (r *Relation) RealizeConcrete(origin, traversed Space, impulse any) (Space, Transcript) {
	if traversed == nil {
		return nil, Transcript{}
	}
	s := NewStack("", origin, traversed, impulse)
	s.Run()
	return s.Realized(), s.Transcript()
}
