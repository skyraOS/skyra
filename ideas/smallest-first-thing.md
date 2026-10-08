# Smallest first thing: two niche maintenance experiments

Recorded October 8, 2026. Opportunity hypotheses, not validated demand or established Skyra capabilities. Consolidates the Foundry campaign idea and the ESPHome panel idea from our conversation.

## Goal
Prove that a small, long-lived task can be done, remembered as an executable procedure, and reused through changing requirements and occasional repairs. Human decisions and AI steps compose together; observations pass explicitly under procedure contracts. Code avoids dependence on retaining the conversation, but durable task state is still required.

## Leading candidate: one homemade ESPHome touchscreen panel

### Tiny promise
“Keep the panel I built working as my household changes.”

Scope: one panel controlling non-safety-critical lights and displaying household information—not the whole smart home. Exclude locks, heating, alarms, and safety-sensitive controls.

### Changes and healing
- Replace the kitchen lights and update the corresponding control.
- Add a garage-door status display (read-only).
- Make buttons larger for a household member.
- Repair configuration or dependency breakage after an update.

### Proposed remembered procedure
1. Capture the existing YAML, device/build versions, known-working baseline, and owner-defined checks.
2. Inspect relevant Home Assistant entities and gather actual observations.
3. Edit configuration, validate, and compile.
4. Obtain owner approval before deployment.
5. Upload and collect logs and entity observations.
6. Ask the owner to confirm physical appearance and behavior. A successful build or online device does not prove that the kitchen button controls the kitchen.
7. Repair failed checks under the contract, preserving recovery information and approval boundaries.
8. Remember the successfully executed procedure and reuse it for the next change.

### Available integration leads
The earlier conversation identified these interfaces; verify current details before implementation:
- ESPHome CLI: YAML validation, compilation, upload, logs.
- ESPHome native API: exposed device entities via its specialized protocol.
- Home Assistant REST API: states and service calls.
- Home Assistant WebSocket API: events and state changes.

### Existing overlap
These are documented-capability leads from prior research, not independently tested reliability:
- ESPHome Designer: visual configuration/display editing and AI layout assistance; touchscreen/LVGL support described as experimental.
- ESPboards LVGL Designer: visual layouts, YAML import/export, hardware presets.
- esphome-mcp: configuration read/write, validation, compilation, deployment, logs, and live entity inspection.

Use the existing MCP integration if suitable rather than building another ESPHome manager. “AI edits and flashes firmware” is not the differentiator. The hypothesis is remembering what working means for this household and preserving it through changes, including explicit human acceptance checks.

Research leads:
- https://github.com/jeeftor/esphome-mcp
- https://github.com/koosoli/ESPHomeDesigner
- https://lvgl.espboards.dev/
- https://esphome.io/components/api/
- https://developers.home-assistant.io/docs/api/rest/
- https://developers.home-assistant.io/docs/api/websocket/

### Smallest experiment
One panel, one requested button change, one interrupted execution/resume, and three owner-confirmed checks. Then reuse the remembered procedure for a second and third change.

Measure elapsed maintenance effort, human interventions, repair effort, recovery after restart, and whether repeat execution is genuinely easier. Validate whether owners experience recurring pain and trust this workflow.

## Alternate candidate: one The One Ring campaign in Foundry VTT

### Tiny promise
“Make my existing campaign ready for our next session, without losing what we’ve built.”

Not an AI dungeon master or general Foundry support: technical upkeep for one GM’s ongoing campaign.

### Pain hypothesis
A long-lived campaign accumulates modules, custom content, adventures, and version dependencies. Updates or new requirements can interrupt play and force the GM to reconstruct compatibility knowledge.

The prior conversation identified a publisher-forum report of a user unable to access their world or continue their campaign after an update. Reverify before relying on it; an individual report does not establish widespread demand.
https://forum.frialigan.se/viewtopic.php?t=15815

### Proposed procedure
1. Capture versions, modules, custom content, and GM-defined essential functionality.
2. Record a working baseline and operate on a backup, not silently on the live campaign.
3. Accept a requirement such as a new adventure or module.
4. Test player connection, character access, essential rolls, and upcoming-session scenes.
5. Repair failed branches, asking the GM before replacing functionality or changing campaign data.
6. Remember the executed checks and repair procedure with explicit observations and contracts.

Foundry already provides backups, snapshots, and compatibility warnings. Do not build another backup manager; test whether campaign-specific verification and recovery add value beyond version metadata.
https://foundryvtt.com/article/backups/

### Smallest experiment
One GM, one campaign, one new adventure installation, three campaign-specific checks. Work on a copy and require approval before applying changes. Measure preparation time saved, interventions, repeatability after restart, and usefulness on a subsequent change.

Interview a few The One Ring GMs about concrete maintenance incidents, recurring failures, existing tools, and trust before expanding scope.

## Decision criterion
Prefer whichever yields a real participant and a repeatable maintenance incident first. The ESPHome panel is the leading candidate because it tests composition across software execution and human physical verification. Neither candidate establishes demand yet.
