# claude-ideas — v11 shape proposals

**Date:** 2026-08-24
Every SHAPE below is mine (Claude's), not the docs'. Terms are the docs'; the *types, returns, and
package split* are proposals — all **rejectable**. Reshape any and say why in `boundary-crossings.md`.
Ordered most-certain → least; the three I'd check first are **[CHECK FIRST]**.

```text
1  BODY vs CONTEXT are separate types.  [CHECK FIRST]
   Body = compiled/static/content-addressed; Context = live/mutable/per-run. Materialize is the loader.
   seam: does Context really need its OWN copy of the move state, or can it index the Body's templates +
   a parallel state array? and does the Body carry contracts or only shapes (think-act-signature)?

2  Capability IS the horizon unit, macaroon-shaped, USER-HELD.  [CHECK FIRST]
   Horizon = []Capability; the user carries it and presents it on the hot path; the CTU verifies offline.
   seam: is a whole Horizon carried per Transfer, or one Capability per target? what exactly is a Caveat
   in Skyra terms? and does offline Verify need the kernel's public key distributed to CTUs (it does)?

3  the cold/hot split is Kernel.Handle vs CTU.Move.  [CHECK FIRST]
   Handle is the only cold entry; Move is the only hot entry. seam: on the hot path, who holds the CACHED
   Body — the CTU (a local cache) or is it re-fetched by content-address on each call? if cached on the
   CTU, does that leak state into a "nearly-stateless" core? (lean: content-addressed fetch, CTU caches
   as a hint only.)

4  Materialize can run on BOTH kernel (cold) and CTU (hot).
   Body + Horizon → Context. seam: is it truly the same op both places, or does the cold path's
   materialize differ (it also provisions/mints)? maybe Materialize is pure and provision/mint are separate.

5  Move is the CTU's single entry; Header dispatches Request/Response.
   subsumes v9's Greet/Recognize/Arrive. seam: does the hot-path entry (present a capability + engram)
   fit Header=Request cleanly, or is "invoke a remembered engram" a third header/shape?

6  the fd table lives on the Context; the socket table on the kernel.
   Context.InFlight (cut-ID → register) = fds; kernel Barrier holds routing/custody/timeouts. seam:
   confirm the split holds when a Response must find both the register (Context) and the route (kernel).

7  MoveTemplate.Return bool is the block-participation switch.
   whether a move expects a return is structural (THINK); the type is negotiated at ACT. seam: is a bool
   enough, or does "expects a return" need to name the return contract here?

8  Entity → Host → Address, host(entity) not modelled as a call yet.
   Entity holds Home HostID; Address is resolved elsewhere. seam: where does host(entity) live — a Host
   interface, a kernel op, or a driver? (bootstrap A is still LEANING.)

9  Think.Compile is one call; Acquire/Compose are named but not wired.
   seam: is THINK a compiler function or a possibility-tree whose children are acquire/compose moves? if
   the latter, Compile dissolves into "run a think-possibility." (carried over from v10 idea 1.)

10 REMEMBER lives on the kernel Table.
   caching a compiled Body = registration. seam: is remember content-addressing the Body, the (request →
   Body) mapping, or both? how is "the same request" matched to a cached Body on the next call?
```

## The one I most expect to be wrong

**#3's caching question.** "The user holds the capability, the CTU verifies offline" is clean, but WHERE
the cached Body lives on the hot path (CTU-local cache vs content-addressed re-fetch) decides whether
the CTU is actually stateless. If it must cache Bodies to be fast, "nearly-stateless" needs re-stating.
