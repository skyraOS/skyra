# Mediated actors over a kernel

Date: 2026-09-20
Status: Load-bearing ideas. Discussion input, not a decision. Records the
execution model the September 20 refinement converges on once "only a Move
crosses the pipe" and "authored code isn't in the runtime binary" are taken
seriously. Companion to
[engrams-surfaces-and-pipe-protocol.md](../engrams-surfaces-and-pipe-protocol.md).

## The model

Go actor — but **mediated actors over a privileged kernel**, not flat/pure
actor. Most of it is already built: a Move is a message, a port is an address,
`Handle(Move)` is receive, and units own their native state. The refinement is
about *which* clauses of "full actor" to keep and which one to refuse.

- **Actors** own their adapter state, are isolated, communicate only by Moves,
  and live behind ports. Above the trust line.
- **The loop is the kernel**: a *distinguished* actor, not a peer. It owns
  authority, routing, and the port table. Below the trust line (substrate).
- **Move = syscall / IPC. Port = the kernel's port table.** This is the OS the
  `grips-and-invariants` thesis already names ("the possibility is a process;
  make the kernel the first real driver"). Unix and Mach are actor-ish in user
  space with a privileged kernel — never flat. Flat actor (peers addressing
  peers, no kernel) is the version that dissolves the authority story.

## The one clause to refuse: authority as a chain of peer actor-hops

If the "authority | transition | services a Move passes through" are **peer
actors the Move physically hops through over ports**, two solved problems come
back:

- **The regress returns.** If an authority-actor stamps a Move and sends it on,
  *who authorized the authority-actor's send?* Another actor? This is `trace()`
  needing its own `trace()`. It does not terminate.
- **The atomicity loss (C3).** The single-writer interpret turn (read → decide →
  commit under one hold) exists precisely so authorization is *one atomic step*.
  Spread across N actor hops you get partial states, a gatekeeper dying mid-chain,
  and lost-update races between concurrent Moves. That un-solves C3.

So authority is **not** a mesh traversal. It is the **loop's single compile turn**
(THINK), in-binary, consuming a raw Move and emitting an AuthorizedMove / Engram.
"The Move passes through authority and the services" is real, but it is *compile
steps inside the loop*, not network hops between peers. The loop terminates the
regress by being substrate.

## Clauses 2 and 4 are the same mechanism

Two of the proposed clauses collapse into one rule:

- (2) a Move is not accepted unless authorized, and
- (4) the runtime manages what actor is on what channel,

both reduce to:

> **The loop is the sole writer of any actor's inbound port.**

Actors emit Moves *to the loop* (one door). The loop authorizes and is the only
thing that dispatches onto a target actor's port. An unauthorized Move therefore
**cannot reach the target** — not because a gatekeeper rejected it, but because
nothing except the loop can put a Move on that port. Rejection by topology, not
by check. This is DNA-17's settlement fence generalized ("a post-fault message
hits a dead fd, enqueues no Move, so no Return can appear"). No authority-actor
is needed; the wiring *is* the authority.

## Clause 1 needs the adapter/control split

"Everyone owns their own state" is true for one kind of state and false for the
other — and taking it literally breaks the requirement that the runtime always
knows where everything is (and trips the review's C8).

- **Adapter state** — yes, actor-private. Native session state, no shared memory,
  only Moves cross. Full actor.
- **Authority / routing / canonical state** — the port table, the Horizons, the
  bindings — is **substrate, not distributed.** It is not "just another actor's
  private state." The kernel-actor that owns it is *distinguished*, not a peer.
  Hand control state to peer actors and the trust boundary that terminates the
  recursion is gone, along with the single place that answers "where is
  everything / is this allowed."

So the rule is **"actors own adapter state; the substrate owns control state"** —
the below-the-line / above-the-line split, now with actors above and the kernel
below.

## The two planes

- **Authority / compile plane = the loop (kernel).** In-binary, single-turn,
  serialized. Input: a raw Move emitted by an actor. Output: an Engram of
  AuthorizedMoves. Authority, transition, and service checks happen here as
  compile steps.
- **Execution plane = actors over ports.** The Engram dispatches authorized Moves
  to backing actors. Actors do work, own adapter state, and emit new Moves that
  re-enter the authority plane.

The runtime cycle: actor emits Move → authority plane (loop compiles + authorizes
→ Engram) → execution plane (dispatch AuthorizedMoves to actors over ports) →
actors emit Moves → back to the authority plane.

## How this lands on the findings

- **C3** (serialized operations lost the old atomicity): keeping authority as the
  loop's single compile turn, not an actor-hop chain, preserves the one-turn
  guarantee. Do not distribute the authorization.
- **C4** (StateOwner is both a Surface and the whole-loop host): resolved by rank.
  StateOwner is an actor on the execution plane speaking the same protocol; the
  loop/kernel is the distinguished coordinator. They are different ranks, not one
  god-object.
- **C8** (adapter vs runtime-control state share an unspecified model): answered
  by the adapter/control split — control state is substrate, never granted to a
  peer actor by a bare port right.

## The one line

> **Actors above, kernel below; the kernel is the sole writer of ports, so
> authority is delivered by topology and authorization stays a single compile
> turn — never a hop through a peer.**
