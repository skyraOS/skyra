import Invariants.Core

/-! Identity, in two levels. The contract level is an equivalence: two
observations are alike when they came through one contract. The refinement
level tells observations apart within it: one invocation returns many
observations, and their refinement types differ. For that to be sound, a
refinement type must never span two contracts, nor two outputs of one: equal
refinements imply one contract, one output shape, one base. -/

variable {o₁ o₂ : Observation} {b b₁ b₂ : ReturnBase}

/-- The contract level: two observations came through one contract. -/
def Observation.SameContract (o₁ o₂ : Observation) : Prop :=
  o₁.type.contract = o₂.type.contract

theorem Observation.sameContract_equivalence : Equivalence Observation.SameContract :=
  ⟨fun _ => rfl, Eq.symm, Eq.trans⟩

instance Observation.contractSetoid : Setoid Observation :=
  ⟨Observation.SameContract, Observation.sameContract_equivalence⟩

/-- Each kind of id hashes its own domain tag first (`skyra/refinement/v1`,
`skyra/seen/v1`): a guard's refinement type and a value's default refinement
never share an id, whatever their parts. `Hashes.Sound` does not say so. -/
def Hashes.Separated (h : Hashes) : Prop :=
  ∀ {b₁ b₂ : ReturnBase} {g : Guard b₁} {v : String}, h.refine b₁ g ≠ h.seen b₂ v

/-- An observation's refinement is keyed to `b`: the base its type mints,
and either a guard's refinement type of `b` or the default's refinement of a
value of `b`. -/
def Observation.Keyed (h : Hashes) (o : Observation) (b : ReturnBase) : Prop :=
  (∃ ps, baseOf h o.type ps = some b) ∧
    ((∃ g : Guard b, o.refined = h.refine b g) ∨ ∃ v, o.refined = h.seen b v)

/-- One refinement, one base. -/
theorem Observation.Keyed.base {h : Hashes} (hs : h.Sound) (sep : h.Separated)
    (k₁ : o₁.Keyed h b₁) (k₂ : o₂.Keyed h b₂) (e : o₁.refined = o₂.refined) : b₁ = b₂ := by
  rcases k₁.2 with ⟨g₁, r₁⟩ | ⟨v₁, r₁⟩ <;> rcases k₂.2 with ⟨g₂, r₂⟩ | ⟨v₂, r₂⟩ <;>
    rw [r₁, r₂] at e
  · exact hs.2.2.1 e
  · exact absurd e sep
  · exact absurd e.symm sep
  · exact (hs.2.2.2 e).1

/-- One base, one observation type: a base names its contract and its
output shape. -/
theorem baseOf_type {h : Hashes} (hs : h.Sound) (h₁ : baseOf h t₁ p₁ = some b)
    (h₂ : baseOf h t₂ p₂ = some b) : t₁ = t₂ := by
  obtain ⟨c, -, o⟩ := (base_eq_iff hs h₁ h₂).mp rfl
  exact (observation_identity t₁ t₂).mpr ⟨c, o⟩

/-- Equal refinements, one observation type: one contract and one output
shape. -/
theorem Observation.Keyed.type {h : Hashes} (hs : h.Sound) (sep : h.Separated)
    (k₁ : o₁.Keyed h b₁) (k₂ : o₂.Keyed h b₂) (e : o₁.refined = o₂.refined) :
    o₁.type = o₂.type := by
  have hb := k₁.base hs sep k₂ e
  subst hb
  obtain ⟨⟨p₁, h₁⟩, -⟩ := k₁
  obtain ⟨⟨p₂, h₂⟩, -⟩ := k₂
  exact baseOf_type hs h₁ h₂

/-- A refinement type never spans two contracts: the refinement level is
finer than the contract level. -/
theorem Observation.Keyed.sameContract {h : Hashes} (hs : h.Sound) (sep : h.Separated)
    (k₁ : o₁.Keyed h b₁) (k₂ : o₂.Keyed h b₂) (e : o₁.refined = o₂.refined) :
    o₁ ≈ o₂ :=
  congrArg ObservationType.contract (k₁.type hs sep k₂ e)

/-- Through two contracts, two refinement types. -/
theorem Observation.Keyed.apart {h : Hashes} (hs : h.Sound) (sep : h.Separated)
    (k₁ : o₁.Keyed h b₁) (k₂ : o₂.Keyed h b₂) (hc : ¬ o₁ ≈ o₂) : o₁.refined ≠ o₂.refined :=
  fun e => hc (k₁.sameContract hs sep k₂ e)

/-- Of one type, from one set of members, observations are one exactly when
their refinements are one. -/
theorem Observation.eq_iff_refined (ht : o₁.type = o₂.type) (hm : o₁.members = o₂.members) :
    o₁ = o₂ ↔ o₁.refined = o₂.refined := by
  constructor
  · intro e; rw [e]
  · intro e
    cases o₁; cases o₂
    simp only [Observation.type, Observation.members, Observation.refined] at ht hm e
    subst ht hm e
    rfl

/-- The refinement decides: keyed observations from one set of members are
one exactly when their refinements are one. Being alike at the contract
level follows; it is never assumed. -/
theorem Observation.Keyed.eq_iff {h : Hashes} (hs : h.Sound) (sep : h.Separated)
    (k₁ : o₁.Keyed h b₁) (k₂ : o₂.Keyed h b₂) (hm : o₁.members = o₂.members) :
    o₁ = o₂ ↔ o₁.refined = o₂.refined := by
  constructor
  · intro e; rw [e]
  · intro e
    exact (Observation.eq_iff_refined (k₁.type hs sep k₂ e) hm).mpr e
