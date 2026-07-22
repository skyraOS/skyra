# dnav2 — archived

The second genome generation and everything that ran it. Superseded by
`src/skyra/skyra-dna-v3/`.

## Why it is kept

v2 is not dead weight — it is the **lineage evidence**. It was built, run, and hit a
real ceiling, and that ceiling is the reason v3 exists:

> Free-text predictions could only be graded by an LLM judge, so the learning signal
> bottomed out in an opinion with no independent source. Causality had to be ASSERTED
> by the model, which made the causal graph an opinion too. And the goal was
> re-derived every turn, so the being could not want anything for longer than one
> impulse.
>
> — `src/skyra/skyra-dna-v3/doc.go`

The thesis cites that history rather than re-running it. `genes_final.go:80`
(`graderSystem`) and `genes_efficient.go:59` (`execGraderSystem`) are the two
adversarial-neutral judges — the concrete artifact behind the claim that a judged
grade bottoms out in an opinion. They are worth reading before anyone proposes
adding a judge back.

`cmd/phase1-ab/` and `cmd/phase2-accumulate/` are the harnesses behind the first
measured learning curve.

## Contents

```text
skyra-dna-v2/            the genome (54 files)
main.go                  the interactive v2 being — was the repo-root `go run .`
cmd/tui/                 TUI runner
cmd/validate-v2/         v2 validation
cmd/phase1-ab/           A/B experiment harness
cmd/phase2-accumulate/   accumulation experiment harness
```

## Why the `_` prefix

The Go toolchain ignores directories beginning with `_` or `.`, so this tree is
excluded from `go build ./...`, `go vet ./...`, and `go test ./...` without needing a
second module or `replace` directives. dnav2 imports live in-module packages
(`src/debug`, `src/skyra`, `src/skyra/relation`, `src/skyra/traverse`,
`src/skyra/world`), which is what made a separate-module archive the wrong shape.

## Running it

You cannot build this in place. The package's own import path
(`skyra-v1/src/skyra/skyra-dna-v2`) no longer resolves, so even naming a file
explicitly fails. Restore it first:

```sh
git mv archive/_skyra-dna-v2/skyra-dna-v2 src/skyra/skyra-dna-v2
git mv archive/_skyra-dna-v2/main.go .
```

The current entry point is `cmd/dnav3/`.
