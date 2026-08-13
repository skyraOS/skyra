# OS parallels — a research handle

**Status: REFERENCE, not decisions.** Decisions live in
[`syscall-read-example/outcomes/`](syscall-read-example/outcomes/); measurements live in
[`syscall-read-example/evidence/`](syscall-read-example/evidence/). This file is the map that made
those easier to find.

## The core mapping

Everything below elaborates this. An engram is not the binary — it is the binary plus what running
it wrote.

```text
THE BINARY          →   Expectation        IMMUTABLE. "authored once and re-run."
PROCESS MEMORY      →   the frame          EPHEMERAL. dies with the opening.
WHAT IT WRITES      →   History            DURABLE. accumulates across openings.
                        Observations
```

> `interfaces.md` — *"Expectations never improve — they are authored once and re-run, and
> everything that accumulates accumulates **AROUND** them."*

It is also a **check on where state may live**: immutable, ephemeral, or written. State in none of
those three has nowhere legitimate to be — which is entity criterion 5 (*no back-channel*) arriving
from the file side.

---

Every parallel below is marked:

```text
✅ VERIFIED    measured on a machine, or quoted from both sides
🔶 PROPOSED    the shape lines up; nothing has tested it
⚠️  WHERE IT    every analogy has one. Stated, because an analogy without its limit
   BREAKS      is how you inherit someone else's mistake.
```

---

## 1 · engram : possibility :: file : process ✅

```text
ENGRAM        durable.     an expectation with a history accumulating around it.   terminology.md
POSSIBILITY   ephemeral.   ONE OPENING OF an engram.
```

`file` / `process` uses the same two words — durable versus ephemeral, and *one opening of*. Written
into `terminology.md` before the analogy was looked for.

**The mapping is three-way** — see [The core mapping](#the-core-mapping) above. A running process
reading a log another process just appended is the same relationship a possibility has to history
written since it opened. **Nothing about that is a break.**

**An engram is a DIRECTORY, and its path is its name.**

```text
mynamespace/engramA/engramB          ← the path IS how engramB is reached
    ├── the expectation              the "binary"
    ├── the history                  what running it wrote
    └── engramC/                     a child engram
```

```go
// ID is what it IS; the name is how it is REACHED. An engram does not hold
// its name — that is the parent's map key.
//
// An engram lives inside exactly ONE parent's map, so two parents
// structurally cannot share one.
```

Single parent means path ↔ engram is **bijective — no hard links** — which is also true of Unix
directories. And a directory holding a binary alongside its logs is ordinary; a macOS `.app` bundle
is exactly that, a directory you run.

```text
WHAT IT BUYS    everything the OS knows about the file/process split: one file, many concurrent
                processes; a process outliving the file it came from. Plus a check on where
                state may live — see The core mapping.

⚠️ WHERE IT     THE ONLY DIFFERENCE IS APPEND-ONLY. Verified: `Graph` exposes Mint · MintRoot ·
   BREAKS       Store · RecordDiscard · Apply · Disperse and NO delete; there is no `delete(`
                in graph or kernel outside tests. A filesystem mutates and unlinks; this one
                cannot.

                TWO EARLIER VERSIONS OF THIS SECTION CLAIMED A DIFFERENT BREAK AND BOTH WERE
                WRONG. First: that engrams are mutable where files are not — wrong, it mapped
                the whole engram onto the binary when only the expectation is binary-shaped.
                Second: that Unix keeps the binary and its log as two objects — wrong,
                directories bundle, which is what an engram is.

                The pattern is worth more than either correction: this file asks every parallel
                to state where it breaks, and the pressure to produce one manufactured a break
                twice. AN ANALOGY THAT HOLDS IS A FINDING TOO.
```

## 2 · The OS has TWO wrappers, and so does this ✅

```text
THE FILE FORMAT   Mach-O · ELF · PE           what the LOADER parses. durable, on disk.
                  ↳ the engram / observation shape

THE ABI           argv/envp on the stack,     how you talk ONCE RUNNING. live, per-invocation.
                  calling convention          ↳ the frame + the emission grammar
```

`execve` enforces the first; the architecture enforces the second. **A program satisfying one and
not the other still fails — just later and worse.**

```text
WHAT IT BUYS    it separates two questions that keep getting merged: what must be true of a
                stored engram, versus what must be true of a live exchange. They have different
                validators and different failure times.
```

## 3 · You emit a claim; the parse decides ✅

```text
a compiler emits BYTES it claims are Mach-O    →  execve parses.   bad magic → ENOEXEC, no process
an entity emits TEXT it claims is an emission  →  grammar.Parse.   unparseable → no possibility
```

```go
type Emission interface{ isEmission() }   // unexported: nothing outside grammar constructs one
```

> `grammar.md` — *"the incapacity is a **parsing** property, not an API property."*

**The producer never constructs the object.** If a compiler could hand the kernel a filled-in
process image instead of bytes, it could hand it a malformed one and the type system would vouch
for it. The unexported method is the same defence.

```text
WHAT IT BUYS    settles "who wraps the expectation" — neither the kernel on your behalf nor the
                entity for itself. The parse is the boundary, and it is the only place a third
                form could be refused.
```

## 4 · Reaching anything is ADDRESS + TRANSPORT ✅

```text
                ADDRESS                                  TRANSPORT
curl            /usr/bin/curl                            fork(2) + execve(59)
Claude          api.anthropic.com:443                    socket(97) + connect(98)
mongod          /tmp/mongodb-27017.sock                  socket(97) + connect(98)
libssl          /opt/homebrew/…/libssl.3.dylib           open(5) + mmap(197)
```

**A name never tells you how to reach it — the namespace does.** Rows 1 and 3 share the filesystem
namespace and resolve to different transports.

*Measured. See `Where E9`.*

## 5 · `execve` and `connect` converge on one object ✅

```text
LOCAL     pipe(42) → fork(2) → execve(59)     →  fd
REMOTE    socket(97) → connect(98)            →  fd
                                                 ↓
                        read(3) · write(4) · close(6)    ← IDENTICAL on both
```

**The transports differ; the handle does not.**

```text
WHAT IT BUYS    "everything is a file descriptor" is why local and remote reach converge. A design
                treating a local handle and a remote handle as two types draws a line the OS
                doesn't — which bears on `host.Handle`, currently shaped only around spawn.

⚠️ WHERE IT     an fd is process-local. It does not survive being sent to another machine, and
   BREAKS       passing one even between local processes needs SCM_RIGHTS — which Go excludes on
                Windows. The convergence is real and NOT portable.
```

## 6 · The kernel gives a byte pipe; everything above is programs ✅

```text
kernel      socket + connect      a byte pipe to an address
userspace   TLS                   a program interpreting those bytes
userspace   HTTP/2                a program interpreting those
userspace   the API contract      a program interpreting those
```

**There is no TLS syscall** — 461 numbers in Darwin's table, zero matches. The handshake, ciphers,
certificate chain and session keys are all userspace `read()`/`write()` on one fd.

```text
WHAT IT BUYS    the layer boundary is unambiguous and measurable: if it is not in the syscall
                table, it is a program, and programs are the things that can be entities.
```

## 7 · A program is a PATH, not a command line ✅

```text
✓  a native binary                       execve'd directly
✓  a #! script                           the KERNEL reads the first two bytes and execs the
                                         interpreter named there. no shell involved.
✗  'cd'                                  a shell BUILTIN. not a file. → No such file or directory
✗  '/bin/ls | grep x'                    a shell EXPRESSION. the pipe is syntax, not a path.
```

*Demonstrated with raw `fork` + `execv`. See `How E9`.*

```text
WHAT IT BUYS    "the runtime needs bash" is false, and so is "the runtime needs curl". Both are
                conveniences above the only essential layer: route · resolution · transport ·
                credential.
```

## 8 · A path and a network address are the same object: a NAME IN A NAMESPACE ✅

```text
file:///usr/bin/curl
https://api.anthropic.com/v1/messages
unix:///tmp/mongodb-27017.sock
```

**The scheme IS the resolution procedure.** Same shape the design already uses for authority —
`terminology.md` makes NAMESPACE a first-class durable object, and capability is name resolution.

```text
WHAT IT BUYS    naming is UNIVERSAL, so it discriminates nothing — which is why entity criterion 2
                could not be "has an address". See Who E3.
```

## 9 · Independent existence = an independent failure domain ✅

```text
before dlopen:  pid 19388
after  dlopen:  pid 19388                 ← same process. you are simply LARGER.
after  fork:    pid 19388 spawned 19389   ← a second process.
```

```text
execve · connect   →  a HANDLE to a party that is not you
dlopen             →  a POINTER, which is an offset INTO you
```

The test: **can it fail without taking you with it?**

```text
⚠️ WHERE IT     failure domains are NESTED — thread · process · machine · network. mongod is
   BREAKS       independent of your process and not of your machine. And a device with no MMU has
                no local isolation at all, so on such a device there are NO local entities.
```

## 10 · Identity and capability are different reads, for a structural reason ✅

```text
IDENTITY     must be STABLE     so history can key on it       IOPlatformSerialNumber
CAPABILITY   must be CURRENT    so it reflects a changing box  hw.machine · kern.osrelease
```

**Opposite requirements.** One value cannot do both: `brew install psql` must change the second and
must not change the first. Which is why every OS separates the reads, and why the inert factory
serial is the right identity precisely because it is useless for capability.

## 11 · Identifiers come in tiers, and the middle one may not exist ✅

```text
PER BOOT      kern.bootsessionuuid    v4 RANDOM — fresh every boot
PER MACHINE   IOPlatformSerialNumber  factory-issued, pre-dates the OS
              IOPlatformUUID          v5 SHA1 name-based — deterministic
              kern.uuid               v3 MD5 name-based — deterministic
```

The UUID **version nibble** is the evidence, not recall: a v4 is random, a v3/v5 is a hash and
cannot change unless its input does.

```text
macOS HAS NO PER-INSTALL TIER.   you get this boot or this hardware, nothing between.
                                 Linux has /etc/machine-id for the middle; there is no equivalent.
```

## 12 · An entity is an interpreter, and interpreters compose ✅

```text
runtime ──execve──▶ curl ──HTTPS──▶ Claude
                    entity          entity
```

`curl 8.7.1` passes all five entity criteria. So reaching Claude went **through** an entity — which
makes "endpoint" recursive rather than special.

```text
WHAT IT PREDICTS    exec.CommandContext(ctx, "codex", args...) is ONE EntityID covering TWO
                    entities — the CLI binary and the model behind it. If codex starts failing,
                    nothing in the graph can say which regressed.

⚠️ THE COST         composition puts the middle entity in the TRUST PATH. curl can misreport what
                    Claude said, and the history blames the outer name either way.
```

## 13 · Leaf-ness is structural, like an interpreter that cannot branch 🔶

An entity whose interpreter can only ever emit `COMMIT` grows no children — and **nothing has to
say so.** No `is_leaf` flag, no capability declaration. `mongod` is a leaf because a query result
has no path by which it becomes a `SELECT`.

```text
REFINEMENT   leaf-ness belongs to the INVOCATION, not the engram. The same expectation interpreted
             by Claude can branch and by mongod cannot — so an engram is not inherently a leaf, an
             OPENING is.
```

## 14 · `content.EngramRef` is a symlink 🔶

An engram lives in exactly one parent's map — that is the tree. But `content.EngramRef{ID, Name}`
references an engram *without* being a parent-child edge:

```text
THE PARENT'S MAP     the tree. one parent, one path, no hard links.
content.EngramRef    a reference ACROSS the tree, carried as a value.
```

Which is what a symlink is: a name that resolves elsewhere, distinct from the directory entry that
owns the target.

```text
WHY IT MATTERS   this is the type that carries REACH. `interfaces.md` separates delegation from
                 injection on exactly this — reach must not travel in the data channel, and only
                 the type keeps them apart. A symlink handed to something that cannot follow it
                 is an inert string; handed to something that can, it is authority.

⚠️ UNTESTED      the shape lines up. Whether EngramRef inherits the problems symlinks have —
                 dangling targets, resolution loops, TOCTOU races — has not been checked.
```

## 15 · Executable versus data is a MODE BIT, not a type 🔶

```text
ENGRAM        an EXECUTABLE file    carries an expression. can be opened and run.
OBSERVATION   a DATA file           written by one invocation, read as input by another.
```

But Unix answers this with a permission bit rather than a type — **a `#!` script is data marked
executable**, and the kernel resolves it (parallel 7).

```text
WHAT IT PREDICTS   if both are files, "can an observation become an engram" is the same question,
                   and `discover` plus `content.EngramRef` are where it would already be
                   happening. UNTESTED — this is the analogy's next claim, not a finding.
```

---

## Where the parallels came from

All of parallels 4–11 were measured on one MacBook Pro on 2026-08-12, in the course of tracing a
single question — *where does `read-local(file.txt)` actually run* — which is
[`syscall-read-example/example.md`](syscall-read-example/example.md).

```text
THE METHOD WAS THE POINT   every one of these came from running the thing rather than recalling
                           it. Three claims that survived recall did NOT survive measurement:
                           that a library has no address, that LAN and internet are different
                           mechanisms, and that arch changes what a syscall DOES rather than how
                           it is reached.
```
