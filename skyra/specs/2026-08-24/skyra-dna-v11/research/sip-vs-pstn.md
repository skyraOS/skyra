# SIP vs PSTN — where the intelligence lives

**Date:** 2026-08-24
**Status: RESEARCH POINTER.** A telephony contrast worth stealing from for the coordination /
distributed-kernel question. Relates to `../ownership.md` (kernel authorizes, doesn't coordinate),
`../bootstrap/sequence.md` (A: entity→host→address), and the deferred distributed kernel (B).

---

## The contrast

```text
PSTN   circuit-switched. SMART NETWORK, DUMB ENDPOINTS. the network sets up the circuit, owns the
       intelligence (SS7 / Intelligent Network), and coordinates the call end-to-end. numbers = E.164.
SIP    packet-based sessions. DUMB NETWORK, SMART ENDPOINTS. endpoints establish sessions peer-to-peer;
       the network just helps them FIND each other (registrar/proxy) and gets out of the way. URIs =
       sip:user@domain. media flows edge-to-edge after setup.
```

## Why it maps to Skyra

Skyra leans **SIP**, and we already decided the pieces:

```text
· kernel AUTHORIZES, does not coordinate       → SIP: the network doesn't own the call; endpoints do.
· ENTITIES coordinate (holding capabilities)   → SIP: intelligence at the EDGES.
· host(entity) = the home-registry / HLR       → SIP REGISTRAR: an endpoint registers its location so
                                                 others can find it; the address stays volatile/edge-resolved.
· media flows peer-to-peer after setup         → after the kernel provisions, entities reach each other
                                                 DIRECTLY (hot path, kernel-free). the kernel is only in
                                                 the loop for setup/authorization, like a SIP proxy.
```

PSTN is the thing to **avoid**: a smart central network that coordinates. That's the "kernel as
Kubernetes control plane" mistake we already rejected. Skyra's kernel is the SIP registrar/proxy —
find + authorize — not the PSTN switch.

## To explore later

```text
· the distributed kernel (bootstrap B, deferred) is really "how do SIP endpoints find + authorize each
  other across domains" — proxies, registrars, trust between domains (SIP federation).
· ENUM already bridges E.164 → SIP (the cellular note) — the PSTN→SIP seam is a known, real thing.
```
