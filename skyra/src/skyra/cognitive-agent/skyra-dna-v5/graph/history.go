package graph

import (
	"sort"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Partition is one cell of a child engram's history: the invocations that
// answered the same question and came back in the same shape.
//
// Both folds are READ-TIME. Shape is derived from content, and same-question is
// a comparison rather than a stored key, so this adds no field and no migration.
// Changing what counts as the same shape invalidates nothing already written.
//
// Nothing declares the shape, because it is a property of neither side alone —
// the same expectation under Claude and under python3.11 returns different
// shapes. It belongs to THE PAIRING, so it can only be observed. The partitioned
// history IS the knowledge of what this pairing returns.
type Partition struct {
	Question string // fingerprint of the observations passed in, plus the entity
	Shape    string // fingerprint of what came back
	Entity   ids.EntityID
	Invs     []*Invocation
}

// Used counts invocations whose return the parent cited.
func (p *Partition) Used() int {
	n := 0
	for _, i := range p.Invs {
		if i.Cited {
			n++
		}
	}
	return n
}

// Diffs reports the successive changes across a partition.
//
// Within a partition, the DIFF is the informative object — not the returns.
// `480 → 500` is the signal; two absolute numbers are not.
func (p *Partition) Diffs() []Diff {
	var out []Diff
	for i := 1; i < len(p.Invs); i++ {
		a, b := p.Invs[i-1].Return, p.Invs[i].Return
		if a == nil || b == nil {
			continue
		}
		d := Diff{From: a.ID, To: b.ID, Fields: map[string][2]string{}}
		for k, bv := range b.Content {
			av, had := a.Content[k]
			if !had {
				d.Fields[k] = [2]string{"", render(bv)}
				continue
			}
			if render(av) != render(bv) {
				d.Fields[k] = [2]string{render(av), render(bv)}
			}
		}
		if len(d.Fields) > 0 {
			out = append(out, d)
		}
	}
	return out
}

type Diff struct {
	From, To ids.ObservationID
	Fields   map[string][2]string
}

func render(v content.Value) string {
	return content.Content{"v": v}.Render()
}

// Partitions folds a child engram's history twice — by question, then by shape.
//
// The entity is part of the question. Same question to Claude and to a human is a
// DIFFERENT question; one engram interpreted by two entities already accumulates
// separate histories, and entity partitioning therefore needs no extra mechanism.
//
// The host is NOT part of it. If one entity answers differently on two hosts,
// that is evidence the entity is under-specified, not a reason for a second key.
func (e *Engram) Partitions(d content.Deref) []*Partition {
	byKey := map[string]*Partition{}
	var order []string

	for _, inv := range e.History {
		cs := make([]content.Content, 0, len(inv.Observations))
		for _, o := range inv.Observations {
			cs = append(cs, o.Content)
		}
		q := string(inv.Entity) + "‖" + content.FingerprintSet(cs, d)

		shape := "<none>"
		if inv.Return != nil {
			shape = inv.Return.Shape()
		}

		key := q + "‖" + shape
		p, ok := byKey[key]
		if !ok {
			p = &Partition{Question: q, Shape: shape, Entity: inv.Entity}
			byKey[key] = p
			order = append(order, key)
		}
		p.Invs = append(p.Invs, inv)
	}

	out := make([]*Partition, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out
}

// Summary is a read-time view of what a possibility sees on one child.
//
//	here is what I have used, what it said, when it worked, and when it did not.
//
// No weights change. The entity sees a richer child set over time, which is what
// topological learning means in practice.
type Summary struct {
	Name        string
	Engram      *Engram
	Invocations int
	CitedCount  int
	Hops        map[int]int
	Partitions  []*Partition
}

// Summarize folds a child for presentation to an entity.
func Summarize(name string, e *Engram, d content.Deref) Summary {
	s := Summary{Name: name, Engram: e, Hops: map[int]int{}}
	for _, inv := range e.History {
		s.Invocations++
		if inv.Cited {
			s.CitedCount++
		}
		for h, n := range inv.HopProfile() {
			s.Hops[h] += n
		}
	}
	s.Partitions = e.Partitions(d)
	return s
}

// HopKeys returns the populated hops in order, for rendering.
func (s Summary) HopKeys() []int {
	out := make([]int, 0, len(s.Hops))
	for h := range s.Hops {
		out = append(out, h)
	}
	sort.Ints(out)
	return out
}
