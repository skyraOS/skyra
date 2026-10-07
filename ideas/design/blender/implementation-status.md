# Blender adapter and native graphics pilot

Implemented against `ideas/skyra-responsive-graphics-blender-spec.md`.
This is a working developer pilot, not completion of every phase in that spec.

## What works

- `blender` is a Skyra procedure, offered to Think and represented in its constellation.
  Its private `blender.tools` leaf is available only within the Blender procedure.
- Typed scene/object inspection, transform changes, material changes, USDZ/GLB
  export, and PNG rendering run against an isolated synthetic working copy.
- Version checks reject stale edits before dispatch. Edits create new files;
  exports are candidates marked unreviewed. No generic Python tool is exposed.
- The desktop backend starts the bridge automatically. Blender 5.1.2 was installed
  at `~/.skyra/tools/Blender-5.1.2.app`, and its application signature verified.
  This adapter uses a fixed headless worker, not a third-party Blender MCP server.
- A bundled sphere and two arcs load in RealityKit over the existing Metal sky.
  Graphics → Core (pilot) offers Off, Static, and Reactive. Default is Off.
  Reactive adds bounded pointer tilt, pending arc movement, and observed call cues.
  Return cues describe a call return, not task success.
- Reduced motion and inactive presentation use the static PNG. Hidden/off modes
  remove the live view. Placement avoids controls, transcripts, and constellations;
  if no safe viewport fits, it hides. Decorative views do not accept clicks.
- Native/fallback resources are packaged in the app and hash checked. Authoring
  outputs cannot replace bundled resources automatically.

The invisible app core was caused by an initially empty SwiftUI Group: its asset
loading task did not mount. Keeping a concrete view mounted fixed both native
rendering and the static fallback. Actual app screenshots accompany this report.
The desktop launcher also now waits for its TCP listener before reporting ready;
a stale parking socket alone no longer passes the startup check.

## Evidence

- `live-bridge-context.json`: context from the running app bridge, recorded and replayed.
- `adapter-acceptance.json`: five recorded calls covering context, inspection,
  a material edit, export, and rendering. Each replayed byte-identically.
- `artifacts/job-am00no0v`: exported candidate; independent inspection found
  2,368 triangles, two materials, three mesh nodes, and expected node bindings.
- `artifacts/job-ngeqge03/preview.png`: Blender-rendered 384 × 384 PNG.
- `core-report.json` and `axis-report.json`: standalone native loading, framing,
  transparent snapshots, click-through, and weak view release after teardown.
- `skyra-static-pilot.png` and `skyra-reactive-pilot.png`: actual app captures.
- Python validation tests, Go adapter/provider/host tests, and the desktop test
  suite pass. One existing desktop test requiring a live agent is skipped.

Native checks ran on this MacBook Pro M1 Pro with macOS 15.6. This does not
satisfy the spec's named macOS 14 and ProMotion performance gates.

## Remaining production work

The broader spec still calls for performance/energy/GPU measurements, a quality
ladder, production asset review and promotion tooling, stronger service isolation,
async job progress/cancellation and explicit idempotency, texture/modifier/animation
operations, and optional web rendering. Arbitrary project selection is not exposed.
USDZ inspection currently checks its archive; scene semantics need native loading
and visual review. The worker's network/filesystem restrictions are enforced, but
its Mach/IOKit/shared-memory grants remain broad and need a production audit.

No changes have been committed as part of this implementation.
