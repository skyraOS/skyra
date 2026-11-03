# 54 — One tool per model: an experiment in structural composability

Date: 2026-11-03

Status: **experiment, for building.** Nothing is built. Code starts
2026-11-04.

**Scope:** a harness rule that makes work fall into composable units by
mechanism, not by prompts or caps, and the experiment that tests it. Security
stays in the background: [53](53-capabilities-and-information-flow-proposal.md)
is the last line of defense, so this proposal does not narrow anything.

## The problem

Agents are greedy. A node that holds every tool does everything itself, and
the run is one tangled unit instead of reusable parts.

Caps do not fix it. A limit on actions or children per node does not limit
what can be expressed, since delegation is unlimited, but it puts the cuts
where the number runs out. A greedy agent does k things and hands off
"continue".

Delegation and carrying context forward are the same thing: a step emits a
value and another step receives it. So the question is not whether work is
delegated, but how wide each edge is. A greedy run is a chain whose every edge
carries everything. A composable run has large subtrees joined to the rest by
small values.

## The rule

> A model may act only when its focus holds at most one tool. Holding more
> than one, its only moves are to **delegate**, giving each child a subset of
> its tools and a set of observations, or to **return**.

There is no number in the rule beyond the one. Zero and one tool are the same
kind of node: with zero, a model returns what it wrote (`return "some
string"`); with one, it may also return what its tool gave back (`return
tostring(someObject)`). The rule changes what each kind of node is:

- **A leaf holds zero or one tool.**
  - **Zero tools: content.** It writes a value (a summary, a file's text, an
    answer) from the observations it was handed, and returns it.
  - **One tool: an effect.** Its interface is fixed by construction: one
    tool's contract in, one value out. It may call its tool many times.
- **An interior node holds two or more, and is pure control.** It partitions
  tools, passes observations, and combines returns. It never touches the
  world and writes no content of its own beyond its return.

The tree therefore splits three ways: control (interior nodes), content
(zero-tool leaves) and effects (one-tool leaves). Interior nodes are what
`remember` compiles into procedures; content leaves stay model calls; effect
leaves are the calls procedures make.

## Observation passing

Edges pass **references**, named by the sender. A delegate names the recorded
observations its child receives, and the child's focus is assembled from
exactly those. Passing everything is possible, but it is an explicit list of
names, and it is recorded. Edge width is the number and size of the references
passed.

What an interior node sees of an observation is the experiment's main
variable:

| Setting | An interior node sees | Expected |
|---|---|---|
| **A: contents** | the bytes of every observation its children return | better-informed splits; interior focus grows, and greed may return at the interior level |
| **B: presentations** | a reference and its presentation (`obs-3: a 412-line Go file`, `test: 2 failures`), never the bytes | tiny interior focus; untrusted content enters only leaves |

In both, a leaf may return a small typed value beside a reference to a large
observation, so an interior node can decide on the small value and pass the
large one on unread.

Setting B makes every interior node a planner that never reads untrusted
content (CaMeL's privileged model). That is a security property in the
background, not a goal of this experiment.

## What it may do wrong

1. **Peeling.** An interior node hands one tool to a leaf and all the rest to a
   single "continue" child: a chain of depth n, the caps failure in another
   shape. Signature: a chain of interior nodes whose tool sets shrink by one
   and whose edges carry every reference.
2. **Fat interiors.** One interior node runs many rounds of delegation. Fine if
   its trace is pure control, since it compiles; a problem in setting A if its
   focus grows like a session.
3. **Iteration across tools.** Edit → test → edit crosses two tools, so the
   loop must live in an interior node. Whether it is factored or stays one long
   interior is likely the most telling result.
4. **Cost.** Every interior node is a model turn. Structure is paid for in
   turns and wall time, and the experiment records both.

## Mechanism

Most of it exists.

| Piece | Where | Change |
|---|---|---|
| The rule | `codex/proceduretools`, `claude/resident` (the tool offering) | more than one tool in focus: offer only `delegate` and `return`; one: offer that tool and `return`; none: offer only `return` |
| Delegation with a subset | apprentice `delegate` (`tools`, `contracts`, `fresh: true`) | add `observations: [names]`; the child's focus is assembled from them (see "Observation types") |
| Focus assembly | the adapter, from the delegate's references | setting A: contents; setting B: presentations, from `apprentice/types` |
| Leaf return | the leaf's contract | a small typed value, and optionally a reference to a large observation |
| A fresh session per node | `fresh: true` | every node, interior and leaf, starts with no session history |
| The trees | `.skyra/` records | read as recorded; nothing new is stored for the experiment |

The experiment leaves narrowing alone. A child's tool subset is its focus, not
a capability; it runs under its parent's Boundary.

### Observation types: the bare minimum

Observation passing needs a small piece of the type system. **Do not go
further than this.** The goal is to see tree shapes, not to build the type
system; anything not listed here waits until the trees say it is needed.

An observation, as the experiment knows it:

| Field | What it is |
|---|---|
| `name` | a short name local to the node: `obs-1`, `obs-2`, … What the model reads and writes. Never a hash |
| `stamp` | the record stamp it came from. Kept by the adapter, never shown to the model |
| `shown` | its presentation, from `apprentice/types.Present` as it exists today: what setting B shows instead of the bytes |
| `bytes` | the value itself: what setting A shows, and what a leaf is handed |

- **Naming.** Each tool return, and each child's return, becomes an
  observation with the next free name in that node. The adapter keeps the
  table from names to stamps for the node it serves.
- **Passing.** `delegate` takes `observations: ["obs-2", "obs-5"]`. The adapter
  turns the names into stamps; the child's adapter gives them new names
  (`obs-1`, `obs-2`) in the child's own table. An unknown name is a refusal of
  the delegate, shown to the model.
- **Returning.** A leaf returns `{"value": <JSON>, "observations": [names]}`:
  a small value the parent can decide on, and the observations it hands back
  unread. Nothing checks the value's shape.

Out of scope for the experiment: lineage and labels (53), shape hashes and
type equality, return contracts, CEL over observations, checking a leaf's
value, and storing anything beyond what the records already hold.

## The task set

Tasks this repository can check, each run in both settings, several times per
model:

1. **Read-only:** answer a question about `controller/core` that needs several
   files.
2. **Read then write:** summarize a package into a new file in scratch.
3. **Edit and test:** fix a seeded bug in a small package and make its tests
   pass (the edit → test loop).
4. **Search and act:** find every caller of a function and record them.
5. **A repeated task:** run task 2 on several packages in a row, to see
   whether what was remembered from the first is reused.

Real Codex and Claude, not the fake app-server: the fake makes the choices
itself.

## Measures

All read from the records after the run.

| Measure | What it shows |
|---|---|
| depth, and fan-out per interior node | the tree's shape |
| the tool partition at each split | how the model divides work |
| edge width: references per edge, and their total bytes | how much context each cut carries |
| rounds per interior node | fat interiors |
| peeling signature count | whether greed survives the rule |
| interior focus size (setting A vs B) | what the visibility setting costs or saves |
| model turns and wall time per task | the price of the structure |
| content leaves vs effect leaves per task | where the model writes, and where it acts |
| subtrees `remember` compiles, and calls to them on later tasks with no model step | whether the cuts are reusable |
| task success | whether the structure still gets the work done |

Baseline: the same tasks with every tool offered to one model, as today.

## Decisions to make

1. **How many observations an interior node may name.** Recommendation:
   unlimited. Width is measured, not capped.
2. **Leaf return shape.** Settled for the experiment: `{"value", "observations"}`,
   with `value` required so setting B has something to decide on, and its
   shape not checked. Typing it is for later.
3. **Depth bound.** Keep `maxDepth` (64) as a safety bound only.

## Later

- Finding cuts mechanically: weight the recorded graph by bytes and take
  narrow waists (dominators, minimum cuts) as candidate modules, then confirm
  them with CEL route equivalence and ablation of single named references.
- The caller states the return contract when it delegates, so the interface
  is fixed before the child runs.
- Running n models per node to widen the distribution, and choosing CEL by
  running every candidate against all traces.
- Narrowing each child to its tool subset once the trees are understood, so
  the focus becomes a capability (53).
