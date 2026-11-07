import Invariants.Walk

/-! Proposal 58: frames and children. A frame's tools are observations in
its scope, each keyed on its binary; the root is granted them, and nothing
that arrives is a tool (`Walk.tools_kept`). An instruction's first position
decides what it is, as `exec`'s first argument does:

* `return $n` returns `$n` to the frame's caller;
* a `$n` holding a tool runs that binary as a process, passing it and the
  rest like any invoke;
* `delegate(goal(a), …)` starts a child frame: its goal is taken from an
  observation, never written, and it is passed observations, tools among
  them, like any other. It is a child only when it passes fewer tool types
  than its emitter holds.

An observation's role is goal, tool or context. A frame's scope is its
goal, `$1`, then its tools, then its context: what it was passed and what
has arrived since, in order. Tools are fixed when a frame starts and
context only grows at the end, so a `$n` a learned guard names keeps its
meaning.

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

/-- An observation's role: what a frame is to do, what it can run, or what
it knows. -/
inductive Role where
  | goal
  | tool
  | context
  deriving DecidableEq, Repr

def role (o : Obs) : Role :=
  if o.goal then .goal else if o.tool.isSome then .tool else .context

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
  deriving DecidableEq, Repr

/-! ## Kinds: act and delegate

A frame's kind is counted from the tool types it holds, under the act
bound `k`: an act frame holds at most `k` and runs them; a delegate holds
more and hands work on. `k` is the owner's: it is set where the tree is
started and is never shown to a model, which sees only what its kind may
write. -/

inductive FrameKind where
  | act
  | delegate
  deriving DecidableEq, Repr

def kindOf (k : Nat) (scope : List Obs) : FrameKind :=
  if (tools scope).length ≤ k then .act else .delegate

/-- An instruction as the core reads it, by its first position: `return`,
a `$n` holding a tool, or `delegate` with the child's goal. A run's and a
delegation's `args` are the other arguments, in order. The tool a run names
is a whole `$n`. -/
inductive Instr where
  | ret (a : Arg)
  | run (tool : Nat) (args : List Arg)
  | delegate (goal : Arg) (args : List Arg)

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
  /-- A delegation passing no tool: a child acts through something. -/
  | noTools
  /-- An act frame runs its tools and returns: it does not hand work on. -/
  | actDelegates
  /-- A delegate hands work on and returns: it does not run tools. -/
  | delegateRuns
  /-- A goal taken from a tool: a goal is what to do, not what to run. -/
  | goalTool (n : Nat)
  deriving DecidableEq, Repr

inductive Kind where
  | ret (o : Obs)
  | run (binary : String) (passed : List Obs)
  /-- A child: its scope, its goal first, then its tools, then its context. -/
  | child (scope : List Obs)
  | reject (r : Reason)

/-- What a child is passed, in its roles' order: the tools, then the rest,
each in the order passed. Nothing passed is a goal: a goal passed is
context, and the child's goal is the one its delegation names. -/
def byRole (ps : List Obs) : List Obs :=
  let qs := ps.map fun o => { o with goal := false }
  qs.filter (·.tool.isSome) ++ qs.filter (·.tool.isNone)

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

variable (k : Nat)

def classify (scope : List Obs) : Instr → Kind
  | .ret a => match argObs cutBase scope a with
    | .ok o => .ret o
    | .error r => .reject r
  | .run t args =>
    if kindOf k scope = .delegate then .reject .delegateRuns else
    match get? scope t with
    | none => .reject (.unbound t)
    | some o => match o.tool with
      | none => .reject (.notTool t)
      | some b => match resolveArgs cutBase scope args with
        | .ok ps => .run b (o :: ps)
        | .error r => .reject r
  | .delegate g args =>
    if kindOf k scope = .act then .reject .actDelegates else
    match argObs cutBase scope g with
    | .error r => .reject r
    | .ok go =>
      if go.tool.isSome then .reject (.goalTool g.n) else
      match resolveArgs cutBase scope args with
      | .error r => .reject r
      | .ok ps =>
        if (tools ps).length = 0 then .reject .noTools
        else if (tools ps).length < (tools scope).length then .child ({ go with goal := true } :: byRole ps)
        else .reject (.narrows (tools ps).length (tools scope).length)

/-- A child frame: what it was passed is its scope, `$1` first. -/
def childFrame (ps : List Obs) : Frame := { scope := ps, pending := none }

variable {cutBase k}

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

theorem filterMap_tool_some : ∀ (ps : List Obs),
    (ps.filter (·.tool.isSome)).filterMap (·.tool) = ps.filterMap (·.tool)
  | [] => rfl
  | o :: ps => by
    cases h : o.tool <;> simp [List.filter, h, filterMap_tool_some ps]

theorem filterMap_tool_none : ∀ (ps : List Obs), (ps.filter (·.tool.isNone)).filterMap (·.tool) = []
  | [] => rfl
  | o :: ps => by
    cases h : o.tool <;> simp [List.filter, h, filterMap_tool_none ps]

/-- Ordering by role keeps what is passed: the same tools. -/
theorem tools_byRole : tools (byRole ps) = tools ps := by
  simp only [tools, byRole, List.filterMap_append, filterMap_tool_some, filterMap_tool_none,
    List.append_nil, List.filterMap_map]
  rfl

theorem tools_goal {o : Obs} (h : o.tool = none) : tools ({ o with goal := true } :: ps) = tools ps := by
  simp [tools, h]

/-- What a delegation that starts a child was: a delegate's, its goal taken
from an observation and not a tool, its arguments resolved, the child's
scope its goal, then what it passes in roles' order. -/
theorem child_inv (h : classify cutBase k scope (.delegate g args) = .child cs) :
    kindOf k scope = .delegate ∧ ∃ go ps, argObs cutBase scope g = .ok go ∧ go.tool = none ∧
      resolveArgs cutBase scope args = .ok ps ∧ cs = { go with goal := true } :: byRole ps ∧
      0 < (tools ps).length ∧ (tools ps).length < (tools scope).length := by
  simp only [classify] at h
  split at h
  · cases h
  · rename_i hk
    have hd : kindOf k scope = .delegate := by
      cases hq : kindOf k scope
      · exact absurd hq hk
      · rfl
    refine ⟨hd, ?_⟩
    cases hg : argObs cutBase scope g with
    | error _ => simp [hg] at h
    | ok go =>
      simp only [hg] at h
      split at h
      · cases h
      · rename_i ht
        have ht' : go.tool = none := by
          cases hq : go.tool
          · rfl
          · simp [hq] at ht
        cases hr : resolveArgs cutBase scope args with
        | error _ => simp [hr] at h
        | ok ps =>
          simp only [hr] at h
          split at h
          · cases h
          · rename_i h0
            split at h
            · rename_i hlt
              simp only [Kind.child.injEq] at h
              exact ⟨go, ps, rfl, ht', rfl, h.symm, Nat.pos_of_ne_zero h0, hlt⟩
            · cases h

/-- A child holds strictly fewer tool types than its emitter. -/
theorem child_fewer (h : classify cutBase k scope (.delegate g args) = .child cs) :
    (tools (childFrame cs).scope).length < (tools scope).length := by
  obtain ⟨-, go, ps, -, ht, -, rfl, -, hlt⟩ := child_inv h
  simpa [childFrame, tools_goal ht, tools_byRole] using hlt

/-- A child holds at least one tool: it acts through something. -/
theorem child_has_tool (h : classify cutBase k scope (.delegate g args) = .child cs) :
    0 < (tools (childFrame cs).scope).length := by
  obtain ⟨-, go, ps, -, ht, -, rfl, h0, -⟩ := child_inv h
  simpa [childFrame, tools_goal ht, tools_byRole] using h0

/-- A child holds only tools its emitter holds. -/
theorem child_sub (h : classify cutBase k scope (.delegate g args) = .child cs) :
    ∀ b ∈ tools (childFrame cs).scope, b ∈ tools scope := by
  obtain ⟨-, go, ps, -, ht, hr, rfl, -, -⟩ := child_inv h
  intro b hb
  simp only [childFrame, tools_goal ht, tools_byRole] at hb
  obtain ⟨o', ho', hb⟩ := mem_tools.mp hb
  obtain ⟨o, ho, hb'⟩ := resolveArgs_tool hr o' ho' b hb
  exact mem_tools.mpr ⟨o, ho, hb'⟩

/-- Only a delegate starts a child: it holds more tool types than the act
bound. -/
theorem child_from_delegate (h : classify cutBase k scope (.delegate g args) = .child cs) :
    k < (tools scope).length := by
  have hd := (child_inv h).1
  simp only [kindOf] at hd
  split at hd
  · cases hd
  · omega

/-- A child's `$1` is its goal, taken from an observation of its parent's
scope, whole or cut, never written and never a tool. -/
theorem child_goal (h : classify cutBase k scope (.delegate g args) = .child cs) :
    ∃ go, argObs cutBase scope g = .ok go ∧ go.tool = none ∧ cs.head? = some { go with goal := true } := by
  obtain ⟨-, go, ps, hg, ht, -, rfl, -, -⟩ := child_inv h
  exact ⟨go, hg, ht, rfl⟩

/-- A child's scope is in roles' order: its goal, then its tools, then its
context. -/
theorem child_roles (h : classify cutBase k scope (.delegate g args) = .child cs) :
    ∃ go ts cx, cs = go :: ts ++ cx ∧ role go = .goal ∧ (∀ o ∈ ts, role o = .tool) ∧ (∀ o ∈ cx, role o = .context) := by
  obtain ⟨-, go, ps, -, -, -, rfl, -, -⟩ := child_inv h
  refine ⟨_, _, _, rfl, by simp [role], ?_, ?_⟩
  · intro o ho
    obtain ⟨ho, ht⟩ := List.mem_filter.mp ho
    obtain ⟨o', -, rfl⟩ := List.mem_map.mp ho
    cases hq : o'.tool <;> simp_all [role]
  · intro o ho
    obtain ⟨ho, ht⟩ := List.mem_filter.mp ho
    obtain ⟨o', -, rfl⟩ := List.mem_map.mp ho
    simp only [Option.isNone_iff_eq_none] at ht
    simp [role, ht]

/-- An act frame never hands work on. -/
theorem act_never_delegates (hk : kindOf k scope = .act) :
    classify cutBase k scope (.delegate sp args) = .reject .actDelegates := by
  simp [classify, hk]

/-- A delegate never runs a tool itself. -/
theorem delegate_never_runs (hk : kindOf k scope = .delegate) :
    classify cutBase k scope (.run t args) = .reject .delegateRuns := by
  simp [classify, hk]

/-- A frame runs only a binary it holds, as a tool in its scope. -/
theorem run_held (h : classify cutBase k scope (.run t args) = .run b ps) : b ∈ tools scope := by
  simp only [classify] at h
  split at h
  · cases h
  · cases hg : get? scope t with
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

/-- In an act frame, a first position that is not a tool never runs. -/
theorem notTool_rejected (hk : kindOf k scope = .act) (hg : get? scope t = some o) (ht : o.tool = none) :
    classify cutBase k scope (.run t args) = .reject (.notTool t) := by
  simp [classify, hk, hg, ht]

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

/-- A chain of delegations: each frame but the last started the next, so
each holds more than the act bound, and fewer tool types than the one
before. -/
def Delegates (k : Nat) : List (List Obs) → Prop
  | [] => True
  | [_] => True
  | a :: b :: rest => k < (tools a).length ∧ (tools b).length < (tools a).length ∧ Delegates k (b :: rest)

/-- Under the act bound, delegation nests no deeper than the root's tool
types past `k`: a chain from the root is at most that many frames, and one
more for the act frame it ends in. -/
theorem delegates_depth : ∀ {root : List Obs} {rest}, Delegates k (root :: rest) →
    (root :: rest).length ≤ (tools root).length - k + 1
  | _, [], _ => by simp
  | root, b :: rest, h => by
    obtain ⟨hk, hlt, hn⟩ := h
    have := delegates_depth hn
    cases rest with
    | nil => simp only [List.length_cons, List.length_nil] at this ⊢; omega
    | cons c rest' =>
      have hkb : k < (tools b).length := hn.1
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
