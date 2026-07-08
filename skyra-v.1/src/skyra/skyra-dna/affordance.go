package dna

import "skyra-v1/src/skyra"

// The AffordanceRecord model (design-docs/being/affordance-record.md): a record
// IS a MemoryCell whose Kind is stamped from the affordance that produced it and
// whose Content is the affordance's own payload — one node type, an open Kind
// vocabulary (new Kinds are born as raw strings from the affordance, no constant
// added). A MemoryCell of Kind "perceive" carrying a PerceiveRecord payload IS
// the PerceiveRecord. Existing typed contents (InteractionRecord, ActionRecord,
// Prediction, Gist) collapse into this model as their record layers are next
// touched — the perceive path is born on it.

// PerceiveRecord is the payload of a Kind:"perceive" record — a sense event:
// entity + coord, no utterance. The coord rides ON the perception event (it is
// old the instant it is written; "current" = the freshest record per entity —
// perceive-internalization.md §2). The record sits between gist and entity:
// gist → record{coord} → entity.
type PerceiveRecord struct {
	EntityID  string
	Coord     skyra.Coord
	Timestamp int
}
