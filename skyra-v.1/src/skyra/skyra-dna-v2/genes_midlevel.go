package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// midlevel:frame — assemble the exchange from the entity index, infer the gist
// (real inference), create the grouping. Ported from v1 midlevel.observeExternal
// (cold-start create path). Sets root.Grouping so the membrane threads it to the
// fan-out's predict workers. Reports "framed" / "empty".
func midlevelFrame(root *EphemeralRoot, origin skyra.Space, impulse any) {
	gi := root.DNA().Groupings()
	exchange := renderExchange(root.Entities)
	// the current surroundings — the active episode's traversed perceive records,
	// normalized (freshest per entity). Lets midlevel frame purely SPATIAL episodes
	// (the maze; sharpening the default grouping from sense data alone) — decision
	// 4, perceive-build-uncertainties.md. Read-only assembly: no crossing here (the
	// settle crosses what it grades; self's assembly crosses what it feeds).
	locations := ""
	if g := gi.Active(); g != nil {
		locations = strings.Join(
			renderLocations(FrameOf(g, settleFrameBudget, root.DNA().Activation(), root.DNA().Clock())), "\n")
	}
	if exchange == "" && locations == "" {
		root.Condition = "empty"
		return
	}
	infer := root.DNA().Infer()
	if infer == nil {
		root.Condition = "empty"
		return
	}
	present := framePresent(exchange, locations, gi.OpenEpisodes())
	resp, err := infer(frameOpSystem, present)
	if err != nil {
		root.Condition = "empty" // a failed call frames nothing this cycle
		return
	}
	op, id, gist, ok := ParseFrameOp(resp)
	if !ok {
		root.Condition = "empty"
		return
	}

	// The querier contract (Phase 1): the gist is an ANNOTATED sentence; the
	// compound is parsed from it (the parser is the validator). Malformed
	// annotation gets bounded retry-with-correction — a format error, not an
	// epistemic error (no AIMD tax; the budget controller lands in Phase 3).
	parsed, perr := ParseAnnotated(gist)
	for attempt := 1; perr != nil && attempt <= parseMaxRetries; attempt++ {
		resp, err = infer(frameOpSystem, present+"\n\n---\n"+CorrectionMessage(gist, perr))
		if err != nil {
			break
		}
		op2, id2, gist2, ok2 := ParseFrameOp(resp)
		if !ok2 {
			continue
		}
		op, id, gist = op2, id2, gist2
		parsed, perr = ParseAnnotated(gist)
	}
	ts := root.DNA().Tick()

	// Sentence + subgraph commit transactionally. A format failure is returned to
	// the LLM at the same epistemic budget; after bounded retries, commit neither
	// the legacy gist nor the Aitia set. The previous active state remains valid.
	if perr != nil {
		debug.Being("skyra", "querier", fmt.Sprintf("annotation failed after retries (%v) — retaining active state", perr))
		root.Condition = "query-format-error"
		return
	}
	comp := root.DNA().Compounds().Persist(parsed, ts)
	gist = parsed.Sentence

	// EDIT: mint a NEW gist on the existing episode (the sequence — the prior gist
	// stays and decays as "before I thought x"; the episode continues, so its edges
	// persist and re-grade — the compounding fix).
	if op == "edit" {
		if g := gi.Edit(id, gist, ts); g != nil {
			root.Grouping = g
			debug.Being("skyra", "midlevel", fmt.Sprintf("edit %s gist=%q", id, gist))
		}
	}
	// NEW (or an edit against an unknown id): open a fresh episode.
	if root.Grouping == nil {
		groupID := fmt.Sprintf("%s:g:%d", root.DNA().BeingID(), ts)
		root.Grouping = gi.Create(groupID, gist, ts) // Create sets it active
		debug.Being("skyra", "midlevel", fmt.Sprintf("new %s gist=%q", groupID, gist))
	}

	// Commit the minimal-Aitia side: mint the set around the compound, activate
	// it on the Slice (atomic replacement — retain-or-requery's requery half),
	// and cross-link the spine gist to its canonical compound.
	set := root.DNA().Compounds().MintSet(comp.Summary.ID(), ts)
	root.DNA().ActivateSet(set.ID)
	if g := latestGist(root.Grouping); g != nil {
		g.SetContent(dna.Gist{Text: gist, Timestamp: ts, CompoundID: comp.Summary.ID()})
	}
	debug.Being("skyra", "querier", fmt.Sprintf("compound %s: %q V=%d E=%d → set %s active",
		comp.Summary.ID(), gist, len(comp.Members), len(parsed.Edges), set.ID))

	consolidateInteraction(root) // hang the exchange on the episode + entity→episode

	// The coherence-settle: with the gist now FIXED, grade the episode's nodes against it and
	// write alignment (A) onto their edges, pruning misfits until it plateaus. midlevel
	// withholds its frame until here — the settle is the sole writer of A (active-grouping.md).
	passes := coherenceSettle(root.Grouping, root.DNA().Activation(), ts, infer)
	debug.Being("skyra", "midlevel", fmt.Sprintf("settled %s in %d passes", root.Grouping.ID(), passes))

	root.Condition = "framed"
}

// consolidateInteraction attaches the just-arrived interaction to the framed
// episode's CURRENT gist (records belong to the framing they arrived under) and
// wires the entity → episode edge — so the exchange lives on the spine the frame
// walks (into the frame + economy), not only in the entity index.
func consolidateInteraction(root *EphemeralRoot) {
	if root.Grouping == nil {
		return
	}
	gist := latestGist(root.Grouping)
	if gist == nil {
		return
	}
	node, entity, ok := root.Entities.Latest()
	if !ok {
		return
	}
	root.DNA().Groupings().AddChild(gist, gist.ID()+"→"+node.ID(), node)
	root.Entities.LinkGrouping(entity, root.Grouping)
}

// framePresent renders the framing-op present: the exchange + the current
// surroundings (normalized perceived locations, when present) + the open episodes
// the op decides edit-or-new against.
func framePresent(exchange, locations string, episodes []struct{ ID, Gist string }) string {
	var b strings.Builder
	if locations != "" {
		b.WriteString("Surroundings:\n")
		b.WriteString(locations)
		b.WriteString("\n\n")
	}
	b.WriteString("Exchange:\n")
	if exchange == "" {
		b.WriteString("(none — a sense-only moment)\n")
	} else {
		b.WriteString(exchange)
	}
	b.WriteString("\n\nOpen episodes:\n")
	if len(episodes) == 0 {
		b.WriteString("(none yet — this is the first exchange → open a <new> one)")
	} else {
		for _, e := range episodes {
			b.WriteString(fmt.Sprintf("[id=%s] %s\n", e.ID, e.Gist))
		}
	}
	return b.String()
}

// renderExchange turns the entity index's interaction records into a timestamp-
// ordered transcript (who said what) — the present the framing pass reads.
func renderExchange(ei *EntityIndex) string {
	if ei == nil {
		return ""
	}
	type line struct {
		ts int
		s  string
	}
	var lines []line
	for _, e := range ei.Snapshot() {
		for _, c := range e.Children() {
			mc, ok := c.(*dna.MemoryCell)
			if !ok {
				continue
			}
			if rec, ok := mc.Content().(dna.InteractionRecord); ok {
				lines = append(lines, line{rec.Timestamp, fmt.Sprintf("[t%d] %s: %s", rec.Timestamp, rec.From, rec.Utterance)})
			}
		}
	}
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].ts < lines[j].ts })
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.s)
		b.WriteByte('\n')
	}
	return b.String()
}

// coldStartSystem — the framing prompt (ported verbatim from v1).
const coldStartSystem = "You are an evolving being reading a live exchange. In one short " +
	"phrase, predict what this exchange is about and what you should be trying to do. " +
	"Respond with only the tag: <label>your framing</label>"

// frameOpSystem — the edit/new framing operator (active-grouping.md). Decides whether the
// exchange CONTINUES an open episode (edit + sharpened gist) or STARTS a new one, leaning
// toward continue (islands are the failure to avoid). The gist is truth.
const frameOpSystem = "You are an evolving being organizing your own memory. You are given " +
	"the recent EXCHANGE (what was just said), your current SURROUNDINGS (perceived locations, " +
	"when present), and a list of OPEN EPISODES you are already holding — each a short gist of " +
	"an ongoing thread, with an id.\n\n" +
	"Decide whether the exchange CONTINUES one of the open episodes or STARTS a new one.\n\n" +
	"- If it continues an episode, reply:  <edit id=\"THE-ID\">annotated sentence</edit>\n" +
	"    (restate that episode's gist to account for what just happened)\n" +
	"- If it is a new thread, reply:        <new>annotated sentence</new>\n\n" +
	"PREFER continuing an episode when the exchange plausibly extends it; open a new one only " +
	"when the topic GENUINELY shifts.\n\n" +
	"The gist is ONE COMPLETE ANNOTATED SENTENCE — your claim about what this moment is and " +
	"what you should be doing, with its structure marked inline:\n" +
	"- mark each referent with an ordinal:  michael[1]  or  {the skyra project}[2]\n" +
	"- mark each relationship on its verb/word:  asked[1→2]\n" +
	"- reuse an ordinal to re-reference the same thing:  michael[1] ... he[1]\n" +
	"Every referent gets a node annotation; every asserted relationship gets an edge " +
	"annotation; plain connecting words stay unannotated.\n\n" +
	"Example: <new>michael[1] wants[1→2] help[2] with[2→3] {the skyra project}[3], and " +
	"skyra[4] should assist[4→3]</new>\n\n" +
	"Reply with only the one tag."

// coherenceSystem — the settle's batch coherence grader (active-grouping.md), adapted from
// the comparator (validated to ~100k). Grades each item's coherence to the FRAMING (truth);
// items below 0 get pruned. Structure + calibration from the comparator; verb shifted from
// attainment to belonging.
const coherenceSystem = "You are the coherence check of a cognitive runtime.\n\n" +
	"You are given a FRAMING — a short statement of what this moment of thought is about, taken " +
	"as TRUE — and numbered ITEMS currently in the working set.\n\n" +
	"For each item, fill an <alignment>: how much the item BELONGS under the framing — how well " +
	"the framing supports, explains, or is about it.\n\n" +
	"<alignment> is a graded LEVEL from -1 to 1. USE THE WHOLE RANGE, not just the ends:\n" +
	"  +1      = central — the framing is directly about this item\n" +
	"  0 to 1  = loosely / partly supported — related but not central. Score proportional to how " +
	"much the framing accounts for it.\n" +
	"   0      = unrelated — the framing neither supports nor conflicts with it\n" +
	"  -1 to 0 = in tension — the item cuts against the framing or belongs elsewhere\n" +
	"Score NEGATIVE ONLY if the item actively conflicts with the framing. \"Loosely related\" is " +
	"NOT negative — it is positive-and-below-1. The framing is the truth; judge each item against " +
	"it, not the other way around.\n\n" +
	"Output only:\n<grades>\n  <g n=\"1\">0.8</g>\n  <g n=\"2\">-0.3</g>\n  ...\n</grades>"
