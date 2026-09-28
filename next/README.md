# Next: capability

Where the controller goes after the prototype. Agreed on paper, not built.

1. **[`14-interpreted-not-stated.md`](14-interpreted-not-stated.md)**: the
   model never emits a control operator. It supplies only operands (which
   capability, and a payload), and the runtime derives invoke, return, fault or
   interrupt. A return is invoking a single-use `reply` capability.
2. **[`research-question.md`](research-question.md)**: the question, the
   metrics, the baseline, the September–December plan, and the open questions
   to decide together.
3. **[`15-adapters.md`](15-adapters.md)**: the dumb actor host and the
   stateless adapter step. Built.
4. **[`16-ingress.md`](16-ingress.md)**: ingress by parking a frame with its
   host; conversations on the controller. Built.
5. **[`17-app-server-custody.md`](17-app-server-custody.md)**: the Codex
   app-server as a registered actor under Seatbelt; Codex asks, never
   executes. Draft.
6. **[`agenda.md`](agenda.md)**: the immediate work. Define input contracts,
   write `interpret` as a total pure function, define `reply`'s lifecycle, and
   fill in the checklist for every transition.
7. **[`21-observation-passing.md`](21-observation-passing.md)**: observations
   travel by hash. Nodes may author them for child invocations; continuations
   pass child observations on return instead of authoring a replacement result.
8. **[`22-read-tool.md`](22-read-tool.md)**: `read`, the first leaf that
   observes the world: a workspace grant approved once, a typed observation
   with provenance, and the adapter's model-facing tool set. Built.
9. **[`23-desktop-app.md`](23-desktop-app.md)**: `Skyra.app`, a SwiftUI client
   of the conversation protocol in the TUI's night-sky theme, and the user's
   entry point: the workspace grant approved once, and later the one place a
   boundary crossing is asked. Cuts 1 and 2 built.
