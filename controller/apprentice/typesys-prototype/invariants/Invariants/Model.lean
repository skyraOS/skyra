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

/-- An interface: an observation's type. A shape; a definition in memory,
`def B`, a tool or a model by its program's endpoints; a composition, a product of interfaces; or a cut,
an interface divided by a selector, kept undivided. -/
inductive Iface where
  | shape (s : Shape)
  | defn (b : DefLoc)
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
voter. Its conditions are 59's, over abstract atoms. -/
structure Impl (A : Type) where
  chain : List (Refine.Cond A × Template)

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

/-- What each operator reaches: the models `delegate` reaches and the tools
`act` reaches. -/
structure Endpoints where
  delegate : List DefLoc
  act : List DefLoc
  deriving DecidableEq, Repr

/-- A program: the definition a process is instantiated from. The language
version, its constants, its global scope by name, its endpoints, and the
entry scope its grant must have. -/
structure Program where
  version : DefLoc
  constants : Constants
  global : List (String × DefLoc)
  endpoints : Endpoints
  entry : Scope

/-- Boot's first check: the program is well formed. -/
def Program.wellFormed (p : Program) : Bool :=
  1 ≤ p.constants.act && 1 ≤ p.constants.transitions && 1 ≤ p.constants.refinements &&
    0 < p.constants.inhibition.2 && p.constants.inhibition.1 ≤ p.constants.inhibition.2 &&
    1 ≤ p.constants.floor && p.constants.floor ≤ 4 &&
    decide (p.global.map (·.1)).Nodup &&
    p.endpoints.act.all (· ∉ p.endpoints.delegate)

/-- A key: what has been seen. Its root is its program's `H(version,
global)`; then its kind; then the scope's interfaces and each arrival's,
in order. -/
structure Key where
  root : DefLoc
  kind : Kind
  seen : List Iface

/-- Memory, the call graph: each key to its refinement, 59's, holding the
implementations authored there as voters; `none` for a key never seen. -/
abbrev Memory (A : Type) := Key → Option (Refine.Refinement A Template)

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

/-- A tree: every frame, linked by dynamic links, and the frames that have
returned. -/
structure Tree where
  frames : List Frame
  closed : List InstLoc

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
pushes. Tool types are counted through the program's endpoints: `act` is
the program's tools. -/

/-- Direction is a function of the operator: every operator but `return`
pushes. -/
def Op.pushes : Op → Bool
  | .ret => false
  | _ => true

/-- A scope's definitions: tools and models alike. -/
def defs (entries : List Iface) : List DefLoc :=
  entries.filterMap fun | .defn b => some b | _ => none

/-- A scope's tool types: its definitions among the program's tools, each
once. Models never count. -/
def toolTypes (act : List DefLoc) (entries : List Iface) : List DefLoc :=
  ((defs entries).filter (· ∈ act)).eraseDups

/-- An agent's kind, counted under `k`: at most `k` tool types acts, more
delegates, and none is no agent. -/
def agentKind (k : Nat) (act : List DefLoc) (entries : List Iface) : Option Kind :=
  let t := (toolTypes act entries).length
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
def pushed (k : Nat) (act : List DefLoc) : Op → List Iface → Option Kind
  | .act, _ => some .leaf
  | .delegate, entries => agentKind k act entries
  | .boot, _ => some .root
  | .ret, _ => none

theorem pushes_iff (o : Op) : o.pushes = false ↔ o = .ret := by
  cases o <;> simp [Op.pushes]

/-- A push has a kind, a pop has none. -/
theorem pushed_ret (k : Nat) (act : List DefLoc) (entries : List Iface) :
    pushed k act .ret entries = none := rfl

theorem agentKind_ne_leaf (k : Nat) (act : List DefLoc) (entries : List Iface) :
    agentKind k act entries ≠ some .leaf := by
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
theorem pushed_leaf (k : Nat) (act : List DefLoc) (o : Op) (entries : List Iface) :
    pushed k act o entries = some .leaf ↔ o = .act := by
  cases o with
  | act => simp [pushed]
  | delegate => simpa [pushed] using agentKind_ne_leaf k act entries
  | boot => simp [pushed]
  | ret => simp [pushed]

/-- Leaves by position: a leaf's parent is an act frame. Whatever pushed a
leaf emitted `act`, and only an act frame emits it. -/
theorem leaf_parent (k : Nat) (act : List DefLoc) (parent : Kind) (o : Op) (entries : List Iface)
    (he : o ∈ ops parent) (hp : pushed k act o entries = some .leaf) : parent = .act :=
  (act_emitted parent).mp ((pushed_leaf k act o entries).mp hp ▸ he)

/-- A delegation pushes an act frame exactly when its scope holds between
one and `k` tool types. -/
theorem delegate_act (k : Nat) (act : List DefLoc) (entries : List Iface) :
    pushed k act .delegate entries = some .act ↔
      1 ≤ (toolTypes act entries).length ∧ (toolTypes act entries).length ≤ k := by
  simp only [pushed, agentKind]
  split
  · rename_i h; simp [h]
  · rename_i h
    split
    · rename_i hk; simp [hk]; omega
    · rename_i hk; simp [hk]

theorem mem_toolTypes {act : List DefLoc} {entries : List Iface} {b : DefLoc} :
    b ∈ toolTypes act entries ↔ b ∈ defs entries ∧ b ∈ act := by
  simp [toolTypes, List.mem_eraseDups]

/-- Models never count against `k`: in a well-formed program no tool type
is a model. -/
theorem toolTypes_not_model {p : Program} (hw : p.wellFormed = true) {entries : List Iface} {b : DefLoc}
    (hb : b ∈ toolTypes p.endpoints.act entries) : b ∉ p.endpoints.delegate := by
  simp only [Program.wellFormed, Bool.and_eq_true, List.all_eq_true, decide_eq_true_eq] at hw
  exact hw.2 b (mem_toolTypes.mp hb).2

/-- A scope passing a subset of another's definitions passes a subset of
its tools. -/
theorem toolTypes_sub {act : List DefLoc} {e₁ e₂ : List Iface}
    (h : ∀ b ∈ defs e₁, b ∈ defs e₂) : ∀ b ∈ toolTypes act e₁, b ∈ toolTypes act e₂ := by
  intro b hb
  obtain ⟨hd, ha⟩ := mem_toolTypes.mp hb
  exact mem_toolTypes.mpr ⟨h b hd, ha⟩

/-! ## Frame identity

A frame is identified by its reference and its interface: the frame that
invoked it and which of its invokes it was, and its scope. The first frame
of a process, by its program and grant. A slot is a frame and a `$n`. Lean
computes no hash: hashes are given, and sound when they are injective and
apart. -/

/-- What boot hands the first frame: definitions, tools and models held by
the booting frame; and context values. -/
structure Grant where
  defs : List DefLoc
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
  | some (f, _), g => ∀ b ∈ g.defs, b ∈ held f

/-- The first frame's id: by program and grant at genesis, else as any
invoke. -/
def bootId (h : Ids) (p : Program) (g : Grant) : Option (Frame × Nat) → InstLoc
  | none => h.first (h.program p) g
  | some (f, n) => h.child f.id n p.entry

/-- The first frame's bindings: the grant's tools, then its context, in the
entry scope's order. -/
def grantBindings (g : Grant) (os : List Observation) : List Member :=
  g.defs.map .tool ++ os.map .obs

/-- Boot's checks: the program is well formed, the grant's interface is the
entry scope position by position, every definition it passes is an
endpoint, and the grant only narrows. -/
def Boots (p : Program) (g : Grant) (os : List Observation) (src : Option (Frame × Nat)) : Prop :=
  p.wellFormed = true ∧ p.entry.kind = .root ∧
    g.defs.map Iface.defn ++ os.map (·.iface) = p.entry.entries ∧
    (∀ b ∈ g.defs, b ∈ p.endpoints.delegate ∨ b ∈ p.endpoints.act) ∧ narrows src g

open Classical in
/-- Boot: the checks, then exactly one frame pushed, its static link the
program passed, its dynamic link the frame that booted it. -/
noncomputable def boot (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) :
    Option Frame :=
  let id := bootId h p g src
  match contextObs h id (g.defs.length + 1) g.context with
  | none => none
  | some os =>
    if Boots p g os src then
      some ⟨id, h.program p, src.map (·.1.id), p.entry, grantBindings g os, []⟩
    else none

/-- Boot gives a frame exactly when its checks pass, and the frame is the
entry scope's, under the program passed, with nothing yet in its state. -/
theorem boot_eq_some (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) (f : Frame) :
    boot h p g src = some f ↔
      ∃ os, contextObs h (bootId h p g src) (g.defs.length + 1) g.context = some os ∧
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
    (grantBindings g os).filterMap (fun | .tool b => some b | .obs _ => none) = g.defs := by
  simp [grantBindings, List.filterMap_append, List.filterMap_map]
  induction os <;> simp_all [Function.comp_def]

/-- Boot only narrows: the first frame holds exactly the grant's tools, and
each is held by the frame that booted it. -/
theorem boot_narrows (h : Ids) (p : Program) (g : Grant) (f r : Frame) (n : Nat)
    (hb : boot h p g (some (r, n)) = some f) : ∀ b ∈ held f, b ∈ held r := by
  obtain ⟨os, -, ⟨-, -, -, -, hn⟩, rfl⟩ := (boot_eq_some h p g _ f).mp hb
  intro b hb'
  have : b ∈ g.defs := by simpa [held, held_grant] using hb'
  exact hn b this

/-- The first frame is a root of the entry scope, under the program passed,
and its dynamic link is the frame that booted it. -/
theorem boot_frame (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) (f : Frame)
    (hb : boot h p g src = some f) :
    f.scope = p.entry ∧ f.scope.kind = .root ∧ f.program = h.program p ∧
      f.parent = src.map (·.1.id) ∧ f.state = [] := by
  obtain ⟨os, -, ⟨-, hr, -, -, -⟩, rfl⟩ := (boot_eq_some h p g src f).mp hb
  exact ⟨rfl, hr, rfl, rfl, rfl⟩

/-- Endpoints are typed at boot: every definition the first frame holds is
a model or a tool of the program passed. -/
theorem boot_endpoints (h : Ids) (p : Program) (g : Grant) (src : Option (Frame × Nat)) (f : Frame)
    (hb : boot h p g src = some f) : ∀ b ∈ held f, b ∈ p.endpoints.delegate ∨ b ∈ p.endpoints.act := by
  obtain ⟨os, -, ⟨-, -, -, he, -⟩, rfl⟩ := (boot_eq_some h p g src f).mp hb
  intro b hb'
  exact he b (by simpa [held, held_grant] using hb')

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

/-! ## The key is lossless

58 keyed a call on its one hop: the interfaces at the positions the call
reads. 60 keys on the whole scope. The one hop is a function of the whole
key and the call; the whole key is not a function of the one hop. -/

/-- The interfaces a template reads, cut as it cuts them: 58's one hop. -/
def project (k : Key) (t : Template) : List (Option Iface) :=
  t.args.filterMap fun
    | .ref (.var n sels) => some ((k.seen[n - 1]?).map (sels.foldl .cut))
    | _ => none

/-- The one hop sees only the positions the call reads: keys agreeing
there project alike, whatever else they hold. -/
theorem project_reads (t : Template) {k₁ k₂ : Key}
    (h : ∀ n sels, Arg.ref (.var n sels) ∈ t.args → k₁.seen[n - 1]? = k₂.seen[n - 1]?) :
    project k₁ t = project k₂ t := by
  unfold project
  generalize t.args = args at h
  induction args with
  | nil => rfl
  | cons a as ih =>
    have ih := ih fun n sels hm => h n sels (List.mem_cons_of_mem _ hm)
    match a with
    | .ref (.var n sels) => simp [h n sels List.mem_cons_self, ih]
    | .ref (.name _) => simpa using ih
    | .lit _ => simpa using ih

/-- The converse fails: two keys, one hop, two scopes. What 58 merged, 60
keeps apart. -/
theorem project_lossy (root : DefLoc) :
    ∃ k₁ k₂ : Key, ∃ t : Template, project k₁ t = project k₂ t ∧ k₁ ≠ k₂ := by
  refine ⟨⟨root, .act, [.shape .str, .shape .num]⟩, ⟨root, .act, [.shape .str, .shape .bool]⟩,
    ⟨.ret, [.ref (.var 1 [])]⟩, rfl, ?_⟩
  intro e
  simp only [Key.mk.injEq, List.cons.injEq, Iface.shape.injEq, true_and] at e
  exact absurd e.1 (by intro h; cases h)

/-! ## Reuse: exact match aligns by identity

A reuse rule says which stored implementations count at a key, and how
their references align with the key's positions. The base rule is exact
match: only the key's own, aligned by identity. -/

/-- Rebinding a template's references through an alignment. -/
def Template.rebind (a : Nat → Nat) (t : Template) : Template :=
  { t with args := t.args.map fun
    | .ref (.var n sels) => .ref (.var (a n) sels)
    | x => x }

theorem rebind_id (t : Template) : t.rebind id = t := by
  cases t with
  | mk op args =>
    simp only [Template.rebind, Template.mk.injEq, true_and]
    induction args with
    | nil => rfl
    | cons x xs ih =>
      match x with
      | .ref (.var _ _) => simpa using ih
      | .ref (.name _) => simpa using ih
      | .lit _ => simpa using ih

/-- A reuse rule: from a stored key to the key being decided, an alignment
of positions, or none when the stored key does not count. -/
def Rule := Key → Key → Option (Nat → Nat)

open Classical in
/-- The base rule. -/
noncomputable def exact : Rule := fun s t => if s = t then some id else none

theorem exact_self (k : Key) : exact k k = some id := by simp [exact]

/-- Exact reuse counts only a key's own, and binds each reference to the
same `$n`. -/
theorem exact_only {s t : Key} {a : Nat → Nat} (h : exact s t = some a) :
    s = t ∧ ∀ tm : Template, tm.rebind a = tm := by
  unfold exact at h
  split at h
  · cases h; exact ⟨by assumption, rebind_id⟩
  · cases h

/-! ## Memory is a fold of the log

The log's events are what has been seen: a `hesitate` records a key never
seen, its refinement `null`; a `startle` commits an implementation at a
key, 59's commit. Memory is the fold, from nothing. -/

inductive Event (A : Type) where
  | hesitate (k : Key)
  | startle (k : Key) (impl : Impl A)

def Event.key {A : Type} : Event A → Key
  | .hesitate k => k
  | .startle k _ => k

section
open Classical
variable {A : Type}

/-- The refinement at one key after one event there. -/
def stepAt : Option (Refine.Refinement A Template) → Event A → Option (Refine.Refinement A Template)
  | none, .hesitate _ => some .empty
  | some r, .hesitate _ => some r
  | r, .startle _ i => some (Refine.commit (r.getD .empty) i.chain)

noncomputable def Memory.apply (m : Memory A) (e : Event A) : Memory A :=
  fun k => if k = e.key then stepAt (m k) e else m k

/-- Memory from the log. -/
noncomputable def replay (log : List (Event A)) : Memory A := log.foldl Memory.apply fun _ => none

/-- One key's refinement, from its events alone. -/
def replayAt (log : List (Event A)) : Option (Refine.Refinement A Template) := log.foldl stepAt none

theorem foldl_apply_at (k : Key) : ∀ (log : List (Event A)) (m : Memory A),
    (log.foldl Memory.apply m) k = (log.filter (·.key = k)).foldl stepAt (m k)
  | [], _ => rfl
  | e :: log, m => by
    rw [List.foldl_cons, foldl_apply_at k log]
    by_cases h : e.key = k
    · simp [h, Memory.apply]
    · have hk : k ≠ e.key := fun e' => h e'.symm
      simp [h, Memory.apply, hk]

/-- What has been seen at a key decides its refinement, and nothing else
does: memory at a key is the fold of that key's events alone. -/
theorem replay_at (log : List (Event A)) (k : Key) :
    replay log k = replayAt (log.filter (·.key = k)) := by
  exact foldl_apply_at k log _

/-- The log appends, memory folds on: replaying more is applying it. -/
theorem replay_append (log more : List (Event A)) :
    replay (log ++ more) = more.foldl Memory.apply (replay log) := by
  exact List.foldl_append

/-- A key with no event is unseen. -/
theorem replay_unseen (log : List (Event A)) (k : Key) (h : ∀ e ∈ log, e.key ≠ k) :
    replay log k = none := by
  rw [replay_at]
  have : log.filter (·.key = k) = [] := List.filter_eq_nil_iff.mpr (by simpa using h)
  rw [this]; rfl

theorem stepAt_some : ∀ (r : Option (Refine.Refinement A Template)) (e : Event A), (stepAt r e).isSome
  | none, .hesitate _ => rfl
  | some _, .hesitate _ => rfl
  | none, .startle _ _ => rfl
  | some _, .startle _ _ => rfl

theorem foldl_stepAt_some : ∀ (log : List (Event A)) (r : Refine.Refinement A Template),
    (log.foldl stepAt (some r)).isSome
  | [], _ => rfl
  | e :: log, r => by
    rw [List.foldl_cons]
    obtain ⟨r', h⟩ := Option.isSome_iff_exists.mp (stepAt_some (some r) e)
    rw [h]; exact foldl_stepAt_some log r'

/-- A key with an event is seen, its refinement recorded, `null` until a
`startle` commits there. -/
theorem replay_seen (log : List (Event A)) (k : Key) (e : Event A) (he : e ∈ log) (hk : e.key = k) :
    (replay log k).isSome := by
  rw [replay_at]
  obtain ⟨l₁, l₂, hl⟩ := List.mem_iff_append.mp he
  have hf : log.filter (·.key = k) = l₁.filter (·.key = k) ++ e :: l₂.filter (·.key = k) := by
    subst hl; simp [List.filter_append, hk]
  rw [hf, replayAt, List.foldl_append, List.foldl_cons]
  obtain ⟨r', h⟩ := Option.isSome_iff_exists.mp (stepAt_some (l₁.filter (·.key = k) |>.foldl stepAt none) e)
  rw [h]; exact foldl_stepAt_some _ r'

/-- `hesitate` never erases: a refinement already there is kept. -/
theorem hesitate_keeps (r : Refine.Refinement A Template) (k : Key) :
    stepAt (some r) (.hesitate k) = some r := rfl

/-- `startle` never erases: every branch already there is in the next
version. -/
theorem startle_keeps {r : Refine.Refinement A Template} {b} (k : Key) (i : Impl A)
    (hb : b ∈ r.branches) : ∃ r', stepAt (some r) (.startle k i) = some r' ∧ b ∈ r'.branches :=
  ⟨_, rfl, Refine.commit_keeps hb⟩

/-! ## A decision is a tally

Under the exact rule, the decision at a key is 59's tally over the
implementations stored there. An unseen key, and a seen one never
committed, are `null`: the path to `startle`. -/

variable {V K : Type} [DecidableEq K]

def Outcome.ofRefine : Refine.Outcome Template → Outcome
  | .null => .null
  | .select t => .select t
  | .conflict => .conflict

/-- The decision at a frame's key, given the values it reads. -/
def decideAt (p : Refine.Policy Template K) (sat : A → V → Bool) (m : Memory A) (k : Key) (v : V) :
    Outcome :=
  match m k with
  | none => .null
  | some r => .ofRefine (Refine.tally p sat r.branches v)

theorem decideAt_tally (p : Refine.Policy Template K) (sat : A → V → Bool) {m : Memory A} {k : Key}
    {r : Refine.Refinement A Template} (h : m k = some r) (v : V) :
    decideAt p sat m k v = .ofRefine (Refine.tally p sat r.branches v) := by
  simp [decideAt, h]

theorem decideAt_unseen (p : Refine.Policy Template K) (sat : A → V → Bool) {m : Memory A} {k : Key}
    (h : m k = none) (v : V) : decideAt p sat m k v = .null := by
  simp [decideAt, h]

/-- A key seen and never committed is `null`: `hesitate` records `null`. -/
theorem decideAt_hesitated (p : Refine.Policy Template K) (sat : A → V → Bool) {m : Memory A}
    {k : Key} (h : m k = some .empty) (v : V) : decideAt p sat m k v = .null := by
  simp [decideAt, h, Refine.tally, Refine.Refinement.empty, Refine.selected, Outcome.ofRefine]

/-- The decision at a key depends on that key's events alone: logs agreeing
on what was seen at a key decide there alike. -/
theorem decideAt_local (p : Refine.Policy Template K) (sat : A → V → Bool)
    {log₁ log₂ : List (Event A)} {k : Key}
    (h : log₁.filter (·.key = k) = log₂.filter (·.key = k)) (v : V) :
    decideAt p sat (replay log₁) k v = decideAt p sat (replay log₂) k v := by
  simp only [decideAt, replay_at, h]

end

/-! ## The tree

Every decision is one edge: an invoke or a boot pushes one frame, a return
pops one. A frame decides only while open and with nothing pending; what
it pushes is the kind its operator pushes; and a return lands its
observation in exactly the frame its dynamic link names. -/

def Tree.find (t : Tree) (id : InstLoc) : Option Frame := t.frames.find? (·.id = id)

def Tree.isOpen (t : Tree) (id : InstLoc) : Bool := !(t.closed.contains id)

/-- A frame waits while a frame it invoked has not returned. -/
def Tree.pending (t : Tree) (id : InstLoc) : Bool :=
  t.frames.any fun c => c.parent = some id && t.isOpen c.id

/-- How many frames a frame has invoked: its next invoke's index. -/
def Tree.invokes (t : Tree) (id : InstLoc) : Nat := (t.frames.filter (·.parent = some id)).length

/-- A decision's edge. An invoke pushes a frame of `scope` by `op`, under
the decider's program; a boot pushes the frame `boot` gave, under the
program passed; a return pops. -/
inductive Edge where
  | invoke (op : Op) (scope : Scope) (bindings : List Member)
  | boot (f : Frame)
  | ret (o : Observation)

/-- Direction is the edge's form, as it is the operator's. -/
def Edge.pushes : Edge → Bool
  | .ret _ => false
  | _ => true

/-- What the deciding frame may emit: an operator its kind emits, pushing
the kind that operator pushes, a delegation narrowing tool types; a boot
only from a kind that boots, of a root linked to the decider. `acts` gives
each program's tools, read through a frame's static link. -/
def admits (k : Nat) (acts : DefLoc → List DefLoc) (f : Frame) : Edge → Bool
  | .invoke op s _ => decide (op ≠ .boot ∧ op ∈ ops f.scope.kind ∧
      pushed k (acts f.program) op s.entries = some s.kind ∧
      (op = .delegate →
        (∀ b ∈ defs s.entries, b ∈ defs f.scope.entries) ∧
        (toolTypes (acts f.program) s.entries).length < (toolTypes (acts f.program) f.scope.entries).length))
  | .boot c => decide (Op.boot ∈ ops f.scope.kind ∧ c.parent = some f.id ∧ c.scope.kind = .root)
  | .ret _ => true

/-- Where a return lands: the frame its dynamic link names. -/
def Tree.land (t : Tree) (p : InstLoc) (o : Observation) : List Frame :=
  t.frames.map fun g => if g.id = p then g.append [.arrived o] else g

def Tree.step (h : Ids) (k : Nat) (acts : DefLoc → List DefLoc) (t : Tree) (fid : InstLoc) (e : Edge) : Option Tree :=
  match t.find fid with
  | none => none
  | some f =>
    if t.isOpen fid && !t.pending fid && admits k acts f e then
      match e with
      | .invoke _ s b => some ⟨t.frames ++ [push h f (t.invokes fid) s b], t.closed⟩
      | .boot c => some ⟨t.frames ++ [c], t.closed⟩
      | .ret o => match f.parent with
        | none => none
        | some p => if (t.find p).isSome then some ⟨t.land p o, fid :: t.closed⟩ else none
    else none

/-- The tree is a fold of the log: its edges, from genesis. -/
def Tree.replay (h : Ids) (k : Nat) (acts : DefLoc → List DefLoc) (t : Tree) : List (InstLoc × Edge) → Option Tree
  | [] => some t
  | (fid, e) :: log => (t.step h k acts fid e).bind (·.replay h k acts log)

section
variable {h : Ids} {k : Nat} {acts : DefLoc → List DefLoc} {t t' : Tree} {fid : InstLoc} {e : Edge}

theorem step_some (hs : t.step h k acts fid e = some t') :
    ∃ f, t.find fid = some f ∧ t.isOpen fid = true ∧ t.pending fid = false ∧ admits k acts f e = true ∧
      match e with
      | .invoke _ s b => t' = ⟨t.frames ++ [push h f (t.invokes fid) s b], t.closed⟩
      | .boot c => t' = ⟨t.frames ++ [c], t.closed⟩
      | .ret o => ∃ p, f.parent = some p ∧ (t.find p).isSome ∧ t' = ⟨t.land p o, fid :: t.closed⟩ := by
  unfold Tree.step at hs
  split at hs
  · cases hs
  · rename_i f hf
    split at hs
    · rename_i hc
      simp only [Bool.and_eq_true, Bool.not_eq_true'] at hc
      refine ⟨f, hf, hc.1.1, hc.1.2, hc.2, ?_⟩
      cases e with
      | invoke _ s b => exact (Option.some.inj hs).symm
      | boot c => exact (Option.some.inj hs).symm
      | ret o =>
        simp only at hs
        split at hs
        · cases hs
        · rename_i p hp
          split at hs
          · rename_i hpf; exact ⟨p, hp, hpf, (Option.some.inj hs).symm⟩
          · cases hs
    · cases hs

/-- A frame decides only while it is open, with nothing pending. -/
theorem step_ready (hs : t.step h k acts fid e = some t') :
    (∃ f, t.find fid = some f) ∧ t.isOpen fid = true ∧ t.pending fid = false := by
  obtain ⟨f, hf, ho, hp, -⟩ := step_some hs
  exact ⟨⟨f, hf⟩, ho, hp⟩

/-- Every decision is one push or one pop: a push adds one frame and
closes none; a return adds none and closes the frame that decided. -/
theorem step_one (hs : t.step h k acts fid e = some t') :
    (e.pushes = true ∧ t'.frames.length = t.frames.length + 1 ∧ t'.closed = t.closed) ∨
      (e.pushes = false ∧ t'.frames.length = t.frames.length ∧ t'.closed = fid :: t.closed) := by
  obtain ⟨f, -, -, -, -, he⟩ := step_some hs
  cases e with
  | invoke _ s b => rw [he]; exact .inl ⟨rfl, by simp, rfl⟩
  | boot c => rw [he]; exact .inl ⟨rfl, by simp, rfl⟩
  | ret o =>
    obtain ⟨p, -, -, he⟩ := he
    rw [he]; exact .inr ⟨rfl, by simp [Tree.land], rfl⟩

/-- An invoke pushes the decider's next child, under its program, linked
to it. -/
theorem step_invoke {op : Op} {s : Scope} {b : List Member}
    (hs : t.step h k acts fid (.invoke op s b) = some t') :
    ∃ f, t.find fid = some f ∧ t'.frames = t.frames ++ [push h f (t.invokes fid) s b] ∧
      (push h f (t.invokes fid) s b).parent = some fid := by
  obtain ⟨f, hf, -, -, -, he⟩ := step_some hs
  rw [he]
  refine ⟨f, hf, rfl, ?_⟩
  have := List.find?_some hf
  simp only [decide_eq_true_eq] at this
  simp [push, this]

/-- A delegation passes a subset of the decider's scope: its child's tools
are tools the decider holds, and strictly fewer of them. -/
theorem step_delegate {s : Scope} {b : List Member}
    (hs : t.step h k acts fid (.invoke .delegate s b) = some t') :
    ∃ f, t.find fid = some f ∧
      (∀ x ∈ toolTypes (acts f.program) s.entries, x ∈ toolTypes (acts f.program) f.scope.entries) ∧
      (toolTypes (acts f.program) s.entries).length < (toolTypes (acts f.program) f.scope.entries).length := by
  obtain ⟨f, hf, -, -, ha, -⟩ := step_some hs
  simp only [admits, decide_eq_true_eq] at ha
  obtain ⟨-, -, -, hn⟩ := ha
  obtain ⟨hsub, hlt⟩ := hn trivial
  exact ⟨f, hf, toolTypes_sub hsub, hlt⟩

/-- A return lands its observation in exactly the frame its dynamic link
names, appended as an arrival: what was returned is seen there. -/
theorem step_ret {o : Observation} (hs : t.step h k acts fid (.ret o) = some t') :
    ∃ f p, t.find fid = some f ∧ f.parent = some p ∧ (t.find p).isSome ∧
      t'.frames = t.land p o ∧ t'.closed = fid :: t.closed := by
  obtain ⟨f, hf, -, -, -, p, hp, hpf, he⟩ := step_some hs
  rw [he]; exact ⟨f, p, hf, hp, hpf, rfl, rfl⟩

/-- Landing changes no other frame. -/
theorem land_other {p : InstLoc} {o : Observation} {g : Frame} (hg : g ∈ t.frames) (hp : g.id ≠ p) :
    g ∈ t.land p o := by
  simp only [Tree.land, List.mem_map]
  exact ⟨g, hg, by simp [hp]⟩

/-- The frame landed in grows by exactly the arrival, and its key by
exactly its interface. -/
theorem land_self (rootOf : DefLoc → DefLoc) {p : InstLoc} {o : Observation} {g : Frame}
    (hg : g ∈ t.frames) (hp : g.id = p) :
    g.append [.arrived o] ∈ t.land p o ∧ (key rootOf (g.append [.arrived o])).seen = seen g ++ [o.iface] := by
  refine ⟨?_, key_arrived rootOf g o⟩
  simp only [Tree.land, List.mem_map]
  exact ⟨g, hg, by simp [hp]⟩

/-- A frame that has returned never decides again. -/
theorem step_closed (hc : fid ∈ t.closed) : t.step h k acts fid e = none := by
  cases hs : t.step h k acts fid e with
  | none => rfl
  | some t' =>
    have := (step_ready hs).2.1
    simp [Tree.isOpen, hc] at this

/-- Every return closes one frame. -/
theorem replay_closed : ∀ {t t' : Tree} (log : List (InstLoc × Edge)), t.replay h k acts log = some t' →
    t'.closed.length = t.closed.length + (log.filter fun x => !x.2.pushes).length
  | t, t', [], hr => by simp [Tree.replay] at hr; subst hr; simp
  | t, t', (fid, e) :: log, hr => by
    simp only [Tree.replay, Option.bind_eq_some_iff] at hr
    obtain ⟨t₁, h₁, h₂⟩ := hr
    rw [replay_closed log h₂]
    rcases step_one h₁ with ⟨hp, -, hc⟩ | ⟨hp, -, hc⟩
    · simp [hc, hp]
    · simp [hc, hp]; omega

/-! ### Isolation

A frame's state changes only by arrivals: what a frame it invoked
returned. Its id, links, scope and bindings never change, and no other
frame's elimination or introduction ever reaches it. -/

theorem append_append (g : Frame) (es es' : List Entry) :
    (g.append es).append es' = g.append (es ++ es') := by
  simp [Frame.append, List.append_assoc]

/-- One edge: every frame is kept, or kept with one arrival appended. -/
theorem step_keeps (hs : t.step h k acts fid e = some t') {g : Frame} (hg : g ∈ t.frames) :
    g ∈ t'.frames ∨ ∃ o, g.append [.arrived o] ∈ t'.frames := by
  obtain ⟨f, -, -, -, -, he⟩ := step_some hs
  cases e with
  | invoke _ s b => rw [he]; exact .inl (List.mem_append_left _ hg)
  | boot c => rw [he]; exact .inl (List.mem_append_left _ hg)
  | ret o =>
    obtain ⟨p, -, -, he⟩ := he
    rw [he]
    by_cases hp : g.id = p
    · exact .inr ⟨o, (land_self (fun d => d) hg hp).1⟩
    · exact .inl (land_other hg hp)

/-- A run: every frame there at its start is there at its end, with only
arrivals appended to its state. -/
theorem replay_keeps : ∀ {t t' : Tree} (log : List (InstLoc × Edge)), t.replay h k acts log = some t' →
    ∀ g ∈ t.frames, ∃ es, g.append es ∈ t'.frames ∧ ∀ x ∈ es, ∃ o, x = .arrived o
  | t, t', [], hr, g, hg => by
    simp [Tree.replay] at hr; subst hr
    exact ⟨[], by simpa [Frame.append] using hg, by simp⟩
  | t, t', (fid, e) :: log, hr, g, hg => by
    simp only [Tree.replay, Option.bind_eq_some_iff] at hr
    obtain ⟨t₁, h₁, h₂⟩ := hr
    rcases step_keeps h₁ hg with hg₁ | ⟨o, hg₁⟩
    · exact replay_keeps log h₂ g hg₁
    · obtain ⟨es, hm, ha⟩ := replay_keeps log h₂ _ hg₁
      refine ⟨[.arrived o] ++ es, by rwa [← append_append], ?_⟩
      intro x hx
      rcases List.mem_append.mp hx with hx | hx
      · simp at hx; exact ⟨o, hx⟩
      · exact ha x hx

/-! ### Well typed

A tree is well typed when every frame without a dynamic link is a root,
and every other frame's parent is in the tree and emitted an operator its
kind emits, pushing the frame's kind, a delegation narrowing tool types. -/

def WellTyped (k : Nat) (acts : DefLoc → List DefLoc) (t : Tree) : Prop :=
  (∀ c ∈ t.frames, c.parent = none → c.scope.kind = .root) ∧
  (∀ c ∈ t.frames, ∀ p, c.parent = some p → ∃ f, t.find p = some f ∧ ∃ op ∈ ops f.scope.kind,
    pushed k (acts c.program) op c.scope.entries = some c.scope.kind ∧
    (op = .delegate →
      (toolTypes (acts c.program) c.scope.entries).length < (toolTypes (acts f.program) f.scope.entries).length))

/-- Genesis: one root, no link, is well typed. -/
theorem wellTyped_genesis {f : Frame} (hp : f.parent = none) (hr : f.scope.kind = .root) :
    WellTyped k acts ⟨[f], []⟩ := by
  refine ⟨?_, ?_⟩
  · intro c hc _; simp at hc; subst hc; exact hr
  · intro c hc p hcp; simp at hc; subst hc; simp [hp] at hcp

theorem find_append {l : List Frame} {c : Frame} {p : InstLoc} {f : Frame}
    (h : l.find? (·.id = p) = some f) : (l ++ [c]).find? (·.id = p) = some f := by
  rw [List.find?_append, h]; rfl

theorem find_land {p q : InstLoc} {o : Observation} :
    (t.land p o).find? (·.id = q) =
      (t.find q).map fun g => if g.id = p then g.append [.arrived o] else g := by
  unfold Tree.land Tree.find
  generalize t.frames = l
  induction l with
  | nil => rfl
  | cons g gs ih =>
    by_cases hq : g.id = q
    · subst hq; by_cases hg : g.id = p <;> simp [hg, append_id]
    · by_cases hg : g.id = p
      · have hpq : ¬ p = q := hg ▸ hq
        simp [hg, hpq, append_id, ih]
      · simp [hg, hq, ih]

/-- The appended frame keeps its scope and links. -/
theorem land_scope {p : InstLoc} {o : Observation} (g : Frame) :
    let g' := if g.id = p then g.append [.arrived o] else g
    g'.scope = g.scope ∧ g'.parent = g.parent ∧ g'.id = g.id ∧ g'.program = g.program := by
  by_cases hg : g.id = p <;> simp [hg, Frame.append]

/-- What a pushed frame needs of its parent to keep the tree well typed. -/
def Admitted (k : Nat) (acts : DefLoc → List DefLoc) (f c : Frame) : Prop :=
  ∃ op ∈ ops f.scope.kind, pushed k (acts c.program) op c.scope.entries = some c.scope.kind ∧
    (op = .delegate →
      (toolTypes (acts c.program) c.scope.entries).length < (toolTypes (acts f.program) f.scope.entries).length)

theorem wellTyped_push {cl : List InstLoc} {f c : Frame} (hw : WellTyped k acts t)
    (hf : t.find fid = some f) (hc : c.parent = some fid) (ha : Admitted k acts f c) :
    WellTyped k acts ⟨t.frames ++ [c], cl⟩ := by
  refine ⟨?_, ?_⟩
  · intro c' hc' hn
    rcases List.mem_append.mp hc' with hc' | hc'
    · exact hw.1 c' hc' hn
    · simp at hc'; subst hc'; simp [hc] at hn
  · intro c' hc' p hp
    rcases List.mem_append.mp hc' with hc' | hc'
    · obtain ⟨f', hf', hr⟩ := hw.2 c' hc' p hp
      exact ⟨f', find_append hf', hr⟩
    · simp at hc'; subst hc'
      rw [hc] at hp; cases hp
      exact ⟨f, find_append hf, ha⟩

theorem wellTyped_land {cl : List InstLoc} {p : InstLoc} {o : Observation} (hw : WellTyped k acts t) :
    WellTyped k acts ⟨t.land p o, cl⟩ := by
  refine ⟨?_, ?_⟩
  · intro c' hc' hn
    simp only [Tree.land, List.mem_map] at hc'
    obtain ⟨g, hg, rfl⟩ := hc'
    obtain ⟨hs, hpar, -, -⟩ := land_scope (p := p) (o := o) g
    rw [hs]; exact hw.1 g hg (hpar ▸ hn)
  · intro c' hc' q hq
    simp only [Tree.land, List.mem_map] at hc'
    obtain ⟨g, hg, rfl⟩ := hc'
    obtain ⟨hs, hpar, -, hpg⟩ := land_scope (p := p) (o := o) g
    obtain ⟨f', hf', op, hop, hpu, hn⟩ := hw.2 g hg q (hpar ▸ hq)
    obtain ⟨hs', -, -, hpf⟩ := land_scope (p := p) (o := o) f'
    refine ⟨if f'.id = p then f'.append [.arrived o] else f', ?_, op, ?_, ?_, ?_⟩
    · show (t.land p o).find? (·.id = q) = _
      rw [find_land, hf']; rfl
    · rw [hs']; exact hop
    · rw [hs, hpg]; exact hpu
    · intro hd; rw [hs, hs', hpg, hpf]; exact hn hd

/-- Every edge keeps the tree well typed: the tree is well typed after
every edge from genesis. -/
theorem step_wellTyped (hw : WellTyped k acts t) (hs : t.step h k acts fid e = some t') : WellTyped k acts t' := by
  obtain ⟨f, hf, -, -, ha, he⟩ := step_some hs
  have hfid : f.id = fid := by have := List.find?_some hf; simpa using this
  cases e with
  | invoke op s b =>
    rw [he]
    simp only [admits, decide_eq_true_eq] at ha
    obtain ⟨-, hop, hpu, hn⟩ := ha
    exact wellTyped_push hw hf (by simp [push, hfid]) ⟨op, hop, hpu, fun hd => (hn hd).2⟩
  | boot c =>
    rw [he]
    simp only [admits, decide_eq_true_eq] at ha
    obtain ⟨hop, hpar, hr⟩ := ha
    exact wellTyped_push hw hf (hfid ▸ hpar) ⟨.boot, hop, by simp [pushed, hr], by simp⟩
  | ret o =>
    obtain ⟨p, -, -, he⟩ := he
    rw [he]; exact wellTyped_land hw

theorem replay_wellTyped : ∀ {t t' : Tree} (log : List (InstLoc × Edge)),
    WellTyped k acts t → t.replay h k acts log = some t' → WellTyped k acts t'
  | t, t', [], hw, hr => by simp [Tree.replay] at hr; subst hr; exact hw
  | t, t', (fid, e) :: log, hw, hr => by
    simp only [Tree.replay, Option.bind_eq_some_iff] at hr
    obtain ⟨t₁, h₁, h₂⟩ := hr
    exact replay_wellTyped log (step_wellTyped hw h₁) h₂

/-! ### Depth is bounded

Within a process, every frame's measure is below its parent's: a leaf's
is 0, an agent's its tool types. A delegation narrows, an act frame holds
a tool, and only a boot, which pushes a root, starts a new process. So a
chain of frames within a process is no longer than its first frame's tool
types. -/

def measure (act : List DefLoc) (s : Scope) : Nat := if s.kind = .leaf then 0 else (toolTypes act s.entries).length

/-- An act frame in a well-typed tree holds a tool. -/
theorem act_has_tool (hw : WellTyped k acts t) {f : Frame} (hf : f ∈ t.frames) (ha : f.scope.kind = .act) :
    1 ≤ (toolTypes (acts f.program) f.scope.entries).length := by
  cases hp : f.parent with
  | none => have := hw.1 f hf hp; rw [ha] at this; cases this
  | some q =>
    obtain ⟨-, -, op, -, hpu, -⟩ := hw.2 f hf q hp
    rw [ha] at hpu
    cases op with
    | delegate => exact ((delegate_act k _ _).mp hpu).1
    | act => simp [pushed] at hpu
    | boot => simp [pushed] at hpu
    | ret => simp [pushed] at hpu

/-- Within a process, a frame's measure is below its parent's. -/
theorem measure_lt (hw : WellTyped k acts t) {c f : Frame} {p : InstLoc} (hc : c ∈ t.frames)
    (hp : c.parent = some p) (hf : t.find p = some f) (hr : c.scope.kind ≠ .root) :
    measure (acts c.program) c.scope < measure (acts f.program) f.scope := by
  obtain ⟨f', hf', op, hop, hpu, hn⟩ := hw.2 c hc p hp
  rw [hf] at hf'; cases hf'
  have hfm : f ∈ t.frames := List.mem_of_find?_eq_some hf
  cases op with
  | act =>
    have hk : f.scope.kind = .act := (act_emitted _).mp hop
    have hl : c.scope.kind = .leaf := by simp [pushed] at hpu; exact hpu.symm
    have := act_has_tool hw hfm hk
    simp [measure, hl, hk]; omega
  | delegate =>
    have hcl : c.scope.kind ≠ .leaf := fun e => agentKind_ne_leaf k _ c.scope.entries (e ▸ hpu)
    have hfl : f.scope.kind ≠ .leaf := by intro e; rw [e] at hop; simp [ops] at hop
    simp [measure, hcl, hfl]; exact hn rfl
  | boot => simp [pushed] at hpu; exact absurd hpu.symm hr
  | ret => simp [pushed] at hpu

/-- A chain of frames, each the dynamic link's target of the next. -/
inductive Chain (t : Tree) : List Frame → Prop
  | one (f : Frame) : Chain t [f]
  | cons (f c : Frame) (cs : List Frame) : c ∈ t.frames → c.parent = some f.id → t.find f.id = some f →
      Chain t (c :: cs) → Chain t (f :: c :: cs)

/-- Depth is bounded: a chain within a process, no frame below its first a
root, is no longer than its first frame's measure. -/
theorem depth (hw : WellTyped k acts t) : ∀ {f : Frame} {cs : List Frame}, Chain t (f :: cs) →
    (∀ c ∈ cs, c.scope.kind ≠ .root) → cs.length ≤ measure (acts f.program) f.scope
  | _, [], _, _ => Nat.zero_le _
  | f, c :: cs, .cons _ _ _ hc hp hf hch, hr => by
    have h₁ := measure_lt hw hc hp hf (hr c List.mem_cons_self)
    have h₂ := depth hw hch fun c' hc' => hr c' (List.mem_cons_of_mem _ hc')
    simp only [List.length_cons]; omega

end

end Model
