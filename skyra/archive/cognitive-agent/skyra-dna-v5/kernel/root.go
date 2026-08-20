package kernel

import (
	"fmt"
	"sync"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/host"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// RootOrigin records where an ask came from.
//
//	USER-ROOTED     the ask came from outside. a human at the root.
//	KERNEL-ROOTED   the kernel originated it. permissions · credit walks · eviction.
//
// Counted so that kernel-rooted work outgrowing user-rooted work — the
// self-maintenance spiral — is MEASURABLE RATHER THAN FELT.
type RootOrigin int

const (
	UserRooted RootOrigin = iota
	KernelRooted
)

// Root is a chain, and it is THE TRANSACTION BOUNDARY.
//
// A possibility is a frame inside a transaction, not a transaction of its own.
// Nested transactions raise "what does it mean for a child to commit into a
// parent that later aborts," and there is no good answer. The chain has one
// commit point, and it is where authority, acceptance and the human already are.
type Root struct {
	Span   ids.SpanID
	Engram *graph.Engram

	Origin   RootOrigin
	Provoker ids.SpanID // empty only for the maintenance class

	mu sync.Mutex

	// Held, not performed. Materialized along the citation chain when the root
	// accepts; anything the walk never reaches is DISCARDED, and the discard is
	// RECORDED.
	staged []*Effect
}

// Disposition is what became of a staged effect. EVERY STAGED EFFECT HAS
// EXACTLY ONE, and that is the whole reason it is an enum.
//
//	Staged      held, not performed. the initial state.
//	Landed      the driver wrote it.
//	Discarded   the credit walk never reached it, or the chain was rejected. RECORDED.
//	Failed      reached, attempted, the driver errored. RECORDED.
//
// Two booleans expressed four combinations, two of them meaningless (both true)
// and one of them — NEITHER — the exact state a failed write fell into
// unnoticed. One field cannot hold a state nothing can name.
type Disposition int

const (
	Staged Disposition = iota
	Landed
	Discarded
	Failed
)

func (d Disposition) String() string {
	switch d {
	case Staged:
		return "staged"
	case Landed:
		return "landed"
	case Discarded:
		return "discarded"
	case Failed:
		return "failed"
	}
	return "unknown"
}

// Effect is a world effect that has been ASKED FOR and not performed.
//
//	STORAGE          stage the EFFECT.   an overlay — B reads back what A wrote.
//	EVERYTHING ELSE  stage the INTENT.   the kernel holds "send this."
//
// The second is the general case; the overlay is a storage-class convenience
// layered on it. An email cannot be un-sent, but it can be un-sent-YET, which is
// why the earlier stageable/not-stageable split was the wrong cut.
type Effect struct {
	// WHO asked. Materialization follows citation, so an effect lands iff the
	// credit walk reaches the engram that staged it.
	Engram ids.EngramID

	Class  string // "storage"
	Path   string
	Body   []byte
	Handle string

	// What became of it. Staged until something disposes of it.
	Disposition Disposition
}

// Stage holds an effect rather than performing it, and hands back the handle a
// possibility will carry forward.
//
// For storage this is an OVERLAY: a later possibility in the same chain reads
// back what an earlier one wrote, without anything having touched the world.
func (r *Root) Stage(e *Effect) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.Handle = fmt.Sprintf("staged-%d", len(r.staged)+1)
	r.staged = append(r.staged, e)
	return e.Handle
}

// ReadStaged serves a read from the overlay, newest write wins. This is what
// makes "B reads back what A wrote" true without the write having happened.
func (r *Root) ReadStaged(path string) ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.staged) - 1; i >= 0; i-- {
		if r.staged[i].Class == "storage" && r.staged[i].Path == path {
			return r.staged[i].Body, true
		}
	}
	return nil, false
}

func (r *Root) Staged() []*Effect {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*Effect(nil), r.staged...)
}

// Materialize realizes the staged effects the credit walk reached, and DISCARDS
// the rest — recording every discard.
//
//	THE CREDIT WALK      which invocations contributed to what was accepted.
//	MATERIALIZATION      exactly the same set.
//
// One traversal, two jobs. "Credit follows data" and "effects follow data" stop
// being two claims.
//
// A DISCARD IS AN EVENT, NOT AN ABSENCE. This is the one place that must not
// simply inherit the credit mechanism: citation was designed with cheap failure —
// a mis-derived citation costs a mis-ranking and the system corrects — where the
// same mistake here costs a LOST WRITE, and the tolerances were never set for
// that. So every staged effect is explicitly disposed, and an unreachable one is
// a reportable fact rather than a no-op.
func (r *Root) Materialize(g *graph.Graph, accepted *graph.Observation, h host.Host) error {
	reached := map[ids.EngramID]bool{}
	if accepted != nil {
		reached = g.Reached(accepted)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for _, e := range r.staged {
		if !reached[e.Engram] {
			e.Disposition = Discarded
			g.RecordDiscard(graph.Discard{
				Span:   r.Span,
				Engram: e.Engram,
				Detail: fmt.Sprintf("%s %s (%d bytes) — the credit walk never reached its producer", e.Class, e.Path, len(e.Body)),
			})
			continue
		}
		switch e.Class {
		case "storage":
			s := h.Storage()
			if s == nil {
				return fmt.Errorf("host %s implements no storage class", h.ID())
			}
			if err := s.Write(e.Path, e.Body); err != nil {
				// REACHED, ATTEMPTED, REFUSED BY THE DRIVER. Disposed like
				// anything else: an effect that was asked for and did not
				// happen is a reportable fact, and the rule is that EVERY
				// staged effect ends in a state something can name. Left
				// undisposed it fell into the fourth state nothing records —
				// which is the whole of what a lost write looks like from
				// outside.
				e.Disposition = Failed
				g.RecordDiscard(graph.Discard{
					Span:   r.Span,
					Engram: e.Engram,
					Detail: fmt.Sprintf("%s %s (%d bytes) — the driver refused the write: %v",
						e.Class, e.Path, len(e.Body), err),
				})
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			e.Disposition = Landed
		default:
			if firstErr == nil {
				firstErr = fmt.Errorf("no driver for staged class %q", e.Class)
			}
		}
	}
	return firstErr
}

// Discard drops every staged effect — an orphaned root, or a chain nobody
// accepted. Each one is still recorded.
func (r *Root) Discard(g *graph.Graph, why string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range r.staged {
		if e.Disposition != Staged {
			continue
		}
		e.Disposition = Discarded
		g.RecordDiscard(graph.Discard{
			Span:   r.Span,
			Engram: e.Engram,
			Detail: fmt.Sprintf("%s %s — %s", e.Class, e.Path, why),
		})
	}
}
