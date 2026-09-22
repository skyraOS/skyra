# Boundary report — 2026-09-06

> Historical probe report. The subsequent accepted [fd-entry handoff](surface-proposal.md#accepted-revision-kernel-hands-executable-fd-entries-to-ctu)
> gives Move and Register a shared storage home and changes Fire to accept entries. Installation
> and execution remain unimplemented; the findings below describe the original frozen surface.

**Result: partial components work; the local loop is blocked.** Passing scoped build/vet/test checks
do not establish execution or a root observation. Authority was README, the source implementation
probe rules, and the three design documents named in the work order. No frozen shapes were changed.

## Interfaces that held

| Boundary | Evidence / limit |
|---|---|
| Runtime.Boot → Kernel/CTU Memory | Existing test preserves one shared preloaded frame and child order. |
| Memory.Land(parent, child, observation) | Can select and fill one ordered child entry, protect producer bytes, and reject conflicts without extra state. This alone does not satisfy a contract. |
| Scheduler.Claim | Existing lifecycle map supports serial claim-before-handoff and duplicate suppression. Failed-handoff recovery is still absent. |
| Scheduler.Apply(Issued/Landed) | Existing map supports retained completion and late acknowledgements without opcode or register access. Publication is a trusted event precondition; dependent admission remains unavailable. |

## Blocked interfaces and proposed minimum changes

These are proposals for review, not permission to extend this run's surface.

1. **Preload / Memory.Resolve.** Memory contains only Registers; Register contains no Move, Op,
   ParentFd, or backing reference. Even fully resident execution cannot recover an executable from
   an fd. A canonical action/params key would still omit return addressing and authority. Specify
   Memory-owned live descriptor storage and an installation method, independent of durable loading.
   Define Resolve's result for unknown fd versus known descriptor with missing backing. Do not encode
   descriptors in Schema, observation bytes, or MoveID.Value.
2. **Scheduler.Admit.** Scheduler holds only lifecycle states and receives only fds. It has no CTU or
   readiness query/result, and non-nil children alone do not establish satisfaction. Specify a CTU
   readiness interface and validated admission handoff, including how dependents are found after
   publication. The loop must not inspect Memory.Registers to manufacture Ready entries.
3. **CTU.Fire → Scheduler.Apply.** Fire returns nothing, CTU has no event sink, and ExecCtx intentionally
   has no scheduler. Even a successful Op.Fire event has no specified route out of batch dispatch.
   For a serial local probe, add an explicit event callback to batch dispatch (one event at a time);
   define later completion delivery separately. No global channel or direct scheduler-map writes.
4. **Ret.Fire validation and root delivery.** Ret.Value can carry a proposed observation and ParentFd
   suffices for child routing. Neither specifies how child values construct the parent's proposed
   output, schema dialect/unknown-schema handling, or required-child satisfaction. Define those
   semantics and their failure result before enabling Ret. Nil ParentFd explicitly requires a separate
   root destination in the design; choose a root publication interface and its relationship to Landed.
   Treating Landed.Value alone as the root destination would silently settle that open design choice.
5. **Exec / MMU.** Perform returns both Observation and Outcome, while Event cannot represent all
   driver failures/waiting results without conflating faults or success. Specify that mapping before
   performing effects. MMU has no verification dependencies/key access; matching target/action alone
   would omit the inherited capability signature requirement. Define trusted verification access.
6. **Recovery and lifetime.** Scheduler lacks parked states, kernel recovery routing, and wakeup
   ingress. Retirement has no eligibility signal/policy. These are unnecessary for the proposed
   preloaded example but remain explicit stubs, alongside durable installation/loading and REMEMBER.

No executable main loop was added: there is no valid first admission or descriptor resolution, and
even manual dispatch cannot transport its event. A driver-backed Move registry, a closure capturing
the graph in Op, or direct test orchestration of Ret would conceal those gaps. The tests exercise
only supported components and explicit boundary failures; they do not bypass the loop and claim
the resulting observations came from execution.
