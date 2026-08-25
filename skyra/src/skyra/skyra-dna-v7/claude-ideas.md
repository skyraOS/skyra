# Claude's ideas — v7 interface choices

**Date:** 2026-08-22
**Author: Claude. These are MY proposals, not design canon and not yours.**

> Everything below is a modeling decision **I** made while translating
> `skyra/specs/2026-08-18/` into Go interfaces. **None of it is in the docs.** The docs settle
> *concepts*; these settle *how I typed them*, and I had to pick. Read each as a proposal to
> accept, reject, or correct. Where a choice takes a side on an open doc question, it says so.

The terms are all from the docs (see `interface-rationale.md`). What's mine is the **shape** —
which types exist, what they return, how the packages split.

---

## 1 · A `causal` leaf package holds the cross-boundary types

**Mine:** one package, imported by ctu/kernel/drivers/entity, holding everything that crosses.
**Why:** Go import cycles — if these types lived in `ctu`, then `kernel` importing them and `ctu`
importing `kernel` would cycle. A leaf package is the standard fix, and it doubles as "here is
exactly what crosses boundaries," which is the whole point of this build.
**Reject if:** you'd rather each party own its types and accept a messier import graph, or you want
the boundary types to live with the boundary they cross rather than pooled.

## 2 · `Content` and `Observation` are TWO types

**Mine:** `entity.Run` returns `Content` (raw); only the CTU makes an `Observation` (Content +
Provenance).
**Doc basis:** decode-and-run.md — "the entity yields content; the CTU stamps provenance." I turned
that sentence into two distinct Go types so the trust boundary is a *type error* to cross, not a
convention.
**Reject if:** you think one type with a "stamped?" field is enough. I think two types is safer —
you can't accidentally return an entity's word as a fact.

## 3 · `Outcome` is a 3-value enum, and it's the return of Move / Greet / Remember

**Mine:** `Success | Failure | NotYet` as one type, used everywhere a crossing can be "not yet."
**Doc basis:** move.md's three outcomes. The *name* `Outcome` is my invention (the values aren't).
**Reject if:** you'd split NotYet out as a separate signal, or want richer failure detail than one enum.

## 4 · Header + Correlation live in the ENVELOPE (I picked a side of a real seam)

**Mine:** `Move.Header` and `Move.Correlation` are envelope fields.
**The conflict is real and not mine:** move.md says both live in the *payload*; traffic-routing.md
treats both as envelope addressing. **They disagree in your docs.** I put them in the envelope
because routing to GREET vs RECOGNIZE needs the header before anyone opens the payload — and I left
a `// SEAM` on `Payload` so it's visible.
**Reject if:** you meant the payload — but then explain how the fabric routes request vs response
without reading it. This is the crossing I'd resolve first.

## 5 · `SysInstruction` is an optional pointer

**Mine:** `*SysInstruction` (nil = absent).
**Doc basis:** move.md — "absent by default." A pointer makes "unspecified → scheduler free" the
zero value, which is the common case.
**Reject if:** you want a value type with an explicit "present" bool, or a different optionality idiom.

## 6 · `Resource{Name, Trusted, Kernel}` — trust/privilege as provenance bools

**Mine:** two bools on a resource: `Trusted` (which table it came from) and `Kernel` (is it THE
named kernel).
**Doc basis:** dispatch.md — "named, not flagged," trust = which table. I modeled those as fixed
provenance on the resource, NOT as flags on the MOVE. If these ever move onto `Move`, that's the
dispatch.md violation to catch.
**Reject if:** you'd rather resources be distinct types (HostResource / EntityResource / the kernel)
than one struct with bools. That might be cleaner and I'm not sure I got this one right.

## 7 · `Interpreter` and `Scheduler` are separate interfaces from `CTU`

**Mine:** three interfaces in `ctu`, not one.
**Doc basis:** decode-and-run.md names the interpreter and the scheduler as distinct steps. I kept
them separate so the interpret→schedule seam is a package boundary you can see.
**Reject if:** you'd fold them into `CTU` since the CTU owns all three anyway.

## 8 · Kernel signatures: `Activate(Name,Horizon)→Possibility`, `Remember(Observation)→Cone`

**Mine:** Activate returns a live `Possibility`; Remember takes the kept `Observation` and returns
the `Cone` it wrote.
**Doc basis:** primitive.md / boundaries.md (load→possibility; remember writes the cone of the kept
observation). The exact params are my reading.
**Reject if:** Remember should take the root/arena explicitly, or return an Outcome instead of the Cone.

## 9 · Driver mirrors the crossings one layer down — and there are TWO `Activate`s

**Mine:** `drivers.Driver.Activate(Name)→Engram` and `Remember(Cone)→Outcome`, distinct from the
kernel's Activate/Remember.
**Doc basis:** durable-storage.md — the driver persists (HOW), the kernel decides (WHAT). So the
driver fetches an `Engram`; the *kernel* turns it into a `Possibility`. Two Activates at two layers.
**Reject if:** you want the driver to return raw bytes and the kernel to parse, or a single crossing
type shared by both.

## 10 · `Bound` attenuates (CTU); `Define` is a KERNEL method — CORRECTED

**Mine (corrected):** `ctu.Bound(Horizon,Caveat)→Horizon` narrows only. `kernel.Define(Pin)→
(Account,Horizon)` is the privileged DEFINE half.
**What I got wrong first:** I said "DEFINE has no method because it names the kernel." That was
backwards — if it NAMES the kernel, it IS a call to the kernel, i.e. a **kernel method**. The
struct-tie audit (§13) caught it: `Account` and `Pin` were orphans, which meant authentication had
no interface. Adding `kernel.Define` ties them and fixes the decomposition.
**Doc basis:** horizon-and-root §2a, boundaries.md SEAM 2 (Model A). Note the kernel now has THREE
methods — two memory crossings (Activate/Remember) + one authority-tier write (Define) — which is
exactly Model A, not a violation of "two crossings" (that was always memory-tier).
**Reject if:** you chose Model B (define crystallizes at REMEMBER) — then Define folds into Remember.

## 11 · Struct fields are deliberately thin

**Mine:** minimal fields — enough to show the crossing, not a full data model (`Name{Value}`,
`Goal{Statement}`, `Engram` with an opaque continuation).
**Why:** fleshing these out is design work, not interface work, and I didn't want to invent a data
model you didn't ask for. Thin fields mark where the shapes still need to grow.
**Reject if:** you want me to push specific ones deeper — say which.

---

## 12 · Every name I introduced (struct + field audit)

You asked for a full accounting of new names. I audited every type and field in `causal.go`.

**Type names — all doc terms except one:**

```
Outcome   the only new TYPE name. declared in §3. (values Success/Failure/NotYet ARE doc terms.)
```

Everything else (Move, Header, Target, Horizon, Caveat, SysInstruction, Payload, Correlation,
Observation, Content, Provenance, Contract, Potential, Goal, Frame, Possibility, Outstanding,
Name, Engram, Cone, Account, Pin, Resource, EligibleSet) is a doc term.

**Field names — these are MINE. Two carry a decision; the rest are placeholder scaffolding:**

```
CARRY A DECISION (see the section noted):
  Resource.Trusted   bool — "which table" as provenance.   §6.
  Resource.Kernel    bool — "the named kernel" as a mark.  §6.

PLACEHOLDER SCAFFOLDING (generic Go names for thin structs — the DOC CONCEPT is what matters,
the field name is throwaway and should be replaced when the shape grows):
  Caveat.Constraint          the doc concept is "a caveat that narrows." name is filler.
  SysInstruction.Constraint  the doc concept is "a narrowing placement constraint." name is filler.
  Correlation.ID             move.md says "an ID, a position, or the connection" — I picked ID, unresolved.
  Content.Bytes              raw content had to be *something*; bytes is a placeholder.
  Goal.Statement             a goal is given to THINK; "Statement" is filler.
  Outstanding.Count          outstanding is a count/set; I used a count.
  Name.Value                 a namespace name needs a value; "Value" is filler.
  Pin.Secret                 a pin authenticates; "Secret" is filler.
```

**Doc-derived field names (not inventions, but worth naming so nothing hides):**

```
  Target.Endpoint        from "one CTU per endpoint."
  Provenance.Invocation  from continuation-language "which engram and invocation produced it."
  Potential.Signature    from "a potential is a method signature."
  Frame.Scope / .Return  from "observations in scope" and "the return contract."
  Possibility.Parent     from "a possibility with no parent is the root."
  Engram.Contracts       from "engram = child contracts + compiled continuation."
```

Nothing in the structs introduces a new *concept*. The only inventions are the type name `Outcome`,
the package name `causal`, and the placeholder/decision field names above — all now listed.

---

## 13 · Struct → interface map (main vs secondary, and the orphan it found)

Your test: every struct must tie to an interface; if it can't, a concept wasn't decomposed. I ran it.

**PRIMARY** — appears directly in an interface signature (it CROSSES a boundary):

```
Move          → ctu.Move / .Greet, ctu.Interpreter.Interpret
Payload       → entity.Entity.Run
Outcome       → ctu.Move/.Greet/.Recognize, kernel.Remember, drivers.Remember  (returns)
Goal, Frame   → ctu.Think
Potential     → ctu.Think  (return)
Horizon       → ctu.Bound, kernel.Activate, kernel.Define
Caveat        → ctu.Bound
EligibleSet   → ctu.Interpreter.Interpret (return), ctu.Scheduler.Schedule (param)
Resource      → ctu.Scheduler.Schedule  (return)
Content       → entity.Entity.Run  (return)
Observation   → ctu.Recognize, ctu.Scheduler.Schedule, entity.Entity.Run, kernel.Remember
Possibility   → kernel.Activate  (return)
Outstanding   → ctu.Orient  (return)
Name          → kernel.Activate, drivers.Activate
Engram        → drivers.Activate  (return)
Cone          → kernel.Remember (return), drivers.Remember (param)
Account, Pin  → kernel.Define
```

**SECONDARY** — a field of a primary; ties to an interface through its parent:

```
Header, Target, SysInstruction, Correlation  → fields of Move
Provenance                                    → field of Observation
Contract                                      → fields of Engram, Frame (frame-resident; never a
                                                 param directly — worth noting for RECOGNIZE)
```

**THE ORPHAN THE TEST FOUND:** `Account` and `Pin` tied to nothing — authentication had no
interface. That is your "didn't break it down well enough." Fix: `kernel.Define` (§10). After it,
**no orphans remain.**

---

## The two I'd double-check first

```
#4  Header/Correlation home     — your own docs disagree; I guessed envelope.
#6  Resource as bools           — might want distinct types instead; least sure of this one.
```
