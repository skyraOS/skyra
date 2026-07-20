package dnav3

import (
	"math"
	"sort"
)

// Segmentation — spans and classes as DERIVED indexes over the observation log.
//
// Everything in this file is a pure function over []Observation: rebuildable, never
// authoritative, no model anywhere. Design and validation:
//
//	specs/2026-07-20/episode-as-region.md         span = one witnessed occurrence;
//	                                              episode = the CLASS of similar spans
//	specs/2026-07-19/observe-paper-test.md        why single-span episodes carry nothing
//	specs/2026-07-20/observe-paper-test-4.md      both rules recover hand segmentation
//	specs/2026-07-20/live-segmentation-run.md     the composition that survives real data:
//	                                              OVER-SEGMENT (finest rhythm), THEN MERGE
//
// The two rules answer different questions off different signals, and conflating them was
// v2's edit/new trap (one LLM judgment doing both jobs at write time):
//
//	BOUNDARY     temporal.  where does one occurrence end?    renewal gaps in WALL time.
//	MEMBERSHIP   spatial.   which occurrences are the same     co-occurrence of resolved
//	                        kind of thing?                     identity above baseline.

const (
	// DefaultGapK is the boundary multiplier: a gap is a boundary when it exceeds K × the
	// finest witnessed rhythm. Deliberately EAGER — the live run showed a single sitting has
	// no hours-band, so the rule must cut at exchange scale and let Classes remend what it
	// over-cuts. Tuning K upward to avoid over-segmentation is the wrong fix.
	DefaultGapK = 5.0

	// DefaultMergeThreshold is the class-merge floor on baseline-weighted shared identity —
	// calibrated to "at least two informative shared terms" (log-weights at df 2/5 give
	// 2×0.92 = 1.84; a single shared term stays below). One shared entity — a place, a
	// promiscuous verb — must never found a class: S1·S2 shared kitchen-fridge at 0.92 in
	// paper test #4 and correctly stayed apart.
	DefaultMergeThreshold = 1.5
)

// A Span is one witnessed occurrence: a contiguous stretch of the observation log, cut by
// renewal gaps. Indices are into the log slice the span was computed from — a span is a
// VIEW, not a store.
type Span struct {
	ID          SpanID
	First, Last int   // log indices, inclusive
	WallStart   int64 // wall of the first moment
	WallEnd     int64 // wall of the last moment
	Open        bool  // the tail span: no boundary has closed it yet. Its class
	//                   membership is PROVISIONAL (strict at close, provisional while open).
}

// Spans cuts the observation log into spans by renewal gaps (rule 1).
//
// The rhythm is the median positive wall-delta in the log — the finest witnessed cadence.
// A boundary falls wherever a gap exceeds k × rhythm. k <= 0 takes DefaultGapK.
//
// A log without wall stamps (all zero) cannot be segmented temporally and comes back as one
// open span — GIGO, by design: the fix is stamping (Observation.Wall), not guessing here.
func Spans(log []Observation, k float64) []Span {
	if len(log) == 0 {
		return nil
	}
	if k <= 0 {
		k = DefaultGapK
	}
	var deltas []float64
	for i := 1; i < len(log); i++ {
		if d := log[i].Wall - log[i-1].Wall; d > 0 {
			deltas = append(deltas, float64(d))
		}
	}
	cut := math.Inf(1)
	if len(deltas) > 0 {
		// The baseline is the FINEST sustained rhythm — lower quartile, not median. The gap
		// spectrum is multi-band (moments / exchanges / sessions), and when boundary gaps
		// rival intra-deltas in count a median lands in the wrong band and yields one giant
		// span (the live run's first failure). The lower quartile stays in the bottom band.
		sort.Float64s(deltas)
		cut = k * deltas[len(deltas)/4]
	}
	var out []Span
	first := 0
	for i := 1; i < len(log); i++ {
		if float64(log[i].Wall-log[i-1].Wall) > cut {
			out = append(out, span(log, first, i-1))
			first = i
		}
	}
	tail := span(log, first, len(log)-1)
	tail.Open = true
	return append(out, tail)
}

func span(log []Observation, first, last int) Span {
	return Span{ID: log[first].SpanID, First: first, Last: last, WallStart: log[first].Wall, WallEnd: log[last].Wall}
}

// A Class is an episode in the episode-as-region sense: the holder of similar spans — the
// precedent class. Statistics (edge EMAs grouped by class) attach HERE, not to spans; the
// wave heats classes and descends into member spans only for specifics.
type Class struct {
	Members   []int              // indices into the spans slice, ascending
	Signature map[string]float64 // identity term → baseline weight (0 for ubiquitous terms)
	Open      bool               // true if any member span is open — membership provisional
}

// Classes groups spans into classes by baseline-weighted co-occurrence (rule 2).
//
// A span's identity is the set of RESOLVED terms committed inside it — triple subjects,
// predicates, and objects. These are canonical by construction in v3 (Triple.Key identity;
// world-given namespaces cost no coreference), which is exactly the property the live run
// found missing in the v2 trace it replayed.
//
// Terms are weighted log(N/df): a term in every span weighs zero (the fan effect's identity
// twin — michael discriminates nothing; 'ask' nearly nothing). Two spans merge when the
// summed weight of their shared terms clears threshold (<= 0 takes DefaultMergeThreshold).
// Merging is transitive (union-find): over-segmentation from an eager Spans remends here.
func Classes(log []Observation, spans []Span, threshold float64) []Class {
	if threshold <= 0 {
		threshold = DefaultMergeThreshold
	}
	sigs := make([]map[string]bool, len(spans))
	df := map[string]int{}
	for i, s := range spans {
		sig := map[string]bool{}
		for j := s.First; j <= s.Last && j < len(log); j++ {
			for _, term := range []string{log[j].Subject, log[j].Predicate, log[j].Object} {
				if term != "" {
					sig[term] = true
				}
			}
		}
		sigs[i] = sig
		for term := range sig {
			df[term]++
		}
	}
	n := float64(len(spans))
	weight := func(term string) float64 {
		if d := df[term]; d < len(spans) {
			return math.Log(n / float64(d))
		}
		return 0
	}

	parent := make([]int, len(spans))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			score := 0.0
			for term := range sigs[i] {
				if sigs[j][term] {
					score += weight(term)
				}
			}
			if score >= threshold {
				parent[find(j)] = find(i)
			}
		}
	}

	byRoot := map[int]*Class{}
	var order []int
	for i := range spans {
		r := find(i)
		c := byRoot[r]
		if c == nil {
			c = &Class{Signature: map[string]float64{}}
			byRoot[r] = c
			order = append(order, r)
		}
		c.Members = append(c.Members, i)
		c.Open = c.Open || spans[i].Open
		for term := range sigs[i] {
			c.Signature[term] = weight(term)
		}
	}
	out := make([]Class, 0, len(order))
	sort.Ints(order)
	for _, r := range order {
		out = append(out, *byRoot[r])
	}
	return out
}

// Lens is the entity → class posting lists: for each informative identity term, which
// classes does it participate in? This is retrieval.md's reverse index grouped one level up
// — the structure the wave's lens-hop reads. Ubiquitous terms (weight 0) are omitted; they
// have nothing to say (and including them would hand every impulse every class).
func Lens(classes []Class) map[string][]int {
	out := map[string][]int{}
	for i, c := range classes {
		for term, w := range c.Signature {
			if w > 0 {
				out[term] = append(out[term], i)
			}
		}
	}
	return out
}
