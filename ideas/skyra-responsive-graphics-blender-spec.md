# Skyra responsive graphics and Blender integration specification

Status: proposed implementation specification, not implemented or benchmarked.
Research and repository inspection: October 7, 2026.
Scope: Skyra's native desktop visual experience, Blender authoring pipeline, bounded fine-grained authoring tools, and a secondary web adapter. This is a full spec for this initiative, not a replacement specification for Skyra's cognitive runtime.

## 1. Executive decision

Keep SwiftUI for the interface and the existing Metal sky for ambient effects. Use Blender offline to author a small sculptural Skyra core, reusable meshes, textures, and optional animation assets. Add a renderer adapter rather than running Blender inside the desktop app. Start with transform-driven interaction on one asset; prove loading, transparency, lifecycle, and performance before introducing rigs or morph targets.

Native candidate: Blender → USDZ → RealityKit. This is a hypothesis to validate, not a proven fit. The current app targets macOS 14, so do not introduce RealityView unconditionally or raise the deployment target silently. Investigate a compatible non-AR ARView wrapper and supported lifecycle controls on the actual SDK. If that fails acceptance, retain Metal with baked graphics while evaluating a separate mesh-rendering pass. Web candidate: Blender → GLB → Three.js, deferred until the native pilot passes.

Blender makes the asset. Skyra supplies responsiveness, application state, picking, camera framing, and runtime effects. Exported assets must never determine authority, tool execution, or conversation truth.

## 2. Verified baseline

Repository evidence, read directly:

- `desktop/Package.swift`: Swift tools 5.10, macOS 14 deployment target; executable and test targets, no declared resources.
- `desktop/README.md`: native client of the conversation protocol; no runtime authority; existing Metal sky, background throttling and accessibility behavior described.
- `desktop/Sources/Skyra/Views/ContentView.swift`: SwiftUI ZStack with SkyView beneath controls; 232-point sidebar; reading width 720; minimum window 640×420; pointer in named sky coordinate space; floating log geometry; Reduce Motion and visibility passed into the sky.
- `desktop/Sources/Skyra/Sky/SkyView.swift`: SkyModel owns pointer, reading column, log rect, roots, selection and ambient simulation; Metal view hit testing returns nil; background rate is 10 FPS.
- `desktop/Sources/Skyra/Sky/SkyRenderer.swift`: existing full-screen background and instanced sprite/segment pipelines, triple in-flight buffers, bgra8Unorm target, premultiplied blend setup, exposed offscreen test hooks. This is not a general 3D renderer.
- `desktop/Sources/Skyra/Sky/Theme.swift`: dark sRGB palette, differentiated amber refusal and red failure.
- `desktop/Sources/Skyra/Protocol.swift`: conversation phases ready/pending/failed/closed; call states running/returned/fault; boundary and contract faults count as refusals; snapshots carry incarnation/revision and root destinations.
- `frontend/package.json`: Next.js/React frontend; no current Three.js dependency.

No asset was loaded, Blender installation checked, desktop build run, Penpot operation attempted, or renderer benchmark performed for this document. Source changes below are proposals.

## 3. Documentation findings and evidence limits

Official documentation consulted through retrieved page/search content:

[D1] Blender glTF manual, version 5.1: https://docs.blender.org/manual/en/5.1/addons/import_export/scene_gltf2.html
[D2] Blender development glTF manual: https://docs.blender.org/manual/en/dev/addons/scene_gltf2.html
[D3] Blender USD manual, version 4.2: https://docs.blender.org/manual/en/4.2/files/import_export/usd.html
[D4] Blender USD limitations, version 4.4: https://docs.blender.org/manual/en/4.4/files/import_export/usd.html
[D5] Blender baking manual, version 5.0: https://docs.blender.org/manual/id/5.0/render/cycles/baking.html
[D6] Blender scripting security manual: https://docs.blender.org/manual/id/latest/advanced/scripting/security.html
[D7] Blender command line, version 4.2: https://docs.blender.org/manual/id/4.2/advanced/command_line/arguments.html
[D8] Apple RealityView: https://developer.apple.com/documentation/realitykit/realityview
[D9] Apple entity loading: https://developer.apple.com/documentation/realitykit/loading-entities-from-a-file
[D10] Apple loadModel API: https://developer.apple.com/documentation/realitykit/entity/loadmodel(contentsof:withname:)
[D11] Three.js documentation index: https://threejs.org/docs/

Export findings: glTF carries meshes, recognized PBR materials, textures, hierarchy and supported animation. GLB packages its related data in one binary. Blender's material graph is not an arbitrary portable shader program. Animation support and action/NLA grouping vary by exporter version; development documentation includes experimental animation-pointer behavior that must not become a baseline requirement. [D1,D2]

USD findings: Blender exports supported scene objects and can package USD plus texture dependencies by using a `.usdz` output extension. Its Preview Surface material conversion is an approximation. Skeletons and blend shapes have export options, but runtime interoperability still requires an asset test. Older documented limitations are version-specific, not proof of today's behavior. [D3,D4]

Baking findings: Cycles can bake shader/lighting information to image textures; mesh UVs and an active bake target are required. Bake procedural appearance for portability; do not assume a static bake preserves a live simulation. [D5]

Security findings: `.blend` files may contain scripts and Python drivers. Disabling automatic execution does not make explicit Python execution safe. Blender provides background/script invocation flags and an exit-code option for script exceptions. Apply OS boundaries in addition to Blender flags. [D6,D7]

Loading findings: Apple's documented loadModel operation flattens a loaded hierarchy. A named-node interactive asset should use a hierarchy-preserving entity loading path, validated against the deployed SDK, rather than flattening it by default. [D9,D10]

Research limits: several direct English/latest and 4.5 manual fetches failed; some available content was search-extracted and some documentation was localized. Apple pages expose much content through JavaScript. Installed Blender/Xcode versions and exact macOS API availability remain unverified. Pin the actual toolchain during Phase 0; do not treat a development manual as the installed release.

## 4. Product goals and non-goals

Goals:
1. Make Skyra's core respond immediately to pointer, window geometry and observed conversation activity.
2. Preserve calm readability, existing stars/constellations and native input behavior.
3. Make individual assets and properties inspectable and editable through bounded tools.
4. Produce repeatable, reviewable asset builds with rollback and visual evidence.
5. Keep the application fully useful with graphics disabled or unavailable.

Non-goals for v1: full 3D UI, AR, replacing the Metal nebula, embedding Blender, arbitrary user scripts, physically accurate fluid simulation, unrestricted mesh editing by a model, cloud asset fetching, animated materials requiring experimental extensions, or claims that animation reveals an agent's hidden cognition.

Success: one Blender-authored core follows the pointer and changes presentation from existing protocol observations without blocking conversation controls, increasing authority, or requiring Blender on an end user's computer.

## 5. Skyra experience specification

### 5.1 Visual language

Retain the palette: sky #080A0E; ink #D6DBE5; accent #7EC4CC; operator green #60C4A6; dim #788092; refusal amber #DEA868; failure red #E27070. These are derived from Theme.swift. Dark-only v1; no second hardcoded palette in the 3D adapter.

Core design proposal: a restrained rounded mineral/orb form, subtle teal surface detail and two sparse orbital arcs. No face or cartoon mascot in the pilot. Use silhouette and slow orientation changes rather than noisy bloom. Warm/violet accents may echo the existing sky. Avoid glass/transmission dependencies in the first export.

The orb is presentation, not a new authority-bearing object. Existing root constellations continue to represent selectable procedures. Do not replace them with ambiguous decorative planets.

### 5.2 Layering

Back to front: Metal ambient sky → bounded decorative 3D viewport → existing constellation labels/control overlays → sidebar, header, conversation, composer → sheets and accessibility controls.

The first 3D viewport is a small local region, not a second full-window scene. If transparent composition is unreliable, use a self-contained inset with the sky background color or fall back to the current core. Never cover conversation text with a bright decorative object.

Use existing layout measurements as exclusions: header, sidebar, reading column, actual floating log, composer, root hit regions. While the log is being arranged, suppress hover reaction and hold decorative framing steady.

### 5.3 Responsive layout proposal

All measurements below are points, not backing pixels.

- Available free area ≥320×240: core diameter 120–180, with 24-point margin.
- Narrow/free area below that: shrink to 64–96 or hide; do not compress text.
- Minimum existing 640×420 window remains supported.
- Wide window: place in a free side region, never center it underneath the reading column.
- Empty conversation: larger core may occupy free central space; active conversation: recede into a side region.
- If no unoccluded region exists: hide 3D and retain existing sky.

Compute placement from geometry, not just width breakpoints. On resize recompute the viewport and camera; do not mutate asset scale repeatedly to compensate for camera errors. Maintain bounding-sphere framing and margins across aspect ratios.

### 5.4 Interaction

Decorative v1 is click-through and not focusable. Pointer tracking comes from ContentView's existing observation, not an independent event interceptor. Leaving the window eases orientation toward neutral. Dragging/selecting text never affects application actions through the orb.

Proposed maximum yaw ±10°, pitch ±6°, eased over approximately 180 ms. Idle drift is optional and small. Hover changes highlight, not physical position toward the cursor. Avoid sudden acceleration, zoom or camera shake.

Future interactive mode must use a native keyboard/VoiceOver-accessible button with an explicit action, label and minimum 44×44 hit area. Clicking graphics must never silently invoke a tool or grant access.

### 5.5 Settings and accessibility

Settings: graphics off/static/reactive; quality automatic/low/standard; reset appearance. Defaults reactive where supported, automatic quality. Reduce Motion removes continuous rotation, orbit, pointer tilt and pulses; use static appearance changes and textual status. Hidden windows stop graphics work. Inactive windows cap presentation at 10 FPS or suspend if the chosen engine cannot honor pacing economically.

Decoration is accessibility-hidden. Essential status remains in native text and iconography; color is supplemental. Respect increased contrast and transparency preferences through an opaque fallback if needed. No audio or flashing effects. App operation must remain identical with graphics off.

## 6. State projection contract

Add a pure `GraphicsProjection` from observed session/backend state. No new backend graphics authority, no reading `.skyra/`, and no new protocol messages in v1.

Suggested normalized state:

```
GraphicsInput {
  connection: connected | connecting | offline
  conversationID: optional string
  phase: empty | ready | pending | failed | closed
  runningCalls: bounded integer
  recentOutcome: none | returned | refused | fault
  eventKey: optional stable key
  pointer: optional normalized vec2
  viewportPoints: vec2
  exclusionsPoints: list<rect>
  visible: bool
  active: bool
  reduceMotion: bool
  graphicsMode: off | static | reactive
}
```

Map existing protocol, not invented cognition:

| Observation | Presentation proposal | Meaning |
|---|---|---|
| No current conversation | neutral core, faint idle | no selected conversation |
| ready | steady teal | ready for input |
| pending | restrained slow orbit | turn pending, not proof of thinking |
| running call | short arc/brief highlight | recorded call running |
| returned call | brief green accent | call returned, not task success |
| boundary/contract refusal | amber notch | boundary/contract refused the call |
| other call fault | red notch | call fault |
| failed phase | persistent small red marker | conversation failed |
| offline | dim core and textual connection status | disconnected, not failed agent |
| closed | neutral static | conversation closed |

Precedence: invisible/off → static preference/Reduce Motion → offline → failed → refusal/fault cue → running → pending → ready/empty. Connection overlay does not erase the selected conversation's recorded phase.

Use `(incarnation, conversationID, call.stamp, state)` as a cue key. Revisions select newer snapshots; do not fire completion cues for every historical call on connection. Establish a baseline on initial snapshot and on conversation selection. Clear old cue queues on incarnation changes. Coalesce concurrent calls; cap at three visual cues/sec. Presentation events expire after 600 ms and never create durable runtime facts.

Simulation uses a monotonic clock, clamped dt (proposed ≤1/30 s) and analytic easing. Resume does not replay a long backlog. Tests inject a fixed clock and seed. UI animation is not part of the runtime's byte-identical replay guarantee.

## 7. Runtime architecture

Proposed interfaces (conceptual, not compiled API):

```
GraphicsProjection.project(observation) -> GraphicsInput
GraphicsLayout.place(input, assetBounds) -> GraphicsPlacement?
GraphicsController.update(input, monotonicTime) -> GraphicsFrame
GraphicsRenderer.load(validatedBundle) async -> LoadResult
GraphicsRenderer.apply(frame)
GraphicsRenderer.setLifecycle(visible, active, motionMode)
GraphicsRenderer.capture(testConfiguration) async -> Preview
GraphicsRenderer.unload()
```

The controller is engine-independent: outputs root orientation, arc angle, scale modifier, intensity, tint and clip intent. The adapter maps these onto supported engine operations. Feature capability flags distinguish transforms, materials, clip playback and morph weights. Missing features degrade individually; no unchecked assumptions about morph or material API support.

Native pilot components:
- `GraphicsProjection.swift`: pure state conversion and cue deduplication.
- `GraphicsLayout.swift`: exclusion-aware placement and framing.
- `GraphicsController.swift`: motion and lifecycle policy.
- `GraphicsAssetManifest.swift`: decoder, limits, version checks.
- `GraphicsView.swift`: SwiftUI/native host with click-through behavior.
- `RealityKitGraphicsAdapter.swift`: isolated compatibility experiment.
- Bundled asset resource directory and packaging verification.

Do not splice a 3D engine into SkyPaint's sprite types. Keep SkyModel's existing star generation/hash semantics untouched. An alternate Metal mesh implementation needs explicit vertex/index buffers, depth attachment, lighting/material shader, camera math, loading/conversion, and test coverage; it is not just another sprite.

All engine mutation must use the executor/thread required by the framework. Load asynchronously with cancellation and generation tokens so a stale load cannot attach to another window/session. Release scene subscriptions, timers, GPU resources and caches on teardown. Graphics errors never throw through message submission.

## 8. Blender asset specification

### 8.1 Source scene

One `.blend` master per asset family. Required export collection `SKYRA_EXPORT`; reference/sculpt/lighting collections excluded. Suggested hierarchy:

```
SkyraCoreRoot
  CoreMesh
  ArcInner
  ArcOuter
  AccentMesh
```

Stable unique names become manifest bindings. Asset root at origin; author in meters with a 1-meter normalized reference diameter. Record actual exported bounds/unit metadata and convert once in the adapter. Do not assume Blender's Z-up equals runtime up: test an asymmetric axis fixture, document conversion, and avoid double rotation.

Apply scale/rotation on static meshes before final export. For rigs, establish rest pose and transforms before skinning and validate any later transform application. Keep topology stable for deformation. Reject accidental lights/cameras/world dependencies in the export collection; runtime owns framing and lighting.

### 8.2 Pilot limits (design budgets, not measured engine limits)

- Standard LOD ≤20,000 exported triangles; low LOD ≤5,000.
- ≤4 material slots and ≤8 mesh nodes.
- Textures normally 1024²; maximum 2048² by exception.
- Export package ≤5 MiB target, ≤10 MiB hard pilot cap.
- Decoded texture budget ≤32 MiB including estimated mip overhead.
- No rig in Phase 1; later rigs ≤32 deform bones, ≤4 influences per vertex.
- No volumetric export, hair, cloth, fluid, topology-changing animation or UDIM dependency in v1.

Measure after export: seams and hard normals may increase runtime vertex counts. Budgets are checked on the exported asset, not Blender's displayed base mesh.

### 8.3 Materials

Portable pilot: Principled-based metallic/roughness PBR with base color, roughness, metalness, tangent normal and modest emissive contribution. Bake procedural surface detail to textures with UV padding. Store base/emissive textures as color data and normal/roughness/metalness as non-color data; confirm runtime channel packing and normal convention per adapter.

Avoid transmission, volumetric appearance, complex clearcoat, custom shader graphs and animated material exports in v1. Dynamic tint/intensity is runtime code on supported material parameters. If unsupported, substitute native materials or a separate accent mesh. Do not promise Blender-render fidelity.

Do not bake key light/specular highlights into base color; it prevents correct response to runtime lighting. Create a neutral reference render plus app-matched lighting reference. Preserve an unlit diagnostic version for export debugging.

### 8.4 Motion authoring

Phase 1 motion is transforms authored by runtime, avoiding clip compatibility risk. Phase 2 optional clips: `idle`, `pending`, `returned`, `refused`, `fault`; neutral entry/exit poses; 30 FPS authoring; loops checked for seams; transitions blended or eased by the adapter.

Validate action/NLA export under the pinned Blender version; enumerate clips in each actual output. Never infer clip presence from a source action name. Bake constraints/drivers where required by the exporter using an explicitly trusted build process. Rig animation, shape keys and named clip playback each require separate native and web capability tests. A successful GLB morph test does not establish USDZ/RealityKit morph support.

## 9. Export and packaging pipeline

Stages: inspect source → preflight → duplicate/export staging scene → bake if required → export USDZ and optionally GLB → validate each output → runtime smoke test → capture previews → publish immutable version → promote approved manifest.

1. Pin Blender binary/version, exporter configuration, script hash and dependency versions. Record OS and GPU for renders.
2. Source preflight verifies allowed collection, names, transforms, units, bounds, UVs, external image dependencies, material slots and banned features.
3. All outputs go into a fresh job directory; never export over an approved artifact.
4. Export selected/staged objects with explicit options appropriate to the pinned exporter. Save a machine-readable option record.
5. Export GLB directly from Blender for web; export USDZ directly for native. Do not chain GLB → USDZ conversion by default.
6. Validate hashes, package paths, references, counts, finite numeric data, bounds, textures and declared features.
7. Load the actual package in each runtime; verify hierarchy/node names, axes, dimensions, shading and available clips.
8. Render reference views at neutral, pointer extremes, pending and refusal/fault appearances.
9. Publish only if tests pass and a reviewer approves fidelity; retain prior version for rollback.

Example command shape, to be verified on the pinned installation:

```
/path/to/blender --background --factory-startup --disable-autoexec \
  --python-exit-code 1 --python /approved/export_skyra_asset.py \
  -- --source /granted/source/core.blend --out /granted/job/output
```

The trusted script parses arguments after `--`, opens the granted source with automatic scripts disabled, stages only permitted objects and calls exporters. The explicit script is executable code and must run under Skyra's recorded sandboxed tool boundary; storing it in ideas does not make it runnable by the current tools. Driver-dependent assets may fail with autoexec disabled: report that failure rather than silently enabling scripts. This command is a proposed developer/build interface, not an executed command.

### 9.1 Manifest contract

Illustrative JSON, not an implemented schema:

```json
{
  "schemaVersion": 1,
  "assetID": "skyra.core",
  "assetVersion": "0.1.0",
  "sourceHash": "<sha256>",
  "toolchain": {"blender": "<pinned version>", "exportScriptHash": "<sha256>"},
  "coordinateContract": {"runtimeUp": "Y", "units": "meters", "referenceDiameter": 1},
  "files": {
    "native": {"path": "core.usdz", "sha256": "<sha256>"},
    "web": {"path": "core.glb", "sha256": "<sha256>"},
    "fallback": {"path": "core-static.png", "sha256": "<sha256>"}
  },
  "nodes": {"root": "SkyraCoreRoot", "core": "CoreMesh", "innerArc": "ArcInner", "outerArc": "ArcOuter"},
  "clips": [],
  "requires": ["transforms", "pbr"],
  "bounds": {"center": [0, 0, 0], "radius": 0.5},
  "limits": {"triangles": 20000, "materials": 4, "textureEdge": 2048},
  "license": {"identifier": "<approved license>", "attribution": "<if required>"}
}
```

Production schema: reject unknown major schema versions, unsafe relative paths, missing mandatory files, malformed hashes, nonfinite numbers, duplicate bindings and out-of-policy requirements. Limits above are maximum budgets; a separate generated inspection report stores actual counts. Enforce local policy even if a supplied manifest claims higher limits. Normalize exported axes into the stated runtime contract and verify it with the fixture.

Bundle immutable assets using Swift package resources plus the app launcher packaging path, after checking how `skyra app` copies resources. Verify launch outside the repository and in a clean user account. Web assets use same-origin versioned public paths, not a `.blend` file or arbitrary external URLs. No Blender executable ships in Skyra.app.

## 10. Fine-grained Blender integration for Skyra authoring

This initiative should expose more than “generate an image.” Proposed bounded operation surface:

| Operation | Parameters/result | Boundary |
|---|---|---|
| inspect_scene | job ID → objects, hierarchy, units, bounds, material/clip summaries | read allowed project |
| inspect_object | stable ID → transforms/topology summary/bindings | no scripts |
| set_transform | object ID, explicit translation/rotation/scale | staged scene only |
| set_material_parameter | material ID, allowed PBR parameter/value | typed and clamped |
| assign_texture | material slot, granted asset hash | no arbitrary path/network |
| set_visibility | object ID, export visibility | preserve source history |
| set_modifier_parameter | approved modifier/type/property | bounded geometry expansion |
| set_keyframe | object/bone ID, approved property/frame/value | bounded frame range |
| select_export_collection | stable collection ID | no unrestricted scene export |
| export_asset | preset ID, job output ID | new artifact directory |
| render_preview | approved camera preset, resolution, state | time/pixel limits |
| validate_asset | artifact ID → structured findings | independent validator |
| compare_preview | baseline/candidate IDs → metrics/images | report, no auto approval |

Each returns structured status, changed object IDs, before/after values, source version, artifact hashes, warnings and typed errors. Never collapse exporter, validation or SDK failures into “not connected.” Long jobs expose queued/running/succeeded/failed/cancelled and recorded timestamps.

Requests include `expectedSceneVersion` and idempotency key. Apply edits on an isolated working copy, validate, then save a new version. Conflicting scene version returns conflict, not last-writer-wins. Rollback selects a prior approved artifact; no destructive overwrite. Interactive Blender UI editing and headless jobs must not write the same source concurrently.

Implementation preference: a narrow request-to-Python dispatcher executed by a registered sandboxed Blender target. A local UI bridge is optional for previews, with authenticated local transport, explicit project scope and no generic eval operation. MCP is only a transport choice; it neither grants fine-grained semantics nor replaces Boundary enforcement. A live Blender connection is not available in this session and must not be represented as existing.

## 11. Security and authority

Respect Skyra's emitter/checker separation: exporter output is untrusted; a separate validator authorizes promotion based on observable artifact contents. Export tool cannot mark its own output approved.

Blender worker Boundary: read only granted source/assets and approved scripts; write only per-job output/cache; no network by default; no `.skyra/`, `.codex`, home-wide reads or arbitrary subprocesses. Verify executable and script identity through the runtime's existing release machinery, not self-reported hashes alone. Do not use an ambient shell to bypass registered tool recording.

Treat archives and parser inputs as attack surfaces. Reject absolute paths, traversal, external URL references, oversized/overcompressed packages, excessive geometry/textures and unsupported extension requirements. Use process memory/time quotas and cancellation; where native GPU access requires broader grants, scope and document that exception rather than claiming a perfect sandbox.

Runtime v1 accepts bundled approved assets only. User import is deferred and would require independent quarantine, validation and an explicit grant. Asset custom properties are inert metadata, never commands or hidden tool instructions. Manifests cannot widen authority. Logs omit conversation text, paths outside the grant and secrets; visual diagnostics use synthetic protocol fixtures unless the user requests a live capture.

## 12. Performance and lifecycle gates

Proposed targets to validate on a named baseline machine (initial candidate: 8-GB M1 MacBook Air, macOS 14) and a ProMotion device:

- No UI main-thread blocking during load/export; message input remains responsive.
- Foreground combined graphics p95 frame time ≤16.7 ms at 60 Hz. A 120-Hz tier requires ≤8.3 ms; otherwise cap appropriately rather than advertise 120 FPS.
- Incremental orb GPU time target ≤2 ms p95; measure with profiler, not CPU callbacks alone.
- Cold local asset ready target ≤500 ms; never gate conversation readiness on it.
- Incremental resident memory target ≤64 MiB; decoded textures separately tracked.
- Hidden window: no application-driven frame updates; verify actual engine/GPU behavior.
- Background: ≤10 visual updates/sec or suspend; no unbounded independent engine loop.
- Ten window open/close cycles: no accumulating subscriptions/GPU allocations; define tolerance after baseline measurement.

Quality ladder: standard asset → low LOD/reduced viewport resolution → static fallback → off. Trigger downgrade on sustained frame-budget misses (proposed 2 seconds); recover only after sustained headroom (10 seconds), with hysteresis. Thermal/low-power signals may force low/static. Viewport resolution uses backing scale with a quality cap, independent of layout points.

Test GPU composition, sRGB/tone mapping and premultiplied alpha against the current renderer. RealityKit and Metal layers do not inherently share depth or lighting. No promise of correct cross-layer 3D occlusion. Reject any solution that makes the sky opaque behind the orb or introduces halos over text.

## 13. Failure behavior and diagnostics

Typed errors: missing_resource, unsupported_schema, hash_mismatch, unsafe_reference, asset_budget_exceeded, exporter_failed, export_timeout, missing_node, missing_clip, unsupported_capability, engine_unavailable, sdk_incompatible, shader_failed, preview_failed, cancelled.

Render-load failure displays the existing sky plus native status if useful; no modal interruption. A missing optional arc/clip disables that effect. Missing root/core rejects the candidate. Failure to import does not imply successful validation. Failure to capture means fidelity remains unverified.

Export reports include stage, job ID, exit status, bounded sanitized stderr, version/config hashes and output inventory. Record engine/OS/render settings with preview captures. Rollback is available without rebuilding source assets. User graphics preferences survive application restarts; animation state does not persist as truth.

## 14. Web adapter (Phase 4)

Add a pinned Three.js dependency deliberately, not implicitly through the native implementation. Browser-only loading component with server-rendered static fallback; load GLB from approved same-origin paths. Map the same normalized controller output to object transforms and supported materials/animations. Validate the actual GLTFLoader API/version during implementation; the documentation index confirms the relevant loader and animation modules, not project compatibility. [D11]

DOM controls own input and accessible status. Decorative canvas uses pointer-events none; pointer observation from the UI feeds normalized values. ResizeObserver updates camera/renderer dimensions. Respect prefers-reduced-motion, visibility and context-loss handling. Dispose renderer, textures, geometries, materials, mixers and event listeners on unmount, accounting for shared resources. No continuously active render loop for a static hidden asset.

Do not require pixel-identical native/web rendering. Share semantics, palette, binding names, state fixtures and silhouette expectations. Each adapter has its own fidelity baselines and budgets.

## 15. Validation plan

### Pure tests

Projection: all phases; refusal versus fault; running count clamp; offline precedence; unknown call states; conversation switches; baseline snapshot without celebration; duplicate revisions; incarnation resets.

Controller: fixed-clock pointer easing and max angles; nil pointer reset; capped dt; Reduce Motion freezes animation; no backlog on resume; bounded cue rate; quality hysteresis.

Layout: minimum window, wide/narrow aspect ratios, sidebar open/closed, floating log positions, title/composer exclusions, backing scale 1×/2×, no free-area hide decision.

Manifest: invalid versions/hashes/paths, missing resources, unsupported features, nonfinite bounds, duplicate names, excess counts, archive traversal and decompression limits.

### Export/runtime fixture matrix

1. Asymmetric axis object: scale/up/front/handedness.
2. Named hierarchy: stable object bindings after actual load.
3. Neutral PBR patch: base/normal/roughness/metalness channels.
4. Alpha edge: compositing halos and backdrop opacity.
5. Two transform clips: output enumeration and playback, optional Phase 2.
6. Skin/morph fixture: separate feature gates, optional Phase 2.
7. Missing/corrupt resource: fallback, no UI crash.
8. Repeated load/unload: memory and cancellation.

Run USDZ on every supported macOS family and GLB on supported browser/GPU combinations when web ships. Do not promote based solely on Blender's viewport or an exported file existing.

### Visual fidelity

Capture app screenshots at 640×420, 1024×768 and 1440×900 points, defined backing scale, fixed time/seed and neutral/extreme pointers. Include empty, pending, returned, refused, failed, offline, Reduce Motion, floating log and graphics-off states. Compare to runtime-specific approved images with masked nondeterministic text/time regions.

Automated metrics are advisory initially: silhouette overlap target ≥0.98 under identical framing; no unexplained pixel differences in protected UI regions; optional image error threshold calibrated from repeat captures. Human review checks shading, silhouette, seam artifacts, readability and calmness. Do not demand Blender/Cycles and real-time PBR pixel equality. Store actual captures and reviewer result; a spec or mockup is not fidelity proof.

### Integration acceptance

Submit/type/select/copy/resize/arrange log/open sheets with graphics enabled; event delivery unchanged. Existing star/hash/offscreen tests pass. Reduce Motion and inactive/hidden tests pass. Launch packaged app away from repo. Benchmarks meet the named-machine gate. All failure modes fall back. No new runtime authority or ambient network access.

## 16. Implementation roadmap and exit criteria

Phase 0 — compatibility spike:
- Pin SDK/Blender versions and inspect resource packaging launcher.
- Export neutral core and axis fixture to USDZ.
- Compile a macOS-14-compatible RealityKit host experiment, validate camera/transparency/click-through/lifecycle/capture.
- Profile next to current sky.
Exit: documented go/no-go. If engine cannot meet deployment/lifecycle requirements, retain Metal/static fallback; do not begin full asset production.

Phase 1 — reactive core:
- Build pure projection/layout/controller; ship one transform-only asset.
- Hook existing pointer and observed phases; add graphics settings/fallback.
- Bundle verified resources, implement load cancellation and typed errors.
Exit: acceptance matrix, package smoke test and performance report pass.

Phase 2 — production assets:
- Add controlled material highlights, optional clips, low LOD and automated exporter/validator.
- Establish reference captures and artifact provenance.
Exit: every enabled feature demonstrated in actual runtime, rollback tested.

Phase 3 — fine-grained authoring:
- Register bounded Blender target, typed edit operations, job lifecycle and isolated source versions.
- Add inspect → edit → export → capture → compare loop.
Exit: end-to-end small property edit with before/after evidence, sandbox-negative tests, concurrency/conflict tests and no generic eval.

Phase 4 — web parity:
- Add GLB adapter, static-first browser loading, lifecycle/disposal tests.
Exit: shared semantic fixtures pass and browser-specific visual/performance baselines approved.

Roles: desktop engineer owns runtime/packaging; technical artist owns Blender master and material/axis fixtures; runtime/security engineer owns worker boundary and validator; reviewer owns asset promotion. One person can fill several roles, but approval remains separate from export logic.

## 17. Risks and unresolved decisions

| Risk | Mitigation / decision gate |
|---|---|
| macOS 14 incompatible convenience APIs | compile on minimum target; no silent target bump |
| Separate engine renders continuously | measure hidden/inactive energy; suspend or reject adapter |
| PBR/alpha diverges from sky | neutral materials, local viewport, runtime references |
|USDZ rig/morph/clip loss | transforms first, independent feature fixtures |
| Procedural Blender look does not export | bake or implement explicit runtime shader |
| Decorative state implies false task success | return cue means returned call only; textual truth remains |
| Source scripts execute unexpectedly | disable autoexec plus isolated trusted worker and OS limits |
| Export nondeterminism | record toolchain/hashes; compare semantics, not promised byte reproducibility |
| Resources missing in custom app bundle | launcher/package smoke test outside repository |
| Fine-grained tool mutates wrong object | stable IDs, expected version, before/after report, isolated copy |
| UI distraction or motion sensitivity | small effect amplitudes, reading exclusions, off/static modes |

Open before implementation: installed Blender release/exporter APIs; exact minimum-target RealityKit host/camera support; actual pause/capture behavior; whether transparent composition is economical; resource copy rules in `skyra app`; baseline hardware and memory tolerances; licenses of any third-party texture/model; whether interactive core is desired at all.

## 18. Definition of done

This initiative is complete only when a reviewed Blender-authored asset ships in a packaged Skyra app, responds to documented observations, preserves input/accessibility and existing sky tests, meets measured budgets, has secure reproducible-in-process export provenance and rollback, and has actual preview evidence. Optional capabilities are disabled until demonstrated. Documentation, an SVG, or a successful bridge connection alone do not satisfy completion.
