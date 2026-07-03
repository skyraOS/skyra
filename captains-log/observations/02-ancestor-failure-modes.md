# Observation: How This Lineage Dies — The Ancestor Failure-Mode Map

**Date:** 2026-06-24 · **Category:** field observation (NOT an activation experiment)

> A due-diligence map, not an experiment. It catalogs how prior cognitive
> architectures (SOAR, ACT-R, OpenCog/CogPrime, Numenta/HTM, Cyc) actually failed,
> abstracts the failure modes, and marks where Skyra stands against each *on paper*.
> Purpose: recognize each failure as a *named face* when it arrives during the build,
> instead of meeting it as a mystery. Reread the day something starts misbehaving.
> No numbers; no next phase. Standing reference.

## Why this is in the log

The hardest risks for this class of system are invisible until the assembled loop is
dynamically live, and they killed beautiful, well-read architectures — *not* because the
ideas were bad, but because the emergent competence didn't show up on assembly, or the
dynamics were an intractable control problem. Skyra is in that lineage. Knowing precisely
*how the ancestors died* is what lets us see the failure modes coming.

## The ancestors, by cause of death

- **SOAR / ACT-R — the content bottleneck, then drift into instruments.** Production-rule /
  activation architectures (ACT-R even has base-level activation *with decay + spreading
  activation*, close to our bump). They work; neither became a mind. (1) The
  **knowledge-acquisition bottleneck**: intelligence lives in *content* that had to be
  hand-authored; their learning (SOAR chunking) only compiled existing problem-solving into
  faster rules, never acquired new concepts from raw experience. (2) **Symbol grounding**:
  tokens with no tie to sensorimotor reality. Outcome: superb executors of encoded
  knowledge that slid into being *research instruments* (ACT-R fits reaction-time data,
  hand-built per task) while the AGI horizon receded.

- **OpenCog / CogPrime — the closest mirror, and the most instructive.** Explicitly aimed at
  AGI: AtomSpace (typed weighted hypergraph), PLN (probabilistic logic), MOSES (evolutionary
  learning), and **ECAN** — attention as spreading *importance* values that decay, with hub
  effects and forgetting. ECAN ≈ our wave + EMA + threshold-eviction by another name. Cause
  of death:
  1. **Synergy never cohered.** Powerful parts; assembled, the whole was not more than the
     parts. Getting them to help each other was harder than building each.
  2. **The attention dynamics were an intractable control problem.** ECAN was *supposed* to
     tame combinatorial explosion by focusing on the right atoms — tuning it to surface the
     right thing at the right time was never solved.
  3. **Perpetual framework that outran demonstrated behavior** — decades of architecture,
     little emergent cognition. Goertzel eventually rewrote it (Hyperon): an admission the
     first cathedral didn't cross the gap.

- **Numenta / HTM — fidelity ≠ function.** Biologically faithful (cortical columns, sparse
  representations, sequence prediction). Kept producing an excellent *time-series anomaly
  detector*; never scaled the hierarchy into general cognition. Getting the local mechanism
  biologically right bought nothing about whether macro cognition appears. Then deep learning
  — uglier, functional — lapped it.

- **Cyc — the anti-pattern.** Decades hand-encoding common-sense logic. Stalled forever on the
  knowledge bottleneck: you cannot hand-author the world, and authored symbols don't ground.

## The failure modes, abstracted

1. **Content / grounding** — the substrate never fills with grounded content (hand-author →
   doesn't scale/ground; learn-from-scratch → concepts never rich enough).
2. **Relevance-by-diffusion** — no cheap, typed, write-time path from context to the relevant
   nodes, so relevance must be *re-derived by diffusion every cycle*: expensive AND an
   uncontrollable tuning problem (ECAN's death).
3. **Two bolted-together systems** — structure (ontology) and dynamics (attention) that don't
   co-evolve; hand-reconciled, never one substrate.
4. **Coherence / synergy** — on assembly, the promised emergent cognition doesn't materialize;
   the whole doesn't exceed the parts.
5. **Tool-drift** — the architecture succeeds at a smaller goal and the mind recedes.
6. **Fidelity ≠ function** — biological/cognitive faithfulness doesn't buy emergent competence.

## Where Skyra stands against each (on paper)

| # | Failure mode | Skyra's standing | Mechanism / the load-bearing reason |
|---|---|---|---|
| 1 | Content / grounding | **Decisively sidestepped** | The **LLM carries world-knowledge and grounded meaning**; the topology only *organizes*. The card no ancestor held. Removes the death that made the others matter. |
| 3 | Two bolted-together systems | **Sidestepped** | **Edge-owned activation**: adjacency and weight are the *same record*. Waves crystallize edges (Hebbian path→edge); edges route waves. One co-evolving substrate, not two hand-reconciled ones. |
| 2 | Relevance-by-diffusion | **Sidestepped on paper; soft spots at scale** | **Indexes-as-heaters**, relevance = *superposition* of cheap heaters (entity anchor + IndexScore overlap + semantic gather). The cheap typed path ECAN lacked. Open seams below. |
| 6 | Fidelity ≠ function | **Mostly sidestepped** | Function is carried by the **LLM**, not by biological fidelity; cognitive theories are *organizing patterns*, not a simulation to be faithful to. |
| 4 | Coherence / synergy | **NOT sidestepped — unresolvable on paper** | But **relocated** into the LLM's judgment (see below). The one only the running loop can settle. |
| 5 | Tool-drift | **Execution risk, not architectural** | Aimed at a being, not a tool. The exposure is *discipline*: design outrunning build. |

## The deepest move: coherence converted from unbounded-static to bounded-improving

OpenCog had to **build** coherence out of hand-tuned logic + attention dynamics — an *open,
static* problem that doesn't improve on its own. Skyra **relocates coherence into the LLM**:
"given well-framed, bounded material, does a frontier model compose it correctly?" That is
**bounded** (≤~100k, one judgment) **and rides the one curve in the field that improves for
free** — every model generation derisks the hardest failure mode without touching the
architecture. No ancestor's fatal problem had that property. This is the strongest on-paper
result in the design.

**Two honest caveats on the relocation:**
- **Presence ≠ framing.** The architecture still owns the *construction* of the judgment call
  (gist, prompt, what's foregrounded). A composition failure with all material present can
  still be *ours* via bad framing — no better model fixes that. The rule is "all material
  present **and well-framed** → model's problem," not "present → not my problem."
- **The relocation has an unproven precondition** — that the architecture *reliably delivers
  well-framed, bounded material every cycle.* That delivery (the closed loop, the cycle not
  going dead or epileptic at scale) is exactly the thing never yet run.

## The open seams — and why they're in the right place

Recognize these by name when behavior degrades:

1. **Cold-context hub saturation** (sharpest, scale-hidden). Branching-divides-amplitude
   stops domination but also kills *within-hub* discrimination; the real discriminator is
   *intersection* of multiple present entities. It fails at the **opening of an interaction**
   (few present entities → nothing to intersect) — compounded because alignment is
   lagged-neutral there too. Looks fine small; rots as hubs accumulate stamps.
2. **Entity-orthogonal relevance** (coverage blindness). Entity-overlap ≠ relevance; the
   anchor is blind to relevance that isn't entity-grounded (abstract principles, absent-entity
   goals, analogical rhymes). *Proposed fix:* a **semantic gather** — read-only similarity
   *gather* (must **vector-jump** to topologically-distant nodes, not edge-walk), LLM cut at
   the apex, **only Express bumps the survivors**. Law: **similarity may gather; only judgment
   may groove** (wide read-in, narrow gated groove — same shape as the membrane). Demotes the
   resonance/embedding work to its correct slot: *one heater among several*, not the selector.
   *Watch:* fuse the apex cut with Global's grouping call (don't pay for a third relevance
   pass); the gather threshold can be loose but not absent.
3. **Frame lock-in** (relevance-realization). Relevance is judged vs the *active* groupings, so
   a relevant *dormant* grouping can be invisible — the system retrieves-relative-to-frame but
   may not *re-frame*. Needs surprise/stress to re-open dormant structure. Mechanism not yet
   clearly specced.

**Why this is a good spot:** each seam is **additive and cheap to test** — an index
add / split / remove is ~minutes and reversible. The design converted relevance from a
*structural* problem ("unbounded mess") into an *incremental* one ("relevant but missing a
dimension → add a heater"). The standing floor under all of it: **entity resolution** —
relevance-correctness ≤ resolution-correctness, so it's not a "later" detail.

## The standing risk

Two questions carry essentially all remaining risk: **(a) does the assembled loop cohere**
(only the running loop can answer — design cannot), and **(b) does the build catch the design
before the design outruns it** (the meta-risk: these conversations are generative enough that
discovering the next elegant recognition can outpace porting the MVP onto the kernel). Most
projects in this graveyard had ten open risks. Skyra has two, and they're the right two.

## Not a next phase

Standing reference, not a hypothesis under test. Cited when the architecture's relevance or
coherence design is questioned, and reread when something starts misbehaving.

## Related

- `observations/01-harness-membrane-leak.md` — the membrane thesis from the negative side
- `conclusions.md` — what's validated (judgment layer) vs. designed (the machinery)
- `design-docs/cognition/{working-memory,entities-and-surroundings,groupings-and-understandings,error-recovery-loop}.md`
- `design-docs/kernel/activation-and-traversal.md` — edge-owned activation (the one-substrate answer to #3)
