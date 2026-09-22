# identity/01 · Identity

**Package:** `ids`

Five types carry over unchanged. One is added, and it is added because `From` needs it.

---

## 1 · The types

```go
type (
    EngramID      string    // SPEC
    ObservationID string    // SPEC
    EntityID      string    // SPEC
    HostID        string    // SPEC
    SpanID        string    // SPEC

    RuntimeID     string    // ADDED — see §3
)
```

Distinct named types rather than a shared `ID`, so the compiler refuses a host where an engram
belongs. Carried from `../../01-ids.md` without change.

## 2 · Global uniqueness stays load-bearing

```text
WHY GLOBAL UNIQUENESS   Producer resolves ACROSS entity trees. the credit walk is bounded by
IS LOAD-BEARING         CITES BEING EMPTY at a crossing, never by ID scope — so nobody should
                        reach for scoped IDs to "fix" a boundary that is already held.
```

Unchanged, and worth repeating in a document proposing a distributed-facing `RuntimeID`, because
that is exactly the point at which someone reaches for scoped IDs.

## 3 · `RuntimeID`, and why it is new

`From` is *namespace × requesting engram × runtime*, and the third has no type in the current
surface. `../../../semantics/terminology.md` defines what it names:

```text
SYSTEM    the whole network of runtimes. specifically the network layer.
RUNTIME   one per device that can run one. AN AUTHORITY DOMAIN.
KERNEL    one per runtime. the authority over it. holds everything HAPPENING there.
HOST      physical hardware addressed over a network — INCLUDING devices with no OS.
```

**A runtime is not a host**, which is why `HostID` cannot stand in. A phone with no runtime on it is
a host and not part of a distributed runtime. And a kernel-backed runtime instance does not span
devices — there is one per device — while an execution may span devices through `Reach`.

```text
AUTHORITY REMAINS LOCAL.   WORK CROSSES A BOUNDARY.   these are different claims.
```

`RuntimeID` is the name of the authority domain, and it appears in `From` because *local and remote
are relations to runtime-supplied `From`, not entity properties* (P13). Without it, P13 has nothing
to be a relation to.

```text
FOR REVIEW    whether RuntimeID is minted like the others or derives from device identity.
              ../../syscall-read-example/outcomes/ decided device identity is "a hash over
              hardware fields read at boot," which argues for derivation rather than minting —
              a minted runtime ID would not survive a restart, and From must mean the same thing
              across one. NOT SETTLED. flagged with H10.
```

---

## 4 · The minter

```go
type Minter struct {        // ADDED, carried
    engram      atomic.Uint64
    observation atomic.Uint64
    span        atomic.Uint64
}

func (m *Minter) Engram()      EngramID
func (m *Minter) Observation() ObservationID
func (m *Minter) Span()        SpanID
```

`2026-08-09/interfaces.md` leaves this open — *"EngramID, ObservationID and SpanID are all globally
unique, all minted by the kernel, and probably one mechanism."* This is that one mechanism, carried
from `../../01-ids.md` unchanged.

**Sequential rather than random on purpose.** An ID that sorts by creation makes a transcript
readable, and nothing in the design reads an ID for meaning.

`RuntimeID` is deliberately **not** on the minter, pending §3's open question. `EntityID` and
`HostID` are not minted either — they are identities registered per runtime, not things the kernel
creates.

```text
CARRIED REVIEW NOTE    Minter is defensible but it is state the kernel holds, and the kernel
                       already holds it as a field. Whether it is a type or three counters on the
                       Kernel struct is a coin flip; it is a type because the spec calls it
                       "one mechanism."
```

---

## 5 · What is not an ID

```text
NO ContractID     a contract is identified by Position — (owning parent, position name).
                  ../contracts/02-child-contract.md §5.

NO PositionID     same reason. Position is a pair, not a minted name.

NO VersionID      a version is (Position, Age). ../../../reach/age.md: "The parent's map key is
                  already the lineage identity, so this needs no new durable type."

NO CandidateID    a candidate IS an observation and already has an ObservationID.
```

Four types that a naive pass would add, each removed by the same argument: the thing already has a
name, and a second name for one fact is the defect `Cites`-is-derived exists to avoid.
