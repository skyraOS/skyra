// Package graph is the durable expectation graph from specs/2026-08-03/graph.md.
//
//	EVERY DURABLE NODE IS AN EXPECTATION
//	EVERY EDGE IS AN OBSERVATION
//
// Expectations and actions are a COMPOSITION, not two types. Following the filesystem:
// an expectation is a DIRECTORY (has entries, can be traversed into, can anchor the
// execution stack) and an action is a FILE (a leaf, is executed, cannot be traversed
// into). Unix has one inode and decides by what the inode contains; so does this.
//
// Nothing here executes. The graph is a place — durable, append-only, and walked by the
// ephemeral handles in handle.go, which are what implement kernel.Space. A process does
// not become the filesystem.
package graph

// NodeID and ObservationID are durable identities. Opaque: compared for equality, never
// parsed, never used to recover position.
type NodeID string
type ObservationID string

// A Node is one inode.
//
// Kind is DERIVED FROM COMPOSITION and never stored — no type tag, not even Unix's mode
// bit. Exactly one of Entries/Act is populated, and which one is what makes this a
// directory or a file:
//
//	Entries populated   → a DIRECTORY  → an EXPECTATION
//	Act populated       → a FILE       → an ACTION
//
// A directory's content is its entries; a file's content is what it does. That is the
// whole of the distinction.
type Node struct {
	ID    NodeID
	Label string

	// Entries is a directory's content — its children, each reached through an
	// observation. Empty on a file.
	//
	// An expectation's entries include at least its own perceive, so a directory is
	// never empty.
	//
	// NOT YET ENFORCED — there is no constructor. It must become one, because the
	// whole point of the universal-perceive rule is that an unfalsifiable expectation
	// is UNCONSTRUCTIBLE rather than merely detectable. A struct literal can currently
	// build one, which means the rule is still a rule.
	Entries []Entry

	// Act is a file's content — the operator this action runs. Nil on a directory.
	//
	// A DIRECTORY IS NEVER EXECUTED. An expectation acts only through the actions it
	// contains, which is why grading arrives entirely through files.
	Act *Act
}

// IsDirectory reports whether this node is an expectation. Derived, so it cannot
// disagree with the structure.
func (n *Node) IsDirectory() bool { return n != nil && n.Act == nil }

// IsFile reports whether this node is an action.
func (n *Node) IsFile() bool { return n != nil && n.Act != nil }

// An Act is what a file does.
type Act struct {
	// Primitive names the world operation. The world's instrument list decides what
	// is measurable, not the model — see cognitive-agent/skyra-dna-v3/affordance.go
	// for the precedent this will reuse.
	Primitive string

	// Perceive marks the falsifying look — the action by which an expectation can be
	// wrong about the WORLD rather than about its own execution.
	//
	// It is a field for now and probably should not stay one. Whether an act perceives
	// is a property of the world's instruments, and deriving it from affordances would
	// make it one less thing anyone can declare about themselves.
	Perceive bool
}

// An Entry is one directory entry: a name, what it points at, and the observation that
// IS the edge.
//
// This is also the expectation → action edge that Q(E,A) lives on, which is what the
// composition bought back — see specs/2026-08-03/questions.md, answered 08-09.
type Entry struct {
	Name   string
	Target NodeID

	// Via is the observation this edge is. There is no edge without one: the structure
	// IS the evidence.
	//
	// OPEN (questions.md Q1): what this observation asserts. A subtask edge and an
	// evidential-ground edge are different relations, and trust.md multiplies along
	// this one, so the answer decides what T, G, and P compute.
	Via ObservationID
}

// An Observation is an edge. Immutable once written.
//
// The observation worker emits TWO numbers and they do different jobs:
//
//	SUPPORT      [-1, 1]   signed. THE ONLY THING THAT ENTERS THE ARITHMETIC.
//	CONFIDENCE   [ 0, 1]   how sure the reading is. ROUTES, NEVER WEIGHTS.
//
// Folding confidence into support as expected-value weighting is the tempting and wrong
// move: novel evidence is exactly what reads as low-confidence, and novel evidence is
// what contradicts established belief. Weighting by it would shrink the mass of
// unfamiliar contradictory observations while looking like principled uncertainty
// handling. See 2026-07-27/trust.md, "Confidence is not a scalar on the edge."
//
// The [-1,1] range on Support is load-bearing twice over: it gives depth attenuation for
// free, and it is what stops chains from amplifying without bound.
type Observation struct {
	ID         ObservationID
	Support    float64
	Confidence float64
	At         int
}
