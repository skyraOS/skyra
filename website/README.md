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

## Hosting

Any static host works (GitHub Pages, Cloudflare Pages). `assets/replay.wasm`
and `assets/wasm_exec.js` are build outputs: run `build.sh` before deploying, or
commit them.

## Not built yet

From the plan's order of work: the call limit on a root's reach (and its run),
the provenance root (after observation passing, next/21), and live mode.
