# 01 · Identity

**Package:** `ids`   ·   26 code lines

Everything durable is named by one of five types, and one thing mints all of them.

## The identity types

```go
type (
    EngramID      string    // SPEC
    ObservationID string    // SPEC
    EntityID      string    // SPEC
    HostID        string    // SPEC
    SpanID        string    // SPEC
)
```

All five are `interfaces.md`'s, unchanged. Distinct named types rather than a shared `ID`, so the
compiler refuses a host where an engram belongs.

## The minter

```go
type Minter struct {        // ADDED
    engram      atomic.Uint64
    observation atomic.Uint64
    span        atomic.Uint64
}

func (m *Minter) Engram()      EngramID
func (m *Minter) Observation() ObservationID
func (m *Minter) Span()        SpanID
```

`interfaces.md` leaves this open — *"EngramID, ObservationID and SpanID are all globally unique, all
minted by the kernel, and probably one mechanism."* This is that one mechanism.

**Sequential rather than random on purpose.** An ID that sorts by creation makes a transcript
readable, and nothing in the design reads an ID for meaning.

```text
WHY GLOBAL UNIQUENESS   Producer resolves ACROSS entity trees. the credit walk is bounded by
IS LOAD-BEARING         CITES BEING EMPTY at a crossing, never by ID scope — so nobody should
                        reach for scoped IDs to "fix" a boundary that is already held.
```

## Review notes

`Minter` is defensible but it is state the kernel holds, and the kernel already holds it as a field.
Whether it is a type or three counters on the `Kernel` struct is a coin flip; it is a type because
the spec calls it "one mechanism."
