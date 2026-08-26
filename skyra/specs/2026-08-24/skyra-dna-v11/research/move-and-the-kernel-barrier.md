# Move and the kernel barrier

**Date:** 2026-08-24
**Status: EVOLVING — actively worked, not settled.** Recorded so the model isn't lost; expect it to
change. Each section is tagged CONVERGING (leaning settled) or OPEN. Builds on `cpu-execution-model.md`
(body/context), `socket-triple-to-fd.md` (fd/socket), and `../ownership.md`.

---

## 1 · The chain — CONVERGING

```text
entities ──▶ INGRESS ──▶ CTU.Move ──▶ EGRESS (approach) ──▶ (out)
```

- **The CTU owns `Move`** — the single entry point. Ingress calls it; local THINK calls it too.
- The CTU calls **egress** to send outward.
- **Only CROSSING moves hit the barrier.** A local (RunHere) move executes *inside* the CTU and never
  touches ingress/egress or the block.
- **Naming caveat:** this INVERTS v9's `Move` (which was the egress send). Here `Move` = the CTU's
  execute-entry; the outward send is egress / approach. The term flips meaning — state it loudly.

## 2 · `Move` subsumes the ingress verbs — CONVERGING

`Move(Transfer)` dispatches on `Header`:

```text
Header = Request   → admit the child possibility, make resident, Drive     (was Greet)
Header = Response  → land the observation at its join, re-drive            (was Recognize)
```

So the four v9 ingress verbs (Greet / Recognize / Arrive / old-Move) collapse to one `Move`; egress is
the outward send.

## 3 · Ingress/egress = the kernel barrier — CONVERGING

The barrier is where trust + authority cross, which is kernel domain:

```text
EGRESS (out)   attenuate the horizon · SIGN the segment (custody) · park in the block.
INGRESS (in)   VERIFY the returned segment (custody) · pop the block · hand to CTU.Move.
```

The fabric moves bytes *below* the barrier; the kernel *guards* it. **D5 reconciliation:** a trap is a
crossing whose *target is the kernel itself*. All exits go through egress — peer CTUs are horizontal,
the kernel/privilege is vertical. One door; the target decides the direction.

## 4 · The in-flight block = kernel socket table; the possibility holds fds — CONVERGING

The socket model settles the ownership: **sockets are kernel objects; a process holds only an fd.**

```text
KERNEL owns the BLOCK      = the SOCKET TABLE. cut-ID → { owning CTU/possibility, custody, pending
                             response, timeout }. the crossing STATE. kernel-global, at the barrier.
POSSIBILITY holds HANDLES  = the FD TABLE. cut-ID → register (which register this handle fills).
                             per-possibility, its own dataflow.
```

A **cut-ID is a file descriptor**, and a cut's lifecycle is a socket's lifecycle:

```text
CTU.Move issues a cut  → EGRESS: open a socket entry · attenuate · SIGN · send · PARK · return the handle.
                         the possibility records handle → register in its fd table; Drive returns parked.
response arrives       → INGRESS: recv · VERIFY · look up routing in the block · deliver to owning CTU · CLOSE.
CTU.Move (resume)      → lands the observation in the register its fd table maps the cut-ID to, re-drives.
```

`frontier / ORIENT` = the possibility's set of open fds — derived. **Reconciles v9:** `owners`
(cut-ID → possibility) becomes the kernel socket table; `InFlight` (cut-ID → register) stays the
possibility's fd table. (This CORRECTS an earlier "one merged barrier block" idea — it's the fd/socket
split, not a merge.)

## 5 · Why it fits — CONVERGING

```text
NEARLY-STATELESS CTU   holds handles, not crossing state — the state lives in the kernel, like sockets.
KERNEL BARRIER         ingress/egress are the kernel operating over its own socket table.
OWNERSHIP MODEL        the kernel owns the separation handlers AND the crossing state (../ownership.md).
```

---

## 6 · Request vs Response — the block is flat and client-side — CONVERGING

The block holds only cut-IDs **this** CTU issued (client-side correlation, like an RPC client's pending
table). That is why it does not split by source:

```text
we ISSUE a cut (egress)       CLIENT → PARK our cut-ID/fd in our block.
a REQUEST arrives (ingress)   SERVER → nothing to pop; new resident work. the cut-ID belongs to the
                                     ISSUER's block — we only echo it back on our reply.
a RESPONSE arrives (ingress)  CLIENT → POP our block by cut-ID.
```

The **Header** decides, not the source; local vs remote is irrelevant to the block.

### The return signature is the participation switch — both ends

```text
EGRESS            return signature present → PARK (expect a Response).
                  fire-and-forget          → send, DON'T park. no entry, ever.
INGRESS Response  → POP our block (only a return-expecting cut we parked can land here).
INGRESS Request   return signature present → process, then send a Response (pops the ISSUER's block).
                  fire-and-forget          → process, send nothing.
```

So the block holds exactly the cuts that **expect a return**. Whether a move expects a return is
**STRUCTURAL — THINK's half** of the signature; the return TYPE is negotiated at ACT (see
`think-act-signature.md`).

### Two corners

```text
UNKNOWN RESPONSE    a Response whose cut-ID has no entry (late / duplicate / post-timeout) → DROP it
                    (the closed-connection analogue). ingress needs a "no such fd → ignore" path.
SERVER THAT SPAWNS  processing an incoming Request can make us issue OUR OWN sub-cuts → those DO park
                    (we are client for them). rule stays: egress-with-return-signature parks, nothing else.
```

---

## 7 · Ingress without an fd — the server case — CONVERGING

A Response pops an fd (we were the client). A Request has **no fd** — that is the DUAL, not a gap:

```text
fd (our block)   client-side backward pointer — where a RESPONSE lands in OUR graph.
return route     server-side backward pointer — where OUR result goes back OUT.
  (on the new     = { issuer address, issuer's cut-ID }, both arrive IN the Request.
   possibility)
```

A Request with no fd creates a **new resident possibility** carrying a **return route** (reply-to =
issuer + their cut-ID). It *starts* a graph here (a root on our CTU, a child of the remote possibility).
On finish, egress stamps the issuer's cut-ID on the Response and pops the **issuer's** block. So:
**cut → fd, request → return route.**

### Inbound authorization is by IDENTITY, not solicitation

The fd proves *solicitation* — but only for Responses. For Requests the authorizing thing is **who it
is from**, so there are three inbound cases:

```text
RESPONSE (has fd)          authorized by SOLICITATION — lands on its fd.
REQUEST from the PRINCIPAL authorized by IDENTITY — SELF-AUTHORIZING. the user is the principal (the
  (the user)               authority SOURCE, not a peer entity); its request SEEDS the root horizon.
                           the barrier's check for it is AUTHENTICATION, not policy.
REQUEST from a FOREIGN     authorized by POLICY — the permission check, gated by the human's policy.
  entity
```

The user is local and unsolicited but **not foreign** — foreign is about identity, not locality.
Treating the principal as a peer entity was the mislabel that made its own request look "unauthorized."

**SCOPE — DECIDED:** both checks (authenticate the principal; policy-gate a foreigner) are **stubbed for
now** — API-authorization / security, a bridge to cross later. What this fixes: inbound is
identity-gated at the barrier under the human's authority, and the principal's request is the bootstrap
(the root THINK seeds the horizon). Internals deferred. See `capability-table-and-horizon.md`.

---

## Still moving — OPEN

```text
TIMEOUTS         live in the kernel socket table (it holds the pending state), like TCP timers are the
                 kernel's, not the process's. mechanism TBD (age/expire → Failure/NotYet to parked work).
HANDLE SHAPE     the exact cut-ID / fd representation — an fd-like small int, or the MoveID as-is?
MOVE SIGNATURE   stay `Move(Transfer)` with Header-dispatch, or split? leaning one method.
D5 vs BARRIER    is the trap literally an egress crossing whose target is the kernel, or a distinct axis
                 that shares the barrier? keep coherent with D5.
LOCAL vs CROSS   local CUTS park client-side (§6); RunHere sync dispatch does not. confirm the boundary.
INBOUND ADMISSION  §7 fixes that unsolicited inbound is gated at the barrier under human policy, and
                 STUBS the check (API-authorization / security — later). the internals stay deferred.
```
