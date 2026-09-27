# Skyra: where it is and where it's going

Date: 2026-09-27  
Status: a snapshot summary of the whole repo, including the uncommitted
reframing of `README.md`, `grips-and-invariants.md` and
`next/research-question.md` around approval fatigue and self-extension.

## What it is

Skyra is an agent runtime. Its one-line pitch: **the human approves the
boundary once, not every tool, and the OS enforces that boundary.** The unit of
trust is a "bounded, auditable agent" rather than the process. Its authority is
one list, the **Horizon**. A child's Horizon is always a subset of its
parent's, and each tool call runs under a macOS Seatbelt sandbox generated from
that Horizon.

## The "why" ([`grips-and-invariants.md`](grips-and-invariants.md))

The doc is written from firsthand experience with Codex and Claude Code. It
names six pains ("grips") and the three invariants they force:

| Grip | Invariant |
|---|---|
| 1. Skills are markdown re-read every turn · 2. The same work is paid for again and again | **I. Cognition is compiled and remembered**: finished work is kept and reused, not re-derived |
| 3. Plan mode is a checklist the model can silently skip | **II. Every action is a typed claim; deviating from it is a fault** |
| 4. Shell access is ambient authority · 5. Five agents can't be watched by hand · 6. Approval-prompt fatigue *(new)* | **III. Authority is one list that only narrows, the OS enforces it, and provenance comes from the structure** |

Its key framing: the contribution is **coupling, not invention**. Permissions,
plans and audit logs all exist already, but today they sit beside execution.
Skyra makes them part of the graph that actually runs. The doc also marks which
parts are driven by real pain (the single-machine core, which is what ships)
and which are a bet (distributing across machines, which "must not front-run
the core").

**Status by invariant:** III is built. II is built for each single action;
negotiating the plan with the human up front is not. I is designed and
deliberately deferred.

## Where it is now (built, tested, running real Codex)

The **stateless controller** ([`controller/`](controller/), about 4.3k lines of
Go including tests, standard library only, macOS):

- There is no server. One `skyra` binary runs once per step and exits. Every
  decision goes through one pure function, `machine.Step`. All I/O happens in
  `instance.Drive`, which decides nothing.
- **Operators are derived, not stated** (thought 14). The model only supplies
  operands, `{"to": capability, "payload": …}`. `interpret` derives whether
  that is an invoke, a return, or a typed fault (`malformed` / `horizon` /
  `contract` / `custody`). Returning a result means invoking a single-use
  `reply` capability.
- **Custody and stamps:** each record is content-addressed and hash-chained in
  `.skyra/`. A parent exits only after its successor has stamped (committed).
  Replay re-derives every record byte for byte from genesis.
- **Dependencies and resolve:** a continuation can invoke, wait on a slot and
  resume. The join rule is enforced by hashing, and a fan-in claim picks
  exactly one resumer.
- **Adapters:** a long-lived "dumb" actor host sits in front of Codex, and a
  stateless adapter step owns all Codex-specific behavior. Codex's tool calls
  become sandboxed leaf tools. If Codex tries its own shell, that is detected
  and turned into a horizon fault.
- **Ingress and conversations:** a browser chat waits by "parking" its frame
  with the webhost. The user side is Codex's parent, so the user outranks the
  agent. Conversations are multi-turn, saved as their record history, and
  replayable. A fault is shown and the conversation continues.

The project's history: seven months of solo work. DNA-19
([`history/dna19/`](history/dna19/), about 7.4k lines) was the long-running
predecessor. It proved the approach against the real Codex app-server, but its
authority lived in Go pointers and it trusted Codex's own sandbox report. That
flaw is why the stateless controller exists.

## Where it's going

**Research question** ([`next/research-question.md`](next/research-question.md)):
can human approval move from every action to one Horizon grant, so an agent can
build and run its own tools without asking, while matching a conventional
runtime on cost and task success and containing strictly more?

| When | Work |
|---|---|
| Late Sept | Semantics: contracts, `interpret`, resolve. **Done** |
| October | Fault routing inside a turn. **Self-extension**: an agent registers a tool it wrote as a narrower leaf and runs it with no approval. This is "the demo the thesis needs" |
| November | Audit and replay. The baseline harness and task set |
| December | Evaluation against Codex in two modes, ask-on-request and full-auto. Metrics: **human interruptions per task** (the headline), task success, overhead, friction (authority faults per task), tools that ran with zero approvals, and containment on something like AgentDojo |

**Open, unowned work:** the evaluation design, fault routing and restore after
a restart, self-extension (today the tool registry is read once, at startup),
and *preventing* Codex's native execution rather than only detecting it.

## Gaps and inconsistencies

1. **Two different sets of "three invariants."** The README's are: the emitter
   is never the checker; authority only narrows; truth is derived.
   `grips-and-invariants.md` has I (compiled cognition), II (typed claims) and
   III (narrowing plus provenance). They overlap but aren't the same list. A
   reader will trip on "Invariant 3" meaning different things in the two docs.
2. **Grip 6 claims more than is built.** The grips doc now leads with approval
   fatigue and says "anything built inside [the Horizon], agent-written tools
   included, runs without asking." Its status table marks III "built," but
   self-extension isn't built, and the thesis rests on it. The "Not built"
   bullet does mention it, but the summary table reads as if it's done.
3. **Grip 5 has no follow-through.** Grip 5 (watching five agents, and the
   "Horizon overlap as set intersection" view) appears in no plan or
   not-built list.
4. **Invariant I timing conflicts.** The grips doc says I is "deferred past the
   December evaluation." `research-question.md` includes H4 (stored procedures
   make repeat runs cheaper) as a December stretch goal. Pick one.
5. **Stale "The system" section in `research-question.md`.** It still says
   intake "parses them as one Move (invoke, return or fault)," which
   contradicts thought 14, and it cites "15 tests" and "boot to a leaf tool."
   The README cites "~3,400 lines, 23 tests." The actual count is about 4.3k
   lines and 26 test functions in `controller/`.
6. **Stale fault rule in `frontend/README.md`.** It says "a fault retires the
   anchor," which is DNA-19 behavior. On the controller, the conversation
   continues after a fault.
7. **`design/summary.md` and `design/proposal.md` lag behind.** Both say so
   themselves. `design/README.md` still marks thoughts 12 and 13 "Not built,"
   but both are built now.
8. **`nohup.out` is committed in git.**

## The biggest risk

The docs raise it themselves (research question, open question 4): if users
just grant broad Horizons, Skyra is "full auto with extra steps." How narrow a
Horizon can be while tasks still succeed is probably the number that decides
the whole result.
