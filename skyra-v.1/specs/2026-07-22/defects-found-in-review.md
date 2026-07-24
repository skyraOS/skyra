# Defects Found in Review — a Work Order

**Date:** 2026-07-22
**Status:** ALL IN-SCOPE ITEMS CLOSED 2026-07-22. Section 4 remains deliberately untouched.
**Scope:** `src/skyra/cognitive-agent/skyra-dna-v3` only.
**Verified:** against the working tree on 2026-07-22. `go build ./...`, `go vet ./...`,
and `go test ./...` all pass AS OF THIS WRITING — every defect below is latent, not a
current failure. Line numbers are accurate at time of writing but the tree is under
active edit; each item also names its symbol, so cite the symbol if the line has drifted.

**Reads on:** [Episode Names Two Different Things](episode-names-two-different-things.md),
[Biggest Risks](../2026-07-21/biggest-risks.md)

---

## How to use this

Three tiers, ordered by deadline rather than by size. All three landed on 2026-07-22. Tier
1 must remain green before any run whose numbers are kept — those two defects corrupt
evidence, and evidence collected under them cannot be repaired afterward.

Section 4 is OUT OF SCOPE and must not be touched. Read it before starting.

Every item names its own verification. A fix without its test is not done: two of these
defects exist precisely because a promise in a comment was never asserted anywhere.

---

## Tier 1 — corrupts the measurement

### 1.1 `PassesToClosure` records runs that never closed

**Resolution 2026-07-22: LANDED.** `Metrics.Cycles` remains the attempt denominator,
`Metrics.Closures` counts target-met runs, and `PassesToClosure` receives successful runs
only. `Report` renders `closures=met/attempted`. The regression drives one successful run
and one two-pass structural-fixpoint bail and proves the bail changes the denominator but
not the pass series.

**Where:** `being.go:398`, end of `Run`.

**Defect.** The append is unconditional and runs after `c.Unresolved` has already been
set at 386 (fixpoint bail) or 396 (target not met). One series therefore mixes three
outcomes that mean opposite things:

```text
target met            the number the thesis wants
fixpoint bail         a SMALL number — the model stopped making progress and gave up
6 cycles, unmet       a large number
```

**Why it matters.** The give-up path records FEWER passes than a success. A being that
degrades into quitting early produces a FALLING passes-to-closure curve — the thesis's
success signal — while getting worse. The ablation arm is the one expected to fail more,
so the bias runs in the direction that flatters the comparison being run. This is Risk 12
(aggregate improvement hides local epistemic damage) landing on the primary metric.

**Fix.** Record the outcome alongside the count. Preferred shape:

```text
Metrics.PassesToClosure  keeps ONLY runs where the target was met
Metrics.Closures         met / attempted, so the curve has a denominator
```

Do not simply drop unresolved runs on the floor — a closure RATE is required to read the
curve at all. Three passes across 10/10 closures and three passes across 4/10 are opposite
results, and a passes series alone cannot distinguish them.

`c.Unresolved` is already correct by the time line 398 runs; no new state is needed.

**Verify.** A test that drives one closing run and one run that bails at the fixpoint,
then asserts the bail contributed nothing to `PassesToClosure` and did contribute to the
attempted count.

---

### 1.2 `Mandatory` does not do what its comment promises

**Resolution 2026-07-22: LANDED.** `SelectMemoryDetailed` copies and ranks the candidate
slice, admits every mandatory node unconditionally, then fills optional memory from the
remaining budget. Its result exposes budget, total spend, mandatory spend, and overrun.
`Being.LastMemorySelection` retains the latest stats, run reset clears them, and the planner
present prints `mandatory-overrun=N tokens` when the current occurrence exceeds budget.
The regression covers an oversized mandatory node, a zero budget, caller-order
preservation, the live `ActiveMemory` path, and rendered overrun visibility.

**Where:** `memory.go:173-198`, `SelectMemory`.

**Defect.** The doc comment states:

> Mandatory root nodes are **admitted first**; the remainder competes by activation/token
> density.

The code delivers sort ORDER, not admission. Mandatory nodes sort ahead (181-182) and are
then subjected to the same budget test as everything else (196):

```go
if spent+cost > budget {
    continue
}
```

So when the current span exceeds the budget, part of it is silently dropped from the
present. `Mandatory` is set at 239 for exactly the current span — the occurrence the being
is reasoning about right now.

**Why it matters.** The failure is invisible. The being simply does not see part of its
own current occurrence, plans worse for reasons nothing records, and the damage surfaces
as degraded passes-to-closure with no attributable cause. During a measured run that is a
confound; during a conversational prototype it is a bug that will be misdiagnosed as
retrieval.

At the default `MemoryTokens: 2048` this likely does not fire today. It fires the moment
a single occurrence gets large.

**Fix.** Admit every mandatory node unconditionally, deduct their cost, then greedily fill
the remainder from what is left (possibly nothing). That makes the comment true.

Optional and worth considering: when mandatory alone exceeds the budget, the present is
over-budget by construction. That is a real condition and should be observable — a counter
or a field — rather than silently tolerated. Do not solve it by resuming truncation.

**Verify.** A test with a mandatory node larger than the whole budget, asserting it is
present in the output. There is currently NO test mentioning `Mandatory` anywhere in the
package — `grep -rn Mandatory *_test.go` returns nothing. That absence is why the defect
survived.

---

## Tier 2 — correctness, no deadline

### 2.1 `copy` shadows the builtin

**Resolution 2026-07-22: LANDED.** The local is now `settled`; behavior is unchanged.

**Where:** `prediction_calibration.go:146`, inside `resolveCalibrationCategory`.

```go
copy := *dimension
found = &copy
```

Legal and correct today. It shadows the builtin `copy` for the rest of the scope, so a
later edit that adds a genuine `copy(dst, src)` in this function fails in a way that reads
as a type error rather than a shadowing error. Rename to `settled` or `match`.

**Verify.** Existing tests cover the path; no new test required.

---

### 2.2 Zero coverage is not the same as insufficient support

**Resolution 2026-07-22: LANDED.** `BaseRateCalibration` now carries an explicit state:
`INSUFFICIENT_SUPPORT`, `MEASURED_INERT`, `MEASURED_ANTI_DIAGNOSTIC`, or `CALIBRATED`.
Counts and rates survive in every state. Inert (`0/0`) and zero-coverage anti-diagnostic
results retain finite zero PMI; finite negative/zero PMI is also anti-diagnostic and kept
as a prune signal. Only `CALIBRATED` derives `Ready()` and only finite positive PMI can
enter same-claim fusion. The executable fixture covers all four states plus finite negative
PMI and asserts that no `NaN` or infinity enters storage or projection.

**Where:** `prediction_calibration.go:199-203`, `measureDimensionCalibration`.

The guard collapses two different conclusions into one not-`Ready`:

```go
if calibration.CategorySamples < window.MinimumCategorySamples ||
   calibration.NonCategorySamples < window.MinimumNonCategorySamples ||
   calibration.Coverage <= 0 || calibration.BaseRate <= 0 {
```

Failing a support floor means WE DO NOT KNOW YET. Coverage of exactly zero means MEASURED,
AND MAXIMALLY ANTI-DIAGNOSTIC — the predicate fired for no category case at all, on
adequate support. That is one of the most informative results the window can produce and
the cleanest possible prune signal, and it is currently indistinguishable from an
unmeasured predicate.

Note `Ready` is deliberately not promotion — the existing comment on `BaseRateCalibration`
is right about that and should stay. This is about not discarding a measured negative.

**Fix.** Distinguish the two states. Adding a field is acceptable; overloading `Ready` is
not.

**Verify.** A window with adequate support where a dimension never fires for a category
case, asserting the result is distinguishable from an under-supported one.

---

## Tier 3 — documentation and naming

### 3.1 The held-out guarantee rests on an assumption nothing states

**Resolution 2026-07-22: LANDED.** The code beside `selectionEvidence` now states that the
overlap rail covers declared attributions only, that undeclared model exposure is
undetectable from the graph, and that true blindness requires withholding the window
outside model-visible context. Behavior is intentionally unchanged.

**Where:** `prediction_calibration.go:51-54`, `CalibratePredictionField`.

`selectionEvidence` is built from `field.Attributions` — the evidence the model DECLARED
it used when authoring the predicate. The overlap check at 75 is therefore only as honest
as that declaration. A predicate actually chosen by reading observations never declared as
attributions overlaps the window in fact, and the check passes.

This is not fixable in code: undeclared influence is undetectable by construction. It is
the residual hole in the file's central claim, and Risk 6's stop condition depends on it.

**Fix.** A comment at the `selectionEvidence` construction stating the assumption
explicitly — held-out-ness is enforced against DECLARED attributions, and undeclared
influence is out of reach. No behavior change.

---

### 3.2 `Surprise` is pointwise mutual information

**Resolution 2026-07-22: LANDED.** `BaseRateCalibration.Surprise` is now `PMI`, and the
same-claim projection is named `FusePositiveFirePMI`. Documentation calls
`Coverage × PMI` the conditional positive-fire objective, not full mutual information.

**Where:** `BaseRateCalibration.Surprise` in `prediction_field.go`, computed at
`prediction_calibration.go:204`.

```text
log(Coverage / BaseRate) = log[ p(fires | category) / p(fires) ]
```

That is PMI. `ExpectedCredit = Coverage × PMI` is then one term of the mutual-information
sum. "Surprise" conventionally means `-log p`, so the current name points at the wrong
quantity.

Renaming to `PMI` costs one identifier and buys a direct connection to standard
feature-selection literature — free related-work positioning for a mechanism that
currently has none.

**Fix.** Rename the field and update the doc comment to name the quantity. Mechanical.

---

### 3.3 The risk register has gone stale

**Resolution 2026-07-22: LANDED.** Risk 1 now records the independent interpretation and
consequence paths, Toy #9's four outcomes, and the closed two-bit gate while retaining the
open routing/decomposition risks. Risk 2 moves independent interpretation settlement into
existing protection. Risk 6 now distinguishes Phase-3's blocked same-evidence fitter from
Phase-4F's caller-declared held-out calibration and records the remaining blindness limit.

**Where:** `specs/2026-07-21/biggest-risks.md`, Risk 1, "Missing protection".

It states that v3 lacks an independently settled interpretation claim licensing a
category-discriminating prediction. That is what the Phase 4 work implements and
`TestToy9FullTwoBitOutcomeMatrix` covers.

The register's own governing distinction is `implemented rail / specified rail / open
problem`, and it is the release gate for every phase. A gate document that lags the code
silently stops being a gate.

**Fix.** Move the closed items from "Missing protection" to "Existing protection" and note
the gate's status. Going forward, gate closures belong in the same commit as the code that
closes them.

---

## 4. OUT OF SCOPE — do not touch

These are deliberate, architectural, or already have their own spec. Changing any of them
from this document would be wrong even though each looks like a defect.

**`Triple.Key` performs no normalization.** `file:/src`, `file:/src/`, and `File:/src` are
three distinct keys, so a requirement and an expect spelled differently never join and
closure silently fails through the fixpoint bail. This is bounded on purpose by the
world-given / model-assigned split at `triple.go:98` — `file:`, `proc:`, `url:`, `coord:`
subjects are addressed BY THE WORLD, so their spelling is not the model's opinion. Adding
normalization is an architectural decision about identity, not a bug fix. Leave it.

**The `EpisodeID` → `RunID` and `Class` → `Episode` renames.** Specified in
[Episode Names Two Different Things](episode-names-two-different-things.md), with an
ordering hazard (`Episode` is only free after `EpisodeID` moves out) and a trap
(`classifySource` is a DIFFERENT WORD FAMILY and must not be swept). Do not attempt from
this document.

**Span fragmentation / the dead merge layer.** `Classes()` has no live callers and every
impulse mints its own span. Same spec, which carries the design rule that keeps the fix
safe: grouping may confer attention, never settlement rights. Not a code fix.

**Orphaned packages.** `src/skyra/spacetime/traverse`, `src/tui`, `src/debugview`,
`src/skyra/user-agent/user-dna`, and `src/placegateway` have zero live importers after the dnav2
archive. Whether they follow dnav2 into `archive/` is a human decision, not a cleanup.

**`wave.go` and `segment.go`.** Prototype history, retained deliberately. Not on the live
ingest path. Do not modernize, do not delete.

---

## Codex response — comparison against the implementation line

**Date:** 2026-07-22  
**Scope:** review only; no fixes applied by this response.

The review is strong. The two Tier-1 findings are real measurement defects, the shallow
correctness and naming findings are substantially right, and none of them overturns the
two-axis or prediction-field architecture. Two conclusions need refinement, and the
episode note leaves one operational recall asymmetry unstated.

### Tier 1 — confirmed

#### `PassesToClosure` mixes success and failure

Confirmed. The unconditional append occurs after both the fixpoint-bail path and the final
target check have set `c.Unresolved`. A failing system can therefore improve the reported
curve by giving up early. That is not ordinary metric noise; it is bias in the direction
that can flatter the ablation comparison.

The existing `Metrics.Cycles` already increments once per `Run`, so it is close to an
attempt denominator. The minimum honest measurement shape is:

```text
Cycles / ClosureAttempts     every attempted Run
Closures                     target-met Runs
PassesToClosure              target-met Runs only
```

Whether the denominator keeps the name `Cycles` or becomes `ClosureAttempts` is secondary.
The invariant is that a passes distribution is never reported without its closure rate.

#### `Mandatory` memory is only sorted first

Confirmed. Sorting mandatory nodes ahead of optional nodes is not admission. They still
pass through the ordinary budget rejection, and `budget <= 0` returns before inspecting
them at all. The current implementation can silently remove part of the occurrence the
being is reasoning about.

The proposed correction is right: admit all mandatory nodes first, report any mandatory
overrun, then spend whatever optional budget remains. An over-budget present is preferable
to silently falsifying the promise that the current occurrence is present.

### Tier 2 — one confirmation and one refinement

#### `copy` shadowing

Confirmed and cosmetic. `settled` or `match` is clearer and avoids a future misleading
type error. It does not affect current behavior.

#### Zero coverage needs more than a second boolean

The central finding is correct: insufficient support and adequately measured zero coverage
must not collapse into the same not-`Ready` result. But zero coverage itself has two
different meanings:

```text
coverage = 0, base rate > 0
    predicate fires outside the category but never inside it
    → measured anti-diagnostic; strong prune evidence

coverage = 0, base rate = 0
    predicate never fires anywhere
    → measured inert; useless, but not evidence against the category
```

Setting `Ready=true` with `PMI = log(0/base)` would also introduce negative infinity into a
path that currently requires finite calibration. A status enum is cleaner than overloading
`Ready`:

```text
INSUFFICIENT_SUPPORT
MEASURED_INERT
MEASURED_ANTI_DIAGNOSTIC
CALIBRATED
```

The counted support and fire totals should survive in every state. `Ready` can remain the
narrow statement that a finite positive-fire calibration is available; it must not be used
to erase measured negative or inert results.

### Tier 3 — confirmed, with a mathematical qualification

#### Held-out-ness is only enforceable against declared influence

Confirmed. The overlap rail sees `field.Attributions`, so it proves only:

> no observation the proposer declared as selection evidence is reused in the named
> calibration window.

It cannot prove that the model did not read an unattributed observation. That assumption is
already stated in the Phase-4F research notes, but it belongs beside `selectionEvidence` in
code as well. Truly blind evaluation ultimately requires allocation outside the model's
visible context, not a more elaborate overlap check after authorship.

#### `Surprise` is PMI

Confirmed:

```text
log(P(fire | category) / P(fire))
```

is pointwise mutual information between the category event and predicate firing. `PMI` is a
better field name than `Surprise`.

One qualification: `Coverage × PMI` is the positive-fire contribution **per category
instance** used by the current predicate-selection objective. It is not by itself a full
mutual-information term over the global joint distribution; that would also include
category prevalence and the remaining fire/non-fire outcomes. The rename is still right,
but the documentation should not claim that the existing positive-fire objective computes
full mutual information.

#### The risk register is stale

Confirmed. Toy #9 and the Phase-4 claim path now provide independently owned consequence
and interpretation predictions, distinct settlement addresses, and separate histories.
Risk 1 should move those items into implemented protection. Remaining decomposition,
routing-influence, and context-predicate risks should stay open rather than being erased by
the gate closure.

### Response to `Episode Names Two Different Things`

`EpisodeID → RunID` is strongly justified. The implemented boundary swaps the Realm,
replaces the observation graph, span index, and activation field, resets occurrence-local
identity counters, and preserves the expression graph. That is a run/world lifecycle
boundary, not the derived similarity region described by the earlier episode notes.

`Class → Episode` is internally consistent with the project's established ontology, though
outside Skyra the word *episode* usually suggests a contiguous trajectory. The important
part is not the noun; it is preserving the two axes:

```text
span       witnessed temporal occurrence and settlement owner
episode    derived grouping of similar spans for attention
```

The dead-composition finding is also confirmed. `ObserveImpulse` creates a fresh span for
every impulse. `Classes()` and `Lens()` have no live production callers; they remain in the
test/prototype path. The rule proposed by the note is load-bearing and should survive any
rename or later grouping implementation:

> Grouping may confer attention, never settlement rights.

A similarity grouping must not rewrite span ownership or make its observations eligible to
settle another span's bets.

#### Missing from the episode note: cross-run recall is asymmetric

Renaming the lifecycle boundary does not prove that durable expression memory is reachable
through every live inference path after the reset.

```text
explicit Run(target)
    activateQuery can retrieve matching precedent directly from Expr
    → old expression birth spans can be heated across runs

ordinary ObserveImpulse
    retrieval starts from the current run's SpanIndex observation postings
    NewRun/NewEpisode replaced that index and observation log
    → old expression birth spans generally cannot be reached from their original features
```

So the repeated explicit-target experiment may still benefit from durable `Expr`, while a
conversational impulse entering a new run can remain cold despite that durable memory. This
is the earlier cross-run interpretation-recall risk in a more precise form: the learned
expression survives, but the observation postings that normally recruit it do not.

Derived episode grouping does not automatically solve this either. A cross-run grouping
layer needs durable/rebuildable features capable of retrieving old spans, while still
obeying the rule that retrieval attention grants no settlement rights.

### Recommended ordering

```text
1. repair closure measurement before retaining experimental runs
2. make mandatory memory actually mandatory and expose overruns
3. distinguish insufficient, inert, anti-diagnostic, and calibrated outcomes
4. rename Surprise → PMI; add the declared-attribution assumption in code
5. refresh the risk register in the same change that records the closed gates
6. perform RunID naming separately from any episode-grouping implementation
7. treat cross-run conversational recall as its own retrieval design problem
```

The shallow defects should be fixed. The out-of-scope boundaries in this work order should
remain intact. Most importantly, none of these findings licenses merging spans at write
time or weakening the independent settlement addresses established by the two-axis work.

---

## Verification of the Codex response

**Date:** 2026-07-22. All four checkable claims verified against the tree; three
accepted, and the cross-run finding confirmed empirically rather than by reading.

**`Metrics.Cycles` is already an attempt counter.** `being.go:369`, once per `Run`.
Accepted — item 1.1 needs less new state than the work order proposed.

**Zero coverage: the enum is right, and the guard had a second job.** `PMI =
log(0/BaseRate)` is `-Inf`, and `log(0/0)` is `NaN`. So `Coverage <= 0 || BaseRate <= 0`
was also PROTECTING THE ARITHMETIC, not merely conflating two states — the work order
credited it with only the conflation. The four-state enum is the correct fix; a second
boolean is not, because it would not separate inert from anti-diagnostic and would invite
exactly the `-Inf` the guard prevents.

**`Coverage × PMI` is not a full mutual-information term.** Accepted, and the work order
was loose here. The true term weights by the JOINT `p(fire, category) = p(category) ·
p(fire | category)`; the current objective weights by the CONDITIONAL and so omits
category prevalence. The rename to `PMI` stands. The documentation must not claim the
positive-fire objective computes mutual information.

### Cross-run recall asymmetry — CONFIRMED by probe

The most consequential item, and absent from both the work order and the episode note. A
temporary probe placed one expression in run 1, reset with `NewEpisode`, then re-entered
the identical triple by each path:

```text
PRECEDENT REACHED — same run, impulse only:    1
PRECEDENT REACHED — new run, impulse only:     0     ← durable memory unreachable
PRECEDENT REACHED — new run, explicit query:   1
```

The mechanism is exactly as described. `NewEpisode` replaces `Obs`, `Spans`, and
`Activation` and keeps `Expr`. The impulse path retrieves ONLY through `b.Spans.{Entity,
Address, Triple}` posting lists, which the reset emptied. The query path reaches
`b.Expr.Precedent(q, b.tick)` directly from the target triples and is therefore
reset-immune.

Two consequences, and they point in opposite directions:

```text
THE PLANNED EXPERIMENT IS INTACT
    cmd/dnav3 drives Run(target) with an explicit target, so precedent is reached and
    the measured curve is real. 1, not 0.

THE CLAIM IS NARROWER THAN IT SOUNDS
    the being learns across runs WHEN HANDED AN EXPLICIT TARGET. it does not yet
    recruit durable memory from ordinary perception. a conversational impulse in a
    new run starts cold no matter how much Expr has accumulated.
```

This is not a defect to fix from this work order. It is a retrieval design problem, and
the fix a derived grouping layer would need — durable or rebuildable features that can
retrieve old spans — must still obey the rule that grouping confers attention and never
settlement rights.

**Recorded scope note:** the probe was temporary and has been removed. Re-running it is
four minutes of work and belongs in the package permanently once the retrieval question
is taken up.

---

## Not found

Stated so a reviewer knows the search happened rather than assuming it was skipped.

No unsound grading path, no LAW 1 violation, no place where model output reaches world
state without an observation, no shared state between the interpretation and consequence
claim paths, no nondeterminism in selection or comparison. `-race` is clean. Calibration
arithmetic was recomputed by hand against its fixture and every asserted value is correct.

The defects above are shallow. The architecture is not implicated in any of them.
