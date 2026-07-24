package dnav3

import (
	"sort"
	"strconv"
	"strings"
)

// The observe wave — the mechanism inside observation → state change.
//
// Spec: specs/2026-07-20/the-observe-wave.md. Paper trail: tests #1–#4 under
// specs/2026-07-19..20. ZERO inference calls anywhere in this file: index lookups and
// arithmetic over the derived segmentation (segment.go) and resolution history, same
// species as every other set operation in v3 (LAW 4).
//
//	ROOTS      each independently resolvable reference proposes candidates
//	           (tier 1: resolution history, counted priors; tier 2 — the model —
//	           happens at the decomposer, NOT here: a term with no history simply
//	           stands as its own reading until a collapse founds its first edge)
//	HEAT       candidates heat classes through the lens, amplitude conserved
//	RANK       distinct-root count first, heat second — a root's candidates are
//	           alternatives and can never corroborate; corroboration is across roots
//	SELECT     bindings by counted statistics (the Stats hook — Born rule)
//	COVERAGE   ungroundable roots stitch from lower-ranked classes
//	JOINT      the committed binding set is checked for co-occurrence in some span;
//	           zero joint precedent = a NOVEL stitch — legal, but flagged
//	COMMIT     interpreted resolved-to observations, DerivedFrom the raw impulse
//
// The wiring seam (spec delta 6): v3's walking skeleton is target-driven and has no
// impulse/decomposer path yet. When ingest lands, the wave runs between decompose and
// present; until then everything here is callable and tested against the fixtures.

// PredicateResolvedTo is the commit predicate: {surface term, resolved-to, entity}.
// Interpreted provenance — a reading, capped fidelity, one day graded by contradiction
// (this is observation.go's tabled feedback loop getting its mechanism). These commits
// are also what BuildResolutionIndex folds into tier-1 priors: every collapse makes the
// next proposal cheaper — cost scales with surprise, at ingest.
const PredicateResolvedTo = "resolved-to"

// A Candidate is one possible reading of a root, with a counted prior.
type Candidate struct {
	ID    string
	Prior float64
}

// RootKind — entity, verb, context. No special machinery for any of them (paper test
// #3: the verb "back" flips outcomes; witnessed context participates identically).
type RootKind string

const (
	RootEntity  RootKind = "entity"
	RootVerb    RootKind = "verb"
	RootContext RootKind = "context"
)

// A Root is one independently resolvable reference in the impulse. Its candidates are
// the superposition — plural readings live here, and only here (within a root).
type Root struct {
	Term       string
	Kind       RootKind
	Candidates []Candidate
}

// ResolutionIndex is tier 1 of the two-tier proposer: surface term → the readings it
// has historically committed to, with counted priors. A materialized cache over
// resolved-to commits — deletable, rebuildable, cheap to be wrong about.
type ResolutionIndex map[string][]Candidate

// BuildResolutionIndex folds resolution history out of an observation log. Priors are
// counted frequencies, never equal-split and never an opinion.
func BuildResolutionIndex(log []Observation) ResolutionIndex {
	counts := map[string]map[string]float64{}
	for _, o := range log {
		if o.Predicate != PredicateResolvedTo {
			continue
		}
		m := counts[o.Subject]
		if m == nil {
			m = map[string]float64{}
			counts[o.Subject] = m
		}
		m[o.Object]++
	}
	out := ResolutionIndex{}
	for term, m := range counts {
		total := 0.0
		for _, n := range m {
			total += n
		}
		cs := make([]Candidate, 0, len(m))
		for id, n := range m {
			cs = append(cs, Candidate{ID: id, Prior: n / total})
		}
		sort.Slice(cs, func(i, j int) bool {
			if cs[i].Prior != cs[j].Prior {
				return cs[i].Prior > cs[j].Prior
			}
			return cs[i].ID < cs[j].ID
		})
		out[term] = cs
	}
	return out
}

// Roots extracts the impulse's roots and proposes candidates. A term with no
// resolution history stands as its own reading (prior 1) — first contact is the
// decomposer's business upstream, and after the first collapse it never is again.
func Roots(impulse []Observation, context []Triple, rx ResolutionIndex) []Root {
	seen := map[string]bool{}
	var out []Root
	add := func(term string, kind RootKind) {
		if term == "" || seen[term] {
			return
		}
		seen[term] = true
		cs := rx[term]
		if len(cs) == 0 {
			cs = []Candidate{{ID: term, Prior: 1}}
		}
		out = append(out, Root{Term: term, Kind: kind, Candidates: cs})
	}
	for _, o := range impulse {
		add(o.Subject, RootEntity)
		add(o.Object, RootEntity)
		add(o.Predicate, RootVerb)
	}
	for _, t := range context {
		add(t.Subject, RootContext)
		add(t.Object, RootContext)
	}
	return out
}

// ClassHeat is one class's share of the wave. Rank by Roots FIRST: heat from one
// root's own candidates is alternatives piling up, not agreement (paper test #2's
// tie: 0.75 from two roots beats 0.75 from one).
type ClassHeat struct {
	Class int
	Roots int
	Heat  float64
}

// HeatClasses runs lens-hop with conservation: each root carries amplitude 1.0, split
// across its candidates by prior, then across each candidate's classes. Ubiquitous
// terms are absent from the lens by construction and heat nothing.
func HeatClasses(roots []Root, lens map[string][]int) []ClassHeat {
	heat := map[int]float64{}
	rootsIn := map[int]map[int]bool{}
	for ri, r := range roots {
		total := 0.0
		for _, c := range r.Candidates {
			total += c.Prior
		}
		if total == 0 {
			continue
		}
		for _, c := range r.Candidates {
			classes := lens[c.ID]
			if len(classes) == 0 {
				continue
			}
			share := (c.Prior / total) / float64(len(classes))
			for _, ci := range classes {
				heat[ci] += share
				if rootsIn[ci] == nil {
					rootsIn[ci] = map[int]bool{}
				}
				rootsIn[ci][ri] = true
			}
		}
	}
	out := make([]ClassHeat, 0, len(heat))
	for ci, h := range heat {
		out = append(out, ClassHeat{Class: ci, Roots: len(rootsIn[ci]), Heat: h})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Roots != out[j].Roots {
			return out[i].Roots > out[j].Roots
		}
		if out[i].Heat != out[j].Heat {
			return out[i].Heat > out[j].Heat
		}
		return out[i].Class < out[j].Class
	})
	return out
}

// Stats is the Born-rule hook: counted reliability of a reading within a class —
// P and N off the class-grouped edges (see ClassEdges). Nil means no statistics yet;
// selection then falls back to prior × signature weight, ties to the earlier
// candidate. An unmeasured edge is not a bad edge, but a measured one outranks it.
type Stats func(class int, term string) (p float64, n int)

// A Binding is one root's committed reading and the class that grounded it.
// Class == -1 is first contact: the root stands as its own reading, ungrounded.
type Binding struct {
	Root      Root
	Candidate Candidate
	Class     int
}

// A Resolution is the wave's whole output. Transient amplitudes are already gone —
// only the readings, the ranking, and the flags survive to the present.
type Resolution struct {
	Bindings []Binding
	Ranked   []ClassHeat
	Primary  int // the top-ranked class; -1 if nothing was heated

	// JointPrecedent counts the member spans in which the grounded readings co-occur.
	// Zero is the novelty flag: a stitch with no witnessed precedent. Legal — first
	// times must be committable — but the express side should see it and prefer a
	// clarifying act over confident execution.
	JointPrecedent int
	Novel          bool
}

// Resolve runs select + coverage + joint precedent over ranked classes.
func Resolve(log []Observation, spans []Span, classes []Class, lens map[string][]int, roots []Root, stats Stats) Resolution {
	ranked := HeatClasses(roots, lens)
	res := Resolution{Ranked: ranked, Primary: -1}
	if len(ranked) > 0 {
		res.Primary = ranked[0].Class
	}

	for _, r := range roots {
		b := Binding{Root: r, Class: -1}
		best := 0.0
		// Walk classes in rank order; the FIRST class that can ground the root does —
		// stitching from lower-ranked classes only happens for roots the winner
		// cannot cover (the coverage gate).
		for _, ch := range ranked {
			sig := classes[ch.Class].Signature
			found := false
			for _, c := range r.Candidates {
				w, ok := sig[c.ID]
				if !ok || w <= 0 {
					continue // grounding on a ubiquitous term is not grounding
				}
				score := c.Prior * (1 + w)
				if stats != nil {
					if p, n := stats(ch.Class, c.ID); n > 0 {
						score *= 0.5 + p
					}
				}
				if score > best {
					best = score
					b.Candidate = c
					b.Class = ch.Class
					found = true
				}
			}
			if found {
				break
			}
		}
		if b.Class == -1 {
			b.Candidate = r.Candidates[0] // first contact: the term is its own reading
		}
		res.Bindings = append(res.Bindings, b)
	}

	// Joint precedent: in how many witnessed spans do ALL grounded readings co-occur?
	// Two exclusions, both from paper-test ground (strict at close, provisional while
	// open): an OPEN span is not witnessed precedent — least of all the impulse's own
	// span, which trivially contains every term it introduced; and an IDENTITY
	// binding (first contact, the term standing as its own reading) is not a reading
	// and cannot corroborate anything.
	var grounded []string
	for _, b := range res.Bindings {
		if b.Class != -1 && b.Candidate.ID != b.Root.Term {
			grounded = append(grounded, b.Candidate.ID)
		}
	}
	if len(grounded) > 0 {
		for _, s := range spans {
			if s.Open {
				continue
			}
			terms := map[string]bool{}
			for i := s.First; i <= s.Last && i < len(log); i++ {
				terms[log[i].Subject] = true
				terms[log[i].Predicate] = true
				terms[log[i].Object] = true
			}
			all := true
			for _, gterm := range grounded {
				if !terms[gterm] {
					all = false
					break
				}
			}
			if all {
				res.JointPrecedent++
			}
		}
	}
	res.Novel = res.JointPrecedent == 0
	return res
}

// Commit renders the resolution as its state-change write: one interpreted resolved-to
// observation per non-identity binding, DerivedFrom the raw impulse. Confidence is the
// COUNTED prior, not a stated opinion; fidelity is capped by provenance as always.
// The Novel flag travels on the Resolution to the present — it is context for the
// framing call, not a property of any single triple.
func (r Resolution) Commit(derivedFrom []string) []Observation {
	var out []Observation
	for _, b := range r.Bindings {
		if b.Class == -1 || b.Candidate.ID == b.Root.Term {
			continue // identity or first contact: nothing worth recording
		}
		out = append(out, Observation{
			Triple: Triple{
				Mode:       ModeInterpret,
				Confidence: b.Candidate.Prior,
				Subject:    b.Root.Term,
				Predicate:  PredicateResolvedTo,
				Object:     b.Candidate.ID,
			},
			Provenance:  ProvenanceInterpreted,
			Confidence:  b.Candidate.Prior,
			DerivedFrom: derivedFrom,
		})
	}
	return out
}

// ClassEdges groups the learned edges by class — the class statistics of
// episode-as-region, with v3's existing learned object doing all the counting. The
// join runs through evidence: a resolved expect's ResolvedBy names its observation,
// the observation sits in a span, the span belongs to a class. Derived, rebuildable.
func (g *ExpressionGraph) ClassEdges(log []Observation, spans []Span, classes []Class) map[int][]*Edge {
	classOfSpan := map[int]int{}
	for ci, c := range classes {
		for _, si := range c.Members {
			classOfSpan[si] = ci
		}
	}
	spanOfIndex := func(i int) (int, bool) {
		for si, s := range spans {
			if i >= s.First && i <= s.Last {
				return si, true
			}
		}
		return 0, false
	}
	indexOfObsID := func(id string) (int, bool) {
		for i, o := range log {
			if o.ID != "" && string(o.ID) == id {
				return i, true
			}
		}
		// Compatibility for historical/replayed observations written before stable IDs.
		at := 0
		key := id
		if cut := strings.LastIndex(id, "@"); cut >= 0 {
			key = id[:cut]
			at, _ = strconv.Atoi(id[cut+1:])
		}
		for i, o := range log {
			if o.Key() == key && o.At == at {
				return i, true
			}
		}
		return 0, false
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	out := map[int][]*Edge{}
	seen := map[int]map[*Edge]bool{}
	for _, id := range g.order {
		e := g.exprs[id]
		am := actMap(e)
		for _, x := range e.Expects {
			if x.Resolution == TruthUnknown {
				continue
			}
			ed := g.edges[edgeKey(am, x.Of, x.Key())]
			if ed == nil {
				continue
			}
			for _, rid := range x.ResolvedBy {
				oi, ok := indexOfObsID(rid)
				if !ok {
					continue
				}
				si, ok := spanOfIndex(oi)
				if !ok {
					continue
				}
				ci, ok := classOfSpan[si]
				if !ok {
					continue
				}
				if seen[ci] == nil {
					seen[ci] = map[*Edge]bool{}
				}
				if !seen[ci][ed] {
					seen[ci][ed] = true
					out[ci] = append(out[ci], ed)
				}
			}
		}
	}
	return out
}
