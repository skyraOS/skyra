package relation

import (
	"sync"

	"skyra-v1/src/skyra"
)

type Relation struct {
	mu          sync.RWMutex
	id          string
	ActiveState map[string]skyra.Slice
	Energy      float64
}

func NewRelation() *Relation {
	return &Relation{
		ActiveState: make(map[string]skyra.Slice),
	}
}

func (r *Relation) ID() string {
	return r.id
}

// Realize — the wave hits a reality node, calls ID on it, and that
// lookup binds the frame. No slice in ActiveState means the thing
// isn't running — the kernel walks past it.
func (r *Relation) Realize(root skyra.Reality, payload any) any {
	return r.realize(root, payload, r.ActiveState[root.ID()])
}

func (r *Relation) realize(root skyra.Reality, payload any, sl skyra.Slice) any {
	if sl == nil {
		return nil
	}
	sl.SetCount(sl.Count() + 1) // proper time — this frame's clock

	var stack []*skyra.Node

	root.Observe(r, payload)
	rootNode := r.score(root, sl)
	stack = append(stack, rootNode)

	var result any

	for len(stack) > 0 {
		top := stack[len(stack)-1]

		if top.NextChild < len(top.Survivors) && r.Energy > 0 {
			child := top.Survivors[top.NextChild]
			top.NextChild++
			r.Energy--

			child.Observe(r, payload)
			childNode := r.score(child, sl)
			stack = append(stack, childNode)
		} else {
			stack = stack[:len(stack)-1]

			// content type check — the boundary crossing.
			// Content is a Reality → ID() is the ActiveState lookup,
			// and the inner traversal runs in the inner frame.
			// Anything else (string, state, nil) is the cell's own business.
			if inner, ok := top.Cell.Content().(skyra.Reality); ok {
				r.realize(inner, payload, r.ActiveState[inner.ID()])
			}

			// the return is the kernel's channel: primitive requests
			// (dispatched once the world is coupled) and the root's
			// result, which becomes Realize's return. Cell-to-cell
			// data travels through the payload, not through returns.
			out := top.Cell.Express(r, payload)

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, top)
			} else {
				result = out
			}
		}
	}

	return result
}

func (r *Relation) score(cell skyra.Reality, sl skyra.Slice) *skyra.Node {
	records := sl.Records()

	var survivors []skyra.Reality
	for id, child := range cell.Children() {
		rec := records[id]
		if rec == nil {
			rec = child.NewActivation()
			records[id] = rec
		}
		// no activation record — the cell doesn't participate in
		// activation in this frame. not a survivor, not a panic.
		if rec == nil {
			continue
		}
		if rec.Activate(true, sl.Count()) > sl.Threshold() {
			survivors = append(survivors, child)
		}
	}

	selfRec := records[cell.ID()]
	if selfRec == nil {
		selfRec = cell.NewActivation()
		if selfRec != nil {
			records[cell.ID()] = selfRec
		}
	}

	node := &skyra.Node{
		CellID:    cell.ID(),
		Cell:      cell,
		Record:    selfRec,
		Survivors: survivors,
	}
	if selfRec != nil {
		node.Score = selfRec.Activate(true, sl.Count())
	}
	return node
}
