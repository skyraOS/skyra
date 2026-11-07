import Invariants.Walk

/-! Proposal 58: frames and children. A frame's tools are observations in
its scope, each keyed on its binary; the root is granted them, and nothing
that arrives is a tool (`Walk.tools_kept`). An instruction's first position
decides what it is, as `exec`'s first argument does:

* `return $n` returns `$n` to the frame's caller;
* a `$n` holding a tool runs that binary as a process, passing it and the
  rest like any invoke;
* a word is a task: the instruction delegates to a child frame, passing
  observations, tools among them, like any other. It is a child only when
  it passes fewer tool types than its emitter holds.

Every child holds strictly fewer tool types than its parent, and only
tools its parent holds, so frames nest no deeper than the root's tool
types. -/

namespace Frames

open Walk Shape

/-- A scope's tool types: the binaries among its observations, each once. -/
def tools (scope : List Obs) : List String := (scope.filterMap (·.tool)).eraseDups

/-- An instruction as the core reads it, by its first position. A run's
and a delegation's `args` are the other `$n`, in order. -/
inductive Instr where
  | ret (n : Nat)
  | run (tool : Nat) (args : List Nat)
  | delegate (task : String) (args : List Nat)

inductive Reason where
  | unbound (n : Nat)
  /-- The first position is a `$n` that is not a tool. -/
  | notTool (n : Nat)
  /-- A delegation passing as many tool types as its emitter holds, or more:
  it does not narrow. -/
  | narrows (passed held : Nat)
  deriving DecidableEq, Repr

inductive Kind where
  | ret (o : Obs)
  | run (binary : String) (passed : List Obs)
  | child (task : String) (passed : List Obs)
  | reject (r : Reason)

def unboundOf : Fault → Reason
  | .unbound n => .unbound n
  | _ => .unbound 0

def classify (scope : List Obs) : Instr → Kind
  | .ret n => match get? scope n with
    | some o => .ret o
    | none => .reject (.unbound n)
  | .run t args => match get? scope t with
    | none => .reject (.unbound t)
    | some o => match o.tool with
      | none => .reject (.notTool t)
      | some b => match resolve scope args with
        | .ok ps => .run b (o :: ps)
        | .error e => .reject (unboundOf e)
  | .delegate task args => match resolve scope args with
    | .error e => .reject (unboundOf e)
    | .ok ps =>
      if (tools ps).length < (tools scope).length then .child task ps
      else .reject (.narrows (tools ps).length (tools scope).length)

/-- A child frame: what it was passed is its scope, `$1` first. -/
def childFrame (ps : List Obs) : Frame := { scope := ps, pending := none }

/-! ## Children narrow -/

theorem resolve_mem : ∀ {scope : List Obs} {args ps}, resolve scope args = .ok ps → ∀ o ∈ ps, o ∈ scope
  | _, [], ps, h, o, ho => by simp [resolve] at h; subst h; simp at ho
  | scope, n :: ns, ps, h, o, ho => by
    unfold resolve at h
    cases hg : get? scope n with
    | none => simp [hg] at h
    | some o' =>
      cases hr : resolve scope ns with
      | error _ => simp [hg, hr] at h
      | ok os =>
        simp only [hg, hr, Except.ok.injEq] at h
        subst h
        rcases List.mem_cons.mp ho with rfl | ho
        · cases n with
          | zero => simp [get?] at hg
          | succ k => exact List.mem_of_getElem? hg
        · exact resolve_mem hr o ho

theorem mem_tools {scope : List Obs} {b : String} : b ∈ tools scope ↔ ∃ o ∈ scope, o.tool = some b := by
  simp [tools, List.mem_eraseDups]

/-- A child holds strictly fewer tool types than its emitter. -/
theorem child_fewer (h : classify scope (.delegate t args) = .child t' ps) :
    (tools (childFrame ps).scope).length < (tools scope).length := by
  simp only [classify] at h
  cases hr : resolve scope args with
  | error _ => simp [hr] at h
  | ok ps' =>
    simp only [hr] at h
    split at h
    · cases h; assumption
    · cases h

/-- A child holds only tools its emitter holds. -/
theorem child_sub (h : classify scope (.delegate t args) = .child t' ps) :
    ∀ b ∈ tools (childFrame ps).scope, b ∈ tools scope := by
  simp only [classify] at h
  cases hr : resolve scope args with
  | error _ => simp [hr] at h
  | ok ps' =>
    simp only [hr] at h
    split at h
    · cases h
      intro b hb
      obtain ⟨o, ho, hb⟩ := mem_tools.mp hb
      exact mem_tools.mpr ⟨o, resolve_mem hr o ho, hb⟩
    · cases h

/-- A frame runs only a binary it holds, as a tool in its scope. -/
theorem run_held (h : classify scope (.run t args) = .run b ps) : b ∈ tools scope := by
  simp only [classify] at h
  cases hg : get? scope t with
  | none => simp [hg] at h
  | some o =>
    simp only [hg] at h
    cases ht : o.tool with
    | none => simp [ht] at h
    | some b' =>
      simp only [ht] at h
      cases hr : resolve scope args with
      | error _ => simp [hr] at h
      | ok _ =>
        simp only [hr, Kind.run.injEq] at h
        obtain ⟨rfl, -⟩ := h
        have : o ∈ scope := by
          cases t with
          | zero => simp [get?] at hg
          | succ k => exact List.mem_of_getElem? hg
        exact mem_tools.mpr ⟨o, this, ht⟩

/-- A first position that is not a tool never runs. -/
theorem notTool_rejected (hg : get? scope t = some o) (ht : o.tool = none) :
    classify scope (.run t args) = .reject (.notTool t) := by
  simp [classify, hg, ht]

/-! ## Depth -/

/-- Each frame in the chain is a child of the one before: fewer tool types. -/
def Nests : List (List Obs) → Prop
  | [] => True
  | [_] => True
  | a :: b :: rest => (tools b).length < (tools a).length ∧ Nests (b :: rest)

/-- Frames nest no deeper than the root's tool types: a chain from the root
is at most one longer than the number of tool types it holds. -/
theorem nests_depth : ∀ {root : List Obs} {rest}, Nests (root :: rest) →
    (root :: rest).length ≤ (tools root).length + 1
  | _, [], _ => by simp
  | root, b :: rest, h => by
    obtain ⟨hlt, hn⟩ := h
    have := nests_depth hn
    simp only [List.length_cons] at this ⊢
    omega

/-! ## The child's scope table -/

/-- The child's table: its `$k+1` is the k-th observation it was passed. -/
def childTable (ps : List Obs) : List (Nat × Position) := passedTable 0 (List.range' 1 ps.length)

/-- The child's first transition stands at the input base with the passed
positions; the table is right: the position the child's `$n` maps to holds
the value `$n` names in the child. -/
theorem childTable_env {ps : List Obs} (hf : find n (childTable ps) = some p) :
    valueAt p (passedEnv 0 ps) = (get? ps n).map (·.value) := by
  obtain ⟨j, rfl, hj⟩ := find_passed hf
  have hjl : j < ps.length := by simpa using (List.getElem?_eq_some_iff.mp hj).1
  simp only [List.getElem?_range', hjl, Nat.one_mul, Option.some.injEq] at hj
  subst hj
  have := valueAt_passed (ps := ps) (k := 0) (j := j)
  rw [this, Nat.add_comm]
  rfl

end Frames
