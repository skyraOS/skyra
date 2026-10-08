# Niche opportunity: maintain one The One Ring campaign in Foundry VTT

Research lead, not validated demand. Recorded October 8, 2026.

## Small promise
“Make my existing campaign ready for our next session, without losing what we’ve built.”

Not an AI dungeon master or general Foundry support: technical upkeep for one GM’s ongoing campaign.

## Pain hypothesis
A long-lived campaign accumulates modules, custom content, adventures, and version dependencies. Occasional updates or new requirements can interrupt play and force the GM to reconstruct compatibility knowledge and repair their setup.

The prior conversation identified a publisher-forum report of a user unable to access their world or continue their campaign after an update. This is a research lead, not evidence of widespread demand; reverify the thread before relying on it:
https://forum.frialigan.se/viewtopic.php?t=15815

## Proposed workflow
1. Capture versions, modules, custom content, and GM-defined essential functionality.
2. Record a working baseline and operate on a backup, never silently on the live campaign.
3. Accept a new requirement, such as an adventure installation or additional module.
4. Test campaign-specific checks: player connection, character access, essential rolls, and upcoming-session scenes.
5. Repair failed branches; ask the GM before replacing functionality or changing campaign data.
6. Remember the successfully executed checks and repair procedure, with explicit observations and contracts.

These are proposed capabilities, not established Skyra functionality. Durable task state is required in addition to executable procedures.

## Why Skyra might fit
The ongoing task is preserving campaign readiness across sessions and changes. Human decisions and AI execution compose in the same workflow. Remembered procedures retain the particular campaign’s acceptance criteria without relying on conversation context. Adaptation handles changed assumptions rather than repeating the entire investigation.

## Existing alternatives / boundary
Foundry already offers backups, snapshots, and compatibility warnings. Avoid building another backup manager. The opportunity hypothesis is campaign-specific verification and recovery beyond version metadata.
Documentation lead from the conversation:
https://foundryvtt.com/article/backups/

## Smallest experiment
One GM, one The One Ring campaign, one new adventure installation, three campaign-specific checks. Work on a copy; require owner approval before applying changes.

Measure preparation time saved, human interventions, repeatability after restart, and whether remembered checks remain useful on a subsequent change.

## Open questions
- Do these GMs repeatedly lose meaningful preparation time to maintenance?
- Which failures recur, and which can be reliably observed and tested?
- Would a GM trust a tool with a copy of the campaign?
- What existing tools already cover campaign-specific verification?
- Does remembering the procedure reduce effort on the second change?

Next step: interview a few The One Ring GMs with existing Foundry campaigns and collect concrete maintenance incidents before expanding scope.
