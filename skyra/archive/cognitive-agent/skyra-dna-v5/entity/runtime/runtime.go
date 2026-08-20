// Package runtime is the entity that mans system possibilities.
//
//	AS KERNEL   pairs, drives, applies writes.   NOT SELECTABLE.
//	AS ENTITY   interprets system expectations.  SELECTABLE, and emits like anything else.
//
// THIS IS A CONSTRAINT ON THE CODE, NOT A NOTE. If a possibility that selects the
// runtime-entity can reach the write path, it can influence what is recorded
// about itself — the one thing the honest-witness split exists to prevent.
//
// It is enforced here by what this package can even SAY: it holds an Index, which
// is read-only by construction. There is no *graph.Graph in this package and
// therefore no Apply, no Mint, no Disperse. The separation is a type, not a
// promise.
package runtime

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/trace"
)

// Index is everything the entity role may see of the graph. Read-only, and
// deliberately narrow.
type Index interface {
	Engrams() []*graph.Engram
}

// Author reaches a model to author an expectation. Supplied rather than
// constructed here, so this package holds no inference wiring.
type Author func(need string) (name string, expression string, err error)

// Mint hands back an engram ID for an expectation that has not run yet.
//
// It does NOT put anything in the index — mint-on-closure is the kernel's, and an
// expectation that has never closed must not be discoverable. This only allocates
// identity so a delegated name can point at something.
type Mint func(x graph.Expectation) *graph.Engram

// The system expressions. These are hand-written and fixed, which is what
// discover.md means by the fixed-body trade: creation gated by a FIXED BODY is
// cheap and local, and cannot learn. What learns is COMPETITION AMONG discover
// engrams — different retrieval strategies accumulating different histories, with
// selection moving between them.
const (
	Discover   = "system:discover"
	ByIndex    = "system:by-index"
	ByRecency  = "system:by-recency"
	Create     = "system:create"
	Interpreta = "system:interpret" // reserved; not built
)

// Entity is the runtime manning a system possibility.
type Entity struct {
	ID    ids.EntityID
	Index Index
	Auth  Author
	Mint  Mint
	Run   *trace.Run

	// Per-possibility state. Dies with the possibility, like any entity's
	// within-invocation context.
	selected bool
}

func New(id ids.EntityID, idx Index, auth Author, mint Mint, run *trace.Run) func() entity.Interpreter {
	return func() entity.Interpreter {
		return &Entity{ID: id, Index: idx, Auth: auth, Mint: mint, Run: run}
	}
}

// Monitor is a no-op for a code entity. There is no context to fold into — the
// frame is read by field access on each Step rather than accumulated.
//
// That is not a shortcut. It is the same duck-typing one universal signature
// requires: a code expectation reads the frame by FIELD ACCESS; a model
// expectation reads it by ATTENTION. Both work over what they were handed.
func (e *Entity) Monitor(o *graph.Observation) {}

func (e *Entity) Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error) {
	switch x.Expression {
	case Discover:
		return e.discover(f)
	case ByIndex:
		return e.byIndex(f)
	case ByRecency:
		return e.byRecency(f)
	case Create:
		return e.create(f)
	}
	// A code expectation reading something it does not understand fails LOUDLY.
	// interfaces.md flags the asymmetry to watch: this fails loudly where a model
	// given the same frame would just do worse quietly — same missing input, very
	// different credit records.
	return nil, fmt.Errorf("runtime entity cannot man expectation %q", x.Expression)
}

// ---------------------------------------------------------------------------
// discover — an orchestrator over resolvers
// ---------------------------------------------------------------------------

// discover tries its resolvers IN ORDER, and the last one creates.
//
// This ordering is not a system rule. It is the strategy embodied in THIS
// discover engram's expectation — and a discover that reaches for create too
// eagerly is OUTCOMPETED by one that does not. The guard is testable rather than
// trusted.
//
//	BEFORE   a rule. asserted, unfalsifiable.
//	NOW      a strategy. embodied, and BEATEN BY A BETTER ONE if it is wrong.
func (e *Entity) discover(f frame.Frame) (grammar.Emission, error) {
	given := f.Observations()
	if len(given) == 0 {
		return nil, fmt.Errorf("discover invoked with nothing to resolve")
	}
	cite := given[0].ID

	// Forward EVERYTHING down to the resolver, not just the need.
	//
	// FOUND BY RUNNING IT, twice — the caller passed the failure back exactly as
	// asked, and discover dropped it on the floor, forwarding only its first
	// argument. by-index went on returning the same wrong engram and the chain
	// livelocked anyway.
	//
	// The same shape as the finding it was meant to fix: THE INFORMATION EXISTED
	// AND NOTHING ROUTED IT. An orchestrator that narrows what it hands down is
	// making a judgement about relevance, and here it made the wrong one silently.
	// A possibility may pass a subset — that is scoping, and it is legitimate —
	// but a resolver cannot exclude what it never saw.
	all := make([]ids.ObservationID, 0, len(given))
	for _, o := range given {
		all = append(all, o.ID)
	}

	rets := f.Returned()

	// Nothing tried yet — the cheap rung first.
	if len(rets) == 0 {
		return grammar.Select{Members: []grammar.Member{{
			Engram:       "by-index",
			Cite:         cite,
			Entity:       e.ID,
			Observations: all,
		}}}, nil
	}

	// A resolver answered. If it found reach, name it and close.
	last := rets[len(rets)-1]
	if len(last.Content.Reach()) > 0 {
		return grammar.Commit{Observed: last.ID}, nil
	}

	// Rung 2: recency, if the index missed.
	if len(rets) == 1 {
		return grammar.Select{Members: []grammar.Member{{
			Engram:       "by-recency",
			Cite:         cite,
			Entity:       e.ID,
			Observations: all,
		}}}, nil
	}

	// Rung 3: create. Expensive — it GROWS structure.
	if len(rets) == 2 {
		return grammar.Select{Members: []grammar.Member{{
			Engram:       "create",
			Cite:         cite,
			Entity:       e.ID,
			Observations: all,
		}}}, nil
	}

	// Everything was tried and nothing resolved. THIS IS THE DIAGNOSIS, and it is
	// why discover needs an expectation rather than a signature: a flat list is
	// raw material, and "there is nothing here that does X" is a CONSTRUCTED
	// claim. A function returns a list; only a possibility can return this.
	//
	// It is still a commit by NAMING — discover selected, so it has forfeited
	// origination and cannot author the diagnosis itself.
	return grammar.Commit{Observed: last.ID}, nil
}

// ---------------------------------------------------------------------------
// the resolvers — leaves, code, one shot
// ---------------------------------------------------------------------------

// byIndex tokenizes the need and queries the index.
//
// An index holds only what a perspective already ACCUMULATED, so it structurally
// cannot find what has never been interacted with. That is the half discover
// cannot do alone.
//
// It walks the INDEX, not the caller's children — so it finds things nobody has
// used yet, including earlier failed attempts. That is memoization rather than
// bloat: the second time a need arises, discovery finds the orphan instead of
// re-improvising it.
func (e *Entity) byIndex(f frame.Frame) (grammar.Emission, error) {
	need, ok := needText(f)
	if !ok {
		return nil, fmt.Errorf("by-index: no need to resolve")
	}
	want := tokens(need)
	spent := rejected(f)

	type scored struct {
		e *graph.Engram
		n int
	}
	var hits []scored
	for _, cand := range e.Index.Engrams() {
		if strings.HasPrefix(cand.Expectation.Expression, "system:") {
			continue // system space is entered from user space, never resolved into
		}
		if spent[cand.ID] {
			continue // already tried and handed back for this need
		}
		n := overlap(want, tokens(cand.Expectation.Expression))
		if n > 0 {
			hits = append(hits, scored{cand, n})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].n > hits[j].n })

	if len(hits) == 0 {
		// Not an error. "Nothing here does X" is the answer.
		return grammar.Commit{Content: content.Content{
			"resolver": content.Str("by-index"),
			"found":    content.Bool(false),
			"searched": content.Num(float64(len(e.Index.Engrams()))),
		}}, nil
	}

	best := hits[0].e
	return grammar.Commit{Content: content.Content{
		"resolver": content.Str("by-index"),
		"found":    content.Bool(true),
		"reach":    content.EngramRef{ID: best.ID, Name: nameFor(best)},
		"matched":  content.Str(best.Expectation.Expression),
	}}, nil
}

// byRecency returns what has been used lately. A different strategy over the
// same index, accumulating its own history — which is the whole point of
// resolvers being engrams rather than branches in one function.
func (e *Entity) byRecency(f frame.Frame) (grammar.Emission, error) {
	var best *graph.Engram
	bestN := 0
	for _, cand := range e.Index.Engrams() {
		if strings.HasPrefix(cand.Expectation.Expression, "system:") {
			continue
		}
		n := 0
		for _, inv := range cand.History {
			if inv.Cited {
				n++
			}
		}
		if n > bestN {
			best, bestN = cand, n
		}
	}
	if best == nil || rejected(f)[best.ID] {
		return grammar.Commit{Content: content.Content{
			"resolver": content.Str("by-recency"),
			"found":    content.Bool(false),
		}}, nil
	}
	return grammar.Commit{Content: content.Content{
		"resolver": content.Str("by-recency"),
		"found":    content.Bool(true),
		"reach":    content.EngramRef{ID: best.ID, Name: nameFor(best)},
		"used":     content.Num(float64(bestN)),
	}}, nil
}

// create authors a new expectation.
//
// It hands back REACH to an engram that has not run. Nothing is in the index yet
// — mint-on-closure means it persists only if it closes, so an improvised
// expectation is never discoverable before it has proven it can close.
//
// NAME THE NEED, NOT THE COMPLAINT. The name is what gets tokenized for the
// index, so a new possibility must be named for what it does — "find a red
// utensil" — never for the rejection that prompted it. The negation is the cite.
//
// FLAGGED: this is where questions.md §2 (the unattributed term) becomes
// concrete. The Author call is a model call whose contribution appears in no
// citation. Everything else in the chain traces back to what entered from
// outside; the authored expression does not.
func (e *Entity) create(f frame.Frame) (grammar.Emission, error) {
	need, ok := needText(f)
	if !ok {
		return nil, fmt.Errorf("create: no need to author for")
	}
	name, expr, err := e.Auth(need)
	if err != nil {
		return nil, fmt.Errorf("create: authoring failed: %w", err)
	}
	born := e.Mint(graph.Expectation{Expression: expr})
	e.Run.Logf("    create authored %q → %s", name, born.ID)

	return grammar.Commit{Content: content.Content{
		"resolver":    content.Str("create"),
		"found":       content.Bool(true),
		"reach":       content.EngramRef{ID: born.ID, Name: name},
		"expectation": content.Str(expr),
	}}, nil
}

// ---------------------------------------------------------------------------

// rejected reads WHAT HAS ALREADY BEEN TRIED for this need, out of the
// observations the caller handed down.
//
// FOUND BY RUNNING IT. discover.md's refinement ladder ends at "discover →
// create a NEW one," and says the failure "is cited as the prompt for a new
// discovery." But a possibility is one opening: each discover invocation gets a
// fresh frame and cannot remember that its own last answer was handed back.
//
// Observed: the root asked for a bloom-filter summary, got handed the
// write-ahead-log engram, refused to name it, and re-selected discover — which
// deterministically returned the SAME engram. Three times, then eviction. The
// ladder's third rung is unreachable while every rung is amnesiac.
//
// The fix uses nothing new. An observation carries its PRODUCER, so a caller that
// passes back what failed has already said which engram failed — no rejection
// type, no new field, no memory anywhere. What the spec set is missing is the
// SENTENCE saying discover must read it, and that a caller must pass it.
//
// NOTE what this is not: it is not credit. A rejection here does not walk, does
// not land on anyone's record, and does not descend — descending a rejection is a
// shape the design refuses. It is scoped to one discovery.
func rejected(f frame.Frame) map[ids.EngramID]bool {
	out := map[ids.EngramID]bool{}
	for _, o := range f.Observations() {
		// THE WORK THAT FAILED. Its producer is the engram that did it.
		if o.Producer != "" {
			out[o.Producer] = true
		}
		// THE SUGGESTION THAT FAILED. A discover result names its engram in a
		// reach, and its OWN producer is discover — which is not the thing to
		// exclude.
		//
		// FOUND BY RUNNING IT, third time on the same mechanism. "Pass back what
		// failed" is AMBIGUOUS, and the run took the other reading: a caller that
		// dislikes a suggestion hands back the suggestion, never having invoked
		// anything. Those are two different observations with two different
		// producers, and only one of them names the engram to skip.
		//
		// Both readings are legitimate — "I tried it and it was wrong" and "I do
		// not want that one" — so both are honoured. discover.md's sentence covers
		// neither explicitly.
		for _, er := range o.Content.Reach() {
			out[er.ID] = true
		}
	}
	return out
}

func needText(f frame.Frame) (string, bool) {
	// The need is what carries a "need" field. A passed-back failure is an
	// observation too, and reading it as the need would make the complaint the
	// query — which is exactly what discover.md forbids: NAME THE NEED, NOT THE
	// COMPLAINT.
	for _, o := range f.Observations() {
		if v, ok := o.Content["need"]; ok {
			if s, ok := v.(content.Str); ok {
				return string(s), true
			}
		}
	}
	for _, o := range f.Observations() {
		if o.IsIngress() {
			return o.Content.Render(), true
		}
	}
	if len(f.Observations()) > 0 {
		return f.Observations()[0].Content.Render(), true
	}
	return "", false
}

var stop = map[string]bool{
	"a": true, "an": true, "the": true, "to": true, "of": true, "for": true,
	"and": true, "or": true, "in": true, "on": true, "is": true, "it": true,
	"that": true, "this": true, "with": true, "as": true, "be": true, "by": true,
}

func tokens(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	}) {
		if len(w) > 2 && !stop[w] {
			out[w] = true
		}
	}
	return out
}

func overlap(a, b map[string]bool) int {
	n := 0
	for w := range a {
		if b[w] {
			n++
		}
	}
	return n
}

// nameFor derives a selectable name from an expectation. The index is keyed by
// ID; a NAME is how it is reached, and the receiving frame needs one.
func nameFor(e *graph.Engram) string {
	f := strings.Fields(strings.ToLower(e.Expectation.Expression))
	var parts []string
	for _, w := range f {
		w = strings.Map(func(r rune) rune {
			if ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') {
				return r
			}
			return -1
		}, w)
		if len(w) > 2 && !stop[w] {
			parts = append(parts, w)
		}
		if len(parts) == 3 {
			break
		}
	}
	if len(parts) == 0 {
		return string(e.ID)
	}
	return strings.Join(parts, "-")
}
