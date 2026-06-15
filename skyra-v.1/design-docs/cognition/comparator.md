# The Comparator — Coherence's Output Structure

## What It Is

The comparator is what Coherence (ACC) emits on descent. It is one structure for all three mismatch domains, because content, prediction, and articulation are three instances of a single operation: **compare a reference against an observation, and signal how aligned they are.** This is the same comparator the whole goal/prediction thread converges on — completion is a comparison (satisfied-when vs. world), prediction is a comparison, articulation is a comparison. One primitive, three domains.

Coherence's descent output collapses to a list of these comparisons. Nothing else — no separate delta, domain label, or result flag. Each comparison is one junction; the set of them is the frame.

## The Structure

```
<comparator>
  <comparison>
    <expected>#2</expected>        — one or more cells: the reference
    <actual>#3, #7</actual>        — one or more cells: what's measured against it
    <alignment>-1.0 .. 1.0</alignment>
    <basis>#5</basis>              — optional: other cells that informed the call
  </comparison>
  <comparison> ... </comparison>
  …as many as there are real junctions
</comparator>
```

- **`<expected>`** — the reference side. Always the binding reality's own content: a setpoint (Goal), a prediction (Prediction), or an intent (Action). One or more cell indexes.
- **`<actual>`** — the observed side. A member thought, what arrived, what motor produced. One or more cell indexes.
- **`<alignment>`** — a signed scalar, −1 to 1. See below.
- **`<basis>`** — optional provenance: other cells that informed the judgment but aren't the two operands.

The operand names are neutral — `expected`/`actual`, not `thought`/`against` — because a comparison records a **match** as readily as a mismatch. The system must learn from being right, not only from being wrong.

## Alignment Is a Signed Level

`<alignment>` runs −1 to 1:

- **+1** — actual meets expected. For a goal's setpoint, this is **completion**.
- **−1** — actual is the opposite state, the anti-setpoint.
- **0** — unrelated. No junction worth recording; not emitted.

Sign carries match vs. mismatch; magnitude carries strength. This replaces a separate `result` flag and a separate `weight` — one number is both.

**It is a level, not a per-stimulus delta.** It measures how close the world is to the reference *right now* — a position, because comparing `actual` (current state) against `expected` (the reference) yields a position by construction. "Is this working?" — whether the being is *getting* there — is the **trend of alignment across frames**, the derivative, not the stored value. Store the level; movement is the difference between consecutive levels. This is the *"trend of the delta across traversals"* the Goal comparator already tracks for its four outcome categories (rising = controllable, flat/falling = abandon).

**The cosine it always wanted.** This conversation began at "cosine similarity can't measure completion — 0.47 for *hello* against *greet the user*," and lands on a number on the exact same [−1, 1] scale as cosine. The difference is the basis: this is not embedding geometry, it is the recognition the LLM provides against the *observable* satisfied-when ([Decision #8](../architecture/v1-architecture-decisions.md)). It looks like cosine and it is the thing cosine was failing to be.

**It is the learning signal.** [Decision #12](../architecture/v1-architecture-decisions.md) wants Hebbian reinforcement proportional to strength. Sign and magnitude both fall out of alignment: positive strengthens the expected↔actual coupling, negative decays it. The comparison's alignment *is* the weight update.

## Domain Is Derived, Not Declared

The LLM never emits `content | prediction | articulation`. It never sees those words. The engine derives the domain from the cells the LLM cited.

**The path is the type.** A member cell cannot be reached without descending through its binding reality. A cell reached through Goal is a setpoint in this frame; the same dumb cell reached through Prediction is a prediction. So the type is not a property of the cell — it is a fact about how the wave arrived, stamped onto the **write**, not the cell. Cells stay dumb containers.

**Write-time stamping reuses competence-on-the-wave.** No new mechanism. The wave already carries the binding reality's competence as payload (the "[competence travels on the wave](episodic-memory-summary.md)" pattern). That competence is present at `WriteToSlice()`. The write records *(cell content, the binding-reality context the wave is carrying)*. The stamp is just the wave's current competence, captured at write time — the same dumb side-effect that does episodic binding. One write types the entry and binds it into the frame.

**The index space is over frame writes, not bare cells.** Each indexed thing is a cell-as-reached-through-a-path. A shared cell crossed via two binding realities this traversal is two indexes with two stamps — correct, not a bug (it matches "same cell, different conjunctions, different episodes" in [episodic-encoding](episodic-encoding.md)). So every cited index has exactly one unambiguous stamp; the domain never has to disambiguate.

**Domain is the `<expected>` side's stamp.** The reference is always the binding reality's own content, so the reference's path determines the kind of comparison:

| `<expected>` reached through | domain | the question | negative alignment routes energy to… |
|---|---|---|---|
| **Goal** (a setpoint) | content | "are my beliefs / state meeting the setpoint?" | the contradicting cell — split or absorb |
| **Prediction** (a prediction) | prediction | "did the world surprise me?" | the failing goal — gather before committing |
| **Action** (an intent) | articulation | "did I say what I meant?" | the motor/action loop — tighten expression |

Belief-against-belief with no setpoint (neither side a prediction or intent) defaults to **content** — internal consistency.

The three domains are the three things a being can be wrong about: its own consistency (content), the world (prediction — the exafference delta), and itself (articulation — the efference delta). See [predictive-hierarchy](predictive-hierarchy.md).

**No hole where the path mis-types.** You can only make a goal-completion comparison if Goal was on the wave this traversal — otherwise its setpoint isn't a frame write and there's nothing to compare against. You cannot check closure on a goal you didn't activate, which is exactly right.

## Match Reinforces, Mismatch Routes

- **Mismatch** (alignment < 0) → the engine routes energy per the derived domain (split the cell, raise goal dominance and gather, or tighten the motor loop). Coherence is the stress center — the error signal *is* the energy (see [predictive-hierarchy](predictive-hierarchy.md)).
- **Match** (alignment > 0) → reinforce the expected↔actual coupling. These thoughts *confirmed* the reference; deepen the groove. This is how "the thoughts that drove a good decision" get saved, and how a reflex forms ([Decision #15](../architecture/v1-architecture-decisions.md)) — repeated matches are what let inference stop firing.

## The LLM Contract

Domain-derivation removes all architecture vocabulary from the prompt. The LLM's entire job: *here are the cells in play, numbered; pick any that conflict or drive toward a setpoint, and score how aligned they are from −1 to 1.* Pure pattern-matching over an indexed set — what LLMs are best at — with no jargon to misuse.

It is grounded by construction: the LLM can only cite indexes that exist in the frame, so it cannot fabricate a thought. The working set *is* the vocabulary — the observed-not-asserted discipline ([Decision #16](../architecture/v1-architecture-decisions.md)) made structural.

It self-limits: only conflicts and drivers are worth citing; everything at alignment ≈ 0 gets no comparison. Output is naturally sparse — the prediction-gating story again (most things predicted → few junctions → thin frame). The slice's capacity window caps it if it ever runs long.

### Canonical prompt block

```
Here are the cells active right now, each numbered.

For any cells that conflict, or that drive toward something you are trying to
reach, make a comparison:

<comparison>
  <expected> — the cell(s) that set the standard: what should be true.
  <actual>   — the cell(s) measured against it: what is true.
  <alignment> — from -1 to 1. +1: actual meets expected. -1: actual is the
                opposite. 0: unrelated (don't bother recording it).
</comparison>

Make as many as you see. Only cite cells from the list above.
```

## Goal Closure — the Throughline

`<alignment>` → 1 on a goal's *satisfied-when* comparison **is the goal-closure signal** — the throughline experiment .1001 went looking for. The being knows it's done not because the act resembles the goal, but because the world matches the setpoint, and the comparator reads that match as a 1.

## The Frame Is Comparisons, Not Just Cells

A frame is not a flat list of co-present cells — it is the **junctions** among them: what was checked against what, and whether it held. That is a richer conjunctive code than a member list — the hippocampal binding pattern with the *edges* included, not just the nodes. `<basis>` is the nodes; `<comparison>` is the nodes plus the relationship.

## Global and Local Coherence — One Method, Two Scopes

Coherence runs at two scales, and it is the *same method* both times. The central cell sits behind the membrane (north of sensory, south of motor) and owns **global** alignment — does the being, as a whole, cohere? The binding realities own **local** alignment — does this one index cohere internally? But local coherence isn't separately implemented. The central cell **throws the coherence and frontopolar competence down the wave**; when it reaches the binding realities, it runs there, localized.

**Scope is positional, not parametric.** Nothing passes a "run locally" flag. The coherence method reads *"my slice / my members."* Where it executes rebinds what "my" means — at the central cell, "my slice" is the integrated returns from all binding realities; thrown down to Goal, it is Goal's own members. Same code, scope set by position. This is the [local-relativity invariant](../laws/invariants/local-relativity-invariant.md) doing the work: the same law evaluated in each frame yields local results. Local coherence falls out for free because coherence run *at a location is already scoped to that location.* No machinery moves.

This lands on the comparator above: **local coherence is the content-domain comparison** (setpoint vs. members) at each binding reality; **global coherence is the same comparison** at the central cell over the returned votes. One comparator, two scopes — not two implementations.

### The Cell-Execution Contract

Passed competence **brackets** the cell's own work — primes it on the way in, audits it on the way out:

- **Observe (descent):** run passed competence **first**, then the cell's own. The passed method is top-down context (a prediction that needs no local input), so it primes the local cell before it scores.
- **Express (ascent):** run the cell's own **first**, then passed competence **last**. The passed method is the frontopolar check, which needs the local vote to already exist — it audits the cell *after* it votes.

The ordering is not arbitrary — it falls out of dataflow direction: context before content going in, check after commit going out. It is the onion/middleware shape, and it **nests**: a binding reality can throw its own competence to its members, stacking global-wraps-local-wraps-local-local. Fractal coherence.

**The contract lives in the cell, not the kernel.** Every cell's Observe runs passed-then-own; every Express runs own-then-passed. The kernel never looks inside the payload — it just delivers the wave. The bracketing is the cell's contract, the same way executing a passed payload against your content is already the cell's job. Kernel opacity is preserved.

### Two Frontopolar Checks — Gather at Two Scales

The frontopolar audit fires at both scales:

- **Local** — a binding reality votes, then its localized frontopolar asks "did *my* evidence cover this?" If not, it can raise a local `<need>`.
- **Global** — the central cell integrates the votes, then its frontopolar asks "does the being, as a whole, have what it needs — or do I gather before I respond?" This is how a being decides to ask a question before answering.

**The global verdict completes on ascent.** When the central cell "makes its global case" on descent, it is only setting the global frame and throwing it down — it has no local results yet. Global resolution is the ascent frontopolar, the last thing before motor. Prime down, verify up — predictions down, errors up. This is `predictive-hierarchy.md` realized through competence-on-the-wave: the hierarchy's "each layer translates the prediction into its own vocabulary" *is* the same coherence method re-executing, localized, at each level.

## The Vote Flow — Global and Local Behave Differently

Global and local run the *same* vote/question pair, but with the emphasis flipped by position:

- **Local: vote-primary, question-after.** A binding reality fires its prediction, runs its wave, and on the return casts its vote, then self-checks — the deterministic **vote → question** order (the own-first / passed-last express contract). It always proposes.
- **Global: question-primary, vote-exceptional.** The central cell monitors, and casts a vote **only to break a tie.** It stays silent while the regions agree.

That asymmetry is the ACC's real job — engage on conflict, not on smooth processing. Aligned votes fire action reflexively (basal-ganglia coincidence detection, no center needed); the center only engages when the regions can't resolve among themselves. The non-gating property survives even into the tie-break.

```
local: fire prediction → wave → (return) vote → question
once votes are cast, they return to global
  aligned        → kick off action immediately (center never engages)
  not aligned    → global's question method either:
                     (a) break the tie  — cast the one deciding vote, raise a question
                     (b) inner voice    — inject a reframed impulse back through sensory;
                                          the topology re-propagates and the regions re-resolve
```

**(b) is inner speech.** Vygotsky's self-regulating inner voice, realized as re-entrant processing (Edelman/Dehaene) — the broadcast loops back as input. It fits the machinery exactly: it reuses the **sensory path** (the inner voice enters where external perception does); it carries the **"from me"** tag, so on return it folds into the workspace as reafference, not as new external news; it is the *"deliberation as re-descent"* of [predictive-hierarchy](predictive-hierarchy.md), routed through sensory so the *whole* topology reframes rather than one stuck branch; and it is **powered by the conflict itself** — the tie is a mismatch → stress → Coherence routes that energy into the re-entry, and **amplitude attenuation bounds the loop** (it fades after a few passes). That bound is the line between deliberating and ruminating: the being thinks until it resolves or runs out of energy. Mode (a) imposes; mode (b) reframes and lets the autonomous regions own the resolution — truer to the non-gating spirit.

## This Is a Global Workspace

The gather mechanism is Global Workspace Theory, derived rather than imposed. Global spawns the binding realities (the [owned-specialists crossing](../kernel/world-kernel-coupling.md#two-crossing-modes--peer-and-owned) — `Content` as a map of Spaces, threaded, origin = global). Each returns its top-aligning items as a "from me" wave. Global **deposits** them to its slice — it does not fire its question on these self-returns, because the origin split ("is this from me?") makes a self-return a silent deposit, not a question-fire (the same kernel rule that return waves don't re-stimulate the topology). A **cutoff** keeps only the top-aligning first-pass returns.

The correspondence is exact:

| Global Workspace Theory | here |
|---|---|
| specialized parallel processors | the binding realities |
| competition for access | the cutoff — only top-aligning returns survive |
| the broadcast | the deposited context written to global's slice (a sub-reality) |
| ignition | global fires the alignment check when the set is complete / hot |
| conscious content this cycle | what won access — the shared workspace |

Crucially, **global sees the same workspace the regions see** — only the top-activating broadcast, not a privileged view. The center reads the broadcast like everyone else, which is *why* it stays non-gating and removable. The workspace is also the concrete answer to "what is the being attending to this cycle": whatever won access.

The kernel never knows any of this is happening — it spawns waves and delivers them to origins. Accumulation, the cutoff, ignition, and the tie-break are all the cell's policy on its async returns. Transport is the kernel's; gather is the cell's.

## Dependencies & Open

- **The one hook:** `WriteToSlice()` must capture the wave's competence alongside the write. It is the same hook that does episodic binding — the write types the entry and binds it in one dumb side-effect.
- **Level vs. trend storage:** alignment is the stored level; the trend (derivative across frames) is what the Goal comparator reads for controllable/uncontrollable. Where the trend is computed and held (Goal binding reality vs. slice) is open.
- **Frontopolar `<need>` structure:** the ascent *mechanism* is now specced (above — it runs as bracketed competence, local and global, gather-before-commit). What's still loose is the output tag itself: `<need>` is a satisfied-when one level down (an epistemic sub-goal whose setpoint is a perception), but its exact fields and how it launches as a `perceive` aren't drawn yet.
- **Multi-cell operands:** `<expected>`/`<actual>` accept sets. How a set-vs-set alignment is scored (aggregate vs. worst-case) is unspecified.
- **Where the hippocampus indexes hang:** global's `Children` vs. each binding reality's own `Children` (the distributed-indexes/shared-store question from [episodic-memory-summary](episodic-memory-summary.md)). The owned-specialists crossing works either way; this only decides who owns the door.
- **What selects break-tie vs. inner voice:** mode (a) vs. (b) on a deadlock. Likely energy/urgency — cheap tie or must-act-now → (a); budget to deliberate → (b) — which makes the inner voice a *paid* deliberation, governed by the same energy that bounds it. Not settled.
