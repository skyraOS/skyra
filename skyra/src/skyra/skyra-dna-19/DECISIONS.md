# DNA-19 contracts decided before implementation

DNA-18.5 is copied, never edited. Its host-only peers move under root. Kernel,
RuntimeState, Memory, Active, Scheduler, ExecutionContext and Bookkeeper retain
ownership. Provider operations and demultiplexing move into the independent Codex
executable; neither executable imports the other's implementation.

## Schema and authority

Protocol and descriptor version 1 require exact compatibility. An operation UTF-8
observation selects exactly one declarative profile. Unique observation keys,
required fields, encodings, size limits and enumerated constraints are validated
before dispatch and again by Codex. Unknown fields fail closed. Memory stores the
full descriptor; a binding contains only admitted profiles. Planning enumeration
returns detached admitted contracts. Continuation profiles cannot be invoked.

## Parent lifetime

Bookkeeper owns persistent parents separately from command frames. Trusted host
provisioning mints a parent sequence in the same allocation domain as frames and
an HMAC authenticated reference containing random boot namespace, owning context,
sequence and binding generation. A private boot key is passed to the chosen trusted
binary through its environment. Neither a bare sequence nor a provider thread ID
confers authority. Host admission checks the registered parent and binding; Codex
checks its signature, namespace, context and incarnation. Command Return releases
only that command. One outstanding command per parent; independent parents overlap.
Fork requires a fresh authorized destination and an idle live source. end-parent
explicitly retires an idle parent; context retirement revokes all parents and work.
No restart persistence is implemented.

## Continuations, faults and isolation

Provider thread, turn and pending request identifiers stay private to Codex.
Bookkeeper allocates a distinct child frame, retains opaque continuation bytes and
returns the child result to the original command. This Return is the tool-result
continuation profile; callers cannot invoke it as a fresh command. Codex checks the
command, pending request, continuation nonce and active session before replying.
Fault can settle a running OR waiting frame atomically, recursively invalidating
its descendants. Late results cannot resume a faulted frame. Contamination retires
the whole conversation and faults once; later commands cannot clear it. Detection
does not establish prevention of a native effect. Unknown provider requests are
never approved. Unknown correlations are discarded, conflicting live correlations
fault the affected conversation. Provider thread separation remains trusted.

## Bounded tool migration

Only the boot-granted echo tool (target 1, operation echo) is grantable in this
proof. Add/remove is idle-only and starts a fresh provider thread with the complete
new tool set. Stable authenticated parent reference stays unchanged. Context retained
is the complete bounded transcript of inputs, final responses and tool results,
encoded as inert historical data in developer instructions, not replayed turns or
effects. Hidden reasoning/provider state is not migrated. Above 64 KiB history,
migration fails explicitly without truncation. The old mapping is invalidated only
after a successful replacement response; failure retains the old mapping. Removed
tools are no longer authorized; old-thread events are ignored. Fork uses the native
fork operation, with a separate mutable session record. Unsupported operations
return explicit faults; no fallback that silently loses context is allowed.

## Shutdown and failure

Root launches Codex; Codex launches and reaps the provider. Graceful pipe closure
lets Codex reap before exit; a process-group fallback handles a stuck actor/provider.
Startup rollback removes registration and reservations, invalidates endpoints and
waits for reaping. Connect is a normal correlated command and must finish before
ordinary workload. Health measures protocol responsiveness, not inference readiness.
A healthy actor reports provider loss as correlated Faults. Actor port loss remains
runtime-fatal. No committed request is retried. Timeouts guard operations, not test
ordering. Live inference requires explicit opt-in configuration; deterministic
provider results do not establish live compatibility or OS containment.

## Implementation and verification notes

The provider adapter uses initialize/initialized, thread/loaded/list, thread/start,
turn/start, thread/fork and item/tool/call. These names occur only inside Codex and
its fixture. Local CLI 0.155.1 experimental schema generation confirms dynamicTools
on thread/start and not on thread/fork; the recorded inventory is in artifacts.
Thread-start/fork responses must report approvalPolicy=never and sandbox readOnly
without network access before the actor dispatches inference. This is a separate
policy-report check, not evidence of complete native-tool containment.

Binding generation is authenticated inside the parent reference; owner-qualified
command frames still use the existing wire Context plus local FrameID. A child
Return carries the original command in Source, its separate child in Answering,
and Bookkeeper's stored continuation nonce in State. Codex does not use a caller
thread ID, OS PID or Move.Identity as its conversation identity. RPC IDs and call
IDs are checked separately and never replace either host identity.

Expected malformed input is rejected before a frame is committed. A failed provider
replacement leaves the old actor thread/tools installed and faults with unsupported;
Bookkeeper releases that command without retiring its parent. Contamination, timeout
and provider failure retire the conversation. Fork locks the source in both owners
until completion. Retired threads and commands have non-reused tombstones; stale
or unknown provider events cannot create work. Context shutdown invalidates every
live parent/frame; a fault while waiting atomically invalidates all descendants.

Final captured verification: go vet ./... and go test -race -count=1 -v ./...
passed; four independent builds and the real-process demo passed. Import dependency
evidence confirms neither executable imports the other. See artifacts/verification.json,
acceptance.txt, builds.json, import-boundary.json, demo.jsonl and live.txt. The initial live test was skipped; the subsequent explicitly authorized live run
passed handshake, health and one inference (see live.txt). Sandbox Go-cache denials were rerun with approved
cache access; no test failure was suppressed.

## Live validation follow-up

The existing ChatGPT login was usable; the original DNA19_LIVE gate was a test
opt-in, not evidence of unavailable credentials. The authorized follow-up used
/opt/homebrew/bin/codex, version codex-cli 0.155.1, argv ["app-server"], with the
existing configured model gpt-6-astra. It checked returned read-only/no-network and
never-approve thread policy, then received the final response "ready". Host Move
trace and test output are captured in live-trace.jsonl and live.txt.

The real provider sends mcpServer/startupStatus/updated with a thread ID but no turn
ID. Its locally generated ServerNotification schema classifies it as administrative
status (starting/ready/failed/cancelled). The adapter validates that shape and ignores
it for settlement. A request disguised with that method still faults and never
receives approval; MCP tool-call effects remain forbidden. A regression test covers
both cases. The version check now reads stdout separately from stderr diagnostics.
The initial smoke covered only inference; the expanded suite below supersedes
that evidence limit.

## Expanded live acceptance

The live suite now runs actual dynamic-tool work, native fork, tool-set removal and
addition with replacement threads, and concurrent-parent recall. It asserts real
Bookkeeper child frames/results, unchanged thread/turn on the tool resume, source
history after branch mutation, retained transcript data through replacements, and
correct random-marker result routing while both parent commands are outstanding.

Opt-in SKYRA_CODEX_AUDIT=1 writes private binding/turn/tool correlation metadata to
stderr. It is evidence only, never an authority channel; parent capability tokens,
keys, prompts and tool contents are excluded. Seven logical parents produced nine
distinct provider threads; the migration parent received exactly two replacements
with echo absent and then restored. The audit validates the old/new binding chain
and the tool request/resume correlation. Root's operational protocol still contains
only Skyra Moves; only the live test reads these diagnostics.

The recorded run uses the existing ChatGPT login, CLI 0.155.1 and the configured
model. These are finite observed isolation and migration scenarios. Full provider
sealing, hidden-state migration and native-effect prevention remain unproved.
