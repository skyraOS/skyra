# 04 · The Grammar

**Package:** `grammar`   ·   161 code lines

Two forms. The incapacity is a **parsing** property, not an API property.

## The forms

```go
type Emission interface{ isEmission() }   // SPEC

type Select struct{ Members []Member }    // SPEC

type Member struct {                      // SPEC
    Engram       string            // the NAME, resolved in frame
    Cite         ids.ObservationID // PER MEMBER, never per group
    Entity       ids.EntityID      // WHO interprets the child
    Observations []ids.ObservationID
    Host         ids.HostID        // optional; only where the host is the TARGET
}

type Commit struct {                      // SPEC
    Content  content.Content       // a LEAF originates
    Observed ids.ObservationID     // an ORCHESTRATOR names
}
```

**Structs, not an open interface, because they are data.** The exhaustive type switch over them is a
guard on the vocabulary: two forms is the central claim, and a switch that fails to compile on a
third case is better than dispatch that silently accommodates one.

`Commit` holds **no `Cites` field**, because citation is derived from the `Ref`s in `Content`. Two
places for one fact would eventually disagree.

## The parser

```go
func Parse(raw string) (Emission, error)

type ErrNoEmission struct{ Raw string }   // ADDED
```

Seventeen cases are asserted in `grammar/parse_test.go` — prose, a tool call, an invented attribute,
a member missing `cite`, a group-level anything. All refused.

```text
THE WIRE FORM, exactly as grammar.md writes it:

  <select>
    <engramName cite=obs-0001 entity=some-entity>obs-0002, obs-0003</engram>
  </select>

  <commit><obs-0007></commit>     an orchestrator NAMES
  <commit>{ … }</commit>          a leaf ORIGINATES

note the asymmetry the spec specifies and the parser honours: a member's OPENING tag is the engram
NAME and its closing tag is the literal </engram>. the name is in tag position because it is the one
thing every member must have.
```

## Review notes

`ErrNoEmission` is an error type carrying one message. It exists so the kernel can distinguish "the
entity said something unparseable" from "the entity was unreachable," which matters because the first
is an ordinary failed invocation and the second is a host problem. Thin, and could be a sentinel.

**Nothing else here is addable to.** This is the package that most directly is the claim, and it is
21 code lines for the forms plus 140 for the parser — the ratio is right, since the forms are the
design and the parser is the enforcement.
