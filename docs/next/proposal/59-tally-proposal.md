# 59 — Deciding by tally

Date: 2026-11-07

Status: **proposed**. Builds on
[58](58-type-and-transition-machine-proposal.md) and replaces its rule for
a conflict, *the older branch keeps the overlap* (58, Reconciling a
conflict; Conflicts found later; At commit). Everything else in 58 stands:
bases, versions, effective conditions, `startle`, `trust`, both bounds.

A refinement holds procedures, each a chain the model wrote at one
`startle`. Each procedure selects at most one branch for any values, so
it is one vote. When values arrive, the core counts the votes by
instruction, at the finest grouping first and coarser after. An
instruction wins when its lead over the runner-up clears a margin set by
one constant, the **inhibition**. With no winner the core startles, and
the model's guard is one more vote. Age only breaks ties.

## Why

The experiment in [`cel/`](../../../cel/INPUT.md) asked Opus 5.5 seven
times, with no tools, for one act frame's guard: the goal "make sure the
service is healthy, and fix it if it isn't", the tools curl, systemctl and
journalctl, and a 503 with `{"restarts":3, "unit", "url"}`. Each guard was
then evaluated on the values it was asked about and on two it never saw:

| values | distinct answers | split |
|---|---|---|
| 200 | 1 | 7 × `return 0 $5` |
| 503, restarts 3 (asked) | 2 | 6 × `return 1 $5 $6`, 1 × `return 1 $5` |
| 503, restarts 1 | 3 | 4 × `return 1 $5 $6`, 2 × `$2 $6.url`, 1 × `return 1 $5` |

Under 58, the 503 rows are conflicts, and the oldest branch takes the
overlap. Which branch is oldest depends on commit order, and nothing
about the guards themselves. Six of seven agree on the case asked about;
age can still choose the one. The split itself is information: a 7–0, a
6–1 and a 4–2–1 call for different confidence, and one rule for all three
throws that away.

## Definitions

**Voter.** A procedure: the chain one `startle` committed. Every branch
carries its voter, the local id its chain was committed under, alongside
its own local id and rank (58, Refinements are versioned). A voter is
local to its base, as a branch is.

**One voter, one vote.** A chain's branches have effective conditions:
each conjoins its own condition with the negation of every earlier one in
its chain. So for any values at most one branch of a voter selects. Each
voter casts one vote, its instruction, or abstains.

**Ballot.** For values `v`, the votes cast: each voter whose branch
selects `v`, with that branch's instruction.

**Grouping levels.** A vote's instruction is canonical (only its form and
its `$n`, in order), so it parses into parts, compared coarsest first:

| level | key | example |
|---|---|---|
| 1 form | run, return or delegate | continue or finish |
| 2 + head | the tool run; a return's exit status; the delegated goal | `return 1` |
| 3 + first | the first argument; a return's stdout | `return 1 $5` |
| 4 exact | the whole canonical instruction | `return 1 $5 $6` |

Two votes are in one group at level `k` when their keys agree to `k`. The
**floor** is the coarsest level a tally may use. It is 2 by default:
continuing is never merged with finishing, one tool with another, or
exit 0 with exit 1.

**Margin.** With `cast` votes, the leader's count minus the runner-up's.
The runner-up's is 0 when the leader is alone.

**Inhibition** `ι ∈ [0, 1]`. The margin a winner needs is

```text
required(cast) = max(1, ⌈(1 − ι) · cast⌉)
```

As in lateral inhibition, a high `ι` lets a slim lead suppress the rest:
`ι = 1` needs a margin of 1, a strict plurality. A low `ι` needs broad
agreement: `ι = 0` needs a margin of every vote cast, unanimity. The
margin is never below 1, so a tie never wins, and at most one group wins
at a level.

## Tally

```text
tally(branches, v, ι, floor):
  ballot ← the instruction of each voter selecting v
  if ballot is empty: null                       → startle, as 58
  for k from 4 (exact) down to floor:
    group the ballot at level k; leader, runner-up by count
    if count(leader) − count(runner-up) ≥ required(|ballot|):
      return one: the exact instruction with the most votes in leader,
                  ties to the lowest rank among its voters
  conflict                                       → startle
```

- **Finest first.** A tally that decides at level 4 never coarsens. Only
  when exact answers split does it look at which differ only in detail.
- **A group runs one instruction.** A group at level 2, such as
  `return 1`, is not an instruction. The tally runs the exact instruction
  in it with the most votes. Age decides only a tie, and only inside the
  winning group.
- **No winner is a conflict, and conflict startles.** The model writes a
  guard without seeing the branches (58). Its chain is one more voter, and
  the core tallies again. The refinement bound limits the asks, so a
  deadlock ends as an exhausted frame, `{error: bytes}`, as any exhausted
  frame does.
- **Determinism.** The tally depends on the branches as a set and on the
  values, never on evaluation order or on which commit landed first,
  except where age breaks a tie.

`distinguish` (58) is the tally at `ι = 0`, exact level only: one
instruction when every vote agrees, otherwise a conflict. Today's
behaviour is one setting of this one.

### The experiment under the tally

`cast = 7`, so `required` is 1, 2, 4, 6 and 7 at `ι` = 1, 0.75, 0.5,
0.25 and 0. Margins at each level (floor 2):

| values | level 4 (exact) | level 3 | level 2 |
|---|---|---|---|
| 200 | 7 − 0 = 7 | 7 | 7 |
| 503, restarts 3 | 6 − 1 = 5 | 7 (`return 1 $5`) | 7 (`return 1`) |
| 503, restarts 1 | 4 − 2 = 2 | 5 − 2 = 3 | 5 − 2 = 3 |

| `ι` | 200 | 503, restarts 3 | 503, restarts 1 |
|---|---|---|---|
| 1 | `return 0 $5` | `return 1 $5 $6` | `return 1 $5 $6` |
| 0.75 | `return 0 $5` | `return 1 $5 $6` | `return 1 $5 $6` (margin 2) |
| 0.5 | `return 0 $5` | `return 1 $5 $6` | startle |
| 0.25 | `return 0 $5` | `return 1 $5 $6`, at level 3 | startle |
| 0 | `return 0 $5` | `return 1 $5 $6`, at level 3 | startle |

At `ι = 0` the agent that gave no why does not block the six who did: they
merge at level 3, where all seven return 1 with `$5`. Nor does it change
which instruction runs: the exact instruction with the most votes in the
group does. The restarts-1 row is the real split, five finishing against
two curling, and only a high inhibition decides it.

## What changes from 58

**Commit no longer narrows.** Under 58, a new branch that decidably
overlaps an older one with a different instruction is retired at commit,
and only its remainder stored. That deletes the dissenting vote before it
is ever counted. Under the tally every committed chain is stored whole, with
its overlaps. Retirement, remainders and *possibly overlapping* records
are no longer needed for correctness. `Overlaps` may stay as a diagnostic,
reporting at commit where voters disagree before any value lands there.

**`resolve` is replaced by `tally`.** No branch is retired at evaluation
either. A version changes only when a `startle` commits a new voter.

**After `startle`** (58, `afterStartle`) is unchanged in what runs: the
step runs the instruction its own model wrote, once. It is *learned* when
the tally over the committed version selects that same instruction.

**Duplicates become weight.** Today two commits of the same chain store
its branches twice (an earlier batch of the experiment committed 16
branches, five of them exact copies of an earlier branch). Under the tally that is two
votes, which is correct, but it need not be two copies: a branch can be
stored once with the set of its voters. Evaluating a base costs one CEL
evaluation per distinct branch, not per voter.

**What the core keeps.** Per branch: its voter. Per base: `ι` and the
floor, or they are global (open question 3). No count is stored: a tally
is computed from the branches and the values at each step, as
`distinguish` is today.

## Required Lean definitions and proofs

In `invariants/Invariants/Refine.lean`, before the Go:

- `Branch` gains `voter`; `tally` over the selected branches, the levels
  and `required`, with the instruction's parts as an abstract key function
  so the Lean stays generic over instructions, as it is now.
- **One voter, one vote:** for any chain and values, at most one of the
  chain's branches selects them (`chainBranches` with effective
  conditions).
- **Order:** `tally` is invariant under any permutation of the branches.
- **Uniqueness:** at most one group clears `required` at any level.
- **Unanimity:** when every vote cast has one exact instruction, it wins
  at every `ι`.
- **Deciding is monotone in `ι`:** if the tally decides at `ι`, it decides
  at every `ι' ≥ ι`. *What* it decides is not: votes `A₁ 3, A₂ 3, B 4`,
  with `A₁` and `A₂` one group at level 3, give `B` at `ι = 1` (level 4,
  margin 1) and `A₁` or `A₂` where a margin of 2 is needed (level 3, 6 − 4).
  Raising `ι` can let a finer level decide first. The proof states the
  weaker property; whether the finer level should then win is open
  question 6.
- **The floor holds:** every vote counted for a winning group agrees with
  the instruction run on every part down to the floor.
- **`distinguish` is a tally:** at `ι = 0` and the exact level only, it
  equals `distinguish`.
- **Witness kept:** a value that was `null` before `startle` and is
  covered by the new chain is no longer `null`; with no other vote it
  selects the new instruction.

The proofs specific to *older keeps the overlap* go: `narrow`,
`narrow_older`, `remainder*`, `reconcile*`, `conflictPair*`, `resolve`,
`resolve_settles`, and the vectors `refineCommit`, `refineResolve` and
`refineOverlaps` (the last kept only if the diagnostic stays). New
vectors: `tally`, over generated ballots and every level and `ι` in the
table above.

## Go

- `tree/refine.go`: `Branch.Voter`; `CommitChain` stores the chain whole;
  `Tally(bs, vals, ι, floor) (Selection, Counts)`; `Distinguish` stays,
  as the special case.
- `tree/frames.go`: the instruction's parts, from the parsed `Instr`, for
  the grouping key.
- `tree/dispatch.go`: `decide` tallies; `Bounds` gains `Inhibition` and
  `Floor`.
- The conflict record (58, Locating a conflict) gains the ballot: each
  group with its count, at the level the tally stopped.

## Open questions

1. **What a cast vote is.** The margin counts votes cast, so an abstaining
   voter does not count against the leader. Counting every voter in the
   base would make an abstention a vote against, so a refinement that has
   learned much elsewhere would decide less here.
2. **Are two voters independent?** Seven asks of one model at one base are
   seven voters here. If one model, asked twice with the same scope, is one
   opinion repeated, voters may need a weight, or identity by model and
   input, so that asking again cannot manufacture a majority.
3. **Where `ι` lives.** One per root, per base, or per frame kind? A base
   whose instructions touch the world (`systemctl restart`) may want a
   lower `ι` than one that only reads.
4. **The order inside level 2.** A run's head is its tool; a delegation's
   is its goal. Is a delegation's goal closer to its form or to its
   arguments?
5. **Startle on conflict, or ask with the ballot.** The model is never
   shown branches (58). A conflict might be the one place it should see
   the split, to break it rather than add to it.
6. **Finer first, or largest margin first.** Finest-first lets a narrow
   exact plurality beat a group that agrees more broadly (the `A₁ A₂ B`
   case under Required Lean). The alternative tallies every level down to
   the floor and takes the largest margin relative to `required`, finer
   on a tie. It is monotone in what it decides, and costs the same.

## Checks and measurements

- Re-run the `cel/` experiment with more goals and record each ballot:
  how often does each `ι` decide, and at which level?
- Count startles saved against 58's age rule on the same commits, and
  where the two pick different instructions.
- Measure tally cost per step against `distinguish`: it evaluates the same
  branches, and the grouping is over the votes cast.
