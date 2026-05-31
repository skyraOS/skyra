package reality

import (
	"strings"

	"skyra-v1/src/inference"
)

// Skyra DNA — the first expression of the Reality interface.
// Every field here is the same as every other DNA. The behavior
// comes from what the genome registers in ObserveFns/ExpressFns
// and where this Reality sits in the topology.
type Skyra struct {
	id             string  // unique per instance, assigned at creation
	BeingID        string  // which being this cell belongs to — used to get the right Slice on the Relation
	Weight         float64 // global weight — EMA across all traversals
	TraversalCount int     // proper time — how many relations have passed through
	LastTraversed  int     // traversal count when last activated
	Alpha          float64 // EMA smoothing factor — slow = stable, fast = reactive
	Type           string  // set by Competence once on descent (before Observe), once on ascent (before Express). One field, two reads. Observe and Express have independent type vocabularies but never fire at the same time.
	Content        string    // what this cell holds — identity for PFC, memory for memory, skill for skill
	Vector         []float64 // embedding of Content — set once at creation, used for relevance in activation
	Description    string    // what this cell type is responsible for — shapes traversal, not shown during inference

	// the topology — where this Reality sits in the graph
	Relationships map[string]Reality // descent targets — what this node knows
	Expressors    map[string]Reality // ascent targets — what this node can do

	// the DNA — keyed by Type, selected by Competence
	ObserveFns map[string]func(*Relation)            // intake behaviors per type — needed: pass-through, pick-up, propagate
	ExpressFns map[string]func(*Relation) *Relation  // output behaviors per type
	// expressors needed: motor (act), compress (provider call — collapse accumulated context)
}

func (s *Skyra) ID() string {
	return s.id
}

// Create — instantiate a new cell with this DNA. The genome calls this.
func (s *Skyra) Create(r *Relation) Reality {
	return &Skyra{
		id:            NewID(),
		Weight:        1.0,
		Alpha:         0.1,
		Relationships: make(map[string]Reality),
		Expressors:    make(map[string]Reality),
		ObserveFns:    make(map[string]func(*Relation)),
		ExpressFns:    make(map[string]func(*Relation) *Relation),
	}
}

// Competence — the cell reads itself before processing.
// Called twice per traversal: once on descent (sets observe type),
// once on ascent (sets express type). Same field, two vocabularies.
func (s *Skyra) Competence(r *Relation) {
	activation := s.Activation(r)

	// thresholds are placeholders
	switch {
	case activation > 0.6:
		s.Type = "propagate"
	case activation > 0.2:
		s.Type = "pick-up"
	default:
		s.Type = "pass-through"
	}
}

// CompetenceExpress — ascent competence read. Decides whether
// to compress based on thought density on the Relation.
func (s *Skyra) CompetenceExpress(r *Relation) {
	density := ThoughtDensity(r, s.BeingID)

	switch {
	case density > 0.5:
		s.Type = "compress"
	default:
		s.Type = "pass-through"
	}
	// motor fires at the exit of Realize(), not through competence
}

// ThoughtDensity — average pairwise cosine similarity across
// thought vectors on the Relation. High = through line exists.
// Low = scattered context.
func ThoughtDensity(r *Relation, beingID string) float64 {
	sl := r.GetSlice(beingID)
	sl.mu.Lock()
	thoughts := make([]*Skyra, 0, len(sl.Thoughts))
	for _, t := range sl.Thoughts {
		if cell, ok := t.(*Skyra); ok && len(cell.Vector) > 0 {
			thoughts = append(thoughts, cell)
		}
	}
	sl.mu.Unlock()

	if len(thoughts) < 2 {
		return 0
	}

	var sum float64
	pairs := 0
	for i := 0; i < len(thoughts); i++ {
		for j := i + 1; j < len(thoughts); j++ {
			sum += inference.Cosine(thoughts[i].Vector, thoughts[j].Vector)
			pairs++
		}
	}
	return sum / float64(pairs)
}

// Realize — one traversal through this cell.
// Competence → Observe → descent → ascent → Express → weight update.
// The Relation enters and leaves transformed. So does the cell.
func (s *Skyra) Realize(r *Relation) *Relation {
	if r.Cancelled {
		s.Decay()
		return r
	}
	sl := r.GetSlice(s.BeingID)
	sl.mu.Lock()
	if sl.Visited[s.id] {
		sl.mu.Unlock()
		return r
	}
	sl.Visited[s.id] = true
	sl.mu.Unlock()

	// first competence read — what am I on intake?
	s.Competence(r)

	// descent phase — the cell absorbs from the signal
	s.Observe(r)

	// deeper into the topology — promoted nodes first, then memories
	for _, rel := range s.Relationships {
		r = rel.Realize(r)
	}

	// decay unvisited neighbors — cancelled traversal, one level deep
	for id, rel := range s.Relationships {
		sl.mu.Lock()
		visited := sl.Visited[id]
		sl.mu.Unlock()
		if !visited {
			rel.Realize(&Relation{Cancelled: true})
		}
	}

	// ascent phase — expressors fire on the way back up
	for _, e := range s.Expressors {
		r = e.Realize(r)
	}

	// second competence read — ascent vocabulary: motor or compress based on thought density
	s.CompetenceExpress(r)

	// the cell outputs
	r = s.Express(r)

	// the cell ages — proper time increments, weights update
	s.TraversalCount++
	s.LastTraversed = s.TraversalCount
	s.Reinforce()

	// act — parse intent graph, add acts to the Relation
	sl.mu.Lock()
	intent := sl.Intent
	sl.mu.Unlock()
	if intent != "" {
		acts := parseIntentGraph(intent)
		for _, act := range acts {
			a := s.Create(&Relation{}).(*Skyra)
			a.Content = act
			sl.mu.Lock()
			sl.Thoughts = append(sl.Thoughts, a)
			sl.mu.Unlock()
		}
	}

	// motor — fires at the exit, executes the result of the traversal
	if fn, ok := s.ExpressFns["motor"]; ok {
		r = fn(r)
	}

	return r
}

// parseIntentGraph splits the intent graph into individual actions.
// <then> is the delimiter between sequential steps.
// Actions without <then> between them are part of the same step.
func parseIntentGraph(intent string) []string {
	steps := strings.Split(intent, "<then>")
	var acts []string
	for _, step := range steps {
		step = strings.TrimSpace(step)
		if step != "" {
			acts = append(acts, step)
		}
	}
	return acts
}

// Observe — mutual observation. The Relation observes itself through
// the encounter first, then the cell's type-specific intake fires.
func (s *Skyra) Observe(r *Relation) {
	sl := r.GetSlice(s.BeingID)
	sl.mu.Lock()
	sl.LastSeen = s
	if sl.Forming != nil {
		sl.Forming.Relationships[s.id] = s
	}
	sl.mu.Unlock()
	r.Observe(r)
	if fn, ok := s.ObserveFns[s.Type]; ok {
		fn(r)
	}
}

// ObservePassThrough — dormant. Don't touch the Relation.
func ObservePassThrough(s *Skyra) func(*Relation) {
	return func(r *Relation) {}
}

// ObservePickUp — absorb content from the cell onto the Relation as a thought cell. Costs signal.
func ObservePickUp(s *Skyra) func(*Relation) {
	return func(r *Relation) {
		if s.Content == "" {
			return
		}
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		sl.Thoughts = append(sl.Thoughts, s)
		sl.Load += len(s.Content)
		sl.Signal -= 0.01
		sl.mu.Unlock()
	}
}

// ObservePropagate — absorb content onto the Relation without costing signal.
// Also reads Goal and Intent to shape descent. If the cell's content updates
// the being's understanding of goal or route, propagate can refine them.
func ObservePropagate(s *Skyra, call func(string, string) (string, error)) func(*Relation) {
	return func(r *Relation) {
		if s.Content == "" {
			return
		}
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		sl.Thoughts = append(sl.Thoughts, s)
		sl.Load += len(s.Content)
		goal := sl.Goal
		intent := sl.Intent
		sl.mu.Unlock()

		if goal == "" {
			return
		}

		prompt := `You are a being navigating a world. You have a goal and an intent graph.

Your goal is what you are trying to reach.
Your intent graph is your prediction of the future — the sequence of actions you believe will get you to the goal. When you update it, it overwrites the current one and shapes what happens next.

You just observed something. Think about what it means for your goal and your predicted future.

Respond with:
<thought>what you think about this observation</thought>
Optionally, if your prediction of the future has changed:
<goal>updated goal</goal>
<intent>updated intent graph</intent>`
		context := "Goal: " + goal + "\nCurrent intent: " + intent + "\nObserved: " + s.Content

		result, err := call(prompt, context)
		if err != nil {
			return
		}

		thought := extractTag(result, "thought")
		newGoal := extractTag(result, "goal")
		newIntent := extractTag(result, "intent")

		sl.mu.Lock()
		if thought != "" {
			t := s.Create(&Relation{}).(*Skyra)
			t.Content = thought
			sl.Thoughts = append(sl.Thoughts, t)
			sl.Load += len(thought)
		}
		if newGoal != "" {
			sl.Goal = newGoal
		}
		if newIntent != "" {
			sl.Intent = newIntent
		}
		sl.mu.Unlock()
	}
}

// ExpressMotor — execute. The leaf where cognition crosses into consequence.
func ExpressMotor(s *Skyra, execute func(string) string) func(*Relation) *Relation {
	return func(r *Relation) *Relation {
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		var parts []string
		for _, t := range sl.Thoughts {
			if cell, ok := t.(*Skyra); ok {
				parts = append(parts, cell.Content)
			}
		}
		sl.mu.Unlock()
		if len(parts) > 0 {
			result := execute(strings.Join(parts, "\n"))
			if result != "" {
				out := s.Create(&Relation{}).(*Skyra)
				out.Content = result
				sl.mu.Lock()
				sl.Thoughts = []Reality{out}
				sl.Load = len(result)
				sl.mu.Unlock()
			}
		}
		return r
	}
}

// ExpressCompress — provider call that collapses accumulated thoughts.
// Fires when Load approaches MaxLoad. Groups similar thoughts by cosine
// similarity, compresses each cluster via provider, replaces raw thoughts
// with compressed cells on the Relation.
func ExpressCompress(s *Skyra, call func(string, string) (string, error)) func(*Relation) *Relation {
	return func(r *Relation) *Relation {
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		if sl.Load < r.MaxLoad/2 || len(sl.Thoughts) < 2 {
			sl.mu.Unlock()
			return r
		}

		thoughts := make([]*Skyra, 0, len(sl.Thoughts))
		for _, t := range sl.Thoughts {
			if cell, ok := t.(*Skyra); ok {
				thoughts = append(thoughts, cell)
			}
		}
		sl.mu.Unlock()

		if len(thoughts) < 2 {
			return r
		}

		clusters := clusterBySimilarity(thoughts, 0.6)

		var compressed []Reality
		totalLoad := 0

		for _, cluster := range clusters {
			if len(cluster) == 1 {
				compressed = append(compressed, cluster[0])
				totalLoad += len(cluster[0].Content)
				continue
			}

			var parts []string
			for _, cell := range cluster {
				parts = append(parts, cell.Content)
			}
			raw := strings.Join(parts, "\n---\n")

			sl.mu.Lock()
			goal := sl.Goal
			intent := sl.Intent
			sl.mu.Unlock()

			prompt := `You are a being navigating a world. You have a goal and an intent graph.

Your goal is what you are trying to reach.
Your intent graph is your prediction of the future — the sequence of actions you believe will get you to the goal. When you update it, it overwrites the current one and shapes what happens next.

You just gathered context during a traversal. Compress it — find the through line, drop redundancy. Then think about what survived and what it means for your predicted future.

Respond with:
<compressed>dense summary preserving the through line</compressed>
<thought>what you think about what survived</thought>
Optionally, if your prediction of the future has changed:
<goal>updated goal</goal>
<intent>updated intent graph</intent>`
			if goal != "" {
				raw = "Goal: " + goal + "\nCurrent intent: " + intent + "\n---\n" + raw
			}

			result, err := call(prompt, raw)
			if err != nil {
				for _, cell := range cluster {
					compressed = append(compressed, cell)
					totalLoad += len(cell.Content)
				}
				continue
			}

			compressedContent := extractTag(result, "compressed")
			if compressedContent == "" {
				compressedContent = result
			}
			thought := extractTag(result, "thought")
			newGoal := extractTag(result, "goal")
			newIntent := extractTag(result, "intent")

			sl.mu.Lock()
			if thought != "" {
				t := s.Create(&Relation{}).(*Skyra)
				t.Content = thought
				sl.Thoughts = append(sl.Thoughts, t)
			}
			if newGoal != "" {
				sl.Goal = newGoal
			}
			if newIntent != "" {
				sl.Intent = newIntent
			}
			sl.mu.Unlock()

			out := s.Create(&Relation{}).(*Skyra)
			out.Content = compressedContent
			for _, cell := range cluster {
				out.Relationships[cell.id] = cell
			}
			vec, err := inference.Embed(compressedContent)
			if err == nil {
				out.Vector = vec
			}

			s.Relationships[out.id] = out
			compressed = append(compressed, out)
			totalLoad += len(compressedContent)
		}

		sl.mu.Lock()
		sl.Thoughts = compressed
		sl.Load = totalLoad
		sl.mu.Unlock()

		return r
	}
}

// clusterBySimilarity groups thought cells by cosine similarity.
// Cells without vectors go into their own singleton cluster.
func clusterBySimilarity(thoughts []*Skyra, threshold float64) [][]*Skyra {
	assigned := make([]bool, len(thoughts))
	var clusters [][]*Skyra

	for i := 0; i < len(thoughts); i++ {
		if assigned[i] {
			continue
		}
		cluster := []*Skyra{thoughts[i]}
		assigned[i] = true

		if len(thoughts[i].Vector) == 0 {
			clusters = append(clusters, cluster)
			continue
		}

		for j := i + 1; j < len(thoughts); j++ {
			if assigned[j] || len(thoughts[j].Vector) == 0 {
				continue
			}
			sim := inference.Cosine(thoughts[i].Vector, thoughts[j].Vector)
			if sim >= threshold {
				cluster = append(cluster, thoughts[j])
				assigned[j] = true
			}
		}
		clusters = append(clusters, cluster)
	}
	return clusters
}

// Express — mutual expression. The Relation expresses itself first,
// then the cell's type-specific output fires.
func (s *Skyra) Express(r *Relation) *Relation {
	r = r.Express(r)
	if fn, ok := s.ExpressFns[s.Type]; ok {
		r = fn(r)
	}
	return r
}

// Activation — base weight * recency * product of all binding fields.
// This is the score that Competence reads to determine what the cell is.
// TODO: add relevance term — cosine similarity between cell's content embedding and Relation's impulse embedding. Local model (nomic-embed-text or ONNX in-process). Embed once at cell creation, store vector on cell.
func (s *Skyra) Activation(rel *Relation) float64 {
	base := s.Weight * Recency(s.TraversalCount, s.LastTraversed)
	if field, ok := rel.Fields[s.id]; ok {
		for _, w := range field {
			base *= w
		}
	}
	return base
}

// Reinforce — EMA update on the return path. The cell was traversed,
// so its weight moves toward 1.0. Alpha controls how fast.
func (s *Skyra) Reinforce() {
	s.Weight = s.Alpha*1.0 + (1-s.Alpha)*s.Weight
}

// Decay — EMA update when NOT traversed. Weight drifts toward 0.
// Starvation is built into the formula.
func (s *Skyra) Decay() {
	s.Weight = (1 - s.Alpha) * s.Weight
}
