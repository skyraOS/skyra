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
	// A reframe Route carries a failure summary — a dramatic grade miss asking for
	// a corrected framing. Thread the summary into the framing present.
	reason := ""
	if r := routeOf(impulse); r != nil && r.Kind == RouteReframe {
		reason = r.Summary
		debug.Being("skyra", "midlevel", "reframe requested")
	}
	g.observeExternal(reason)
}

// observeExternal — the grouping pass. The read-half (built): assemble the frame
// from sensory's warm surroundings. The write-half (TODO): score by overlap ×
// recency, run the grouping inference call, write setpoint→groupID.
// reframeReason (non-empty on a grade-driven reframe) puts the failure in the
// present so the new gist is predicted with the surprise in view.
// See specs/2026-06-16/skyra-dna/midlevel-observe-gather.md.
func (g *MidLevelCell) observeExternal(reframeReason string) {
	var trigger *MemoryCell
	g.frame, trigger = g.assembleFrame()

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
	system := coldStartSystem
	if reframeReason != "" {
		// Reframe: the last framing's predictions all missed. Put the failure in
		// the present and ask for a corrected framing.
		present = fmt.Sprintf("Your previous framing led to predictions that all failed:\n%s\n\n---\nThe exchange so far:\n%s",
			reframeReason, present)
		system = reframeSystem
	}
	debug.Being("skyra", "midlevel", fmt.Sprintf("════════════ frame t=%d ════════════\n%s\n════════════════════════════════════", g.slice.Count(), present))
	debug.Being("skyra", "midlevel", "inference start")
	resp, err := g.infer(system, present)
	if err != nil {
		debug.Being("skyra", "midlevel", fmt.Sprintf("inference error: %v", err))
		return // a failed call leaves last cycle's gist standing
	}
	debug.Being("skyra", "midlevel", "inference end")
	if label, ok := ParseLabel(resp); ok {
		g.gist = label
		g.startGrouping(label, trigger)
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
//
// Two sets of children are wired at creation time:
//
//  1. Gist tokens — each content word in the gist becomes a shared token node
//     child. These are the attention roots: the first nodes traversed when entering
//     this grouping, defining the current working memory frontier.
//
//  2. Triggering interaction — the interaction record that caused this grouping
//     (the latest inbound utterance). Its children are already tokenized utterance
//     tokens (some resolving to entity nodes). Adding it here connects utterance
//     tokens to gist tokens through the grouping: hello→interaction→grouping→greeting.
//     Traversed repeatedly, that path accumulates activation — the semantic
//     relationship between surface word and framing grows from topology, not
//     from any explicit declaration.
func (g *MidLevelCell) startGrouping(gist string, trigger *MemoryCell) {
	ts := 0
	if g.slice != nil {
		ts = g.slice.Count()
	}
	groupID := fmt.Sprintf("%s:g:%d", g.id, ts)
	idx := g.groupingIndex()
	if idx == nil {
		return
	}
	groupingCell := NewMemoryCell(groupID, KindGrouping, Gist{Text: gist, Timestamp: ts}, g.slice)
	// gist tokens — attention roots / working memory seeds; no pronoun map
	// (gist is LLM-generated third-person framing, speaker context absent).
	// tokenize MUTATES sensory's entity index (the shared token web), so it is a
	// cross-cell write — hold sensory's lock around it (substrate pass). (This
	// cross-cell write is a known single-writer wart; a cleaner factoring has
	// sensory own all its writes — deferred.)
	if g.sensory != nil {
		var toks []*MemoryCell
		if sl := g.sensory.Slice(); sl != nil {
			sl.Lock()
			toks = g.sensory.tokenize(gist, nil)
			sl.Unlock()
		} else {
			toks = g.sensory.tokenize(gist, nil)
		}
		var tokIDs []string
		for _, tok := range toks {
			groupingCell.AddChild(groupID+"→"+tok.ID(), tok)
			tokIDs = append(tokIDs, tok.ID())
		}
		debug.Being("skyra", "midlevel", fmt.Sprintf("grouping tokens=%v", tokIDs))
	}
	// triggering interaction — the utterance that caused this framing
	if trigger != nil {
		groupingCell.AddChild(groupID+"→trigger", trigger)
		if rec, ok := trigger.Content().(InteractionRecord); ok {
			debug.Being("skyra", "midlevel", fmt.Sprintf("grouping trigger utterance=%q tokens=%d", rec.Utterance, len(trigger.Children())))
		}
	}
	idx[groupID] = groupingCell
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

// reframeSystem runs when the prior framing's predictions all missed — the
// surprise is in the present. Ask for a corrected framing that accounts for it.
const reframeSystem = "You are an evolving being. Your previous framing of this exchange was " +
	"wrong — every prediction it produced failed. Given what actually happened, re-frame: in " +
	"one short phrase, say what this exchange is really about and what you should be trying to " +
	"do now. Respond with only the tag: <label>your corrected framing</label>"

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
func (g *MidLevelCell) assembleFrame() ([]ParsedNode, *MemoryCell) {
	if g.sensory == nil || g.sensory.Slice() == nil {
		return nil, nil
	}
	sl := g.sensory.Slice()
	sl.Lock()
	defer sl.Unlock()

	var frame []ParsedNode
	var latestInteraction *MemoryCell
	var latestTS int = -1
	for _, entity := range g.sensory.ByEntity {
		mc, ok := entity.(*MemoryCell)
		if !ok {
			continue
		}
		frame = append(frame, mc.Parse())
		for _, child := range mc.Children() { // entity-outward
			cc, ok := child.(*MemoryCell)
			if !ok {
				continue
			}
			frame = append(frame, cc.Parse())
			if cc.Kind() == KindResponse {
				if ts := contentTimestamp(cc.Content()); ts > latestTS {
					latestTS = ts
					latestInteraction = cc
				}
			}
		}
	}
	sort.SliceStable(frame, func(i, j int) bool {
		return frame[i].Timestamp < frame[j].Timestamp
	})
	return frame, latestInteraction
}

// Frame is the assembled read-half output (parsed, timestamp-ordered
// surroundings) for the current cycle.
func (g *MidLevelCell) Frame() []ParsedNode { return g.frame }

// WorldState returns a formatted snapshot of known positions for the prompt:
// skyra's own position (if known) and each entity's last seen coord.
// Reads directly from sensory — no lock needed (single writer, called from
// the same goroutine that runs the pipeline).
func (g *MidLevelCell) WorldState() string {
	if g.sensory == nil {
		return ""
	}
	// Cross-cell read of sensory's ByEntity + SelfCoord — hold sensory's lock
	// across the whole iteration (substrate pass; serializes against sensory's
	// Observe writes). SelfCoord is called only here and does not re-lock.
	if sl := g.sensory.Slice(); sl != nil {
		sl.Lock()
		defer sl.Unlock()
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Below is your world state — the locations of each entity within the world you inhabit.\n")
	if coord, ok := g.sensory.SelfCoord(); ok {
		fmt.Fprintf(&b, "Self: %d,%d\n", coord.X, coord.Y)
	} else {
		fmt.Fprintf(&b, "Self: unknown — fire perceive to orient\n")
	}
	for id, entity := range g.sensory.ByEntity {
		mc, ok := entity.(*MemoryCell)
		if !ok {
			continue
		}
		var latestTS int = -1
		var latestCoord skyra.Coord
		for _, child := range mc.Children() {
			cc, ok := child.(*MemoryCell)
			if !ok || cc.Kind() != KindPresence {
				continue
			}
			if es, ok := cc.Content().(EntityState); ok && es.Timestamp > latestTS {
				latestTS = es.Timestamp
				latestCoord = es.Location
			}
		}
		// only show entities with real position data — skip thin token-only nodes
		if latestTS >= 0 {
			fmt.Fprintf(&b, "  %s: %d,%d\n", id, latestCoord.X, latestCoord.Y)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

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

// Sensory exposes the boundary's sensory cell so motor can record outgoing utterances.
func (g *MidLevelCell) Sensory() *SensoryCell { return g.sensory }

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
