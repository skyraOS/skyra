# Surface refactor

Date: 2026-09-30  
Status: to do after [24](24-the-room.md) is walked through. Items are added as
the walk finds them.

## The rule

> **A surface is what a program is, not what it may do.** Its identity (the
> binary, pinned by hash) and its contracts (input and output). Its reach and
> its OS grants belong to the Horizon it runs under, and everything a call
> varies travels in the payload, checked by the input contract.

`horizon.Surface` (`core/horizon/horizon.go:85`) mixes five concerns today:
identity (`Path`, `SHA256`), contract (`Input`, `Output`), authority (`Reach`,
`Authority`, `Delegate`), the resident lifecycle (nine fields, Codex only),
and the execution mode (`Host`).

## 1. Delete `Surface.Reach`, `Authority` and `Delegate`

**Now.** A program declares the names it wants in its `contract.json`
(`"reach": [...]`: `user`, `agent`, `codex/resident`, `codex/adapter`), and
its OS grants in its registry entry. The registry copies both onto the surface
(`registry.go:73`, `:148`) and checks each name is registered (`:151`).
`Narrow` grants a child the intersection of its surface's `Reach` and the
parent's reach, attenuated by `Delegate` (`machine.go:469`). `Render` builds
the profile from `h.Surface.Authority` (`horizon.go:157`), so every surface
gets its own. `Leaf()` is `len(Reach) == 0` (`horizon.go:111`), and decides
whether a program reads a `Continuation` or its bare payload
(`machine.go:359`).

**Why they go.** All three are authority, and authority belongs to the
Horizon, derived from the state and the proposal (24). A program listing what
it may call or touch is the program stating its own authority.

**What replaces them.** Nothing new: `Horizon.Reach` already exists.

- **Reach and grants live only on the Horizon,** derived from state at every
  step. The runtime builds the `Horizon` struct after the lookup, as the
  in-memory form of that derivation, but it is never a source.
- **Every invocation shares its parent's Horizon** unless the parent requests
  a narrowing (24). `read` runs with the same reach and grants as its caller;
  being invoked is already within that authority.
- **The profile is rendered from the Horizon,** one per Horizon, plus the
  exclusions (the floor, and exec of the running program only).
- **Authority is not coupled to leaf or continuation.** `Leaf()` goes. A
  program that emits without addressing anything gets a `leaf` fault carrying
  its output (24), so a plain binary works without knowing the protocol.

**Open.**

- **What decides the stdin format.** Today `Leaf()` picks `Continuation` or
  bare payload. With reach gone from the surface, that is a property of what
  the program *is*, so it belongs on the surface as identity: an explicit
  kind, or read from its input contract.

## 2. Delete `Surface.Args`

**Now.** Fixed argv for a resident, from the registry entry
(`registry.go:79`, `:148`), appended at launch (`instance.go:361`,
`appserver.go:178`). No registry in the repo sets it.

**Why it goes.** Anything that varies how a program runs is input, and input
is checked by the input contract. A fixed argv in the registry is a second,
unchecked input channel.

**What replaces it.** Nothing, today: it is unused. If a resident needs
launch options, they are fields of its start call's input contract, and the
runtime renders them to argv. Authority is still bounded by the sandbox, not
by argv.

## 3. Move the network and resident fields off the surface

**Now.** Nine fields exist only for the Codex resident: `Resident`,
`Connection`, `CloseConnection`, `ResourceOwnerDepth`, `ResidentProfile`,
`ProxyHosts`, `ProxyPath`, `ProxySHA256` (and `Args`, item 2). They say how a
surface is kept alive, which connection it is handed, and where it may
connect.

**Why they go.** They are two different things, and neither is what a
program is:

- **Where it may connect is authority.** `ProxyHosts` and the proxy are
  network grants, so they belong to the Horizon and its rendered profile,
  like every other grant (item 1).
- **Keeping it alive is control flow.** Starting a resident, handing its
  connection to each operation, and closing it are steps of the system, and
  belong in the control flow, not on the surface. That needs a loop
  operator: something that keeps a program in custody across operations.

**Open.**

- The loop operator. Like every state, it has to be derived, not stated.
  Worth checking first what the continuation already gives: a continuation
  resumed round after round (resolve → execute → emit) is a loop, and the
  Codex adapter's `ioCall` steps already run on it. What it lacks is a
  resource that outlives one step, which is what custody of the connection
  (`Resources`) patches today.

**Order.** After items 1 and 2 and the changes in 24 are in. Take it up with
improvements item 4 (the Codex adapter).
