# Blender authoring adapter

Skyra exposes a `blender` procedure in Think and a narrowed `blender.tools`
leaf inside that procedure. This is a typed, offline authoring adapter;
no general Python execution or third-party MCP server is exposed to the model.

Start the desktop app with `./skyra app`. Its `serve` backend starts a private
Unix-socket bridge. Requires macOS and Blender 5.1+. Set `SKYRA_BLENDER_BINARY`
before launching to select a binary; otherwise the bridge checks
`~/.skyra/tools/Blender-5.1.2.app` and `/Applications/Blender.app`.
Other launcher modes currently do not start the bridge automatically.

## Operations

Call `context` first. Every other operation requires its returned
`scene_version` as `expected_scene_version`:

- `inspect_scene`: named objects, transforms, materials, and triangle counts.
- `inspect_object`: one object by `object_id`.
- `set_transform`: translation in meters, rotation in radians, or positive
  scale, each as three finite numbers bounded to [-10, 10].
- `set_material_parameter`: `roughness`, `metallic`, or RGBA `base_color` in [0, 1].
- `export_asset`: fresh USDZ and GLB candidates from `SKYRA_EXPORT`.
- `render_preview`: a bounded transparent PNG, supplied as image content to Codex.

The initial working copy is the synthetic fixture in `fixtures/core.blend`.
Edits save new copies in `~/.skyra/blender/jobs`; a successful edit advances
an atomic working-copy receipt. The original fixture is never overwritten.
Outputs are copied to new directories under `ideas/design/blender/artifacts`.
The model cannot select project paths, output paths, scripts, or executable code.

## Execution and errors

A pinned, operator-owned script runs headless under macOS Seatbelt, with
scripts in the blend file disabled. The worker can read the fixed source,
Blender/system libraries, and its job directory, and write only to its job
and `/dev/null`. It has no network grant. Mach service, IOKit, and shared-memory
permissions remain broad in this developer pilot; this is not a completed
production isolation audit.

Jobs are serialized and limited to 90 seconds. Scene-version conflicts fail
before starting Blender. Failures include stage and specific codes such as
`binary_missing`, `scene_version_conflict`, `object_not_found`,
`material_not_found`, `job_timeout`, and `asset_validation_failed`.
A lost bridge response after dispatch reports an unknown edit outcome;
inspect context before retrying. There is no asynchronous cancellation API yet.

Exports are independently inspected for package limits, external references,
geometry bounds, and basic archive structure. All exports remain **unreviewed**.
USD scene semantics still need native loading and visual review. The app loads
only its bundled, hash-checked pilot asset; authoring never promotes a candidate.

## Verification

From `controller`, run `go test ./blender ./codex/resident ./hosts/web ./targets/user`.
Run `python3 -m unittest discover -s controller/blender -p 'test_*.py'` from the
repo root. The native tests run with `swift test --package-path desktop` using
full Xcode. The standalone native load/lifecycle spike is in
`desktop/GraphicsPilot`.
