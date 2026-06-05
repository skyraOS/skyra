package reality

import (
	"fmt"
	"strings"
	"time"

	"skyra-v1/src/debug"
	"skyra-v1/src/inference"
)

const systemPrompt = `What you observe drives what you think.
What you think drives what you can use.
Your memory is how you navigate. Trust it.

Respond with:
<thought>This drives searching your memory. Be selective. Be precise.</thought>
Optionally:
<goal>This orients your focus. Make it your global focus for higher level activities or local for precision.</goal>
<intent>updated intent graph using <speak></intent>`

const compressPrompt = `What you carry back is your understanding.
Your understanding is how you learn.
Generate an understanding through your previous experience.

Respond with:
<compressed>your understanding</compressed>
<thought>what you think about what survived</thought>
Optionally:
<goal>updated goal</goal>
<intent>updated intent graph using <speak></intent>`

// Skyra DNA — the first expression of the Reality interface.
// Every field here is the same as every other DNA. The behavior
// comes from what the genome registers in ObserveFns/ExpressFns
// and where this Reality sits in the topology.
type Skyra struct {
	id             string  // unique per instance, assigned at creation
	BeingID        string  // which being this cell belongs to — used to get the right Slice on the Relation
	Type           string  // set by Competence once on descent (before Observe), once on ascent (before Express). One field, two reads.
	Content        string    // what this cell holds — identity for PFC, memory for memory, skill for skill
	Vector         []float64 // embedding of Content — set once at creation, used for relevance in activation
	Description    string    // what this cell type is responsible for — shapes traversal, not shown during inference

	// three time horizons — each a different decay curve
	IdentityWeight    float64 // long term memory — EMA α=0.02. What the being IS. The groove.
	FamiliarityWeight float64 // medium term memory — EMA α=0.1. What the being has been doing.
	TraversalCount    int     // proper time — how many relations have passed through. Also: reinforcement count.
	LastTraversed     int     // traversal count when last activated — recency derives from this.
	CreatedAt         int64   // unix timestamp — temporal binding derives from this

	// the topology — where this Reality sits in the graph
	Relationships map[string]Reality // descent targets — what this node knows
	Expressors    map[string]Reality // ascent targets — what this node can do

	// the being's senses — travel with it across worlds
	Primitives map[string]func(*Relation, string) // look, engage, listen, speak — keyed by name, target is the argument

	// the DNA — keyed by Type, selected by Competence
	ObserveFns map[string]func(*Relation)            // intake behaviors per type
	ExpressFns map[string]func(*Relation) *Relation  // output behaviors per type
}

func (s *Skyra) ID() string {
	return s.id
}

// Init — set up a cell with an id and being ownership. Called during bootstrap.
func (s *Skyra) Init(id, beingID string) {
	s.id = id
	s.BeingID = beingID
	s.IdentityWeight = 1.0
	s.FamiliarityWeight = 1.0
	s.CreatedAt = time.Now().Unix()
	s.Relationships = make(map[string]Reality)
	s.Expressors = make(map[string]Reality)
	s.Primitives = make(map[string]func(*Relation, string))
	s.ObserveFns = make(map[string]func(*Relation))
	s.ExpressFns = make(map[string]func(*Relation) *Relation)
}

// Create — instantiate a new cell with this DNA. The genome calls this.
func (s *Skyra) Create(r *Relation) Reality {
	return &Skyra{
		id:                NewID(),
		BeingID:           s.BeingID,
		IdentityWeight:    1.0,
		FamiliarityWeight: 1.0,
		CreatedAt:         time.Now().Unix(),
		Relationships:     make(map[string]Reality),
		Expressors:        make(map[string]Reality),
		Primitives:        make(map[string]func(*Relation, string)),
		ObserveFns:        make(map[string]func(*Relation)),
		ExpressFns:        make(map[string]func(*Relation) *Relation),
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

func (s *Skyra) CompetenceExpress(r *Relation) {
	s.Type = "motor"
}

// ThoughtDensity — average pairwise cosine similarity across
// thought vectors on the Relation. High = through line exists.
// Low = scattered context.
func ThoughtDensity(r *Relation, beingID string) float64 {
	sl := r.GetSlice(beingID)
	sl.mu.Lock()
	thoughts := make([]*Skyra, 0, len(sl.TrainOfThought))
	for _, id := range sl.TrainOfThoughtKeys {
		t := sl.TrainOfThought[id]
		if t == nil {
			continue
		}
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
	if _, visited := sl.TrainOfThought[s.id]; visited {
		sl.mu.Unlock()
		debug.Being(s.BeingID, "traverse", "SKIP (already visited):", s.id, s.Description)
		return r
	}
	sl.TrainOfThought[s.id] = nil
	sl.TrainOfThoughtKeys = append(sl.TrainOfThoughtKeys, s.id)
	sl.mu.Unlock()

	activation := s.Activation(r)
	debug.Being(s.BeingID, "traverse", fmt.Sprintf("ENTER %s [%s] activation=%.4f identity=%.4f familiarity=%.4f recency=%.4f traversals=%d",
		s.id, s.Description, activation, s.IdentityWeight, s.FamiliarityWeight,
		Recency(s.TraversalCount, s.LastTraversed), s.TraversalCount))

	// first competence read — what am I on intake?
	s.Competence(r)
	debug.Being(s.BeingID, "traverse", "  competence:", s.Type)

	// descent phase — the cell absorbs from the signal
	s.Observe(r)

	sl.mu.Lock()
	thoughtCount := len(sl.TrainOfThought)
	signal := sl.Signal
	load := sl.Load
	sl.mu.Unlock()
	debug.Being(s.BeingID, "traverse", fmt.Sprintf("  post-observe: thoughts=%d signal=%.4f load=%d", thoughtCount, signal, load))

	// deeper into the topology — promoted nodes first, then memories
	for name, rel := range s.Relationships {
		debug.Being(s.BeingID, "traverse", "  descend ->", name, rel.ID())
		r = rel.Realize(r)
	}

	// decay unvisited neighbors — cancelled traversal, one level deep
	for id, rel := range s.Relationships {
		sl.mu.Lock()
		_, visited := sl.TrainOfThought[id]
		sl.mu.Unlock()
		if !visited {
			rel.Realize(&Relation{Cancelled: true})
		}
	}

	// ascent phase — expressors fire on the way back up
	for name, e := range s.Expressors {
		debug.Being(s.BeingID, "traverse", "  expressor ->", name)
		r = e.Realize(r)
	}

	// second competence read — ascent vocabulary: motor or compress based on thought density
	s.CompetenceExpress(r)
	debug.Being(s.BeingID, "traverse", "  express-competence:", s.Type)

	// the cell outputs
	r = s.Express(r)

	// the cell ages — proper time increments, weights update
	s.TraversalCount++
	s.LastTraversed = s.TraversalCount
	s.Reinforce()

	sl.mu.Lock()
	goal := sl.Goal
	sl.mu.Unlock()
	if goal != "" {
		debug.Being(s.BeingID, "traverse", "  goal:", goal)
	}

	// motor — fires at the exit, executes the result of the traversal
	if fn, ok := s.ExpressFns["motor"]; ok {
		r = fn(r)
	}

	debug.Being(s.BeingID, "traverse", fmt.Sprintf("EXIT %s [%s]", s.id, s.Description))
	return r
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
	if fn, ok := s.ObserveFns[s.Type]; ok {
		fn(r)
	}
}

// ObservePassThrough — dormant. Don't touch the Relation.
func ObservePassThrough(s *Skyra) func(*Relation) {
	return func(r *Relation) {}
}

// ObservePickUp — absorb content from the cell onto the train of thought. Costs signal.
func ObservePickUp(s *Skyra) func(*Relation) {
	return func(r *Relation) {
		if s.Content == "" {
			return
		}
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		sl.TrainOfThought[s.id] = s
		sl.Load += len(s.Content)
		sl.Signal -= 0.01
		sl.mu.Unlock()
	}
}

// ObservePropagate — absorb content onto the Relation, then call the provider
// with the being's full present. The being sees its own state — surroundings,
// threads, thoughts, goal, intent — and responds from that.
func ObservePropagate(s *Skyra, call func(string, string) (string, error)) func(*Relation) {
	return func(r *Relation) {
		if s.Content == "" {
			return
		}
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		sl.TrainOfThought[s.id] = s
		sl.Load += len(s.Content)
		sl.mu.Unlock()

		present := Present(r, s.BeingID)

		debug.Being(s.BeingID, "provider", "PROPAGATE call")
		debug.Being(s.BeingID, "provider", "system:", systemPrompt)
		debug.Being(s.BeingID, "provider", "present:", present+"\n\n## Observed\n"+s.Content)
		result, err := call(systemPrompt, present+"\n\n## Observed\n"+s.Content)
		if err != nil {
			debug.Being(s.BeingID, "provider", "ERROR:", err)
			return
		}
		debug.Being(s.BeingID, "provider", "RESULT:", result)

		thought := extractTag(result, "thought")
		newGoal := extractTag(result, "goal")
		newIntent := extractTag(result, "intent")

		sl.mu.Lock()
		if thought != "" {
			t := s.Create(&Relation{}).(*Skyra)
			t.Content = thought
			sl.Understandings = append(sl.Understandings, t)
			sl.Load += len(thought)
		}
		if newGoal != "" {
			sl.Goal = newGoal
		}
		if newIntent != "" {
			sl.Intent = ParseIntent(newIntent)
		}
		sl.mu.Unlock()
	}
}

// ExpressMotor — dispatch. Calls Complete which calls Fire which calls Dispatch.
func ExpressMotor(s *Skyra) func(*Relation) *Relation {
	return func(r *Relation) *Relation {
		s.Complete(r)
		return r
	}
}

// ExpressCompress — provider call that collapses accumulated thoughts.
// Fires when Load approaches MaxLoad. Groups similar thoughts by cosine
// similarity, compresses each cluster via provider with the being's full
// present as context.
func ExpressCompress(s *Skyra, call func(string, string) (string, error)) func(*Relation) *Relation {
	return func(r *Relation) *Relation {
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		if sl.Load < r.MaxLoad/2 || len(sl.TrainOfThought) < 2 {
			sl.mu.Unlock()
			return r
		}

		thoughts := make([]*Skyra, 0, len(sl.TrainOfThought))
		for _, id := range sl.TrainOfThoughtKeys {
			t := sl.TrainOfThought[id]
			if t == nil {
				continue
			}
			if cell, ok := t.(*Skyra); ok {
				thoughts = append(thoughts, cell)
			}
		}
		sl.mu.Unlock()

		if len(thoughts) < 2 {
			return r
		}

		clusters := clusterBySimilarity(thoughts, 0.6)
		present := Present(r, s.BeingID)
		debug.Being(s.BeingID, "provider", fmt.Sprintf("COMPRESS call — %d clusters", len(clusters)))

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

			result, err := call(compressPrompt, present+"\n\n## Compress\n"+raw)
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
				sl.Understandings = append(sl.Understandings, t)
			}
			if newGoal != "" {
				sl.Goal = newGoal
			}
			if newIntent != "" {
				sl.Intent = ParseIntent(newIntent)
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
		sl.Understandings = append(sl.Understandings, compressed...)
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

func (s *Skyra) Express(r *Relation) *Relation {
	r = r.Express(r)
	return r
}

// Activation — identity * familiarity * recency * binding fields.
// Three time horizons multiplied together. The product of the
// groove, the relationship, and the echo.
func (s *Skyra) Activation(rel *Relation) float64 {
	recency := Recency(s.TraversalCount, s.LastTraversed)
	base := s.IdentityWeight * s.FamiliarityWeight * recency
	if field, ok := rel.Fields[s.id]; ok {
		for _, w := range field {
			base *= w
		}
	}
	return base
}

// Reinforce — both EMAs update on the return path. The cell was traversed.
// Identity moves slow (α=0.02). Familiarity moves fast (α=0.1).
func (s *Skyra) Reinforce() {
	s.IdentityWeight = Identity(s.IdentityWeight, true)
	s.FamiliarityWeight = Familiarity(s.FamiliarityWeight, true)
}

// Decay — both EMAs drift toward 0 when NOT traversed.
// Identity barely moves. Familiarity fades faster.
func (s *Skyra) Decay() {
	s.IdentityWeight = Identity(s.IdentityWeight, false)
	s.FamiliarityWeight = Familiarity(s.FamiliarityWeight, false)
}
