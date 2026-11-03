# 55 — One tool trees: the invariant

Date: 2026-11-03

Status: **draft; step 1 done.** The functional core of [54](54-one-tool-per-model-experiment-proposal.md),
linked into the apprentice. Typechecks in core Lean 4.35; every theorem rests
only on `propext`, `Classical.choice` and `Quot.sound`.

## Contract identity proof

```lean
/-- A binary by its hash, with its scheme: `cdhash:` and 40 hex digits on
macOS, `sha256:` and 64 on Linux. -/
structure Binary where
  val : String
  deriving DecidableEq

/-- A schema by its canonical JSON, as `contract.Normal` writes it. -/
structure Schema where
  val : String
  deriving DecidableEq

def hexDigits (s : String) (n : Nat) : Bool :=
  s.length == n && s.all fun c => c.isDigit || ('a' ≤ c && c ≤ 'f')

/-- A known scheme and its digest. -/
def Binary.wellFormed (b : Binary) : Bool :=
  (b.val.startsWith "cdhash:" && hexDigits (b.val.drop 7).toString 40) ||
  (b.val.startsWith "sha256:" && hexDigits (b.val.drop 7).toString 64)

/-- The one check on a shape, wherever one is taken. Non-empty for now;
checking that it is canonical (`contract.Normal`) goes here. -/
def Schema.wellFormed (s : Schema) : Bool := !s.val.isEmpty

def wellFormed (binary : Binary) (shape : Schema) : Bool :=
  binary.wellFormed && shape.wellFormed

/-- A contract is a binary and the shape of its input. Its constructor is
private: every contract is made by `identify`, and carries the proof it passed. -/
structure Contract where
  private mk ::
  binary : Binary
  shape  : Schema
  wf     : wellFormed binary shape = true
  deriving DecidableEq

/-- The one way to a contract. Refuses a binary without a known scheme and
digest, or an empty shape. -/
def identify (binary : Binary) (shape : Schema) : Option Contract :=
  if h : wellFormed binary shape = true then some ⟨binary, shape, h⟩ else none

theorem identify_parts (c : Contract) : identify c.binary c.shape = some c := by
  cases c; simp [identify, *]

theorem identify_eq_some : identify b s = some c ↔ c.binary = b ∧ c.shape = s := by
  constructor
  · intro h; unfold identify at h; split at h <;> cases h; exact ⟨rfl, rfl⟩
  · rintro ⟨rfl, rfl⟩; exact identify_parts c

theorem contract_identity (a b : Contract) :
    a = b ↔ a.binary = b.binary ∧ a.shape = b.shape := by
  cases a; cases b; simp
```

Every contract goes through `identify`, which refuses a binary without a
known scheme and digest, or a shape that is not well formed, and every contract carries the
proof that it passed. So rebuilding a contract from its parts always gives it
back (`identify_parts`). The scheme is part of the identity: the cdhash and
the SHA-256 of one binary are different contracts, never confused. Whether
the hash belongs to the binary is the runtime's to compute. `Binary` and
`Schema` are distinct types, so swapped arguments do not typecheck. Every shape, input
or output, is checked by one predicate, `Schema.wellFormed`: non-empty for
now, and the one place a canonical check goes when Lean takes it from
`contract.Normal`.

A contract is identified by the pair, never by a second hash of it. Two
contracts are the same exactly when their binaries and shapes are; no proof
assumes anything about collisions. One binary may have several shapes, each
its own contract. A return shape waits until the trees need one.

## Observation identity proof

```lean
/-- The type of an observation: the contract it came back through, and its
own output shape. One invocation of a contract gives many observations, each
with its own shape. Its constructor is private: every one is made by
`ObservationType.of`, and carries the proof it passed. -/
structure ObservationType where
  private mk ::
  contract : Contract
  output   : Schema
  wf       : output.wellFormed = true
  deriving DecidableEq

/-- The one way to an observation type. Refuses an output shape that is not
well formed. -/
def ObservationType.of (contract : Contract) (output : Schema) : Option ObservationType :=
  if h : output.wellFormed = true then some ⟨contract, output, h⟩ else none

theorem ObservationType.of_parts (o : ObservationType) :
    ObservationType.of o.contract o.output = some o := by
  cases o; simp [ObservationType.of, *]

theorem ObservationType.of_eq_some :
    ObservationType.of c s = some o ↔ o.contract = c ∧ o.output = s := by
  constructor
  · intro h; unfold ObservationType.of at h; split at h <;> cases h; exact ⟨rfl, rfl⟩
  · rintro ⟨rfl, rfl⟩; exact ObservationType.of_parts o

theorem observation_identity (a b : ObservationType) :
    a = b ↔ a.contract = b.contract ∧ a.output = b.output := by
  cases a; cases b; simp
```

An observation is typed by the contract it came back through (binary and
input) and its own output shape. The contract has one input; the many
observations of one invocation share it and differ in their shapes. Every
observation type goes through `ObservationType.of`, which refuses an output
shape that is not well formed, so rebuilding one from its parts always gives
it back (`ObservationType.of_parts`).

Blame runs both ways: an input that does not fit the contract is the
caller's, an output that does not fit is the agent's. What an output must fit
is not settled: shapes the contract declares, or the input it is passed to.

## Observe: the one way to an observation

```lean
/-- The shape of a context: an array of references to observations, as
`contract.Normal` writes it. Each member carries its own type. -/
def Schema.context : Schema := ⟨"{\"type\":\"array\",\"items\":{\"type\":\"string\"}}"⟩

/-- The shape of a goal: the text a parent hands its child. -/
def Schema.goal : Schema := ⟨"{\"type\":\"string\"}"⟩

/-- The shape of an exit status: why an agent gave up, in its own words. -/
def Schema.exitStatus : Schema := ⟨"{\"type\":\"string\"}"⟩

/-- The narrow interfaces through which `observe` is granted. -/
inductive Interface
  | effect       -- what came back from a leaf's one contract, or what it wrote
  | computation  -- a goal or a context, as a node delegates; a node's return
  | exit         -- an agent giving up, with a reason
  deriving DecidableEq

/-- An agent is a leaf or a node; `classify` decides which (see "Leaf or node"). -/
inductive Kind
  | leaf
  | node
  deriving DecidableEq

/-- The interfaces each kind observes through. -/
def Kind.interfaces : Kind → List Interface
  | .leaf => [.effect, .exit]
  | .node => [.computation, .exit]

/-- An observation's place in the records, assigned as it is recorded. -/
abbrev Stamp := Nat

/-- An observation: its stamp, the contract and kind of the agent that created
it, the interface it was created through, its type, and the observations it
was made from. Its constructor is private: `observe` is the only way to make
one. -/
inductive Observation
  | private mk (stamp : Stamp) (creator : Contract) (kind : Kind) (via : Interface)
      (type : ObservationType) (members : List Observation)

def Observation.stamp   : Observation → Stamp            | .mk s _ _ _ _ _  => s
def Observation.creator : Observation → Contract         | .mk _ c _ _ _ _  => c
def Observation.kind    : Observation → Kind             | .mk _ _ k _ _ _  => k
def Observation.via     : Observation → Interface        | .mk _ _ _ v _ _  => v
def Observation.type    : Observation → ObservationType  | .mk _ _ _ _ t _  => t
def Observation.members : Observation → List Observation | .mk _ _ _ _ _ ms => ms

mutual
def Observation.decEq : (a b : Observation) → Decidable (a = b)
  | .mk s₁ c₁ k₁ v₁ t₁ ms₁, .mk s₂ c₂ k₂ v₂ t₂ ms₂ =>
    if h : s₁ = s₂ ∧ c₁ = c₂ ∧ k₁ = k₂ ∧ v₁ = v₂ ∧ t₁ = t₂ then
      match Observation.decEqList ms₁ ms₂ with
      | isTrue hm  => isTrue (by obtain ⟨rfl, rfl, rfl, rfl, rfl⟩ := h; subst hm; rfl)
      | isFalse hm => isFalse (by intro e; cases e; exact hm rfl)
    else isFalse (by intro e; cases e; exact h ⟨rfl, rfl, rfl, rfl, rfl⟩)

def Observation.decEqList : (a b : List Observation) → Decidable (a = b)
  | [], [] => isTrue rfl
  | [], _ :: _ => isFalse (by simp)
  | _ :: _, [] => isFalse (by simp)
  | a :: as, b :: bs =>
    match Observation.decEq a b with
    | isTrue h =>
      match Observation.decEqList as bs with
      | isTrue h'  => isTrue (by subst h; subst h'; rfl)
      | isFalse h' => isFalse (by intro e; cases e; exact h' rfl)
    | isFalse h => isFalse (by intro e; cases e; exact h rfl)
end

instance : DecidableEq Observation := Observation.decEq

/-- What each interface admits, given `observe`'s own contract. -/
def admits (self : Contract) : Interface → ObservationType → List Observation → Bool
  | .effect,      t, ms => ms.isEmpty && t.contract != self
  | .exit,        t, ms => ms.isEmpty && t.contract == self && t.output == Schema.exitStatus
  | .computation, t, [] => t.contract == self && t.output == Schema.goal
  | .computation, t, ms => t.contract == self && t.output == Schema.context &&
                           decide (ms.map (·.stamp)).Nodup

/-- The one function that creates observations. `self` is `observe`'s own
contract; `caller` is the agent creating it, and `kind` decides the
interfaces it holds. `stamp` is the record it will be, after every member's. -/
def observe (self caller : Contract) (kind : Kind) (stamp : Stamp) (via : Interface)
    (type : ObservationType) (members : List Observation) : Option Observation :=
  if via ∈ kind.interfaces ∧ admits self via type members = true ∧
     members.all (·.stamp < stamp) = true
  then some (.mk stamp caller kind via type members) else none

theorem observe_eq_some :
    observe self caller kind stamp via type members = some o ↔
      via ∈ kind.interfaces ∧ admits self via type members = true ∧
      (∀ m ∈ members, m.stamp < stamp) ∧
      o.stamp = stamp ∧ o.creator = caller ∧ o.kind = kind ∧
      o.via = via ∧ o.type = type ∧ o.members = members := by
  unfold observe
  split
  · rename_i h
    obtain ⟨hg, ha, hs⟩ := h
    simp only [List.all_eq_true, decide_eq_true_eq] at hs
    constructor
    · intro e; cases e; exact ⟨hg, ha, hs, rfl, rfl, rfl, rfl, rfl, rfl⟩
    · rintro ⟨-, -, -, h₁, h₂, h₃, h₄, h₅, h₆⟩
      cases o
      simp_all [Observation.stamp, Observation.creator, Observation.kind,
        Observation.via, Observation.type, Observation.members]
  · rename_i h
    simp only [reduceCtorEq, false_iff]
    rintro ⟨hg, ha, hs, -⟩
    exact h ⟨hg, ha, by simpa [List.all_eq_true] using hs⟩
```

`observe` is the boundary. It is the only function that creates an
observation: `Observation`'s constructor is private, so nothing outside this
file can make one any other way, and every member of an observation was made
by `observe` before it. Validity throughout a lineage is by construction;
there is nothing else to check. Protecting it means protecting this file (see
"The Go shell"), and re-validating what comes back from outside it (see
"Records").

`observe` is never granted whole. It is granted through narrow interfaces,
each admitting one kind of observation:

- **`effect`**, granted to leaves. Typed by the contract the leaf invoked (a
  tool's, or for a leaf that writes, its own) and the shape of what came
  back; never by `observe`'s own contract. No members.
- **`computation`**, granted to nodes, through `delegate`. A goal (no
  members, the goal shape) or a context (members, in order, none twice, the
  context shape). Typed by `observe`'s own contract: it is what made them.
  A node's return is a computation too (see "Scope").
- **`exit`**, granted to both. An agent giving up, willingly, with a reason.
  An exit the system forces is a fault, never an exit status.

So effects come only from leaves, and computations only from nodes; an exit
from either. An observation records the interface it came through, so a goal
and an exit status, both strings, are never confused. `classify` decides an
agent's kind, and the kind decides its interfaces (`Kind.interfaces`); the
observation records both, with its creator, so it carries its own grant.
They are granted after the kind is known, so they never count toward the
reach that decides it.

Every observation has a stamp: its place in the records, assigned as it is
recorded, and later than every member's. Two observations of the same type
from the same creator, two sibling listings, are two stamps and so two
observations; Lean never compares their values, which it never holds. The
stamps are the records' to keep unique; a context's members are checked by
stamp, none twice.

An observation's type has two halves, typed apart here: the contract it came
back through (the input half) and its own output shape (the return half). At
runtime they are joined in its hash.

A context's members are a list, in the order passed: order is part of what a
child is shown, so `[a, b]` and `[b, a]` are different contexts. No
observation appears twice in one context. The members' types stay on the
members, so a context holds them without restating them.

## Inputs

```lean
/-- The empty input: an object with no fields. The caller sends `{}`. -/
def Schema.empty : Schema := ⟨"{\"type\":\"object\"}"⟩
```

A child's input is one context: its goal first, then the observations passed
to it, in order. A delegated child's context is never empty: it always holds
its goal. `{}` is the input of a contract that takes none. Nothing else is
assumed of a callee's declared shape.

## Scope

```lean
/-- Which way an observation is going. -/
inductive Dir
  | invoke
  | ret

/-- Everything passed is in the sender's scope. -/
def mayPass (sender passed : List Observation) : Bool :=
  passed.all (· ∈ sender)

theorem mayPass_iff : mayPass sender passed = true ↔ passed ⊆ sender := by
  simp [mayPass, List.all_eq_true, List.subset_def]

/-- The receiver's scope after the edge. A child starts with only what it is
passed; a parent keeps its scope and adds what is returned. -/
def receive : Dir → (receiver passed : List Observation) → List Observation
  | .invoke, _,     passed => passed
  | .ret,    scope, passed => scope ++ passed

/-- An observation and every observation it was made from. -/
def lineage : Observation → List Observation
  | .mk s c k v t ms => .mk s c k v t ms :: ms.attach.flatMap fun ⟨m, _⟩ => lineage m

/-- What a node's return may hold: nothing in the lineage of what it was
passed. -/
def nodeMayReturn (received returned : List Observation) : Bool :=
  returned.all fun o => (lineage o).all fun x => !decide (x ∈ received.flatMap lineage)

theorem nodeMayReturn_iff :
    nodeMayReturn received returned = true ↔
      ∀ o ∈ returned, ∀ x ∈ lineage o, ∀ r ∈ received, x ∉ lineage r := by
  simp [nodeMayReturn, List.all_eq_true]

/-- A contract cannot call itself to create observations: of what a node
returns, only an exit status may be its own. -/
def creatorMayReturn (self : Contract) (returned : List Observation) : Bool :=
  returned.all fun m => m.creator != self || m.via == .exit

theorem creatorMayReturn_iff :
    creatorMayReturn self returned = true ↔
      ∀ m ∈ returned, m.creator = self → m.via = .exit := by
  simp [creatorMayReturn, List.all_eq_true, Decidable.imp_iff_not_or]

/-- A node returns one observation: a computation of its own, with members,
its children's observations or its exit status, never content of its own. -/
def nodeReturn (self : Contract) (received : List Observation) (o : Observation) : Bool :=
  o.via == .computation && o.creator == self && !o.members.isEmpty &&
  nodeMayReturn received o.members && creatorMayReturn self o.members

theorem nodeReturn_iff :
    nodeReturn self received o = true ↔
      o.via = .computation ∧ o.creator = self ∧ o.members ≠ [] ∧
      (∀ m ∈ o.members, ∀ x ∈ lineage m, ∀ r ∈ received, x ∉ lineage r) ∧
      (∀ m ∈ o.members, m.creator = self → m.via = .exit) := by
  simp only [nodeReturn, Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true',
    List.isEmpty_eq_false_iff, nodeMayReturn_iff, creatorMayReturn_iff, and_assoc]

/-- The core supplies a child's actor, never its parent's binary: Claude
delegates only to Codex, and Codex only to Claude. The model sees `delegate`
and nothing of which actor runs. -/
def alternates (parent child : Contract) : Bool := parent.binary != child.binary

/-- So what a child creates is never its parent's own: the creator rule
refuses only what a node created itself. -/
theorem alternates_child_may_return (h : alternates parent child = true)
    (hm : ∀ m ∈ returned, m.creator = child) : creatorMayReturn parent returned = true := by
  rw [creatorMayReturn_iff]
  intro m hmr hc
  have := hm m hmr
  simp [alternates] at h
  exact absurd (by rw [← hc, this]) h
```

An agent may pass only what is in its scope: what it was passed, what its
children returned, and what it created through the interfaces it holds. The
direction says who the sender is: on invoke the parent, on return the child.
An observation is obtained only by endowment (the root's spec), creation
(through `observe`), or introduction (passed on invoke or return), the
object-capability rule. The runtime supplies the sender's scope as a value,
from the records; Lean checks it. On invoke what is checked is the context's
members: the goal and the context are new, created as they are passed.

A node returns one observation, and only its outer shape is the node's: a
computation it created through `observe`. Its members are its children's
observations, or its exit status:

```
node_n_observation { child_observation: <some_obs> }
node_n_observation { exit: "reason" }
```

Never content of its own. A node that returns `""` or `{}` has written an
observation itself, which only `observe` may do, and a computation with no
members is not a return.

Two rules hold of the members, each closing what the other leaves open:

- **A contract cannot call itself to create observations** (`creatorMayReturn`,
  the rule once called echo). Of what a node returns, only an exit status may
  be its own creation; a goal or a context it made for a child is never
  handed back up as a result.
- **No member may meet the lineage of what the node was passed**
  (`nodeMayReturn`): it would be found twice on one edge, once going down and
  once coming back. A goal the parent made is the parent's creation, not the
  node's, so the creator rule alone would let it back up; this rule does not.

Both are checked by instance: a member is its stamp, so two siblings of one
type never collide.

So a node that has not delegated has one thing it may return, its exit
status. Nothing is assumed about what a node should return beyond this.

The creator rule checks by contract, and it does not refuse a node's
children, because a child never runs its parent's binary (`alternates`). The
core supplies every child's actor: Claude delegates only to Codex, and Codex
only to Claude. The model sees only `delegate`; which actor runs the child is
never in its prompt, its tools or its payload. So what a child creates is
never its parent's own (`alternates_child_may_return`). A grandchild may
share its grandparent's contract, but it is never a direct member: a node's
return is checked with `mayPass` against its scope, like every other pass,
and its scope holds what its children returned, not what is inside them.

## Novelty

```lean
/-- An input at base type: the binary that received it, and the types of the
observations passed, in order. -/
structure InputBase where
  binary       : Binary
  observations : List ObservationType
  deriving DecidableEq

inductive InputBaseNovelty
  | seen
  | unseen

inductive ObservationBaseNovelty
  | seen
  | unseen

def inputBaseNovelty (history : List InputBase) (i : InputBase) : InputBaseNovelty :=
  if i ∈ history then .seen else .unseen

def observationBaseNovelty (history : List ObservationType) (o : ObservationType) :
    ObservationBaseNovelty :=
  if o ∈ history then .seen else .unseen
```

Novelty is by base type, never by value, and split by direction, as one input
gives many observations. An input is unseen when this binary has not been
handed this sequence of types before. An observation is unseen when this
contract (binary and input shape) has not returned this output shape before.
The goal is typed like any observation; its base type never varies, so it
does not change an input's novelty.

Novelty is apart from validity: validity refuses, novelty measures. The
runtime supplies the history; Lean answers.

## Guards

```lean
/-- The default guard, until the model writes one: the value last returned
(`returns.r0c0`), or the input (`input`), matched whole. `json` is the value
as `contract.Normal` writes it; JSON is a CEL literal. -/
def defaultGuard (ref json : String) : String := ref ++ " == " ++ json

/-- What the apprentice does after an observation comes back. -/
inductive Next
  | continue  -- take the next instruction; no model
  | fallback  -- an assumption broke; the model runs

/-- An unseen shape broke an assumption; a seen one continues only while its
guard holds. -/
def next : ObservationBaseNovelty → (guardHolds : Bool) → Next
  | .unseen, _     => .fallback
  | .seen,   true  => .continue
  | .seen,   false => .fallback

theorem next_continue_iff : next n g = .continue ↔ n = .seen ∧ g = true := by
  cases n <;> cases g <;> simp [next]
```

There are two paths, never a third:

| | Guard holds | Guard breaks |
|---|---|---|
| **Shape unseen** | | fall back to the model |
| **Shape seen** | continue: the apprentice takes the next instruction, with no model | fall back to the model, which writes or extends that shape's guard |

The system handles shape; guards handle assumptions. A guard is CEL, in the
apprentice's refinement environment (`apprentice/refine`: `input`,
`input_type`, `returns`, `types`; no `map`, no `filter`, no clock). Each
(input, output shape) has exactly one guard, so one input, which may return
many shapes, has as many guards as shapes it has returned.

Until the model writes one, a shape's guard is the default: a string match
on the value last returned (`defaultGuard`: the value as JSON, which is a CEL
literal, matched whole). It is overwritten when the same input returns the
same shape with a different value: the world changed under the procedure, and
the model, falling back, writes a guard that covers more.

The objective is coverage: for a given input, as many of its outputs as
possible fall inside a guard that holds, so the procedure runs
deterministically, with no model. A model's own writing rarely repeats, so it
mostly falls back; that is low coverage, and it is measured.

### Example

A run on `{"path": "~/Documents/A"}` was remembered as:

```
r0c0 = list_dir({path: input.path})        → ["notes.md", "q3.xlsx"]
r1c0 = categorize({files: returns.r0c0})   → {"docs": ["notes.md"], "sheets": ["q3.xlsx"]}
```

1. **The default guard**, before any model has written one:

   ```cel
   returns.r0c0 == ["notes.md", "q3.xlsx"]
   ```

   The next run with the same input and the same listing: the guard holds,
   `categorize` runs compiled, no model.

2. **The world changes.** `report.pdf` appears. The shape (an array of
   strings) is seen; the string match breaks; the apprentice falls back. The
   model writes a guard that covers more:

   ```cel
   returns.r0c0.all(f, f.endsWith(".md") || f.endsWith(".xlsx") || f.endsWith(".pdf"))
   ```

   Any listing of those kinds now runs compiled: one input, many outputs.

3. **Outside the assumption.** `clip.mov` appears. The guard breaks; the
   model extends it, or handles it another way.

4. **An unseen shape.** The directory becomes unreadable, and `list_dir`
   returns `{"error": "permission denied"}`. This input has never returned
   that shape, so it falls back with no guard involved. Handled, the shape
   gets its own guard:

   ```cel
   types.r0c0 == "list_dir → {error: string}" && returns.r0c0.error.startsWith("permission")
   ```

   One input, two shapes, a guard for each.

### The invoke path

```lean
/-- What the core does when a child is invoked. -/
inductive Enter
  | procedure  -- the entry guard holds: the remembered procedure runs; no model
  | model      -- an assumption broke; the model runs

def enter : InputBaseNovelty → (guardHolds : Bool) → Enter
  | .unseen, _     => .model
  | .seen,   true  => .procedure
  | .seen,   false => .model

theorem enter_procedure_iff : enter n g = .procedure ↔ n = .seen ∧ g = true := by
  cases n <;> cases g <;> simp [enter]
```

The invoke path mirrors the return path. An unseen input base type (this
binary has never been handed this sequence of types) runs the model. A seen
one runs its entry guard over the input's values: if it holds, the remembered
procedure runs in place of the model, for the whole subtree; if it breaks,
the model runs, and writes or extends the guard.

A return guard decides whether to continue inside a procedure; an entry guard
decides whether to enter one. Goals share one base type, so many procedures
may sit under one input base type ("categorize `~/A`" and "summarize `~/A`").
Which one runs is whatever the model encodes in the entry guard's CEL; the
system runs what is encoded and decides nothing. The default entry guard is a
string match on the whole input: reuse only for an identical input, until the
model writes one that covers more.

A procedure replays its calls under the original leaf grants, and creates
each observation through the interface it was first created through. So a
procedure in place of a subtree is still that subtree in the records:
lineage, scope and the circular check hold as they did for the model.

### Who executes

Guards decide; they never run anything. On both paths:

```
CEL guard         holds / breaks                       pure, no effects
Lean next/enter   continue / fallback / procedure      policy
apprentice        writes the plan                      meaning; cannot exec
skyra-spawn       checks the grant, execs              mechanism; the only exec
```

Execution is enforced by the OS, never by CEL, Lean or a Go package any code
could route around ([52](52-kernel-provisions-apprentice-spawns-proposal.md)):
the apprentice's seccomp filter denies `fork`, `pipe` and `exec`, and its one
authority is a socket to `skyra-spawn`, bound to its grant. `skyra-spawn`
alone execs pinned binaries, and only those the asking socket's grant allows.
A procedure's replayed call runs in the Landlock domain of the leaf grant it
was learned under, as the model's leaf would have.

**For now, on macOS:** 52 is Linux only and not built. The experiment runs
under today's kernel, which mediates every hop, with Seatbelt allowing exec
only for the program being run. That enforces part of the rule, not "only
`skyra-spawn` execs"; accepted until 52 exists.

## Leaf or node

```lean
def kind (reach : List Contract) : Kind :=
  if reach.length ≤ 1 then .leaf else .node

/-- What the core runs: the agent's contract and the contracts in its reach,
as values. `none` refuses an agent whose own binary is in its reach. -/
def classify (self : Contract) (reach : List Contract) : Option Kind :=
  if self.binary ∈ reach.map (·.binary) then none else some (kind reach.eraseDups)

theorem classify_none_iff :
    classify self reach = none ↔ self.binary ∈ reach.map (·.binary) := by
  unfold classify; split <;> simp_all

theorem classify_leaf_iff :
    classify self reach = some .leaf ↔
      self.binary ∉ reach.map (·.binary) ∧ reach.eraseDups.length ≤ 1 := by
  unfold classify kind; split <;> (try split) <;> simp_all

theorem classify_node_iff :
    classify self reach = some .node ↔
      self.binary ∉ reach.map (·.binary) ∧ 2 ≤ reach.eraseDups.length := by
  unfold classify kind; split <;> (try split) <;> simp_all <;> omega
```

The core passes an agent's contract and its reach as values, and Lean
decides: refused, a leaf, or a node. No contract of an agent's own binary is
in its reach, and duplicates count once. The theorems say what a leaf is and
what it is not: one contract in reach or none, never its own binary.

Two identities, each the stricter for its question. The self check compares
binaries, so an agent cannot reach itself through another of its shapes. The
count compares contracts, so two shapes of one binary are two contracts, and
make a node. Either other choice would open a hole.

Two things this invariant does not cover, on purpose:

- **A contract may be wide.** The count is per contract, and nothing bounds
  what one contract covers: a shape of `{"type":"any"}`, or one contract over
  many subcommands, is one contract and so a leaf. The rule is only as narrow
  as the contracts registered. A concession for now; the experiment registers
  narrow ones.
- **`self` must be honest.** The theorems hold for the `self` they are given.
  It is the actor being provisioned, supplied by the kernel at pass time,
  never by a parent's payload. The Go enforces this.

## Pass time and provision time

**Pass time** is the parent's act: it chooses the contracts and the
observations a child gets. Checked, in order:

0. **The actor.** The core picks the child's actor, the other binary from
   the parent's (`alternates`). The model never names one.
1. **The Boundary.** Every contract passed is held by the parent
   (`passable`).
2. **The observations.** Every one passed is in the parent's scope
   (`mayPass`); the goal and the context are created through `computation`.
3. **The kind.** `classify` runs on the child's contract and the contracts
   passed. Only now is it known to be a leaf or a node.

On return, a node's one observation is checked by `nodeReturn`, and its
members by `mayPass` against the node's scope.

```lean
/-- Pass time: every contract passed is held by the parent. -/
def passable (held passed : List Contract) : Bool :=
  passed.all (· ∈ held)

theorem passable_iff : passable held passed = true ↔ passed ⊆ held := by
  simp [passable, List.all_eq_true, List.subset_def]
```

A child's actor (the model that runs it) is supplied by the core through
`delegate`, never drawn from the parent's reach, and never the parent's own
binary. So reach need not shrink
from parent to child, and a node may grant its whole reach to one child: its
actor's binary is not among what it is passed. A run may loop; the depth
bound stops it.

**Provision time** is the kernel's act: it grants the child what its kind
allows. A contract is held with one of two rights: to invoke it, or only to
pass it on.

```lean
/-- What an agent is provisioned with, from its kind. -/
structure Offer where
  invoke     : List Contract   -- contracts it may call
  passOnly   : List Contract   -- contracts it may only pass on
  delegate   : Bool
  interfaces : List Interface  -- how it may observe

def offered : Kind → (passed : List Contract) → Offer
  | .leaf, passed => ⟨passed, [], false, Kind.leaf.interfaces⟩
  | .node, passed => ⟨[], passed, true, Kind.node.interfaces⟩

theorem offered_interfaces : (offered k p).interfaces = k.interfaces := by
  cases k <;> rfl

theorem offered_effect_iff : .effect ∈ (offered k p).interfaces ↔ k = .leaf := by
  cases k <;> simp [offered, Kind.interfaces]

theorem offered_computation_iff : .computation ∈ (offered k p).interfaces ↔ k = .node := by
  cases k <;> simp [offered, Kind.interfaces]

theorem offered_exit : .exit ∈ (offered k p).interfaces := by
  cases k <;> simp [offered, Kind.interfaces]

theorem offered_delegate_iff : (offered k p).delegate = true ↔ k = .node := by
  cases k <;> simp [offered]

theorem offered_node_invokes_nothing : (offered .node p).invoke = [] := rfl
```

- **A leaf** may invoke the contract it was passed, if any, and observes
  through `effect` and `exit`. It does not hold `delegate`.
- **A node** may only pass on the contracts it was passed, holds `delegate`,
  and observes through `computation` and `exit`. It invokes nothing else.

Pass-only is the right a pure-control node needs; capability systems keep
passing on apart from using (seL4's Grant right, SPKI's propagate flag).

## Records: re-validated when rebuilt

```lean
/-- A contract as Go keeps it: its binary and its shape, as text. -/
structure ContractRecord where
  binary : String
  shape  : String
  deriving DecidableEq

def Contract.record (c : Contract) : ContractRecord := ⟨c.binary.val, c.shape.val⟩

def ContractRecord.identify (r : ContractRecord) : Option Contract :=
  _root_.identify ⟨r.binary⟩ ⟨r.shape⟩

theorem ContractRecord.identify_record (r : ContractRecord) (h : r.identify = some c) :
    c.record = r := by
  obtain ⟨hb, hs⟩ := identify_eq_some.mp h
  simp [Contract.record, hb, hs]

/-- An observation as Go keeps it: plain values, its members by their
records. Anything may build one; only `rebuild` makes an observation of it. -/
inductive Record
  | mk (stamp : Stamp) (creator : ContractRecord) (kind : Kind) (via : Interface)
      (contract : ContractRecord) (output : String) (members : List Record)

mutual
def Observation.record : Observation → Record
  | .mk s c k v t ms => .mk s c.record k v t.contract.record t.output.val (Observation.records ms)

def Observation.records : List Observation → List Record
  | []      => []
  | o :: os => o.record :: Observation.records os
end

mutual
/-- Re-validate a record: its contracts identified, its type made by
`ObservationType.of`, its members rebuilt first, and the observation made by
`observe`, as if for the first time. -/
def Record.rebuild (self : Contract) : Record → Option Observation
  | .mk stamp creator kind via contract output members => do
    let caller ← creator.identify
    let c ← contract.identify
    let type ← ObservationType.of c ⟨output⟩
    let ms ← Record.rebuildAll self members
    observe self caller kind stamp via type ms

def Record.rebuildAll (self : Contract) : List Record → Option (List Observation)
  | []      => some []
  | r :: rs => do
    let o ← r.rebuild self
    let os ← Record.rebuildAll self rs
    pure (o :: os)
end

mutual
/-- What a rebuild gives back is the record it was given. -/
theorem Record.rebuild_record : ∀ (r : Record) (o : Observation),
    r.rebuild self = some o → o.record = r
  | .mk stamp creator kind via contract output members, o, h => by
    simp only [Record.rebuild, Option.bind_eq_bind, Option.bind_eq_some_iff] at h
    obtain ⟨caller, hcaller, c, hc, type, htype, ms, hms, ho⟩ := h
    obtain ⟨-, -, -, hs, hcr, hk, hv, ht, hm⟩ := observe_eq_some.mp ho
    obtain ⟨htc, hto⟩ := ObservationType.of_eq_some.mp htype
    have hrs := Record.rebuildAll_records members ms hms
    cases o
    simp only [Observation.stamp, Observation.creator, Observation.kind, Observation.via,
      Observation.type, Observation.members] at hs hcr hk hv ht hm
    subst hs hcr hk hv ht hm
    simp [Observation.record, hrs, htc, hto, ContractRecord.identify_record _ hcaller,
      ContractRecord.identify_record _ hc]

theorem Record.rebuildAll_records : ∀ (rs : List Record) (os : List Observation),
    Record.rebuildAll self rs = some os → Observation.records os = rs
  | [], os, h => by simp [Record.rebuildAll] at h; subst h; rfl
  | r :: rs, os, h => by
    simp only [Record.rebuildAll, Option.bind_eq_bind, Option.bind_eq_some_iff,
      Option.pure_def, Option.some.injEq] at h
    obtain ⟨o, ho, os', hos, rfl⟩ := h
    simp [Observation.records, Record.rebuild_record r o ho, Record.rebuildAll_records rs os' hos]
end
```

`Observation`'s constructor is private, which protects observations only
while they are inside Lean. Go keeps them as records, and hands them back for
`mayPass`, `nodeReturn` and the next `observe`. A record is plain values and
anything may build one, so a record is never trusted: `rebuild` makes it an
observation again by the same path it was first made, identifying every
contract, making its type with `ObservationType.of`, rebuilding every member
first, and calling `observe`. A forged record has no observation, and what a
rebuild gives back is exactly the record it was given (`rebuild_record`).

So the exported functions take records, never observations: every one
rebuilds what it is handed before it checks anything. Records are re-checked
on every use, never cached as checked.

What `rebuild` cannot know is whether the record is the one the records hold.
A stamp the records never assigned, or a creator and kind other than what the
kernel provisioned, rebuild as long as they are consistent. Those facts are
the runtime's: Go reads records from `.skyra/` and never from a payload.

## The Go shell

Lean is the functional core: policy. Go is the imperative shell: mechanism.
Go computes facts, asks Lean, and carries out the answer; it decides nothing.

| Go does | With Lean's |
|---|---|
| Computes each binary's hash (cdhash or SHA-256) and canonical shapes | `identify`, `ObservationType.of` |
| Supplies `self`: the actor being provisioned, from the kernel, never from a payload | `classify` |
| Picks the child's actor: the other of Claude and Codex. The model is never shown it | `alternates` |
| Reads the parent's reach at pass time | `passable` |
| Reads each agent's scope from the records, and what a node was passed | `mayPass`, `receive`, `nodeReturn`, `creatorMayReturn` |
| Creates every observation by calling `observe`, with the caller's contract and kind and the next record stamp, and records it | `observe` |
| Hands back records, never observations; each is rebuilt before any check | `Record.rebuild` |
| Grants what the kind allows: invoke or pass-only rights, `delegate`, interfaces. Faults an invoke of a pass-only contract at intake | `offered` |
| Builds the prompt and the tools offered from the `Offer` | `offered` |
| Reads the history of inputs and observations from the records | `inputBaseNovelty`, `observationBaseNovelty` |
| Evaluates a shape's guard in CEL (`apprentice/refine`), writing the default guard until the model writes one | `defaultGuard`, `next` |
| At invoke, evaluates the entry guard, and runs the remembered procedure in place of the model, its calls under the original leaf grants | `enter` |
| Runs the model; stops a run at the depth bound. A forced stop is a fault, never an exit status | |

`observe` is the boundary, so the file that defines it is protected: its
hash is pinned in the Go build, CI runs `#print axioms` on every theorem and
accepts only `propext`, `Classical.choice` and `Quot.sound`, Go calls only
functions that take plain values, and the Go that calls them is protected
with it.

Left free, to observe: children may share a contract, a node may grant its
whole reach to one child, and a run that loops stops at the depth bound.

The rule for building the shell: Go that decides something Lean decides is
deleted, not kept beside it. Known so far:

- `delegate`'s `actor` and `model` fields, and its "actor in reach" check
  (`controller/apprentice/delegate.go`). The core picks the actor
  (`alternates`); the model's `delegate` takes a goal and observations only.
- The base-type branches of `refusedInput` and `refusedEdge`
  (`controller/apprentice/main.go`): an input or return of a type the
  procedure was not compiled for is `enter` or `next` on an unseen type. Go
  keeps only the CEL evaluation, holds or breaks, and hands Lean the fact.

## Known gaps

What this invariant does not cover, written down so it is not mistaken for
covered:

- **Endowment has no case in Lean.** The root's spec is where its first
  observations come from, and nothing in `observe` admits them: no interface
  is the root's. Until one is added, the root's endowment is created by Go,
  outside the boundary.
- **A leaf that paraphrases its input breaks lineage.** `nodeMayReturn`
  follows lineage, and a leaf that copies what it was passed into a new
  effect creates an observation with no members. The copy is new; the rule
  cannot see it.
- **Contracts may be wide** (see "Leaf or node"): a leaf is as narrow as its
  one contract.
- **Records are rebuilt, not authenticated.** `rebuild` refuses a record no
  `observe` would accept, not one the records never held (see "Records").
- **52 is not built on macOS.** Only the kernel's mediation and Seatbelt
  enforce who execs (see "Who executes").

## Plan

Base types are contracts and binaries only: `ObservationType` stays
(contract, output shape). Value is CEL's, configured per shape, on both
paths. For the first runs, guards use the default implementation and
`remember` is off, so no procedure exists and every path falls back to the
model; guards still run, and whether each would have held is recorded.

1. **Finish this proposal.** Done (2026-11-03):
   - The default guard, `defaultGuard`: the value as JSON, a CEL literal,
     matched whole, for return and entry guards.
   - Stamped observations: two sibling listings are two stamps.
   - The creator rule, `creatorMayReturn`, beside `nodeMayReturn`: neither
     covers the other. It checks by contract, which holds because Claude and
     Codex alternate (`alternates`), hidden from the model.
   - Records re-validated when rebuilt (`Record.rebuild`, `rebuild_record`).
   - Known gaps, written down.
2. **The Lean core as a package.** A Lake project pinned to Lean 4.35; its
   files become the source of truth, and this proposal points to them.
   Linked into the apprentice through cgo, not executed: under today's macOS
   sandbox the apprentice may exec only the program being run. Exported
   functions take plain values only. CI pins the source hash in the Go build
   and runs `#print axioms` on every theorem, accepting only `propext`,
   `Classical.choice` and `Quot.sound`.
3. **The Go shell** (apprentice, `proceduretools`, `claude/resident`).
   - The tools offered follow the kind: a node gets `delegate` and `exit`; a
     leaf its one contract, `effect` and `exit`.
   - `delegate` runs `alternates`, `passable`, `classify` and `offered`;
     goals and contexts are created through `observe`; it takes
     `observations: [names]`, with a name table per node (`obs-1`, …)
     mapping names to stamps. It takes no actor.
   - A leaf's result is created through `observe` (`effect`); a node's
     return is checked by `nodeReturn`, and its members by `mayPass`.
   - The Go that Lean replaces is deleted (see "The Go shell").
   - Guards are evaluated with `apprentice/refine` (default string match);
     holds or breaks is recorded; with `remember` off, always fall back.
   - Both novelty measures are recorded; the depth bound is kept.
4. **The treemap harness.** Choose the slice (count candidate directories
   first); the spec sections, written by the user; narrow tools (`list_dir`,
   `stat`) with tight shapes; a `find` manifest taken before the run for the
   coverage check; a renderer that draws the tree from `.skyra/` records.
5. **First run.** Real Claude or Codex, the small slice, several runs. Read
   the trees.

Stop for review after steps 1, 2 and 3.

Open: link Lean into the apprentice through cgo (recommended) or register it
as its own target; and where the Lake project lives
(`controller/apprentice/policy/` or a top-level `lean/`).
