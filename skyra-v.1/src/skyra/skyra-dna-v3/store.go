package dnav3

import (
	"fmt"
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

	am := actMap(e)
	for i := range e.Expects {
		x := &e.Expects[i]
		key := x.Key()

		// reverse index: "how have I made this true before?"
		g.index[key] = append(g.index[key], id)

		// the learned edge. Keyed by what the acts DO, not by which expression emitted them — so it
		// ACCUMULATES across episodes. Source is a SET: cardinality gives joints for free.
		ek := edgeKey(am, x.Of, key)
		if g.edges[ek] == nil {
			g.edges[ek] = &Edge{Source: x.Of, Target: key}
		}
	}
}

// Edge returns the learned edge for one expect (creating it if this is its first sighting). An edge
// with N=0 is not a bad edge — it is an UNMEASURED one.
//
// The edge is SHARED across every expression whose acts do the same thing. That is the whole point:
// it is the memory.
func (g *ExpressionGraph) Edge(exprID string, x Expect) *Edge {
	g.mu.Lock()
	defer g.mu.Unlock()
	e := g.exprs[exprID]
	if e == nil {
		return &Edge{Source: x.Of, Target: x.Key()}
	}
	ek := edgeKey(actMap(e), x.Of, x.Key())
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
		am := actMap(e)
		best := 0.0
		for _, x := range e.Expects {
			if x.Key() != t.Key() {
				continue
			}
			if ed := g.edges[edgeKey(am, x.Of, x.Key())]; ed != nil {
				if w := ed.RetrievalWeight(tick); w > best {
					best = w
				}
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

// edgeKey identifies the learned edge: the SET of acts responsible → the world state claimed.
//
// # It must NOT include the expression instance
//
// This was the bug the first live run found, and it was fatal. The key used to be
// (exprID, actIDs, target) — so every run minted a NEW expression (e1, e2, e3…), and therefore a
// NEW, UNMEASURED edge for the same claim:
//
//	P=0.75 N=1  {a1} → file:/greet.sh exists true   [claimed 0.99]
//	P=0.75 N=1  {a1} → file:/greet.sh exists true   [claimed 0.98]
//	P=0.75 N=1  {a1} → file:/greet.sh exists true   [claimed 0.99]
//
// Four edges. Same claim. N=1 forever. P never converges, confidence(N) stays at rock bottom, and
// THE BEING CANNOT LEARN ACROSS EPISODES — which is the one thing this entire architecture exists to
// do.
//
// The same bug class as every other one we have hit: THE REUSABLE THING WAS KEYED BY THE INSTANCE
// THAT MADE IT, NOT BY WHAT IT IS.
//
// So the key is the acts' REUSABLE IDENTITY — what they actually do — sorted, so {a1,a2} and
// {a2,a1} are one edge:
//
//	{write→file:/greet.sh} → file:/greet.sh exists true
//
// That accumulates across every run that ever writes that file, which is exactly what an edge is
// for: "given this expression, does this world state come back?"
func edgeKey(acts map[string]Act, of []string, target string) string {
	sigs := make([]string, 0, len(of))
	for _, id := range of {
		a, ok := acts[id]
		if !ok {
			sigs = append(sigs, "?"+id) // an expect attributed to an act that is not here
			continue
		}
		sigs = append(sigs, a.Primitive+"→"+a.Target)
	}
	sort.Strings(sigs)
	return strings.Join(sigs, "+") + "\x00" + target
}

// actMap indexes an expression's acts by ID, for edgeKey.
func actMap(e *Expression) map[string]Act {
	m := make(map[string]Act, len(e.Acts))
	for _, a := range e.Acts {
		m[a.ID] = a
	}
	return m
}

// EdgeReport dumps what the being has actually LEARNED — the counted conditional probabilities, not
// anyone's opinion. This is the memory.
func (g *ExpressionGraph) EdgeReport() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	seen := map[string]bool{}
	var out []string
	for _, id := range g.order {
		e := g.exprs[id]
		am := actMap(e)
		for _, x := range e.Expects {
			ek := edgeKey(am, x.Of, x.Key())
			ed := g.edges[ek]
			if ed == nil || ed.N == 0 || seen[ek] {
				continue // an unmeasured edge is not a bad edge. it has nothing to say yet.
			}
			seen[ek] = true
			out = append(out, fmt.Sprintf("P=%.2f N=%-2d  %s  →  %s %s %s",
				ed.P, ed.N, strings.SplitN(ek, "\x00", 2)[0], x.Subject, x.Predicate, x.Object))
		}
	}
	sort.Strings(out)
	return out
}
