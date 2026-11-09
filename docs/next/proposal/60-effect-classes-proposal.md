# 60 — Effect classes: what an instruction does sets how sure the tally must be

Date: 2026-11-07

Status: **stale** (2026-11-09): written against 58's contracts, which
[61](61-static-and-dynamic-proposal.md) replaces; not to be built as is.
Builds on [59](59-tally-proposal.md) and answers its
open questions 2 (are two voters independent?) and 3 (where `ι` lives).
Uses the lineage of [53](53-capabilities-and-information-flow-proposal.md).
Everything else in 59 stands: voters, ballots, levels, `required`, finest
first.

Every tool's contract carries an **effect class**, set by the operator at
registration. The class of the instruction a tally is about to run sets
three things: the most `ι` it may use, the coarsest level it may merge at,
and a **quorum**, the fewest *independent* voters that must back it. A
voter is independent of another when it was learned from a different
situation or a different model. A voter whose lineage the tool's integrity
allowlist refuses does not vote for that tool. A tool with no class is the
strictest class. The compiler picks the caution, not whoever writes `ι`.

## Why

59 has one `ι` and one floor for every base. Three things go wrong with
that.

**One voter wins at every `ι`.** `required(1) = max(1, ⌈(1 − ι) · 1⌉) = 1`
for every `ι`, and a lone vote has margin 1. The first chain ever committed
at a base therefore runs with no model call even at `ι = 0`. Unanimity of
one is still unanimity. For `curl` that is the point of learning. For
`systemctl restart`, one model call becomes a habit at once, and no setting
of `ι` prevents it.

**Seven samples are not seven opinions.** The `cel/` experiment's seven
voters are seven `claude -p` processes: one model, one prompt, one input.
Knight and Leveson (1986) ran independently written versions of one
program and found they failed together on the same inputs far more often
than independence predicts. The Condorcet jury theorem, the reason a
majority is worth more than one voter, assumes independence. Without it,
more votes add confidence and not accuracy. A 6–1 between samples is a good
measure of how sure *one model* is (self-consistency; Wang et al., 2022).
It says nothing about whether a second situation would agree.

**A floor of 2 merges what must not be merged.** At level 2, `run
systemctl $6.unit $7` and `run systemctl $6.unit $8` are one group, and the
exact instruction with the most votes runs. For a read that is fine. For a
tool where one argument means `stop` and another means `restart`, the
group's agreement covers votes that wanted something else.

All three are the same mistake: one caution for instructions with
different consequences.

## Polarity

As 59 defines it, **high `ι` is permissive** (`ι = 1`: a lead of 1 wins)
and **low `ι` is cautious** (`ι = 0`: unanimity). This proposal keeps
that and writes it everywhere the same way. A class *caps* `ι`; it never
raises it. "Prod writes cap `ι` at 0.25" means: never decide those on less
than near-unanimity.

## Definitions

**Effect class.** One of an ordered set, least to most caution:

| class | means | examples |
|---|---|---|
| `none` | changes nothing in the world | `return`, a pure transform |
| `read` | reads, changes nothing | `curl` GET, `journalctl`, `cat` |
| `write` | changes local state, undoable | writing in the workspace |
| `external` | changes state outside the workspace, undoable | restarting a service, posting a draft |
| `irreversible` | cannot be undone | deleting, sending mail, moving money |

**Where a class is set.** On the contract, at registration, by the
operator: the same place and the same owner as 53's allowlists and as
binary pins. No syscall sets or lowers one, and the model never sees it. A
class is part of the contract's identity, as its output declaration is
(58, Shapes): changing it is a different contract, so its outputs land in
different bases and nothing learned under the old class carries over.

**An instruction's class.**

| form | class |
|---|---|
| run | the tool's contract class |
| return | `none`: a return acts on nothing. What the parent does with it is tallied where it acts |
| delegate | the most cautious class among the child's tool types |

A delegation's class is finite and static. The child holds strictly fewer
tool types than its emitter (58, `child_fewer`), all from the emitter's
scope, so the class is a maximum over a set known when the instruction is
admitted.

**Instructions are made only of parameters.** A tool's literal arguments
come from observations, so the class cannot depend on them. A tool whose
subcommands differ in consequence (`systemctl status` and `systemctl
restart`) is registered as two contracts, each with its own binary pin and
class.

**Unknown is irreversible.** A contract registered without a class is
`irreversible`. The cautious path costs model calls; the permissive one
costs whatever the tool does. This fails closed, as 53's allowlists do.

**Forbidden is not a class.** What a position may not invoke at all is a
capability (53): it is not in the scope, and `sense` rejects it. The tally
only ever sees instructions a position may run.

## Settings

Each class has three settings:

| class | `ι` cap | floor | quorum `q` |
|---|---|---|---|
| `none` | 1 | 2 | 1 |
| `read` | 1 | 2 | 1 |
| `write` | 0.5 | 3 | 2 |
| `external` | 0.25 | 4 | 2 |
| `irreversible` | 0 | 4 | 3 |

- **`ι` cap.** The effective `ι` is `min(ι_root, cap(class))`. A root
  can be more cautious than its classes, never less.
- **Floor.** The coarsest level a group may win at. Floor 4 means exact
  instructions only: no merging across arguments.
- **Quorum.** The fewest independent, admitted voters (below) that must
  back the winning group.

The table is the default, kept with the registration configuration. An
operator may make a class stricter, never looser than the class below it,
so the order of the classes stays the order of their caution.

**Settings are the candidate's, not the base's.** One base can see a
`return`, a `curl` and a `systemctl` on one ballot. The bar is set by what
would run. Running a read in place of a restart risks being wrong, not
damage, and being wrong is what the next arrival's tally and `startle`
handle. At level 2 and finer, a group has one form and one head, so one
class: the floor of 2 from 59 is what makes "the group's class" defined.

## Independence

**A voter's source** is the pair `(model, focus)`: the pinned model
descriptor that wrote its chain (47), and the hash of what that `startle`
was shown (53, Need-to-know). Two voters with one source are one opinion
asked twice.

**Two counts.**

| count | counts | used for |
|---|---|---|
| votes | every voter, as 59 | the margin and `required` |
| sources | distinct sources among the winning group's voters | the quorum |

The margin keeps every vote, because agreement between samples is real
information about how sure one model is. The quorum counts sources, because
only a second situation, or a second model, tests what one model run could
not.

In practice sources accrue across arrivals: different values reach one
base, and each `startle` there sees a different focus. A path backed by
three sources is one that three different situations each led the model to
write.

**A `startle` from a source that already voted adds a vote and no source.**
Asking again at the same values cannot manufacture a quorum.

## Admitted voters

53's integrity rule says what may shape what enters a sink: a sink's
contract lists the sources a value's lineage may have. A voter's chain is
shaped by its focus, so its lineage is its focus's leaves.

**A voter votes for a tool only if the tool's integrity allowlist admits
the voter's lineage.** Otherwise it abstains on that ballot: it is not
cast, and counts towards neither margin nor quorum. A chain the model wrote
while a web page was in focus cannot be the vote that runs
`bank.transfer`, whatever it agrees with.

This is the Byzantine case. Prompt injection does not have to defeat a
guard. It has to get one chain committed. Under 59 at `ι = 1` with few
voters, one committed chain can win on plurality. Here it cannot vote at a
sink whose allowlist excludes where it came from, and where no allowlist
exists, a single poisoned source still cannot meet a quorum above 1.

Like 53, this is opt-in: a tool with no integrity allowlist admits every
voter, and the check walks a lineage only where a policy exists.

## Tally with classes

59's tally, with the settings taken from the leader:

```text
tally(branches, v, ι_root):
  ballot ← the instruction of each admitted voter selecting v
  if ballot is empty: null                                   → startle, as 58
  for k from 4 (exact) down to 2:
    group the ballot at level k; leader L, runner-up by votes
    s ← settings(class(L))
    if k < s.floor: continue
    ι ← min(ι_root, s.cap)
    if votes(L) − votes(runner-up) ≥ required(|ballot|, ι)
       and sources(L) ≥ s.quorum:
      return one: the exact instruction with the most votes in L,
                  ties to the lowest rank among its voters
  conflict                                                   → startle
```

- **Skipping a level is not deciding.** A leader below its floor skips
  that level, and the tally goes on to coarser levels, where the leader
  may be a different group of a different class. A strict instruction
  itself can only win at or above its floor: a coarser group containing it
  keeps its head, so its class, so its floor.
- **Short of a quorum is a conflict, and a conflict startles.** The
  model's chain is one more voter. If its source is new, it may complete
  the quorum, and the tally decides on the same step. If not, the step runs
  the model's own instruction once, unlearned at these values, as 58 does
  when another branch is selected after `startle`. A strict path keeps the
  model in the loop on every arrival until `q` situations have agreed.
- **59 is the case where every class is `none`.** Cap 1, floor 2,
  quorum 1, every voter admitted: the tally above is 59's.

### The experiment under classes

Registering `curl` and `journalctl` as `read` and `systemctl` as
`external`, and taking `ι_root = 1`:

| values | ballot | leader's class | sources | result |
|---|---|---|---|---|
| 200 | 7 × `return 0 $5` | `none` | 1 | `return 0 $5` |
| 503, restarts 3 | 6 × `return 1 $5 $6`, 1 × `return 1 $5` | `none` | 1 | `return 1 $5 $6` |
| 503, restarts 1 | 4 × `return 1 $5 $6`, 2 × `$2 $6.url`, 1 × `return 1 $5` | `none` | 1 | `return 1 $5 $6` (margin 2) |

Nothing changes, because every leader is a return, and that is right: no
row runs anything. All seven voters share one source, so the quorum is 1
whatever the margin.

Had the restarts-1 row instead been 5 × `$3 restart-args` (a `systemctl`
run) against 2 returns, 59 at `ι = 1` would restart the service on the
first arrival. Here the leader is `external`: cap 0.25 needs a margin of 6
of 7, and the quorum needs 2 sources. It startles. The service is restarted
without the model only after a second, different situation has led a model
to the same exact instruction, and on near-unanimous ballots.

## What changes from 59

- **Open question 3 is answered.** `ι` is per root, and each class caps
  it. 59's `Bounds.Inhibition` becomes `ι_root`, and 59's single floor
  becomes the floor of each class, never below 2.
- **Open question 2 is answered** by the two counts. The margin is
  unchanged. The quorum is new.
- **The ballot is filtered** by integrity before it is counted.
- **The conflict record gains** the leader's class, its settings, and its
  sources, so a conflict says *which* bar it missed: margin, floor or
  quorum.
- **Nothing in commit changes.** A voter stores its source with its
  chain. Classes live on contracts and are read at tally time.

## Required Lean definitions and proofs

In `invariants/Invariants/Refine.lean`, after 59's `tally`, with the class
and the source abstract (an order and a key), as the instruction's parts
are:

- `Class`, a linear order; `Settings` with `cap`, `floor`, `quorum`; and
  `strict : Class → Class → Prop` (higher in the order, cap no greater,
  floor and quorum no lower).
- **One class per group:** at every level `k ≥ 2`, every vote in a group
  has one class.
- **Quorum holds:** if the tally returns `i`, the voters in `i`'s group at
  the deciding level have at least `quorum(class(i))` distinct sources.
- **No lone habit:** if `quorum(class(i)) ≥ 2`, a ballot whose votes all
  share one source never returns `i`. This closes 59's `required(1) = 1`
  case.
- **Resampling is inert:** adding a voter with an existing source never
  increases `sources` for any group.
- **Admission:** every vote counted for a winner is admitted by the
  winner's integrity allowlist.
- **Stricter decides less:** making one class's settings stricter never
  makes the tally decide a ballot it did not decide before, for ballots
  whose leader at every level is of that class. (59's caveat stands:
  deciding is monotone, *what* is decided across levels is not.)
- **59 is a special case:** with every class `none` and every voter
  admitted, `tally` here equals 59's `tally`.
- **Unknown is strictest:** a contract with no class is `irreversible`.
- **Delegation:** a delegation's class is at least the class of every
  tool type in the child's scope.

Vectors: `tallyClasses`, over generated ballots crossed with classes,
sources and admission, including both experiment tables above.

## Go

- Registration: `Contract.Class`, default `Irreversible`, part of the
  contract hash.
- `tree/refine.go`: `Branch.Source` (model descriptor hash, focus hash);
  `Tally(bs, vals, ιRoot, classes, admit)` returns `(Selection, Counts)`,
  with counts per group of votes and sources.
- `tree/dispatch.go`: `Bounds.Inhibition` becomes the root `ι`; a
  `Settings` table per class, loaded with the registration configuration.
- The `startle` path: after a conflict, a new voter whose source already
  voted is committed as a vote and the step runs its own instruction once.
- The admission check reuses 53's integrity walk over the voter's focus
  leaves, only where the tool has an allowlist.

## Open questions

1. **Are the classes the right ones?** `write` and `external` split on
   *where* a change lands; 53 splits sinks by allowlist. A class may be
   better derived from a contract's capabilities (what its sandbox profile
   can reach) than declared beside them. Declared is simpler and auditable,
   derived cannot drift.
2. **What makes a focus different?** Hashing the whole focus makes any
   changed byte a new source, such as a timestamp. A source may need to be
   the focus's base types and the values its guards read, not its bytes.
3. **Cross-model sources.** A second model is a stronger second opinion
   than a second situation for one model. Should a quorum for
   `irreversible` require two models, not just two sources?
4. **How many poisoned voters to tolerate.** Admission removes voters
   whose lineage is refused. Where no allowlist exists, `f` committed
   poisoned sources defeat a quorum of `f`. Classical quorum systems need
   `n ≥ 3f + 1` to tolerate `f` Byzantine members. Is `q(irreversible) = 3`
   meant to tolerate one?
5. **Choosing settings from a target error rate.** Selective prediction
   (Chow, 1970) sets an abstention threshold from the error rate it may
   accept. With enough recorded ballots and their outcomes, each class's
   `ι` cap could be computed from a target error rate per class instead of
   chosen.
6. **Decay.** A habit backed by `q` sources stays backed. Should sources
   age, so a strict path is re-confirmed after the world may have changed?

## Checks and measurements

- Re-run `cel/` with goals whose leaders are runs of each class, and
  record per class how often the tally decides, at which level, and which
  bar a conflict missed.
- Count model calls per arrival on a strict path until its quorum is met,
  against 59: the cost of caution, per class.
- Measure correlated failure: for paths backed by `q` sources, how often a
  later arrival's model disagrees, against paths backed by one source
  sampled `q` times. If they do not differ, the source definition (open
  question 2) is not capturing independence.
- Inject a poisoned chain at a sink with and without an allowlist, and
  confirm it is never the deciding vote where the allowlist exists.
