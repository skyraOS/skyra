# The Example — `read-local(file.txt)`

**Status: WALKTHROUGH.** Traced against `skyra-dna-v5` as it exists. No conclusions — where the
trace stops, it stops.

---

## The setup

Three things, deliberately in three places:

```text
THE RUNTIME    the kernel, the graph, the registry.     ON A LAPTOP.
THE MODEL      whatever interprets `read-local`.        A REMOTE API.
THE FILE       file.txt                                 ON A SERVER.
```

Nothing about this is exotic. It is the ordinary case the moment there is more than one machine,
and `terminology.md` already says a runtime reaches devices rather than spanning them.

## The ask

A parent possibility has an observation naming a file and wants its contents.

```go
// what the parent emits
grammar.Member{
    Engram:       "read-local",
    Cite:         "obs-0007",
    Entity:       "codex",
    Observations: []ids.ObservationID{"obs-0008"},   // {path: "file.txt"}
    Host:         "",                                 // optional. and never read.
}
```

```go
// what it names
graph.Engram{
    ID:          "eng-0042",
    Expectation: graph.Expectation{Expression: "read a local file and return its contents"},
}

// what it passes
graph.Observation{
    ID:      "obs-0008",
    Content: content.Content{"path": content.Str("file.txt")},
}
```

**That is everything the system holds about this invocation.** An expectation, a path, and an entity
name.

---

## Step 1 · The kernel resolves the entity

```go
interp, err := k.Reg.Open("codex")
```

```go
type Registry struct {
    interpreters map[ids.EntityID]func() Interpreter
}
```

The registry maps an entity ID to a constructor. **No host is involved.** There is nothing to ask
about where `codex` is, which hosts can reach it, or whether more than one can.

`Member.Host` was left empty by the parent — but it would not have mattered, because nothing in
`kernel/` reads it. Checked: every reference to `.Host` in the kernel is either the `Kernel.Host`
field or `k.Host.ID()`.

## Step 2 · The kernel records where it landed

```go
graph.Invocation{
    Entity: "codex",
    Host:   k.Host.ID(),     // "here". hardcoded. the only host the kernel has.
}
```

```go
type Kernel struct {
    Host host.Host    // ONE host. a field, not a table.
}
```

So the record says `here`, and will say `here` for every invocation ever made, on every machine,
regardless of what actually happened.

## Step 3 · The model runs somewhere

`codex` assembles a prompt and reaches the model:

```go
entity/codex.Step  →  inference.CallCodex(system, present)
                   →  exec.CommandContext(ctx, "codex", args...)
                      cmd.Dir = <a temp dir>
                      cmd.Env = os.Environ() + TMPDIR      // CODEX_HOME left for auth
```

A process is spawned **on the machine the runtime is on** — the laptop — because `exec.Command`
has no other option. This does not pass through `host.ExecutionOps`. Nothing in `skyra-dna-v5`
does; the only reference is a test returning `nil`.

The model itself is remote. So by this point the work has touched two machines, and the record
holds one field that says `here`.

## Step 4 · The model decides to open the file

The expectation says *read a local file*. The observation says `file.txt`. The model emits
something that means "open it."

**Stop here. Ask two questions of everything the system holds.**

```text
WHICH MACHINE IS THE FILE ON?

  the expectation           "read a local file"        — says LOCAL. relative to what?
  the observation           {path: "file.txt"}         — a path. no machine.
  the member                Host: ""                   — empty, and unread.
  the invocation            Host: "here"               — the laptop, by hardcoding.
  the entity                "codex"                    — an identity. no location.

WHICH MACHINE IS THE MODEL ON?

  the same five places, and none of them says.
```

Neither question is answerable from anything in the frame, the emission, the record, or the
registry. **"Local" has no referent**: local to the runtime, local to the process that was spawned,
or local to the model — and the example was constructed so those are three different machines.

## Step 5 · What `open()` would need anyway

Assume the location question is somehow settled and a process on the right machine issues the call.

```c
open("file.txt", O_RDONLY)
```

```text
CARRIED IN THE FRAME    the path                            obs-0008

NOT CARRIED ANYWHERE    the working directory   "file.txt" is RELATIVE and resolves against
                                                 the process's cwd
                        the credentials          uid/gid decide whether the open is permitted
                        the machine              step 4
                        the mount table          the filesystem must be mounted there
```

`cmd.Dir` is set to a **temporary directory** created per call, so under the current code the
working directory that a relative path would resolve against is a fresh empty temp dir that is
deleted afterwards.

## Where the trace stops

Four places, in the order they were reached:

```text
1  THE REGISTRY HAS NO HOST            step 1. entity → interpreter, with no location in between.

2  Invocation.Host IS HARDCODED        step 2. the field exists and records one constant.

3  "LOCAL" HAS NO REFERENT             step 4. three candidate machines, and nothing in the
                                       expectation, the observation, the member, the record or
                                       the registry distinguishes them.

4  THE CONTEXT IS NOT REPRESENTED      step 5. cwd, credentials and mount are what decide whether
                                       the call succeeds, and none of them appears in any type
                                       in the tree.
```

## For reference — every place a location could have been written

```go
grammar.Member{      Host ids.HostID }    // exists. optional. NEVER READ.
graph.Invocation{    Host ids.HostID }    // exists. RECORDED, NOT KEYED. hardcoded to "here".
kernel.Kernel{       Host host.Host  }    // ONE host. a field.
entity.Registry{     — }                  // no host field at all.
graph.Engram{        — }                  // no host field. correctly: an engram is WHAT.
graph.Expectation{   Expression string }  // free text. could SAY a machine; nothing would read it.
host.Host{ ID() · Execution() · Storage() }  // a host cannot be asked what it can reach.
```
