# Continuation language — ordered observations as operands

**Date:** 2026-09-11  
**Status: DECIDED DIRECTION.** This note records the continuation-language shape
settled in discussion. It does not authorize changing DNA15. The implementation
target and migration sequence remain to be chosen.

## Decision

Remove `Action` as a top-level field. An invocation carries one ordered list of
observations. The position of an observation in that list determines its role:

```text
Invoke.Operands[0]   method header
Invoke.Operands[1:]  method parameters
Return.Observation   returned value
```

The language rule is:

> The way an observation appears in the operand list is the way it is interpreted.

The same `Observation` representation may therefore carry a method header, a
parameter, or a return value. These are different compositions and lifecycle
positions, not different underlying value types.

## An observation is not executable

An observation remains inert data. Execution comes from the containing `Invoke`,
its authorized target, and the observation's position within its operand list.

```text
Observation alone
    = data

Observation in Invoke.Operands[0]
    = method header

Observation in Invoke.Operands[1:]
    = method parameter

Observation in Return
    = returned value
```

The explicit evaluation boundary is:

```text
Invoke
    + Target(entity/address/capability)
    + ordered Operands([]Observation)
```

No executable bit, executable observation subtype, or ambient evaluation rule is
introduced.

## Minimal application grammar

The continuation language has a small application form:

```text
Application := Header Operand*
```

For example:

```text
Operands
+-- [0] "what is the weather"   method header
+-- [1] location = "Boston"     parameter
+-- [2] units = "fahrenheit"    parameter
```

The method header retains its field name and value as an observation. Its header
role comes from position zero, not from a separate `Action` field or Go type.

An implementation may give the ordered slice a named type so the rule has one
home:

```go
type Operands []Observation

func (o Operands) Header() (Observation, bool)
func (o Operands) Parameters() []Observation
func (o Operands) Validate() error
```

This is conceptual. The important decision is the composition law, not these exact
method signatures.

## Scheduler and driver ownership

Continuation-language semantics remain laws of the runtime. Representation and
entity-specific interpretation remain driver concerns.

```text
Scheduler
    recognizes Invoke versus Return
    requires a non-empty operand list for Invoke
    assigns header/parameter roles by position
    enforces operand-name uniqueness
    preserves operand order and identity
    evaluates continuation/register readiness
    applies parent, child, wait, resume, and return transition laws

Dispatcher
    resolves entity/address to the installed driver binding
    carries the capability address and operands to that binding

Driver
    interprets the header's entity-specific meaning
    translates operands into the addressed entity API
    translates entity output back into canonical Move(s)
```

The driver may decode the continuation language's external representation, but it
does not decide whether a decoded transition is legal. Parsed moves re-enter the
same runtime admission and transition laws as every other move.

```text
external / LLM representation
        |
        +-- driver decodes
                |
                +-- canonical Move
                        |
                        +-- scheduler validates and transitions
```

Owning the representation of a law is not owning the law.

## Capability address and method header are not duplicates

The invocation target remains:

```text
entity / entity-address / capability-address
```

That target selects the authorized semantic face and callable entrypoint of an
entity. The method-header observation is input to that entrypoint.

For example:

```text
Target:
    weather / api / answer

Operands:
    [0] method = "what is the weather"
    [1] location = "Boston"
```

The capability address answers **where this operand list may be interpreted**. The
method header answers **what this particular continuation application asks within
that interpretation**.

Authority and continuation-language composition therefore remain separate:

```text
Target capability    authority and dispatch
Invoke/Return        runtime transition form
ordered Operands     continuation-language application
```

See
[`entity-address-driver-transport.md`](../2026-09-10/entity-address-driver-transport.md)
for the stable target and capability-address decisions.

## Current shape and intended refinement

DNA15 and the DNA16 seed currently carry:

```go
type Invoke struct {
	InvocationID InvocationID
	Action       Action
	Params       []Observation
	ReturnFd     InvocationID
	Register     Register
	Horizon      Horizon
	Focus        []Capability
}

type Capability struct {
	Target    Target
	Action    Action
	ExpiresAt int64
	Sig       []byte
}
```

The intended conceptual refinement is:

```go
type Target struct {
	Invocation InvocationID
	Entity     EntityID
	Address    EntityAddress
	Capability CapabilityAddress
	Host       HostID
}

type Invoke struct {
	InvocationID InvocationID
	Operands     Operands
	ReturnFd     InvocationID
	Register     Register
	Horizon      Horizon
	Focus        []Capability
}

type Capability struct {
	Target    Target
	ExpiresAt int64
	Sig       []byte
}
```

These are explanatory shapes, not an approved Go surface. The decided changes are:

```text
remove Invoke.Action
rename/reframe Invoke.Params as ordered Operands
place the method header at Operands[0]
place parameters at Operands[1:]
place CapabilityAddress in the target
remove Capability.Action because the target names the exact callable capability
```

## Identity consequence

The older identity statement was:

```text
EngramID = hash(Action + ordered Params)
```

With `Action` composed into the operand list, that reduces to:

```text
EngramID = hash(ordered Operands)
```

This preserves the old information rather than deleting it: the former action is
now the first identity-bearing observation. Canonical observation encoding and the
final identity algorithm remain separate open decisions.

## Return and recomposition

A return remains one observation:

```go
type Return struct {
	Observation Observation
}
```

A returned observation does not become executable merely because it was produced.
It can become a future method header only when a continuation explicitly composes
it into position zero of a new admitted invocation:

```text
child Return.Observation
        |
        +-- lands as parent data
                |
                +-- explicit composition into new Operands[0]
                        |
                        +-- admitted Invoke
```

This makes data-to-application promotion visible, authorized, and subject to the
ordinary continuation transition rules.

## Related programming patterns

This is closest to a tagged-message or term-algebra design: headers, arguments,
and results share one value representation, while their position and surrounding
form determine their role. It is not the conventional RPC envelope, where method,
parameters, and result are distinct top-level fields.

It also stops short of unrestricted homoiconicity. Code and data may share a
carrier, but an observation never evaluates itself. `Invoke` is the explicit and
authorized evaluation boundary.

## Invariants

```text
Action is not a top-level Invoke or Capability field
an Invoke operand list is non-empty
operand zero is always the method header
remaining operands are parameters
operand names are unique under the continuation-language contract
operand order is preserved
method headers, parameters, and returns share Observation representation
Observation is inert outside an admitted Invoke composition
only explicit placement in Operands[0] grants the header role
entity/address/capability remains separate from continuation-language operands
scheduler owns transition legality; driver owns entity-specific interpretation
```

## Open implementation questions

- The exact structural rule for the header observation's name and content.
- Whether uniqueness covers every operand or only parameters after the header.
- Whether parameter order is semantic, identity-bearing only, or canonicalized by
  name before identity is calculated.
- Whether `Operands` earns a named type immediately or begins as
  `[]Observation` with validation at admission.
- How existing `Action` values migrate into header observations without changing
  established engram identity unexpectedly.
- Where capability-address input/output descriptors attach to the observation
  contract used for driver validation and Focus projection.
