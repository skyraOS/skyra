# Approach is the gate; root is the account

**Date:** 2026-08-15
**Status: CONVERSATION RECORD. UNRATIFIED.** Third session. Written down before it is lost. Not
checked against the code beyond the citations given.

> **Approach is the only gate, and it has two uses: admitting an outside party, and elevating an
> inside one. Sign-up is interior provisioning and was never part of it. `root` is the namespace
> with no provisioner — dormant, never in the call path, and the word is free because the sense
> that was retired was a different one.**

---

## 0 · Where this came from

The sandbox thread ended at *namespaces are engrams, so authority differences are positional*,
which walked straight back into `root` — a word `../../semantics/terminology.md` retired. The
question was whether AWS's root account is a counter-example. It is not; it is a confirmation.

---

## 1 · Approach's decomposition is already written, and it is not meet/greet

`../../semantics/terminology.md` addendum:

```text
APPROACH        the primitive. should this be let in?
  authenticate  answers WHO — and this is the half §5 says is missing
  Mount         answers WHAT, already designed: grants ⊆ ceiling
```

**Both halves are sign-in.** So the question *"gate in front of meet/greet, or dissolve Approach
into them"* is malformed:

```text
APPROACH  ==  GREET  ==  authenticate + Mount
```

That is `login` + `setuid` exactly. Nothing is fronted and nothing dissolves.

## 2 · Meet is interior. It was never part of the gate

`useradd` is not a second entrance in Linux:

```text
login     PUBLIC.    reachable from outside. the gate.
useradd   INTERIOR.  you must already hold the authority. NOT AN ENTRY POINT.
```

Creating a namespace is **provisioning**, which already exists without touching Approach:

> *"PROVISION — the ACT of handing down access from access you hold. it can only narrow. boot
> provisions the admin namespace; the admin namespace provisions the rest."*

```text
GREET / SIGN-IN   = APPROACH     authenticate + Mount. the only way in.
MEET  / SIGN-UP   = PROVISION    interior. an admitted namespace narrowing what it holds.
```

### One door, and that is the property being protected

> `../../continuation-language/primitives.md` §1: *"It is the only way anything enters, which puts
> the trust boundary and the gate on the same object — a gate anywhere else would be guarding a
> door with no wall attached."*

Two public entrances means two walls. Dissolving Approach into meet + greet buys exactly that.

### Public sign-up does not break it

```text
ANONYMOUS PARTY
    ↓ APPROACH
PUBLIC NAMESPACE, near-zero ceiling
    ↓ ordinary invocation
SIGNUP BEHAVIOR — permitted to that namespace
```

The anonymous party never holds authority to create an account; it asks something that does.
Supported by `primitives.md` §1 as written: *"Approach opens a namespace to an outside party; it
does not create the branch that later executes an ask."*

```text
PROPOSED, NOT RULED   the thin-anonymous-namespace mechanism. it needs no new machinery, but
                      whether a near-zero ceiling is expressible depends on H3.
```

---

## 3 · Bootstrap resolves outside the gate, in both systems

```text
LINUX   the installer creates root before login can run.
SKYRA   "boot provisions the admin namespace."
```

Which is why `primitives.md` can say *"Decide cannot express the first entry: it is a parent
choosing, and the first decision has no parent"* without that being a problem. **The first entry is
not a decision. It is boot writing a ceiling.**

---

## 4 · root = the namespace with no provisioner

**DECIDED: the root account is the admin account.** A rename plus a derivation, not a new type.

```text
NOT   a flag, a reserved name, or a special row on the rack.
BUT   read off the provisioning lineage. boot wrote its ceiling; nothing else did.
```

Which keeps `terminology.md`'s formula intact — *"what makes admin special is not a position on
the rack. It is that its ceiling was written by boot rather than by another namespace"* — and keeps
the rack flat.

```text
PROPOSED, NOT RULED   the derived formulation. Michael ruled the identification (root == admin);
                      "no provisioner" as the definition is this document's.
```

### Why the word is available

`terminology.md` retired `root`, but for a different sense:

```text
ROOT-OF-A-TRAVERSAL   the top of an ask.      RETIRED — "it implied a special node and was not
                                              one." what it named is a BRANCH.

ROOT-AS-ACCOUNT       the identity with no    NEVER EXISTED. admin filled the role unnamed.
                      provisioner.
```

The retirement removed the *conflicting* meaning. And the implication that made it wrong before —
that it names something special — is **correct** in the account sense.

`terminology.md` already reaches for the analogy while retiring the name:

> *"admin signs out and I am logged out with them, which is wrong **the same way `root` not being
> logged in does not stop you using your account**."*

---

## 5 · Root is not in the call path

Verified against AWS, which was the case that prompted this:

```text
CREATED       with the account itself. tied to the account email. HAS NO PROVISIONER.
USED FOR      a short list nothing else can do — change account email/name, close the account,
              change support plan, restore IAM access after lockout, some billing/marketplace ops.
NOT USED FOR  anything routine. AWS guidance: MFA it, issue no access keys, never use it daily.
RESTRICTABLE  no — IAM cannot constrain root in a standalone account. Only an Organization SCP
              can constrain a MEMBER account's root.
```

Ordinary operations authorize against IAM policies. **Root is not consulted, signs nothing, and
need not exist in any active sense.**

The design already matches this:

> *"`Mount` reads it and compares, so **no live admin is needed to authorize a sign-in**."*

### And you do not act AS root

```text
LINUX   useradd needs root's AUTHORITY — but you run `sudo useradd`, as yourself.
AWS     you do NOT create IAM users as root. create an admin identity early and use that.
```

So `meet` requires **sufficient ceiling**, not **being root**. Root is the origin of the authority
chain, not who walks it. This is what keeps it dormant.

---

## 6 · Approach-as-gate is not admin-as-account

Linux keeps these separate and so should this:

```text
`login` RUNS AS ROOT   the privileged PATH. no account of its own. runs before anyone is admitted.
uid 0                  the privileged ACCOUNT. dormant. not consulted during login.
```

```text
APPROACH        the gate. kernel-owned, absolute, no namespace, no provisioner, not an account.
ADMIN / ROOT    an account. the namespace with no provisioner. dormant.
```

The load-bearing property behind why the gate must be maximally privileged:

> **The thing that authenticates you must be more privileged than any account it can log you into,
> because it has to be able to become anyone.**

And it never hands that privilege over — `login` drops to the user rather than granting root.
Authority flows down, never sideways.

---

## 7 · The credential check is an endpoint, not a behavior

The worry: if auth is compiled behavior, an authored or trace-derived continuation could be the
thing checking passwords. Linux answers this with PAM, and the split is exactly the one the sandbox
session already chose:

```text
/etc/pam.d/*   THE STACK.  declarative config — which modules, in what order, what is required.
               root-owned, among the most protected files on the system.

pam_unix.so    THE CHECK.  vetted native code loaded into the privileged process.
               not authored, not derived, not scripted.
```

Mapped:

```text
THE LOOP     compiled behavior. skyra language. safe to instrument, safe to solidify.
THE CHECK    AN ENDPOINT PACKAGE. runtime-provided, vetted, held ONLY by Approach.
```

Reading a credential store is an **effect**, not a computation, so it is mediated like every other
effect. A derived continuation can only ever call an endpoint it was handed — which is why the auth
loop can solidify from traces without danger.

---

## 8 · Linux ran the H3 experiment

H3's addendum asks whether Horizon narrows per-axis or monolithically.

```text
uid 0 WAS MONOLITHIC   all-or-nothing root.
IT WAS DECOMPOSED      into ~40 capabilities — CAP_NET_ADMIN, CAP_SYS_ADMIN, CAP_DAC_OVERRIDE, …
BECAUSE                all-or-nothing was too coarse to express real policy.
```

**Evidence for per-axis narrowing, from a system that started where the opaque `Horizon` currently
is and moved away from it.** The monolithic version is the one that got abandoned.

---

## 9 · Sudo is Approach

**DECIDED: they are the same thing.** One gate, two uses.

```text
GREET / SIGN-IN   outside party → admitted namespace.
SUDO / ELEVATE    admitted party → a branch with a raised, ceiling-bounded Horizon.
```

Both are `authenticate` + a ceiling comparison. The second half differs only in what it compares
against: `Mount` checks `grants ⊆ ceiling` at admission; elevation checks the same containment at
branch open.

### The gap sudo fills already exists

```text
CEILING   the maximum grants a namespace may EVER be mounted with.   DURABLE.
GRANTS    what this session actually holds.                          chosen at Mount.
```

That gap **is** sudo's gap. `sudoers` is the ceiling; an ordinary session is the grants. Mount with
little; elevate toward the ceiling for one act. No new authority concept is introduced.

```text
THE INVARIANT   ELEVATION NEVER EXCEEDS THE CEILING.
                the same check Mount already performs, at a different moment.
```

### It does not widen a Horizon — it starts a chain with a higher one

**DECIDED.** This is what keeps `PROVISION`'s can-only-narrow rule intact.

Real `sudo cmd` does not elevate your shell. It forks a new process with different authority and
leaves your shell untouched. The analogue:

```text
BOOT → admin ceiling → your ceiling      PROVISIONING. only narrows. durable.
your ceiling → session grants            MOUNT. a subset.
                                              ↓
                                   SUDO OPENS A NEW BRANCH
                                   initial Horizon raised, still ⊆ ceiling
                                              ↓
                                   within that chain, Horizon only narrows
```

So `CHILD HORIZON ⊆ PARENT HORIZON` is untouched. Elevation happens at **the root of a chain**,
before any narrowing begins — the same position at which Approach already establishes an initial
Horizon in §1. Nothing mid-tree ever widens.

Which settles what skyra-sudo is structurally:

> **A branch, not an operation.**

### Re-authentication keeps one gate and one endpoint

`sudo` re-prompts because holding a session is not proof you are still present. The credential
check lives behind an endpoint only Approach holds (§7), so either elevation gets that endpoint —
and Approach stops being the only gate — or elevation *is* Approach, invoked from inside.

Taking the second is what makes "Approach and sudo are the same thing" load-bearing rather than
cosmetic: **one gate, one credential endpoint, one place the trust boundary lives.**

It is defensible on substance too — the credential comes from outside the system, a human's memory,
so proving you are still that party is genuinely outside-in.

### The contradiction this creates, stated plainly

```text
kernel/02-approach.md §1   "NOT a method on Kernel. Approach is encountered from OUTSIDE."
kernel/01-protocol.md §1   APPROACH is deliberately absent from the possibility-facing interface —
                           "a possibility already inside the system cannot call it."
```

If elevation is Approach and elevation is requested from inside, **Approach must be callable from
inside**, and both statements above are wrong as written.

Two ways out, neither taken here:

```text
A  APPROACH GAINS A POSSIBILITY-FACING FORM.  the gate is reachable from within, and
                                              "always outside-in" narrows to "always crosses a
                                              trust boundary" — which re-authentication does.

B  ELEVATION IS A DISTINCT PRIMITIVE          sharing Approach's endpoint. keeps the interface
   THAT LOOKS LIKE APPROACH                   rule, at the cost of the unification just decided.
```

A is what the decision implies. It requires amending `primitives.md` §1's *"always outside-in,"*
which is the phrase both interface documents are derived from.

### Audit comes free, and is better than syslog

`sudo`'s logging is bolted on. Here the elevated branch is a possibility with a span, and Remember
retains its cone:

```text
SYSLOG GIVES      who ran what command, when.
THE CAUSAL CONE   the exact expectation, the entity, EXACTLY WHICH OBSERVATIONS WERE IN FRAME,
                  what was decided, what returned, and what was remembered.
```

Nothing to build. An elevated branch is auditable because every branch is.

### Deferred: `sudo -u`

**DEFERRED BY DECISION. Not being built now.**

Real sudo does two things under one name:

```text
sudo cmd          RAISE MY OWN AUTHORITY toward my ceiling.     ← this is what is decided above
sudo -u postgres  ACT AS ANOTHER PRINCIPAL.                     ← deferred
```

The second is different in kind — authority that was never yours, and the causal record has to say
whose. It also depends on `PRINCIPAL`, which is currently undefined in two of the three loose docs
(`creator-and-compiler.md` §9 edge 4, and §9 edge 3 below). Settle that first.

---

## 10 · Open edges

```text
1  IS meet PUBLIC OR CEILING-GATED?   §2 takes the Linux separation, which makes it interior.
                                      if a deployment wants genuinely public sign-up WITHOUT the
                                      thin-namespace indirection, provisioning gains a second
                                      origin besides boot, and "root = no provisioner" degrades
                                      from a definition to a usually.

2  WHAT IS A NEAR-ZERO CEILING?       §2's public namespace needs one, and H3 is unresolved on
                                      what a ceiling is made of.

3  THE ANONYMOUS PRINCIPAL            terminology.md: "PRINCIPAL — WHO ASKS. holds a namespace. an
                                      entity that can INITIATE." an anonymous approacher holds a
                                      namespace only after admission. what it is BEFORE that is
                                      unnamed. compounds creator-and-compiler.md §9 edge 4, where
                                      PRINCIPAL is already undefined once Entity dissolves.

4  APPROACH'S OWN ENGRAMS             Michael's "approach has 2/3 engrams." if greet is Approach
                                      and meet is interior, greet's internal decomposition
                                      (prompt · check · retry · admit) is unspecified. the retry
                                      loop is the part that may legitimately be compiled behavior.

5  DOES ROOT NEED THE SHORT LIST?     AWS reserves ~6 operations only root may perform. whether
                                      skyra has an equivalent — things no provisioned authority
                                      may do, only the unprovisioned namespace — is not asked
                                      anywhere.

6  APPROACH'S INTERFACE               §9 makes Approach callable from inside, contradicting
                                      kernel/02-approach.md §1 and kernel/01-protocol.md §1.
                                      option A — a possibility-facing form, narrowing "always
                                      outside-in" to "always crosses a trust boundary" — is what
                                      the decision implies but was not explicitly ruled.

7  ELEVATION TIMEOUT                  sudo caches authentication ~15 minutes. whether an elevated
                                      branch's authority expires, and whether that is wall-clock
                                      or scoped to the branch's life, is unaddressed. the branch
                                      form suggests the latter comes free — the authority ends
                                      when the chain closes — which may be strictly better than
                                      sudo's timestamp cache.

8  WHAT REQUESTS ELEVATION            a branch is opened by an ask. whether elevation is a property
                                      of the ask, a separate call before it, or an outcome of
                                      Approach returning a wider Admission, is unspecified.
```

---

## 11 · Provenance of the claims

```text
CITED FROM SOURCE   Approach's authenticate + Mount decomposition · the one-door argument ·
                    "Approach opens a namespace... does not create the branch" · PROVISION ·
                    "boot provisions the admin namespace" · "no live admin is needed" ·
                    the root analogy already in the flat-rack argument.

VERIFIED EXTERNALLY AWS root's behavior and reserved operations · the Linux login/PAM path ·
                    the uid-0 → capabilities decomposition. Not from the spec set.

DECIDED IN SESSION  meet and greet separate · Approach stays as the gate · root account == admin
                    account · take the Linux experiment · APPROACH AND SUDO ARE THE SAME THING ·
                    elevation starts a new chain rather than widening a Horizon · sudo -u is
                    DEFERRED.

PROPOSED, NOT RULED Approach IS greet with meet fully interior · the thin-anonymous-namespace
                    route for public sign-up · "root = the namespace with no provisioner" as the
                    derivation · re-authentication as the reason one endpoint suffices ·
                    resolution A for the interface contradiction in §9 · everything in §10.
```
