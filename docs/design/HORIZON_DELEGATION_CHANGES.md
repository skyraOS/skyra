# Horizon inheritance and delegation visibility

## Purpose

Proposed changes based on the preceding source investigation. This document is a change plan, not a claim that fixes have been implemented or verified against the running application. Source locations below were reported in that investigation and may move during refactoring.

## 1. Show delegation in tool activity

### Reported issue

`delegate`, `remember`, and `procedure` are model-facing contracts routed through the `interpreter` target. The activity filtering in `controller/hosts/web/calls.go` (reported lines 343–348) excludes that target, as well as internal `system` and `system.ref.*` calls. Consequently, a delegation invocation can be recorded without appearing in the user-visible activity list.

### Required changes

- Identify visible tool activity by the selected model-facing contract, not solely by its underlying execution target.
- Show `delegate`, `remember`, and `procedure` with their invocation status and result or failure.
- Preserve invocation identifiers and parent/child relationships so delegation can be traced.
- Keep internal execution plumbing collapsed beneath the public invocation rather than exposing it as unrelated tool calls.
- Apply existing authorization and sensitive-data redaction rules to displayed arguments and results.

### Acceptance criteria

- A `delegate` invocation appears in activity while running and reaches a visible terminal status.
- Failed or cancelled delegation is visible rather than silently filtered out.
- Internal interpreter/system events do not create duplicate public activity entries.
- Tests cover contract-to-target routing and parent/child activity correlation.

## 2. Make Horizon the source of inherited capability authority

### Reported issue

The registered tool path already inherits authority through Horizon narrowing in `controller/core/machine/room.go`. However, provider-native capabilities can bypass that catalog and receive delegation-specific restrictions.

In `controller/codex/resident/main.go`, the investigation reported:

- Delegated threads disable native image viewing and web search (lines 500–501).
- Native `imageView` and `webSearch` activity is accepted for parents but treated as a violation for delegated children (lines 448–454).

This creates a parent/child capability asymmetry that is not derived solely from the inherited Horizon. Whether this adapter path explains the observed live session still requires runtime verification.

### Required changes

- Make Horizon-held capability bindings and restrictions the single source of execution authority for both parents and children.
- Derive the tools exposed by each agent adapter directly from the effective Horizon.
- Inherit the caller's authorized context and tool reach by default; apply explicit narrowing when requested.
- Remove capability exceptions based solely on whether an invocation is delegated.
- Represent provider-native capabilities through Horizon-governed bindings, or disable them consistently when they cannot satisfy the same authority and recording requirements.
- Preserve authorization checks, provenance, effect recording, and replay semantics for native capabilities; do not simply enable unrecorded child web access.
- Distinguish authorization from adapter support. If an adapter cannot execute an inherited capability, report that limitation explicitly rather than silently changing the child's authority.

### Acceptance criteria

- Equivalent effective Horizons produce equivalent authorized capability sets regardless of delegation status.
- An explicitly narrowed child cannot access excluded capabilities through provider-native paths.
- Every executed capability has an auditable authorized invocation and outcome.
- Integration tests cover parent/child web and image access, adapter limitations, and explicit narrowing.
- Runtime records establish which adapter path is used by the observed session.

## 3. Keep provider sessions in the agent adapter, not kernel policy

### Architectural direction

A provider session may be useful as an agent-adapter implementation detail. It should be configured from the effective Horizon, not become a separate source of tool authority or a generic session abstraction enforced by the kernel.

### Required changes

- Review kernel-level `delegated`, `inherit_model`, agent availability checks, and the `agent-context` syscall for agent-specific policy.
- Move provider-session creation, model selection defaults, and provider-specific behavior into the agent abstraction where appropriate.
- Retain generic kernel enforcement of authority, custody, effect execution, and replay. Validate existing authority on each invocation; do not implicitly invoke narrowing on every invocation.
- Define explicitly which context and capability bindings are inherited and which are narrowed; do not equate inheritance with copying an entire provider transcript or granting unavailable capabilities.
- Ensure session reuse or recreation cannot expand authority beyond the effective Horizon.

### Acceptance criteria

- Kernel authority decisions do not depend on provider-session identity or delegated-thread status.
- Agent adapters consume Horizon-derived context and capability configuration.
- Tests demonstrate that session reuse, fresh sessions, and nested delegation preserve the same authority constraints.

## 4. Expose narrowing as an explicit Horizon syscall

### Architectural direction

`narrow` should be an explicit syscall exposed by the Horizon, not a method automatically invoked for every invocation. The interpreter invokes it when attenuation is requested, analogous to its explicit state-read syscall when implementing `remember`. This is the requested architectural direction; the exact syscall contract and current state-read path still require source verification.

### Required changes

- Expose narrowing through the Horizon's syscall surface rather than an unconditional invocation-time method.
- Ordinary invocation uses the existing effective Horizon; delegation inherits it by default without an implicit narrowing operation.
- When a caller requests restricted context or capability reach, the interpreter explicitly invokes the narrowing syscall and uses the resulting Horizon for the intended execution scope.
- Keep per-invocation authority validation separate from narrowing. Removing automatic narrowing must not remove authorization checks.
- Define narrowing as monotonic attenuation: it cannot add authority, restore excluded capabilities, or mutate the parent's Horizon.
- Record explicit narrowing requests and outcomes with sufficient provenance for semantic replay, including failures.
- Keep provider-session construction and agent defaults outside this generic syscall contract.

### Acceptance criteria

- Ordinary invocations and unrestricted delegation do not invoke the narrowing syscall.
- Explicit attenuation invokes the syscall and supplies its resulting Horizon to the intended execution scope.
- Attempts to expand authority fail; the parent's Horizon remains unchanged.
- Per-invocation authorization still rejects access outside the effective Horizon, including provider-native paths.
- Tests distinguish inheritance, explicit narrowing, and authorization validation, and cover nested delegation and replay.

## 5. Select execution policy by contract and maintain resource lifetime in the kernel

### Architectural direction

A target identifies a program; a contract `(target, schema)` selects an operation
that program implements. One program can expose both ephemeral operations and
operations that acquire or use a persistent process. Execution policy that varies
by operation belongs to the selected contract, not an actor-wide resident flag.

The contract declares supported execution semantics and permitted lifetime inputs.
An invocation selects concrete options within that contract and the caller's
effective Horizon. The kernel records and enforces the resulting lifetime on the
acquired resource instance. Declaring a contract does not grant authority.

### Required changes

- Expose generic resource acquisition, connection, and release through kernel
  syscall contracts. Adapters and procedures compose these operations.
- Allow one target to expose, for example, `inspect`, `start`, `exchange`, and
  `stop` contracts with distinct behavior; do not apply one contract's execution
  policy to its siblings merely because they share a target.
- Keep procedure steps ephemeral when appropriate while their acquired provider
  process remains alive across those steps.
- Record each resource's identity, owning execution scope, endpoint, effective
  authority, and lifetime condition. The kernel allocates an endpoint and returns
  a scoped handle; a port number or process ID alone is not authority.
- Support conditions the kernel can observe, such as owning-scope completion or
  fault, cancellation, deadline expiry, and explicit release. The owner may be a
  procedure invocation or a longer-lived conversation scope, within held authority.
- Have adapters translate provider-specific completion events into explicit
  release calls. Keep provider-message interpretation outside the kernel.
- Preserve the lifetime established at acquisition across later connection or
  inspection calls. Any lifetime change requires an explicit authorized operation.
- Define cleanup and recovery for cancellation, process death, runtime restart,
  and interrupted acquisition or release, without duplicate acquisition or release
  of another scope's resource. Recovery must report lost ambient process state;
  restarting a process does not reconstruct it by itself.

### Acceptance criteria

- One program supports ephemeral and persistent operations through distinct
  contracts without an actor-wide mode switch or sibling authority expansion.
- Multiple acquisitions through the same contract have independent identities,
  owners, and permitted lifetimes; ordinary calls cannot overwrite those lifetimes.
- The kernel enforces recorded termination conditions and scoped handle access.
- Cancellation and recovery reclaim the correct resources while preserving the
  waiting parent's resources and recorded results.
- Acquisition, lifetime enforcement, and release have deterministic recorded
  semantics for replay; replay does not blindly launch a duplicate process.

## Suggested implementation order

1. Add regression tests for missing delegation activity and parent/child capability asymmetry.
2. Fix public contract activity visibility and invocation correlation.
3. Map provider-native capabilities into Horizon-governed execution.
4. Remove delegation-specific capability exceptions after equivalent enforcement exists.
5. Replace unconditional invocation-time narrowing with an explicit Horizon syscall, preserving independent authority validation.
6. Refactor agent/session policy out of the kernel where appropriate.
7. Implement contract-specific resource operations and kernel-maintained lifetime conditions before removing resident lifecycle fields from Surface.
8. Verify with runtime traces and host integration tests, including nested delegation, resource interruption/recovery, and replay.

## Scope

This plan does not authorize broadening capability ceilings, bypassing sandbox restrictions, or weakening recording and replay guarantees. The goal is consistent Horizon-derived inheritance and truthful user-visible activity.
