# v12 implementation gaps

**Date:** 2026-09-01  
**Scope:** gaps found while comparing `overnight.md`, the frozen v12 type surface, and the authoritative
08-27/09-01 specifications. This is an inspection record, not an implementation decision log.

## Bottom line

The execution model is coherent at the architectural level, but it cannot be implemented literally while
the current surfaces are both frozen and exhaustive. The named seams A-E are real, and several additional
boundary gaps prevent the requested local reactor, fault re-drive, remote return path, and golden test from
being expressed without inventing private structures or changing a contract.

If "frozen signatures" permits private runtime structures, private methods, indexes, adapters, and an
executable wiring package, the probe can proceed while recording those inventions in `DECISIONS.md`. If it
means the existing types and methods are exact and exhaustive, Phase 1 and Phase 2 cannot reach the stated
definition of done.

## Load-bearing contract gaps

### 1. Materialize has no result channel

`kernel.Faults.Materialize` returns `(*c.Context, c.Outcome)`, but the uniform CTU trap interface is:

```go
type Handler interface {
	Handle(f Fault) c.Outcome
}
```

At the same time, the boundary says that the kernel never writes CTU memory: it returns a provisioned
Context, and the CTU installs it and re-drives the faulting Move. `Handler.Handle` cannot return that Context,
and the current `Fault` frame contains no result slot through which an adapter can return it.

One of the following has to be invented:

- a result field in the trap frame;
- a richer handler return value;
- a CTU-owned adapter with a private result channel;
- kernel mutation of CTU state, contradicting the documented ownership boundary.

This blocks the demanded Materialize -> install -> re-drive path.

### 2. The required in-flight table has no home

The settled state model requires three tables:

```text
READY   WorkList
ISSUED  in-flight table
LANDED  Context.Registers
```

`ctu.CTU` contains `Cache`, `Bodies`, and `Work`, but no local in-flight table. `Core` contains only an ID.
The remote socket table is described behind `kernel.Barrier`, but that does not provide the local issued
table required for asynchronous `Driver.Perform` calls.

Without a concrete in-flight table the implementation cannot:

- make removal from Work an atomic move into ISSUED;
- prevent duplicate local dispatch;
- pop the issued entry on Landing;
- distinguish quiescence from "Work is temporarily empty while operations are running";
- implement `Work empty AND in-flight empty` as the retirement condition.

A private CTU runtime sidecar or a new CTU field is required.

### 3. Landing has no ingress API

Local and remote operations must both produce `c.Landing`, handled by the same LAND half of the reactor.
The only CTU door is:

```go
func (ct *CTU) Transfer(mv c.Move) c.Outcome
```

It cannot accept a `Landing`, and neither CTU nor Barrier exposes another Landing method. The comment on
`Transfer` explicitly leaves this open.

This prevents expression of:

- local asynchronous completion;
- remote Landing ingress;
- write -> pop in-flight -> flip;
- `T-remote` without an extra private entry point.

### 4. Asynchronous local completion has no delivery mechanism

`Driver.Perform` synchronously returns `(Observation, Outcome)`, while ISSUE is required to be
fire-and-forget and all completions must arrive as Landings. No completion channel, callback, event queue,
or Landing method exists.

A goroutine alone is insufficient: it still needs a defined, synchronized route back into LAND and a place
to retain in-flight state until completion.

### 5. Move-to-Body resolution has no index

A wire `Move` names `(TargetFd, Action)` and intentionally carries no `BodyID`. The target must resolve its
own Body from `(Target, Action)`. However:

- `BodyCache` is keyed only by `BodyID`;
- `engram.Store.Activate` accepts only a `BodyID`;
- no `(Target, Action) -> BodyID` registry or resolver exists.

Therefore neither a warm i-cache lookup nor an engram page-in can begin from the information present in a
Move. A content-addressed body store and a callable-name index are two different structures; the second one
currently has no home.

### 6. THINK cannot return a complete Body DAG

`Body.Children` contains child `BodyID`s, but:

```go
Compile(request c.Observation, registered []c.Target) (c.Body, c.Outcome)
```

returns only one Body. A pure THINK implementation cannot return the child bodies referenced by that root,
and THINK is explicitly forbidden from executing, provisioning, or mutating CTU/kernel stores as an ambient
side effect.

The one-node seed avoids this. The required three-node golden Body does not. It needs one of:

- a compile artifact containing the root plus all DAG objects;
- a caller-supplied content-addressed sink;
- a separate child-resolution mechanism;
- hand-loading child bodies outside the Compile contract.

The last option can make the test run, but does not solve the production boundary.

### 7. Remember persists one Body, not the complete DAG

`engram.Store.Persist(body)` writes one Body. The authoritative model says Remember CAS-writes each quiesced
subtree, but neither `Context` nor the Remember interface supplies the complete reachable set directly.

Persisting only the root leaves its child digests unresolved on a future warm activation unless all children
were separately persisted by an undocumented traversal or side effect.

### 8. Root move identity and output register are undefined

`WorkList` is keyed by `MoveID`, whose `Writes` field is the output register in the parent. A root Move has
`ReturnFd == nil` because it has no parent. The docs say the root output/register is derivable as the DAG
sink, but do not define:

- the root's `MoveID` for insertion into Work;
- the root output `RegKey`;
- how its Landing is represented when there is no parent fd;
- the exact transition from root output to `Context.Result`.

A reserved root register, a separate root issue path, or another explicit convention must be invented.

### 9. Request data disappears before execution

`Cold.Handle` receives the request Observation and THINK consumes it to select a Body. Neither `Body`,
`Move`, nor `Context` retains that request as an operand. The seed greeter can return a constant while ignoring
its input, but a driver that needs request data cannot receive it.

This is the concrete cold-path manifestation of seam A.

### 10. `PossibilityID = hash(body + inputs)` is not computable yet

The child PossibilityID is meant to be minted at the target from the resolved Body and its inputs. Until the
operand model says what the inputs are and how they are canonically encoded, the hash input set is undefined.

This also leaves open:

- ordering of multiple observations;
- inclusion of literal/bound operands;
- distinction between semantic content and provenance segments;
- handling of NOW/RANDOM/side-effecting moves that cannot safely memoize;
- whether contracts or remembered type information participate in identity.

## Named design seams

### Seam A: operand model

A move's child returns account for upward-flowing computed values, but there is no home for:

- leaf inputs;
- constants and literals;
- bound action arguments;
- the original request;
- a parent's datum threaded down;
- non-child observations used by a move.

Because `Driver.Perform(action, in []Observation)` receives only a slice of observations, the reactor also
needs a deterministic rule for ordering a map of child registers into that slice.

This seam gates readiness, driver invocation, PossibilityID minting, and any useful non-constant leaf.

### Seam B: flip index

LAND must find every move that consumes a newly landed register. No reverse index exists, and Work contains
only moves that are already ready. Non-ready parents therefore need a discoverable runtime home.

The implementation must choose between:

- a `consumers[PossibilityID][RegKey]` reverse index built while expanding the Body DAG;
- scanning the resident Body graph on every Landing;
- a private pending-move table containing unresolved dependency counts.

Whichever choice is made is the coupling between ISSUE and LAND and must be synchronized with both.

### Seam C: consume and free

The spec says a parent consumes all its inputs when it fires and frees those register slots. The exact atomic
trigger is unspecified.

Questions forced by the golden test include:

- are inputs copied and deleted in the same critical section as the Work claim?
- does a failed authorization or dispatch consume them?
- if dispatch faults and re-drives, must inputs remain resident?
- how does fan-out work if two parents consume the same content-addressed child value?
- is freeing per parent edge or per physical register slot?

Deleting a shared slot on the first consumer conflicts with fan-out unless each edge has its own runtime
register or a consumer count.

### Seam D: Fault vs NotYet vs Failure

The three paths are semantically distinct, but the stubs do not yet provide distinct state transitions:

- `NotYet`: an operand is absent; no kernel call and no ready Work entry;
- `Fault`: body/authority backing is missing; trap, repair residency, then re-drive;
- `Failure`: a semantic errno Observation lands as a value.

The current `Transfer` and `Transition` return `NotYet` unconditionally, so no executable distinction exists.
The implementation also needs to decide where a faulting move resides while the handler runs so it is neither
lost nor double-issued.

### Seam E: child PossibilityID hash input

The target must mint a child identity from Body plus inputs, but the operand representation and canonical
digest format are unresolved. The identity rule must also explicitly disable or salt reuse for impure actions.

## Authority and fault classification gaps

### Authorization miss and materialization miss share an insufficient outcome vocabulary

ISSUE must perform Interpret plus Verify per move. A missing Body is a Materialize page fault. A missing or
invalid capability is a protection/authorization fault, but the vector table contains only Materialize and
Remember. Reusing Materialize for an authority failure would conflate memory residency with permission and
could accidentally turn a denial into a provisioning request.

The spec mentions a possible future Provision vector but does not settle it. Phase 1 still needs an explicit
failure/fault rule for Interpret or Verify misses.

### Signature semantics are absent

`Capability.Sig` is bytes, but there is no signing/verification/key interface, canonical signed payload, or
clock seam for expiry. A probe implementation can invent deterministic mock signatures, but this is security-
critical behavior and must be recorded rather than presented as the settled offline verification scheme.

### Bound cannot necessarily enumerate the entire workflow

`Provisioning.Bound` receives one root Body and must authorize every `(Target, Action)` in the complete DAG.
Because children are referenced only by digest and Bound receives no Body resolver, a general implementation
cannot traverse them through the interface alone. A concrete provisioner would need private access to the
BodyCache/engram or a precomputed target set.

## Contract and persistence gaps

### Remember cannot write the promised discovered contract

The authoritative contract spec says structural return information is discovered by running and authored at
Remember. The frozen `Contract` contains only a name, `Context` has no discovered contract field, and
`engram.Engram` stores only a Body. There is no structural type representation, observation-type registry,
or remembered-contract field.

Therefore a probe may persist the original named Contract, but it cannot implement the described type-
feedback behavior or a meaningful warm-run return guard.

### Digest canonicalization is unspecified

`Body.Children` is a Go map. A content digest must define deterministic ordering and encoding; hashing a
default map serialization is not a portable content-addressing contract. The digest must also say whether
`Body.ID` is excluded from its own preimage to avoid a recursive identity definition.

### Cache's Materializing state is only implicit

`Cache` is `map[PossibilityID]*Context`, so a present nil pointer can represent Materializing, but no wait-list,
wake mechanism, or failure cleanup is defined. If materialization fails, the implementation must decide
whether the nil latch is removed, retained as a negative cache entry, or woken with a failure.

## Remote and concurrency gaps

### Barrier ingress cannot carry the response type

`kernel.Barrier.Ingress(from, mv c.Move)` accepts only the forward activation. The barrier prose requires it
to distinguish a forward Move from a return Landing, but Go has no sum type at this boundary and no second
Landing method exists.

### Reply-to-sender state is not represented by the interfaces

`ReturnFd` omits a destination because the barrier is supposed to derive it from the sender and retain it in
the socket table. The Barrier interface exposes no socket entry type, source association, or response method.
That can be private implementation state, but it is mandatory rather than optional plumbing.

### Frozen maps are unsafe for N > 1

`Cache`, `BodyCache`, `WorkList`, and `Context.Registers` are ordinary maps. Concurrent core issue, async
driver completion, and ingress LAND require a shared lock or a complete ownership/event-loop discipline.
Neither is represented. CAS cannot be applied directly to Go map entries.

Phase 3 therefore requires either a CTU-wide synchronized runtime sidecar or confinement of all table
mutations to one reactor goroutine.

### Quiescence races with asynchronous Landing

Even with an in-flight count, retirement must be linearized with LAND and any flip that enqueues new Work.
Checking two maps independently can observe both empty between operations and retire a context prematurely.

## Boot and executable-delivery gaps

### `boot` is not an executable

The current `boot/boot.go` declares `package boot` and a `Boot` interface. There is no `package main` or
`func main()`. The requested command that prints `hello, human` needs either a separate command directory or
a change in package structure.

### The concrete ownership graph is not surfaced

The interfaces describe Boot, Cold, Provisioning, Faults, Barrier, Handshake, Thinker, Disk, and Table, but
there is no concrete kernel type that owns and connects them. That is expected implementation work, yet the
frozen surface does not specify how Cold reaches the CTU to execute or how the handler adapter returns its
materialized Context.

## Documentation conflicts and stale vocabulary

The work order declares `Message`, `Route`, `RegState`, and `Work[Issued]` dead terms and says that typing one
means following a stale document. The supposedly authoritative/current v12 files still contain them:

- `causal.go` repeatedly describes a `Message` even though Move replaced it;
- the v12 README says `Route` and `Work[Issued]`;
- `kernel.go` says "now the wire unit is a Message" inside the Materialize commentary;
- `causal.go`'s package authority still points first to the older v11 execution model;
- `move-and-possibility.md` says its code application included `RegState`, while the later state model deletes
  it.

The implementation can follow the explicit precedence banner and current concrete types, but the comments
should not simultaneously be treated as an error detector for stale terminology.

## Definition-of-done constraints

### The targeted v12 packages currently pass only as stubs

The following targeted commands are green:

```text
go build ./src/skyra/skyra-dna-v12/...
go vet ./src/skyra/skyra-dna-v12/...
go test ./src/skyra/skyra-dna-v12/...
```

There are currently no v12 tests, and the execution bodies return `NotYet` or do nothing, so this establishes
only that the frozen surface type-checks.

### Module-wide `go build ./...` is already red outside v12

At inspection time, module-wide package discovery/build fails because unrelated archived commands and
packages import paths that are absent from the module. Therefore the work order's literal module-wide green
requirement cannot be attributed solely to v12 work unless those existing archive failures are excluded or
separately repaired.

### The provided worktree is already dirty

At inspection time, v11/spec files have existing modifications and the entire v12 directory is untracked.
The hard requirement to work in an isolated worktree needs to be satisfied before implementation begins, or
the supplied workspace must be explicitly confirmed as the intended isolated tree. Existing unrelated work
must be preserved.

## Minimum decisions required before or during the probe

The smallest coherent probe implementation must explicitly record choices for:

1. the Materialize result/re-drive channel;
2. the CTU runtime sidecar: in-flight state, locks/event ownership, completion delivery, and waiters;
3. a Landing entry point shared by local and remote completion;
4. the `(Target, Action) -> BodyID` resolver;
5. representation and persistence of a complete compiled Body DAG;
6. root MoveID/output-register convention;
7. operands, their deterministic order, and seed-request placement;
8. consumer indexing, fan-out accounting, and consume/free atomicity;
9. canonical Body and Possibility digests, including impurity handling;
10. authorization-fault classification distinct from Materialize;
11. the reduced probe meaning of Remember, since structural contract feedback has no type surface;
12. the executable boot location and the scoped build/test command.

Each invention should be tagged at its implementation site and repeated in `DECISIONS.md`, as required by
`overnight.md`.
