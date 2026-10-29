# The release is the Boundary

Date: 2026-09-23  
Status: agreed 2026-09-23 (D6). Names the handle from [05](05-the-release-is-a-handle.md).

## The idea

The parent reads the child's commit and returns the child's **Boundary**. The
release, the handle and the Boundary are one object. Provisioning happens exactly
at the handoff, which is where
[possibility-as-process](https://github.com/skyraOS/tony-stark-was-able-tobuild-this-in-a-cave/blob/b243b038919ea0c97781cf487c648f9e156fa1d2/history/specs/2026-09-21/possibility-as-process.md) says it
belongs: "authority is checked once, at provision". Execution against what the
Boundary holds is then free.

```text
N+1 intake:  commit R_{N+1}           ──stamp──▶  N reads R_{N+1}, runs its 3 checks
                                                  N computes H_{N+1} = narrow(H_N, R_{N+1})
N+1 execute: Execute(H_{N+1})         ◀─Boundary─  N sends H_{N+1}, exits
```

## What a Boundary holds

| Part | Kind | Contents |
|---|---|---|
| `stamp` | data | `hash(R_{N+1})`. Binds this Boundary to one commit |
| `surface` | data | The binary, args and Seatbelt profile this frame may execute, taken from `R` and checked against `H_N` |
| `grants` | data | What T's emissions may invoke next. This is the ceiling for the child's own children |
| `node` | fd | This frame's node. The capability to act |
| `pipes` | fd | T's stdin and stdout, pre-wired by the parent (optional; see 05, open point 2) |

The data parts say what the frame may do. The fd parts are what make it able to
do it. Only the parent holds the fds, so only the parent can issue a Boundary.

## Rules

1. **A Boundary only narrows.** `H_{N+1} ⊆ H_N`, always. `narrow` intersects what
   `R_{N+1}` asks for with what `H_N` holds. Asking for more yields the
   intersection, or a fault if the surface itself is not covered.
2. **A Boundary is computed, not stored.** `narrow` is a pure function of `H_N`
   and `R_{N+1}`, so a Boundary can be derived again on replay or restore by
   walking the chain of records from the root. This keeps the "no second commit"
   decision: the release still writes nothing.
3. **The fds are live and never recorded.** On restore they are gone. Records
   written before a crash are data, and the frames that were running are faulted.
4. **`Execute` takes only a Boundary.** `Execute(h Boundary)`. The executable, args
   and profile come from `h.surface`, which was derived from the commit.

## What this settles

- **Boundary narrowing now has one home.** It happens at the release, computed by
  the parent, which is trusted and is not the emitter.
- **Frame 0.** The root's Boundary comes from the boot seed: the first `narrow`
  starts from the operator's grant.

## Open

1. **Which Boundary authorizes the emission during intake?** The emission comes
   from T, running under `H_N`, so N+1 must check it against `H_N`. Either N
   sends `H_N`'s data part alongside the bytes (N is trusted), or N+1 derives it
   from the chain. Sending it is cheaper. Deriving it trusts less.
2. **Executing by handle.** macOS has no `fexecve`, so the surface is executed by
   path, and the Seatbelt profile is what confines it. The binary cannot be an fd
   in the Boundary, only a path plus a content hash, checked at spawn.
