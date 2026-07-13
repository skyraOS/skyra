package dnav2

import (
	"fmt"
	"sort"
	"sync"

	"skyra-v1/src/skyra"
)

// MemoryCell is the passive, universal node of the durable knowledge graph.
// Kind describes a storage/debugging role; topology and typed ports carry
// behavior and relationship meaning.
type MemoryCell struct {
	id       string
	kind     Kind
	mu       sync.RWMutex
	content  any
	children map[string]skyra.Space
	ports    map[string]map[string]skyra.Space
	slice    skyra.Slice
}

type Kind string

const (
	KindEntity     Kind = "entity"
	KindResponse   Kind = "response"
	KindAction     Kind = "action"
	KindEpisode    Kind = "episode"
	KindFrame      Kind = "frame"
	KindPrediction Kind = "prediction"
)

type Episode struct{ Created int }

func NewMemoryCell(id string, kind Kind, content any, slice skyra.Slice) *MemoryCell {
	return &MemoryCell{id: id, kind: kind, content: content, children: map[string]skyra.Space{}, ports: map[string]map[string]skyra.Space{}, slice: slice}
}

func (m *MemoryCell) ID() string         { return m.id }
func (m *MemoryCell) Kind() Kind         { return m.kind }
func (m *MemoryCell) Slice() skyra.Slice { return m.slice }
func (m *MemoryCell) Content() any       { m.mu.RLock(); defer m.mu.RUnlock(); return m.content }
func (m *MemoryCell) Children() map[string]skyra.Space {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]skyra.Space, len(m.children))
	for k, v := range m.children {
		out[k] = v
	}
	return out
}
func (m *MemoryCell) ChildrenIn(port string) map[string]skyra.Space {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]skyra.Space, len(m.ports[port]))
	for k, v := range m.ports[port] {
		out[k] = v
	}
	return out
}
func (m *MemoryCell) ChildPorts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.ports))
	for port := range m.ports {
		out = append(out, port)
	}
	sort.Strings(out)
	return out
}
func (m *MemoryCell) Agents() map[string]skyra.Space        { return nil }
func (m *MemoryCell) NewActivation() skyra.ActivationRecord { return &rootActivation{} }
func (m *MemoryCell) Observe(skyra.Space, any)              {}
func (m *MemoryCell) Express(skyra.Space, any) any          { return m.Content() }
func (m *MemoryCell) AddChild(key string, cell skyra.Space) { m.AddChildTo("default", key, cell) }
func (m *MemoryCell) AddChildTo(port, key string, cell skyra.Space) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ports[port] == nil {
		m.ports[port] = map[string]skyra.Space{}
	}
	m.ports[port][key] = cell
	m.children[key] = cell
}
func (m *MemoryCell) SetContent(content any) { m.mu.Lock(); defer m.mu.Unlock(); m.content = content }

type Provenance string

const (
	ProvenanceObserved Provenance = "observed"
	ProvenanceAsserted Provenance = "asserted"
	ProvenanceComposed Provenance = "composed"
)

type Entity struct {
	ID         string
	Provenance Provenance
}
type InteractionRecord struct {
	From, To, Utterance, Action string
	Payload                     any
	Timestamp                   int
}
type PerceiveRecord struct {
	EntityID  string
	Coord     skyra.Coord
	Timestamp int
}
type Frame struct {
	Text      string
	Timestamp int
}
type Prediction struct {
	Text, Role  string
	Probability float64
	Timestamp   int
	Graded      bool
	Grade       float64
	Evidence    string
}
type ActionRecord struct {
	ProposalID, Primitive, EntityID, ActionID string
	Content                                   any
	Impulse                                   skyra.Impulse
	Timestamp                                 int
}

func (a ActionRecord) String() string {
	switch a.Primitive {
	case "engage":
		return fmt.Sprintf("<engage><%s/></engage>", a.EntityID)
	case "move":
		return fmt.Sprintf("<move><%s></%s></move>", a.EntityID, a.EntityID)
	case "perceive":
		return "<perceive/>"
	case "wait":
		return fmt.Sprintf("<wait><%s/></wait>", a.EntityID)
	default:
		return fmt.Sprint(a.Content)
	}
}
