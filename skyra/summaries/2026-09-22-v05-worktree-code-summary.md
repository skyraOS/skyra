# Skyra v.05: a code read of the `v.05` worktree

**Date:** 2026-09-22
**Reviewed build:** worktree `worktrees/skyra-v05`, branch `v.05` at `f3573fe`
("Align frontend with v.05 runtime snapshots"). The tree has no uncommitted changes.
**Scope:** I read all of the code: the Go runtime in `skyra-v.05/` (about 5.3k lines
across 38 files), the Next.js frontend in `frontend/` and `scripts/deploy.sh`. For
docs, I read only `README.md` and `bugs.md`. I skipped `archive/`, which holds the
v.01–v.04 generations.
**Health:** `go build`, `go vet` and `go test ./...` all pass. Only
`src/reality` has tests: `universe_test.go` covers the snapshot shape and
`ws_test.go` covers the WS component.

---

## 1. What v.05 is

The README describes v.05 as *"a universe simulator that happens to use LLMs as one of its base
cases… just a hashmap that calls itself."* In the code, it is one Go binary that:

1. reads a text **genome** (`genome.skyra`) that declares devices, components and
   beings,
2. builds a graph of objects that all implement one interface, and
3. runs a single synchronous loop that passes one mutable `*Relation` from being to
   being until a response comes back empty.

Each LLM being has two layers: an inner **Think** layer that loops over operators,
and an outer **Act** layer that must address exactly one peer as
`<peer>message</peer>`. Humans, Claude Code, CLI tools and a Zork process are
beings too, and they sit in the same routing graph.

## 2. The one interface

```go
type Reality interface {
    ID() string
    Create(r *Relation) Reality   // constructor-by-convention
    Realize(r *Relation) string   // do the thing / or export a snapshot
}
```

`Relation` (`relation.go`) carries everything:

| field | role |
|---|---|
| `ID` | target being (who this is addressed to) |
| `Origin` | sender |
| `ThreadID` | which thread the loop is in |
| `Impulse` | the message text |
| `Parsers map[string]func() string` | lazily rendered prompt sections ("the present") |
| `Realities map[string]Reality` | what is reachable: beings, `think:*` ops, memory, desk… |
| `Collecting` + `Exports` | a second mode: walk the tree and export snapshots instead of acting |

Every type branches on `r.Collecting`. In collecting mode it exports a snapshot
(`being:<name>`, `thread:<id>`, `exchange:<key>`, `node:<id>`, …). In live mode it
acts. `Universe.Realize(Collecting)` assembles those exports into one
`UniverseState` JSON document, which the WS broadcasts after every resolved hop.

**How prompts are built.** Each layer lists the parser keys it subscribes to, then
copies only those parsers into a fresh `Relation`. It adds its own `system` parser
and hands the result to the `Provider`. `Provider.derivePresent` concatenates the
parser outputs in map order (so the order is random) and appends `message:
<impulse>`. The prompt is whatever the parsers render at call time.

`Port` and `Panel` interfaces exist but nothing implements them.

## 3. Boot (`main.go`)

1. Load `../.env`, then fill in `OPENROUTER_API_KEY` from the macOS Keychain if
   it isn't set. Start `debug` logging, which wipes and recreates `logs/` on every
   run, and create `~/.skyra/beings/`.
2. Build `Levels` and `Exchange`, then `NewThread`, the router. It gets the think
   operators `recall, remember, skill, browse, search, plan`. `ActOps` is empty.
3. `bootstrap()` parses the genome one line at a time. Values come from `Extract(impulse,
   "~key")`, which runs up to the next `~` or `|`:
   - `device` → `MacOS` (just a component map)
   - `component` → `Terminal` (stdin), `Provider` (llm), `WS` (started on its port
     right away)
   - `being` → a type-specific constructor. Only `user` beings get `Access=true`,
     which means only users can open threads.
4. Every `Provider` is given `inference.Call` (OpenRouter chat completions,
   temperature 0.2, 4096 max tokens, 120s timeout).
5. `thread.OnResolve` rebuilds the universe snapshot and broadcasts it over WS.
6. The terminal reads the first line, `Impress("michael", line)` wraps it,
   and `universe.Realize(rel)` starts a loop that doesn't return while the
   conversation keeps going.

### The genome that ships

| being | type | notes |
|---|---|---|
| skyra | llm | hub. Peers with everyone, including zork |
| michael | user | entrypoints terminal, ws. Opens threads |
| louise | llm | Arrival-flavored identity |
| claude | agent | shells out to the `claude` CLI |
| builder | llm | `~operators bash`, so it gets a **real shell** |
| philosopher | llm | |
| zork | process | `dfrotz ../DATA/ZORK1.DAT` |

Every LLM being uses one OpenRouter component, `anthropic/claude-sonnet-4-5`.

## 4. The main loop: Thread → Exchange → Being

```
Universe.Realize
  └─ NewThread.Realize          loop { … }
       ├─ open/lookup Thread (access-gated), attach thread + levels parsers
       ├─ expose all beings + think:* ops on r.Realities
       ├─ Exchange.Realize
       │    ├─ if r.ID empty → peel first word as target
       │    ├─ find/create Conversation keyed "a:b" (sorted)
       │    ├─ optional <ref>peer:START-END</ref> → copy context across exchanges
       │    ├─ append entry, Levels.Award(+5 xp each side)
       │    ├─ >20 entries → compress oldest into a "trace" memory, keep 10
       │    ├─ attach exchange / conversation(last 10) / ref-context parsers
       │    └─ Process? call directly and bounce back : being.Realize(r)
       ├─ Thread.Spread(from,to)  → append-only edge graph
       ├─ OnResolve()             → broadcast snapshot
       └─ reset parsers, continue (r.ID/r.Origin now point at the next hop)
```

When a being answers, it rewrites `r.ID`, `r.Origin` and `r.Impulse`, and the loop
routes the answer to whoever it named. The loop only stops when a hop returns `""`.

**Thread commands.** These are checked only when `r.ID == ""`:
- `being|grow ~name … ~type … ~devices …` → `Grow()` creates a being at runtime
  (llm/user/cli/agent only) and adds it to the relationship lists of the peers
  it names.
- `accept|reject <being> <task>` → moves a task on that being's desk out of `review`.

**The ref gate** (`<ref>`) is still parsed, but the code that required it is
commented out (commit `7829db3`, "disable ref enforcement"). The thread still has
an error-routing path for `r.Realities["error"]`, but nothing sets that error now.
`bugs.md` records the design reasoning.

## 5. Being types

| type | Realize does |
|---|---|
| `Self` (llm) | two-layer cognition. See §6 |
| `User` | prints the impulse and blocks on device input. `MacOS` sends it to WS if a browser is connected, otherwise to the terminal (`;;` starts multi-line input). **The reply is automatically addressed back to whoever spoke** (`r.ID = r.Origin`) |
| `Agent` | `claude -p <impulse> --output-format json`, then `--resume <session_id>` on later calls |
| `CLI` | `exec(entrypoint, impulse)`, returns stdout |
| `Process` | a long-lived child process (Zork). Writes a line to stdin, then drains stdout with a 500ms settle and a 2s cap. Exchange calls it directly and sends the output back to the caller |

`Being` is the shared value type: identity, purpose, relationships, entrypoints,
and a home dir at `~/.skyra/beings/<name>`. It renders a "peers you can address"
list, but **routing doesn't enforce it**. Exchange resolves any key in
`r.Realities`.

Every call is preceded by `Preamble`: *"You survive through your relationships. If you
lose all of them, you end…"*

## 6. Self: the Think/Act layers

`Self.Create` wires `being`, `desk`, `memory` (loaded from disk), `context`,
`think` and `act`. On `Realize` it:

1. puts being/memory/context/desk on the shared relation,
2. calls `Context.Heat(origin)`: it loads the memories linked to the entities of that
   relationship and attaches `memory-context` (understandings + tensions only),
3. runs **Think**, then **Act**. If Act returns a think-back, it loops to Think again.

**Think** (`think.go`) is private and has a budget of 5 passes:
- The present includes desk (scoped), exchange, conversation, ref-context, thread,
  memory-context, levels, the operator list, time pressure ("time remaining: 1. You
  must emit <surface-thought> now"), up to 10 recent thoughts scoped by peer, and the
  current think session.
- Each pass, the model emits one thing:
  - an operator tag, which runs and whose result is appended;
  - a `<peer>` tag, which narrows the scope of desk and history;
  - a `<surface-thought>`, which ends Think.
- Outer-layer operators are blocked with a system message.
- The result goes into `History` (at most 10) and is handed to Act as `inner`.

**Act** (`act.go`) is public and gets 3 attempts:
- It must output `<target>message</target>`. If there is no tag, or the target is
  itself, it retries with a warning.
- `<think>…</think>` sends the thought back to Think, and Self re-enters the loop.
  **There is no limit on how many times this can happen.**
- On success it sets `r.ID = target`, `r.Impulse = message`, and Self sets
  `r.Origin = self`.

**Think operators:**

| op | backing |
|---|---|
| `recall` | `Context.Retrieve` → `Memory.QueryGraph` (entities → `mentions` edges → substring match → falls back to the 5 most recent). Sorted by weight, top 10 |
| `remember` | `Context.Store` → an **LLM curation pass** (clean / dedupe / classify / extract entities / store or discard) → `Memory.StoreArtifact` |
| `plan` | desk tasks: create (with parent/subtasks), complete→`review`, drop, open/close/focus |
| `skill` | reads `~/.skyra/beings/<name>/skills/<name>.md` |
| `browse` | Firecrawl scrape, truncated to 4k |
| `search` | Firecrawl search, top 5 |
| `bash` | `bash -c`, 30s timeout, no sandbox (only builder has it) |

## 7. Memory, desk, levels

- **Memory graph** (`memory.go`, `memgraph.go`): each being has its own graph of
  `entity:<relationship>:<name>` and `mem:<ts>` nodes, joined by `mentions` edges.
  Memory weight depends on artifact type: understanding 1.0 > tension 0.8 >
  salience 0.5 > trace 0.2. Entities come from `Extractor`, which matches known
  names plus capitalized words that aren't stopwords. `Resolver` merges near-duplicate
  names (Levenshtein distance under 3). The graph is saved to
  `~/.skyra/beings/<name>/memory/graph.json` on every store. Everything is
  partitioned by relationship, so a being's memory of michael and its memory of
  louise don't mix. A vector index exists, but `Embed` is never set, so it doesn't
  run.
- **Desk** (`desk.go`, `task.go`, `plan.go`): each being has a tree of tasks per
  relationship. State goes `active → review → done/dropped`, and a rejected task
  goes back to `active`. Only a user can accept or reject, through the thread
  command. There are view states for open, closed, focused and collapsed. The desk
  is **kept in memory only**.
- **Levels** (`levels.go`): +5 XP to both parties for every exchange entry, on
  thresholds 0/50/150/…/2250. Shown to the being as "your level: N".
- **Economics**: the type and its snapshot exist, but `main` never sets
  `Universe.Econ`, so the snapshot always has an empty `economics` map.

## 8. I/O surfaces

- **WS** (`ws.go`): `:8080/ws`, on `golang.org/x/net/websocket`.
  - Outbound messages are `{type:"universe", payload: UniverseState}` after every
    hop, and `{type:"impulse", payload:{from,content}}` when a being addresses the
    user.
  - Inbound messages are `{type:"input", payload:{content}}`, pushed onto an
    **unbuffered** inbox that only `User.Realize` reads.
- **Frontend** (`frontend/`): Next.js + Zustand + zod, with two views.
  - *User present* (`/`, `RelationStream`): the perspective being's exchanges,
    with an input box.
  - *Universe present* (`/map`, `RealityMap`): a list of beings, the reality-graph
    tree, and a detail panel for layers, level, memories and skills.
  - The store replaces all of its state on every snapshot (no deltas).
    `mock/server.mjs` sends canned v.05 state so the UI can be built without the
    runtime. Components are marked TRANSITIONAL, waiting on a "Phase 2 UI rework".
- **Deploy** (`scripts/deploy.sh`): reads the key from the Keychain, `git push`, writes
  `.env` on the Mac mini (`skyra@10.0.0.50`), then pull, `go build` and kill the old
  process there. It doesn't restart the runtime; you have to start it by hand.
- **Logs**: `logs/system.log` plus one file per layer per being
  (`logs/<being>/{self,inner,outer,memory,context,desk}.log`), full prompts included.



## 10. Where v.05 sits relative to main

v.05 is the **generation before** the current `skyra/src/skyra` DNA line on main.
Its ideas are relational and read like a simulation:
- beings that survive through relationships,
- private thought versus public act,
- XP, desks and task review,
- memory partitioned by relationship.

Its mechanics are a single mutable `Relation` handed around one synchronous loop. There is no
authority boundary: any being can address any reality, and one LLM being has a
shell. The DNA-series summaries (the ports and kernel work, then the external
binary boundary in DNA-19) read as the fix for exactly that: they keep the
being/relationship vocabulary and add ownership, isolation and a mediated edge.
These ideas from v.05 are worth carrying forward:
- **Think/Act split with think-back:** private reasoning is structurally separate
  from addressed speech.
- **Relationship-partitioned memory with a curator pass on write:** each memory
  belongs to a relationship, and an LLM cleans, dedupes, classifies and links it
  before it is stored.
- **Tasks go to review and the user accepts them:** a being can't mark its own
  work done.
- **One snapshot type the UI renders from:** every reality exports a snapshot, and
  the UI renders straight from the combined result.
