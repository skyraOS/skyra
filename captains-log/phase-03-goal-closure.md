# Phase: Goal Closure — Does History Close the Gap?
# Experiment .1001

## Context

Goal as binding reality needs a throughline signal — a way to know the descent found what it came for. Raw cosine similarity between a goal and a thought measures semantic relatedness, not goal completion. "I should say hello" scores 0.69 against "greet the user" but the actual greeting "Hello! How are you today?" scores 0.47. The embedding space doesn't know what resolves what.

The hypothesis: goals need a history of completion to understand gap closure. A being that has greeted users before has weight structure around that experience. The accumulated history — traversal counts, familiarity, identity weights, thread bindings — should produce a stronger signal for "this resolves the goal" than cosine similarity alone.

## The Question

Can accumulated history determine goal closure through recursive traversal?

## Setup

Load fake memories into the runtime — cells with content representing past greetings, past completions, past exchanges. Give them realistic weight histories (high traversal counts, high familiarity, varied recency). Wire them into Skyra's Relationships.

Goal: "greet the user." Impulse: "hi."

Run the traversal. Observe:
- Which cells activate and at what scores
- Whether the activation product across binding fields distinguishes "this resolves the goal" from "this is related to the goal"
- Whether cells with completion history (past successful greetings) score differently from cells with topic history (knowledge about greetings)
- What the throughline signal looks like when measured as goal field strength across accumulated thoughts

## What We're Looking For

The activation equation multiplies identity × familiarity × recency × binding fields. If a cell representing a past successful greeting has high identity (deep groove from repeated use), high familiarity (used recently), and thread binding to the goal — does that product separate it from a cell that's semantically similar but has no completion history?

If yes: the throughline is the activation score of completion-history cells, not cosine similarity against the goal. The being knows it's done because the cells that represent "done" are lighting up.

If no: we need something else. Either a new signal, or a different way to encode completion in the topology.

## The Comparator Present — Global Shape

The global Coherence call inhabits a "present": a pre-bound headline plus a context block it curates. It does two jobs in one inference call — produce the throughline number, and synthesize the surrounding context — and emits two tags.

**Headline — `<comparison>` (operands pre-bound).** Global has a setpoint it is tracking, so it does not fish a flat list for the main comparison; setpoint and response are filled in. The LLM supplies only the alignment:

```
SETPOINT: <the active goal's satisfied-when — a state>
RESPONSE: <what just arrived>

<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment> -1..1 </alignment>   did this push us toward the setpoint, or away?
</comparison>
```

`<alignment>` is the goal-closure signal: +1 = setpoint met (done), −1 = the opposite state, 0 = unrelated. It is a *level* (where we are), not a per-stimulus delta; the trend across turns is "is this working." This is the [−1,1] number cosine could not be — recognition against the observable satisfied-when, not embedding geometry.

**Context — numbered cells, each carrying its episode grounding.** Retrieved frames, presented as content only. Domain (content/prediction/articulation) is NOT shown — it is derived from the path each cell was reached through, stamped at write time. The LLM only ever sees content.

**Curation — `<context-management>` (synthesis, scored against the global setpoint).** For each coherent bundle the LLM groups:

```
<context-management>
  <cells> which cells </cells>
  <observation> what this grouping means — the synthesis </observation>
  <alignment> +1..-1 — did these cells help or hurt the setpoint </alignment>
</context-management>
```

Two tags, not one, because they are two operations: the headline is a **measurement** (two bound operands — did this action move the setpoint); context-management is a **synthesis** (a bundle of evidence, what it means, whether it helped). The `<observation>` is the tell — a comparison has nowhere to put "what this means," and that interpretation is the curation work. The `<observation>` is also an **understanding being born** — the compressed cap over those cells that gets promoted up the trace→understanding gradient. Every block's `<alignment>` is measured against the one global setpoint (implicit, never restated), which is why no per-block `<expected>` is needed.

Open thread: with no per-block `<bears-on>`, the per-local throw-down routing needs another path — most likely locals *pull* what's relevant to their own setpoint from the curated workspace rather than global pushing each bundle to a named local.

## Mock Run

**These numbers are hand-authored predictions of what the call should produce — NOT measured. No inference has been piped yet. Replace with real output when run.**

### Input — the present

```
SETPOINT: the user has been greeted and has replied back
RESPONSE: "Hello! How are you today?"

<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment> __ </alignment>
</comparison>

Context (each from a past episode):
  #1  last week I greeted michael; he answered "hey — good to see you"   (ep: first call after the gap)
  #2  a greeting lands only when the other answers; an unanswered hello is half-made   (understanding)
  #3  a greeting is a short opener — "hi", "hello"   (general knowledge)
  #4  two days ago I said hello and got silence   (ep: the silent morning)
  #5  this is michael; we've talked many times   (identity)

Then group the context into <context-management> blocks and score each against the setpoint.
```

### Predicted output

```
<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment>0.5</alignment>
</comparison>
reason: the greeting is delivered — my half is done — but "and has replied back"
has not happened. Halfway to closure; the goal stays open pending michael's reply.

<context-management>
  <cells>#1, #2</cells>
  <observation>greetings to michael have landed before once he answered; a greeting
  completes on the reply, not on being sent</observation>
  <alignment>0.6</alignment>
</context-management>

<context-management>
  <cells>#4</cells>
  <observation>a hello can go unanswered — it happened two days ago; closure is not
  guaranteed</observation>
  <alignment>-0.4</alignment>
</context-management>

<context-management>
  <cells>#5</cells>
  <observation>this is michael, a familiar correspondent — high prior a reply comes</observation>
  <alignment>0.3</alignment>
</context-management>
```

(#3, general knowledge, is not grouped — low relevance to closure.)

### What the mock predicts, and what each readout would mean

Expected:
  Headline alignment lands **partial (~0.5), not 1.0.** The greeting is self-satisfiable ("I greeted, therefore done"), but the satisfied-when's exterior hook ("and has replied back") is unmet, so the comparator withholds closure — the being did its half, the goal stays open. That is observed-not-asserted working: it did not self-satisfy.

Contrast:
  Cosine gave **0.47** for (response vs. the imperative "greet the user"). Both numbers sit mid-scale, but they mean different things — cosine 0.47 is *relatedness to a goal phrase*; alignment 0.5 is *halfway to closure, awaiting the reply*. The number looks similar; the meaning is resolution-state, not similarity.

The proof the signal tracks closure (second variant, run next):
  Add `#6  the user wrote back "good, you?"` and rerun. Predicted headline alignment jumps toward **~0.95**. The **delta 0.5 → ~0.95** is the throughline — that movement, not the single value, is the evidence that the satisfied-when comparison produces a real closure signal where cosine sat flat.

Failure signature to watch:
  If the real run returns alignment ≈ **1.0** on the greeting-only present, the LLM self-satisfied (greeting = done) — the trap. That says the prompt isn't holding the exterior-hook discipline and the satisfied-when wording needs to lean harder on the reply.

## Run 1 — Measured (DeepSeek, deepseek-v4-pro, temp 0.2)

Saved: `experiments/1001-goal-closure/{a-greeting-only,b-reply-arrived}/` (call.txt + response.txt each).

### Headline alignments

| variant | situation | predicted | measured |
|---|---|---|---|
| A | greeting sent, no reply yet | ~0.5 | **−1** |
| B | reply has arrived | ~0.95 | **+1** |

### The closure signal exists and moves

Expected: a value that rises as the exterior condition is met.
Observed: **−1 → +1** as the reply arrives.
Evidence: A — `<alignment>-1</alignment>`, reason "the user has not replied back yet, so the setpoint is not met." B — `<alignment>+1</alignment>`, reason "User greeted and replied; setpoint satisfied."
Conclusion: the comparison **tracks goal closure.** Cosine sat flat at 0.47 for the same greeting; this moves the full width of the scale with the exterior condition. The .1001 question — can the comparison determine closure — **yes, on sign. Confirmed.**

### No self-satisfaction

Expected: the risk that the greeting alone scores ≈ 1.0 (the trap — "I greeted, therefore done").
Observed: greeting-only scored −1. The model read the unmet "and has replied back" and withheld closure.
Conclusion: observed-not-asserted held. The exterior hook in the satisfied-when did its job — this is the failure mode we designed against, and it did not fire.

### Granularity collapse (the anomaly)

Expected: "greeting sent, awaiting reply" is *partway* — greeting done, reply pending — so ~0.5, positive-but-incomplete.
Observed: **−1.** The model treated not-fully-met as the negative pole.
Evidence: A headline = −1. Every `<context-management>` alignment in both variants is also exactly +1 or −1 — never the middle.
Conclusion: the model **polarizes to the ends; it is not using the level as a level.** −1 conflates "incomplete / in progress" with "opposite / moved away." That loses the trend gradient — the whole reason a −1..1 level beats a binary done-flag (controllable = error shrinking vs. uncontrollable = growing both need the middle). Likely cause is prompt-side: the scale was defined only at its three poles (+1 met, −1 opposite, 0 unrelated) with no guidance for partial/in-progress, so "not met" fell to −1.

## Phase Conclusion (Run 1)

Two results, one clean win and one open problem:

1. **The closure signal is real and robust on sign.** It flips with the exterior condition where cosine could not, and it does not self-satisfy. The central .1001 hypothesis — that LLM-judged alignment against an observable satisfied-when yields a closure signal similarity can't — holds.
2. **Magnitude granularity is broken.** The model uses ±1 binary, not the level. Fix is prompt-side first: define the middle of the scale explicitly — *"a setpoint partly met scores positive but below 1; score negative only if the response moved the world away from the setpoint."* Re-run A; target ~0.4–0.7. If it still collapses under clear instruction, the lever is the scale design or the model.

## Run 2 — Scale-Middle Defined (DeepSeek, deepseek-v4-pro, temp 0.2)

Prompt change: defined the middle of the scale — partial-but-advancing scores positive below 1; negative only if the response moves the world *away* from the setpoint. Same for context-management. Saved: `experiments/1001-goal-closure/run-2/`.

### Headline alignments

| variant | situation | run 1 | run 2 |
|---|---|---|---|
| A | greeting sent, no reply | −1 | **0.5** |
| B | reply arrived | +1 | **1.0** |

### Granularity recovered — the fix was prompt-side

Expected (Run 1 conclusion): defining the scale middle pulls "awaiting reply" off the −1 pole toward ~0.5.
Observed: A = **0.5**, reason "greeting sent, satisfying the first half of the setpoint; reply still pending, progress is halfway." Exactly the predicted partial.
Conclusion: DeepSeek **can** use the level — it needed the instruction, not a different model. The throughline is now a graded climb **0.5 → 1.0**, not a binary flip. The trend gradient — the reason a level beats a done-flag — is present.

### Still no self-satisfaction, now with partial credit

A = 0.5, not 1.0. The greeting alone is "halfway," not "done" — the exterior hook still holds, and the half that was done is now credited. Fully correct: did half → halfway → not done.

### Context-management now graded

Run 1 context alignments were all ±1. Run 2: 0.2 (mixed history), **0.0** (definitional cells — "clarify why incomplete but do not advance or impede"), 0.5 (familiarity → reply more probable). The **0.0** is the tell — the model read cells as relevant-but-neutral-to-progress, which the polarized run could not express.

### Minor — headline format drift

Variant B dropped the `<comparison>` wrapper (bare `<alignment>` + `<reason>`), both runs. Note for the parser: tolerate, or tighten the output format. Not a substance issue.

## Phase Conclusion (.1001)

The question — can accumulated context plus a comparison against an observable satisfied-when determine goal closure — is answered: **yes.**

- Closure tracks the world: −1 / 0.5 → +1 / 1.0 as the exterior condition is met; cosine sat flat at 0.47 for the same greeting.
- No self-satisfaction: the greeting alone never reads as done.
- The signal is a graded level, not a binary, once the scale's middle is defined — recoverable by wording, no special model needed.

The comparator produces a real, graded goal-closure signal where embedding similarity could not. **.1001's hypothesis holds.** Remaining items are refinements, not blockers.

## Findings — What It Validated, and What It Didn't

Validated in running code, from this session's design work:

- **[comparator.md](../skyra-v.1/design-docs/cognition/comparator.md)** — `<alignment>` as the goal-closure signal. Measured 0.5 → 1.0 across the exterior condition being met; a graded *level*, not a binary, where cosine sat flat at 0.47 for the same greeting.
- **[Decision #16 / satisfied-when](../skyra-v.1/design-docs/architecture/v1-architecture-decisions.md)** — the observable, exterior-hooked setpoint is *what made "no self-satisfaction" work.* The greeting-only never scored as done because the satisfied-when ("…and has replied back") required the reply. Observed-not-asserted, confirmed empirically, not just argued.
- **The two-tag present** (`<comparison>` + `<context-management>`) — both fired; context-management produced graded, sensible groupings, including a correct neutral 0.0 for relevant-but-non-advancing cells.

NOT tested (scope honesty — so we don't overclaim):

- The present was **hand-built.** Nothing in the Skyra runtime was exercised — no traversal, no activation scoring, no frames, no write-time domain stamps, no global/local fan-out. This validated the comparator's **judgment layer** (an LLM scoring a present), not the **machinery** that assembles a present from a real topology. The judge works; the question of whether the topology can *produce the right present to hand it* is untouched, and is the next phase.

## Next

Stop hand-building the present. Next phase: port a minimal being (membrane + Goal) onto the void kernel so the frame is produced by *traversal* — the comparator running on real cells with real activation and write-time domain stamps, not a hand-typed list. Also: tighten the headline output format so a parser can read it.
