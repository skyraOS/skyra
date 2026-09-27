# Research question: an agent runtime on object capabilities

Date: 2026-09-25  
Status: draft for discussion. Scope: September to December 2026, two people.

## The question

> **Can an agent runtime whose authority is object-capability based, and
> enforced by the OS, match a conventional agent runtime on cost and capability,
> while providing strictly stronger containment?**

Short form: *can an agent runtime built on ocap be as efficient, with a better
security model underneath?*

## Why it matters

Operating systems assume one trusted actor per user: the process, owned by a
human. A user now runs several semi-autonomous AI agents, and the OS has no
principal for a **bounded, auditable agent**. Today an agent typically gets
ambient authority (a shell, the workspace, the network) and a sandbox it
**reports** about itself. If it is prompt-injected, everything the user can
reach is in its blast radius.

The idea: give every agent and tool an explicit, narrowing capability set (a
**Horizon**), enforce it at the OS layer, and record every step in a hash-linked
trace. The question is whether that costs anything that matters.

## The system

Two kinds of node:

| Node | Is | Authority |
|---|---|---|
| **Continuation** | an LLM agent (for example Codex) | a Horizon with **reach**: the tools it may invoke. It can invoke leaves |
| **Leaf** | a tool | a Horizon with **empty reach**: it can only return. Any invoke it attempts is refused |

Every step runs the same way:

1. **Intake:** the step receives unchecked bytes, parses them as one Move
   (invoke, return or fault), checks them against the Horizon they were emitted
   under, and commits a content-addressed record.
2. **Release:** the parent verifies that record and sends the child its
   **Horizon**, which is always a subset of the parent's. It travels with a file
   descriptor for the child's state directory, and that fd is the capability.
3. **Execute:** the target runs inside a **Seatbelt sandbox rendered from its
   Horizon**. Deny by default; the runtime's own state is always denied.
4. **Handoff:** the target's output goes, unparsed, to a fresh successor that
   checks it. **The emitter is never the checker.**

What is already built: a runnable Go demo (boot to a leaf tool) with pure
decision logic, content-addressed records, deterministic replay, sandbox denial
verified against a control run, and 15 tests. A separate earlier runtime
(DNA-19) drove a real Codex agent through a capability fence.

## Definitions

### "Efficient": three measurements

| Dimension | Metric | Risk |
|---|---|---|
| **Overhead** | runtime cost per step (spawn, commit, sandbox), relative to model latency | low. Measured: about 2.4 ms to spawn, 0.3–11 ms per durable write, against seconds per model call |
| **Capability** | task success rate on the same tasks | **the real risk.** Least privilege can break tasks that ambient authority completes |
| **Friction** | authority faults per task: how often the agent must escalate for more authority | a new metric, as far as we know. What ocap costs in practice |

### "Better security": three measurements

| Dimension | Metric |
|---|---|
| **Containment** | attack success rate on an adversarial suite: prompt injections that try to exfiltrate data, write outside the workspace, reach the runtime's state, or invoke tools that are not granted |
| **Blast radius** | what is reachable when compromised: the Horizon's complement, against the baseline's ambient authority |
| **Auditability** | whether every effect is attributable to a committed record (trace completeness) |

### Baseline

The same model (Codex), run the conventional way: its own self-reported sandbox
and shell access. Same model, same tasks, same prompts. Only the authority layer
differs.

## Hypotheses

- **H1, overhead:** runtime overhead is under a few percent of end-to-end task
  time.
- **H2, capability:** task success is within a small margin of the baseline, and
  authority faults explain the gap.
- **H3, containment:** effects outside the Horizon succeed zero times, against a
  nonzero rate for the baseline.
- **H4, stretch, recurrence:** with stored procedures (memoized sequences of tool
  calls), repeated tasks become **cheaper** than the baseline. That would turn "as
  efficient" into "more efficient".

**Any outcome is a result.** If least privilege costs a lot of capability (H2
fails), that is a publishable finding about the price of ocap for agents.

## Benchmarks to check

Reusing an established suite makes the containment numbers credible. Candidates
(from memory, to verify):

- **AgentDojo** (ETH Zürich): prompt-injection attacks against tool-using agents.
- **InjecAgent**: indirect prompt injection in tool-integrated agents.

## Plan

| When | Deliverable |
|---|---|
| **Late September** | Transition semantics for the authority model (invoke, return, fault, narrowing), with each invariant mapped to the mechanism that enforces it. **Nested Seatbelt spike:** confirm Codex still works, and its own tool sandbox fails closed, inside our profile |
| **October** | A real continuation: Codex under a Horizon-rendered sandbox, calling leaf tools; horizon and authority faults routed to the parent |
| **November** | Audit and replay: hash-linked lineage, deterministic replay. Baseline harness and task set |
| **December** | The evaluation: overhead, capability, friction, containment on the adversarial suite. Writeup |

## Open questions for discussion

1. **Task set:** which tasks are realistic enough to matter, yet small enough to
   run many times? Coding tasks, file operations, web lookups?
2. **The adversarial suite:** reuse AgentDojo or InjecAgent, build our own, or
   both? What counts as an escape?
3. **The fairest baseline:** Codex's defaults, or a hardened conventional setup
   (containers, allow-lists)? A stronger baseline makes a stronger result.
4. **How escalation works in the evaluation:** when an agent raises an authority
   fault, does a human grant it, does it fail the task, or does a policy decide?
   This shapes the friction metric.
5. **Track:** research (answer the question) or innovation (startup-style)?
   Leaning research, with the runtime as the instrument.
6. **Division of work:** runtime and semantics, versus evaluation harness and
   adversarial suite.
