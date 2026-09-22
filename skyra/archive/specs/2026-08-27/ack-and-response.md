# Ack and response — two planes, two tables

**Date:** 2026-08-27
**Status: EXPLORING — MECHANISM PREDATES the wire-unit collapse (09-01).** The two-plane CONCEPT (system ack
vs entity response) may survive, but its hook is gone: **`Header{Request, Ack, Response}` no longer exists** —
request is a `Move`, response is a `Landing` (two types, no enum on a Move), so "Transfer's Header switch
grows a third arm" doesn't apply as written. `Work[Issued]` is now the separate in-flight/socket table, not a
bucket of Work. Re-derive the ack/response split against the Move/Landing + socket-table model
(`move-and-possibility.md`, `register-lifecycle-and-transition.md`) before combing.

A crossing carries TWO kinds of "response" that our single `Response` conflates:
the SYSTEM ack (transport/custody — control plane) and the ENTITY response (the semantic result — data
plane). Making the split explicit, and naming the two tables that track them. Refines `dispatch-and-
decode.md` (the LAND arm) and `move-and-the-kernel-barrier`.

---

## The two axes

```text
ENTITY response   the SEMANTIC result — the observation the target entity produced. lands in the register
                  (the fd). DATA PLANE. answers "what did it compute?"
SYSTEM ack        the TRANSPORT acknowledgment — received / parked / accepted (or denied / faulted). says
                  nothing about the answer; it is about CUSTODY. CONTROL PLANE. answers "did it arrive?"
```

Every real protocol keeps these apart: TCP's ACK (got your bytes) ≠ the app's reply; SIP's 1xx provisional
(Trying / Ringing) ≠ the 2xx final answer. **Reliability is separable from semantics:** a missing ACK means
RETRANSMIT (lost in transport); a missing RESPONSE means STILL WORKING (semantics pending). One message type
cannot tell those apart — hence the split.

## Two tables, two owners (they already exist)

We do not invent two new structures — we recognize that each plane already has its table and its owner:

```text
PLANE             TABLE                     OWNER    concern
─────────────     ───────────────────────   ──────   ────────────────────────────────────────────
system / ack      the BARRIER socket table  KERNEL   custody · signing · retransmit · timeout   (control)
entity / response Work[Issued]              CTU      the fd · dataflow · LAND                     (data)
```

Ownership-clean: the kernel already owns the barrier and custody — transport reliability IS its job; the CTU
already owns `Work[Issued]` and dataflow — the semantic wait IS its job. The split just stops the socket
table from doing double duty.

## The cut is a two-hop state machine

```text
SENT ──(system ACK)──▶ ACCEPTED ──(entity RESPONSE)──▶ COMPLETED
        retransmit if                  still-working if
        no ack (TRANSPORT)             no response (SEMANTIC)
```

- **Hop 1 (transport):** egress signs + parks in the socket table + sends; the far barrier ACKs custody.
  No ack before timeout → RETRANSMIT. This hop is entirely between the two KERNELS/barriers.
- **Hop 2 (semantic):** the entity does the work and returns its observation; ingress pops the socket table
  and hands the RESPONSE to `CTU.Transfer`, which LANDs it on the fd. This hop is the CTU's dataflow.

## Message types — Header gains a plane

```text
Request    entity, forward   — "do this action"                        (data plane, outbound)
Ack        system, backward  — "received / parked" (or Nack: denied)   (control plane)
Response   entity, backward  — "here is the result"                    (data plane, the semantic LAND)
```

(COMB: is it `Header{Request, Ack, Response}` — a flat enum — or a 2-axis tag {leg: fwd|back} × {plane:
system|entity}? the 2-axis form is truer but heavier. lean: flat enum now, note the axes.)

## Consequence for dispatch — LAND splits

From `dispatch-and-decode.md`, the inbound `LAND` arm becomes two:

```text
ACK inbound        update the TRANSPORT table (socket table): stop the retransmit timer, mark ACCEPTED.
                   NO data lands; the move stays Issued awaiting the entity response. (kernel/barrier)
RESPONSE inbound   the real semantic LAND: write register · POP Issued · FLIP dependents. (CTU)
```

So `Transfer`'s coarse Header switch grows a third arm (Ack), and only the Response arm reaches the
dataflow. An Ack never touches registers.

---

## Ideas to dig into (Claude's seeds)

```text
A  FIRE-AND-FORGET STILL GETS AN ACK. a move with Return=false has no semantic RESPONSE — but it still
   wants a transport ACK (did it arrive?). so the two planes are INDEPENDENT: ack-yes/response-no is a valid
   combination (a delivered notification). this is why they cannot be one table.

B  NACK / DENIAL IS CONTROL-PLANE. a request rejected at ingress (identity gate, custody fail, unprovisioned)
   comes back as a system NACK — NOT an entity failure. contrast idea C in dispatch-and-decode (a Perform
   that returns Failure is an ENTITY response carrying failure DATA). two different failures, two planes.

C  WHAT THE ACK CARRIES. minimally a custody confirmation. maybe more: a park-id, an ETA, a "queued behind
   N". provisional responses (SIP 1xx) suggest the ack channel can carry progress, not just a boolean.

D  RETRANSMIT / IDEMPOTENCY. retransmit on missing ack requires the receiver to DEDUPE (idempotent on the
   move's id, like Materialize is on move.ID) so a re-sent request is not double-executed. the ack plane and
   the fault-retry plane want the SAME idempotency key.

E  TIMEOUT vs EXPIRY. the transport timeout (no ack → retransmit) is different from the capability ExpiresAt
   (the semantic lease). two clocks: one for delivery, one for authority.

F  DOES THE CTU EVER SEE THE ACK? maybe not — the barrier (kernel) could absorb the whole transport hop and
   only hand the CTU a RESPONSE. then the ack table is PURELY kernel-internal and the CTU's Work[Issued] only
   ever sees semantic landings. (lean: yes — keep transport OUT of the CTU, consistent with the two-fault
   surface. the CTU is oblivious to retransmit.)
```

## Open — first to settle

```text
· idea F — is the ack plane entirely kernel-internal (CTU sees only Response)? decides whether Header even
  needs an Ack value the CTU handles, or the barrier absorbs it.
· Header shape — flat {Request, Ack, Response} vs the 2-axis tag.
· the ack table's relationship to the existing socket table — same table with a state field, or two.
```
