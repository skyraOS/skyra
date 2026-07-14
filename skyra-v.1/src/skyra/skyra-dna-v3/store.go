package dnav3

import (
	"sort"
	"strings"
	"sync"
)

// The two durable stores, and the derived index over them.
//
//	DURABLE
//	  observation log     append-only. the raw record of what crossed the boundary.
//	  expression graph    acts, expects, framings — AND THEIR GRADES. the being's memory.
//
//	DERIVED (rebuildable, never authoritative)
//	  observation graph   a fold over the log. holds world state.
//	  reverse index       posting lists over the expression graph.

// ObservationGraph is the world state, and it is a FOLD OVER THE OBSERVATION LOG (LAW 1). Only the
// decomposer writes it. Expression has no write access, so the being cannot believe its own output
// into reality.
type ObservationGraph struct {
	mu  sync.Mutex
	log []Observation          // append-only. THE durable thing.
	cur map[string]Observation // (subject,predicate) → latest. derived: the fold.
}

func NewObservationGraph() *ObservationGraph {
	return &ObservationGraph{cur: map[string]Observation{}}
}

// Write appends an observation and folds it into world state. THIS IS THE ONLY WRITER.
func (g *ObservationGraph) Write(o Observation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.log = append(g.log, o)
	g.cur[sp(o.Triple)] = o
}

// Resolve is THE DIFF, and it is a set operation — no model, ever.
//
//	present, object matches      → TRUE
//	present, object differs      → FALSE   (wrong-value: predicted ok, observed running)
//	absent                       → UNKNOWN (ABSENCE IS NOT FALSITY — the being simply has not
//	                                        looked. the `within` timer will send it to look.)
//
// That last line is the single most destructive bug available in this design if you get it wrong:
// treating "I didn't check" as "it didn't happen" drags every edge in the graph toward zero.
func (g *ObservationGraph) Resolve(t Triple) (Truth, Observation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	o, ok := g.cur[sp(t)]
	if !ok {
		return TruthUnknown, Observation{}
	}
	if o.Object == t.Object {
		return TruthTrue, o
	}
	return TruthFalse, o
}

// Log returns the append-only record. Everything else in this file is a fold over it.
func (g *ObservationGraph) Log() []Observation {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]Observation, len(g.log))
	copy(out, g.log)
	return out
}

// sp keys world state by (subject, predicate) — NOT by the full triple. The object is the VALUE at
// that address, which is what lets Resolve tell "wrong value" apart from "never looked."
func sp(t Triple) string { return t.Subject + "\x00" + t.Predicate }

// ExpressionGraph is the being's MEMORY: every claim it ever made and how those claims fared. The
// observation log is what happened; this is what the being made of it, and it is the only part
// nobody else could reconstruct.
type ExpressionGraph struct {
	mu    sync.Mutex
	exprs map[string]*Expression
	edges map[string]*Edge    // edgeKey → the learned P(W | E)
	index map[string][]string // triple Key() → expression IDs that PRODUCED it (the reverse index)
	order []string            // insertion order, so retrieval is deterministic
}

func NewExpressionGraph() *ExpressionGraph {
	return &ExpressionGraph{
		exprs: map[string]*Expression{},
		edges: map[string]*Edge{},
		index: map[string][]string{},
	}
}

// Place inserts an emission and wires its edges + posting lists. This is the placer worker: no
// model, pure graph insertion.
func (g *ExpressionGraph) Place(id string, e *Expression) {
	g.mu.Lock()
	defer g.mu.Unlock()
	e.ID = id
	g.exprs[id] = e
	g.order = append(g.order, id)

	for i := range e.Expects {
		x := &e.Expects[i]
		key := x.Key()

		// reverse index: "how have I made this true before?"
		g.index[key] = append(g.index[key], id)

		// the learned edge. Source is a SET — cardinality gives joints for free.
		ek := edgeKey(id, x.Of, key)
		if g.edges[ek] == nil {
			g.edges[ek] = &Edge{Source: x.Of, Target: key}
		}
	}
}

// Edge returns the learned edge for one expect (creating it if this is its first sighting). An edge
// with N=0 is not a bad edge — it is an UNMEASURED one.
func (g *ExpressionGraph) Edge(exprID string, x Expect) *Edge {
	g.mu.Lock()
	defer g.mu.Unlock()
	ek := edgeKey(exprID, x.Of, x.Key())
	if g.edges[ek] == nil {
		g.edges[ek] = &Edge{Source: x.Of, Target: x.Key()}
	}
	return g.edges[ek]
}

// Precedent answers the reverse-index query: which prior expressions produced this world state?
// Ranked by how reliably they actually produced it — P × confidence(N) × F(now).
//
// This is where the grade changes what gets recalled. Not "the model remembers better": THE EDGES
// RERANK THE POSTING LISTS, and P is a counted conditional probability, not anyone's opinion.
func (g *ExpressionGraph) Precedent(t Triple, tick int) []*Expression {
	g.mu.Lock()
	defer g.mu.Unlock()

	type scored struct {
		e *Expression
		w float64
	}
	var out []scored
	for _, id := range g.index[t.Key()] {
		e := g.exprs[id]
		if e == nil {
			continue
		}
		best := 0.0
		for _, x := range e.Expects {
			if x.Key() != t.Key() {
				continue
			}
			if w := g.edges[edgeKey(id, x.Of, x.Key())].RetrievalWeight(tick); w > best {
				best = w
			}
		}
		out = append(out, scored{e, best})
	}
	// deterministic: weight desc, then ID — replay must not depend on map order
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].w != out[j].w {
			return out[i].w > out[j].w
		}
		return out[i].e.ID < out[j].e.ID
	})

	es := make([]*Expression, len(out))
	for i, s := range out {
		es[i] = s.e
	}
	return es
}

// Open returns every expect that has not resolved yet, with the expression that made it.
func (g *ExpressionGraph) Open() []OpenBet {
	g.mu.Lock()
	defer g.mu.Unlock()
	var out []OpenBet
	for _, id := range g.order {
		e := g.exprs[id]
		for i := range e.Expects {
			if e.Expects[i].Resolution == TruthUnknown {
				out = append(out, OpenBet{Expr: e, Expect: &e.Expects[i]})
			}
		}
	}
	return out
}

// OpenBet is one unresolved claim and the expression that staked it.
type OpenBet struct {
	Expr   *Expression
	Expect *Expect
}

// edgeKey identifies the learned edge: the SET of acts responsible (within this expression) → the
// world state claimed. Sorted, so {a1,a2} and {a2,a1} are the same edge.
func edgeKey(exprID string, of []string, target string) string {
	src := append([]string(nil), of...)
	sort.Strings(src)
	return exprID + "\x00" + strings.Join(src, "+") + "\x00" + target
}
