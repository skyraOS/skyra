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
those three has nowhere legitimate to be — which is the *no back-channel* entity-boundary invariant
arriving from the file side. It is separate from entity criterion 5, which places continuity through
the history accumulated with the entity's observer/parent.

---

Every parallel below is marked:

```text
✅ VERIFIED    measured on a machine, or quoted from both sides
🔶 PROPOSED    the shape lines up; nothing has tested it
⚠️  WHERE IT    stated, because an analogy without its limit is how you inherit someone
   BREAKS      else's mistake.
```

> **Read every ⚠️ block with this in mind: AN ANALOGY THAT HOLDS IS A FINDING TOO.**
>
> Asking each parallel where it breaks creates pressure to produce one, and that pressure
> manufactured a false break **twice on parallel 1 alone** — first claiming engrams are mutable
> where files are not, then claiming Unix keeps a binary and its log apart. Both were the template
> being satisfied rather than the thing being reported. Both retractions are left in place.
>
> Where a section says the analogy holds, that is a result, not an omission.

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

                Both are left in place. The pattern they set is in the legend at the top of
                this file, because it governs how every ⚠️ block here should be read.
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

⚠️ THE SUBJECT   `../structs and interfaces/aug-14-int-structs/` finding 1: PROVENANCE TYPING
   MAY NOT       DISSOLVES EngramRef. If a candidate is an ordinary observation with a producer,
   SURVIVE       there is no distinct type for this parallel to be about.

                 AND THE FOLLOW-ON IS SHARPER THAN THE ONE ABOVE. EngramRef was deliberately not
                 a Ref — "it never enters Cites, so reach carries NO CREDIT and the walk cannot
                 descend through it." Under provenance typing the walk CAN descend through a
                 thinking behavior, which is open-questions.md HOLE H13.

                 This parallel asked whether the symlink analogy imports symlink PROBLEMS. H13
                 asks whether dissolving the type reopens a credit path the type was shaped to
                 close. The second matters more, and it arrived from a direction this section
                 did not anticipate. LEFT STANDING, CONTINGENT — if H13 dissolves the type, this
                 parallel is about a thing that no longer exists.
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

## 16 · A solidified continuation is an eBPF program, and admission is a VERIFIER 🔶

```text
eBPF PROGRAM            bytecode authored in userspace, submitted to run IN the kernel.
THE VERIFIER            runs at LOAD. proves termination and memory safety, or REFUSES.
ONCE LOADED             runs many times, unverified, at full speed.

COMPILED CONTINUATION   behavior authored by an entity, submitted to run IN the kernel loop.
ADMISSION               — does not exist —
ONCE SOLIDIFIED         runs many times with NO CROSSING.
```

The structural match is the lifecycle, not the language: **an artifact authored by an untrusted
party, admitted once into a privileged execution context, then run repeatedly without further
checks.** Admission is paid for once; the guarantee it establishes has to hold for every run after.

This is parallel 3 extended in time. There, an entity emits a claim and the parse decides — a
check on **form**, per emission. A verifier is the same boundary asking about **behavior over
time**, on an artifact that will outlive the emission that produced it.

The historical trajectory is the useful part. eBPF originally rejected all backward jumps: no loops
at all, programs had to be DAGs — a total language by construction. Bounded loops were admitted
later (recalled as 5.3), once the verifier could *prove* termination. It never moved to admitting
arbitrary loops.

```text
WHAT IT PREDICTS   the question [`../continuation-language/conditions-and-loops.md`](../continuation-language/conditions-and-loops.md)
                   leaves open — whether the kernel should admit a continuation it cannot prove
                   terminates — has been answered once already by a system with the same shape,
                   and the answer was BOUNDED, NOT REFUSED and NOT UNRESTRICTED.

                   It also predicts where the pressure lands. `while` + a content condition + an
                   accumulator carried between iterations is Turing complete, and recursion is
                   already refused structurally ("REACH MAY NOT REACH UPWARD"). So iteration is
                   the only door, which is exactly the door eBPF spent a decade narrowing.

⚠️ WHERE IT        THE VERIFIER'S CLOSURE DOES NOT EXIST HERE. eBPF verifies against a fixed
   BREAKS          instruction set and a whitelist of helpers — it can see everything the program
                   can do. A continuation invokes child positions whose occupants are chosen at
                   runtime by Think and whose behavior is authored elsewhere. Proving the loop
                   skeleton terminates says nothing about a child that never returns. Verify-once
                   is only true of the SKELETON; the occupants change every run.

⚠️ AND THE STAKES  eBPF rejects aggressively because the downside is a kernel panic. Here the
   ARE NOT THE     downside is a wasted possibility, and eviction already handles that —
   SAME            "a RUNAWAY is depth or breadth past a bound, and it is an EVICTION decision."
                   The analogy argues FOR a verifier; the difference in blast radius argues that
                   the existing operational answer may already be proportionate. Both are stated
                   because picking the mechanism before the stakes is how this gets over-built.

⚠️ RECALLED, NOT   every eBPF claim above is from recall. Nothing here was measured, and this
   MEASURED        machine cannot measure it — Darwin has no eBPF. TO VERIFY: load a program with
                   an unbounded backward jump via bpf(2) BPF_PROG_LOAD on Linux and read the
                   verifier log. Flagged loudly because this file's method is the opposite, and
                   three recalled claims elsewhere in it did not survive measurement.
```

## 17 · The frame is a NAMESPACE, and the seven methods sort by their relation to it 🔶

The seven kernel protocol methods do not read as one list. Sorted against the OS, they fall into
three groups — and the sort is not by "does work" versus "does not."

```text
DECIDE · REACH · REMEMBER    THE CAUSAL SPINE.  calls that cross and cause.        → syscalls
APPROACH                     CONSTRUCTS the resolution context.                    → mount · unshare
THINK · ORIENT               READ THROUGH one. bounded, relative, never global.    → getdents · /proc
RECOGNIZE                    — DOES NOT SORT —  see the break, below.
```

**The namespace is not a method.** A Linux namespace is not something you call: it is a resolution
context that changes the answer to a call *without changing the call*. `open("/etc/passwd")` in two
mount namespaces returns two different files, and neither caller did anything differently. So
Approach/Think/Orient are not namespaces; they are the operations that create and read one.

The namespace itself already exists in the tree, under two names:

```text
frame.Frame     what ONE POSSIBILITY can name.        the per-opening view.
Session         all live work under one namespace.    the rack row. kernel.md.
```

```go
// Resolve turns a NAME into an engram. This is the whole of capability:
// you can select what you can name; you can name what you were given.
Resolve(name string) (*graph.Engram, bool)
```

That is `namei`. Two possibilities resolving the same name get different engrams, and it is correct
rather than a bug — which is the namespace property stated exactly.

Note this separates `Approach` from `Think`/`Orient`, which an earlier grouping bundled together as
"the non-causal ones." `mount(2)` and `getdents(2)` are not the same kind of thing.

### The bounding set arrives twice, at two scales

`Mount`'s one check is the Linux **capability bounding set**, essentially unchanged:

```text
CapBnd            can only SHRINK across execve. you cannot gain a capability you did not inherit.
grants ⊆ Ceiling  kernel.md. arithmetic, no semantics, NO LIVE ADMIN — which is what keeps the
                  rack flat, the same way uid 0 inside a user namespace buys nothing outside it.
```

Then the same rule reappears one layer down, in code that was written without reference to it:

```go
// kernel/runGroup — Reach narrows on the way down: a child gets its parent's
// reach, and a select may drop from it but can never mint into it.
```

Ceiling at admission, monotone narrowing at every descent. **One invariant, stated twice, at two
scales.** That is the strongest single hit in this parallel.

### `Orient` is `/proc` under a pid namespace

```text
kernel.md   Rooted() is NOT on Topology — "a global count with no node to start from," and
            introspection is bounded by the subtree, "which is not a rule but a fact about which
            way the pointers point."

pid ns      there is no view from inside that sees the whole machine. PID 1 is YOUR init.
```

Reached independently, before the analogy was looked for. Per this file's legend, that is a result.

```text
WHAT IT BUYS    isolation, monotone narrowing, relative-only views, and capability arithmetic —
                every one of which the design had already arrived at separately. The analogy
                supplies the vocabulary and the reason they belong together, not new mechanism.
```

### The eight kinds, each against its counterpart

Linux does not have *a* namespace. It has orthogonal KINDS, and a process holds one of each,
independently composable. If the mapping is real, the frame should decompose the same way.

```text
KIND     SINCE   VIRTUALIZES                            SKYRA COUNTERPART                  VERDICT
mnt      2.4.19  the mount table — what a PATH means    Engrams() · Resolve() · View       ✅ direct
uts      2.6.19  hostname + NIS domainname              ids.EntityID — READ-ONLY           ✅ setter cut
ipc      2.6.19  SysV IPC · POSIX message queues        — NOTHING TO ISOLATE —             finding
pid      2.6.24  the process ID table                   Orient · Topology · Head           ✅ view
                                                        ids.Minter                         ✗ numbering
net      2.6.24  interfaces · routes · ports            horizon — SPECIFIED, NOT BUILT     the WHERE hole
user     3.8     uid/gid + capability sets              Grant · Ceiling · Mount            ✅ direct
cgroup   4.6     your POSITION in the resource tree     Ancestry, hidden from possibility  ✅ direct
time     5.6     CLOCK_MONOTONIC / BOOTTIME offset      — UNNEEDED —                       finding
```

**`mnt`** is parallel 1 restated: an engram is a directory, its path is its name, single parent, no
hard links. There are **TWO mnt-shaped axes**, not one — `Resolve()` over engram names and
`Entities()` over who may be addressed, which `frame.Frame` itself calls *"scoped, exactly like
names."* Two independent resolution namespaces bundled into one object.

**`uts` is the setter, not the name.** `sethostname(2)` lets a process name itself, and that is what
`entity.go` rejects — *"a declaration is made by whoever REGISTERED the entity, not by the entity."*
Same reason an `Expectation` carries no language tag. The name survives; only the write is gone.
This is what lets WHO sit under Reach without becoming a self-report — see parallel 18.

**`ipc` has nothing to isolate.** Linux isolates SysV segments and message queues because two
unrelated processes can find each other through them. Here everything travels parent→child as
arguments and child→parent as returns, and the entity boundary forbids a back channel. No shared
object exists to be virtualized.

**`pid` splits, and the numbering half is inverted ON PURPOSE.** The view half matches exactly —
bounded by subtree, no inside view of the whole machine, which is why `Rooted()` was refused. But a
pid namespace *renumbers*: a process holds a different pid in each ancestor ns. `ids.go` refuses
that in advance:

```go
// Global uniqueness is load-bearing in a way that is easy to misread: Producer
// resolves across entity trees. The credit walk is bounded by CITES BEING EMPTY at
// a crossing, never by ID scope — so nobody should reach for scoped IDs to "fix" a
// boundary that is already held.
```

Linux scopes the identifier to enforce the boundary. This design holds the boundary structurally and
keeps identifiers global. **Same isolation, opposite mechanism** — and the comment names the Linux
answer as the misread to avoid, before the parallel existed.

**`cgroup` is not cgroups.** The controllers do the limiting; the namespace only hides where you sit,
so a container reads `/proc/self/cgroup` and sees itself at the root. The counterpart is `Ancestry`,
and the consequence is load-bearing rather than cosmetic:

```go
// Refused here rather than filtered in discovery, because discovery is a
// POSSIBILITY and possibilities cannot see their ancestry — only the kernel can.
```

Hide the position, and every operation depending on it moves up a level. That is the cgroup-ns shape
exactly.

**`time` exists for a reason this design does not have.** It offsets `CLOCK_MONOTONIC` and
`CLOCK_BOOTTIME` — never `REALTIME` — and it was built for checkpoint/restore, so a restored process
sees no jump. There is nothing here to restore: *"the kernel dies, everything in flight evaporates…
no reconciliation pass, no recovery protocol — ever."* What exists instead is logical, not wall
clock: [`../reach/age.md`](../reach/age.md)'s version coordinate `(owning parent, child position,
age)`, a pointer chain resolved at Reach and made durable at Remember. The only duration in the
surface is `Head.Since`, which exists solely to make STALLED detectable. Ordering, not clocks —
*"trajectory comes free because the log is ordered."*

```text
THE PATTERN ACROSS THE ABSENCES   uts, ipc and time virtualize something this design either
                                  refuses to read (a self-report), forbids (a back channel), or
                                  cannot need (restored state). ONLY net IS ABSENT BY OMISSION
                                  RATHER THAN BY DECISION — which is how to rank the hole.
```

Those axes are bundled into one `Frame` today. The prediction is that they are independent.

```text
IT ALSO DISAGREES WITH kernel.md ABOUT WHERE. That document wants an entity → hosts TABLE — a
registry fact, global, rebuildable from configuration. Namespace logic says reachability should be
SCOPED PER SESSION instead, the way netns makes "which network exists" a per-process answer rather
than a lookup. UNTESTED, and recorded because the two designs differ and drifting into one by
accident is how the choice gets made without being made.
```

```text
⚠️ SETNS IS THE     Linux namespaces come with a way IN: a privileged process calls setns(2) and
   ESCAPE HATCH,    its view EXPANDS. This design refuses that structurally — REACH MAY NOT REACH
   AND IT IS        UPWARD, enforced in the kernel because only the kernel can see Ancestry.
   REFUSED HERE     Taken wholesale, the analogy hands you an operation you must reject. It is the
                    inherited-mistake case the legend at the top of this file warns about.

⚠️ AND IT IMPORTS   RECOGNIZE HAS NO COUNTERPART. A mount namespace does not get better at
   NOTHING ABOUT    resolving names because a previous lookup succeeded. Namespaces scope; they
   LEARNING         never learn — they have no history, which is the one thing an engram is.

                    That is why Recognize does not sort. It writes into the structure Think reads,
                    which makes them the learn/lookup pair and makes Recognize the carrier of
                    "all learning is topological: selection learns."

                    READ AS A FINDING, NOT A GAP: the OS parallel covers the SCOPING half of the
                    kernel completely and the LEARNING half not at all. The seam between them is
                    exactly where the one unsortable method sits.

⚠️ RECALLED, NOT    every Linux claim above is recall — CapBnd's shrink-only rule, setns
   MEASURED         semantics, the namespace kind list. The Skyra side is quoted from the tree;
                    the OS side is not measured, and Darwin cannot measure it. TO VERIFY on Linux:
                    read CapBnd in /proc/self/status across an execve, and `capsh --print`;
                    `lsns` for the kind list; `unshare -U` for uid 0 that buys nothing outside.
                    Flagged because three recalled claims elsewhere in this file did not survive
                    measurement.
```

## 18 · `Reach` is `uts` + `net` fused at the interface — which places `Project`'s missing host 🔶

Parallel 17 lists uts and net as separate kinds because Linux keeps them separate. **This design
presents them as one method.** Reach takes a name and yields a contacted party; the decomposition
into WHO and WHERE is interior to it.

That is parallels 4 and 5 promoted from observation to method:

```text
parallel 4   reaching anything is ADDRESS + TRANSPORT. a name never tells you how to reach it.
parallel 5   execve and connect converge on ONE OBJECT. the transports differ; the handle does not.
             ↳ REACH is that convergence, given a name.
```

```text
NOT A DISCOVERY — THE TYPE SURFACE STATES IT ALREADY. `../structs and interfaces/aug-14-int-structs/`

P12                  ADDRESS AND TRANSPORT REMAIN BEHIND THE KERNEL BOUNDARY.
P9 · reach.md §3     Reach RESOLVES the entity and may never substitute it; address and transport
                     stay below. "An entity is an identity and does not carry its endpoints.
                     HOSTS HOLD ENDPOINTS."
kernel/05-reach.md   Realization{ Live · BeyondCurrentReach · BeyondYourHorizon } — the two
                     failures below are already a typed sum, not an inference from age.md.

WHAT THIS PARALLEL ADDS is not the boundary. It is WHY THERE ARE EXACTLY TWO FAILURES rather than
one, and where the unbuilt half attaches.
```

### The evidence: Reach's return type already decomposes

From [`../reach/age.md`](../reach/age.md), Reach's caller-facing conclusions:

```text
POSSIBILITY               live — and the version this possibility is acting through.
BEYOND CURRENT REACH      no version. nothing was reached.
BEYOND YOUR HORIZON       no version. nothing was reached.
```

**Two distinct failures that mean the same thing.** You would never distinguish them if one scope
were consulted. They are different because two are:

```text
BEYOND CURRENT REACH   the name does not resolve from here.        →  uts / mnt half.  WHO.
BEYOND YOUR HORIZON    it resolves, and you may not go that far.   →  net half.        WHERE.
```

The method is fused; its failure modes are not. That is the decomposition showing through the seam.

### Both halves are already named

`age.md` lists the four things a parent owns for each child position — and two of them are these:

```text
signature      what observations this position receives
return type    what observations this position hands back
entity         WHO interprets it                            ← the uts half
horizon        HOW FAR OUTWARD it may see or act            ← the net half
```

**`horizon` is a network namespace**: the scope of what is reachable, owned per position, versioned
when it changes. Parallel 17 marks net as the one absence by omission rather than decision — this
narrows that. It is not unnamed. It is specified in `age.md` and unbuilt in the tree.

### What it places

`interface changes/entity.md` records `Project` as blocked on something that does not exist:

```go
Project(r Rendering, h ids.HostID)   // "NOTHING SUPPLIES h."
```

Under this parallel, **Reach supplies it.** Resolution picks the entity, horizon bounds where it may
run, routing lands on a host, and `Project` is Reach's transport half crossing the entity boundary.

```text
REACH   resolve (WHO)  →  bound by horizon (WHERE)  →  Project (aim)  →  Step (cross)
```

`Registry.Open(id)` handing back an interpreter with zero host involvement is not the design — it is
the net half unbuilt. The three missing pieces in `kernel.md` (entity→hosts table · residency ·
Spawn as a place) all live *under* Reach on this reading, which is why none of them surface in the
seven-method protocol.

### Why the fusion is safe HERE

`age.md` forbids Reach from substituting *"another entity, expectation, or input."* So Reach picks
the route and never the destination — two scopes consulted, neither re-decided. That is exactly a
resolver's authority, and it is what keeps one method over two namespaces from becoming one decision
over two questions.

```text
✅ FUSED AT THE      ALREADY HELD, AND HELD BETTER THAN AS A PROHIBITION. An earlier draft of this
   INTERFACE IS      section warned that `salesdb@mac-A` bakes a host into an EntityID, and that
   FINE. FUSED IN    Invocation.Host being RECORDED, NOT KEYED would then be keyed through the
   THE IDENTIFIER    back door while the fold still reported it unkeyed. The type surface does not
   IS NOT.           forbid it — it gives the fact somewhere else to live:

                         P13                 local and remote are RELATIONS to runtime-supplied
                                             From, NEVER entity properties.
                         identity/02-from.md From{ Namespace, Requesting, Runtime }. no setter,
                                             no constructor. the runtime fills it.
                         reach.md §3         hosts hold endpoints; the entity does not.

                     "The identifier stays uts-only" names what is forbidden. `From` names where
                     the fact goes instead, which is why it carries a RuntimeID at all. The
                     prohibition is subsumed; recorded because the reasoning arrived separately
                     and agreed.

                     THE PRECEDENT STILL STANDS AS A COST, PARALLEL 12: exec.CommandContext(ctx,
                     "codex", …) is ONE EntityID covering TWO entities, and "if codex starts
                     failing, nothing in the graph can say which regressed." From fixes WHERE,
                     not composition — that name is still two things.

⚠️ AND THE NAME      parallel 4 is emphatic that a name never determines a transport — rows 1 and 3
   STILL DOES NOT    share the filesystem namespace and resolve to different transports. Read-only
   ROUTE            uts does not change that. The fusion is justified by CONVERGENCE (parallel 5),
                     not by the name becoming trustworthy enough to route on.

⚠️ UNTESTED          nothing here is built. Reach's signature is a proposal, horizon has no type,
                     and Project cannot run. This parallel argues WHERE the missing pieces attach;
                     it does not demonstrate that they work.
```

## 19 · A possibility is a STACK FRAME THAT DOES NOT POP — and escalation is unwinding 🔶

**Date added:** 2026-08-17. Came out of typing
[`../structs and interfaces/aug-14-int-structs/possibility/possibility.md`](../structs%20and%20interfaces/aug-14-int-structs/possibility/possibility.md),
not out of measurement. Nothing here was run.

```text
A FRAME         one call of a function.
A POSSIBILITY   "ONE OPENING OF an engram."   terminology.md, written before the analogy existed.
```

The definition is already frame-shaped, and so is most of the type:

```text
Parent                    the saved frame pointer. Ancestry-by-walking IS walking the stack, and
                          it is why Ancestry stopped needing to be a stored field.
Branch                    the activation record — args in, locals, return slot.
DOWN, NEVER UP            a frame cannot read its callers' locals. Same rule, same reason, and
                          in both systems it is a fact about which way pointers point.
EVICTION                  pop it and everything in it goes. v5's comment that eviction safety
                          "falls out rather than being built" is stack discipline, unnamed.
MaxDepth · MaxSteps       stack limits. kernel.md's "a RUNAWAY is depth or breadth past a bound,
                          and it is an EVICTION decision" is a stack overflow with a policy.
```

### The finding: the two repair directions are PUSH and POP

This is what the parallel buys, and it is not visible without it.
[`../structs and interfaces/aug-15-proposal/10-escalation-and-the-world-boundary.md`](../structs%20and%20interfaces/aug-15-proposal/10-escalation-and-the-world-boundary.md)
describes two directions and calls them two aspects of one mechanism:

```text
ZOOM        toward the world. what a parent does WHILE IT STILL HAS authority, material, budget.
            each descent strictly narrower.                                    ← PUSH A FRAME

ESCALATE    inward along causality, when the zoom EXHAUSTS. one position at a time,
            arriving at the next parent as a SPEC problem.                     ← POP A FRAME
```

**Escalation is unwinding**, and the correspondence is exact enough to be load-bearing:

```text
A HANDLER THAT CANNOT HANDLE      a parent whose repair exhausts.
RETHROW                           the failure becomes the next parent's.
THE FRAME WIDENS AS YOU UNWIND    "repair cost rises with every escalation, because the frame
                                   widens as it travels inward."
THE LAST HANDLER                  the user.
```

```text
WHAT IT PREDICTS   THE GIVE-UP PREDICATE IS A RETHROW CONDITION, and that is a thing language
                   runtimes have had to specify precisely rather than leave to judgement.
                   §10's Exhaustion — HorizonExhausted · BudgetExhausted · AttemptsExhausted ·
                   Unrepairable — is a catch clause list. Only the first is derivable today;
                   the design calls the others F1 and F4.

ALSO PREDICTS      A TAIL CALL HAS NO ANALOGUE YET. A possibility whose only act is to invoke
                   one child and return its result unchanged is a frame that need not be kept.
                   Whether that is worth collapsing — and whether collapsing it would destroy a
                   position that is supposed to accumulate history — is not asked anywhere.
                   UNTESTED.
```

### ⚠️ WHERE IT BREAKS — and this is the one that matters

```text
A STACK FRAME DIES ON RETURN. A POSSIBILITY DOES NOT.
```

> `../semantics/closure.md` — *"A possibility RETURNS when it has an observation. It CLOSES when
> the root commits. **Those were one word, and almost everything expensive came from that.**"*

Refinement-in-place *requires* surviving your own return: the parent revises, sends a new
expression into the same live possibility, and what that possibility sees is a sequence rather
than an event. An activation record that outlives its return and is re-entered with new arguments
is not a stack frame — it is a heap-allocated continuation.

**This is the `setns` case again.** Taken wholesale, the analogy hands you an operation you must
reject, and here that operation is *pop on return* — the exact defect `closure.md` exists to remove.

Three smaller breaks:

```text
A STACK IS LIFO        this is a tree and children run concurrently. terminology.md: "cardinality
                       > 1 means several heads at once, so the answer is a SET." A call stack has
                       one head; this has many, which is why Orient returns a set and Introspect's
                       counter was wrong.

FRAMES DO NOT LEARN    the ENGRAM accumulates history around the invocation. Nothing in a stack
                       says "this position has been tried five times and here is what worked."
                       Same seam as §17's — the OS covers scoping completely and learning not at
                       all.

THE CALLER OWNS THE     a function's return type is the callee's. Here the PARENT owns the return
RETURN TYPE            contract, and "no engram owns the contracts by which every possible caller
                       in the system must use it." Contract satisfaction is contextual; a
                       signature is not.
```

### And the word is not available

Recorded because the analogy will tempt a rename, and the set has a rule about this that has
fired five times — ROOT, CHAIN, TREE, MOUNT, INGRESS, each retired because *the word implied
something false*. `span` went the same way on 2026-08-17.

```text
FRAME IS TAKEN     frame.Frame already means what a possibility SEES — the restricted view
                   rendered for the crossing. Two Frames in one system is the defect MOUNT was
                   retired for: "five existing mechanisms under one word."

AND FRAME IMPLIES  dies-on-return, which is wrong here in the one place it matters most.
THE BROKEN HALF    "possibility" carries what frame does not and the design needs: POTENTIAL,
                   NOT YET ACTUAL — which is the whole distinction Think · Decide · Reach rests on.
```

```text
USE IT TO REASON, NOT TO NAME. The unwinding correspondence is a genuine finding and should be
used on the give-up predicate. The lifetime correspondence is false and should never be built on.
```

---

## Where the parallels came from

All of parallels 4–11 were measured on one MacBook Pro on 2026-08-12, in the course of tracing a
single question — *where does `read-local(file.txt)` actually run* — which is
[`../syscall-read-example/example.md`](../syscall-read-example/example.md).

```text
THE METHOD WAS THE POINT   every one of these came from running the thing rather than recalling
                           it. Three claims that survived recall did NOT survive measurement:
                           that a library has no address, that LAN and internet are different
                           mechanisms, and that arch changes what a syscall DOES rather than how
                           it is reached.
```
