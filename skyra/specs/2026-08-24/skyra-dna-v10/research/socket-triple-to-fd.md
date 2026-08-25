# The socket triple → fd

**Date:** 2026-08-24
**Status: RESEARCH — keep for later.** Grounded on this machine (macOS, darwin). How a
`(domain, type, protocol)` lookup becomes a file descriptor, and why the fd is a handle, not the
endpoint. Worth remembering: this is the OS mirror of the THINK-proposes / executor-binds line.

---

After the triple lookup confirms the protocol exists, the kernel does three allocations and hands you
an integer:

```text
socket(AF_INET, SOCK_STREAM, 0)
   │
   ├─ 1. allocate the SOCKET object   ← the real endpoint; carries the looked-up protosw (TCP behavior)
   ├─ 2. wrap it in a FILE object     ← an open-file entry with the socket's read/write vtable
   └─ 3. install that file in the process's FD TABLE at the lowest free slot
                                       └─ that slot's INDEX is the fd you get back  (here: 3)
```

So the fd is a **chain of indirection**, not the socket itself:

```text
fd  (3, an int, per-process)
  → fd-table slot
  → struct file            (open-file object)
  → struct socket          (the endpoint + protocol state)
  → protosw                (the (domain,type,protocol) behavior the lookup found)
```

## The proof is in the output

- **The endpoint is typed but nameless.** Right after `socket()`, `getsockname()` is `0.0.0.0:0` — it
  is a TCP endpoint (the triple is resolved), but it has **no address**. The triple → fd step binds
  the **protocol**, not the **location**.
- **`bind()` is the separate step** that gives it a name (`127.0.0.1:58164`). And `connect()` would
  give it the remote name. **Addressing is a later act.**
- **The fd is just an opaque index.** `dup()` produced fd `4` pointing at the **same** endpoint — the
  number is a per-process handle, the endpoint lives behind it and can have many handles. The next
  socket took `5`, the lowest free slot.

## Why remember it

```text
socket() — lookup the triple → typed, NAMELESS endpoint + fd    ≈  admit the CAPABILITY (what network)
bind()/connect() — resolve the concrete ADDRESS                  ≈  resolve the DEPENDENCY (which endpoint)
```

The triple lookup answers "what kind of network" (the capability THINK bounds the space to, nameless).
`bind`/`connect` answers "which actual address" (the concrete resolution, learned at the moment of use
— the executor's job at dispatch). The fd is the narrow handle: the process passes a small integer to
`read`/`write` and never touches the socket object directly. The kernel hands out a **capability
reference, not the thing.**
