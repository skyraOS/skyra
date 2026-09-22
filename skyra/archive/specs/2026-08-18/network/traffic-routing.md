# Traffic routing

**Date:** 2026-08-20
**Status: WORKING.** The topology is agreed. Socket shape, addressing, acknowledgment, and retry
mechanisms remain open.

_Instantiates `the-quadrant.md` — this is the **TRUST line**, transport (how envelopes reach GREET / RECOGNIZE)._

> **One logical egress. Two ingress paths. A resolved connection gives the host network stack
> everything it needs to route the MOVE.**

This document separates host traffic from causal admission. The port subsystem moves and
normalizes envelopes. GREET and RECOGNIZE remain CTU primitives performed after an envelope
arrives.

---

## The topology

The single logical egress is provisionally called the **APPROACH socket**. `APPROACH` names the submission
socket only. It is not a primitive, and it is not ACTIVATE.

```text
                                      ┌─→ LOCAL INGRESS ────┐
CTU ─→ APPROACH SOCKET ─→ HOST STACK ─┤                      ├─→ NORMALIZE ─→ DISPATCH ─→ CTU
                                      └─→ remote host ─→ EXTERNAL INGRESS
```

Every outbound MOVE enters APPROACH. The CTU supplies the live connection required by MOVE, plus its
horizon and opaque payload. That connection already contains the destination address. The host
network stack applies its routing table and selects loopback or an external interface.

```text
LOCAL CONNECTION     addresses LOCAL INGRESS; the host stack routes it over loopback/local IPC.
REMOTE CONNECTION    addresses the remote runtime's EXTERNAL INGRESS; the host stack routes it
                     through the appropriate network interface and next hop.
```

The destination normalizes both ingress paths into the same arrival envelope. From that point
forward, locality is no longer observable to CTU semantics.

---

## Why one egress

Two caller-visible egress operations would leak placement into execution:

```text
MoveLocal(target, ...)
MoveRemote(target, ...)
```

That would make the caller choose topology and make migration change program behavior. One APPROACH
submission point keeps MOVE stable while connections and host routes change underneath it. The
implementation may still use separate local and remote connection pools, queues, or transports;
they remain host mechanism beneath APPROACH.

---

## Why two ingress paths

Local and external arrival are different trust and transport boundaries even though they become
the same causal event.

```text
LOCAL INGRESS      host-only. loopback or another local IPC mechanism. not externally exposed.
EXTERNAL INGRESS   network-facing. authenticates peers and carries external refusal policy.
```

Keeping them separate permits independent authentication, rate limits, queues, observability,
and backpressure. Normalization prevents those host differences from leaking above the driver.

The split is not request versus response. Both headers can travel through either ingress path:

```text
normalized REQUEST   → GREET    checks authority; admits work or refuses it.
normalized RESPONSE  → RECOGNIZE   correlates the parent and checks its frame-resident
                                    contract; admits a result or refuses it.
```

Thus the physical topology has two ingress paths, while the protocol still has two semantic
receiving ports. They are different decompositions over different axes.

---

## Addressing

An OS socket identifies a runtime transport endpoint, not necessarily one CTU. The envelope
identifies the destination within that runtime.

```text
HOST ADDRESS + OS PORT   which runtime receives the bytes.
ENVELOPE TARGET          which CTU / causal endpoint receives the MOVE.
HEADER                   REQUEST or RESPONSE.
CORRELATION              which outstanding parent position receives a RESPONSE.
```

This avoids assigning one scarce OS port per CTU and allows an endpoint to move between hosts by
changing routing state rather than changing its causal identity.

---

## Custody and completion

The path exposes three distinct facts:

```text
1  APPROACH accepted the envelope.             local egress now holds it.
2  destination ingress accepted it.       destination runtime now holds it.
3  GREET / RECOGNIZE admitted it.      destination CTU accepted the transition.
```

A successful socket write proves only the first fact. It cannot by itself mean that custody
crossed to the destination. The working direction is that MOVE completes on a destination-ingress
acknowledgment, with the exact ownership transfer and retry rule still to be decomposed.

The required invariant is stronger than reliable bytes:

> **At every point, exactly one side is responsible for the envelope, including during retry and
> ambiguous disconnect.**

---

## Ownership

```text
CTU               issues MOVE; performs GREET and RECOGNIZE after normalized arrival.
PORT SUBSYSTEM    owns APPROACH, local/external ingress, host connections, and queues.
FABRIC            the end-to-end custody-transfer relationship realized through those mechanisms.
HOST KERNEL       routes the addressed traffic; supplies sockets, descriptors, protocol state,
                  buffering, and devices.
SKYRA KERNEL      not interposed. its two CTU crossings remain ACTIVATE and REMEMBER.
```

Host-kernel socket activity is an implementation mechanism, not an additional Skyra-kernel
crossing.

---

## Open

```text
LOCAL TRANSPORT    TCP loopback, Unix-domain sockets, or another host IPC mechanism.
CONNECTION SHAPE   per MOVE, per endpoint, per host, or a pooled connection set.
DELIVERY ACK       wire shape and exact point at which custody changes owner.
RETRY              how ambiguous disconnect avoids both loss and duplicate admission.
CONNECTION SOURCE  how THINK supplies and refreshes the live connection MOVE requires.
BACKPRESSURE       one combined signal or separate local/external pressure.
EXTERNAL SECURITY  peer identity, encryption, replay protection, and refusal placement.
```
