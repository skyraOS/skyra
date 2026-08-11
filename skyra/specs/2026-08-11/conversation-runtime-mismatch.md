# Conversation / Runtime Mismatch

**Date:** 2026-08-11
**Status: NOTE.** Records mismatches only. It does not choose between the conversation and the
current `skyra-dna-v5` runtime.

| current runtime | conversation |
|---|---|
| Failed children produce stored error observations. | Errors remain ephemeral. |
| Every invocation enters child history. | Only successful replacement structure matters. |
| Discovery passes rejected outputs back to avoid retrying them. | The node directly authors a better expectation. |
| Discovered engrams are reused by pointer and ID. | Reuse copies the expectation into a new locally owned engram. |
| An internal root commit disperses credit immediately. | The user is the ultimate acceptance boundary. |
