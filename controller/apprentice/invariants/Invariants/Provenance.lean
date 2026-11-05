import Invariants.Core

/-! Provenance. `Observation`'s constructor is private, but privacy hides a
name, not a value: Lean's logic ranges over every observation, and a tactic
can still reach the constructor. So the claim is made of the system instead.
A run starts with nothing held; a node gains an observation only by `observe`
or the root's endowment, and moves one only by passing what it holds, down
an edge or back up. Every observation any node holds was then made through
`observe` or endowed, and so was everything it was made from. The moves here
check less than the runtime does (any node may be endowed, any kind may
observe, a return is checked only against what the child holds), so the
claim covers every stricter runtime too. -/

/-- Made by the system: the root's endowment, or what `observe` admitted from
observations already made. -/
inductive Made (a : Apprentice) : Observation → Prop
  | root (h : rootObservation a t ρ = some o) : Made a o
  | observed (h : observe a k t ms ρ = some o) (hm : ∀ m ∈ ms, Made a m) : Made a o

/-- What every node holds, by node. -/
abbrev Scopes := Nat → List Observation

def Scopes.set (s : Scopes) (n : Nat) (v : List Observation) : Scopes :=
  fun m => if m = n then v else s m

/-- How a run moves. Only `endow` and `observe` add an observation the run
did not hold; `invoke` and `ret` move what a node already holds. -/
inductive Move (a : Apprentice) : Scopes → Scopes → Prop
  | endow (h : rootObservation a t ρ = some o) : Move a s (s.set n (s n ++ [o]))
  | observe (hm : ms ⊆ s n) (h : observe a k t ms ρ = some o) : Move a s (s.set n (s n ++ [o]))
  | invoke (h : mayPass (s parent) passed = true) :
      Move a s (s.set child (receive .invoke (s child) passed))
  | ret (h : mayPass (s child) returned = true) :
      Move a s (s.set parent (receive .ret (s parent) returned))

/-- Every state a run reaches from nothing held. -/
inductive Run (a : Apprentice) : Scopes → Prop
  | empty : Run a fun _ => []
  | move (r : Run a s) (h : Move a s s') : Run a s'

/-- Everything the system holds is made. -/
def Scopes.Holds (a : Apprentice) (s : Scopes) : Prop := ∀ n, ∀ o ∈ s n, Made a o

theorem rootObservation_members (h : rootObservation a t ρ = some o) : o.members = [] := by
  unfold rootObservation at h; split at h <;> cases h; rfl

theorem Made.members (h : Made a o) (hm : m ∈ o.members) : Made a m := by
  cases h with
  | root h => rw [rootObservation_members h] at hm; cases hm
  | observed h hs => rw [(observe_eq_some.mp h).2.2.1] at hm; exact hs m hm

theorem lineage_cases (hx : x ∈ lineage o) : x = o ∨ ∃ m ∈ o.members, x ∈ lineage m := by
  cases o
  simp only [lineage, List.mem_cons, List.mem_flatMap, List.mem_attach, true_and,
    Subtype.exists] at hx
  rcases hx with rfl | ⟨m, hm, hx⟩
  · exact Or.inl rfl
  · exact Or.inr ⟨m, hm, hx⟩

/-- Everything a made observation was made from, at every depth, is made. -/
theorem Made.lineage (h : Made a o) (hx : x ∈ lineage o) : Made a x := by
  induction h generalizing x with
  | root h' =>
    rcases lineage_cases hx with rfl | ⟨m, hm, -⟩
    · exact .root h'
    · rw [rootObservation_members h'] at hm; cases hm
  | observed h' hs ih =>
    rcases lineage_cases hx with rfl | ⟨m, hm, hx⟩
    · exact .observed h' hs
    · rw [(observe_eq_some.mp h').2.2.1] at hm; exact ih m hm hx

/-- A made observation came through `observe`, or is an endowed effect. -/
theorem Made.through (h : Made a o) :
    (∃ k t ms ρ, observe a k t ms ρ = some o) ∨
      ∃ t ρ, rootObservation a t ρ = some o ∧ o.isEffect = true := by
  cases h with
  | observed h => exact Or.inl ⟨_, _, _, _, h⟩
  | root h => exact Or.inr ⟨_, _, h, Observation.effect_iff.mpr (rootObservation_members h)⟩

/-- A made computation came through `observe`: the root is endowed only effects. -/
theorem Made.computation (h : Made a o) (hc : o.isComputation = true) :
    ∃ k t ms ρ, observe a k t ms ρ = some o := by
  rcases h.through with h | ⟨-, -, -, he⟩
  · exact h
  · rw [Observation.effect_xor_computation, hc] at he; cases he

theorem Move.made (hs : s.Holds a) (h : Move a s s') : s'.Holds a := by
  intro m o ho
  cases h with
  | endow h =>
    rename_i n
    simp only [Scopes.set] at ho
    split at ho
    · rcases List.mem_append.mp ho with old | new
      · subst m; exact hs _ o old
      · simp only [List.mem_singleton] at new; subst new; exact .root h
    · exact hs m o ho
  | observe hm h =>
    rename_i n
    simp only [Scopes.set] at ho
    split at ho
    · rcases List.mem_append.mp ho with old | new
      · subst m; exact hs _ o old
      · simp only [List.mem_singleton] at new; subst new
        exact .observed h fun x hx => hs n x (hm hx)
    · exact hs m o ho
  | invoke h =>
    simp only [Scopes.set, receive] at ho
    split at ho
    · exact hs _ o (mayPass_iff.mp h ho)
    · exact hs m o ho
  | ret h =>
    simp only [Scopes.set, receive] at ho
    split at ho
    · rcases List.mem_append.mp ho with old | back
      · subst m; exact hs _ o old
      · exact hs _ o (mayPass_iff.mp h back)
    · exact hs m o ho

/-- No observation exists in a run that was not made: every one any node
holds came through `observe`, or is the root's endowment, and so did
everything it was made from. -/
theorem Run.made (r : Run a s) : s.Holds a := by
  induction r with
  | empty => intro _ _ ho; cases ho
  | move _ h ih => exact Move.made ih h

/-- What a node holds in a run came through `observe`, unless the root was
endowed it, as an effect. -/
theorem Run.through (r : Run a s) (ho : o ∈ s n) :
    (∃ k t ms ρ, observe a k t ms ρ = some o) ∨
      ∃ t ρ, rootObservation a t ρ = some o ∧ o.isEffect = true :=
  (r.made n o ho).through

/-- Every computation a node holds in a run came through `observe`, and so
did every computation in its lineage. -/
theorem Run.computation (r : Run a s) (ho : o ∈ s n) (hx : x ∈ lineage o)
    (hc : x.isComputation = true) : ∃ k t ms ρ, observe a k t ms ρ = some x :=
  ((r.made n o ho).lineage hx).computation hc
