# Motivation — Dependency Graph Architecture

## The Insight

Motivation (OFC) maintains a dependency graph of what must remain true for each goal to be valid. It doesn't track goal progress — Goal does that. Motivation tracks whether the *reason* for the goal still holds.

## Transitive Dependency Chain

```
Attention (Surroundings on Slice)
    ↓ episodes depend on attention (scored by overlap with active cells)
Episodes (Episodic binding reality)
    ↓ goals depend on episodes (goals span episodes, indexed by them)
Goals (Goal binding reality)
    ↓ dependencies are conditions that must remain true
Dependencies (Motivation's graph)
```

Goals are transitively dependent on attention. If the attention map shifts, episodes shift, and the goals that depend on those episodes lose their grounding.

## Two Monitoring Modes

**Passive** — the dependency cell is in Surroundings. Motivation reads it for free every traversal. No inference, no goal, no action needed. The condition is in the being's perceptual field.

**Active** — the dependency faded from Surroundings. Motivation can't verify passively. It raises an epistemic goal: "bring X back into attention." Routes through Goal → Action → Motor → PrimitiveLook. Result lands back in Surroundings, Motivation reads it, reinforces or kills the dependent goal.

The threshold is attention decay, not condition failure. Motivation detects that a condition *left its field of view*, not that it broke. Uncertainty triggers the check.

## Indexing Structure

```
Base layer:    episodeID → Episode (member set, from Episodic)
Second layer:  goalID → Goal (spans multiple episodes)
Dependencies:  goalID → []cellID (conditions that must remain true)
```

- A goal can span multiple episodes. Born in episode A, persists through B and C, completed in D.
- Episodic segments underneath. Goal persists across segments. Motivation monitors dependencies regardless of which episode is current.
- The active episode tail (Episodic.Score gradient) tells Motivation which goals are potentially relevant — reinstated episodes can wake their associated goals.

## Episode ↔ Goal Linkage

Goals don't belong to a single episode. They slide across episode boundaries:

- **Same goal, new episode:** When Episodic segments (prediction spike → goal evaluates → new episode), Motivation determines whether the active goals carry forward into the new episode or close with the old one.
- **Motivation decides the slide:** If the goal's dependencies are still valid → goal slides to the new episode. If dependencies broke → goal stays with the old episode (abandoned).
- **This is the segmentation signal source.** The doc says "Motivation revalues → conditions change → signals Goal → goal changes → signals Episodic → new episode boundary." Motivation's dependency check is what initiates the whole chain.

## OFC Biology Mapping

- OFC maintains a "cognitive map" of task contingencies — a dependency graph
- OFC is episode-indexed but not episode-confined — persistent graph, context-sensitive access
- OFC raises its own epistemic goals ("verify this condition") — information-seeking with no reward value
- OFC neurons become more active when contingencies shift — change detector on preconditions
- Same mechanism as Episodic retrieval (CA3 pattern completion) but asking a different question: not "what matches?" but "are my conditions still present?"

## Multiple Episodes Active Simultaneously

One dominant episode plus a gradient tail. Same mechanism (CA3 pattern completion), different input strengths:

- Current episode: full sensory input, scores ~1.0
- Reinstated episodes: partial cues, graded activation (0.3–0.7)
- They coexist until overlap gets too high → system separates or integrates

The tail gives Motivation access to goals from partially active episodes. A reinstated episode can wake dormant goal dependencies for passive monitoring.

## Open Questions

- When exactly does Motivation decide a goal slides vs closes? Is it threshold-based (dependency activation below X = close) or binary (any dependency gone = close)?
- How does Motivation handle conflicting dependencies across coactive goals?
- Does the dependency graph need decay? Old goals with stale dependencies that never got checked — do they just accumulate?
- Relationship to the doc's "four conditions on ascent" — those map to dependency states: still valid, strengthened, weakened, gone.
