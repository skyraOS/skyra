import Invariants.Refine
import Invariants.Shape

/-! Proposal 58: the walk. `sense` takes a frame and what its emitter sent,
an invoke or a return, and gives where the step stands: the base, the
values at the base's positions, and the scope table mapping the frame's
`$n` to those positions. An invoke walks forward to its input base; a
return walks back to the invoke it answers. The direction decides which
base, and nothing else: the same transition runs at either.

A frame holds at most one pending invoke. A return therefore answers the
frame's one pending invoke, and names nothing. Parallel work is several
frames, each with its own.

Everything here is pure. Lean computes no hash: `Hashes` are given. -/

namespace Walk

open Shape

/-- An observation in a frame's scope: its value, the hash of the base it
arrived at, and, for a tool, its binary. Tools are granted at the root;
nothing that arrives is a tool. A goal is what a frame is to do: its `$1`. -/
structure Obs where
  value : Value
  base : String
  tool : Option String := none
  goal : Bool := false

/-- A position of a base: what arrived there, or the k-th observation passed
to the invoke a return answers (or, at an input base, the invoke itself). -/
inductive Position where
  | arrived
  | passed (k : Nat)
  deriving DecidableEq, Repr

/-- The invoke a frame waits on: its input base, how its contract's output
reads, the `$n` it was passed and what they held. -/
structure Pending where
  input : String
  output : Output
  args : List Nat
  passed : List Obs

/-- A frame: its scope, `$1` first, appended to and never rewritten; and
its one pending invoke, if any. -/
structure Frame where
  scope : List Obs
  pending : Option Pending

/-- How the runtime hashes an input base (the contract and the bases of
what is passed, in order) and a return base (the input base it answers and
the encoding of the shape that came back). -/
structure Hashes where
  input : String → List String → String
  ret : String → String → String

inductive Event where
  /-- Invoke a contract, declared `json` or `text`, passing the frame's `$n`. -/
  | invoke (contract : String) (output : Output) (args : List Nat)
  /-- What the pending invoke's process gave: its stdout (`decoded` is the
  shell's JSON decoding of it, as in `Shape.arrive`), its exit status and
  its stderr. -/
  | ret (raw : ByteArray) (decoded : Except (Nat × Failure) Value) (exit : Nat) (stderr : ByteArray)
  /-- What the pending invoke's child frame returned: a value, already
  shaped, or the child's error observation. -/
  | returned (v : Value) (hs : (shapeOf v).isSome)

inductive Fault where
  /-- An invoke while one is pending: one at a time. -/
  | busy
  /-- A return with nothing pending: it answers nothing. A custody fault. -/
  | idle
  /-- `$n` names nothing in the frame's scope. -/
  | unbound (n : Nat)
  deriving DecidableEq, Repr

inductive Dir where
  | forward
  | back
  deriving DecidableEq, Repr

/-- Where the step stands. `table` maps `$n` to positions; the first entry
for a number is its position. -/
structure Arrival where
  dir : Dir
  base : String
  env : List (Position × Value)
  table : List (Nat × Position)
  next : Frame

/-! ## Scope -/

/-- `$n`: `$1` is the first observation; there is no `$0`. -/
def get? (scope : List Obs) : Nat → Option Obs
  | 0 => none
  | k + 1 => scope[k]?

def resolve (scope : List Obs) : List Nat → Except Fault (List Obs)
  | [] => .ok []
  | n :: ns => match get? scope n with
    | none => .error (.unbound n)
    | some o => match resolve scope ns with
      | .ok os => .ok (o :: os)
      | .error e => .error e

def passedTable : Nat → List Nat → List (Nat × Position)
  | _, [] => []
  | k, n :: ns => (n, .passed k) :: passedTable (k + 1) ns

def passedEnv : Nat → List Obs → List (Position × Value)
  | _, [] => []
  | k, o :: os => (.passed k, o.value) :: passedEnv (k + 1) os

/-- The position of `$n`: its first entry. -/
def find (n : Nat) : List (Nat × Position) → Option Position
  | [] => none
  | (m, p) :: t => if m = n then some p else find n t

/-- The `$n` at a position: the first entry holding it. -/
def back (p : Position) : List (Nat × Position) → Option Nat
  | [] => none
  | (n, q) :: t => if q = p then some n else back p t

def valueAt (p : Position) : List (Position × Value) → Option Value
  | [] => none
  | (q, v) :: t => if q = p then some v else valueAt p t

/-! ## The walk -/

/-- Forward: the invoke's base, its passed positions, and the frame waiting. -/
def forwardTo (h : Hashes) (f : Frame) (c : String) (o : Output) (args : List Nat)
    (passed : List Obs) : Arrival :=
  let base := h.input c (passed.map (·.base))
  { dir := .forward, base, env := passedEnv 0 passed, table := passedTable 0 args,
    next := { f with pending := some ⟨base, o, args, passed⟩ } }

/-- Back: the return base of the pending invoke and the shape that came
back; the value appended, never a tool; nothing pending. -/
def backTo (h : Hashes) (f : Frame) (p : Pending) (v : Value) (s : Shape) : Arrival :=
  let base := h.ret p.input s.enc
  { dir := .back, base,
    env := (.arrived, v) :: passedEnv 0 p.passed,
    table := (f.scope.length + 1, .arrived) :: passedTable 0 p.args,
    next := { scope := f.scope ++ [{ value := v, base }], pending := none } }

def sense (h : Hashes) (f : Frame) : Event → Except Fault Arrival
  | .invoke c o args => match f.pending with
    | some _ => .error .busy
    | none => match resolve f.scope args with
      | .error e => .error e
      | .ok passed => .ok (forwardTo h f c o args passed)
  | .ret raw decoded exit err => match f.pending with
    | none => .error .idle
    | some p =>
      let v := arriveExit p.output raw decoded exit err
      .ok (backTo h f p v ((shapeOf v).get (arriveExit_shaped p.output raw decoded exit err)))
  | .returned v hs => match f.pending with
    | none => .error .idle
    | some p => .ok (backTo h f p v ((shapeOf v).get hs))

/-- Every step is a forward walk from a frame waiting on nothing, or a back
walk to the frame's pending invoke. -/
theorem sense_ok (hs : sense h f e = .ok a) :
    (∃ c o args passed, e = .invoke c o args ∧ f.pending = none ∧
      resolve f.scope args = .ok passed ∧ a = forwardTo h f c o args passed) ∨
    (∃ p v s, e ≠ .invoke c' o' args' ∧ f.pending = some p ∧ a = backTo h f p v s ∧
      shapeOf v = some s) := by
  cases e with
  | invoke c o args =>
    unfold sense at hs
    cases hp : f.pending with
    | some _ => simp [hp] at hs
    | none =>
      simp only [hp] at hs
      cases hr : resolve f.scope args with
      | error _ => simp [hr] at hs
      | ok passed =>
        simp only [hr, Except.ok.injEq] at hs
        exact .inl ⟨c, o, args, passed, rfl, rfl, hr, hs.symm⟩
  | ret raw d x err =>
    unfold sense at hs
    cases hp : f.pending with
    | none => simp [hp] at hs
    | some p =>
      simp only [hp, Except.ok.injEq] at hs
      exact .inr ⟨p, _, _, by simp, rfl, hs.symm, by simp⟩
  | returned v hv =>
    unfold sense at hs
    cases hp : f.pending with
    | none => simp [hp] at hs
    | some p =>
      simp only [hp, Except.ok.injEq] at hs
      exact .inr ⟨p, v, _, by simp, rfl, hs.symm, by simp⟩

/-- The step at the arrival: the base's current refinement, distinguished on
this execution's values. The direction is not read. -/
def step {A I : Type} [DecidableEq I] (refs : String → Refine.Refinement A I)
    (sat : A → List (Position × Value) → Bool) (a : Arrival) : Refine.Outcome I :=
  Refine.distinguish sat (refs a.base).branches a.env

theorem step_dir {A I : Type} [DecidableEq I] (refs : String → Refine.Refinement A I)
    (sat : A → List (Position × Value) → Bool) (a : Arrival) (d : Dir) :
    step refs sat { a with dir := d } = step refs sat a := rfl

/-! ## One at a time -/

theorem invoke_busy (h : Hashes) (hp : f.pending = some p) :
    sense h f (.invoke c o args) = .error .busy := by
  simp [sense, hp]

theorem return_idle (h : Hashes) (hp : f.pending = none) :
    sense h f (.ret raw d x err) = .error .idle := by
  simp [sense, hp]

theorem returned_idle (h : Hashes) (hp : f.pending = none) :
    sense h f (.returned v hv) = .error .idle := by
  simp [sense, hp]

/-- An invoke walks forward, waits, and leaves the scope as it was. -/
theorem invoke_waits (hs : sense h f (.invoke c o args) = .ok a) :
    a.dir = .forward ∧ f.pending = none ∧ a.next.scope = f.scope ∧
      ∃ passed, resolve f.scope args = .ok passed ∧
        a.base = h.input c (passed.map (·.base)) ∧
        a.next.pending = some ⟨a.base, o, args, passed⟩ := by
  rcases sense_ok (c' := c) (o' := o) (args' := args) hs with
    ⟨c₁, o₁, args₁, passed, he, hp, hr, rfl⟩ | ⟨_, _, _, hne, -⟩
  · cases he; exact ⟨rfl, hp, rfl, passed, hr, rfl, rfl⟩
  · exact absurd rfl hne

/-- A waiting frame cannot invoke again until the return. -/
theorem one_at_a_time (hs : sense h f (.invoke c o args) = .ok a) :
    sense h a.next (.invoke c' o' args') = .error .busy := by
  obtain ⟨-, -, -, _, -, -, hp⟩ := invoke_waits hs
  exact invoke_busy h hp

/-- A return walks back to the invoke it answers: its base is the return
base of the pending invoke's input base and the shape that came back; the
value is appended to the scope, and nothing is pending. -/
theorem return_answers (hs : sense h f (.ret raw d x err) = .ok a) :
    a.dir = .back ∧ ∃ p, f.pending = some p ∧
      a.base = h.ret p.input ((shapeOf (arriveExit p.output raw d x err)).get
        (arriveExit_shaped p.output raw d x err)).enc ∧
      a.next.pending = none ∧
      a.next.scope = f.scope ++ [{ value := arriveExit p.output raw d x err, base := a.base }] := by
  unfold sense at hs
  cases hp : f.pending with
  | none => simp [hp] at hs
  | some p =>
    simp only [hp, Except.ok.injEq] at hs
    subst hs
    exact ⟨rfl, p, rfl, rfl, rfl, rfl⟩

/-- A child's return walks back the same way: to the parent's pending
invoke, at the return base of the shape of the value returned. -/
theorem returned_answers (hs : sense h f (.returned v hv) = .ok a) :
    a.dir = .back ∧ ∃ p, f.pending = some p ∧
      a.base = h.ret p.input ((shapeOf v).get hv).enc ∧ a.next.pending = none ∧
      a.next.scope = f.scope ++ [{ value := v, base := a.base }] := by
  unfold sense at hs
  cases hp : f.pending with
  | none => simp [hp] at hs
  | some p =>
    simp only [hp, Except.ok.injEq] at hs
    subst hs
    exact ⟨rfl, p, rfl, rfl, rfl, rfl⟩

/-- `$n` never changes what it names: a step only appends. -/
theorem scope_prefix (hs : sense h f e = .ok a) : f.scope <+: a.next.scope := by
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, _, -, -, -, rfl⟩ | ⟨_, _, _, -, -, rfl, -⟩
  · exact List.prefix_refl _
  · exact List.prefix_append _ _

/-- Nothing that arrives is a tool: a frame's tools are those it started
with. -/
theorem tools_kept (hs : sense h f e = .ok a) :
    ∀ o ∈ a.next.scope, o.tool.isSome → o ∈ f.scope := by
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, _, -, -, -, rfl⟩ | ⟨_, _, _, -, -, rfl, -⟩
  · intro o ho _; exact ho
  · intro o ho ht
    simp only [backTo, List.mem_append, List.mem_singleton] at ho
    rcases ho with ho | rfl
    · exact ho
    · simp at ht

/-- What arrives is context: a step keeps the frame's scope, and what it
appends is neither a goal nor a tool. A frame's goal and tools are those it
started with. -/
theorem arrives_context (hs : sense h f e = .ok a) :
    ∃ more, a.next.scope = f.scope ++ more ∧ ∀ o ∈ more, o.tool = none ∧ o.goal = false := by
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, _, -, -, -, rfl⟩ | ⟨_, _, _, -, -, rfl, -⟩
  · exact ⟨[], by simp [forwardTo], by simp⟩
  · exact ⟨_, rfl, by simp⟩

/-! ## Locations are general -/

/-- An invoke's base depends on the contract and the bases of what is
passed, never their values: executions passing different values from the
same bases stand at one location. -/
theorem invoke_general (h1 : sense h f (.invoke c o args) = .ok a)
    (h2 : sense h f' (.invoke c o' args') = .ok a')
    (hb : ∀ ps ps', resolve f.scope args = .ok ps → resolve f'.scope args' = .ok ps' →
      ps.map (·.base) = ps'.map (·.base)) :
    a.base = a'.base := by
  obtain ⟨-, -, -, ps, hr, hb1, -⟩ := invoke_waits h1
  obtain ⟨-, -, -, ps', hr', hb2, -⟩ := invoke_waits h2
  rw [hb1, hb2, hb ps ps' hr hr']

/-! ## The scope table is right -/

/-- A frame's pending invoke holds what its `$n` held: true from the
invoke on, since nothing is appended while it waits. -/
def Frame.Valid (f : Frame) : Prop :=
  ∀ p, f.pending = some p → resolve f.scope p.args = .ok p.passed

theorem sense_valid (hs : sense h f e = .ok a) : a.next.Valid := by
  intro p hp
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, ps, -, -, hr, rfl⟩ | ⟨_, _, _, -, -, rfl, -⟩
  · simp only [forwardTo] at hp; cases hp; exact hr
  · simp [backTo] at hp

theorem resolve_get {scope : List Obs} : ∀ {args : List Nat} {ps : List Obs},
    resolve scope args = .ok ps → ∀ {k n : Nat}, args[k]? = some n → ∃ o, get? scope n = some o ∧ ps[k]? = some o
  | [], ps, _, k, n, h => by simp at h
  | m :: ms, ps, hr, k, n, h => by
    unfold resolve at hr
    cases hg : get? scope m with
    | none => simp [hg] at hr
    | some o =>
      cases hr' : resolve scope ms with
      | error _ => simp [hg, hr'] at hr
      | ok os =>
        simp only [hg, hr', Except.ok.injEq] at hr
        subst hr
        cases k with
        | zero => simp at h; subst h; exact ⟨o, hg, rfl⟩
        | succ k => simpa using resolve_get hr' (by simpa using h)

theorem find_passed : ∀ {args : List Nat} {k n p}, find n (passedTable k args) = some p →
    ∃ j, p = .passed (k + j) ∧ args[j]? = some n
  | [], _, _, _, h => by simp [passedTable, find] at h
  | m :: ms, k, n, p, h => by
    simp only [passedTable, find] at h
    split at h
    · rename_i e; cases h; exact ⟨0, by simp, by simp [e]⟩
    · obtain ⟨j, rfl, hj⟩ := find_passed h
      exact ⟨j + 1, by simp [Nat.add_assoc, Nat.add_comm 1], by simpa using hj⟩

theorem valueAt_passed : ∀ {ps : List Obs} {k j}, valueAt (.passed (k + j)) (passedEnv k ps) =
    ps[j]?.map (·.value)
  | [], _, _ => by simp [passedEnv, valueAt]
  | o :: os, k, j => by
    simp only [passedEnv, valueAt]
    cases j with
    | zero => simp
    | succ j =>
      have : (Position.passed k = Position.passed (k + (j + 1))) = False := by simp
      simp only [this, ite_false]
      have e : k + (j + 1) = (k + 1) + j := by omega
      rw [e, valueAt_passed]; simp

theorem get_append (h : get? scope n = some o) : get? (scope ++ more) n = some o := by
  cases n with
  | zero => simp [get?] at h
  | succ k =>
    simp only [get?] at h ⊢
    have hk : k < scope.length := by
      rcases Nat.lt_or_ge k scope.length with hk | hk
      · exact hk
      · simp [List.getElem?_eq_none hk] at h
    rw [List.getElem?_append_left hk, h]

/-- The scope table is right: the position a `$n` maps to holds, at this
base, the value `$n` names in the frame. -/
theorem table_env (hv : f.Valid) (hs : sense h f e = .ok a) (hf : find n a.table = some p) :
    valueAt p a.env = (get? a.next.scope n).map (·.value) := by
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, ps, -, -, hr, rfl⟩ | ⟨pd, v, s, -, hp, rfl, -⟩
  · obtain ⟨j, rfl, hj⟩ := find_passed hf
    obtain ⟨o', hg, hpj⟩ := resolve_get hr hj
    simp only [forwardTo]
    rw [valueAt_passed, hpj]; simp [hg]
  · simp only [backTo, find] at hf
    split at hf
    · rename_i e; cases hf; subst e
      simp [backTo, valueAt, get?]
    · obtain ⟨j, rfl, hj⟩ := find_passed hf
      obtain ⟨o', hg, hpj⟩ := resolve_get (hv pd hp) hj
      simp only [backTo, valueAt, reduceCtorEq, ite_false, Nat.zero_add]
      have := valueAt_passed (ps := pd.passed) (k := 0) (j := j)
      simp only [Nat.zero_add] at this
      rw [this, hpj, get_append hg]

/-! ## Translation: `$n` to positions and back -/

/-- An instruction template: text, and references to the frame's `$n`. -/
inductive Tok where
  | lit (s : String)
  | ref (n : Nat)
  deriving DecidableEq, Repr

/-- Its stored, base-relative form: `@arrived`, `@passed[k]`. -/
inductive Stored where
  | lit (s : String)
  | pos (p : Position)
  deriving DecidableEq, Repr

/-- Store a template under a scope table. A `$n` the table lacks is not in
the base: the template runs once, unlearned. -/
def store (t : List (Nat × Position)) : List Tok → Except Nat (List Stored)
  | [] => .ok []
  | .lit s :: xs => (Stored.lit s :: ·) <$> store t xs
  | .ref n :: xs => match find n t with
    | none => .error n
    | some p => (Stored.pos p :: ·) <$> store t xs

/-- Bind a stored template in a frame: each position becomes the `$n` that
holds it there. -/
def bind (t : List (Nat × Position)) : List Stored → Option (List Tok)
  | [] => some []
  | .lit s :: xs => (Tok.lit s :: ·) <$> bind t xs
  | .pos p :: xs => match back p t with
    | none => none
    | some n => (Tok.ref n :: ·) <$> bind t xs

theorem back_find : ∀ {t : List (Nat × Position)}, (t.map (·.2)).Nodup →
    find n t = some p → back p t = some n
  | [], _, h => by simp [find] at h
  | (m, q) :: t, hu, h => by
    simp only [List.map_cons, List.nodup_cons] at hu
    simp only [find] at h
    simp only [back]
    split at h
    · rename_i e; cases h; simp [e]
    · have hq : q ≠ p := by
        rintro rfl
        obtain ⟨j, hj⟩ : ∃ x ∈ t, x.2 = q := by
          clear hu
          induction t with
          | nil => simp [find] at h
          | cons x t ih =>
            simp only [find] at h
            split at h
            · cases h; exact ⟨x, List.mem_cons_self .., rfl⟩
            · obtain ⟨y, hy, e⟩ := ih h; exact ⟨y, List.mem_cons_of_mem _ hy, e⟩
        exact hu.1 (List.mem_map.mpr ⟨j, hj.1, hj.2⟩)
      simp only [hq, ite_false]
      exact back_find hu.2 h

/-- Storing then binding in the same frame gives the template back. -/
theorem bind_store (hu : (t.map (·.2)).Nodup) :
    ∀ {xs : List Tok} {ys}, store t xs = .ok ys → bind t ys = some xs
  | [], ys, h => by simp [store] at h; subst h; rfl
  | .lit s :: xs, ys, h => by
    simp only [store] at h
    cases hx : store t xs with
    | error _ => simp [hx] at h; cases h
    | ok zs =>
      simp only [hx] at h; cases h
      simp [bind, bind_store hu hx]
  | .ref n :: xs, ys, h => by
    simp only [store] at h
    cases hf : find n t with
    | none => simp [hf] at h
    | some p =>
      cases hx : store t xs with
      | error _ => simp [hf, hx] at h; cases h
      | ok zs =>
        simp only [hf, hx] at h; cases h
        simp [bind, back_find hu hf, bind_store hu hx]

theorem passedTable_pos : ∀ {k : Nat} {args : List Nat} {q},
    q ∈ (passedTable k args).map (·.2) → ∃ j, q = .passed (k + j)
  | _, [], _, h => by simp [passedTable] at h
  | k, m :: ms, q, h => by
    simp only [passedTable, List.map_cons, List.mem_cons] at h
    rcases h with rfl | h
    · exact ⟨0, by simp⟩
    · obtain ⟨j, rfl⟩ := passedTable_pos h; exact ⟨j + 1, by simp; omega⟩

theorem passedTable_nodup : ∀ {k : Nat} {args : List Nat}, ((passedTable k args).map (·.2)).Nodup
  | _, [] => by simp [passedTable]
  | k, m :: ms => by
    simp only [passedTable, List.map_cons, List.nodup_cons]
    refine ⟨fun h => ?_, passedTable_nodup⟩
    obtain ⟨j, e⟩ := passedTable_pos h
    simp at e; omega

/-- No two `$n` share a position, so every table binds back. -/
theorem table_nodup (hs : sense h f e = .ok a) : (a.table.map (·.2)).Nodup := by
  rcases sense_ok (c' := "") (o' := .text) (args' := []) hs with
    ⟨_, _, _, _, -, -, -, rfl⟩ | ⟨_, _, _, -, -, rfl, -⟩
  · exact passedTable_nodup
  · simp only [backTo, List.map_cons, List.nodup_cons]
    refine ⟨fun h => ?_, passedTable_nodup⟩
    obtain ⟨j, e⟩ := passedTable_pos h; cases e

/-- What a step learns binds back in the frame that learned it. -/
theorem learned_binds (hs : sense h f e = .ok a) (hx : store a.table xs = .ok ys) :
    bind a.table ys = some xs :=
  bind_store (table_nodup hs) hx

end Walk
