# Real app-server OS sandbox and direct-tool probes

Current direct-tool validation: `python3 controller/experiments/appserver/tool_surface.py`
launches the real installed binary against a local HTTP capture server. It
asserts that the inference request contains exactly the direct `functions.echo`
function, with no `exec`/`wait`, other built-ins, or injected skills inventory.
No external model is called by this check. Evidence:
`/private/tmp/skyra-tools-8e83ggiq/request.json`. Direct-mode live runtime echo,
interruption and recovery passed in `/private/tmp/skyra-e2e-by7wqlkl`.

The launcher now selects `standard` tool mode in a protected copy of the
installed version's model catalog, clears model-driven extra tools, explicitly
disables agents, and disables bundled/host skills. Both the code-mode host
feature and its executable grant are removed. The earlier code-mode results
below describe the initial experiment, before this direct-mode integration.

Production integration now lives in `controller/codex/provider/` and the root
`skyra` launcher. The original probe below shares that profile renderer.
Run `python3 controller/experiments/appserver/integration.py` for deterministic
adapter/Drive/replay tests, or add `--live` to use the actual sandboxed Codex
launcher. The live suite tests OS denials, conversation, a real runtime echo
invocation, direct interruption of a real turn, and recovery through the
adapter. The deterministic suite tests the adapter's native-attempt detection,
interruption, horizon fault and recovery, without relying on model compliance.

Integration evidence (2026-09-27): deterministic run
`/private/tmp/skyra-e2e-k9fnws3d`, successful live run
`/private/tmp/skyra-e2e-etoku60u`, and actual shell-launch/reuse plus invalid-profile
failure checks in `/private/tmp/skyra-launch-1emz7pk2`. The Go controller and
DNA-19 Codex package tests also pass. A real model declined the native-command
prompt, so the live suite intentionally separates direct OS probes and real
interrupt RPC verification from deterministic adapter native-event testing.

`probe.py` launches the installed native Codex binary through the existing
`controller/hosts/actor`, with an external macOS Seatbelt profile. It creates a separate
host socket, private evidence directory, CODEX_HOME, workspace and scratch
directory. It does not change or restart the production Skyra service.

Run from the repository root, outside another Seatbelt sandbox:

```sh
python3 controller/experiments/appserver/probe.py
python3 controller/experiments/appserver/probe.py --live
```

Requires macOS, Go, Python 3, an installed native Codex binary, and an existing
file-based `~/.codex/auth.json` login. `--codex /absolute/native/binary` overrides
discovery. `--live` makes a real model request. The login is symlinked into the
temporary home with read permission only; cleanup removes that symlink. Auth
refresh and long-running credentials are not tested.

## Results

Verified on macOS 15.6 with Codex 0.157.1 on 2026-09-27:

- App-server initialization, an allowed in-process file read and a scratch
  write succeed. Writes to `config.toml` are denied.
- In-process reads outside the grant and writes to the test workspace fail
  with `Operation not permitted (os error 1)`.
- `process/spawn` cannot start `/bin/sh`. This endpoint does not rely on the
  thread's internal sandbox, so the external OS profile enforces the denial.
- `command/exec` with a read-only policy also cannot execute its command.
- A real model turn calls the registered `echo` dynamic tool, receives its
  response, and completes with `SKYRA_DYNAMIC_OK`.
- The thread retains `approvalPolicy: never` and `sandbox: readOnly`.
  `environments: []` is accepted but returned as `null`; do not interpret that
  as proof of native-tool removal. Managed requirements returned `null`.

The live tool response is supplied by this probe through the host protocol.
This verifies provider communication under Seatbelt; it is not a production
adapter/Drive/replay integration test.

Evidence is printed as `/private/tmp/skyra-os-...`. Each run saves `actor.sb`,
`report.json`, `host.log`, and the host's protocol journals in that private
directory. Keep raw evidence local. The final successful run is
`/private/tmp/skyra-os-fyvs8eh0/report.json` (12 passing checks).

## Grants discovered and limits

The native executable avoids needing the npm shim, Node, or a shell. The initial
experiment required `codex-code-mode-host` while the model still used code
mode. Selecting standard tool mode removes that dependency; the current
profile does not grant execution to the helper, shells, or `sandbox-exec`.

Startup required a literal read grant for `/` (not recursive access). Without
it the macOS dynamic loader aborted before app-server initialization. Inference
needed system networking and certificate dependencies, including the system
MDS database, user MDS cache and its writable lock, and the named
`com.apple.AppleDatabaseChanged` shared-memory object. The renderer lists the
exact grants. They are a working experimental set, not a proven minimum.

Outbound IP access remains broad, including local IP destinations. No egress
proxy or destination allowlist is implemented. Metadata and executable mapping
permissions are broad; several system directories are readable. The actor can
read the granted login and write its temporary home and scratch. This does not
prove absence of every escape or an authority-derived production ceiling.

The initial experiment established that external OS enforcement can be added
through today's host without waiting for resident custody. Launch/config
isolation and adapter interruption are now integrated. Egress restriction,
auth refresh, authority-derived grants, and resident custody remain unfinished.
Already running processes are not migrated; see `startup-doc.md` for restart.
