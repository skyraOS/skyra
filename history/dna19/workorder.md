# Work order: DNA-19 — boot-register the separate Codex executable

Status: implemented and deterministically verified. See README.md, DECISIONS.md,
issues.md and artifacts/verification.json. Live handshake, health, inference, fork, tool resume, tool-set migration and
concurrent-parent isolation scenarios passed with CLI 0.155.1. Universal provider
sealing and OS containment are not claimed.

## Outcome

Build on DNA-18.5 in this new directory. Move all host-root implementation and
host-only peers under `root/`. Build `codex/` as a separate executable binary that
receives and emits Skyra Moves. Boot registers its declarative profiles, binds the
external executable, invokes connection setup and exposes the admitted capabilities.

The persistent parent's invocation identity represents a conversation on the Skyra
side. The Codex binary maps that parent to its provider thread, keeps unrelated
parents isolated and preserves the association across continuation and child work.

Do not modify DNA-18.5 or earlier implementations in place. Use the existing
repository Go module; do not add a nested module. No sockets are required.

## Sources and precedence

- Accepted direction: [Codex registration proposal](../specs/2026-09-21/proposal/codex-registration.md),
  especially external-process separation, profile registration, thread isolation
  and contamination fencing.
- Baseline: DNA-18.5, its implementation,
  decisions, tests and captured evidence.
- Contracts: [shared proposal](../specs/2026-09-21/proposal/proposal.go),
  particularly Executable / ArgContract / OutputContract / FaultContract.
- Capability inventory: [Codex integration](../specs/2026-09-19/codex/codex-integration.md).
  Preserve the behavior; historical provider-version observations require fresh
  verification against the implementation environment.
- Identity/lifecycle evidence: DNA-16 work order,
  DNA-17 Codex implementation,
  acceptance tests and
  findings.
- Isolation background: [possibility-as-process](../specs/2026-09-21/possibility-as-process.md)
  and [persistent parent](../specs/2026-09-19/persistent-parent-and-separate-pools.md).

The accepted external-binary boundary takes precedence over the older in-process
Driver/HostCommand interfaces and the earlier host-kernel demux placement. Preserve
semantics, not those APIs. Record this ownership change explicitly in DECISIONS.md.

## Required directory and binary layout

```text
skyra-dna-19/
  workorder.md
  README.md
  DECISIONS.md
  issues.md
  demo.sh
  protocol/                   shared wire types and declarative contract vocabulary only
  root/                       all host/root responsibilities
    main.go                   executable entrypoint and boot wiring
    contracts.go              host owner declarations
    ...                       Kernel, RuntimeState, Memory, Active, Scheduler,
                              ExecutionContext, Bookkeeper, ports, trace
    parent/                   inherited host test/demo peer
    echo/                     inherited independent echo executable
    crosscontext/             inherited host activation peer
    creation/                 inherited host creation peer
    ..._test.go               host acceptance and subprocess integration tests
  codex/                      independent executable, not imported by root
    main.go                   Move input/output and lifecycle
    contracts.go              private actor/session declarations
    profiles.json             versioned, declarative profile descriptor
    ...                       profile handling, session demux, provider adapter
    internal/
      provider/               private provider protocol and transport, if useful
    testdata/
      fake-provider/          deterministic provider process fixture
    ..._test.go
  artifacts/                  captured traces and verification results
```

The exact split of implementation files is flexible. Preserve owner responsibilities;
placing code beneath `root/` does not merge the owners into RootProgram. Shared
`protocol/` is deliberately outside both executables because it defines their wire
boundary, not host runtime internals. Host-only packages all belong under `root/`.

Build targets from this directory:

```sh
go build -o artifacts/root ./root
go build -o artifacts/codex ./codex
go build -o artifacts/echo ./root/echo
```

`root` must not import `codex` implementation packages; `codex` must not import
`root` implementation packages. Shared declarations contain no runtime pointers,
callbacks, provider-native commands or OS handles. Stdio carries only framed Moves;
diagnostics use stderr. If Codex wraps a provider subprocess, it supervises/reaps
that subprocess; root supervises/reaps the Codex binary.

## 1. Relocate the baseline and retain its proofs

Copy DNA-18.5 into the new layout, update imports/build paths and preserve its
existing acceptance scenarios. Retain recoverable Move rejection, explicit fatal
transport failures, cross-context invocation, bounded creation, reservations,
binding generations, RootSink/frame/transport separation and shutdown behavior.

Keep Kernel / RuntimeState / Memory / Active / Scheduler / ExecutionContext /
Bookkeeper / Executable ownership. No host Active entry is required merely because
Codex is an external executable. The caller's runtime frames track its invocations
through the bound port. Any runtime inside the Codex binary is separate.

## 2. Declare and register usable argument profiles

Restore observation-based argument declarations or specify an equivalent concrete
schema. DNA-18.5's operation-only declaration and universal operation/text validator
are insufficient. Record the schema choice before implementing dependent behavior.

The descriptor must declare:

- Protocol/descriptor version and compatibility rules.
- Each profile's selector and required/optional observations, encodings and constraints.
- Result alternatives and failures associated with each profile.
- Continuation-only profiles and lifecycle restrictions.

Use an `operation` observation to select exactly one profile. Reject unknown,
ambiguous, duplicate or malformed observations before dispatch. Runtime-assigned
ExecutableID names the program; profiles do not create separate executable IDs.

Boot reads the descriptor shipped with the chosen binary, validates it, registers
it in Memory and installs admitted access into the caller's horizon. Expose those
registered contracts to authorized calling/planning code so it can construct
arguments. Prove that profile visibility and execution permission agree: an
advertised but ungranted profile is not executable through a narrower binding.
The external binary also validates inputs; its descriptor and implementation must
have conformance tests. A descriptor cannot self-grant authority.

## 3. Implement the top-level capability inventory

| Profile | Behavior and required proof |
|---|---|
| `connect` | Establish configured provider connection and handshake; return an opaque scoped connection reference/readiness. No conversation starts merely because boot registered the actor. |
| `health` | Probe connection responsiveness with bounded correlation; distinguish timeout/connection failure from inference readiness. |
| `start-thread` | Start the provider thread for an authorized stable logical parent and run initial input; preserve that parent/thread association. |
| `submit-input` | Run subsequent input against that same parent/thread; a command result does not terminate the conversation parent. |
| `fork-thread` | Create a distinct parent/thread association from authorized source history. Never alias the source's mutable session record. |
| `tools-add` | Produce a replacement provider thread with the complete updated tool set, within grantable authority. |
| `tools-remove` | Produce a replacement provider thread with the reduced tool set and revoke removed execution authority. |
| `tool-result` | Resume exactly the outstanding correlated dependency request through Bookkeeper's child-result path; never an unrestricted input route. |

For tool replacement, define the bounded migration contract before implementation:
what context is retained, which logical reference remains stable, when the old
binding is invalidated and what happens to pending work. An idle-only first
implementation may reject changes while work is outstanding. Do not silently
lose history, replay effects or treat historical API assumptions as current facts.
Unsupported provider features must return explicit errors and remain identified
as incomplete integration, not be advertised as successful implementations.

The executable translates profiles to provider operations internally. No provider
JSON-RPC vocabulary is added to Kernel or Scheduler. Implementations may have
private helpers; root communicates with the real separate binary in acceptance tests.

## 4. Make parent lifetime and per-command results explicit

Before implementation, declare the conversation-parent lifecycle and its relation
to command frames in DECISIONS.md and the relevant contracts. This is required work,
not a reason to silently reduce the task to one generic ask.

- A persistent parent is the logical conversation identity; individual profile
  commands and tool children have separate invocation/correlation identities.
- Preserve the parent across multiple completed input commands and tool resumes.
  Command Return settles that command only. Parent termination/revocation is explicit.
- Define how the initial parent is minted, how later commands are attached and how
  its authority is verified. Do not use a model-supplied bare ID as proof of ownership.
- DNA-18.5's local FrameID needs owner qualification and an authenticated external
  reference. Decide and document the concrete wire representation; do not equate
  Move.Identity, frame sequence, process ID and provider threadId.
- Keep canonical host frame/continuation state with Bookkeeper/context. Keep provider
  thread/turn/pending-call state inside Codex, scoped by the authenticated parent.
  Do not add an unrelated host session map that silently replaces Bookkeeper.
- The first implementation may serialize turns per parent. Concurrent independent
  parents must remain separate; a busy parent must not overwrite another's correlation.
- Define end-of-parent cleanup, outstanding-child disposition and late-event rejection.
  Restart/resume across process failure remains outside the first proof.

## 5. Enforce thread isolation in the trusted Codex binary

The host fences protocol ownership; it cannot inspect or guarantee private provider
thread selection. The external binary owns the isolation obligation:

```text
scoped parent P  ↔ provider thread T
scoped parent P2 ↔ provider thread T2
provider event T → only P's bound protocol output
```

Demux must validate thread, turn, pending request and live binding/incarnation.
Unknown, stale or conflicting correlations fail closed. Parent identities from
different host namespaces must not collide even if local sequence numbers match.
Never accept a provider threadId from an arbitrary caller as authorization to
continue, fork or reconfigure that thread. Child Return resumes its original
parent/thread without a new thread/start.

A shared provider process is permissible; shared process does not mean shared
conversation. Provider sealing between threads remains a declared trust assumption.
Use interleaved events from two fake-provider threads to prove binary routing, and
record separately what a live provider run does and does not establish.

## 6. Preserve the contamination fence and asynchronous protocol

Provider acknowledgments, partial text and progress do not settle invocations.
The executable keeps reading events while dependencies run; a scheduler turn never
waits for model completion. A permitted dependency emits a child Invoke through the
bound protocol, waits via Bookkeeper and continues with the correlated child result.

A forbidden native effect-bearing provider event taints the affected command and
emits a terminal Fault. Subsequent clean text/completion cannot produce a Return or
second result. Preserve DNA-17's rejection semantics, including a fault while a
parent command is waiting on child work. DNA-18.5's running-frame-only checks are
not sufficient for that case: define atomic fault settlement and descendant
invalidation so late child results cannot resume faulted work.

Keep contamination sticky for the affected work and explicitly define whether the
conversation is retired; do not let a later profile call silently clear taint.
Detection and result rejection do not prove the native effect was prevented.
Unknown provider requests must not trigger invented approvals. Effective provider
policy, allowed tools and configuration must be verified separately from this fence.

## 7. Boot, failure and shutdown

Boot prepares descriptor, reservations and endpoint before accepting ordinary
work; launch the binary and perform connect via a normal correlated Move. Mark
ready only after a validated response. A mismatched descriptor/version, launch
failure, handshake error or timeout must unwind live mappings/reservations and reap
all started processes. Tests must inspect cleanup, not merely observe a returned error.

An expected Move rejection must preserve unrelated progress. A provider failure
reported over a healthy actor port should produce a correlated Fault. Unexpected
binary/transport loss may retain DNA-18.5's explicit runtime-fatal policy; do not
claim process-crash isolation unless implemented and tested. Never retry a committed
side effect implicitly. Cancellation closes blocked reads/writes, releases scheduler
claims and waits for helper/provider reaping.

## Acceptance and completion artifacts

Use deterministic channels, scripted provider events and barriers. Timeouts guard
failure, never establish order. The fake provider runs behind the actual Codex
binary, so provider event translation, port serialization and process boundaries
are exercised together.

| Scenario | Required evidence |
|---|---|
| Separate executable | Independently built root/Codex binaries communicate with Moves; no implementation imports across the boundary |
| Registration | Boot loads real descriptor; caller can enumerate admitted profiles and form valid arguments; connect precedes workload |
| Contract enforcement | Invalid/unknown/ungranted profile rejected before external dispatch; valid profile reaches binary intact |
| Persistent conversation | Start followed by submit-input preserves logical parent/provider thread; command results leave parent live |
| Thread isolation | Interleaved P/P2 events route only to their own results; forged references and colliding local IDs fail |
| Dependency resume | Tool child has separate identity; one correlated result resumes existing parent/thread and pending call |
| Fork and tools | Fork does not alias source; tool replacement respects authority and documented migration/in-flight rules |
| Contamination | Forbidden event faults once; later final text, completion and child returns cannot yield clean success |
| Independent progress | Provider wait/rejected work does not block another eligible context |
| Cleanup | Startup rollback and explicit cancellation release reservations, close endpoints, release claims and reap processes |
| Regression | All DNA-18.5 acceptance scenarios still pass after relocation |

Completion requires:

- Runnable root and Codex programs, versioned descriptor, deterministic fake-provider
  fixtures, tests, and README commands for both binaries and each demonstration.
- DECISIONS.md describing lifecycle/schema/isolation decisions; issues.md identifying
  remaining provider/containment/recovery limits without treating them as proved.
- Captured real-process demo trace for registration, two commands on one parent,
  a second isolated parent, dependency resume and contamination rejection.
- `go vet ./...` and `go test -race -count=1 ./...` passing from this directory;
  captured results and successful independent builds.
- An opt-in live-provider integration command, with configured provider version and
  policy recorded. Run when configured and authorized; report unavailable credentials,
  version incompatibility or skipped live validation explicitly. Fake-provider success
  alone must not be described as verified live Codex compatibility.

This work order authorizes implementation decisions needed to make the bounded
contracts concrete. It does not require sockets, general OS services, arbitrary
parallel child joins, crash persistence or a general provider framework.
