# Execution brainstorm — operand model, live notes

**Date:** 2026-09-02
**Status: BRAINSTORM / OPEN.** Working notes from a back-and-forth on the operand seam (seam A). NOT decisions
— a captured trail so the thinking is not lost between sessions. The authoritative doc is still
`../operand-model.md`; when anything here hardens, it graduates there (or into causal.go), not into this dir.

The trigger: seam A (the operand model) is the blocker that gates the whole engine — the ready predicate,
`Perform`'s `in`, `PossibilityID = hash(body+inputs)` (seam E), and the entire WARM/reuse path all wait on
it. This dir works the problem out loud.

---

## The docs here

```text
sys-vs-payload.md        the reframe that IS the seam: a wire Move carries a full SYS half (addressing +
                         authority) and an EMPTY data half. the operand = the payload half the wire dropped.
                         includes the printed Move-field table.
leaf-and-directions.md   the two-directions model (forward selects a child · back returns an observation),
                         the leaf definition (orchestrates no dependencies; only returns backward), and
                         leaf-as-producer (a value is PRODUCED by a leaf, not baked as a field — Option 2).
args-decision.md         THE load-bearing bit. "forward carries no data" ⇒ external/request data has no way
                         down except (a) baking it in (kills reuse) or (b) a materialize-time ARG. the model's
                         own axioms push toward YES-args. the WARM path is what's at stake.
code-emission-and-act.md the "two things, not one" thread: an LLM emitting code in its return. why
                         world-touching code is NOT a leaf return but new forward moves (ACT/JIT), and why
                         it must stay OUT of the operand corner. leaf-KIND (crossing taxonomy) vs leaf-opcode.
decision.md              running ledger of leaning calls + explicit open flags. D1: (partly superseded) dissolve
                         action + params into payload. D2: payload undetermined — work the opcode/operand model.
                         D3: target stays in the operand; the arg channel is a separate ABI-owned layer, not the
                         opcode — leaves OPEN where observation params live + how a compiler links them.
                         D4: bind operands NOMINALLY (named registers ⇒ no positional convention); a convention
                         survives only at the remote boundary (nominal Contract) + the Driver edge (marshal).
                         D5: split ACT and COMPILE — declaration is THINK/ACT, COMPILE-proper is REMEMBER
                         (JIT-shaped; revises think-and-act.md).
                         D6: Move is the universal wire; CALL and RETURN are opcodes (reverses the Move/Landing
                         split; closes the barrier-ingress sum-type gap). the WIRE axis, orthogonal to operands.
                         D7: opcodes are TRANSITIONS — two pops (in-flight pop during the run for quiescence vs
                         structural retirement at REMEMBER), and opcodes append to a TRACE that REMEMBER compiles.
                         D8: effectful leaves compute-not-fire, staged at a kernel confirm-gate that is a FAULT
                         (confirm-bit on the horizon @ Bound; new IDT vector); refusal = an interposed corrective
                         observation (seam H), with an in-type (local redo) vs out-of-type (renegotiation) fork.
                         D9: leaf = what it TOUCHES (target), not read/write (those are caps); EXEC is a distinct
                         opcode. reshaped set: CALL = edge, RETURN/EXEC = the two terminals (pure / world-touch).
                         D10: the SOURCE collapse — every fd is written by a move; landed/baked/bound become the
                         writer's TARGET-KIND; fault = a move-to-kernel opcode. reactor STATE survives; commits to
                         Option 2 (everything is a move). open fork: injection-as-move vs Materialize pre-seed.
```

## Where seam A stands

```text
· leaf-KIND: CLOSED by D9 — every special leaf is EXEC(target, action); ambient read = EXEC read-action, no
  operand; effectful write = EXEC + D8's confirm-gate. read/write are caps, not leaf kinds.
· operand-SOURCE: the last axis — confirm {landed, baked, bound} holds for every move incl. EXEC (a no-operand
  EXEC simply has none). the args=YES call + Materialize pre-seed is the frame.
· NEXT: EXEC mechanics under the hood — how EXEC issues, drives, faults (D8), lands, and traces (D7) — to
  clarify the execution model fully. open fork carried in: is EXEC childless or a terminal that may CALL children.
```

## The one open decision this all converges on

```text
ARGS: yes or no — is there a per-invocation argument, bound at MATERIALIZE time, distinct from a compile-time
      literal AND from a child's return? (operand-model.md item 1 — the load-bearing bit.)
      the brainstorm LEANS YES, derived from the model's own "forward carries no data" rule. see args-decision.md.
```

Everything else in seam A (Option 1 vs 2 for literals; the ordered `Operand` list; the PossibilityID hash
input set) is secondary and falls out once the args bit is settled.
