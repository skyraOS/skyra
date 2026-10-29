# Separate Boundary authority from adapter configuration and delegation

Status: proposed. This document specifies a refactor; it does not change runtime behavior.

## Decision

Delegation is the interpreter's `delegate` procedure. It is not a Boundary mode.
Remove `Boundary.Delegated` and `Grant.Delegated`. Do not replace them with another
agent-mode flag in the kernel.

Move agent/model configuration and model-facing tool bindings into contracts
exposed by the Codex and Claude adapters. Procedures compose those contracts.
Remove `AgentContext`, `ToolBinding`, and `WorkTools` from `core/boundary` and remove
their special continuation fields and syscalls.

Preserve kernel enforcement of concrete invocation capabilities. An adapter's
tool catalog is a presentation of capabilities, not a grant of authority.
The kernel must be able to restrict individual contracts on one actor without
knowing tool labels, providers, models, sessions, or delegation.

## Current dependency trace

Paths below are relative to the repository root. This trace includes the current
working-tree adapter changes, rather than only committed code.

### AgentContext

Definition: `controller/core/boundary/boundary.go`, `AgentContext{Actor, Model}`.
Both `Boundary.Agent` and `Grant.Agent` carry it. `Surface.Agent` is a separate
`AgentConfig{Operations}` that enables agent-specific kernel behavior.

| Dependency | Current responsibility | Replacement owner |
| --- | --- | --- |
| `core/machine/system.go`: `AgentContext`, `Syscalls`, `checkSyscall`, `changesWaiter`, `agentContext` | Registers `system.agent-context`; accepts a resolved model only from a configured agent; prevents changing an inherited model | Adapter configuration contract and adapter validation |
| `core/machine/authority.go`: `Narrow` | Updates the waiting frame's model metadata; copies it into ordinary calls and reference invocations | Explicit procedure arguments and adapter-owned state |
| `core/machine/system.go`: `narrowing` | Interprets `inherit_model`; identifies agents; validates available provider operations; sets child actor/model | Interpreter composition of adapter contracts |
| `core/machine/continuation.go`: `Continuation`, `agentName`, `agents`, `continuation` | Constructs agent identity, model metadata, and an available-agent catalog | Adapter contracts plus operator-provisioned procedure configuration |
| `core/machine/machine.go`: `release` | Adds the same agent fields to the initial continuation envelope | Ordinary input payload; no special agent envelope |
| `controller/interp/main.go`: `envelope`; `controller/interp/delegate.go`: `runDelegate` | Uses `Agent` for omitted actor/model and `Agents` to validate choices | Adapter's delegate wrapper supplies resolved configuration; interpreter selects a configured adapter contract |
| `controller/codex/resident/main.go` | Reads inherited model, observes the provider-resolved model, publishes it through `system.agent-context` | Adapter state and an ordinary configuration result |
| `controller/claude/resident/main.go`: `resolvedModel`, `decide`, provider request construction | Resolves model and publishes it through `system.agent-context` | Adapter state and an ordinary configuration result |
| `runtime/registry/registry.go`; `skyra` launcher | Validates `agent.operations` and stores `AgentConfig` on surfaces | Validates provisioned adapter/procedure definitions outside the generic kernel |

Current flow: registry agent declaration → Surface.Agent → adapter model
publication → syscall modifies waiting Boundary.Agent → interpreter reads it →
narrow copies it into Grant.Agent → reference invocation copies it into child
Boundary.Agent → child continuation delivers it to the adapter.

### ToolBinding and WorkTools

Definition: `core/boundary/boundary.go`, `ToolBinding{Target, Contracts}` and
`WorkTools *map[string]ToolBinding` on both Boundary and Grant.

| Dependency | Current responsibility | Replacement owner |
| --- | --- | --- |
| `core/boundary/tools.go`: `Internal` | Exempts system and configured agent operations from the work-tool restriction | Explicit kernel capability scope; no agent exemption |
| `core/boundary/tools.go`: `AvailableTools` | Builds model-facing names, interpreter aliases, procedure titles, and selectable work tools | Adapter tool-catalog contract |
| `core/boundary/tools.go`: `ResolveTool` | Resolves names and transport aliases such as `repo_write` | Adapter binding contract |
| `core/machine/system.go`: `narrowing` | Resolves `tools`, preserves omitted versus empty selection, rejects widening, stores concrete bindings | Adapter resolves labels; generic narrowing validates concrete targets/contracts |
| `core/machine/authority.go`: `Narrow` | Carries the work-tool ceiling through ordinary calls and references | Generic restricted reach inherited by every invocation |
| `core/boundary/tools.go`: `PermitsTool`; `core/machine/interpret.go`: `interpret` | Rejects calls outside the target/contract ceiling | Generic invocation-capability check |
| `core/machine/continuation.go`: `workToolNames`, `Continuation`; `machine.go`: `release` | Serializes the ceiling back into model-facing labels | Explicit adapter input/result; generic envelope exposes reachable contracts |
| Codex and Claude `resident/main.go`: `tools`, prompt construction | Filters tools offered to the provider and conditions repository prompt text | Adapter configuration and binding contracts |
| `controller/codex/proceduretools/tools.go`: `Allows`, `Procedures`, `Payload` | Applies labels, discovers interpreter procedures, and translates model arguments into actual calls | Shared adapter support code, outside core |
| `runtime/registry/registry.go`; `skyra` | Marks `Surface.WorkTool`; automatically marks bundles | Adapter catalog/procedure provisioning metadata, outside generic Surface |

Current flow: registry work-tool metadata + reach → kernel builds labeled catalog
→ `narrow.tools` resolves labels into bindings → Grant/Boundary carries ceiling →
continuation carries labels to adapter → adapter filters advertised tools →
intake separately enforces concrete bindings.

### Delegated

Definition: a boolean on Boundary and Grant, set through `system.narrow`.

| Dependency | Current responsibility | Replacement owner |
| --- | --- | --- |
| `interp/delegate.go`: `runDelegate` | Sends `delegated:true` in `narrow`, then invokes the returned reference | Interpreter performs explicit configure/narrow/invoke/result composition |
| `core/machine/system.go`: narrow schema and `narrowing` | Checks registered agents/operations, applies a 64-frame guard, propagates the flag with logical OR | Procedure selection validation; generic execution-depth limit |
| `core/machine/authority.go`: `Narrow` | Propagates the flag through calls and grants | Delete propagation |
| `core/machine/continuation.go`; `machine.go`: `release` | Sends the flag to adapters | Explicit adapter session and lifetime configuration |
| `core/machine/execution.go`: `edgeSocket` | Adds Boundary.Origin to the socket key so children and siblings get fresh residents | Generic resource-instance identity/lifetime, explicitly acquired by contract |
| `core/boundary/resident.go`: `RenderResident` | Removes the native image-read exception for child providers | Explicit provider resource grants; no renderer agent-mode branch |
| `codex/resident/main.go` | Disables native tools, changes event handling, stops the provider before completion, omits thread metadata from result | Adapter run/session contract and explicit resource close |
| `claude/resident/main.go` | Avoids loading/saving the parent's session and resuming it | Adapter session policy and per-invocation state |

Current flow: interpreter `delegate` → narrow flag → Grant → child Boundary →
provider socket identity + resident sandbox + continuation → adapter session,
native-tool, result, and cleanup behavior. One boolean currently controls all
these independent responsibilities.

### Tests and documentation that depend on this representation

- `controller/core/machine/delegate_test.go`: inheritance, contract ceilings,
  empty tool sets, alias resolution, model publication, and socket isolation.
- `controller/core/boundary/boundary_os_test.go`: native image denial under
  `Delegated`.
- `controller/interp/delegate_test.go`: expects `inherit_model` and `tools` in
  emitted narrow requests.
- `controller/skyra/delegate_test.go`: real runtime acceptance; inspects child
  Boundary.Delegated and Boundary.WorkTools.
- `controller/codex/resident/delegate_test.go` and
  `controller/claude/resident/delegate_test.go`: provider/session behavior.
- Both adapters' `tool_visibility_test.go`, Codex `main_test.go`, and
  `controller/codex/proceduretools` coverage: catalog visibility and dispatch.
- `DELEGATION.md`, adapter tool descriptions, launcher registrations, and
  `controller/experiments/appserver/tool_surface.py`: protocol and acceptance
  expectations must be updated with the implementation.

## Proposed ownership boundaries

| Concern | Owner |
| --- | --- |
| Executable identity and declared input contracts | Registered Surface |
| Execution semantics and permitted lifetime inputs for an operation | Selected contract `(target, schema)` |
| Concrete lifetime options requested for an execution | Invocation payload, bounded by the contract and effective Boundary |
| Acquired resource identity, owner, endpoint, authority, and termination condition | Kernel-maintained resource instance |
| Which target/contract can actually be invoked | Boundary reach and generic kernel intake |
| Files, sockets, services, scratch access | Boundary.Authority and generic sandbox rendering |
| Human/model tool names, descriptions, aliases, argument translation | Adapter binding contracts |
| Provider/model selection and resolution | Adapter configuration contracts and state |
| Parent session reuse or fresh session | Adapter run/session contracts |
| Goal-only child input, selection, composition, and failure outcome | Interpreter delegate procedure |
| Resident acquisition, connection custody, release, isolation | Generic runtime resource operations composed by adapters/procedures |
| Native provider behavior and compatibility requirements | Provider support code outside core/boundary |

### 1. Adapter contracts expose configuration and bindings

Both adapters expose equivalent operations, even if their provider protocols
differ. Use disjoint object wrappers so payloads fit exactly one contract.
Illustrative operation names and payloads:

```json
{"configure":{"model":"selected-model","session":"fresh","bindings":[]}}
{"bindings":{"names":["read","repo_write"]}}
{"run":{"configuration":"adapter-owned-reference","text":"goal"}}
```

`configure` resolves a model and session policy and returns an invocation-local
configuration reference and the resolved model. The reference is validated
against adapter state and the invocation scope; a plain string is not itself
authority. It must not be a global mutable setting shared by concurrent tasks.

`bindings` returns named bindings containing target, concrete contract, input
translation, and provider-facing schema/description. This is a declared
adapter contract with JSON input/output, not `boundary.ToolBinding`. Implement
shared binding logic under adapter support code, including moving the relevant
catalog logic out of `core/boundary/tools.go`.

Resolve aliases deterministically and reject collisions. A binding can select
one interpreter contract without selecting all contracts of the interpreter.
Returned bindings must be checked against actual held capabilities before a
child scope is created. Neither adapter declarations nor configuration inputs
can manufacture kernel authority.

Omitted tools inherit the caller's effective binding selection; an explicit
empty list selects no work bindings. The adapter still needs its transport
capabilities, supplied explicitly in its scoped invocation.

Model-generated `delegate` arguments remain goal/actor/tools. The adapter's
binding translates these into an interpreter call containing the caller's
resolved adapter configuration. That configuration comes from the adapter's
own recorded configuration/state, not model-supplied inheritance metadata.
An explicit actor uses that adapter's default; an omitted actor uses the
current adapter and resolved model. No `system.agent-context` is needed.

### 2. Kernel narrowing restricts generic target/contract capabilities

Replace the two overlapping mechanisms, Reach plus WorkTools, with reach that
can authorize a subset of a surface's declared contracts.

A concrete proposed representation:

```go
type ReachEntry struct {
    Surface    Surface            // pinned identity and declared contracts
    Contracts  []*contract.Schema // authorized subset; empty means no calls
    Expandable bool               // whether actor-owned extensions enter this scope
}
// Boundary.Reach: map[string]ReachEntry
// Grant carries a restricted reach snapshot, not a model-facing tool catalog.
```

Registration/bootstrap creates initial entries. Ordinary calls preserve the
same scope. Narrow accepts target names and authorized contract identities,
checks them against the parent's entries, and returns a reference to the
restricted scope. Omission means inherit; an explicit empty contract selection
does not mean all contracts. Narrowing cannot change executable identity or
turn Expandable from false to true.

Select payload contracts against the full declared surface contract set first,
preserving existing ambiguity faults. Then require that selected contract to
be authorized by the ReachEntry. Restriction must not make an otherwise
ambiguous payload appear valid.

In a fixed child scope, all entries are non-expandable: `extend-reach` can store
or expose an actor's new procedure in an eligible broader scope but cannot
authorize it in that fixed child. Normal expandable scopes retain existing
remember/extend behavior. This is a generic capability rule, not delegation
awareness. Define how reference invocation and ancestor extension propagation
preserve these bounds before removing WorkTools.

Include system operations and adapter transport contracts explicitly when
constructing a scope. They receive no automatic provider/agent exemption.
Every descendant, including an indirectly invoked actor or procedure, inherits
the restricted scope. A reference cannot resurrect contracts absent from the
current caller's capability bounds. Test stored-reference escapes as well as
new narrow requests.

The exact wire format requires updating release derivation, identity encoding,
and replay fixtures. Preserve Surface identity independently of the authorized
contract subset; do not silently rewrite pinned registrations to narrow them.

### 3. Interpreter delegate composes ordinary contracts

Keep the public `delegate` contract and existing failure representation.
Its execution is an explicit recorded composition:

1. Resolve the requested adapter from provisioned procedure definitions and
   obtain its configuration/binding contracts. Reject unavailable choices.
2. Resolve caller/default model and requested bindings through adapter contracts.
3. Ask generic `system.narrow` for the concrete invocation capability subset and
   OS authority. Pass neither `delegated`, `inherit_model`, nor tool labels.
4. Invoke the scoped adapter with a fresh session configuration and only the
   goal as task context. Do not forward parent history, images, or session IDs.
5. Complete required resource cleanup, then return the child observation or
   `{failed:{kind,cause}}` to the caller.

The procedure's recorded calls and interpreter frame position supply resumable
control. Existing `runDelegate` is currently an embedded Go implementation of
the composition, not a JSON-installed DAG; migrate it to contract composition
without introducing a delegation state into the machine.

Keep the adapter responsible for its provider handshake, run, and stop sequence.
The interpreter waits for the adapter's terminal return rather than duplicating
provider protocol logic. Success requires completed cleanup under the chosen
lifetime policy. Preserve the existing distinction between ordinary child
failures and integrity/persistence failures that stop execution.

### 4. Separate session policy, resource lifetime, and OS permissions

Use explicit adapter session policy (`fresh` or `resume` with an adapter-owned
session reference). Model metadata and result shape belong to the adapter
contract. A delegate procedure can project the task observation without requiring
a boolean to change every adapter return.

Fresh sessions require fresh provider resources when their ambient state or
sandbox differs. Generic resource acquisition returns an instance handle scoped
to its owning execution scope. Connection operations use that handle; close
releases it. Resource identity comes from recorded acquisition/ownership, not
`Boundary.Delegated`. Repeated resumes of one acquisition reuse it; sibling
acquisitions do not. The parent remains alive while waiting.

#### Contract policy and kernel-maintained resource lifetime

A target identifies a program; a contract `(target, schema)` selects an operation
that program implements. A program can behave differently for each contract.
Execution semantics that vary by operation belong to the selected contract,
rather than an actor-wide resident mode. This does not make a contract a source
of authority: Boundary still bounds which operation can be invoked and which
resources it can acquire.

For example, one target may expose the following contracts:

| Contract | Behavior |
| --- | --- |
| `inspect` | Execute once and return |
| `start` | Acquire a persistent process and return a scoped handle |
| `exchange` | Use an existing process through an authorized handle |
| `stop` | Release that process through an authorized handle |

These are illustrative operations, not final wire names. Generic acquisition,
connection, and release are kernel syscall contracts; adapters compose them to
implement their own declared operations. A persistent resource can outlive many
ephemeral procedure steps without keeping the procedure's own process alive.

Separate declaration, request, and enforcement:

- The registered target pins executable identity and its declared contracts.
- The selected contract declares supported execution semantics and permitted
  lifetime inputs. Its policy does not apply to sibling contracts automatically.
- The invocation supplies concrete lifetime options within the selected contract
  and the caller's effective Boundary.
- The kernel records the acquired resource's identity, owning execution scope,
  endpoint, effective authority, and termination condition, then maintains that
  resource until the recorded condition occurs or the resource dies.

The kernel allocates the endpoint and returns a scoped handle. A port number,
process ID, or plain string alone is not authority to connect, inspect, or stop.
The owning scope can be a procedure invocation or a longer-lived conversation
scope when authorized. A call cannot select an arbitrary scope or gain a longer
lifetime merely by supplying its name.

Lifetime conditions must be directly observable by the kernel: owning-scope
completion or fault, cancellation, deadline expiry, or explicit release. An
adapter observes provider-specific events and invokes release; the kernel does
not interpret provider messages. Later inspection or connection calls preserve
the established lifetime. Changing it requires an explicit authorized operation.

Define cancellation, process death, runtime restart, and interrupted acquisition
or release before migration. Recorded ownership must let recovery reclaim the
correct resource without closing the waiting parent or a sibling, and replay
must not blindly duplicate an acquisition. A dead process's ambient state is not
reconstructed merely by restarting its executable; resource loss and any
authorized restart must be explicit outcomes. Record acquisition and release
outcomes and define how pending lifetime conditions are recovered.

Provider sandbox requirements are explicit grants/launch policy. A fresh task
with no image capability receives no workspace/scratch image grant. Provider
native tool disabling remains adapter/provider configuration, with the sandbox
providing independent enforcement. No renderer branches on agent/task mode.

Move `core/boundary/resident.go` and `claude_resident.go` provider-specific policy
into provider runtime support as an intermediate step. The final design uses
generic authority rendering with explicit provider resource grants. Moving files
alone does not fix `runtime/instance/startResident` replacing the generic profile
with a provider profile. In particular, the current Claude profile is allow-default
except for the runtime floor, and ignores its ClaudePaths argument; it must be
accounted for explicitly rather than described as Boundary-derived enforcement.

Apply runtime frame/resource limits generically, or procedure-local budgets
explicitly. Remove the delegation-only depth branch from `narrowing`.

## Migration sequence

1. Add adapter configuration, bindings, and run contracts with fixtures for both
   providers. Keep existing contracts working during migration; new wrappers
   must not overlap old input schemas.
2. Implement generic restricted reach and extension rules. Prove equivalent
   enforcement before retiring WorkTools. Update all reach consumers, registry
   loading, release verification, continuation construction, and replay encoding.
3. Add generic resource syscall contracts, contract-specific execution semantics,
   kernel-maintained instance ownership and lifetime conditions, and explicit
   provider grants. Verify parent/child/sibling isolation, cleanup, cancellation,
   and interrupted acquisition/release recovery.
4. Rewrite interpreter delegation and adapter delegate bindings to compose the
   new contracts, preserving public omitted/empty tools and actor semantics.
5. Remove AgentContext, WorkTools, Delegated from Boundary and Grant; remove
   agent-context syscall, inherit_model/delegated/tools from narrow, specialized
   envelope fields, kernel catalog/alias logic, and delegated socket/profile branches.
6. Remove Surface.Agent and Surface.WorkTool after provisioning and adapter
   catalogs no longer depend on them. Revisit resident Surface fields separately
   using the resource design; do not claim they disappear merely by moving types.
7. Update launcher bundles, DELEGATION.md, provider probes, and acceptance tests.

Existing conversations pin registrations and serialize old Boundaries. Do not
reinterpret those bytes as the new schema. Provision new conversations with the
new protocol; either retain versioned legacy decoding/replay or explicitly require
old conversations to finish on the old runtime. Choose and document that policy
before rollout. Release checks compare canonical Boundaries, so compatibility
cannot be handled only at adapter inputs.

## Acceptance criteria

- Core Boundary, Grant, generic continuation, and machine syscall definitions
  contain no agent model, tool label/catalog, or delegated mode.
- A non-AI actor can use the same restricted target/contract capabilities and
  resource lifetime operations.
- One target can expose both ephemeral and persistent operations through distinct
  contracts; selecting one does not apply its execution policy to its siblings.
- Separate acquisitions through the same contract have independent recorded
  identities, owners, and permitted lifetimes. Inspection and connection calls
  cannot implicitly overwrite acquisition policy.
- Resource handles enforce scoped access, and the kernel enforces recorded
  lifetime conditions without interpreting provider-specific messages.
- Cancellation, process death, runtime restart, and interrupted acquisition or
  release have recorded recovery semantics that avoid duplicate acquisition and
  reclaim the correct resources without closing the parent or siblings.
- Both adapters expose binding/configuration behavior through declared contracts;
  interpreter procedures compose them without provider-specific kernel checks.
- Omitted tools inherit, empty tools permit no work calls, and aliases resolve
  without widening capabilities or changing parent state.
- Direct calls, indirect actors, procedures, nested narrow scopes, stored
  references, and extend-reach cannot bypass a fixed contract subset.
- A selected interpreter contract does not authorize its siblings or dependencies.
- The original declared-contract ambiguity behavior remains intact.
- Fresh tasks do not load/save the parent's session or inherit parent transcript,
  images, native image permissions, or resident ambient state.
- Parent, child, and siblings have correct resource ownership; cleanup runs before
  success and failure returns without prematurely closing the waiting parent.
- Provider model inheritance is explicit adapter/procedure data and is reproducible
  from recorded configuration, with no kernel model-publication syscall.
- Replay/release derivation stays deterministic, and legacy compatibility has an
  explicit tested policy.

This spec adds no implementation tests by itself. Implementation should migrate
the dependency tests listed above into generic kernel capability tests, adapter
contract tests, interpreter composition tests, and real-runtime isolation tests.
