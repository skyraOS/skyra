package dnav2

import (
	"fmt"
	"sort"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// midlevel:frame — assemble the exchange from the entity index, infer the frame
// (real inference), create the episode.observeExternal
// (cold-start create path). Sets root.Episode so the membrane threads it to the
// fan-out's predict workers. Reports "framed" / "empty".
func midlevelFrame(root *EphemeralRoot, origin skyra.Space, impulse any) {
	episodes := root.DNA().Episodes()
	exchange := renderExchange(root.Nodes)
	// the current surroundings — the active episode's traversed perceive records,
	// normalized (freshest per entity). Lets midlevel frame purely SPATIAL episodes
	// (the maze; sharpening the default episode from sense data alone) — decision
	// 4, perceive-build-uncertainties.md. Read-only assembly: no crossing here (the
	// settle crosses what it grades; self's assembly crosses what it feeds).
	locations := ""
	if g := episodes.Active(); g != nil {
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
	activeSet, hasActive := root.DNA().ActiveSet()
	activeID := ""
	if hasActive {
		activeID = activeSet.PresentID
	}
	offered := structuralCandidateIDs(root.DNA().CausalNodes(), activeID, root.ObservationID)
	present := structuralPresent(root.DNA(), exchange, locations, activeID, root.ObservationID, offered)
	frameInfer := root.Infer("frame-op")
	resp, err := frameInfer(frameOpSystem, present)
	if err != nil {
		root.Condition = "empty" // a failed call frames nothing this cycle
		return
	}
	op, parsed, perr := validateStructuralOp(root.DNA().CausalNodes(), resp, hasActive, offered)
	for attempt := 1; perr != nil && attempt <= parseMaxRetries; attempt++ {
		root.Trace.Reject("frame-op", resp, perr)
		resp, err = frameInfer(frameOpSystem, present+"\n\n---\n"+StructuralCorrectionMessage(resp, perr))
		if err != nil {
			break
		}
		op, parsed, perr = validateStructuralOp(root.DNA().CausalNodes(), resp, hasActive, offered)
	}
	if perr != nil {
		root.Trace.Reject("frame-op", resp, perr)
	}
	ts := root.DNA().Tick()

	// Sentence + subgraph commit transactionally. A format failure is returned to
	// the LLM at the same epistemic budget; after bounded retries, commit neither
	// the active frame or Aitia set. The previous active state remains valid.
	if perr != nil {
		debug.Being("skyra", "querier", fmt.Sprintf("annotation failed after retries (%v) — retaining active state", perr))
		root.Condition = "query-format-error"
		return
	}
	if op == StructuralRetain {
		root.Episode = episodes.Active()
		if activeSet.EfficientEdgeID == "" || activeSet.FinalEdgeID == "" {
			set, err := buildFullCausalSet(root.DNA(), root.Trace, activeID, ts)
			if err != nil {
				debug.Being("skyra", "causal", fmt.Sprintf("expansion failed: %v — retaining prior active state", err))
				root.Condition = "causal-format-error"
				return
			}
			root.DNA().ActivateSet(set.ID)
		}
		root.Trace.Note("frame-op", "retain "+activeID)
		debug.Being("skyra", "querier", fmt.Sprintf("retain causal-node %s", activeID))
	} else {
		comp := root.DNA().CausalNodes().Persist(parsed, ts)
		frame := parsed.Sentence
		set, err := buildFullCausalSet(root.DNA(), root.Trace, comp.Cell.ID(), ts)
		if err != nil {
			root.Trace.Note("frame-op", fmt.Sprintf("expansion failed — %s persisted but orphaned (no set)", comp.Cell.ID()))
			debug.Being("skyra", "causal", fmt.Sprintf("expansion failed: %v — retaining prior active state", err))
			root.Condition = "causal-format-error"
			return
		}
		if activeEpisode := episodes.Active(); activeEpisode != nil {
			root.Episode = episodes.Edit(activeEpisode.ID(), frame, ts)
		}
		if root.Episode == nil {
			episodeID := fmt.Sprintf("%s:episode:%d", root.DNA().BeingID(), ts)
			root.Episode = episodes.Create(episodeID, frame, ts)
		}
		if op == StructuralCompose && root.ObservationID != "" {
			attachCompositionEvidence(root.DNA().CausalNodes(), comp.Cell.ID(), root.ObservationID)
		}
		root.DNA().ActivateSet(set.ID)
		if g := latestFrame(root.Episode); g != nil {
			g.SetContent(Frame{Text: frame, Timestamp: ts})
		}
		root.Trace.Note("frame-op", fmt.Sprintf("%s → causal-node %s %q, set %s active", op, comp.Cell.ID(), frame, set.ID))
		debug.Being("skyra", "querier", fmt.Sprintf("%s causal-node %s: %q V=%d E=%d → set %s active",
			op, comp.Cell.ID(), frame, len(comp.Parents), len(parsed.Edges), set.ID))
	}

	consolidateInteraction(root) // hang the exchange on the episode + entity→episode

	// The coherence-settle: with the frame now FIXED, grade the episode's nodes against it and
	// write alignment (A) onto their edges, pruning misfits until it plateaus. midlevel
	// withholds its frame until here — the settle is the sole writer of A (active-grouping.md).
	passes := coherenceSettle(root.Episode, root.DNA().Activation(), ts, root.Infer("settle"))
	root.Trace.Note("settle", fmt.Sprintf("settled %s in %d passes", root.Episode.ID(), passes))
	debug.Being("skyra", "midlevel", fmt.Sprintf("settled %s in %d passes", root.Episode.ID(), passes))

	root.Condition = "framed"
}

func buildFullCausalSet(d *DNA, tr *RunTrace, presentID string, ts int) (AitiaSet, error) {
	efficientParsed, finalParsed, err := expandCausalBeliefs(d, tr, presentID)
	if err != nil {
		return AitiaSet{}, err
	}
	efficient := d.CausalNodes().Persist(efficientParsed, ts)
	final := d.CausalNodes().Persist(finalParsed, ts)
	return d.CausalNodes().MintFullSet(presentID, efficient.Cell.ID(), final.Cell.ID(), ts)
}

func validateStructuralOp(ci *CausalNodes, raw string, hasActive bool, offered map[string]float64) (string, ParsedSubgraph, error) {
	op, annotated, ok := ParseStructuralOp(raw)
	if !ok {
		return "", ParsedSubgraph{}, fmt.Errorf("expected exactly one <retain/>, <compose>...</compose>, or <requery>...</requery> operation")
	}
	if op == StructuralRetain {
		if !hasActive {
			return "", ParsedSubgraph{}, fmt.Errorf("retain requires an active Material/Formal subgraph")
		}
		return op, ParsedSubgraph{}, nil
	}
	parsed, err := ParseAnnotated(annotated)
	if err != nil {
		return "", ParsedSubgraph{}, err
	}
	causalRefs := 0
	for _, node := range parsed.Nodes {
		if node.StableID != "" && ci.Get(node.StableID) != nil {
			if _, allowed := offered[node.StableID]; !allowed {
				return "", ParsedSubgraph{}, fmt.Errorf("CausalNode %q was not offered in this structural pass", node.StableID)
			}
			causalRefs++
		}
	}
	if op == StructuralCompose && causalRefs == 0 {
		return "", ParsedSubgraph{}, fmt.Errorf("compose must reference at least one existing CausalNode by stable ID")
	}
	if op == StructuralRequery && causalRefs != 0 {
		return "", ParsedSubgraph{}, fmt.Errorf("requery must emit a standalone subgraph without CausalNode references")
	}
	return op, parsed, nil
}

func StructuralCorrectionMessage(raw string, err error) string {
	return fmt.Sprintf("Your previous response was:\n%s\n\nIt was invalid: %v. Re-emit the same decision as exactly one operation: <retain/>, <compose>annotated sentence</compose>, or <requery>annotated sentence</requery>. Compose must cite an offered CausalNode ID; requery must be standalone.", raw, err)
}

func attachCompositionEvidence(ci *CausalNodes, presentID, observationID string) {
	for _, record := range ci.EdgesOf(presentID) {
		if record.Kind == EdgeInherited {
			_ = ci.AttachEvidence(record.ID, observationID)
		}
	}
}

// consolidateInteraction attaches the just-arrived interaction to the framed
// episode's CURRENT frame (records belong to the framing they arrived under) and
// wires the entity → episode edge — so the exchange lives on the spine the frame
// walks (into the frame + economy), not only in the entity index.
func consolidateInteraction(root *EphemeralRoot) {
	if root.Episode == nil {
		return
	}
	frame := latestFrame(root.Episode)
	if frame == nil {
		return
	}
	node, entity, ok := root.Nodes.Latest()
	if !ok {
		return
	}
	root.DNA().Episodes().AddChild(frame, frame.ID()+"→"+node.ID(), node)
	root.Nodes.LinkEpisode(entity, root.Episode)
}

// framePresent renders the framing-op present: the exchange + the current
// surroundings (normalized perceived locations, when present) + the open episodes
// the op decides edit-or-new against.
func framePresent(exchange, locations string, episodes []struct{ ID, Frame string }) string {
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
			b.WriteString(fmt.Sprintf("[id=%s] %s\n", e.ID, e.Frame))
		}
	}
	return b.String()
}

func structuralPresent(d *DNA, exchange, locations, activeID, observationID string, offered map[string]float64) string {
	var b strings.Builder
	if locations != "" {
		b.WriteString("SURROUNDINGS:\n")
		b.WriteString(locations)
		b.WriteString("\n\n")
	}
	b.WriteString("EXCHANGE:\n")
	if exchange == "" {
		b.WriteString("(none)\n")
	} else {
		b.WriteString(exchange)
	}
	b.WriteString("\nACTIVE MATERIAL/FORMAL:\n")
	if activeID == "" {
		b.WriteString("(none — cold start requires requery)\n")
	} else {
		b.WriteString(renderCausalStructure(d.CausalNodes(), activeID))
	}
	b.WriteString("\nOBSERVATION SUBGRAPH:\n")
	if observationID == "" {
		b.WriteString("(none)\n")
	} else {
		b.WriteString(renderCausalStructure(d.CausalNodes(), observationID))
	}
	b.WriteString("\nOFFERED COMPOSITION CANDIDATES:\n")
	ids := make([]string, 0, len(offered))
	for id := range offered {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		b.WriteString(fmt.Sprintf("[candidate=%s amplitude=%.4f] %s\n", id, offered[id], causalSentence(d.CausalNodes().Get(id))))
	}
	if len(ids) == 0 {
		b.WriteString("(none)\n")
	}
	return b.String()
}

func structuralCandidateIDs(ci *CausalNodes, activeID, observationID string) map[string]float64 {
	scores := map[string]int{}
	if query, ok := ci.QueryOf(observationID); ok {
		for _, candidate := range ci.Candidates(query) {
			scores[candidate.CausalNodeID] = candidate.Occurrences
		}
	}
	// The active and observation nodes are always legal composition referents,
	// even when reverse retrieval has no overlap score for one of them.
	for _, id := range []string{activeID, observationID} {
		if id != "" && scores[id] == 0 {
			scores[id] = 1
		}
	}
	total := 0
	for _, score := range scores {
		total += score
	}
	out := map[string]float64{}
	if total == 0 {
		return out
	}
	for id, score := range scores {
		out[id] = float64(score) / float64(total)
	}
	return out
}

func renderCausalStructure(ci *CausalNodes, id string) string {
	cell := ci.Get(id)
	if cell == nil {
		return fmt.Sprintf("[id=%s] (missing)\n", id)
	}
	content, _ := cell.Content().(CausalNodeContent)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[id=%s] %s\n", id, content.Sentence))
	for _, edge := range ci.EdgesOf(id) {
		if edge.Kind == EdgeStructural {
			b.WriteString(fmt.Sprintf("  %s --%s--> %s\n", edge.FromID, edge.Label, edge.ToID))
		}
	}
	return b.String()
}

// renderExchange turns the entity index's interaction records into a timestamp-
// ordered transcript (who said what) — the present the framing pass reads.
func renderExchange(ei *NodeRegistry) string {
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
			mc, ok := c.(*MemoryCell)
			if !ok {
				continue
			}
			if rec, ok := mc.Content().(InteractionRecord); ok {
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

// coldStartSystem — the framing prompt.
const coldStartSystem = "You are an evolving being reading a live exchange. In one short " +
	"phrase, predict what this exchange is about and what you should be trying to do. " +
	"Respond with only the tag: <label>your framing</label>"

// frameOpSystem is the ungraded Material/Formal structural controller.
const frameOpSystem = "You are an evolving being organizing your own memory. You are given " +
	"the OBSERVATION SUBGRAPH and the ACTIVE MATERIAL/FORMAL representation, when one exists. " +
	"Material/Formal is structural and ungraded. Choose exactly one operation:\n\n" +
	"- <retain/> when the active representation already fits without structural change.\n" +
	"- <compose>annotated sentence</compose> when the observation extends it. Compose MUST " +
	"reference at least one offered CausalNode using [ordinal=causal:id].\n" +
	"- <requery>annotated sentence</requery> when no active representation exists or replacement " +
	"is required. Requery MUST be standalone with no CausalNode stable reference.\n\n" +
	"For compose/requery, emit ONE COMPLETE ANNOTATED SENTENCE:\n" +
	"- mark each referent with an ordinal:  michael[1]  or  {the skyra project}[2]\n" +
	"- mark each relationship on its verb/word:  asked[1→2]\n" +
	"- reuse an ordinal to re-reference the same thing:  michael[1] ... he[1]\n" +
	"Every referent gets a node annotation; every asserted relationship gets an edge " +
	"annotation; plain connecting words stay unannotated. Reply with only the operation."

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
