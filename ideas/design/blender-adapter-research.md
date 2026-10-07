# Blender adapter research for Skyra

Researched October 7, 2026. Research and proposed design only; nothing installed or connected.

## Recommendation

Investigate the **official Blender Lab MCP server** first. It gives Skyra a direct route to Blender scene data and Python automation with bundled API/manual lookup and PNG screenshots. Integrate it as a recorded Skyra capability, following the Penpot bridge pattern, rather than configuring a global Codex MCP connection.

Primary sources:

- [Official Blender Lab MCP page](https://www.blender.org/lab/mcp-server/).
- [Official source repository](https://projects.blender.org/lab/blender_mcp).
- [Inspected revision](https://projects.blender.org/lab/blender_mcp/src/commit/dbbf836ad4b1025f14a2b3b504c43903f39e0b04).
- [Community MCP for Blender](https://github.com/ahujasid/mcp-for-blender).
- [blender-remote](https://github.com/igamenovoer/blender-remote).
- [Blender Python module documentation](https://docs.blender.org/api/main/info_advanced_blender_as_bpy.html).

The official web repository blocked browser-tool access, but a read-only source checkout succeeded. The findings below about tools, transport, requirements and weak sandbox come from that checkout, not third-party setup guides. No Blender executable was found at /Applications/Blender.app/Contents/MacOS/Blender or on this shell's PATH; other installations were not exhaustively searched.

## Candidates

| Candidate | Fit | Tradeoff |
| --- | --- | --- |
| Official Blender Lab MCP | Preferred for a small integration using Blender's own API and documentation | Add-on requires Blender 5.1+; unrestricted Python is powerful and needs a deliberate execution boundary |
| Community MCP for Blender, previously ahujasid/blender-mcp | Alternative for asset libraries, generated models and richer visual workflows | More integrations and configuration; advertised safe mode still permits normal Blender saving/import/export and is not read-only |
| blender-remote | Alternative if a Python automation client is more useful than the official connector | Additional community dependency; requires its own compatibility assessment |
| Direct Blender background process / bpy | Useful for repeatable asset generation without a live desktop session | Does not automatically provide a live interactive Blender connection; Skyra must implement job management and outputs |

These alternatives were reviewed at documentation/source level, not installed or benchmarked. The recommendation is an architectural judgment, not a measured reliability comparison.

## Official adapter facts verified in source

- Two parts: add-on inside Blender and separate Python MCP server.
- Default MCP transport is stdio; source also supports Streamable HTTP via --transport http.
- Add-on socket defaults to localhost:9876. MCP-to-add-on variables are BLENDER_MCP_HOST and BLENDER_MCP_PORT, different from the community project's variables.
- Python package requires Python >=3.10 and mcp[cli]>=1.2.0,<2. Add-on manifest requires Blender >=5.1.0.
- Source lists 26 tools: live/background Python execution, file/data/object summaries, missing-reference checks, linked libraries, bundled API/manual searches, workspace/view navigation, PNG area/window screenshots, thumbnail and viewport rendering.
- render_thumbnail_to_path and render_viewport_to_path accept filesystem output paths. The Skyra bridge must assign those paths rather than accepting arbitrary model-supplied locations.
- The live socket client has a 300-second timeout. Render lifecycle should not inherit Penpot's short synchronous timeout unchanged.
- The inspected package version is 1.0.2 and add-on manifest version is 1.0.0. Pin the actual tested revision and paired artifacts rather than assuming matching version strings.
- Source uses GPL-3.0-or-later notices. Keep dependency provenance and distribution review separate from adapter implementation.

## Proposed Skyra composition

think → blender procedure → narrowed Codex → blender_tools → private Skyra bridge → official MCP server → Blender add-on

Runtime capability: blender.tools. Provider name: blender_tools. The think constellation shows a blender procedure star; its internal dependency stays beneath that star. A dedicated Blender root can use the same procedure.

Local briefs and reports: ideas/design/blender/. Generated .blend files, previews and exports: a bridge-owned project directory under that location. Exported binary assets need their own output handling; the existing text writer cannot create them.

Proposed public operations:

| Operation | Purpose |
| --- | --- |
| context | Report connected Blender version, file, scene, dirty state and bridge incarnation |
| inspect | Allowlisted scene/object/file summaries |
| docs | API/manual lookup through upstream documentation tools |
| preview | PNG area/window screenshot returned through Skyra's existing image validation |
| execute | Explicit editing mode; Python runs inside the bound Blender session |
| render | Bridge chooses output path and returns a validated preview/job result |
| save_copy | Save into bridge-owned project space, preserving the original project |

Do not expose an unrestricted upstream tool-name/arguments passthrough. Pin the executable and upstream catalog; serialize edits for one Blender instance. Context must bind the process/session and project; filepath alone is insufficient for unnamed files or reopening the same path. Check identity immediately before edits, with an execution-time guard where possible.

## Read-only and write boundaries

A read-only first milestone is feasible: offer only reviewed summary, documentation and screenshot tools; exclude execute, render-to-path, save, import/export and navigation. Upstream readOnlyHint annotations are descriptions, not enforcement. Inspect the implementation of each selected tool and validate arguments in Skyra.

Once execute is offered, it is not read-only. Python inside Blender can change/delete scene objects and access the filesystem. Skyra's Codex read-only sandbox does not constrain a separately running Blender process. The official weak_sandbox.py explicitly describes itself as guidance rather than isolation and blocks only a few disruptive calls. Prompt rules or Python filters cannot turn unrestricted Blender Python into a trustworthy filesystem boundary.

For bounded editing, launch a dedicated Blender instance on a copied project with OS-level filesystem/network restrictions, or implement specific trusted operations instead of arbitrary Python. Keep the add-on listener local; localhost limits network exposure but is not authentication. A copy protects the original .blend file only, not the rest of the filesystem.

## Implementation and acceptance sequence

1. Locate/install a compatible Blender and pin the official MCP/add-on revision. Start one dedicated instance and confirm actual tools/list plus live context. No account or cloud key is needed for the local official transport.
2. Build the operator-owned bridge and recorded blender.tools leaf. Use stdio to the MCP server and a private socket grant to Skyra actors. Keep endpoints/process launch settings outside model input.
3. Ship the read-only tool set and specific errors: binary missing, unsupported version, add-on unavailable, wrong project, invalid response, script error and timeout. Preserve upstream details. Never replace all failures with a reconnect prompt.
4. Add bounded editing and project/output management after the Blender execution boundary is established. Mark outcomes unknown after dispatched edit timeouts; inspect before retrying, never automatically rerun edits.
5. Add procedure/root/constellation wiring and validated PNG display. Keep rendering asynchronous or use recorded polling with visible status.
6. Live acceptance: inspect a fixture, create an object in a copied project, apply material/light/camera, export a PNG, inspect it visually, save a .blend copy and reopen it to verify persistence. Test wrong-project rejection and uncertain-edit recovery. Confirm offline record replay does not dispatch Blender changes again.

Not yet proven: installed-version compatibility, handshake, screenshot bounds, rendering throughput, save/reopen behavior, or a functioning Blender sandbox on this Mac.
