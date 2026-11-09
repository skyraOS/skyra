import Invariants.Frames

/-! Proposal 60: definition and instantiation. The system has two sides.
A definition is what every execution shares; an instantiation is what one
execution holds. Six pairs, each a definition and its instantiation:

| definition     | instantiation |
|----------------|---------------|
| interface      | observation   |
| implementation | decision      |
| reference      | location      |
| scope          | frame         |
| program        | process       |
| call graph     | tree          |

The rules, by construction of the types below:

* **A definition never holds a value.** No definition type has a field of
  `Shape.Value`, so a key is a function of types alone (`key_values`).
* **Locations point from instantiation to definition, never back.** There
  are two spaces, `DefLoc` and `InstLoc`; no definition type has a field of
  `InstLoc`, so a key never reads the log (`key_locs`).
* **What has been seen and what is live are apart.** A key reads a frame's
  scope and its arrivals; eliminations and introductions are its live
  state and never enter it (`key_live`).

Everything here is pure. Lean computes no hash: hashes are given. -/

namespace Model

open Shape

/-! ## Two spaces -/

/-- A location in definition space: memory and programs, by hash. -/
structure DefLoc where
  hash : String
  deriving DecidableEq, Repr

/-- A location in instantiation space: the log, by hash. -/
structure InstLoc where
  hash : String
  deriving DecidableEq, Repr

/-- What a cut divides out: a field path, or runs of lines. A selector is
fixed in the template, never computed from a value. -/
inductive Sel where
  | field (path : List String)
  | lines (runs : List (Nat × Nat))
  deriving DecidableEq, Repr

/-! ## Definitions -/

/-- The operators. To invoke is to push, to return is to pop; each agent
kind's operator is named for the kind. -/
inductive Op where
  | delegate
  | act
  | boot
  | ret
  deriving DecidableEq, Repr

/-- An interface: an observation's type. A shape; a leaf, by its
definition in memory; a composition, a product of interfaces; or a cut,
an interface divided by a selector, kept undivided. -/
inductive Iface where
  | shape (s : Shape)
  | leaf (b : DefLoc)
  | op (o : Op) (parts : List Iface)
  | cut (i : Iface) (s : Sel)
  deriving Repr

/-- A frame's kind, a prefix of its scope's type: an agent decides, a leaf
only returns. -/
inductive Kind where
  | root
  | delegate
  | act
  | leaf
  deriving DecidableEq, Repr

/-- A reference: a relative name. A local by its `$n`, cut by selectors in
order; or a global, by name, relative to the program. -/
inductive Ref where
  | var (n : Nat) (sels : List Sel)
  | name (s : String)
  deriving DecidableEq, Repr

/-- What an introduction constructs from nothing: a literal, which is
code, or an exit status. -/
inductive Lit where
  | str (s : String)
  | exit (n : Nat)
  deriving DecidableEq, Repr

/-- An argument of a stored template: a reference, or an introduction,
inlined at commit. A template holds no free text. -/
inductive Arg where
  | ref (r : Ref)
  | lit (l : Lit)
  deriving DecidableEq, Repr

/-- A template: an operator and its arguments, made entirely of variables. -/
structure Template where
  op : Op
  args : List Arg
  deriving DecidableEq, Repr

/-- An implementation: one guard chain, authored at one `startle`, 59's
voter. Its conditions are abstract, as `Refine.Cond` is. -/
structure Impl (A : Type) where
  chain : List (A × Template)

/-- A scope: a frame's interface, what it is created with. Its kind, and
the interfaces of what it was passed, in roles' order: goal, tools,
context. -/
structure Scope where
  kind : Kind
  entries : List Iface

/-- A program's constants. `ι` is a ratio, `num / den`. -/
structure Constants where
  act : Nat
  transitions : Nat
  refinements : Nat
  inhibition : Nat × Nat
  floor : Nat
  deriving DecidableEq, Repr

/-- A program: the definition a process is instantiated from. The language
version, its constants, its global scope by name, and the entry scope its
grant must have. -/
structure Program where
  version : DefLoc
  constants : Constants
  global : List (String × DefLoc)
  entry : Scope

/-- Boot's first check: the program is well formed. -/
def Program.wellFormed (p : Program) : Bool :=
  1 ≤ p.constants.act && 1 ≤ p.constants.transitions && 1 ≤ p.constants.refinements &&
    0 < p.constants.inhibition.2 && p.constants.inhibition.1 ≤ p.constants.inhibition.2 &&
    1 ≤ p.constants.floor && p.constants.floor ≤ 4 &&
    decide (p.global.map (·.1)).Nodup

/-- A key: what has been seen. Its root is its program's `H(version,
global)`; then its kind; then the scope's interfaces and each arrival's,
in order. -/
structure Key where
  root : DefLoc
  kind : Kind
  seen : List Iface

/-- Memory, the call graph: each key to its refinement, `none` for a key
never seen, then the implementations authored there. -/
def Memory (A : Type) := Key → Option (List (Impl A))

/-! ## Instantiations -/

/-- An observation: a value in a frame, its interface, and where it is in
the log. -/
structure Observation where
  value : Value
  iface : Iface
  loc : InstLoc

/-- A location: an absolute address, in one space, with the selectors of a
cut. -/
inductive Location where
  | inst (l : InstLoc) (sels : List Sel)
  | defn (l : DefLoc) (sels : List Sel)

/-- What a frame binds from its scope: a tool, a definition; or an
observation. Not everything in a frame is an observation. -/
inductive Member where
  | tool (b : DefLoc)
  | obs (o : Observation)

/-- What is appended to a frame's state after its creation: what arrived,
which has been seen; or an elimination or an introduction, which are
live. -/
inductive Entry where
  | arrived (o : Observation)
  | eliminated (n : Nat) (sels : List Sel)
  | introduced (l : Lit)

/-- A frame: one activation of a scope. Its id; its static link, the
program it is defined under; its dynamic link, the frame that invoked it;
its scope; what it binds from it; and its state. -/
structure Frame where
  id : InstLoc
  program : DefLoc
  parent : Option InstLoc
  scope : Scope
  bindings : List Member
  state : List Entry

/-- What a decision gives. -/
inductive Outcome where
  | null
  | select (t : Template)
  | conflict

/-- A decision: the key it was taken at, its outcome, and the frame's
dynamic chain, recorded. -/
structure Decision where
  key : Key
  outcome : Outcome
  chain : List InstLoc

/-- A process: a program's instantiation, every frame whose static link is
it. -/
structure Process where
  program : DefLoc
  frames : List InstLoc

/-- A tree: every frame, linked by dynamic links. -/
structure Tree where
  frames : List Frame

/-! ## The key -/

/-- The interfaces of what arrived, in order: eliminations and
introductions are skipped. -/
def arrivals : List Entry → List Iface
  | [] => []
  | .arrived o :: es => o.iface :: arrivals es
  | .eliminated _ _ :: es => arrivals es
  | .introduced _ :: es => arrivals es

/-- What a frame has seen: its scope, then what arrived. -/
def seen (f : Frame) : List Iface := f.scope.entries ++ arrivals f.state

/-- A frame's key. `rootOf` gives a program's root, `H(version, global)`. -/
def key (rootOf : DefLoc → DefLoc) (f : Frame) : Key :=
  ⟨rootOf f.program, f.scope.kind, seen f⟩

theorem arrivals_append : ∀ (as bs : List Entry), arrivals (as ++ bs) = arrivals as ++ arrivals bs
  | [], _ => rfl
  | .arrived o :: as, bs => by simp [arrivals, arrivals_append as bs]
  | .eliminated _ _ :: as, bs => by simp [arrivals, arrivals_append as bs]
  | .introduced _ :: as, bs => by simp [arrivals, arrivals_append as bs]

/-! ## Definitions hold no value -/

def Observation.mapValue (g : Value → Value) (o : Observation) : Observation :=
  { o with value := g o.value }

def Member.mapValue (g : Value → Value) : Member → Member
  | .tool b => .tool b
  | .obs o => .obs (o.mapValue g)

def Entry.mapValue (g : Value → Value) : Entry → Entry
  | .arrived o => .arrived (o.mapValue g)
  | .eliminated n s => .eliminated n s
  | .introduced l => .introduced l

/-- Every value a frame holds, changed by `g`. -/
def Frame.mapValues (g : Value → Value) (f : Frame) : Frame :=
  { f with bindings := f.bindings.map (·.mapValue g), state := f.state.map (·.mapValue g) }

theorem arrivals_mapValue (g : Value → Value) :
    ∀ es : List Entry, arrivals (es.map (Entry.mapValue g)) = arrivals es
  | [] => rfl
  | .arrived o :: es => by
    simp [arrivals, Entry.mapValue, Observation.mapValue, arrivals_mapValue g es]
  | .eliminated _ _ :: es => by simp [arrivals, Entry.mapValue, arrivals_mapValue g es]
  | .introduced _ :: es => by simp [arrivals, Entry.mapValue, arrivals_mapValue g es]

/-- A key is a function of types alone: changing every value a frame holds
leaves its key unchanged. -/
theorem key_values (rootOf : DefLoc → DefLoc) (g : Value → Value) (f : Frame) :
    key rootOf (f.mapValues g) = key rootOf f := by
  simp [key, seen, Frame.mapValues, arrivals_mapValue]

/-- An arrival's interface: its shape, or none when it has none (a
repeated key). -/
def arrive (v : Value) (loc : InstLoc) : Option Observation :=
  (shapeOf v).map fun s => ⟨v, .shape s, loc⟩

/-- Two values of one shape arrive at one interface, wherever they are in
the log: the interface reads the shape, never the value. -/
theorem arrive_iface (h₁ : arrive v₁ l₁ = some o₁) (h₂ : arrive v₂ l₂ = some o₂)
    (hs : shapeOf v₁ = shapeOf v₂) : o₁.iface = o₂.iface := by
  simp only [arrive, Option.map_eq_some_iff] at h₁ h₂
  obtain ⟨s₁, e₁, rfl⟩ := h₁
  obtain ⟨s₂, e₂, rfl⟩ := h₂
  rw [hs, e₂] at e₁
  cases e₁
  rfl

/-! ## Locations point one way -/

def Observation.mapLoc (g : InstLoc → InstLoc) (o : Observation) : Observation :=
  { o with loc := g o.loc }

def Member.mapLoc (g : InstLoc → InstLoc) : Member → Member
  | .tool b => .tool b
  | .obs o => .obs (o.mapLoc g)

def Entry.mapLoc (g : InstLoc → InstLoc) : Entry → Entry
  | .arrived o => .arrived (o.mapLoc g)
  | .eliminated n s => .eliminated n s
  | .introduced l => .introduced l

/-- Every location in the log a frame holds, renamed by `g`: its id, its
dynamic link, and each observation's place. -/
def Frame.mapLocs (g : InstLoc → InstLoc) (f : Frame) : Frame :=
  { f with
    id := g f.id
    parent := f.parent.map g
    bindings := f.bindings.map (·.mapLoc g)
    state := f.state.map (·.mapLoc g) }

theorem arrivals_mapLoc (g : InstLoc → InstLoc) :
    ∀ es : List Entry, arrivals (es.map (Entry.mapLoc g)) = arrivals es
  | [] => rfl
  | .arrived o :: es => by
    simp [arrivals, Entry.mapLoc, Observation.mapLoc, arrivals_mapLoc g es]
  | .eliminated _ _ :: es => by simp [arrivals, Entry.mapLoc, arrivals_mapLoc g es]
  | .introduced _ :: es => by simp [arrivals, Entry.mapLoc, arrivals_mapLoc g es]

/-- A key never reads the log: renaming every instantiation location a
frame holds leaves its key unchanged, so memory depends on no run. -/
theorem key_locs (rootOf : DefLoc → DefLoc) (g : InstLoc → InstLoc) (f : Frame) :
    key rootOf (f.mapLocs g) = key rootOf f := by
  simp [key, seen, Frame.mapLocs, arrivals_mapLoc]

/-! ## Seen and live apart -/

/-- Live state never enters the key: appending an elimination or an
introduction to a frame's state leaves its key unchanged. -/
theorem key_live (rootOf : DefLoc → DefLoc) (f : Frame) (e : Entry)
    (he : ∀ o, e ≠ .arrived o) : key rootOf { f with state := f.state ++ [e] } = key rootOf f := by
  cases e with
  | arrived o => exact absurd rfl (he o)
  | eliminated n s => simp [key, seen, arrivals_append, arrivals]
  | introduced l => simp [key, seen, arrivals_append, arrivals]

/-- What arrives is seen: appending an arrival grows the key by exactly its
interface, at the end. -/
theorem key_arrived (rootOf : DefLoc → DefLoc) (f : Frame) (o : Observation) :
    (key rootOf { f with state := f.state ++ [.arrived o] }).seen = seen f ++ [o.iface] := by
  simp [key, seen, arrivals_append, arrivals]

/-! ## Kinds and direction

To invoke is to push, to return is to pop, and each operator's direction is
its form. A frame's kind decides which operators it emits: an agent's by
its scope's tool types under `k`, a leaf's by position, the frame an `act`
pushes. -/

/-- Direction is a function of the operator: every operator but `return`
pushes. -/
def Op.pushes : Op → Bool
  | .ret => false
  | _ => true

/-- A scope's tool types: the leaves among its interfaces, each once. -/
def toolTypes (entries : List Iface) : List DefLoc :=
  (entries.filterMap fun | .leaf b => some b | _ => none).eraseDups

/-- An agent's kind, counted under `k`: at most `k` tool types acts, more
delegates, and none is no agent. -/
def agentKind (k : Nat) (entries : List Iface) : Option Kind :=
  let t := (toolTypes entries).length
  if t = 0 then none else if t ≤ k then some .act else some .delegate

/-- The operators a kind emits. -/
def ops : Kind → List Op
  | .root => [.delegate, .boot, .ret]
  | .delegate => [.delegate, .ret]
  | .act => [.act, .ret]
  | .leaf => [.ret]

/-- The kind of the frame an operator pushes, given its scope's
interfaces: an act pushes a leaf, a delegation an agent counted under
`k`, a boot a root, and a return pushes nothing. -/
def pushed (k : Nat) : Op → List Iface → Option Kind
  | .act, _ => some .leaf
  | .delegate, entries => agentKind k entries
  | .boot, _ => some .root
  | .ret, _ => none

theorem pushes_iff (o : Op) : o.pushes = false ↔ o = .ret := by
  cases o <;> simp [Op.pushes]

/-- A push has a kind, a pop has none. -/
theorem pushed_ret (k : Nat) (entries : List Iface) : pushed k .ret entries = none := rfl

theorem agentKind_ne_leaf (k : Nat) (entries : List Iface) : agentKind k entries ≠ some .leaf := by
  simp only [agentKind]
  split
  · simp
  · split <;> simp

/-- A leaf emits only `return`. -/
theorem leaf_emits (o : Op) : o ∈ ops .leaf ↔ o = .ret := by simp [ops]

/-- Only an act frame emits `act`. -/
theorem act_emitted (kd : Kind) : Op.act ∈ ops kd ↔ kd = .act := by
  cases kd <;> simp [ops]

/-- A leaf is pushed exactly by `act`. -/
theorem pushed_leaf (k : Nat) (o : Op) (entries : List Iface) :
    pushed k o entries = some .leaf ↔ o = .act := by
  cases o with
  | act => simp [pushed]
  | delegate => simpa [pushed] using agentKind_ne_leaf k entries
  | boot => simp [pushed]
  | ret => simp [pushed]

/-- Leaves by position: a leaf's parent is an act frame. Whatever pushed a
leaf emitted `act`, and only an act frame emits it. -/
theorem leaf_parent (k : Nat) (parent : Kind) (o : Op) (entries : List Iface)
    (he : o ∈ ops parent) (hp : pushed k o entries = some .leaf) : parent = .act :=
  (act_emitted parent).mp ((pushed_leaf k o entries).mp hp ▸ he)

/-- A delegation pushes an act frame exactly when its scope holds between
one and `k` tool types. -/
theorem delegate_act (k : Nat) (entries : List Iface) :
    pushed k .delegate entries = some .act ↔
      1 ≤ (toolTypes entries).length ∧ (toolTypes entries).length ≤ k := by
  simp only [pushed, agentKind]
  split
  · rename_i h; simp [h]
  · rename_i h
    split
    · rename_i hk; simp [hk]; omega
    · rename_i hk; simp [hk]

/-! ## Frame identity

A frame is identified by its reference and its interface: the frame that
invoked it and which of its invokes it was, and its scope. The first frame
of a process, by its program and grant. A slot is a frame and a `$n`. Lean
computes no hash: hashes are given, and sound when they are injective and
apart. -/

/-- What boot hands the first frame: tools, definitions held by the
booting frame; and context values. -/
structure Grant where
  tools : List DefLoc
  context : List Value

structure Ids where
  child : InstLoc → Nat → Scope → InstLoc
  first : DefLoc → Grant → InstLoc
  slot : InstLoc → Nat → InstLoc
  program : Program → DefLoc

def Ids.Sound (h : Ids) : Prop :=
  (∀ {p₁ p₂ n₁ n₂ s₁ s₂}, h.child p₁ n₁ s₁ = h.child p₂ n₂ s₂ → p₁ = p₂ ∧ n₁ = n₂ ∧ s₁ = s₂) ∧
  (∀ {d₁ d₂ g₁ g₂}, h.first d₁ g₁ = h.first d₂ g₂ → d₁ = d₂ ∧ g₁ = g₂) ∧
  (∀ {p n s d g}, h.child p n s ≠ h.first d g)

/-- A frame's id from its path: the first frame's program and grant, then
each invoke, latest first. -/
def pathId (h : Ids) (d : DefLoc) (g : Grant) : List (Nat × Scope) → InstLoc
  | [] => h.first d g
  | (n, s) :: ps => h.child (pathId h d g ps) n s

/-- A frame's id is its path: under sound hashes, one id is one program,
one grant and one path from the first frame. -/
theorem pathId_inj {h : Ids} (hs : h.Sound) :
    ∀ {ps₁ ps₂ : List (Nat × Scope)} {d₁ d₂ g₁ g₂},
      pathId h d₁ g₁ ps₁ = pathId h d₂ g₂ ps₂ → d₁ = d₂ ∧ g₁ = g₂ ∧ ps₁ = ps₂
  | [], [], _, _, _, _, e => by
    obtain ⟨rfl, rfl⟩ := hs.2.1 e; exact ⟨rfl, rfl, rfl⟩
  | [], (_, _) :: _, _, _, _, _, e => absurd e.symm hs.2.2
  | (_, _) :: _, [], _, _, _, _, e => absurd e hs.2.2
  | (n₁, s₁) :: ps₁, (n₂, s₂) :: ps₂, _, _, _, _, e => by
    obtain ⟨ep, rfl, rfl⟩ := hs.1 e
    obtain ⟨rfl, rfl, rfl⟩ := pathId_inj hs ep
    exact ⟨rfl, rfl, rfl⟩

/-- Invoke: a frame pushes its `n`th child, of `scope`, bound to
`bindings`. Its static link is its parent's program, its dynamic link its
parent, and it starts with no state. -/
def push (h : Ids) (parent : Frame) (n : Nat) (scope : Scope) (bindings : List Member) : Frame :=
  ⟨h.child parent.id n scope, parent.program, some parent.id, scope, bindings, []⟩

theorem push_id_inj {h : Ids} (hs : h.Sound)
    (e : (push h f₁ n₁ s₁ b₁).id = (push h f₂ n₂ s₂ b₂).id) :
    f₁.id = f₂.id ∧ n₁ = n₂ ∧ s₁ = s₂ :=
  hs.1 e

/-- A child is defined under its parent's program: every frame of a
process reads one global scope and one set of constants. -/
theorem push_static (h : Ids) (f : Frame) (n : Nat) (s : Scope) (b : List Member) :
    (push h f n s b).program = f.program := rfl

theorem push_dynamic (h : Ids) (f : Frame) (n : Nat) (s : Scope) (b : List Member) :
    (push h f n s b).parent = some f.id := rfl

/-- What happens to a frame after its creation: entries appended. -/
def Frame.append (f : Frame) (es : List Entry) : Frame := { f with state := f.state ++ es }

/-- A frame's id does not change as its state grows: bindings are state,
not identity. -/
theorem append_id (f : Frame) (es : List Entry) : (f.append es).id = f.id := rfl

theorem append_scope (f : Frame) (es : List Entry) : (f.append es).scope = f.scope := rfl

/-- 58's stamp: which frame, and which state of it. -/
def stamp (f : Frame) : InstLoc × Nat := (f.id, f.state.length)

theorem stamp_append (f : Frame) (es : List Entry) :
    stamp (f.append es) = (f.id, f.state.length + es.length) := by
  simp [stamp, Frame.append]

/-- Every entry appended is a new state: no two states of one frame share a
stamp. -/
theorem stamp_apart (f : Frame) {es : List Entry} (h : es ≠ []) : stamp (f.append es) ≠ stamp f := by
  rw [stamp_append]
  intro e
  have := (Prod.mk.inj e).2
  exact h (List.eq_nil_of_length_eq_zero (by omega))

/-! ## Boot

`boot(program, grant)` pushes a process's first frame, or refuses. It runs
from a root, as its `n`th invoke, or at genesis from nothing. -/

/-- The definitions a frame holds as tools. -/
def held (f : Frame) : List DefLoc := f.bindings.filterMap fun | .tool b => some b | .obs _ => none

/-- The grant's context, arrived at slots from `n` on. -/
def contextObs (h : Ids) (id : InstLoc) : Nat → List Value → Option (List Observation)
  | _, [] => some []
  | n, v :: vs => match arrive v (h.slot id n), contextObs h id (n + 1) vs with
    | some o, some os => some (o :: os)
    | _, _ => none

/-- Boot's third check: the grant only narrows. At genesis nothing holds
tools, and what may be passed is open (60's question 8). -/
def narrows : Option (Frame × Nat) → Grant → Prop
  | none, _ => True
  | some (f, _), g => ∀ b ∈ g.tools, b ∈ held f

/-- The first frame's id: by program and grant at genesis, else as any
invoke. -/
def bootId (h : Ids) (p : Program) (g : Grant) : Option (Frame × Nat) → InstLoc
  | none => h.first (h.program p) g
  | some (f, n) => h.child f.id n p.entry

/-- The first frame's bindings: the grant's tools, then its context, in the
entry scope's order. -/
def grantBindings (g : Grant) (os : List Observation) : List Member :=
  g.tools.map .tool ++ os.map .obs

/-- Boot's checks: the program is well formed, the grant's interface is the
entry scope position by position, and the grant only narrows. -/
def Boots (p : Program) (g : Grant) (os : List Observation) (src : Option (Frame × Nat)) : Prop :=
  p.wellFormed = true ∧ p.entry.kind = .root ∧
    g.tools.map Iface.leaf ++ os.map (·.iface) = p.entry.entries ∧ narrows src g

open Classical in
/-- Boot: the checks, then exactly one frame pushed, its static link the
program passed, its dynamic link the frame that booted it. -/
noncomputable def boot (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) :
    Option Frame :=
  let id := bootId h p g src
  match contextObs h id (g.tools.length + 1) g.context with
  | none => none
  | some os =>
    if Boots p g os src then
      some ⟨id, h.program p, src.map (·.1.id), p.entry, grantBindings g os, []⟩
    else none

/-- Boot gives a frame exactly when its checks pass, and the frame is the
entry scope's, under the program passed, with nothing yet in its state. -/
theorem boot_eq_some (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) (f : Frame) :
    boot h p g src = some f ↔
      ∃ os, contextObs h (bootId h p g src) (g.tools.length + 1) g.context = some os ∧
        Boots p g os src ∧
        f = ⟨bootId h p g src, h.program p, src.map (·.1.id), p.entry, grantBindings g os, []⟩ := by
  unfold boot
  simp only
  split
  · rename_i hc; simp [hc]
  · rename_i os hc
    simp only [hc, Option.some.injEq, exists_eq_left']
    split
    · rename_i hb; simp [hb, eq_comm]
    · rename_i hb; simp [hb]

theorem held_grant (g : Grant) (os : List Observation) :
    (grantBindings g os).filterMap (fun | .tool b => some b | .obs _ => none) = g.tools := by
  simp [grantBindings, List.filterMap_append, List.filterMap_map]
  induction os <;> simp_all [Function.comp_def]

/-- Boot only narrows: the first frame holds exactly the grant's tools, and
each is held by the frame that booted it. -/
theorem boot_narrows (h : Ids) (p : Program) (g : Grant) (f r : Frame) (n : Nat)
    (hb : boot h p g (some (r, n)) = some f) : ∀ b ∈ held f, b ∈ held r := by
  obtain ⟨os, -, ⟨-, -, -, hn⟩, rfl⟩ := (boot_eq_some h p g _ f).mp hb
  intro b hb'
  have : b ∈ g.tools := by simpa [held, held_grant] using hb'
  exact hn b this

/-- The first frame is a root of the entry scope, under the program passed,
and its dynamic link is the frame that booted it. -/
theorem boot_frame (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) (f : Frame)
    (hb : boot h p g src = some f) :
    f.scope = p.entry ∧ f.scope.kind = .root ∧ f.program = h.program p ∧
      f.parent = src.map (·.1.id) ∧ f.state = [] := by
  obtain ⟨os, -, ⟨-, hr, -, -⟩, rfl⟩ := (boot_eq_some h p g src f).mp hb
  exact ⟨rfl, hr, rfl, rfl, rfl⟩

/-- A refused program never boots. -/
theorem boot_wellFormed (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat))
    (hw : p.wellFormed = false) : boot h p g src = none := by
  cases hb : boot h p g src with
  | none => rfl
  | some f =>
    obtain ⟨os, -, ⟨hw', -⟩, -⟩ := (boot_eq_some h p g src f).mp hb
    simp [hw] at hw'

/-- The first frame's key is its program's root, a root's kind, and the
entry scope: boot's order is canonical, so one program boots to one key. -/
theorem boot_key (h : Ids) (rootOf : DefLoc → DefLoc) (p : Program) (g : Grant)
    (src : Option (Frame × Nat)) (f : Frame) (hb : boot h p g src = some f) :
    key rootOf f = ⟨rootOf (h.program p), .root, p.entry.entries⟩ := by
  obtain ⟨hs, hr, hp, -, hst⟩ := boot_frame h p g src f hb
  simp [key, seen, hs, hp, hst, arrivals, ← hr]

/-! ## Cuts

A cut divides an interface by a selector. The key keeps `I / sel`
undivided; reducing it to a shape is a function, and agrees with the shape
of the value the cut takes. Cutting a value is 58's (`Frames.cutValue`). -/

def Sel.toCut : Sel → Frames.Sel
  | .field path => .field path
  | .lines runs => .lines runs

/-- A value cut by selectors in order. -/
def cutAll : Value → List Sel → Option Value
  | v, [] => some v
  | v, s :: ss => (Frames.cutValue s.toCut v).bind (cutAll · ss)

/-- A field path through a shape. -/
def fieldShape : Shape → List String → Option Shape
  | s, [] => some s
  | .obj fs, k :: ks => (fs.lookup k).bind (fieldShape · ks)
  | _, _ :: _ => none

/-- A shape divided by a selector: a field path's shape, or text's lines,
which are text. An empty selector divides nothing, as `cutValue` refuses
it. -/
def reduceShape : Sel → Shape → Option Shape
  | .field [], _ => none
  | .field path, s => fieldShape s path
  | .lines [], _ => none
  | .lines _, .str => some .str
  | .lines _, .bytes => some .str
  | .lines _, _ => none

/-- Reducing an interface: a cut of a shape is that shape divided, a cut of
a cut is divided in turn, and nothing else is a cut. -/
def Iface.reduce : Iface → Option Iface
  | .cut i sel => i.reduce.bind fun
    | .shape s => (reduceShape sel s).map .shape
    | _ => none
  | i => some i

theorem lookup_mem {α : Type} : ∀ {l : List (String × α)} {k : String} {v : α},
    l.lookup k = some v → (k, v) ∈ l
  | [], _, _, h => by simp at h
  | (k', v') :: l, k, v, h => by
    by_cases e : k = k'
    · subst e; simp at h; subst h; exact List.mem_cons_self ..
    · have : (k == k') = false := by simpa using e
      simp [List.lookup, this] at h
      exact List.mem_cons_of_mem _ (lookup_mem h)

theorem mem_lookup {α : Type} : ∀ {l : List (String × α)} {k : String} {v : α},
    (l.map (·.1)).Nodup → (k, v) ∈ l → l.lookup k = some v
  | [], _, _, _, h => by simp at h
  | (k', v') :: l, k, v, nd, h => by
    simp only [List.map_cons, List.nodup_cons] at nd
    rcases List.mem_cons.mp h with e | hl
    · simp only [Prod.mk.injEq] at e; obtain ⟨rfl, rfl⟩ := e; simp
    · have : k ≠ k' := by
        intro e; subst e; exact nd.1 (List.mem_map.mpr ⟨(k, v), hl, rfl⟩)
      have hb : (k == k') = false := by simpa using this
      simp [List.lookup, hb, mem_lookup nd.2 hl]

/-- A field's value has the field's shape: dividing a shape by a key
agrees with looking the key up in the value. -/
theorem lookup_shape {kvs : List (String × Value)} {s : Shape} {k : String} {v : Value}
    (h : shapeOf (.obj kvs) = some s) (hv : kvs.lookup k = some v) :
    ∃ t, shapeOf v = some t ∧ (match s with | .obj fs => fs.lookup k | _ => none) = some t := by
  obtain ⟨nd, fs, hf, rfl⟩ := shapeOf_obj h
  obtain ⟨t, ht, hm⟩ := fieldsOf_mem hf (lookup_mem hv)
  refine ⟨t, ht, ?_⟩
  have ndk : ((sortKeys fs).map (·.1)).Nodup :=
    (sortKeys_perm.map _).nodup_iff.mpr (by rw [fieldsOf_keys hf]; exact nd)
  exact mem_lookup ndk (sortKeys_perm.symm.subset hm)

theorem fieldAt_shape : ∀ (path : List String) {v v' : Value} {s : Shape},
    shapeOf v = some s → Frames.fieldAt v path = some v' →
      ∃ s', shapeOf v' = some s' ∧ fieldShape s path = some s'
  | [], _, _, _, hs, hf => by simp [Frames.fieldAt] at hf; subst hf; exact ⟨_, hs, rfl⟩
  | k :: ks, v, v', s, hs, hf => by
    cases v with
    | obj kvs =>
      simp only [Frames.fieldAt, Option.bind_eq_some_iff] at hf
      obtain ⟨w, hw, hf⟩ := hf
      obtain ⟨t, ht, hl⟩ := lookup_shape hs hw
      obtain ⟨s', h', hp⟩ := fieldAt_shape ks ht hf
      obtain ⟨-, fs, -, rfl⟩ := shapeOf_obj hs
      exact ⟨s', h', by simp only [fieldShape, Option.bind_eq_some_iff]; exact ⟨t, hl, hp⟩⟩
    | _ => simp [Frames.fieldAt] at hf

/-- A cut's value has its divided interface's shape. -/
theorem cut_sound {v v' : Value} {s : Shape} (sel : Sel) (hs : shapeOf v = some s)
    (hc : Frames.cutValue sel.toCut v = some v') :
    ∃ s', shapeOf v' = some s' ∧ reduceShape sel s = some s' := by
  cases sel with
  | field path =>
    cases path with
    | nil => simp [Sel.toCut, Frames.cutValue] at hc
    | cons k ks => exact fieldAt_shape (k :: ks) hs (by simpa [Sel.toCut, Frames.cutValue] using hc)
  | lines runs =>
    cases runs with
    | nil => simp [Sel.toCut, Frames.cutValue] at hc
    | cons r rs =>
      simp only [Sel.toCut, Frames.cutValue, Option.bind_eq_some_iff, Option.map_eq_some_iff] at hc
      obtain ⟨t, ht, ls, -, rfl⟩ := hc
      refine ⟨.str, rfl, ?_⟩
      cases v <;> simp [Frames.textOf] at ht <;> simp [shapeOf] at hs <;> subst hs <;> rfl

/-- Cut by selectors in order, a value keeps the interface reduced: the
key's `I / sel / …` reduces to the shape of what the cut takes. -/
theorem cutAll_sound : ∀ (sels : List Sel) {v v' : Value} {s : Shape} (i : Iface),
    shapeOf v = some s → i.reduce = some (.shape s) → cutAll v sels = some v' →
      ∃ s', shapeOf v' = some s' ∧ (sels.foldl .cut i).reduce = some (.shape s')
  | [], _, _, _, _, hs, hi, hc => by
    simp [cutAll] at hc; subst hc; exact ⟨_, hs, hi⟩
  | sel :: sels, v, v', s, i, hs, hi, hc => by
    simp only [cutAll, Option.bind_eq_some_iff] at hc
    obtain ⟨w, hw, hc⟩ := hc
    obtain ⟨t, ht, hr⟩ := cut_sound sel hs hw
    exact cutAll_sound sels (.cut i sel) ht (by simp [Iface.reduce, hi, hr]) hc

/-! ## Variables

A frame's `$n` are its bindings, then its state, in order, from `$1`. The
system instantiates the scope's and each arrival; the agent eliminates
(cuts something held) or introduces (constructs from nothing). -/

/-- What a `$n` holds: a tool, which has no value; a value seen, bound
from the scope or arrived; an introduction; or an elimination of an
earlier `$n`. -/
inductive Item where
  | tool (b : DefLoc)
  | seen (v : Value)
  | intro (l : Lit)
  | elim (n : Nat) (sels : List Sel)

def Member.item : Member → Item
  | .tool b => .tool b
  | .obs o => .seen o.value

def Entry.item : Entry → Item
  | .arrived o => .seen o.value
  | .eliminated n s => .elim n s
  | .introduced l => .intro l

def items (f : Frame) : List Item := f.bindings.map (·.item) ++ f.state.map (·.item)

/-- `$n`, from `$1`. -/
def item (f : Frame) (n : Nat) : Option Item := if n = 0 then none else (items f)[n - 1]?

def Lit.value : Lit → Value
  | .str s => .str s
  | .exit n => .exit n

/-- Variables are append-only: what a `$n` holds never changes as the
frame's state grows. -/
theorem item_append (f : Frame) (es : List Entry) {n : Nat} (h : n ≤ (items f).length) :
    item (f.append es) n = item f n := by
  unfold item
  split
  · rfl
  · have : items (f.append es) = items f ++ es.map (·.item) := by
      simp [items, Frame.append, List.append_assoc]
    rw [this, List.getElem?_append_left (by omega)]

/-- Appending an entry binds it to the next `$n`. -/
theorem item_next (f : Frame) (e : Entry) :
    item (f.append [e]) ((items f).length + 1) = some e.item := by
  simp [item, items, Frame.append]

/-- Reading `$n` cut by `sels`. An elimination reads the earlier `$n` it
eliminates; one of a later `$n` reads nothing. An introduction is never
cut: what is constructed from nothing is constructed whole. A tool has no
value. -/
def read (f : Frame) (n : Nat) (sels : List Sel) : Option Value :=
  match item f n with
  | some (.seen v) => cutAll v sels
  | some (.intro l) => if sels = [] then some l.value else none
  | some (.elim m s) => if m < n then read f m (s ++ sels) else none
  | _ => none
termination_by n

/-! ### Inlining at commit

At commit each elimination an implementation reads is replaced by its
definition and each introduction by its literal, so a stored
implementation reads only what was seen. -/

def inline (f : Frame) (n : Nat) (sels : List Sel) : Arg :=
  match item f n with
  | some (.elim m s) => if m < n then inline f m (s ++ sels) else .ref (.var n sels)
  | some (.intro l) => if sels = [] then .lit l else .ref (.var n sels)
  | _ => .ref (.var n sels)
termination_by n

/-- An argument's value in a frame. A global names a definition, which has
no value. -/
def evalArg (f : Frame) : Arg → Option Value
  | .lit l => some l.value
  | .ref (.var n sels) => read f n sels
  | .ref (.name _) => none

/-- Inlining at commit and evaluating gives the same value as reading. -/
theorem inline_read (f : Frame) (n : Nat) (sels : List Sel) :
    evalArg f (inline f n sels) = read f n sels := by
  have hr : read f n sels = match item f n with
      | some (.seen v) => cutAll v sels
      | some (.intro l) => if sels = [] then some l.value else none
      | some (.elim m s) => if m < n then read f m (s ++ sels) else none
      | _ => none := by rw [read]
  rw [inline, hr]
  cases e : item f n with
  | none => simp [evalArg, hr, e]
  | some it =>
    cases it with
    | tool b => simp [evalArg, hr, e]
    | seen v => simp [evalArg, hr, e]
    | intro l => by_cases hs : sels = [] <;> simp [evalArg, hr, e, hs]
    | elim m s =>
      by_cases hm : m < n
      · simp only [hm, ↓reduceIte]; exact inline_read f m (s ++ sels)
      · simp [evalArg, hr, e, hm]
termination_by n

/-- An inlined argument reads only what was seen: a literal, or a `$n`
holding a value seen, whenever the reference reads anything. -/
theorem inline_seen (f : Frame) (n : Nat) (sels : List Sel) (h : read f n sels ≠ none) :
    (∃ l, inline f n sels = .lit l) ∨
      ∃ j sels' v, inline f n sels = .ref (.var j sels') ∧ item f j = some (.seen v) := by
  rw [read] at h
  rw [inline]
  split
  · rename_i m s hm
    rw [hm] at h
    split
    · rename_i hlt; simp only [hlt, ↓reduceIte] at h; exact inline_seen f m (s ++ sels) h
    · rename_i hlt; simp [hlt] at h
  · rename_i l hl
    rw [hl] at h
    split
    · exact .inl ⟨l, rfl⟩
    · rename_i hs; simp [hs] at h
  · rename_i h₁ h₂
    cases e : item f n with
    | none => rw [e] at h; simp at h
    | some it =>
      cases it with
      | tool b => rw [e] at h; simp at h
      | seen v => exact .inr ⟨n, sels, v, rfl, e⟩
      | intro l => exact absurd e (h₂ l)
      | elim m s => exact absurd e (h₁ m s)
termination_by n

/-! ### Introduction and the copy check

An introduction constructs from nothing. A string introduced whose text
appears in a value the frame holds is refused, with the `$n` that holds
it: that is a copy, and the cut takes it instead. An exit status is never
a copy. -/

mutual
/-- Every string a value holds, at any depth, keys included. -/
def valueStrings : Value → List String
  | .str s => [s]
  | .arr xs => stringsList xs
  | .obj kvs => stringsMembers kvs
  | _ => []

def stringsList : List Value → List String
  | [] => []
  | x :: xs => valueStrings x ++ stringsList xs

def stringsMembers : List (String × Value) → List String
  | [] => []
  | (k, v) :: kvs => k :: valueStrings v ++ stringsMembers kvs
end

/-- Whether `$n` holds a value with `s` in its text. -/
def holds (f : Frame) (s : String) (n : Nat) : Bool :=
  match item f n with
  | some (.seen v) => (valueStrings v ++ (Frames.textOf v).toList).any fun t => decide (s.toList <:+: t.toList)
  | _ => false

/-- Introduce: the entry to append, or the `$n` it would copy. -/
def introduce (f : Frame) : Lit → Except Nat Entry
  | .exit k => .ok (.introduced (.exit k))
  | .str s => match (List.range' 1 (items f).length).find? (holds f s) with
    | some n => .error n
    | none => .ok (.introduced (.str s))

/-- A refused introduction names a `$n` holding a value with its text. -/
theorem introduce_refused {f : Frame} {s : String} {n : Nat} (h : introduce f (.str s) = .error n) :
    ∃ v, item f n = some (.seen v) ∧
      ∃ t ∈ valueStrings v ++ (Frames.textOf v).toList, s.toList <:+: t.toList := by
  simp only [introduce] at h
  split at h
  · rename_i n' hf
    cases h
    have hh := List.find?_some hf
    simp only [holds] at hh
    split at hh
    · rename_i v hv
      obtain ⟨t, ht, hi⟩ := List.any_eq_true.mp hh
      exact ⟨v, hv, t, ht, by simpa using hi⟩
    · simp at hh
  · cases h

/-- An introduction no held value contains is accepted. -/
theorem introduce_fresh {f : Frame} {s : String}
    (h : ∀ n, holds f s n = false) : introduce f (.str s) = .ok (.introduced (.str s)) := by
  simp only [introduce]
  rw [List.find?_eq_none.mpr (fun n _ => by simp [h n])]

/-- An exit status is never refused. -/
theorem introduce_exit (f : Frame) (k : Nat) : introduce f (.exit k) = .ok (.introduced (.exit k)) := rfl

/-- What introduce accepts is an introduction: live, so never in the key. -/
theorem introduce_live (rootOf : DefLoc → DefLoc) {f : Frame} {l : Lit} {e : Entry}
    (h : introduce f l = .ok e) : key rootOf (f.append [e]) = key rootOf f := by
  have : e = .introduced l := by
    cases l with
    | exit k => simp [introduce] at h; exact h.symm
    | str s =>
      simp only [introduce] at h
      split at h
      · cases h
      · cases h; rfl
  subst this
  exact key_live rootOf f _ (by simp)

end Model
