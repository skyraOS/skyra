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
