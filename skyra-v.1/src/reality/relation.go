package reality

import "sync"

// Node — the kernel's traversal frame. Wraps a cell and its activation record.
type Node struct {
	CellID    string
	Cell      Reality
	Record    ActivationRecord
	Score     float64
	Children  []*Node   // all children — above and below threshold
	Survivors []Reality // above threshold — queued for descent
	NextChild int       // index into Survivors for iterative DFS
}

type TraversalPair struct {
	Current []*Node
	Prior   []*Node
}

type Slice struct {
	mu             sync.Mutex
	TraversalCount int                       // monotonic global clock — increments once per BeginTraversal
	Records        map[string]ActivationRecord // cellID → persistent activation record
	LastChildren   map[string][]string       // cellID → child IDs from last observation
	Attention      map[string]float64        // filtered subset of surroundings — what survived the prediction gate
	Threshold      float64                   // DNA-defined minimum score to descend
	Stack          []*Node                   // the DFS state machine
	Trees          map[string]*TraversalPair // keyed by binding reality ID
}

type Relation struct {
	mu     sync.RWMutex
	id     string
	Slices map[string]*Slice
	Energy float64
}

func (r *Relation) ID() string {
	return r.id
}

// Realize — the kernel. Iterative DFS. Calls Observe on descent, Express on ascent.
func (r *Relation) Realize(beingID string, brID string, root Reality) {
	sl := r.GetSlice(beingID)

	sl.mu.Lock()
	sl.TraversalCount++
	pair := sl.Trees[brID]
	if pair == nil {
		pair = &TraversalPair{}
		sl.Trees[brID] = pair
	}
	pair.Prior = pair.Current
	pair.Current = nil
	sl.Stack = nil
	sl.mu.Unlock()

	root.Observe(r)
	rootNode := r.observe(beingID, root.ID(), root, root.Children())
	if rootNode == nil {
		return
	}

	sl.mu.Lock()
	sl.Stack = append(sl.Stack, rootNode)
	sl.mu.Unlock()

	for {
		sl.mu.Lock()
		if len(sl.Stack) == 0 {
			sl.mu.Unlock()
			break
		}

		top := sl.Stack[len(sl.Stack)-1]

		if top.NextChild < len(top.Survivors) && r.Energy > 0 {
			child := top.Survivors[top.NextChild]
			top.NextChild++
			sl.mu.Unlock()

			r.Energy--
			child.Observe(r)
			childNode := r.observe(beingID, child.ID(), child, child.Children())

			sl.mu.Lock()
			if childNode != nil {
				sl.Stack = append(sl.Stack, childNode)
			}
			sl.mu.Unlock()
		} else {
			sl.Stack = sl.Stack[:len(sl.Stack)-1]
			sl.mu.Unlock()

			top.Cell.Express(r)

			sl.mu.Lock()
			if len(sl.Stack) > 0 {
				parent := sl.Stack[len(sl.Stack)-1]
				parent.Children = append(parent.Children, top)
			} else {
				pair.Current = append(pair.Current, top)
			}
			sl.mu.Unlock()
		}
	}
}

// observe — scores children via their activation records. The kernel owns traversal, cells own scoring.
func (r *Relation) observe(beingID string, cellID string, cell Reality, cells map[string]Reality) *Node {
	sl := r.GetSlice(beingID)
	sl.mu.Lock()
	defer sl.mu.Unlock()

	var children []*Node
	var survivors []Reality

	// Score present children
	for id, child := range cells {
		rec := sl.Records[id]
		if rec == nil {
			rec = child.NewActivation()
			sl.Records[id] = rec
		}
		score := rec.Activate(true, sl.TraversalCount)

		children = append(children, &Node{
			CellID: id,
			Cell:   child,
			Record: rec,
			Score:  score,
		})

		if score > sl.Threshold {
			survivors = append(survivors, child)
		}
	}

	// Decay absent children
	if prior, ok := sl.LastChildren[cellID]; ok {
		present := make(map[string]bool, len(cells))
		for id := range cells {
			present[id] = true
		}
		for _, childID := range prior {
			if !present[childID] {
				rec := sl.Records[childID]
				if rec != nil {
					score := rec.Activate(false, sl.TraversalCount)
					children = append(children, &Node{
						CellID: childID,
						Record: rec,
						Score:  score,
					})
				}
			}
		}
	}

	// Update last children for next traversal
	childIDs := make([]string, 0, len(cells))
	for id := range cells {
		childIDs = append(childIDs, id)
	}
	sl.LastChildren[cellID] = childIDs

	// Self
	selfRec := sl.Records[cellID]
	if selfRec == nil {
		selfRec = cell.NewActivation()
		sl.Records[cellID] = selfRec
	}
	selfScore := selfRec.Activate(true, sl.TraversalCount)

	return &Node{
		CellID:    cellID,
		Cell:      cell,
		Record:    selfRec,
		Score:     selfScore,
		Children:  children,
		Survivors: survivors,
	}
}

func (r *Relation) GetSlice(beingID string) *Slice {
	r.mu.RLock()
	s, ok := r.Slices[beingID]
	r.mu.RUnlock()
	if ok {
		return s
	}

	s = &Slice{
		Records:      make(map[string]ActivationRecord),
		LastChildren: make(map[string][]string),
		Attention:    make(map[string]float64),
		Trees:        make(map[string]*TraversalPair),
	}

	r.mu.Lock()
	r.Slices[beingID] = s
	r.mu.Unlock()
	return s
}
