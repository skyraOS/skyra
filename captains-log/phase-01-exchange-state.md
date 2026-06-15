# Phase: Exchange State — First Contact

## Setup

v.1 branch. Skyra + Michael. DeepSeek as provider. TUI input. Three express functions registered: compress, speak (internal), motor. Four maps on the Slice: Thread, TrainOfThought, Understandings, PreviousActions. User Express fires before Observe. World routing removed from Skyra's Observe. Speak is the only primitive.

Observe prompt: systemPrompt (propagate)
Express prompt: compressPrompt (compress), speakInternalPrompt (speak), speakActPrompt (motor)

## Run

Michael typed "skyra hi how are u" and "skyra hi are you there?" — no response displayed in the TUI.

## Observations

### Motor never fired

Expected:
  After propagate and speak internal complete, motor fires with Thread state + speak intent content, calls LLM with speakActPrompt, produces spoken words, writes to Relation via SetImpulse.

Observed:
  Motor checked `speakContent` (extracted from `<speak>` tag in intent) and `threadParts` (from Slice.Thread). Both were empty. Motor returned early without calling the LLM.

Evidence:
  No MOTOR call in provider.log. Traverse log shows EXIT without any motor output. Michael's observe log shows `origin: michael` — the Relation still carried Michael's impulse because motor never wrote skyra's response.

Conclusion:
  Motor had no input. Two causes:

  1. **Intent had no `<speak>` tag.** The LLM emitted `<then>` instead of `<speak>` because the observe prompt referenced `<then>` in the intent format. Fixed: removed `<then>` from both observe and compress prompts, only `<speak>` remains.

  2. **Skyra's Slice.Thread was empty.** Michael's impulse landed on the Relation via `r.SetImpulse("michael", msg)` but nothing wrote it to Skyra's Thread map. The Thread binding reality exists in Skyra's Relationships but it only manages member weights — it doesn't capture exchange content from the Relation onto the Slice.Thread map.

### Express prompt chain unclear

Expected:
  Three express functions fire in sequence: compress → speak internal → motor. Each uses its own prompt. The compressPrompt fires only when competence selects "compress" (high thought density). The speakInternalPrompt always fires after. The speakActPrompt fires in motor at the exit.

Observed:
  CompetenceExpress selected "pass-through" (thought density was low — only one cell on TrainOfThought). So compress did not fire. Speak internal fired (speakInternalPrompt) and produced output that landed on Understandings. Motor fired but bailed because it reads Thread + intent speak tag, not Understandings.

Evidence:
  Provider log shows PROPAGATE call → SPEAK call → no MOTOR call. Express-competence in traverse log: "pass-through". Speak internal result was: "We're maintaining system coherence as a high-priority intent..."

Conclusion:
  **Motor is disconnected from speak internal's output.** Speak internal writes to Understandings. Motor reads from Thread + intent `<speak>` tag. These are two different places. The speak internal output never reaches motor.

  The chain needs to be: speak internal produces context → that context flows to motor. Currently the chain is broken because they read/write to different maps.

### Visited map not resetting

Expected:
  Each exchange (impulse from Michael) triggers a fresh traversal through Skyra's topology.

Observed:
  First message triggered a full traversal. Second message hit `sl.Visited[s.id]` still true from the first traversal. Skyra's Realize returned immediately with "SKIP (already visited)".

Evidence:
  Traverse log shows only one ENTER/EXIT cycle despite two messages from Michael.

Conclusion:
  The visited map on the Slice persists across exchanges. It needs to reset at the start of each new impulse so the being can traverse its topology again. Without this, the being only thinks once and then goes silent forever.

## Phase Conclusion

Three issues blocking the exchange round-trip:

1. **Motor reads the wrong source.** Motor reads Thread + intent `<speak>`. Speak internal writes to Understandings. The chain is broken. Motor needs access to what speak internal produced.

2. **Thread map not populated.** Michael's impulse lands on the Relation but never reaches Skyra's Slice.Thread. The Thread binding reality doesn't capture exchange content — it only manages member weights.

3. **Visited map not resetting.** The being can only traverse once. Second impulse gets blocked by visit-once guard from the first traversal.

Next phase: fix these three, get a clean round-trip.
