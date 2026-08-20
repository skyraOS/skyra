// Package ids holds the identity types and the one minting mechanism.
//
// interfaces.md leaves ID minting open, noting only that EngramID, ObservationID
// and SpanID are all globally unique, all minted by the kernel, and are "probably
// one mechanism." This is that one mechanism.
//
// Global uniqueness is load-bearing in a way that is easy to misread: Producer
// resolves across entity trees. The credit walk is bounded by CITES BEING EMPTY at
// a crossing, never by ID scope — so nobody should reach for scoped IDs to "fix" a
// boundary that is already held.
package ids

import (
	"fmt"
	"sync/atomic"
)

type (
	EngramID      string
	ObservationID string
	EntityID      string
	HostID        string
	SpanID        string
)

// Minter is held by the kernel. It is the only source of durable identity.
//
// Sequential rather than random on purpose: an ID that sorts by creation makes a
// transcript readable, and nothing in the design reads an ID for meaning.
type Minter struct {
	engram      atomic.Uint64
	observation atomic.Uint64
	span        atomic.Uint64
}

func (m *Minter) Engram() EngramID {
	return EngramID(fmt.Sprintf("eng-%04d", m.engram.Add(1)))
}

func (m *Minter) Observation() ObservationID {
	return ObservationID(fmt.Sprintf("obs-%04d", m.observation.Add(1)))
}

func (m *Minter) Span() SpanID {
	return SpanID(fmt.Sprintf("span-%04d", m.span.Add(1)))
}
