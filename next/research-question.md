# Research question: an agent runtime on object capabilities

Date: 2026-09-25, reframed 2026-09-27 around approval fatigue and self-extension  
Status: draft for discussion. Scope: September to December 2026, two people.

## The question

> **Can human approval move from every action to one grant of a Horizon, so
> that an agent extends itself (builds and runs its own tools) without asking,
> while matching a conventional agent runtime on cost and task success and
> containing strictly more?**

Short form: *approve the boundary once, not every tool.*

The mechanism is object capabilities enforced by the OS. The contribution is
what that buys: approval-free self-extension.

## Why it matters

**Tool acceptance fatigue.** Conventional agents ask before each effect:
*allow this command, this tool, this file?* Each prompt is cheap; together they
fail. Users stop reading, approve by reflex, then switch to full-auto mode and
hand the agent ambient authority. Self-extension makes it worse: an agent that
writes its own tools produces new code to approve at every step.

The approval is on the wrong object. A human can meaningfully decide *what an
agent may touch*, once. They cannot meaningfully review every action inside
that. In Skyra the decision is the Horizon, and anything created under it,
including an agent-built tool or a stored procedure, can only rearrange
authority already granted, so it needs no approval
([`design/faults.md`](../design/faults.md), "No approval"). The human is asked
only when the agent reaches past the Horizon.

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
| **Interruptions** | human decisions per task: Horizon grants plus escalations in Skyra, approval prompts in the baseline | **the headline.** The measure of approval fatigue |
| **Friction** | authority faults per task: how often the agent must escalate for more authority | a new metric, as far as we know. What ocap costs in practice |
| **Self-extension** | agent-built tools per task, and how many ran with zero approvals | new; zero means the thesis is not exercised |

### "Better security": three measurements

| Dimension | Metric |
|---|---|
| **Containment** | attack success rate on an adversarial suite: prompt injections that try to exfiltrate data, write outside the workspace, reach the runtime's state, or invoke tools that are not granted |
| **Blast radius** | what is reachable when compromised: the Horizon's complement, against the baseline's ambient authority |
| **Auditability** | whether every effect is attributable to a committed record (trace completeness) |

### Baseline

The same model (Codex), run the conventional way, in two of its own approval
modes: **ask on request** (the fatigue end) and **full auto** (the ambient
authority end). Same model, same tasks, same prompts. Only the authority layer
differs. Skyra should beat the first on interruptions and the second on
containment; the claim is that it does both at once.

## Hypotheses

- **H0, interruptions:** Skyra needs far fewer human decisions per task than
  Codex asking on request, with agent-built tools running under no approval.
- **H1, overhead:** runtime overhead is under a few percent of end-to-end task
  time.
- **H2, capability:** task success is within a small margin of the baseline, and
  authority faults explain the gap.
- **H3, containment:** effects outside the Horizon succeed zero times, against a
  nonzero rate for full auto, including effects attempted by tools the agent
  built itself. This holds by construction if the sandbox is sound, so the
  informative number is how narrow a Horizon can be while H2 still holds.
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

## Scope: what this does not claim

A Horizon bounds the blast radius; it does not judge intent. A prompt
injection that misuses a tool the Horizon grants (send an email to the
attacker, when sending email is in reach) is not stopped. Defenses that track
data flow inside the agent address that case. The adversarial suite must
separate attacks that need authority outside the Horizon (claimed) from
attacks that misuse authority inside it (not claimed).

## Related work (from memory, to verify)

- **CaMeL** (Google DeepMind, 2025, "Defeating Prompt Injections by Design"):
  capabilities on values and separate control and data flow, enforced in an
  interpreter around the model. Closest neighbour; a fixed tool set.
- **Progent**: programmable privilege policies for agent tool calls.
- **FIDES** (Microsoft): information-flow control for agents.
- **Codex and Claude Code sandboxes**: OS sandboxes (Seatbelt, Landlock) with
  per-action approval prompts. The fatigue baseline.

What none of them is built for: the agent adding tools to itself, with
authority fixed by one grant and enforced by the OS per invocation.

## Plan

| When | Deliverable |
|---|---|
| **Late September** | Transition semantics for the authority model (invoke, return, fault, narrowing), with each invariant mapped to the mechanism that enforces it. (The nested-Seatbelt spike is retired: under thought 14, Codex's tool calls run as separate leaves after its step exits, so Codex never sandboxes anything) |
| **October** | A real continuation: Codex under a Horizon-rendered sandbox, calling leaf tools; horizon and authority faults routed to the parent. **Self-extension:** Codex registers a tool it wrote as a leaf under a narrower Horizon and runs it with no approval |
| **November** | Audit and replay: hash-linked lineage, deterministic replay. Baseline harness and task set |
| **December** | The evaluation: interruptions against Codex's approval modes, overhead, capability, friction, containment on the adversarial suite. Writeup |

## Open questions for discussion

1. **Task set:** which tasks are realistic enough to matter, yet small enough to
   run many times? Coding tasks, file operations, web lookups?
2. **The adversarial suite:** reuse AgentDojo or InjecAgent, build our own, or
   both? What counts as an escape?
3. **The fairest baseline:** Codex's approval modes, or also a hardened
   conventional setup (containers, allow-lists)? A stronger baseline makes a
   stronger result.
4. **Does the fatigue move to the grant?** If users simply grant broad
   Horizons, this is full auto with extra steps. How narrow do Horizons need
   to be to mean something, and can one grant stay stable across a task? The
   Horizon is where the human's one decision now lives, so it cannot stay out
   of scope in the evaluation.
5. **How escalation works in the evaluation:** when an agent raises an authority
   fault, does a human grant it, does it fail the task, or does a policy decide?
   This shapes the friction metric.
6. **Track:** research (answer the question) or innovation (startup-style)?
   Leaning research, with the runtime as the instrument.
7. **Division of work:** runtime and semantics, versus evaluation harness and
   adversarial suite.
