# Startup

Everything needed to get Skyra running on a fresh machine. **Keep this current:**
when anything about setup changes, update this file in the same commit.

**Last verified: 2026-09-26** on macOS 15.6 (Apple Silicon), Go 1.25.0,
Node 26.0.0, Codex CLI 0.157.1:

- `./install.sh` and `skyra doctor` pass.
- `skyra` with real Codex: a browser message made Codex call the `echo` tool,
  which ran as a separate host process and resumed on the same turn.
- `skyra chat --fake`, `skyra demo` and `skyra test` all pass.
- `skyra live`: all 6 scenarios pass.

## 1. Prerequisites

| Need | Why | Install |
|---|---|---|
| **macOS** | the sandbox is Seatbelt (`sandbox-exec`) | — |
| **Go 1.25+** | builds everything; standard library only, no other Go dependencies | `brew install go` |
| **Node 20+** | DNA-19's browser UI (Next.js) | `brew install node` |
| **Codex CLI**, logged in | only for `skyra` with a real model | `brew install codex`, then `codex login` (a ChatGPT login works) |

Without Codex, everything else still works: use `skyra chat --fake`.

## 2. Install (once)

```sh
git clone https://github.com/skyraOS/tony-stark-was-able-tobuild-this-in-a-cave.git skyra
cd skyra
./install.sh
```

`install.sh` installs the UI's npm dependencies, builds every binary into
`build/`, links the `skyra` command onto your `PATH` (into `~/.local/bin`, or
the first writable directory already on `PATH`), then runs `skyra doctor`. If it
prints an `export PATH=…` line, add that to `~/.zshrc` and open a new terminal.

## 3. Run

```sh
skyra
```

This builds, starts DNA-19 with real Codex, starts the UI, and opens
http://localhost:3000. The terminal shows the runtime's trace: every Move
admitted, committed and routed. Ctrl-C stops everything.

| Command | Does |
|---|---|
| `skyra` | chat with Codex in the browser, through DNA-19's capability fence |
| `skyra chat --fake` | the same against a deterministic fake Codex: no model, no login |
| `skyra chat --echo` | the same with the echo agent: no Codex at all |
| `skyra demo` | **the stateless controller**: three scenarios (the full chain, the sandbox probe, a misbehaving leaf), each ending in a replay check |
| `skyra test` | every Go test, for the controller and DNA-19 |
| `skyra live` | DNA-19's live suite against your installed Codex (about a minute; rewrites the evidence in `history/dna19/artifacts/`) |
| `skyra doctor` | check prerequisites and ports |
| `skyra help` | list the commands |

`skyra` always rebuilds first, so after a `git pull` just run it again.

## 4. What to try in the chat

- **"Call the echo tool with the text hi, then tell me what it returned."**
  Watch the trace: the tool call leaves Codex as a Move, the host runs `echo`
  as a separate process, and the result resumes Codex on the same turn.
- Ordinary questions work: it's a normal Codex conversation.
- Type `/fault` to see the fault path: the conversation is retired and cannot
  be revived.

**What it cannot do, by design.** Every Codex thread starts read-only, with no
network, `approvalPolicy=never`, and exactly one tool, `echo`. It cannot edit
files or run commands. If Codex tries anything outside that vocabulary (a shell
command, an approval request, an unknown event), the fence retires the whole
conversation with a `contamination:` fault. Restart `skyra` to get a new one.

## 5. Troubleshooting

| Symptom | Fix |
|---|---|
| `skyra: command not found` | open a new terminal; or add `export PATH="$HOME/.local/bin:$PATH"` to `~/.zshrc`; or run `./skyra` from the repo |
| `port 8080 is in use` / `port 3000 is in use` | stop the other process, or `SKYRA_WS_PORT=8090 SKYRA_UI_PORT=3001 skyra` |
| `codex is not logged in` | `codex login`, or use `skyra chat --fake` |
| the first message fails with `contamination:` | a newer Codex emits an event the fence doesn't know. Run `skyra live` to see which; the allowlist is in `history/dna19/codex/events.go` |
| the UI says `open` but stays on "waiting for the interaction to open…" | the UI is dropping the server's snapshots as invalid. Open the browser console and look for `schema_invalid` drops: the server and `frontend/lib/protocol/schemas.ts` disagree (fixed 2026-09-26 for empty lists sent as `null`) |
| the UI loads but never connects | the runtime isn't up: check the terminal for an error above the trace |
| the UI fails to start | read `build/ui.log`; rerun `./install.sh` to reinstall the npm dependencies |
| anything else | `skyra doctor` |

## 6. Keeping this current

Before telling anyone to run this:

1. `git pull`, then `skyra doctor` and `skyra test`.
2. `skyra demo`: all three scenarios end in `byte-identical`.
3. `skyra chat --fake`, then `skyra`: send the echo prompt above.
4. If Codex was upgraded: `skyra live`, then update the versions and date in
   **Last verified** above.
