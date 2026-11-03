# 55 — One tool trees: the invariant

Date: 2026-11-03

Status: **draft.** The functional core of [54](54-one-tool-per-model-experiment-proposal.md),
likely a binary within the apprentice. Typechecks in core Lean 4.35.

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
with its own shape. Its constructor is private: every one is made by `observe`,
and carries the proof it passed. -/
structure ObservationType where
  private mk ::
  contract : Contract
  output   : Schema
  wf       : output.wellFormed = true
  deriving DecidableEq

/-- The one way to an observation type. Refuses an output shape that is not
well formed. -/
def observe (contract : Contract) (output : Schema) : Option ObservationType :=
  if h : output.wellFormed = true then some ⟨contract, output, h⟩ else none

theorem observe_parts (o : ObservationType) : observe o.contract o.output = some o := by
  cases o; simp [observe, *]

theorem observe_eq_some : observe c s = some o ↔ o.contract = c ∧ o.output = s := by
  constructor
  · intro h; unfold observe at h; split at h <;> cases h; exact ⟨rfl, rfl⟩
  · rintro ⟨rfl, rfl⟩; exact observe_parts o

theorem observation_identity (a b : ObservationType) :
    a = b ↔ a.contract = b.contract ∧ a.output = b.output := by
  cases a; cases b; simp
```

An observation is typed by the contract it came back through (binary and
input) and its own output shape. The contract has one input; the many
observations of one invocation share it and differ in their shapes. Every
observation type goes through `observe`, which refuses an output shape that
is not well formed,
so rebuilding one from its parts always gives it back (`observe_parts`).

Blame runs both ways: an input that does not fit the contract is the
caller's, an output that does not fit is the agent's. What an output must fit
is not settled: shapes the contract declares, or the input it is passed to.

## Observations

```lean
/-- The shape of a context: an array of references to observations, as
`contract.Normal` writes it. Each member carries its own type. -/
def Schema.context : Schema := ⟨"{\"type\":\"array\",\"items\":{\"type\":\"string\"}}"⟩

/-- The shape of a goal: the text a parent hands its child. -/
def Schema.goal : Schema := ⟨"{\"type\":\"string\"}"⟩

/-- An observation: its type, and the observations it was made from. A
leaf's has no members; a context's members are the observations passed in
it, in order. -/
inductive Observation
  | mk (type : ObservationType) (members : List Observation)

mutual
def Observation.decEq : (a b : Observation) → Decidable (a = b)
  | .mk t₁ ms₁, .mk t₂ ms₂ =>
    if ht : t₁ = t₂ then
      match Observation.decEqList ms₁ ms₂ with
      | isTrue hm  => isTrue (by subst ht; subst hm; rfl)
      | isFalse hm => isFalse (by intro h; cases h; exact hm rfl)
    else isFalse (by intro h; cases h; exact ht rfl)

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

/-- Valid observations, given the contract of `create_observation`. -/
inductive Valid (create : Contract) : Observation → Prop
  /-- What a leaf creates: typed by the contract it came back through, never
  `create_observation`'s, and its own output shape. -/
  | leaf : type.contract ≠ create →
           Valid create (.mk type [])
  /-- A goal: registered by `create_observation` when a parent delegates. -/
  | goal : type.contract = create →
           type.output = Schema.goal →
           Valid create (.mk type [])
  /-- A context: typed by `create_observation`, which registered it, and the
  context shape. Never empty, no member twice, every member valid. -/
  | context : type.contract = create →
              type.output = Schema.context →
              members ≠ [] →
              members.Nodup →
              (∀ m ∈ members, Valid create m) →
              Valid create (.mk type members)

/-- What the core runs. -/
def valid (create : Contract) : Observation → Bool
  | .mk t [] => t.contract != create || t.output == Schema.goal
  | .mk t ms => t.contract == create && t.output == Schema.context &&
                decide ms.Nodup && ms.attach.all fun ⟨m, _⟩ => valid create m

theorem valid_iff : ∀ o : Observation, valid create o = true ↔ Valid create o
  | .mk t [] => by
    simp only [valid, Bool.or_eq_true, bne_iff_ne, beq_iff_eq]
    constructor
    · rintro (h | h)
      · exact .leaf h
      · by_cases hc : t.contract = create
        · exact .goal hc h
        · exact .leaf hc
    · intro h; cases h with
      | leaf h => exact .inl h
      | goal _ h => exact .inr h
      | context _ _ hne _ _ => exact absurd rfl hne
  | .mk t (m :: ms) => by
    have ih : ∀ x ∈ m :: ms, valid create x = true ↔ Valid create x :=
      fun x _ => valid_iff x
    simp only [valid, Bool.and_eq_true, beq_iff_eq, decide_eq_true_eq, List.all_eq_true]
    constructor
    · rintro ⟨⟨⟨hc, ho⟩, hn⟩, hall⟩
      exact .context hc ho (by simp) hn fun x hx => (ih x hx).1 (hall ⟨x, hx⟩ (List.mem_attach _ _))
    · intro h; cases h with
      | context hc ho _ hn hall =>
        exact ⟨⟨⟨hc, ho⟩, hn⟩, fun ⟨x, hx⟩ _ => (ih x hx).2 (hall x hx)⟩
termination_by o => sizeOf o
```

An observation's type has two halves, typed apart here: the contract it came
back through (the input half) and its own output shape (the return half). At
runtime they are joined in its hash.

What a leaf creates is typed by the contract the leaf invoked (a tool's, or
for a leaf that writes, its own), and the shape of what came back. Never
`create_observation`'s: an observation typed by it is always a context, so an
empty context cannot pass as a leaf's observation. A context is typed by `create_observation`'s contract,
since that is what registered it, and by `Schema.context`. Not the child's:
it has not run yet. Not the parent's: a node creates nothing. The members'
types stay on the members, so a context holds them without restating them.

`Valid` is the rule; `valid` is what runs, and `valid_iff` proves they agree.
The core calls `valid`, so the rule is checked by Lean, never re-implemented
in Go.

`create_observation` is a system call, with its own binary and contract. It
is a leaf, so the creation invariant holds for it. It is granted at
provisioning, after the kind is known (see "Pass time and provision time"),
so it never counts toward the reach that decides the kind.

A leaf creates an observation through it. A goal or a context passed to a
child is not created by the node that passes it: the system registers each
through `create_observation`, the context only when every observation in it
is valid. The goal's text is still the node's; registering it makes it an
observation, with a type and a place in lineage, rather than an unrecorded
instruction. A node
still creates nothing. A context is never empty: an empty context is not
passed (see "Inputs"). Its members are a list, in the order passed: order is
part of what a child is shown, so `[a, b]` and `[b, a]` are different
contexts. The same members in the same order are one observation however
often they are passed. No observation appears twice in one context.

## Inputs

```lean
/-- The empty input: an object with no fields. The caller sends `{}`. -/
def Schema.empty : Schema := ⟨"{\"type\":\"object\"}"⟩
```

A child's input is one context: its goal first, then the observations passed
to it, in order. The goal is an observation like the rest, so a delegated
child's context is never empty. `{}` is the input of a contract that takes
none. Nothing else is assumed of a callee's declared shape.

## Scope

```lean
/-- Which way an observation is going. -/
inductive Dir
  | invoke
  | ret

/-- Everything passed is in the sender's scope, and valid. -/
def mayPass (create : Contract) (sender passed : List Observation) : Bool :=
  passed.all fun o => decide (o ∈ sender) && valid create o

theorem mayPass_iff :
    mayPass create sender passed = true ↔ ∀ o ∈ passed, o ∈ sender ∧ Valid create o := by
  simp [mayPass, List.all_eq_true, valid_iff]

/-- The receiver's scope after the edge. A child starts with only what it is
passed; a parent keeps its scope and adds what is returned. -/
def receive : Dir → (receiver passed : List Observation) → List Observation
  | .invoke, _,     passed => passed
  | .ret,    scope, passed => scope ++ passed
```

An agent may pass only what is in its scope: what it was passed, what its
children returned, and, for a leaf, what it created. The direction says who
the sender is: on invoke the parent, on return the child. An observation is
obtained only by endowment (the root's spec), creation (a leaf, through
`create_observation`), or introduction (passed on invoke or return), the
object-capability rule. The runtime supplies the sender's scope as a value,
from the records; Lean checks it. On invoke what is checked is the context's
members: the goal and the context are new, registered as they are passed.

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

## Leaf or node

```lean
inductive Kind
  | leaf
  | node
  deriving DecidableEq

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

1. **The Boundary.** Every contract passed is held by the parent (`passable`);
   the kernel faults a narrowing that asks for anything else.
2. **The observations.** Every one passed is valid; the system registers the
   context through `create_observation`.
3. **The kind.** `classify` runs on the child's contract and the contracts
   passed. Only now is it known to be a leaf or a node.

```lean
/-- Pass time: every contract passed is held by the parent. -/
def passable (held passed : List Contract) : Bool :=
  passed.all (· ∈ held)

theorem passable_iff : passable held passed = true ↔ passed ⊆ held := by
  simp [passable, List.all_eq_true, List.subset_def]
```

A child's actor (the model that runs it) is supplied by the core through
`delegate`, never drawn from the parent's reach. So reach need not shrink
from parent to child, and a node may grant its whole reach to one child: its
actor's binary is not among what it is passed. A run may loop; the depth
bound stops it.

**Provision time** is the kernel's act: it populates the child's reach from
its kind. A contract is held with one of two rights: to invoke it, or only to
pass it on.

- **A leaf** holds the contract it was passed, if any, to invoke, and
  `create_observation`.
- **A node** holds the contracts it was passed only to pass on, and
  `delegate`. It never holds `create_observation`, and cannot invoke what it
  holds.

Pass-only is the right a pure-control node needs; capability systems keep
passing on apart from using (seL4's Grant right, SPKI's propagate flag). For
the experiment the runtime enforces it, by offering a node only `delegate`.
It moves into the kernel, as a boundary fault at intake, once the trees show
nodes trying to invoke what they hold.

## Enforced by the runtime

Not part of the invariant:

- **What each kind is given:** its prompt, the tools it is offered, how it
  may return.
- **A node has at least one child.** The runtime refuses a node's return
  until it has delegated.

Left free, to observe: children may share a contract, a node may grant its
whole reach to one child, and a run that loops stops at the depth bound.
