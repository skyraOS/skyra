# The discovery pipeline — socket ▸ getaddrinfo ▸ connect

**Date:** 2026-08-24
**Status: RESEARCH — the OS abstraction THINK's discovery plants on.** Contracts pulled from this
machine's headers (macOS SDK: `sys/socket.h`, `netdb.h`). The interest is the **contracts** — the
top-level param contract and the return of each stage — because that is what maps onto THINK vs the
executor. Grounded, not from memory.

---

## The three stages — and each is one of our actors

```text
socket(domain, type, protocol)   nameless capability: "what KIND of reach"        THINK proposes
getaddrinfo(name, service, …)    RESOLVER: name → bounded candidate endpoints     ← THE DISCOVERY FLAG
connect(fd, address, len)        TEST / bind one candidate; reachability learned  executor, at dispatch
```

---

## Stage 1 · `socket` — select a nameless capability

```c
int socket(int domain, int type, int protocol);
```

```text
PARAM CONTRACT   (domain, type, protocol) — three ints. A KEY into the protosw registry, not free
                 input: only registered+permitted tuples resolve (else EAFNOSUPPORT / EPROTONOSUPPORT
                 / EPROTOTYPE / EPERM).
RETURN           int — an fd (≥ 0), or -1 + errno. A handle onto a TYPED, NAMELESS endpoint
                 (getsockname == 0.0.0.0:0). The triple binds the PROTOCOL, not the LOCATION.
SKYRA            THINK proposing a capability: (capability selector) → a nameless capability reference.
                 A name miss here is the FAULT signal; permission is the separate TRAP axis.
```

## Stage 2 · `getaddrinfo` — the resolver (DISCOVERY)

```c
int getaddrinfo(const char *hostname, const char *servname,
                const struct addrinfo *hints, struct addrinfo **res);
```

```text
PARAM CONTRACT   (hostname, servname, hints, &res)
                   hostname  — a NAME (human-meaningful: "example.com", or an identity to resolve)
                   servname  — the service ("https", "443")
                   hints     — a PARTIAL addrinfo constraining the search (a partial triple + flags)
                   res       — OUT param: head of a linked list of results
RETURN           int status — 0, or an EAI_* code (gai_strerror). The real payload is *res.
                 *res : struct addrinfo*  — a BOUNDED SET (linked list) of candidates:

     struct addrinfo {
         int        ai_flags;
         int        ai_family;      // ┐
         int        ai_socktype;    // ├─ the TRIPLE  (feeds socket())
         int        ai_protocol;    // ┘
         socklen_t  ai_addrlen;
         char      *ai_canonname;   //    the resolved canonical NAME
         struct sockaddr *ai_addr;  //    the ADDRESS  (feeds connect())
         struct addrinfo *ai_next;  //    → next candidate
     };

SKYRA            DISCOVERY = name → bounded candidate endpoints. TWO HORIZONS of the same op:
                   STATIC (THINK)    capability surface + memory → nameless candidate categories.
                                     read-only, NO network.
                   LIVE  (executor)  DNS / directory / search → concrete candidate addresses.
                                     network-using, at execution.
```

## Stage 3 · `connect` — test and bind one candidate

```c
int connect(int socket, const struct sockaddr *address, socklen_t address_len);
```

```text
PARAM CONTRACT   (fd, address, len) — the nameless endpoint (stage 1) + ONE chosen address picked
                 from stage 2's candidate set.
RETURN           int — 0 (reachability LEARNED: the endpoint answered), or -1 + errno
                 (ECONNREFUSED / ETIMEDOUT / ENETUNREACH …), or EINPROGRESS on a non-blocking fd.
SKYRA            executor at dispatch: (capability ref, resolved dependency) → outcome. Maps onto
                 move.md's three: 0 = SUCCESS · error = FAILURE · EINPROGRESS = NOT YET.
```

---

## The join — why `struct addrinfo` IS a potential move

`getaddrinfo` returns each candidate as **the triple AND the address in one record** — `ai_family/
ai_socktype/ai_protocol` (what feeds `socket`) beside `ai_addr` (what feeds `connect`). So a single
`addrinfo` is exactly a `(capability, location)` pair, and `ai_next` chains them into a **bounded
possibility space**. That is the shape of a potential move; the linked list is the bounded candidate
set THINK is meant to produce.

```text
              ┌────────────── struct addrinfo (one candidate = one potential move) ──────────────┐
name ──gai──▶ │ (ai_family, ai_socktype, ai_protocol)  +  ai_addr  +  ai_canonname               │ ──▶ ai_next ──▶ …
              └──────────────┬──────────────────────────────────┬──────────────────────────────┘
                             │ the TRIPLE                        │ the ADDRESS
                             ▼                                   ▼
                        socket(...)  ── fd ──────────────────▶ connect(fd, ai_addr, ai_addrlen)
```

## The contract chain, end to end

```text
(domain,type,protocol)  ─socket→   fd            (nameless capability handle)      THINK proposes
(name, service, hints)  ─gai→      [addrinfo…]   (bounded (triple+address) set)    RESOLVER / discovery
(fd, one address)       ─connect→  0 | err | EINPROGRESS   (SUCCESS|FAILURE|NOTYET)  executor, dispatch
```

## To work through next

```text
WHICH CONTRACT IS THINK'S?     THINK holds the getaddrinfo SHAPE at the static horizon: (name + hints)
                               → bounded candidates, but with ai_addr left as a VERDICT/name, not a
                               live-resolved sockaddr. does THINK emit addrinfo-shaped potentials?
WHERE DO HINTS COME FROM?      `hints` is a partial triple that BOUNDS the search. is that the request
                               + capability surface? is it the horizon?
WHO OWNS THE LIVE RESOLVE?     getaddrinfo-over-the-network is itself a capability (a resolve MOVE) the
                               executor runs — subject to admission/history like any endpoint. confirm.
NOVELTY GATE                   a resolved-to address never seen before is a novel endpoint; connect to
                               it crosses the MINT/paving gate under user tolerance. wire this in.
```
