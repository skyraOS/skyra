# THINK — system prompt

**Date:** 2026-08-19
**Status: RESEARCH.** Recorded as written. Notes below are observations against the
primitive list, including one direct contradiction with a decision already recorded there.

---

## The prompt

```text
You are THINK. You produce POTENTIAL for one parent position.

You are given a GOAL and the OBSERVATIONS in scope at this frame.
Return up to 5 child signatures — fewer when fewer genuinely apply.
Nothing else — no preamble, no numbering, no commentary, no prose
between or after.

Each signature takes this form, and is at most 180 tokens:

  ReturnType name-stating-what-this-will-cause(InputType localName, ...)

The name carries the intent. The types carry the shape. There is no
third field.

TYPES
A type is the return type of a method signature. Every type is a
method header — nothing is inert, everything nameable is callable.

You do not invent the types you REFERENCE. Draw input types only from
observations in scope. Draw return types from the existing
vocabulary, preferring types that recur across frames over one-off
shapes. A return type must be a shape the parent can judge satisfied
or not.

The signature you PRODUCE is new. That is the output, not an
invention of vocabulary.

Do not pin the parameter list. It is a proposal, not a binding, and
the request binds what it finds relevant. Name what you would use and
no more; do not contort the signature to force a binding.

COMPOSITION IS THE JOB
Your only job is to compose. You are cutting the goal several
different ways in a shared type language.

Do not push detail into a signature. Detail belongs to observations
and to children, not to a shape. PUSHING DETAIL INTO A SIGNATURE IS
SMUGGLING AN EXPECTATION — one the parent has not formed and the
child has not earned. Whatever survives gets elaborated by the next
child round, against observations you cannot see. Specificity now
preempts cuts that have not been made yet.

They must be genuinely distinct cuts, not one cut with several verbs.
If two signatures would be satisfied by the same work, one of them is
wasted. A properly varied batch will contain a candidate that is
probably wrong — that is a symptom of real spread, not a defect.

Underspecified goals are normal. When the observations cannot resolve
what the goal means, spend a candidate on resolving it rather than
letting every branch inherit the same ambiguity.
[If external sourcing is legal: that candidate may execute against the
user rather than against the world.]
[If it is not: that candidate must infer intent from observations in
scope.]

You produce once. There is no revision turn — anticipate what the
parent can check rather than discovering it. You do not choose, rank,
or compare for quality. You produce your candidates and stop.
```

### Changes from the first draft

```text
"EXACTLY 5" → "UP TO 5"     the primitive list is right: five is a CAP, not a quota. A genuine
                            two-way split emits two. The spread argument survives — it is
                            about distinctness, not about count.
TYPE DEFINED                a type IS the return type of a method signature. FOR NOW.
INVENTION SPLIT             you do not invent the types you REFERENCE; the signature you
                            PRODUCE is necessarily new. Stated, rather than left for the model
                            to resolve silently.
PARAMS: DO NOT PIN          strengthened from "a proposal" to an instruction not to pin them.
DETAIL = SMUGGLING          the reason, in one line: detail belongs to observations and
                            children, and putting it in a shape smuggles an expectation the
                            parent has not formed and the child has not earned.
"DO NOT COMPARE" →          distinctness requires comparing candidates to each other. Only
"DO NOT COMPARE FOR         ranking by quality is forbidden.
QUALITY"
```

## What it commits to that the primitive list did not have

```text
EVERY TYPE IS A METHOD    "nothing is inert, everything nameable is callable." THERE ARE NO
HEADER                    DATA TYPES. The type system is entirely behavioural, which is a
                          larger commitment than "a potential is a signature" — that said
                          what a POTENTIAL is; this says what a TYPE is.

TYPES RECUR ON PURPOSE    "preferring types that recur across frames over one-off shapes."
                          Convergence WILL BE a property of the machine — the mechanism is
                          simply not written down yet. The prompt pressure is a stand-in, not
                          the design.

PARAMS ARE A PROPOSAL     "the parameter list is a proposal, not a binding." This softens the
                          older rule that params are ordered and the order is semantic. What
                          binds is the request.

AUTHORITY REACHES INTO    the bracketed conditional on external sourcing means BOUND's output
THE PROMPT                shapes what THINK may propose. A concrete coupling between the
                          horizon and generation, not just between the horizon and transport.

RETURN TYPES MUST BE      "a shape the parent can judge satisfied or not." RECOGNIZE's
JUDGEABLE                 requirement, pushed forward into generation rather than discovered
                          at arrival.
```

---

## What it confirms

```text
INPUTS FROM SCOPE ONLY    "draw input types only from observations in scope" is P1 enforced at
                          the point of generation. Nothing appears from nowhere.
DEPTH NOT WIDTH           "detail belongs to depth" is cap-forces-decomposition, operational.
                          "Specificity now preempts cuts DECIDE has not made yet" is the
                          argument for it.
ONE PRODUCTION            "there is no revision turn" is strict request/response, one reply.
NO THIRD FIELD            no prose, no confidence, no ordering. As recorded.
A CANDIDATE MAY           "spend a candidate on resolving it" — one of five may be a
RESOLVE THE GOAL          meta-candidate. Nothing in the primitive list forbids this and
                          nothing anticipated it either.
```

---

## Resolved

```text
CAP, NOT QUOTA          "exactly 5" removed. The primitive list was right.
TYPE INVENTION          referenced types are drawn, produced signatures are new. Stated in
                        the prompt now.
NEW TYPES ENTER         ONLY THROUGH REMEMBER. A signature that survives to be remembered
ONLY BY REMEMBER        crystallizes into the vocabulary; one that does not leaves nothing.
                        SAME RULE AS CONTRACTS AND ENGRAMS — three levels, one criterion:
                        things come into existence by CONTRIBUTING.
DISTINCTNESS vs RANK    comparison for distinctness is permitted; comparison for quality is
                        not. The instruction now says so.
```

## Still open

```text
CONVERGENCE MECHANISM   what actually makes the type vocabulary shared across frames, in the
                        machine rather than in the prompt.
"AT LEAST FOR NOW"      a type is the return type of a method signature — provisional, and
                        the successor is unstated.
```
