# Dispatch and operand ownership

**Date:** 2026-09-11  
**Status: DECIDED DIRECTION.**

```text
Target path:
    entity / address / capability
    └──────────────┘   └────────┘
       dispatcher         driver

Operand composition:
    [method header, parameter, parameter...]
             └──────── scheduler interprets structure
                          driver interprets meaning
```

## Ownership

- **Scheduler:** owns the continuation-language laws. It recognizes the ordered
  operand composition, assigns position zero as the method header, treats the
  remaining observations as parameters, and decides when the invocation is
  structurally valid and executable.
- **Dispatcher:** resolves `entity/address` to the active driver binding. It does
  not interpret the capability or operand contents.
- **Driver:** owns the addressed entity API. It selects the `capability`,
  interprets the method header and parameters in that capability's context, and
  performs the operation.

An observation is inert by itself. The scheduler makes the composition
executable; the driver gives its contents entity-specific meaning.

Authority remains a kernel concern: resolving a target or understanding its
operands does not grant permission to invoke it.

See [continuation-language-operands.md](continuation-language-operands.md) for the
full continuation-language decision.
