import Invariants.Refine

/-! Proposal 58: dispatch, one transition's control. An arrival stands at a
base (`Walk.sense`); the base's refinement is distinguished on its values.
One instruction is trusted. A conflict is reconciled, the older branch
keeping the values, and the settled instruction trusted. Null startles:
the model writes a guard, which is committed, and the step runs its own
chain's instruction. A guard that does not store or check is rejected and
the model asked again; a guard that does not cover what arrived is
committed, and the model asked again. Every path that runs an instruction
ends in exactly one `trust`: one transition. The model is asked at most
`refinementBound` times in one transition, its refinement bound: past
that the frame ends, returning an error observation to its caller, as an
exhausted transition bound does.

What the model writes is given, in order, as `writes`: the model is the
shell's, and its answers are inputs. -/

namespace Dispatch

open Refine

variable {A V I : Type}

/-- What the model wrote when startled. -/
inductive Wrote (A I : Type) where
  /-- A guard, stored and checked against the base's shapes. -/
  | guard (chain : List (Cond A × I))
  /-- A guard reading a `$n` that is not in the base: it is not stored, and
  runs once, unlearned. -/
  | unlearned (chain : List (Cond A × I))
  /-- A guard that does not store or check: rejected. -/
  | refused

/-- One transition's result: its states, the refinement after it, what
`trust` runs (none while the model has not written a guard that covers the
arrival), whether the refinement now selects it, and how many times the
model was asked, an ask still unanswered included; `ended` is whether the
asks ran out, ending the frame. -/
structure Out (A I : Type) where
  states : Transition
  refinement : Refinement A I
  instr : Option I
  learned : Bool
  asks : Nat
  ended : Bool

variable [DecidableEq I] (sat : A → V → Bool) (ov : Cond A → Cond A → Option Bool) (v : V)
  (refinementBound : Nat)

/-- The asks ran out: nothing runs, and the frame ends. -/
def ended (r : Refinement A I) (acc : Transition) (n : Nat) : Out A I :=
  ⟨acc, r, none, false, n, true⟩

/-- About to ask, `n` asks made. -/
def startle : Refinement A I → List (Wrote A I) → Transition → Nat → Out A I
  | r, [], acc, n => if refinementBound ≤ n then ended r acc n
    else ⟨acc ++ [.startle], r, none, false, n + 1, false⟩
  | r, .refused :: ws, acc, n => if refinementBound ≤ n then ended r acc n
    else startle r ws (acc ++ [.startle, .reject]) (n + 1)
  | r, .unlearned c :: ws, acc, n => if refinementBound ≤ n then ended r acc n
    else match chainSelect sat c v with
    | none => startle r ws (acc ++ [.startle, .reject]) (n + 1)
    | some i => ⟨acc ++ [.startle, .trust], r, some i, false, n + 1, false⟩
  | r, .guard c :: ws, acc, n => if refinementBound ≤ n then ended r acc n
    else
    let r' := (commit ov r c).1
    match afterStartle sat c r'.branches v with
    | .startleAgain => startle r' ws (acc ++ [.startle, .reconcile]) (n + 1)
    | .trust i l => ⟨acc ++ [.startle, .reconcile, .trust], r', some i, l, n + 1, false⟩

def decide (r : Refinement A I) (ws : List (Wrote A I)) : Out A I :=
  match distinguish sat r.branches v with
  | .select i => ⟨[.trust], r, some i, true, 0, false⟩
  | .null => startle sat ov v refinementBound r ws [] 0
  | .conflict =>
    let x := resolve sat r.branches.length r v
    match x.1 with
    | .select i => ⟨[.reconcile, .trust], x.2.1, some i, true, 0, false⟩
    | _ => startle sat ov v refinementBound x.2.1 ws [.reconcile] 0

variable {sat ov v refinementBound}

theorem startle_wf : ∀ {r : Refinement A I} {ws acc n i}, State.trust ∉ acc →
    (startle sat ov v refinementBound r ws acc n).instr = some i →
    (startle sat ov v refinementBound r ws acc n).states.WellFormed
  | _, [], _, n, _, _, h => by
    simp only [startle] at h; split at h <;> simp [ended] at h
  | r, .refused :: ws, acc, n, i, ht, h => by
    simp only [startle] at h ⊢
    by_cases hm : refinementBound ≤ n
    · simp [hm, ended] at h
    · simp only [hm, ite_false] at h ⊢; exact startle_wf (by simp [ht]) h
  | r, .unlearned c :: ws, acc, n, i, ht, h => by
    simp only [startle] at h ⊢
    by_cases hm : refinementBound ≤ n
    · simp [hm, ended] at h
    · simp only [hm, ite_false] at h ⊢
      cases hc : chainSelect sat c v with
      | none => simp only [hc] at h ⊢; exact startle_wf (by simp [ht]) h
      | some j => exact ⟨acc ++ [.startle], by simp, by simp [ht]⟩
  | r, .guard c :: ws, acc, n, i, ht, h => by
    simp only [startle] at h ⊢
    by_cases hm : refinementBound ≤ n
    · simp [hm, ended] at h
    · simp only [hm, ite_false] at h ⊢
      cases ha : afterStartle sat c (commit ov r c).1.branches v with
      | startleAgain => simp only [ha] at h ⊢; exact startle_wf (by simp [ht]) h
      | trust j l => exact ⟨acc ++ [.startle, .reconcile], by simp, by simp [ht]⟩

/-- Every transition that runs an instruction ends in exactly one `trust`. -/
theorem decide_wf {r : Refinement A I} {ws i} (h : (decide sat ov v refinementBound r ws).instr = some i) :
    (decide sat ov v refinementBound r ws).states.WellFormed := by
  unfold decide at h ⊢
  cases hd : distinguish sat r.branches v with
  | select j => exact ⟨[], rfl, by simp⟩
  | null => simp only [hd] at h ⊢; exact startle_wf (by simp) h
  | conflict =>
    simp only [hd] at h ⊢
    cases hx : (resolve sat r.branches.length r v).1 with
    | select j => exact ⟨[.reconcile], rfl, by simp⟩
    | null => simp only [hx] at h ⊢; exact startle_wf (by simp) h
    | conflict => simp only [hx] at h ⊢; exact startle_wf (by simp) h

/-- The model is asked at most `refinementBound` times; a frame whose asks run out
has asked exactly that many, and runs nothing. -/
theorem startle_asks : ∀ {r : Refinement A I} {ws acc n}, n ≤ refinementBound →
    (startle sat ov v refinementBound r ws acc n).asks ≤ refinementBound ∧
    ((startle sat ov v refinementBound r ws acc n).ended = true →
      (startle sat ov v refinementBound r ws acc n).asks = refinementBound ∧
      (startle sat ov v refinementBound r ws acc n).instr = none)
  | _, [], _, n, hn => by
    simp only [startle]
    by_cases hm : refinementBound ≤ n
    · simp only [hm, ite_true, ended]; exact ⟨hn, fun _ => ⟨by omega, trivial⟩⟩
    · simp only [hm, ite_false]; simp; omega
  | r, .refused :: ws, acc, n, hn => by
    simp only [startle]
    by_cases hm : refinementBound ≤ n
    · simp only [hm, ite_true, ended]; exact ⟨hn, fun _ => ⟨by omega, trivial⟩⟩
    · simp only [hm, ite_false]; exact startle_asks (by omega)
  | r, .unlearned c :: ws, acc, n, hn => by
    simp only [startle]
    by_cases hm : refinementBound ≤ n
    · simp only [hm, ite_true, ended]; exact ⟨hn, fun _ => ⟨by omega, trivial⟩⟩
    · simp only [hm, ite_false]
      cases chainSelect sat c v with
      | none => exact startle_asks (by omega)
      | some j => simp; omega
  | r, .guard c :: ws, acc, n, hn => by
    simp only [startle]
    by_cases hm : refinementBound ≤ n
    · simp only [hm, ite_true, ended]; exact ⟨hn, fun _ => ⟨by omega, trivial⟩⟩
    · simp only [hm, ite_false]
      cases afterStartle sat c (commit ov r c).1.branches v with
      | startleAgain => exact startle_asks (by omega)
      | trust j l => simp; omega

theorem decide_asks {r : Refinement A I} {ws} :
    (decide sat ov v refinementBound r ws).asks ≤ refinementBound ∧
    ((decide sat ov v refinementBound r ws).ended = true →
      (decide sat ov v refinementBound r ws).asks = refinementBound ∧ (decide sat ov v refinementBound r ws).instr = none) := by
  unfold decide
  cases hd : distinguish sat r.branches v with
  | select j => simp
  | null => exact startle_asks (Nat.zero_le _)
  | conflict =>
    simp only
    cases hx : (resolve sat r.branches.length r v).1 with
    | select j => simp
    | null => exact startle_asks (Nat.zero_le _)
    | conflict => exact startle_asks (Nat.zero_le _)

/-- The model is asked only when nothing in the refinement settles the
arrival: what was learned is reused with no model call. -/
theorem decide_select {r : Refinement A I} {ws i} (h : distinguish sat r.branches v = .select i) :
    decide sat ov v refinementBound r ws = ⟨[.trust], r, some i, true, 0, false⟩ := by
  unfold decide; rw [h]

/-- An unlearned run changes nothing stored. -/
theorem unlearned_keeps {r : Refinement A I} {c : List (Cond A × I)} {ws i} (hm : 0 < refinementBound)
    (hnull : distinguish sat r.branches v = .null) (hc : chainSelect sat c v = some i) :
    decide sat ov v refinementBound r (.unlearned c :: ws) = ⟨[.startle, .trust], r, some i, false, 1, false⟩ := by
  unfold decide; rw [hnull]; simp [startle, hc, Nat.not_le.mpr hm]

/-- What a startle learns is reused: when the model's first guard covers
what arrived, the next arrival at the base with those values selects its
instruction, trusted at once, and the model is not asked. -/
theorem learned_reused {r : Refinement A I} {c : List (Cond A × I)} {ws ws' i} (hm : 0 < refinementBound)
    (hnull : distinguish sat r.branches v = .null) (hc : chainSelect sat c v = some i) :
    decide sat ov v refinementBound (decide sat ov v refinementBound r (.guard c :: ws)).refinement ws' =
      ⟨[.trust], (commit ov r c).1, some i, true, 0, false⟩ := by
  have hw := commit_witness (ov := ov) (chain := c) (distinguish_null.mp hnull) hc
  have hr : (decide sat ov v refinementBound r (.guard c :: ws)).refinement = (commit ov r c).1 := by
    unfold decide; rw [hnull]
    simp only [startle, Nat.not_le.mpr hm, ite_false, afterStartle, hc, hw]
  rw [hr]; exact decide_select hw

/-- The first startle's guard, when it covers the arrival, is trusted at
once and is learned: one ask. -/
theorem first_guard {r : Refinement A I} {c : List (Cond A × I)} {ws i} (hm : 0 < refinementBound)
    (hnull : distinguish sat r.branches v = .null) (hc : chainSelect sat c v = some i) :
    decide sat ov v refinementBound r (.guard c :: ws) =
      ⟨[.startle, .reconcile, .trust], (commit ov r c).1, some i, true, 1, false⟩ := by
  have hw := commit_witness (ov := ov) (chain := c) (distinguish_null.mp hnull) hc
  unfold decide; rw [hnull]
  simp [startle, afterStartle, hc, hw, Nat.not_le.mpr hm]

end Dispatch
