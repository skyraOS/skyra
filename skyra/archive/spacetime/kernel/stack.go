package kernel

// A Stack is one execution, reified.
//
// The recursive kernel kept its frames on Go's own stack, which made a wave atomic: once
// started it ran to completion, and nothing could look at it, interleave with it, or stop
// it partway. That was fine while a wave was a single short traversal.
//
// It stops being fine the moment an agent holds SEVERAL live executions and advances them
// one at a time. Interleaving requires the stack to be a value the runner can step, and
// eviction requires it to be a value the runner can drop. So the frames move off the Go
// stack and into here.
//
// This is what specs/2026-08-03/questions.md Q3 called the trampoline. Note what it does
// NOT buy on its own: a Stack is inspectable and droppable, but not yet serializable, so
// an execution still dies with its process.
type Stack struct {
	// WaveID identifies this execution. It is what a crossing records when one agent
	// spawns a wave at another — the edge of the wave TREE, since a wave never spans
	// agents and correlation has to be explicit rather than inferred.
	WaveID string

	origin  Space
	impulse any

	frames []frame
	hops   []Hop

	// realized is the root's concrete return, set when the last frame pops.
	realized Space
}

// A frame is one position in a live execution — the Space, and the child realization it
// is waiting to fold.
type frame struct {
	space Space
	stack *Stack

	// hop indexes this frame's entry in Stack.hops, so Realized can be stamped when
	// the frame's Observe finally returns.
	hop int

	// incoming is the value a just-finished child realized, waiting to be folded by
	// this frame's Observe. present says whether there is one, since nil is a
	// legitimate realization and must not read as "nothing arrived."
	incoming Space
	present  bool

	// exhausted records that Perturb already returned nil, so the pending Observe is
	// a step of its own. Without it a leaf would call Perturb AND Observe in one
	// step, making some steps twice the work of others — which a scheduler has no way
	// to see and no way to account for.
	exhausted bool

	// value is this frame's latest Observe return — the fold so far. A frame that
	// yielded several children has Observed once per child, and the LAST return is
	// what it hands up.
	value Space

	// folded says whether Observe has run at least once. A frame that yielded nothing
	// still gets one Observe with a nil realization, so a leaf may say what it is;
	// a frame that yielded must NOT get that extra call.
	folded bool
}

// NewStack opens an execution at root. The root's hop is recorded immediately: it is a
// Space the kernel is committed to running, and law 4 says that is recorded before it
// runs, not after it succeeds.
func NewStack(waveID string, origin, root Space, impulse any) *Stack {
	s := &Stack{WaveID: waveID, origin: origin, impulse: impulse}
	if root != nil {
		s.push(root)
	}
	return s
}

func (s *Stack) push(sp Space) {
	// Depth and parent are STORED again. They were derivable while one Perturb meant
	// one child and the transcript was a chain; an iterating Perturb makes it a tree,
	// and position no longer carries either.
	parent, depth := -1, 0
	if n := len(s.frames); n > 0 {
		parent, depth = s.frames[n-1].hop, s.frames[n-1].depth()+1
	}
	s.hops = append(s.hops, Hop{SpaceID: sp.ID(), Parent: parent, Depth: depth})
	s.frames = append(s.frames, frame{space: sp, hop: len(s.hops) - 1, stack: s})
}

func (f frame) depth() int { return f.stack.hops[f.hop].Depth }

// Done reports whether this execution has finished.
func (s *Stack) Done() bool { return s == nil || len(s.frames) == 0 }

// Depth is how many frames are live. Useful for a scheduler; carries no meaning to the
// kernel itself.
func (s *Stack) Depth() int {
	if s == nil {
		return 0
	}
	return len(s.frames)
}

// Transcript is the record of what this execution ran. Valid at any point, not only at
// the end — a stack evicted partway still says honestly how far it got.
func (s *Stack) Transcript() Transcript {
	if s == nil {
		return Transcript{}
	}
	return Transcript{WaveID: s.WaveID, Hops: s.hops}
}

// Realized is the concrete Space this execution produced, valid once Done. It is the
// root's Observe return, and reaching it is the assembly boundary's privilege — see
// Relation.RealizeConcrete.
func (s *Stack) Realized() Space {
	if s == nil {
		return nil
	}
	return s.realized
}

// Step advances this execution by exactly ONE call into a Space and returns whether it
// did any work. False means finished.
//
// One Step is one of four things, and never more than one:
//
//	FOLD      a child finished; hand its realization to this frame's Observe
//	DESCEND   Perturb promoted a continuation; push it
//	EXHAUST   Perturb promoted nothing; mark the base case
//	FINISH    pop, and hand this frame's value up
//
// A frame is asked for a continuation REPEATEDLY — Perturb, fold, Perturb again — until it
// promotes nothing. So a Space must eventually return nil from Perturb; one that keeps
// handing back the same child never terminates. Under the old one-shot Perturb that was
// free, and it is now the Space's responsibility.
//
// EXHAUST and FINISH are separate on purpose. A Space's Perturb may do arbitrary work, so
// folding it into the same step as Observe would make a leaf cost two calls while every
// other transition costs one — an unevenness the scheduler cannot see and cannot account
// for. Uniform granularity is what makes round-robin fair.
func (s *Stack) Step() bool {
	if s.Done() {
		return false
	}
	i := len(s.frames) - 1
	top := &s.frames[i]

	// FOLD — a child's realization is waiting. The frame does NOT pop: it will be
	// asked for another continuation, and that next Perturb sees whatever this Observe
	// just wrote. That re-entry is what lets one frame run a pipeline — set up, fold,
	// then act on what the fold produced.
	if top.present {
		realized := top.incoming
		top.incoming, top.present = nil, false
		top.value = top.space.Observe(s.origin, s.impulse, realized)
		top.folded = true
		s.hops[top.hop].Realized = top.value != nil
		return true
	}

	// FINISH — Perturb has promoted nothing further.
	if top.exhausted {
		value := top.value
		if !top.folded {
			// yielded nothing at all: Observe still runs with a nil realization,
			// so a leaf gets to say what it concretely is.
			value = top.space.Observe(s.origin, s.impulse, nil)
			s.hops[top.hop].Realized = value != nil
		}
		s.finish(i, value)
		return true
	}

	// DESCEND, or EXHAUST.
	if next := top.space.Perturb(s.origin, s.impulse); next != nil {
		s.hops[top.hop].Promoted = true
		s.push(next)
		return true
	}
	top.exhausted = true
	return true
}

// finish pops frame i and hands its value to the frame beneath, or keeps it as the
// execution's result when there is nothing beneath.
func (s *Stack) finish(i int, value Space) {
	s.frames = s.frames[:i]
	if len(s.frames) == 0 {
		s.realized = value
		return
	}
	parent := &s.frames[len(s.frames)-1]
	parent.incoming, parent.present = value, true
}

// Run advances to completion. Convenience for callers with one execution and no reason
// to interleave.
func (s *Stack) Run() {
	for s.Step() {
	}
}
