# The reach abstraction

**Date:** 2026-08-13
**Status: REFERENCE.** What `reach` can inherit from the operating system, what it cannot, and what
this design has to carry itself because the OS drops it.

> **Address in, file descriptor out. The topology is the kernel's problem, below the call.**

---

## Why this document exists

`Where E9` recorded that LAN and internet are the same mechanism, measured on one machine. That
looked like a macOS detail. It is not — it is the most portable abstraction in networking, and
knowing *why* it is portable tells you exactly what it costs.

---

## 1 · The abstraction

```c
int fd = socket(AF_INET, SOCK_STREAM, 0);   // SYS_socket   97
connect(fd, addr, addrlen);                 // SYS_connect  98
read(fd, …) / write(fd, …) / close(fd);     //          3 / 4 / 6
```

You supply an **address**. You receive a **file descriptor**. Nothing in that exchange mentions
distance, medium, or route.

---

## 2 · Where it is strong

### It is the same call at every distance ✅

```console
LAN gateway  socket()+connect() -> OSError
internet     socket()+connect() -> fd 4  peer 160.79.104.10
```

Identical calls, identical code. Only the bytes in the `sockaddr` differ. **The routing decision —
one hop across your kitchen, or fourteen hops through three autonomous systems — is made in the
kernel's routing table, and userspace never sees it.**

### It is the same call on operating systems that share nothing else ✅

Berkeley sockets (1983) won so completely that Windows ported the API rather than inventing one.
Go's `net` package is the check that can be run locally — a mature runtime that must work on both:

```text
224   .go files in net/
 13   *_windows.go
 13   *_unix.go            ← ~88% of the package is shared
```

```go
// net/hook_unix.go
connectFunc func(int, syscall.Sockaddr) error            = syscall.Connect
// net/hook_windows.go
connectFunc func(syscall.Handle, syscall.Sockaddr) error = syscall.Connect
```

**The same function name on both.** The platform files exist to bridge one type difference — an
`int` fd against a `syscall.Handle` — not a difference in model.

And `dialSingle` branches on **address type, not platform**:

```go
switch ra := ra.(type) {
case *TCPAddr:  …
case *UnixAddr: …
}
```

```text
THE ABSTRACTION IS ADDRESS-SHAPED, NOT PLATFORM-SHAPED.
```

### It converges with local reach ✅

```text
LOCAL     pipe(42) → fork(2) → execve(59)     →  fd
REMOTE    socket(97) → connect(98)            →  fd
                                                 ↓
                        read(3) · write(4) · close(6)    ← IDENTICAL on both
```

So *reach an entity on this machine* and *reach an entity in another state* end in the same object,
operated by the same three syscall numbers.

*Every ✅ above except the Windows/Winsock history was measured on one machine, 2026-08-12/13.*

---

## 3 · What it hides — and hiding it is the strength

```text
MEDIUM      ethernet · wifi · cellular · a VPN tunnel      never visible
ROUTE       one hop or fourteen                            decided below the call
TOPOLOGY    who is adjacent to whom                        the routing table's business
```

**This is why "wifi" was never going to appear in this runtime.** It is three layers below the only
thing the API exposes, and a design that mentioned it would be reaching past an abstraction that
exists precisely so nothing has to.

```text
CONSEQUENCE FOR REACH   the resolver picks an ADDRESS, not a ROUTE. Routing is not the resolver's
                        job and cannot be — the abstraction that makes reach portable is the same
                        one that hides the path.
```

---

## 4 · What it does not hide, and never could

Waldo et al., *A Note on Distributed Computing* (1994): the attempt to make remote calls look like
local ones fails irreducibly, on four things.

```text
LATENCY             connect() to 10.0.0.1 and to 160.79.104.10 are the same call at very
                    different speeds. The API reports neither.
PARTIAL FAILURE     one side can vanish mid-exchange. A local call cannot.
MEMORY ACCESS       no shared address space across the boundary.
CONCURRENCY         the far side runs whether or not you do.
```

**The API is silent on all four.** It gives you one call for any distance and tells you nothing
about which distance you got.

```text
THE API IS SHARED         socket + connect. one call, any distance, every OS.
THE PROPERTIES ARE NOT    10.0.0.1 and 160.79.104.10 fail differently, at different speeds, in
                          different failure domains — and connect() will not say which.
```

---

## 5 · What the design has to carry

Every item below exists because the abstraction drops something. This is the payload of the
document.

```text
WHERE IS TWO QUESTIONS        the API gives one call for "where am I stored" and "where is my
Where · decision 3            entity". They are independent — the walkthrough had the runtime on a
                              laptop, the model remote, the file on a server — so the design must
                              carry a distinction the API does not.

CRITERION 2 IS FAILURE        an address does not tell you whether the thing behind it can fail
DOMAIN, NOT ADDRESS           without taking you with it. Everything has a name; only some things
Who · decision 2              have their own crash. See Who E3.

THE RESOLVER PICKS AN         and learns from failure, because failure is the only thing the API
ADDRESS, AND LEARNS FROM      does report. A fallback event is the observation.
FAILURE                       Where · decisions 5 and 10

REACH FAILS IMMEDIATELY       connect() blocks and then errors; it does not tell you in advance.
AND LOUDLY                    So the design places reachability BEFORE the possibility opens, and
Where E10                     requires the failure to be loud — the same rule world.md already
                              wrote for projection.

FAILURE DOMAINS ARE NESTED    thread · process · machine · network. mongod is independent of your
Who · decision 4              process and not of your machine. The API exposes none of this, so
                              co-location is learned from CORRELATED FAILURE instead.

THE HANDLE IS ONE TYPE        execve and connect both end in an fd, so a design treating local and
How · decision 4              remote handles as two things draws a line the OS does not.
                              ⚠️ but an fd is PROCESS-LOCAL and does not travel — the convergence
                              is real and not portable.
```

---

## 6 · Where the abstraction stops entirely

```text
kernel      socket + connect      a byte pipe to an address
userspace   TLS                   a program interpreting those bytes
userspace   HTTP/2                a program interpreting those
userspace   the API contract      a program interpreting those
```

**There is no TLS syscall** — 461 numbers in Darwin's table, zero matches. Everything above the byte
pipe is a program, which is also the line above which things can be entities at all.

---

## The shape of the lesson

```text
A GOOD ABSTRACTION HIDES ONE THING COMPLETELY AND IS HONEST THAT IT HID NOTHING ELSE.
```

Sockets hide topology totally — that is what makes one call work at every distance on every OS. They
hide latency and failure not at all, and never claimed to. The failure mode is not the abstraction;
it is a design that reads *"one call for any distance"* as *"distance does not matter."*

Which is the same finding as 2026-08-12's other one, from the opposite side: **knowing where you are
is fundamental in a distributed system.** The API drops it. Something has to carry it. That is
sections 3 and 5 of this document.
