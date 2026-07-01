package dna

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// MidLevel — the orchestrator (supersedes the old Coherence cell).
//
// Three jobs: vote tally, tie-break, and grouping. It holds two agents (self,
// other), each its own descent/ascent. Its Children() are GROUPINGS — higher-
// order setpoints over the active field — not the raw setpoints (those are
// below MidLevel's altitude: the kernel filter + the raw predict/score layer).
//
// Visibility is scoped: MidLevel reads only crystallized understandings of
// completed setpoints and the potential gists of active groupings.
//
// See specs/2026-06-16/skyra-dna/midlevel.md and
// design-docs/cognition/groupings-and-understandings.md.
type MidLevelCell struct {
	id    string
	slice skyra.Slice

	// sensory — back-reference to the boundary's sensory agent. MidLevel reads its
	// warm surroundings (under sensory's slice lock) to assemble the frame.
	sensory *SensoryCell

	// kernel — the send primitive; used to spawn the next pipeline stage on
	// Express.
	kernel skyra.Time

	// other — the "other" agent: the next pipeline stage after MidLevel, spawned on
	// Express (MidLevel → other). self/voting come after.
	other *OtherCell

	// synthesis — the transient held vote state for the current cycle.
	synthesis any

	// frame — the assembled read-half output this cycle: the parsed,
	// timestamp-ordered surroundings (the grouping inference call's input).
	frame []ParsedNode

	// infer — the injected LLM seam (nil → skip inference).
	infer InferFunc

	// gist — the current framing (the grouping label predicted this cycle). The
	// cold-start form: a plain string. Becomes a framing node on a stable grouping
	// once the warm case lands. See specs/2026-06-16/skyra-dna/midlevel.md
	// ("Grouping vs. framing").
	gist string
}

func NewMidLevelCell(ctx BuildContext, sensory *SensoryCell) *MidLevelCell {
	g := &MidLevelCell{
		id:      ctx.BeingID + ":midlevel",
		slice:   ctx.sliceFor(":midlevel"),
		sensory: sensory,
		kernel:  ctx.Kernel,
		infer:   ctx.Infer,
	}
	g.other = NewOtherCell(ctx, g) // the next pipeline stage (MidLevel → other); g = the back-ref
	return g
}

func (g *MidLevelCell) ID() string                            { return g.id }
func (g *MidLevelCell) Slice() skyra.Slice                    { return g.slice }
func (g *MidLevelCell) NewActivation() skyra.ActivationRecord { return &midLevelActivation{} }

// Children — the groupings (the kernel walks these). Past-only adjacency; the
// future lives here as the grouping index, applied top-down as context.
func (g *MidLevelCell) Children() map[string]skyra.Space { return g.groupingIndex() }

func (g *MidLevelCell) Agents() map[string]skyra.Space {
	if g.other == nil {
		return nil
	}
	return map[string]skyra.Space{g.other.ID(): g.other}
}

// groupingsKey is the slice-map key under which the grouping index lives.
const groupingsKey = "groupings"

// groupingIndex is MidLevel's grouping map, resident on its SLICE (Maps), not a
// private field — so the agents (self/other) can read it cross-slice and append
// their predictions to the grouping cells under the lock. Lazily created; the map
// key is the groupID, each value a higher-order setpoint carrying a gist.
func (g *MidLevelCell) groupingIndex() map[string]skyra.Space {
	if g.slice == nil {
		return nil
	}
	idx, ok := g.slice.Maps()[groupingsKey].(map[string]skyra.Space)
	if !ok {
		idx = map[string]skyra.Space{}
		g.slice.Maps()[groupingsKey] = idx
	}
	return idx
}

// Content — the transient vote synthesis (the grouping index lives on the slice;
// see groupingIndex). synthesis is per-cycle.
func (g *MidLevelCell) Content() any { return g.synthesis }

// Observe routes to one of two internal functions based on the impulse. The
// discriminator is not generalized beyond MidLevel yet.
//   - external impulse  -> observeExternal: the grouping pass (score impulse vs
//     active groupings in the buffer's context; skip if no completed setpoint).
//   - returning result  -> observeSelf: integrate the returned scores.
func (g *MidLevelCell) Observe(origin skyra.Space, impulse any) {
	if isReturn(impulse) {
		g.observeSelf(impulse)
		return
	}
	g.observeExternal(impulse)
}

// observeExternal — the grouping pass. The read-half (built): assemble the frame
// from sensory's warm surroundings. The write-half (TODO): score by overlap ×
// recency, run the grouping inference call, write setpoint→groupID.
// See specs/2026-06-16/skyra-dna/midlevel-observe-gather.md.
func (g *MidLevelCell) observeExternal(impulse any) {
	g.frame = g.assembleFrame()

	// Cold-start grouping pass: predict the framing/gist for this exchange. No
	// setpoints to group yet, no existing groupings to match — just impulse → gist.
	// (Warm case — match/re-frame/create over existing groupings + setpoint
	// grouping — is the next layer.) TODO (write-half after this): score by
	// overlap × recency, group setpoints under the gist, write setpoint→groupID.
	if g.infer == nil {
		return
	}
	present := renderFrame(g.frame)
	if present == "" {
		return // nothing perceived yet — nothing to frame
	}
	resp, err := g.infer(coldStartSystem, present)
	if err != nil {
		return // a failed call leaves last cycle's gist standing
	}
	if label, ok := ParseLabel(resp); ok {
		g.gist = label
		g.startGrouping(label)
		debug.Being("skyra", "midlevel", "grouping gist="+label)
	}
}

// Gist is a grouping's framing: the label + when it was set (for "most recent
// gist" recency). The swappable framing on the stable grouping anchor; cold-start
// holds it directly on the grouping, the separate-framing-node split is the warm
// case. See specs/2026-06-16/skyra-dna/midlevel.md ("Grouping vs. framing").
type Gist struct {
	Text      string
	Timestamp int
}

// startGrouping creates a fresh grouping anchor carrying the gist and publishes it
// into the slice-resident grouping index (groupingIndex), for the agents to read,
// fire a wave on, and append predictions to. Cold-start = the "create" action;
// warm = match/re-frame an existing grouping instead (later).
func (g *MidLevelCell) startGrouping(gist string) {
	ts := 0
	if g.slice != nil {
		ts = g.slice.Count()
	}
	groupID := fmt.Sprintf("%s:g:%d", g.id, ts)
	if idx := g.groupingIndex(); idx != nil {
		idx[groupID] = NewMemoryCell(groupID, KindGrouping, Gist{Text: gist, Timestamp: ts}, g.slice)
	}
}

// Groupings publishes the grouping index by reference — the slice-resident map
// (the same one Children() walks). The agents read it (cross-slice, under
// MidLevel's lock), filter by most recent gist, fire a wave on one, and append
// their predictions to that grouping cell.
func (g *MidLevelCell) Groupings() map[string]skyra.Space { return g.groupingIndex() }

// coldStartSystem instructs the grouping pass to predict a single framing for the
// exchange — the gist that becomes the cycle's top-down context.
const coldStartSystem = "You are an evolving being reading a live exchange. In one short " +
	"phrase, predict what this exchange is about and what you should be trying to do. " +
	"Respond with only the tag: <label>your framing</label>"

// renderFrame turns the timestamp-ordered frame into a readable transcript for
// the prompt — the incoming utterances (who said what), the content the framing
// is predicted from. Filters by Kind (KindResponse = the exchange), so presence/
// state and entity nodes are skipped.
//
// CLEANUP TARGET — the sectioned view (2026-06-26 framing). This flat transcript
// is the degenerate case (only the Exchange section populated). The full view is
// MidLevel's "conscious field": group by Kind into labeled sections —
//
//	## Surroundings        entities + presence (who's here, where)
//	## Exchange            the utterances (the transcript)
//	## Active groupings    each grouping's gist + its member setpoints
//	## Standalone setpoints ungrouped, by gist
//
// When we build it:
//   - Recency-ordered, but direction differs by section: Exchange ascending
//     (it's a transcript, reading order is the meaning); knowledge sections
//     descending (most recent/relevant on top). "recency" = timestamp for now,
//     the Familiarity EMA later (it folds in relevance too).
//   - CAPPED per section (top-K) — a curated readout, never a dump. The gather's
//     overlap × recency already selects which setpoints are relevant; this just
//     renders the selection. Respects the ~100k recall ceiling.
//   - Cognitive Kinds render at GIST altitude — a setpoint shows as its
//     understanding/label, not its raw internals (MidLevel's scoped visibility).
//   - Within the setpoint section, organize by grouping (the warm case).
func renderFrame(frame []ParsedNode) string {
	var b strings.Builder
	for _, e := range frame {
		if e.Kind != KindResponse {
			continue
		}
		if rec, ok := e.Content.(InteractionRecord); ok {
			fmt.Fprintf(&b, "[t%d] %s: %s\n", e.Timestamp, rec.From, rec.Utterance)
		}
	}
	return b.String()
}

// Gist is the current framing (the grouping label predicted this cycle).
func (g *MidLevelCell) Gist() string { return g.gist }

// assembleFrame reads sensory's warm surroundings and parses them entity-outward
// — each entity plus its children (interaction/state records now; stamped
// setpoints once they exist) — into a chronological (timestamp-sorted) frame.
//
// MidLevel already holds the references (via the sensory back-ref), so it calls
// each node's Parse() directly rather than routing through the kernel's Express
// payload. Read under SENSORY's slice lock (SWMR): waits for sensory's wave to
// finish, safe against the next one. Lock order is MidLevel → sensory, never the
// reverse.
func (g *MidLevelCell) assembleFrame() []ParsedNode {
	if g.sensory == nil || g.sensory.Slice() == nil {
		return nil
	}
	sl := g.sensory.Slice()
	sl.Lock()
	defer sl.Unlock()

	var frame []ParsedNode
	for _, entity := range g.sensory.ByEntity {
		mc, ok := entity.(*MemoryCell)
		if !ok {
			continue
		}
		frame = append(frame, mc.Parse())
		for _, child := range mc.Children() { // entity-outward
			if cc, ok := child.(*MemoryCell); ok {
				frame = append(frame, cc.Parse())
			}
		}
	}
	sort.SliceStable(frame, func(i, j int) bool {
		return frame[i].Timestamp < frame[j].Timestamp
	})
	return frame
}

// Frame is the assembled read-half output (parsed, timestamp-ordered
// surroundings) for the current cycle.
func (g *MidLevelCell) Frame() []ParsedNode { return g.frame }

// observeSelf — integrate returning agent scores toward the vote. [stub]
func (g *MidLevelCell) observeSelf(impulse any) {
	// TODO: accumulate returned self/other scores for the Hopfield tally.
}

// Express forwards into the next pipeline stage (the other agent) and returns
// the held synthesis. origin=other (the agent owns its own wave); async (go) so
// the spawn locks other's own slice, not MidLevel's. The vote/orchestration
// (per-grouping self/other, cascade > closure > open, Hopfield tally) is the
// write-half [stub] — it lands once the agents predict and score.
func (g *MidLevelCell) Express(origin skyra.Space, impulse any) any {
	if g.kernel != nil && g.other != nil {
		debug.Being("skyra", "midlevel", "express → other")
		go g.kernel.Realize(g.other, g.other, impulse)
	}
	return g.synthesis
}

// Other is MidLevel's forward target (the "other" agent, next pipeline stage).
func (g *MidLevelCell) Other() *OtherCell { return g.other }

// AddGrouping registers a grouping under its edge-identity key.
func (g *MidLevelCell) AddGrouping(key string, grouping skyra.Space) {
	if idx := g.groupingIndex(); idx != nil {
		idx[key] = grouping
	}
}

// midLevelActivation — minimal activation record. The full bump equation
// (IndexScore × Alignment × stress_modulator into Identity/Familiarity EMAs)
// lives in design-docs/kernel/activation-and-traversal.md and is the next step.
type midLevelActivation struct{}

func (a *midLevelActivation) Activate(_ bool, _ int) float64 { return 1.0 }

// isReturn reports whether an impulse is a return-perception wave.
func isReturn(impulse any) bool {
	if imp, ok := impulse.(skyra.Impulse); ok {
		return imp.Return
	}
	if imp, ok := impulse.(*skyra.Impulse); ok && imp != nil {
		return imp.Return
	}
	return false
}
