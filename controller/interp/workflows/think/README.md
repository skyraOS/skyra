# think

The first continuation procedure ([proposal 38](../../../../docs/next/proposal/38-think-procedure-proposal.md)).
The user's text is its one parameter. It attaches read-only tools to this
invocation only, then runs Codex on the text under them:

- `scope` narrows Codex to the model's tools, `echo`, `read`, `roots` and
  `write`, and its own operations, for this invocation only. The workspace is
  read-only at the OS, as everywhere: only `read` is granted it, and `write`
  reaches nothing but the conversation's scratch space.
- `run` invokes the reference `scope` returned, with `header: "think"`. The
  Codex adapter starts a new thread whose whole prompt is
  `You are inside Skyra, a cognitive architecture, fulfilling: think`.

Install it by naming this file as the interpreter's `bundle` in the registry.
Earlier turns as `context`, and selection by a star, come with steps 5–7.
