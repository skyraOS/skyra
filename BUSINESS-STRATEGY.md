# Business Strategy

**Date opened:** 2026-08-15
**Status: FIRST DRAFT.** The first time the commercial question has been thought about
deliberately. Everything here is a position to argue with, not a plan to execute.

**Scope.** This is the commercial track only. The research track is
[`thesis-proposal.md`](thesis-proposal.md) and is **not** decided here; the licensing position is
already set in [`COMMERCIAL.md`](COMMERCIAL.md) and is not reopened. The one place they touch is
noted in §8.

---

## 1 · The strategy in one line

> **Sell the ability to prove what an AI agent could and could not have done — starting with a
> report that measures the gap, before the runtime that closes it exists.**

Everything below is the argument for that sentence and the sequencing it implies.

---

## 2 · The product is at the world boundary, not the causal origin

The instinct is to put the product where the user types. That is the wrong end.

```text
CAUSAL ORIGIN     the ask. "tell it what to do."
                  UNDIFFERENTIATED. every agent product on earth sells this.

WORLD BOUNDARY    where effects land, integrations live, credentials are spent, and
                  failures originate.
                  THIS IS THE SURFACE. it is where the risk is, and it is the only place
                  a claim can be made that a competitor cannot also make.
```

Two consequences follow, and both are already mechanisms in the design rather than aspirations:

**Ownership.** Every position has an owner. The parent owns the signature and return; the child
owns the body. When something fails, the owner nearest the failure attempts repair first, because
that is where the frame is smallest and the fix is cheapest.

**Escalation.** When the nearest owner cannot fix it, the failure travels inward along causality
one position at a time, and the last stop is a human. Most failures never get there.

```text
THE PROMISE   every part of your system has an owner that fixes it.
              you only see what genuinely needed you.
```

```text
⚠️  THE METAPHOR IS ON-CALL, NOT CODEOWNERS. GitHub's CODEOWNERS routes notifications; this
    repairs. And saying "AI code owners" drops the product into a crowded category — PR review,
    autocomplete — that has nothing to do with what it is.
```

---

## 3 · The wedge: a permission scanner, delivered before payment

The runtime cannot be sold yet (§9). The **measurement** can be, immediately.

```text
WHAT IT DOES     inventory what an organization's agents are connected to, what scopes were
                 granted, and what authority the work actually requires.
                 Output: the gap, plus a reviewed implementation plan.

WHY IT SELLS     it is evaluable before purchase. Security buyers cannot buy what they cannot
                 assess, and a report is assessable. Handing it over before money changes hands
                 removes the only real objection.

WHY IT IS        the finding is dramatic and true in almost every deployment: agents hold
POWERFUL         orders of magnitude more authority than their work requires.
```

Build it in two halves, and the easy half is the high-signal one:

```text
1  INTEGRATION INVENTORY    connectors, OAuth manifests, granted scopes, reachable data classes.
                            CONFIG-LEVEL. no dataflow analysis. cheap, dramatic, defensible.
                            → DO THIS FIRST.

2  CODE ANALYSIS            what the code actually reaches for. static inference across dynamic
                            dispatch, reflection, config-driven behavior and third-party
                            dependencies OVER-APPROXIMATES BADLY.
                            → SECOND, OR NOT AT ALL.
```

```text
PRECEDENT WORTH KNOWING   the closest existing analogue — cloud least-privilege policy
                          generation — derives policies from OBSERVED ACTIVITY rather than
                          static analysis, for exactly the over-approximation reason. That is
                          not an argument against the scanner. It is an argument for what the
                          scan is FOR (§5).

DO NOT PROMISE "EXACT"    an exact plan is a promise that breaks on the first large monorepo.
                          Confidence levels and explicitly marked unknowns are more credible to
                          a security buyer anyway; they are trained to distrust completeness.
```

---

## 4 · Why the SaaS surface beats cloud IAM

This is the strongest single argument in the strategy and it should lead every conversation.

Cloud IAM is fine-grained. **SaaS OAuth scopes are not.** Slack grants workspace-level scopes.
Gmail's read scope means the entire mailbox, all history, with no per-thread or per-label
narrowing. Drive is the same. So when a company connects an agent to those systems:

> **They are not over-granting through carelessness. They are over-granting because no narrower
> grant exists to give.**

Which flips the pitch from *better* to *necessary*:

```text
THE PROVIDER'S SCOPE IS THE COARSEST THING IN THE CHAIN.
IF NARROWING HAPPENS AT ALL, IT HAS TO HAPPEN AT THE RUNTIME.
```

No one fixes this at the OAuth layer. Google is not going to ship per-thread scopes. The only
place the boundary can be drawn is between the credential and the agent.

```text
AND THE SEVERITY IS HIGHER   cloud over-permission is a chronic condition security teams are
                             numb to. "this agent can read every email your company has ever
                             sent" is a one-page finding that reaches a CISO the same day, and
                             drags GDPR / HIPAA / SOC 2 in with it.
```

---

## 5 · The mechanic: the scan is the ceiling, the runtime finds the floor

The scanner and the runtime are not two products. They are the two halves of one mechanism that
already exists in the architecture as two types.

```text
THE SCAN      static. over-approximates. what this system COULD need.        → CEILING
THE RUNTIME   observed. narrows on contact. what it ACTUALLY used.           → HORIZON
              the delta between them is the entire value proposition,
              rendered as a number that goes down over time.
```

State this openly rather than pretending the scan is precise. It converts the weakest technical
claim into a demonstration of the mechanism: the scan gives a defensible ceiling on day one, the
runtime narrows it from what actually happens, and **the narrowing is the product working.**

### The metrics

```text
OVER-GRANT DELTA    granted authority ÷ exercised authority.  ↓ over time = working.

ESCALATION RATE     share of world-boundary failures reaching a human.
                    ↓ the system is learning.
                    → structure accumulating without improving fit.
                    ↑ THE WORLD IS MOVING FASTER THAN THE SPECS — which is not a product
                      failure. It is the product reporting a true fact, and no competitor
                      can report it at all.
```

Escalation rate is also the answer to alert fatigue, which is the canonical complaint of the
buyer being sold to. That is not a coincidence worth wasting.

---

## 6 · What is actually defensible

Most "moats" in this space are not. These two are.

```text
1  YOU CANNOT BOLT AN OBJECT-CAPABILITY MODEL ONTO AN EXISTING AGENT FRAMEWORK.

   The incumbent frameworks have ambient authority by design — a tool gets the process's
   credentials and the process has everything. Retrofitting monotone narrowing is not a feature
   addition, it is a rewrite of the execution model. THIS IS THE MOAT, and it has a shelf life:
   it holds until someone else builds a capability-first substrate, not until someone else
   ships an agent product.

2  CAPABILITY PROFILES COMPOUND ACROSS CUSTOMERS.

   "Here is the authority a Slack reader actually needs" is learned once and worth something
   everywhere. It is structure, not customer data, so it can be shared without touching the
   thing a security buyer will never let you touch.
```

```text
WHAT IS NOT A MOAT   the scanner. It is a few weeks of work for a funded competitor, and it is
                     lead generation, not defensibility. Treat it as the wedge it is.

THE REAL RISK        an incumbent identity or security vendor ships agent permissions as a
                     feature. Speed and depth are the only answers, and depth is the one you
                     already have several years of.
```

---

## 7 · Sequencing, against the actual constraints

The constraints are real and the plan has to survive them: **one year of school left, a target of
~40 hours a week, and the goal is to fund the work rather than to maximize outcome.**

```text
NOW → 6 MONTHS      the scanner's integration-inventory half. Scan real environments. The
                    finding is the artifact and the artifact is the sales motion.

                    GOAL: evidence about what real permission surfaces look like, from more
                    than one organization. Nothing else at this stage compounds.

6 → 12 MONTHS       let the scan data determine the runtime's shape. Do NOT build the runtime
                    first — the whole advantage of leading with measurement is that it tells
                    you which integrations and which effect classes actually matter.

POST-GRADUATION     the company decision, made with data instead of conviction. Whether it is a
                    product, a consultancy, or a portfolio piece that gets you hired should be
                    answerable by then from the scan results alone.
```

```text
⚠️  THE TRAP        a scanner plus a report plus a review call IS A CONSULTANCY. That is a
                    legitimate way to fund the work and a fatal way to spend the year by
                    accident. Decide deliberately which of the three it is:

                       LEAD-GEN ARTIFACT   free, automated, drives runtime interest.
                       PRODUCT             recurring, automated, compounds.
                       SERVICES LINE       pays now, compounds never, consumes all available time.

                    Only the middle one is a business. The first is the honest default for
                    someone with a thesis to write.
```

---

## 8 · Where this touches the research track

One intersection, stated so it is used rather than stumbled into:

> **Scanning real environments produces an empirical result — the measured gap between granted
> and required authority — which is evidence no design argument can supply.**

Whether that belongs in [`thesis-proposal.md`](thesis-proposal.md) is a research decision and is
not made here. Recorded only because the same fieldwork can serve both tracks, and forking effort
across a thesis and a company is the failure mode most likely to end this year badly.

```text
⚠️  IP. If employment or funded research enters the picture, get a written carve-out naming this
    project BEFORE signing, not after. Agent infrastructure is exactly the field where an
    employer's assignment clause bites, and the git history is the prior-art record.
```

---

## 9 · Three things that must not be claimed yet

Each is already written down elsewhere in this repo. Each will be under enormous commercial
pressure to blur, and blurring any of them ends the credibility the whole strategy rests on.

```text
1  COMPLETE MEDIATION DOES NOT EXIST YET.

   Until every exit — network, storage, execution, credentials, inheritance, denials — is
   mediated, the confinement proof is a DESIGN TARGET, not a property of the system. Selling it
   as a property is the classic security-product failure, and it is unrecoverable.

2  THE THREE PROOFS MUST NOT COLLAPSE INTO "VERIFIED".

   Causal, confinement, and artifact proofs answer three different questions and none
   substitutes for another. One word covering all three is the marketing instinct and it is
   wrong.

3  DEPLOYMENT STRENGTH IS NOT UNIFORM.

   Customer-controlled, attested, and vendor-controlled execution produce genuinely different
   claims. Vendor-controlled bottoms out at operational trust, NOT cryptographic non-custody.
   Say so. Competitors will not, and the buyers who notice are the ones worth having.
```

The licensing position already supports the honest version of this: source-available under
FSL-1.1-ALv2 means the boundary is inspectable, which is the only way a confinement claim is
worth anything. **This is not security by obscurity, and the license is the evidence.**

---

## 10 · One substrate, many domains — and the discipline that makes it work

The observation that started this document: the underneath does not change, the skin on top does.
That is correct and it is platform economics. Two cautions on executing it.

```text
SKINS ARE CHEAP TO IMAGINE AND EXPENSIVE TO SUPPORT.
Each one is a sales motion, a compliance surface, a support burden, and a set of integrations
that break independently. At 40 hours a week you can carry ONE, well.

IN SECURITY, THE SUBSTRATE IS THE CREDENTIAL.
Buyers are paying for a boundary that has been reviewed and beaten on for years. A vendor
visibly rotating brands over one engine reads as churn even when the engine is excellent.
```

```text
⚠️  DO NOT USE THE RESTAURANT ANALOGY. What good chains reuse — supply chain, training,
    equipment — is the right model. What the analogy makes people hear is the other pattern:
    same kitchen, new sign, because the last name burned its reputation. In an enterprise
    security market that association is fatal.

    And the analogy hides your actual advantage: restaurant concepts do not teach each other
    anything, while a capability profile learned in one domain is worth something in the next.
    Yours COMPOUNDS; theirs merely AMORTIZES.
```

**The discipline:** build so the substrate is portable, pick one domain, and stay in it long
enough that the substrate earns a name. The word is *platform*, never *rebrand*.

---

## 11 · Parked deliberately

```text
CROSS-COMPANY CAUSALITY   already anticipated in the design set and already fenced: no global
                          registry, no implicit crossing, no global mass, no selection shortcut.
                          The naive version — reputation aggregated across organizations — is
                          the global goodness prior the architecture rejects by name.

                          AND THE MARKET FIGHTS IT: the customers who buy confinement proofs are
                          precisely the ones who will never share causal records. The version
                          that could survive is sharing STRUCTURE, NOT DATA (§6.2).

NON-SECURITY DOMAINS      real, and later. See §10.

THE RUNTIME AS A PRODUCT  not until §9.1 is closed.
```

---

## 12 · Decisions to make, in order

```text
1  IS THE SCANNER LEAD-GEN, PRODUCT, OR SERVICES?          §7. blocks everything after it.
2  WHICH INTEGRATIONS FIRST?                               probably Slack + Google Workspace,
                                                           on scope coarseness and data
                                                           sensitivity. UNVALIDATED.
3  DOES THE SUBSTRATE GET A SEPARATE NAME FROM THE
   PRODUCT?                                                §10 requires the answer to be yes
                                                           eventually; not urgent now.
4  WHO IS THE FIRST SCAN FOR?                              a design partner who will let you
                                                           publish the anonymized delta is worth
                                                           more than one who pays.
5  WHAT CLOSES §9.1?                                       the shortest path to complete
                                                           mediation for ONE integration class
                                                           is the first real engineering
                                                           milestone on the commercial track.
```
