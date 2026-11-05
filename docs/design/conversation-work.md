# Conversations, specs, and trees

Each conversation owns one spec identity and one tree identity. Chat invokes
its existing conversation procedure. Run submits a saved spec revision to the
apprentice's existing tree contract. Conversation messages are never used as
tree input implicitly.

`work.json` beside the conversation's runtime records keeps the operator's
spec revisions and the latest execution reference. This is authored input,
not runtime truth; execution views still come from replaying tree records.
Older conversations acquire deterministic empty spec/tree identities when
loaded. Invalid work metadata is not replaced with a fresh executable spec.

The webhost exposes two conversation-scoped, request-ledger operations:

- `save-spec`: `conversation`, `text`, and expected `specRevision`. Changed
  text creates a revision. Saving never starts execution.
- `run-spec`: `conversation`, expected `specRevision`, and selected `reach`.
  The host reads the saved text and invokes the tree contract with it. Text
  in the request does not supply the goal. Tools remain restricted to the
  host's existing tree allowlist, and the provider follows the conversation.

Both also require `requestId` and `uiIncarnation`. Stale revisions are
refused. Duplicate admitted requests do not repeat work. Each execution
records its conversation, tree identity, spec identity, and revision before
starting. Only one execution per conversation tree may run at once; the next
spec revision can be edited while it runs. A restart marks interrupted tree
runs failed and never reruns them automatically. Deletion waits for both chat
and tree work to finish. Finished tree run records remain in the tree archive.

Conversation snapshots carry the current spec and latest execution reference;
revision history stays on disk. Chat receives the saved spec and last execution
result as separate context observations on its next turn. In the desktop app,
**Spec** opens an editor. **Use latest reply** copies a chat reply into that
editor; **Save spec** publishes it, and **Run saved spec** starts execution.

Selecting a conversation resolves its exact execution reference, rather than
choosing the newest run globally. Before its first run, the selected conversation shows a tree anchor labeled
Awaiting spec or Ready to run, with a background glow derived from its tree
identity. Selecting another conversation changes that glow immediately.
Its accepted tree frames then drive the Metal sky: star nodes, growing branches, and activity lights. Positions depend on
node and tree identities and stay fixed as siblings arrive. Switching trees
crossfades their constellations without replaying historical activity. Light
is subdued under the transcript; Reduce Motion draws a static state. The Trees
window retains detailed inspection and replay, including standalone runs.
