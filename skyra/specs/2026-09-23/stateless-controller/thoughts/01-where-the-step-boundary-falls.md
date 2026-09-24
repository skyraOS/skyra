# Where the step boundary falls

Date: 2026-09-23  
Status: decided 2026-09-23: B (interpret, then spawn).

The pipeline is always the same alternation:

```text
… → interpret → spawn → interpret → spawn → …
```

The question is where one `skyra` process begins and ends, and so which half
shares a process with which.

## A: spawn, then interpret

```text
skyra(N):   spawn target → wait for its output → interpret it → hand result to skyra(N+1)
skyra(N+1): spawn the next target from the already-interpreted Move → …
```

- The process that authorizes a spawn (N) is not the process that performs it
  (N+1). The check and the act are separated by a process boundary and a delay.
  In between, the tree can change: an ancestor faults, a registration race is
  lost, authority is revoked. N+1 acts on a stale check.
- Write-before-execute spans two processes. N writes and N+1 executes, so a
  crash between them leaves intent written by one process and acted on by
  another that must trust it.
- N+1 receives a Move that is already checked and trusts it. That conflicts with
  "the emitter is never the checker": it holds only if N+1 re-checks, which
  makes N's interpretation redundant.
- N must outlive its child to read the child's output. That conflicts with "the
  parent dies once the child is running".
- In its favour, the process that spawned the child naturally holds the child's
  stdout. That answers the stream and live-parent problem, but only by making
  the holder smart, since it interprets.

## B: interpret, then spawn

```text
skyra(N):   receive a raw Move → interpret (read, authorize, write) → spawn target → exit
target:     output is raw, unchecked bytes
skyra(N+1): receive those bytes as its Move → interpret → spawn → exit
```

- The check and the act it authorizes happen in one process, back to back, with
  the write between them. This is read → write → execute with nothing in between.
- Everything that crosses a process boundary is unchecked, and the receiver
  checks it. "The emitter is never the checker" holds structurally.
- A crash leaves a clean cut. Before the write, nothing happened. After the
  write, a recorded intent with no result is faulted on restore.
- The spawning `skyra` does not need the child's output, so it can exit. It
  wires the child's stdout to N+1's stdin before exiting.
- Cost: N+1 exists, blocked on stdin, from the moment the child starts. For a
  one-shot child that is fine. For a streaming child (Codex, `user`), one N+1
  reads one Move, and someone must start the N+2 for the next one. That is the
  holder, and in B it stays dumb because interpretation never lives there.

## Decision

B. An emission is unchecked output, so it must be interpreted by the step that
receives it, and that step then spawns. This also keeps "the emitter is never
the checker". B is the shape the 2026-09-22 decisions already describe, it keeps check
and act in one process, and it leaves the stream problem to a relay that never
interprets. A's one advantage, owning the child's stdout, is really the holder
question, and B answers it without moving the check.

## Side point: how the Move enters

The Move should arrive on stdin, not in args. argv is visible in `ps` to other
local users, which would expose prompts and tool content. It is also capped at
about 1 MB on macOS, and a prompt with history can exceed that. Keep only the
opcode (and target) in argv.
