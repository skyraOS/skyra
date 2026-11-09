import Invariants.Shape

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

end Model
