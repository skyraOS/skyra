package relation

import (
	"sync"

	"skyra-v1/src/skyra"
)

const (
	defaultAmplitude = 1.0 // initial wave amplitude — the agent's investment (constant for now)
	amplitudeDecay   = 0.9 // attenuation per depth step
)

type Relation struct {
	mu          sync.RWMutex
	id          string
	ActiveState map[string]skyra.Slice
	Amplitude   float64 // initial amplitude, set at spawn; per-node amplitude attenuates from it
}

func NewRelation() *Relation {
	return &Relation{
		ActiveState: make(map[string]skyra.Slice),
		Amplitude:   defaultAmplitude,
	}
}

func (r *Relation) ID() string {
	return r.id
}

// Realize runs a wave. origin owns it (the discriminator threaded to every
// Observe, and who the result is delivered to); traversed is the realm this
// wave walks. The slice is bound from the traversed reality's ID — no slice
// means it isn't running, and the kernel walks past it.
func (r *Relation) Realize(origin, traversed skyra.Reality, impulse any) any {
	return r.realize(origin, traversed, impulse, r.ActiveState[traversed.ID()])
}

func (r *Relation) realize(origin, traversed skyra.Reality, impulse any, sl skyra.Slice) any {
	if sl == nil {
		return nil
	}
	sl.SetCount(sl.Count() + 1) // proper time — this frame's clock

	// double buffer: last traversal's tree becomes Prior; Current rebuilds.
	tree := sl.Tree()
	tree.Prior = tree.Current
	tree.Current = nil

	// visited-once (the rail): nodes on the current path. A traversal never
	// re-descends a node already on its path, so it always terminates.
	onPath := map[string]bool{}

	traversed.Observe(origin, impulse)
	rootNode := r.rootNode(traversed, sl)
	r.scoreChildren(rootNode, sl)
	onPath[rootNode.CellID] = true
	stack := []*skyra.Node{rootNode}

	var result any

	for len(stack) > 0 {
		top := stack[len(stack)-1]

		if top.NextChild < len(top.Survivors) {
			childNode := top.Survivors[top.NextChild]
			top.NextChild++

			if onPath[childNode.CellID] {
				continue // already on the current path — the rail
			}

			onPath[childNode.CellID] = true
			childNode.Cell.Observe(origin, impulse)
			r.scoreChildren(childNode, sl)
			stack = append(stack, childNode)
		} else {
			stack = stack[:len(stack)-1]
			onPath[top.CellID] = false

			// content type check — the boundary crossing into another actor.
			// origin is threaded unchanged, so the inner actor reads the wave
			// as external. (Path A2 converts this to a goroutine spawn.)
			if inner, ok := top.Cell.Content().(skyra.Reality); ok {
				r.realize(origin, inner, impulse, r.ActiveState[inner.ID()])
			}

			out := top.Cell.Express(origin, impulse)

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, top)
			} else {
				result = out
				tree.Current = []*skyra.Node{top} // the finished tree
			}
		}
	}

	// Perception Invariant: every wave delivers its result to its origin.
	origin.Observe(origin, result)
	return result
}

// rootNode scores the root once — it has no parent to carry its score.
func (r *Relation) rootNode(cell skyra.Reality, sl skyra.Slice) *skyra.Node {
	rec := recordFor(cell, sl)
	node := &skyra.Node{
		CellID:    cell.ID(),
		Cell:      cell,
		Record:    rec,
		Amplitude: r.Amplitude,
	}
	if rec != nil {
		node.Score = rec.Activate(true, sl.Count()) * r.Amplitude
	}
	return node
}

// scoreChildren activates each child of node exactly once, builds child nodes
// for the survivors (carrying that single score and the depth-attenuated
// amplitude), and fills node.Survivors. No cell is ever activated twice within
// one parent — a survivor's Score becomes its own node score when descended.
// Amplitude folds into the score, so a quiet wave simply fails threshold.
func (r *Relation) scoreChildren(node *skyra.Node, sl skyra.Slice) {
	childAmp := node.Amplitude * amplitudeDecay
	for id, child := range node.Cell.Children() {
		rec := recordFor(child, sl)
		if rec == nil {
			continue // no record — doesn't participate in activation this frame
		}
		score := rec.Activate(true, sl.Count()) * childAmp
		if score > sl.Threshold() {
			node.Survivors = append(node.Survivors, &skyra.Node{
				CellID:    id,
				Cell:      child,
				Record:    rec,
				Score:     score,
				Amplitude: childAmp,
			})
		}
	}
}

// recordFor returns the cell's activation record, creating it on first sight.
func recordFor(cell skyra.Reality, sl skyra.Slice) skyra.ActivationRecord {
	records := sl.Records()
	rec := records[cell.ID()]
	if rec == nil {
		rec = cell.NewActivation()
		if rec != nil {
			records[cell.ID()] = rec
		}
	}
	return rec
}
