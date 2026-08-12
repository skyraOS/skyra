# 02 · Defects

**Status: WORK ORDER.** Two live bugs, and the type shape that permitted one of them.

> One commit per item. `go test ./...` green before each. **Add a test for `D1` and `D2`** — these
> are the two items in this order where a new test is expected rather than a stop.

---

## D1 · `commit` silently drops all but one delegated reach

**File:** `kernel/kernel.go:422`, in `commit`, the naming branch.

```go
// Reach passes through a naming commit — otherwise discover could resolve
// something and the caller could never name it.
for _, er := range named.Content.Reach() {
    out.Content["reach"] = er        // ← every iteration writes THE SAME KEY
}
```

`Content.Reach()` returns a **slice**. Every iteration overwrites the previous one, so a naming
commit that passes up two or more `EngramRef`s hands the caller exactly one — the last, by map
iteration order, which is not even stable.

**Fix:** carry all of them. A `content.List` under one key, or indexed keys — either is fine, so long
as the receiving side in `runGroup` still folds every returned `EngramRef` into `Delegated`.

```go
// runGroup already iterates correctly — it is the producer that is lossy.
for _, er := range ret.Content.Reach() { … p.Delegated[er.Name] = child … }
```

**Test:** an orchestrator names a child whose return carries two `EngramRef`s; assert both names
resolve in the parent's next frame.

**Why it matters:** silent loss of delegation is the failure class this design refuses everywhere
else. A capability the caller was granted simply is not there, with no error.

---

## D2 · `Materialize` leaves a failed write undisposed

**File:** `kernel/root.go:142`, in `Materialize`.

```go
if err := s.Write(e.Path, e.Body); err != nil {
    if firstErr == nil {
        firstErr = err
    }
    continue                 // ← neither Landed nor Discarded, and NO RecordDiscard
}
e.Landed = true
```

The rule is that **every staged effect is explicitly disposed** — landed or discarded — and an
unreachable one is a *reportable fact, never a no-op*. Unreachable is handled. **Failed is not.** The
effect ends in a fourth state nothing can name and nothing records.

**Fix:** dispose it. Record a discard carrying the write error as the detail, and mark it.

**Test:** stage an effect the credit walk reaches, use a storage driver whose `Write` returns an
error, accept, and assert a discard record exists for it.

---

## B1 · `Effect` has two booleans for four states

**File:** `kernel/root.go`

```go
// NOW
Landed    bool
Discarded bool

// TO
Disposition Disposition   // Staged · Landed · Discarded · Failed
```

**This is why `D2` exists.** Two booleans express four combinations, two of which are meaningless
(`both true`) and one of which — *neither* — is the exact state a failed write falls into
unnoticed. An enum makes forgetting a disposition **a compile error** instead of a silent fourth
state.

```text
Staged      held, not performed. the initial state.
Landed      the driver wrote it.
Discarded   the credit walk never reached it, or the chain was rejected. RECORDED.
Failed      reached, attempted, the driver errored. RECORDED.
```

**Order:** do `B1` first, then `D2` — with the enum in place, `D2` is a two-line change and the
compiler points at every site that has to say which state it means.

**Callers to update:** `Materialize` and `Discard` in `kernel/root.go`, and anything in the substrate
tests that asserts on `Landed` or `Discarded`.

```text
NOTE   if a substrate test reads e.Landed directly, updating that read is EXPECTED and is not a
       "test must change" stop — the field is being renamed, not the behaviour. A stop is when a
       test's ASSERTION about behaviour would have to change.
```
