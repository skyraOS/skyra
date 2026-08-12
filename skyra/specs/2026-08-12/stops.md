# Stops

**Status: LOG.** Append-only. One entry per item that could not be completed as written.

An entry is not a failure — it is the work order being wrong, which is the thing worth knowing in
the morning.

```text
FORMAT

## <ITEM ID> — <one line>
WHAT WAS ASKED    the instruction, quoted.
WHAT WAS HIT      the specific thing that required a decision, with file:line.
WHAT WAS LEFT     the state the code was left in. should be UNCHANGED.
```

---

## E1 / E2 — `Raw` is text, and reach cannot be written as text

**WHAT WAS ASKED**

> `Step(d Delivery) (Raw, error)` … `type Raw string // the entity's response, unparsed.`
>
> `Step` now returns `Raw`. **The kernel calls `grammar.Parse`**, not the adapter.

**WHAT WAS HIT**

Two of the three adapters do not produce text, and one of them cannot.

```text
entity/runtime/runtime.go:244   byIndex    "reach": content.EngramRef{…}
entity/runtime/runtime.go:278   byRecency  "reach": content.EngramRef{…}
entity/runtime/runtime.go:312   create     "reach": content.EngramRef{…}
```

All three resolvers commit content containing a `content.EngramRef`. **The grammar has no surface
form for one.** `grammar/parse.go:172 toValue` turns a decoded string into exactly `Ref`,
`WorldRef` or `Str`, and nothing anywhere produces an `EngramRef` from text. So routing every
entity's emission through `Raw` + `grammar.Parse` means the entire discover subsystem loses its
ability to hand back reach — silently, since a commit with no `EngramRef` still parses fine.

Making it expressible is the part that is not mine to decide, because `content/content.go:51-71`
says the type IS the enforcement:

```text
DELEGATION   reach arriving from an entity I CHOSE to invoke.   legitimate.
INJECTION    reach arriving in the data channel.                whoever can talk to me picks
                                                                what I can touch.
"Same bytes, different provenance, and only the type keeps them apart."
```

A textual `<engram …>` that `toValue` accepts is reach travelling in the data channel by
construction. That is not scope — it is the invariant inverted.

The three exits I can see, none of them written down here:

```text
a   the grammar gains an EngramRef form.        breaks delegation/injection, above.
b   Raw stops being text.                       contradicts `type Raw string`, and the kernel is
                                                then not parsing one grammar.
c   code entities skip the parse path.          two classes of entity, and E1's whole point is
                                                that the kernel holds what produced a decision.
```

**Worth saying plainly:** this would have passed `go test ./...`. No test in the tree emits an
`EngramRef` from a scripted body, so the fence would have stayed green while the runtime entity
quietly stopped being able to delegate. That is the failure rule 3 is for.

**WHAT WAS LEFT** UNCHANGED. `entity/entity.go`, `kernel/kernel.go`'s call sites,
`entity/codex`, `entity/runtime` and `entity/scripted` are all as they were. Nothing of E1 or E2
was started.

---

## C1 — `View.Child` cannot reach a `Deref`, and the fold degrades silently

**WHAT WAS ASKED**

> **The decision, so this is not one you have to make:** `View.Summary(name)` becomes
> `View.Child(name) (*graph.Engram, bool)` returning the child engram **with `Cites` stripped**.
> … `entity/codex/assemble.go` builds whatever prose shape it wants from that, locally.

**WHAT WAS HIT**

The child engram is not enough to build the prose from, and the missing piece is the one thing
only the graph has.

```text
graph/history.go:91    func (e *Engram) Partitions(d content.Deref) []*Partition
frame/frame.go:186     return graph.Summarize(name, c, v.g.Deref()), true
                       ← the ONLY place a Deref reaches this code, and it is inside frame
```

`assemble.go:140` renders the partitions. With `View` handing back only the engram, the adapter
has no `Deref` to pass, and `Partitions(nil)` is not a degraded-but-fine option:
`content/question.go:66-68` falls back to `"?@" + id` for an unresolvable ref, so two askings
carrying the same material through different observation instances stop comparing equal. The
history fragments into one partition per instance — *"comparing Refs by ID would make every
question unique again, which is exactly what killed the first InvocationKey."*

Keeping the fold at its current fidelity means `View` must also expose a `Deref` (or a
`Partitions(name)` accessor). That is the stop this document names:

> **STOP IF** removing `Summary` from `View` requires deciding what else `View` should expose.

**Second, smaller, and also undecided:** *"with `Cites` stripped"* has no implementation on a
returned pointer. `Cites` is not a field — it is derived at read time by scanning content for
`Ref`s, including refs embedded in prose (`graph/types.go:86`, `content/content.go:104-141`). To
strip it from a `*graph.Engram` you must hand back a COPY with rewritten content, which requires
deciding what a removed `Ref` leaves in its place and what happens to `<obs-0001>` sitting inside
a sentence. Worth noting that today's `graph.Summary` already carries the raw `*Engram`, so the
protection `View` claims is presently nominal either way.

**WHAT WAS LEFT** UNCHANGED. `graph.Summary`, `graph.Summarize`, `Summary.HopKeys` and
`View.Summary` are all still in place, and `assemble.go` is untouched.

*(C3 did NOT need to be deferred into this item — the minimal form landed on its own, with both
call sites left alone as `01` requires. See commit `C3`.)*

---

## E3 — the sink's own signature is built out of E1's types

**WHAT WAS ASKED**

> ```go
> Crossed(span ids.SpanID, e ids.EngramID, ent ids.EntityID, r Rendering, raw Raw, err error)
> ```
> **Delete these** … `kernel/kernel.go indent · shortExpr · oneLine` … `9 Run.Logf calls`

**WHAT WAS HIT**

`Rendering` and `Raw` are E1, which stopped. Two of the five `Events` methods cannot be written
without them, and re-specifying `Crossed` to carry something else is designing the sink rather
than implementing it.

The deletions were not done alone on purpose. `indent`, `shortExpr`, `oneLine` and the nine
`Logf` calls are the only channel the kernel currently has; removing them before a sink exists
removes the observability and delivers nothing in exchange.

**Recorded for whoever picks this up:** the two lines that actually breach *the kernel never
reads content* are `kernel/kernel.go:538 shortExpr` (reads `Expectation.Expression`) and
`kernel/kernel.go:546 oneLine` (reads `Observation.Content`), reached from `open` and `commit`.
They are the whole of it. Everything else the kernel holds it holds without looking at.

**WHAT WAS LEFT** UNCHANGED. `kernel.Kernel.Run`, `entity/codex.Run`, `entity/runtime.Entity.Run`,
all nine `Logf` calls, the three helpers, and `trace.Run`'s existing `Logf`/`Step`/`Write`
surface.
