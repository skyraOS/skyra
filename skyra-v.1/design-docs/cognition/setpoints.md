# Setpoints — The Global Setpoint and the Subgoal Tree

> **Reconciliation (2026-06-23).** Vocabulary drift; reasoning holds. Read
> **"Coherence" → Global**, **"binding realities" → the self/other agents**. A
> setpoint is now **one `MemoryCell`** (`Content` = the understanding/gist;
> `Children` = the closure snapshot — entities + responses + actions + stresses;
> past-only, no parent/future edge — the future is Global's grouping index). Closure
> is judged by **`completed`** (signed −1..1; `+thr → crystallize`, supporting `< 0`
> → cascade), distinct from the relational `alignment` in the activation bump. Current
> spec: [../../specs/2026-06-16/skyra-dna/setpoint.md](../../specs/2026-06-16/skyra-dna/setpoint.md);
> [working-memory.md](working-memory.md).

## Status

**Planning — discuss before implementing.** Records the ownership of the global setpoint (settled this session) and holds the subgoal-structure work (open, next). Pairs with [comparator.md](comparator.md) (alignment as the closure signal) and the [voting set](../kernel/owned-specialist-crossing.md#6-the-voting-set--brain-region-grounding).

## The Global Setpoint Is Defined by the User at Birth

The being's **root drive** — the anchor everything else keys on — is **authored by the user when the agent is born**, the way the genome encodes hypothalamic homeostatic setpoints. For the prototype this can stay plain: **"be helpful to the user."** It does not need special identity machinery to carry v.1. The creator who writes the genome plays the role evolution plays in biology: it specifies the root setpoint, and the being is born already wanting it.

This resolves the anchor question behind the reward signal: **dopamine is alignment** (the comparator's signed `<alignment>`, read as a trend), but alignment needs something to be aligned *toward*. The birth-given global setpoint is that referent. Without it the reward signal has no ground and is just noise.

### Why this is structural, not assumed

Identity + purpose were founding prompts for five prior runtimes. They were pulled for v.1 to test whether the architecture would *need* them. It re-derived one piece — a birth-given global setpoint — from an unrelated constraint (the reward loop needs an anchor), not from assumption. A thing you remove that reassembles itself from a requirement you weren't thinking about is structural. What did **not** come back is the rest of the old prompt (voice, character, values-essay): the result is "needs *an* anchor," not "needs the identity essay." Identity and voice remain unproven — they may precipitate out of the anchor plus accumulated history, or get added later, one isolated field at a time, each earning it.

## Ownership

> **Superseded — Motivation dissolved.** The "Motivation" owner/home references in this section and in *DNA Placement* below predate the dissolution: the root drive now lives in the **membrane's `Content`**, and Motivation is no longer a binding reality (it became the setpoint's *descent method*). See **Motivation Dissolves** below. Kept for history.

**Keying-on is reachability, not ownership** ([cell-types.md](../being/cell-types.md): "ownership is separate from reachability"). Coherence *reads* the setpoint to compare against it; it does not own it. The owners split by level:

| Object | What it is | Owner | Where it lives |
|---|---|---|---|
| **Root drive** (the anchor) | a standing drive, **no `satisfied-when`** — never completes, only ever *served* | **Motivation** (hypothalamic-nuclei role) | **DNA-specified**, user-authored at genesis |
| **Derived setpoints** (the tree) | goals with an observable `satisfied-when` — closable | **Goal** (PFC role) | emergent — born by convergence at runtime, never in the DNA |
| **Alignment monitoring** | reads dominant setpoint, scores `<alignment>` | **Coherence** (ACC role) | owns *neither* |

The root drive is **not** a setpoint in the `satisfied-when` sense — it has no completion (you don't finish being not-hungry). The setpoints with completion are the *derived* goals that serve it.

## DNA Placement

The genome today has no setpoint field:

```
being ~type llm ~bind episodic,action,goal,motivation,emotion,working_memory ...
```

The root drive is a **new, user-authored genome attribute** (e.g. `~drive "..."` / `~serves michael`), installed into Motivation at boot. The bind list is unchanged. Derived setpoints stay out of the genome entirely — they emerge ([cell-types.md](../being/cell-types.md): "goal formation = goal convergence, no author").

**~~Open fork~~ → resolved:** the root drive lives in the **membrane's `Content`** — Motivation dissolved as a binding reality (see **Motivation Dissolves** below), so there is no "Motivation" to hold it. (Originally a fork between Motivation-holds-it vs. membrane-Content; the dissolution settled it on Content.)

## Ownership Fades — the Setpoint → Drive Seam

The user owns the *first* root, but that ownership is meant to **fade**, by two of the project's own principles:

- **Biology** — born with given drives, develops self-authored values that can outrank them.
- **Skyra-as-God** — the creator "maintains universal principles, active early then fades."

The mechanism already exists: the **promotion gradient** (trace → understanding → skill) extended one level — **setpoint → drive**. A derived setpoint that keeps closing, keeps mattering, gets promoted to a root the being now owns itself. The user authors root #1; the being earns the rest.

**Build decision:** for the minimal being, the global setpoint is **fixed at birth (user-owned)**. The promotion-to-drive path is a **marked seam** (v.2), written down here so it isn't re-derived — it needs the full reinforcement loop working first.

## Subgoals — Open / Next

How the derived tree below the root is structured is the next piece. Captured starting points:

- **Nested local setpoints** — the Thanksgiving shape: `enjoy Thanksgiving` → `keep auntie comfortable` → `don't bring up politics`. Each level a setpoint whose `satisfied-when` serves its parent.
- **Local setpoints are dense reward, not just organization.** A sparse distant root reward is unlearnable (credit assignment). Local setpoints give proximal alignment-closure — a near-term reward the distant root can't. Nested setpoints are *how the manufactured dopamine becomes usable.*
- **Relational setpoints** — a distinct kind: a *standing state maintained about a relationship*, context-activated when that being is present ("with auntie: keep her comfortable"). Neither Goal (own commitment) nor Prospective (their explicit ask). Surfaced by the contextual binding fields ("People → which beings are present"). Candidate home for *one* root drive — "be in good standing with michael" as the relational root the whole tree bottoms out in.

Open questions for the subgoal structure: how a subgoal binds to its parent (an edge? a `serves` link? containment?); how parent alignment composes from child alignments; whether relational setpoints are a Goal subtype or their own cell; and how the tree is born/pruned by convergence under the root drive.

## Motivation Dependency Setpoints

> **Superseded in part — see "Motivation Dissolves" below.** Motivation is no longer a binding reality; its two jobs became methods (a past-completion reality-monitor and Coherence catching defeated assumptions). The dependency-validity content here still holds; only the "Motivation as a binding reality / Observe-face" framing is retired.

Refinement from the subgoal discussion: **Goal owns target decomposition; Motivation owns dependency validity.**

A subgoal and a dependency can both be expressed with the comparator's `<alignment>` scale, but they answer different questions:

| Structure | Owner | Question | Failure mode |
|---|---|---|---|
| **Subgoal** | **Goal** | What local state should become true to serve the parent setpoint? | create pressure to act toward closure |
| **Dependency setpoint** | **Motivation** | What condition must remain aligned for the parent setpoint to stay worth pursuing? | weaken, pause, repair, or abandon the parent |

So Motivation can be read as a graph of **dependency setpoints**: conditions that need to stay near `1` on the alignment scale for work on a parent setpoint to remain justified.

Example for the prototype root:

```text
root drive:
  be helpful to the user

parent setpoint:
  help the user reason through Skyra's motivation/subgoal architecture

Motivation dependency setpoints:
  the user wants brainstorming, not implementation
  the current answer preserves runtime boundaries
  enough context has been loaded
  the concepts are not being prematurely frozen
```

If these dependency setpoints remain aligned, Motivation continues funding the parent setpoint. If one drops, Motivation chooses the response:

```text
dependency aligned
  -> continue parent setpoint

dependency misaligned but repairable
  -> birth a Goal subgoal to restore it

dependency misaligned and not repairable
  -> weaken, pause, or abandon the parent setpoint
```

This keeps the two graphs distinct while letting them interact:

```text
Goal tree:
  desired states and substates

Motivation graph:
  reasons, prerequisites, blockers, and value conditions
```

A Motivation dependency can birth a Goal subgoal, but it is not itself the subgoal. The dependency says "this must hold for the wanting to remain valid." The subgoal says "make this repair state true."

This implies Motivation has the same global/local shape as Coherence:

```text
Local Motivation:
  evaluate dependency setpoints for one parent setpoint

Global Motivation:
  evaluate which active parent setpoints remain justified under the root drive
```

## The Setpoint Is the Recursive Primitive — Motivation/Goal as Descent/Ascent

The split resolves into the same one-cell-two-phases shape as Coherence (Predict/Resolve). The **setpoint** is the universal node; **Motivation is its Observe (descent) face, Goal is its Express (ascent) face**:

- **Descent (Motivation / Observe):** flow down through the dependency graph — the prior setpoints this one stands on — and **gate-check** each. Are my preconditions still aligned (~1)?
- **Ascent (Goal / Express):** if the gates held, flow back up **composing alignment** toward closure. Given my ground holds, am I satisfied?

Same graph, **read twice** — descended as gates, ascended as build. This is the comparator's "prime down, verify up" pointed at setpoints, and the "one field read twice" pattern (attention, Coherence, now the setpoint). Alignment is the quantity that flows; the recursion bottoms out at a **sensory leaf** — a setpoint with no dependencies, alignment read straight from the world ("auntie is smiling").

In kernel terms this is free: a setpoint's `Children()` *is* its dependency setpoints. The kernel descends them (Motivation checks), Expresses on ascent (Goal composes). The setpoint graph is just another plane the kernel already walks. No new machinery.

**The gate short-circuits a false closure.** Putting Motivation on descent means a dependency that comes back misaligned *on the way down* stops the wave before it can ascend into closure — Motivation routes instead (pause / repair-via-subgoal / abandon). You can't close a goal whose ground gave out. Observed-not-asserted applied to the preconditions, not just the target.

### Why this reads as a step in the right direction

It removed machinery instead of adding it: Motivation and Goal stopped being two structures and became the two faces every cell already has; the setpoint graph became a plane the kernel already traverses. In this architecture, a right step usually subtracts.

### Open forks (not resolved)

1. **One cell or two.** A single `Setpoint` cell wearing both faces (tightest — revises the genome to one bind and the voting set to one vote whose Express carries the Goal-proposal), vs. Goal and Motivation as two binds coordinated across the descent/ascent of one setpoint traversal (keeps genome/voting as-is).
2. **Subgoals vs. dependencies as children.** Whether every child gets both reads (gate on descent, build on ascent) or some children are pure dependencies (descent-gate only, never compose). I.e. one child-set read twice, or two child-kinds. ("auntie happy" builds *and* gates; "the user wants brainstorming" gates but isn't a part.)

## Proof of Completion — Where Design Becomes Code

A setpoint reaching alignment ~1 is not *proof* it closed — closure must be **observed, not asserted**, which means something has to actively watch for it: a **completion monitor**. That monitor is the comparator running **continuously** (Coherence reading alignment against the `satisfied-when` until the world reads back ~1), not a one-shot judge of a hand-built present.

This is the **design→code boundary** for this whole thread: the monitor is real runtime, and it depends on the kernel port (the comparator on a traversed frame) and the continuous-coherence loop — the same dependency chain as `.1004` ([phase-06](../../../captains-log/phase-06-compaction-vs-coherence.md)). **Deferred** — the design is good enough to pause here; the monitor is the first piece of code this work will demand once the loop exists.

## Motivation Dissolves — Two Methods, Not a Binding Reality

Motivation was never a clean binding reality: a binding reality has *members*; Motivation never did — it was a *process* riding on the history every binding reality already keeps. "Motivation lives in the past, and all binding realities have history" is the whole argument for dissolving it. It splits into **two methods** that read existing structure, plus the perspective overlay:

- **Descent = other.** The Motivation/Observe face from the recursive-primitive section is the **descent**, and under the self/other split ([self-and-other-modeling.md](self-and-other-modeling.md)) descent = **other-modeling**: Motivation reads from outside (the *why* comes from the other), which is why "self relies on other" = "Goal relies on Motivation."

Motivation's two old jobs now have separate homes:

1. **"Did the prerequisite complete?"** → a **transitive reality-monitor.** It walks the dependency chain (transitive closure over previous setpoints) and checks each actually completed — via a **completion index** (`setpoint → link-to-the-artifact`, e.g. the sent email). It is **doubt-driven**: it fires *occasionally*, only when the expected completion-output is **absent from frame** (a prediction-error trigger). Most waves it never fires. This is **source/reality monitoring** (Marcia Johnson) — *"did I do that, or only imagine it?"* — and it is the **necessary complement to simulation**: once the being can run actions in its head, it must check the *external proof*, never the internal memory, or it hallucinates its own completions.
2. **"Should I still care?"** → **Coherence catching a defeated assumption** (next section). The revaluation isn't a separate check; it's Coherence detecting that an assumption underneath the setpoint was contradicted.

**Consequence for root-drive ownership.** With Motivation no longer a binding reality, the earlier "root drive lives in Motivation" placement is moot — the root drive's home defaults to the membrane's `Content` (the being's identity) or a designated root setpoint. Revisit the [Ownership](#ownership) fork accordingly.

## Dependencies Are Assumptions — A Truth-Maintenance System

A dependency is best understood not as a generic precondition but as an **assumption setpoint**: a setpoint held **true by default — defeasibly — until something contradicts it.** Defeasibility is the defining property, not provenance. Three sources, all converging on "earned the right to be assumed instead of checked":

1. **Persistence** — a previously-completed/verified setpoint assumed to still hold ("the venue is booked"). The **frame problem**: things stay as they were unless changed.
2. **Default** — never verified, high prior ("Justin is healthy" — the typical state). **Default logic / non-monotonic reasoning.**
3. **Promoted prediction** — a prediction that's been right enough to stop actively checking. The existing promotion gradient (`comparator.md`: "repeated matches let inference stop firing") extended: **prediction → assumption.**

This is a **Truth-Maintenance System** (Doyle; ATMS, de Kleer) plus **belief revision** (AGM). The link from a setpoint to its assumption-setpoints *is* the justification structure, and **defeating an assumption cascades** to its dependents:

```
assumption "Justin is healthy"  ──justifies──►  "Michael is at Justin's party June 18"
incoming "Justin is sick, party cancelled"
   → defeats the assumption
   → cascade: the parent setpoint is invalidated → abandon
```

- **Not a binding reality — a status stamp.** "Assumed" is one value of an epistemic-status dimension on setpoints (`{assumed, verified, predicted, refuted}`). *Both* categorical and episodic episodes carry assumptions, which is why it's a stamp, not a home.
- **Fires before other-modeling.** Assumptions are the **ground** predictions stand on — the assumed world-state is established before modeling what the other wants. They bottom out at the sensory leaf (a setpoint whose alignment is *assumed* rather than *read*, until read or contradicted).
- **Confidence is the defeasibility line.** "More likely than not" = the assumption's confidence sitting above the build-on-it threshold. Contradicting evidence drops it below → defeated → cascade. (Same confidence machinery as prediction.)
- **Coherence is the monitor.** "Should I still care" = Coherence (the contradiction monitor) detecting an assumption-violation, then the TMS cascade abandons the dependents. This is the home for the revaluation half of old-Motivation, and it is exactly `cell-types.md`'s "conditions gone → abandon."

So old-Motivation's two jobs land cleanly and separately: **completion** → the transitive reality-monitor (past); **validity** → Coherence catching defeated assumptions (present). Neither needs Motivation to be a binding reality.

## Setpoints Must Be Observable AND Falsifiable

Observed-not-asserted gets you **observable** — the satisfied-when is a world-state the being reads back, not one it declares. That is necessary but not sufficient: a setpoint must also be **falsifiable** — specific enough that it could turn out *wrong*. "Michael is talking about the project" is observable but trivially true; it carries no information and produces no error, so the being learns nothing and helps no one. "Michael is stating *[the specific claim]*" can be wrong — so it does work.

The danger is the **vacuity trap** (Goodhart, for predictions): an error-minimizing system drifts toward vague, always-true setpoints *because* they're never wrong. Two rules guard against it:

1. **Form the most specific grounded setpoint** — the actual claim, not the category; maximally falsifiable while still true.
2. **Don't reward vacuity** — a "match" on an unfalsifiable setpoint is *not* a success and must not reinforce. Otherwise the being *learns* to go vague (the path of least error).

The backstop is the **root drive**: a vague setpoint doesn't *help* the user, so even when it reads locally satisfied, "be helpful" reads unsatisfied — the vacuity that's invisible locally surfaces as misalignment at the root. (Popper applied to setpoints: a claim that can't fail is empty.)
