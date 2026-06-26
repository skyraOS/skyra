package dna

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/skyra"
)

// Global — the orchestrator (supersedes the old Coherence cell).
//
// Three jobs: vote tally, tie-break, and grouping. It holds two agents (self,
// other), each its own descent/ascent. Its Children() are GROUPINGS — higher-
// order setpoints over the active field — not the raw setpoints (those are
// below Global's altitude: the kernel filter + the raw predict/score layer).
//
// Visibility is scoped: Global reads only crystallized understandings of
// completed setpoints and the potential gists of active groupings.
//
// See specs/2026-06-16/skyra-dna/global.md and
// design-docs/cognition/groupings-and-understandings.md.
type GlobalCell struct {
	id    string
	slice skyra.Slice

	// sensory — back-reference to the boundary's sensory agent. Global reads its
	// warm surroundings (under sensory's slice lock) to assemble the frame.
	sensory *SensoryCell

	// groupings — Global's index over the active field; the cells the kernel
	// walks as Global's Children(). Each is a higher-order setpoint carrying a
	// gist (an understanding that hasn't crystallized yet). Keyed by edge
	// identity (the map key), so activation is edge-owned on the slice.
	groupings map[string]skyra.Space

	// kernel — the send primitive; used to spawn the next pipeline stage on
	// Express.
	kernel skyra.Time

	// other — the "other" agent: the next pipeline stage after Global, spawned on
	// Express (Global → other). self/voting come after.
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
	// once the warm case lands. See specs/2026-06-16/skyra-dna/global.md
	// ("Grouping vs. framing").
	gist string
}

func NewGlobalCell(ctx BuildContext, sensory *SensoryCell) *GlobalCell {
	g := &GlobalCell{
		id:        ctx.BeingID + ":global",
		slice:     ctx.sliceFor(":global"),
		sensory:   sensory,
		kernel:    ctx.Kernel,
		infer:     ctx.Infer,
		groupings: map[string]skyra.Space{},
	}
	g.other = NewOtherCell(ctx) // the next pipeline stage (Global → other)
	return g
}

func (g *GlobalCell) ID() string                            { return g.id }
func (g *GlobalCell) Slice() skyra.Slice                    { return g.slice }
func (g *GlobalCell) NewActivation() skyra.ActivationRecord { return &globalActivation{} }

// Children — the groupings (the kernel walks these). Past-only adjacency; the
// future lives here as the grouping index, applied top-down as context.
func (g *GlobalCell) Children() map[string]skyra.Space { return g.groupings }

// Content — the grouping index (durable future layer) + the transient vote
// synthesis. (The grouping index is g.groupings; synthesis is per-cycle.)
func (g *GlobalCell) Content() any { return g.synthesis }

// Observe routes to one of two internal functions based on the impulse. The
// discriminator is not generalized beyond Global yet.
//   - external impulse  -> observeExternal: the grouping pass (score impulse vs
//     active groupings in the buffer's context; skip if no completed setpoint).
//   - returning result  -> observeSelf: integrate the returned scores.
func (g *GlobalCell) Observe(origin skyra.Space, impulse any) {
	if isReturn(impulse) {
		g.observeSelf(impulse)
		return
	}
	g.observeExternal(impulse)
}

// observeExternal — the grouping pass. The read-half (built): assemble the frame
// from sensory's warm surroundings. The write-half (TODO): score by overlap ×
// recency, run the grouping inference call, write setpoint→groupID.
// See specs/2026-06-16/skyra-dna/global-observe-gather.md.
func (g *GlobalCell) observeExternal(impulse any) {
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
	}
}

// coldStartSystem instructs the grouping pass to predict a single framing for the
// exchange — the gist that becomes the cycle's top-down context.
const coldStartSystem = "You are an evolving being reading a live exchange. In one short " +
	"phrase, predict what this exchange is about and what you should be trying to do. " +
	"Respond with only the tag: <label>your framing</label>"

// renderFrame turns the timestamp-ordered frame into a readable transcript for
// the prompt — the interaction records (who said what), the content the framing
// is predicted from. Entity/state nodes are metadata and omitted.
func renderFrame(frame []ParsedNode) string {
	var b strings.Builder
	for _, e := range frame {
		if rec, ok := e.Content.(InteractionRecord); ok {
			fmt.Fprintf(&b, "[t%d] %s: %s\n", e.Timestamp, rec.From, rec.Utterance)
		}
	}
	return b.String()
}

// Gist is the current framing (the grouping label predicted this cycle).
func (g *GlobalCell) Gist() string { return g.gist }

// assembleFrame reads sensory's warm surroundings and parses them entity-outward
// — each entity plus its children (interaction/state records now; stamped
// setpoints once they exist) — into a chronological (timestamp-sorted) frame.
//
// Global already holds the references (via the sensory back-ref), so it calls
// each node's Parse() directly rather than routing through the kernel's Express
// payload. Read under SENSORY's slice lock (SWMR): waits for sensory's wave to
// finish, safe against the next one. Lock order is Global → sensory, never the
// reverse.
func (g *GlobalCell) assembleFrame() []ParsedNode {
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
func (g *GlobalCell) Frame() []ParsedNode { return g.frame }

// observeSelf — integrate returning agent scores toward the vote. [stub]
func (g *GlobalCell) observeSelf(impulse any) {
	// TODO: accumulate returned self/other scores for the Hopfield tally.
}

// Express forwards into the next pipeline stage (the other agent) and returns
// the held synthesis. origin=other (the agent owns its own wave); async (go) so
// the spawn locks other's own slice, not Global's. The vote/orchestration
// (per-grouping self/other, cascade > closure > open, Hopfield tally) is the
// write-half [stub] — it lands once the agents predict and score.
func (g *GlobalCell) Express(origin skyra.Space, impulse any) any {
	if g.kernel != nil && g.other != nil {
		go g.kernel.Realize(g.other, g.other, impulse)
	}
	return g.synthesis
}

// Other is Global's forward target (the "other" agent, next pipeline stage).
func (g *GlobalCell) Other() *OtherCell { return g.other }

// AddGrouping registers a grouping under its edge-identity key.
func (g *GlobalCell) AddGrouping(key string, grouping skyra.Space) {
	g.groupings[key] = grouping
}

// globalActivation — minimal activation record. The full bump equation
// (IndexScore × Alignment × stress_modulator into Identity/Familiarity EMAs)
// lives in design-docs/kernel/activation-and-traversal.md and is the next step.
type globalActivation struct{}

func (a *globalActivation) Activate(_ bool, _ int) float64 { return 1.0 }

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
