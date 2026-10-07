# Native graphics compatibility pilot

Standalone macOS 14-targeted RealityKit experiment. It never starts Skyra's backend or exposes tools. Generates real PNG snapshots, verifies named hierarchy, records load timing and tests click-through/view detachment. Reports do not establish GPU/energy/memory budgets or compatibility on an actual macOS 14 device.

```sh
DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer swift build --package-path desktop/GraphicsPilot
desktop/GraphicsPilot/.build/debug/GraphicsPilot /tmp/skyra-core-preview core
desktop/GraphicsPilot/.build/debug/GraphicsPilot /tmp/skyra-axis-preview axis
```

Assets are fixtures from `controller/blender/build_pilot.py`, Blender 5.1.2. The core is a neutral PBR sphere with two rings; the asymmetric RGB axis fixture verifies Blender Z-up to runtime Y-up conversion. Source `.blend`, GLB and provenance report are in `controller/blender/fixtures`. The production app consumes a separately hashed bundle of the pilot core and static preview.

Use the public post-processing callback only if its source texture is actually copied to the target: a no-op callback produces a blank frame. This experiment preserves pixels while counting frames; the app adapter uses scene update notifications and no post-processing override.
