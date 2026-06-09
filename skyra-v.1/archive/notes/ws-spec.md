# WS Component Spec

## What It Is

A component on the macbook device, same as Terminal and OpenRouter. Implements Reality. It's Terminal for the browser — instead of stdin/stdout over the keyboard, it's messages over a persistent network connection.

## How WebSockets Work

Normal HTTP: client sends request → server responds → connection closes. One-shot.

WebSocket: client sends an HTTP request with "Upgrade: websocket" → server says yes → the connection stays open. Now both sides can send messages at any time. The server can push without being asked. The connection lives until one side closes it.

That's it. It's a pipe that stays open.

## What WS Does in Skyra

Two directions:

**Out (runtime → browser):** Push universe state on every change. The OnResolve hook already fires after every relation resolution and builds the full JSON snapshot. WS takes that snapshot and writes it to every connected browser.

**In (browser → runtime):** Receive user input. When michael types in the browser instead of the terminal, the message arrives over WS. WS creates a Relation (same as Impress does for terminal input) and feeds it into the universe.

## Interface

```go
type WS struct {
    id      string
    Device  Reality
    Port    int
    clients map[*websocket.Conn]bool  // connected browsers
    mu      sync.Mutex
    send    chan string                // outbound messages
}
```

Implements Reality:
- `ID()` → "ws"
- `Create(r)` → initializes server, starts listening on port
- `Realize(r)` → two modes:
  - **Collecting:** broadcasts the JSON state to all connected clients
  - **Normal:** blocks waiting for a message from a connected client (same as Terminal blocks on stdin)

## Realize Behavior

### Collecting mode (r.Collecting)

Called by OnResolve after every relation resolution. The universe JSON is already built — WS just needs to send it.

```
Universe.Realize(collecting) → Thread.Realize(collecting) → ... → JSON string
WS receives the JSON string
WS writes it to every connected client
```

This is the push. The frontend doesn't ask for updates. They arrive.

### Normal mode (user input)

When a relation is routed to a user being whose device resolves to WS:

```
Thread routes to michael
  → User.Realize → MacOS.Realize → WS.Realize
    → WS prints the impulse (sends skyra's message to the browser)
    → WS blocks waiting for input from the browser
    → browser user types, sends message
    → WS returns the message string
```

Same contract as Terminal. Print the incoming message, wait for a response, return it.

## Connection Lifecycle

1. Browser opens `ws://localhost:PORT/ws`
2. Server upgrades the HTTP connection
3. Server sends full universe snapshot immediately (the client is caught up)
4. Connection stays open
5. Runtime pushes new snapshots on every change (OnResolve)
6. Browser can send messages at any time (user input)
7. On disconnect, server removes client from the set. Reconnecting browser gets a fresh full snapshot.

## Message Format

### Server → Client (outbound)

Two message types:

```json
{
  "type": "universe",
  "ts": 1714650000,
  "payload": { ... full universe state ... }
}
```

```json
{
  "type": "impulse",
  "ts": 1714650000,
  "payload": {
    "from": "skyra",
    "content": "hello michael"
  }
}
```

`universe` is the full snapshot — pushed on connect and on every change.
`impulse` is a message directed at the user — what Terminal would print to stdout.

### Client → Server (inbound)

```json
{
  "type": "input",
  "payload": {
    "content": "hey skyra, what do you think about memory?"
  }
}
```

This is what Terminal would read from stdin.

## Where It Sits

```
MacOS (macbook)
├── Terminal (stdin/stdout)
├── WS (websocket server)      ← this
└── OpenRouter (inference)
```

WS is a peer of Terminal. Both are components on the same device. A user being can be wired to either or both. For alpha, michael gets both — terminal for dev, WS for the frontend.

## Genome

Already declared:

```
component ~name ws ~type websocket ~port 8080 ~device macbook
```

## Integration Points

### OnResolve (push state)

Currently:
```go
thread.OnResolve = func() {
    present := universe.Realize(&Relation{Collecting: true})
    debug.Log("[universe]:", present)
}
```

After WS:
```go
thread.OnResolve = func() {
    present := universe.Realize(&Relation{Collecting: true})
    debug.Log("[universe]:", present)
    ws.Broadcast(present)  // push to all connected browsers
}
```

### User device routing

When michael's device is WS instead of Terminal, the flow changes:

```
User.Realize → MacOS.Realize → WS.Realize
  → sends impulse to browser
  → waits for browser input
  → returns input string
```

MacOS routes to WS based on which component the relation targets. For alpha, the simplest path: if a WS client is connected, route to WS. If not, fall back to Terminal.

## Dependencies

- `golang.org/x/net/websocket` — already in go.mod (x/net is there)
- Or `gorilla/websocket` — more common, better API, would need to add to go.mod

Recommend `golang.org/x/net` since it's already a dependency. Keep it simple for alpha.

## What Doesn't Change

- Reality interface — WS implements it like everything else
- Universe / collecting — already builds the JSON, WS just sends it
- Exchange / Thread / Self / Think / Act — no changes
- Terminal — stays as-is, WS is a peer not a replacement

## Open Questions

1. **Multiple users over WS** — For alpha, one user (michael). But WS naturally supports multiple connections. Each connection could be a different user being. Deferred.

2. **Which component for michael** — Terminal, WS, or both? If both, which one does MacOS route to? Simplest: check if WS has a connected client, prefer it, fall back to Terminal.

3. **Streaming vs snapshot** — The data spec discusses deltas vs full snapshots. For alpha, full snapshot on every change. The state is small. Deltas are optimization for later.

## File Changes

| File | Change |
|------|--------|
| `ws.go` | **New file.** The WS component. |
| `main.go` | Bootstrap parses `websocket` component type, wires Broadcast into OnResolve. |
| `macos.go` | May need smarter routing (WS vs Terminal). |
| `genome.skyra` | Already has the component line. |
| `go.mod` | Maybe add gorilla/websocket, or use existing x/net. |
