package world

import (
	"sync"

	"skyra-v1/src/skyra"
)

type Slice struct {
	mu        sync.Mutex
	id        string
	BeingID   string
	count     int
	records   map[string]skyra.ActivationRecord
	threshold float64
	tree      skyra.TraversalPair
	result    any
	maps      map[string]any
	stress    float64 // global stress (neuromodulator); mechanics wired later
}

func NewSlice(id, beingID string) *Slice {
	return NewSliceWithThreshold(id, beingID, 0)
}

func NewSliceWithThreshold(id, beingID string, threshold float64) *Slice {
	return &Slice{
		id:        id,
		BeingID:   beingID,
		records:   make(map[string]skyra.ActivationRecord),
		threshold: threshold,
		maps:      make(map[string]any),
	}
}

func (s *Slice) ID() string                                 { return s.id }
func (s *Slice) Tree() *skyra.TraversalPair                 { return &s.tree }
func (s *Slice) Result() any                                { return s.result }
func (s *Slice) SetResult(result any)                       { s.result = result }
func (s *Slice) Maps() map[string]any                       { return s.maps }
func (s *Slice) Records() map[string]skyra.ActivationRecord { return s.records }
func (s *Slice) Threshold() float64                         { return s.threshold }
func (s *Slice) Count() int                                 { return s.count }
func (s *Slice) SetCount(c int)                             { s.count = c }
func (s *Slice) Stress() float64                            { return s.stress }
func (s *Slice) SetStress(v float64)                        { s.stress = v }
func (s *Slice) Lock()                                      { s.mu.Lock() }
func (s *Slice) Unlock()                                    { s.mu.Unlock() }
