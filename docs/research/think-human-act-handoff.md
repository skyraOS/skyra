# Think → human → Act: implementation assessment

Date: 2026-11-04
Status: research and proposed implementation order. No runtime changes in this assessment.

The human owns a Mission and invokes Think. Think returns an immutable spec
observation to that human. The human passes that observation to Act, granting
the requested capabilities within the Mission's authority ceiling. Think and
Act are sibling invocations under the human's authority. Mission is the
product/work identity; the human's runtime adapter holds its execution position.

This is already the central flow of `docs/next/think.md` and `act.md`.
Those documents also propose read-only Think, write-only Act, automatic
compaction, and interrupts. Those additional policies are not all implied by
this handoff and must not be silently adopted as implementation requirements.

## What exists, and what is missing

| Area | Existing mechanism | Missing connection |
|---|---|---|
| Human parent | `controller/targets/user/main.go` invokes the selected conversation procedure and waits through the browser surface | A recorded human action that passes a received spec to an Act child |
| Think | `controller/apprentice/workflows/think/bundle.json` narrows a model and returns its answer | A spec-producing contract and a typed returned spec observation |
| Spec storage | `controller/hosts/web/work.go` stores editable text revisions in `work.json` | Runtime-authored provenance, citations, and a grant proposal bound to the same immutable version |
| Human Run | `handleWork` reads stored text but takes tools separately from `inbound.Reach` | A pass of an exact observation already held by the human, with its grant bound to that observation |
| Act tree | `controller/hosts/web/trees.go` boots another root using the operator registry | Invocation beneath the Mission's human position, or an explicitly verified cross-root transfer |
| Observations | `controller/apprentice/tree/core.go` has scoped passing, receive, provenance, and return rules | An import/edge from a returned Think observation into the human scope and then into Act |
| Tree start | `shell.Ask` takes a string goal and target names; `Started` takes an endowment shape and digest | An endowment that preserves external observation origin and its verified source edge |
| Enforcement | `machine.narrowing`, `boundary.Authority.Allows`, sandbox rendering, and replay | Provision the approved contracts and OS authority from the human's held ceiling before Act runs |

The current shell creates a fresh `RootObservation` when starting a tree.
Putting the spec's text into that start path would preserve its bytes, but
would not preserve the fact that Think produced it and the human passed it.
Likewise, a `specId` in webhost metadata is not a scoped observation reference.

There is also an identity distinction: `shell.binaryOf` currently derives
an internal identity from a target name. The approval path should pin actual
registry program identities and canonical contracts, using the kernel's
existing verification. A hash of a name alone must not stand in for the
identity of executable code receiving a grant.

## Recommended implementation order

1. Define a versioned spec contract: goal, proposed callable contracts and
   permitted resource scopes, expected outputs, and cited observations.
   Runtime identity must bind payload, originating invocation/return, and
   source observations. The model supplies data and local references; the
   runtime assigns and checks identities. A capability proposal is not a grant.
2. Add an explicit spec-generation operation to Think. Ordinary conversation
   continues as ordinary turns. Under the older directional-return design,
   a spec/compaction leaf authors the structured result, and Think returns it
   unchanged. Reuse the current bounded provider/tool protocol rather than
   trusting prose that claims to be a spec. Automatic threshold compaction
   is not necessary for the first working handoff.
3. Extend the human adapter's protocol with a structured pass-to-Act action.
   The human position must have received the referenced spec through its
   Think child. Resolve references from that scope, validate the source
   return, and commit the human action before dispatch. Keeping the handoff
   beneath this parent avoids inventing an unrestricted cross-root importer.
   Cold recovery must reconstruct this scope from committed edges/receipts.
4. Resolve the spec's requested capabilities against the human's grant.
   Refuse unsupported contracts or scopes outside that ceiling. Provision
   exactly the approved subset, with provider/session plumbing kept separate
   from task capabilities. Descendants only narrow. Do not broaden today's
   read-only tree allowlist merely because a model requested a writer.
5. Extend the tree entry and shell replay so the initial spec and any cited
   inputs retain their original observation identity and provenance. The
   current `Started → RootObservation` path must not erase that ancestry.
   Any change to the formal observation rules must be reflected in Lean,
   its generated vectors, and Go; interface renaming alone needs no change
   to the internal `KindAct` classification.
6. Replace the Mission's editable spec-as-truth with a projection of returned
   spec observations. The UI becomes Think / Spec / Act. Generate Spec
   produces a reviewable artifact. Run passes that artifact. Human edits or
   requested tool changes create a new authored revision; they do not mutate
   a Think-authored observation. The execution records the spec, human pass,
   effective grant, and result. Retry/reconnect uses the same action identity.

A first delivery can use today's macOS kernel and sandbox. It does not need
proposal 52's Linux process supervisor or all of proposal 53's information-flow
system. It does need a real, verifiable observation edge at this boundary;
merely adding source IDs to display metadata would leave the central claim
unenforced.

## Scope decisions to make explicit

- Current Think can write scratch and, in repository mode, use repository
  tools; the launcher extends its bundle accordingly. If Think is to be a
  planner that cannot perform the work before approval, those capabilities
  must leave Think's grant. This is a policy change, not a UI label change.
- The old Act proposal prohibits reads. The current tree demonstration grants
  `find`, `read`, and `echo`. This handoff can work with explicitly approved
  readers; strict write-only Act is a separate choice, especially for tests
  and verification tools that read files.
- An Act result or request for clarification can return to the human, who
  passes it into another Think turn. Full mid-execution interrupts and resume
  should be scheduled separately from the first complete handoff.
- Approval authorizes a bounded execution; it does not prove the proposed
  actions are correct or that the human detected every malicious instruction.

## Acceptance checks

- Think produces a typed spec but cannot trigger Act or grant itself tools.
- A human pass accepts only a spec received in that Mission's scope.
- Forged, foreign-Mission, malformed, or changed references are refused before
  effects. Two equal texts from different source invocations retain provenance.
- Changing tools, scopes, or goal after approval creates another observation
  and requires another human pass. Empty grants never fall back to ambient reach.
- Act receives only the approved spec, selected cited inputs, and its grant;
  it does not receive the entire Think transcript implicitly.
- Tool and OS grants cannot exceed the human's ceiling; child grants cannot
  exceed Act's. Attempted calls beyond either fail through the normal boundary.
- Replay reconstructs the generation, receipt, human pass, provisioning, and
  execution. A reconnect or crash does not automatically dispatch effects again.

## External cross-check

Cap'n Proto's primary documentation separates typed capability references from
arbitrary global identifiers, and describes permission being conveyed by an
actual reference transfer. That supports keeping a spec's proposed tool names
separate from the runtime act of granting held capabilities; adopting Cap'n
Proto itself is not required:
https://capnproto.org/rpc.html#security

W3C PROV-DM distinguishes an artifact, its generating activity, the responsible
agent, derivation, and later use. Those distinctions support recording the
Think-produced spec and the human authorization as separate events, while
preserving the spec's origin. PROV-DM is a conceptual cross-check, not an
execution or authorization mechanism:
https://www.w3.org/TR/prov-dm/
