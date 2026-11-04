# 55 — One tool trees: the invariant

Date: 2026-11-03

Status: **draft.** The functional core of [54](54-one-tool-per-model-experiment-proposal.md):
specified and proven in Lean, run as a Go package that agrees with it on
every vector Lean writes. Typechecks in core Lean 4.35.

The Lean is [`controller/apprentice/invariants/Invariants/Core.lean`](../../../controller/apprentice/invariants/Invariants/Core.lean),
the source of truth. The blocks below are a copy of it, in order;
`check-proposal.sh` there fails if they differ. A change goes to `Core.lean`
first.

## The invariants

Every invariant the system holds, with the theorem that proves it, or what
it rests on when Lean does not. The sections below are the proofs.

**Identity**

1. A contract is its binary and its input shape, nothing else; every one is
   made by `identify` (`contract_identity`, `identify_parts`).
2. An observation type is its contract and its own output shape; every one is
   made by `ObservationType.of` (`observation_identity`).
3. Values never enter a type. Value is CEL's.

**Observations**

4. `observe` is the only way to an observation from an instruction, and
   `rootObservation` the root's endowment. `Observation`'s constructor is
   private, so validity throughout a lineage is by construction
   (`observe_eq_some`). Across the Go package boundary it rests on unexported
   fields.
5. An effect has no members; a computation has members. Every observation is
   exactly one (`Observation.effect_iff`, `Observation.computation_iff`,
   `Observation.effect_xor_computation`).
6. The caller never says what it sends: `observe` decides by what it has, and
   admits by the kind of node sending it. An exit comes from every kind; a
   goal from a delegate; what an effect node writes, through
   `effect_observation`, from an effect node; what came back through a
   contract outside the apprentice, from an invoke node.
7. Only a delegate makes a computation, through `delegate`, no member twice
   (`observe_computation`); an effect or invoke node makes only effects
   (`observe_effect_or_invoke`). Nothing becomes an effect through `observe`,
   `root`, `return`, `invoke` or `delegate` (`observe_not_through`). An exit
   the system forces is a fault.

**Kind and authority**

8. The apprentice is one binary and its contracts: `observe`, `root`, `exit`,
   `return`, `effect_observation`, `invoke`, `delegate`, `goal`; each of that
   binary, none twice (`Apprentice.contracts_mine`,
   `Apprentice.contracts_nodup`).
9. The apprentice's contracts never count, for the self check or the kind
   (`classify_apprentice_before`, `classify_apprentice_after`).
10. No agent reaches its own binary, at every count (`classify_none_iff`).
11. No contract in reach is an effect node, one an invoke node, two or more a
    delegate (`classify_effect_iff`, `classify_invoke_iff`,
    `classify_delegate_iff`). Decided in the `observe` phase.
12. The kind decides what is loaded: `exit` and `return` for every kind;
    `effect_observation` for an effect node; `invoke` for an invoke node;
    `delegate` and `goal` for a delegate. `observe` and `root` are never
    loaded, and loading never changes the kind (`loaded_mine`,
    `loaded_never_privileged`, `classify_loaded`).
13. Only an invoke node invokes a contract outside the apprentice; a delegate
    only passes on (`offered_invoke`, `offered_delegate_invokes_nothing`,
    `offered_effect_invokes_nothing`, `offered_apprentice`).
14. A parent passes only contracts it holds (`passable_iff`).
15. `self` comes from the kernel, never a payload. The Go enforces it.

**Scope**

16. An agent passes only what is in its scope: what it was passed, what its
    children returned, what it created (`mayPass_iff`). An observation is
    obtained only by endowment, creation or introduction.
17. A child starts with only what it is passed; a parent keeps its scope and
    adds what is returned (`receive`).
18. A return holds something. An invoke or delegate node returns nothing its
    model created directly; an effect node may return what it wrote
    (`mayReturn_iff`). Passing is not creating. Nothing returned meets the
    lineage of what the node was passed (`nodeMayReturn_iff`).
19. Refusal is by refinement type: two observations of one refinement type
    are the same. Uniqueness is the system's interpretation; there is no
    stamp. Until Lean takes refinement ids, it compares `ObservationType`.

**Base types, refinement types and guards**

20. A base type is the binary, the base types passed, in order, and the
    shape back, hashed; not the parent, the path, the depth, the goal, or any
    guard (`base_eq_iff`). Refinement types are never part of a base type.
21. A refinement type is a base type and its guard, hashed.
22. Two levels, kept apart: the base types the system recognizes (`known`),
    and a guard slot for each (`guards`). Every known base type has exactly
    one guard, and only a known one has any (`Store.WellFormed`,
    `refinement_known`); so one refinement type each (`one_refinement`).
23. A guard belongs to one base type and cannot move (`refinement_base`;
    `Guard b` is indexed by `b`).
24. Unseen is a hash the system does not recognize: no slot at all
    (`see_unseen_iff`). On first sight the base type becomes known and its
    slot is made with the default guard; no other slot is touched
    (`see_known`, `see_other`, `see_wellFormed`).
25. A rewrite overwrites the one slot of a known base type; an unknown one
    has no slot (`rewrite`, `rewrite_wellFormed`).
26. `known` only grows: seen stays seen (`see_grows`, `rewrite_known`).
27. The one assumption is no hash collisions, written once (`Hashes.Sound`)
    and taken as a hypothesis. Lean never computes a hash.

**Control**

28. Two paths, never a third: a return continues without the model only when
    its base type is seen and its guard holds (`next_continue_iff`); an
    invoke runs a procedure only when its input base type is seen and its
    entry guard holds (`enter_procedure_iff`). Anything else falls back.
29. Novelty measures, never refuses; validity refuses.
30. A replayed procedure is still its subtree in the records: the original
    invoke grants, each observation through the kind it was first made by.
    Not yet proven.
31. Guards decide and never execute; only `skyra-spawn` execs, enforced by
    the OS. On macOS, in part, until 52 is built.

Not settled: the default guard's value, which `Guard.default` does not yet
carry.

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
of a contract's identity assumes anything about collisions (base and
refinement types do, and say so: see "Base types and refinement types"). One binary may have several shapes, each
its own contract. A return's shape is `create_return`'s (see [56](56-tree-machine-proposal.md)).

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
/-- The apprentice: one binary, and its contracts, each of that binary and
none twice. Some are privileged (`observe`, `root`): never loaded into a
reach. Its constructor is private: every one is made by `Apprentice.of`. -/
structure Apprentice where
  private mk ::
  binary            : Binary
  observe           : Contract
  root              : Contract
  exit              : Contract
  ret               : Contract  -- `return`
  effectObservation : Contract
  invoke            : Contract
  delegate          : Contract
  goal              : Contract
  wf : ([observe, root, exit, ret, effectObservation, invoke, delegate, goal].all
          (·.binary == binary) && decide
          [observe, root, exit, ret, effectObservation, invoke, delegate, goal].Nodup) = true

def Apprentice.contracts (a : Apprentice) : List Contract :=
  [a.observe, a.root, a.exit, a.ret, a.effectObservation, a.invoke, a.delegate, a.goal]

/-- The one way to an apprentice. Refuses a contract of another binary, or
one contract in two places. -/
def Apprentice.of (binary : Binary) (observe root exit ret effectObservation invoke delegate goal : Contract) :
    Option Apprentice :=
  if h : ([observe, root, exit, ret, effectObservation, invoke, delegate, goal].all
            (·.binary == binary) && decide
            [observe, root, exit, ret, effectObservation, invoke, delegate, goal].Nodup) = true
  then some ⟨binary, observe, root, exit, ret, effectObservation, invoke, delegate, goal, h⟩
  else none

/-- A contract of the apprentice's binary. -/
def Apprentice.mine (a : Apprentice) (c : Contract) : Bool := c.binary == a.binary

theorem Apprentice.contracts_mine (a : Apprentice) (h : c ∈ a.contracts) : a.mine c = true := by
  have hw := a.wf
  simp only [Bool.and_eq_true, List.all_eq_true, beq_iff_eq] at hw
  simp only [Apprentice.mine, beq_iff_eq]
  exact hw.1 c h

theorem Apprentice.contracts_nodup (a : Apprentice) : a.contracts.Nodup := by
  have hw := a.wf
  simp only [Bool.and_eq_true, decide_eq_true_eq] at hw
  exact hw.2

/-- What a node is: decided from its reach (see "Effect, invoke or delegate"). -/
inductive Kind
  | effect    -- no contract in reach: it writes
  | invoke    -- one: it invokes it
  | delegate  -- two or more: it delegates
  deriving DecidableEq

/-- An observation: its type, and the observations it was made from. Its
constructor is private: `observe` and `rootObservation` are the only ways to
make one. -/
inductive Observation
  | private mk (type : ObservationType) (members : List Observation)

def Observation.type    : Observation → ObservationType  | .mk t _  => t
def Observation.members : Observation → List Observation | .mk _ ms => ms

/-- An effect has no members; a computation has members. -/
def Observation.isEffect (o : Observation) : Bool := o.members.isEmpty
def Observation.isComputation (o : Observation) : Bool := !o.members.isEmpty

theorem Observation.effect_iff {o : Observation} : o.isEffect = true ↔ o.members = [] := by
  simp [Observation.isEffect]

theorem Observation.computation_iff {o : Observation} : o.isComputation = true ↔ o.members ≠ [] := by
  simp [Observation.isComputation]

/-- Every observation is exactly one of the two. -/
theorem Observation.effect_xor_computation {o : Observation} : o.isEffect = !o.isComputation := by
  simp [Observation.isEffect, Observation.isComputation]

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

/-- What an effect may come through, by the kind of node that sends it: an
exit, from every kind; a goal, from a delegate; what an effect node wrote;
what came back through an invoke node's contract, outside the apprentice. -/
def effectThrough (a : Apprentice) (k : Kind) (c : Contract) : Bool :=
  c == a.exit || (k == .delegate && c == a.goal) ||
  (k == .effect && c == a.effectObservation) || (k == .invoke && !a.mine c)

/-- What `observe` admits. No members: an effect, through what its kind may
make one through. Members: a computation, only from a delegate, through
`delegate`, none twice. -/
def admits (a : Apprentice) (k : Kind) (t : ObservationType) (ms : List Observation) : Bool :=
  if ms.isEmpty then effectThrough a k t.contract
  else k == .delegate && t.contract == a.delegate && decide ms.Nodup

/-- The one function that creates observations from an instruction. The
caller never says what it sends: an effect or a computation is decided by
what it has, and admitted by the kind of node that sent it. -/
def observe (a : Apprentice) (k : Kind) (t : ObservationType) (ms : List Observation) :
    Option Observation :=
  if admits a k t ms = true then some (.mk t ms) else none

/-- The root's endowment: an effect through `root`, with no instruction
before it. Privileged: no node sends it. -/
def rootObservation (a : Apprentice) (t : ObservationType) : Option Observation :=
  if t.contract == a.root then some (.mk t []) else none

theorem observe_eq_some :
    observe a k t ms = some o ↔ admits a k t ms = true ∧ o.type = t ∧ o.members = ms := by
  unfold observe
  split
  · rename_i h
    constructor
    · intro e; cases e; exact ⟨h, rfl, rfl⟩
    · rintro ⟨-, ht, hm⟩
      cases o; simp_all [Observation.type, Observation.members]
  · rename_i h
    simp only [reduceCtorEq, false_iff]
    rintro ⟨ha, -⟩; exact h ha

/-- Only a delegate makes a computation, and only through `delegate`. -/
theorem observe_computation (h : observe a k t ms = some o) (hc : o.isComputation = true) :
    k = .delegate ∧ t.contract = a.delegate := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, hm⟩ := h
  rw [Observation.computation_iff, hm] at hc
  simp only [admits, List.isEmpty_iff, hc, ite_false, Bool.and_eq_true, beq_iff_eq] at ha
  exact ⟨ha.1.1, ha.1.2⟩

/-- An effect node and an invoke node make only effects. -/
theorem observe_effect_or_invoke (h : observe a k t ms = some o) (hk : k ≠ .delegate) :
    o.isEffect = true := by
  rw [Observation.effect_xor_computation]
  cases hc : o.isComputation
  · rfl
  · exact absurd (observe_computation h hc).1 hk

/-- Nothing is made through `observe`, `root`, `return`, `invoke` or
`delegate` as an effect: `observe` and `root` are privileged, `return`
passes, `invoke` brings back what its contract's type says, and `delegate`
composes. -/
theorem observe_not_through (h : observe a k t [] = some o)
    (hc : t.contract ∈ [a.observe, a.root, a.ret, a.invoke, a.delegate]) : False := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  have hn := a.contracts_nodup
  have hm : a.mine t.contract = true := a.contracts_mine (by
    simp only [Apprentice.contracts, List.mem_cons, List.not_mem_nil, or_false] at hc ⊢
    rcases hc with h | h | h | h | h <;> simp [h])
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  simp only [admits, List.isEmpty_nil, ite_true, effectThrough, Bool.or_eq_true,
    Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true'] at ha
  simp only [List.mem_cons, List.not_mem_nil, or_false] at hc
  rcases ha with (((e | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩)
  all_goals first
    | (rw [e] at hm; simp at hm; done)
    | (rcases hc with hc | hc | hc | hc | hc <;> rw [hc] at e <;>
        first | (rw [e] at hn; simp at hn; done) | (simp_all; done))
```

`observe` is the boundary. It is the only function that creates an
observation from an instruction: `Observation`'s constructor is private, so
nothing outside this file can make one any other way, and every member of an
observation was made before it. Validity throughout a lineage is by
construction; there is nothing else to check. Protecting it means protecting
this file (see "The Go shell"). The root's endowment is the one observation
with no instruction before it, made by `rootObservation`, through `root`.

An observation is an effect or a computation, by what it has:

- **An effect** has no members. Every observation-producing syscall is one:
  an exit, from every kind; a goal, from a delegate, through `goal`; what an
  effect node wrote, through `effect_observation`; what came back through an
  invoke node's contract, outside the apprentice; the root's endowment.
- **A computation** has members, none twice, in order. Only a delegate makes
  one, through `delegate`: the description, the goal, the reach packaged, and
  the observations passed. A child's input is that computation.

The caller never says which it sends. `observe` takes the kind of node
sending, decided in the `observe` phase by `classify` (see "Effect, invoke or
delegate"), and admits what that kind may make. Nothing is made as an effect
through `observe` or `root`, which are privileged, `return`, which passes,
`invoke`, which brings back what the contract's type says, or `delegate`,
which composes. An exit the system forces is a fault, never an exit status.

An observation's type has two halves, typed apart here: the contract it came
back through (the input half) and its own output shape (the return half). At
runtime they are joined in its hash.

A computation's members are a list, in the order passed: order is part of
what a child is shown, so `[a, b]` and `[b, a]` are different computations.
The members' types stay on the members, so a computation holds them without
restating them.

## Inputs

```lean
/-- The empty input: an object with no fields. The caller sends `{}`. -/
def Schema.empty : Schema := ⟨"{\"type\":\"object\"}"⟩
```

A child's input is one computation, the delegate's: the description, the
goal, the reach packaged, then the observations passed to it, in order. It is
never empty: it always holds its goal. `{}` is the input of a contract that takes none. Nothing else is
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
  | .mk t ms => .mk t ms :: ms.attach.flatMap fun ⟨m, _⟩ => lineage m

/-- What a node's return may hold: nothing in the lineage of what it was
passed. -/
def nodeMayReturn (received returned : List Observation) : Bool :=
  returned.all fun o => (lineage o).all fun x => !decide (x ∈ received.flatMap lineage)

theorem nodeMayReturn_iff :
    nodeMayReturn received returned = true ↔
      ∀ o ∈ returned, ∀ x ∈ lineage o, ∀ r ∈ received, x ∉ lineage r := by
  simp [nodeMayReturn, List.all_eq_true]

/-- What `return` may pass: something; and for an invoke or delegate node,
nothing its model created directly. An effect node may return what it wrote.
`created` is the frame's own: what its model made, never a field on an
observation. -/
def mayReturn (k : Kind) (created returned : List Observation) : Bool :=
  !returned.isEmpty && (k == .effect || returned.all fun r => !decide (r ∈ created))

theorem mayReturn_iff :
    mayReturn k created returned = true ↔
      returned ≠ [] ∧ (k = .effect ∨ ∀ r ∈ returned, r ∉ created) := by
  simp [mayReturn, List.all_eq_true]
```

An agent may pass only what is in its scope: what it was passed, what its
children returned, and what it created. The direction says who the sender
is: on invoke the parent, on return the child. An observation is obtained
only by endowment (`rootObservation`), creation (through `observe`), or
introduction (passed on invoke or return), the object-capability rule. The
runtime supplies the sender's scope as a value, from the records; Lean
checks it.

Every kind returns through `return`, and its shape is `create_return`'s: the
observations returned, in one array, under the name of the node returning
them. The model emits integers into it, references to observations in scope;
passing is not creating. A return holds something, and:

- **An invoke or delegate node** returns nothing its model created directly.
  A violation is a fault. What its model created is the frame's own, known
  from its state, never a field on an observation.
- **An effect node** may return what it wrote.

Nothing returned may meet the lineage of what the node was passed
(`nodeMayReturn`): it would be found twice on one edge, once going down and
once coming back. Nothing is assumed about what a node should return beyond
this.

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

def inputBaseNovelty (history : List InputBase) (i : InputBase) : InputBaseNovelty :=
  if i ∈ history then .seen else .unseen
```

Novelty is by base type, never by value, and split by direction, as one input
gives many observations. An input is unseen when this binary has not been
handed this sequence of types before. An observation is unseen when the system
does not recognize its base type's hash (`see`, below).
The goal is typed like any observation; its base type never varies, so it
does not change an input's novelty.

Novelty is apart from validity: validity refuses, novelty measures. The
runtime supplies the history; Lean answers.

## Base types and refinement types

```lean
/-- A refinement type, by its hash: a base type and its guard's code. -/
structure RefinementId where
  val : String
  deriving DecidableEq

/-- A base type, by its hash: the binary, the base types it was passed, in
order, and the shape of what came back. No guard is in it. -/
structure BaseId where
  val : String
  deriving DecidableEq

/-- A guard over one base type. `default` until the model writes one: the
shell registers the default CEL for it. A guard for one base type is not a
guard for another. -/
inductive Guard (b : BaseId)
  | default
  | written (cel : String)
  deriving DecidableEq

/-- How the runtime hashes. Lean never computes a hash: Go does, and Lean
takes the ids as values. -/
structure Hashes where
  base   : Binary → List BaseId → Schema → BaseId
  refine : (b : BaseId) → Guard b → RefinementId

/-- What is assumed of the hashes: no collisions. Every theorem that needs it
takes it as a hypothesis. -/
def Hashes.Sound (h : Hashes) : Prop :=
  (∀ {b₁ b₂ i₁ i₂ o₁ o₂}, h.base b₁ i₁ o₁ = h.base b₂ i₂ o₂ →
      b₁ = b₂ ∧ i₁ = i₂ ∧ o₁ = o₂) ∧
  (∀ {b₁ b₂} {g₁ : Guard b₁} {g₂ : Guard b₂}, h.refine b₁ g₁ = h.refine b₂ g₂ → b₁ = b₂)

/-- The guards: a function from base type to at most one guard. -/
def Guards := (b : BaseId) → Option (Guard b)

/-- Put a guard in a base type's one slot. -/
def Guards.set (g : Guards) (b : BaseId) (x : Guard b) : Guards :=
  fun a => if h : a = b then some (h ▸ x) else g a

theorem Guards.set_same (g : Guards) (b : BaseId) (x : Guard b) : g.set b x b = some x := by
  simp [Guards.set]

theorem Guards.set_other (g : Guards) (h : a ≠ b) (x : Guard b) : g.set b x a = g a := by
  simp [Guards.set, h]

/-- Two levels, kept apart. `known` is the base level: the base types the
system recognizes. A hash not in it is unseen; it has no slot. `guards` is
the refinement level: one slot per known base type. Not keyed by parent: the
base type already holds the binary and what it was passed. -/
structure Store where
  known  : List BaseId
  guards : Guards

/-- Every known base type has exactly one guard, and only a known one has
any. -/
def Store.WellFormed (s : Store) : Prop :=
  ∀ b, (s.guards b).isSome = true ↔ b ∈ s.known

def Store.empty : Store := ⟨[], fun _ => none⟩

theorem Store.empty_wellFormed : Store.empty.WellFormed := by
  intro b; simp [Store.empty]

inductive Seen
  | seen
  | unseen
  deriving DecidableEq

/-- A return of base type `b` came back. Lean decides whether the system
recognizes the hash. On unseen, `b` becomes known, and its slot is made with
its one guard, the default. The shell reacts: on unseen it registers the
default CEL for `b`. -/
def see (s : Store) (b : BaseId) : Seen × Store :=
  if b ∈ s.known then (.seen, s)
  else (.unseen, ⟨b :: s.known, s.guards.set b .default⟩)

/-- Unseen is a hash the system does not recognize. -/
theorem see_unseen_iff : (see s b).1 = .unseen ↔ b ∉ s.known := by
  unfold see; split <;> simp_all

/-- After `see`, the base type is known. -/
theorem see_known : b ∈ (see s b).2.known := by
  unfold see; split <;> simp_all

/-- `known` only grows. -/
theorem see_grows (ha : a ∈ s.known) : a ∈ (see s b).2.known := by
  unfold see; split <;> simp_all

/-- `see` touches no other base type's guard. -/
theorem see_other (h : a ≠ b) : (see s b).2.guards a = s.guards a := by
  unfold see; split <;> simp [Guards.set_other _ h]

/-- `see` keeps the store well formed: a slot is made exactly when a base
type becomes known. -/
theorem see_wellFormed (hs : s.WellFormed) : (see s b).2.WellFormed := by
  intro a
  unfold see
  split
  · exact hs a
  · by_cases h : a = b
    · subst h; simp [Guards.set_same]
    · simp [Guards.set_other _ h, h, hs a]

/-- A guard broke, the model ran, and wrote another. It overwrites the one
slot. An unknown base type has no slot to overwrite. -/
def rewrite (s : Store) (b : BaseId) (cel : String) : Option Store :=
  if b ∈ s.known then some ⟨s.known, s.guards.set b (.written cel)⟩ else none

theorem rewrite_known (hr : rewrite s b cel = some s') : s'.known = s.known := by
  unfold rewrite at hr; split at hr <;> cases hr; rfl

theorem rewrite_wellFormed (hs : s.WellFormed) (hr : rewrite s b cel = some s') :
    s'.WellFormed := by
  unfold rewrite at hr
  split at hr
  · rename_i hb
    cases hr
    intro a
    by_cases h : a = b
    · subst h; simp [Guards.set_same, hb]
    · simp [Guards.set_other _ h, hs a]
  · cases hr

/-- A base type's refinement type: its base and its one guard, hashed. -/
def refinement (h : Hashes) (s : Store) (b : BaseId) : Option RefinementId :=
  (s.guards b).map (h.refine b)

/-- In a well-formed store, a base type has a refinement type exactly when it
is known. -/
theorem refinement_known (hs : s.WellFormed) :
    (refinement h s b).isSome = true ↔ b ∈ s.known := by
  simp only [refinement, Option.isSome_map]; exact hs b

/-- One refinement type per base type. -/
theorem one_refinement (h₁ : refinement h s b = some r₁) (h₂ : refinement h s b = some r₂) :
    r₁ = r₂ := by
  rw [h₁] at h₂; exact Option.some.inj h₂

/-- A refinement type belongs to one base type: a guard cannot move. -/
theorem refinement_base {h : Hashes} (hs : h.Sound) (h₁ : refinement h s b₁ = some r)
    (h₂ : refinement h s' b₂ = some r) : b₁ = b₂ := by
  simp only [refinement, Option.map_eq_some_iff] at h₁ h₂
  obtain ⟨x₁, -, rfl⟩ := h₁
  obtain ⟨x₂, -, e⟩ := h₂
  exact (hs.2 e).symm

/-- Same binary, same base types passed, same shape back: one base type,
whichever node it came from; and only then. -/
theorem base_eq_iff {h : Hashes} (hs : h.Sound) :
    h.base b₁ i₁ o₁ = h.base b₂ i₂ o₂ ↔ b₁ = b₂ ∧ i₁ = i₂ ∧ o₁ = o₂ := by
  constructor
  · exact hs.1
  · rintro ⟨rfl, rfl, rfl⟩; rfl
```

A base type is a hash of three things: the binary, the base types it was
passed, in order, and the shape of what came back. A refinement type is a
hash of a base type and its guard's code. Refinement types are never part of
a base type: two nodes with one binary, handed observations of the same base
types and returning the same shape, are one base type, wherever they sit in
the tree, whatever guards stand upstream.

Equality is by hash, so it is string equality, and Lean never computes a
hash: Go does, and hands Lean the ids. What that assumes is written once,
`Hashes.Sound` (no collisions), and every theorem that needs it takes it as a
hypothesis.

Two levels, kept apart (`Store`). The base level is `known`: the base types
the system recognizes. A base type is unseen when the system does not
recognize its hash: it has no slot at all, not an empty one. The refinement
level is `guards`: one slot per known base type, holding its one guard. A
slot is made exactly when a base type becomes known, so every known base type
has exactly one guard and only a known one has any (`Store.WellFormed`).
Neither level is keyed by parent: the base type already holds the binary and
what it was passed, and a child never knows its parent exists. The parent
matters to the records of what happened, and to remembering, which is off.

When a return comes back, Lean decides (`see`): seen, or unseen. On unseen
the base type becomes known, and its slot is made with its one guard, the
default. That is an invariant of the system, so it is Lean's. The shell
reacts: on unseen it registers the default CEL for that base type. When a
guard breaks, the model runs and writes another, which overwrites the one
slot (`rewrite`); the base type is the same and stays known, and its
refinement type is the new hash. An unknown base type has no slot to
rewrite. `known` only grows (`see_grows`, `rewrite_known`), and both keep the
store well formed (`see_wellFormed`, `rewrite_wellFormed`).

So there is one refinement type per base type (`one_refinement`), and a
guard cannot move to another base type (`refinement_base`): its refinement
type hashes in the base type it was written for, and `Guard b` is indexed by
it.

A rewrite of base type `T`'s guard may cascade, by design: a new guard means
the behavior changed. It runs both ways along the tree's edges: down, to
every node passed an observation of `T`, and up, to the node it was returned
to. On each side,
each guard meets what now comes to it. Where one holds, it absorbs the change
and the cascade stops there. Where one breaks, the model runs and rewrites
it, and that new refinement type is a change for its own neighbours in turn,
up and down. `T` itself stays seen; the cascade goes only as far as guards
cannot absorb it. A rewrite changes no base type, so every guard downstream
keeps its slot, and the guard decides.

### Example: the path is not in the type

One observation is passed down a tree by two paths: 11 deep on the left, 9
deep on the right. At the bottom of each, the same binary is handed it, and
returns the same shape.

```
left:   n1 → n2 → … → n11 → B(obs) → [string]
right:  m1 → m2 → … → m9  → B(obs) → [string]
```

Both returns have one refinement type. The base type is the binary, the
base types passed, and the shape back, and all three agree:

- **The binary**: `B` on both sides.
- **What it was passed**: `obs`, whose base type travels with it. A
  pass is an introduction, never a new observation, so 11 hops or 9 leave it
  unchanged.
- **The shape back**: `[string]` on both sides.

So one base type, one guard, one refinement type. Depth is not in the hash,
and neither is the path.

They differ only if what `B` is handed differs: another observation in its
input, one of another base type, or the same ones in another order
(`[a, b]` is not `[b, a]`). The goal does not count: its base type never
varies.

## Guards

```lean
/-- What the apprentice does after an observation comes back. -/
inductive Next
  | continue  -- take the next instruction; no model
  | fallback  -- an assumption broke; the model runs

/-- An unseen shape broke an assumption; a seen one continues only while its
guard holds. -/
def next : Seen → (guardHolds : Bool) → Next
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
on the value last returned. It is overwritten when the same input returns the
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

A procedure replays its calls under the original invoke grants, and creates
each observation as the kind that first created it. So a
procedure in place of a subtree is still that subtree in the records:
lineage, scope and `nodeMayReturn` hold as they did for the model.

### Who executes

Guards decide; they never run anything. On both paths:

```
CEL guard         holds / breaks                       pure, no effects
Lean next/enter   continue / fallback / procedure      policy
apprentice        writes the instruction               meaning; cannot exec
skyra-spawn       checks the grant, execs              mechanism; the only exec
```

Execution is enforced by the OS, never by CEL, Lean or a Go package any code
could route around ([52](52-kernel-provisions-apprentice-spawns-proposal.md)):
the apprentice's seccomp filter denies `fork`, `pipe` and `exec`, and its one
authority is a socket to `skyra-spawn`, bound to its grant. `skyra-spawn`
alone execs pinned binaries, and only those the asking socket's grant allows.
A procedure's replayed call runs in the Landlock domain of the invoke grant it
was learned under, as the model's invoke node would have.

**For now, on macOS:** 52 is Linux only and not built. The experiment runs
under today's kernel, which mediates every hop, with Seatbelt allowing exec
only for the program being run. That enforces part of the rule, not "only
`skyra-spawn` execs"; accepted until 52 exists.

## Effect, invoke or delegate

```lean
/-- The kind, from the contracts counted. -/
def kind (counted : List Contract) : Kind :=
  match counted.eraseDups.length with
  | 0 => .effect
  | 1 => .invoke
  | _ => .delegate

/-- The contracts that count: the reach, the apprentice's taken out. -/
def counted (a : Apprentice) (reach : List Contract) : List Contract :=
  reach.filter (!a.mine ·)

/-- What the core runs, in the `observe` phase: the apprentice, the agent's
contract and the contracts in its reach, as values. The apprentice's
contracts count for neither the self check nor the kind. `none` refuses an
agent whose own binary is in the rest of its reach, at every count. -/
def classify (a : Apprentice) (self : Contract) (reach : List Contract) : Option Kind :=
  if self.binary ∈ (counted a reach).map (·.binary) then none else some (kind (counted a reach))

theorem classify_none_iff :
    classify a self reach = none ↔ self.binary ∈ (counted a reach).map (·.binary) := by
  unfold classify; split <;> simp_all

theorem classify_effect_iff :
    classify a self reach = some .effect ↔
      self.binary ∉ (counted a reach).map (·.binary) ∧ (counted a reach).eraseDups.length = 0 := by
  unfold classify kind; split <;> (try split) <;> simp_all

theorem classify_invoke_iff :
    classify a self reach = some .invoke ↔
      self.binary ∉ (counted a reach).map (·.binary) ∧ (counted a reach).eraseDups.length = 1 := by
  unfold classify kind; split <;> (try split) <;> simp_all

theorem classify_delegate_iff :
    classify a self reach = some .delegate ↔
      self.binary ∉ (counted a reach).map (·.binary) ∧ 2 ≤ (counted a reach).eraseDups.length := by
  unfold classify kind
  by_cases hb : self.binary ∈ (counted a reach).map (·.binary)
  · simp [hb]
  · simp only [hb, ite_false, Option.some.injEq, not_false_eq_true, true_and]
    generalize (counted a reach).eraseDups.length = n
    match n with
    | 0 => simp
    | 1 => simp
    | n + 2 => simp

/-- The apprentice's contracts never count: adding any of them, before or
after, leaves the result unchanged, refused included. -/
theorem classify_apprentice_after (h : ∀ c ∈ extra, a.mine c = true) :
    classify a self (reach ++ extra) = classify a self reach := by
  have : extra.filter (!a.mine ·) = [] := by
    simp only [List.filter_eq_nil_iff, Bool.not_eq_true']; intro c hc; simp [h c hc]
  unfold classify counted; rw [List.filter_append, this, List.append_nil]

theorem classify_apprentice_before (h : ∀ c ∈ extra, a.mine c = true) :
    classify a self (extra ++ reach) = classify a self reach := by
  have : extra.filter (!a.mine ·) = [] := by
    simp only [List.filter_eq_nil_iff, Bool.not_eq_true']; intro c hc; simp [h c hc]
  unfold classify counted; rw [List.filter_append, this, List.nil_append]

/-- The apprentice's contracts for a kind, loaded into its reach after the
kind is known: `exit` and `return` for every kind; `effect_observation` for
an effect node; `invoke` for an invoke node; `delegate` and `goal` for a
delegate. -/
def loaded (a : Apprentice) : Kind → List Contract
  | .effect   => [a.exit, a.ret, a.effectObservation]
  | .invoke   => [a.exit, a.ret, a.invoke]
  | .delegate => [a.exit, a.ret, a.delegate, a.goal]

theorem loaded_mine (h : c ∈ loaded a k) : a.mine c = true := by
  apply a.contracts_mine
  cases k <;> simp only [loaded, List.mem_cons, List.not_mem_nil, or_false] at h <;>
    rcases h with rfl | rfl | rfl | rfl <;> simp [Apprentice.contracts]

/-- Loading never changes the kind. -/
theorem classify_loaded :
    classify a self (reach ++ loaded a k) = classify a self reach :=
  classify_apprentice_after fun _ h => loaded_mine h

/-- `observe` and `root` are never loaded. -/
theorem loaded_never_privileged : a.observe ∉ loaded a k ∧ a.root ∉ loaded a k := by
  have hn := a.contracts_nodup
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  cases k <;> simp only [loaded, List.mem_cons, List.not_mem_nil, or_false, not_or] <;> simp_all
```

The core passes the apprentice, an agent's contract and its reach as values,
and Lean decides, in the `observe` phase: refused, an effect node, an invoke
node, or a delegate. The apprentice's contracts are taken out first; they
count for neither the self check nor the kind. Then no contract of the
agent's own binary may be in what is left, at every count, and duplicates
count once:

- **No contract: an effect node.** It writes, through `effect_observation`.
- **One: an invoke node.** It invokes that contract.
- **Two or more: a delegate.** It passes them on, through `delegate`.

Two identities, each the stricter for its question. The self check compares
binaries, so an agent cannot reach itself through another of its shapes. The
count compares contracts, so two shapes of one binary are two contracts, and
make a delegate. Either other choice would open a hole.

Once the kind is known, the apprentice's contracts for it are loaded into
the reach (`loaded`). They are handles with their CEL filled in, run when
selected. `observe` and `root` are privileged and never loaded, and loading
never changes the kind.

Two things this invariant does not cover, on purpose:

- **A contract may be wide.** The count is per contract, and nothing bounds
  what one contract covers: a shape of `{"type":"any"}`, or one contract over
  many subcommands, is one contract and so an invoke node. The rule is only
  as narrow as the contracts registered. A concession for now; the
  experiment registers narrow ones.
- **`self` must be honest.** The theorems hold for the `self` they are given.
  It is the actor being provisioned, supplied by the kernel at pass time,
  never by a parent's payload. The Go enforces this.

## Pass time and provision time

**Pass time** is the parent's act: it chooses the contracts and the
observations a child gets. Checked, in order:

1. **The Boundary.** Every contract passed is held by the parent
   (`passable`).
2. **The observations.** Every one passed is in the parent's scope
   (`mayPass`); the goal is an effect made through `goal`, and the child's
   input the computation made through `delegate`.
3. **The kind.** `classify` runs on the child's contract and the contracts
   passed. Only now is it known to be an effect, invoke or delegate node.

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

**Provision time** is the kernel's act: it grants the child what its kind
allows. A contract outside the apprentice is held with one of two rights: to
invoke it, or only to pass it on. The apprentice's contracts for the kind are
loaded beside them.

```lean
/-- What an agent is provisioned with, from its kind: the contracts it may
invoke, those it may only pass on, and the apprentice's contracts loaded for
its kind. -/
structure Offer where
  invoke     : List Contract
  passOnly   : List Contract
  apprentice : List Contract

def offered (a : Apprentice) : Kind → (passed : List Contract) → Offer
  | .effect,   _      => ⟨[], [], loaded a .effect⟩
  | .invoke,   passed => ⟨passed, [], loaded a .invoke⟩
  | .delegate, passed => ⟨[], passed, loaded a .delegate⟩

/-- Only an invoke node invokes a contract outside the apprentice. -/
theorem offered_invoke (h : (offered a k p).invoke ≠ []) : k = .invoke := by
  cases k <;> simp_all [offered]

theorem offered_delegate_invokes_nothing : (offered a .delegate p).invoke = [] := rfl
theorem offered_effect_invokes_nothing : (offered a .effect p).invoke = [] := rfl

theorem offered_apprentice : (offered a k p).apprentice = loaded a k := by
  cases k <;> rfl
```

- **An effect node** invokes nothing outside the apprentice. It writes,
  through `effect_observation`.
- **An invoke node** may invoke the contract it was passed, through
  `invoke`.
- **A delegate** may only pass on the contracts it was passed, through
  `delegate`, and makes goals through `goal`. It invokes nothing.

Every kind holds `exit` and `return`.

Pass-only is the right a pure-control node needs; capability systems keep
passing on apart from using (seL4's Grant right, SPKI's propagate flag).

## The Go shell

The functional core is policy; the imperative shell is mechanism. The core
is specified and proven in Lean, and runs as one Go package, `core`, that
agrees with the Lean on every vector the Lean writes (see "Plan"). Where this
proposal says Lean decides, the running decision is `core`'s. The shell
computes facts, asks `core`, and carries out the answer; it decides nothing.

| The shell does | With the core's |
|---|---|
| Computes each binary's hash (cdhash or SHA-256) and canonical shapes | `identify`, `ObservationType.of` |
| Supplies the apprentice's binary and contracts | `Apprentice.of` |
| Supplies `self`: the actor being provisioned, from the kernel, never from a payload | `classify` |
| Reads the parent's reach at pass time | `passable` |
| Reads each agent's scope from the records, what a node was passed, and what its model created | `mayPass`, `receive`, `mayReturn`, `nodeMayReturn` |
| Creates every observation by calling `observe` with the sender's kind, and records it; the root's by `rootObservation` | `observe`, `rootObservation` |
| Grants what the kind allows: invoke or pass-only rights, and the apprentice's contracts for it. Faults an invoke of a pass-only contract in the `observe` phase | `offered`, `loaded` |
| Builds the prompt and the contracts offered from the `Offer` | `offered` |
| Reads the history of inputs from the records | `inputBaseNovelty` |
| Hashes base and refinement types, and hands Lean the ids | `Hashes` |
| On unseen, registers the default CEL for that base type; on a rewrite, the model's | `see`, `rewrite` |
| Evaluates a shape's guard in CEL (`apprentice/refine`), writing the default string match until the model writes one | `next` |
| At invoke, evaluates the entry guard, and runs the remembered procedure in place of the model, its calls under the original invoke grants | `enter` |
| Runs the model; stops a run at the depth bound. A forced stop is a fault, never an exit status | |

`observe` is the boundary. In `core`, `Observation` has only unexported
fields and `observe` is the only function that makes one, so nothing outside
the package can make one any other way: Go's form of the private
constructor. `core` is pure, holds no state and imports nothing of the shell.
What protects it is CI: the Lean compiles, `#print axioms` on every theorem
accepts only `propext`, `Classical.choice` and `Quot.sound`, and `core`
agrees with every vector.

Left free, to observe: children may share a contract, a node may grant its
whole reach to one child, and a run that loops stops at the depth bound.

## Plan

Values never enter a type: `ObservationType` stays (contract, output shape),
and base and refinement types are hashes of binaries, base types, shapes
and guards. Value is CEL's, configured per shape, on both
paths. For the first runs, guards use the default implementation and
`remember` is off, so no procedure exists and every path falls back to the
model; guards still run, and whether each would have held is recorded.

1. **Finish this proposal.**
   - The default CEL implementation: a string match on the value, for return
     and entry guards.
   - The refusal rules (no member twice, `mayPass`, `nodeMayReturn`) refuse
     by refinement type, and run after refinement types are assigned: two
     observations of one refinement type are the same, and a second is a
     repeat. There is no stamp; uniqueness is the system's interpretation,
     not an observation's identity. Until Lean takes refinement ids, it
     compares `ObservationType`, which is coarser.
   - Known gaps, written down: endowment has no case in Lean; an invoke
     node that paraphrases its input breaks lineage; contracts may be wide; 52 is not
     built on macOS.
   - The README entry; commit.
2. **The Lean spec and its vectors.** A Lake project pinned to Lean 4.35, in
   `controller/apprentice/invariants`; its files are the source of truth, and
   this proposal points to them.
   Never linked or executed by the apprentice. CI compiles it and runs
   `#print axioms` on every theorem, accepting only `propext`,
   `Classical.choice` and `Quot.sound`. A Lean executable writes test
   vectors as JSON: inputs and Lean's answers for every decision, exhaustive
   where the domain is small, generated across it where it is not.
3. **The Go core.** One pure package, `core`, a port of the Lean, decision
   for decision. `Observation`'s fields unexported; `observe` its only maker.
   Its tests replay every vector and fail on any disagreement, so a change
   to either side alone fails CI. Each invariant is also a Go property test,
   named for its Lean theorem.
4. **The Go shell** (apprentice, `proceduretools`, `claude/resident`), as
   [56](56-tree-machine-proposal.md) specifies it.
   - The contracts offered follow the kind: `exit` and `return` for every
     kind; `effect_observation` for an effect node; `invoke` for an invoke
     node; `delegate` and `goal` for a delegate.
   - `delegate` runs `passable`, `classify` and `offered`; its input is the
     computation made through `observe`; the model passes observations by
     integer references into one array.
   - What comes back is created through `observe` with the sender's kind;
     every return is checked by `mayReturn` and `nodeMayReturn`.
   - Guards are evaluated with `apprentice/refine` (default string match);
     holds or breaks is recorded; with `remember` off, always fall back.
   - Both novelty measures are recorded; the depth bound is kept.
5. **The treemap harness.** Choose the slice (count candidate directories
   first); the spec sections, written by the user; narrow contracts (`list_dir`,
   `stat`) with tight shapes; a `find` manifest taken before the run for the
   coverage check; a renderer that draws the tree from `.skyra/` records.
6. **First run.** Real Claude or Codex, the small slice, several runs. Read
   the trees.

Stop for review after steps 1, 2, 3 and 4.

Open: where `core` lives beside the Lake project in
`controller/apprentice/invariants`.
