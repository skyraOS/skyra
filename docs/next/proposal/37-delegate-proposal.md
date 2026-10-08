# 37 — Delegate: the entrypoint to remembered procedures

Date: 2026-10-06

Status: superseded by the [goal-based delegate contract](../../../ideas/delegate-procedure-spec.md),
implemented for Codex and Claude. The generated-tool design below is retained
as historical context.

Extracted from §4 of 36 (*Composition by
construction*, a draft kept uncommitted in the `codex/procedure-convergence`
worktree), with the decisions of 2026-10-06: `delegate` is how procedures are
created, a header is short by contract, and every invocation is its own
thread.

Builds on [32 — Procedures](../32-procedures.md) (`remember`, refinements,
`extend-reach`) and the built `system.narrow` (`core/machine/room.go`).
Independent of the rest of 36: it needs no Unix leaves, call limit or
Starlark.

## Why remembering needs delegate

`remember` compiles a continuation's control: its rounds, the calls each
round made, and the bindings between them. A leaf has no rounds, and
`Compile` refuses it (`interp/remember/remember.go:109`: *"ran as a leaf: it
called nothing, so there is nothing to compile"*). So a tool result one call
deep cannot become a procedure. When a model calls `read` directly, the
control lives in the model's own turn, interleaved with the rest of the
conversation, and the only continuation to compile is the whole conversation.
The unit is wrong.

`delegate` makes the right unit. It starts a fresh continuation with a
header and parameters; that child's calls are rounds of its own, so its
strand compiles. Every remembered procedure is therefore entered through a
delegation:

```text
model ── delegate {header, observations} ──▶ made tool <header>
           │                                     │ each call: a fresh child thread
           │                                     ▼
           │                               child's calls (rounds) ──▶ return
           ▼
       remember(return) ──▶ procedure, contract {title: header, input: observations}
                            installed by extend-reach in the remembering Horizon
```

## The header is a function name; observations are its parameters

```text
delegate  { header: string, observations: observation[], reach?: string[] }
          → makes tool <header>, runs it once, returns its result

<header>  { observations: observation[] }
          → { answer: observation, observations: observation[] }
```

- **The header** names the work at the granularity of the actor that
  delegates, for example *"Help me think about problem"*. It becomes the
  made tool's title, and the remembered procedure's contract title. Titles
  are labels, never part of fit (`core/contract/contract.go:24`).
- **The observations** are the arguments. A made tool accepts no text of
  its own, so parent and child never converse: all of the parent's intent
  travels in the header and the observations it passes.
- **A made tool is partial application.** `delegate` fixes the header, the
  reach and the initial observations. Calling the tool again runs a fresh
  child on the same header with more observations; a different header is a
  different delegation, and a different tool.
- **`reach` given:** `delegate` calls `system.narrow`, and the made tool is
  that reference with a header attached. Its children run under the
  narrower Horizon. Omitted: they run under the caller's Horizon.

### The header is capped

A long header lets the parent micromanage the child: the more it can say,
the more it prescribes how instead of what. Specificity belongs at the
level of the actor that delegates; the child decides the rest. So the
header has a configurable maximum, enforced at intake like any other fit:

- The cap is part of `delegate`'s input contract, so an over-long header is
  a `contract` fault the model sees and can correct, not a silent truncation.
- **It is counted deterministically: words, not provider tokens.** Fit runs
  in the pure machine and replay re-derives it, so it cannot depend on a
  provider tokenizer that changes between versions. This needs one new schema
  keyword, a string maximum (for example `maxWords`), in `core/contract`.
- The default and its configuration (registry, or per Horizon) are open
  below. Start small; the measurement in §Done when decides.

A short header also remembers better. A header that copies values from the
input (*"find the config that sets port 8080"*) compiles as a constant and is
wrong for port 3000. A header that names the work (*"find the server
config"*) with `port 8080` passed as an observation compiles as a constant
header and a bound parameter.

## One thread per invocation; the header is the prompt

Every call to a made tool starts a new Codex thread on the conversation's
running app-server. Its whole base prompt is:

```text
You are inside Skyra, a cognitive architecture, fulfilling: <header>
```

passed as `baseInstructions` on `thread/start`, which replaces Codex's
default prompt (that prompt describes `shell`, `apply_patch` and approval
modes this runtime never offers). The observations arrive as the first turn's
input. The child's tools are the dynamic tools derived from its Horizon, as
today. The prompt goes through `codex.exchange`, so it is in the lineage like
every other byte Codex is given.

Codex 0.160.0 accepts `baseInstructions` and `developerInstructions` on
`thread/start`, `thread/resume` and `thread/fork`; `turn/start` has no
instructions field, which is one reason the header belongs to the thread.

## From made tool to procedure

A made tool lives as long as the node that made it. Remembering keeps it:
`remember` on a made tool's accepted return compiles that child's strand, and
`extend-reach` installs the procedure under the made tool's contract, only in
the remembering Horizon (32). Later calls with fitting observations run the
procedure without the model while its refinements hold. When a refinement
refuses, the fallback is the original actor: a fresh child thread on the same
header, exactly what the made tool did the first time.

Parameters are bound by provenance: a child's leaf input that came from one
of its observations becomes a parameter; everything else is a constant.
Typed observation references (36 §2, [21](../21-observation-passing.md))
would make that binding explicit rather than inferred; they improve
`delegate` but are not required for it.

## Guarantees kept

| Guarantee | Why it holds |
|---|---|
| The emitter is never the checker | `delegate` is an ordinary invoke of an interpreter contract, checked at intake |
| Authority only narrows | `reach` goes through `system.narrow`, which refuses anything the caller does not hold |
| Self-extension adds no authority | `extend-reach` adds a contract to the caller's own Horizon, nothing else |
| Replay is byte for byte | the header cap is a deterministic schema check in the pure machine; the prompt is recorded |
| Hashes never reach a model | the child sees the header and observation contents, not stamps |

## Build order

1. A string-length keyword in `core/contract` (words), with fit and replay
   tests.
2. Per-invocation threads with the header as `baseInstructions` in the
   resident Codex adapter; the fake Codex requires the field.
3. The `delegate` procedure in the interpreter: makes the tool through
   `extend-reach`, runs it once, `system.narrow` when `reach` is given.
4. `remember` of a made tool's return, with the made tool's contract as the
   procedure's contract; refinement fallback to a fresh child.
5. The calls view draws delegations and made-tool runs, and `skyra metric`
   counts them.

## Done when

- A model with only `delegate` and leaves in reach delegates, its child's
  strand is remembered, and a later call with new observations runs the
  procedure with no model step.
- An over-long header is a `contract` fault, and replay re-derives it.
- A child's prompt in the lineage is exactly the header line.
- Measured on a task set: retries per made tool and header length, before a
  default cap is chosen.

## Open questions

1. **The default cap** and where it is configured: the registry, the
   Horizon, or the procedure actor's own contract.
2. **Two made tools with the same header** in one Horizon: titles are not
   dispatch keys, so they must differ in input contract or be refused.
3. **Remember automatically or on request:** whether an accepted made-tool
   return is remembered without the model asking.
4. **Depth and width:** a child may delegate in turn; measure before capping.
