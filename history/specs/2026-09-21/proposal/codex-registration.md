# Proposal: boot registration of an external Codex executable and its profiles

Status: draft for discussion; not implemented. Builds on DNA-18.5 and the earlier
Codex surface contracts. Codex integration runs in a separate process, built as a
different binary. Its interoperability boundary is the Skyra Move protocol.

## What is registered

Boot registers one addressable executable plus declarations of the argument
profiles it accepts. Those declarations let the runtime expose the available
capabilities, validate a requested profile's observations and dispatch the correct
arguments. The external binary interprets those arguments and performs the work.

```text
host runtime                                      separate binary
Memory: Codex executable + ArgContracts            Codex protocol executable
caller horizon: authorized bound endpoint   Moves → profile interpreter
Bookkeeper: parent/child correlation         ← Moves provider/session implementation
                                                       ↕ private provider transport
                                                   Codex app-server, if used
```

This does not require a Codex ExecutionContext in the host's Active map, an
in-process CodexDriver, Go handler registration, or a second local echo-style
activation bridge. Active owns host contexts and their actual-port reservations;
Memory owns the executable descriptor; the caller's context owns its mapping and
frames. The external actor owns its process-local implementation. If that binary
uses a runtime internally, its state and authority namespace are separate.

“Unique actor” means the distinct provisioned executable instance for this boot.
It does not mean all conversations share one provider thread. Executable identity,
logical parent identity, process identity and provider session identity differ.

## Carry forward the top-level capabilities

The [September-19 integration](../../2026-09-19/codex/codex-integration.md)
consolidated earlier start-app-server / start-thread / continue-thread /
check-app-server-health names into `codex.connect`, `codex.start`,
`codex.continue`, `codex.health`, and `codex.fork`. It also proposed tool-set
addition/removal. Preserve that behavior inventory; do not replace it with a
single generic `ask` operation.

The [September-21 executable contract](proposal.go) makes a profile an
`ArgContract` of one executable, explicitly naming start-thread, submit-input and
tool-result. Proposed public profiles below reconcile those names. Names are
selecting observations, not extra executable addresses or required native methods.

| Proposed profile | Earlier capability | Required argument content | Intended result |
|---|---|---|---|
| `connect` | `codex.connect` | Provisioned connection/configuration reference | Logical connection reference / readiness |
| `health` | `codex.health` | Connection reference, response deadline | Connection responsiveness outcome |
| `start-thread` | `codex.start` | Connection reference, authorized parent reference, thread settings, initial input, allowed tool definitions | Parent/thread association and initial turn outcome |
| `submit-input` | `codex.continue` | Parent/thread reference, next input | Turn outcome for that conversation |
| `fork-thread` | `codex.fork` | Authorized source parent/thread reference, optional boundary/settings | New logical parent/thread association |
| `tools-add` | proposed `codex.tools.add` | Parent reference, tool definitions and permitted targets | Explicit replacement-thread outcome |
| `tools-remove` | proposed `codex.tools.remove` | Parent reference, tool identities | Explicit replacement-thread outcome |
| `tool-result` | child-result continuation | Bound pending-call correlation and child observation | Continue the existing provider work |

Connection references are opaque scoped protocol data, not file descriptors or
unrestricted host handles. Settings cannot widen the boot-provisioned authority.
Tool-set changes retain the earlier new-thread proposal; migration of context and
pending work remains open. Historical provider-version observations are not a
claim about the currently installed provider API.

`tool-result` is a continuation profile. It cannot authorize an arbitrary caller
to fabricate a dependency result. Normal child Return processing resolves the
waiting parent and delivers a correlated resume to its existing endpoint. Encoding
that resume as the declared tool-result argument form is an explicit contract to
specify; it must not bypass Bookkeeper or mint an unrelated parent invocation.

## Descriptor and boot registration

Use a versioned declarative descriptor packaged with the external executable and
admitted by trusted boot configuration. That avoids requiring a model call or an
unregistered discovery operation to learn how to invoke the program. A future
protocol-based descriptor exchange can use a standard bootstrap profile; it is
not required for the first proof.

The conceptual descriptor is the shared proposal's shape:

```text
Executable {
    ID: runtime-assigned CodexExecutableID
    ArgContracts: [connect, health, start-thread, submit-input, ...]
    OutputContract: declared result observations
    FaultContract: declared failure outcomes
}
```

Each ArgContract must describe the observations required by its profile, including
names, accepted encoding, required/optional values and constraints. The exact
observation vocabulary is a design task. Merely recording profile names is not
sufficient for the runtime to construct or validate arguments.

1. Boot selects the external binary and its compatible descriptor, validates the
   protocol version and admits only the intended profiles/configuration.
2. Register the executable and accepted argument contracts in Memory. Provision its
   bound endpoint into the authorized caller's horizon; Active owns reservations.
3. Launch the separate binary and attach its ingress to normal Move admission.
   The binary exposes no runtime object pointers or provider-native methods.
4. Invoke its `connect` profile with declared observations over the bound port.
   The correlated Return establishes readiness. Only then admit normal workload.
5. Expose the registered profiles to authorized planning/calling code as executable
   capabilities. An invocation targets the executable ID and supplies the matching
   profile arguments; runtime validation precedes dispatch.
6. On startup failure, invalidate the binding, remove the failed live registration,
   release reservations and reap the process. Registration grants no model-generated
   authority and creates no conversation thread on its own.

For this draft, use an `operation` observation to select exactly one registered
profile. Unknown, malformed or ambiguous profiles fail before dispatch. This is
an explicit proposed resolution of the shared sketch's open selection rule.
Profile visibility is not authority: the binding's admitted profile set must also
be enforced if some callers receive a narrower set than the executable advertises.

## Invocation and thread identity

Keep the historical stable parent mapping:

- DNA-16 work order: a Codex-backed
  InvocationID owns a thread; tool work gets a child identity and resumes its parent.
- DNA-17 implementation:
  `live map[InvocationID]*codexLiveInvocation`, reverse `threads` lookup, and opaque
  thread/turn/pending-call state.
- DNA-17 findings and
  [persistent-parent sketch](../../2026-09-19/persistent-parent-and-separate-pools.md):
  logical invocation, worker and provider resource have different lifetimes.

The persistent parent's InvocationID is the Skyra-side conversation identity.
The external executable maps it to the provider-assigned threadId. It is not
necessarily the same literal value. Child calls and individual profile requests
have their own correlation identities; neither replaces the conversation parent.

```text
registered external Codex actor C
  logical parent P ↔ provider thread T
    start-thread request S
    submit-input request U
      dependency child Q → Return(Q) → resume pending work under P
  logical parent P2 ↔ different provider thread T2
```

**Thread isolation is a trusted-binary obligation.** Because Codex is an external
executable and the provider protocol is private to it, nothing in the host enforces
that two unrelated logical parents get two unrelated provider threads — the *binary*
must key one thread per logical parent and never cross them (`P↔T`, `P2↔T2`), and a
child return must never create a new thread. This is the same demultiplexing
obligation named in possibility-as-process §19, now discharged inside the binary.
The first proof must configure the binary so this holds and verify it, not assume it.

A profile request completing must not accidentally terminate the persistent
conversation parent. Conversely, an already-terminal parent cannot be reopened by
presenting its old ID. DNA-16/17 demonstrated one turn per parent with tool resumes;
the broader start/continue capability set requires explicit parent lifetime and
per-command result semantics. That is a required contract, not an implemented claim.

DNA-18.5 uses local FrameID rather than the older InvocationID type. Host-side
correlation therefore needs an owner-qualified parent reference, and the external
wire contract needs a scoped logical reference bound to the authenticated port.
The bare frame number or an untrusted thread-id observation is insufficient.
Exact encoding and multi-turn parent lifecycle remain open; do not invent a new
parent on every child resume or derive identity from the OS process.

## Example call shape

Illustrative observations, not a finalized schema or literal JSON byte encoding:

```text
Invoke {
  Target: CodexExecutableID
  Args: [
    operation = start-thread,
    connection = provisioned-connection-reference,
    parent = authorized-logical-parent-reference,
    settings = admitted-thread-settings,
    input = "Inspect the parser and propose a change",
    tools = delegated-tool-descriptions
  ]
}
```

The runtime knows this profile's argument contract and how to send a Move. The
separate executable knows how to start the provider thread. Provider JSON-RPC,
thread/turn IDs and streaming normalization stay behind its port. If app-server
is used, its lifecycle API is an implementation reference, not the Skyra contract.
[Official app-server documentation](https://learn.chatgpt.com/docs/app-server)

## Required work and proof

DNA-18.5's `ArgContract{Operation string}` plus its universal two-observation check
cannot represent these profiles. Restore the observation-based declarations and
implement concrete validation, or explicitly introduce an equivalent schema.
Specify profile-specific result alternatives as well: the single output-key
fixture is insufficient. Keep the existing Invoke/Return/Fault protocol boundary.

**Contamination fence (carried forward from DNA-17).** When the external binary
detects a forbidden native provider action — a `commandExecution`, `fileChange`,
`mcpToolCall`, `webSearch`, or any effect-bearing item it was not authorized to
perform as a mediated Move — it emits a terminal Fault for the bound frame instead of
a Return. The caller-context frame then settles faulted, so its fence rejects any
*later* clean agent message or `turn/completed` for that turn: a contaminated turn
cannot produce a Skyra Return. This is DNA-17's rejection proof (`codex.go`
contamination + the settlement fence) at the external-binary boundary. It does NOT
confine the native tool — that needs the containment work called out below — but it
guarantees a contaminated turn cannot launder a native side effect into a clean
result. This is the minimum structural check available before real containment lands.

First prove with a separate fake-Codex binary that consumes the same declared
profiles, then verify the provider integration:

- Boot registers/exposes the intended profile inventory and waits for connect.
- Registered observations reach the separate process intact; malformed profiles
  are rejected before the external executable receives them.
- Start and continue refer to the same logical parent/thread, while independent
  parents remain separate. Fork creates a new association.
- A child result resumes the correct parent and pending call without creating a
  second thread; fabricated or stale correlations fail.
- Provider events do not masquerade as multiple terminal Returns. Profile-command
  completion is distinguished from conversation-parent termination.
- Contamination fence: the fake binary emits a forbidden native action mid-turn; it
  faults the bound frame, and a subsequent clean agent message and `turn/completed`
  produce no Return and no second outcome. (Result-path containment even though the
  native tool itself is not yet confined; the provider-policy check below still applies.)
- Startup rollback, cancellation, binding invalidation and subprocess reap work.

Do not import the old in-process Driver/HostCommand APIs into the new runtime.
Port their semantics into the separate protocol-speaking binary. Host runtime
changes concern registration, contract validation, binding and correlation.
Provider containment, effect mediation, crash recovery and live API compatibility
still require their own evidence; declared profiles alone do not establish them.
