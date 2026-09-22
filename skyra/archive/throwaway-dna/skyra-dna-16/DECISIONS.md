# DNA16 implementation decisions

## Status

W0–W7 in [`workorder.md`](workorder.md) are complete within the stated local
scope. The proof uses a fake App Server protocol peer; the generic process host is
tested separately with a real child process. No live inference or production OS
containment claim is made.

## W0 — transition contract

- `DriverInput` carries a detached `Move` and previous opaque `DriverState`.
- `DriverEmission` carries complete replacement state, optional raw driver output,
  post-commit `HostCommand`s, or an atomic failure.
- Controller binds source `InvocationID` and handoff sequence, normalizes raw
  output, validates all Moves, and produces `DriverTransition`.
- Scheduler atomically commits state with Moves or failure. It restores a cloned
  resident snapshot if any effect is rejected.
- Placement commits before `Driver.Ingress`. Host commands execute through the
  optional `CommandExecutor` only after the transition commits.

## W1–W3 — transport, state, and host

- `Transport` moves complete opaque frames and is not embedded in `Driver`.
- `JSONLTransport` serializes writes and owns framing; closure is the guaranteed
  way to release a blocked stream read.
- `Continuation.DriverState` is cloned at dispatch and snapshot boundaries.
- `ProcessHost` uses direct `exec`, an explicit environment, stdio transport,
  captured stderr, and deterministic close/wait ownership. Kernel composition
  supplies its policy.

## W4–W6 — Codex

- One `CodexDriver` owns a possibility-scoped transport and live routing cache.
- Thread, turn, phase, tool mapping, pending request, final text, and contamination
  live in versioned JSON inside `DriverState`; scheduler knows none of them.
- Focus compiles to opaque `cap_NNN` dynamic tool names.
- Dynamic tool calls become ordinary child Invokes. A child observation answers
  the stored App Server request while the parent invocation remains stable.
- Native effect-bearing and unknown items atomically persist contamination and
  fault the workflow. Successful final Return requires a clean completed turn.
- Recreating a host with live state faults unless safe reconciliation can be
  proven. DNA16 does not guess or silently start a replacement thread.

## W5 — measured launch admission

The local `codex-cli 0.154.0` was checked with `--strict-config`. The accepted
disable surface uses `features.view_image=false`; `tools.view_image=false` is not
recognized by this build. `CodexLaunchPolicy` requires an exact version, canonical
absolute paths, isolated `CODEX_HOME`, explicit environment, read-only sandbox,
never-approve policy, and strict native-feature overrides before starting App
Server. Its Node launcher receives only a kernel-declared search path containing
absolute canonical directories; no user environment is inherited.

The configuration keys were cross-checked against the
[official OpenAI configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference).

## Verification

From the `skyra` module with an isolated Go cache:

```text
go build ./src/skyra/skyra-dna-16/...       PASS
go vet ./src/skyra/skyra-dna-16/...         PASS
go test -race ./src/skyra/skyra-dna-16/...  PASS
```

The opt-in live test also launched the installed pinned Codex binary with the
hardened policy and completed the real `initialize` response → `initialized`
notification handshake. It performed no inference:

```text
SKYRA_LIVE_CODEX=1 go test -run TestLiveCodexAppServerHandshake \
  ./src/skyra/skyra-dna-16/...              PASS
```

A repository-wide `go test ./...` also reached and passed DNA16, DNA15, and the
active Skyra packages, but its overall result is `FAIL` because archived commands
still import historical packages that are no longer present. Those archive
failures predate and are outside DNA16.
