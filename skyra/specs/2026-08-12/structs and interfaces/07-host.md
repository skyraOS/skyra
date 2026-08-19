# 07 · Hosts

**Package:** `host`

**WHERE it lands.** The Linux device-class model: the kernel calls through a class interface, a
driver fills it in, and the kernel never knows what is behind it.

## The classes

```go
type Host interface {                     // SPEC
    ID() ids.HostID
    Execution() ExecutionOps
    Storage()   StorageOps
    // display · audio — the projection substrate. NOT NAMED YET.
}

type ExecutionOps interface {             // SPEC
    Spawn(argv []string, dir string) (Handle, error)
}

type StorageOps interface {               // SPEC(n) — "storage soon"
    Read(path string)  ([]byte, error)
    Write(path string, body []byte) error
}

type Handle struct {                      // SPEC(n)
    Host ids.HostID
    Ref  string
}
```

**Classes rather than one method**, because an ops struct with optional members grows by addition
where a single interface cannot grow at all — and this is the one surface **selection cannot
correct**.

**An absent implementation is not a declaration.** It is a build-time fact, and calling into a class
a host does not implement fails loudly. That is what lets capability vary without reintroducing the
claims that dissolved.

## The local host

```go
type Local struct {                       // ADDED — implements Host
    Root string
}

type localExec  struct{ l *Local }        // ADDED
type localStore struct{ l *Local }        // ADDED
```

**`here` is the axiom** — the host the runtime is executing on is given, not discovered. The same
shape as the standing root: the thing you cannot derive because you are standing on it.

`localStore.resolve` refuses to leave its root. **That is what a grant would have to be realized by** —
if the kernel checks a grant and the platform layer does not confine, the grant described nothing and
the record reads as bounded.

## Review notes

```text
Spawn IS DECLARED AND UNIMPLEMENTED.   `localExec.Spawn` returns "execution class not implemented."
                                       world.md says "execution now, storage soon"; the build has
                                       storage and not execution, which is backwards.

Handle IS {Host, Ref}.                  an opaque reference plus the host on which it is valid.
                                       Stdout, Stderr and Exit were removed; process output is not
                                       part of what a handle is.
```

**Network is not a class** and is not here, correctly — it is transport, beneath everything. *Fetch
this URL* is either spawning an HTTP client or invoking an HTTP entity, and a network class would
absorb both.
