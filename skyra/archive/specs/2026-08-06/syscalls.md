# Syscalls

**Date:** 2026-08-06
**Status: SPEC.** New. The process-facing surface implied by `execution.md`'s program/process
direction, worked out once that direction stopped being speculative.

> **A process may report on itself. It may never report on the world.**

---

## Interrupt was the wrong word

`2026-08-03/execution.md` filed *"the process throws when it needs something"* under **interrupt**,
and this set carried it. In the program/process vocabulary those docs are committed to, it is not
one:

```text
INTERRUPT   asynchronous, unrequested, imposed from OUTSIDE the process.
SYSCALL     the process voluntarily traps, BLOCKS, and gets a return value.
```

Everything described there is the syscall direction. Calling it an interrupt is what let a *grading*
behavior get filed as the handler — and grading runs the other way. The genuine interrupt table is
runtime → process and has exactly one entry.

## The surface

```text
TRAPS — process → runtime, blocking

  blocked(condition)    the ONLY trap. states a CONDITION, never a remedy.
                        returns: more context, or the process does not return.

  exit(status)          licenses the perceive. grounds NOTHING.
                        returns: nothing. the process is already over.

SIGNALS — runtime → process

  kill                  the only one.
```

**Two out, one in, and no grading verb anywhere.** There is no way for a process to ask *am I done*
or assert *I am done*. `execution.md`:

> Letting an exit code close an expectation hands the grading channel to the thing being graded.

A process-callable INTERPRET is that with extra steps, so INTERPRET has no process-facing form at
all. The process reports, the perceive fires, INTERPRET grades **what the perceive returned**.

## `blocked` — one trap, not two

*"I need a value"* and *"I cannot do this as specified"* are the same condition. If the process
traps, it did not have what it needed; if it did not have what it needed, the program did not
specify it. One trap.

The condition may be specific. The response may never be.

```text
LEGITIMATE   "I need the config path."      a need.
FORBIDDEN    "re-assemble at budget 4000."  a remedy — the process has chosen.
             "replan me."                    same.
```

Four responses are available and all four are the runtime's:

```text
CONTINUE      hand back context. the program was fine.
RE-ASSEMBLE   the node existed and the budget missed it. the INVOCATION was thin,
              not the program — closer to a missing argv than a wrong binary.
REWRITE       the program was underspecified. EXPECT runs.
KILL          the condition is not repairable here.
```

**The process cannot tell which it got.** It resumes, or it does not.

### A trap queues

It arrives as an impulse under `ADMISSION MUST NOT CONSULT ATTENTION`, so a blocked process waits
its turn like anything else. Otherwise *"I am blocked"* is a stack-hijack primitive handed to the
least trustworthy party in the system — Q4's attention hijack, with a caller.

## `exit` — a trigger, carrying almost nothing

Status, and nothing else. Not results, not a summary. **The moment `exit` carries content it is a
results channel and the perceive becomes ceremony.**

Status survives only because it reports on the *process* — exited vs. hung vs. killed, which is
liveness bookkeeping. It is not a claim about the world.

```text
exit(0)        "I think I am done."     an EXPRESSION.
exit(nonzero)  "I think I failed."      also an expression.
```

**Neither changes the grade.** A process can crash after doing the work correctly, or exit clean
having done nothing. Both license the same perceive.

## What the perceive is

```text
RUNTIME-COMPUTED     the worktree diff, plus observations the runtime INTERCEPTED.
                     graded. this is the whole perceive.

PROCESS-AUTHORED     "I edited these files, the tests pass."
                     NOT COLLECTED. the runtime already has the intercepted observations,
                     so it is noise with a credibility problem attached.
```

`execution.md` already blesses the first: *"Worktree and diff is a perceive, because the runtime
computes it and the process does not report it."*

---

## Deliberately not in the minimum

```text
SPAWN       can a process start a child process? that asks whether the process or the
            runtime owns tree growth — and EXPECT is the only writer of state. probably
            answers itself.

HEARTBEAT   "hung" is a third state beside running and exited. a runtime-side timeout,
            not a call.

SUSPEND     attention moving does NOT suspend a limb. ONE STACK bounds attention;
            processes keep running. there is no signal here.
```

## Retrieval behind `blocked`

The CONTINUE and RE-ASSEMBLE responses both mean *walk the graph for a value the process needs*. A
**retrieval** walk — it heats nothing — over the active subgraph first.

Retrieval is itself an expectation, private per agent, and its perceive is **whether the process
traps again for the same condition**: [behaviors.md](behaviors.md). The walk's shape is still
undesigned and is the largest thing behind this surface.
