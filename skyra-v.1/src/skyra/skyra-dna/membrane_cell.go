package dna

import "skyra-v1/src/skyra"

// Membrane — the being's sensorimotor boundary.
//
// Observe = the attention gate (pass-through stub for MVP; ownership: alignment vs. working memory → pass or discard).
// Express = the motor (stub for MVP; ownership: coincidence + fire, conflict → Coherence).
// Content = the root drive (durable; contrast with Coherence's transient synthesis).
type MembraneCell struct {
	id        string
	rootDrive string
	brain     map[string]skyra.Space
	slice     skyra.Slice
}

func NewMembraneCell(id, rootDrive string, slice skyra.Slice) *MembraneCell {
	return &MembraneCell{
		id:        id,
		rootDrive: rootDrive,
		brain:     map[string]skyra.Space{},
		slice:     slice,
	}
}

func (m *MembraneCell) ID() string                            { return m.id }
func (m *MembraneCell) Slice() skyra.Slice                    { return m.slice }
func (m *MembraneCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (m *MembraneCell) Content() any                          { return m.rootDrive }
func (m *MembraneCell) Children() map[string]skyra.Space      { return m.brain }

// Observe — sensory intake. On an external (non-return) wave that names a
// sender, the membrane records who/where/what into the surroundings map it owns.
// The sender self-identifies on the impulse (Source.SpaceID), the location is
// where they sent from (Source.Coord), the payload is what was said. This is the
// start of the pipeline: perception lands here as the surroundings.
func (m *MembraneCell) Observe(origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	if !ok || imp.Return || imp.Source == nil || imp.Source.SpaceID == "" {
		return // not an external signal that names a sender → nothing to record
	}
	if imp.Source.SpaceID == m.id {
		return // a self-originated wave isn't a surrounding
	}
	entity := resolveEntity(imp.Source.SpaceID) // who (exact-match for now)
	said, _ := imp.Payload.(string)             // what was said (best-effort)
	sur := m.surroundings()

	// WM recency: everyone absent this cycle decays; the perceived sender bumps.
	for id, e := range sur {
		if id != entity {
			e.act.Activate(false, 0)
		}
	}
	e, exists := sur[entity]
	if !exists {
		e = SurroundingEntry{act: newMemoryActivation()}
	}
	e.act.Activate(true, 0)
	e.Location = imp.Source.Coord
	e.LastSaid = said
	sur[entity] = e
}

func (m *MembraneCell) Express(origin skyra.Space, impulse any) any { return nil }

func (m *MembraneCell) AddChild(cell skyra.Space) {
	m.brain[cell.ID()] = cell
}

// surroundings returns the membrane-owned surroundings map on the slice, creating
// it on first use. It lives on the slice (one being-plane) so other cells can
// read it; the membrane is the only writer.
func (m *MembraneCell) surroundings() Surroundings {
	maps := m.slice.Maps()
	s, ok := maps[surroundingsKey].(Surroundings)
	if !ok {
		s = Surroundings{}
		maps[surroundingsKey] = s
	}
	return s
}

type memActivation struct{}

func (a *memActivation) Activate(_ bool, _ int) float64 { return 1.0 }

// --- the surroundings map (sensory working memory; membrane-owned) ---

const surroundingsKey = "surroundings"

// SurroundingEntry is the last-known state of an entity: where it was, what it
// last said, and its WM recency activation (bumped on perceive, decayed when
// absent). The membrane writes it on Observe; cells read it for IndexScore.
type SurroundingEntry struct {
	Location skyra.Coord
	LastSaid string
	act      *memoryActivation // recency/presence (membrane-managed, the WM pool)
}

// Activation is the entity's current surroundings recency (0 if none).
func (e SurroundingEntry) Activation() float64 {
	if e.act == nil {
		return 0
	}
	return e.act.Value()
}

// Surroundings is the entity → last-known-state index (the sensory WM). A map,
// so storing it on the slice and mutating it in place are the same reference.
type Surroundings map[string]SurroundingEntry

// SurroundingsOf reads the surroundings map off a slice (nil if none yet).
func SurroundingsOf(sl skyra.Slice) Surroundings {
	if sl == nil {
		return nil
	}
	s, _ := sl.Maps()[surroundingsKey].(Surroundings)
	return s
}

// resolveEntity maps a perceived descriptor to a stable entity id. Exact-match
// for now (the descriptor IS the id); upgrade path is a registry / fuzzy
// resolution. Entity resolution is the real work; this is the seam for it.
func resolveEntity(descriptor string) string { return descriptor }

// asImpulse normalizes the kernel's `any` impulse to a value Impulse.
func asImpulse(impulse any) (skyra.Impulse, bool) {
	switch v := impulse.(type) {
	case skyra.Impulse:
		return v, true
	case *skyra.Impulse:
		if v != nil {
			return *v, true
		}
	}
	return skyra.Impulse{}, false
}
