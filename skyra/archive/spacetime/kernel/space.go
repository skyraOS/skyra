// Package kernel is the two-method substrate from specs/2026-07-27/kernel-alteration.md.
//
// Eight methods to two, both closed over Space. Descent and ascent never leave the
// algebra, and the kernel cannot classify what it walks because it has no method that
// would let it.
//
// It is deliberately a NEW package rather than an edit to spacetime/space.go: the
// eight-method contract is load-bearing for world, world3d, and relation, and those
// keep working while this is built beside them. Same posture traverse/ took ("the
// collapse onto skyra.Space is Phase 3").
package kernel

// Space — the whole traversal contract.
//
//	Perturb    returns one ephemeral executable Space, or nil
//	Observe    receives one realized Space and returns one concrete Space, or nil
//
// The kernel never receives a map, slice, iterator, or tuple of children. A Space owns
// its grouped data AND the behavior governing that grouping, so returning a collection
// would force the kernel to assign meaning to it — which is the one thing it must not
// do (§7).
//
// nil from Perturb is the descent base case. It means "I PROMOTED NOTHING" — a fixed
// point of the promotion, never a judgement about what kind of thing the target is.
// The kernel must never ask "is this a resource?"; that is a classification.
type Space interface {
	// ID is the transcript identity. It is the ONLY thing the kernel reads off a
	// Space, and it reads it to record, never to decide.
	ID() string

	Perturb(origin Space, impulse any) Space
	Observe(origin Space, impulse any, realized Space) Space
}

// Time runs waves. Realize stays VOID for its cognitive result: a returned outcome
// would be a second, chosen results channel — the caller collecting what its own send
// produced — which the Perception Invariant forbids. The concrete Space that ascends
// out of the root goes to the payload log or returns as an involuntary perception
// wave; it is never handed back to the sender.
//
// What Realize DOES return is the Transcript, and that is not a results channel. See
// the Transcript doc for the line between the two and why it has to hold.
type Time interface {
	ID() string
	Realize(origin, traversed Space, impulse any) Transcript
}

// A Transcript is the kernel's record of what it executed: the authoritative answer to
// "which Spaces actually ran," returned to the caller instead of written to a slice.
//
// # Why the kernel returns this rather than the Space writing it
//
// Runtime Spaces disappear when a wave completes, so something must record them before
// they go. That something may not be the Space itself. A Space that records its own
// execution can OMIT ITSELF, and then the audit record is self-reported — the same
// defect confidence has in FLAGS.md ("the one number in the system with no blind
// assessor"), except here it is the record you would use to check everything else.
//
// The split that keeps it honest:
//
//	the Space controls WHAT it returns
//	the kernel controls THAT the return is recorded
//
// The kernel is the only thing that saw every Perturb return, so it is the only honest
// witness. Returning the record preserves that witness without giving the kernel write
// access to anything.
//
// # THE RAIL — this must never carry content
//
//	FORBIDDEN   the caller collects the cognitive OUTCOME of its own send
//	FINE        the caller collects a mechanism record that the wave ran, and where
//
// The test: could a being use this to read back what it thought, bypassing perception?
// IDs and booleans, no. A payload, a Content(), or an Observe return, YES — and at that
// moment this becomes a results channel and the Perception Invariant is broken. Every
// field here is a string, an int, or a bool, and that is a constraint on the type rather
// than a description of it.
//
// # It is a SPINE, not a tree — and the spine is a CHAIN
//
// §10 asks the transcript to carry "grouped member IDs and boundaries." Law 8 says the
// kernel never enumerates grouped members. Both cannot hold, so this holds the second:
// the kernel reports only what it was HANDED. Composition inside a grouping Space is
// invisible here by construction. If member detail is ever needed, the Space has to
// volunteer it, and that field alone goes back to being self-reported — worth marking
// as an explicit exception rather than blurring this now.
//
// It is a TREE. Perturb still returns one Space per call, but it is called REPEATEDLY on a
// frame, so one frame can promote several children in sequence. Depth and parentage are
// therefore stored on the Hop rather than derived from position — a cost paid deliberately
// to let one frame run a pipeline (set up, fold, act on what the fold produced).
type Transcript struct {
	// WaveID identifies the execution this record describes.
	WaveID string

	// Hops are in execution order: index 0 is the root, and index IS depth.
	Hops []Hop
}

// A Hop is one Space the kernel executed. Identity, position, and the two facts only the
// kernel was in a position to witness.
type Hop struct {
	SpaceID string

	// Parent is the index of the hop that promoted this one; -1 at the root.
	// Depth is how far down it sits.
	//
	// STORED, and they were not always. While one Perturb meant one child the transcript
	// was a chain and both fell out of position — but an iterating Perturb makes it a
	// TREE, and a tree needs them written down. Storing a derivation is a cost; this is
	// what it is being spent on.
	Parent int
	Depth  int

	// Promoted is whether this Space's Perturb returned a non-nil continuation.
	// False is the descent base case, NOT a failure.
	Promoted bool

	// Realized is whether this Space's Observe returned a non-nil concrete Space.
	// A nil realization is a legitimate outcome — an observed omission does not
	// erase the corresponding member from the transcript or declare it false (§10).
	Realized bool
}

// Executed is how many Spaces this wave actually ran.
func (t Transcript) Executed() int { return len(t.Hops) }

// Depth is how far the wave descended before hitting its base case.
func (t Transcript) Depth() int {
	d := 0
	for _, h := range t.Hops {
		if h.Depth > d {
			d = h.Depth
		}
	}
	return d
}

// ParentOf is the Space that promoted the hop at i. The root has no parent.
func (t Transcript) ParentOf(i int) (Hop, bool) {
	if i < 0 || i >= len(t.Hops) || t.Hops[i].Parent < 0 {
		return Hop{}, false
	}
	return t.Hops[t.Hops[i].Parent], true
}
