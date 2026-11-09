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

A frame whose own key has nothing may borrow: the program's classes are
read-only tools held interchangeable, each a hypothesis. The key with one
member swapped for another of an unrefuted class is tallied, and what it
selects runs here as it is, its tools by position. A borrowed act running
the swapped member is a trial: the member it was learned with runs on the
same arguments as a shadow, and the class is refuted when the two outputs
refine differently at the key it was learned at, by shape or by the
template the tally there selects. One selecting what the other does is no
evidence for the class, and a null none against. A frame that borrowed and
ends with a bound exhausted refutes its classes too. A refuted class is
never borrowed under again.

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
  /-- A trial's shadow: the member a borrowed act was learned with, run on
  the same arguments, its output compared and then dropped. -/
  | shadow (tool : DefLoc) (args : List Value)
  | ask (k : Key) (kind : Kind) (depth : Nat) (refused : Option Refusal)
  | finished (v : Value)
  | exhausted (v : Value)

inductive Phase where
  | start | ask | run | shadow | done
  deriving DecidableEq, Repr

/-- A borrowing: the class it is under, the key position swapped, the
frame's member there, and the member the template was learned with. -/
structure Borrow where
  cls : DefLoc
  pos : Nat
  own : DefLoc
  src : DefLoc
  deriving DecidableEq, Repr

/-- A trial in progress: a borrowed act, the frame that ran it, its
arguments, and its leaf's output once it came. -/
structure Trial where
  borrow : Borrow
  frame : InstLoc
  args : List Value
  real : Option Value

/-- A run: memory, which outlives it, and the log that folds to it, with
the classes refuted, which outlive it too; the tree; the frame deciding or
the leaf running; the asks of the transition in progress; each frame's
transitions; the classes each frame borrowed under; and a trial waiting
on its shadow. -/
structure State (A : Type) where
  memory : Memory A
  log : List (Event A)
  refuted : List DefLoc
  tree : Tree
  cur : InstLoc
  phase : Phase
  asks : Nat
  taken : InstLoc → Nat
  trials : InstLoc → List DefLoc
  pending : Option Trial

def State.init {A : Type} (log : List (Event A)) (refuted : List DefLoc) : State A :=
  ⟨replay log, log, refuted, Tree.mk [] [], ⟨""⟩, .start, 0, fun _ => 0, fun _ => [], none⟩

/-- What a run is of: its ids, its program, the tally's policy, and when
an atom holds of the values a frame has seen. -/
structure Env (A K : Type) where
  ids : Ids
  program : Program
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
  | trust (id : InstLoc) (st : Template) (borrowed : Option Borrow)
  | finish (id : InstLoc) (v : Value) (exhausted : Bool)
  | ask (r : Option Refusal)

section
variable {A K : Type} [DecidableEq K]

def reaches (E : Env A K) : DefLoc → String → List DefLoc := fun _ name => E.program.reach name

/-- Memory's root: the program's version and its operators' forms. Its
constants, entry scope, read-only tools, classes, and what its operators
reach stay out: a key already holds each tool its frame holds, and its
kind, so a program reaching one more tool shares every key whose frames
do not hold it. -/
def rootOf (E : Env A K) : DefLoc := E.ids.root E.program.version E.program.forms

def keyOf (E : Env A K) (f : Frame) : Key := key (fun _ => rootOf E) f

/-! ### Borrowing and trials -/

/-- A class's location: its members, sorted, as the shell hashes them. -/
def classId (E : Env A K) (c : List DefLoc) : DefLoc :=
  E.ids.cls (c.mergeSort fun a b => !decide (b.hash < a.hash))

/-- Each way a key's member may be swapped for another of the class: the
position, the member there, and the member put in its place. -/
def swaps (c : List DefLoc) (seen : List Iface) : List (Nat × DefLoc × DefLoc) :=
  seen.zipIdx.flatMap fun p => match p.1 with
    | .defn b => if b ∈ c then (c.filter (· ≠ b)).map fun b' => (p.2, b, b') else []
    | _ => []

def swapKey (k : Key) (n : Nat) (b : DefLoc) : Key := { k with seen := k.seen.set n (.defn b) }

/-- What one swap borrows: the template the swapped key selects. -/
def borrowAt (E : Env A K) (m : Memory A) (k : Key) (v : List Value) (id : DefLoc)
    (sw : Nat × DefLoc × DefLoc) : Option (Template × Borrow) :=
  match decideAt E.policy E.program.constants.quorum E.sat m (swapKey k sw.1 sw.2.2) v with
  | .select st => some (st, ⟨id, sw.1, sw.2.1, sw.2.2⟩)
  | _ => none

/-- The reuse rule over classes: in the program's order, under each class
not refuted, each swap in the key's order, the first whose key selects. -/
def borrow (E : Env A K) (refuted : List DefLoc) (m : Memory A) (k : Key) (v : List Value) :
    Option (Template × Borrow) :=
  E.program.classes.findSome? fun c =>
    if classId E c ∈ refuted then none else (swaps c k.seen).findSome? (borrowAt E m k v (classId E c))

/-- A borrowed act running the frame's own member is a trial. -/
def trialOf (b : Option Borrow) (frame : InstLoc) (tool : DefLoc) (args : List Value) : Option Trial :=
  b.bind fun b => if tool = b.own then some ⟨b, frame, args, none⟩ else none

def addTrial (b : Option Borrow) (id : InstLoc) (tr : InstLoc → List DefLoc) : InstLoc → List DefLoc :=
  match b with
  | some b => fun x => if x = id then b.cls :: tr x else tr x
  | none => tr

/-- Classes refuted, each once, in the order refuted. -/
def refute (r cs : List DefLoc) : List DefLoc := cs.foldl (fun r c => if c ∈ r then r else r ++ [c]) r

/-- An output as the key it was learned at reads it: the frame with it
arrived, its member swapped back, and the values as a guard reads them. -/
def asLearned (E : Env A K) (f : Frame) (b : Borrow) (o : Observation) : Key × List Value :=
  let g := f.append [.arrived o]
  (swapKey (keyOf E g) b.pos b.src, (vals g).set b.pos (.str b.src.hash))

/-- Whether a trial refutes its class: the real output and the shadow's
land at different keys where the template was learned, or the tally there
selects a different template for each. -/
def judged (E : Env A K) (m : Memory A) (f : Frame) (tr : Trial) (real shadow : Value) : Bool :=
  let loc := E.ids.slot tr.frame ((items f).length + 1)
  match arrive real loc, arrive shadow loc with
  | some o₁, some o₂ =>
    let (k₁, v₁) := asLearned E f tr.borrow o₁
    let (k₂, v₂) := asLearned E f tr.borrow o₂
    if k₁ ≠ k₂ then true else
    match decideAt E.policy E.program.constants.quorum E.sat m k₁ v₁,
        decideAt E.policy E.program.constants.quorum E.sat m k₂ v₂ with
    | .select t₁, .select t₂ => t₁ ≠ t₂
    | _, _ => false
  | _, _ => false

/-- A trial judged: its class refuted, or the state as it was. -/
def judge (E : Env A K) (s : State A) (tr : Trial) (real shadow : Value) : State A :=
  match s.tree.find tr.frame with
  | some f => if judged E s.memory f tr real shadow then { s with refuted := refute s.refuted [tr.borrow.cls] } else s
  | none => s

/-- A frame ending: with a bound exhausted, the classes it borrowed under
are refuted. -/
def exhaust (ex : Bool) (id : InstLoc) (s : State A) : State A :=
  if ex then { s with refuted := refute s.refuted (s.trials id) } else s

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
      | .select st => go E fuel (.trust id st none) s
      | .null => match borrow E s.refuted s.memory k (vals f) with
        | some r => go E fuel (.trust id r.1 (some r.2)) s
        | none => go E fuel (.ask none) s
      | .conflict => go E fuel (.ask none) s
  | fuel + 1, .ask r, s => match s.tree.find s.cur with
    | none => none
    | some f =>
      if E.program.constants.refinements ≤ s.asks then
        go E fuel (.finish f.id (errorValue s!"refinement bound exhausted: the model was asked {s.asks} times at this key without a guard that runs") true) s
      else some ({ s with asks := s.asks + 1, phase := .ask }, .ask (keyOf E f) f.scope.kind (depth s.tree f.id) r)
  | fuel + 1, .trust id st b, s => match s.tree.find id with
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
              tree := record s.tree id (recorded t)
              trials := addTrial b id s.trials }
            match s.tree.find id with
            | none => none
            | some f => match t.op with
              | .ret => match vs with
                | e :: out :: why => go E fuel (.finish id (returnValue e out why) false) s
                | _ => none
              | .act => match actEdge E s.tree id f t vs with
                | none => none
                | some r => some ({ s with
                    tree := r.2.1
                    cur := r.1
                    phase := .run
                    pending := trialOf b id r.2.2.1 r.2.2.2 }, .runLeaf r.2.2.1 r.2.2.2)
              | .delegate => (delegateEdge E s.tree id f t vs).bind fun r => go E fuel (.decide r.1) { s with tree := r.2 }
              | .boot => none
  | fuel + 1, .finish id v ex, s => match s.tree.find id with
    | none => none
    | some f => match f.parent with
      | none => some ({ exhaust ex id s with phase := .done }, if ex then .exhausted v else .finished v)
      | some pid => (popEdge E s.tree id pid v).bind fun tr => go E fuel (.decide pid) { exhaust ex id s with tree := tr }

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
      | .trust st _ => go E fuel (.trust f.id st none) s

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
    | .run => match s.pending with
      | some tr => some ({ s with phase := .shadow, pending := some { tr with real := some v } },
          .shadow tr.borrow.src tr.args)
      | none => go E fuel (.finish s.cur v false) s
    | .shadow => match s.pending with
      | some tr => match tr.real with
        | some rv => go E fuel (.finish s.cur rv false) (judge E { s with pending := none } tr rv v)
        | none => none
      | none => none
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

/-! ## Programs share memory

Two programs of one version and the same operators key every frame alike,
whatever tools or models they reach. A template learned under one may
name what the other does not reach; it is never run there: trust refuses
a template that does not fit, and the model is asked. -/

section
variable {A K : Type} [DecidableEq K]

omit [DecidableEq K] in
theorem keyOf_forms (E : Env A K) (p : Program) (hv : p.version = E.program.version)
    (hf : p.forms = E.program.forms) (f : Frame) : keyOf { E with program := p } f = keyOf E f := by
  simp [keyOf, rootOf, hv, hf]

theorem trust_unfit {E : Env A K} {fuel : Nat} {id : InstLoc} {st : Template} {b : Option Borrow}
    {s : State A} {f : Frame} {t : Template} (hf : s.tree.find id = some f) (hb : st.bind f = some t)
    (hu : fits E.program f t = false) : go E (fuel + 1) (.trust id st b) s = go E fuel (.ask (some .fits)) s := by
  simp [go, hf, hb, hu]

end

/-! ## What is borrowed

Reuse across tools happens only under a declared class not refuted at that
point in the run: the template is stored at the frame's key with one
member swapped for another of that class, and nowhere else. -/

section
variable {A K : Type} [DecidableEq K]

theorem swaps_mem {c : List DefLoc} {seen : List Iface} {sw : Nat × DefLoc × DefLoc}
    (h : sw ∈ swaps c seen) :
    sw.2.1 ∈ c ∧ sw.2.2 ∈ c ∧ sw.2.2 ≠ sw.2.1 ∧ seen[sw.1]? = some (.defn sw.2.1) := by
  simp only [swaps, List.mem_flatMap] at h
  obtain ⟨p, hp, h⟩ := h
  have hp' := List.mem_zipIdx_iff_getElem?.mp hp
  split at h
  · rename_i b hb
    split at h
    · rename_i hbc
      obtain ⟨b', hb', rfl⟩ := List.mem_map.mp h
      obtain ⟨hb'c, hne⟩ := List.mem_filter.mp hb'
      exact ⟨hbc, hb'c, by simpa using hne, by rw [hp', hb]⟩
    · exact absurd h List.not_mem_nil
  · exact absurd h List.not_mem_nil

theorem borrow_sound {E : Env A K} {r : List DefLoc} {m : Memory A} {k : Key} {v : List Value}
    {st : Template} {b : Borrow} (h : borrow E r m k v = some (st, b)) :
    ∃ c ∈ E.program.classes, b.cls = classId E c ∧ b.cls ∉ r ∧
      b.own ∈ c ∧ b.src ∈ c ∧ b.src ≠ b.own ∧ k.seen[b.pos]? = some (.defn b.own) ∧
      ∃ rf, m (swapKey k b.pos b.src) = some rf ∧ st ∈ rf.branches.map (·.instr) := by
  unfold borrow at h
  obtain ⟨c, hc, h⟩ := List.exists_of_findSome?_eq_some h
  split at h
  · cases h
  · rename_i hr
    obtain ⟨sw, hs, h⟩ := List.exists_of_findSome?_eq_some h
    unfold borrowAt at h
    split at h
    · rename_i st' hd
      simp only [Option.some.injEq, Prod.mk.injEq] at h
      obtain ⟨rfl, rfl⟩ := h
      obtain ⟨ho, hsrc, hne, hn⟩ := swaps_mem hs
      obtain ⟨rf, hrf, hmem⟩ := decided_stored hd
      exact ⟨c, hc, rfl, hr, ho, hsrc, hne, hn, rf, hrf, hmem⟩
    · cases h

/-- A decision trusts a borrowed template only when its own key selects
nothing: what a frame learned at its own key always comes first. -/
theorem borrow_only_null {E : Env A K} {fuel : Nat} {id : InstLoc} {s : State A} {f : Frame}
    (hf : s.tree.find id = some f) {t : Template}
    (hs : decideAt E.policy E.program.constants.quorum E.sat
      (if (s.memory (keyOf E f)).isNone then s.memory.apply (.hesitate (keyOf E f)) else s.memory)
      (keyOf E f) (vals f) = .select t) :
    go E (fuel + 1) (.decide id) s = go E fuel (.trust id t none)
      { s with
        cur := id
        asks := 0
        memory := if (s.memory (keyOf E f)).isNone then s.memory.apply (.hesitate (keyOf E f)) else s.memory
        log := if (s.memory (keyOf E f)).isNone then s.log ++ [.hesitate (keyOf E f)] else s.log } := by
  simp only [go, hf, hs]

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

theorem exhaust_fields (ex : Bool) (id : InstLoc) (s : State A) :
    (exhaust ex id s).taken = s.taken ∧ (exhaust ex id s).asks = s.asks := by
  unfold exhaust; split <;> exact ⟨rfl, rfl⟩

theorem judge_fields (E : Env A K) (s : State A) (tr : Trial) (a b : Value) :
    (judge E s tr a b).taken = s.taken ∧ (judge E s tr a b).asks = s.asks := by
  unfold judge; split
  · split <;> exact ⟨rfl, rfl⟩
  · exact ⟨rfl, rfl⟩

theorem go_bounded (E : Env A K) : ∀ (fuel : Nat) (c : Call) (s : State A) {s' : State A} {e : Effect},
    Bounded E s → go E fuel c s = some (s', e) → Bounded E s'
  | 0, _, _, _, _, _, h => by simp [go] at h
  | fuel + 1, c, s, s', e, hb, h => by
    have ih := go_bounded E fuel
    cases c with
    | decide id =>
      simp only [go] at h
      repeat' split at h
      all_goals first
        | cases h; done
        | exact ih _ _ (by exact ⟨hb.1, Nat.zero_le _⟩) h
    | ask r =>
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · exact ih _ _ (by exact hb) h
        · cases h
          exact ⟨hb.1, by show s.asks + 1 ≤ _; omega⟩
    | trust id st b =>
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
      have hx := exhaust_fields ex id s
      have hb' : Bounded E (exhaust ex id s) := ⟨hx.1 ▸ hb.1, hx.2 ▸ hb.2⟩
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · cases h; exact hb'
        · obtain ⟨_, h⟩ := bind_some h
          exact ih _ _ (by exact hb') h

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
    · split at h
      · cases h; exact hb
      · exact go_bounded E _ _ _ (by exact hb) h
    · split at h
      · split at h
        · refine go_bounded E _ _ _ ?_ h
          constructor
          · intro x; rw [(judge_fields E _ _ _ _).1]; exact hb.1 x
          · rw [(judge_fields E _ _ _ _).2]; exact hb.2
        · cases h
      · cases h
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

/-! ## Refutation is kept

A class once refuted stays refuted: no step forgets a refutation, so
nothing is borrowed under it again in this run, and the classes refuted
outlive it. -/

theorem refute_mono {x : DefLoc} : ∀ {r : List DefLoc} (cs : List DefLoc), x ∈ r → x ∈ refute r cs
  | _, [], h => h
  | r, c :: cs, h => by
    simp only [refute, List.foldl_cons]
    exact refute_mono cs (by split <;> simp_all)

theorem refute_mem {x : DefLoc} : ∀ {r : List DefLoc} (cs : List DefLoc), x ∈ cs → x ∈ refute r cs
  | _, [], h => absurd h List.not_mem_nil
  | r, c :: cs, h => by
    simp only [refute, List.foldl_cons]
    rcases List.mem_cons.mp h with rfl | h
    · exact refute_mono cs (by split <;> simp_all)
    · exact refute_mem cs h

theorem exhaust_refuted {x : DefLoc} (ex : Bool) (id : InstLoc) (s : State A) (h : x ∈ s.refuted) :
    x ∈ (exhaust ex id s).refuted := by
  unfold exhaust; split
  · exact refute_mono _ h
  · exact h

theorem judge_refuted {x : DefLoc} (E : Env A K) (s : State A) (tr : Trial) (a b : Value) (h : x ∈ s.refuted) :
    x ∈ (judge E s tr a b).refuted := by
  unfold judge; split
  · split
    · exact refute_mono _ h
    · exact h
  · exact h

theorem go_refuted (E : Env A K) {x : DefLoc} : ∀ (fuel : Nat) (c : Call) (s : State A) {s' : State A} {e : Effect},
    x ∈ s.refuted → go E fuel c s = some (s', e) → x ∈ s'.refuted
  | 0, _, _, _, _, _, h => by simp [go] at h
  | fuel + 1, c, s, s', e, hx, h => by
    have ih := go_refuted E (x := x) fuel
    cases c with
    | finish id v ex =>
      have hx' := exhaust_refuted ex id s hx
      simp only [go] at h
      split at h
      · cases h
      · split at h
        · cases h; exact hx'
        · obtain ⟨_, h⟩ := bind_some h
          exact ih _ _ (by exact hx') h
    | _ =>
      simp only [go] at h
      repeat' split at h
      all_goals first
        | cases h; done
        | (cases h; exact hx)
        | exact ih _ _ (by exact hx) h
        | (obtain ⟨_, h⟩ := bind_some h; exact ih _ _ (by exact hx) h)

theorem step_refuted (E : Env A K) {x : DefLoc} (fuel : Nat) (s : State A) (i : Input A) {s' : State A} {e : Effect}
    (hx : x ∈ s.refuted) (h : step E fuel s i = some (s', e)) : x ∈ s'.refuted := by
  cases i with
  | answered v =>
    simp only [step] at h
    split at h
    · simp only [genesis] at h
      repeat' split at h
      all_goals first
        | cases h; done
        | exact go_refuted E _ _ _ (by exact hx) h
    · split at h
      · cases h; exact hx
      · exact go_refuted E _ _ _ (by exact hx) h
    · split at h
      · split at h
        · exact go_refuted E _ _ _ (judge_refuted E { s with pending := none } _ _ _ hx) h
        · cases h
      · cases h
    all_goals cases h
  | wrote chain =>
    simp only [step] at h
    split at h
    · simp only [wrote] at h
      repeat' split at h
      all_goals first
        | cases h; done
        | exact go_refuted E _ _ _ (by exact hx) h
    · cases h

/-- A trial whose shadow gave what its leaf gave refutes nothing: only a
difference is evidence. -/
theorem judged_same (E : Env A K) (m : Memory A) (f : Frame) (tr : Trial) (v : Value) :
    judged E m f tr v v = false := by
  unfold judged
  dsimp only
  cases Model.arrive v (E.ids.slot tr.frame ((items f).length + 1)) with
  | none => rfl
  | some o =>
    dsimp only
    generalize asLearned E f tr.borrow o = q
    obtain ⟨k, w⟩ := q
    dsimp only
    simp only [ne_eq, not_true_eq_false, ↓reduceIte]
    cases decideAt E.policy E.program.constants.quorum E.sat m k w <;> simp

end

end Machine
