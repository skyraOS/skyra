# Proposal: a root-invokable user executable with a browser frontend

Date: 2026-09-22  
Status: proposed; not implemented. Builds on DNA-19 and the copied v0.5 frontend.

## Outcome

Root registers and invokes a separate `user` executable. That executable serves
its browser frontend, receives human input, invokes an agent as its child, and
renders the agent's Return. The user invocation remains alive between messages.

The frontend is the human-facing part of the user executable. Its WebSocket server
is private transport plumbing, just as the provider connection is private to Codex.
Root reaches the executable through a bound Skyra Move port.

```text
Browser ← HTTP / WebSocket → user executable ← Moves → host runtime
                                                        │
                                                      Moves
                                                        │
                                                  Codex executable
                                                        │
                                                    provider

Invocation ancestry maintained by Bookkeeper:

Root caller
  └─ U: persistent user interaction
       ├─ A1: agent command → Return resumes U
       │    └─ T1: tool child → Return resumes A1
       └─ A2: next agent command → Return resumes U
```

A1 and A2 are successive children, not simultaneously outstanding children. They
share the conversation anchored to U. A1's Return means the agent has returned to
the user; it does not terminate U or require a separate reply-routing service.

## Sources and retained direction

- [DNA-19 decisions](../../../src/skyra/skyra-dna-19/DECISIONS.md): external binary
  boundary, authenticated parents, command lifetimes, faults and provider isolation.
- [DNA-19 parent peer](../../../src/skyra/skyra-dna-19/root/parent/peer.go): child
  Invoke, serialized continuation and Bookkeeper-delivered child Return. Unlike this
  fixture, the user does not immediately Return after its child returns.
- [Copied frontend](../../../src/skyra/skyra-dna-19/frontend/README.md): conversation
  panels, snapshot store, schema validation, connection state and mock server.
- [Codex registration proposal](../../2026-09-21/proposal/codex-registration.md):
  one executable ID, declarative profiles, boot admission and private transports.

The current frontend is a copy of the v.05 worktree, not a working DNA-19 client.
Its `input` / `universe` / `impulse` messages and older domain types must be adapted.
This proposal does not change the accepted Move ownership boundaries.

## Program and directory shape

Keep the copied frontend in the latest DNA implementation directory. Add a sibling
`user/` program when this proposal is implemented:

```text
skyra-dna-19/
  frontend/                 browser source, schemas, mock and asset build
  user/
    main.go                 process entry, Move reader and event loop
    contracts.go            private interaction/UI declarations
    profiles.json           versioned executable descriptor
    browser.go              local HTTP/WebSocket transport
    interaction.go          input → child Invoke; child Return/Fault → UI
    web/                    generated browser assets for embedding
    ..._test.go
  protocol/                 shared Moves and declarative contracts
  root/                     boot, grants, frames and parent authority
  codex/                    provider implementation, unchanged boundary
```

Build the browser to static assets and embed them in the Go `user` binary. Confirm
static-export compatibility of the copied Next app during implementation; it currently
has no export configuration. Node is a build/development tool, not a required
production sidecar. Development may serve browser assets separately while using the
same user executable's WebSocket endpoint.

Root imports neither user nor Codex implementation packages. User imports neither
root nor Codex implementation packages. Its stdio carries only framed Moves;
diagnostics go to stderr. Root launches and reaps user; user owns its listener and
browser connections. A network port number is not a Skyra executable or port ID.

## Root invocation and boot

Register one runtime-assigned UserExecutableID with the chosen binary's descriptor.
Install its admitted profiles and its delegated child permissions before launch.
Registration itself neither starts a browser listener nor starts model inference.

Proposed profiles:

| Profile | Arguments | Return and lifetime |
|---|---|---|
| `open-ui` | Trusted listener-policy reference, descriptor compatibility digest | Finite `ui-ready` result containing the local URL and UI incarnation. Starts the listener; no conversation or inference. |
| `interact` | UI incarnation, authorized agent executable, provisioned connection reference, admitted agent profiles/tool set, runtime-provided conversation reference | Long-lived invocation U. Its sole terminal Return is `interaction-closed`; individual agent answers are child Returns that resume U. |
| `close-ui` | UI incarnation | Finite `ui-closed` result after all interactions have settled; reject while an interaction is live. |

Observation keys, encodings, required fields and result/fault alternatives must be
concrete in `profiles.json`. `operation` selects exactly one profile. Configuration
references resolve only to boot-provisioned settings; the browser cannot choose an
arbitrary listen address, executable path, executable grant or provider command.
A descriptor declares argument forms, never grants authority.

Boot sequence:

1. Register descriptors, reservations, bindings and grants for user and Codex.
2. Launch the binaries; initialize Codex through its ordinary `connect` invocation.
3. Invoke `open-ui` through the user binding and validate its correlated readiness
   result. Report the URL to the human; opening a browser is optional host UX.
4. Invoke `interact` as U, atomically provisioning its conversation anchor as
   described below. User publishes an input-ready browser snapshot only after it
   has accepted that activation. Browser connection alone is not readiness.
5. Continue running the scheduler while U awaits human input and child outcomes.

Listener failure, bad descriptor, bad readiness response or boot timeout unwinds
registrations, bindings and reservations, closes listeners and reaps started peers.
The URL/readiness Return settles only the `open-ui` command, not the later U.

## User parent and agent conversation identity

U is a real Bookkeeper invocation frame, not a browser tab ID or a UI-only session.
It is the parent of each agent command and the stable logical conversation anchor.
The provider thread remains private to Codex.

DNA-19 currently allocates a separate persistent-parent sequence through the
trusted `mintParent` helper. The proposed extension is to anchor that record to U:

- At U's frame commit, Bookkeeper creates the parent record using U's sequence in
  the existing context namespace. It retains a link to U and its authenticated
  user binding, plus the authorized Codex binding and generation.
- The existing signed ParentRef shape can remain: namespace, owner, sequence,
  provider-binding generation and token. Here its sequence identifies U's anchor.
- A trusted boot provisioning policy on this activation authorizes the attachment.
  Frame allocation, anchor creation and runtime-filled activation arguments commit
  together before dispatch. A caller-supplied `interact` name or descriptor cannot
  activate this policy or select a frame number. This is new host work, not a
  claim that DNA-19 already supports runtime-filled arguments.
- The resulting reference reaches user through its authenticated activation. It
  stays server-side; the browser receives an opaque UI conversation identifier.
- On an agent child Invoke, host admission checks both the reference and that the
  emitting live frame/binding is the authorized U. Possession of a copied reference
  by another executable or another user frame is insufficient.
- A1, A2 and tool children receive their own Bookkeeper frame IDs. Move.Identity is
  per-request correlation; none of those identities replaces U or the provider ID.

No new root-owned browser-session map substitutes for Bookkeeper. User may keep a
private map from browser conversation identifiers to its live interaction records;
that map is presentation/transport state, not canonical frame or authority state.

For the first proof, Root provisions one interaction U and one authorized browser
session. Dynamic creation of additional U frames from the UI is a later capability,
not an implied power of the user executable. The actor's design should allow multiple
interactions later, but the initial UI does not advertise unsupported creation/fork
controls. Existing Codex fork/tool profiles remain independently available to their
already-authorized callers.

## Input, child work and Return

For the first accepted human message, user emits a child Invoke targeting the
admitted agent's `start-thread` profile. Later messages use `submit-input` with the
same authenticated reference. The server constructs these observations; it does
not forward arbitrary browser-authored Moves.

Before emitting a child Invoke, user serializes its continuation into Move.State:
UI conversation/incarnation, accepted browser request ID, and the next interaction
phase. Bookkeeper retains and returns that continuation. Transcripts and drafts
are not smuggled into host frame authority.

```text
Human sends request R
  user validates R and its interaction state
  user emits Invoke(Source=U, Target=agent, State=continuation)
  Bookkeeper allocates A, records U → A, marks U waiting
  agent may invoke tool T; T's result resumes A
  agent emits Return(A)
  Bookkeeper settles A and resumes U with A's result + saved continuation
  user appends the agent response and enables the next human input
```

After a child result, U is running in the host's frame lifecycle and idle with
respect to human input in the user's private UI lifecycle. It holds no scheduler
claim or Go runtime turn while waiting. The user process's reader/event loop remains
available for Moves, browser input and shutdown. No host `Realize` call waits for a
human, a browser socket write or model completion.

The first UI permits one pending agent command per U, matching DNA-19's one-child
contract. Reject a second distinct input as `busy`; do not hide a queue that could
execute after the user closes the conversation. The composer shows pending state.

An agent Fault settles A and resumes U on the existing child-result path. User
renders failure rather than inventing a successful answer. A contamination/provider
fault that retires the anchor disables further agent input for that interaction;
U remains able to display the result and close. Do not silently mint a replacement
conversation to clear taint. Host fault contracts must make this disposition
explicit so the UI does not infer authority from free-form diagnostic prose.

## Browser contract and reconnect

Reuse the transport-wrapper and schema-validation pattern, with a versioned UI
protocol scoped to this authenticated browser session. It is a projection of user
interaction state, not the entire host universe.

Illustrative messages (not Skyra Move variants):

```text
browser → user
  submit { uiIncarnation, conversation, requestId, text }
  close-interaction { uiIncarnation, conversation, requestId }

user → browser
  snapshot { uiIncarnation, revision, conversations, requests }
  accepted { requestId }
  rejected { requestId, code, message }
```

A conversation snapshot contains ordered user/agent entries and explicit phases
such as ready, pending, failed and closed. Entries associate with request IDs.
Actual child Returns append final responses. A socket acknowledgement, activity
indicator or partial provider text never marks an agent command complete. Streaming
provider text is outside this first proof; DNA-19 currently exposes terminal results.

User retains a bounded request ledger for its process lifetime. The same request ID
and identical payload returns its recorded status without another Invoke; a reused
ID with different contents is rejected. Admission into this ledger precedes emitting
work. `accepted` means the user process accepted the request, not that the runtime
or provider completed it. Dispatch rejection must surface as a correlated failure,
not leave the browser indefinitely pending: expected host admission rejections need
a defined diagnostic outcome to user, separate from forging a child Return.

Reconnect authenticates the browser and immediately sends a snapshot from user.
It does not create U again or automatically replay uncertain commands. The client
reconciles requests against the ledger, and can resubmit the same ID within the
same live incarnation without duplicating work. Process restart invalidates the
incarnation and makes unresolved delivery explicitly unknown. Crash persistence and
exactly-once effects across restarts are not promised.

Bind to loopback by default. Use a locally issued bootstrap ticket exchanged for a
browser session; validate WebSocket origin and session ownership. Keep tickets out
of trace output and remove them from the visible URL after exchange. Multiple tabs
of the same authorized session can observe the same interaction; the request ledger
and one-pending-command rule arbitrate writes. A tab choosing another perspective
in the UI does not acquire another parent's authority.

## UI shape to retain

Retain the copied frontend's dark styling, conversation panels, message-entry
components, connection indicator, validation boundary, snapshot/view-state split
and mock-server workflow.

- A panel represents a persistent interaction U. Its composer targets the selected
  conversation, removing the old first-word target convention.
- Agent Returns become response entries. Faults and request status have explicit
  presentation. A Return does not close the panel.
- Browser drafts and selection remain local; accepted entries and request state
  come from the user executable's snapshots.
- Keep the conversation/inspector layout pattern. An initial inspector can show
  the interaction's own admitted capabilities and known command status. Detailed
  agent/tool frame trees require an authorized diagnostic projection; the user
  binding does not automatically receive its grandchildren's Moves or all runtime
  state. Do not manufacture tool activity from assistant text.

Replace the old Being/Universe/Exchange wire schemas and full-universe broadcast
with this scoped projection. Their visual components can remain reusable without
preserving the older routing or lifecycle semantics.

## Closure, failure and process lifetime

Browser disconnect leaves U and any admitted work alive. Results update the
server-side snapshot for reconnect; delivery to a browser is not part of settling
an agent command. Give each socket a bounded writer queue and close/resynchronize
slow clients so a browser cannot stall the Move loop.

Explicit interaction close is idle-only in the first proof. If a provider thread
exists, user first invokes its authorized `end-parent` as a child; after that child
returns, user emits its own terminal `Return(U, interaction-closed)` to Root. If no
thread was started, or the anchor is already retired, it can close without invoking
an invalid provider operation. Host settlement revokes the anchor alongside U.

A close request while work is pending returns `busy`; the UI must not present it as
successful cancellation. General per-command cancellation requires a separate
provider/host contract. Root shutdown remains authoritative: atomically invalidate
U, its anchor and outstanding descendants; close endpoints/listener/browser sockets;
reap processes. Late results cannot revive an interaction. Invalidation does not
undo effects already committed by children.

Unexpected executable/transport loss retains DNA-19's explicit runtime-fatal policy.
Do not claim browser reconnection also restores a crashed user or provider process.

## Required changes and acceptance

Implement as bounded stages within the latest DNA directory:

1. Add the separate user program, versioned profiles and root boot registration.
   Prove open-ui readiness, interact activation and final user Return using a
   deterministic browser client and agent peer.
2. Add Bookkeeper's U-anchored conversation provisioning and source-frame checks.
   Preserve existing ownership, reservation, rejection and continuation tests.
3. Adapt the copied frontend and mock to the scoped UI protocol; package static
   assets and use the selected conversation's composer.
4. Exercise the actual user and Codex binaries with the deterministic provider,
   then run an explicitly configured live browser-to-agent conversation.

Required evidence:

| Scenario | Assertion |
|---|---|
| Boot | Root invokes user through Moves; listener readiness is correlated; no inference before human input. |
| Parent lifetime | Two human messages produce A1/A2 with the same U/anchor; both agent Returns resume user; only explicit close returns U to Root. |
| Tool nesting | A → T → A → U settles correctly; a tool Return is never rendered as the agent's final answer. |
| Nonblocking | While a human is idle, a provider is waiting, or a browser is slow, unrelated eligible runtime work progresses. |
| Authority | Forged browser conversation IDs, copied parent references, ungranted targets/profiles and client-chosen frame IDs cannot dispatch work. |
| Admission failure | Expected host rejection becomes a correlated UI failure and releases the pending request; it cannot look like a successful child result. |
| Reconnect | Immediate current snapshot; duplicate request ID does not cause another invocation; conflicting payload is rejected. |
| Fault | Agent contamination renders one failure, retires its anchor and cannot be cleared by reconnect or a late child result. |
| Close/shutdown | Idle close retires the provider association before Return(U); busy close rejects; forced shutdown invalidates descendants and reaps peers. |
| Build/UI | Independent user binary serves built assets; UI validates messages and displays pending/returned/faulted states. |
| Regression/live | DNA-19 race/vet/process tests pass; live two-turn conversation and tool resume have captured traces, with limitations stated. |

This is a proposal for the executable and lifecycle boundary, not authorization to
quietly add a global session service, unrestricted browser-to-Kernel API, general
OS services, multi-user hosting or crash recovery. Those remain separate work.
