# Owned-Specialist Crossing — Concurrent Binding Realities + Gather-Ignite

## Status

**Planning — discuss before implementing.** First consumer is the central **Coherence** cell (the membrane's first child). This spec covers two things that ship together: a kernel change (a second `Content` crossing mode — a *map* of Realities spawned concurrently) and the cell-side **gather-ignite** policy that fires inference once enough specialists have returned. Builds on the [Two Crossing Modes — Peer and Owned](world-kernel-coupling.md#two-crossing-modes--peer-and-owned) design and the [global-workspace](../cognition/comparator.md#this-is-a-global-workspace) account in the comparator.

## Goal

> **Worked-example note (post-dissolution):** the Goal / Motivation / Prospective specialists used throughout this spec are *illustrative*. Those cells dissolved — Goal/Motivation became the setpoint's ascent/descent methods, Prospective folded in. The **crossing + gather-ignite mechanics here remain valid**; only the example's cell names are stale. The live consumers are the two binding realities (**Temporal + Episodic**), and the voting set is **two, not five** (see §6).

Coherence runs global alignment over Goal / Motivation / Prospective. Those specialists each fire their own inference; running them on the in-plane stack would **serialize** the LLM round-trips, which is the latency the whole owned-specialist crossing exists to avoid. So Coherence must fan them out **concurrently** and gather their votes, firing its own (Resolve) inference only once the set is in.

## The Kernel as it stands

`relation.go` implements exactly **one** crossing — the peer being, a single Reality in `Content`, spawned on the ascent (pop):

```go
// traverse(), on pop:
if inner, ok := top.Cell.Content().(skyra.Space); ok && !isReturnImpulse(impulse) {
    go r.realize(inner, inner, impulse)   // peer: origin = self (inner)
}
out := top.Cell.Express(origin, impulse)
```

There is no map case. The owned-specialist crossing does not exist yet. That is the work.

## 1. Kernel Change — `Content` as a map of Realities

Generalize the `Content` type-check to the second case from [world-kernel-coupling.md](world-kernel-coupling.md#two-crossing-modes--peer-and-owned). The difference from the peer case is **who owns the spawned wave**:

```go
switch c := top.Cell.Content().(type) {
case skyra.Space:                       // peer being — detached, its own life
    if !isReturnImpulse(impulse) {
        go r.realize(c, c, impulse)     // origin = self (inner)
    }
case map[string]skyra.Space:            // owned specialists — report back to me
    if !isReturnImpulse(impulse) {
        for _, sub := range c {
            go r.realize(top.Cell, sub, impulse)  // origin = parent
        }
    }
}
```

- **Peer** (`Content` is a `Space`): origin = the inner being. Detached. The parent has no claim on its result. (Unchanged.)
- **Owned** (`Content` is a `map[string]Space`): origin = **the parent cell**. Each specialist runs its own in-plane wave on its own goroutine, and on completion delivers its result to its origin — which is the parent — as an ordinary "from me" perception wave. The kernel **does not wait, tally, or know a vote is being collected.** It spawns and moves on.

The kernel stays opaque: it spawns waves and delivers each to its origin. Fan-out is the kernel's; **gather is the cell's** (§2).

### Return waves still don't cross

The `!isReturnImpulse` guard holds for both cases. A return/`perceive` wave perceiving a result must not re-stimulate the specialists — otherwise the gather-ignite below would re-trigger the fan-out on every vote that comes home. The returning votes are themselves return waves; they deposit, they do not re-cross.

### Amplitude bounds the tree

Spawned waves derive amplitude from the impulse and attenuate across the crossing (already true for the peer case). The map case spawns N waves, not one, so the goroutine-tree bound matters more — but it is the same mechanism, and visited-once + threshold still terminate each inner wave.

## 2. Gather-Ignite — the cell's policy on returning votes

Each specialist's "from me" wave comes home as **reafference** — Coherence *perceives* its specialist's result. By the Perception Invariant that lands via `origin.Observe(origin, result)`, so the return runs a wave through Coherence: **Observe deposits the vote and bumps the counter; the ascent reaches Express, which checks the gather condition and ignites Resolve only when it trips.** Counter in Observe, ignite-gate in Express. What the parent does is **its own policy**, on its own slice, under the slice lock. This is where "fire inference once enough have finished" lives — and it resolves the earlier open question of whether returns hit Observe or Express: **both, in order** (perceive-deposit, then gate-ignite on the ascent).

```
return wave through Coherence (origin == self, a "from me" perception wave):
  Observe:  deposit the vote into the workspace (Coherence's slice); counter++
  Express:  if counter >= x of n  → ignite: fire Resolve over the workspace
            else                  → do nothing (no inference yet)
```

- **Deposit, not block.** "The parent does not wait" stays intact: there is no barrier that stalls Coherence. Votes arrive as events; Coherence accumulates and ignites on the one that trips the count. Each returning wave locks the slice, so deposits serialize cleanly — no race.
- **The gather condition is the cell's, not the kernel's.** Different cells can gather differently (full count, quorum, race, re-fire stragglers) with zero kernel change. Coherence's is a **quorum `x of n`, not `n of n`** — partial, for straggler tolerance (a specialist that never returns must not hang the ignite). The straggler cost is in Open.
- **Ignition is one inference call.** When the count trips, Coherence (Space) fires Resolve over the deposited workspace — the global verdict. This is GWT ignition: parallel processors → competition for access → broadcast → ignite.

## 3. Coherence Wiring

Putting §1 and §2 together for the first consumer:

| Slot | Coherence holds | Why |
|---|---|---|
| **`Content()`** | `map[string]Space{Goal, Motivation, Prospective}` | the owned-specialist crossing — spawn them concurrently, votes return to Coherence |
| **`Children()`** | the top-aligning deposited votes (cut to the top matches) | the scored workspace — what won access this cycle; global coherence keeps only the top returns |
| **`Observe` → Predict** | the ACC comparator: score `<alignment>`, flag the error, set + throw down the global frame | runs first, before the fan-out — global frame primes the locals |
| **`Resolve`** (ignite) | the frontopolar verdict: votes resolve the error → commit; don't → gather (raise `<need>`) or inner-voice re-entry | fires on the gather condition, after votes return — global verdict last |

The descent/ascent bracket gives the phase order for free: **Predict before the specialists** (Observe precedes the Content spawn in every node's lifecycle), **Resolve after them** (the ignite fires on the returning votes). No reliance on sibling order — see §4.

### Children is dynamic, and empty on the first pass

`Coherence.Children()` returns the *returned votes*, which don't exist until specialists report. So on the wave that *starts* the fan-out, Children is empty/stale — there is nothing to score yet. The cell scores its workspace on the **ignite** pass, not the spawning pass. (Mirrors how the world's `Point` recomputes its neighbors in `Observe` rather than holding them statically.) Bootstrapping detail flagged in Open.

## 4. Lifecycle — Two-Speed: Fast Response, Slow Audit

The fan-out and the ignition are **decoupled in time**, and that decoupling *is* the cognitive payoff — the being does not wait for all of itself to finish thinking before it answers. Three events, two clocks:

1. **Fast express — the being responds.** The wave that ran Predict and fired the fan-out does **not** block on the specialists. It finishes its own deliberation with what it has, and the motor acts — the being speaks or handles the request. The reflex path. (`comparator.md`: "A being can act reflexively without a goal.")
2. **Vote return — a specialist finishes.** Each specialist delivers its "from me" wave back to Coherence (§2): Observe deposits + counts, Express gates.
3. **Resolve ignition — the slow audit.** When `x/n` trips, Coherence fires Resolve over the full workspace — **possibly after the being has already spoken.** Resolve judges whether the late thinking was valuable: it confirmed the response → **suppress**; it added or contradicted → **raise**.

So the response and the audit run on different clocks: **speak, then reconsider.** This is System-1/System-2 with neither a special case — the reflex is the fast wave, the reflection is the late Resolve. It is also frontopolar's two outcomes (resolve-the-error → pass/suppress; don't → raise the `<need>`), fired *after* commit rather than before.

- **`x/n` is for ignition, not for responding.** The quorum gates the *audit*; it never gates the *answer*. The being is not blocked on its own quorum.
- **Raise has two targets.** A valuable late thought can become a **spoken** follow-up / self-correction (impose — `comparator.md` mode a) or an **internal** deposit that only updates episodic/weights and stays quiet (inner-voice — mode b). Resolve picks; the likely lever is whether the late thought *contradicts* what was said (lean spoken) or merely *adds* (lean internal). Open.
- **Amplitude bounds the late raises.** A post-commit Resolve runs under the same attenuation that bounds the inner-voice loop — it fades after a few passes. That is the line between reconsidering and ruminating: the audit cannot raise forever.

## 5. Sequencing — what guarantees the order (and what doesn't)

**Sequential within a plane: yes.** The in-plane DFS fully descends and Expresses one survivor before the next sibling — `relation.go`'s explicit stack.

**Left-to-right among siblings: no.** Survivors are built by ranging `node.Cell.Children()`, a Go map — iteration order is randomized by the runtime. There is no sibling-order guarantee, and nothing may depend on one.

**So global-before-locals is NOT enforced by ordering Coherence's neighbors.** It is enforced by the **descent/ascent bracket**, which is structural:

- Coherence's `Observe` (Predict) runs before the kernel processes Coherence's `Content` — Observe always precedes the Content-check in a node's lifecycle. Global frame first, by construction.
- Coherence's Resolve fires on the returning votes (the gather event), which by definition arrive after the specialists ran. Global verdict last, by construction.

This is why a **separate "neighbor cell to hold the binding realities"** is not needed for sequencing: Coherence holding them in its own `Content` already runs Predict-before-spawn. (See Open for the one case that might still want the split.)

## 6. The Voting Set — Brain-Region Grounding

The five voters map 1:1 onto the convergent inputs to the striatum in the cortico-basal-ganglia-thalamocortical loop. The neuroscience independently reproduces the genome's non-buffer bind set: cortex (and limbic regions) propose weighted bids; the basal ganglia tallies and gates (winner-take-all, disinhibition through the thalamus). No central chairman.

| Brain region | What its vote carries | Binding reality |
|---|---|---|
| **dlPFC / vlPFC** | goal representation, cognitive control | **Goal** |
| **OFC / vmPFC** | reward value, expected outcome, "binding stimulus to response" | **Motivation** (value) |
| **Amygdala** | emotional valence / affective salience | **Emotion** |
| **Hippocampus / temporal cortex** | context, episodic memory, "predicting what might come next" | **Episodic** |
| **Motor / premotor cortex** | the action candidates themselves | **Action** |
| **ACC** | conflict / action-value monitoring | **Coherence** |
| **Basal ganglia / striatum** | *tallies and gates* — winner-take-all | the convergence point itself |

**The sharpening — the tally point is the basal ganglia, not the ACC.** In biology the convergence/selection happens at the striatum; the **ACC is one input, a conflict monitor, not the tally point** — it engages only when the votes conflict. Routine selection is coincidence detection: aligned bids disinhibit the winner and fire, no central region involved. This matches `comparator.md` ("aligned votes fire action reflexively — basal-ganglia coincidence detection, no center needed; the center only engages when the regions can't resolve") and refines this spec: the `x/n` gather + Resolve is Coherence's **conflict** job, not its routine one. The routine tally is a basal-ganglia-like / motor convergence we have not yet named.

**~~Open~~ → resolved: the convergence is the membrane's Express.** The basal-ganglia coincidence-and-fire is the **membrane's motor face** (`Express`): above-threshold-and-agree → fire the action, reflexively, no center. Coherence (ACC) engages only on conflict (the tie-break). So there is no separate basal-ganglia cell — it's the sensorimotor boundary. See [skyra-dna/membrane.md](../../specs/2026-06-16/skyra-dna/membrane.md).

### Voting-set reorg — 5 → 2 (pending)

The five-voter table above is the **biological/prior** model. The setpoint-cognition thread reorganized it, and this is **not yet fully reconciled:**

- **The voters are now two binding realities: categorical + episodic** (semantic vs. episodic memory — Tulving). Each generates a self-model with confidence and votes; a clear winner fires, a tie goes to Coherence.
- **Goal and Motivation are no longer voters — they became *methods*:** ascent = Goal = self, descent = Motivation = other ([self-and-other-modeling.md](../cognition/self-and-other-modeling.md), [setpoints.md](../cognition/setpoints.md)). Perspective is the Observe/Express faces, burned at execution.
- **Action is now placed; Emotion is conjectured.** **Action** is the self-pass's substrate in the [error-recovery loop](../cognition/error-recovery-loop.md#this-places-action) — the store of past actions indexed by the assumptions that justified them, which the self-recursion descends. **Emotion** is conjectured as the *fast trigger* (the "this feels off" valence that kicks the recovery loop), not yet settled.

Sources: [MRC BNDU — basal ganglia & cortex optimal decision making](https://www.mrcbndu.ox.ac.uk/papers/basal-ganglia-and-cortex-implement-optimal-decision-making-between-alternative-actions) · [Subcortical contribution to basal ganglia action selection (PMC)](https://pmc.ncbi.nlm.nih.gov/articles/PMC11097989/) · [Interacting cortico-basal-ganglia-thalamocortical loops (Trends in Neurosciences)](https://www.cell.com/trends/neurosciences/fulltext/S0166-2236(25)00192-4) · [Combinatorial inputs to ventral striatum (eNeuro)](https://www.eneuro.org/content/4/6/ENEURO.0392-17.2017) · [Parafascicular & OFC inputs to striatum (PMC)](https://pmc.ncbi.nlm.nih.gov/articles/PMC8029974/) · [OFC, amygdala, reward value, emotion, decision-making (Oxford)](https://academic.oup.com/book/46855/chapter/413761343)

## 7. Generate + Align Atomic; Collapse at Voting; the Composition Boundary

**Generation and alignment are one inference call — a shared trunk, not a disconnected critic.** The alignment must be computed from the *same representation* that generated the prediction. A second call has to re-interpret what the first produced and can drift — it would score against its own reconstruction, not the actual generative state. This is the actor-critic "value head on the shared trunk" pattern, and the comparator already does it: `<comparison>` emits the cells *and* their `<alignment>` in one structured output, never a second pass. One call can still emit *multiple* `<comparison>` blocks — so the atomic unit is "a generation and all its alignments," not "one alignment."

**Predictions stay latent until the vote.** The generate-step is a recurrent latent move — `past assumptions + previous predictions → generate → future assumptions + future predictions (+ alignment)` — and nothing collapses until it is time to act. **The collapse is the vote**: the categorical and episodic binding realities each present their self-model with its **alignment-as-confidence** (on the self side, confidence ≡ alignment — see [self-and-other-modeling.md](../cognition/self-and-other-modeling.md)); a clear winner fires; a tie is broken by Coherence (`x/n` gather → Resolve, §2). Uncertainty is preserved all the way to the collapse.

**Operations are composable prompt-fragments** (`dna.md`: "the system prompt is assembled from the DNA composition"), so the dissolved pieces (Motivation's methods, the assumption-check, the alignment) compose into the frame's call as fragments — data, not code, hence extensible. But composition has a **boundary**, set by one question:

> **Must these operations agree with each other?** → same call (shared trunk — e.g. generate + align).
> **Are they independent?** → separate calls (fan-out — e.g. the categorical vs. episodic vote, the owned-specialist crossing).

Couple tightly where consistency is load-bearing; fan out where it isn't. The brake is the same one §6 and `.1003` already flag: a stuffed mega-call *degrades* (recall fell apart past ~100k; one prompt doing ten things does each worse). So composability does **not** mean "compose everything into the frame" — it means compose the consistency-coupled, fan out the independent.

## Resolved (this session)

- **Returns hit Observe then Express — both, in order.** A "from me" return wave roots at Coherence as origin and runs a wave through it: Observe deposits the vote + bumps the counter, the ascent reaches Express, which gates the ignite. Counter in Observe, ignite-gate in Express (§2).
- **Gather is a quorum `x of n`, not `n of n`.** Partial ignition, for straggler tolerance (§2). The cost it introduces is the straggler question below.

## Pending revision — Space owns firing

A larger decision reached this session supersedes §1's framing and is **not yet written into the body**: the kernel should **not** own the crossing. Wave-firing ownership lands on **Space**, not the **Kernel** — the kernel transports the one wave it's handed (the in-plane DFS) and delivers to origin; it **never inspects `Content`, never decides to cross**. A Space fires a new wave by invoking the kernel through its environment (another Space that holds the kernel ref); the firing Space chooses *whether*, *when*, and *what to send*, and passes the origin (`self` for peer, `parent` for owned). Consequences: the `Content` type-switch in §1 is **retired**; `Content` becomes kernel-blind private cell data; "peer vs owned" becomes which origin the firing Space passes, not a kernel branch; and Coherence fires its fan-out from inside Predict carrying the computed frame (which the kernel could never author). §1 and §3 are to be rewritten around this once the environment-layer pin (below) is settled.

## Open Questions

- **Which Space is Coherence's environment?** The relay it calls to reach the kernel — the **membrane** (an inner-world facilitator, `innerWorld.observe`, binding realities never reach past it) or the **world directly** (flat; every Space holds an env ref to the world; membrane stops facilitating). This decides whether "every level is an environment" is literal or just world-and-being. Blocks the §1/§3 rewrite.
- **Stragglers after ignition.** With `x/n` quorum, the `n−x` late votes arrive *after* Resolve already fired. Second Resolve, silent deposit, or re-open? Needs a "Resolve already fired this episode" guard or an explicit re-fire policy.
- **Raise external vs internal.** When a post-commit Resolve judges the late thinking valuable, does it speak again (impose / mode a) or just deposit and stay quiet (inner-voice / mode b)? Likely lever: contradicts-what-was-said → spoken; merely-adds → internal. Decides whether a late audit can make the being talk again.
- **Bootstrapping the first ignite.** Children is empty until votes land. Leaning: ignite is purely event-driven off the returning votes; the spawning wave does Predict + fan-out and nothing on its own ascent.
- **Separate holder cell — when it earns its place.** Coherence-owns-Content is the default. A separate child holding the specialists is warranted only if Coherence's `Content` gets a second job (e.g. durable self-state). Not the case for the minimal being.
- **Descent-spawn vs ascent-spawn timing.** [world-kernel-coupling.md](world-kernel-coupling.md#when-the-crossing-fires--descent-after-observe) specs firing on **descent, after Observe**; the live kernel fires on the **pop (ascent)**. Under Space-owned firing this becomes *where in the cell's own method the firing call sits* — naturally after Predict's work, i.e. the descent-overlap the doc wants. Folds into the §1 rewrite.
- **Concurrency safety of the workspace deposit.** N specialists deliver to one slice concurrently. The slice lock serializes them; confirm deposit + counter + ignite-decision sit inside one lock acquisition so two near-simultaneous quorum-completing votes can't double-ignite.
