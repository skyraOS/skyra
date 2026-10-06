# Repository review fixes — 2026-10-06

This records the implementation and validation following [the original review](review-2026-10-06.md). The original review is a historical baseline.

| Finding | Change | Evidence |
|---|---|---|
| 1. Premature and duplicate discharges | Replay requires terminal, distinct discharges belonging to distinct slots of the correct round. New genesis version 2 histories commit the full emission; replay derives every expected slot and checks completeness and order. | Regression tests reject unfinished, duplicated, omitted, reordered and missing-evidence resolves; existing website bundles still replay. |
| 2. Unauthenticated ingress | Every launcher generates a token unless supplied. The host refuses empty credentials, non-loopback Host headers and unauthorized browser origins. Browser subprotocol credentials work only with the configured origin; native clients use Bearer headers. | Handshake tests cover empty/wrong tokens, foreign origin/Host and the authorized browser path; live Swift rejects a wrong token. |
| 3. WebSocket allocation panic | Bounds precede allocation; masked, final frames, canonical lengths, valid opcodes/text and control sizes are checked. Header and write deadlines bound stalled connections. | Malformed frame regressions and a valid masked frame test pass. |
| 4. Interpreter temporary-file collisions | Unique temporary files and exclusive immutable publication replace shared `.tmp` paths. Shared nodes must have identical bytes; a contract has one winning initial procedure publication. | 32 concurrent procedures share a node successfully; competing installations have exactly one winner. |
| 5. Interpreter durability | Actor publications sync bytes before linking/renaming and sync their directory and actor-state directory before returning. Position/remembering state is written before an emission is returned to the runtime. | Interpreter race tests, existing reconciliation tests, and all procedure demos pass. This is not a power-loss certification. |
| 6. Invisible resume/runtime failures | Resume waits for validated custody and reports child exit failures, with a deadline and descriptor cleanup. Drive records I/O interruptions separately from terminal discharges; the host reconciles these to `delivery-unknown` and refuses late parks into a failed interaction. | Failing/nonexistent resume executable tests, interrupted Drive publication test and host pending-to-failed reconciliation test pass. No external effect is automatically repeated. |
| 7. Mock/production request drift | Durable admission precedes execution. An identity hash covers operation, conversation, incarnation and text without retaining message content in the ledger. Snapshots include pending IDs and request statuses. Restarts retain admitted IDs and label unfinished outcomes unknown. | The same protocol harness passes against mock and real webhost with fake Codex: admission, completion, busy, conflict, stale UI, reconnect, close and delete. Durable restart/identity regressions pass. |
| 8. Lost client drafts | Browser and desktop keep text until matching admission is established. Rejection/unknown delivery preserves the draft and retry identity. Browser sends are refused offline; stale socket callbacks cannot corrupt a replacement connection. | Three browser regressions, Swift draft regressions, and all three live desktop/backend tests pass. |
| 9. Incomplete verification/docs | `skyra test` runs `scripts/check.sh`, covering both Go modules, race/vet, configured ESLint, types/build, client/protocol checks, Swift including live fake-backend tests, website bundles/WASM and demos. macOS CI runs the same script. Demos fail on replay errors and unexpected terminal outcomes. Current docs describe built fan-out, authenticated launch and recovery limits. | Full aggregate checks passed; subsequently added browser/live-desktop checks and stricter demos also passed. CI configuration is added; no remote CI run was triggered. |

Additional hardening removes process-global `Fchdir`, tests concurrent claims, publishes tree objects exclusively, restricts new runtime files/directories to the owning user, bounds target diagnostics/output and park messages, closes received descriptors on decoding failures, and uses random conversation IDs.

Compatible frontend dependency patches were installed, including Next, WebSocket and image-processing dependencies. The Next PostCSS dependency uses the patched direct PostCSS range through an override. `npm audit --omit=dev` reports zero vulnerabilities. The full audit still reports nine development-tool advisories (seven high, two moderate), involving the Tailwind 3/ESLint glob and selector dependency trees. Those require a separate dependency migration; no forced framework/CSS major migration was performed.

## Validation

- Controller/TUI/website Go tests with the race detector and vet passed.
- Separate interpreter Go race tests and vet passed.
- Frontend lint, typecheck, production build, client tests and both protocol implementations passed.
- All 41 Swift tests passed with a running real webhost and fake Codex, including the three formerly skipped live tests.
- Every tracked website run bundle replayed; the WASM build passed.
- Every controller demo replayed and matched its expected terminal outcome.
- `git diff --check` passed.

## Limits that remain

Old histories cannot retroactively supply the full emission evidence required of version 2 histories. They receive the stronger terminal-state, uniqueness and ownership checks, but full round coverage remains unprovable from their existing artifacts.

Recovery reporting does not revive lost process/socket custody, certify power-loss recovery, detect every hard-killed descendant immediately, or justify repeating external effects. Failure publication itself can fail when storage is unavailable. The operator registry, runtime and adapters remain trusted; replay proves derivation and hash consistency with trusted anchors, not truth of external observations.

No real model inference, deployment, comparative task-success/cost benchmark or long-history scalability benchmark was performed. Those research/operational recommendations from the review remain work to evaluate, rather than bugs resolved by these changes.
