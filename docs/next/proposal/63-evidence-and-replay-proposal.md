# 63 — Amendment to 61 and 62: evidence counts where it was cast

Date: 2026-11-10

Status: **draft, being worked through**. Amends
[61](61-static-and-dynamic-proposal.md) and
[62](62-citation-and-learned-equivalence-proposal.md); everything in them
it does not name stands. It settles 62's open questions 4, 7 and 8, part
of 1, 2 and 6, and changes one thing 62 proposed (how a template holds an
introduction). Sections 2, 5 and 7 change code that is built; the rest
amend what 62 has not built yet. The syntax shown is provisional.

62 found the gap: what was learned is reused where nothing ties it to
what it was learned from, and citation is the tie. The same gap is open
in six more places. In each, evidence gathered in one place is counted,
or takes effect, in another.

Seven changes:

1. **Quorum is counted on the ballot**, not at the key.
2. **A chain is committed only if it votes** for the values it was
   written for.
3. **A template pins its derivation.** It holds the literal it was
   authored with; the shared entry proposes, and never rebinds.
4. **An operator cites**, and an act with effects reads its goal
   **closed**.
5. **A refutation is an event** in the log.
6. **`span` counts an antichain**, and two alternatives agree only link
   for link.
7. **The copy check searches**; it does not enumerate.

## Why

**A key is not a situation once citations are compiled.** A goal's
interface is `string`, so every goal under the same tools shares a key
(62, Why). 62 keeps the key and moves the difference into the guard: each
chain votes only for the goals its citations read. The key then holds
chains for many goals, and anything counted per key is counted across
all of them.

Quorum is counted per key. `decideAt` is `null` while
`implementations r < q` (Model.lean), and `Memory.Decide` does the same
(`memory.go`). Under 62, with a quorum of 2:

```text
key  [string, leaf rm]
  voter 0   $1 reads "delete"   ?  "$2 $1[2]"       authored for: delete this file
  voter 1   $1 reads "list"     ?  "return 1 $1"    authored for: list this file
```

Two implementations, so the quorum is met. The next *delete this file*
gets one vote, from voter 0, and is trusted on one sighting. Quorum was
defined as how many independent sightings before anything is trusted (61,
What stays). Here it is how many goals the key has seen.

**A chain that never voted is a voter.** `wrote` appends the startle and
then asks whether the chain selects the values it was written for; when
it does not, the model is asked again and the chain stays (`machine.go`,
Machine.lean's `wrote`: the event, then `afterStartle`). It counts toward
the quorum above, and it was never tied to any observation at all.

**An entry looked up at `trust` moves under a stored template.** 62 has
the template name the entry and bind its value at `trust`. The entry is
tallied and shared, so any frame, at any key, can add a voter to it. A
template trusted yesterday with `-a` runs `-la` today because two other
frames derived `-la` from the same span, and no run ever saw this
template with `-la`. If the entry falls to a conflict, the template has
no value to bind, and 62 does not say which key startles.

**An act can run for a goal it never read.** 62's open question 4:
`$2 $3` introduces nothing, cites nothing, and selects for every goal at
its key. The return's `exit(0)` cites the goal and so startles on a goal
never seen, but the act has already run. For `ls` that is a wasted
process. For `rm` it is the failure the system exists to prevent.

**Refutations are a second source of truth.** 61 says memory is a fold of
the log and nothing else (Memory; invariant 3). Refutations are not in
the log. Machine.lean holds them in `State.refuted`, passed to
`State.init` beside the log; `Replays` is `memory = replay log` and says
nothing of them. In Go they are keys in the memory map under a
`refuted:` prefix, and `Store` appends them after the run's events and
applies them after every event on load. Where a refutation fell among
the decisions is lost, so a borrow made before it replays as refused.

**`span` counts keys, and there are few.** 62 counts distinct keys at
which two alternatives hold the same value. With every goal a `string`,
an act frame holding one tool has a handful of keys, most of them
prefixes of one another: one run's successive steps. And 62 does not say
what *the same value* compares. If it is the template alone, `cat` and
`wc` agree on `$2 $1[2]`.

**The copy check is cubic.** `candidates` gives every run of a text's
lines, and `holds` cuts and compares each (`variables.go`; Model.lean's
`candidates`, `holds`). A value of `n` lines is `n²/2` cuts of up to `n`
lines, for each literal, at each ask. 62 adds runs of words.

## 1. Quorum is counted on the ballot

**The quorum counts voters that cast a vote for these values.** A voter
casts at most one (59), so it is the number of votes:

```text
decideAt q m k v  =  null                              k unseen
                     null                              fewer than q votes cast on v
                     tally (the votes cast on v)       otherwise
```

- **A silent voter changes nothing.** An implementation that casts no
  vote for `v` enters neither the quorum nor the tally for `v`. The
  decision at `v` is a function of the chains that select `v`.
- **Quorum is again sightings of a situation**, where a situation is
  what the guards tell apart: under 62, a goal's reading.
- **It still only withholds.** `decideAt_quorum_mono` holds as stated;
  `decideAt_tally` takes `q ≤ votes cast` for `q ≤ implementations`.
- **A quorum of 1 is unchanged**: one vote cast is a ballot that is not
  empty, which the tally already asks for. The default stays 1, and every
  vector decided at quorum 1 stands.
- **A borrow is gated the same way**, since the swapped key is decided
  by the same function (62, `span`).
- **An introduction's entry is unchanged.** Its voters have no
  conditions, so every one casts, and votes cast are implementations.

This is the defense 62 names for under-attention, *several voters
attending differently* (62, The reading). It works only if the several
voters all read the goal at hand.

## 2. A chain is committed only if it votes

**`wrote` asks whether the chain selects before it records.** A chain
with no link holding on the values its frame holds is refused,
`unselected`, and nothing is appended to the log or to memory. The ask
still counts against the refinement bound.

```text
61   check → commit → selects? no → refuse (the chain stays)
63   check → selects? no → refuse (nothing stays)  |  yes → commit → trust
```

So every implementation in memory cast a vote at least once: at its
authoring. 62 already requires that a cited text is read where the link
runs (62, What the system checks). This is the same requirement on the
whole chain.

## 3. A template pins its derivation

**A committed template holds what it was authored with**: the entry's
key and the literal derived, together. It names one voter's derivation,
not whatever the entry's tally selects later.

```text
62   template:  $2  ⟨entry: $1 / "all files" → def ls⟩  $1[2]         value bound at trust
63   template:  $2  ⟨$1 / "all files" → def ls : "-a"⟩   $1[2]         value fixed at commit
```

- **`trust` reads a template and a frame, and nothing else.** What a
  stored template runs is fixed at commit and cannot be changed by a
  commit at another key. This is 61's property, which inlining a literal
  gave for free.
- **The entry stays, as 62 defines it**: keyed by what was cited, then
  the destination; tallied; shared. Each commit that introduces is still
  one voter at its entry.
- **The entry proposes.** It is read in two places, both before a
  commit, never at `trust`:
  1. **By a reuse rule.** When a frame's own key selects nothing, the
     cited-spans rule (62, The reading) finds implementations at other
     keys whose citations this frame's values read. The entry trie is its
     index: by span, then destination.
  2. **By the model's ask.** A span the frame holds that some entry
     cites is shown with what was derived from it, as `Mates` are shown
     today. The model may cite it and take the value, or derive another,
     which is one more voter at the entry.
- **Disagreement is settled where it is seen.** `-a` and `-la` from one
  span are two voters at the entry and two templates at whichever keys
  hold them. Each key's tally decides for that key (59). The entry's
  tally decides only what is proposed.

The saving 62 wanted is kept: a derivation made once is offered
everywhere its span is read. What is dropped is one frame's commit
changing another key's decision.

## 4. An operator cites

**Every operator is an introduction** (61, Variables): a delegation, an
act, a boot and a return each construct one composite observation. 62
makes every introduction cite, and applies it to literals and to
`exit(n)`. It applies to the operator as well.

```text
$1: "delete this file:\n/tmp/x"              goal
$2: leaf rm                                  tool, not read-only

eliminate $3 := $1 / "delete this file:"     cited
eliminate $4 := $1 / [2:2]                   passed
$2 $4 citing $3
```

The operator's citation is compiled into its link's condition as any
citation is. The link selects only where `$1` reads the cited text.

**Which operators must cite.**

| The link's template | Must cite its goal | Reading |
|---|---|---|
| an act on a tool not declared `readOnly` | yes | closed |
| an act on a `readOnly` tool | no | open, if it cites |
| a delegation | no: its child decides at its own key | open, if it cites |
| a return | yes, through `exit(n)` (62) | open |

This settles 62's open question 4: the return's citation is enough only
where the act before it has no effect. An act with effects answering an
arrival cites the arrival as 62 allows, and its goal as well (62, open
question 6, for this case).

### The closed reading

62's reading is open: cited literals in order, a hole at each passed
span, and anything uncited between them ignored. *do not delete* is the
case 62 names. A model that cites `delete` and passes the path writes a
link that selects with or without `do not`, and a quorum does not help:
every voter cites `delete`.

**An act with effects reads its goal closed: every word of the goal is
cited or passed.**

```text
open     … "delete" … ⟨string⟩            fits  "do not delete /tmp/x"
closed   "delete this file:" ⟨string⟩     does not: "do" and "not" are neither
```

- **It is the same reading with no gaps**: the literals and holes, in
  order, are the whole goal.
- **A goal the closed reading does not fit selects nothing**: `null`,
  `startle`, the path that exists. *please delete /tmp/x* startles too.
  That is one ask, set against running `rm` on a goal that was not read.
- **A closed reading has a literal between any two holes.** Two passed
  spans side by side have no boundary (62, open question 1); in a closed
  reading that is refused at commit, with the reason. A hole is one or
  more words, bounded by the literals beside it or by the goal's ends.
- **Open stays the default** for everything else, and 62's open question
  2 stays open for it.

`readOnly` is already the person's declaration that a tool has no effect
(61, The program), and a shadow is already safe only under it. This adds
one use: an act outside it is held to the closed reading.

## 5. A refutation is an event

```text
event ::= hesitate key
        | startle  key implementation
        | refute   what                61: a class;  62: a pair at a position
```

- **It is logged at the step that judged it**: a trial whose outputs
  refine differently, or a frame that borrowed and exhausted a bound.
- **What is refuted is a fold of the log**, as memory is. `State.init`
  takes the log alone. `Replays` covers both: `memory = replay log` and
  `refuted = refutations log`.
- **A borrow replays.** It reads the refutations before it in the log,
  so a decision made before a refutation is the same on replay.
- **The store is one stream.** Each line is an event; no line is written
  from a difference between two memories, and none is applied out of
  order.
- **Refuted stays final**: `go_refuted` and `step_refuted` are unchanged,
  and follow from the log only growing.

## 6. `span` counts an antichain

**Agreement is link for link.** Two alternatives agree at a pair of keys
that differ only at their position when each key holds an implementation
whose chain equals the other's: every condition, compiled citations
among them, and every template with its pinned derivations. A template
alone is not enough.

`cat` and `wc` show why. Their acts are both `$2 $1[2]`, and read-only,
so both may go uncited: the pair agrees at the key where the tool is run,
and running `cat` for `wc` there is harmless. At the next key their
returns differ, `exit(0)` citing *show the contents* against *count the
lines*, so the pair does not agree there, and nothing is borrowed there.
Agreement is per position because sameness is.

**Distinct keys are keys no one of which is a prefix of another.** A key
and its prefix are one frame's successive steps: one context, seen
twice. `span` is the size of the largest set of agreeing keys that is an
antichain in the trie. This settles 62's open question 7 with its first
candidate made exact. Its stricter one, keys of different goal readings,
follows once readings are in the key: two readings part the trie, so
neither key is a prefix of the other.

**`span` is off unless set** (62, open question 8). Until a reading is
promoted into the key (62, open question 3), an act frame holding one
tool has two or three keys in any antichain: its arrivals' shapes. A
finite default would either never fire or fire on one conversation.
`span` gets its range from promotion, so promotion is built first.

```text
constants: {
  …
  span:  none | nat ≥ 1      none: nothing is borrowed without a count
}
```

## 7. The copy check searches

**`holds` stays the definition; a search computes it.** A cut of text
reads exactly `s` when `s` is the whole text or a run of its lines, which
is `s` found in the text starting at a line's start and ending at a
line's end. That is one pass over the text. A field's cuts recurse
through the field as they do now.

- The core proves the search agrees with `holds` on every value, and the
  Go runs the search. No refusal changes, so no vector changes.
- 62's runs of words are the same search on word boundaries, and a cited
  text is already matched that way (62, Lexical units).
- The check's false positives (61, open question 17; 62, open question
  9) are untouched. Finer selectors make them more frequent, which is a
  reason to settle that question before word runs are built.

## What changes

| 61 / 62 | 63 |
|---|---|
| 61, What stays: a key's tally runs once it holds `quorum` implementations | once `quorum` votes are cast on the values |
| 61, Memory: the implementation the model authors is committed | committed only if it selects the values it was authored for |
| 61, Memory: refutations kept in memory beside the keys | events in the log; what is refuted is a fold |
| 62, An introduction is a memory entry: the template names the entry, bound at `trust` | the template holds the entry's key and its literal; the entry proposes |
| 62, Citation: every introduction cites (literals, `exit(n)`) | every operator too, where the table in section 4 requires it |
| 62, The reading: open | closed for an act on a tool not `readOnly` |
| 62, `span`: distinct keys holding the same value | an antichain of keys agreeing link for link; off unless set |
| 62, open question 4 | settled: section 4 |
| 62, open question 7 | settled: section 6 |
| 62, open question 8 | settled: off unless set |
| 62, open questions 1, 2, 6 | settled for the closed reading; open otherwise |

In the code: `Memory.Decide` counts the votes `tree.Cast` gives, in place
of `Implementations`; `Machine.wrote` evaluates the chain before
`record`; `Event` gains a refutation, `Memory.Refute`'s `refuted:` keys
and `Store`'s `refuteLine` go, and `Store.Record` writes the log as it
is; `Frame.Holds` searches. With 62: a stored `Arg` holds an entry's key
and its `Lit` together; `Template` gains what its operator cites; `Sel`'s
text selector gains the closed form; `Constants.Span` is optional.

## Open

1. **Closed, and how strict.** Whether a closed reading matches case and
   punctuation exactly (`Delete`, `file:`), or the language version
   defines a normal form for words. Exact is safe and costs asks.
2. **Effects passed down.** A delegation passing a tool with effects
   cites nothing here: its child's act is held to the closed reading at
   its own key, under the goal the delegation wrote. Whether the
   delegation's `goal(…)` must then be a cut of its own goal and never a
   join, so the child reads words the user wrote.
3. **What the ask shows from an entry.** Every derivation at the entry,
   or only what its tally selects; and whether showing a derivation
   biases the next voter enough that two voters from one entry are not
   independent (59, question 2; 61, open question 5).
4. **A pinned derivation the entry has turned against.** A template
   pinned to `-a` keeps running where its key's tally selects it, while
   the entry now proposes `-la`. Whether the entry's tally should lower
   the template's weight at its own key: a reuse rule's weight, never a
   rebinding.
5. **Quorum for effects.** Whether an act with effects takes a quorum of
   its own, higher than the program's, now that the quorum counts
   sightings of one reading.
6. **Committed chains that stop voting.** A chain that selected at its
   authoring may never select again (its citations read one goal, seen
   once). Whether memory ever drops it, or lineage keeps everything, as
   now.
7. **Trials under a closed reading.** A borrow is made only for a
   `readOnly` tool, so no borrowed act is closed. Whether a learned
   equivalence between two tools with effects can ever be tried, with no
   shadow to run.

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean` and `Machine.lean`, before any Go:

- **Quorum on the ballot:** `decideAt` is `null` when fewer than `q`
  voters select `v`. `decideAt_quorum_mono` as stated; `decideAt_tally`
  and `decideAt_below` restated over votes cast.
- **Silent voters:** committing an implementation that selects nothing
  for `v` leaves `decideAt … v` unchanged, at any quorum.
- **Quorum 1 is 61's:** at `q = 1` the new `decideAt` equals the old.
- **Committed chains voted:** `wrote` appends a `startle` only when the
  stored chain selects `vals f`; a refused chain leaves the log and
  memory as they were.
- **`trust` is local:** the template `trust` runs, and each value it
  binds, is a function of the stored template and the frame. No commit
  at another key, and none at an entry, changes it.
- **An entry is a fold of the log** (62), and is read by no step but
  `borrow` and `ask`.
- **Effects read the goal closed:** a committed link whose act runs a
  tool outside `readOnly` selects only values whose goal is exactly its
  cited literals and holes, in order; two goals that differ in a word
  outside every hole do not both select.
- **No adjacent holes:** a committed closed reading has a literal
  between any two holes.
- **Refutations replay:** `Replays` extended to `refuted`; `step_replays`
  for both; `State.init` takes no refutations. `go_refuted` unchanged.
- **A borrow is a function of the log before it:** replaying a log to
  any prefix gives the borrow made at that point.
- **Agreement is symmetric and per position;** a pair agreeing at no
  antichain of size `span` is never borrowed for; `span = none` borrows
  nothing; raising `span` only withholds (62's obligation, restated).
- **The search is `holds`:** for every frame, string and `$n`, the
  search and `holds` agree; `introduce_refused` follows unchanged.
