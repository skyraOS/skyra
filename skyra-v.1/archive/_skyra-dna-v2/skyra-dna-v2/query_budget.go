package dnav2

import (
	"fmt"
	"math"
	"sync"
)

// QueryBudgetConfig contains tuning values, not architectural constants.
type QueryBudgetConfig struct {
	Min     float64
	Max     float64
	Initial float64
	Alpha   float64
	Beta    float64
}

func DefaultQueryBudgetConfig() QueryBudgetConfig {
	return QueryBudgetConfig{Min: 32, Max: 256, Initial: 128, Alpha: 10, Beta: 0.5}
}

func (c QueryBudgetConfig) validate() error {
	if c.Min <= 0 || c.Max < c.Min || c.Initial < c.Min || c.Initial > c.Max {
		return fmt.Errorf("invalid query budget bounds: min=%v initial=%v max=%v", c.Min, c.Initial, c.Max)
	}
	if c.Alpha <= 0 || c.Beta <= 0 || c.Beta >= 1 {
		return fmt.Errorf("invalid AIMD tuning: alpha=%v beta=%v", c.Alpha, c.Beta)
	}
	return nil
}

// QueryBudgetEvent is an immutable instrumentation record. HasGrade separates
// explicit grade zero from no-grade. EdgeID is sample provenance: the
// contextual edge whose grade produced this update, not the budget's key.
type QueryBudgetEvent struct {
	Lineage  string
	EdgeID   string
	Before   float64
	After    float64
	Grade    float64
	HasGrade bool
	Sequence uint64
}

type queryBudgetState struct {
	current float64
	history []QueryBudgetEvent
}

// QueryBudgetStore owns one independent controller per emitting LINEAGE — the
// Efficient and Final worker voices. The budget is a property of the querying
// machinery, not of any contextual edge: edges die with their sets (item 6,
// option B) while the lineage carries its earned specificity across sets.
// Grades arrive per-edge and update the owning lineage's controller.
// Material/Formal has no budget because it is structurally updated and never
// signed-graded.
type QueryBudgetStore struct {
	mu        sync.Mutex
	config    QueryBudgetConfig
	byLineage map[string]*queryBudgetState
	seq       uint64
}

func NewQueryBudgetStore(config QueryBudgetConfig) (*QueryBudgetStore, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &QueryBudgetStore{config: config, byLineage: map[string]*queryBudgetState{}}, nil
}

func MustNewQueryBudgetStore(config QueryBudgetConfig) *QueryBudgetStore {
	store, err := NewQueryBudgetStore(config)
	if err != nil {
		panic(err)
	}
	return store
}

func (s *QueryBudgetStore) ensure(lineage string) *queryBudgetState {
	state := s.byLineage[lineage]
	if state == nil {
		state = &queryBudgetState{current: s.config.Initial}
		s.byLineage[lineage] = state
	}
	return state
}

// Current returns the lineage's current budget, lazily initializing it at B0.
// The bool reports whether the lineage existed before this read.
func (s *QueryBudgetStore) Current(lineage string) (budget float64, existed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, existed = s.byLineage[lineage]
	return s.ensure(lineage).current, existed
}

// ApplyGrade consumes one continuous grade against a lineage; edgeID records
// which contextual edge produced the sample. Positive magnitude earns tokens
// additively; negative magnitude reduces multiplicatively; explicit zero holds.
// Invalid grades are rejected rather than clamped.
func (s *QueryBudgetStore) ApplyGrade(lineage, edgeID string, grade float64) (QueryBudgetEvent, error) {
	if lineage == "" {
		return QueryBudgetEvent{}, fmt.Errorf("query budget requires a lineage")
	}
	if math.IsNaN(grade) || math.IsInf(grade, 0) || grade < -1 || grade > 1 {
		return QueryBudgetEvent{}, fmt.Errorf("grade %v outside [-1,1]", grade)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.ensure(lineage)
	before := state.current
	switch {
	case grade > 0:
		state.current = math.Min(s.config.Max, before+s.config.Alpha*grade)
	case grade < 0:
		state.current = math.Max(s.config.Min, before*math.Pow(s.config.Beta, -grade))
	}
	return s.recordLocked(state, lineage, edgeID, before, state.current, grade, true), nil
}

// NoGrade records an unresolved episode for instrumentation and holds budget.
// It is not equivalent to ApplyGrade(lineage, edgeID, 0).
func (s *QueryBudgetStore) NoGrade(lineage, edgeID string) (QueryBudgetEvent, error) {
	if lineage == "" {
		return QueryBudgetEvent{}, fmt.Errorf("query budget requires a lineage")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.ensure(lineage)
	return s.recordLocked(state, lineage, edgeID, state.current, state.current, 0, false), nil
}

func (s *QueryBudgetStore) recordLocked(state *queryBudgetState, lineage, edgeID string, before, after, grade float64, hasGrade bool) QueryBudgetEvent {
	s.seq++
	event := QueryBudgetEvent{Lineage: lineage, EdgeID: edgeID, Before: before, After: after, Grade: grade, HasGrade: hasGrade, Sequence: s.seq}
	state.history = append(state.history, event)
	return event
}

func (s *QueryBudgetStore) History(lineage string) []QueryBudgetEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.byLineage[lineage]
	if state == nil {
		return nil
	}
	out := make([]QueryBudgetEvent, len(state.history))
	copy(out, state.history)
	return out
}

// All returns current budgets without creating missing lineages.
func (s *QueryBudgetStore) All() map[string]float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]float64, len(s.byLineage))
	for lineage, state := range s.byLineage {
		out[lineage] = state.current
	}
	return out
}
