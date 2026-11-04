import Lean.Data.Json
import Invariants

/-! Writes the test vectors: Lean's own answers on a fixed, finite set of
inputs, one JSON file per function. Every input is enumerated within its
bounds, with no randomness, so the files are the same on every run. Every
contract is made by `identify` and every observation by `observe`; nothing
here reaches a private constructor. The Go core replays every case.

Every case that needs the apprentice uses one: the first case of
`apprenticeOf.json`, the apprentice as built. It is written once there, not
in every case. -/

open Lean (Json)

namespace Vectors

/-! ## Enumeration -/

def subsets : List α → List (List α)
  | [] => [[]]
  | x :: xs => let r := subsets xs; r ++ r.map (x :: ·)

def exactly (xs : List α) : Nat → List (List α)
  | 0 => [[]]
  | n + 1 => xs.flatMap fun x => (exactly xs n).map (x :: ·)

/-- Every list of length at most `n` drawn from `xs`, repeats allowed. -/
def upTo (xs : List α) (n : Nat) : List (List α) :=
  (List.range (n + 1)).flatMap (exactly xs)

/-! ## Encoding -/

def jList (f : α → Json) (xs : List α) : Json := .arr (xs.map f).toArray

def jOpt (f : α → Json) : Option α → Json
  | none => .null
  | some a => f a

def jBinary (b : Binary) : Json := .str b.val
def jSchema (s : Schema) : Json := .str s.val

def jContract (c : Contract) : Json :=
  Json.mkObj [("binary", jBinary c.binary), ("shape", jSchema c.shape)]

def jType (t : ObservationType) : Json :=
  Json.mkObj [("contract", jContract t.contract), ("output", jSchema t.output)]

partial def jObs (o : Observation) : Json :=
  Json.mkObj [("type", jType o.type), ("refined", .str o.refined.val),
    ("members", .arr (o.members.map jObs).toArray)]

def jDir : Dir → Json
  | .invoke => .str "invoke"
  | .ret => .str "ret"

def jKind : Kind → Json
  | .act => .str "act"
  | .delegate => .str "delegate"

def jOffer (o : Offer) : Json :=
  Json.mkObj [("invoke", jList jContract o.invoke), ("passOnly", jList jContract o.passOnly),
    ("apprentice", jList jContract o.apprentice)]

def jApprentice (a : Apprentice) : Json :=
  Json.mkObj [("binary", jBinary a.binary), ("observe", jContract a.observe),
    ("root", jContract a.root), ("exit", jContract a.exit), ("return", jContract a.ret),
    ("perform_effect", jContract a.performEffect), ("create_input", jContract a.createInput),
    ("delegate", jContract a.delegate), ("goal", jContract a.goal),
    ("create_delegation", jContract a.createDelegation), ("create_act", jContract a.createAct),
    ("package_reach", jContract a.reach)]

def jInputBase (i : InputBase) : Json := .str i.val

def jSeen : Seen → Json
  | .seen => .str "seen"
  | .unseen => .str "unseen"

def jReturnResult : ReturnResult → Json
  | .continue => .str "continue"
  | .fallback => .str "fallback"

def jInvokeResult : InvokeResult → Json
  | .procedure => .str "procedure"
  | .model => .str "model"

def jBase (b : ReturnBase) : Json := .str b.val

def jGuard : Guard b → Json
  | .default => .str "default"
  | .written cel => Json.mkObj [("written", .str cel)]

/-- A store over a finite domain of base types: what is known, and the guard
in every slot of the domain that has one. -/
def jStore (domain : List ReturnBase) (s : Store) : Json :=
  Json.mkObj [("known", jList jBase s.known),
    ("guards", .arr (domain.filterMap fun b =>
      (s.guards b).map fun g => Json.mkObj [("base", jBase b), ("guard", jGuard g)]).toArray)]

def case (i o : Json) : Json := Json.mkObj [("in", i), ("out", o)]

/-! ## The world: a few contracts, types and observations -/

def hex (c : Char) (n : Nat) : String := String.ofList (List.replicate n c)

def binaries : List Binary :=
  [ ⟨"cdhash:" ++ hex 'a' 40⟩, ⟨"sha256:" ++ hex '1' 64⟩, ⟨"cdhash:" ++ hex 'f' 40⟩,
    ⟨"cdhash:" ++ hex 'a' 39⟩, ⟨"sha256:" ++ hex 'A' 64⟩, ⟨"cdhash:" ++ hex 'g' 40⟩,
    ⟨"md5:" ++ hex 'a' 32⟩, ⟨""⟩ ]

def shapeIn : Schema := ⟨"{\"properties\":{\"path\":{\"type\":\"string\"}},\"type\":\"object\"}"⟩
def shapeOut : Schema := ⟨"{\"items\":{\"type\":\"number\"},\"type\":\"array\"}"⟩

def shapeStr : Schema := ⟨"{\"type\":\"string\"}"⟩

def schemas : List Schema := [⟨""⟩, Schema.empty, shapeIn, shapeStr, shapeOut]

def outputs : List Schema := [shapeStr, shapeOut]

def kinds : List Kind := [.act, .delegate]

def bApp : Binary := ⟨"cdhash:" ++ hex 'f' 40⟩

/-- A contract of the apprentice, by its title. -/
def appContract (title : String) : Option Contract :=
  identify bApp ⟨"{\"title\":\"" ++ title ++ "\",\"type\":\"object\"}"⟩

def titles : List String :=
  ["observe", "root", "exit", "return", "perform_effect", "create_input", "delegate", "goal",
   "create_delegation", "create_act", "package_reach"]

structure World where
  app : Apprentice
  /-- Contracts outside the apprentice: `c1` and `c2` are two shapes of one
  binary. -/
  c1 : Contract
  c2 : Contract
  c3 : Contract
  t1 : ObservationType
  t2 : ObservationType
  /-- `e1` again: the same contract and shape, another value, so another
  refinement type and another observation. -/
  e3 : Observation
  /-- Effects: two that came back through an act node's `perform_effect`,
  an exit, an act description, and the root's endowment; and a goal, cut
  from the endowment, its one member. -/
  e1 : Observation
  e2 : Observation
  g1 : Observation
  x1 : Observation
  w1 : Observation
  r0 : Observation
  /-- Computations a delegate composed: `k3` holds `k1`, so lineages are two
  deep. -/
  k1 : Observation
  k2 : Observation
  k3 : Observation

def world : Option World := do
  let [observe', root, exit, ret, pe, ci, delegate, goal, cd, ca, reach] ← titles.mapM appContract
    | none
  let app ← Apprentice.of bApp observe' root exit ret pe ci delegate goal cd ca reach
  let c1 ← identify ⟨"cdhash:" ++ hex 'a' 40⟩ Schema.empty
  let c2 ← identify ⟨"cdhash:" ++ hex 'a' 40⟩ shapeIn
  let c3 ← identify ⟨"sha256:" ++ hex '1' 64⟩ Schema.empty
  let t1 ← ObservationType.of c1 shapeOut
  let t2 ← ObservationType.of c3 shapeOut
  let tg ← ObservationType.of app.goal shapeStr
  let tx ← ObservationType.of app.exit shapeStr
  let tw ← ObservationType.of app.createAct shapeStr
  let tr ← ObservationType.of app.root shapeStr
  let td ← ObservationType.of app.delegate shapeOut
  let ρ (v : String) : RefinementId := ⟨"seen(" ++ v ++ ")"⟩
  let e1 ← observe app .act t1 [] (ρ "e1")
  let e2 ← observe app .act t2 [] (ρ "e2")
  let e3 ← observe app .act t1 [] (ρ "e3")
  let r0 ← rootObservation app tr (ρ "r0")
  let g1 ← observe app .delegate tg [r0] (ρ "g1")
  let x1 ← observe app .act tx [] (ρ "x1")
  let w1 ← observe app .delegate tw [] (ρ "w1")
  let k1 ← observe app .delegate td [g1, e1] (ρ "k1")
  let k2 ← observe app .delegate td [g1, e2] (ρ "k2")
  let k3 ← observe app .delegate td [k1, e2] (ρ "k3")
  return { app, c1, c2, c3, t1, t2, e1, e2, e3, g1, x1, w1, r0, k1, k2, k3 }

/-- Every contract: the agents', then the apprentice's. -/
def contracts (w : World) : List Contract := [w.c1, w.c2, w.c3] ++ w.app.contracts

def types (w : World) : List ObservationType :=
  (contracts w).flatMap fun c => outputs.filterMap (ObservationType.of c)

/-! ## The two edges, and refinement types -/

/-- The hashes the vectors use. Lean never computes a real hash; the Go test
passes the same functions. -/
def toy : Hashes where
  input c ps := "input(" ++ c.binary.val ++ "|" ++ c.shape.val ++ "," ++ toString (ps.map (·.val)) ++ ")"
  ret i o := "return(" ++ i.val ++ "," ++ o.val ++ ")"
  refine b g := ⟨"refine(" ++ b.val ++ "," ++
    (match g with | .default => "default" | .written cel => "written:" ++ cel) ++ ")"⟩
  seen b v := ⟨"seen(" ++ b.val ++ "," ++ v ++ ")"⟩

def bA : Binary := ⟨"cdhash:" ++ hex 'a' 40⟩
def bB : Binary := ⟨"sha256:" ++ hex '1' 64⟩

/-- The contracts the edges are crossed to: two shapes of one binary, and
another binary. -/
def edgeContracts : List Contract :=
  [(bA, Schema.empty), (bA, shapeIn), (bB, Schema.empty)].filterMap fun (b, s) => identify b s

/-- Two inputs with nothing passed, to two binaries. -/
def inputs : List InputBase :=
  match edgeContracts with
  | [cA, _, cB] => [cA, cB].map (createInvoke toy · [])
  | _ => []

/-- Three return base types: two shapes back from one input, one from the
other. -/
def domain : List ReturnBase :=
  match inputs with
  | [iA, iB] => [(iA, shapeOut), (iA, shapeIn), (iB, shapeOut)].filterMap
      fun (i, o) => createReturn toy i o
  | _ => []

def cels : List String := ["cel-1", "cel-2"]

inductive Op
  | see (b : ReturnBase)
  | rewrite (b : ReturnBase) (cel : String)

def Op.apply (s : Store) : Op → Option Store
  | .see b => some (_root_.see s b).2
  | .rewrite b cel => _root_.rewrite s b cel

def ops : List Op :=
  domain.map .see ++ domain.flatMap fun b => cels.map (.rewrite b)

/-- Every store reachable from the empty one in at most three operations,
each once. -/
def stores : List Store := Id.run do
  let mut seen : List String := []
  let mut out : List Store := []
  let mut frontier : List Store := [Store.empty]
  for _ in [0:4] do
    let mut next : List Store := []
    for s in frontier do
      let k := (jStore domain s).compress
      if !seen.contains k then
        seen := seen ++ [k]
        out := out ++ [s]
        next := next ++ ops.filterMap (·.apply s)
    frontier := next
  return out

end Vectors

/-! ## The files -/

open Vectors in
def files (w : World) : List (String × List Json) :=
  let ts := types w
  let cs := contracts w
  let agents := [w.c1, w.c2, w.c3]
  let members := upTo [w.e1, w.e3, w.g1, w.k1] 2
  let ρv : RefinementId := ⟨"seen(made)"⟩
  [ ("identify", binaries.flatMap fun b => schemas.map fun s =>
      case (Json.mkObj [("binary", jBinary b), ("shape", jSchema s)])
        (jOpt jContract (identify b s))),
    ("observationTypeOf", cs.flatMap fun c => schemas.map fun s =>
      case (Json.mkObj [("contract", jContract c), ("output", jSchema s)])
        (jOpt jType (ObservationType.of c s))),
    -- The apprentice as built, then refused: a contract of another binary,
    -- and one contract in two places.
    ("apprenticeOf",
      let a := w.app
      let cands : List (List Contract) :=
        [a.contracts, a.contracts.set 3 w.c1, a.contracts.set 7 a.exit]
      cands.map fun l => match l with
        | [o, r, e, t, pe, ci, d, g, cd, ca, rc] =>
          case (Json.mkObj [("binary", jBinary bApp), ("contracts", jList jContract l)])
            (jOpt jApprentice (Apprentice.of bApp o r e t pe ci d g cd ca rc))
        | _ => .null),
    ("admits", kinds.flatMap fun k => ts.flatMap fun t => members.map fun ms =>
      case (Json.mkObj [("kind", jKind k), ("type", jType t),
          ("members", jList jObs ms)])
        (.bool (admits w.app k t ms))),
    ("observe", kinds.flatMap fun k => ts.flatMap fun t => members.map fun ms =>
      case (Json.mkObj [("kind", jKind k), ("type", jType t),
          ("members", jList jObs ms), ("refined", .str ρv.val)])
        (jOpt jObs (observe w.app k t ms ρv))),
    ("rootObservation", ts.map fun t =>
      case (Json.mkObj [("type", jType t), ("refined", .str ρv.val)])
        (jOpt jObs (rootObservation w.app t ρv))),
    ("mayPass", (subsets [w.e1, w.e2, w.g1, w.k1]).flatMap fun sender =>
      (upTo [w.e1, w.e2, w.e3, w.g1, w.k1] 2).map fun passed =>
        case (Json.mkObj [("sender", jList jObs sender), ("passed", jList jObs passed)])
          (.bool (mayPass sender passed))),
    ("receive", [Dir.invoke, .ret].flatMap fun d => (subsets [w.e1, w.g1]).flatMap fun r =>
      (upTo [w.e1, w.e2] 2).map fun p =>
        case (Json.mkObj [("dir", jDir d), ("receiver", jList jObs r), ("passed", jList jObs p)])
          (jList jObs (receive d r p))),
    ("lineage", [w.e1, w.e2, w.g1, w.x1, w.w1, w.r0, w.k1, w.k2, w.k3].map fun o =>
      case (Json.mkObj [("observation", jObs o)]) (jList jObs (lineage o))),
    ("nodeMayReturn", (upTo [w.g1, w.e1, w.k1, w.x1] 2).flatMap fun received =>
      (upTo [w.e1, w.e2, w.x1, w.k1, w.k3] 2).map fun returned =>
        case (Json.mkObj [("received", jList jObs received), ("returned", jList jObs returned)])
          (.bool (nodeMayReturn received returned))),
    ("mayReturn", (upTo [w.g1, w.w1, w.e1] 2).flatMap fun created =>
      (upTo [w.g1, w.w1, w.e1, w.e2, w.k1] 2).map fun returned =>
        case (Json.mkObj [("created", jList jObs created), ("returned", jList jObs returned)])
          (.bool (mayReturn created returned))),
    ("classify", agents.flatMap fun self =>
      (upTo (agents ++ [w.app.exit, w.app.goal, w.app.observe]) 3).map fun reach =>
        case (Json.mkObj [("self", jContract self),
            ("reach", jList jContract reach)])
          (jOpt jKind (classify w.app self reach))),
    ("rootKind", agents.flatMap fun self => kinds.flatMap fun k =>
      (upTo (agents ++ [w.app.exit]) 2).map fun reach =>
        case (Json.mkObj [("self", jContract self), ("reach", jList jContract reach),
            ("kind", jKind k)])
          (jOpt jKind (rootKind w.app self reach k))),
    ("parts", kinds.map fun k =>
      case (Json.mkObj [("kind", jKind k)]) (jList jContract (parts w.app k))),
    ("loaded", kinds.map fun k =>
      case (Json.mkObj [("kind", jKind k)])
        (jList jContract (loaded w.app k))),
    ("passable", (subsets agents).flatMap fun held => (upTo agents 2).map fun passed =>
      case (Json.mkObj [("held", jList jContract held), ("passed", jList jContract passed)])
        (.bool (passable held passed))),
    ("offered", kinds.flatMap fun k => (upTo agents 2).map fun passed =>
      case (Json.mkObj [("kind", jKind k),
          ("passed", jList jContract passed)])
        (jOffer (offered w.app k passed))),
    ("createInvoke", edgeContracts.flatMap fun c => (upTo domain 2).map fun ps =>
      case (Json.mkObj [("contract", jContract c), ("passed", jList jBase ps)])
        (jInputBase (createInvoke toy c ps))),
    ("createReturn", (inputs ++ match edgeContracts with
        | cA :: _ => (domain.take 2).map (createInvoke toy cA [·])
        | [] => []).flatMap
      fun i => schemas.map fun o =>
        case (Json.mkObj [("input", jInputBase i), ("shape", jSchema o)])
          (jOpt jBase (createReturn toy i o))),
    ("do",
      let ibs : List InputBase := match domain, edgeContracts with
        | [d0, d1, _], [cA, cA', cB] =>
            [(cA, []), (cA, [d0]), (cA, [d0, d1]), (cA, [d1, d0]), (cA', []), (cB, [d0])].map
              fun (c, ps) => createInvoke toy c ps
        | _, _ => []
      (subsets ibs).flatMap fun history => ibs.map fun i =>
        case (Json.mkObj [("history", jList jInputBase history), ("input", jInputBase i)])
          (jSeen («do» history i))),
    ("baseOf", ts.flatMap fun t => (upTo domain 1).map fun ps =>
      case (Json.mkObj [("type", jType t), ("passed", jList jBase ps)])
        (jOpt jBase (baseOf toy t ps))),
    ("see", stores.flatMap fun s => domain.map fun b =>
      let (n, s') := see s b
      case (Json.mkObj [("store", jStore domain s), ("base", jBase b)])
        (Json.mkObj [("seen", jSeen n), ("store", jStore domain s')])),
    ("rewrite", stores.flatMap fun s => domain.flatMap fun b => cels.map fun cel =>
      case (Json.mkObj [("store", jStore domain s), ("base", jBase b), ("cel", .str cel)])
        (jOpt (jStore domain) (rewrite s b cel))),
    ("refinement", stores.flatMap fun s => domain.map fun b =>
      case (Json.mkObj [("store", jStore domain s), ("base", jBase b)])
        (jOpt (fun r => .str r.val) (refinement toy s b))),
    ("ret", [Seen.seen, .unseen].flatMap fun n => [true, false].map fun g =>
      case (Json.mkObj [("seen", jSeen n), ("guardHolds", .bool g)]) (jReturnResult (ret n g))),
    ("invoke", [Seen.seen, .unseen].flatMap fun n => [true, false].map fun g =>
      case (Json.mkObj [("seen", jSeen n), ("guardHolds", .bool g)])
        (jInvokeResult (invoke n g))) ]

/-- One case per line, so a change shows as a line in a diff. -/
def render (name : String) (cases : List Json) : String :=
  "{\"function\":\"" ++ name ++ "\",\"cases\":[\n" ++
    ",\n".intercalate (cases.map (·.compress)) ++ "\n]}\n"

def main (args : List String) : IO UInt32 := do
  let dir : System.FilePath := args.headD "vectors"
  let some w := Vectors.world
    | IO.eprintln "vectors: the world did not build"; return 1
  IO.FS.createDirAll dir
  for (name, cases) in files w do
    IO.FS.writeFile (dir / (name ++ ".json")) (render name cases)
    IO.println s!"{name}: {cases.length}"
  return 0
