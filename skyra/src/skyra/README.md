# Skyra source

`skyra-dna-19/` is the permanent base going forward. The disposable throwaway
generations that led here (`skyra-dna-v7`…`v14`, `skyra-dna-15`…`18.5`, the
CTU engine probe, and the older overnight lineage notes) have been moved to
[`../archive/throwaway-dna/`](../archive/throwaway-dna/) and are kept only as
historical reference — they are not built or maintained.

## Current work

- Base runtime: [`skyra-dna-19/`](skyra-dna-19/) — external-binary boundary
  (`root/` host + separate `codex/` provider actor over framed Skyra Moves),
  authenticated persistent parents, contamination fence, bounded tool migration.
  See its `README.md`, `DECISIONS.md`, and `issues.md`.
- Next increment (proposed, not yet implemented): a root-invokable `user/`
  executable with a browser frontend — see
  [`../../specs/2026-09-22/frontend.md/proposal.md`](../../specs/2026-09-22/frontend.md/proposal.md).

Module root is `skyra/` (module `skyra-v1`); build from within `skyra-dna-19/`.
