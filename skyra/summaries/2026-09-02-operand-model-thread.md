# Operand model — thread summary (for the road)

**Date:** 2026-09-02
**What this is:** a portable digest of a working session on Skyra's biggest open seam — the **operand model**
(seam A): where a move's input data comes from. Readable without the code in front of you. Full notes live in
`specs/2026-09-01/execution-brainstorm/`; the decision ledger is `.../execution-brainstorm/decision.md`.

---

## The problem in one paragraph

Skyra's execution model made an elegant claim: a move's inputs are its children's returns — data flows UP the
tree, nothing is pushed down. That let the design delete the old `Reads` field. But it left three things with
no home: **literals** ("hello"), **the file path / bound arguments** ("read file X"), and **the request
itself** (it arrives, THINK reads it to pick a plan, then it vanishes). A leaf move — one with no children —
literally has no way to receive data. That single hole blocks the whole engine: the "ready" check, the driver
call, the run-identity hash, and the entire reuse/WARM tier all depend on knowing what an operand is.

## The reframe that clarified it: SYS params vs payload

A move on the wire carries two categorically different kinds of thing:

- **SYS params** — what the kernel needs to FIND and AUTHORIZE the executor: who/where (Target), the method
  (Action), the authority (Horizon), where the result returns (ReturnFd). *These are all present today.*
- **PAYLOAD** — the actual DATA the computation reasons over (the Observations). *This is entirely missing.*

So the operand seam, stated cleanly: **the wire carries a full SYS half and an empty data half.** Give the
payload half a home the way the SYS half already has one.

## What we learned from the CPU (the CALL instruction)

We studied how a real CPU does a function call, and it reshaped the whole question.

- A `CALL` instruction's operand is **only the target address** — "where to jump." Nothing else.
- The **arguments are NOT part of the call.** They're placed in registers/stack *before* the call, by
  **convention** (the ABI). The CPU has no idea "parameters" exist; it just saves a return address and jumps.
- Why are they separated? Several forces, all pointing the same way:
  1. **The target is stable; the args vary every call.** Fusing them would kill reuse — you couldn't call the
     same function with different inputs. (This is the big one, and it's the same reuse the WARM tier needs.)
  2. An instruction is fixed-size; an argument list is unbounded — it physically can't fit.
  3. Args are usually *computed* (results of prior work), not constants you could bake in.
  4. **Mechanism vs policy:** the CPU provides the raw jump; *where args go* is a swappable software
     convention (the ABI), so many languages/OSes coexist on one chip.

**Verdict recorded (decision D3):** keep the **target in the operand** (in Skyra: Target + Action stay
together as the callee identity), and treat the **arguments as a separate channel** the call doesn't encode.
This overturned an earlier lean ("dissolve action + parameters into one payload") — the CPU says: don't fuse.

## The three ABIs, and who owns the convention

A compiler juggles **three overlapping contracts (ABIs)** at once: the **CPU/architecture** ABI, the
**OS/platform** ABI, and the **language** ABI. Important corrections we pinned down:

- A compiler **conforms to** the CPU and OS ABIs (handed down as specs it must obey) and **authors** only the
  **language** ABI. It doesn't invent them all.
- The convention does **not** live inside the compiler — it lives in an external, published **spec**. The
  compiler just emits code that obeys it.
- **Who wants a convention?** Two parties who are built independently and must agree **without ever
  communicating** (a caller and callee compiled years apart). Convention is the coordination rule for an
  anonymous channel. *This is the test for whether Skyra needs one.*

## How a compiler couples parameters to a call

There is no stored "coupling." Both sides — caller and callee — independently obey the **same rule** derived
from the **function signature (the contract)**:

- **Contract** = the signature: how many params, their types, their ORDER.
- **Compiling** = turning that contract into placement code: the caller puts args in the agreed slots; the
  callee's prologue reads them back out as named locals.
- **Timing** = the runtime sequence: caller places args FIRST, then calls; callee reads them on entry.

The parameter *name* is compiled away into a *position* (register #1, #2, …) because CPU registers are
**anonymous** — position is the only thing both sides can agree on.

## The Skyra insight: bind NOMINALLY, not positionally

Here's the payoff. **Skyra's registers are named** (`RegKey{Name}`), unlike the CPU's anonymous registers. A
positional calling convention is only forced when the channel is anonymous. Since Skyra's isn't, **it can bind
operands by NAME** — pure dataflow: a child writes register "path", the parent reads register "path." No
positional convention needed. (This is closer to keyword arguments than to C's positional call.)

**Leaning decision (D4): bind nominally; no internal calling convention.** Ordering isn't needed for the
dataflow — the dependency graph handles "what feeds what." A convention only re-enters in two spots, and both
are "conform," not "author":

1. **Remote boundary** — when a target resolves its own body independently, caller and callee still need to
   agree on parameter *names + meanings*. But that's a **nominal contract** (already Skyra's `Contract`), not a
   positional convention.
2. **The driver edge** — the actual `Perform(action, in)` call takes an *ordered* list, and a real OS driver
   may have a genuinely positional expectation (it's conforming to the real machine's ABI). So at that one
   seam, named registers get marshaled into the driver's expected order — owned by the driver's declaration.

Only at the hardware edge does a positional convention survive. Everything internal is nominal.

Caveat to remember: nominal binding puts a **burden on THINK to name slots unambiguously** (e.g. `minuend` vs
`subtrahend` for a subtract). And for the run-identity hash, a named map still needs **canonical key ordering**
— but that's mechanical (sort the keys), not a semantic convention.

## A structural idea that came up: split ACT and COMPILE

Proposal (revises `think-and-act.md`, so it's on the record as a change): **THINK and ACT share the
DECLARATION phase; the COMPILE step is REMEMBER.** This maps cleanly onto a JIT's lifecycle:

- THINK = the plan (structure + names + the contract with the user).
- ACT + the first (COLD) run = the live execution where the LLM routes observations to the children that need
  them, and types get **discovered by running**.
- REMEMBER = records the discovered contract (the "profiler / type-feedback").
- The reusable, guarded WARM body is **produced from what REMEMBER recorded** — i.e. compilation-proper
  consumes REMEMBER's output.

One refinement flagged: keep **record** (REMEMBER) and **emit** (produce the guarded body) as two distinct
functions even if they sit at the same point — otherwise it gets murky later where the type-guards are added.

## Open items to carry forward

```text
1. WHERE the payload/arg channel physically lives — pre-seeded named registers vs a payload slot on the Move
   vs the ordered operand list. (D3 open.)
2. Confirm D4 (nominal binding): try to break it — is there any internal case that truly needs positional order?
3. The map -> ordered-slice marshal at the Driver edge — pin down that it's owned by the driver's Decl.
4. The ARGS question (materialize-time argument vs compile-time literal) — the lean is YES (args exist), because
   without them reusable bodies die and the WARM tier evaporates. Nominal binding is compatible with this.
5. Formalize the ACT/COMPILE/REMEMBER split against think-and-act.md (it's a revision).
6. Seam E (run-identity hash): canonical ordering of the named inputs; exclude impure (NOW/RANDOM) leaves from
   reuse.
```

## The one-line takeaway

**Keep the target in the call; make arguments a separate channel; bind them by NAME (Skyra's registers are
named, so no positional convention is needed) — except at the driver edge, where you marshal to the order the
real machine expects.**
