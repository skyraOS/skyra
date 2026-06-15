# Remaining Changes — Motor as Dispatch

8 items left from the Motor Dependencies list. Items 1–10 are done (Thread struct, Intent struct, motor.go, intent.go, ExpressMotor one-liner, PrimitiveSpeak removed, engage primitive stubbed, main.go bootstrap, intent parsing in Realize removed, present intent rendering).

---

## 1. CompetenceExpress — collapse to motor-only

**File:** `src/reality/skyra.go:117`

Currently selects between `"compress"` and `"pass-through"`. Should only ever set `s.Type = "motor"`. Compress fires during traversal on density threshold — it's not an express phase anymore.

ThoughtDensity stays (it gates whether compress fires during traversal), but CompetenceExpress stops using it to pick a type.

```go
func (s *Skyra) CompetenceExpress(r *Relation) {
    s.Type = "motor"
}
```

---

## 2. Express method — only fire motor

**File:** `src/reality/skyra.go:470`

Currently fires `s.ExpressFns[s.Type]` and then also fires `"speak"` if type isn't speak. All of this collapses. Express fires motor. That's it.

```go
func (s *Skyra) Express(r *Relation) *Relation {
    r = r.Express(r)
    if fn, ok := s.ExpressFns["motor"]; ok {
        r = fn(r)
    }
    return r
}
```

---

## 3. main.go ExpressFns — remove compress registration

**File:** `main.go:84`

Currently registers both `compress` and `motor`. Remove compress from ExpressFns. Compress fires inside the traversal (called from propagate on density threshold), not as an express phase.

```go
skyra.ExpressFns["motor"] = reality.ExpressMotor(skyra)
// compress removed — fires during traversal, not express
```

Compress still exists as a function, it just isn't registered on ExpressFns anymore. It fires from within propagate when ThoughtDensity crosses the threshold.

---

## 4. Present primitives section — show real primitives

**File:** `src/reality/present.go:19`

Currently hardcoded to `"speak"`. Should reflect what the being actually has registered.

```go
// primitives — what the being can do
if s, ok := r.World[beingID]; ok {
    if skyra, ok := s.(*Skyra); ok {
        var names []string
        for name := range skyra.Primitives {
            names = append(names, name)
        }
        sort.Strings(names)
        sections = append(sections, "## Primitives\n"+strings.Join(names, ", "))
    }
}
```

Requires passing the being (or its primitive map) into Present, or looking it up from the world.

---

## 5. Present thread rendering — use sl.Thread (nested map)

**File:** `src/reality/present.go:42–70`

Currently renders from `r.Threads` (the Relation's binding threads). Should render from `sl.Thread` — the being's private nested map.

```go
// threads — the being's thread state
if len(sl.Thread) > 0 {
    for threadID, realities := range sl.Thread {
        var lines []string
        for realityID, entries := range realities {
            lines = append(lines, fmt.Sprintf("%s: %d entries", realityID, len(entries)))
        }
        sort.Strings(lines)
        sections = append(sections, "## Thread: "+threadID+"\n"+strings.Join(lines, "\n"))
    }
}
```

---

## 6. Thread write paths — primitives write into sl.Thread

**File:** `src/reality/primitives.go`

When a primitive fires and gets a result, it writes into the Slice's nested Thread map:

```go
sl.ThreadAppend(threadID, targetID, resultReality)
```

- **look** — each entity found gets written as its own slot: `sl.ThreadAppend(activeThread, entityID, entity)`
- **engage** — the LLM response gets wrapped as a Reality and written: `sl.ThreadAppend(activeThread, target, responseCell)`
- **listen** — opens the gate; incoming events write themselves via ThreadAppend

The active thread ID comes from the Relation's ThreadID or from the intent entry's context.

---

## 7. debugview — reflect new structures

**File:** `src/debugview/server.go`

SliceSnapshot already uses `sl.Intent.Render()` and `len(sl.Thread)`. The HTTP handler that serves the debug view needs to render:

- Intent graph as structured phases (not a flat string)
- Thread map showing thread ID → reality ID → entry count
- PreviousActions with status (success/failed)

---

## 8. tui sidebar — render new structures

**File:** `src/tui/model.go`

The sidebar currently renders the old flat structure. Needs to show:

- Intent graph phases with status indicators
- Thread map hierarchy
- Active primitive state

---

## Order of implementation

1–3 are pure cleanup (collapse express flow). Do them together.
4–5 fix Present for inference accuracy. Do them together.
6 is the real work — primitives writing results into thread. Depends on engage being fully implemented (not just stubbed).
7–8 are display. Do last.
