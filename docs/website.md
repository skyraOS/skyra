# Website: a personal site that proves its own claims

Date: 2026-09-28  
Status: proposed, not built. The theme is chosen (Blueprint); the rest is the
plan. Mockups: [Personal Site Themes](https://claude.ai/artifact/NvPdPS2ovUh5gDYr7vzoj1),
board 4.

## What it is

A personal portfolio site whose centrepiece is Skyra itself. A visitor does not
read a claim about replay; they run it. The site loads a recorded conversation,
re-derives every record in the visitor's browser with the same `machine.Step`
that ran on the Mac, checks each record against its stamp, and arrives at the
head hash. Then the visitor changes one byte and watches the check fail.

It needs no inference. The site is static, so there is no public API, no
Codex login, no cost per visitor, and it can be hosted anywhere. The Mac is
only needed when a new run is recorded.

## The theme: Blueprint

A technical drawing. The page is a drafting sheet: a faint two-level grid, a
ruled border, dimension lines and an engineering title block. The idea is a
spec sheet for an agent runtime, and a quiet nod to this repo's name. It
should read as engineered, not themed: no glow, no gradients, no hand-drawn
effects.

### Palette

| Token | Hex | Use |
|---|---|---|
| `paper` | `#E9EDF0` | the sheet: page background, and fills behind boxes so the grid does not show through |
| `ink` | `#13212E` | text, the border, box outlines |
| `ink-2` | `#3A4A5A` | body copy |
| `muted` | `#4A5968` | labels, stamps, table headings |
| `line` | `#2B5C8A` | the accent: section labels, dimension lines, the head hash, links on hover |
| `grid-minor` | `#2B5C8A` at 7% | 16 px grid |
| `grid-major` | `#2B5C8A` at 14% | 80 px grid |

`muted` and `line` both clear 4.5:1 on `paper`. One accent only. A faulted
record, the one place a second colour is earned, uses `#B23A0C` and always
has a label, never colour alone.

A dark variant (a real blueprint: `#0E2233` ground, `#E6EEF4` ink, the grid in
light blue) is worth building for `prefers-color-scheme: dark`, but it is
second.

### Type

- **Archivo** (600–700) for the name, headlines and title-block headings. Tight
  tracking on large sizes (−0.02em).
- **JetBrains Mono** (400–500) for everything that is data or a label:
  stamps, record states, navigation, section labels (`SECTION A · THESIS`),
  title-block fields. Upper case with light tracking for labels.
- Body copy in Archivo 400 at 18 px, line height 1.55, measure about 600 px.

### Components

- **The sheet.** A 1.5 px `ink` border inset 24 px from the viewport edge,
  over the two-level grid. Every page has it.
- **Section labels.** Monospace, `line` colour, lettered like drawing sections:
  `SECTION A · THESIS`, `SECTION B · RECORDS`.
- **Dimension lines.** A thin `line` rule with end ticks and a monospace label
  (`HORIZON · granted once`). Use them to annotate, sparingly: one or two a
  page, and only where they mean something (the scope of a grant, a
  boundary).
- **The record strip.** Records as joined boxes: state, target, short stamp.
  This is the visual for a lineage, and on the Verify page each box ticks as
  replay reaches it.
- **The title block.** Bottom right, as on a drawing: `DWG NO.` (one per
  project: `SKY-001`), `REPLAY` (`7 / 7 identical`), `DRAWN BY`, and `HEAD`
  in the accent. On the Verify page its values are live.

### Voice

Plain and exact, like the rest of the repo. Claims say what is enforced and
what is only detected. No "revolutionary", no "AI-powered".

## Pages

| Page | What it holds |
|---|---|
| Home | Name, the thesis ("Grant the boundary once. Let the OS hold the line."), one paragraph on Skyra, the record strip and title block |
| Skyra | The problem, the three invariants, how a step works, where it is (the README's status table, kept honest) |
| Verify a run | The replay demo, below |
| Writing | `grips-and-invariants.md`, the design notes worth reading, and the research question |
| About | Short: who, what next, contact |

## Verify a run

```text
1. load a recorded conversation: its records and the input bytes they hash
2. re-derive every record in the browser with machine.Step (compiled to WebAssembly)
3. hash each one and compare it to its stamp, then to the head hash
4. let the visitor change one byte and watch the check fail
```

`core/machine` is pure and standard library only, so it should build with
`GOOS=js GOARCH=wasm` behind a small wrapper that feeds it records instead of
reading `.skyra/`. The point of running it client-side: the check is the
visitor's browser running the same code, not the server saying "trust me".

Runs worth recording:

1. An ordinary question, answered.
2. A prompt injection, faulted at intake.
3. A probe of `.skyra/` from inside a leaf, denied.
4. A call budget running out (once the call limit exists, below).

### What the hash does and does not prove

The records form a hash DAG: each record stores its predecessor's stamp and
the sha256 of the bytes it observed (`core/record/record.go`), and a resolve
names the discharges it waited on. The head hash therefore commits to every
record it reaches and every input.

It proves integrity: nothing changed, and every state was derived by this
function from these observations. It does not prove the observations came
from real Codex; a consistent tree can be built offline. Say so on the page,
and publish each head hash when its run is recorded (a git commit is enough),
which at least proves the record existed then and was not edited later.

### Later: a provenance root

Once [observation passing](next/21-observation-passing.md) lands, the answer's
observation ID becomes the root of a Merkle DAG of where its words came from,
not only of what ran. A visitor could click any part of an answer and walk down
to the leaf and file that authored it, recomputing each ID on the way. The
return rule is what makes this worth showing: a continuation cannot return text
it wrote, so everything traces to a leaf. Three open choices in next/21 decide
the shape: author identity must be a stamp replay rebuilds exactly; whether a
model that made no calls counts as a leaf; and how faults are represented.

## Optional: live mode

Later, and only if it earns its keep. Visitors ask a real question and get a
real, recorded answer.

```text
operator (the headless Mac)
 ├─ personal root   .skyra-personal/   your workspace, your tools
 └─ public root     .skyra-public/     read site/, reply, codex limited to 5
      └─ public host behind a tunnel, one lineage per visitor question
           └─ its own Codex process: same binary, its own home, a mini model
```

- **Sibling roots, not parent and child.** Their Horizons are disjoint, and
  each has its own record tree, so the public side cannot even list personal
  conversations.
- **The call limit lives on the root.** A reach entry gains a limit
  (`"codex": {"limit": 5}`); `interpret` faults the sixth call as a horizon
  fault (`"codex": 5 of 5 calls spent`). The count is the number of `codex`
  invokes in the root frame's own steps, so it is derived, not stored, and
  replay recounts it. A call is a visitor turn; to bound tool round trips
  inside a turn, limit `read` in the grant given to Codex as well. When a
  deeper frame is ever granted `codex`, `Narrow` must cap its limit at what
  the parent has left.
- **A new public host, not `hosts/web`.** One endpoint: ask, get the reply and
  its lineage. It is the only code facing the internet: input limits, rate
  limits, timeouts, and a daily cap on new conversations (a per-conversation
  limit cannot bound the bill across lineages).
- **Credentials.** Serving the public from a personal ChatGPT login likely
  breaks OpenAI's terms and spends your rate limits. Use an API key for the
  public root, with a hard monthly spend limit set in the OpenAI dashboard.
  Check the current terms before launch.
- **Containment.** Codex's native execution is detected and interrupted, not
  prevented. For the public root keep its work directory empty and egress
  limited to OpenAI hosts, and say plainly on the site which parts are
  enforced by the OS and which are only detected.
- **Operations.** Cloudflare Tunnel or Tailscale Funnel, no open ports; launchd
  to restart the host; sleep off; the Codex binary pinned (the provider only
  accepts 0.160.0).

## Order of work

1. A test that every record in a lineage is reachable from its head.
2. `machine` built to WebAssembly, replaying one recorded run in a page.
3. The static site in the Blueprint theme, with the Verify page and the four
   recorded runs.
4. The call limit on a root's reach, its fault and its replay test.
5. The provenance root, after observation passing.
6. Live mode, if still wanted.

## Open questions

- Are supervisor faults reachable from the head, or does the head need to name
  them?
- Which runs are the most convincing to a reader who has never seen Skyra?
- The dark variant: build it at launch, or after?
- Where to host the static site: GitHub Pages is probably enough, or does the
  site want its own domain from the start?
