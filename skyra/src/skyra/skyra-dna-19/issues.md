# DNA-19 remaining limits

- Live handshake, inference, tool calls/resumes, fork divergence, tool-set replacement
  with retained history, and concurrent-parent marker isolation passed with CLI
  0.155.1 and the existing ChatGPT login. The audit records nine distinct threads.
  This is evidence for the tested scenarios, not a proof that the provider universally
  seals arbitrary histories or behaves identically across models/versions. Live
  native-effect contamination was not deliberately induced; deterministic tests
  cover that fence and adversarial correlations.
- Provider thread sealing and the selected binary are trusted. HMAC references
  authenticate host provisioning across the Move boundary; they are not protection
  against a compromised binary, a same-user debugger or malicious OS. No sockets
  or sandbox implementation is added.
- Requested thread policy is checked against the provider-reported approval/sandbox
  response. Dynamic tool execution is limited to echo by both descriptor and actor.
  Full provider configuration, plugin/native-tool availability and actual enforcement
  still require deployment-specific auditing. The effect fence rejects results after
  detection; it cannot establish that an effect was prevented or undone.
- Migration retains input/final-response/tool-result text as inert historical data,
  bounded to 64 KiB. It does not preserve hidden reasoning, provider-native items or
  guarantee identical model behavior. No effects are replayed. Oversize history and
  unsupported replacement return explicit faults. No silent truncation/fallback.
- One command and one pending dependency per parent. A fork locks its source until
  the fork response. Idle-only explicit parent termination; cancellation retires all
  local work. Descendant invalidation rejects results, but cannot undo child effects
  already committed. There is no general concurrent child join or provider interrupt.
- Actor/pipe loss remains runtime-fatal. Provider failure on a healthy actor port
  produces correlated faults. No retries, crash persistence, cross-process restart
  resume or general connection reconnection. Provider request deadlines are five
  seconds (health is caller-bounded); a command deadline is thirty seconds.
- Graceful cancellation closes streams, lets the actor reap its provider and reaps
  the actor. A two-second process-group kill is the fallback for an unresponsive
  actor; after a forced actor death, descendant reaping belongs to the OS, not the
  dead actor. This is not a production process supervisor. Process-group plumbing
  currently targets Unix (the implementation environment is macOS).
- Protocol input and catalog/frame/history retention are not production resource
  quotas. Old command/parent/thread tombstones prevent reuse and remain until process
  shutdown. Streams use standard-library JSON framing, not hardened size-bounded IPC.
- Inherited bounded-runtime differences remain: direct frame commits instead of an
  Engram compiler, diagnostic rather than general typed recovery, explicit plumbing
  rather than an OS service, one pending child, and one fixed creation template.
