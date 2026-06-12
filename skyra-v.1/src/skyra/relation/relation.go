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
	id          string
	ActiveState map[string]skyra.Slice
	Amplitude   float64        // initial amplitude, set at spawn; per-node amplitude attenuates from it
	wg          sync.WaitGroup // tracks cross-actor sub-waves spawned during this wave tree
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

// Realize runs a wave and waits for the whole wave tree (it and every
// cross-actor sub-wave it spawned) to settle. origin owns the wave (the
// discriminator threaded to every Observe, and who the result is delivered
// to); traversed is the realm this wave walks.
func (r *Relation) Realize(origin, traversed skyra.Reality, impulse any) any {
	result := r.realize(origin, traversed, impulse, r.ActiveState[traversed.ID()])
	r.wg.Wait()
	return result
}

// realize is one wave: a locked traversal, then delivery to the origin. The
// two locks are never held at once, so cross-actor waves can't deadlock.
func (r *Relation) realize(origin, traversed skyra.Reality, impulse any, sl skyra.Slice) any {
	if sl == nil {
		return nil
	}
	result := r.traverse(origin, traversed, impulse, sl)
	r.deliver(origin, result) // Perception Invariant
	return result
}

// traverse walks `traversed` under its slice lock and returns the wave's
// result. A cell whose Content is a Reality is another actor: the kernel
// spawns a goroutine for it, threading origin unchanged, and does not wait.
func (r *Relation) traverse(origin, traversed skyra.Reality, impulse any, sl skyra.Slice) any {
	sl.Lock()
	defer sl.Unlock()

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

			// boundary crossing — Content is a Reality → another actor.
			// Spawn its wave on its own goroutine, origin threaded unchanged.
			if inner, ok := top.Cell.Content().(skyra.Reality); ok {
				isl := r.ActiveState[inner.ID()]
				r.wg.Add(1)
				go func(in skyra.Reality, s skyra.Slice) {
					defer r.wg.Done()
					r.realize(origin, in, impulse, s)
				}(inner, isl)
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

	return result
}

// deliver hands the wave's result to its origin under the origin's slice lock
// (held separately from the traversed lock), so concurrent deliveries to one
// being serialize and can't race.
func (r *Relation) deliver(origin skyra.Reality, result any) {
	if osl := r.ActiveState[origin.ID()]; osl != nil {
		osl.Lock()
		defer osl.Unlock()
	}
	origin.Observe(origin, result)
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
