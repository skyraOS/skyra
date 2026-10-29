# Think UI investigation

Initial investigation (before implementation): inspected live Skyra window (1280x758) and current working-tree source. No source edits or runtime profiling performed. Existing modifications were preserved.

## Confirmed structural problems
- desktop/Sources/Skyra/Sky/Roots.swift: RootConstellation.size is always 220x150; actor spacing depends on actor count but each actor's tools fan across the same fixed radius. There is no collision avoidance or label-aware spacing. Think's large tool inventory overlaps. RootSky scales this only 1.1x.
- ContentView.swift: floating log default width 720, height window minus 200. ConversationView adds 48 horizontal padding; FloatingLog adds another 56. Resizing the log cannot increase the inner reading column beyond 720. Dragging anywhere in the conversation competes with selection/scroll interaction; resize affordances are invisible.

## Performance risks, not measured root causes
- controller/hosts/web/main.go broadcast builds snapshots containing all conversation entries, including withCalls. Connection.swift Session is MainActor and synchronously JSON-decodes incoming snapshots; snapshot path also JSONSerialization-parses raw data for request handling. Large histories/screenshots can therefore consume UI-thread time.
- SkyRenderer.draw performs scene generation, blocking gate.wait(), drawable acquisition, text texture lookup and uploads in the draw callback. GPU backlog can cause blocking; thread and duration need profiling before attribution.
- ContentView observes entire current entries for followCalls; CallLights.follow scans all calls each change. ConversationView constructs an enumerated array and Message.body parses markdown on body evaluation. Increasing transcript size increases this work.
- Current ContentView uses LivingSkyView, already capped at 15 FPS and paused when inactive/hidden/history shown. Older SkyView has display-rate pacing, but it is NOT the active root view; do not claim the live sky runs at 120 FPS from that code alone.

## Suggested implementation order
1. Compact constellation overview with tool counts; expandable, scrollable procedure inspector for actors/tools. Labels on hover/selection rather than every tool always visible. Keep the Think root and conversation working-set navigation available.
2. Explicit Focus/Expand control for the transcript; clear resize/drag handles and quieter background. Keep readable line length while giving tools/long content more room.
3. Decode off the main actor, then apply snapshots on main actor with generation/revision checks; send history separately or per-conversation deltas.
4. Instrument decode/apply/scene-generation/GPU-wait durations and snapshot byte/entry counts. Profile idle, typing, scrolling, tool bursts, screenshot-heavy histories and resizing; compare sky enabled vs paused.
5. Replace blocking renderer admission with safe frame skipping if profiling shows GPU waits, retaining buffer-slot correctness. Cache markdown by immutable text and update call lights incrementally if measured costs justify it.

No measured CPU/GPU timings are available through the provided tools; these findings establish code paths and likely causes, not a proven ranking of runtime bottlenecks.


## Implementation — 2026-10-07

- The sky now draws root/actor stars with tool counts, using actor-aware vertical spacing. A procedure inspector has expandable actor sections and a scrollable complete tool inventory. Tool traffic lands on its actor in the overview.
- The transcript has a dedicated drag bar, visible resize edges/corner, Expand/Reading width and Focus/Exit focus controls. Its placement stays between launches; the content no longer has a move gesture. Focus pauses the sky and preserves a readable text column. The header's Context menu keeps procedure/model/working-set navigation available while focused or expanded. Bounds follow the composer's actual position.
- WebSocket decoding runs in a detached task, with ordered receive/apply and connection-generation checks after decoding. The request ledger uses the same typed decoder traversal; the second JSONSerialization pass is removed.
- The desktop negotiates per-conversation replacement deltas. Each frame includes a baseline revision and ordered membership list; unchanged histories are reused, deletions clear missing membership, and a missing baseline requests a full resync. Other clients keep full snapshots. Snapshot broadcasts are serialized to preserve revision order. The host still constructs full views and enforces the existing frame limit before delta selection, so server-side history construction remains a profiling candidate.
- Instruments signposts cover decode/apply, scene generation, GPU admission wait, drawable acquisition and encode/upload, with snapshot byte/entry events. `SKYRA_PROFILE_UI=1` adds webhost snapshot build/size logs and an opt-in desktop diagnostic workload.
- The transcript iterates entry indices without materializing an enumerated array. Measured repeated markdown parsing justified a bounded immutable-text cache (512 entries, 4 MiB text cost). Call lights remain unchanged because their measured scan cost was small.

### Measured diagnostic workload

Xcode debug build on this Mac, synthetic data; 20 repetitions per workload and 100 for markdown. These are wall-clock means from a suite running alongside other tests, not live interaction or GPU profiles.

| Workload | Mean |
| --- | ---: |
| Decode 20 entries / 8,219 bytes | 0.085 ms |
| Decode 1,001 entries / 2,501,445 bytes, including a synthetic screenshot payload | 7.075 ms |
| Call-light scan, 1,001 entries | 0.202 ms |
| Unique markdown message | 0.351 ms |
| Cached markdown message | 0.003 ms |
| Warm scene generation at 1280×758 | 5.039 ms |
| Warm scene generation at 1800×1000 | 8.573 ms |

The scene workload advances frames, but excludes drawable acquisition, uploads and GPU execution. The screenshot workload measures JSON payload handling, not image decoding/display. Focus pauses recurring sky frames; comparing actual typing/scrolling/resizing with sky enabled vs paused still requires live Instruments capture. There is no measured GPU-admission backlog, so frame-skipping admission was not introduced. This preserves buffer-slot correctness until a real capture justifies that change.

Validation: the desktop suite and diagnostic workload passed (9 XCTest tests plus 58 Swift Testing cases, with the three environment-gated live cases skipped in that run). Webhost tests passed under the race detector, and `go vet ./hosts/web` passed. Regression coverage includes 100-tool overview sizing, delta history retention/order/deletion, stale or missing baselines, and acknowledgement-before-transcript behavior.


The isolated live Think round-trip also passed: first submission, tool/Boundary results, a second turn in the same conversation, and deletion over the actual WebSocket. Its test fixture now selects Think explicitly, independent of destination ordering. The fake backend used temporary storage and was stopped after verification.
