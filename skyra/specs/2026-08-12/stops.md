# Stops

**Status: LOG.** Append-only. One entry per item that could not be completed as written.

An entry is not a failure — it is the work order being wrong, which is the thing worth knowing in
the morning.

```text
FORMAT

## <ITEM ID> — <one line>
WHAT WAS ASKED    the instruction, quoted.
WHAT WAS HIT      the specific thing that required a decision, with file:line.
WHAT WAS LEFT     the state the code was left in. should be UNCHANGED.
```

---

*(empty)*
