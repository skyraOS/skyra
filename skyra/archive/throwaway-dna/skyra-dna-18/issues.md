# Open boundaries

This experiment does not decide final main-loop ownership, activation envelopes,
save/resume persistence, preparation/trap execution, output contract taxonomy,
provider or OS containment, arbitrary multi-child joins, fairness/backpressure,
crash recovery, reliable delivery, or observation signing.

The root and shared sketches differ about enqueue ownership and expose more state
than the work order permits. The chosen bounded declarations are in DECISIONS.md.
Production compilation/memoization of Engrams is not exercised; this fixture commits
frame transitions directly. Binding replacement abandons old pending frames; recovery
of those activations requires a future cancellation/fault contract.

The exposed context executable is a detached declaration; this experiment's trusted
external endpoint creates root frames directly against authorized local targets.
Routing to another context by its executable descriptor, delegation, and a full
activation envelope are not demonstrated. Frame identity on the fixture wire is a
local integer plus the Move's context field; it reduces the sketch's compound
FrameID. Admission checks context identity and endpoint provenance before lookup.

Fault strings and dispatch errors are fixture diagnostics. The parent peer exercises
successful child returns; fault recovery, retries, and guaranteed delivery remain
open. There is no transport retry after a committed send fails. Frames and request
IDs remain retained for this short-lived experiment; bounded retention and production
resource limits are unimplemented. The two scheduler workers demonstrate independent
context progress, not unlimited progress when every worker blocks on transport.

No unresolved blocker remains for the work order's bounded acceptance scenarios.
Passing tests do not establish the production properties listed above.
