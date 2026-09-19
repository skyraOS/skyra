# Message — the wire unit (Seed dissolved)

**Date:** 2026-08-27
**Status: SUPERSEDED (2026-08-28) by `move-and-possibility.md`.** Pass 2's premise — "the Body already IS the
moves (`Body.Moves` + `MoveTemplate.Reads`), so a Move field on the wire is redundant" — no longer holds: the
move/possibility collapse removed `Body.Moves` and `MoveTemplate.Reads`, and a remote invoke now carries
{Target, Action, ReturnTo} (the target resolves its own Body). So **`Message` is retired and the Move IS the
wire format** (re-aligning with `what-travels.md`); the response leg became the `Landing` type this doc left
OPEN. The envelope-dedup and Seed history below remain useful context. Original status follows.

**Status (original): DECISION, CODE APPLIED (two passes).** Pass 1: `Seed` (a Context-shaped pre-image) dissolved into
`Message`. Pass 2 (below, "collapsed AGAIN"): since the **Body already IS the moves** (`Body.Moves`) and their
inputs (`MoveTemplate.Reads`), the `Move` and `Inputs` fields were redundant and removed — the forward
Message is now `{Possibility, Body, ReturnTo}`. Moves come OFF the wire (MoveTemplate travels in the Body; the
runtime Move is local). Applied to causal.go + threaded (Transfer / Egress / Ingress). Build+vet green.
The narrative below is pass 1; the current shape is the "collapsed AGAIN" section.

---

## Why Seed was awkward (the diagnosis)

```text
· Seed = {Possibility, Body, Registers, ReturnTo} — literally a Context minus its runtime accumulation +
  Result. a Context PRE-IMAGE. reads like a half-duplicate of Context.
· it embedded a full Possibility, while the Move ALSO carries MoveID.Possibility → felt duplicated.
```

## Why the obvious fixes don't work, and the one that does

```text
· "Possibility IS the wire format"     — NO. we just moved Body + ReturnTo OFF Possibility (it is now
                                         identity + authority only). the target could not build a Context.
· "{Possibility, Move} is the wire"    — CLOSE, but short two fields: the target also needs Body (what code)
                                         and ReturnTo (where home). the exact bits we relocated to Context.
· MESSAGE = {run-context} + Move       — YES. the run-context (Possibility + Body + ReturnTo) is the SAME for
                                         every move of a run; carry it ONCE per delivery, not embedded per
                                         move. Seed dissolves into these envelope fields + the one Move.
```

## The shape (causal.go) — collapsed AGAIN: the Body already IS the moves

A second pass cut it further. The Body already contains the moves (`Body.Moves []MoveTemplate`) and each
move's inputs (`MoveTemplate.Reads`). So a separate `Move` field and an `Inputs` register-map were BOTH
redundant — over-decomposed. Removed. The forward/distribute Message is now just:

```go
type Message struct {   // FORWARD (distribute) leg
    Possibility *Possibility // WHICH RUN — identity + authority (Horizon rides here; can't be reconstructed)
    Body        BodyID       // the run's code = a series of moves; digest, bytes pulled from the engram
    ReturnTo    *Route       // the caller's fd this run's result fills (To is sender-derivable)
}
```

Consequences:
```text
· MOVES COME OFF THE WIRE. MoveTemplate (static, inside the Body) travels; the runtime Move (with State) is
  instantiated LOCALLY in the WorkList. Move is no longer a wire field — it is a runtime instance.
· INPUTS ARE NOT PUSHED. MoveTemplate.Reads names the inputs; the VALUES arrive as normal dataflow LANDINGS
  into registers, not pre-filled on the message. (this answers the old push/pull question toward PULL/flow.)
· THE RESPONSE LEG IS A LANDING, not a Move. {fd, observation} — needs its OWN small wire type (OPEN).
· ReturnTo is thin — effectively just the Fd (the caller's register to fill); its To is reply-to-sender.
```

## What this does and does not dedup

```text
· DEDUPS: the whole Possibility STRUCT (+ Body + ReturnTo) — carried ONCE per delivery, not per move.
· DOES NOT (and cannot) dedup: MoveID.Possibility — the id STAYS on the Move because it is the local
  WorkList/Cache key (loose coupling: MoveID.Possibility says which Context to read/write). the id
  necessarily appears; only the heavy struct is lifted to the envelope.
```

## Threaded through the wire path

```text
· ctu.CTU.Transfer(msg c.Message)         — the door now takes a Message (was Move).
· ctu.Egress interface  Egress(from, msg) — a cut sends a Message.
· kernel.Barrier.Egress/Ingress(msg)      — sign/park/verify a Message; socket table still keys on MoveID.
· kernel.Faults.Materialize — NOTE added: it wants the MESSAGE (run-context) to BUILD the Context, not a
  bare Move. signature reconciliation DEFERRED with the handler bodies (kernel-handler-block.md).
```

## Open

```text
· THE LANDING TYPE — the response leg is {fd, observation}, not a Move. give it a small wire type; decide
  whether Transfer takes a sum (forward Message | landing) or two entry points. (the clearest next step.)
· ReturnTo → could shrink to just the Fd (To = reply-to-sender, barrier-derived), or move ENTIRELY to the
  barrier (making the forward Message exactly {Possibility, Body}). confirm.
· Move.Header {Request,Response} — with moves off the wire, is Header still earned on the runtime Move? (COMB)
· the NAME: `Message` is standard/clear; could take a more Skyra-flavored word (Delivery, Dispatch, Parcel).
  rename is cosmetic — the shape is the decision.
```
