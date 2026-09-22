# 09 · Trace

**Package:** `trace`

**Outside the design, and the package comment says so.**

```go
type Run struct {                             // ADDED — by admission
    Dir   string
    mu    sync.Mutex
    main  *os.File
    steps int
}

func Open(root string) (*Run, error)
func (r *Run) Close() error
func (r *Run) Logf(format string, a ...any)
func (r *Run) Step(engram, ent, kind, prompt, response string) string
func (r *Run) Write(name, body string)
```

It exists because `kernel.md` is explicit that **you cannot ask the graph what the system is doing** —
every operational question is a kernel-side facility, out of band, unavailable to anything inside the
model — and something still had to be debuggable.

```text
WHY IT IS NAMED HERE   so nobody later mistakes it for structure and builds a dashboard out of
                       engrams. kernel.md warns that will be tempting the first time someone wants
                       one, and it stays tempting.
```

## Review notes

**This package is fine. Its placement is not.**

```text
34 LINES OF FORMATTING sit in kernel.go — indent · shortExpr · oneLine — truncating and rendering
                       EXPECTATIONS for a log. that is THE KERNEL READING CONTENT, even if only to
                       shorten it, and it is the wrong side of NO SEMANTICS.

9 TRACE CALLS          threaded through the loop, plus a *trace.Run field on the Kernel struct.
```

The fix is that a trace facility should be **handed events**, not reach into state. The kernel emits
`(engram, entity, kind)` and something else decides how to render it — which also moves the
truncation, and with it the content-reading, out of the component that must hold least.

`Run` is nil-safe on every method, so the substrate tests pass `nil` and no tracing happens. That is
the shape the fix wants generalized: **an optional sink, not a held dependency.**
