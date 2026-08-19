# Host and reach

**Date:** 2026-08-15
**Status: PROPOSAL, and the thinnest file here.** Most of WHERE is specified and unbuilt.
`divergence.md` is right that this is a different class of gap: there is no line of code that is a
worse implementation of reach, because there is nothing to be worse.

> **An entity is an identity and does not carry its endpoints. Hosts hold endpoints. Address and
> transport stay behind the kernel boundary, and the resolver ranks while the kernel routes.**

---

## 1 · `Host` — capability classes

```go
package host

// Constitutive rather than declared: something with no execution class is not a host, it is a
// device we cannot use. An ABSENT implementation is a BUILD-TIME FACT, not a declaration.
type Host interface {
	ID() ids.HostID
	Execution() ExecutionOps // nil if not implemented
	Storage() StorageOps
	Endpoints() EndpointOps  // NEW. see §2.
}

// The trust boundary itself, not merely a call across it.
//   THE CALL     TRUSTED.   fixed semantics, kernel code.
//   THE PAYLOAD  UNTRUSTED. the process does whatever it wants.
type ExecutionOps interface {
	Spawn(program string, args []string) (Handle, error)
}

type StorageOps interface {
	Read(path string) ([]byte, error)
	Write(path string, body []byte) error
}

// Cut to what a handle IS: an opaque reference and the host it is valid on.
type Handle struct {
	Host ids.HostID
	Ref  string
}
```

```text
WHY Handle SHRANK   Stdout · Stderr · Exit are a SHELL'S SHAPE leaking into a class interface, on
                    the one surface selection cannot correct later. Spawn has never run — the
                    local driver returns "execution class not implemented" — so those three fields
                    are a guess nobody was forced to make.
                    → 01-removals.md C4.

WHY spawn(program)  of everything a spawn call takes on any platform, exactly two elements are
   → handle         cleanly universal: THE PROGRAM and A HANDLE BACK. Arguments do not round-trip
                    as a vector on Windows; environment names differ in case sensitivity; the
                    working directory is not a POSIX spawn parameter at all; extra descriptors are
                    not universal in any form.
                    → How-outcomes 2; How-evidence E8.

AND A PROGRAM IS    not a command line. A #! script works on POSIX because the KERNEL reads the
A PATH              first two bytes; `cd` is a builtin and `/bin/ls | grep x` is an expression,
                    and both fail with "no such file or directory."
                    → How-evidence E9.
```

```text
THE PROJECTION SUBSTRATE IS STILL UNNAMED.   host.go's own diagram has display and audio under a
                                             class with no name. future-research/projection-and-
                                             representation.md parks it deliberately and says why:
                                             nothing there should be settled before WHERE is.
```

---

## 2 · `EndpointOps` — the missing half

```go
// Hosts hold endpoints. An entity is reachable more than one way — through an API, through a
// gateway, through a CLI — and THOSE ENDPOINTS ARE NOT THE ENTITY'S. An entity carrying its own
// would be a declaration, and it would go stale the moment anything moved.
type EndpointOps interface {
	// Eligible endpoints for one entity, from this host. VOLATILE. never durable structure.
	For(e ids.EntityID) []Endpoint
}

// An Endpoint is A NAME PLUS THE PROCEDURE BY WHICH ITS NAMESPACE RESOLVES IT.
// The scheme IS the resolution procedure — file:// resolves by open/execve, https:// by
// socket/connect, unix:// by socket/connect on a path.
type Endpoint struct {
	Host   ids.HostID
	Scheme string
	Name   string
}
```

```text
THIS IS THE ONE ABSENCE BY OMISSION.   os-parallels.md §17 maps eight Linux namespace kinds onto
                                       this design. uts, ipc and time are absent BY DECISION — the
                                       design refuses a self-report, forbids a back channel, and
                                       has nothing to restore. `net` is the only one absent because
                                       nobody built it. That is how to rank the hole.
```

```text
SOURCE   Where-outcomes 4 (hosts have endpoints) and 9 (address and transport are two columns);
         Who-evidence E4 (the scheme is the resolution procedure);
         reach-boundaries-proposal.md §"Layer 3 · Actual Reach and the endpoint boundary".
```

> **QUESTION 8.1 — `Endpoints()` on `Host` or scoped per `Session`?**
> `os-parallels.md` §17 flags the disagreement and says drifting into one by accident is how the
> choice gets made without being made. `kernel.md` wants an entity→hosts **table**: a registry
> fact, global, rebuildable from configuration. Namespace logic says reachability should be
> **scoped per session**, the way netns makes "which network exists" a per-process answer rather
> than a lookup. **Taken: on `Host`**, with `Horizon.CanReach` doing the scoping — because the
> endpoint is a property of the device and the *permission* is what varies per session. But the
> netns reading is the more consistent one with everything else here, and I may have this
> backwards. → 09 F16.

---

## 3 · Reach, which has no method

Per D1, `Reach` is a protocol driven inside `Decide` and its public type is `Realization`
(05 §4). What belongs here is the shape beneath it.

```text
LAYER 1   reach-planning policy    entity → local or remote relationship expectations
LAYER 2   mode planning            request × entity × From × Horizon × experience
                                   → ranked expectation-observations THROUGH THINK
LAYER 3   actual reach             one decided composition
                                   → live possibility | bounded impossibility
```

**Only layer 3 may resolve and contact.** Layers 1 and 2 are `think-reach-entity` — learnable
behavior beneath Think that creates potential and contacts nothing.

```go
// The local/remote split is by the TERMINUS OF THE IMMEDIATE RELATIONSHIP, relative to
// From.Runtime. Not by transport, not by latency, not by trust.
type Locality uint8
const (
	Local  Locality = iota // an independent failure domain ON From.Runtime's host
	Remote                 // beyond it, crossing into another host or runtime authority domain
)
```

```text
AN INTERMEDIARY DOES NOT DISAPPEAR.   runtime ──local──▶ CLI ──remote──▶ model is TWO reaches and
                                      two independently attributable entities. If an intermediary
                                      can independently fail or alter what crosses, it is part of
                                      the causal path. Routers stay hidden because they route
                                      BELOW the endpoint abstraction; an application proxy
                                      interprets the exchange and therefore does not.

REACH FAILS BEFORE THE POSSIBILITY    world.md placed reachability at DISPATCH months before
OPENS, AND LOUDLY                     resolution was discussed: "an undeliverable member returns
                                      instantly and does not block its group." So an exhausted
                                      resolver fails on the same clock as a TV with no speaker.
                                      One mechanism, not two.
```

### `Distance` is not a type

```text
ORIENT × ENDPOINT × EXPERIENCE → DISTANCE
```

Computed at each reach, from where you actually are, and **never stored**. A stored distance is a
declaration with a timestamp on it, wrong the moment anything moves; computed, it cannot go stale —
only be out of date, which the next observation fixes. That is why the third term is `Experience`
rather than `Topology`.

```text
THE KERNEL SUPPLIES   orient (live topology) · endpoints (host-held) · experience (history)
THE RESOLVER RANKS    and THE RANKING IS DISTANCE. So it is not an eighth primitive — it is what a
                      reach engram computes, which makes it learnable and lets strategies compete.

MEASURE AT REACH,     REACH = time to obtain a handle ← execution distance
NOT AT RETURN         WORK  = time to obtain an answer ← not distance at all
                      Conflating them makes a hard question look like a far entity, and the
                      resolver learns to avoid an endpoint for being asked difficult things.
```

Two distances, pointing opposite ways:

```text
EXECUTION DISTANCE   low is good.   it cost little to get there.
CAUSAL DISTANCE      high is good.  its influence survived a long way from its origin.
                     ← already computed. HopProfile() exists and predates the word.
```

```text
SOURCE   distance.md, whole; reach-abstraction.md §5 for what the OS drops and this must carry;
         Where-evidence E10.
```

---

## 4 · What is unbuilt, named plainly

```text
NO RESOLVER                    nothing ranks anything.
NO ENDPOINT TABLE              EndpointOps above is proposed here and exists nowhere.
NO DEVICE IDENTITY             Where-outcomes 2 specifies a hash over hardware fields read at
                               boot. Nothing reads any of them.
Invocation.Host IS HARDCODED   written once as k.Host.ID(), read nowhere. Under Where-outcomes 8
                               it stops being dead weight: it records the resolver's choice, and
                               the resolver reads its own history back.
Project HAS NO HOST            "NOTHING SUPPLIES h." Under os-parallels.md §18, REACH SUPPLIES IT:
                               resolve (WHO) → bound by horizon (WHERE) → Project (aim) → Step
                               (cross).
Spawn HAS NEVER RUN            the driver surface is three ops and one does not exist, which makes
                               "execution now, storage soon" currently backwards.
```

> **QUESTION 8.2 — `{path: "file.txt"}` must stay unresolved, and no type here enforces it.**
> The correct outcome for the walkthrough is a **loud refusal**: a relative path names no host and
> no namespace, Horizon can bound authority to an anchored world but cannot decide which world was
> meant, and guessing creates both location and authority. `content.WorldRef` (02 §1) is the
> anchored form — it carries a `HostID` — but nothing forces a path to be one. A `Str` containing
> `"file.txt"` is expressible and will be written. **The type surface cannot fix this**; it is a
> contract question, which means it lands on the parent that owns the position. Recorded so it is
> not assumed handled.
