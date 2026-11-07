import Invariants.Core

/-! A human root is outside the worker tree's `Kind`. This module adds no
worker instruction and changes no existing classification or observation rule.
Ingress is supplied by the trusted adapter, not selected ingress payload bytes.
This model covers root authoring, scoped passing, return attribution and
capability-subset decisions. Hashing, durable commit and ingress authentication
are the surrounding runtime's responsibilities. -/

namespace HumanRoot

inductive Ingress
  | human | runtime | storage | worker
  deriving DecidableEq

inductive Origin
  | human (principal request : String)
  | child (call target receipt : String) (position : Nat)
  deriving DecidableEq

structure Observation where
  private mk ::
  origin : Origin
  value : String
  deriving DecidableEq

structure Call where
  id : String
  target : String
  inputs : List Observation
  grant : List String
  deriving DecidableEq

structure State where
  principal : String
  scope : List Observation
  entries : List String
  ceiling : List String
  pending : List Call

/-- A trusted ingress decision. Merely being worker node `0` never suffices. -/
def mayAuthor (ingress : Ingress) : Bool := ingress == .human

def mayPass (scope passed : List α) [DecidableEq α] : Bool := passed.all (fun p => decide (p ∈ scope))

def mayGrant (ceiling grant : List String) : Bool := grant.all (· ∈ ceiling)

def mayInvoke (entries : List String) (target : String)
    (scope inputs : List α) [DecidableEq α] (ceiling grant : List String) : Bool :=
  decide (target ∈ entries) && mayPass scope inputs && mayGrant ceiling grant

/-- Repeated human authoring adds an observation, with the principal fixed
by root configuration. The payload never chooses its own origin. -/
def author (s : State) (ingress : Ingress) (request value : String) : Option State :=
  if mayAuthor ingress then
    some { s with scope := s.scope ++ [⟨.human s.principal request, value⟩] }
  else none

/-- The human chooses a child, inputs already held, and a bounded grant.
The observations themselves are passed without changing their origin. -/
def invoke (s : State) (ingress : Ingress) (c : Call) : Option State :=
  if ingress == .human && mayInvoke s.entries c.target s.scope c.inputs s.ceiling c.grant then
    some { s with pending := s.pending ++ [c] }
  else none

/-- Only verified runtime ingress introduces a child's outputs. The output
value never chooses its author: call, target, receipt and position do. -/
def receive (s : State) (ingress : Ingress) (c : Call) (receipt : String)
    (values : List String) : Option State :=
  if ingress == .runtime && decide (c ∈ s.pending) && !(receipt == "") then
    let outputs := values.zipIdx.map fun (value, position) =>
      Observation.mk (.child c.id c.target receipt position) value
    some { s with scope := s.scope ++ outputs, pending := s.pending.erase c }
  else none

end HumanRoot

theorem HumanRoot.mayAuthor_iff : HumanRoot.mayAuthor ingress = true ↔ ingress = .human := by
  simp [HumanRoot.mayAuthor]

theorem HumanRoot.worker_cannot_author : HumanRoot.author s .worker request value = none := by
  simp [HumanRoot.author, HumanRoot.mayAuthor]

theorem HumanRoot.runtime_cannot_author_human : HumanRoot.author s .runtime request value = none := by
  simp [HumanRoot.author, HumanRoot.mayAuthor]

theorem HumanRoot.author_holds (h : HumanRoot.author s .human request value = some next) :
    ∃ o ∈ next.scope, o.origin = .human s.principal request ∧ o.value = value := by
  simp only [HumanRoot.author, HumanRoot.mayAuthor, BEq.rfl, ite_true, Option.some.injEq] at h
  subst next
  exact ⟨⟨.human s.principal request, value⟩, by simp, rfl, rfl⟩

theorem HumanRoot.author_preserves_scope (h : HumanRoot.author s ingress request value = some next) :
    s.scope ⊆ next.scope := by
  unfold HumanRoot.author at h
  split at h
  · cases h; intro o ho; exact List.mem_append_left _ ho
  · contradiction

theorem HumanRoot.author_preserves_authority (h : HumanRoot.author s ingress request value = some next) :
    next.entries = s.entries ∧ next.ceiling = s.ceiling := by
  unfold HumanRoot.author at h
  split at h <;> cases h
  exact ⟨rfl, rfl⟩

theorem HumanRoot.mayPass_iff [DecidableEq α] (scope passed : List α) :
    HumanRoot.mayPass scope passed = true ↔ passed ⊆ scope := by
  simp [HumanRoot.mayPass, List.all_eq_true, List.subset_def]

theorem HumanRoot.mayGrant_iff : HumanRoot.mayGrant ceiling grant = true ↔ grant ⊆ ceiling := by
  simp [HumanRoot.mayGrant, List.all_eq_true, List.subset_def]

theorem HumanRoot.mayInvoke_iff [DecidableEq α] (entries : List String) (target : String) (scope inputs : List α) (ceiling grant : List String) :
    HumanRoot.mayInvoke entries target scope inputs ceiling grant = true ↔
      target ∈ entries ∧ inputs ⊆ scope ∧ grant ⊆ ceiling := by
  simp [HumanRoot.mayInvoke, HumanRoot.mayPass_iff, HumanRoot.mayGrant_iff, and_assoc]

theorem HumanRoot.invoke_bounded (h : HumanRoot.invoke s ingress c = some next) :
    ingress = .human ∧ c.target ∈ s.entries ∧ c.inputs ⊆ s.scope ∧ c.grant ⊆ s.ceiling := by
  unfold HumanRoot.invoke at h
  split at h
  · rename_i allowed
    simp only [Bool.and_eq_true, beq_iff_eq] at allowed
    exact ⟨allowed.1, (HumanRoot.mayInvoke_iff _ _ _ _ _ _).mp allowed.2⟩
  · contradiction

theorem HumanRoot.invoke_preserves_observations (h : HumanRoot.invoke s ingress c = some next) :
    next.scope = s.scope ∧ c ∈ next.pending := by
  unfold HumanRoot.invoke at h
  split at h
  · cases h; exact ⟨rfl, by simp⟩
  · contradiction

theorem HumanRoot.receive_requires_runtime (h : HumanRoot.receive s ingress c receipt values = some next) :
    ingress = .runtime ∧ c ∈ s.pending ∧ receipt ≠ "" := by
  unfold HumanRoot.receive at h
  split at h
  · rename_i allowed
    simpa [Bool.and_eq_true, and_assoc] using allowed
  · contradiction

theorem HumanRoot.human_cannot_claim_child_return : HumanRoot.receive s .human c receipt values = none := by
  simp [HumanRoot.receive]

theorem HumanRoot.receive_preserves_scope (h : HumanRoot.receive s ingress c receipt values = some next) :
    s.scope ⊆ next.scope := by
  unfold HumanRoot.receive at h
  split at h
  · cases h; intro o ho; exact List.mem_append_left _ ho
  · contradiction

theorem HumanRoot.receive_attributes_child (h : HumanRoot.receive s ingress c receipt values = some next)
    (ho : o ∈ next.scope) (fresh : o ∉ s.scope) :
    ∃ position, o.origin = .child c.id c.target receipt position := by
  unfold HumanRoot.receive at h
  split at h
  · cases h
    simp only [List.mem_append] at ho
    rcases ho with old | returned
    · exact False.elim (fresh old)
    · obtain ⟨pair, -, eq⟩ := List.mem_map.mp returned
      subst o
      exact ⟨pair.2, rfl⟩
  · contradiction

/-! Traces. A root moves only by `author`, `invoke` and `receive`; what holds
after every step holds of every state the root can reach from its start. -/

namespace HumanRoot

inductive Step : State → State → Prop
  | author  (h : author s ingress request value = some next) : Step s next
  | invoke  (h : invoke s ingress c = some next) : Step s next
  | receive (h : receive s ingress c receipt values = some next) : Step s next

inductive Reachable (start : State) : State → Prop
  | refl : Reachable start start
  | step (r : Reachable start s) (h : Step s next) : Reachable start next

/-- Every pending call was in bounds when the human made it, and stays so. -/
def State.Bounded (s : State) : Prop :=
  ∀ c ∈ s.pending, c.target ∈ s.entries ∧ c.inputs ⊆ s.scope ∧ c.grant ⊆ s.ceiling

/-- Where every observation in scope came from: the start, the principal, or
a receipted return from an entry. -/
def State.Attributed (s start : State) : Prop :=
  ∀ o ∈ s.scope, o ∈ start.scope ∨ (∃ request, o.origin = .human s.principal request) ∨
    ∃ id target receipt position, o.origin = .child id target receipt position ∧
      target ∈ s.entries ∧ receipt ≠ ""

end HumanRoot

theorem HumanRoot.author_fixes (h : HumanRoot.author s ingress request value = some next) :
    next.principal = s.principal ∧ next.entries = s.entries ∧ next.ceiling = s.ceiling ∧
      next.pending = s.pending := by
  unfold HumanRoot.author at h
  split at h <;> cases h
  exact ⟨rfl, rfl, rfl, rfl⟩

theorem HumanRoot.invoke_fixes (h : HumanRoot.invoke s ingress c = some next) :
    next.principal = s.principal ∧ next.entries = s.entries ∧ next.ceiling = s.ceiling ∧
      next.scope = s.scope ∧ next.pending = s.pending ++ [c] := by
  unfold HumanRoot.invoke at h
  split at h <;> cases h
  exact ⟨rfl, rfl, rfl, rfl, rfl⟩

theorem HumanRoot.receive_fixes (h : HumanRoot.receive s ingress c receipt values = some next) :
    next.principal = s.principal ∧ next.entries = s.entries ∧ next.ceiling = s.ceiling ∧
      next.pending = s.pending.erase c := by
  unfold HumanRoot.receive at h
  split at h <;> cases h
  exact ⟨rfl, rfl, rfl, rfl⟩

/-- A worker moves the root by no step at all. -/
theorem HumanRoot.worker_moves_nothing :
    HumanRoot.author s .worker request value = none ∧ HumanRoot.invoke s .worker c = none ∧
      HumanRoot.receive s .worker c receipt values = none := by
  simp [HumanRoot.author, HumanRoot.mayAuthor, HumanRoot.invoke, HumanRoot.receive]

/-- Storage replays nothing into the root: it neither authors, invokes nor returns. -/
theorem HumanRoot.storage_moves_nothing :
    HumanRoot.author s .storage request value = none ∧ HumanRoot.invoke s .storage c = none ∧
      HumanRoot.receive s .storage c receipt values = none := by
  simp [HumanRoot.author, HumanRoot.mayAuthor, HumanRoot.invoke, HumanRoot.receive]

/-- One return per call: once a call made once is received, its receipt is refused again. -/
theorem HumanRoot.receive_once (h : HumanRoot.receive s ingress c receipt values = some next)
    (once : s.pending.count c = 1) : HumanRoot.receive next ingress' c receipt' values' = none := by
  have hp := (HumanRoot.receive_fixes h).2.2.2
  have hn : c ∉ next.pending := by
    rw [hp, ← List.count_eq_zero, List.count_erase_self, once]
  unfold HumanRoot.receive
  simp [hn]

theorem HumanRoot.step_fixes (h : HumanRoot.Step s next) :
    next.principal = s.principal ∧ next.entries = s.entries ∧ next.ceiling = s.ceiling := by
  cases h with
  | author h => obtain ⟨a, b, c, -⟩ := HumanRoot.author_fixes h; exact ⟨a, b, c⟩
  | invoke h => obtain ⟨a, b, c, -⟩ := HumanRoot.invoke_fixes h; exact ⟨a, b, c⟩
  | receive h => obtain ⟨a, b, c, -⟩ := HumanRoot.receive_fixes h; exact ⟨a, b, c⟩

theorem HumanRoot.step_preserves_scope (h : HumanRoot.Step s next) : s.scope ⊆ next.scope := by
  cases h with
  | author h => exact HumanRoot.author_preserves_scope h
  | invoke h => rw [(HumanRoot.invoke_fixes h).2.2.2.1]; exact List.Subset.refl _
  | receive h => exact HumanRoot.receive_preserves_scope h

/-- No trace changes who the principal is, what may be invoked, or what may be granted. -/
theorem HumanRoot.reachable_fixes (r : HumanRoot.Reachable start s) :
    s.principal = start.principal ∧ s.entries = start.entries ∧ s.ceiling = start.ceiling := by
  induction r with
  | refl => exact ⟨rfl, rfl, rfl⟩
  | step _ h ih =>
    obtain ⟨p, e, c⟩ := HumanRoot.step_fixes h
    exact ⟨p.trans ih.1, e.trans ih.2.1, c.trans ih.2.2⟩

/-- Nothing held is ever lost. -/
theorem HumanRoot.reachable_preserves_scope (r : HumanRoot.Reachable start s) :
    start.scope ⊆ s.scope := by
  induction r with
  | refl => exact List.Subset.refl _
  | step _ h ih => exact List.Subset.trans ih (HumanRoot.step_preserves_scope h)

theorem HumanRoot.step_bounded (hb : s.Bounded) (h : HumanRoot.Step s next) : next.Bounded := by
  have hs := HumanRoot.step_preserves_scope h
  obtain ⟨-, he, hc⟩ := HumanRoot.step_fixes h
  rw [HumanRoot.State.Bounded, he, hc]
  cases h with
  | author h =>
    rw [(HumanRoot.author_fixes h).2.2.2]
    intro c hm
    obtain ⟨t, i, g⟩ := hb c hm
    exact ⟨t, List.Subset.trans i hs, g⟩
  | invoke h =>
    rename_i call
    rw [(HumanRoot.invoke_fixes h).2.2.2.2]
    intro c hm
    rcases List.mem_append.mp hm with old | new
    · obtain ⟨t, i, g⟩ := hb c old
      exact ⟨t, List.Subset.trans i hs, g⟩
    · simp only [List.mem_singleton] at new
      subst new
      obtain ⟨-, t, i, g⟩ := HumanRoot.invoke_bounded h
      exact ⟨t, List.Subset.trans i hs, g⟩
  | receive h =>
    rw [(HumanRoot.receive_fixes h).2.2.2]
    intro c hm
    obtain ⟨t, i, g⟩ := hb c (List.mem_of_mem_erase hm)
    exact ⟨t, List.Subset.trans i hs, g⟩

/-- From a bounded start, every call still pending targets an
entry, passes only what the root holds, and grants within the ceiling. -/
theorem HumanRoot.reachable_bounded (hb : start.Bounded) (r : HumanRoot.Reachable start s) :
    s.Bounded := by
  induction r with
  | refl => exact hb
  | step _ h ih => exact HumanRoot.step_bounded ih h

theorem HumanRoot.step_attributed (hb : s.Bounded) (ha : s.Attributed start)
    (h : HumanRoot.Step s next) : next.Attributed start := by
  obtain ⟨hp, he, -⟩ := HumanRoot.step_fixes h
  unfold HumanRoot.State.Attributed
  rw [hp, he]
  cases h with
  | author h =>
    unfold HumanRoot.author at h
    split at h
    · cases h
      intro o ho
      rcases List.mem_append.mp ho with old | new
      · exact ha o old
      · simp only [List.mem_singleton] at new
        subst new
        exact Or.inr (Or.inl ⟨_, rfl⟩)
    · contradiction
  | invoke h =>
    rw [(HumanRoot.invoke_fixes h).2.2.2.1]
    exact ha
  | receive h =>
    rename_i call receipt values
    obtain ⟨-, hm, hr⟩ := HumanRoot.receive_requires_runtime h
    have ht := (hb call hm).1
    intro o ho
    by_cases fresh : o ∈ s.scope
    · exact ha o fresh
    · obtain ⟨position, e⟩ := HumanRoot.receive_attributes_child h ho fresh
      exact Or.inr (Or.inr ⟨_, _, _, position, e, ht, hr⟩)

/-- Every observation the root holds came from its start, from its own
principal, or back from an entry under a receipt. Nothing else gets in. -/
theorem HumanRoot.reachable_attributed (hb : start.Bounded) (r : HumanRoot.Reachable start s) :
    s.Attributed start := by
  induction r with
  | refl => intro o ho; exact Or.inl ho
  | step r h ih => exact HumanRoot.step_attributed (HumanRoot.reachable_bounded hb r) ih h

/-- A start with nothing pending is bounded. -/
theorem HumanRoot.bounded_of_idle {start : HumanRoot.State} (h : start.pending = []) : start.Bounded := by
  intro c hc; rw [h] at hc; cases hc

/-- Nothing pending at the start: a reached state's human observations
name the start's principal, and its returns name the start's entries. -/
theorem HumanRoot.reachable_from_idle (idle : start.pending = [])
    (r : HumanRoot.Reachable start s) (ho : o ∈ s.scope) :
    o ∈ start.scope ∨ (∃ request, o.origin = .human start.principal request) ∨
      ∃ id target receipt position, o.origin = .child id target receipt position ∧
        target ∈ start.entries ∧ receipt ≠ "" := by
  obtain ⟨hp, he, -⟩ := HumanRoot.reachable_fixes r
  have := HumanRoot.reachable_attributed (HumanRoot.bounded_of_idle idle) r o ho
  rwa [hp, he] at this
