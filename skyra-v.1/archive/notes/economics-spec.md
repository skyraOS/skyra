# Economics — Task Economy

A Reality that tracks the value flow between beings. Who asked for what, who delivered, what's outstanding.

---

## What it is

Economics is the accounting layer for the task economy. When being A asks being B to do something, that's a task. When B delivers and A accepts, that's a resolution. Economics tracks all of it — open tasks, completed tasks, who's doing the most work, who's asking for the most.

Economics builds on top of the task layer (see task-layer.md). The task layer manages individual task objects — heartbeats, nudges, lifecycle. Economics is the aggregate view — the ledger.

---

## Task recap (from task-layer.md)

```
task {
  id             string
  issued_by      string      // the being who created it (owns closure)
  assigned_to    string      // the being doing the work
  impulse        string      // the original request
  last_heartbeat time.Time   // updated on every relation from assigned being
  resolved       bool        // only issuer can set true
  created_at     time.Time
  resolved_at    time.Time   // zero until resolved
}
```

The task layer watches traffic. If a relation passes through with the assigned being's name, heartbeat resets. 10 minutes idle → nudge. Only the issuer closes.

---

## What Economics tracks

### Per-being ledger

```
being_ledger {
  name           string
  tasks_issued   int       // tasks this being asked others to do
  tasks_assigned int       // tasks assigned to this being
  tasks_completed int      // tasks this being completed (resolved by issuer)
  active_tasks   int       // currently open tasks assigned to this being
}
```

Every being has a ledger. The universe exposes all ledgers. The frontend can show: skyra has 3 active tasks, has completed 12, issued 5 to others.

### Active tasks

The full list of open tasks. Who asked, who's working, how long it's been, last heartbeat.

```
active_task {
  id             string
  issued_by      string
  assigned_to    string
  impulse        string    // what was asked (truncated for display)
  age            int       // seconds since created
  idle           int       // seconds since last heartbeat
}
```

### History

Resolved tasks. Kept for the universe view — the full record of who did what for whom.

```
resolved_task {
  id             string
  issued_by      string
  assigned_to    string
  impulse        string
  duration       int       // seconds from creation to resolution
}
```

---

## How tasks get created

A task is implicit in a relation. When being A addresses being B with a request, that's a task. But not every message is a task — "hey how are you" isn't a task.

Two options:

1. **Explicit** — the being uses a protocol tag: `<task>do this thing</task>` inside their message. The task layer intercepts it.
2. **Implicit** — every being-to-being message that crosses an exchange is a task until the issuer says it's done.

Option 1 is cleaner. Option 2 is simpler but noisy — casual conversation becomes tasks. For alpha: option 1. The Act layer already enforces protocol tags. Adding `<task>` is natural.

A being's outer layer (Act) can emit:
```
<louise><task>research what memory architectures exist and report back</task></louise>
```

The task layer sees the `<task>` tag, creates the task object, strips the tag, and passes the message through to the exchange normally. Louise receives the message. The task layer watches for Louise's heartbeat.

---

## Resolution

The issuer resolves. When being A receives being B's response and is satisfied, A can emit:
```
<task-resolve>task_id</task-resolve>
```

Or simpler: the task layer watches for the issuer to address someone else or to go idle after receiving the response. If the issuer moves on, the task resolves implicitly.

For alpha: explicit resolution with the tag. Keeps it clean.

---

## Where it sits

Economics is a Reality that sits at the NewThread level — it can see all relations passing through. It watches traffic, manages task objects, updates ledgers.

```
Universe
├── NewThread
│   ├── Economics    ← watches all relations, manages tasks
│   ├── Exchange
│   ├── skyra (Self)
│   └── michael (User)
└── ...
```

Or: Economics wraps Exchange the way Universe wraps Thread. Relations pass through Economics on their way to Exchange. Economics intercepts task tags, tracks heartbeats, then passes through.

The second option is cleaner — Economics is a layer in the descent, not a side observer.

```
NewThread → Economics → Exchange → Being
```

---

## Collecting

When a collecting relation passes through Economics:

```go
if r.Collecting {
    for _, ledger := range e.Ledgers {
        r.Export("ledger:"+ledger.Name, ledger)
    }
    for _, task := range e.ActiveTasks {
        r.Export("task:"+task.ID, taskSnapshot(task))
    }
    r.Export("node:economics", RealityNode{...})
    return ""
}
```

The universe sees all ledgers and all active tasks.

---

## Universe state addition

```json
{
  "beings": [...],
  "threads": [...],
  "exchanges": [...],
  "economics": {
    "ledgers": [
      {
        "name": "skyra",
        "tasks_issued": 5,
        "tasks_assigned": 12,
        "tasks_completed": 10,
        "active_tasks": 2
      }
    ],
    "active_tasks": [
      {
        "id": "a1b2",
        "issued_by": "skyra",
        "assigned_to": "louise",
        "impulse": "research memory architectures...",
        "age": 340,
        "idle": 15
      }
    ]
  },
  "reality_graph": {...}
}
```

This replaces the current `economics: map[string]int` with structured data. The simple key-value store becomes a real economy.

---

## What this is not

- Not inference tracking. That's the Inference Reality's job.
- Not a currency system. No tokens, no coins. It's a ledger of work.
- Not a marketplace. Beings don't bid on tasks. The issuer assigns directly.
- Not coupled to inference energy. A being doesn't need energy credits to issue a task. It needs energy to think — that's a different constraint.

---

## Open questions

1. **Task detection** — explicit `<task>` tag vs implicit (every cross-being message). Explicit is cleaner but requires the being to learn the protocol. The Act system prompt already teaches protocol — adding one more tag is trivial.

2. **Multi-step tasks** — being A asks B, B asks C for help. Is that one task or two? Probably two — B issues a sub-task to C. The ledger tracks both independently.

3. **Idle threshold** — 10 minutes is arbitrary. Should it be per-task? Per-being? Configurable in the genome?

4. **Task history depth** — how many resolved tasks do we keep? All of them for alpha. Trim later if it matters.

5. **Economics + Inference bridge** — the deliberate non-coupling. Later, completing tasks could recharge energy. For now they're independent. The bridge is a post-alpha decision.
