# Resolution — the frozen-interface gaps Codex surfaced

**Date:** 2026-08-26
**Status: PROPOSE-LANE RESOLUTION.** The CTU hot-path handoff (`overnight.md`) hit five frozen-interface
gaps and the implementer STOPPED without editing — as instructed. This records the design calls that
close them, and the concrete type changes to apply before the handoff is re-issued.

> **One root cause under most of it: the `Transfer → Move` fold was too aggressive.** Folding the wire
> envelope into `Move` dropped the `Child *Context` seed — and a REQUEST is the move that MATERIALIZES a
> context, so it must carry the seed that context is built from. There is no context yet to read from.

---

## The horizon question (the tell)

"Why does a move carry a horizon — isn't the horizon on the context?" Correct for every RESIDENT move: it
reads the horizon off its context, carries nothing. The exception is the **Request leg**: a Request
materializes a *fresh* context on the far side, so there is no context yet — the horizon (the held
capability) is verified OFFLINE at ingress, before any context exists, and therefore must travel WITH the
request. So it is not "moves carry horizons"; it is "a Request carries a **seed**, and the horizon is one
part of that seed" (with the BodyID and the input values). That seed is the `Child *Context` we deleted.

---

## The five gaps and their resolutions

```text
#1  REQUEST LEG LOST ITS SEED  (BodyID + Horizon + seed inputs)
    ROOT      the Transfer→Move fold dropped Child *Context. A Request can't verify/materialize without it.
    RESOLVE   restore a seed on the Request leg as `Seed *Context`. A Context is now LEAN (we stripped all
              move-state — no Attention), so a seed is genuinely small: body ref + seed registers + horizon
              + return route, Result=nil, no moves loaded. Materialize consumes it → keeps Materialize
              dealing in Contexts (symmetric with Remember). One field restores all three dropped pieces.
    NOTE      Seed.Body is OPTIONAL: present = a known/migrated body (the user's REMEMBERED body, hot-path
              entry); zero = a cut whose callee resolves its own body (out of scope for this handoff — the
              network is mocked, the response is supplied).

#5  RETURN ROUTE CAN'T CORRELATE  (ReturnTo is only a Target, no fd)
    ROOT      a completed child must land its Response on the PARENT's specific move (the fd). ReturnTo
              carried only a Target — no MoveID — so the response can't be correlated.
    RESOLVE   introduce `Route{ To Target; Fd MoveID }`. The child completes → Response{ ID: route.Fd,
              Target: route.To, Returned: result }. Same root cause as #1 — the request leg was underspecified.

#2  DRIVERDECL IS METADATA ONLY  (no callable mechanism)
    ROOT      Core.Transition can't RUN a local move — DriverDecl is just {Target, Actions} for the kernel.
    RESOLVE   add a `Driver` interface (the mechanism): Decl() DriverDecl + Perform(action, in) →
              (Observation, Outcome). Core holds `Local map[Target]Driver` — its installed drivers.
              DriverDecl stays the kernel-facing ADVERTISEMENT; Drivers.Advertise() = Decl() over Local.

#3  CORE HAS NO INTERPRET
    ROOT      the contract calls Interpret(ctx, move); Core has no handle to it, tests can't stub it.
    RESOLVE   Core holds `Interp Interpret` (a field/seam). Interpret answers the HORIZON question (is the
              target provisioned? no → Fault). Core answers the MECHANISM question (is it in Local? →
              LOCAL, else CUT). Two orthogonal checks — authorization vs locality.

#4  CORE HAS NO SELF  (can't set from / return-to-self)
    ROOT      issuing a cut needs `from` for Egress and a return route pointing back HERE; Core knows
              neither.
    RESOLVE   add `Self Target` to Core. Cut for move M in ctx → outbound Seed.ReturnTo = { To: Self,
              Fd: M.ID }; Egress(from: Self, m: outbound).
```

---

## The type changes to apply

```go
// causal.go — the return route (new)
type Route struct {
	To Target // where the response goes
	Fd MoveID // which move it lands on (the parent's fd)
}

// causal.go — Context.ReturnTo becomes a Route
type Context struct {
	ID        PossibilityID
	Body      BodyID
	Registers map[RegKey]Register
	Horizon   Horizon
	Result    *Observation
	ReturnTo  *Route // was *Target — now carries the fd for correlation
}

// causal.go — Move: drop ReturnTo, add the Request-leg Seed
type Move struct {
	ID       MoveID
	Template MoveTemplate
	State    MoveState
	Header   Header       // Request | Response
	Returned *Observation // Response leg
	Seed     *Context     // Request leg: the seed to materialize (Body+Horizon+seed Registers+ReturnTo). nil else.
}

// ctu.go — the driver MECHANISM (new), alongside the existing DriverDecl advertisement
type Driver interface {
	Decl() DriverDecl
	Perform(action c.Name, in []c.Observation) (c.Observation, c.Outcome)
}

// ctu.go — Core gains Self, Interp, Local
type Core struct {
	Self    c.Target             // this CTU's identity — from for Egress, To for the return route
	Cache   c.Cache
	Bodies  c.BodyCache
	Work    c.WorkList
	Interp  Interpret            // the horizon block-selection seam Transition calls
	Local   map[c.Target]Driver  // installed local mechanisms — resolves LOCAL vs CUT
	Out     Egress
	// + a wake signal for Run
}
```

`Drivers.Advertise()` becomes `Decl()` mapped over `Local`. `EligibleSet` and the rest are unchanged.

---

## What did NOT change

- **The horizon still lives on the Context** for every resident move — the fix is only about the Request
  leg, which has no context to read from yet.
- **No `Attention` reintroduced** — the fd table is still `Work[Issued]`; `Route.Fd` is how a response
  finds its slot.
- **Scope is unchanged** — still hot-path only; the cold path, engram disk, and callee-side body
  resolution stay deferred.

## Next

Apply these to `causal.go` / `ctu.go`, re-run `go build && go vet`, then re-issue the `overnight.md`
handoff against the re-frozen surface. The five gaps become five satisfied preconditions; T1/T3/T5 can
then be built honestly.
