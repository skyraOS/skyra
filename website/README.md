# website

The personal site from [`docs/website.md`](../docs/website.md), in its
Blueprint theme. It is static: plain HTML and CSS, and the controller's replay
compiled to WebAssembly so a visitor's browser checks a recorded run itself.

| Page | |
|---|---|
| `index.html` | the thesis, a run's record strip, the title block |
| `skyra.html` | the problem, the three invariants, a step, the status |
| `verify.html` | replay a recorded run in the browser; change a byte and watch it fail |
| `writing.html` | the design notes worth reading |
| `about.html` | who, what next, contact |

## Run it locally

`fetch` and WebAssembly need HTTP, not `file://`:

```sh
website/build.sh                        # replay.wasm and wasm_exec.js, into assets/
python3 -m http.server -d website 8000  # then open http://localhost:8000
```

## How Verify works

```text
record.sh        boots each scenario under the real sandbox (fake Codex app-server),
                 then export/ writes runs/<name>.json: exactly the records and inputs
                 replay reads to reach the head, after replaying them
replay/          a js/wasm main around controller/runtime/replay.Lineage: skyraVerify(bundle)
                 replays every record from genesis and reports which are identical
assets/verify.js loads a bundle, calls skyraVerify, draws the strip, and lets the
                 visitor edit any record's or input's bytes and replay again
```

The replay that runs in the browser is the controller's own: `replay.Lineage`
reads through a `Store`, and on the Mac that store is the tree on disk
(`replay.Replay`). Nothing is reimplemented in JavaScript.

## Recording runs

```sh
website/record.sh            # rebuild every run in runs/, dated today
```

The binaries are staged in `/private/tmp/skyra-site` so the paths stored in
each genesis record do not name this machine's home directory. Each bundle is
committed with its head hash, which is the "publish the head hash when the run
is recorded" from the plan. Re-recording changes every hash (genesis carries a
nonce), so record once per change worth showing.

## Hosting: the Mac mini

```sh
website/deploy.sh     # build here, publish to the ssh host `skyra`
```

`deploy.sh` builds the site, copies only the published files (pages,
`assets/`, `runs/`) to `~/site/www` on the mini, and runs `deploy/install.sh`
there: Caddy (installed with Homebrew if missing) serves them on
`127.0.0.1:8080` with compression, under a launchd agent
(`com.skyra.site`) that starts it at login and restarts it if it stops. Logs
are in `~/site/`.

It listens on loopback only. The public way in is **Tailscale Funnel**, with no
domain: a second launchd agent (`com.skyra.tailscaled`) runs `tailscaled` as
the login user in userspace networking (no root), its state and socket under
`~/site/tailscale`, as the node `skyra-site`. Once, on the mini:

```sh
alias ts="tailscale --socket=$HOME/site/tailscale/tailscaled.sock"
ts up --hostname=skyra-site   # prints a login link
ts funnel --bg 8080           # https://skyra-site.<tailnet>.ts.net → Caddy
```

Funnel keeps its config in Tailscale's state, so deploys do not touch it. To
look at the site over ssh instead:

```sh
ssh -L 8080:127.0.0.1:8080 skyra    # then open http://localhost:8080
```

The public, read-only ingress that would sit beside it is
[next/32](../docs/next/32-the-public-horizon.md).

## Not built yet

From the plan's order of work: the call limit on a root's reach (and its run),
the provenance root (after observation passing, next/21), and live mode.
