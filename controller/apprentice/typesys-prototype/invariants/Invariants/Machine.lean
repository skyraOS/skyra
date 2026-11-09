/-
The machine: 61's runtime, composed of what Model.lean defines, as the Go
shell runs it (model/machine.go). It is pure: each step takes one input
and gives one effect for the shell to perform (run a leaf, ask the model,
or end).

Genesis pushes the first frame: the goal as `$1`, then the program's
tools. At each frame its key is decided: an unseen key is hesitated,
recorded with nothing committed; the tally over the templates stored
there selects one, or startles the model, whose chain is committed as one
more implementation and its own template run. Running a template records
what it eliminated and introduced in the deciding frame's state, and its
operator is one tree edge: an act pushes a leaf the world runs, a
delegation a child, a return pops into the parent. A bound exhausted pops
too, with an error observation.

Conditions here are abstract (`Refine.Cond A`), already over the key's
positions, as memory stores them; the shell's CEL, and what it refuses
before a condition is stored, are the shell's. The loop runs on fuel, so
it is a function: a run that needs more steps than its fuel gives none.
-/
import Invariants.Model

namespace Machine

open Model Shape

/-! ## Inputs and effects -/

/-- A template as the model writes it: each argument its pieces, written
together. One piece is the argument itself; two or more are a join. -/
structure Written where
  op : Op
  args : List (List Arg)

inductive Input (A : Type) where
  /-- What a leaf's process gave, or the goal. -/
  | answered (v : Value)
  /-- The chain the model wrote when asked. -/
  | wrote (chain : List (Refine.Cond A × Written))

/-- Why the model is asked again: what it wrote cannot run in the frame;
a literal is text a cut of `$n` takes; a join is refused; an argument
reads nothing seen; the chain selects nothing for the values it was
written for; or an argument of the template run does not read. -/
inductive Refusal where
  | fits
  | copies (n : Nat)
  | join (r : JoinRefusal)
  | unstored
  | unselected
  | unread
  deriving DecidableEq, Repr

inductive Effect where
  | runLeaf (tool : DefLoc) (args : List Value)
  | ask (k : Key) (kind : Kind) (depth : Nat) (refused : Option Refusal)
  | finished (v : Value)
  | exhausted (v : Value)

inductive Phase where
  | start | ask | run | done
  deriving DecidableEq, Repr

/-- A run: memory, which outlives it, and the log that folds to it; the
tree; the frame deciding or the leaf running; the asks of the transition
in progress; and each frame's transitions. -/
structure State (A : Type) where
  memory : Memory A
  log : List (Event A)
  tree : Tree
  cur : InstLoc
  phase : Phase
  asks : Nat
  taken : InstLoc → Nat

def State.init {A : Type} (log : List (Event A)) : State A :=
  ⟨replay log, log, Tree.mk [] [], ⟨""⟩, .start, 0, fun _ => 0⟩

/-- What a run is of: its ids, its program and memory's root (the program
without its constants), the tally's policy, and when an atom holds of the
values a frame has seen. -/
structure Env (A K : Type) where
  ids : Ids
  program : Program
  root : DefLoc
  policy : Refine.Policy Template K
  sat : A → List Value → Bool

/-! ## What the machine reads -/

/-- What a frame has seen as a guard reads it, by key position: a tool as
its name, a value as itself. -/
def vals (f : Frame) : List Value :=
  (seenAt f).map fun n => match item f n with
    | some (.tool b) => .str b.hash
    | some (.seen v) => v
    | _ => .null

def depthOf (t : Tree) : Nat → InstLoc → Nat
  | 0, _ => 0
  | n + 1, id => match t.find id with
    | some { parent := some p, .. } => depthOf t n p + 1
    | _ => 0

/-- How deep a frame is: the first frame is 0. -/
def depth (t : Tree) (id : InstLoc) : Nat := depthOf t t.frames.length id

def errorValue (msg : String) : Value := .obj [("error", .bytes msg.toUTF8)]

/-- What a return gives its caller: its exit status, the value it returns
as stdout, and its why as stderr when it names one. -/
def returnValue (exit out : Value) (why : List Value) : Value :=
  .obj ([("exit", exit)] ++ (why.head?.map fun w => [("stderr", w)]).getD [] ++ [("stdout", out)])

def toolAt (f : Frame) : Arg → Option DefLoc
  | .ref (.var n _) => match item f n with
    | some (.tool b) => some b
    | _ => none
  | _ => none

/-! ## Fitting, recording, making -/

def argFits (f : Frame) (op : Op) (i : Nat) : Arg → Bool
  | .lit _ => !(op == .act && i == 0)
  | .join _ => !(op == .act && i == 0)
  | .ref (.name _) => false
  | .ref (.var n sels) =>
    1 ≤ n && n ≤ (items f).length &&
      match item f n with
      | some (.tool _) => sels.isEmpty && ((op == .act && i == 0) || (op == .delegate && 0 < i))
      | _ => !(op == .act && i == 0)

def passedTools (f : Frame) (args : List Arg) : List DefLoc := (args.filterMap (toolAt f)).eraseDups

/-- Whether a template can run in a frame, whatever its values: its
operator is one the frame's kind emits, a tool is run from a `$n` holding
one, only values are arguments, and a delegation passes at least one tool
and fewer than the frame holds. -/
def fits (p : Program) (f : Frame) (t : Template) : Bool :=
  decide (t.op ∈ ops f.scope.kind) &&
    (t.args.zipIdx.all fun (a, i) => argFits f t.op i a) &&
    (t.op != .delegate ||
      (!(p.reach "delegate").isEmpty && 0 < (passedTools f t.args.tail).length &&
        (passedTools f t.args.tail).length < (toolTypes (p.reach "act") f.scope.entries).length))

/-- What running a template records: each literal an introduction, each
join its cuts then itself, each cut an elimination, in argument order. -/
def recorded (t : Template) : List Entry :=
  t.args.flatMap fun
    | .lit l => [.introduced l]
    | .join ps =>
      (ps.filterMap fun | .seen n s => if s.isEmpty then none else some (Entry.eliminated n s) | _ => none) ++
        [.joined ps]
    | .ref (.var n sels) => if sels.isEmpty then [] else [.eliminated n sels]
    | _ => []

def record (t : Tree) (id : InstLoc) (es : List Entry) : Tree :=
  ⟨t.frames.map fun g => if g.id = id then g.append es else g, t.closed⟩

/-- A written template made: each join's parts made (`joinParts`). -/
def made (f : Frame) (w : Written) : Except Refusal Template := do
  let args ← w.args.mapM fun
    | [a] => pure a
    | as => match joinParts f as with
      | .ok ps => pure (.join ps)
      | .error r => throw (.join r)
  pure ⟨w.op, args⟩

/-- The first literal that is text a cut takes, by the `$n` it names. -/
def copied (f : Frame) (t : Template) : Option Nat :=
  t.args.findSome? fun
    | .lit (.str s) => match introduce f (.str s) with
      | .error n => some n
      | .ok _ => none
    | _ => none

/-- One link stored: made, fitted, its literals checked, then stored. -/
def storeLink (p : Program) (f : Frame) (w : Written) : Except Refusal Template := do
  let t ← made f w
  if !fits p f t then throw .fits
  if let some n := copied f t then throw (.copies n)
  match t.store f with
  | some st => pure st
  | none => throw .unstored

/-- Each argument's value, a tool's none. -/
def argValues (f : Frame) (t : Template) : Option (List Value) :=
  t.args.mapM fun a => if (toolAt f a).isSome then some .null else evalArg f a

/-- A leaf's argument, bound at its slot. -/
def leafMember (ids : Ids) (leaf : InstLoc) (p : Value × Nat) : Option Member :=
  (arrive p.1 (ids.slot leaf (p.2 + 2))).map Member.obs

/-- What a delegation passes at each position: the tool a `$n` holds, past
the goal, or none for a value. -/
def passing (f : Frame) (t : Template) : List (Option DefLoc) :=
  t.args.zipIdx.map fun (a, i) => if i = 0 then none else toolAt f a

def passedIface (p : Option DefLoc × Value) : Option Iface :=
  match p.1 with
  | some b => some (.defn b)
  | none => (shapeOf p.2).map .shape

def passedMember (ids : Ids) (child : InstLoc) (p : (Option DefLoc × Value) × Nat) : Option Member :=
  match p.1.1 with
  | some b => some (.tool b)
  | none => (arrive p.1.2 (ids.slot child (p.2 + 1))).map .obs

def shapeIface (v : Value) : Option Iface := (shapeOf v).map .shape

def storeLinkOf {A : Type} (p : Program) (f : Frame) (l : Refine.Cond A × Written) :
    Except Refusal (Refine.Cond A × Template) :=
  (storeLink p f l.2).map (l.1, ·)

/-! ## The loop -/

inductive Call where
  | decide (id : InstLoc)
  | trust (id : InstLoc) (st : Template)
  | finish (id : InstLoc) (v : Value) (exhausted : Bool)
  | ask (r : Option Refusal)

section
variable {A K : Type} [DecidableEq K]

def reaches (E : Env A K) : DefLoc → String → List DefLoc := fun _ name => E.program.reach name

/-- The key a frame decides at: its key, each definition by its class. -/
def keyOf (E : Env A K) (f : Frame) : Key := E.program.classKey (key (fun _ => E.root) f)

/-- An act's edge: the leaf pushed, running the tool on the values past
it; the leaf, the tree, the tool and its arguments. -/
def actEdge (E : Env A K) (tr : Tree) (id : InstLoc) (f : Frame) (t : Template) (vs : List Value) :
    Option (InstLoc × Tree × DefLoc × List Value) := do
  let tool ← t.args.head?.bind (toolAt f)
  let args := vs.tail
  let ifs ← args.mapM shapeIface
  let sc : Scope := ⟨.leaf, .defn tool :: ifs⟩
  let leaf := E.ids.child id (tr.invokes id) sc
  let bs ← args.zipIdx.mapM (leafMember E.ids leaf)
  let tr ← tr.step E.ids E.program.constants.act (reaches E) id (.invoke .act sc (.tool tool :: bs) none)
  pure (leaf, tr, tool, args)

/-- A delegation's edge: the child pushed, of what it passes, its kind
counted under the act bound, consulting the frame's model or the
program's first. -/
def delegateEdge (E : Env A K) (tr : Tree) (id : InstLoc) (f : Frame) (t : Template) (vs : List Value) :
    Option (InstLoc × Tree) := do
  let ifs ← ((passing f t).zip vs).mapM passedIface
  let kind ← agentKind E.program.constants.act (E.program.reach "act") ifs
  let sc : Scope := ⟨kind, ifs⟩
  let child := E.ids.child id (tr.invokes id) sc
  let bs ← ((passing f t).zip vs).zipIdx.mapM (passedMember E.ids child)
  let tr ← tr.step E.ids E.program.constants.act (reaches E) id
    (.invoke .delegate sc bs (f.model.orElse fun _ => (E.program.reach "delegate").head?))
  pure (child, tr)

/-- A return's edge: the value landed in the parent, at its next `$n`. -/
def popEdge (E : Env A K) (tr : Tree) (id pid : InstLoc) (v : Value) : Option Tree := do
  let p ← tr.find pid
  let o ← arrive v (E.ids.slot pid ((items p).length + 1))
  tr.step E.ids E.program.constants.act (reaches E) id (.ret o)

def go (E : Env A K) : Nat → Call → State A → Option (State A × Effect)
  | 0, _, _ => none
  | fuel + 1, .decide id, s => match s.tree.find id with
    | none => none
    | some f =>
      let k := keyOf E f
      let fresh := (s.memory k).isNone
      let s := { s with
        cur := id
        asks := 0
        memory := if fresh then s.memory.apply (.hesitate k) else s.memory
        log := if fresh then s.log ++ [.hesitate k] else s.log }
      match decideAt E.policy E.program.constants.quorum E.sat s.memory k (vals f) with
      | .select st => go E fuel (.trust id st) s
      | _ => go E fuel (.ask none) s
  | fuel + 1, .ask r, s => match s.tree.find s.cur with
    | none => none
    | some f =>
      if E.program.constants.refinements ≤ s.asks then
        go E fuel (.finish f.id (errorValue s!"refinement bound exhausted: the model was asked {s.asks} times at this key without a guard that runs") true) s
      else some ({ s with asks := s.asks + 1, phase := .ask }, .ask (keyOf E f) f.scope.kind (depth s.tree f.id) r)
  | fuel + 1, .trust id st, s => match s.tree.find id with
    | none => none
    | some f => match st.bind f with
      | none => none
      | some t =>
        if !fits E.program f t then go E fuel (.ask (some .fits)) s else
        match argValues f t with
        | none => go E fuel (.ask (some .unread)) s
        | some vs =>
          let T := E.program.constants.transitions
          if T ≤ s.taken id then
            go E fuel (.finish id (errorValue s!"transition bound exhausted: {T} transitions") true) s
          else
            let s := { s with
              taken := fun x => if x = id then s.taken x + 1 else s.taken x
              tree := record s.tree id (recorded t) }
            match s.tree.find id with
            | none => none
            | some f => match t.op with
              | .ret => match vs with
                | e :: out :: why => go E fuel (.finish id (returnValue e out why) false) s
                | _ => none
              | .act => match actEdge E s.tree id f t vs with
                | none => none
                | some r => some ({ s with tree := r.2.1, cur := r.1, phase := .run }, .runLeaf r.2.2.1 r.2.2.2)
              | .delegate => (delegateEdge E s.tree id f t vs).bind fun r => go E fuel (.decide r.1) { s with tree := r.2 }
              | .boot => none
  | fuel + 1, .finish id v ex, s => match s.tree.find id with
    | none => none
    | some f => match f.parent with
      | none => some ({ s with phase := .done }, if ex then .exhausted v else .finished v)
      | some pid => (popEdge E s.tree id pid v).bind fun tr => go E fuel (.decide pid) { s with tree := tr }

/-- The model wrote a chain: each link stored, then the chain committed as
one more implementation at the frame's key and its own template run, or
the model asked again. -/
def wrote (E : Env A K) (fuel : Nat) (chain : List (Refine.Cond A × Written)) (s : State A) : Option (State A × Effect) :=
  match s.tree.find s.cur with
  | none => none
  | some f =>
    match chain.mapM (storeLinkOf E.program f) with
    | .error r => go E fuel (.ask (some r)) s
    | .ok stored =>
      let k := keyOf E f
      let e : Event A := .startle k ⟨stored⟩
      let s := { s with memory := s.memory.apply e, log := s.log ++ [e] }
      match Refine.afterStartle E.policy E.sat stored (((s.memory k).map (·.branches)).getD []) (vals f) with
      | .startleAgain => go E fuel (.ask (some .unselected)) s
      | .trust st _ => go E fuel (.trust f.id st) s

/-- Genesis: the first frame, the goal then the tools, its kind counted
under the act bound. -/
def genesis (E : Env A K) (fuel : Nat) (goal : Value) (s : State A) : Option (State A × Effect) :=
  let tools := E.program.reach "act"
  let loc := E.ids.program E.program
  let id := E.ids.first loc ⟨tools, []⟩
  match arrive goal (E.ids.slot id 1) with
  | none => none
  | some o =>
    let ifs := o.iface :: tools.map Iface.defn
    match agentKind E.program.constants.act tools ifs with
    | none => none
    | some kind =>
      let f : Frame := ⟨id, loc, none, ⟨kind, ifs⟩, .obs o :: tools.map Member.tool, [],
        (E.program.reach "delegate").head?⟩
      go E fuel (.decide id) { s with tree := ⟨[f], []⟩ }

/-- One step: an input, the next effect; none for an input the machine is
not waiting for, or a run past its fuel. -/
def step (E : Env A K) (fuel : Nat) (s : State A) : Input A → Option (State A × Effect)
  | .answered v => match s.phase with
    | .start => genesis E fuel v s
    | .run => go E fuel (.finish s.cur v false) s
    | _ => none
  | .wrote chain => if s.phase = .ask then wrote E fuel chain s else none

end

/-! ## What the machine keeps

What a template records is live: a frame's key changes only by what
arrives. -/

theorem arrivals_nil : ∀ {es : List Entry}, (∀ e ∈ es, ∀ o, e ≠ .arrived o) → arrivals es = []
  | [], _ => rfl
  | e :: es, h => by
    have ih := arrivals_nil (fun e he => h e (List.mem_cons_of_mem _ he))
    cases e with
    | arrived o => exact absurd rfl (h _ List.mem_cons_self o)
    | eliminated => simp [arrivals, ih]
    | introduced => simp [arrivals, ih]
    | joined => simp [arrivals, ih]

theorem recorded_live (t : Template) : ∀ e ∈ recorded t, ∀ o, e ≠ .arrived o := by
  intro e he o
  simp only [recorded, List.mem_flatMap] at he
  obtain ⟨a, _, ha⟩ := he
  match a with
  | .lit _ => simp at ha; subst ha; simp
  | .ref (.name _) => simp at ha
  | .ref (.var n sels) =>
    by_cases hs : sels.isEmpty <;> simp [hs] at ha
    subst ha; simp
  | .join ps =>
    simp only [List.mem_append, List.mem_filterMap, List.mem_singleton] at ha
    rcases ha with ⟨p, _, hp⟩ | rfl
    · cases p with
      | str => simp at hp
      | seen n s => by_cases hs : s.isEmpty <;> simp [hs] at hp; subst hp; simp
    · simp

/-- Running a template leaves the deciding frame's key as it was: what it
eliminates and introduces is never seen. -/
theorem recorded_key (rootOf : DefLoc → DefLoc) (f : Frame) (t : Template) :
    key rootOf (f.append (recorded t)) = key rootOf f := by
  simp [key, seen, Frame.append, arrivals_append, arrivals_nil (recorded_live t)]

/-- So too the key the machine decides at, through the program's classes. -/
theorem recorded_keyOf (E : Env A K) (f : Frame) (t : Template) :
    keyOf E (f.append (recorded t)) = keyOf E f := by
  simp only [keyOf, recorded_key]

/-- With no classes, the machine decides at the frame's key itself. -/
theorem keyOf_unclassed (E : Env A K) (h : E.program.classes = []) (f : Frame) :
    keyOf E f = key (fun _ => E.root) f := classKey_none h _

/-! ## What the machine runs is stored

The loop trusts a template at exactly two places: a decision's select,
and after a startle the template the model's own chain gives. Either is
stored at the frame's key: nothing runs that memory does not hold. -/

section
variable {A K : Type} [DecidableEq K]

/-- A decision selects only a template stored at the key. -/
theorem decided_stored {p : Refine.Policy Template K} {q : Nat} {sat : A → List Value → Bool} {m : Memory A}
    {k : Key} {v : List Value} {t : Template} (h : decideAt p q sat m k v = .select t) :
    ∃ r, m k = some r ∧ t ∈ r.branches.map (·.instr) := by
  unfold decideAt at h
  split at h
  · cases h
  · rename_i r hr
    split at h
    · cases h
    · refine ⟨r, hr, ?_⟩
      cases ht : Refine.tally p sat r.branches v with
      | null => rw [ht] at h; cases h
      | conflict => rw [ht] at h; cases h
      | select i =>
        rw [ht] at h
        cases h
        obtain ⟨_, _, x, hx, hi, _⟩ := Refine.tally_group ht
        exact List.mem_map.mpr ⟨x, (List.mem_filter.mp hx).1, hi⟩

theorem chainSelect_mem {sat : A → List Value → Bool} {v : List Value} :
    ∀ {chain : List (Refine.Cond A × Template)} {t : Template},
      Refine.chainSelect sat chain v = some t → t ∈ chain.map (·.2)
  | [], _, h => by simp [Refine.chainSelect] at h
  | (c, i) :: rest, t, h => by
    simp only [Refine.chainSelect] at h
    split at h
    · cases h; simp
    · exact List.mem_cons_of_mem _ (chainSelect_mem h)

theorem chainBranches_instrs : ∀ {voter next : Nat} {prev : List (Refine.Cond A)}
    {chain : List (Refine.Cond A × Template)},
    (Refine.chainBranches voter next prev chain).map (·.instr) = chain.map (·.2)
  | _, _, _, [] => rfl
  | _, _, _, (_, _) :: _ => by simp [Refine.chainBranches, chainBranches_instrs]

/-- After a startle, the template run is one the commit stored. -/
theorem written_stored {p : Refine.Policy Template K} {sat : A → List Value → Bool}
    {r : Refine.Refinement A Template} {chain : List (Refine.Cond A × Template)} {v : List Value}
    {t : Template} {l : Bool}
    (h : Refine.afterStartle p sat chain (Refine.commit r chain).branches v = .trust t l) :
    t ∈ (Refine.commit r chain).branches.map (·.instr) := by
  unfold Refine.afterStartle at h
  split at h
  · cases h
  · rename_i i hi
    cases h
    have := chainSelect_mem hi
    simp only [Refine.commit, List.map_append, List.mem_append]
    exact .inr (by rw [chainBranches_instrs]; exact this)

end

/-! ## The bounds hold

Every state the machine reaches keeps its bounds: no frame has run more
templates than the transition bound, and no transition has asked more
than the refinement bound. A frame past either ends, with an error
observation. -/

section
variable {A K : Type} [DecidableEq K]

def Bounded (E : Env A K) (s : State A) : Prop :=
  (∀ x, s.taken x ≤ E.program.constants.transitions) ∧ s.asks ≤ E.program.constants.refinements

theorem bind_some {α β : Type} {o : Option α} {g : α → Option β} {y : β} (h : o.bind g = some y) :
    ∃ a, g a = some y := by
  cases o with
  | none => simp at h
  | some a => exact ⟨a, h⟩

theorem go_bounded (E : Env A K) : ∀ (fuel : Nat) (c : Call) (s : State A) {s' : State A} {e : Effect},
    Bounded E s → go E fuel c s = some (s', e) → Bounded E s'
  | 0, _, _, _, _, _, h => by simp [go] at h
  | fuel + 1, c, s, s', e, hb, h => by
    have ih := go_bounded E fuel
    cases c with
    | decide id =>
      simp only [go] at h
      split at h
      · cases h
      · split at h <;> exact ih _ _ (by exact ⟨hb.1, Nat.zero_le _⟩) h
    | ask r =>
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · exact ih _ _ (by exact hb) h
        · rename_i hr
          cases h
          exact ⟨hb.1, by show s.asks + 1 ≤ _; omega⟩
    | trust id st =>
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · cases h
        · split at h
          · exact ih _ _ (by exact hb) h
          · split at h
            · exact ih _ _ (by exact hb) h
            · split at h
              · exact ih _ _ (by exact hb) h
              · rename_i hT
                have htk : ∀ x, (if x = id then s.taken x + 1 else s.taken x) ≤ E.program.constants.transitions := by
                  intro x
                  split
                  · subst_vars; omega
                  · exact hb.1 x
                repeat (first | (split at h <;> try (cases h; done)) | exact ih _ _ (by exact ⟨htk, hb.2⟩) h)
                all_goals first
                  | exact ih _ _ (by exact ⟨htk, hb.2⟩) h
                  | (cases h; exact ⟨htk, hb.2⟩)
                  | (obtain ⟨_, h⟩ := bind_some h; exact ih _ _ (by exact ⟨htk, hb.2⟩) h)
    | finish id v ex =>
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · cases h; exact hb
        · obtain ⟨_, h⟩ := bind_some h
          exact ih _ _ (by exact hb) h

theorem step_bounded (E : Env A K) (fuel : Nat) (s : State A) (i : Input A) {s' : State A} {e : Effect}
    (hb : Bounded E s) (h : step E fuel s i = some (s', e)) : Bounded E s' := by
  cases i with
  | answered v =>
    simp only [step] at h
    split at h
    · simp only [genesis] at h
      split at h
      · cases h
      · split at h
        · cases h
        · exact go_bounded E _ _ _ (by exact hb) h
    · exact go_bounded E _ _ _ (by exact hb) h
    all_goals cases h
  | wrote chain =>
    simp only [step] at h
    split at h
    · simp only [wrote] at h
      split at h
      · cases h
      · split at h
        · exact go_bounded E _ _ _ (by exact hb) h
        · split at h <;> exact go_bounded E _ _ _ (by exact hb) h
    · cases h

end

end Machine
