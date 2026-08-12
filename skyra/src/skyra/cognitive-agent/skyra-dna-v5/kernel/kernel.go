// Package kernel HOLDS EVERYTHING THAT IS HAPPENING. The graph holds only what
// happened.
//
// That split is not a division of labour. It is what makes three claims true
// rather than aspirational:
//
//   - "The kernel writes nothing" holds only under a strict split. If in-flight
//     work lived in the graph you would mark a possibility running and later
//     unmark it — a durable write, performed by the one component forbidden one.
//   - Mint-on-closure needs it. An improvised expectation existing in the graph
//     before it closed would be discoverable mid-flight, and something could be
//     selected that has never proven it can close.
//   - Crash semantics collapse to nothing. The kernel dies, everything in flight
//     evaporates, and the graph is exactly the set of completed commits. No
//     half-written state, no reconciliation pass, no recovery protocol — ever.
//
// The kernel is THE ONLY GLOBAL THING, and only over EXECUTION TOPOLOGY:
//
//	IT HOLDS     roots · heads · ancestry · bindings.   EXECUTION TOPOLOGY.
//	IT CANNOT    read content · judge · assign credit.  NO SEMANTICS.
//
// A scheduler's view, not a world model. Knowing the shape of live chains is not
// knowing their order.
package kernel

import (
	"fmt"
	"sync"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/host"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/trace"
)

// Kernel pairs an engram with an entity, drives the loop, and hands a RECORD to
// the graph. It applies no write and holds no policy.
type Kernel struct {
	G    *graph.Graph
	Reg  *entity.Registry
	Host host.Host
	IDs  *ids.Minter
	Run  *trace.Run

	// Builtin names, present in every frame. discover is nameable everywhere — a
	// builtin name, not something you must discover in order to discover. That is
	// what stops the regress.
	Builtin map[string]*graph.Engram

	// Bounds. A RUNAWAY is depth or breadth past a bound, and it is an EVICTION
	// decision — which requires exactly the view the kernel has and nobody else
	// does. A possibility sees down, an engram sees its own children, no entity
	// spans invocations.
	MaxDepth int
	MaxSteps int

	mu    sync.Mutex
	roots map[ids.SpanID]*Root
	heads map[ids.SpanID]int // live head count per root — the concurrent frontier

	// Engrams given identity that have not run. They are NOT in the graph,
	// because mint-on-closure means an expectation that has never closed must not
	// be discoverable. In-flight things belong in the kernel; this is one.
	unminted map[ids.EngramID]*graph.Engram

	// Counted so the self-maintenance spiral is countable rather than felt.
	userRooted   int
	kernelRooted int
}

func New(g *graph.Graph, reg *entity.Registry, h host.Host, run *trace.Run) *Kernel {
	return &Kernel{
		G: g, Reg: reg, Host: h, Run: run,
		IDs:      &ids.Minter{},
		Builtin:  map[string]*graph.Engram{},
		MaxDepth: 6,
		MaxSteps: 8,
		roots:    map[ids.SpanID]*Root{},
		heads:    map[ids.SpanID]int{},
	}
}

// Ingress puts material into the tree from outside.
//
// INGRESS IS NOT ORIGINATION. Two acts wear the word create, and only one is
// governed by the origination rule: asking whether an inbound observation is
// originated or composed is a CATEGORY ERROR, because no possibility produced it.
// Producer is empty precisely because there is no act of production to describe.
//
// NOTHING GATES THIS (questions.md §5). Any caller is an ingress channel — a
// human, a webhook, a file watcher, structurally identical. Opening a root is the
// maximum-privilege operation in the design and there is no gate on it. The hole
// is left open rather than papered over.
func (k *Kernel) Ingress(c content.Content) *graph.Observation {
	o := &graph.Observation{ID: k.IDs.Observation(), Content: c}
	k.G.Store(o)
	return o
}

// Ask opens a user root and drives it to completion.
//
// THE ROOT'S EXPECTATION is whatever was typed. interfaces.md leaves this open —
// either the root is a special case with none, or its expectation is the ask,
// which makes the human an entity reading natural language. The second is chosen
// here because it needs no special case anywhere else.
// A ROOT IS REUSED WHEN THE SAME ASK RECURS, and that is not a convenience.
//
// FOUND WHILE BUILDING, and recorded because the spec set does not say it
// anywhere. All learning is topological, and topology accumulates on a PARENT'S
// children. If every ask minted a fresh root, the root's child map would be empty
// on arrival every time — so the outermost possibility, the one whose selections
// matter most, would be the only thing in the system that can never learn. Eight
// identical asks would make eight unrelated one-shot trees.
//
// Reuse-by-expectation is the same mechanism by-index already implements, applied
// at ingress: the second time a need arises, you find the engram instead of
// re-improvising it. Nothing new — but it has to be SAID, or an implementation
// gets it wrong in a way that silently disables the central claim.
func (k *Kernel) Ask(ask string, ent ids.EntityID, in []*graph.Observation) (*Root, *graph.Observation, error) {
	span := k.IDs.Span()

	rootEngram := k.rootFor(ask)

	r := &Root{Span: span, Engram: rootEngram, Origin: UserRooted}

	k.mu.Lock()
	k.roots[span] = r
	k.userRooted++
	k.mu.Unlock()

	k.Run.Logf("ROOT %s  %q", span, ask)

	out, err := k.Drive(r, ent, in)

	// Mint-on-closure applies to the root like anything else: it persists because
	// it closed.
	if err == nil {
		k.G.MintRoot(rootEngram)
	}

	k.mu.Lock()
	delete(k.roots, span)
	k.mu.Unlock()

	return r, out, err
}

// rootFor finds the engram for a recurring ask, or gives a new one identity.
//
// It matches on the expectation's expression, which is the coarsest possible
// reading of "the same question" — the ask is the whole of what a root is for.
// Anything finer belongs to the same-question fold over observations, one level
// down, where it already lives.
func (k *Kernel) rootFor(ask string) *graph.Engram {
	for _, e := range k.G.Roots() {
		if e.Expectation.Expression == ask {
			return e
		}
	}
	return graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: ask})
}

// Drive runs an already-opened root to completion.
//
// Separate from Ask because the root must EXIST before anything under it stages
// an effect — which is the chain being the transaction boundary rather than the
// possibility, made concrete: there is nowhere else for a staged effect to live.
func (k *Kernel) Drive(r *Root, ent ids.EntityID, in []*graph.Observation) (*graph.Observation, error) {
	return k.open(r, r.Engram, "", ent, in, k.Reg.Known(), nil, 0)
}

// Accept is the root's synthesis, and it is TWO THINGS ON TWO AXES.
//
//	DISPERSAL         every commit performs it identically. THE ROOT IS NOT SPECIAL.
//	MATERIALIZATION   happens once per chain. THE ROOT IS SPECIAL.
//
// Easy to conflate. Acceptance IS citation — the human is simply the outermost
// parent, running the same operation with a different surface. And it is the same
// event that materializes, which is a better grounding for the one signal the
// whole pressure model rests on.
func (k *Kernel) Accept(r *Root, o *graph.Observation) error {
	k.Run.Logf("ROOT %s ACCEPTED — materializing along the citation chain", r.Span)
	return r.Materialize(k.G, o, k.Host)
}

// Reject is not-citing. Staged effects die with the chain.
func (k *Kernel) Reject(r *Root, why string) {
	k.Run.Logf("ROOT %s REJECTED — %s", r.Span, why)
	r.Discard(k.G, "the root did not accept")
}

// ---------------------------------------------------------------------------
// the loop
// ---------------------------------------------------------------------------

// open pairs an engram with an entity and drives it until it commits.
//
//	Open a possibility · Step(expectation, frame) · on Select, run the group
//	concurrently, await ALL of it, append to Returned, then Monitor each return
//	IN EMITTED ORDER · on Commit, mint the observation and close.
func (k *Kernel) open(
	r *Root,
	e *graph.Engram,
	name string,
	ent ids.EntityID,
	obs []*graph.Observation,
	reach []ids.EntityID,
	ancestry []ids.EngramID,
	depth int,
) (*graph.Observation, error) {

	if depth > k.MaxDepth {
		return nil, fmt.Errorf("runaway: depth %d past bound %d", depth, k.MaxDepth)
	}

	interp, err := k.Reg.Open(ent)
	if err != nil {
		return nil, err
	}

	p := &Possibility{
		Engram: e, Entity: ent, Name: name,
		Observations: obs,
		Reach:        reach,
		Ancestry:     append(append([]ids.EngramID{}, ancestry...), e.ID),
		Delegated:    map[string]*graph.Engram{},
	}

	k.mu.Lock()
	k.heads[r.Span]++
	k.mu.Unlock()
	defer func() {
		k.mu.Lock()
		k.heads[r.Span]--
		k.mu.Unlock()
	}()

	pad := indent(depth)
	k.Run.Logf("%s┌ open %s [%s] via %s", pad, shortExpr(e.Expectation.Expression), e.ID, ent)

	for step := 0; step < k.MaxSteps; step++ {
		f := k.frame(p)

		em, err := interp.Step(e.Expectation, f)
		if err != nil {
			// The crossing failed. This is an ORDINARY FAILED INVOCATION, not a
			// runtime error — observed and recorded on the INVOKER'S engram, and
			// the invoker's selection is what learns from it.
			k.Run.Logf("%s└ ✗ %v", pad, err)
			return nil, err
		}

		switch t := em.(type) {

		case grammar.Select:
			if err := authorizeSelect(f, t); err != nil {
				k.Run.Logf("%s│ ✗ unauthorized select: %v", pad, err)
				return nil, err
			}
			if err := k.runGroup(r, p, f, t, depth); err != nil {
				return nil, err
			}
			// THE FORFEIT, fired at the first select and permanent thereafter.
			p.Selected = true

			// Monitor at GROUP CLOSURE, in EMITTED order — never as members land.
			// Arrival order is a scheduling accident; pushing on arrival makes the
			// entity's context depend on it, breaks replay, and is partial
			// completion arriving as a side effect rather than as a decision.
			//
			// A LEAF NEVER RECEIVES IT — but a possibility that has selected is by
			// definition not a leaf, so the guard is structural here.
			for _, ret := range p.Returned[len(p.Returned)-len(t.Members):] {
				interp.Monitor(ret)
			}

		case grammar.Commit:
			if err := authorizeCommit(f, t, p.Selected); err != nil {
				k.Run.Logf("%s│ ✗ unauthorized commit: %v", pad, err)
				return nil, err
			}
			out, err := k.commit(r, p, t)
			if err != nil {
				return nil, err
			}
			k.Run.Logf("%s└ ✓ %s  %s", pad, out.ID, oneLine(out.Content.Render()))
			return out, nil
		}
	}

	// A STALLED HEAD. Eviction, and the possibility wrote nothing — so nothing
	// happened, in the graph AND in the world, because its staged effects will
	// not be reached by any walk.
	return nil, fmt.Errorf("evicted: %d steps without a commit", k.MaxSteps)
}

// runGroup executes an invocation group. EVERY GROUP AWAITS.
//
// A GROUP IS A BARRIER, and that cost is accepted deliberately: one slow or
// trapped member delays the next selection even if the others have returned
// enough to continue. Whether partial completion is ever necessary is open, and
// it must not be smuggled in as an execution flag without a case that demands it.
func (k *Kernel) runGroup(r *Root, p *Possibility, f frame.Frame, s grammar.Select, depth int) error {
	n := len(s.Members)
	rets := make([]*graph.Observation, n)
	errs := make([]error, n)
	kids := make([]*graph.Engram, n)

	var wg sync.WaitGroup
	for i, m := range s.Members {
		child, _ := f.Resolve(m.Engram) // authorized already
		kids[i] = child

		var args []*graph.Observation
		for _, id := range m.Observations {
			if o, ok := k.G.Observation(id); ok {
				args = append(args, o)
			}
		}

		wg.Add(1)
		go func(i int, m grammar.Member, child *graph.Engram, args []*graph.Observation) {
			defer wg.Done()
			// Reach narrows on the way down: a child gets its parent's reach, and
			// a select may drop from it but can never mint into it.
			rets[i], errs[i] = k.open(r, child, m.Engram, m.Entity, args, p.Reach, p.Ancestry, depth+1)
		}(i, m, child, args)
	}
	wg.Wait()

	// Collected in EMITTED order, never completion order — which is what keeps
	// the barrier real and makes the sequence reconstructible from the durable
	// record, so re-running a possibility rebuilds the same context.
	for i, m := range s.Members {
		ret := rets[i]
		if errs[i] != nil {
			// A failure still returns an observation. It is A CROSSING: Producer
			// is the engram I invoked, Cites is empty — nothing to descend into,
			// and the walk terminates there with no special case.
			ret = &graph.Observation{
				ID:       k.IDs.Observation(),
				Producer: kids[i].ID,
				Content: content.Content{
					"error": content.Str(errs[i].Error()),
				},
			}
			k.G.Store(ret)
		}

		var args []*graph.Observation
		for _, id := range m.Observations {
			if o, ok := k.G.Observation(id); ok {
				args = append(args, o)
			}
		}

		p.Returned = append(p.Returned, ret)
		p.Pending = append(p.Pending, &graph.Pending{
			Child: kids[i],
			Name:  m.Engram,
			Inv: &graph.Invocation{
				Span:         r.Span,
				Observations: args,
				Entity:       m.Entity,
				Host:         k.Host.ID(),
				Cite:         m.Cite,
				Return:       ret,
			},
		})

		// DELEGATION. Reach arriving from an entity I CHOSE to invoke becomes
		// nameable here. An EngramRef in an ARGUMENT or at INGRESS is ignored —
		// that would be injection, and whoever can talk to me would pick what I
		// can touch.
		for _, er := range ret.Content.Reach() {
			// REACH MAY NOT REACH UPWARD. Binding a name that resolves to an
			// ancestor lets a chain re-enter itself; see Possibility.Ancestry.
			// Refused here rather than filtered in discovery, because discovery
			// is a POSSIBILITY and possibilities cannot see their ancestry — only
			// the kernel can.
			if p.wouldRecur(er.ID) {
				k.Run.Logf("%s│ ⊘ refused reach to %s (%q) — it is an ancestor of this chain",
					indent(depth), er.ID, er.Name)
				continue
			}
			if child, ok := k.lookupPending(er.ID); ok {
				p.Delegated[er.Name] = child
			}
		}
	}
	return nil
}

// commit closes the possibility. THE ONLY DURABLE WRITE.
func (k *Kernel) commit(r *Root, p *Possibility, c grammar.Commit) (*graph.Observation, error) {
	var out *graph.Observation

	switch {
	case c.Observed != "":
		// An orchestrator NAMES. The named observation is handed up as-is: the
		// parent sees what the child CONCLUDED.
		named, ok := k.G.Observation(c.Observed)
		if !ok {
			return nil, fmt.Errorf("commit names unknown observation %s", c.Observed)
		}
		// Handing the child's observation up verbatim would make this possibility
		// a PIPE — the parent would reach past it into its subtree. So the
		// composition is its own observation that REFERENCES what it named, which
		// is also what puts the reference in Cites and lets credit descend.
		out = &graph.Observation{
			ID:       k.IDs.Observation(),
			Producer: p.Engram.ID,
			Content:  content.Content{"named": content.Ref(named.ID)},
		}
		// Reach passes through a naming commit — otherwise discover could resolve
		// something and the caller could never name it.
		//
		// ALL OF IT PASSES. Reach() returns a slice, and writing each element to
		// one key handed the caller exactly one — the last, by map iteration
		// order, which is not even stable. A capability the caller was granted
		// simply was not there, with no error, which is the failure class this
		// design refuses everywhere else.
		//
		// A List is enough: Reach() walks nested values, so every EngramRef in it
		// is collected on the receiving side, and runGroup already folds each one
		// into Delegated.
		if reach := named.Content.Reach(); len(reach) > 0 {
			carried := make(content.List, 0, len(reach))
			for _, er := range reach {
				carried = append(carried, er)
			}
			out.Content["reach"] = carried
		}

	default:
		// A LEAF ORIGINATES. Producer is the leaf engram; Cites is whatever it
		// carried forward.
		out = &graph.Observation{
			ID:       k.IDs.Observation(),
			Producer: p.Engram.ID,
			Content:  c.Content,
		}
	}

	rec := &graph.Record{
		Span:        r.Span,
		Engram:      p.Engram,
		Name:        p.Name,
		Invocations: p.Pending,
		Return:      out,
	}
	if err := ownership(rec); err != nil {
		return nil, err
	}

	// MINT ON CLOSURE. It persists because it closed — which is what makes the
	// index safe to discover from: nothing in it is a promise.
	//
	// A ROOT PERSISTS WITHOUT BECOMING DISCOVERABLE. It is an ask, not a
	// capability — see graph.roots. Note that this is the SECOND path into the
	// index, and the first attempt at the fix only closed the other one: Ask
	// routed the root correctly and this line put it back, so by-index went on
	// proposing the root for two more runs. Mint-on-closure is written in the set
	// as one rule, and it turns out to have two call sites.
	if p.Engram == r.Engram {
		k.G.MintRoot(p.Engram)
	} else {
		k.G.Mint(p.Engram)
	}

	// The kernel hands the record over. THE GRAPH APPLIES IT. If an interpreter
	// performed the write it could omit it; if the kernel did, "writes nothing"
	// would be a policy of restraint rather than an incapacity.
	k.G.Apply(rec)

	return out, nil
}

// ---------------------------------------------------------------------------

func (k *Kernel) frame(p *Possibility) frame.Frame {
	return &frame.Live{
		Engram:    p.Engram,
		Ent:       p.Entity,
		Reach:     p.Reach,
		Obs:       p.Observations,
		Rets:      p.Returned,
		Impulses:  p.Impulses,
		Delegated: p.Delegated,
		Builtin:   k.Builtin,
		G:         k.G,
	}
}

// lookupPending finds an engram that may not be in the index yet.
//
// An engram created this chain has NOT been minted — mint-on-closure means it
// persists only if it closes — so it cannot be found in the graph. It is held
// here, in ephemeral kernel state, exactly where in-flight things belong.
func (k *Kernel) lookupPending(id ids.EngramID) (*graph.Engram, bool) {
	if e, ok := k.G.Engram(id); ok {
		return e, true
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	e, ok := k.unminted[id]
	return e, ok
}

// Hold registers an engram that has been given identity but has not run.
func (k *Kernel) Hold(e *graph.Engram) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.unminted == nil {
		k.unminted = map[ids.EngramID]*graph.Engram{}
	}
	k.unminted[e.ID] = e
}

// Introspect answers what is running, from the only view that has it.
//
// BOUNDED BY THE SUBTREE, not by the grammar. Never upward, so a possibility
// cannot learn its own reception; never sideways, so there is no coordination in
// flight.
func (k *Kernel) Introspect(from ids.SpanID) (heads int, ok bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	h, ok := k.heads[from]
	return h, ok
}

// Spiral reports whether kernel-rooted work is outgrowing user-rooted work.
// Countable rather than felt.
func (k *Kernel) Spiral() (user, kernelRooted int) {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.userRooted, k.kernelRooted
}

func indent(d int) string {
	s := ""
	for i := 0; i < d; i++ {
		s += "│ "
	}
	return s
}

func shortExpr(s string) string {
	s = oneLine(s)
	if len(s) > 60 {
		s = s[:60] + "…"
	}
	return s
}

func oneLine(s string) string {
	out := []rune{}
	space := false
	for _, r := range s {
		if r == '\n' || r == '\t' || r == ' ' {
			space = true
			continue
		}
		if space && len(out) > 0 {
			out = append(out, ' ')
		}
		space = false
		out = append(out, r)
	}
	s = string(out)
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}
