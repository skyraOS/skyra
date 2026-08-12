# 01 · Removals

**Status: WORK ORDER.** Three types that are unused, or built past what anything needs.

> **There is no `C5`, and if you find one written down elsewhere, ignore it.** It named
> `entity.Entity` — a one-field `struct{ ID EntityID }` that was proposed and **never implemented**.
> The only `type Entity struct` in the tree is `entity/runtime.Entity`, which is a real
> implementation and **must not be touched**. Three types are removed here, not five.

> One commit per item. `go test ./...` green before each. If a test must change, STOP and record it
> in `stops.md`.

---

## C2 · `graph.UsedSeen` → a function

**File:** `graph/history.go`

```go
// REMOVE
type UsedSeen struct{ Used, Seen int }
func (u UsedSeen) Ratio() float64
```

Two ints and a ratio. Replace with a function returning the two values, or the ratio directly —
whichever reads better at the call site.

```text
FIRST   grep for UsedSeen across the tree, tests included. If nothing constructs it, delete it
        outright and note that in the commit message.
```

**Why:** `used / seen` is a compression ratio and a *gauge, not a verdict*. It is derived at read
time from an engram's history. A type for it invites someone to store it.

---

## C3 · `graph.Diff` → a method on `Partition`

**File:** `graph/history.go`

```go
// REMOVE the type
type Diff struct {
    From, To ids.ObservationID
    Fields   map[string][2]string
}

// KEEP the behaviour, on Partition, returning whatever shape the caller needs
func (p *Partition) Diffs() []Diff       // ← currently returns the type being removed
```

**Keep `Partition`.** It earns its place: the two folds are the knowledge of what a pairing returns,
and a real run showed it separating one right pairing from two wrong ones that had an identical
shape.

**Why `Diff` does not:** *within a partition the diff is the informative object* is a real claim, and
it is one method rather than a type. If a caller needs the pairs, return them inline.

```text
CHECK   Partition.Diffs() may be unreferenced outside graph/. If so, this is a pure internal
        change and the diff is small.
```

---

## C4 · `host.Handle` → `{Host, Ref}`

**File:** `host/host.go`

```go
// NOW
type Handle struct {
    Host   ids.HostID
    Ref    string
    Stdout string      // ← remove
    Stderr string      // ← remove
    Exit   int         // ← remove
}
```

**Why:** `Stdout`, `Stderr` and `Exit` are a **shell's shape leaking into a class interface**.
`Spawn` has never run — `localExec.Spawn` returns *"execution class not implemented"* — so these
three fields are a guess nobody was forced to make, on the one surface that cannot be corrected by
selection later.

Cut back to what a handle is: **an opaque reference, and the host it is valid on.** A process's
output is not part of what a handle *is*; it is part of what some future execution driver returns
beside one.

```text
BLAST RADIUS   host/host.go only. `Handle` appears in ExecutionOps.Spawn's signature and in
               localExec.Spawn's one-line stub. Nothing constructs a populated Handle anywhere.

NOT THE SAME   content.WorldRef also has a field called Handle (grammar/parse.go:179,
THING          content/question.go:57). LEAVE IT ALONE. Different type, different package.
```
