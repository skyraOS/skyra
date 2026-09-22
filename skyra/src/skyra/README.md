# Skyra runtime

The permanent base (originating from the DNA-19 increment; the disposable
generations that led here are in [`../../archive/throwaway-dna/`](../../archive/throwaway-dna/)).
Building on it, [`user/`](user/) is a root-invokable external executable that
serves a browser over WebSocket and drives an agent — the first slice of
[`../../specs/2026-09-22/frontend.md/proposal.md`](../../specs/2026-09-22/frontend.md/proposal.md)
(see the "Browser interaction" section below).

The host runtime and peers live under `root/`. `codex/` builds an
independent executable that exchanges JSON-framed Skyra Moves over stdio and owns
its private app-server process, conversations, event demultiplexing and effect
contamination fence. Shared `protocol/` contains wire vocabulary and declarative
validation only. No nested module or additional dependency is introduced.

From this directory:

```sh
go build -o artifacts/root ./root
go build -o artifacts/codex ./codex
go build -o artifacts/echo ./root/echo
go build -o artifacts/fake-provider ./codex/testdata/fake-provider
./demo.sh > artifacts/demo.jsonl 2> artifacts/demo.stderr
go vet ./... > artifacts/vet.txt 2>&1
go test -race -count=1 -v ./... > artifacts/acceptance.txt 2>&1
# Run the opt-in live suite using the configuration shown below.
```

Run the Codex demonstration alone:

```sh
./artifacts/root codex-demo ./artifacts/codex ./codex/profiles.json \
  ./artifacts/fake-provider ./artifacts/echo
```

It registers the shipped descriptor, connects through a correlated Move, enumerates
admitted profiles, starts a parent, submits another command on its existing thread,
starts an isolated second parent, resumes a tool child and rejects contamination.
`demo.sh` first runs the inherited echo, parent/child, cross-context and bounded
creation scenarios in another root process. Trace numbering is per root process.
The fixture's thread labels in result text are evidence, never authorization inputs.

The actor is launched by root with a private random parent-signing key and boot
namespace. Its argv selects the trusted provider executable and provider arguments;
its stdout is exclusively Moves and stderr is diagnostic output. The actual Codex
provider can be selected without changing Kernel or Scheduler. Raw provider commands,
thread IDs and turn IDs are absent from their code and the public descriptor.

`protocol.Args` builds observations from a selected profile. For example, health
requires `operation=health`, the opaque `connection` returned by connect, and
`deadline_ms=1000`. Planning code calls `ExecutionContext.Capabilities` for detached
admitted contracts. Memory retains the complete registered descriptor. The binding
still checks its narrower grant set on every dispatch. `tool-result` is declared
as continuation-only: only Bookkeeper's correlated child Return can activate it.
Trusted host provisioning calls `mintParent`; caller-supplied bare IDs do not mint
or authenticate conversations. `end-parent` explicitly retires an idle parent.

| Evidence | Tests |
|---|---|
| Real descriptor registration, persistent conversation, fork, tool replacement and resume | `TestCodexProcessAcceptance` |
| Interleaved parents while independent echo progresses | `interleaved independent parents and echo progress` subtest |
| Fault while the real actor waits on a deliberately withheld child; late Return rejected | `contamination faults while real binary waits on child` subtest |
| Startup cleanup after missing executable, descriptor mismatch, handshake error and timeout | `rollback` subtests, including endpoint/reservation/claim/process checks |
| Provider loss over healthy actor port and bounded health timeout | provider-exit and health-deadline subtests |
| Narrow binding, malformed arguments, foreign parent and old incarnation rejected before dispatch | `TestProfileAuthorityBeforeDispatch`, `TestHostParentOwnershipAndBusyFence` |
| Sticky contamination, exact tool correlation, replay rejection, old-thread fencing, transcript bound, effective policy mismatch | `codex/actor_test.go` |
| Descriptor compatibility, inventory and malformed observations | protocol tests and descriptor conformance test |
| All inherited DNA-18.5 acceptance scenarios | relocated boundary, contract, extension and demo tests |

See [DECISIONS.md](DECISIONS.md) for ownership, lifecycle, schema and migration
contracts, and [issues.md](issues.md) for limits.

## Opt-in live provider acceptance suite

The recorded live suite passed with the existing ChatGPT login and
`codex-cli 0.155.1`, using the configured `gpt-6-astra` model. It covers:

- Handshake, bounded health check and inference.
- An actual echo-tool invocation, distinct host child frame, correlated resume on
  the original provider thread/turn, and a subsequent command recalling the result.
- Native fork with inherited history; changing the branch's marker leaves the
  source conversation's marker unchanged.
- Tool removal and re-addition using two replacement provider threads, the same
  logical parent, and a retained random marker from the historical transcript.
  Echo is reported unavailable after removal; after re-addition it actually executes
  through a host child using the retained marker.
- Two overlapping commands on independent parents. Identical recall prompts return
  each parent's own random marker with the correct result identity.
- Nine distinct provider threads across seven parents; the migration parent changes
  threads twice. Child requests and resumes retain their provider thread and turn.

Evidence is in `artifacts/live.txt`, host Moves in `artifacts/live-trace.jsonl`, and
private correlation diagnostics in `artifacts/live-provider.stderr`. The suite enables
`SKYRA_CODEX_AUDIT=1` for metadata-only actor diagnostics; no parent tokens, signing
keys, prompts or tool contents enter that audit. No credential contents are printed.

To reproduce, configure credentials using the provider's normal mechanism, set an
absolute binary path, pin its exact version output and supply reviewed launch
arguments as JSON:

```sh
DNA19_LIVE=1 \
DNA19_PROVIDER=/absolute/path/to/codex \
DNA19_PROVIDER_VERSION='codex-cli 0.155.1' \
DNA19_PROVIDER_ARGS='["app-server"]' \
./live.sh > artifacts/live.txt 2>&1
```

The default deterministic suite skips live work unless opted in. The actor requests
`approvalPolicy=never`, `sandbox=read-only`, and only granted dynamic tools; it
checks the returned approval/sandbox policy before starting a turn. These observed
live scenarios do not prove universal provider sealing or OS containment. Native
contamination and adversarial stale/replayed events are tested deterministically,
not provoked through real effect-bearing tools.

Fresh experimental schema inventory and hashes are in
`artifacts/provider-schema.json`. The [official app-server documentation](https://learn.chatgpt.com/docs/app-server)
describes the initialization and experimental dynamic-tool boundary used here.

## Browser interaction (`user/` + `frontend/`)

The [`user/`](user/) executable is a root-registered external peer (structurally
like `codex/`): it selects between framed Moves from the host and browser
messages from its private WebSocket listener. Root opens a persistent
interaction U with an `interact` Invoke; each browser message becomes a child
Invoke to the agent, and the agent's Return resumes U (U stays alive across
messages and settles only on an explicit close). This first slice drives the
`echo` agent, so the full loop runs with no provider:

```text
browser (WS) → user → root → echo → Return → user → browser (WS)
```

The browser contract (`submit`/`close-interaction` up; `snapshot`/`accepted`/
`rejected` down) enforces: `accepted` ≠ completion, one pending command per U
(second is `busy`), a request ledger (idempotent replay, `conflict` on reuse),
and a fault that retires the anchor. The `frontend/` client speaks this contract.

The agent is selectable. With `echo` the loop runs with no provider. With
`codex` a browser message drives a real Codex conversation through the isolation
fence — the host connects Codex and mints U's parent, and the user relays each
message (`start-thread`, then `submit-input`) without minting or forging either;
Codex's tool-child routing and contamination fence are unchanged:

```text
browser (WS) → user → root → codex → provider
                        \→ echo (Codex's tool child)
```

Run the echo stack, or the codex stack with the deterministic fake provider,
then `cd frontend && npm run dev` (default WS URL `ws://localhost:8080/ws`):

```sh
go build -o artifacts/root ./root
go build -o artifacts/user ./user
go build -o artifacts/echo ./root/echo
go build -o artifacts/codex ./codex
go build -o artifacts/fake-provider ./codex/testdata/fake-provider

# echo agent (no provider):
./artifacts/root user-demo ./artifacts/user ./artifacts/echo 127.0.0.1:8080

# codex agent (fake provider; swap in the real codex app-server for a live model):
./artifacts/root user-codex-demo ./artifacts/user ./artifacts/codex \
  ./codex/profiles.json ./artifacts/fake-provider ./artifacts/echo 127.0.0.1:8080
```

| Evidence | Test |
|---|---|
| Full mediated browser→echo loop over a real WebSocket; U terminal Return on close | `root.TestUserBrowserAgentLoop` |
| Browser message drives a real Codex thread (start-thread then submit-input) through the fence | `root.TestUserCodexBrowserLoop` |
| accepted≠completion, one-pending `busy`, ledger idempotency/`conflict`, fault retirement, not-ready/close | `user` engine unit tests |

Not yet done: U-anchored Bookkeeper provisioning as first-class runtime-filled
activation arguments (here the trusted host packs the minted connection/parent
into the `interact` args), a versioned `user` `profiles.json` descriptor, and an
opt-in live run against the real Codex app-server.
