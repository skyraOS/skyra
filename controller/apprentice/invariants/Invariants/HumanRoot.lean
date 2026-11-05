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
