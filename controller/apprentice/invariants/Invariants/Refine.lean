/-! Proposal 58: one base's versioned refinement. A refinement holds
branches, each a condition (CEL, abstracted here as `Cond`) selecting an
instruction. The model writes a guard, a chain of conditions, and never sees
the existing branches; the system stores the chain as independent branches,
narrows each new branch where it meets an older one with a different
instruction (the older keeps the overlap), and selects. A transition is
one instruction's life, ending in `trust`; a frame takes at most its transition bound
of them, and an exhausted frame returns an error.

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
orders it: a lower rank is older. -/
structure Branch (A I : Type) where
  id : Nat
  rank : Nat
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

omit [DecidableEq I] in
/-- Retiring a branch never widens another: what the rest select is
unchanged. -/
theorem retire_selected (k : Nat) :
    selected sat (bs.filter (·.id != k)) v = (selected sat bs v).filter (·.id != k) := by
  simp only [selected, List.filter_filter, Bool.and_comm]

end

/-! ## Reconciling: the older keeps the overlap -/

/-- The newer branch's remainder outside the older: a new local id, the
newer's rank and instruction. -/
def remainder (o n : Branch A I) (id : Nat) : Branch A I :=
  { id, rank := n.rank, cond := .and n.cond (.not o.cond), instr := n.instr }

/-- Retire the newer branch `n` by local id and add its remainder. -/
def reconcile (r : Refinement A I) (o n : Branch A I) : Refinement A I :=
  { next := r.next + 1,
    branches := r.branches.filter (·.id != n.id) ++ [remainder o n r.next] }

section
variable {sat : A → V → Bool} {r : Refinement A I} {o n b : Branch A I} {v : V}

theorem remainder_rank : (remainder o n k).rank = n.rank := rfl
theorem remainder_instr : (remainder o n k).instr = n.instr := rfl

/-- In the overlap the older keeps the values: the remainder never selects
what the older selects. -/
theorem remainder_disjoint (h : o.selects sat v = true) : (remainder o n k).selects sat v = false := by
  simp_all [remainder, Branch.selects, Cond.eval]

/-- Outside the older, the newer's selection is kept. -/
theorem remainder_outside (hn : n.selects sat v = true) (ho : o.selects sat v = false) :
    (remainder o n k).selects sat v = true := by
  simp_all [remainder, Branch.selects, Cond.eval]

theorem remainder_sub (h : (remainder o n k).selects sat v = true) : n.selects sat v = true := by
  simp_all [remainder, Branch.selects, Cond.eval]

/-- No erasure: every branch but the retired one is carried unchanged. -/
theorem reconcile_keeps (hb : b ∈ r.branches) (hid : b.id ≠ n.id) :
    b ∈ (reconcile r o n).branches := by
  simp [reconcile, hb, hid]

/-- The values that triggered `startle` were `null`, so no older branch
selects them; they lie in the remainder and still select the new
instruction. -/
theorem witness_kept (hn : n.selects sat v = true) (hnull : selected sat r.branches v = [])
    (ho : o ∈ r.branches) : (remainder o n k).selects sat v = true := by
  apply remainder_outside hn
  cases h : o.selects sat v
  · rfl
  · have : o ∈ selected sat r.branches v := List.mem_filter.mpr ⟨ho, h⟩
    rw [hnull] at this; cases this

theorem reconcile_wellFormed (hr : r.WellFormed) (hn : n ∈ r.branches) :
    (reconcile r o n).WellFormed := by
  obtain ⟨lt, ids, ranks⟩ := hr
  have keep : ∀ b ∈ r.branches.filter (·.id != n.id), b ∈ r.branches ∧ b.id ≠ n.id := by
    intro b hb; simpa using List.mem_filter.mp hb
  refine ⟨?_, ?_, ?_⟩
  · intro b hb
    simp only [reconcile, List.mem_append, List.mem_singleton] at hb
    rcases hb with hb | rfl
    · obtain ⟨hb, -⟩ := keep b hb
      exact ⟨Nat.lt_succ_of_lt (lt b hb).1, (lt b hb).2⟩
    · exact ⟨Nat.lt_succ_self _, Nat.le_of_lt (Nat.lt_of_le_of_lt (lt n hn).2 (lt n hn).1)⟩
  · simp only [reconcile, List.map_append, List.map_cons, List.map_nil]
    rw [List.nodup_append]
    refine ⟨(List.filter_sublist).map _ |>.nodup ids, by simp, ?_⟩
    intro a ha c hc
    simp only [List.mem_singleton] at hc
    subst hc
    obtain ⟨x, hx, rfl⟩ := List.mem_map.mp ha
    exact Nat.ne_of_lt (lt x (keep x hx).1).1
  · simp only [reconcile, List.map_append, List.map_cons, List.map_nil]
    rw [List.nodup_append]
    refine ⟨(List.filter_sublist).map _ |>.nodup ranks, by simp, ?_⟩
    intro a ha c hc
    simp only [List.mem_singleton] at hc
    subst hc
    obtain ⟨x, hx, rfl⟩ := List.mem_map.mp ha
    obtain ⟨hx, hid⟩ := keep x hx
    intro e
    exact hid (congrArg Branch.id (nodup_map_inj ranks hx hn e))

end

/-! ## Commit: the model's chain, narrowed against the older branches -/

/-- The branches of a chain `c₀ ? i₀ : c₁ ? i₁ : … : null`, each with its
effective condition (its own, and no earlier one in the chain), local ids
and ranks from `next`. -/
def chainBranches : Nat → List (Cond A) → List (Cond A × I) → List (Branch A I)
  | _, _, [] => []
  | next, prev, (c, i) :: rest =>
    { id := next, rank := next, cond := .and c (notAny prev), instr := i } ::
      chainBranches (next + 1) (prev ++ [c]) rest

/-- What the chain itself selects: its first condition that holds. -/
def chainSelect (sat : A → V → Bool) : List (Cond A × I) → V → Option I
  | [], _ => none
  | (c, i) :: rest, v => if c.eval sat v then some i else chainSelect sat rest v

/-- One reconcile: the older branch, the retired one, and the remainder. -/
structure Record where
  older : Nat
  retired : Nat
  remainder : Nat
  deriving Repr, DecidableEq

structure Narrowed (A I : Type) where
  branch : Branch A I
  next : Nat
  records : List Record
  possible : List (Nat × Nat)

/-- Narrow a new branch against each older one with a different
instruction. `overlaps` decides an overlap (`some`) or cannot (`none`):
an undecided pair is recorded as possibly overlapping. -/
def narrow [DecidableEq I] (overlaps : Cond A → Cond A → Option Bool) :
    List (Branch A I) → Nat → Branch A I → Narrowed A I
  | [], next, cur => ⟨cur, next, [], []⟩
  | o :: os, next, cur =>
    if o.instr = cur.instr then narrow overlaps os next cur
    else match overlaps o.cond cur.cond with
      | some true =>
        let m := narrow overlaps os (next + 1) (remainder o cur next)
        { m with records := ⟨o.id, cur.id, next⟩ :: m.records }
      | some false => narrow overlaps os next cur
      | none =>
        let m := narrow overlaps os next cur
        { m with possible := (o.id, cur.id) :: m.possible }

structure Committed (A I : Type) where
  branches : List (Branch A I)
  next : Nat
  records : List Record
  possible : List (Nat × Nat)

def commitChain [DecidableEq I] (overlaps : Cond A → Cond A → Option Bool)
    (olds : List (Branch A I)) : Nat → List (Cond A) → List (Cond A × I) → Committed A I
  | next, _, [] => ⟨[], next, [], []⟩
  | next, prev, (c, i) :: rest =>
    let m := narrow overlaps olds (next + 1)
      { id := next, rank := next, cond := .and c (notAny prev), instr := i }
    let k := commitChain overlaps olds m.next (prev ++ [c]) rest
    ⟨m.branch :: k.branches, k.next, m.records ++ k.records, m.possible ++ k.possible⟩

/-- `startle`'s commit: the new version is the current branches, unchanged,
plus the chain's branches, each already narrowed. The raw new branch is
retired and its remainder added at once. -/
def commit [DecidableEq I] (overlaps : Cond A → Cond A → Option Bool) (r : Refinement A I)
    (chain : List (Cond A × I)) : Refinement A I × List Record × List (Nat × Nat) :=
  let k := commitChain overlaps r.branches r.next [] chain
  (⟨k.next, r.branches ++ k.branches⟩, k.records, k.possible)

section
variable [DecidableEq I] {sat : A → V → Bool} {ov : Cond A → Cond A → Option Bool}
  {olds : List (Branch A I)} {r : Refinement A I} {b : Branch A I}
  {chain : List (Cond A × I)} {i : I} {v : V}

/-- No erasure: every current branch is in the new version. -/
theorem commit_keeps (hb : b ∈ r.branches) : b ∈ (commit ov r chain).1.branches := by
  simp [commit, hb]

theorem narrow_instr : ∀ {os : List (Branch A I)} {nx cur}, (narrow ov os nx cur).branch.instr = cur.instr
  | [], _, _ => rfl
  | o :: os, nx, cur => by
    unfold narrow
    split
    · exact narrow_instr
    · split
      · exact narrow_instr
      · exact narrow_instr
      · exact narrow_instr

/-- Narrowing only narrows. -/
theorem narrow_sub : ∀ {os : List (Branch A I)} {nx cur},
    (narrow ov os nx cur).branch.selects sat v = true → cur.selects sat v = true
  | [], _, _, h => h
  | o :: os, nx, cur, h => by
    unfold narrow at h
    split at h
    · exact narrow_sub h
    · split at h
      · exact remainder_sub (narrow_sub h)
      · exact narrow_sub h
      · exact narrow_sub h

/-- Narrowing keeps whatever no older branch selects. -/
theorem narrow_keeps : ∀ {os : List (Branch A I)} {nx cur},
    cur.selects sat v = true → (∀ o ∈ os, o.selects sat v = false) →
      (narrow ov os nx cur).branch.selects sat v = true
  | [], _, _, h, _ => h
  | o :: os, nx, cur, h, hos => by
    have ho := hos o (by simp)
    have hos' : ∀ o ∈ os, o.selects sat v = false := fun o' h' => hos o' (by simp [h'])
    unfold narrow
    split
    · exact narrow_keeps h hos'
    · split
      · exact narrow_keeps (remainder_outside h ho) hos'
      · exact narrow_keeps h hos'
      · exact narrow_keeps h hos'

/-- The overlaps decision is sound: `some false` means no value satisfies both. -/
def Sound (sat : A → V → Bool) (ov : Cond A → Cond A → Option Bool) : Prop :=
  ∀ p q, ov p q = some false → ∀ v, p.eval sat v = false ∨ q.eval sat v = false

/-- The older keeps the overlap: a narrowed branch selects a value an older
branch with a different instruction selects only where the pair was
recorded as possibly overlapping. -/
theorem narrow_older (hs : Sound sat ov) : ∀ {os : List (Branch A I)} {nx cur},
    ∀ o ∈ os, o.instr ≠ cur.instr → o.selects sat v = true →
      (narrow ov os nx cur).branch.selects sat v = true →
        ∃ k, (o.id, k) ∈ (narrow ov os nx cur).possible
  | [], _, _, o, ho, _, _, _ => by simp at ho
  | o' :: os, nx, cur, o, ho, hne, hv, hf => by
    by_cases heq : o'.instr = cur.instr
    · simp only [narrow, heq, ite_true] at hf ⊢
      rcases List.mem_cons.mp ho with rfl | ho
      · exact absurd heq hne
      · exact narrow_older hs o ho hne hv hf
    · simp only [narrow, heq, ite_false] at hf ⊢
      cases hov : ov o'.cond cur.cond with
      | none =>
        simp only [hov] at hf ⊢
        rcases List.mem_cons.mp ho with rfl | ho
        · exact ⟨cur.id, by simp⟩
        · obtain ⟨k, hk⟩ := narrow_older hs o ho hne hv hf
          exact ⟨k, by simp [hk]⟩
      | some bb =>
        cases bb
        · simp only [hov] at hf ⊢
          rcases List.mem_cons.mp ho with rfl | ho
          · rcases hs _ _ hov v with h | h
            · simp [Branch.selects, h] at hv
            · have := narrow_sub hf; simp [Branch.selects, h] at this
          · exact narrow_older hs o ho hne hv hf
        · simp only [hov] at hf ⊢
          rcases List.mem_cons.mp ho with rfl | ho
          · have := remainder_disjoint (n := cur) (k := nx) hv
            rw [narrow_sub hf] at this; cases this
          · exact narrow_older hs o ho (by rwa [remainder_instr]) hv hf

theorem commitChain_sub : ∀ {nx prev} {chain : List (Cond A × I)},
    ∀ b ∈ (commitChain ov olds nx prev chain).branches, b.selects sat v = true →
      ∀ c ∈ prev, c.eval sat v = false
  | _, _, [], b, hb, _ => by simp [commitChain] at hb
  | nx, prev, (c, i) :: rest, b, hb, hv => by
    simp only [commitChain, List.mem_cons] at hb
    rcases hb with rfl | hb
    · have := narrow_sub hv
      simp only [Branch.selects, Cond.eval, Bool.and_eq_true] at this
      exact (notAny_eval.mp this.2)
    · intro c' hc'
      exact commitChain_sub b hb hv c' (by simp [hc'])

/-- After `startle`, the values that were `null` select exactly the
instruction the model's chain gives them: the new branches are disjoint,
and narrowing keeps what no older branch selects. -/
theorem commitChain_witness (hnull : ∀ o ∈ olds, o.selects sat v = false) :
    ∀ {nx prev} {chain : List (Cond A × I)}, (∀ c ∈ prev, c.eval sat v = false) →
      ∀ b ∈ (commitChain ov olds nx prev chain).branches,
        b.selects sat v = true → chainSelect sat chain v = some b.instr
  | _, _, [], _, b, hb, _ => by simp [commitChain] at hb
  | nx, prev, (c, i) :: rest, hprev, b, hb, hv => by
    simp only [commitChain, List.mem_cons] at hb
    rcases hb with rfl | hb
    · have := narrow_sub hv
      simp only [Branch.selects, Cond.eval, Bool.and_eq_true] at this
      simp [chainSelect, this.1, narrow_instr]
    · have hc := commitChain_sub b hb hv c (by simp)
      simp only [chainSelect, hc]
      exact commitChain_witness hnull (fun c' h' => by
        rcases List.mem_append.mp h' with h' | h'
        · exact hprev c' h'
        · simp at h'; exact h' ▸ hc) b hb hv

theorem commitChain_covers (hnull : ∀ o ∈ olds, o.selects sat v = false) :
    ∀ {nx prev} {chain : List (Cond A × I)}, (∀ c ∈ prev, c.eval sat v = false) →
      chainSelect sat chain v = some i →
        ∃ b ∈ (commitChain ov olds nx prev chain).branches, b.selects sat v = true
  | _, _, [], _, h => by simp [chainSelect] at h
  | nx, prev, (c, i') :: rest, hprev, h => by
    simp only [chainSelect] at h
    split at h
    · rename_i hc
      simp only [commitChain, List.mem_cons]
      exact ⟨_, .inl rfl, narrow_keeps (by
        simp [Branch.selects, Cond.eval, hc, notAny_eval.mpr hprev]) hnull⟩
    · rename_i hc
      simp only [commitChain, List.mem_cons]
      exact (commitChain_covers hnull (nx := _) (prev := prev ++ [c]) (fun c' h' => by
        rcases List.mem_append.mp h' with h' | h'
        · exact hprev c' h'
        · simp at h'; simpa [h'] using hc) h).imp fun _ ⟨hb, hv⟩ => ⟨.inr hb, hv⟩

/-- The case that triggered `startle` still runs the new instruction: if no
current branch selected the values and the model's chain covers them, the
committed version selects the chain's instruction for them. -/
theorem commit_witness (hnull : selected sat r.branches v = [])
    (hc : chainSelect sat chain v = some i) : distinguish sat (commit ov r chain).1.branches v = .select i := by
  have olds : ∀ o ∈ r.branches, o.selects sat v = false := by
    intro o ho
    cases h : o.selects sat v
    · rfl
    · have : o ∈ selected sat r.branches v := List.mem_filter.mpr ⟨ho, h⟩
      rw [hnull] at this; cases this
  have news := commitChain_witness (ov := ov) (nx := r.next) (prev := []) (chain := chain) olds
    (by simp)
  obtain ⟨b, hb, hv⟩ := commitChain_covers (ov := ov) (nx := r.next) (prev := []) olds (by simp) hc
  apply distinguish_select.mpr
  refine ⟨fun e => ?_, fun b' hb' => ?_⟩
  · have : b ∈ selected sat (commit ov r chain).1.branches v :=
      List.mem_filter.mpr ⟨by simp [commit, hb], hv⟩
    rw [e] at this; cases this
  · obtain ⟨hb', hv'⟩ := List.mem_filter.mp hb'
    simp only [commit, List.mem_append] at hb'
    rcases hb' with hb' | hb'
    · rw [olds b' hb'] at hv'; cases hv'
    · have := news b' hb' hv'; rw [hc] at this; exact (Option.some.inj this).symm

end

/-! ## Conflicts found later -/

/-- The oldest branch of a list: the lowest rank, the first on a tie. -/
def lowest : List (Branch A I) → Option (Branch A I)
  | [] => none
  | b :: bs => match lowest bs with
    | none => some b
    | some c => if c.rank < b.rank then some c else some b

theorem lowest_mem : ∀ {bs : List (Branch A I)} {b}, lowest bs = some b → b ∈ bs
  | [], _, h => by simp [lowest] at h
  | x :: xs, b, h => by
    unfold lowest at h
    split at h
    · cases h; simp
    · split at h
      · cases h; rename_i c hc _; exact List.mem_cons_of_mem _ (lowest_mem hc)
      · cases h; simp

theorem lowest_none : ∀ {bs : List (Branch A I)}, lowest bs = none ↔ bs = []
  | [] => by simp [lowest]
  | x :: xs => by unfold lowest; split <;> (try split) <;> simp

/-- A conflict's pair: the oldest selected branch, and the oldest selected
branch with a different instruction. -/
def conflictPair [DecidableEq I] (sat : A → V → Bool) (bs : List (Branch A I)) (v : V) :
    Option (Branch A I × Branch A I) := do
  let s := selected sat bs v
  let o ← lowest s
  let n ← lowest (s.filter (·.instr != o.instr))
  pure (o, n)

/-- Resolve the values: reconcile each conflict, older keeping the overlap,
until the outcome is `null` or one instruction. -/
def resolve [DecidableEq I] (sat : A → V → Bool) :
    Nat → Refinement A I → V → Outcome I × Refinement A I × List Record
  | 0, r, v => (distinguish sat r.branches v, r, [])
  | fuel + 1, r, v => match conflictPair sat r.branches v with
    | none => (distinguish sat r.branches v, r, [])
    | some (o, n) =>
      let x := resolve sat fuel (reconcile r o n) v
      (x.1, x.2.1, ⟨o.id, n.id, r.next⟩ :: x.2.2)

section
variable [DecidableEq I] {sat : A → V → Bool} {r : Refinement A I} {v : V}

theorem conflictPair_none (h : conflictPair sat r.branches v = none) :
    distinguish sat r.branches v ≠ .conflict := by
  simp only [conflictPair, Option.bind_eq_bind, Option.pure_def] at h
  cases ho : lowest (selected sat r.branches v) with
  | none => rw [distinguish_null.mpr (lowest_none.mp ho)]; simp
  | some o =>
    rw [ho] at h
    simp only [Option.bind_some, Option.bind_eq_none_iff, reduceCtorEq, imp_false] at h
    cases hn : lowest ((selected sat r.branches v).filter (·.instr != o.instr)) with
    | some n => exact absurd hn (h n)
    | none =>
      have e := lowest_none.mp hn
      rw [distinguish_select.mpr ⟨fun e' => by rw [e'] at ho; simp [lowest] at ho, fun b hb => Decidable.byContradiction fun hne => by
        have : b ∈ (selected sat r.branches v).filter (·.instr != o.instr) :=
          List.mem_filter.mpr ⟨hb, by simpa using hne⟩
        rw [e] at this; cases this⟩]
      simp

theorem conflictPair_some (h : conflictPair sat r.branches v = some (o, n)) :
    o ∈ selected sat r.branches v ∧ n ∈ selected sat r.branches v := by
  simp only [conflictPair, Option.bind_eq_bind, Option.pure_def] at h
  cases ho : lowest (selected sat r.branches v) with
  | none => rw [ho] at h; cases h
  | some o' =>
    rw [ho] at h
    simp only [Option.bind_some] at h
    cases hn : lowest ((selected sat r.branches v).filter (·.instr != o'.instr)) with
    | none => rw [hn] at h; cases h
    | some n' =>
      rw [hn] at h
      simp only [Option.bind_some, Option.some.injEq, Prod.mk.injEq] at h
      obtain ⟨rfl, rfl⟩ := h
      exact ⟨lowest_mem ho, (List.mem_filter.mp (lowest_mem hn)).1⟩

theorem length_filter_lt {p : α → Bool} : ∀ {l : List α} {x}, x ∈ l → p x = false →
    (l.filter p).length < l.length
  | [], _, h, _ => by simp at h
  | a :: l, x, h, hp => by
    rcases List.mem_cons.mp h with rfl | h
    · simp only [List.filter_cons, hp]
      exact Nat.lt_succ_of_le (List.length_filter_le _ _)
    · simp only [List.filter_cons]
      split
      · exact Nat.succ_lt_succ (length_filter_lt h hp)
      · exact Nat.lt_succ_of_lt (length_filter_lt h hp)

/-- Each reconcile takes one branch away from the values: the newer is
retired, and its remainder does not select them. -/
theorem reconcile_fewer (h : conflictPair sat r.branches v = some (o, n)) :
    (selected sat (reconcile r o n).branches v).length < (selected sat r.branches v).length := by
  obtain ⟨⟨ho, hov⟩, ⟨hn, hnv⟩⟩ := (conflictPair_some h).imp List.mem_filter.mp List.mem_filter.mp
  have : selected sat (reconcile r o n).branches v = selected sat (r.branches.filter (·.id != n.id)) v := by
    simp [selected, reconcile, List.filter_append, remainder_disjoint (n := n) (k := r.next) hov]
  rw [this, retire_selected]
  exact length_filter_lt (List.mem_filter.mpr ⟨hn, hnv⟩) (by simp)

/-- With fuel for every selected branch, resolving never ends in a
conflict. -/
theorem resolve_settles : ∀ {fuel} {r : Refinement A I}, r.WellFormed →
    (selected sat r.branches v).length ≤ fuel → (resolve sat fuel r v).1 ≠ .conflict
  | 0, r, _, hl => by
    simp only [resolve]
    rw [distinguish_null.mpr (List.eq_nil_of_length_eq_zero (Nat.le_zero.mp hl))]; simp
  | fuel + 1, r, hr, hl => by
    simp only [resolve]
    split
    · rename_i h; exact conflictPair_none h
    · rename_i o n h
      have hn := (List.mem_filter.mp (conflictPair_some h).2).1
      exact resolve_settles (reconcile_wellFormed hr hn)
        (Nat.le_of_lt_succ (Nat.lt_of_lt_of_le (reconcile_fewer h) hl))

end

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
its own model's chain gives these values, whatever that root selects: an
older branch committed concurrently keeps the stored values, and the step
runs its own instruction once. -/
def afterStartle [DecidableEq I] (sat : A → V → Bool) (chain : List (Cond A × I))
    (committed : List (Branch A I)) (v : V) : After I :=
  match chainSelect sat chain v with
  | none => .startleAgain
  | some i => .trust i (distinguish sat committed v == .select i)

/-- What the step runs does not depend on what any other commit stored. -/
theorem afterStartle_own [DecidableEq I] {sat : A → V → Bool} (c₁ c₂ : List (Branch A I)) :
    (afterStartle sat chain c₁ v).instr = (afterStartle sat chain c₂ v).instr := by
  unfold afterStartle; split <;> rfl

theorem afterStartle_instr [DecidableEq I] {sat : A → V → Bool} {c : List (Branch A I)} :
    (afterStartle sat chain c v).instr = chainSelect sat chain v := by
  unfold afterStartle; split <;> simp_all [After.instr]

/-! ## Transitions and the transition bound -/

/-- The states of a transition: everything the system does with one
instruction. -/
inductive State where
  | reject | hesitate | startle | reconcile | trust
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
