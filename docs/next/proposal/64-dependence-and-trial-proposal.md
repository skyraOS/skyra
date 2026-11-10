# 64 — Two primitives: dependence assumed, removed by trial

Date: 2026-11-10

Status: **draft, being worked through**. Builds on
[61](61-static-and-dynamic-proposal.md), whose data model stands. It
replaces the guard: 58's conditions, 59's voter as a chain the model
wrote, and 62's citation compiled into one. It restates
[62](62-citation-and-learned-equivalence-proposal.md) and
[63](63-evidence-and-replay-proposal.md) in its own terms and supersedes
parts of each (What changes). Nothing here is built. The syntax shown is
provisional.

**The model writes two things: eliminations and introductions.** It
writes no condition. An elimination is a derivation, which the system
computes and checks. An introduction is an assertion, which the system
cannot check. When what was learned applies is no longer something the
model says. It is something the system finds:

1. **Every unit depends on everything its frame has seen.** That is where
   each starts.
2. **A dependency is removed only by trial**: the same unit, derived
   again where that dependency read something else.
3. **The system's work is to remove dependencies**, as many as trials
   support, at a confidence the program sets by what a wrong act costs.

## Why

**A guard is an assertion.** It says: this is the condition under which
this holds. The system cannot check that. The return Claude wrote with
`cat` was `size($4) > 0 ? "return 0 $4"` (61, What a trial cannot see):
a condition true of nearly anything, trusted as written. 59's experiment
asked one model seven times for one frame's guard and got guards that
agreed on the values asked about and split three ways on values never
seen. The instruction was stable. The condition around it was not.

**A citation is an assertion too.** 62 ties an introduction to the span
it was derived from, and 63 extends that to operators. Both make the
model say what its output depends on: `Γ ⊢ "-a"`, with `Γ` written by the
model. A model that leaves `not` out of `Γ` has asserted that `not` does
not matter, and 62 names that risk without closing it (62, The reading).
Writing `Γ` down does not make it true.

**Agreement is not proof, and nothing else is on offer.** What a phrase
means has no ground truth the system can consult. Several models giving
one answer is evidence, weaker the more alike the models are. What is
left is what people already do with a prompt: run it again, vary it, and
see whether the output holds. This proposal is that practice with the
runs recorded, the variations named, and the threshold a constant.

## What stands

- **61's data model**: the six pairs, the key as everything seen and
  never a value, memory as a fold of the log, the tree, boot.
- **Elimination and introduction**, told apart by form; variables
  append-only; an instruction made entirely of variables; the copy check.
- **59's tally**, inhibition and floor, now over the units that vote.
- **63's sections 1, 3, 5 and 7**: quorum counted on the ballot, a
  template pinning what it was authored with, a refutation as an event,
  the copy check as a search.
- **`readOnly`**, the person's declaration of a tool's effects.

## What the model writes

At a startle the model writes a **step**: eliminations, introductions,
and one operator, itself an introduction (61, Variables).

```text
$1: "list all files in this directory:\n/tmp"      goal
$2: leaf ls                                         tool

eliminate $3 := $1 / [2:2]
introduce $4 := "-a"
$2 $4 $3
```

There is no `c ? "a" : null`, no chain, and no `citing`. The refusals
that concern a template stay (`fits`, `copies`, `join`, `unstored`,
`unread`); those that concern a condition go (`cond`, `guard`,
`unselected`).

### A predicate is an elimination

A step sometimes turns on a value: restart the service at three
restarts, not at one. That is a derivation from something held, so it is
an elimination:

```text
eliminate $7 := $3.restarts > 2          bool
```

- **The system computes it**, as it computes a cut, and synthesizes its
  type. It never needs the model, and it replays.
- **Its literal is code** (`2`), as a guard's `x > 7` was (61, The rule).
- **The expression language may stay CEL**, one expression over values
  seen. What goes is the chain around it.
- **The model chooses which predicate to compute. It does not say that
  the step depends on it.** The system finds that out, below.

62's text match (`$1 / "all files"`) is the same kind of elimination: it
reads its text or nothing.

## Dependence

A **unit** is one introduction a step makes: a literal, an exit status,
or the operator. Each unit is stored with its **dependencies**.

A **dependency** is an elimination and what it read when the unit was
authored: `$1 / [2:2]` read `/tmp`; `$3.restarts > 2` read `true`; `$2`
is `def ls`. Both halves are code, as 62's cited span is, and a long
text is held by its hash. A definition still holds no value.

**A unit starts with every dependency its frame offers:**

- each `$n` seen, whole: the goal, each tool, each context value, each
  arrival;
- each elimination the step wrote, predicates among them;
- the other units of its step.

**A cut splits its source.** When a step cuts `$1 / [2:2]`, the
dependency on `$1` becomes two: the part cut, and the rest. Parts are
runs of lines, fields, and runs of words (62, Lexical units). So the
model's eliminations are still its attention, recorded (62): they decide
where a value can come apart. They no longer decide what matters.

**A dependency is in one of three states**, per unit:

| State | It holds in a frame when | How it is reached |
|---|---|---|
| **equal** | the elimination reads what it read at authoring | where every dependency starts |
| **among** | it reads one of the values seen | each agreeing sighting adds its value |
| **free** | always | `span` distinct values, all agreeing |

*Among* is 61's sum, `I₁ + … + Iₙ`, written by memory: it records what
has been seen and assumes nothing. *Free* is a hypothesis: that the next
value will agree as the seen ones did.

## Deciding

**A unit votes in a frame when every dependency it has holds there.** A
step votes when all its units do. The votes are tallied as 59 tallies
them, under 63's quorum, counted on the ballot. This is the guard,
derived: a condition no one wrote, made only of what was observed.

A frame none of whose stored steps vote has an empty ballot: `null`,
`startle`, the path that exists. The model is asked, and writes a step.

## Removing a dependency

**Each startle is a sighting.** The step the model writes is compared,
unit by unit, with the steps stored at its key:

```text
stored    $1 rest = "list all files in this directory:"   $1[2] = "/tmp"    →  "-a"
sighted   $1 rest = "list all files in this directory:"   $1[2] = "/var"    →  "-a"
```

- **Agreement** is the same unit at the same place in the template: the
  same literal for the same destination, or the same operator on the same
  position. Every dependency that read something else gains the new
  value: equal becomes among, and among grows.
- **At `span` distinct values a dependency is free**, for that unit.
- **Dissent** is another unit where every dependency of a stored one
  holds. Both stand as votes, and the tally decides (59). A free
  dependency at which the two differ returns to among, and stays there.
- **Evidence from the world is final; evidence from a model is a vote.**
  A shadow whose output refines differently, a bound exhausted, or the
  person rejecting a result refutes (63, section 5). A model writing
  something else is one more voter.

**The predicate needs no rule of its own.** Restart at `restarts` 3, 5
and 9: the dependency on `$3.restarts` grows to among `{3, 5, 9}` and
goes free, while the dependency on `$3.restarts > 2` has read `true`
every time and stays equal. A sighting at `1` reads `false`, and the unit
does not vote. The threshold was learned from sightings. The model only
proposed looking.

**Order of removal is given, not searched.** There are `2ⁿ` sets of
dependencies and nobody enumerates them. A dependency moves only when a
sighting happens to differ at it, and each sighting moves all it differs
at, together.

## Trials

A sighting that arrives in use costs nothing extra: the model was asked
anyway. A **trial** is a sighting the system makes.

- **A trial ask varies what the model is shown**: one dependency reads
  something else, everything else is as it was, and the model writes a
  step. **The step is never run.** Nothing reaches the world, so a unit
  can be tried even when its tool has effects.
- **Variations are named in the language version**: delete a part,
  substitute a part with one of its type seen elsewhere, reorder parts.
  Order is in the reading (62), so reordering is how the system learns
  whether order matters, and for language it usually does.
- **A trial is logged**, with what was varied and the step written, so
  what it supports replays.
- **A trial measures the model.** A model that ignores `not` will show,
  under every variation, that `not` does not matter. Trials make a
  model's behavior stable and known. They do not make it right, and the
  world's refutation stays the backstop.
- **Trials cost asks**, which the system exists to remove. They are an
  investment: spent before behavior is stable, repaid by every reuse
  after. They belong in idle time and under a budget (open question 3).

## Every reuse rule is a removed dependency

61 to 63 name several ways what was learned at one key counts at
another. Each is one dependency, in one state:

| 61 to 63 | Here |
|---|---|
| the exact rule | every dependency equal |
| a passed span, a hole in a reading (62) | a cut part, free |
| a cited span (62) | a part that stayed equal |
| the cited-spans reuse rule (62) | the uncited rest, free |
| the closed reading (63) | the rest held equal, by rule |
| a class (61); a learned sum (62) | a tool's dependency, among |
| `span` (62, 63) | the count that takes among to free |
| reuse across a prefix (61) | later arrivals, free |
| promotion of a reading (62) | the states at a position, moved into the key |

So 62's open question 2, *uncited text: literal or hole*, has one answer:
equal until trials free it.

## Stakes

One threshold cannot serve `ls` and `rm`. A regulator does not hold a
cancer drug and an allergy pill to one standard either: what is
acceptable moves with what a failure costs.

```text
constants: {
  …
  quorum:  nat ≥ 1                 votes cast before a tally is trusted (63)
  span:    nat ≥ 2                 distinct values before a dependency is free
  passed:  nat ≥ 1                 the same, for a part the step passes; default 1
  effects: { quorum, span, passed }   the same three, for an act on a tool
                                      not readOnly; each at least the above
}
```

- **`passed` defaults to 1**: a part the step passes on is free from its
  first sighting, which is 61's behavior today (a second run on another
  file asks nothing). It is a default, not a truth: `rm` on a path is a
  passed part, and `effects.passed` is where that is raised.
- **For an act with effects, only a passed part goes free.** Every other
  part of its goal stays equal or among. This is 63's closed reading,
  now a restriction on states.
- **A tool's dependency is never free.** It reaches among, a sum, and a
  borrowed act is still tried against a shadow where the tool is
  read-only (61, Reuse rules).
- **What a count buys.** `n` agreeing sightings and no dissent bound the
  rate of dissent near `3/n` at 95% confidence: 10 sightings, under 30%;
  100, under 3%. A failure seen once in a thousand needs about three
  thousand sightings to find. Trials catch what is common. What is rare
  is found in use, which is why refutation is final and cheap to record.

The stages are a clinical trial's:

| Stage | Here | Reaches the world |
|---|---|---|
| safety, few subjects | trial asks; shadows of read-only tools | no |
| efficacy, at scale | `quorum`, then `span` | only what is trusted |
| surveillance after release | refutation, final, in the log | yes |

## The log and memory

```text
event ::= hesitate key
        | derive   key step values        a startle: the step, and what each
                                          dependency read
        | trial    key variation step     a trial ask: never run
        | refute   what                   63
```

- **Memory is a fold of the log**: each key's steps, each unit's
  dependencies, each dependency's values seen.
- **A dependency's state is computed, not stored.** It is a function of
  the values seen and the program's constants, so two programs differing
  only in `span` share memory, as two differing in `quorum` do (61,
  Process and heap).
- **The key is 61's.** Dependencies are held with the unit, beside the
  key; they are not in it, until promotion moves them (open question 6).

## What changes

| 58 to 63 | 64 |
|---|---|
| 58: the model writes a CEL chain, `c ? "a" : null` | a step: eliminations, introductions, one operator |
| 58: a branch's effective condition | a unit's dependencies, each in a state |
| 59: a voter is a chain | a voter is a unit; the tally is unchanged |
| 61: an implementation is a guard chain | a step and its units' dependencies |
| 62: an introduction cites; the citation is compiled into the guard | no citation: it depends on everything, until trials say otherwise |
| 62: a reading is literals and holes the model marked | derived: parts equal, among or free |
| 62: an introduction is an entry keyed by what it cites | a unit, shared wherever its remaining dependencies hold |
| 62: equivalence learned, gated by `span` | a tool's dependency, among; `span` for every dependency |
| 63, section 2: a chain commits only if it votes | moot: no chain; a step is authored for the values at hand |
| 63, section 4: an operator cites; the closed reading | no citation; closed is a restriction on states |
| 63, section 6: `span` counts an antichain | `span` counts distinct values at one dependency |

In the code: `tree`'s `StoreConds`, `CheckGuard`, `CommitChain` and
effective conditions go, with the chain grammar in `drive`'s prompts;
`tree.Cast` evaluates dependencies in place of conditions; `Link.Cond`
goes and `Event` carries a step and what it read; `Sel` gains
predicates; `Constants` gains `span`, `passed` and `effects`;
`Program.Classes` goes (62). In Lean, `Refine.lean` is rewritten over
units; `tally` and its theorems keep their statements.

## Open

1. **Interactions.** *Among* is held per dependency, so a frame reading
   `a` at one and `y` at another votes though `(a, y)` was never seen
   together. Whether combinations are assumed independent until a
   dissent, or whether a unit also keeps the joint sightings and votes
   only on those until `span` of them.
2. **What a dissent does to a free dependency.** Returning it to among
   is conservative and final. Whether a model's dissent should be final,
   or only the world's, with a model's tallied against the agreeing
   sightings under the inhibition.
3. **The trial budget.** How many trial asks, on which units first. The
   candidates: units at keys seen most often, units nearest `span`, and
   units of acts with effects, where a sighting in use is most costly.
4. **Which variations are sound.** Deletion and substitution leave text
   a person might write. Reordering often does not, and a model's answer
   to a scrambled goal may say little about real ones.
5. **Whether the model is a dependency.** 61 keeps the model out of the
   key, as provenance. Under *depends on everything* a unit depends on
   the model that wrote it, and is freed when another model writes the
   same. That would make quorum count distinct models, which is the
   stronger evidence, and costs an ask per model.
6. **Promotion.** When a position's states enter the key (62, open
   question 3), so that goals of one reading share a prefix.
7. **Granularity.** How finely a value comes apart before the model cuts
   it: whole, or by line and field from the start. Finer parts free
   sooner and interact more.
8. **A unit's dependence on its step's other units.** `-a` and `-l`
   written together as `-la` are one literal; written apart they are two
   units, each depending on the other. Whether that dependency is ever
   freed.
9. **`passed` at 1.** It keeps today's behavior and it is the one
   dependency freed with no trial. Whether the default should be 2, at
   one more ask for every new step.
10. **What the model is shown.** Whether an ask shows the states learned
    so far at its key. It would help the model agree with itself, and for
    the same reason weakens each sighting as independent evidence (63,
    open question 3).

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean`, `Refine.lean` and `Machine.lean`,
before any Go:

- **The model writes no condition:** a step is eliminations,
  introductions and one operator; no step holds a predicate over values
  but an elimination.
- **Eliminations are the system's:** every elimination, predicates among
  them, is a total function of what was seen, with a synthesized type.
- **Definitions still hold no value:** a dependency is an elimination
  and a literal; `key_values` holds, the key unchanged.
- **Dependence starts total:** a unit just authored votes in a frame
  only if the frame reads, at every dependency, what the authoring frame
  read.
- **Authoring votes:** a step votes in the frame it was written in.
- **States are monotone in agreement:** an agreeing sighting never makes
  a unit vote in fewer frames; a dissent never makes it vote in more.
- **Free needs `span`:** a dependency is free only when `span` distinct
  values were seen at it, each in an agreeing sighting; raising `span`
  only withholds.
- **Effects stay closed:** a unit of an act on a tool not `readOnly` has
  no free dependency on its goal but a passed part; a tool's dependency
  is never free.
- **A trial reaches nothing:** a trial ask gives no `runLeaf` and
  changes no tree.
- **Memory is a fold of the log**, with `derive` and `trial`; a
  dependency's state is a function of that fold and the constants.
- **The tally is 59's:** with every dependency equal, the decision at a
  key is 59's tally over the steps authored for exactly these values.
- **Refutation is final** (63), and overrides any count.
