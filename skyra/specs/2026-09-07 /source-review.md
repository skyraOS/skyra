# Source review ledger

**Reviewed:** 2026-09-07

All 59 source files were read and reviewed against the latest checkpoint and entity/host thread. No file was copied whole.
Selected surviving sections from 10 sources were reconciled into four current documents; 49 source documents were excluded from the port.

An exclusion can mean superseded, speculative/out of current scope, or already represented by a newer source. It does not delete the historical document. In particular, old implementation-language choices are not promoted to new engine requirements.

Current conversation decisions take precedence over old “decided” banners. This ledger records the review; it is not another source of runtime requirements.

## 2026-08-24

| Source | Port action | Review result |
|---|---|---|
| [skyra-dna-v10/potentials/test.md](../2026-08-24/skyra-dna-v10/potentials/test.md) | Excluded | ConnectionContract/Move.Contract and universal probing rungs not current contract design. |
| [skyra-dna-v10/potentials/user-as-privilege-boundary.md](../2026-08-24/skyra-dna-v10/potentials/user-as-privilege-boundary.md) | Excluded | Resident privilege entity/cut/mint recipe not current; policy/mechanism covered by v11 ownership extract. |
| [skyra-dna-v10/questions.md](../2026-08-24/skyra-dna-v10/questions.md) | Excluded | Old THINK framing, no-state doctrine and quiescence assumptions superseded. |
| [skyra-dna-v10/research/decisions.md](../2026-08-24/skyra-dna-v10/research/decisions.md) | Excluded | Old routing/grant/trust-block mechanisms; principal/mechanism covered by later source. |
| [skyra-dna-v10/research/pipeline.md](../2026-08-24/skyra-dna-v10/research/pipeline.md) | Excluded | Historical analogy and explicitly superseded stage ordering. |
| [skyra-dna-v10/research/socket-triple-to-fd.md](../2026-08-24/skyra-dna-v10/research/socket-triple-to-fd.md) | Excluded | OS socket research; does not specify the current one-table fd model. |
| [skyra-dna-v10/think/acquire.md](../2026-08-24/skyra-dna-v10/think/acquire.md) | Excluded | Fixed discover/test/admit and two-grant onboarding protocol not reaffirmed. |
| [skyra-dna-v10/think/discover.md](../2026-08-24/skyra-dna-v10/think/discover.md) | Excluded | RegistrationProposal/getaddrinfo graph not accepted current routing surface. |
| [skyra-dna-v10/think/questions.md](../2026-08-24/skyra-dna-v10/think/questions.md) | Excluded | Unresolved planning/risk questions and old probe restrictions, not accepted local-loop requirements. |
| [skyra-dna-v11/bootstrap/ctu.md](../2026-08-24/skyra-dna-v11/bootstrap/ctu.md) | Selected sections → [ownership-and-drivers.md](ownership-and-drivers.md) | Mechanism versus authority and platform-backend separation only; no old bootstrap/handshake recipe. |
| [skyra-dna-v11/bootstrap/kernel.md](../2026-08-24/skyra-dna-v11/bootstrap/kernel.md) | Excluded | Kernel-does-not-coordinate thesis reversed; seed/authority principle covered elsewhere. |
| [skyra-dna-v11/bootstrap/sequence.md](../2026-08-24/skyra-dna-v11/bootstrap/sequence.md) | Excluded | Main-exits, one-kernel-per-old-host and handshake recipe not selected; address placement superseded. |
| [skyra-dna-v11/ownership.md](../2026-08-24/skyra-dna-v11/ownership.md) | Selected sections → [ownership-and-drivers.md](ownership-and-drivers.md) | Human policy/kernel provisioning/CTU execution only; no live-view registry or direct CTU recovery. |
| [skyra-dna-v11/research/README.md](../2026-08-24/skyra-dna-v11/research/README.md) | Excluded | Historical research index. |
| [skyra-dna-v11/research/authorization-and-registration.md](../2026-08-24/skyra-dna-v11/research/authorization-and-registration.md) | Selected sections → [ownership-and-drivers.md](ownership-and-drivers.md) | Availability versus authorization only; no selected token protocol, lease scheme, or complete-safety claim. |
| [skyra-dna-v11/research/capability-table-and-horizon.md](../2026-08-24/skyra-dna-v11/research/capability-table-and-horizon.md) | Excluded | Live horizon view, kernel=MMU, page-fault authorization and dual-table model superseded. |
| [skyra-dna-v11/research/cpu-execution-model.md](../2026-08-24/skyra-dna-v11/research/cpu-execution-model.md) | Excluded | CPU analogy and unapproved inference; current executable/live separation sourced directly. |
| [skyra-dna-v11/research/decisions.md](../2026-08-24/skyra-dna-v11/research/decisions.md) | Excluded | Historical Starlark leaf choice/graph-frontend deferral retained at source; not reaffirmed as new engine dependency or requirement. |
| [skyra-dna-v11/research/memory-hierarchy-and-faults.md](../2026-08-24/skyra-dna-v11/research/memory-hierarchy-and-faults.md) | Excluded | Seed/BodyID/two-fault/Context loader superseded by current filesystem Resolve direction. |
| [skyra-dna-v11/research/move-and-the-kernel-barrier.md](../2026-08-24/skyra-dna-v11/research/move-and-the-kernel-barrier.md) | Excluded | Transfer Request/Response and dual socket/fd tables superseded by Move/opcode ingress. |
| [skyra-dna-v11/research/sip-vs-pstn.md](../2026-08-24/skyra-dna-v11/research/sip-vs-pstn.md) | Excluded | Kernel-free/no-coordination thesis reversed; current entity-host thread owns routing. |
| [skyra-dna-v11/research/starlark-computation-and-dataflow.md](../2026-08-24/skyra-dna-v11/research/starlark-computation-and-dataflow.md) | Excluded | Historical evaluator/frontend design not reaffirmed for this engine; no new language, IR or opcode imported. |
| [skyra-dna-v11/research/starlark-frontend-surface.md](../2026-08-24/skyra-dna-v11/research/starlark-frontend-surface.md) | Excluded | Explicitly speculative frontend/IR/operator surface, deferred at source. |
| [skyra-dna-v11/research/think-act-paths.md](../2026-08-24/skyra-dna-v11/research/think-act-paths.md) | Excluded | Kernel-free hot path, capability caches and old IDs not current. |
| [skyra-dna-v11/research/think-act-signature.md](../2026-08-24/skyra-dna-v11/research/think-act-signature.md) | Excluded | Unconfirmed negotiation/quadrant design; full parent contract checking deliberately deferred. |

## 2026-08-27

| Source | Port action | Review result |
|---|---|---|
| [README.md](../2026-08-27/README.md) | Excluded | Historical index and obsolete status claims. |
| [ack-and-response.md](../2026-08-27/ack-and-response.md) | Excluded | Ack/Header and dual-plane/table protocols unchosen; Issued removed. No acknowledgement architecture imported. |
| [carry-over-from-08-24.md](../2026-08-27/carry-over-from-08-24.md) | Excluded | Prior review manifest; original sources reviewed directly instead. |
| [children-and-the-frame.md](../2026-08-27/children-and-the-frame.md) | Excluded | Context/Work menu placement superseded by Focus; surviving visibility principle already sourced from current Focus doc. |
| [context-identity-and-races.md](../2026-08-27/context-identity-and-races.md) | Excluded | Old context identities, membership latch, CAS and memoization assumptions superseded or unresolved. |
| [contracts-and-remember.md](../2026-08-27/contracts-and-remember.md) | Selected sections → [move-register-return.md](move-register-return.md) | Discovery/REMEMBER direction only; no name-only contract, type registry, mandatory guard or deopt machinery. |
| [dispatch-and-decode.md](../2026-08-27/dispatch-and-decode.md) | Excluded | Four-stage Local/Cut/Trap dispatch, CTU publication and direct trap mechanics superseded. |
| [horizon-and-attention.md](../2026-08-27/horizon-and-attention.md) | Selected sections → [authority-and-focus.md](authority-and-focus.md) | Kernel-created workflow authority; no Attention field rules, Work/Issued menu, or kernel-free hot path. |
| [kernel-handler-block.md](../2026-08-27/kernel-handler-block.md) | Excluded | CTU handler table and Host-as-privilege-block model superseded; surviving ownership covered by later source. |
| [message-the-wire-unit.md](../2026-08-27/message-the-wire-unit.md) | Excluded | Explicitly superseded by universal Move; no Message or Seed carrier. |
| [move-and-possibility.md](../2026-08-27/move-and-possibility.md) | Excluded | Old Move/ID/input/whole-blob expansion mechanics superseded; portable/live distinction sourced elsewhere. |
| [possibility-vs-context.md](../2026-08-27/possibility-vs-context.md) | Excluded | Context partition/Seed/cut identity model superseded. |
| [register-lifecycle-and-transition.md](../2026-08-27/register-lifecycle-and-transition.md) | Excluded | Three-table state model, scalar slots and ISSUE/LAND loop superseded. |
| [research/README.md](../2026-08-27/research/README.md) | Excluded | Research index, not current design. |
| [research/cpu-privilege-mechanism-vs-policy.md](../2026-08-27/research/cpu-privilege-mechanism-vs-policy.md) | Excluded | Analogy/research, not accepted interfaces; no CPU topology imported as architectural proof. |
| [system-and-semantic.md](../2026-08-27/system-and-semantic.md) | Excluded | Exploratory four interpreters, nominal contract, Host privilege taxonomy and guard/wake protocol not accepted. |
| [think-and-act.md](../2026-08-27/think-and-act.md) | Excluded | Fixed ACT/JIT compiler and error-handler roles not reaffirmed; discovery direction sourced from contracts doc. |
| [what-travels.md](../2026-08-27/what-travels.md) | Selected sections → [move-register-return.md](move-register-return.md) | Portable Move/authority versus resident live state only; no Seed/Context wire or destination-CTU loader. |

## 2026-09-01

| Source | Port action | Review result |
|---|---|---|
| [README.md](../2026-09-01/README.md) | Excluded | Historical index with contradictory status banners and obsolete runtime model. |
| [body-registers-and-return.md](../2026-09-01/body-registers-and-return.md) | Selected sections → [move-register-return.md](move-register-return.md) | Return obligations, roles/types and discovery only; omitted Action+Schema and proposed durable parent/child fields. |
| [closed-loop-outstanding.md](../2026-09-01/closed-loop-outstanding.md) | Excluded | Old transition table and fault traces replaced by current checklist/checkpoint. |
| [engram-loading-and-fault-return.md](../2026-09-01/engram-loading-and-fault-return.md) | Selected sections → [engrams-and-resolution.md](engrams-and-resolution.md) | Durable unit, storage and current driver Resolve; shared-Memory ownership also extracted into ownership doc. Old identities/decoder/events omitted. |
| [execution-brainstorm/README.md](../2026-09-01/execution-brainstorm/README.md) | Excluded | Brainstorm index and obsolete operand open tasks. |
| [execution-brainstorm/args-decision.md](../2026-09-01/execution-brainstorm/args-decision.md) | Excluded | Derived from no-forward-data rule, now superseded; argument ABI not adopted. |
| [execution-brainstorm/code-emission-and-act.md](../2026-09-01/execution-brainstorm/code-emission-and-act.md) | Excluded | Unapproved ACT/JIT and leaf taxonomy; no current need to import the brainstorming document. |
| [execution-brainstorm/decision.md](../2026-09-01/execution-brainstorm/decision.md) | Excluded | Mixed historical ledger; confirmed Move/opcode/authority principles already covered by later sources, old mechanisms excluded. |
| [execution-brainstorm/leaf-and-directions.md](../2026-09-01/execution-brainstorm/leaf-and-directions.md) | Excluded | No-forward-data and required const-leaf model superseded. |
| [execution-brainstorm/sys-vs-payload.md](../2026-09-01/execution-brainstorm/sys-vs-payload.md) | Excluded | Obsolete Move carrier and missing-payload premise; old host/address framing replaced. |
| [focus-and-move.md](../2026-09-01/focus-and-move.md) | Selected sections → [authority-and-focus.md](authority-and-focus.md) | Horizon/Focus/dependencies, Action, and replay only; omitted old backing field and durable contract fields. |
| [ingress-and-fire.md](../2026-09-01/ingress-and-fire.md) | Excluded | Three-peer ownership, claims, closed events, and CTU publication superseded. |
| [operand-model.md](../2026-09-01/operand-model.md) | Excluded | Missing-forward-input premise resolved by invocation Params; proposed Operand carrier not adopted. |
| [register-decision-review.md](../2026-09-01/register-decision-review.md) | Excluded | Historical inventory of superseded register/identity proposals; not a current specification. |
| [register-frame.md](../2026-09-01/register-frame.md) | Selected sections → [move-register-return.md](move-register-return.md) | Fd/register shapes, ordered values and parent addressing; replaced claim/events and CTU publication with latest loop. |
| [the-reactor.md](../2026-09-01/the-reactor.md) | Excluded | ISSUE/LAND and table-membership lifecycle superseded. |
