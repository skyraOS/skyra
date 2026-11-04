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

/-- The apprentice: one binary, and its contracts, each of that binary and
none twice. Some are privileged (`observe`, `root`): never loaded into a
reach. `createDelegation`, `createAct` and `reach` are run by `delegate`'s
own CEL, before the model's point: never loaded either. Its constructor is
private: every one is made by `Apprentice.of`. -/
structure Apprentice where
  private mk ::
  binary            : Binary
  observe           : Contract
  root              : Contract
  exit              : Contract
  ret               : Contract  -- `return`
  performEffect     : Contract  -- `perform_effect`
  createInput       : Contract  -- `create_input`
  delegate          : Contract
  goal              : Contract
  createDelegation  : Contract  -- `create_delegation`
  createAct         : Contract  -- `create_act`
  reach             : Contract  -- `package_reach`
  wf : ([observe, root, exit, ret, performEffect, createInput, delegate, goal,
          createDelegation, createAct, reach].all (·.binary == binary) && decide
          [observe, root, exit, ret, performEffect, createInput, delegate, goal,
          createDelegation, createAct, reach].Nodup) = true

def Apprentice.contracts (a : Apprentice) : List Contract :=
  [a.observe, a.root, a.exit, a.ret, a.performEffect, a.createInput, a.delegate, a.goal,
   a.createDelegation, a.createAct, a.reach]

/-- The one way to an apprentice. Refuses a contract of another binary, or
one contract in two places. -/
def Apprentice.of (binary : Binary)
    (observe root exit ret performEffect createInput delegate goal createDelegation createAct reach :
      Contract) :
    Option Apprentice :=
  if h : ([observe, root, exit, ret, performEffect, createInput, delegate, goal,
            createDelegation, createAct, reach].all (·.binary == binary) && decide
            [observe, root, exit, ret, performEffect, createInput, delegate, goal,
            createDelegation, createAct, reach].Nodup) = true
  then some ⟨binary, observe, root, exit, ret, performEffect, createInput, delegate, goal,
    createDelegation, createAct, reach, h⟩
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

/-- What a node is (see "Act or delegate"). A child's kind is decided from
the contracts its delegate passed; the root's is given. There is no effect
node: an effect is the observation, and `perform_effect` the generic that
makes one, taking on the contract passed to it. -/
inductive Kind
  | act       -- one contract: it performs effects through it
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
exit, from every kind; a description and a packaged reach, from a delegate;
an input, from an act node, to dispatch; what came back through an act
node's `perform_effect`, typed by the contract it was passed, outside the
apprentice. A goal is never an effect: it is cut from what a delegate has. -/
def effectThrough (a : Apprentice) (k : Kind) (c : Contract) : Bool :=
  c == a.exit ||
  (k == .delegate && c == a.createDelegation) || (k == .delegate && c == a.createAct) ||
  (k == .delegate && c == a.reach) || (k == .act && c == a.createInput) ||
  (k == .act && !a.mine c)

/-- What `observe` admits. No members: an effect, through what its kind may
make one through. Members: a computation, none twice, from a delegate
through `delegate`; from a delegate through `goal`, a part cut from the one
observation it is its member; or from an act node through `create_input`:
an input filled from names in its scope, which are its members. -/
def admits (a : Apprentice) (k : Kind) (t : ObservationType) (ms : List Observation) : Bool :=
  if ms.isEmpty then effectThrough a k t.contract
  else (k == .delegate && t.contract == a.delegate || k == .act && t.contract == a.createInput ||
    k == .delegate && t.contract == a.goal && ms.length == 1) && decide ms.Nodup

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

/-- A computation comes three ways only: from a delegate, through
`delegate` or `goal`, or from an act node, through `create_input`. -/
theorem observe_computation (h : observe a k t ms = some o) (hc : o.isComputation = true) :
    (k = .delegate ∧ t.contract = a.delegate) ∨ (k = .act ∧ t.contract = a.createInput) ∨
    (k = .delegate ∧ t.contract = a.goal) := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, hm⟩ := h
  rw [Observation.computation_iff, hm] at hc
  simp only [admits, List.isEmpty_iff, hc, ite_false, Bool.and_eq_true, Bool.or_eq_true,
    beq_iff_eq] at ha
  rcases ha.1 with (h₁ | h₂) | ⟨h₃, -⟩
  · exact Or.inl h₁
  · exact Or.inr (Or.inl h₂)
  · exact Or.inr (Or.inr h₃)

/-- An act node makes a computation only as an input: everything else it
makes is an effect. -/
theorem observe_act (h : observe a .act t ms = some o) :
    o.isEffect = true ∨ t.contract = a.createInput := by
  rw [Observation.effect_xor_computation]
  cases hc : o.isComputation
  · exact Or.inl rfl
  · rcases observe_computation h hc with ⟨hk, -⟩ | ⟨-, ht⟩ | ⟨hk, -⟩
    · exact absurd hk (by decide)
    · exact Or.inr ht
    · exact absurd hk (by decide)

/-- A goal is cut, never written: a computation of exactly one member, what
it was cut from, made by a delegate. Its value is lines of that member's,
which the shell cuts; values never enter the type. -/
theorem observe_goal (h : observe a k t ms = some o) (hc : t.contract = a.goal) :
    k = .delegate ∧ ms.length = 1 := by
  have hd := a.contracts_nodup
  have hm : a.mine t.contract = true := a.contracts_mine (by rw [hc]; simp [Apprentice.contracts])
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hd
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  by_cases he : ms = []
  · subst he
    simp only [admits, List.isEmpty_nil, ite_true, effectThrough, Bool.or_eq_true,
      Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true'] at ha
    rcases ha with (((((e | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩)
    all_goals first
      | (rw [e] at hm; simp at hm; done)
      | (rw [hc] at e; simp_all; done)
  · simp only [admits, List.isEmpty_iff, he, ite_false, Bool.and_eq_true, Bool.or_eq_true,
      beq_iff_eq, decide_eq_true_eq] at ha
    rcases ha with ⟨(⟨-, e⟩ | ⟨-, e⟩) | ⟨⟨hk, -⟩, hl⟩, -⟩
    · rw [hc] at e; simp_all
    · rw [hc] at e; simp_all
    · exact ⟨hk, by simpa using hl⟩

/-- An input's members are names it was filled from, none twice; it is
still one input, sent by an act node. -/
theorem observe_filled_input (h : observe a k t ms = some o) (hc : t.contract = a.createInput)
    (hm : ms ≠ []) : k = .act ∧ ms.Nodup := by
  have hd := a.contracts_nodup
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hd
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  simp only [admits, List.isEmpty_iff, hm, ite_false, Bool.and_eq_true, Bool.or_eq_true,
    beq_iff_eq, decide_eq_true_eq] at ha
  rcases ha with ⟨(⟨-, e⟩ | ⟨hk, -⟩) | ⟨⟨-, e⟩, -⟩, hn⟩
  · rw [hc] at e; simp_all
  · exact ⟨hk, hn⟩
  · rw [hc] at e; simp_all

/-- Nothing is made through `observe`, `root`, `return`, `perform_effect`,
`delegate` or `goal` as an effect: `observe` and `root` are privileged,
`return` passes, `perform_effect` takes on the contract passed to it,
`delegate` composes, and `goal` cuts. -/
theorem observe_not_through (h : observe a k t [] = some o)
    (hc : t.contract ∈ [a.observe, a.root, a.ret, a.performEffect, a.delegate, a.goal]) : False := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  have hn := a.contracts_nodup
  have hm : a.mine t.contract = true := a.contracts_mine (by
    simp only [Apprentice.contracts, List.mem_cons, List.not_mem_nil, or_false] at hc ⊢
    rcases hc with h | h | h | h | h | h <;> simp [h])
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  simp only [admits, List.isEmpty_nil, ite_true, effectThrough, Bool.or_eq_true,
    Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true'] at ha
  simp only [List.mem_cons, List.not_mem_nil, or_false] at hc
  rcases ha with (((((e | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩)
  all_goals first
    | (rw [e] at hm; simp at hm; done)
    | (rcases hc with hc | hc | hc | hc | hc | hc <;> rw [hc] at e <;>
        first | (rw [e] at hn; simp at hn; done) | (simp_all; done))

/-- A description and a packaged reach come only from a delegate: they are
parts of what it composes. -/
theorem observe_composed (h : observe a k t [] = some o)
    (hc : t.contract ∈ [a.createDelegation, a.createAct, a.reach]) : k = .delegate := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  have hn := a.contracts_nodup
  have hm : a.mine t.contract = true := a.contracts_mine (by
    simp only [Apprentice.contracts, List.mem_cons, List.not_mem_nil, or_false] at hc ⊢
    rcases hc with h | h | h <;> simp [h])
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  simp only [admits, List.isEmpty_nil, ite_true, effectThrough, Bool.or_eq_true,
    Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true'] at ha
  simp only [List.mem_cons, List.not_mem_nil, or_false] at hc
  rcases ha with (((((e | ⟨hk, -⟩) | ⟨hk, -⟩) | ⟨hk, -⟩) | ⟨-, e⟩) | ⟨-, e⟩)
  · rcases hc with hc | hc | hc <;> rw [hc] at e <;> simp_all
  all_goals first
    | exact hk
    | (rw [e] at hm; simp at hm; done)
    | (rcases hc with hc | hc | hc <;> rw [hc] at e <;> simp_all; done)

/-- An input comes only from an act node: what it makes to dispatch. -/
theorem observe_input (h : observe a k t [] = some o) (hc : t.contract = a.createInput) :
    k = .act := by
  rw [observe_eq_some] at h
  obtain ⟨ha, -, -⟩ := h
  have hn := a.contracts_nodup
  have hm : a.mine t.contract = true := a.contracts_mine (by rw [hc]; simp [Apprentice.contracts])
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  simp only [admits, List.isEmpty_nil, ite_true, effectThrough, Bool.or_eq_true,
    Bool.and_eq_true, beq_iff_eq, Bool.not_eq_true'] at ha
  rcases ha with (((((e | ⟨-, e⟩) | ⟨-, e⟩) | ⟨-, e⟩) | ⟨hk, -⟩) | ⟨hk, -⟩)
  all_goals first
    | exact hk
    | (rw [hc] at e; simp_all; done)

/-- The empty input: an object with no fields. The caller sends `{}`. -/
def Schema.empty : Schema := ⟨"{\"type\":\"object\"}"⟩

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

/-- What `return` may pass: something, and nothing its model created
directly. `created` is the frame's own: what its model made, never a field
on an observation. -/
def mayReturn (created returned : List Observation) : Bool :=
  !returned.isEmpty && returned.all fun r => !decide (r ∈ created)

theorem mayReturn_iff :
    mayReturn created returned = true ↔ returned ≠ [] ∧ ∀ r ∈ returned, r ∉ created := by
  simp [mayReturn, List.all_eq_true]

/-- An input base type, by its hash: the contract invoked, and the return
base types of what it was passed, in order. Its constructor is private: every one
is made by `createInvoke`. -/
structure InputBase where
  private mk ::
  val : String
  deriving DecidableEq

/-- A return base type, by its hash: the input base type it answers, and the
shape of what came back. No guard is in it. Its constructor is private: every
one is made by `createReturn`, which takes an input base type, so every
return answers an invoke. -/
structure ReturnBase where
  private mk ::
  val : String
  deriving DecidableEq

/-- A refinement type, by its hash: a return base type and its guard's code. -/
structure RefinementId where
  val : String
  deriving DecidableEq

/-- A guard over one return base type. `default` until the model writes one:
the shell registers the default CEL for it. A guard for one return base type
is not a guard for another. -/
inductive Guard (b : ReturnBase)
  | default
  | written (cel : String)
  deriving DecidableEq

/-- How the runtime hashes. Lean never computes a hash: Go does, and Lean
takes the hashes as values. -/
structure Hashes where
  input  : Contract → List ReturnBase → String
  ret    : InputBase → Schema → String
  refine : (b : ReturnBase) → Guard b → RefinementId

/-- What is assumed of the hashes: no collisions. Every theorem that needs it
takes it as a hypothesis. -/
def Hashes.Sound (h : Hashes) : Prop :=
  (∀ {c₁ c₂ p₁ p₂}, h.input c₁ p₁ = h.input c₂ p₂ → c₁ = c₂ ∧ p₁ = p₂) ∧
  (∀ {i₁ i₂ o₁ o₂}, h.ret i₁ o₁ = h.ret i₂ o₂ → i₁ = i₂ ∧ o₁ = o₂) ∧
  (∀ {b₁ b₂} {g₁ : Guard b₁} {g₂ : Guard b₂}, h.refine b₁ g₁ = h.refine b₂ g₂ → b₁ = b₂)

/-- Down an edge (`create_invoke`): the one way to an input base type. From
the contract, not its binary: two contracts of one binary, the apprentice's
included, are two input base types. A contract is well formed by
construction, so it never refuses. -/
def createInvoke (h : Hashes) (c : Contract) (passed : List ReturnBase) : InputBase :=
  ⟨h.input c passed⟩

/-- Up an edge (`create_return`): the one way to a return base type. It takes
the input base type it answers. Refuses a shape that is not well formed. -/
def createReturn (h : Hashes) (i : InputBase) (shape : Schema) : Option ReturnBase :=
  if shape.wellFormed then some ⟨h.ret i shape⟩ else none

/-- Same contract, same return base types passed: one input base type; and
only then. -/
theorem input_eq_iff {h : Hashes} (hs : h.Sound) :
    createInvoke h c₁ p₁ = createInvoke h c₂ p₂ ↔ c₁ = c₂ ∧ p₁ = p₂ := by
  constructor
  · intro e; exact hs.1 (congrArg InputBase.val e)
  · rintro ⟨rfl, rfl⟩; rfl

/-- Same input base type, same shape back: one return base type; and only
then. -/
theorem return_eq_iff {h : Hashes} (hs : h.Sound) (h₁ : createReturn h i₁ o₁ = some r₁)
    (h₂ : createReturn h i₂ o₂ = some r₂) : r₁ = r₂ ↔ i₁ = i₂ ∧ o₁ = o₂ := by
  unfold createReturn at h₁ h₂
  split at h₁ <;> split at h₂ <;> cases h₁ <;> cases h₂
  constructor
  · intro e; exact hs.2.1 (congrArg ReturnBase.val e)
  · rintro ⟨rfl, rfl⟩; rfl

/-- Same contract, same return base types passed, same shape back: one
return base type, whichever node it came from; and only then. -/
theorem base_eq_iff {h : Hashes} (hs : h.Sound)
    (hr₁ : createReturn h (createInvoke h c₁ p₁) o₁ = some r₁)
    (hr₂ : createReturn h (createInvoke h c₂ p₂) o₂ = some r₂) :
    r₁ = r₂ ↔ c₁ = c₂ ∧ p₁ = p₂ ∧ o₁ = o₂ := by
  rw [return_eq_iff hs hr₁ hr₂, input_eq_iff hs, and_assoc]

/-- Nothing comes from nowhere: what a call to a contract made, the
apprentice's included (a goal, an input, a description, a packaged reach),
has a return base type, minted the one way: down the edge to the contract
of its type, and back up with its type's output shape. -/
def baseOf (h : Hashes) (t : ObservationType) (passed : List ReturnBase) : Option ReturnBase :=
  createReturn h (createInvoke h t.contract passed) t.output

/-- An observation type's output shape is well formed, so it always has one. -/
theorem baseOf_isSome : (baseOf h t p).isSome := by
  unfold baseOf createReturn; rw [t.wf]; rfl

/-- Made through two contracts, two return base types: a goal and an input
are never one, nor `x` and `x2`. -/
theorem baseOf_apart {h : Hashes} (hs : h.Sound) (hc : t₁.contract ≠ t₂.contract)
    (h₁ : baseOf h t₁ p₁ = some r₁) (h₂ : baseOf h t₂ p₂ = some r₂) : r₁ ≠ r₂ := by
  intro e
  exact hc ((base_eq_iff hs h₁ h₂).mp e).1

inductive Seen
  | seen
  | unseen
  deriving DecidableEq

/-- Do: an input base type came down an edge. Seen when the system has been
handed it before. -/
def «do» (history : List InputBase) (i : InputBase) : Seen :=
  if i ∈ history then .seen else .unseen

theorem do_unseen_iff : «do» history i = .unseen ↔ i ∉ history := by
  unfold «do»; split <;> simp_all

/-- The guards: a function from return base type to at most one guard. -/
def Guards := (b : ReturnBase) → Option (Guard b)

/-- Put a guard in a return base type's one slot. -/
def Guards.set (g : Guards) (b : ReturnBase) (x : Guard b) : Guards :=
  fun a => if h : a = b then some (h ▸ x) else g a

theorem Guards.set_same (g : Guards) (b : ReturnBase) (x : Guard b) : g.set b x b = some x := by
  simp [Guards.set]

theorem Guards.set_other (g : Guards) (h : a ≠ b) (x : Guard b) : g.set b x a = g a := by
  simp [Guards.set, h]

/-- Two levels, kept apart. `known` is the base level: the return base types
the system recognizes. A hash not in it is unseen; it has no slot. `guards`
is the refinement level: one slot per known return base type. Not keyed by
parent: the return base type already holds, through its input base type, the
binary and what it was passed. -/
structure Store where
  known  : List ReturnBase
  guards : Guards

/-- Every known base type has exactly one guard, and only a known one has
any. -/
def Store.WellFormed (s : Store) : Prop :=
  ∀ b, (s.guards b).isSome = true ↔ b ∈ s.known

def Store.empty : Store := ⟨[], fun _ => none⟩

theorem Store.empty_wellFormed : Store.empty.WellFormed := by
  intro b; simp [Store.empty]

/-- A return of return base type `b` came back. Lean decides whether the system
recognizes the hash. On unseen, `b` becomes known, and its slot is made with
its one guard, the default. The shell reacts: on unseen it registers the
default CEL for `b`. -/
def see (s : Store) (b : ReturnBase) : Seen × Store :=
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
def rewrite (s : Store) (b : ReturnBase) (cel : String) : Option Store :=
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

/-- A return base type's refinement type: its base and its one guard, hashed. -/
def refinement (h : Hashes) (s : Store) (b : ReturnBase) : Option RefinementId :=
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
  exact (hs.2.2 e).symm

/-- The return guard: what the apprentice does after a return comes back. -/
inductive ReturnResult
  | continue  -- take the next instruction; no model
  | fallback  -- an assumption broke; the model runs

/-- An unseen return broke an assumption; a seen one continues only while its
guard holds. -/
def ret : Seen → (guardHolds : Bool) → ReturnResult
  | .unseen, _     => .fallback
  | .seen,   true  => .continue
  | .seen,   false => .fallback

theorem ret_continue_iff : ret n g = .continue ↔ n = .seen ∧ g = true := by
  cases n <;> cases g <;> simp [ret]

/-- The invoke guard: what the core does when a child is invoked. -/
inductive InvokeResult
  | procedure  -- the invoke guard holds: the remembered procedure runs; no model
  | model      -- an assumption broke; the model runs

def invoke : Seen → (guardHolds : Bool) → InvokeResult
  | .unseen, _     => .model
  | .seen,   true  => .procedure
  | .seen,   false => .model

theorem invoke_procedure_iff : invoke n g = .procedure ↔ n = .seen ∧ g = true := by
  cases n <;> cases g <;> simp [invoke]

/-- The kind, from the contracts counted: none for none, as a delegate
passes at least one contract. -/
def kind (counted : List Contract) : Option Kind :=
  match counted.eraseDups.length with
  | 0 => none
  | 1 => some .act
  | _ => some .delegate

/-- The contracts that count: the reach, the apprentice's taken out. -/
def counted (a : Apprentice) (reach : List Contract) : List Contract :=
  reach.filter (!a.mine ·)

/-- What the core runs, in the `observe` phase, for a child: the apprentice,
the child's contract and the contracts its delegate passed, as values. The
apprentice's contracts count for neither the self check nor the kind.
`none` refuses a child whose own binary is in the rest of its reach, at
every count, and a delegate that passed nothing. -/
def classify (a : Apprentice) (self : Contract) (reach : List Contract) : Option Kind :=
  if self.binary ∈ (counted a reach).map (·.binary) then none else kind (counted a reach)

theorem classify_none_iff :
    classify a self reach = none ↔
      self.binary ∈ (counted a reach).map (·.binary) ∨ (counted a reach).eraseDups.length = 0 := by
  unfold classify kind
  by_cases hb : self.binary ∈ (counted a reach).map (·.binary)
  · simp [hb]
  · simp only [hb, ite_false, false_or]
    generalize (counted a reach).eraseDups.length = n
    match n with
    | 0 => simp
    | 1 => simp
    | n + 2 => simp

theorem classify_act_iff :
    classify a self reach = some .act ↔
      self.binary ∉ (counted a reach).map (·.binary) ∧ (counted a reach).eraseDups.length = 1 := by
  unfold classify kind
  by_cases hb : self.binary ∈ (counted a reach).map (·.binary)
  · simp [hb]
  · simp only [hb, ite_false, not_false_eq_true, true_and]
    generalize (counted a reach).eraseDups.length = n
    match n with
    | 0 => simp
    | 1 => simp
    | n + 2 => simp

theorem classify_delegate_iff :
    classify a self reach = some .delegate ↔
      self.binary ∉ (counted a reach).map (·.binary) ∧ 2 ≤ (counted a reach).eraseDups.length := by
  unfold classify kind
  by_cases hb : self.binary ∈ (counted a reach).map (·.binary)
  · simp [hb]
  · simp only [hb, ite_false, not_false_eq_true, true_and]
    generalize (counted a reach).eraseDups.length = n
    match n with
    | 0 => simp
    | 1 => simp
    | n + 2 => simp

/-- A delegate passes at least one contract: a child with nothing counted
in its reach is refused. -/
theorem classify_passes_one (h : classify a self reach = some k) : counted a reach ≠ [] := by
  intro hc
  have : classify a self reach = none := classify_none_iff.mpr (Or.inr (by rw [hc]; rfl))
  rw [this] at h; cases h

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

/-- The root is not classified by its reach: no delegate chose it. Its kind
is given, by whoever started it; only the self check applies. -/
def rootKind (a : Apprentice) (self : Contract) (reach : List Contract) (k : Kind) :
    Option Kind :=
  if self.binary ∈ (counted a reach).map (·.binary) then none else some k

theorem rootKind_none_iff :
    rootKind a self reach k = none ↔ self.binary ∈ (counted a reach).map (·.binary) := by
  unfold rootKind; split <;> simp_all

theorem rootKind_given (h : rootKind a self reach k = some k') : k' = k := by
  unfold rootKind at h; split at h <;> cases h; rfl

/-- The act transformation: what `delegate`'s CEL makes before the model's
point, from the child's kind. A delegate child is described by
`create_delegation` and handed the reach packaged; an act child by
`create_act`, its one contract in its offer, with nothing to pass on. -/
def parts (a : Apprentice) : Kind → List Contract
  | .act      => [a.createAct]
  | .delegate => [a.createDelegation, a.reach]

/-- An act child is never described as a delegate, nor handed a reach to
pass on; a delegate child is never described as an act. -/
theorem parts_apart :
    a.createDelegation ∉ parts a .act ∧ a.reach ∉ parts a .act ∧
      a.createAct ∉ parts a .delegate := by
  have hn := a.contracts_nodup
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  simp only [parts, List.mem_cons, List.not_mem_nil, or_false, not_or]
  simp_all [Ne.symm]

/-- The apprentice's contracts for a kind, loaded into its reach after the
kind is known: `exit` and `return` for every kind; `create_input` and
`perform_effect` for an act node; `delegate` and `goal` for a delegate. -/
def loaded (a : Apprentice) : Kind → List Contract
  | .act      => [a.exit, a.ret, a.createInput, a.performEffect]
  | .delegate => [a.exit, a.ret, a.delegate, a.goal]

theorem loaded_mine (h : c ∈ loaded a k) : a.mine c = true := by
  apply a.contracts_mine
  cases k <;> simp only [loaded, List.mem_cons, List.not_mem_nil, or_false] at h <;>
    rcases h with rfl | rfl | rfl | rfl <;> simp [Apprentice.contracts]

/-- Loading never changes the kind. -/
theorem classify_loaded :
    classify a self (reach ++ loaded a k) = classify a self reach :=
  classify_apprentice_after fun _ h => loaded_mine h

/-- `observe` and `root` are never loaded, nor `create_delegation`,
`create_act` and `package_reach`, which `delegate`'s CEL runs before the
model's point. -/
theorem loaded_never_privileged :
    a.observe ∉ loaded a k ∧ a.root ∉ loaded a k ∧ a.createDelegation ∉ loaded a k ∧
      a.createAct ∉ loaded a k ∧ a.reach ∉ loaded a k := by
  have hn := a.contracts_nodup
  simp only [Apprentice.contracts, List.nodup_cons, List.mem_cons, List.not_mem_nil,
    or_false, not_or] at hn
  cases k <;> simp only [loaded, List.mem_cons, List.not_mem_nil, or_false, not_or] <;>
    simp_all [Ne.symm]

/-- Pass time: every contract passed is held by the parent. -/
def passable (held passed : List Contract) : Bool :=
  passed.all (· ∈ held)

theorem passable_iff : passable held passed = true ↔ passed ⊆ held := by
  simp [passable, List.all_eq_true, List.subset_def]

/-- What an agent is provisioned with, from its kind: the contracts it may
invoke, those it may only pass on, and the apprentice's contracts loaded for
its kind. -/
structure Offer where
  invoke     : List Contract
  passOnly   : List Contract
  apprentice : List Contract

def offered (a : Apprentice) : Kind → (passed : List Contract) → Offer
  | .act,      passed => ⟨passed, [], loaded a .act⟩
  | .delegate, passed => ⟨[], passed, loaded a .delegate⟩

/-- Only an act node invokes a contract outside the apprentice. -/
theorem offered_invoke (h : (offered a k p).invoke ≠ []) : k = .act := by
  cases k <;> simp_all [offered]

theorem offered_delegate_invokes_nothing : (offered a .delegate p).invoke = [] := rfl

theorem offered_apprentice : (offered a k p).apprentice = loaded a k := by
  cases k <;> rfl

