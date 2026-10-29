# The public Boundary: ingress from strangers, read-only

Proposal: **Not in the proposal.** A direction for the website
([`website/`](../../website/README.md)), replacing its "live mode": no model, nothing
that writes.

Date: 2026-10-02  
Status: proposed. The static site it sits beside is built and deployed to the
Mac mini (`website/deploy.sh`: Caddy on loopback, kept up by launchd; the
tunnel is Tailscale Funnel). Builds on [20](20-boundry-crossing.md) (the rule for a
new crossing), [24](24-the-room.md) (roots, Boundaries) and 16 (ingress by
parking), and [32](../design/procedures.md) (procedures and system syscalls).

## The idea

The website's live mode put Codex behind a public endpoint. That brings
everything a public model needs: prompt injection as the normal case, a spend
cap, and native execution
that is only detected. All of it goes away if the public Boundary holds **no
model, no interpreter, and nothing that writes**:

> A stranger's request is a run. It is committed under a Boundary that can only
> read what is already published, and its records are handed back so the
> stranger can replay them.

What the public can do is read, and try to read further. What they get back is
a lineage, including the fault when they reach past the Boundary. The Verify
page replays it: **your own request, and your own failed attempt, checked in
your own browser.**

## The root

A sibling of the operator's root, not a child: its own tree
(`~/site/public/.skyra/`), its own genesis, no edge to the personal tree. The
public side cannot list, read or name a personal conversation, because nothing
in its reach leads there.

| Surface | What it does | Authority |
|---|---|---|
| `read` | reads one published file or lists one directory (the existing target) | `read`: the published folder only, subpath scope |
| `roots` | reports what the Boundary can see (the existing target) | none |
| `replay` | new: given a head in the public tree, re-derives it and returns the verdict and the bundle `website/export` writes | `read`: the public tree only |

No `write`, no `codex`, no interpreter, no `connect`, no network. Syscalls go
to `system`: `{"to":"system","payload":{"narrow":{…}}}` only narrows.
The public root exposes only that syscall's contract, not `strand` or
`extend-reach`. No actor here remembers work or adds a contract. Every
instruction still passes intake under the public Boundary; composition
cannot give a procedure more authority than its caller holds.

The sandbox is what enforces this; the targets check nothing. A request to read
`~/.ssh`, `../../.skyra` or the personal tree is either a boundary fault at
intake or an OS denial reported as data, and either way it is a record.

## The host

The only code facing the internet, kept small: `hosts/public`, one endpoint.

```text
POST /run   {"to": "read", "payload": {"path": "runs/index.json"}}
  → 1. size and rate limits, before anything is committed
    2. boot one lineage under the public root (skyra boot, as record.sh does)
    3. reply {head, answer, bundle}: the bundle is what export writes, so the
       page replays it with the same skyraVerify
```

- **Limits:** request body ≤ 4 KiB; per-address and global rate limits; a
  timeout per run; a daily cap on lineages. Over a limit, nothing is committed.
- **The tree grows with every request.** One tree per day
  (`public/2026-10-02/.skyra/`), archived whole, never pruned: deleting records
  breaks lineages. A day's head hashes can be published like the recorded runs'.
- **Content:** `read` returns file contents, so the published folder holds only
  what is already on the site: the runs, and the docs chosen for Writing.
- Caddy routes `/run` to the host on loopback, beside the static files; the
  tunnel does not change.

## The rule for a new crossing, answered

| Question (20) | For public ingress |
|---|---|
| Who emits it, who checks it? | a stranger emits the request; the host checks only size and rate. The payload is untrusted bytes, interpreted at intake against the public Boundary like any target output (invariant 1) |
| Is it authority? | no. A request names a surface and a payload. The Boundary comes from the public root's genesis, never from the request |
| Is it state the runtime acts on? | it is an observation: the bytes received, stored and hashed. What the runtime acts on is derived from it |
| Does it name an actor? | no new one. `read` and `roots` exist; `replay` is a leaf like them |
| Is it a descriptor? | no. The host passes bytes to `skyra boot`; no descriptor reaches it from outside |
| Who holds it while nothing runs? | the day's public tree. The host holds no state between requests |

## What it shows

A "try to break it" page beside Verify: a box to ask for a path, a few
suggestions (`runs/index.json`, `../`, `/Users`, `.skyra/records`), and each
answer drawn as a record strip that replays in place. Adversarial, live, and
free to run: no model is called.

## Order

1. `replay` as a target (and its contract), with a test that it refuses a head
   outside the public tree.
2. `hosts/public`: the endpoint, its limits, one lineage per request, the
   bundle in the reply. Tests for every limit.
3. The public registry and the day's tree; the published folder.
4. Caddy routes `/run`; the "try to break it" page.
5. The tunnel.

## Open

- **Tunnel:** the static site already uses Tailscale Funnel (a `ts.net` name);
  public ingress would use the same tunnel and Caddy route.
- **Is server-side `replay` worth having** when the browser replays anyway? It
  is a second opinion, and it lets a short link (`/verify?head=…`) fetch a
  bundle by head. Without it, the host returns the bundle itself and the
  surface goes.
- **Which docs are published**, and whether `read` should see them as files or
  only the runs.
- **Rate limit by address** behind a tunnel: the tunnel's forwarded address is
  a header the host must trust, which is its own small crossing.
