# Engrams and resolution

**Reviewed:** 2026-09-07  
**Status:** Durable-unit and lookup decisions; storage and declaration mechanics remain unimplemented.

## Current versus committed

The [working-context/compile/Remember decision](host-entity-address-horizon/execution-history-and-live-work.md)
separates working history and live work from the compiled artifact. Latest successful compile selects
current for new invocations; Remember is the user-initiated commit boundary. Compilation alone does
not commit. How these forms map to the durable unit below, including what Remember persists, remains
open. An immutable compiled artifact must not be confused with the mutable working context.

## Remembered behavior and live execution

An engram is a remembered executable: one body and its immediate child/dependency references, with
the reusable specification needed to produce its return. Each child engram owns its own layer.
An engram does not physically contain the entire descendant graph.

The body/register-specification direction survives; final Engram fields do not yet exist. Do not
restore the earlier parentContract/childContracts struct proposal as an accepted shape. Live register
values and return addresses belong to the execution, not to another run restored from disk.

Structure and loading policy are separate. Whether descendants are loaded eagerly or on demand is
still open. Version-specific relationships make behavioral changes visible, but the digest/version
protocol must be reconciled with the latest identity decision: **MoveID and EngramID are the same ID**,
with action plus ordered observation params as the starting rule. Neither automatic memoization nor
content-hash version propagation is claimed as an implemented identity protocol.

## Lookup and provisioning

```go
// Filesystem driver method
Resolve(id MoveID) (*Fd, error)
```

Scheduler looks for the incoming MoveID in Memory first. On absence, kernel calls the filesystem
driver's Resolve and installs the returned Fd through Memory. The returned entity already contains
Move and Register; no new result wrapper or intervening EngramStore/Pager layer is required.

If nothing is remembered, declaration must supply executable behavior while kernel provisions its
register. Creating an empty register alone does not supply that behavior. I/O failure or invalid stored
content must remain distinguishable from no remembered executable; errors do not silently authorize
a declaration fallback or count as a successful load.

Resolve lives on the filesystem driver. A retained Kernel.Load method would coordinate the call and
installation; it would not restore the older kernel-owned read/decode/verify division. How the returned
fd binds to the arriving Move's live authority and parent address, and the absence/error conventions,
still need definition. Success alone does not dispatch or authorize the work.

## Durable storage

Use a configurable application-data directory through the same filesystem-driver mechanism used
for ordinary file work. No separate partition is needed. The intended macOS default is:

```text
~/Library/Application Support/Skyra/
```

Engrams are durable application data, not disposable cache. Resident copies may be cached. Files meant
for the user and exported artifacts go to locations the user chooses. Hidden status is presentation,
not an access-control boundary.

The initial storage direction is one immutable file per engram with entry-point references separate.
Exact filename, revision, canonical encoding, and publication protocols remain open. Exact-ID lookup
is distinct from discovering suitable behavior by intent. Updating an entry-point reference does not
mutate an existing remembered executable.

Sharing the driver does not share kernel storage authority with workflows. The driver and enough
loading infrastructure must exist at boot independently of the missing engram. Recovery must preserve
the incoming identity, usable inputs, return address, and authority; it must not dispatch a second
independent recovery or fabricate a completed return. No separate Attempt ID is adopted.

## Open boundaries

Durable encoding and integrity checks, safe publication, eviction and retention, declaration output,
shared-load behavior, fault transport, and engram availability across entity/host bindings remain
open. Do not import a two-fault-only API, a full Context checkpoint format, or a stored grant as renewed
authority. These choices are not prerequisites to implementing every part of the minimal local probe.

## Reviewed source

Selective port of [09-01 loading and fault returns](../2026-09-01/engram-loading-and-fault-return.md),
using its latest Resolve/identity corrections and the current checkpoint. The old distinct backing ID,
CTU dispatch-time Resolve, kernel decoder, and Issued/Landed fault walkthrough are excluded.
