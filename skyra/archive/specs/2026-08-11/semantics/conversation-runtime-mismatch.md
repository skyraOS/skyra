# Conversation / Runtime Mismatch

**Date:** 2026-08-11
**Status: RESOLVED, 2026-08-11.** Written as a note recording mismatches without choosing between
them. All five have since been chosen, and they went the same way.

| current runtime | conversation | resolved by |
|---|---|---|
| Failed children produce stored error observations. | Errors remain ephemeral. | open-until-commit |
| Every invocation enters child history. | Only successful replacement structure matters. | open-until-commit |
| Discovery passes rejected outputs back to avoid retrying them. | The node directly authors a better expectation. | open-until-commit |
| Discovered engrams are reused by pointer and ID. | Reuse copies the expectation into a new locally owned engram. | discovery is always a copy |
| An internal root commit disperses credit immediately. | The user is the ultimate acceptance boundary. | open-until-commit |

**One change decides four of them.** If a possibility returns without closing, and only the root's
commit is durable, then errors are ephemeral because nothing persisted, only successful structure is
kept because closure mints what worked, revision happens in place because nothing died to pass a
failure back to, and dispersal is once at the root because that is the only commit that closes
anything.

The fifth is independent: **discovery is always a copy.** The index is shared and the tree is owned,
so citation is what makes structure yours.

> **The conversation was right in every row.** This note is therefore no longer a list of tensions —
> it is a specification of the conversation side, and the runtime is what disagrees with it.

See [open-until-commit.md](open-until-commit.md) and
[roots-and-admission.md](roots-and-admission.md).
