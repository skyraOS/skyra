# think

The first continuation procedure ([proposal 38](../../../../docs/next/proposal/38-think-procedure-proposal.md)).
The user's text is its one parameter. It attaches read-only tools to this
invocation only, then runs Codex on the text under them:

- `scope` narrows Codex to `echo`, `read`, `roots` and its own operations,
  dropping all write and scratch authority (`without`), so the whole subtree
  is read-only at the OS, and `write` is not in reach.
- `run` invokes the reference `scope` returned, with `header: "think"`. The
  Codex adapter starts a new thread whose whole prompt is
  `You are inside Skyra, a cognitive architecture, fulfilling: think`.

Install it by naming this file as the interpreter's `bundle` in the registry.
Earlier turns as `context`, and selection by a star, come with steps 5–7.
