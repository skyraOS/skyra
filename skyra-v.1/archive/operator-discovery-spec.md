# Operator Discovery Spec

## Premise

Think doesn't get a menu of action operators. It thinks freely. The correct operators appear in Act based on what was thought. This is Salience applied at the Think→Act boundary.

## Flow

```
Think (introspective ops only)
  → surfaces thought
    → Salience matches action operators to thought content
      → Act receives matched operators
        → executes OR passes back to Think
          → Think sees what Act had (one-line append)
            → adjusts, surfaces again
              → Act gets new matched set
```

## Think Layer

Think sees only introspective operators:
- recall, remember, skill, browse, search

Think's system prompt includes:
> "Think about what you want to do. The correct operators will appear when you act."

Think does NOT see action operators. It doesn't know what's available. It just thinks about intent.

## The Boundary (Salience)

When Think surfaces a thought, before Act fires:

1. Take the surfaced thought content
2. Run it against the registered ActOps (embeddings, keyword match, or both)
3. Return the top-N operators whose descriptions are semantically relevant to the thought
4. Attach only those to the Act layer for this pass

This is not retrieval in the traditional sense. Nothing was queried. The thought existed, and relevant capabilities surfaced because of proximity — like things coming to mind.

## Act Layer

Act receives:
- The surfaced thought (as `inner`)
- Only the operators Salience matched for this turn
- The outer parsers (peers, etc.)

Act either:
- Executes using the available operators → done
- Passes back to Think (think-back) → loop continues

## Think-Back Visibility

When Act passes back to Think, Think gets a one-line append in its exchange:

```
system: act's operators this turn: [plan, deploy, notify]
```

This gives Think awareness of what was available without giving it the ability to call those operators directly. Think can now adjust:
- "plan was available but I need something else — let me rethink what I actually want"
- "deploy was there, that's close — let me be more specific about what to deploy"

## Why This Works

1. No static menu. Beings don't memorize tool lists.
2. Intent drives capability. Think about deploying → deploy appears. Think about searching → search was already there in Think. The layers have different operator sets for a reason.
3. Discovery loop. If the wrong operators show up, the being naturally adjusts its thinking. No error state — just refinement.
4. Scales without cognitive load. 100 registered operators don't create a 100-item menu. Only 2-3 appear per turn based on thought content.

## Implementation Notes

- Salience Reality needs an embed function (provider-agnostic, same as LLM calls)
- Each operator needs a description field (one line, what it does)
- Matching can start simple: keyword overlap between thought and operator descriptions
- Graduate to embeddings when the operator count justifies it
- The matched set should be small (3-5 max per turn) — if too many match, rank and truncate

## Operator Descriptions (example)

```
plan: "break a goal into sequential steps and track progress"
deploy: "push code or configuration to a running environment"
notify: "send a message to a peer or external channel"
shell: "execute a command on the host device"
```

## Plan: Collaborative Deliberation

Plan is not a one-shot operator. It's a mode that Think and Act enter together.

### How it works

1. Think calls `<plan>` with its current thoughts as input
2. Act enters plan mode — it cannot execute. It can only collaborate.
3. Act responds with: what operators are available, what it thinks should happen, constraints it sees
4. This passes back to Think. Think refines — adjusts intent, narrows scope, asks for different angles.
5. They loop — Think bringing intent, Act bringing capability awareness — until they converge.
6. A plan artifact is produced. Act exits plan mode.
7. Act now executes the plan, step by step, with Salience matching operators per step.

### Why plan lives in Think

Plan is the act of deliberation. It's introspective. Think owns it because Think is where intent forms. But it requires Act's participation because Act knows what's possible — it sees the matched operators, it knows the world's constraints.

### What Act looks like in plan mode

- Cannot fire operators
- Can only respond to Think with: available operators for this thought, suggestions, constraints
- Essentially becomes a mirror — Think says "I want X", Act says "here's what I can see for X, and here's what I'd consider"
- The back-and-forth uses the existing think-back mechanism (same loop, just constrained)

### Plan artifact — the Task Reality

The output of deliberation is a Task: a tree of Realities that attaches to Self.

#### Task as Reality

A Task implements Reality. Each node in the tree is itself a Reality. Resolving a task means realizing it — same recursive descent as everything else in the system. A step doesn't "call" an operator. It realizes. Salience matches operators to its intent at realization time. If a node has children, it descends into them first.

#### Structure

```
Task (Reality, owned by Self)
├── Node: "set up the environment"
│   ├── Node: "check what's running"
│   └── Node: "provision what's missing"
├── Node: "run the migration"
├── Node: "verify it worked"
└── Node: "let the team know"
```

Each node has:
- Intent (what/why) — not operator names, not tool calls
- State: pending | active | done | revised
- Children (optional) — subtasks that resolve first

#### Lifecycle

1. Being operates freely. No task exists.
2. Think calls `<plan>`. Deliberation loop begins (Think ↔ Act in plan mode).
3. They converge. A Task tree is produced and attaches to Self as a Reality.
4. Act walks the tree. Each node realizes — Salience matches operators to that node's intent.
5. At any node, Act can pass back to Think to revise remaining branches.
6. Tree completes → detaches from Self. Being returns to free operation.
7. Completed tasks persist as raw material for skill crystallization.

#### Why it's a Reality

- Participates in collecting. The present shows the task tree. Other beings can see what you're working on.
- Same physics as everything else. No special execution engine. Just realize.
- Revision is mutation — add nodes, prune branches, rewrite intent. The tree is alive during execution.
- Think and Act operate at each node. Deliberation can happen at any depth, not just top-level.

#### No task by default

A being without a task just thinks and acts. Free cognition. Tasks only exist when a process produces one. The being isn't a task executor — it's a being that sometimes has tasks.

### Skills as crystallized plans

When the same plan pattern emerges repeatedly — same sequence of intents leading to same operator chains — that becomes a skill. A skill is a chainable callable action: one invocation that encapsulates a deliberated sequence. The being doesn't need to re-plan what it's already learned.

Salience can detect repeated chains and surface the skill instead of individual operators. The being's context stays tight — one slot for a learned sequence instead of re-deliberating every time.

## What Changes

- Plan moves from ActOps to ThinkOps
- Think's `renderOpsWithOuter` drops the outer ops section entirely. Think just doesn't see action operators.
- Act gets a `planMode` state — when active, it responds but cannot execute
- Act's `collectOps` becomes `collectMatchedOps(thought string)` — filters based on Salience score
- Self's loop passes the surfaced thought to Salience before firing Act
- Think-back appends one line to exchange: what Act had available
