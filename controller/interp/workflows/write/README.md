# Write and commit ideas

For the Skyra repository selection, `write` is this interpreter procedure:

1. `write.file` saves `{path,text,append?}` under the granted `ideas/` directory.
2. `ideas.commit` receives the successful write observation through park custody.
3. The host verifies the file hash, uses a private index, and updates the separate
   `refs/heads/skyra/ideas` ref with only the observed file.

The initial notes commit descends from the current HEAD; later note commits
extend the notes ref. No checkout occurs. The user's current branch, HEAD and
index remain intact, and other untracked or staged notes are excluded. The
commit author is `Skyra agent <skyra@localhost>` and the committer uses the
configured user identity. Hooks and signing are disabled; nothing is pushed.
Actors receive no Git execution or metadata write authority from this procedure.

The result preserves the write observation and adds `commit`, `branch`, and
`commit_unchanged` when the same content needs no new commit. `commit_error`
reports failures; the saved file remains available. A refused write does not
run Git. Runtime conversation identity records before/after hashes and commits
for `roots` with `changes:true`.

Other selected workspaces retain plain scratch writes without commits.
Rebuild and restart Skyra, then use a new conversation to load this workflow.
