# syscall-read-example

**Status: WORKING.** A single worked example, traced against the code, used to find where the
entity/host seam actually falls.

```text
example.md    THE WALKTHROUGH. one ask, traced step by step until the information runs out.
              poses the problem. draws no conclusions.

evidence.md   OBSERVATION. commands and their output, numbered E1..E6 so decisions can cite them.
              no conclusions either.

outcomes/     DECISIONS, one document per question. WRITTEN BY MICHAEL ONLY.
                Where-outcomes.md   which machine work runs on
                What-outcomes.md    the work itself
                How-outcomes.md     below the kernel boundary
                Why-outcomes.md     what prompted an invocation
                When-outcomes.md    ordering and time
```

> **Rule for this directory:** nothing in `outcomes/` is mine to fill in. Proposals go in the chat
> log and land there only when Michael puts them there. The example and the evidence may be
> extended; the outcomes may not be inferred from them by anyone else.

**Evidence and decisions are kept apart on purpose.** A measurement is true whoever reads it; a
decision is a choice someone made. Mixing them makes the second look like the first.

The example stops where the information stops. Every place the trace cannot continue is a place the
design has no representation for, and naming those is what the example is for. Deciding what to do
about them is not.
