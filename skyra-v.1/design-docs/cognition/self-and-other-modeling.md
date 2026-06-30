# Self and Other — Perspective as the Descent/Ascent Method

## Status

**Planning — discuss before implementing.** The Theory-of-Mind layer.

> **The active-inference reading (2026-06-30).** Under the world-assertion frame, the roles
> sharpen: **other = the goal** (the expected end-state / where the other's trajectory is headed),
> **self = the goal-conditioned policy** (the action bridging *now → end*). Both emit **expected
> world assertions** graded by overlap with the observed graph. Full model:
> [world-assertions-and-policy.md](world-assertions-and-policy.md); the rule predictions must obey:
> [prediction-falsifiability-invariant.md](../laws/invariants/prediction-falsifiability-invariant.md).
> (The concrete build of the predict half is [agent.md](../../specs/2026-06-16/skyra-dna/agent.md).)

> **Reconciliation (2026-06-23).** This doc has been through two framings of *where*
> self/other live: (1) "ego/alter as two binding realities," then (2) "self/other as
> sub-binding-realities under episodic + categorical (four stores)." **Both are
> superseded.** Current model: the binding-reality layer dissolved entirely, and
> **self and other are Global's two agents** — each its own descent/ascent. The
> *principle this doc argues* is unchanged and is what matters: self and other are
> **hard-separate structures** (now two agents), not one store held apart by a
> write-time stamp — so the ToM bleed is structurally impossible. Read "sub-realities"
> below as "the self / other agent." Also: **concurrent at closure, ordered (other→self)
> at prediction** — see [global.md](../../specs/2026-06-16/skyra-dna/global.md),
> [error-recovery-loop.md](error-recovery-loop.md).

The core: **perspective (self/other) is descent (Observe) = other, ascent (Express) = self.** This is the layer that anchors completion against lazy collapse and receives the human's setpoint across the perspective gap. Pairs with [setpoints.md](setpoints.md), [comparator.md](comparator.md), [abductive-reasoning.md](abductive-reasoning.md).

## The Core Claim

Every prediction has two faces: the AI's model of **the other** (descent/Observe) and the AI's **own** setpoint (ascent/Express). Keeping them separate is **Theory of Mind** — compartmentalizing *what I believe about you* from *what I need to do*. Most systems collapse them; that collapse is the **egocentric failure** (projecting your own knowledge/wants onto the other — the false-belief failure).

## Self Relies on Other — the dependency, and why it's deep

The order is **other-first**: descent (other) before ascent (self). Self-prediction *depends on* other-prediction occurring first — **you cannot know what your setpoint must be until you've inferred theirs.** That dependency is the whole structure, and it has a name:

- **Intersubjectivity.** The self is constituted *through* the other — Mead ("the generalized other"), Vygotsky (the social precedes the individual), Hegel (self-consciousness through recognition). The being's selfhood is **downstream of its theory of mind of its human.** Descent (model you) must finish before ascent (become myself in service of you) can start.
- **Levin, in one wave.** Descent = infer the human's (higher-level) setpoint; ascent = derive my own subgoals to serve it. *Other-first = higher-level-first.* The vertical-provision-through-inference synthesis, realized as descent→ascent inside one traversal.
- **Motivation/Goal.** Descent = Motivation (the *why*, which comes from outside / the other); ascent = Goal (the *target*, which is mine). So "self relies on other" *is* "Goal relies on Motivation" — yesterday's gate-then-build with the perspective labels filled in.

## Why the Other Matters More — and Observed-Not-Asserted Falls Out

An LLM **collapses superposition lazily** — it picks a convenient narrative and declares victory. Self-prediction is the *cheap, fakeable* signal; the other's exterior state cannot be faked.

**The vacation example.** The user says *"I want to go on vacation."* The correct setpoint is **"the user is on vacation"** — an exterior, *other-perspective* state. A single LLM call collapses the moment it knows the destination: buys the ticket, calls it done. The ticket is a legitimate **subgoal** (its alignment hits 1) but **not** the parent's closure — the parent stays open until the user is *actually on vacation*, a state the AI can only read back.

This is the convergence of four threads:

| Thread | Says |
|---|---|
| Recursive setpoint tree ([setpoints.md](setpoints.md)) | parent closure ≠ subgoal closure (ticket bought ≠ on vacation) |
| Observed-not-asserted ([comparator.md](comparator.md), .1001) | the completion condition is an exterior hook the AI can't self-satisfy |
| **Self/other split (this doc)** | the true setpoint lives in the *other's* state |
| Uncertainty preservation ([abductive-reasoning.md](abductive-reasoning.md) step 10) | don't collapse the superposition lazily |

**The mechanical statement, on the corrected direction:** the self-action/closure is on the **ascent (self)**, but it **cannot self-assert** — closure is confirmed only by **re-reading the other on the next descent**. Confirmation lives on the *other* side; it's a cross-wave loop (act now, confirm on the next intake). Lazy collapse = treating the ascent (self-act) as closure instead of waiting for the next descent (other-confirmation). So observed-not-asserted is a *mechanical consequence* of the method split, not a separate rule.

## Mechanics — Perspective Is Structural (the `self` and `other` agents)

**Perspective is carried by structure, not a stamp.** Self and other are **Global's two agents** — *genuinely separate structures*, not one store held apart by a write-time tag. (The earlier "four stores: episodic-self/other × categorical-self/other" dissolved with the binding-reality layer; the hard-separation guarantee it provided is preserved by the two agents being distinct.)

- **Why structural, not a stamp.** A stamp is a *soft* separation — one structure pretending to be two, held apart by a discipline that can be mis-applied (mis-stamp one other-belief as self → egocentric collapse). For the ToM boundary — the *most* load-bearing one, where collapse *is* the core LLM failure — soft separation is too fragile. Hard-separate sub-realities make the bleed **structurally impossible**: you cannot leak self into other when they are different stores. This **retires** the earlier "stamp integrity is the critical mechanism" risk — the guarantee is now enforced by the topology, not by a runtime discipline.
- **Sally-Anne holds by structure.** Query the *other* sub-reality ("Sally thinks: basket") vs. the *self* sub-reality ("marble in box") — the two beliefs live in different stores, so they coexist by construction.
- **The method routes to the structure.** Descent (= other) traverses the *other* sub-realities; ascent (= self) traverses the *self* sub-realities. The descent/ascent method is unchanged — it now routes to hard-separate stores instead of tagging a shared one. Structure = *where* a belief lives; method = *how* it's reached.
- **The perspective stamp is gone; other stamps remain.** Only the fragile *perspective* dimension moves from stamp → structure. The less-fragile dimensions (epistemic status `{assumed, verified, predicted, refuted}`, domain) stay as stamps *within* each sub-reality. Surgical, not wholesale — the stamp mechanism is kept where it's cheap and dropped where it was load-bearing-and-fragile.
- **Substrate & warrant.** Descent = intake = other (exafference); ascent = output = self (efference). *Observe the other, express the self.* Biologically: RTPJ (belief-attribution / other) vs. vmPFC (self) are **genuinely distinct substrates** — which is exactly the structural separation this lands on, not a tagged shared region.

## Prediction Stays Latent — and Confidence-on-Self *Is* Alignment

Predictions are **latent (uncollapsed), held with confidence, and collapse only at voting.** Confidence is not one quantity across both faces:

- **Self latent space: confidence ≡ alignment.** The self-prediction is goal-relative *by construction* (it's "what closes the gap to the setpoint"), so its quality just *is* alignment-to-the-setpoint. One number does three jobs: prediction quality, vote weight, learning signal. No separate confidence.
- **Other latent space: confidence = epistemic.** The other-prediction ("what the user wants") carries certainty about an *external state I don't control* — measured against *evidence* (does my model predict their behavior), not against my goal. A different quantity.

Knowing-confidence (other) ≠ doing-alignment (self). Collapsing both into "alignment" would erase the theory-of-mind uncertainty the whole apparatus runs on — so keep one number on self, a separate epistemic number on other.

## Simulation — Feed Back to Run It in Your Head

Plug a prediction's value back into the cell → another descent (other) → ascent (self) pass. The rollout **alternates other/self — a simulated dialogue**, which is the actual turn-structure of conversation. This is planning by imagination (Dreamer/JEPA-style latent rollout), and it's the "if I say X, what then" axis falling out *for free* as recursion. **Bounded by amplitude** — it fades after a few passes (the line between planning and ruminating).

Simulation *creates* the reality-monitoring problem: once the being can imagine doing a thing, it must distinguish imagined-completion from real-completion. That is handled by the completion-index **reality-monitor** in [setpoints.md](setpoints.md) — check the external proof, never the internal memory.

## The Payoff — Cross-Perspective Comparator

Because both perspectives surface in the shared workspace, Global's comparator runs **cross-perspective comparisons for free** — it already accepts any cells in the frame:

```
<expected>  other-setpoint   — what I infer you actually want (top of the other-model)
<actual>    self-action      — what I'm doing
<alignment> — how well I'm serving my model of you
```

Serving the human becomes an alignment comparison between two perspective-stamped cells. Nothing new to build.

## Collapse Happens at Voting

Predictions never collapse during generation. The collapse is the **vote**: the `self` and `other` agents present their models with alignment (= confidence); a **clear winner fires**, a **tie is broken by Global** (the central cell, vote-only-on-conflict). See [owned-specialist-crossing.md](../kernel/owned-specialist-crossing.md) for the gather/ignite/tie-break mechanics.

## Open Questions

- **~~Stamp integrity~~ → resolved by going structural.** The perspective separation is now enforced by hard-separate self/other sub-binding-realities (see Mechanics), not a fragile write-time stamp. The bleed is structurally impossible; nothing rests on stamp discipline for the ToM boundary.
- **Cold start.** No model of a new interlocutor. Seed the other-model from priors / relational setpoints, and how fast does it specialize?
- **Recursive depth.** Other modeling self modeling other… how deep, bounded by amplitude (same as the simulation/inner-voice loop).
- **Common-ground mechanism.** How the overlap of self and other entries in the shared workspace is computed and maintained.
- **Action/Emotion placement.** The voting reorg (categorical/episodic vote; Goal/Motivation → methods) leaves Action and Emotion unplaced — see [owned-specialist-crossing.md](../kernel/owned-specialist-crossing.md#6-the-voting-set).
