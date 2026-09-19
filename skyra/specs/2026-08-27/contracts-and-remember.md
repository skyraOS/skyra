# Contracts are written at Remember — types flow bottom-up

> **TYPE UPDATE (2026-09-06):** [Contract = Action + Schema](../2026-09-01/body-registers-and-return.md#contract--action--schema)
> replaces the name-only design with Action (the sole model-visible header, capped at 160 tokens) and
> opaque JSON Schema. The discovery/REMEMBER lifecycle still applies; code and historical name-only
> examples below have not been migrated.

> **REFINEMENT (2026-09-06):** [Body, registers, and return](../2026-09-01/body-registers-and-return.md)
> frames Body as the register requirements for producing a return. Declaration can name observations
> and use already-known contract information while unseen structural details remain to be discovered.
> The discovery/REMEMBER lifecycle below still applies; concrete representation needs further work.

**Date:** 2026-08-27
**Status: DECISION / CORRECTION.** The caller does NOT declare a structural return contract at CALL — it
knows only the NAME of the child it invokes (and its own inputs). Types are NEGOTIATED (ACT) and DISCOVERED
by running; the return contract is WRITTEN at REMEMBER, inferred bottom-up from what the run composed.
Corrects the "CALL writes Register.Contract" prose in system-and-semantic.md. Reconciles with the JIT:
Remember is the type-feedback store.

---

## The correction

```text
AT CALL the caller knows only:  the NAME of the invoked child ("I expect what X returns") + its OWN inputs
                                (what was passed to it). NOT the child's return shape — it has never seen it.
TYPES are NEGOTIATED (ACT), not declared. params and return types are settled at ACT, not by the caller.
A RETURN is a STRUCTURED Observation — each field's value is a type, and each field must be an EXISTING
Observation. a return type is a COMPOSITION of existing observations.
THE PARENT learns a child's return ONLY when it returns; it composes children's observations into its OWN
return; the parent's return CONTRACT is written at REMEMBER (crystallized on persist).
```

Types flow UP from execution and are recorded on the way OUT — not DOWN from a declaration on the way in.

## The contract lifecycle — three moments, not one

```text
CALL      (SemanticCall)   WRITE THE NAME — "I invoked child X." a reference, not a shape. Contract == a Name.
RETURN    (SemanticReturn) COLD: DISCOVER — record what came back (each field an existing observation). no
                           judgment; nothing to judge against yet.
                           WARM: GUARD — check the return against the REMEMBERED contract (prior run). mismatch
                           → deopt.
REMEMBER  (persist)        WRITE THE CONTRACT — the run's return type, INFERRED bottom-up from what it composed
                           (children's observations). the structural type crystallizes and is persisted HERE.
```

The contract is an OUTPUT of the first run, not an INPUT to it.

## The JIT, completed — Remember is the type-feedback store

```text
COLD run     types unknown → DISCOVER by running → WRITE the contract at Remember.  (the profiling run)
ACT compiles against the REMEMBERED contract as its type assumption; inserts GUARDS.
WARM run     the guard at SemanticReturn checks the return vs the REMEMBERED contract; deopt re-Remembers.
```

You cannot statically type the base case (an LLM), so you observe · record · compile · guard — textbook
adaptive JIT. Remember was already the persistence point; it is ALSO the type-inference commit point.

## What this fixes / what survives (from system-and-semantic.md)

```text
SURVIVES  the type-error flag is still raised at SemanticReturn — value meets criterion there.
FIXED     the CRITERION (Register.Contract) is NOT written at SemanticCall. at CALL you write only the NAME;
          the structural contract is WRITTEN AT REMEMBER; the guard checks against that REMEMBERED contract.
          on the COLD run there is no criterion yet — SemanticReturn DISCOVERS, it does not judge.
```

### The post-wake fork, now sharp

```text
guard fails, return is a valid composition of existing observations but a DIFFERENT shape
    → re-Remember + recompile (the type EVOLVED). deopt.
guard fails, a field is NOT an existing observation
    → GENUINE error → emit an errno Observation (Skyra's errno). the semantic error signature.
```

## Consistency with the code (already aligned)

```text
· causal.go  Contract struct{ Name Name }  — Contract IS a name. matches "the type is whatever the name of
             the invoked child is."
· causal.go  MoveTemplate.Contract comment already says "the input/return TYPE is negotiated at ACT."
· so the CODE was right; the system-and-semantic PROSE ("CALL writes the contract") was the overstatement.
  tighten the Register.Contract / Contract comments to say: name at CALL, structural type written at REMEMBER.
```

## Open

```text
· the "existing observation" registry — WHERE the universe of known observation types lives (the ABI /
  the semantic interpreter's surface). "each field must be an existing observation" needs that registry.
· HOW Remember stores the contract — alongside the Body? is it part of the engram digest (so a body's
  content address includes its inferred type)? ties to content-addressed PossibilityID.
· the structural composition rules — how a composite return type is built from existing observations, and
  what "different shape" (re-Remember) vs "invalid field" (errno) means precisely.
· first-contact on another machine: a WARM (remembered) contract lives where it was Remembered — does the
  remembered contract travel/fetch with the Body (content-addressed), or is each machine's first run cold?
```
