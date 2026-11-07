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
types.

An argument is a `$n` or a cut of one: runs of its lines (`$2[3:7,9]`) or a
field path (`$2.dir`). The model never writes the value: the instruction
says where to take it from, so what is learned applies to new values. A cut
is part of binding an instruction, not a transition or an ask; its value is
a function of the value it is cut from, and it is never a tool. -/

namespace Frames

open Walk Shape

/-- A scope's tool types: the binaries among its observations, each once. -/
def tools (scope : List Obs) : List String := (scope.filterMap (·.tool)).eraseDups

/-! ## Cuts -/

/-- What an argument takes of its `$n`: all of it, runs of its lines, from
`a` to `b` inclusive and counted from 1, or the value at a field path. -/
inductive Sel where
  | whole
  | lines (runs : List (Nat × Nat))
  | field (path : List String)
  deriving DecidableEq, Repr

/-- A selector as it is written after `$n`, and as its base is keyed. -/
def Sel.enc : Sel → String
  | .whole => ""
  | .lines runs => "[" ++ ",".intercalate (runs.map fun (a, b) => toString a ++ ":" ++ toString b) ++ "]"
  | .field path => String.join (path.map ("." ++ ·))

/-- A value's text: a string, or bytes that are UTF-8. -/
def textOf : Value → Option String
  | .str s => some s
  | .bytes b => String.fromUTF8? b
  | _ => none

/-- Text's lines: split at newlines, a final newline ending the last line
rather than starting another. -/
def lines (s : String) : List String :=
  let ls := s.splitOn "\n"
  if ls.length > 1 ∧ ls.getLast? = some "" then ls.dropLast else ls

def cutRuns (ls : List String) : List (Nat × Nat) → Option (List String)
  | [] => some []
  | (a, b) :: rest =>
    if 1 ≤ a ∧ a ≤ b ∧ b ≤ ls.length then
      (cutRuns ls rest).map (((ls.drop (a - 1)).take (b - a + 1)) ++ ·)
    else none

def fieldAt : Value → List String → Option Value
  | v, [] => some v
  | .obj kvs, k :: ks => (kvs.lookup k).bind (fieldAt · ks)
  | _, _ :: _ => none

/-- A cut's value: the selected lines joined, or the field's value. -/
def cutValue : Sel → Value → Option Value
  | .whole, v => some v
  | .lines [], _ => none
  | .lines runs, v => (textOf v).bind fun s => (cutRuns (lines s) runs).map fun ls => .str ("\n".intercalate ls)
  | .field [], _ => none
  | .field path, v => fieldAt v path

/-- A cut observation: its value cut, its base keyed on the base it was cut
from and the selector, never a tool. All of `$n` is `$n` itself. -/
def cutObs (cutBase : String → String → String) : Sel → Obs → Option Obs
  | .whole, o => some o
  | s, o => (cutValue s o.value).map fun v => { value := v, base := cutBase o.base s.enc }

/-- An argument: a `$n` and what it takes of it. -/
structure Arg where
  n : Nat
  sel : Sel := .whole

/-- An instruction as the core reads it, by its first position. A run's
and a delegation's `args` are the other arguments, in order. The tool a
run names is a whole `$n`. -/
inductive Instr where
  | ret (a : Arg)
  | run (tool : Nat) (args : List Arg)
  | delegate (task : String) (args : List Arg)

inductive Reason where
  | unbound (n : Nat)
  /-- A cut that does not apply to `$n`'s value: not text, lines it does
  not have, or a field it does not hold. -/
  | cut (n : Nat)
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

variable (cutBase : String → String → String)

/-- An argument's observation: `$n`, cut. -/
def argObs (scope : List Obs) (a : Arg) : Except Reason Obs :=
  match get? scope a.n with
  | none => .error (.unbound a.n)
  | some o => match cutObs cutBase a.sel o with
    | none => .error (.cut a.n)
    | some o' => .ok o'

def resolveArgs (scope : List Obs) : List Arg → Except Reason (List Obs)
  | [] => .ok []
  | a :: as => match argObs cutBase scope a with
    | .error r => .error r
    | .ok o => match resolveArgs scope as with
      | .ok os => .ok (o :: os)
      | .error r => .error r

def classify (scope : List Obs) : Instr → Kind
  | .ret a => match argObs cutBase scope a with
    | .ok o => .ret o
    | .error r => .reject r
  | .run t args => match get? scope t with
    | none => .reject (.unbound t)
    | some o => match o.tool with
      | none => .reject (.notTool t)
      | some b => match resolveArgs cutBase scope args with
        | .ok ps => .run b (o :: ps)
        | .error r => .reject r
  | .delegate task args => match resolveArgs cutBase scope args with
    | .error r => .reject r
    | .ok ps =>
      if (tools ps).length < (tools scope).length then .child task ps
      else .reject (.narrows (tools ps).length (tools scope).length)

/-- A child frame: what it was passed is its scope, `$1` first. -/
def childFrame (ps : List Obs) : Frame := { scope := ps, pending := none }

variable {cutBase}

/-! ## Cuts are taken, never written -/

/-- All of `$n` is `$n`. -/
theorem cut_whole : cutObs cutBase .whole o = some o := rfl

/-- A cut is never a tool: cutting a tool passes its name as text. -/
theorem cut_not_tool (hs : s ≠ .whole) (h : cutObs cutBase s o = some o') : o'.tool = none := by
  cases s with
  | whole => exact absurd rfl hs
  | lines _ | field _ =>
    simp only [cutObs, Option.map_eq_some_iff] at h
    obtain ⟨_, -, rfl⟩ := h; rfl

/-- A cut's value is a function of the value it is cut from: the model
never writes it, and the same selector on a new value takes from that. -/
theorem cut_of_value (hv : o.value = o'.value) :
    (cutObs cutBase s o).map (·.value) = (cutObs cutBase s o').map (·.value) := by
  cases s <;> simp [cutObs, hv, Option.map_map, Function.comp_def]

/-- A cut's base is the base it was cut from and the selector, never the
value: executions cutting the same way from the same base stand at one
location. -/
theorem cut_base (hs : s ≠ .whole) (h : cutObs cutBase s o = some o') : o'.base = cutBase o.base s.enc := by
  cases s with
  | whole => exact absurd rfl hs
  | lines _ | field _ =>
    simp only [cutObs, Option.map_eq_some_iff] at h
    obtain ⟨_, -, rfl⟩ := h; rfl

/-- What an argument holds of a tool, it holds of the `$n` it names: the
tool, whole, or none. -/
theorem argObs_tool (h : argObs cutBase scope a = .ok o') (ht : o'.tool = some b) :
    ∃ o ∈ scope, o.tool = some b := by
  unfold argObs at h
  cases hg : get? scope a.n with
  | none => simp [hg] at h
  | some o =>
    simp only [hg] at h
    cases hc : cutObs cutBase a.sel o with
    | none => simp [hc] at h
    | some o'' =>
      simp only [hc, Except.ok.injEq] at h
      subst h
      have hm : o ∈ scope := by
        cases hn : a.n with
        | zero => simp [hn, get?] at hg
        | succ k => rw [hn] at hg; exact List.mem_of_getElem? hg
      by_cases hw : a.sel = .whole
      · rw [hw, cut_whole] at hc; cases hc; exact ⟨o, hm, ht⟩
      · rw [cut_not_tool hw hc] at ht; cases ht

theorem resolveArgs_tool : ∀ {args : List Arg} {ps}, resolveArgs cutBase scope args = .ok ps →
    ∀ o' ∈ ps, ∀ b, o'.tool = some b → ∃ o ∈ scope, o.tool = some b
  | [], ps, h, o', ho, _, _ => by simp [resolveArgs] at h; subst h; simp at ho
  | a :: as, ps, h, o', ho, b, ht => by
    unfold resolveArgs at h
    cases ha : argObs cutBase scope a with
    | error _ => simp [ha] at h
    | ok o =>
      cases hr : resolveArgs cutBase scope as with
      | error _ => simp [ha, hr] at h
      | ok os =>
        simp only [ha, hr, Except.ok.injEq] at h
        subst h
        rcases List.mem_cons.mp ho with rfl | ho
        · exact argObs_tool ha ht
        · exact resolveArgs_tool hr o' ho b ht

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
theorem child_fewer (h : classify cutBase scope (.delegate t args) = .child t' ps) :
    (tools (childFrame ps).scope).length < (tools scope).length := by
  simp only [classify] at h
  cases hr : resolveArgs cutBase scope args with
  | error _ => simp [hr] at h
  | ok ps' =>
    simp only [hr] at h
    split at h
    · cases h; assumption
    · cases h

/-- A child holds only tools its emitter holds. -/
theorem child_sub (h : classify cutBase scope (.delegate t args) = .child t' ps) :
    ∀ b ∈ tools (childFrame ps).scope, b ∈ tools scope := by
  simp only [classify] at h
  cases hr : resolveArgs cutBase scope args with
  | error _ => simp [hr] at h
  | ok ps' =>
    simp only [hr] at h
    split at h
    · cases h
      intro b hb
      obtain ⟨o', ho', hb⟩ := mem_tools.mp hb
      obtain ⟨o, ho, hb'⟩ := resolveArgs_tool hr o' ho' b hb
      exact mem_tools.mpr ⟨o, ho, hb'⟩
    · cases h

/-- A frame runs only a binary it holds, as a tool in its scope. -/
theorem run_held (h : classify cutBase scope (.run t args) = .run b ps) : b ∈ tools scope := by
  simp only [classify] at h
  cases hg : get? scope t with
  | none => simp [hg] at h
  | some o =>
    simp only [hg] at h
    cases ht : o.tool with
    | none => simp [ht] at h
    | some b' =>
      simp only [ht] at h
      cases hr : resolveArgs cutBase scope args with
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
    classify cutBase scope (.run t args) = .reject (.notTool t) := by
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
