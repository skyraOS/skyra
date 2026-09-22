# Proposal 1 — kernel driver registration

Status: aligned with [scheduler-loop proposal 3](../loop/loop-proposal.md)
and its [placement-gated continuation algorithm](../loop/placement-gated-continuation-algorithm.md).
This records registration ownership and the minimal boot composition for a working
local loop. It is not an implementation order; no runtime changes or test results
are claimed. The current-v14 review below describes the inspected baseline, while
the proposed composition uses the new loop boundaries.

Review scope: only `skyra-dna-v14` and the supplied `loop improvements` proposal
directory. No kernel implementation outside that scope was inspected. Within
v14, kernel privileges are declared, but no kernel driver or boot handoff consumes
them. The review is based on source inspection; tests were not executed.

## Current registration path

```text
NewCTU(rootAddress, rootRegistration, authority, complete)
  create DriverRegistry
  register root driver and input validator
  give the same registry to ContinuationController as Resolver
  construct Scheduler
  return CTU exposing Submit, RegisterDriver, and Snapshot

CTU.RegisterDriver(address, registration)
  DriverRegistry.Register(address, registration)

Controller.Validate / Dispatch
  Resolver.Resolve(address)
  validate input / authorize each handoff and execute
```

`driver.PrivilegedCTU` declares `Registrar`, `Admit`, and `Fault` capabilities.
Nothing in this directory constructs or delivers that bundle. The current root
driver is an ordinary registered `Driver`; the constructor does not establish
that it is a trusted kernel.

## Strengths to preserve

| Existing behavior | Why it helps |
| --- | --- |
| Separate `Resolver` and `Registrar` interfaces | Expresses the intended distinction between endpoint lookup and installation. |
| Registry protected by `sync.RWMutex` | Serializes registration and permits synchronized lookup, including concurrent duplicate attempts. |
| Driver and validator installed together | Lookup sees a complete registration rather than an endpoint awaiting its validator. |
| Blank addresses, missing dependencies, and typed-nil drivers/validators rejected | Prevents malformed entries from entering the concrete registry. |
| Duplicate addresses rejected; no replacement or removal | Preserves the registration between input validation and subsequent dispatch. |
| Lookup performs no authority callback or kernel trap | Keeps endpoint resolution separate from authorization; the new loop checks Invoke and Return in continuation processing. |
| Driver callbacks run after registry lookup releases its lock | Registry synchronization does not surround endpoint execution. |
| Privileges expressed as narrow operations | Provides a useful starting point for granting kernel operations without sharing scheduler state. |

The existing integration and API tests exercise registration through real echo
execution. `TestMissingDependencies` also covers a missing validator and typed-nil
driver. These are useful foundations, although they do not establish kernel-only
registration.

## Weaknesses and concrete consequences

### 1. Kernel-only registration is not established

`CTU.RegisterDriver` is available to every holder of the public CTU and directly
calls the registry. There is no kernel policy decision on that path. This is
consistent with the current host-configured probe, but does not implement the
kernel-only ownership described by the `Registrar` comments.

`PrivilegedCTU` currently adds vocabulary without wiring: no boot recipient,
initialization ordering, or failure behavior is implemented. Its `Admit` and
`Fault` fields also have no supplied implementations here.

### 2. Interface narrowing alone does not separate capabilities

The controller receives the same concrete `*DriverRegistry`, which implements
both interfaces. A holder of that `Resolver` can type-assert it to `Registrar`.
The normal controller code does not do this, but the claimed read-only boundary
is an API convention rather than a property of the object handed over.

This is an in-process ownership concern. Go interfaces do not sandbox arbitrary
native code or revoke references already held elsewhere.

### 3. The resolver constructor misses typed nils

`controller.NewContinuationController` checks `registry == nil`, while its
parameter is now a `driver.Resolver` interface. A nil `*DriverRegistry` inside
that interface passes the check and can panic on a later `Resolve` call.

The public `dna.NewContinuationController` still accepts `*DriverRegistry`, so
passing nil through that facade reaches the same problem. The facade also
prevents callers from supplying a different resolver implementation even though
the component constructor supports one.

### 4. Address and object-lifetime contracts are implicit

Registration rejects whitespace-only addresses but accepts `" echo "` as a
different key from `"echo"`. Lookup uses exact strings. This need not be an
implementation bug, but callers have no documented canonical-address rule.

The registry lock protects the map, not the internals of the registered driver
or validator. `Resolve` returns interface references to those same objects.
Mutating their configuration after publication can change behavior or introduce
races despite the registry mutex. Concurrent invocations can also call the same
endpoint object.

### 5. Dedicated registration and boot proofs are missing

There are no dedicated driver registry tests in the inspected directory. Existing
tests do not cover competing registrations, capability separation, typed-nil
resolvers, or delivery/failure of kernel boot privileges.

## Proposed ownership

Keep one registry per runtime. Boot owns construction; the trusted kernel receives
the registration operation; the controller receives endpoint resolution. Ordinary
drivers receive only their existing ingress/egress contract and bound output
callback.

| Participant | Proposed access |
| --- | --- |
| Trusted boot code | Construct runtime, install designated kernel endpoint, deliver privileges, publish successful runtime. |
| Kernel | Register approved local driver/validator pairs through a registrar capability. |
| Controller | Resolve endpoints through an object exposing only lookup. |
| Public CTU client | Submit root work and inspect snapshots. |
| Ordinary driver | Receive invocation snapshots and emit candidate work or results. |

Use separate unexported adapter types for the resolver and registrar, each
delegating to the same registry. Do not anonymously embed the registry in either
adapter, since that would promote its methods. This makes a registrar type
assertion on the runtime's resolver fail without creating another registry.

## Alignment with the new loop

Keep registration distinct from continuation authorization. Installing an endpoint
does not grant permission to invoke it or accept its returned observation.
`transition()` owns event application, assessment, selection, and reservation under
the scheduler mutex. Its continuation-opcode helper checks and stamps the trusted
Horizon for both **Invoke and Return**, before their opcode effects. Return resolves
its trusted Horizon from the bound resident source.

Dispatch resolves routing and places selected work. A dispatch callback carrying
`runtimeEvent` (`dispatchEvent`) performs no Horizon authorization check and carries
no authorization verdict. Source/handoff binding is still required. If placement
success releases buffered Invoke/Return output, that output enters continuation
processing and receives its own authorization check. Prior Invoke authorization
or placement success does not exempt Return.

Parent resumption retains the loop's one-resumption rule and existing InvocationID,
with a new handoff sequence. This registration proposal adds no separate dispatch
or resumption authorization hook; continuation authorization follows proposal 3.

The scheduler's resident `Continuation.Handoff` owns placement status, sequence,
pending output, and consumption. `Move.Invoke.Register` contains the semantic
continuation frame. Drivers receive detached Moves without Handoff bookkeeping.
The registry and kernel do not become owners of execution state or another event
loop. Endpoint validators remain outside the scheduler mutex; complete batch
validation must precede atomic installation and selection. Kernel initialization,
driver calls, dispatch, and host callbacks also remain outside scheduler locks.

## Proposed boot sequence

Use a registration-only initialization contract, conceptually
`Initialize(registrar) error`; final Go names and constructor signatures remain
implementation detail. Supply the designated kernel and ordinary root registration
separately. The kernel endpoint is explicitly installed by trusted boot, while the
kernel initializer installs the supplied root and child endpoint registrations.
The ordinary root receives no kernel privileges.

```text
trusted boot
  validate designated kernel, root, local registrations, and loop dependencies
  create private registry and separate resolver/registrar adapters
  install designated kernel endpoint with its input validator
  wire controller and scheduler using proposal 3 boundaries
  initialize kernel with only the registration capability
    fixture kernel installs fixed local registrations, including root and echo
  verify required root registration is available
  if initialization succeeds, publish public CTU exposing Submit and Snapshot
  if initialization fails, return error without publishing CTU or starting work
```

Kernel initialization completes before public admission and does not start
workflows. It runs without holding registry or scheduler locks; individual
registration operations acquire the registry lock internally. Deliver only the
registrar, not the existing `PrivilegedCTU` bundle with unusable `Admit` or `Fault`
functions. Admission and fault capabilities remain outside this increment.

Failed boot leaves its private registry unpublished. That is not rollback of
external effects or revocation of references retained by trusted boot code. Keep
initialization limited to local installation for this increment.

The registrar remains private to the kernel after boot. The fixture installs all
endpoints during initialization; no dynamic registration request protocol is needed
for the working loop. Registration remains append-only. Replacement, removal, and
revocation require separate lifecycle semantics.

## Minimal policies and working-loop acceptance

Stub policy while implementing the mechanics. The fixture kernel installs a fixed
list of trusted, locally supplied driver/validator pairs. This is an explicit test
policy, not production approval or dynamic code loading. Use an explicit test
Horizon policy and trusted fixture grants to exercise authorization for both Invoke
and Return. Do not replace the continuation check-and-stamp path with an implicit
allow-all shortcut or move the arbitrary v14 authority callback under the lock.

Capability separation, initialization failure, endpoint validation, placement
association, buffering, and event ordering must be implemented rather than stubbed.
Compiled matching can retain proposal 3's completeness stub for one resumption.
The acceptance path is:

```text
boot and register endpoints
  public Submit(root Invoke)
  validate and authorize root; reserve and dispatch initial handoff
  buffer root's child batch until matching placement success
  validate and authorize child Invokes; admit whole batch and dispatch
  gate each child Return on its placement; authorize and apply each Return
  once both observations are present, reserve the existing root's resumption
  dispatch resumed root with updated semantic Register and new handoff sequence
  gate and authorize root Return; deliver one root-result notification attempt
```

Use valid grants and successful placement for the baseline echo proof. Add explicit
rejection and placement-failure cases. The loop proposal's open expired-stamp
response, late-fault precedence, and rejected-output consumption decisions remain
owned by that proposal. They do not block composing registration and the successful
loop, but their behavioral paths need explicit outcomes before being considered
complete; this registration increment does not silently decide them.

## Proposed registration contract

- Require a nonempty address with no leading or trailing whitespace. Reject
  noncanonical input rather than silently rewriting a routing identity. Preserve
  exact, case-sensitive lookup; do not add URI or network routing semantics.
- Continue installing the driver and validator atomically and rejecting missing
  or typed-nil values. Continue rejecting every duplicate, including an attempt to
  re-register the same object. Never overwrite an existing endpoint.
- Treat successful publication as a commitment that both objects are ready for
  use and safe for the concurrency they can receive. Configuration becomes
  immutable after registration, or the implementation synchronizes changes itself.
- Keep `InputValidator.Validate` authoritative for endpoint input checking.
  `Declaration` remains descriptive; registration need not pretend to compile
  general JSON Schema.
- Use `support.Missing` for resolver dependency checks. Document that any supplied
  resolver must return valid registrations and stable endpoint bindings, preserving
  the concrete registry's assumptions.

No new registration wire opcode is proposed. The existing API installs Go
objects; a future driver request to load an endpoint would need a kernel-owned
mapping from a validated request to a trusted local factory. Raw output must not
be treated as a serialized Go driver or a source of registrar privileges.

## Public API migration

Removing `RegisterDriver` from the public CTU changes an API exercised by
`api_test.go` and the integration fixtures. Record this as an intentional migration:
provide registrations through trusted boot/kernel setup, update those fixtures,
and keep public submission and snapshot behavior.

Removing only the method from an interface is insufficient if its returned
concrete object still exposes it: callers could recover it by type assertion.
Remove the public concrete method too, or return a separate public facade whose
method set has no registration operation.

Align the `dna.NewContinuationController` forwarding signature with
`driver.Resolver`. Existing ordinary calls passing `*DriverRegistry` still work;
callers relying on the exact constructor function type may need changes.

If compatibility requires retaining the legacy CTU, explicitly label that path
host-managed and expose the kernel-owned runtime through a distinct constructor
and public view. Do not describe the legacy registration path as kernel-only.
The preferred direction is one migrated API rather than two permanent models.

## Verification for a later implementation

- [ ] Nil and typed-nil resolvers fail construction through both component and
  public constructors, without deferred panics.
- [ ] Empty/noncanonical addresses and missing or typed-nil registration
  components fail without installing an entry.
- [ ] Concurrent duplicate registration produces one success and one stable
  driver/validator pair; concurrent lookup and registration pass the race detector.
- [ ] The supplied resolver cannot be asserted to `Registrar`; the public CTU
  cannot be asserted to an interface exposing `RegisterDriver`.
- [ ] Only the designated kernel receives registration capability. Ordinary root
  and child drivers receive no privileges through runtime wiring.
- [ ] Kernel initialization failure returns no public runtime and starts no work.
  Successful initialization makes installed endpoints available to submission.
- [ ] Registration does not bypass continuation authorization: both Invoke and
  Return pass the Horizon check before opcode effects. Denied Invokes are not
  dispatched; denied Returns do not install observations or report success.
- [ ] RuntimeEvent dispatch callbacks perform no authorization check or verdict.
  Buffered continuation output released by placement success is still authorized.
- [ ] Boot followed by root Invoke, two child Invokes, child Returns, one root
  resumption, and root Return produces the ordered echo result and one notification.
- [ ] Output-before-placement is buffered without blocking acknowledgement;
  placement failure discards pending output and produces one failure notification
  attempt. Stale initial feedback cannot satisfy the resumed handoff's gate.
- [ ] Driver-facing registers exclude resident Handoff bookkeeping. Whole child
  batches are validated before atomic admission and selection.
- [ ] Preserve immediate, delayed/reversed, concurrent, source-binding,
  fault-suppression, snapshot, and callback-error integration behavior.

Build, vet, and race-test the affected v14 packages once implementation is
authorized. These are proposed checks, not results from this document-only review.

## Remaining decisions and scope

The minimal increment uses separate kernel/root inputs, registration-only kernel
initialization, a fixed local registration policy, a kernel-private registrar, and
a migrated public CTU without registration. Finalize concrete Go signatures and
fixture wiring during implementation; do not add placeholder Admit/Fault powers.
Production kernel approval policy and dynamic registration requests remain future
work. The registry's synchronization does not establish concurrency safety inside
individual endpoint implementations.

Dynamic code loading, network discovery, process isolation, registration
replacement/removal, retries, and persistence remain outside this registration
proposal. Scheduler-loop restructuring, placement gating, authorization, and the
compiled-matching stub follow proposal 3; they are integration dependencies, not
alternative rules defined here.

## Current code references

- [Driver contracts and registry](../../../skyra-dna-v14/driver/driver.go)
- [Declared kernel privileges](../../../skyra-dna-v14/driver/privilege.go)
- [CTU construction and public registration](../../../skyra-dna-v14/ctu/ctu.go)
- [Controller resolver dependency and authorization](../../../skyra-dna-v14/controller/controller.go)
- [Public API facade](../../../skyra-dna-v14/api.go)
- [Public API compatibility test](../../../skyra-dna-v14/api_test.go)
- [Integration and dependency tests](../../../skyra-dna-v14/ctu/ctu_test.go)
- [Scheduler-loop proposal 3](../loop/loop-proposal.md)
- [Placement-gated continuation algorithm](../loop/placement-gated-continuation-algorithm.md)
