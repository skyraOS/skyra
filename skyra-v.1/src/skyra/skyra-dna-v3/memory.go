package dnav3

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

// TokenCounter measures the model-facing cost of rendered memory. Different inference
// providers tokenize differently, so the counter is an injected seam rather than an
// ontological property of English text. The default is a deterministic conservative
// approximation; a provider tokenizer can replace it without changing traversal.
type TokenCounter func(string) int

// ApproxTokenCount is the provider-neutral default. It uses Unicode rune count rather
// than bytes and rounds up at three characters per token so the budget errs toward
// excluding a node rather than overflowing the model context.
func ApproxTokenCount(s string) int {
	n := utf8.RuneCountInString(s)
	if n == 0 {
		return 1
	}
	return (n + 2) / 3
}

func observationMemoryText(o Observation) string {
	return fmt.Sprintf("%s %s %s", o.Subject, o.Predicate, o.Object)
}

func expressionMemoryText(e *Expression) string {
	if e == nil {
		return ""
	}
	var s strings.Builder
	fmt.Fprintf(&s, "framing: %s\n", e.Framing)
	for _, a := range e.Acts {
		fmt.Fprintf(&s, "<act id=%q target=%q primitive=%q>%s</act>\n",
			a.ID, a.Target, a.Primitive, truncate(a.Payload, 200))
	}
	for _, x := range e.Expects {
		fmt.Fprintf(&s, "<expect id=%q> %.2f %d %s %s %s </expect>\n",
			x.PredictionID, x.Confidence, x.Within, x.Subject, x.Predicate, x.Object)
	}
	for _, attribution := range e.PredictionAttributions {
		fmt.Fprintf(&s, "<evidence>%s</evidence>\n", formatEvidenceAttribution(attribution))
	}
	return strings.TrimSpace(s.String())
}

// MemoryNode is the only thing the context renderer budgets. Durable storage types
// stamp their base TokenCount; a node stamps the actual rendered text again because
// labels and learned statistics can change the presentation cost.
type MemoryNode struct {
	ID         string
	Kind       string
	EpisodeID  EpisodeID
	SpanID     SpanID
	TokenCount int
	Activation float64
	Text       string
	Mandatory  bool
}

// Density is relevance delivered per model token. Activation remains a distinct
// quantity: token cost changes selection efficiency, never belief or causal truth.
func (n MemoryNode) Density() float64 {
	tokens := n.TokenCount
	if tokens < 1 {
		tokens = 1
	}
	return n.Activation / float64(tokens)
}

// SpanIndex is the direct, derived memory index. It replaces the class hop:
// entity/address/triple -> witnessed spans, plus the observations and token weight
// owned by each span. It is rebuildable from the append-only observation log.
type SpanIndex struct {
	mu sync.Mutex

	bySpan     map[SpanID][]Observation
	spanTokens map[SpanID]int
	spanOrder  []SpanID

	entity  map[string][]SpanID
	address map[string][]SpanID
	triple  map[string][]SpanID
	seen    map[string]map[SpanID]bool
}

func NewSpanIndex() *SpanIndex {
	return &SpanIndex{
		bySpan:     map[SpanID][]Observation{},
		spanTokens: map[SpanID]int{},
		entity:     map[string][]SpanID{},
		address:    map[string][]SpanID{},
		triple:     map[string][]SpanID{},
		seen:       map[string]map[SpanID]bool{},
	}
}

func (x *SpanIndex) Add(o Observation) {
	if x == nil || o.SpanID == "" {
		return
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	if _, exists := x.bySpan[o.SpanID]; !exists {
		x.spanOrder = append(x.spanOrder, o.SpanID)
	}
	x.bySpan[o.SpanID] = append(x.bySpan[o.SpanID], o)
	x.spanTokens[o.SpanID] += maxInt(1, o.TokenCount)
	for _, term := range []string{o.Subject, o.Predicate, o.Object} {
		x.addPosting("entity\x00"+term, term, o.SpanID, x.entity)
	}
	x.addPosting("address\x00"+sp(o.Triple), sp(o.Triple), o.SpanID, x.address)
	x.addPosting("triple\x00"+o.Key(), o.Key(), o.SpanID, x.triple)
}

func (x *SpanIndex) addPosting(seenKey, key string, span SpanID, dst map[string][]SpanID) {
	if key == "" {
		return
	}
	if x.seen[seenKey] == nil {
		x.seen[seenKey] = map[SpanID]bool{}
	}
	if x.seen[seenKey][span] {
		return
	}
	x.seen[seenKey][span] = true
	dst[key] = append(dst[key], span)
}

func (x *SpanIndex) Entity(term string) []SpanID { return x.postings(x.entity, term) }
func (x *SpanIndex) Address(t Triple) []SpanID   { return x.postings(x.address, sp(t)) }
func (x *SpanIndex) Triple(t Triple) []SpanID    { return x.postings(x.triple, t.Key()) }
func (x *SpanIndex) SpanTokens(id SpanID) int {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.spanTokens[id]
}
func (x *SpanIndex) Observations(id SpanID) []Observation {
	x.mu.Lock()
	defer x.mu.Unlock()
	out := make([]Observation, len(x.bySpan[id]))
	copy(out, x.bySpan[id])
	return out
}

func (x *SpanIndex) postings(src map[string][]SpanID, key string) []SpanID {
	if x == nil {
		return nil
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	out := make([]SpanID, len(src[key]))
	copy(out, src[key])
	return out
}

func (x *SpanIndex) Spans() []SpanID {
	if x == nil {
		return nil
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	out := make([]SpanID, len(x.spanOrder))
	copy(out, x.spanOrder)
	return out
}

// SelectMemory greedily fills a fixed memory budget. Mandatory root nodes are
// admitted first; the remainder competes by activation/token density. Selection is
// deterministic so a replay with the same impulses and parameters yields one present.
func SelectMemory(nodes []MemoryNode, budget int) []MemoryNode {
	if budget <= 0 {
		return nil
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Mandatory != nodes[j].Mandatory {
			return nodes[i].Mandatory
		}
		if nodes[i].Density() != nodes[j].Density() {
			return nodes[i].Density() > nodes[j].Density()
		}
		if nodes[i].Activation != nodes[j].Activation {
			return nodes[i].Activation > nodes[j].Activation
		}
		return nodes[i].ID < nodes[j].ID
	})
	out := make([]MemoryNode, 0, len(nodes))
	spent := 0
	for _, n := range nodes {
		cost := maxInt(1, n.TokenCount)
		if spent+cost > budget {
			continue
		}
		spent += cost
		out = append(out, n)
	}
	return out
}

func (b *Being) activeMemoryNodes(query []Triple) []MemoryNode {
	if b.Activation == nil {
		return nil
	}
	counter := b.CountTokens
	if counter == nil {
		counter = ApproxTokenCount
	}
	nodes := map[string]MemoryNode{}
	for _, active := range b.Activation.RankedSpans() {
		for _, o := range b.Spans.Observations(active.SpanID) {
			kind := "memory"
			label := "observed"
			if truth, current := b.Obs.Resolve(o.Triple); truth != TruthUnknown && current.ID == o.ID {
				kind = "current"
				label = "current"
			}
			if o.Provenance == ProvenanceMark {
				kind = "mark"
				label = "mark"
			}
			// Preserve the legacy contiguous triple text for existing consumers while
			// exposing the stable observation ID and an explicit quoted object surface
			// for Phase-4F character-range attribution.
			text := fmt.Sprintf("%s [observation=%s span=%s] %s [object-text=%q]",
				label, o.ID, o.SpanID, observationMemoryText(o), o.Object)
			nodes[string(o.ID)] = MemoryNode{
				ID:         string(o.ID),
				Kind:       kind,
				EpisodeID:  o.EpisodeID,
				SpanID:     o.SpanID,
				TokenCount: counter(text),
				Activation: b.Activation.Nodes[string(o.ID)],
				Text:       text,
				Mandatory:  o.SpanID == b.currentSpan,
			}
		}
		for _, e := range b.Expr.ExpressionsInSpan(active.SpanID) {
			text := expressionMemoryText(e)
			nodes["expression:"+e.ID] = MemoryNode{
				ID:         "expression:" + e.ID,
				Kind:       "precedent",
				EpisodeID:  e.EpisodeID,
				SpanID:     e.SpanID,
				TokenCount: counter(text),
				Activation: b.nodeActivation("expression:"+e.ID, active.Activation),
				Text:       text,
			}
		}
	}

	// Exact target precedent gets its measured base-rate label. It still competes
	// under the same token budget and activation/token density as every other memory.
	for _, q := range query {
		for _, e := range b.Expr.Precedent(q, b.tick) {
			x := matching(e, q)
			ed := b.Expr.Edge(e.ID, x)
			text := fmt.Sprintf("[worked %.0f%% of %d tries]\n%s",
				ed.P*100, ed.N, expressionMemoryText(e))
			a := b.nodeActivation("expression:"+e.ID, b.Activation.Spans[e.SpanID])
			id := "expression:" + e.ID
			if old, ok := nodes[id]; !ok || a > old.Activation || len(text) > len(old.Text) {
				nodes[id] = MemoryNode{
					ID:         id,
					Kind:       "precedent",
					EpisodeID:  e.EpisodeID,
					SpanID:     e.SpanID,
					TokenCount: counter(text),
					Activation: a,
					Text:       text,
				}
			}
		}
	}
	out := make([]MemoryNode, 0, len(nodes))
	for _, n := range nodes {
		if n.Activation >= b.Activation.Config.Epsilon {
			out = append(out, n)
		}
	}
	return SelectMemory(out, b.Activation.Config.MemoryTokens)
}

func (b *Being) nodeActivation(id string, spanFallback float64) float64 {
	if b.Activation == nil {
		return 0
	}
	if a := b.Activation.Nodes[id]; a > 0 {
		return a
	}
	// Compatibility for durable expressions placed before node activation existed.
	return spanFallback
}

// ActiveMemory exposes the exact ephemeral projection used for inference. It is
// inspection, not durable state, and is useful for endurance tests and live traces.
func (b *Being) ActiveMemory(query []Triple) []MemoryNode {
	b.activateQuery(query)
	return b.activeMemoryNodes(query)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
