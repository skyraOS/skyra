# Living wallpaper implementation — 2026-10-07

Branch: feat/living-wallpaper-ui (switch succeeded with local changes preserved).

## Implemented
- Desktop: removed permanent left procedure/history rail and full-window procedure hit targets; centered transcript, quiet reading matte, compact History/New/Procedure/Tools controls.
- Native history sheet: all conversations/procedures and legacy records, local transcript search, procedure filter, stable ordering while open, explicit refresh for newly arriving conversations, separate guarded delete action, native Escape dismissal and initial search focus.
- Desktop rendering: existing instanced Metal scene through LivingSkyView, 15 FPS foreground target, paused while inactive/hidden/history-open/reduced-motion, explicit static draw. Independent running/working animation clocks removed; textual Working status retained.
- Desktop transcript: reader-following gate and Latest messages action; assistant text mostly unboxed; recorded screenshot decoding delayed until disclosure opens.
- Web: no mounted sidebar; integrated native modal history, local search, stable row order, guarded host actions; scope-unavailable disclosure; no per-message backdrop blur; focus-visible and reduced-motion styles.
- Web renderer: capped ~15 redraws/sec, no continuous hidden/static RAF loop, live reduced-motion listener, ResizeObserver instead of 500ms layout polling, frame delta clamp, selection/incarnation stamp reset, avoids rescanning unchanged entries.
- CI added for frontend typecheck/lint/build/client tests and native build/test. CI results not obtained locally.
- Blender matte prototype refined separately, not promoted to production. Preview visually inspected: muted cyan sphere and indigo rings; still too object-like for the reading column. Exports: job-9aapoqqe/scene.blend, job-far4v35s/asset.glb + asset.usdz, job-e9cktarn/preview.png under ideas/design/blender/artifacts. 2,368 triangles; native USDZ load remains unverified.

## Actual verification
- swift.test: exit 0, latest report ideas/checks/skyra-check-2772948321.json. Nine XCTest passes; Swift Testing reports 53 tests with three live-backend tests skipped. Two added history regression tests pass.
- swift.build (release): exit 0, ideas/checks/skyra-check-2620830376.json.
- git diff --check: exit 0.
- No sandbox denial or output truncation on these checks.
- Browser opened localhost:3000; actual screenshot shows ERR_CONNECTION_REFUSED. This is NOT visual validation of the implementation.
- Observed Chrome/Penpot tree was truncated; its screenshot is a previous design, not this implementation.
- Check tool removes temporary build products. Installed desktop binary was not rebuilt/launched by these recipes. No tool is advertised for packaging/launching a freshly built binary, running npm, or interacting with browser controls.
- Git remote -v refused (subcommand unavailable, exit -1). git log -1 --format refused (option unavailable, exit -1). These are inspection restrictions, not inability to switch/edit/commit/push.

## Still required before merge
This is an implementation candidate, not completion of every acceptance gate in performance-and-tests.md. Run frontend CI; build/launch desktop and browser app for screenshot/keyboard/screen-reader review at compact/wide/zoomed sizes; verify sheet deletion focus and composer/send flow against a live host. Profile baseline/candidate CPU/GPU/input/energy rather than claiming measured gains. Verify static scene after resize and dynamic accessibility changes. Search remains synchronous (no debounced/indexed search yet); decorative constellation view, compositor-focus feedback, optional floating log/3D-pilot integration, and full reconnect light-reset validation are deferred. Production graphics assets remain unchanged. Existing floating-log storage is not deleted, but the centered UI currently has no arrange gesture; existing Graphics menu settings do not mount a pilot in this layout.

## Research reconfirmed
Apple MTKView documentation explains paused/requested-draw and preferredFramesPerSecond controls:
https://developer.apple.com/documentation/metalkit/mtkview/
web.dev animation performance guidance:
https://web.dev/articles/animations-and-performance/
W3C motion prevention:
https://www.w3.org/WAI/WCAG22/Techniques/css/C39
These justify lifecycle pacing, avoiding unnecessary paint/compositing and static preference handling; they do not establish measured application performance.

Unrelated pre-existing controller, Composer and GrantSheet changes were preserved and are excluded from the UI commit.

Final cleanup: removed the unused Graphics pilot menu from SkyraApp rather than leave controls that have no effect. The opt-in pilot code/assets remain in the repository for future reviewed integration.

Research retrieval caveat: web.dev article content fetched successfully. Apple page returned a JavaScript requirement; its linked Markdown fetch failed with unsupported content type. Apple API behavior is exercised by the compiling renderer, but the fetched page did not independently supply its full documentation. The 15 FPS ambient policy is our unmeasured design choice, not web.dev's general recommendation for interactive animations.
Final release build after app-menu cleanup: exit 0, ideas/checks/skyra-check-2483862360.json.
