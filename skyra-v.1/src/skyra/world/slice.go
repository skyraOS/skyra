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
}

func NewSlice(id, beingID string) *Slice {
	return &Slice{
		id:      id,
		BeingID: beingID,
		records: make(map[string]skyra.ActivationRecord),
	}
}

func (s *Slice) ID() string                                  { return s.id }
func (s *Slice) Tree() *skyra.TraversalPair                  { return &s.tree }
func (s *Slice) Records() map[string]skyra.ActivationRecord  { return s.records }
func (s *Slice) Threshold() float64                          { return s.threshold }
func (s *Slice) Count() int                                  { return s.count }
func (s *Slice) SetCount(c int)                              { s.count = c }
