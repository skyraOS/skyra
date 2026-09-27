# Authority owner

Date: 2026-09-20

Status: Proposed composition to test. Not a kernel or process-placement decision.

Power: provision Horizons, change grants, and establish trusted identities.
This role also protects the canonical records that make those decisions meaningful.

## Composition

| Existing component | Why it is placed here |
|---|---|
| [Possibility](possibility/proposal.go) | Owns the run's authority context and invocation frames logically. |
| [Horizon](horizon/proposal.go) | Defines authority and gates each Move's executable resolution. |
| [AuthorityChecker](authority-checker/README.md) | Receives a Move, resolves its Possibility, and returns the checked Move with its Horizon. Namespace ownership remains open. |
| [StateOwner](state-owner/README.md) | Stores canonical state and serializes complete interpret/commit turns. |
| [SurfaceRegistry](surface-registry/README.md) | Controls what implementation a granted SurfaceID names. Mutation can change executable meaning. |
| [TrapController](trap-controller/proposal.go) | Runtime-owned handling of traps/faults; precise recovery powers remain open. |
| [DNA-17 reference](reference/dna17.go) | Historical snapshot only; contains contracts spanning all three roles. |

Grouping these components does not give each of them every power of the role.
StateOwner is a storage/serialization component, not permission for every caller
to mint grants. TrapController remains a Surface with explicit, bounded authority.

## Boundary with execution

Interpretation admits the Move and prepares its exact executable binding while
the Horizon and registry view are consistent. Commit records that authorization.
The broker receives committed work; it does not choose an arbitrary registered
handler after getting a boolean permission result.

Broker outcomes re-enter interpretation with trusted source attribution. They
cannot directly mark canonical invocations returned, mutate grants, or install
new handlers. State operations requested as Moves enter the same admission path.

## Placements to challenge

- **Registry:** mutation belongs here because replacement can alter existing
  grants' meaning. Binding realization may need broker-owned private resources.
  If the current `Surface.Driver` field mixes these responsibilities, split the
  registration record from the live resource binding rather than add a second
  registry writer.
- **TrapController:** placed here as runtime recovery logic, not merely because
  it is named a controller. If it invokes an agent to propose recovery, that
  agent belongs in the worker role and its proposal still needs admission.
  Whether the controller itself needs authority beyond ordinary grants is open.
- **StateOwner:** its pool hosts state turns, but authorization, dependency
  tracking, and prepared writes remain separate responsibilities within a turn.
- **Provisioning:** we still lack a concrete provisioning service/interface.
  Possibility and Horizon structs describe its records, not its implementation.

The question for this role is: which operations can change what another component
is allowed to do? Those operations must not leak through general state or registry
access offered to the broker or workers.
