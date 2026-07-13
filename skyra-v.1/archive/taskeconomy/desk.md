# Desk

A task is a commitment to change something. It sits between purpose and impulse. Purpose is permanent direction — it never completes. Impulse is momentary input — it arrives and passes. A task is a commitment with a horizon: it persists across exchanges but it resolves.

## Structure

Two realities. One operator.

**Desk** — a Reality inside Self, alongside Being, Memory, Think, and Act. The desk holds the being's tasks and view state. When Self builds the Relation for cognition, the desk attaches a parser that renders the being's current task view into its present. The being sees its desk as part of its world, shaped by its own prior commands about what to show and hide.

```go
type Desk struct {
    id    string
    Items map[string]*Task
    Views map[string]string // task name → "open" | "closed"
}
```

The desk is not a to-do list bolted onto the being. It's a section of the being's present that the being controls. A being working on "fix the websocket" can have that task open — seeing all its subtasks, their states, their detail — while "ship v.06" is minimized to just a name and status line. The being manages its own attention.

**Task** — the artifact. Recursive. A task holds an ordered slice of tasks inside it. Same shape all the way down. Full definition in the task object doc.

## Desk Inside Self

The desk lives inside Self, same level as Memory:

```
Self holds: Being, Memory, Desk, Think, Act
```

When Self.Realize fires, it attaches the desk's parser to the Relation alongside the being and memory parsers. By the time the Relation reaches Think, the being's present includes its desk view as a given. The being doesn't query the desk during thought — the desk is already there, already shaped by the being's prior decisions about what to pay attention to.

## Relationship Slicing

The desk is partitioned by relationship. Tasks belong to the relationship that produced them. Thought history is partitioned the same way. The being explicitly names the relationship in every task command — the system does not infer it from the current exchange.

This gives the being control over context weight. It can collapse entire relationship slices it doesn't need right now, keeping only the relevant ones visible. A being working on a task from michael doesn't need to see its full task history with louise unless the work overlaps — in which case it opens both slices.

Tasks that span relationships name multiple parties. The task appears in each slice.

When a plan resolves and the delta gets saved to memory, it's tagged by relationship. The being doesn't just learn "I underestimated this" — it learns "I underestimated this with michael." The learning is relational.

## What the Being Sees

A being with tasks across two relationships, one slice open:

```
desk:
  michael:
    ▸ ship v.06 [active]
    ▾ fix websocket timeout [active]
        ▸ reproduce the drop [done]
        ▸ check keepalive interval [active]
        ▸ test under load [held]
  louise:
    (collapsed)
```

The open task shows its subtasks (which themselves can be open or closed). Closed tasks are one line. Collapsed relationship slices show nothing. The being chose this view. It can change it on the next think pass.

A being that opens both slices:

```
desk:
  michael:
    ▸ ship v.06 [active]
    ▸ fix websocket timeout [active]
  louise:
    ▸ deployment review [active]
    ▸ memory compression [held]
```

Context shifts. The being sees what it chose to see.

## Why Inside Self

The desk is the being's own workspace — not physics, not a world layer. It belongs alongside Being and Memory as part of what makes a Self. Like Memory, it's state the being owns and mutates through its own operators.

- The desk persists across exchanges as part of Self's state
- Collection can snapshot it through Self's collecting pass
- Multiple beings could still see each other's desks (future: shared projects via exchange)
- The desk doesn't consume think budget to render — it's attached to the Relation before Think fires

## Collection

During universe collecting, the desk exports its full state: all tasks (expanded, regardless of view state), all subtasks recursively, all view settings. The browser sees everything. The being sees what it chose to see.

## Task Object

```go
type Task struct {
    id           string
    Name         string
    Description  string
    Assumptions  []string
    Commands     []string
    Validation   string
    AcceptedBy   string
    State        string   // active, held, done, dropped
    Items        []*Task
}
```

- **Name** — what the task is called
- **Description** — what should change
- **Assumptions** — what the being believes to be true before starting. Stated explicitly at creation, not after. These are the reference point for learning — when a plan resolves, assumption drift is the sharpest signal of what the being got wrong
- **Commands** — necessary commands to execute
- **Validation** — how the being knows the task is done (internal check)
- **AcceptedBy** — who confirms the task is actually done. Since tasks are scoped by relationship, the natural acceptor is the party who gave the task. Validation is the being's own sense of done. Acceptance is the relationship confirming it
- **State** — active, held, done, dropped
- **Items** — ordered subtasks. Recursive. Same shape all the way down

## Task vs. ReAct

In a ReAct system, the agent has no control over its own context. The loop accumulates observations and the agent sees all of them. The only context management is external — the framework truncates or summarizes.

In Skyra, the being manages its own context window. It decides what to expand, what to minimize, what to focus on. This is not prompt engineering from outside — it's the being making attention decisions from inside. The desk is the mechanism.

This is the difference between a system that manages an agent's context and a being that manages its own attention.

## Open Questions

**Persistence.** Does the desk write to disk? If so, where? Same pattern as memory (~/.skyra/beings/{name}/desk/)?

**Shared desks.** The owner-keyed map supports it structurally. Should beings be able to see each other's desks? Shared projects?

**Relationship to Exchange.** Could Exchange compress old entries differently based on what the desk says is active? (A being focused on "websocket fix" might want exchange history about that topic preserved longer.)

**View as physics.** Is the view state something economics should tax? Opening more tasks costs attention budget? Or is that over-engineering a constraint that should just be felt naturally through context window limits?
