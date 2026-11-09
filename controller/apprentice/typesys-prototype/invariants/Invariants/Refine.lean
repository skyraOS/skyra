/-! Proposals 58 and 59: one base's versioned refinement. A refinement
holds branches, each a condition (CEL, abstracted here as `Cond`)
selecting an instruction. The model writes a guard, a chain of conditions,
and never sees the existing branches; the system stores the chain whole,
as branches under one voter. When values arrive, each voter whose branch
selects them casts that branch's instruction, and `tally` counts the votes
(59). A transition is one instruction's life, ending in `trust`; a frame
takes at most its transition bound of them, and an exhausted frame returns
an error.

Everything here is pure. Which version is current, the store, and the
scope table are not in this module. -/

namespace Refine

/-! ## Conditions -/

/-- A condition: atoms, `true`, conjunction and negation. Atoms are what
the guard compares; `sat` says when an atom holds of a value. -/
inductive Cond (A : Type) where
  | atom (a : A)
  | tt
  | and (p q : Cond A)
  | not (p : Cond A)
  deriving Repr, DecidableEq

variable {A V I : Type}

def Cond.eval (sat : A → V → Bool) : Cond A → V → Bool
  | .atom a, v => sat a v
  | .tt, _ => true
  | .and p q, v => p.eval sat v && q.eval sat v
  | .not p, v => !p.eval sat v

/-- No condition of the list holds. -/
def notAny : List (Cond A) → Cond A
  | [] => .tt
  | c :: cs => .and (.not c) (notAny cs)

theorem notAny_eval {sat : A → V → Bool} :
    ∀ {cs : List (Cond A)}, (notAny cs).eval sat v = true ↔ ∀ c ∈ cs, c.eval sat v = false
  | [] => by simp [notAny, Cond.eval]
  | c :: cs => by simp [notAny, Cond.eval, notAny_eval]

/-! ## Branches and refinements -/

/-- A branch. `id` is its local id, unique only within its base's
refinement; `rank` is the local id it was first committed under, which
orders it: a lower rank is older; `voter` is the local id its chain was
committed under. -/
structure Branch (A I : Type) where
  id : Nat
  rank : Nat
  voter : Nat
  cond : Cond A
  instr : I
  deriving Repr, DecidableEq

def Branch.selects (sat : A → V → Bool) (b : Branch A I) (v : V) : Bool := b.cond.eval sat v

/-- One version of a base's refinement: its branches, unordered, and the
next local id. Ids start at 1 (`b1`). -/
structure Refinement (A I : Type) where
  next : Nat
  branches : List (Branch A I)
  deriving Repr

def Refinement.empty : Refinement A I := ⟨1, []⟩

/-- Local ids are below the counter and never repeat; a rank is an id
already issued, and no two branches share one. -/
def Refinement.WellFormed (r : Refinement A I) : Prop :=
  (∀ b ∈ r.branches, b.id < r.next ∧ b.rank ≤ b.id) ∧
    (r.branches.map (·.id)).Nodup ∧ (r.branches.map (·.rank)).Nodup

theorem Refinement.empty_wellFormed : (Refinement.empty : Refinement A I).WellFormed := by
  simp [Refinement.WellFormed, Refinement.empty]

theorem nodup_map_inj {f : α → β} :
    ∀ {l : List α}, (l.map f).Nodup → ∀ {x y}, x ∈ l → y ∈ l → f x = f y → x = y
  | [], _, _, _, hx, _, _ => by simp at hx
  | a :: l, h, x, y, hx, hy, e => by
    rw [List.map_cons, List.nodup_cons] at h
    rcases List.mem_cons.mp hx with rfl | hx' <;> rcases List.mem_cons.mp hy with rfl | hy'
    · rfl
    · exact absurd (e ▸ List.mem_map_of_mem hy') h.1
    · exact absurd (e.symm ▸ List.mem_map_of_mem hx') h.1
    · exact nodup_map_inj h.2 hx' hy' e

/-! ## Selection -/

def selected (sat : A → V → Bool) (bs : List (Branch A I)) (v : V) : List (Branch A I) :=
  bs.filter (·.selects sat v)

inductive Outcome (I : Type) where
  | null
  | select (i : I)
  | conflict
  deriving Repr, DecidableEq

/-- `null` when nothing selects the values, one instruction when every
selected branch agrees on it, otherwise a conflict. Two branches selecting
one instruction are not a conflict. -/
def distinguish [DecidableEq I] (sat : A → V → Bool) (bs : List (Branch A I)) (v : V) :
    Outcome I :=
  match selected sat bs v with
  | [] => .null
  | b :: rest => if rest.all (·.instr = b.instr) then .select b.instr else .conflict

section
variable [DecidableEq I] {sat : A → V → Bool} {bs bs' : List (Branch A I)} {v : V}

theorem distinguish_null : distinguish sat bs v = .null ↔ selected sat bs v = [] := by
  unfold distinguish; split <;> simp_all <;> split <;> simp

theorem distinguish_select :
    distinguish sat bs v = .select i ↔
      selected sat bs v ≠ [] ∧ ∀ b ∈ selected sat bs v, b.instr = i := by
  unfold distinguish
  split
  · simp_all
  · rename_i b rest h
    rw [h]
    split
    · rename_i ha
      simp only [List.all_eq_true, decide_eq_true_eq] at ha
      constructor
      · rintro ⟨⟩; exact ⟨by simp, fun c hc => by
          rcases List.mem_cons.mp hc with rfl | hc
          · rfl
          · exact ha c hc⟩
      · rintro ⟨-, hi⟩; rw [hi b (by simp)]
    · rename_i ha
      obtain ⟨c, hc, hne⟩ : ∃ c ∈ rest, ¬ c.instr = b.instr := by simpa using ha
      simp only [reduceCtorEq, false_iff, not_and]
      intro _ hi
      exact absurd ((hi c (by simp [hc])).trans (hi b (by simp)).symm) hne

omit [DecidableEq I] in
theorem selected_perm (h : bs.Perm bs') : (selected sat bs v).Perm (selected sat bs' v) :=
  h.filter _

/-- Branches are unordered: the outcome does not depend on their order. -/
theorem distinguish_perm (h : bs.Perm bs') : distinguish sat bs v = distinguish sat bs' v := by
  have hp := selected_perm (sat := sat) (v := v) h
  have mem : ∀ b, b ∈ selected sat bs v ↔ b ∈ selected sat bs' v := fun b => hp.mem_iff
  have nil : selected sat bs v = [] ↔ selected sat bs' v = [] := by
    rw [← List.length_eq_zero_iff, ← List.length_eq_zero_iff, hp.length_eq]
  cases ho : distinguish sat bs v with
  | null => exact (distinguish_null.mpr (nil.mp (distinguish_null.mp ho))).symm
  | select i =>
    obtain ⟨ne, all⟩ := distinguish_select.mp ho
    exact (distinguish_select.mpr ⟨fun e => ne (nil.mpr e),
      fun b hb => all b ((mem b).mpr hb)⟩).symm
  | conflict =>
    cases ho' : distinguish sat bs' v with
    | null => rw [distinguish_null.mpr (nil.mpr (distinguish_null.mp ho'))] at ho; cases ho
    | select j =>
      obtain ⟨ne, all⟩ := distinguish_select.mp ho'
      rw [distinguish_select.mpr ⟨fun e => ne (nil.mp e), fun b hb => all b ((mem b).mp hb)⟩]
        at ho
      cases ho
    | conflict => rfl

end

/-! ## Commit: the model's chain, stored whole -/

/-- The branches of a chain `c₀ ? i₀ : c₁ ? i₁ : … : null`, each with its
effective condition (its own, and no earlier one in the chain), local ids
and ranks from `next`, and all under one voter. -/
def chainBranches (voter : Nat) : Nat → List (Cond A) → List (Cond A × I) → List (Branch A I)
  | _, _, [] => []
  | next, prev, (c, i) :: rest =>
    { id := next, rank := next, voter, cond := .and c (notAny prev), instr := i } ::
      chainBranches voter (next + 1) (prev ++ [c]) rest

/-- What the chain itself selects: its first condition that holds. -/
def chainSelect (sat : A → V → Bool) : List (Cond A × I) → V → Option I
  | [], _ => none
  | (c, i) :: rest, v => if c.eval sat v then some i else chainSelect sat rest v

/-- `startle`'s commit: the current branches, unchanged, and the chain's,
whole. The chain is one new voter, named by the local id of its first
branch. Nothing is narrowed: a branch that overlaps an older one with a
different instruction is a dissenting vote, counted by `tally`. -/
def commit (r : Refinement A I) (chain : List (Cond A × I)) : Refinement A I :=
  ⟨r.next + chain.length, r.branches ++ chainBranches r.next r.next [] chain⟩

section
variable {sat : A → V → Bool} {r : Refinement A I} {b : Branch A I}
  {chain : List (Cond A × I)} {i : I} {v : V} {vt : Nat}

/-- No erasure: every current branch is in the new version. -/
theorem commit_keeps (hb : b ∈ r.branches) : b ∈ (commit r chain).branches := by
  simp [commit, hb]

theorem chainBranches_sub : ∀ {nx prev} {chain : List (Cond A × I)},
    ∀ b ∈ chainBranches vt nx prev chain, b.selects sat v = true → ∀ c ∈ prev, c.eval sat v = false
  | _, _, [], b, hb, _ => by simp [chainBranches] at hb
  | nx, prev, (c, i) :: rest, b, hb, hv => by
    simp only [chainBranches, List.mem_cons] at hb
    rcases hb with rfl | hb
    · simp only [Branch.selects, Cond.eval, Bool.and_eq_true] at hv
      exact notAny_eval.mp hv.2
    · intro c' hc'
      exact chainBranches_sub b hb hv c' (by simp [hc'])

/-- One voter, one vote: at most one of a chain's branches selects any
values. -/
theorem chain_one_vote : ∀ {nx prev} {chain : List (Cond A × I)},
    (selected sat (chainBranches vt nx prev chain) v).length ≤ 1
  | _, _, [] => by simp [chainBranches, selected]
  | nx, prev, (c, i) :: rest => by
    simp only [chainBranches, selected, List.filter_cons]
    split
    · rename_i h
      have hc : c.eval sat v = true := by
        simp only [Branch.selects, Cond.eval, Bool.and_eq_true] at h; exact h.1
      rw [List.filter_eq_nil_iff.mpr (fun b hb hv => by
        have := chainBranches_sub b hb (by simpa using hv) c (by simp)
        rw [hc] at this; cases this)]
      simp
    · exact chain_one_vote

theorem chainBranches_witness : ∀ {nx prev} {chain : List (Cond A × I)},
    (∀ c ∈ prev, c.eval sat v = false) →
      ∀ b ∈ chainBranches vt nx prev chain, b.selects sat v = true → chainSelect sat chain v = some b.instr
  | _, _, [], _, b, hb, _ => by simp [chainBranches] at hb
  | nx, prev, (c, i) :: rest, hprev, b, hb, hv => by
    simp only [chainBranches, List.mem_cons] at hb
    rcases hb with rfl | hb
    · simp only [Branch.selects, Cond.eval, Bool.and_eq_true] at hv
      simp [chainSelect, hv.1]
    · have hc := chainBranches_sub b hb hv c (by simp)
      simp only [chainSelect, hc]
      exact chainBranches_witness (fun c' h' => by
        rcases List.mem_append.mp h' with h' | h'
        · exact hprev c' h'
        · simp at h'; exact h' ▸ hc) b hb hv

theorem chainBranches_covers : ∀ {nx prev} {chain : List (Cond A × I)},
    (∀ c ∈ prev, c.eval sat v = false) → chainSelect sat chain v = some i →
      ∃ b ∈ chainBranches vt nx prev chain, b.selects sat v = true
  | _, _, [], _, h => by simp [chainSelect] at h
  | nx, prev, (c, i') :: rest, hprev, h => by
    simp only [chainSelect] at h
    split at h
    · rename_i hc
      simp only [chainBranches, List.mem_cons]
      exact ⟨_, .inl rfl, by simp [Branch.selects, Cond.eval, hc, notAny_eval.mpr hprev]⟩
    · rename_i hc
      simp only [chainBranches, List.mem_cons]
      exact (chainBranches_covers (prev := prev ++ [c]) (fun c' h' => by
        rcases List.mem_append.mp h' with h' | h'
        · exact hprev c' h'
        · simp at h'; simpa [h'] using hc) h).imp fun _ ⟨hb, hv⟩ => ⟨.inr hb, hv⟩

end

/-! ## Tally

Each voter whose branch selects the values casts that branch's
instruction; the votes cast are the ballot. They are grouped by key, at the
finest level first: `group k i` is instruction `i`'s key at level `k`.
A group wins a level when its lead over every other group, and over none,
clears the margin `required`, which the inhibition `ι` sets. The tally then
runs the exact instruction in the group with the most votes, ties to the
oldest. With no winner at any level, it is a conflict. -/

/-- How a base tallies: each instruction's key at a level, the levels tried
finest first and down to the floor, and the inhibition `ι`, in thousandths
(1000 is 1). -/
structure Policy (I K : Type) where
  group : Nat → I → K
  levels : List Nat
  inhibition : Nat

/-- The margin a winner needs among `cast` votes: `⌈(1 − ι) · cast⌉`, never
below 1. `ι = 1` is a plurality; `ι = 0` is every vote cast. -/
def required (ι cast : Nat) : Nat := max 1 (((1000 - ι) * cast + 999) / 1000)

theorem one_le_required : 1 ≤ required ι c := Nat.le_max_left _ _

theorem required_le (h : 0 < c) : required ι c ≤ c := by
  have : (1000 - ι) * c ≤ 1000 * c := Nat.mul_le_mul_right _ (Nat.sub_le _ _)
  unfold required
  generalize (1000 - ι) * c = a at *
  omega

theorem required_anti (h : ι ≤ ι') : required ι' c ≤ required ι c := by
  have : (1000 - ι') * c ≤ (1000 - ι) * c := Nat.mul_le_mul_right _ (by omega)
  unfold required
  generalize (1000 - ι') * c = a at *
  generalize (1000 - ι) * c = b at *
  omega

theorem required_zero (h : 0 < c) : required 0 c = c := by
  simp only [required, Nat.sub_zero]; omega

section
variable {K : Type} [DecidableEq K] (p : Policy I K)

/-- The votes in group `g` at level `k`. -/
def groupVotes (k : Nat) (ballot : List (Branch A I)) (g : K) : Nat :=
  (ballot.filter fun y => p.group k y.instr = g).length

/-- `x`'s group wins level `k`: it has the margin, and leads every other
group by it. -/
def wins (k : Nat) (ballot : List (Branch A I)) (x : Branch A I) : Bool :=
  let req := required p.inhibition ballot.length
  let n := groupVotes p k ballot (p.group k x.instr)
  decide (req ≤ n) && ballot.all fun y =>
    decide (p.group k y.instr = p.group k x.instr) || decide (groupVotes p k ballot (p.group k y.instr) + req ≤ n)

variable [DecidableEq I]

/-- The votes for `x`'s exact instruction. -/
def exactVotes (ballot : List (Branch A I)) (x : Branch A I) : Nat :=
  (ballot.filter fun y => y.instr = x.instr).length

/-- `x` is ahead of `y`: more votes for its exact instruction, or as many
and older. -/
def ahead (ballot : List (Branch A I)) (x y : Branch A I) : Bool :=
  decide (exactVotes ballot y < exactVotes ballot x) ||
    (decide (exactVotes ballot x = exactVotes ballot y) && decide (x.rank < y.rank))

/-- The vote furthest ahead. -/
def best (ballot : List (Branch A I)) : List (Branch A I) → Option (Branch A I)
  | [] => none
  | x :: xs => match best ballot xs with
    | none => some x
    | some y => if ahead ballot y x then some y else some x

def tallyAt (ballot : List (Branch A I)) : List Nat → Outcome I
  | [] => .conflict
  | k :: ks => match best ballot (ballot.filter (wins p k ballot)) with
    | some x => .select x.instr
    | none => tallyAt ballot ks

/-- `null` when no voter selects the values; otherwise the instruction the
first level with a winner runs, or a conflict. -/
def tally (sat : A → V → Bool) (bs : List (Branch A I)) (v : V) : Outcome I :=
  match selected sat bs v with
  | [] => .null
  | b :: rest => tallyAt p (b :: rest) p.levels

variable {p} {sat : A → V → Bool} {bs bs' ballot ballot' l : List (Branch A I)}
  {x y z : Branch A I} {v : V} {i : I} {k : Nat}

theorem ahead_irrefl : ahead ballot x x = false := by simp [ahead]

theorem ahead_asymm (h : ahead ballot x y = true) : ahead ballot y x = false := by
  simp only [ahead, Bool.or_eq_true, Bool.and_eq_true, decide_eq_true_eq, Bool.or_eq_false_iff,
    Bool.and_eq_false_iff, decide_eq_false_iff_not] at *
  omega

theorem not_ahead_trans (h₁ : ahead ballot x y = false) (h₂ : ahead ballot y z = false) :
    ahead ballot x z = false := by
  simp only [ahead, Bool.or_eq_false_iff, Bool.and_eq_false_iff, decide_eq_false_iff_not] at *
  omega

theorem best_mem : ∀ {l : List (Branch A I)} {x}, best ballot l = some x → x ∈ l
  | [], _, h => by simp [best] at h
  | z :: zs, x, h => by
    unfold best at h
    split at h
    · cases h; simp
    · split at h
      · cases h; rename_i w hw _; exact List.mem_cons_of_mem _ (best_mem hw)
      · cases h; simp

theorem best_none : ∀ {l : List (Branch A I)}, best ballot l = none ↔ l = []
  | [] => by simp [best]
  | z :: zs => by unfold best; split <;> (try split) <;> simp

/-- Nothing in the list is ahead of the best. -/
theorem best_max : ∀ {l : List (Branch A I)} {x}, best ballot l = some x →
    ∀ y ∈ l, ahead ballot y x = false
  | [], _, h, _, _ => by simp [best] at h
  | z :: zs, x, h, y, hy => by
    unfold best at h
    split at h
    · rename_i hn
      cases h
      have := best_none.mp hn; subst this
      simp at hy; subst hy; exact ahead_irrefl
    · rename_i w hw
      split at h
      · rename_i hwz
        cases h
        rcases List.mem_cons.mp hy with rfl | hy
        · exact ahead_asymm hwz
        · exact best_max hw y hy
      · rename_i hwz
        cases h
        rcases List.mem_cons.mp hy with rfl | hy
        · exact ahead_irrefl
        · exact not_ahead_trans (best_max hw y hy) (by simpa using hwz)

/-- Two votes neither ahead of the other are one, ranks being distinct. -/
theorem ahead_eq (hr : (ballot.map (·.rank)).Nodup) (hx : x ∈ ballot) (hy : y ∈ ballot)
    (h₁ : ahead ballot x y = false) (h₂ : ahead ballot y x = false) : x = y := by
  have : x.rank = y.rank := by
    simp only [ahead, Bool.or_eq_false_iff, Bool.and_eq_false_iff, decide_eq_false_iff_not] at h₁ h₂
    omega
  exact nodup_map_inj hr hx hy this

omit [DecidableEq I] in
theorem groupVotes_perm (h : ballot.Perm ballot') (g : K) :
    groupVotes p k ballot g = groupVotes p k ballot' g := (h.filter _).length_eq

theorem exactVotes_perm (h : ballot.Perm ballot') (x : Branch A I) :
    exactVotes ballot x = exactVotes ballot' x := (h.filter _).length_eq

omit [DecidableEq I] in
theorem all_perm (h : ballot.Perm ballot') (f : Branch A I → Bool) : ballot.all f = ballot'.all f := by
  apply Bool.eq_iff_iff.mpr
  simp only [List.all_eq_true]
  exact ⟨fun H y hy => H y (h.mem_iff.mpr hy), fun H y hy => H y (h.mem_iff.mp hy)⟩

omit [DecidableEq I] in
theorem wins_perm (h : ballot.Perm ballot') : wins p k ballot' = wins p k ballot := by
  funext x
  simp only [wins, groupVotes_perm h, h.length_eq, all_perm h]

theorem best_ballot_perm (h : ballot.Perm ballot') : ∀ {l : List (Branch A I)}, best ballot' l = best ballot l
  | [] => rfl
  | z :: zs => by
    have e : ahead ballot' = ahead ballot := by
      funext a c; simp only [ahead, exactVotes_perm h]
    simp only [best, best_ballot_perm h (l := zs), e]

omit [DecidableEq I] in
/-- A group wins a level only when no other does. -/
theorem wins_unique (hx : x ∈ ballot) (hy : y ∈ ballot) (wx : wins p k ballot x = true)
    (wy : wins p k ballot y = true) : p.group k x.instr = p.group k y.instr := by
  apply Decidable.byContradiction
  intro hne
  simp only [wins, Bool.and_eq_true, decide_eq_true_eq, List.all_eq_true, Bool.or_eq_true] at wx wy
  have r := one_le_required (ι := p.inhibition) (c := ballot.length)
  rcases wx.2 y hy with h1 | h1
  · exact hne h1.symm
  rcases wy.2 x hx with h2 | h2
  · exact hne h2
  omega

/-- Ballots are unordered: a tally does not depend on the order of the
votes, ranks being distinct. -/
theorem tallyAt_perm (hr : (ballot.map (·.rank)).Nodup) (h : ballot.Perm ballot') :
    ∀ ls, tallyAt p ballot ls = tallyAt p ballot' ls
  | [] => rfl
  | k :: ks => by
    simp only [tallyAt]
    rw [wins_perm h, best_ballot_perm h]
    have hc : (ballot.filter (wins p k ballot)).Perm (ballot'.filter (wins p k ballot)) := h.filter _
    cases hx : best ballot (ballot.filter (wins p k ballot)) with
    | none =>
      have e : ballot'.filter (wins p k ballot) = [] := by
        rw [best_none.mp hx] at hc; exact hc.symm.eq_nil
      rw [e]
      simp only [best]
      exact tallyAt_perm hr h ks
    | some x =>
      have hmx := best_mem hx
      cases hy : best ballot (ballot'.filter (wins p k ballot)) with
      | none =>
        rw [best_none.mp hy] at hc
        rw [hc.eq_nil] at hmx; cases hmx
      | some y =>
        have hmy := best_mem hy
        have hmx' := hc.mem_iff.mp hmx
        have hmy' := hc.mem_iff.mpr hmy
        have := ahead_eq hr (List.mem_filter.mp hmx).1 (List.mem_filter.mp hmy').1
          (best_max hy x hmx') (best_max hx y hmy')
        simp [this]

theorem tally_perm (hr : (bs.map (·.rank)).Nodup) (h : bs.Perm bs') :
    tally p sat bs v = tally p sat bs' v := by
  have hs := selected_perm (sat := sat) (v := v) h
  have hrs : ((selected sat bs v).map (·.rank)).Nodup := (List.filter_sublist.map _).nodup hr
  unfold tally
  cases e : selected sat bs v with
  | nil => rw [e] at hs; rw [hs.symm.eq_nil]
  | cons b rest =>
    rw [e] at hs hrs
    cases e' : selected sat bs' v with
    | nil => rw [e'] at hs; exact absurd hs.length_eq (by simp)
    | cons b' rest' => rw [e'] at hs; exact tallyAt_perm hrs hs _

theorem tallyAt_ne_null : ∀ {ls}, tallyAt p ballot ls ≠ .null
  | [] => by simp [tallyAt]
  | k :: ks => by
    simp only [tallyAt]
    cases best ballot (ballot.filter (wins p k ballot)) with
    | some x => simp
    | none => exact tallyAt_ne_null

theorem tally_null : tally p sat bs v = .null ↔ selected sat bs v = [] := by
  unfold tally
  cases selected sat bs v with
  | nil => simp
  | cons b rest => exact ⟨fun h => absurd h tallyAt_ne_null, fun h => by cases h⟩

theorem tallyAt_conflict : ∀ {ls}, tallyAt p ballot ls = .conflict ↔
    ∀ k ∈ ls, ∀ x ∈ ballot, wins p k ballot x = false
  | [] => by simp [tallyAt]
  | k :: ks => by
    simp only [tallyAt]
    cases hx : best ballot (ballot.filter (wins p k ballot)) with
    | some x =>
      have hm := List.mem_filter.mp (best_mem hx)
      constructor
      · intro h; cases h
      · intro H; have := H k (by simp) x hm.1; rw [hm.2] at this; cases this
    | none =>
      have he := List.filter_eq_nil_iff.mp (best_none.mp hx)
      simp only
      rw [tallyAt_conflict]
      constructor
      · intro H k' hk' x hx'
        rcases List.mem_cons.mp hk' with rfl | hk'
        · simpa using he x hx'
        · exact H k' hk' x hx'
      · intro H k' hk' x hx'; exact H k' (by simp [hk']) x hx'

theorem tallyAt_select : ∀ {ls}, tallyAt p ballot ls = .select i →
    ∃ k ∈ ls, ∃ x ∈ ballot, x.instr = i ∧ wins p k ballot x = true ∧
      ∀ y ∈ ballot, wins p k ballot y = true → p.group k y.instr = p.group k i
  | [], h => by simp [tallyAt] at h
  | k :: ks, h => by
    simp only [tallyAt] at h
    cases hx : best ballot (ballot.filter (wins p k ballot)) with
    | some x =>
      simp only [hx, Outcome.select.injEq] at h
      subst h
      have hm := List.mem_filter.mp (best_mem hx)
      exact ⟨k, by simp, x, hm.1, rfl, hm.2, fun y hy wy => wins_unique hy hm.1 wy hm.2⟩
    | none =>
      simp only [hx] at h
      obtain ⟨k', hk', rest⟩ := tallyAt_select h
      exact ⟨k', by simp [hk'], rest⟩

/-- The floor holds: the instruction run is a vote of the group that won,
at a level the policy tries, and every vote that won with it shares its
key there. With levels no coarser than the floor, and keys that agree at
every coarser level, no vote counted for it differs from it above the
floor. -/
theorem tally_group (h : tally p sat bs v = .select i) :
    ∃ k ∈ p.levels, ∃ x ∈ selected sat bs v, x.instr = i ∧ wins p k (selected sat bs v) x = true ∧
      ∀ y ∈ selected sat bs v, wins p k (selected sat bs v) y = true → p.group k y.instr = p.group k i := by
  unfold tally at h
  cases e : selected sat bs v with
  | nil => rw [e] at h; cases h
  | cons b rest => rw [e] at h; exact tallyAt_select h

/-- Unanimity: when every vote cast is one instruction, it wins, at any
inhibition. -/
theorem tally_unanimous (hne : selected sat bs v ≠ []) (hall : ∀ b ∈ selected sat bs v, b.instr = i)
    (hl : p.levels ≠ []) : tally p sat bs v = .select i := by
  unfold tally
  cases e : selected sat bs v with
  | nil => exact absurd e hne
  | cons b0 rest =>
    rw [e] at hall
    obtain ⟨k, ks, hk⟩ : ∃ k ks, p.levels = k :: ks := by
      cases h : p.levels with
      | nil => exact absurd h hl
      | cons k ks => exact ⟨k, ks, rfl⟩
    rw [hk]
    simp only [tallyAt]
    have hw : (b0 :: rest).filter (wins p k (b0 :: rest)) = b0 :: rest := by
      apply List.filter_eq_self.mpr
      intro x hx
      have hcount : groupVotes p k (b0 :: rest) (p.group k x.instr) = (b0 :: rest).length := by
        unfold groupVotes
        rw [List.filter_eq_self.mpr]
        intro y hy; simp [hall y hy, hall x hx]
      simp only [wins, hcount, Bool.and_eq_true, decide_eq_true_eq, List.all_eq_true, Bool.or_eq_true]
      exact ⟨required_le (by simp), fun y hy => .inl (by rw [hall y hy, hall x hx])⟩
    rw [hw]
    cases hb : best (b0 :: rest) (b0 :: rest) with
    | none => exact absurd (best_none.mp hb) (by simp)
    | some x => simp [hall x (best_mem hb)]

omit [DecidableEq I] in
theorem wins_mono {q : Policy I K} (hg : q.group = p.group) (h : p.inhibition ≤ q.inhibition)
    (w : wins p k ballot x = true) : wins q k ballot x = true := by
  obtain ⟨g, ls, ι⟩ := p
  obtain ⟨g', ls', ι'⟩ := q
  simp only at hg h; subst hg
  simp only [wins, groupVotes, Bool.and_eq_true, List.all_eq_true, Bool.or_eq_true] at w ⊢
  have r := required_anti (c := ballot.length) h
  obtain ⟨w1, w2⟩ := w
  have w1 := of_decide_eq_true w1
  exact ⟨decide_eq_true (by omega), fun y hy => (w2 y hy).imp id fun hy' =>
    decide_eq_true (by have := of_decide_eq_true hy'; omega)⟩

/-- Deciding is monotone in the inhibition: a tally that decides at `ι`
decides at every higher `ι'`. What it decides may differ, a finer level
deciding first (proposal 59, open question 6). -/
theorem tally_decides_mono (h : ι ≤ ι') (d : tally { p with inhibition := ι } sat bs v ≠ .conflict) :
    tally { p with inhibition := ι' } sat bs v ≠ .conflict := by
  unfold tally at *
  cases e : selected sat bs v with
  | nil => simp
  | cons b rest =>
    rw [e] at d
    intro hc
    apply d
    rw [tallyAt_conflict] at hc ⊢
    intro k hk x hx
    cases hw : wins { p with inhibition := ι } k (b :: rest) x
    · rfl
    · have := wins_mono (p := { p with inhibition := ι }) (q := { p with inhibition := ι' }) rfl h hw
      rw [hc k hk x hx] at this; cases this

theorem filter_full {q : α → Bool} {l : List α} (h : l.length ≤ (l.filter q).length) :
    ∀ a ∈ l, q a = true :=
  List.filter_eq_self.mp
    ((List.filter_sublist (p := q) (l := l)).eq_of_length (Nat.le_antisymm (List.length_filter_le _ _) h))

end

/-- `distinguish` is a tally: the exact instruction as the key, one level,
and `ι = 0`, every vote cast. -/
theorem distinguish_tally [DecidableEq I] {sat : A → V → Bool} {bs : List (Branch A I)} {v : V} (k : Nat) :
    tally ⟨fun _ i => i, [k], 0⟩ sat bs v = distinguish sat bs v := by
  cases hd : distinguish sat bs v with
  | null => exact tally_null.mpr (distinguish_null.mp hd)
  | select i =>
    obtain ⟨ne, all⟩ := distinguish_select.mp hd
    exact tally_unanimous ne all (by simp)
  | conflict =>
    unfold tally
    cases e : selected sat bs v with
    | nil => rw [distinguish_null.mpr e] at hd; cases hd
    | cons b rest =>
      apply tallyAt_conflict.mpr
      intro k' _ x hx
      cases hw : wins ⟨fun _ i => i, [k], 0⟩ k' (b :: rest) x
      · rfl
      · exfalso
        have hreq : required 0 (b :: rest).length = (b :: rest).length := required_zero (by simp)
        simp only [wins, groupVotes, hreq, Bool.and_eq_true, decide_eq_true_eq] at hw
        have hall : ∀ y ∈ selected sat bs v, y.instr = x.instr := by
          rw [e]; intro y hy; simpa using filter_full hw.1 y hy
        have := distinguish_select.mpr ⟨by rw [e]; simp, hall⟩
        rw [this] at hd; cases hd

/-- What triggered `startle` still runs the new instruction: if no current
branch selected the values and the model's chain covers them, the new
chain is the only vote, and the committed version selects its
instruction. -/
theorem commit_witness {K : Type} [DecidableEq I] [DecidableEq K] {p : Policy I K} {sat : A → V → Bool}
    {r : Refinement A I} {chain : List (Cond A × I)} {v : V} {i : I}
    (hnull : selected sat r.branches v = []) (hc : chainSelect sat chain v = some i)
    (hl : p.levels ≠ []) : tally p sat (commit r chain).branches v = .select i := by
  have olds : ∀ o ∈ r.branches, o.selects sat v = false := by
    intro o ho
    cases h : o.selects sat v
    · rfl
    · have : o ∈ selected sat r.branches v := List.mem_filter.mpr ⟨ho, h⟩
      rw [hnull] at this; cases this
  refine tally_unanimous ?_ ?_ hl
  · obtain ⟨b, hb, hv⟩ := chainBranches_covers (vt := r.next) (nx := r.next) (prev := []) (by simp) hc
    intro e
    have : b ∈ selected sat (commit r chain).branches v :=
      List.mem_filter.mpr ⟨by simp [commit, hb], hv⟩
    rw [e] at this; cases this
  · intro b' hb'
    obtain ⟨hb', hv'⟩ := List.mem_filter.mp hb'
    simp only [commit, List.mem_append] at hb'
    rcases hb' with hb' | hb'
    · rw [olds b' hb'] at hv'; cases hv'
    · have := chainBranches_witness (prev := []) (by simp) b' hb' hv'
      rw [hc] at this; exact (Option.some.inj this).symm

/-! ## After `startle` -/

inductive After (I : Type) where
  | startleAgain
  /-- Run `i`; `learned` says whether the version committed selects it for
  these values, or it runs once, unlearned. -/
  | trust (i : I) (learned : Bool)
  deriving Repr, DecidableEq

def After.instr : After I → Option I
  | .startleAgain => none
  | .trust i _ => some i

/-- The step evaluates over the root it committed. It runs the instruction
its own model's chain gives these values, whatever that root's tally
selects: other voters may outvote it, and the step runs its own
instruction once, unlearned. -/
def afterStartle {K : Type} [DecidableEq I] [DecidableEq K] (p : Policy I K) (sat : A → V → Bool)
    (chain : List (Cond A × I)) (committed : List (Branch A I)) (v : V) : After I :=
  match chainSelect sat chain v with
  | none => .startleAgain
  | some i => .trust i (tally p sat committed v == .select i)

/-- What the step runs does not depend on what any other commit stored. -/
theorem afterStartle_own {K : Type} [DecidableEq I] [DecidableEq K] {p : Policy I K}
    {sat : A → V → Bool} (c₁ c₂ : List (Branch A I)) :
    (afterStartle p sat chain c₁ v).instr = (afterStartle p sat chain c₂ v).instr := by
  unfold afterStartle; split <;> rfl

theorem afterStartle_instr {K : Type} [DecidableEq I] [DecidableEq K] {p : Policy I K}
    {sat : A → V → Bool} {c : List (Branch A I)} :
    (afterStartle p sat chain c v).instr = chainSelect sat chain v := by
  unfold afterStartle; split <;> simp_all [After.instr]

/-! ## Transitions and the transition bound -/

/-- The states of a transition: everything the system does with one
instruction. `tally` is a count that did not settle on the instruction
run: a conflict, or a startle's commit, counted with the votes already
there. -/
inductive State where
  | reject | hesitate | startle | tally | trust
  deriving Repr, DecidableEq

/-- A transition is the life of one instruction: the states it passed
through. It ends in `trust` running the instruction, exactly once; the
outcome starts the next transition. -/
abbrev Transition := List State

def Transition.WellFormed (t : Transition) : Prop :=
  ∃ pre : List State, t = pre ++ [.trust] ∧ State.trust ∉ pre

/-- A well-formed transition runs exactly one instruction. -/
theorem Transition.WellFormed.runs_one {t : Transition} (h : t.WellFormed) :
    List.count State.trust t = 1 := by
  obtain ⟨pre, rfl, hpre⟩ := h
  rw [List.count_append, List.count_eq_zero.mpr hpre]
  simp

/-- So the instructions a frame runs are exactly its transitions: the transition bound
bounds its effects. -/
theorem runs_eq_length : ∀ {ts : List Transition}, (∀ t ∈ ts, t.WellFormed) →
    (ts.map (List.count State.trust)).sum = ts.length
  | [], _ => rfl
  | t :: ts, h => by
    simp only [List.map_cons, List.sum_cons, List.length_cons]
    rw [(h t (by simp)).runs_one, runs_eq_length (fun t' h' => h t' (by simp [h']))]
    omega

inductive Ending where
  | done
  | exhausted
  deriving Repr, DecidableEq

/-- Run a frame under its transition bound: `step` gives the next transition, or none
when the frame is done. Every transition counts, whether or not it calls
the model. -/
def runFrame (step : S → Option (Transition × S)) : Nat → S → List Transition × Ending × S
  | 0, s => match step s with
    | none => ([], .done, s)
    | some _ => ([], .exhausted, s)
  | n + 1, s => match step s with
    | none => ([], .done, s)
    | some (t, s') =>
      let x := runFrame step n s'
      (t :: x.1, x.2)

/-- What a frame returns to its caller: its outcome, or, exhausted, the
error observation, saying after how many transitions. The caller receives
it as any outcome: `null` at its base, so its model is startled. -/
inductive Return (S : Type) where
  | outcome (s : S)
  | error (transitions : Nat)
  deriving Repr, DecidableEq

def frameReturn (x : List Transition × Ending × S) : Return S :=
  match x.2.1 with
  | .done => .outcome x.2.2
  | .exhausted => .error x.1.length

theorem runFrame_bound {step : S → Option (Transition × S)} :
    ∀ {n s}, (runFrame step n s).1.length ≤ n
  | 0, s => by unfold runFrame; split <;> simp
  | n + 1, s => by
    unfold runFrame; split
    · simp
    · exact Nat.succ_le_succ runFrame_bound

/-- A frame ends either done, with nothing left to take, or exhausted,
having taken exactly its transition bound of transitions. -/
theorem runFrame_ends {step : S → Option (Transition × S)} :
    ∀ {n s}, let x := runFrame step n s;
      (x.2.1 = .done ∧ step x.2.2 = none) ∨ (x.2.1 = .exhausted ∧ x.1.length = n)
  | 0, s => by unfold runFrame; split <;> simp_all
  | n + 1, s => by
    unfold runFrame; split
    · simp_all
    · rcases runFrame_ends (step := step) (n := n) (s := by assumption) with h | h
      · exact .inl h
      · exact .inr ⟨h.1, by simp [h.2]⟩

/-- A learned cycle that never calls the model still ends: a frame that
always has a next transition is exhausted after exactly its transition bound of them,
and returns the error. -/
theorem runFrame_cycle {step : S → Option (Transition × S)}
    (h : ∀ s, ∃ t s', step s = some (t, s')) :
    ∀ {n s}, (runFrame step n s).2.1 = .exhausted ∧ (runFrame step n s).1.length = n
  | 0, s => by
    obtain ⟨t, s', e⟩ := h s; unfold runFrame; rw [e]; simp
  | n + 1, s => by
    obtain ⟨t, s', e⟩ := h s
    unfold runFrame; rw [e]
    have ih := runFrame_cycle h (n := n) (s := s')
    simp [ih.1, ih.2]

theorem runFrame_cycle_error {step : S → Option (Transition × S)}
    (h : ∀ s, ∃ t s', step s = some (t, s')) : frameReturn (runFrame step n s) = .error n := by
  obtain ⟨e, l⟩ := runFrame_cycle h (n := n) (s := s)
  simp [frameReturn, e, l]

end Refine
