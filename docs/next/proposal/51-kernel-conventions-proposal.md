# 51 — Conventions the kernel still knows

Date: 2026-10-28

Status: **to look at.** Not a proposal yet: a list of places to examine, and
the rule to examine them by.

**Scope:** Skyra conventions that live in `core` and `runtime` although the
kernel does not need them. The same kind of special case
[49](49-services-behind-adapters-proposal.md)'s cleanse removed for
residents.

## The rule

The apprentice (today's interpreter, renamed in
[50](50-write-policies-proposal.md)'s tabled notes) is an actor, registered
like Codex or `echo`. The kernel never names it, and the split between the
kernel and the apprentice is a Skyra convention, not something the kernel
enforces. That is right, because the apprentice holds no special power:
security comes from intake checks, grants, contracts and the record, which
apply to every actor alike.

So: **the kernel serves capabilities, never roles.** Anything a convention
needs from the kernel is a generic capability, granted to whichever actor's
grant holds it, never tied to a name or a kind of actor. Whatever fails that
test is a candidate to move out.

## Candidates

### 1. Procedure bundles

The runtime knows the apprentice's bundle format.

- `runtime/registry/registry.go` (`Entry.Bundle`, the bundle block in
  `load`): parses the bundle (version 1, its procedures), turns each
  procedure's contract into one of the surface's contracts, and requires a
  stateful continuation. An entry with a bundle is also forced to be a work
  tool (`WorkTool: e.WorkTool || e.Bundle != ""`).
- `runtime/boot/boot.go` (the `BundlePath` block in `Run`): checks the
  bundle's hash, then runs `<binary> install-bundle <state> <bundle>` before
  genesis.
- `core/horizon/horizon.go` (`Surface.BundlePath`, `BundleSHA256`) and
  `core/machine/continuation.go` (`against`): a bundle changes which
  contract a continuation is told it executes.

To look at: can provisioning be an ordinary first invocation of the actor,
recorded like any other, and its contracts declared in `contract.json`
like everyone else's?

### 2. Agent fields and the agent-context syscall

Agent conventions live in the Horizon and the room.

- `core/horizon/horizon.go`: `Surface.Agent` (`AgentConfig`, the agent's
  private operations), and `Agent`, `WorkTools` and `Delegated` on both
  `Horizon` and `Grant`.
- `core/machine/room.go`: the `system.agent-context` syscall
  (`AgentContext`, `agentContext`), which records a provider model; and the
  delegation branches of `narrowing`.
- `core/machine/continuation.go`: `agentName`, `agents`, `workToolNames`,
  and the `Actor`, `Agent`, `Agents`, `WorkTools`, `Delegated` fields every
  continuation is shown.
- `runtime/registry/registry.go`: the work-tool and agent-operation checks
  at the end of `load`.

`docs/summary.md` already calls these "current mechanisms, not the intended
final ownership boundaries." To look at: can each become contract data, or
a grant the kernel checks without knowing what an agent is?

### 3. Comments

`core/horizon/tools.go` ("the same deterministic interpreter binding as the
transports") and `core/contract/contract.go` (the interpreter's contracts)
describe kernel code in the apprentice's terms. Small, but a sign of where
the convention leaks.

## What to answer first

- Which of these does the kernel *check*, not just carry? A field the kernel
  only copies to a continuation is information and can travel as contract
  data. A field the kernel enforces must become a generic capability.
- Does 50's "parent of processes" (a lifecycle handle, join and write
  policies) land as a generic grant? If so, it is the first capability built
  under this rule, and the agent fields should follow the same pattern.
