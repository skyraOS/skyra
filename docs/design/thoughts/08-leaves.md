# Leaves

Date: 2026-09-23  
Status: agreed 2026-09-23 (D8). The delivery mechanism is proposed. Answers
proposal open question 6.

## The idea

Registration marks an executable as a **leaf** or not. A leaf can only return.
It cannot invoke, and it cannot be continued. Its instance handles exactly one
emission, and that emission must be a Return (or a Fault).

## Two classes of executable

| | Leaf | Non-leaf |
|---|---|---|
| Examples | `echo`, any tool | Codex, `user` |
| May emit | One Return or Fault | Invokes, and eventually its own Return |
| Can be resumed (receive a Return) | Never | Yes: it is live while its children run |
| Instance lifetime | One execute, one handoff, exit | Repeated execute and handoff while T runs |
| Boundary `grants` | Empty. `narrow` enforces this from the registration | Whatever `narrow` allows |

A leaf that emits an Invoke is caught at intake: the Invoke is not covered by
the Boundary it ran under, so it is committed as a fault.

## Why this answers open question 6

The question was how a tool's Return gets back into a live Codex when the
process holding Codex's stdin is two instances up the chain:

```text
N    supervises Codex (non-leaf, live, holds Codex's stdin)
N+1  supervises echo  (leaf)
N+2  intakes echo's Return
```

The leaf rule makes the path fixed and short.

1. **Only non-leaves are ever resumed, and a non-leaf's supervisor is alive while
   it waits.** N cannot have exited, because Codex is still running. So there is
   always a live party to deliver to.
2. **The resume channel travels a fixed, short distance.** When N hands off
   Codex's tool Invoke, it puts a one-shot **resume fd** in `H_{N+1}`, the write
   end of a pipe N reads. N+1 passes it on in `H_{N+2}` when it hands off echo's
   Return. N+2 is the only instance that uses it. A leaf has no further
   descendants, so the fd never travels any further.
3. **Delivery is a handoff in reverse.** N+2 commits the Return and writes its
   stamp to the resume fd. N reads `R_{N+2}` by stamp and verifies it (it exists,
   it links to the Invoke N handed off, and its verdict is a Return), then writes
   the result into Codex's stdin.

N is the only writer into Codex's stdin, so deliveries and anything else N sends
cannot interleave. N is not checking its own emission here: echo emitted the
Return, and N+2 checked it. N only verifies custody, as in any handoff.

## Why a leaf is also simpler

- **No continuation state.** A leaf is never resumed, so it has no
  `tool-result`-style profile and no pending-dependency bookkeeping.
- **No streaming.** One emission means the instance never loops.
- **Fan-out stays shallow.** Codex can issue several tool Invokes. Each goes to a
  leaf, with its own resume fd back to N.

## Non-leaf children

A non-leaf child (for example a sub-agent that runs its own tools) is itself
resumable, with its own live supervisor. The same rule applies one level down:
each non-leaf's supervisor hands out resume fds to its own children. Resume fds
go one Invoke deep and are never forwarded past the Return they belong to.
Leaves are what guarantee the tree ends.

## Consequences for the proposal

- Registry entries carry `leaf: bool`. It is part of the registered surface, so
  it is checked at registration and cannot be claimed in a Move.
- The Boundary gains a `resume` fd: present for any frame whose caller is live,
  and consumed by the frame's Return instance.
- Open question 6 is closed.
