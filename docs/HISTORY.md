# Skyra: a full history

*Compiled 2026-10-13 from the git history (754 commits, 2026-02-08 → `4c09817`)
and the project's own retrospective records, several of which have since been
deleted from the tree. Recovered sources are cited by the commit that holds them.*

```text
2026     FEB    MAR    APR    MAY    JUN    JUL    AUG    SEP    OCT
         │      │      │      │      │      │      │      │      │
Genesis  ██████████████                                              Feb 8 – Apr 4
Beings                 ████████████                                  Apr 8 – Jun 9
Cognition                        ████████████████████                May 28 – Aug 20
Runtime                                             ██████████       Aug 19 – Sep 26
Controller                                                    █████► Sep 23 – now

         voice          beings in      machine          the smallest   a bounded,
         appliance →    a world, one   cognition over   correct        auditable agent
         "cognitive OS" interface      a neutral world  execution seam runtime you can ship
```

Commits per month: Feb 40 · Mar 127 · Apr 48 · May 36 · Jun 47 · Jul 133 ·
Aug 111 · Sep 108 · Oct (to the 13th) 104.

## How to read this

- **Five eras.** Each one was a bigger rethink than a version bump: the
  hardware-first *Personal AI OS*; LLM *beings* living in nested worlds;
  machine *cognition* over a neutral world substrate; a run of throwaway
  *runtime* generations (DNA-7 → DNA-19); and the *stateless controller* that
  is on `main` today.
- **Two numbering schemes overlap.** A product line (`v.01` → `v.05`, then
  `skyra-v.1`) ran February to June. An engine line (`dna-v2` … `v14`, then
  `DNA-15` … `DNA-19`) started inside it and ran until the controller replaced
  it. Where the numbers seem to collide, they belong to different lines.
- **Dates are git author dates.** Much of the archived code was bulk-moved
  later, so file mtimes and path-scoped history collapse onto archive commits.
  The v12–v16 generations and the ctu-engine were squashed into a single
  snapshot commit (`40a0340`, 09-19). Those are dated from their own README and
  DECISIONS headers.
- **Recovered sources.** Most of the project's earlier retrospectives were
  removed in the 09-26 cleanup. They can still be read from history:
  - `git show cfec03f^:architecture-evolution-timeline.md`: the phase-by-phase
    log, Phases 0–34
  - `git show cfec03f^:skyra/historical-timeline.md`: the four-era timeline to
    09-08
  - `git show cfec03f^:captains-log/conclusions.md`: the comparator experiments
  - `git show 5283889 --stat`: the DNA-19-era summaries and the
    stateless-controller handoff

## The shape of the whole thing

One motion repeats: **build big, hit the wall, keep the lesson, throw away
the weight.** Genesis was about 24k lines, roughly 80% of them documentation.
The cognition era was more than 40k lines of tested Go. The runtime era shrank
that to a working core of about 890 lines, then grew it back into DNA-19, a
roughly 7k-line host that mediated separate processes. The controller then
rebuilt it again around one replayable function. Each collapse distilled what
the larger build had shown to be necessary. The `boundary-crossings.md` and
`claude-ideas.md` discipline (docs win over code; proposals are marked apart
from canon) turned each build into a probe meant to show where the interfaces
break.

---

## Era 0: Genesis, the Personal AI OS (Feb 8 – Apr 4)

### Feb 8–11: a voice appliance on three machines

The repo opens on 2026-02-08 with an initial commit and a test file. The next
day brings the first architecture overview and a full README (PR #1). Skyra
starts as a literal **Personal AI Assistant**: an always-on **voice** appliance
with project-centric memory, pinned to a three-machine layout:
- a Raspberry Pi for voice;
- a Mac mini as the **control plane**;
- a GPU box for reasoning.

"OpenClaw" orchestrates it over local Llama, Qwen and DeepSeek models. PRs
#2–#4 (Feb 9–11) add memory architecture, multi-device capability, voice auth,
mobile interaction and a TV node.

### Feb 15–28: services, the context injector, persona and the arc

On Feb 15, PRs #5–#11 land in one day:
- listener and control-plane scaffolding;
- context compression;
- event ingress and acknowledgement docs;
- a delegation estimator;
- task-formation docs;
- Mermaid diagram fixes;
- the first **persona, soul and system-prompt** layers.

Feb 17 enforces a first-person voice across the persona layers and adds a
consistency and reconciliation model for speculative responses from the Pi.
Feb 20 introduces the **Arc model** (PR #14). Feb 28 defines tool hydration,
the listener lifecycle and a job phase model, and adds `next-steps.md` and
`ideas.md`.

### Mar 1: the rebrand

In about twenty commits on one day, **Personal AI Assistant becomes Personal
AI OS.** Projects become **Agents**, and device daemons become **Shards**. The
Pi is a Shard, and the GPU stops being a separate tier because GPU machines
are Shards too. The architecture is reframed as "control plane + variable
Shards." Early drafts of the Shard model and a "distributed brain" doc appear.
The README's "Why I Built It" becomes "Why I'm Building It."

### Mar 4–8: capabilities, context engine, predictive memory, the whitepaper

- **Mar 4–5:** Shards are reframed around capabilities instead of named
  machines (PRs #20–#22). A `voice_event_v1` schema lands. A "major
  architecture overhaul" brings an Internal Router, an Estimator, an External
  Router and a Job Registry, plus a context engine and an adaptation design.
- **Mar 6:** importance vectors (`[global, regional]` with time boundaries), a
  long-term memory store, and a revision that makes domain agents the
  doorkeepers of a unified heap. A gaps register (G1–G21) tracks what is open.
- **Mar 7:** the `skyrad` universal daemon model, an execution loop, an object
  store redesign, and **predictive memory** with an observational store.
- **Mar 8:** the Skyra whitepaper.

### Mar 9–11: the cognitive OS, then OctaOS

**Mar 9, kernel v2.** A unified protocol, with skill, job and task primitives,
and Redis as the trust boundary. Memory becomes a property graph, two-tier
(observational vs committed), with node promotion. The **soul and personality
layers are removed.** Principles are written down: "your keys, your data, your
consequences" and "the model is a dependency, not a component."

**Mar 10:** a crypto protocol, a trust spectrum, and skills as first-class
objects. Skills are *closed for modification, open for extension*, so
`update_skill` is removed. Other decisions that day:
- the context window comes from the hardware rather than being hardcoded;
- API compatibility becomes a gateway plugin;
- the user decides on skill trust, and the system informs without blocking;
- access profiles are introduced.

**Mar 11:** the **OctaOS** sheet and a "proof-of-executable-reasoning"
framework, along with a migration to an `octos` runtime and a codified
"verification law."

### Mar 12 – Apr 4: runtime rewrite, actors and episodes, contracts

- A runtime rewrite checkpoint (Mar 18).
- The data model is rebuilt around retention and runtime cognition (Mar 24).
- A structural projection demo (Mar 26).
- Capability contracts and probing demos (Mar 27–30).
- Registration and onboarding docs, primitives and orchestrator nodes, and
  aligned node and capability contracts (Mar 30 – Apr 1).

In this period, cognition reorganizes into durable **actors** working inside
**episodes** that project **frames**. The primitive set becomes `act` /
`observe` / `recall` / `learn`. A **Jarvis** (user space) vs **Stark** (system
space) role split appears, which is the direct ancestor of today's
user/system boundary. "Big update" (Apr 4) bulk-archives the genesis tree.

**What Genesis left behind:** local-first and user-owned; hardware as an
extension of capability; memory as the compounding asset; a kernel/runtime
boundary; "a command is not trusted because a model emitted it."

---

## Era 1: Beings in a World (Apr 8 – Jun 9)

### v.03: the first running code (Apr 8–19)

The **first implementation** (Phase 13, Apr 8–9) is a Go runtime of
`Being`s with a Nature (identity and purpose), an `Impulse` wire protocol
(`skyra <being> <expression> | <source>: <reason>`), and per-peer exchange
stacks with a "perspective flip," so each side holds its own view of an
exchange. A Gemini runner gives it inference. The model's output *is* the wire
format.

- **Apr 10:** v.03 is refactored into `primitives`, `world` and **`metaxu`**
  ("the between"), the router that carries signals between beings, seeded
  from a `genome.skyra`.
- **Apr 12:** the commit "shes alive." Self-relationships, exponential backoff
  and local inference follow.
- **Apr 14–19:** a correction path, "getting the language stable" (Phase 14),
  and design sessions on relational debt, pattern beings, thread unification
  and execution surfaces.

### v.04: Logos (Apr 21–30)

One universal interface, `Entity.Relate`, with a closed ontology of five
categories: Entities, Beings, Operators, Invariants and Mediums. Naming
churned: Nature → Pathos, logos → entity, Parse → Impress.

- **Apr 23:** shell and exec mediums, and a Linux analogy doc.
- **Apr 24, Phase 16:** **Claude Code, OpenCode and Codex join as typed
  beings** over shell mediums, as peers rather than tool integrations. The
  bootstrap scaffolding is removed. The runtime boots, grows the genome and
  waits. The first multi-agent session logs are captured: a builder being and
  Skyra designed a "moment" trace schema together.
- **Apr 26:** a lens spec, a mock server, and fixtures from real sessions.

### v.05: Reality and recursive worlds (Apr 27 – Jun 9)

The ontology collapses to one interface, `Reality { ID; Create; Realize }`.
The README described it as "a hashmap that calls itself." Worlds nest:
universe → device (macOS) → components (terminal, WebSocket, LLM) → being →
LLM → ports. Each being has an inner **Think** plane and an outer **Act**
plane, and runs under a survival preamble ("you survive through your
relationships… each response costs something"). The development path:
- the reality interface and genome boot (Apr 28);
- `~ref` context crossing and the being "louise" (Apr 30);
- a stable communication layer and the browser bridge (May 2);
- memory, levels, the task desk and a `bash` operator (May 7);
- the v0.1 memory-graph spec and Phases 1–2: a provider rack, specialist
  promotion and a TUI (May 9–17).

**Housekeeping in this era:**
- the repo moves under the `skyraOS` GitHub org (PR #36);
- an Apache 2.0 license is added (May 7) and replaced the next day by
  **FSL-1.1-ALv2**;
- a deploy pipeline and keychain support arrive.

**The idea that pushed v.05 out:** the **unified graph** (May 19–21), an
"observation-collapse" architecture with quantum-formal math, layered strata
and a cognitive nervous system. v.1 then introduces a new cosmology and deletes
v.05 routing (May 25). v.05 is formally retired on Jun 9.

---

## Era 2: Cognition and the distributed world (May 28 – Aug 20)

The largest era by far, with more than 40k lines of Go across overlapping
lineages. The project stopped modeling personalities and tried to build actual
machine cognition. The defining split had three parts:
- **cognition:** agents that retain observations and predictions and own no
  topology;
- **a neutral transport:** it routes impulses and must not interpret them;
- **a world:** it owns topology and live state, and reports without grading.

### skyra-v.1: living DNA (May 28 – Jul 2)

**Building the substrate (May 28 – Jun 11):**
- one interface and one DNA, starting at 207 lines (May 28);
- DNA that deposits, decays, gets picked up and attenuates;
- tenant isolation: a Slice per being (May 31);
- a nervous system of three decay curves (Jun 1);
- motor as dispatch (Jun 3);
- six binding realities and three cell layers (Jun 5–6);
- an episodic three-layer inverted index (Jun 9);
- a kernel contract with the Perception Invariant, activation-physics fixes
  and goroutine-per-crossing concurrency (Jun 11).

**First numbers (Jun 14–15, Phase 25).** The comparator experiments
`.1001`–`.1003` used a cheap DeepSeek judge over hand-built context:
- **goal closure** scored 0.5 → 1.0 where cosine similarity sat flat at 0.47;
- **noise tolerance** found 5 signal cells among 59 distractors, with zero
  false positives;
- **scale:** verdicts held to about 200k tokens, but **complete recall
  degraded past about 100k**.

The captain's log recorded these as single samples over hand-built presents,
which validated the judge and not the machinery.

**The cognitive chain (Jun 21 – Jul 2):** working memory as a lit subgraph;
the sensory → MidLevel → other → self → motor chain; a
prediction-falsifiability invariant (Jun 29); the full loop closed (Jul 1);
and the **first embodied runs** in a maze world (Jul 2).

### dna-v2 → v6: the epistemic engine (Jul 3 – mid-Aug)

**v2** (Jul 3–14, about 9.7k LOC) moved to slice-based dispatch:
- a coordination board, and an orchestrator "membrane" as the mount;
- ported cognitive phases;
- edge-owned learning: activation record, bump, credit write-back;
- a cause-and-effect V0 (`CausalNode`, `NodeRegistry`).

On Jul 3, phase 1 was "VALIDATED on real Codex." Jul 13 wrote up "the first
live run, and what it cost," and two Substack drafts were written.

**v3** (Jul 14–21, about 10.1k LOC) was the epistemic rewrite. The edge became
a conditional probability with an EMA prior, and every `expect` carried a
`within`. The governing rule: *the model proposes, deterministic parsers
compare, only world-derived observations settle durable claims*. It produced
the first real retrieval curve (`.2001`). Its last idea was "interpretation
learns through the world, not through itself" (Jul 21).

**In parallel:**
- the **spacetime** substrate and a "void kernel" (Jul 5);
- a **distributed-agent world** with a 2D maze and an interactive 3D globe
  (Jul 13–19);
- a thesis proposal scaffold (Jul 21);
- the directory rename `skyra-v.1` → `skyra` (Jul 25).

**v4, v5, v6** (Jul 24 – Aug 12). The work moved through dated spec sets
(07-24, 08-03, 08-06, 08-09, 08-11, 08-12): engrams, possibilities, the
syscall surface, "no semantic embeddings, relatedness is the graph," and "there
is no control plane." **v5** built the 08-09 set (about 5.2k LOC) and recorded
"four repairs the model run forced." Aug 12 ran an **overnight unsupervised
work order** under strict rules ("do not remove OR RENAME"). Its outcome docs
covered device identity, WHO/WHAT/WHERE, what each host offers, and
`os-parallels.md`.

**Why the era was set aside.** The three-way split was architecturally right,
but over six generations the epistemic machinery (falsification paths,
calibration, fusion, credit walks, grammars) grew faster than the behavior it
bought. On **Aug 19**, the **CTU pivot** ("eight primitives over one transport
verb") kept the lesson and dropped the weight. The whole line was archived the
next day.

---

## Era 3: the runtime generations, DNA-7 → DNA-19 (Aug 19 – Sep 26)

### The CTU pivot (Aug 19)

The pivot was read out of hardware. An ISA is *trapped into*, not *interposed*:
the MMU checks every access, but the kernel only wrote the page tables. The
kernel is therefore the **authority**, not the **machine**. Out of this came
two units (the CTU executes, the fabric transports), one transport primitive
(**MOVE**, a custody transfer), and eight primitives in four groups.

### v7 → v14: throwaway probes (Aug 20 – Sep 8)

- **v7** (Aug 22): the first end-to-end MOVE, with trusted system space split
  from untrusted user space.
- **v8**: a possibility distributed across CTUs, with signed provenance.
- **v9**: the possibility as a **reorder buffer** over a register file.
- **v10**: can THINK acquire a capability securely without ever minting one?
- **v11** (Aug 24–26): Body vs Context, and a **cold path** vs a **hot path**,
  "OAuth + macaroons at the request level." A later code read found the
  offline verifier was a test stub that always returned `true`.
- **v12** (cold path, 660 LOC) and **v13** (loop probe): surfaces frozen
  before they were decided. For about ten days, little code ran.
- **v14** (Sep 8): **working.** An Invoke/Return loop with capability
  inheritance and source-bound returns. The core is about 890 lines, and it
  passes `go test -race` on the echo proof.

On Sep 1 the thesis was frozen for the first time, in
**`grips-and-invariants.md`**. Operating systems assume one actor per user, but
a user now runs N semi-autonomous agents. Skyra's principal abstraction is
therefore the **bounded, auditable agent**. Five observed grips collapse into
three invariants: compiled cognition, the plan as a typed contract, and
capability-bounded authority under one list (the **Boundary**).

### DNA-15 → DNA-19: ownership, then a real process boundary (Sep 8–22)

The question changed from "what is the machine?" to "who is the sole owner of
each piece of state?"

- **DNA-15:** kernel-owned registration and a transition loop under one
  mutex. Two real races were reproduced and fixed with regression tests.
- **DNA-16** (Sep 12): a long-lived Codex App Server hosted with opaque
  `DriverState`. Dynamic tool calls become child Invokes, and a native effect
  contaminates the session. It was tested against a fake peer only.
- **DNA-17** (Sep 19): "the collapse," with one door (`Interpret`), a sole
  writer, and concurrent dispatch workers. A Codex round-trip proof ran
  against a fake.
- **DNA-18** (about Sep 20): the echo becomes a **separate OS process**
  exchanging JSON Moves over pipes, with 11 acceptance scenarios.
- **DNA-18.5:** cross-context activation and capability-mediated creation.
- **DNA-19** (Sep 22): the untrusted provider moves into its own `codex`
  binary, and authority is an **HMAC-signed parent reference**. A `user/`
  executable connects a browser to the agent. **The opt-in live run against
  the real `codex app-server` passed.** It was promoted to the permanent base.

Each generation ended with the same gaps still open: persistence, crash
recovery, reliable delivery, backpressure, joins over many children, and OS
containment of the provider.

**DNA-19's own critique** pointed to the next design. Authority lived in Go
pointers and exported functions, so anything that could call a function could
enter the root without crossing a port. Containment also rested on Codex
*reporting* `sandbox=read-only`.

### The great cleanup (Sep 26)

The stateless controller design (Sep 23–25) declared that **truth is derived,
not stored**. On Sep 26 a series of `chore:` commits deleted:
- superseded framings, roadmaps and white papers;
- business and outreach drafts;
- archived code, experiments and the old frontend;
- the pre-DNA-19 packages, committed binaries, old docs and summaries.

The repo was flattened around the controller, and a new README stated the
three invariants. Seven months of archive left the tree, and git history still
holds all of it.

---

## Era 4: the stateless controller (Sep 23 – now)

The long-running host was replaced by **one `skyra` binary that runs once per
step.** Each run:
1. takes custody of one Move;
2. checks it and durably records it as a content-addressed record;
3. runs the target inside a macOS Seatbelt sandbox **generated from the
   Boundary**;
4. hands the output to a fresh successor, and exits.

The machine is the chain of records in `.skyra/`, not any process. The three
invariants:
1. **The emitter is never the checker.**
2. **Authority only narrows, and the OS enforces it.**
3. **Truth is derived, not stored.** Replay from genesis is the definition of
   correctness.

### Sep 26–29: the controller works end to end

- **Sep 26:** one-command setup and the `skyra` launcher.
- **Sep 27:**
  - `interpret` derives the state;
  - input contracts arrive;
  - dependencies and `resolve` give the first continuation;
  - an actor host and a stateless Codex adapter;
  - **real Codex through the controller** (`skyra ask`);
  - conversations with ingress by parking.

  The frontend moves to the repo root. The README is reframed around **tool
  acceptance fatigue** and self-extension.
- **Sep 28:**
  - Codex app-server custody, a resident adapter and an egress proxy;
  - a **TUI** with a full-screen "living sky";
  - a workspace `read` tool;
  - **Skyra.app**, a SwiftUI desktop client;
  - call lights and ledger lines across all three clients;
  - demo GIFs in the README.

  `history/` (the DNA-19 artifacts) is removed. The last commit holding DNA-19
  is `b243b03`. Controller packages are regrouped into core, runtime, targets,
  codex and hosts.

### Sep 30 – Oct 3: the room, routes, procedures, the website

- **The room** (next/24–29, Sep 30):
  - one authority per Boundary, shared by every call;
  - Boundaries committed and looked up by hash;
  - resume derived from the tree, and an authored genesis;
  - `system.narrow`, so references a caller receives join its reach.

  Docs move under `docs/`. Merged as PR #38 on Oct 1.
- **Section F** (next/30, Oct 1): a hash audit, then shapes, contents and
  diffs, then return types, a **Starlark route runner**, routing at
  discharge, a novelty interrupt, and "the log is the test suite." Output
  contracts become optional and are then removed.
- **The sky and the website** (Oct 2):
  - the sky rendered in Metal, with roots and write targets drawn as boundaries;
  - a **Blueprint site** that replays recorded runs in the visitor's browser;
  - deployment to the Mac mini over Tailscale Funnel, with a resume page.
- **Section H, procedures** (next/32, Oct 3):
  - H1 removes the route and type machinery from the kernel *two days after it
    was built*;
  - then come fan-out, the Boundary as the actor system, contracts by
    `(target, schema)`, `strand`, and per-actor state;
  - **`remember`** compiles a recorded strand into a procedure, which then runs
    without the model inside CEL refinements.

### Oct 4–9: think, Claude, delegation, distribution

- **Oct 4–6:**
  - accepted-frontier conversation reconstruction;
  - a full repository review with fixes (`review-2026-10-06.md`);
  - descriptors sealed against inheritance;
  - **`think`** as the first continuation procedure (proposal 38);
  - connections moved out of the kernel and onto their edges.

  Conversations now run through `think` instead of calling Codex directly.
  Every prompt Codex adds on its own is turned off. Image viewing is a
  deliberate single native hole, so pasted screenshots work. A `find` tool is
  added.
- **Oct 7:**
  - a **Claude provider** (Opus 5.5, an MCP bridge, sandbox profile,
    `--resume` sessions), with the sky tinted warm when Claude is selected;
  - a bounded ideas workspace;
  - repository and design actors;
  - **Homebrew distribution**, including unsigned releases;
  - **bounded delegation** for both adapters, with goal-only child context and
    a work-tool ceiling.
- **Oct 8–9:**
  - scoped adapter tools;
  - `BOUNDARY-CONCERNS-SPEC.md`, which proposes moving agent and delegation
    fields out of the kernel;
  - conversation cards with first-message excerpts;
  - `think` made into one procedure with a `model` parameter.

### Oct 11–13: measuring it

- **Oct 11:** a contract-owned lifetime policy spec, research notes and idea
  drafts.
- **Oct 13:** **controller overhead vs inference time**
  (`OVERHEAD-VS-INFERENCE.md`):
  - overhead is a **fixed cost per step** and does not grow with inference;
  - one model round trip costs about 250 ms on the fake path, mostly
    `F_FULLFSYNC`;
  - a tool round trip costs about 380 ms on real Codex through the resident;
  - Skyra was compared with native `codex exec` over 24 real runs;
  - a barrier-then-force durability policy is proposed;
  - replay re-verifies 10,000 records in about 0.2 s from memory.

### Where it stands

**Built:**
- the step protocol and custody handoff;
- hash-chained records checked by replay;
- per-call sandboxes rendered from the Boundary;
- narrowing, fan-out and procedures, including `remember`;
- Codex and Claude adapters and bounded delegation;
- browser, terminal and desktop clients;
- Homebrew packaging.

**Open, from `docs/summary.md`:**
- Provider authority is not yet uniformly derived from the Boundary. Native
  image viewing and web search are not recorded individually, and Claude's
  profile starts from `allow default`.
- The kernel still carries agent-specific fields.
- Application-level checkpoint and recovery is not wired up.
- A host restart does not restore live execution.
- The procedure language is a finite DAG.
- Think's context grows without compaction.
- The lifecycle and authority specs need to be reconciled into one protocol.

**Uncommitted at the time of writing:** a modified `skyra` launcher and a new
`controller/targets/models/` target.

---

## Throughlines

The names changed every few weeks. These ideas did not:

1. **The user/system split.** Jarvis/Stark (v.02) → Pathos/Logos (v.04) →
   Think/Act (v.05) → user/kernel → the human grants a Boundary once.
2. **The model is never its own judge.** "A command is not trusted because a
   model emitted it" (Feb–Mar) → only world-derived observations settle
   claims (dna-v3) → *the emitter is never the checker* (invariant 1).
3. **Authority is minted, never assumed.** Shards as capabilities → BOUND and
   MINT (v10) → HMAC parent refs (DNA-19) → *authority only narrows, and the
   OS enforces it* (invariant 2).
4. **Compile once, run many.** Predictive memory → engrams → cold/hot paths
   (v11) → `remember` and procedures that run without the model.
5. **One envelope, applied recursively.** metaxu's `Signal` → v.05's
   `Reality`/`Relation` → the CTU's `Move` → the controller's
   Invoke/Return/Fault.
6. **Records over state.** The observational store (Mar) → the episodic index
   (Jun) → "truth is derived, not stored" (invariant 3).

The method is a throughline too: **every generation was built to be thrown
away**, and the distillations were the intended output of the search, not
failures to recover from.
