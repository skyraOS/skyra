# Create-spec: an inspectable artifact from conversation observations

Date: 2026-10-06
Status: proposed, not implemented.

Builds on [Think](../docs/next/think.md), [Observation passing](../docs/next/21-observation-passing.md),
and [the think procedure](../docs/next/proposal/38-think-procedure-proposal.md).

## Purpose

The user actor invokes an authored `create-spec` procedure with source
observations. The procedure returns a structured spec artifact whose content
and placement decisions can be inspected through citations.

Creating or inspecting a spec does not replace model context, approve the spec,
or authorize its execution. The user decides what to do with the artifact later.

The procedure has its own root and outline on the human actor's UI. Its artifact
inspector exposes the evidence behind the result, separately from the procedure
outline.

## Requirements and their sources

These requirements come from the design conversation that requested this spec.
The quotations below identify that conversational basis; they are not runtime
observation IDs and must not be represented as verified runtime citations.

| Requirement | User observation in the design conversation |
| --- | --- |
| Invocation is manual. | “manuel: its a procedure most likely.” |
| Produce a spec observation from context. | “it turns the models context into a spec type observation.” |
| Account for every content word with conversation evidence. | “all written words in the spec need to be produced by cited observations. so every one is tied to something in the conversation. no word unaccounted for.” |
| Leave context replacement to the user. | “we probably shouldnt make that decision for the user yet.” |
| Make content and placement inspectable. | “the spec creator creating a spec that the user can inspect the citations for reasons why it put stuff in places is a good artifact.” |
| The user actor supplies observations; the creator is a visible procedure. | “the user passes the contact as a observation to the spec creator. The spec creator is a procedure so it has its own tree on the human actors ui.” |
| Clicking a constellation selects the contract context targeted by the user. | “contract context switches on the app when you click on it. thats where the user tells the adapter where im targettign the selection of the constellation.” |

The last quotation leaves the exact input terminology ambiguous. This proposal
distinguishes source context observations from contract observations: a contract
identifies an interface, but does not supply the conversation evidence by itself.

The mechanisms below are proposed implementation choices, not decisions already
approved by those quotations.

## Invocation and source boundary

The user actor selects the conversation evidence and invokes `create-spec`.
The invocation identifies:

- The source observations or a reference to a bounded source context observation.
- The source invocation and accepted observation frontier when history must be
  reconstructed.
- The target and history contract used for that reconstruction.
- The versioned spec schema to produce.
- A size budget.
- Any contract observation the artifact is intended to describe.

The history contract and the output spec schema are separate inputs. Passing a
schema as data does not grant authority or automatically enforce the output.

Use `system.strand` when source history needs reconstruction. Its existing
contract and frontier checks remain authoritative. Resolve the source before
narrowing the model's Horizon.

For the first implementation, invoke the procedure beneath the source user actor
and pass resolved evidence downward. Merely opening a new `create-spec`
conversation creates another room; possession of a source hash does not permit
that room to read the original conversation. Separate-room execution would need
an explicit evidence-transfer mechanism.

The source boundary is fixed for each invocation. Later messages do not silently
change an artifact already being generated.

## Procedure composition

```text
create-spec
  ├─ resolve-source
  ├─ project-evidence
  ├─ propose-spec       → narrowed Codex
  └─ validate-spec      → final artifact observation
```

### Resolve-source

Obtain the selected, authorized source observations. When reconstruction is
needed, use the history contract and accepted frontier with `system.strand`.
Reject inaccessible or invalid references rather than treating arbitrary supplied
hashes as evidence.

### Project-evidence

A deterministic leaf extracts citeable passages. Each passage retains:

- Its source record identity and source field.
- Its exact text range and original text.
- Its author or speaker attribution, to the extent established by the source.
- Its ordering and enough surrounding context to preserve qualifications.

Give the model short local evidence handles, such as `e17`, mapped to those
passages. The model does not invent hashes or calculate source byte offsets.

### Propose-spec

Codex selects and organizes evidence into requirements, proposals, decisions,
and unresolved questions. Every entry names its supporting passages and the
evidence supporting its placement.

In the strict first version, the model selects exact source excerpts rather than
writing unrestricted prose. A schematic candidate is:

```json
{
  "requirements": [
    {
      "source": "e17",
      "placement_evidence": ["e18"]
    }
  ],
  "proposals": [],
  "decisions": [],
  "open_questions": []
}
```

This example illustrates references and placement evidence; it is not the final
runtime contract. Section vocabulary, source-range representation, and budget
units still need definition.

The model invocation is narrowed to the operations needed for this task. It
does not gather unrelated evidence or perform implementation effects.

### Validate-spec

A separate deterministic leaf validates and assembles the artifact. It rejects
unknown references, invalid ranges, and uncited content, and resolves candidate
handles against the actual source evidence.

Every content word is drawn from a cited source passage. Fixed section labels
and formatting come from the versioned spec contract, rather than uncited model
output. Placement explanations obey the same sourcing rule as other content.

The validating leaf authors the final artifact result. The interpreter returns
that child result unchanged. Its source references remain attached to the
original observations, not merely to the model's candidate.

## Placement is an inspectable judgment

Word provenance and classification are different checks. A quotation can establish
where text came from without establishing that it belongs under “decision.”

An assistant proposal remains an assistant proposal unless the evidence establishes
user acceptance. Ordering, qualifications, rejection, and later corrections must
remain inspectable. Conflicting statements stay visible unless evidence explicitly
resolves them. Unclear acceptance remains unresolved.

Deterministic validation can establish reference validity and exact text sourcing.
It cannot generally prove semantic classification or completeness. The UI must
expose placement as a judgment with evidence, not portray it as a proven fact.

If paraphrasing is introduced later, complete citation coverage can be checked,
but entailment assessment is fallible. It must not be described as a guarantee
that every generated word is justified.

## Artifact and UI

Register `create-spec` in an authored procedure bundle. The existing destination
loader derives its outline from that definition, giving it a procedure root and
constellation without a separately authored diagram.

### Constellation selection is the target selector

Clicking a procedure constellation selects its contract context. This is the
human's explicit indication of which procedure and contract they are targeting;
the adapter must receive that selection from the user actor rather than infer it
from message text or ask the model to choose it.

Selecting `create-spec` changes the target for the next invocation. It does not
invoke the procedure, approve an artifact, or replace a conversation's context.
The selected contract, the selected conversation, and the source evidence passed
to the procedure are distinct values. Evidence selected from a source conversation
must remain explicitly attached when the human changes the target constellation.

The desktop already changes its local root selection on a constellation click
and puts away a conversation belonging to another root. The implementation must
carry the selected target and contract through structured invocation to the user
actor and adapter. A local visual selection alone is insufficient.

The contract context described here is invocation targeting, distinct from
adopting a spec as replacement model context.

Let the human select source evidence, target the `create-spec` constellation,
and explicitly submit the invocation. Persist its structured result as a
lineage-backed artifact.

The artifact inspector presents content and lets the user select an entry to see:

- The exact supporting source passages, their attribution, and their location.
- The evidence supplied for placing the entry in its section.
- The procedure invocation that produced the artifact.

The procedure outline shows composition. The artifact inspector shows provenance
and placement evidence. An installed outline alone does not provide artifact
inspection or a live invocation tree.

The existing conversation remains usable. Artifact creation introduces no
replacement model context, approval requirement, or automatic implementation step.

Conversation navigation can move out of the left sidebar into an ephemeral
translucent planet picker. This is a UI design proposal, described in
[the planet-picker draft](conversation-planets/README.md), separate from the
spec-creation mechanism.

## Existing mechanisms

- [`interpreter/proc/plan.go`](../controller/interp/proc/plan.go) runs
  dependency-ordered procedure nodes and evaluates a return bound to a child result.
- [`machine/room.go`](../controller/core/machine/room.go) implements
  `system.strand`, including contract and bounded history checks.
- [`machine/machine.go`](../controller/core/machine/machine.go) exposes child
  discharge stamps as `StepView.Observation`.
- [`web/destinations.go`](../controller/hosts/web/destinations.go) discovers
  authored procedure roots and derives their outlines from bundle definitions.
- [`codex/resident/main.go`](../controller/codex/resident/main.go) accepts
  context supplied by its caller and starts a thread for an invocation.

These mechanisms are starting points. They do not already implement the complete
observation-passing or artifact-inspection behavior proposed here.

## Implementation gaps and problems

### Checked observation passing

Contracts currently support ordinary JSON types, not a checked `observation`
type. Result stamps exist, but the general observation register and directional
ownership checks in [21](../docs/next/21-observation-passing.md) are still proposed.

The implementation must explicitly validate source accessibility and provenance.
A structural string check on a hash is insufficient. A first implementation may
use the bounded syscall evidence path, but must state its scope rather than claim
to implement the entire observation-passing proposal.

### Structured human invocation

The user actor currently knows only `think` and sends turns using the
`text`, `context`, and `images` shape. The UI ingress likewise submits text and
images. Add a structured invocation path for `create-spec`; registering another
root alone does not make that invocation work.

### Structured artifact transport

Conversation results currently become chat text. Add a result representation and
projection that preserve artifact identity, entries, citations, and placement
evidence through the webhost and UI protocol.

### Source boundaries across runs

Conversations may span multiple runs. Preserve original run and record identities
when carrying evidence into an artifact. Copied transcript text must not silently
acquire the provenance of the original observation.

### Missing source observations

Hidden model reasoning is not available source evidence. Native photo/web content
cannot be cited as directly observed source content unless captured as an
observation. An assistant's description of that content remains an assistant
assertion.

### Selection and context limits

Exact excerpts prevent uncited wording but do not guarantee that the model selects
every important requirement or qualification. Large source contexts need bounded
processing that retains original citations. Selection quality must be evaluated
separately from citation validity.

## Initial implementation sequence

1. Define the versioned artifact and evidence projection shapes, including source
   identity, ranges, attribution, placement evidence, and invocation boundary.
2. Build the deterministic evidence projection and validation leaves, with tests
   for forged references, invalid ranges, and uncited content.
3. Install the authored `create-spec` procedure and add structured invocation from
   the source user actor. Use a controlled model proposal step between projection
   and validation.
4. Add structured artifact transport and the citation/placement inspector to the
   human UI.
5. Verify reload and replay preserve artifact identity and source references;
   evaluate handling of corrections, conflicting statements, and unresolved
   acceptance on real conversations.

## Acceptance criteria

- A human explicitly invokes `create-spec` on selected conversation evidence.
- The procedure is visible through an outline derived from its installed bundle.
- The output is a structured artifact with a stable lineage identity.
- Every content word resolves to an exact cited source passage; fixed presentation
  text is accounted for by the versioned contract.
- Each placement exposes its evidence, with ambiguity left unresolved rather than
  presented as proven acceptance.
- Invalid or inaccessible references and uncited content are rejected.
- Reload and replay preserve the artifact and its source identities without
  rerunning inference to invent a replacement artifact.
- Creating or inspecting the artifact changes no model context and implies no
  approval or execution authorization.
