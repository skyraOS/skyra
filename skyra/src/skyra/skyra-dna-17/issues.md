# DNA-17 — open issues

Date: 2026-09-19. Exploratory generation; mocked/fake App Server proof only.

## Carried forward from the specs (still open after building)

1. **Operation vocabulary split.** DNA-17 uses an ad-hoc closed set
   (`opRecordInvocation`, `opAuthorizeStart`, `opSettleReturn`, `opSettleFault`)
   on the system plane and a separate `EmissionCommit` on the execution plane. The
   spec's clean "one closed vocabulary tagged system vs execution, and which ops
   may co-commit" is unresolved. In particular the execution plane does not yet
   express its writes as `[]Operation`.

2. **Identifier encoding / unforgeability.** `Identifier` is a `uint64` table key.
   It is unforgeable only because nothing hands the raw integer across a real trust
   boundary in-process. There is no opaque-reference / capability encoding, and the
   scope granularity `stateVia(id)` keys to is "the whole invocation record," not a
   finer grant.

3. **Horizon revocation mid-execution.** Not implemented. `fdEntry` records the
   `gen` it was minted under, but nothing pulls a grant or rechecks generation on an
   in-flight execution-plane request. The spec wants one liveness check covering
   revocation and settlement together; DNA-17 only checks settlement (`live`).

4. **Waiting-record contract.** `waitingOn` is a single `InvocationID` (one child).
   The general contract — how interpret identifies *all* the moves a satisfied
   dependency unblocks and authorizes them in the same commit — is not built; only
   the single-child Codex case is exercised. No multi-dependency join.

5. **Single-operation rule for execution-plane read-modify-write.** Not enforced.
   The Codex driver happens never to read canonical state during execute (it gets
   the child result via `DriverInput.ChildResult`), so the "no worker holds a
   snapshot across turns" condition is respected by accident, not by a rule.

6. **`state.Commit` vs `cu.Execute` naming.** Adopted the spec's preferred split:
   `Commit` for applying operations, `Surface.Ingress`/`Execute` for the world
   phase. Left as a note in case the project settles differently.

7. **Queue capacity, cancellation, shutdown coordination.** The queue is a fixed
   4096-buffer channel. `CU.Close` cancels the context and drains workers, but a
   worker blocked inside a real (non-fake) transport would depend on that
   transport's Close to unblock. Not a durable shutdown contract.

## New questions surfaced while building

8. **fd-per-session vs fd-per-dispatch (the big one).** DNA-17 reuses one fd for a
   parent invocation's whole life and re-dispatches to it on resume, contradicting
   `minimal-loop-v2.md`'s "mint an identifier for the resumed move." Which is
   canonical? If the spec's mint-on-resume stands, the model needs an explicit
   notion of a *provider session* distinct from the fd, and a rule for correlating
   async emissions that arrive between resumes to the right session.

9. **Async emissions have no owning Move.** `turn/started`, `item/tool/call`, and
   `turn/completed` arrive unprompted. DNA-17 correlates them via the persistent
   output callback bound to the fd. This means "one execution emits multiple Moves"
   is realized by a callback that outlives `Ingress`, not by the `range
   surface.Emissions()` shape in the spec sketch. The spec should state which is
   intended.

10. **Settlement-point asymmetry.** `Return` settles at interpret; `Failure`
    settles the fd at emission-commit (fence) then re-enters a Fault Move. Should
    both settle at the same seam? Making Return also settle at emission-commit would
    unify it but push protocol correlation into the dumb writer.

11. **Exactly-once root delivery is enforced with a `sync.Map` guard**
    (`rootOnce`), belt-and-suspenders on top of the settlement fence. Whether the
    fence alone should be sufficient (it appears to be, in these two paths) is worth
    confirming before removing the guard.

12. **`OpenPossibility` atomicity.** Open records the Possibility and enqueues the
    first Move; the enqueue is to a channel outside the writer lock's transaction,
    so "possibility record + first move commit together or neither" is only
    approximately true (a crash between them is not modeled — in-process only).
