# Boot is genesis, then `skyra setup`

Status: idea, not started. Written from a code read on 2026-10-09.

## The idea

Boot does one thing: it commits genesis and emits a seed to `setup`, then gets out of the way. At that point a Horizon already exists, so everything after genesis runs as ordinary Horizon execution: recorded, sandboxed, replayable. Setup is the skyra binary running itself as a target (`skyra setup`). It reads a predefined manifest, emits the installs (models, adapters, procedure bundles) as invocations, and then hands off to the user experience.

```
skyra boot            genesis → seed {to: setup, manifest, then} → exit
  skyra setup         a continuation under the root Horizon
    round 1           install / admit what the manifest names    (fan-out, one spawn each)
    gather            every install's result returns to setup
    round 2           seed the user experience: `then`
  setup's reply discharges frame 0 (the operator's frame)
```

Three stages:

1. **Genesis.** Only what a Horizon needs to run: the syscalls, `setup`, and the room's ceiling.
2. **Setup.** Reads the manifest and installs what it names (codex, claude, interpreter bundles, tools).
3. **User experience.** What the user sees when it starts: the chosen procedure (think, design…), the model, and the first prompt.

## Why: what boot does today

`controller/runtime/boot/boot.go` `Run`:

1. `tree.Init`
2. `registry.Load`: parses every contract strictly, reads every bundle, reads and hashes **every binary**, including the large Codex native binary.
3. A provisioning loop (`boot.go:58–84`). For each stateful actor with a bundle, it re-reads the bundle and binary, re-checks hashes that `Load` computed moments earlier, then runs `<binary> install-bundle <state> <bundle>`. This runs **unsandboxed, before genesis**, so nothing in the lineage records it.
4. Genesis (frame 0, author `operator`, with a nonce).
5. `instance.Drive` with the seed. With `-detach` it returns here, which is how the webhost uses it (`hosts/web/main.go:704`).

What goes wrong:

- **The full boot cost is paid every conversation turn.** The webhost runs `skyra boot … -detach` for every conversation and every continuation run.
- **Provisioning changes `.skyra/<actor>/` before any record exists**, so replay can't see it.
- **The room is assembled in two languages.** A ~170-line Python heredoc in `skyra` (`registry()`) builds the registry and rewrites the `think` bundle's `scope.narrow.reach` up to five times. Go validates the result.
- **Install failures surface badly.** They are boot errors ("could not start the conversation") or `die`s in shell, not faults the UI can explain.

## Why it fits the existing machinery

- **Skyra already runs itself** for `instance`, `hold` and `resume`. Target binaries are the one thing it launches directly.
- **The machine already has a read → event → decide loop.** Effects like `GatherWaiter → Settled` and `ReadStrand → StrandRead` read, and `Commit`/`StoreInput` write. Nothing new is needed for setup to read state and decide.
- **Fan-out** (one emission with N instructions, one round, gather) already exists. Each instruction is a spawned `skyra instance`, which gives "n spawns of the runtime to write the contracts".
- **Continuations already resume after a gather** (the agent: "calls echo, waits, resumes").
- **`remember` already installs procedures and then calls `extend-reach` once per contract.** Bundle installation can take the same path, so `registry.Load` would stop parsing bundles and adding their contracts up front.
- **The operator can't be the stage driver.** It has no binary to resume: `machine.go:366` sets `s.Phase = Done // the operator's continuation`. So the stages need a continuation actor, and that's setup.

## Setup's properties

- **Almost no authority.** Setup doesn't hash, launch or write. It reads the manifest from its input and emits instructions, and the kernel does every real step. It can be sandboxed with no read/write/connect grants.
- **No new trusted binary.** It's the pinned skyra binary in mode `setup`, registered in genesis like any other surface.
- **Installs run in the actor's own sandbox,** writing only its own state (`.skyra/<actor>/`), as invocations recorded in the lineage.
- **`install` should be reachable only by setup.** Setup narrows it out before emitting `then`, so no later agent can reinstall anything.

## Design rules to keep

- **`machine.Step` stays pure.** "Is it installed?" is a read effect whose answer returns as an event.
- **The read has to be of the tree, not of mutable actor state.** `replay/strand.go`: "records are immutable, so the same walk gives the same strand." Installs are records, and "already installed" means an install record exists in the room.
- **The ceiling.** `room.go:36–39`: a narrowing may only drop what is held, and an extension adds a contract, never authority. Below genesis, nothing can add an actor or a grant.

## Open decisions

1. **What genesis holds.**
   - A. Genesis keeps the full registry, as now, and setup only provisions. No kernel change. **Ship this first.**
   - B. Genesis is minimal and pins the manifest's sha as its ceiling. Setup adds entries through a new `admit` syscall, which the kernel accepts only if the entry is in the pinned manifest, its binary hashes match, and its contract checks out. The invariant becomes "nothing runs that genesis didn't authorize": genesis authorizes the *set*, setup picks from it.
2. **Once per launch or once per conversation.**
   - Installing models is per machine: the Codex version check, the app-server proxy, the claude login check.
   - Per conversation means setup repeats every turn, in parallel but still repeated.
   - Per launch means setup is its own lineage, and each conversation's genesis cites its head stamp as the ceiling. That's verifiable through the pinned hashes, but it contradicts "two boots are two histories, never one" (`boot.go:89`), so it needs a new idea: a conversation rooted in a recorded setup.
   - Leaning per launch: "starting skyra" happens once, "new chat" is instant.
3. **Setup as Go code or as an authored procedure.** A procedure needs the interpreter, and installing the interpreter is setup's job. Either setup is Go code inside `skyra setup` (start here), or its first fixed step installs the interpreter's bundle and then hands the rest of the script to it.
4. **How bundle bytes reach a sandboxed installer.** Today `install-bundle` reads by path, unsandboxed. Either put the bytes in the install payload (checked against the pinned sha at intake; watch `maxEmission`), or grant read on `pinned/`.
5. **The extension rule.** "An actor cannot extend itself with what it holds" means a bundle contract mustn't also be declared. The primary contract stays as declared; only the extras come from extension (today's workaround: `registry.go:127`).
6. **Fan-out races.** Two concurrent first invokes of the same actor must not corrupt its store. Check whether `proc.Store.Put` is atomic (temp file plus rename).

## Manifest vs. user choice

- **The manifest (stage 2) is what can run.** That's the operator's decision, and it replaces the Python registry builder with a declarative file.
- **Stage 3 is what this user runs:** model, procedure, first prompt. The user can only narrow within the manifest, never widen. Narrowing already enforces this.
- **The manifest names models but never carries credentials.** Admitting codex checks that a login exists without bringing `~/.codex` into the room.
- **Setup is recorded,** so the webhost's 200 ms lineage watch can show startup as it happens ("verifying codex…", "claude: not logged in") and turn failures into fixable prompts.

## Related: one launch trampoline

There are four separate launch sites today: `run` (`instance.go:313`), `startResident` (`instance.go:358`), `bootstrapCatalog` (`appserver.go:187`) and boot's `install-bundle`. Each does its own read, hash, `sandbox-exec`.

Proposed: `sandbox-exec -p PROFILE skyra target <sha> <path> [args]`. Inside the sandbox, skyra checks the hash, seals inherited file descriptors (`custody.SealInherited`), sets the environment, and calls `syscall.Exec`, replacing itself in the same process.

- The trampoline takes everything from argv and never reads `.skyra`.
- macOS has no `fexecve`, so the exec is still by path. The guarantee is that nothing in the room can write to `pinned/`.
- Once every launch is checked at the point of use, `registry.Load` could take digests from the `pinned/<digest>/` paths instead of re-hashing every binary on every boot.

## What gets deleted

- `boot.go:58–84`: the provisioning loop.
- The bundle parsing in `registry.Load`.
- Eventually, the bundle-rewriting parts of the Python `registry()` in `skyra`.

## First step

Start a worktree. Add `skyra setup` as a Go continuation that installs the `interpreter` bundle through an invocation, then seeds `then`. Change boot to seed `setup`. Check it against `controller/run.sh` (the interpreter scenario, plus every scenario's replay check) and the replay tests.
