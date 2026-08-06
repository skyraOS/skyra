# Execution

**Date:** 2026-08-06
**Status: FOUNDATION.** Carries `2026-08-03/execution.md` with the rules separated from the
direction — two of them applied to what is built today and were sitting under a *working toward*
heading.

> **The runtime separates what the agent knows from what the agent is currently doing.**

---

## The two structures

```text
DURABLE     THE GRAPH          structure and history. persists. evolves only when learning
                               occurs. holds NO claim about the present.

EPHEMERAL   EXECUTION STACK    a traversal over it. its top frame is the current expectation.
                               begins, pauses, resumes, terminates without altering the graph.
```

Execution is not represented by the graph. The graph is a place; execution is a walk. A process
does not become the filesystem — it traverses it, holds a position, and resolves names against
that position.

## The anchor

> **The active expectation is the execution anchor — a process's current working directory.**

Everything resolves **relative to that position**. The runtime does not reason globally; it reasons
locally from where it is standing. Which is what keeps context load small: it only gets large in
one place, *"what would close me,"* and that is a retrieval-and-plan task where a large context is
earning something.

### Handle and frame are different things

```text
HANDLE   a POSITION — where execution is standing. one stack entry.
FRAME    the NEIGHBOURHOOD that position exposes — what is in view from here.
```

A handle's position **determines** the attention frame. It is not one. `2026-07-27/workers.md`
keeps "frame"; the execution object is a **handle** — Unix's open file description, coupling a
durable node to a behavior plus the transient state of this use.

### Nothing is stored that a walk can derive

The pattern that resolved most of this set:

```text
where are we                  the frontier: where a descent stops
what is the active subgraph   what the walk touched
where does the plan live      it IS the subtree
what has fired                the transcript, queried
how do phases get ordered     nesting depth
```

Each started as *"we need a field for this"* and ended as a query. That is a reasonable prior for
the next thing that feels like it needs a home.

## ONE STACK

> **One subgraph is active at a time. It is connected, entered at a root. There is one stack.**

**This limits ATTENTION, not limbs.**

```text
ATTENTION   ONE stack. one subgraph. what the runtime is THINKING ABOUT.
LIMBS       N processes. what is happening IN THE WORLD on its behalf.
```

> **Just because one subgraph is active does not mean one process is alive.**

They were never the same resource. Reading ONE STACK as a limit on concurrent action is the
misreading to avoid; each in-flight process is addressed by its own `WaveID`, so several can be
outstanding without ambiguity about which returning effect belongs to which.

### Why one stack is safe

Stack-per-impulse existed to stop goal-directed and stimulus-driven work competing for one
pointer. That competition only matters if an execution can **run long** — blocked on the model,
blocked on the world.

Actions-as-residual removed both. A wave is graph reads and graph writes and nothing else: it runs
to completion and ends. There is nothing long-running for an arriving impulse to compete with, so
the premise of the earlier argument is gone.

### The property that has to survive

Several stacks were carrying Toy 5's ingress guarantee. Under one stack it moves to the queue:

```text
SAFE    impulses queue in ARRIVAL ORDER. every one is processed. nothing is ranked.
SEALS   the queue is ordered, or dropped, by relevance · trust · activation.
```

> **Admission must not consult attention.**

The property was never "many stacks." A FIFO queue gives it more simply. What must never appear is
a queue that decides which impulse matters — that is the sealing failure relocated into the
intake.

### Enter at the standing root

Under one stack, whatever is entered is the agent's **entire** attention for that turn.

```text
THE STANDING ROOT   always the entry. the active subgraph is the whole tree, BOUNDED BY
                    activation and closure. the bound is derived; nobody chooses it.
A SUBGRAPH ROOT     enter partway down. cheaper — and whoever picks the entry has chosen
                    what the agent attends to, by hand.
```

`chroot` remains possible and is how a resumed execution re-enters, but it should not be the
ordinary path.

## Interception: route and record, never filter

Cross-agent traffic passes through the runtime rather than going directly. The payoff is larger
than routing: **it makes the agent society auditable.** Direct channels would leave the runtime
blind to its own crossings.

```text
INTERCEPT TO ROUTE and RECORD     legitimate. which agent, and that it happened.
INTERCEPT TO FILTER by relevance  the sealing failure, at the front door.
```

*"The runtime is in charge of attention"* is true for **ordering and delivery**. It must not become
true for **admission**.

## The perceive is the interface

```text
NOT THIS RUNTIME'S CONCERN   how a delegated agent edits files behind its harness.
THIS RUNTIME'S CONCERN       what is alive within it, and what came back.
```

The runtime never inspects a process. It observes effects.

> **A process's self-report is not grounding. A perceive still has to fire.**

An agent saying *"I wrote the file"* is an expression, not an observation — dnav3's LAW 1 exactly:
*"world state does not contain it until something came back and said so."* Delegation makes this
**more** important: the more capable and opaque the actor, the more the only thing under control is
whether you looked. Letting an exit code close an expectation hands the grading channel to the
thing being graded.

**Worktree and diff is a perceive**, because the *runtime* computes it and the process does not
report it. Hold it as the fallback: a targeted perceive is cheaper and sharper, and a full diff is
what you reach for when you **cannot name what you expect** — the same condition as unplaceable
residual.

## The worktree is a third lifetime

The two structures at the top of this doc are not enough once work happens outside the runtime.

```text
DURABLE     THE GRAPH       structure and history. persists.
EPHEMERAL   THE PROCESS     has a pid, exits.
WORLD       THE WORKTREE    outlives the process that made it.
                            dies by RUNTIME DECISION, never by the process.
```

### Exit freezes the worktree. It does not drop it.

Continue-vs-scrap is decided by firing the perceive against the worktree and grading it. **If exit
dropped the worktree there would be nothing left to perceive**, and the only surviving artifact
would be the exit status — the exit-code failure arriving by a new route: not by *trusting* the
status, but by destroying the evidence until the status is all that remains.

```text
exit  →  worktree FROZEN, process gone
      →  the runtime computes the diff
      →  INTERPRET grades it against the expectation
      →  continue · re-assemble · rewrite · scrap
      →  the worktree is dropped HERE, and only on scrap
```

**Dropping is the scrap branch**, which is a decision the process is structurally least qualified to
make: it would be grading its own accumulated work, and a process that spent an hour building
something will not tell you to throw it away.

It is also what makes CONTINUE constructible at all. An exited process cannot be resumed; what
continues is the **work**, by re-exec'ing on the same worktree under the rewritten program. That
requires the worktree to outlive the process.

### Scrap is cheaper than it looks

Append-only means the learning already landed. Support edges written during the process's life are
in the graph and stay there. **The worktree is not the graph** — scrapping loses world-work, never
history, including the record that this approach did not get there.

That matters because otherwise the runtime has a thumb on the scale toward continuing, and
continuing on invalid state is the silent failure. Scrap is loud and bounded.

### Confinement is what makes the diff a valid perceive

Not hygiene. If a process can write outside its worktree, **the diff no longer captures what it
did** — the perceive is silently incomplete and every grade downstream of it is wrong. The runtime
would be grading a subset of the effects without knowing which subset.

The process should also not know it is confined:

```text
HIDE   the path does not exist. the process treats it as an ordinary absent file.
DENY   the path exists and returns EPERM. the process TRAPS — asking the runtime for
       access to something the runtime itself walled off.
```

Denying manufactures spurious `blocked` calls indistinguishable from genuine underspecification.
Hiding does not. A design argument, not a preference.

**Available where:** Linux user + mount namespaces give unprivileged hiding. macOS has no
equivalent — `sandbox_exec` denies rather than hides — so hiding there means a container, which
gets its own mount namespace and therefore hides correctly. Host directories bind-mount in without
copying; the cost is that many-small-file work is slower than native.

### Put the runtime inside the VM

Which collapses the macOS problem into the Linux one.

```text
RUNTIME ON THE HOST   every process file crosses the host↔VM virtiofs boundary. UID/GID
                      mapping, ownership, and diff performance are all problems.

RUNTIME IN THE VM     the runtime is a Linux process. it creates the clone on VM-native
                      filesystem, OWNS IT OUTRIGHT, and confines its own children with
                      namespaces — the mechanism that was good all along.
```

**One VM, runtime inside, N confined processes** — not one VM per process. Cheaper, and the diff
becomes a local git operation instead of one over virtiofs, which retires the performance caveat
too.

A host↔VM mount still exists, but it now serves only the **runtime**, which is trusted and may run
as root in a disposable VM. A far weaker requirement than serving an untrusted non-root process.

### A git worktree does not confine. Use a clone.

The word **worktree** is right for the concept — one process's slice of the world — and wrong for
the implementation.

`git worktree add` writes a `.git` **file** holding an absolute path back into the main repository's
`.git/worktrees/`. It is not self-contained:

```text
MOUNT ONLY THE WORKTREE   git breaks. the gitdir it points at is not there.
MOUNT THE REPO TOO        git works, and confinement is gone — the process sees every
                          branch and every other worktree, and can WRITE the shared .git.
```

The second is worse than a leak: a process corrupting the shared `.git` damages every other
process's slice at once, which is the opposite of the isolation this is for.

```text
SO   the confined unit is a CLONE. self-contained .git, mounts cleanly, and one process
     cannot reach another's history.
COST a clone across a mount boundary cannot hardlink, so it genuinely copies.
     --depth 1 or --filter=blob:none if that bites.
```

### macOS runtimes — checked, 2026-08-06

Two options, both real.

```text
DOCKER DESKTOP     one shared Linux VM, all containers inside it. bind mounts via
                   virtiofs. mature.

apple/container    at 1.0. OCI images, Swift, built on the Virtualization framework.
                   ONE VM PER CONTAINER — stronger isolation than a shared VM, which
                   is the direction this design wants.
```

**But its bind mounts currently reintroduce DENY.** UID/GID mapping across the virtiofs boundary is
incomplete — there is no `--userns keep-id` equivalent, so a non-root process in the container finds
bind-mounted files owned by root and cannot write them. Cross-volume single-file mounts also fail to
propagate writes back to the host, silently.

That is the failure this doc already argued against: **the process hits EPERM on a path the runtime
walled off by accident, and traps for a condition the runtime manufactured.**

```text
CONVERGES   don't bind-mount the repo. CLONE INTO the container.

            the git argument and the permission argument arrive at the same place
            from different directions, which is the strongest reason to trust it.
```

Bind mounts stay useful for handing *in* small read-only inputs. They should not carry the worktree.

### Reaping and re-exec are BOUNDARY

Worktrees outlive processes, so something must re-exec on them, and something must reap the ones
whose expectation finally closed. These are the same job: **acting on world state after a grade.**

They are probably not a new worker. BOUNDARY owns crossings, a worktree is the residue of a
crossing, and the four behaviors should be defended against growing a fifth.

## Why the graph may be walked while it is learned

Nothing is removed or rewritten, so a stack cannot be standing on something that ceases to exist.
Learning **adds**. A traversal in flight sees the new material or not, and both are consistent —
the same property that let the kernel drop its lock, arriving from the other direction.

---

## PROGRAM and PROCESS — a direction

**Not built.** Recorded because it is where this is heading, and because the vocabulary names the
durable/ephemeral split in words that carry their own intuitions.

```text
PROGRAM   the expectation.  durable. on disk. the thing you can run again.
PROCESS   one running of it. ephemeral. has a pid, a lifecycle, an exit.
```

An expectation is not a process any more than `/bin/ls` is. It is the thing you **exec**.

```text
WaveID          the pid
the transcript  the process trace
ack = +1        it STARTED. not that it worked.
the perceive    checking the world yourself instead of trusting the exit code
in-flight       running vs exited-nonzero
kill            SIGTERM — and "can it kill itself" is just exit
sequential      `&&` versus `&`
ONE STACK       a shell with one foreground job
```

### Actions delegated to an external agent

An expectation spawns a process, and that process may be an off-the-shelf coding agent rather than
anything this runtime implements. `2026-07-27/actions.md` argued for this without saying so:

> The action worker no longer needs to speak the target's language — it needs only to know that
> running this tends to move this expectation.

The action layer improves on its own. What this runtime is for is the expectation and the
observation.

**It does not break ONE STACK.** A live process is a crossing that *streams*: the wave spawns and
returns, the process runs beyond the boundary, results arrive as impulses. The wave still never
blocks. What is new is that a crossing becomes an open channel with a lifecycle rather than a send
with a receipt.

### The process-facing surface

Polling a running process does not work; the process traps when it needs something and the trap is
an impulse. That much was right. **Calling it an interrupt was not** — a trap that blocks for a
return value is a syscall, and filing it as an interrupt is what let a *grading* behavior get named
as the handler.

The surface is two traps out, one signal in, and no grading verb: [syscalls.md](syscalls.md).

### It is octopus-shaped

One brain holding expectation and attention; limbs with real local competence acting without
central command. The center learns *whether reaching that way works*, never how the arm does it.

### What the SDKs expose — checked 2026-08-06

This was listed as *"check rather than assume."* Checked. Both provide the shape.

```text
CLAUDE AGENT SDK   createSdkMcpServer / create_sdk_mcp_server registers tools that run
                   IN-PROCESS, not as a subprocess. streaming messages out.

                   tools: []   REMOVES EVERY BUILT-IN. the agent can then call nothing
                               except what the runtime authored.

                   a sandbox runtime in beta takes directory and network-host allowlists
                   without a container.

CODEX SDK          startThread / resumeThread, sessions persisted on disk, and
                   `codex exec` for non-interactive runs. tools registered as functions,
                   and built-ins are OVERRIDABLE BY NAME (read_file, local_shell,
                   apply_patch).

                   sandbox_mode: read-only | workspace-write | danger-full-access.
                   network off by default under workspace-write.
```

### `tools: []` is the interesting one

A process whose entire action vocabulary is runtime-authored makes **interception total** — the
runtime does not observe the effect, it *performs* it. `THE COUPLING IS CHOSEN BY THE AGENT, NEVER
BY THE DATA`, arriving at the process boundary.

And it upgrades what those results are:

```text
A TOOL CALL      what the process ASKED FOR.        an expression.
A TOOL RESULT    what the runtime's handler DID.    runtime-computed. gradeable.
```

**It does not retire the diff.** Any tool general enough to be useful — a shell most of all — lets
effects exceed what the tool signature modelled. So this instantiates a rule already written here:
a targeted perceive is cheaper and sharper, and the diff is the fallback and the ground truth.

### The asymmetry to plan around

Removing every built-in is a Claude SDK capability. Codex's override-by-name is weaker — replacing
`local_shell` is not the same as a guarantee that no other path exists. **The strong confinement
story does not port evenly**, so it should not be assumed when choosing a runner.

Codex's macOS seatbelt also has open bugs where `network_access` is silently ignored, and a setting
that silently fails to apply is worse than one that errors. Under the VM decision above the runtime
is on Linux and seatbelt never runs, so this is dodged rather than solved.

### On depending on someone else's runner

The exposure is real and it is small. What the runtime needs is **spawn · stream · interrupt · kill,
plus a tool ABI** — and the tool layer is ours. If the runner is replaced, the tools survive it,
because `the perceive is the interface` was already the rule.

That both SDKs independently converged on thread-resume, function-registered tools, and mode-based
sandboxing is evidence this is the shape of the category rather than one vendor's idea.

### Support both runners, and let confinement vary

> **The perceive is the interface, so it does not care which runner produced the worktree.**

Which is what makes two runners cheap. But they do not confine equally, and designing to the weaker
floor would throw away the stronger guarantee for nothing:

```text
DO NOT   design to the weakest common confinement. that discards tools: [] — total
         interception — to match a runner that cannot offer it.

DO       let the runtime KNOW which runner a limb has, and shift where grading leans.
```

```text
STRONG CONFINEMENT   the tool layer is the whole action vocabulary. tool RESULTS are
                     runtime-computed, so the sharp cheap perceive carries most of the load.

WEAK CONFINEMENT     effects can escape the tools that were modelled. THE DIFF CARRIES MORE.
```

The diff was already ground truth, so nothing new is needed to absorb the difference — the weaker
runner is simply more expensive to observe. **Confinement strength is a property of the limb**, and
this design already tolerates limbs of differing trust, because it never trusted any of them.

**Cost is not a design input.** Switching runners when one gets expensive is glue, and glue belongs
outside the model. What the model owes is that the switch changes nothing structural — which the
above is what buys.
