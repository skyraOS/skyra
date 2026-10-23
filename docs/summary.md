# Skyra: where we are

Date: 2026-10-09
Status: current implementation on `main`, checked against source and recent
commits. Proposed changes are separated below. This summary is not a claim that
all providers or failure paths have been verified with live inference.

The [README](../README.md) introduces the project; [startup](startup-doc.md)
and the [desktop guide](../desktop/README.md) describe running it. Older design
notes preserve the reasoning, but their status tables can lag the code.

## Current position

Skyra runs agent work through explicit contracts under a Horizon: the calls and
OS resources an execution holds. The controller runs once per step, derives
invocations, returns and faults from untrusted output, records them, and hands
custody to the successor. The core decision function is pure; the runtime shell
performs its effects. Durable records and released Horizons can be checked by
replay. Provider processes and hosts can remain alive across these ephemeral
steps; their live state is not reconstructed by replay alone.

Procedures are the application programming model. The interpreter executes
operator-installed or remembered compositions of ordinary contracts. Think is
now an authored procedure, with Codex or Claude selected by a model parameter.
Bounded delegation is implemented for both adapters. The desktop presents
procedures as roots, their composition as inspectable constellations, and their
conversations as vertical cards.

The next architectural work is separating generic authority from adapter
configuration and process lifecycle. That refactor is proposed, not implemented.
The current kernel still carries agent metadata, a work-tool ceiling, delegation
state, and resident-specific execution fields. Provider-native operations and
sandbox exceptions also prevent claiming uniform Horizon enforcement today.

## What is implemented

| Area | Current behavior |
| --- | --- |
| Step protocol | Successor intake derives and commits output before effects; release verification derives the Horizon; custody travels with frame directory descriptors. |
| Contracts | A capability is identified by `(target, schema)`. Structural fit selects exactly one declared contract; missing or ambiguous matches fault. Returns carry observations rather than declared output contracts. |
| Authority | Ordinary calls inherit the caller's effective Horizon. Explicit `system.narrow` returns a reference for a restricted execution. Generic Seatbelt rendering denies by default and protects the runtime floor. Provider exceptions are listed below. |
| Fan-out and joining | A round opens all child slots before release; it resumes after all discharges, in emission order, with one winning claim. Generic cancellation is not implemented. |
| Recording and replay | Content-addressed records, inputs and Horizons preserve provenance. Version 2 histories commit complete emissions so replay checks round coverage as well as discharge ownership. |
| Accepted history | Durable acceptance receipts distinguish an accepted release from a proposed one. `system.strand` supports both single-observation strands and invocation-root plus nonempty observation-frontier reads. |
| Reconstruction support | `ReadTreeInRoom` reconstructs the bounded accepted closure and preserves narrowed-child opacity. Operator-side `RecoverTree` reconciles recorded progress; it does not revive processes or repeat effects. |
| Interpreter | Durable procedure bodies and frame positions; bindings, dynamic call targets, `each`, independent ready-node fan-out, and procedure-to-procedure calls. |
| Authored procedures | Version 1 bundles are validated and provisioned before genesis. Think, design, Blender and write workflows use this path; deployment/migration versioning remains incomplete. |
| Remembering | `remember` compiles completed recorded work, exposes a contract through actor-owned `extend-reach`, and runs it inside base refinements. CEL refinement and model-fallback machinery exists; real-provider generalization is not established by the scripted demos. |
| Provider adapters | Codex app-server and Claude CLI adapters translate provider tool requests into recorded invocations. Their transport operations remain separate registered targets in the current implementation. |
| Ingress | Authenticated WebSockets, bounded framing, durable request admission and conflict detection. Interrupted requests become `delivery-unknown`; they are not automatically retried. |
| Clients | Browser, terminal and desktop clients consume the conversation protocol. The desktop has procedure roots, model selection, inspectable tool outlines, history, attachments and vertical conversation cards. |
| Verification tooling | `skyra test`, controller demos and provider probes cover module checks, fake-provider behavior, real runtime boundaries and replay. Live subscription inference is a separate check. |

### Think and the conversation path

The current turn path is:

```text
human → browser/terminal/desktop → webhost → user actor
      → interpreter / think(text, model, context, images)
      → system.narrow → selected adapter → recorded tool calls
      → procedure result → user actor → parked browser interaction
```

`think` remains an interpreter bundle. Its dynamic target comes from the `model`
parameter (`codex` or `claude`); it narrows on each invocation, then calls the
returned reference. Prior turns still travel as a growing `context` value.
The October 9 merge unifies Think's procedure identity and preserves old
`think-claude` conversations through compatibility mapping. It does not implement
all of [proposal 39](next/proposal/39-unified-think-procedure-proposal.md):
a compiled Think target, one-time conversation narrowing and adapter-owned
memory are not the current path.

The launcher provisions tools and workflow dependencies according to the
workspace. These include read, find, roots and write, plus scoped repository,
design, Blender and desktop-observation operations where configured. Both
adapters expose tools from the effective scoped catalog, rather than hiding
otherwise authorized non-work bindings by presentation alone.

The desktop scopes conversation navigation by procedure and selected model.
The newly merged cards show the first user message, selection and phase, with
scrolling and a new-conversation action. Activity is grouped under the user's
request; provider adapter steps are hidden for both Codex and Claude, including
configured adapters. Interpreter calls are still filtered, so public delegation
and remembering activity need the visibility fix in the change plan.

Conversation creation, listing, loading, closing and deletion are still partly
webhost/user-actor responsibilities. An authored Think turn is not the complete
procedure-native conversation application proposed in
[proposal 33](next/proposal/33-procedure-adapter-extraction-proposal.md).

### Delegation and current resource lifetime

The public interpreter contract is:

```json
{"delegate":{"goal":"Review the change","actor":"claude","tools":["read","find"]}}
```

Omitting `actor` inherits the caller's adapter and resolved model; an explicit
actor uses that adapter's configured default. Omitted tools inherit the effective
work-tool selection; `tools: []` permits no work calls. Concrete target/contract
bindings enforce the ceiling across indirect calls and nested delegation.
Selecting a procedure does not grant its dependencies. Transport operations
currently receive an internal exemption from this work-tool restriction.

A child receives only its goal as task context, without parent history, images
or provider session identifiers. It gets a separate provider edge. Ordinary
child faults and refusals return `{failed:{kind,cause}}`; integrity and persistence
failures retain the runtime's stop behavior. Delegated Codex records a stop before
returning, while Claude uses a one-shot provider process. The parent's holder
stays alive while its frame is unfinished, then resumes idle reclamation.

The implementation still uses `Horizon.Agent`, `WorkTools` and `Delegated`,
`system.agent-context`, and delegated socket/profile branches. These are current
mechanisms, not the intended final ownership boundaries. See
[DELEGATION.md](design/DELEGATION.md) for exact semantics and verification limits.

## Boundaries and gaps

- **Provider authority is not uniformly Horizon-derived.** Generic workers use
  the deny-default Horizon profile, but provider launch profiles replace it.
  Codex parents may use native image viewing and provider web search without
  individual reads being recorded in lineage; delegated children disable those
  capabilities. Claude's current profile starts with `allow default` and denies
  the runtime floor; disabling native tools is not equivalent to OS containment.
  The universal claims that every effect is recorded and every process has only
  Horizon-rendered authority are therefore goals, not current guarantees.
- **The kernel still knows agent-specific representation.** Agent declarations,
  model publication, work-tool labels and delegated mode remain in core. Surface
  also retains resident/connection/profile fields. Moving policy out requires
  equivalent authority enforcement before deleting these mechanisms.
- **History reconstruction is ahead of application recovery.** Root/frontier
  reads, acceptance evidence and operator reconciliation exist. User-owned
  checkpoint publication and full conversation recovery are not wired into the
  application. A checkpoint must retain all accepted branch endpoints, not one
  latest record when work fans out.
- **A host restart does not restore live execution.** History remains inspectable;
  unfinished admitted work is reported as `delivery-unknown`. Recorded intent
  cannot establish the outcome of an unobserved external effect or recreate a
  provider process, socket or session.
- **The procedure language remains a finite DAG.** It has bindings and `each`,
  but no general conditional node, explicit fault handler or durable event loop.
  Bundles exist; migrations and multiple representations per contract do not.
- **Containment is shared within a Horizon.** A granted actor can invoke other
  contracts that Horizon holds. Restrictions require explicit scope construction;
  a prompt injection that misuses a granted tool is not prevented by the grant.
- **Some replay and durability limits remain.** Version 1 histories cannot prove
  complete round coverage; records still contain checked derived fields; the
  stack is serialized through `Resume`. Publication is synced, but no power-loss
  certification is claimed.
- **Large views and context remain bounded imperfectly.** Snapshots over 16 MiB
  omit old history without a client truncation notice. Think's context grows
  with prior turns; compaction is not implemented.

## Current architectural direction

The [Horizon concern spec](design/HORIZON-CONCERNS-SPEC.md) proposes generic
contract-level reach, adapter-owned configuration and tool bindings, removal of
agent/delegation fields from core, explicit resource ownership, and a tested
legacy replay policy. It does not change runtime behavior by itself.

The latest lifecycle discussion refines the resource design further and
supersedes the earlier idea of attaching lifecycle policy metadata to contracts:

- Normal authorized invocation remains the execution and communication path.
  A resource handle must not create a second unchecked exchange channel.
- Different contracts of one target may implement different lifecycle behavior
  by composition; persistence is not an actor-wide mode or authority granted by
  a contract declaration.
- A separate lifecycle actor is the candidate owner of start, retention, reuse
  and shutdown policy. The kernel supplies generic ownership/custody mechanisms
  and validates execution against the held target/contract authority.
- Endpoint reservation must be atomic and scoped. A port lock or discovered port
  alone does not identify an authorized process or confer invocation authority.
- The primitive boundary is still being designed. Kernel-managed acquire/release
  was an earlier option; lifecycle-actor control with lower-level kernel support
  is the latest direction, not an implemented or finalized protocol.

The detailed specs still describe the earlier acquire/release approach and need
reconciliation with this discussion before implementation. Process reuse must
preserve authority bounds, independent ownership and recovery semantics under
either design.

## Next work

1. Reconcile the authority and lifecycle specs into one implementable protocol,
   including fixed versus expandable scopes, provider-native recording,
   cancellation, resource loss and old-conversation compatibility.
2. Fix public interpreter activity visibility and parent/child correlation.
3. Establish equivalent generic contract authority and provider enforcement before
   removing `WorkTools`, `AgentContext`, `Delegated` or resident-specific branches.
4. Build user-owned root/frontier checkpoints and projection, then move the rest
   of the conversation application into procedures without duplicating state.
5. Validate live provider isolation, restart behavior and remembered recurrence
   separately from deterministic fixtures and replay tests.

The research question remains whether one bounded grant can replace repeated
human approvals while preserving task success and containing effects outside the
grant. Procedures add a second measure: how much repeated work displaces model
calls. Existing metrics read procedure execution and fallback from records;
comparative task-success, containment and cost evaluation remains future work.

## Where to read next

| Topic | Source |
| --- | --- |
| Runtime implementation | [controller guide](../controller/README.md), [machine](../controller/core/machine/machine.go) |
| Procedure language and remembering | [procedures](next/32-procedures.md), [interpreter](../controller/interp/proc/proc.go) |
| Current Think composition | [bundle](../controller/interp/workflows/think/bundle.json), [proposal 38](next/proposal/38-think-procedure-proposal.md) |
| Unified model selection and remaining ideas | [proposal 39](next/proposal/39-unified-think-procedure-proposal.md) |
| Accepted tree reconstruction | [tree reader](../controller/runtime/replay/tree.go), [reconciliation](../controller/runtime/replay/recover.go), [proposal 33](next/proposal/33-procedure-adapter-extraction-proposal.md) |
| Bounded delegation | [DELEGATION.md](design/DELEGATION.md) |
| Authority and adapter refactor | [Horizon concern spec](design/HORIZON-CONCERNS-SPEC.md) |
| Desktop behavior | [desktop guide](../desktop/README.md) |
| Research and invariants | [research question](next/research-question.md), [grips and invariants](grips-and-invariants.md) |
