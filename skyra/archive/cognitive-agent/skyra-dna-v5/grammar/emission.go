// Package grammar holds the two forms and the parser that admits nothing else.
//
//	Two forms — SELECT and COMMIT. Everything else is a name.
//
// Because this is a grammar rather than an interface, the incapacity is a PARSING
// property, not an API property. A possibility cannot do anything but emit these
// two forms, and nothing else is expressible.
//
// That is the one item of the traversal bill that was DROPPED rather than kept:
// an API split is checkable by a compiler, where a grammar is checked by a parser
// at runtime. What is bought is that the constraint covers everything a
// possibility could emit rather than the two methods someone remembered to split.
package grammar

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// Emission is what a possibility produces. Two implementations, and the type
// switch over them is a GUARD ON THE VOCABULARY: two forms is the design's
// central claim, and a switch that fails to compile on a third case is better
// than dispatch that silently accommodates one.
type Emission interface{ isEmission() }

// Select is an INVOCATION GROUP.
//
// Cardinality is the concurrency declaration: one member is a sequential step, N
// members run together, and EVERY GROUP AWAITS. There is no async, no await, no
// execution flag — arity is the whole declaration.
//
// Select IS invoke. There is no separate call form, which makes "selection
// happens at the call site" literal rather than descriptive.
type Select struct{ Members []Member }

// Member is one named invocation inside a group.
type Member struct {
	// The NAME, resolved in frame. Capability is name resolution: you can select
	// what you can name; you can name what you were given.
	Engram string

	// PER MEMBER, never per group. Grouping is a scheduling fact and must not
	// become a shared cause — a group-level cite would be ONE CAUSE FOR N
	// SELECTIONS, reintroducing sibling blindness at the one place the design has
	// it solved, and making only whole groups prunable.
	Cite ids.ObservationID

	// WHO interprets the child. Resolved in frame, exactly like names.
	Entity ids.EntityID

	// WHAT — bound as arguments.
	Observations []ids.ObservationID

	// OPTIONAL, and usually empty. Delivery picks a host on the way out, so
	// naming one is only meaningful where the host is the TARGET rather than the
	// interpreter — turning on THIS TV, not that one.
	Host ids.HostID
}

// Commit closes the possibility and is THE ONLY DURABLE WRITE.
//
// No Cites field: citation is DERIVED from the Refs appearing in Content.
type Commit struct {
	// Exactly one of these.
	//
	//	NEVER SELECTED   a LEAF. atomic, one-shot. ORIGINATES an observation.
	//	HAS SELECTED     an ORCHESTRATOR. may only NAME what came back. PERMANENTLY.
	//
	// Originate only where it could not be mistaken for composition. The forfeit
	// fires at the FIRST select, because that is when ambiguity becomes possible,
	// and it is permanent, because ambiguity never goes away once children have
	// returned.
	Content  content.Content   // a LEAF originates.
	Observed ids.ObservationID // an ORCHESTRATOR names.
}

func (Select) isEmission() {}
func (Commit) isEmission() {}

// IsOrigination reports whether this commit invents rather than names.
func (c Commit) IsOrigination() bool { return c.Observed == "" }
