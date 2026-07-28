# Simplifications Found in Review — a Work Order

**Date:** 2026-07-22
**Status:** OPEN. None of these are applied.
**Scope:** `src/skyra/cognitive-agent/skyra-dna-v3` only.
**Verified:** against the working tree on 2026-07-22. `go build ./...`, `go vet ./...`,
and `go test ./...` all pass AS OF THIS WRITING. Nothing below is a current failure —
this document is about COST and DIVERGENCE RISK, not about wrongness. Line numbers are
accurate at time of writing but the tree is under active edit; each item also names its
symbol, so cite the symbol if the line has drifted.

**Reads on:** [Defects Found in Review](../2026-07-22/defects-found-in-review.md) — the correctness
pass over the same package on the same day. This document is its complement and does not
restate anything it found. Section 5 records where the two overlap.

---

## How to use this

Quality only. No item here changes behavior, and each one should land with the test suite
unchanged and still green. If a change in this document requires editing a test's
expectations, it has stopped being a simplification and needs its own argument.

Four tiers, ordered by payoff:

```text
Tier 1   measured cost           one item, 7.3x, benchmarked
Tier 2   one rule, two bodies    the divergence risk — worth more than its line count
Tier 3   dead and redundant      lines that do nothing
Tier 4   stdlib and boilerplate  mechanical
```

Section 5 is ALREADY ADJUDICATED and must not be acted on from this document.
Section 6 records what was searched for and not found.

Tier 2 is the tier that matters. Every item in it is one rule implemented twice, which is
the shape that survives review and then diverges quietly six months later. Line count is
the wrong way to rank it.

---

## Tier 1 — measured cost

### 1.1 The activation path SHA-256s every act payload two to three times per expect

**Where:** `activation_field.go:306-315` (`propagateActivation`) and `:386-390`
(`deriveActivatedTarget`), against `ExpressionGraph.Edge` at `store.go:431-443`.

**Cost.** `propagateActivation` computes the key it needs and then throws it away:

```go
am := actMap(e)                          // 306  builds the act map
ek := edgeKey(am, x.Of, x.Key())         // 309  hashes every act payload
...
ed := b.Expr.Edge(e.ID, x)               // 315  builds actMap AGAIN, hashes AGAIN
```

`Edge(exprID, x)` takes only the expression ID, so it must reconstruct `actMap(e)` and
recompute `edgeKey` internally — under the mutex — even when the caller is holding both.
`deriveActivatedTarget` is worse: it computes `edgeKey`, then `edgeSourceKey`, then calls
`Edge`, which computes both a third time.

This is not free because `edgeSourceKey` (`store.go:638`) SHA-256s each act's payload,
and payloads are FILE BODIES. The hash is the right identity — that is the fix from the
first live run, recorded at length in `edgeKey`'s comment, and it is not in question. What
is in question is computing it three times for one lookup.

**Measured.** 40 expressions x 5 expects x 8 KB payloads, `propagateActivation` as written
versus an identical body using a key-taking accessor and one memoized `edgeSourceKey` per
`Of`-set:

```text
BenchmarkPropagate         2,197,011 ns/op   3,458,783 B/op   3,275 allocs/op
BenchmarkPropagateKeyed      299,236 ns/op     411,687 B/op   1,104 allocs/op
                                 7.3x               8.4x
```

The benchmark was written to a scratch file, run, and reverted; it is not in the tree. Any
fix should land with it restored as a permanent benchmark.

**Why it matters.** `propagateActivation` runs on every impulse, every settle, and every
observation write — `activateObservation` calls it directly (`:188`). Its cost scales with
the size of the act payloads the being has ever written, not with the size of the region
being activated. That is the wrong scaling law for the component that decides what the
being is currently paying attention to, and it will bite exactly when the being has done
enough work to be interesting.

**Fix.** Add a key-taking accessor beside the existing one:

```text
Edge(exprID string, x Expect) *Edge      unchanged — keep it, grade() has no key
edgeByKey(ek string, of []string, target string) *Edge
```

Then in both activation sites, compute `edgeSourceKey` once per distinct `Of`-set per
expression and build the target keys from it. Do not change `edgeKey`'s definition, do not
cache across calls, and do not memoize on the `Expression` — a payload is mutable until
`Place` and a stored hash would be a new class of bug for a purely mechanical win.

**Verify.** The benchmark above, restored. Plus an assertion that `edgeByKey` and
`Edge` return the SAME POINTER for the same expression and expect — the edge is the
memory, and two accessors that mint two edges for one claim is the exact failure
`edgeKey`'s comment was written about.

### 1.2 `heatPosting` copies a span's whole observation slice to filter it

**Where:** `activation_field.go:252`, inside the posting loop.

**Cost.** `b.Spans.Observations(span)` (`memory.go:143`) allocates and copies the span's
entire observation slice, and the caller immediately filters it down with `matches`. This
sits inside `propagateActivation`'s doubly-nested loop across hops, so the copy is paid
per span per expect per hop.

**Fix.** A filtered accessor on `SpanIndex` that applies the predicate while holding the
lock and returns only matches. The copy exists to avoid handing out the index's backing
array; filtering under the lock preserves that without the full-slice allocation.

**Verify.** Covered by the same benchmark. This is second-order next to 1.1 — do it in the
same pass or not at all.

---

## Tier 2 — one rule, two bodies

### 2.1 Prediction-ID namespacing is implemented twice, with different dedup schemes

**Where:** `prepareExpressionPredictionFields` (`prediction_field.go:171-197`) and
`ExpressionGraph.Place` (`store.go:354-367`).

**Defect.** Both namespace local prediction IDs under the expression ID, and both
de-duplicate collisions — by different rules:

```text
prepare   suffix is #<count of prior occurrences>    p1, p1#1, p1#2
Place     suffix is #<expect index>                  p1, p1#3, p1#7
```

In the `Run` path `prepare` runs first (`being.go:465`) and `Place`'s branch degrades to a
guarded no-op: the `HasPrefix` test at `store.go:358` sees the prefix `prepare` already
applied and skips, and the `seenPredictionIDs` map sees no collision because `prepare`
already broke them. So the second implementation is live only for callers that reach
`Place` directly — where it assigns DIFFERENT IDS FOR THE SAME EMISSION than the main path
would.

**Why it matters.** `PredictionID` is a stable identity by construction —
`prediction_field.go:11` says so, and the reason given is that attributions point at names
rather than slice positions so that ordering cannot silently redirect evidence. Two
namespacing bodies is that guarantee resting on which door the caller came through. Today
the fallback is unreachable from `Run` and harmless; the moment a second entry point
places an expression, evidence attribution follows the other scheme.

**Fix.** One owner. `Place` is the better home — it is the placer, it already owns
`exprs`, `bets`, and the reverse index, and it cannot be bypassed. Move the namespacing
and dedup there whole, and reduce `prepare` to what only it can do: validating attributions
against the birth boundary. If `prepare` must keep assigning IDs because attribution
validation needs them before placement, then it should assign them and `Place` should
ASSERT they are already namespaced rather than silently re-deriving them.

**Verify.** A test that places an expression with two expects carrying the same authored
`id` through `Run`, and a second that places the identical expression through `Place`
directly, then asserts both produce the same prediction IDs. That test does not exist
today, which is why the divergence is invisible.

### 2.2 `grade` and `gradeInterpretations` share a settlement tail

**Where:** `being.go:200-260` and `being.go:265-298`.

**Defect.** The two functions duplicate the same block: dedup against `ResolvedBy`,
compare, fetch the edge, downgrade `Mode` when observation fidelity is lower, grade,
observe, stamp resolution, append `ResolvedBy`, bump metrics.

**Why it matters — and the part to get right.** THE SEPARATION OF THE TWO PATHS IS
CORRECT AND LOAD-BEARING. It is the lucky-hit correction
([The Two-Bit Outcome](../2026-07-21/the-two-bit-outcome.md)), it is why
`ExpressionGraph` keeps `interpretationEdges` disjoint from `edges`, and it is why
`Metrics` carries a second set of counters. Nothing in this item proposes merging the
paths, the indexes, the edges, or the metrics.

What is duplicated is the SETTLEMENT ARITHMETIC, which is identical by intent — both
paths deliberately share one `Edge` probability primitive, which is the design recorded in
[Toy #9](../2026-07-21/toy-9-two-address-claim-paths.md): "two typed claim paths and
histories sharing one `Edge` probability primitive — not two duplicated edge
implementations." Right now the primitive is shared but its CALLING SEQUENCE is copied,
which is half the rule.

**Fix.** Extract the tail as a helper taking the expect pointer, the edge, the observation,
and a pointer to the metric triple to bump. The eligibility checks stay where they are —
they are genuinely different, and that difference is the whole point:

```text
grade                  span activation, episode match, ArmedAt
gradeInterpretations   ArmedAfter sequence boundary, episode match
```

Extract only below the `truth == TruthUnknown` bail. Do not unify the guards.

**Verify.** Existing coverage is adequate — both paths are exercised. The check is that
the interpretation metrics still move independently of the consequence metrics after the
extraction, which the current tests already assert.

### 2.3 `RecordSettlement` builds two identical three-level maps by hand

**Where:** `store.go:448-478`.

**Defect.** `settlementSpans` and `settlementSourceSpans` are the same type,
`map[string]map[Truth]map[SpanID]struct{}`, and the function nil-inits both by hand — 18
lines of stanza, the second one with an extra outer nil check the first does not have.

**Why it matters.** The two maps answer genuinely different questions and must stay
separate — `store.go:493` is explicit that density features must read the prediction-time
side and never the resolving side. That distinction is worth protecting, and it is easier
to protect when the mechanical part is not obscuring it.

**Fix.** One unexported helper, `recordSpan(dst, ek, outcome, span)`, called twice. The
comment explaining why there are two maps then sits directly above two one-line calls
instead of two paragraphs of map construction.

**Verify.** Existing settlement tests cover both maps.

### 2.4 The range verdict is written twice inside one switch

**Where:** `comparePredictionDimension`, `prediction_field.go:537-546` (`ParserTokenCount`)
and `:559-569` (`ParserParagraphCount`).

**Defect.** Twelve identical lines — below / above / in-range, with the same `Residual`
arithmetic — in two arms of the same switch.

**Fix.** A small helper taking the actual number and the `NumberRange`, returning
`(Truth, relation string, residual float64)`. The two arms keep their own evidence
gathering, which is where they legitimately differ.

**Verify.** Existing Phase-4F parser tests.

### 2.5 Differentiation config defaulting happens in three places

**Where:** `NewActivationField` (`activation_field.go:73-85`), `classifySource`
(`:433-439`), `differentiateSource` (`density_pockets.go:47-49`).

**Defect.** Three sites re-derive defaults from `DefaultActivationConfig().Differentiation`,
and they do not agree on how much they cover: the constructor defaults all four fields,
`classifySource` defaults two, `differentiateSource` defaults one.

**Why it matters.** A caller that constructs a `DifferentiationConfig` literal gets a
different set of defaults depending on which function reads it first. The thresholds
decide when a branch is committed versus held plural, which is the machinery
[Phase 3](../2026-07-21/phase-3-audit-geometry-is-not-trusted-interpretation.md) exists to
keep honest.

**Fix.** One method, `func (c DifferentiationConfig) withDefaults() DifferentiationConfig`,
applied at the constructor and asserted — not re-applied — downstream.

**Verify.** A test that constructs a partial `DifferentiationConfig` and asserts every
consumer sees the same effective values.

### 2.6 Two Jaccard implementations with deliberately different empty semantics

**Where:** `jaccardSpans` (`activation_field.go:552`) and `jaccardFeatures`
(`density_pockets.go:223`).

**Defect.** Same measure, two bodies. The divergence is intentional — `jaccardFeatures`
special-cases two empty feature sets to 0 with the comment "shared absence is not a
density observation," while `jaccardSpans` reaches 0 through an empty union incidentally
rather than by decision.

**Fix.** Do NOT merge them into one generic function; the empty-set decision is a real
semantic difference and collapsing it would silently change one caller. Instead make the
difference legible: have each name the other, and state at `jaccardSpans` whether its
empty-union zero is a decision or an accident. If it is a decision, say so; if nobody
knows, that is itself the finding.

**Verify.** Documentation only, unless the `jaccardSpans` empty case turns out to be
undecided — in which case it needs a test that pins whichever answer is chosen.

### 2.7 `meanWithinSimilarity` and `meanBetweenSimilarity`

**Where:** `density_pockets.go:193` and `:208`.

**Defect.** Identical accumulate-and-average bodies over different pair generators.

**Fix.** Low value, listed for completeness. Only worth doing if 2.6 is being touched
anyway, since both live in the same neighborhood.

---

## Tier 3 — dead and redundant

### 3.1 `ObserveImpulse` builds a slice it never reads

**Where:** `ingest.go:38-43`.

```go
var derivedFrom []string
for _, t := range impulse {
    o := b.appendObservation(...)
    written = append(written, o)
    derivedFrom = append(derivedFrom, obsID(o))   // never read
}
```

The compiler permits it because appending counts as use. `b.currentBornFrom` is built
immediately afterward from `written` and is the value actually carried forward.

**Fix.** Delete the variable. If it was scaffolding for passing raw provenance into a
decomposer call, that call does not exist here and the intent belongs in a comment, not in
a live slice that allocates on every impulse.

### 3.2 `appendObservation` re-assigns embedded fields

**Where:** `being.go:163`.

```go
o := Observation{... Triple: t ...}
o.Subject, o.Predicate, o.Object = t.Subject, t.Predicate, t.Object   // already set
```

`Triple` is embedded and was assigned whole on the line above. The three assignments are
no-ops. Worth deleting specifically because the line implies the embedding is partial —
which would be alarming if true.

### 3.3 `Expression.Due` has zero callers

**Where:** `expression.go:305`.

No caller in the package, including tests. `Settle` (`being.go:307`) reimplements it inline
against `Expr.Open()`, and correctly so — `Settle` needs the owning expression for the span
ID, which `Due` discards.

**Fix.** Delete it, or state in one line that it is retained as the documented shape of
the settle contract. Its doc comment is the best statement in the package of why `within`
solves the scope problem, so a straight delete loses something — move that paragraph to
`Settle` if the function goes.

### 3.4 `Affordances.Standing` is exercised only by its own test

**Where:** `affordance.go:107`.

The compass-versus-bet distinction it encodes is real and is referenced in the type's own
documentation. This reads as reserved API rather than rot, so the action is to SAY SO —
one line marking it as not yet on the live path — rather than to delete it.

### 3.5 Five nil-map guards that cannot fire

**Where:** `store.go:167`, `:211`, `:267`, `:368`, `:467`.

Each re-initializes a map that `NewExpressionGraph` (`:144-159`) already allocates. There
is no constructor-bypassing path — `ExpressionGraph` is never built as a bare literal
outside tests.

**Fix.** Delete. Defensive code that cannot fire is not defense; it is an assertion that
the constructor might be incomplete, which invites someone to add a sixth.

---

## Tier 4 — stdlib and boilerplate

The module is on `go 1.25.0`. All of these are mechanical.

| Where | Now | Should be |
|---|---|---|
| `memory.go:306` | `maxInt` | builtin `max` |
| `being.go:558` | `contains([]string, string)` | `slices.Contains` |
| `activation_field.go:586,597` | `decayStrings` + `decaySpans` | one generic over `~map[K]float64` |
| `affordance.go:117` | `namespaceOf` hand-rolled byte loop | `strings.Cut` — used for the same job at `:68` of the same file |
| `being.go:164`, `being.go:469`, `memory.go:209` | `counter := b.CountTokens; if counter == nil {...}` x3 | one `func (b *Being) countTokens(string) int` |

The last one is the only one with any weight: three copies of a nil-fallback for an
injected seam means three places to forget it. `appendObservation` and `plan` both do it;
a fourth caller will not.

---

## 5. Already adjudicated elsewhere — do not act from this document

Recorded so a reader knows these were reached independently and then found to be settled,
rather than missed.

**`wave.go` and `segment.go`.** This review independently confirmed that all of
`Roots`, `HeatClasses`, `Resolve`, `Resolution.Commit`, `ClassEdges`, `Spans`, `Classes`,
`Lens`, and `BuildResolutionIndex` have no callers outside `wave_test.go` and
`segment_test.go` — 618 production lines, 986 with tests.
[Defects Found in Review](../2026-07-22/defects-found-in-review.md) section 4 has already ruled on this:
"Prototype history, retained deliberately. Not on the live ingest path. Do not modernize,
do not delete." That ruling stands and this document does not reopen it.

ONE NARROW EXCEPTION IS PROPOSED, and only because it is a comment rather than code.
`wave.go:29-31` currently reads:

> The wiring seam (spec delta 6): v3's walking skeleton is target-driven and has no
> impulse/decomposer path yet. When ingest lands, the wave runs between decompose and
> present; until then everything here is callable and tested against the fixtures.

Ingest has landed. It did not wire the wave — `memory.go:76` states the successor
directly ("SpanIndex is the direct, derived memory index. It replaces the class hop"), and
`ingest.go:20` calls the wave's own output "the retired class-wave inspection types." So
the file's header now promises a wiring step that is not coming. Retaining prototype
history is a decision; retaining a comment that misdescribes the live architecture to the
next reader is not part of that decision. Amending those three lines to say "prototype,
superseded by SpanIndex, retained deliberately" costs nothing and contradicts nothing.

**`IngestResult.Resolution`.** Vestigial in the same way and for the same reason:
`ingest.go:65` fills only `Primary: -1` and `Novel`, leaving `Bindings`, `Ranked`, and
`JointPrecedent` zero, and all four consuming tests read only `.Novel` — which is exactly
`len(groups) == 0`. Since the `Resolution` type is `wave.go`'s, this is downstream of the
ruling above and is NOT proposed as a change.

**`Classes()` has no live callers.** Covered by section 4 of the defects document and by
[Episode Names Two Different Things](episode-names-two-different-things.md). Not a code fix.

**`copy` shadowing at `prediction_calibration.go:146`.** Reached independently; already
filed as defect 2.1. Not restated here.

**Orphaned sibling packages.** `traverse`, `tui`, `debugview`, `user-dna`,
`placegateway`. Already recorded as a human decision, not a cleanup.

---

## 6. Not found

Stated so a reviewer knows the search happened rather than assuming it was skipped.

No simplification was found that would reduce the number of distinct concepts in the
architecture. Every abstraction in the package earns its place: the three-valued `Truth`,
the provenance/fidelity split, the separate interpretation and consequence claim paths,
the durable `ComparisonField` retained alongside the scalar `FusionProjection`, the
`Mode`-versus-`Provenance` distinction, and the derived/durable split across the two
stores are all load-bearing, and each is cheaper than what it replaces. Several look like
duplication on a first read and are not.

In particular, the two obvious "collapse these" candidates should be refused if proposed
later:

```text
interpretations/interpretationEdges/interpretationBets  mirroring
exprs/edges/bets                                        -> refuse. this is the
                                                           lucky-hit correction.

ComparisonField retained beside FusionProjection        -> refuse. the scalar is a
                                                           reversible projection and
                                                           the dimensional record is
                                                           the authority.
```

Nothing in Tier 1 through Tier 4 touches either.

The duplication that does exist is shallow and local. No item in this document requires an
architectural decision, and none of them should require touching a test's expectations. If
one does, stop and write it up instead.
