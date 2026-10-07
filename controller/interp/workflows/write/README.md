# Write and commit ideas

When the Skyra repository is the selected workspace, the launcher registers
`write` as this interpreter procedure. Its input stays `{path, text, append?}`:

1. `write.file` writes text into the granted `ideas/` directory.
2. `ideas.commit` receives the write observation through the park protocol.
3. The operator-side Git host stages `ideas/` and runs
   `git commit --only -m "ideas: save agent write" -- ':(top,literal)ideas'`.

The commit includes changes throughout `ideas/`, including untracked files
and deletions. Unrelated staged changes remain staged and are excluded from
the commit. Repository hooks and commit signing are disabled for this fixed
operation. Actors receive no Git execution or `.git` write grants.

The result preserves the write observation and adds `commit` (the hash),
`commit_unchanged` (identical content needs no new commit), or `commit_error`.
A refused write is returned without running Git. If Git fails, the written
file remains available and the failure is returned to the agent.

The Git host uses the repository's configured author identity. Other selected
workspaces keep the plain scratch writer and do not create Git commits.
Restart Skyra to load the new registry and procedure.
