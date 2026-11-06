# Missions: Plan, Spec, Launch

Each conversation is a Mission, with one spec identity and one tree identity.
Its human holds a root above its trees (`core/humanroot`, journaled by
`runtime/humanroot`). Chat, the Plan conversation, invokes its existing
conversation procedure. Every spec revision is a typed Spec observation held
at that root: a goal, the observations it cites and the capabilities it
proposes. Launch is the human passing exactly one held Spec to the
apprentice's existing tree contract. Conversation messages are never used as
tree input implicitly.

`work.json` beside the conversation's runtime records keeps its spec revisions,
each with its tools, its source (human or Plan), the observations it cites and
the identity of its Spec at the root, and the latest execution reference. The
root's journal, under `mission/<epoch>/` in the conversation's home, is the
record of what was authored, planned and passed; execution views still come
from replaying tree records. The epoch names the pinned kernel and tool
binaries: after a rebuild, a Mission starts a new root, and a spec saved
against the old pins must be saved again before it can launch.
Older conversations acquire deterministic empty spec/tree identities when
loaded. Invalid work metadata is not replaced with a fresh executable spec.

The webhost exposes three conversation-scoped, request-ledger operations:

- `save-spec`: `conversation`, `text`, the proposed tools as `reach`, an
  optional `profile`, and expected `specRevision`. The human authors a Spec at
  the root; changed text, tools or profile create a revision. With no
  `profile`, the latest revision's is kept. Saving never starts execution.
- `submit`, in a Plan conversation: a message to Plan. A Mission's
  conversation (the think root, on a host with a tree executor) is Plan: there
  is no other agent in it. Each message is one Plan turn. The host authors the
  message and the conversation so far at the root, then invokes Plan with them
  and the latest Spec. Plan is granted a read-only way around the workspace
  (`pwd`, `cd`, `ls`, `find`, `read`, which pass a directory observation
  between them), a contract that returns the contracts the human holds for
  Launch, and `propose-spec`, which takes that returned observation and the
  tools it names from it. Plan is one act node, under its own profile, that uses these itself and never delegates. It returns the Spec as it
  now stands with a `note`, its reply, which the conversation shows; a Spec
  that asks something new of Launch (goal, sources, inputs or tools) becomes
  the turn's revision, and the reply names it. Plan may cite only what it was
  passed and propose only catalogue capabilities. Its goal is segments:
  quotes of the message, the conversation or the latest Spec, by id, and words
  of its own. The Spec records each segment's source and its cite coverage,
  and every quote is checked against what the root holds; Plan's words lower
  coverage, and the human's own edits never do. A quote of a Spec counts only
  where a human wrote or cited it. Every revision builds on the last: saving
  diffs the edit against it, so kept words keep their source and only changed
  ones are the human's. A failed turn is reported as Plan's reply and leaves
  the spec unchanged; a turn the host stopped in is reported, never rerun. The
  turns are kept in the conversation's work file, and are its transcript.
- `launch-spec`: `conversation` and the `specRevision` to run, any revision
  the root holds. The host passes that revision's Spec, as the root holds it, with exactly the
  capabilities it proposes, and its tree runs under the Spec's profile. Neither the request's text nor its tools supply
  the goal or the grant. The catalogue remains the host's tree allowlist, and
  the provider follows the conversation. Launch runs once, on a registry
  restricted to those capabilities; its Spec and cited inputs enter the tree
  as its root's endowment.

All three also require `requestId` and `uiIncarnation`. A save from a stale
revision is refused. Duplicate admitted requests do not repeat work. Each execution
records its conversation, tree identity, spec identity, and revision before
starting. Only one execution per conversation tree may run at once; the next
spec revision can be edited while it runs. A restart marks interrupted tree
runs failed and never reruns them automatically. Deletion waits for both chat
and tree work to finish. Finished tree run records remain in the tree archive.

Conversation snapshots carry the current spec and latest execution reference;
revision history stays on disk. Chat receives the saved spec and last execution
result as separate context observations on its next turn. In the desktop app, each bundle is presented as a **Mission**. Its workspace
has Plan, Spec, and Launches tabs. **Plan** is the conversation with Plan; a
reply that made a revision links to it. **Spec** shows the spec as it stood
at a revision, the latest unless a reply's was chosen, each part marked by
its source, with its coverage, tools and tree profile; **Edit** saves a revision built on
the latest, and **Launch** passes the revision shown. Changing the tools is
an unsaved edit.

Selecting a conversation resolves its exact execution reference, rather than
choosing the newest run globally. Before its first run, the selected conversation shows a tree anchor labeled
Awaiting spec or Ready to launch, with a background glow derived from its tree
identity. Selecting another conversation changes that glow immediately.
Its accepted tree frames then drive the Metal sky: star nodes, growing branches, and activity lights. Positions depend on
node and tree identities and stay fixed as siblings arrive. Switching trees
crossfades their constellations without replaying historical activity. Light
is subdued under the transcript; Reduce Motion draws a static state. The Trees
window retains detailed inspection and replay, including standalone runs.

The Mission redesign archives previous conversations on the first desktop
launch for each workspace. The archive is persistent desktop navigation state,
keyed by workspace; it does not delete or move runtime files and does not
cancel active work. Restore brings the same conversation/spec/tree identities
back to the active Mission list. New Missions are not archived on later
snapshots or reconnects. Procedure constellations are no longer mounted in the
main window; procedure and model selection live in Mission settings.
