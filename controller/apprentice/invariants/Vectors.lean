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

/-! ## Proposal 58: refinements

The atoms compare one field, `arrived.size`, with a constant, or test it
for evenness; the values are that field's. Conditions are written as CEL, the way the Go receives them.
Overlap is decided exactly for comparisons by testing every half-point
around each constant (so `x > 7 && x < 8` overlaps, as it does over
numbers), and left undecided whenever `even` appears. -/

inductive Atom where
  | gt (k : Int) | lt (k : Int) | eq (k : Int) | even
  deriving Repr, DecidableEq

def Atom.sat : Atom → Int → Bool
  | .gt k, x => x > k
  | .lt k, x => x < k
  | .eq k, x => x == k
  | .even, x => x % 2 == 0

/-- The atom on doubled values: `p` stands for `p / 2`. -/
def Atom.half : Atom → Int → Bool
  | .gt k, p => p > 2 * k
  | .lt k, p => p < 2 * k
  | .eq k, p => p == 2 * k
  | .even, _ => false

/-- Written as a stored guard reads it: the arrival's `size` field, which
the model wrote as `$1.size`. -/
def Atom.cel : Atom → String
  | .gt k => s!"arrived.size > {k}"
  | .lt k => s!"arrived.size < {k}"
  | .eq k => s!"arrived.size == {k}"
  | .even => "int(arrived.size) % 2 == 0"

def Atom.const : Atom → List Int
  | .gt k | .lt k | .eq k => [k]
  | .even => []

def condCel : Refine.Cond Atom → String
  | .atom a => a.cel
  | .tt => "true"
  | .and p q => "(" ++ condCel p ++ ") && (" ++ condCel q ++ ")"
  | .not p => "!(" ++ condCel p ++ ")"

def condAtoms : Refine.Cond Atom → List Atom
  | .atom a => [a]
  | .tt => []
  | .and p q => condAtoms p ++ condAtoms q
  | .not p => condAtoms p

def overlaps (p q : Refine.Cond Atom) : Option Bool :=
  let atoms := condAtoms p ++ condAtoms q
  if atoms.contains .even then none
  else
    let ks := atoms.flatMap Atom.const
    let points := if ks.isEmpty then [0] else ks.flatMap fun k => [2 * k - 1, 2 * k, 2 * k + 1]
    some (points.any fun x => p.eval Atom.half x && q.eval Atom.half x)

abbrev RBranch := Refine.Branch Atom String
abbrev RRef := Refine.Refinement Atom String

def jNat (n : Nat) : Json := Lean.Json.num (Lean.JsonNumber.fromNat n)

def jBranch (b : RBranch) : Json :=
  Json.mkObj [("id", jNat b.id), ("rank", jNat b.rank), ("cond", .str (condCel b.cond)),
    ("instr", .str b.instr)]

def jRef (r : RRef) : Json :=
  Json.mkObj [("next", jNat r.next), ("branches", jList jBranch r.branches)]

def jRecord (x : Refine.Record) : Json :=
  Json.mkObj [("older", jNat x.older), ("retired", jNat x.retired), ("remainder", jNat x.remainder)]

def jPossible (x : Nat × Nat) : Json := .arr #[jNat x.1, jNat x.2]

def jOutcome : Refine.Outcome String → Json
  | .null => .null
  | .select i => Json.mkObj [("select", .str i)]
  | .conflict => .str "conflict"

def jAfter : Refine.After String → Json
  | .startleAgain => .str "startle"
  | .trust i l => Json.mkObj [("trust", .str i), ("learned", .bool l)]

/-- The guard as the model writes it: `c₀ ? "i₀" : c₁ ? "i₁" : null`. -/
def guardCel : List (Refine.Cond Atom × String) → String
  | [] => "null"
  | (c, i) :: rest => condCel c ++ " ? \"" ++ i ++ "\" : " ++ guardCel rest

def links : List (Refine.Cond Atom × String) :=
  let conds : List (Refine.Cond Atom) :=
    [.atom (.gt 2), .atom (.gt 5), .atom (.gt 8), .atom (.lt 5), .atom (.eq 5), .atom .even, .tt]
  conds.flatMap fun c => ["A", "B"].map (c, ·)

def chains1 : List (List (Refine.Cond Atom × String)) := links.map ([·])

def chains2 : List (List (Refine.Cond Atom × String)) :=
  (links.flatMap fun a => links.map fun b => [a, b]).take 40

def commitR (r : RRef) (chain : List (Refine.Cond Atom × String)) : RRef :=
  (Refine.commit overlaps r chain).1

/-- The empty refinement and every one a single one-link guard makes. -/
def refs1 : List RRef := Refine.Refinement.empty :: chains1.map (commitR .empty)

/-- Every refinement two one-link guards make, one after the other. -/
def refs2 : List RRef := chains1.flatMap fun a => chains1.map fun b => commitR (commitR .empty a) b

/-- Refinements with possibly overlapping branches, as commit leaves them
when overlap is undecided: conflicts for `resolve` to find. -/
def refsPossible : List RRef :=
  chains1.flatMap fun a => chains1.filterMap fun b =>
    let r := commitR (commitR .empty a) b
    if (Refine.commit overlaps (commitR .empty a) b).2.2.isEmpty then none else some r

def values : List Int := [0, 2, 3, 5, 6, 8, 9]

def jInt (x : Int) : Json := Lean.Json.num (Lean.JsonNumber.fromInt x)

def refineFiles : List (String × List Json) :=
  let case (i o : Json) : Json := Json.mkObj [("in", i), ("out", o)]
  [ ("refineOverlaps", links.flatMap fun (p, _) => links.filterMap fun (q, i) =>
      if i = "A" then some (case (Json.mkObj [("p", .str (condCel p)), ("q", .str (condCel q))])
        (jOpt Json.bool (overlaps p q))) else none),
    ("refineCommit", refs1.flatMap fun r => (chains1 ++ chains2).map fun chain =>
      let (r', recs, poss) := Refine.commit overlaps r chain
      case (Json.mkObj [("refinement", jRef r), ("guard", .str (guardCel chain))])
        (Json.mkObj [("refinement", jRef r'), ("records", jList jRecord recs),
          ("possible", jList jPossible poss)])),
    ("refineDistinguish", refs2.flatMap fun r => values.map fun v =>
      case (Json.mkObj [("refinement", jRef r), ("x", jInt v)])
        (jOutcome (Refine.distinguish Atom.sat r.branches v))),
    ("refineResolve", refsPossible.flatMap fun r => values.map fun v =>
      let (o, r', recs) := Refine.resolve Atom.sat r.branches.length r v
      case (Json.mkObj [("refinement", jRef r), ("x", jInt v)])
        (Json.mkObj [("outcome", jOutcome o), ("refinement", jRef r'), ("records", jList jRecord recs)])),
    ("refineAfterStartle", (chains1 ++ chains2.take 10).flatMap fun chain =>
      refs1.flatMap fun r => [3, 6, 9].map fun v =>
        case (Json.mkObj [("guard", .str (guardCel chain)), ("committed", jRef r), ("x", jInt v)])
          (jAfter (Refine.afterStartle Atom.sat chain r.branches v))) ]

/-! ## Proposal 58: shapes

Each case is bytes and the contract's declared output; Lean gives what
arrives and its shape's canonical encoding. JSON cases are rendered from
values, duplicate keys included, so the Go decodes the same bytes. -/

mutual
def svText : Shape.Value → String
  | .null => "null"
  | .bool b => if b then "true" else "false"
  | .num lit => lit
  | .str s => "\"" ++ Shape.escape s ++ "\""
  | .arr xs => "[" ++ svTexts xs ++ "]"
  | .obj kvs => "{" ++ svMembers kvs ++ "}"
  | .bytes _ => ""
def svTexts : List Shape.Value → String
  | [] => ""
  | [x] => svText x
  | x :: xs => svText x ++ "," ++ svTexts xs
def svMembers : List (String × Shape.Value) → String
  | [] => ""
  | [(k, v)] => "\"" ++ Shape.escape k ++ "\":" ++ svText v
  | (k, v) :: kvs => "\"" ++ Shape.escape k ++ "\":" ++ svText v ++ "," ++ svMembers kvs
end

mutual
def svJson : Shape.Value → Json
  | .null => .null
  | .bool b => .bool b
  | .num lit => match lit.toNat? with
    | some n => Lean.Json.num (Lean.JsonNumber.fromNat n)
    | none => .str lit
  | .str s => .str s
  | .arr xs => .arr (svJsons xs).toArray
  | .obj kvs => Json.mkObj (svJsonMembers kvs)
  | .bytes b => .arr (b.toList.map fun x => Lean.Json.num (Lean.JsonNumber.fromNat x.toNat)).toArray
def svJsons : List Shape.Value → List Json
  | [] => []
  | x :: xs => svJson x :: svJsons xs
def svJsonMembers : List (String × Shape.Value) → List (String × Json)
  | [] => []
  | (k, v) :: kvs => (k, svJson v) :: svJsonMembers kvs
end

def svKeys : List String := ["a", "b", "é", "k\"1", "c\n"]
def svLeaves : List Shape.Value := [.null, .bool true, .num "1", .str "a", .str "x\"\\"]

def svValues : List Shape.Value :=
  let one := svKeys.flatMap fun k => svLeaves.map fun v => Shape.Value.obj [(k, v)]
  let two := (svKeys.take 3).flatMap fun k₁ => (svLeaves.take 3).flatMap fun v₁ =>
    (svKeys.take 3).flatMap fun k₂ => (svLeaves.take 3).map fun v₂ => Shape.Value.obj [(k₁, v₁), (k₂, v₂)]
  let arrs := [Shape.Value.arr []] ++ svLeaves.map (fun v => .arr [v]) ++
    svLeaves.flatMap (fun a => svLeaves.map fun b => .arr [a, b]) ++
    [.arr [.num "1", .num "1", .num "1"], .arr [.arr [.num "1"], .arr []], .arr [.arr [], .arr [.num "1"]]]
  let nested := (one.take 10 ++ arrs.take 8).map (fun v => Shape.Value.obj [("a", v)]) ++
    (one.take 10).map (fun v => Shape.Value.arr [v, v]) ++
    [.obj [("a", .obj [("b", .num "1"), ("b", .num "2")])], .arr [.obj [("a", .num "1"), ("a", .null)]],
     .arr [.obj [("a", .num "1")], .obj [("b", .num "1")]], .arr [.obj [("b", .num "1")], .obj [("a", .num "1")]]]
  svLeaves ++ one ++ two ++ arrs ++ nested

def svTextBytes : List String := ["5", "true", "{\"a\":1}", "hello\n", "", "[1,2]", "é"]

def jBytes (b : ByteArray) : Json :=
  .arr (b.toList.map fun x => Lean.Json.num (Lean.JsonNumber.fromNat x.toNat)).toArray

mutual
/-- The first repeated key in the rendered text, in reading order, and the
offset just past it: what a streaming decoder reports. -/
def dupIn : Shape.Value → Nat → Option (Nat × String)
  | .arr xs, pos => dupInElems xs (pos + 1)
  | .obj kvs, pos => dupInMembers kvs [] (pos + 1)
  | _, _ => none
def dupInElems : List Shape.Value → Nat → Option (Nat × String)
  | [], _ => none
  | x :: xs, pos => match dupIn x pos with
    | some r => some r
    | none => dupInElems xs (pos + (svText x).utf8ByteSize + 1)
def dupInMembers : List (String × Shape.Value) → List String → Nat → Option (Nat × String)
  | [], _, _ => none
  | (k, v) :: kvs, seen, pos =>
    let keyEnd := pos + ("\"" ++ Shape.escape k ++ "\"").utf8ByteSize
    if k ∈ seen then some (keyEnd, k)
    else match dupIn v (keyEnd + 1) with
      | some r => some r
      | none => dupInMembers kvs (k :: seen) (keyEnd + 1 + (svText v).utf8ByteSize + 1)
end

/-- What the shell's decoder gives for a rendered value. -/
def svDecoded (v : Shape.Value) : Except (Nat × Shape.Failure) Shape.Value :=
  match dupIn v 0 with
  | some (off, k) => .error (off, .duplicate k)
  | none => .ok v

/-- Bytes that do not decode as declared. Syntax errors carry the decoder's
offset: the bytes it had read when it found the error. -/
def svBad : List (Shape.Output × List UInt8 × Except (Nat × Shape.Failure) Shape.Value) :=
  [(.text, [111, 107, 255, 110, 111], .error (0, .syntax)), (.text, [0xc3], .error (0, .syntax)),
   (.text, [0xe9, 0x41], .error (0, .syntax)), (.text, [0x61, 0xed, 0xa0, 0x80], .error (0, .syntax)),
   (.text, [0xf4, 0x90, 0x80, 0x80], .error (0, .syntax)), (.text, [0xc0, 0x80], .error (0, .syntax)),
   (.json, "{\"a\":".toUTF8.toList, .error (5, .syntax)),
   (.json, "{\"a\":1} x".toUTF8.toList, .error (9, .syntax)),
   (.json, "nul".toUTF8.toList, .error (3, .syntax)),
   (.json, [], .error (0, .syntax)),
   (.json, "{\"a\":1,}".toUTF8.toList, .error (8, .syntax)),
   (.json, [34, 255, 34], .error (0, .syntax))]

def shapeFiles : List (String × List Json) :=
  let case (o : Shape.Output) (raw : ByteArray) (d : Except (Nat × Shape.Failure) Shape.Value) : Json :=
    let v := Shape.arrive o raw d
    Json.mkObj [("in", Json.mkObj [("output", .str (if o = .json then "json" else "text")),
        ("bytes", jBytes raw)]),
      ("out", Json.mkObj [("value", svJson v),
        ("shape", jOpt (fun s => .str (Shape.Shape.enc s)) (Shape.shapeOf v))])]
  [ ("shapeArrive",
      svValues.map (fun v => case .json (svText v).toUTF8 (svDecoded v)) ++
      (svTextBytes ++ (svValues.take 12).map svText ++ ["é ok", String.singleton (Char.ofNat 0x10FFFF)]).map
        (fun b => case .text b.toUTF8 (.error (0, .syntax))) ++
      svBad.map fun (o, b, d) => case o ⟨b.toArray⟩ d) ]

/-! ## Proposal 58: the walk

Each case is a frame's starting scope and a sequence of events, invokes and
returns, applied in turn; a fault leaves the frame as it was. Hashes are
stand-ins, `in(contract;bases)` and `ret(input;shape)`, which the Go replay
supplies too. For each arrival, a template naming every `$n` of the next
scope is stored under the arrival's table and bound back. -/

def wHashes : Walk.Hashes :=
  { input := fun c bs => "in(" ++ c ++ ";" ++ ",".intercalate bs ++ ")"
    ret := fun i s => "ret(" ++ i ++ ";" ++ s ++ ")" }

def wPos : Walk.Position → String
  | .arrived => "arrived"
  | .passed k => "passed" ++ toString k

def wRef : Walk.Position → String
  | .arrived => "@arrived"
  | .passed k => "@passed[" ++ toString k ++ "]"

def wTok : Walk.Tok → String
  | .lit s => s
  | .ref n => "$" ++ toString n

def wStored : Walk.Stored → String
  | .lit s => s
  | .pos p => wRef p

def wObs (o : Walk.Obs) : Json := Json.mkObj [("value", svJson o.value), ("base", .str o.base)]

def wFrame (f : Walk.Frame) : Json :=
  Json.mkObj [("scope", .arr (f.scope.map wObs).toArray),
    ("pending", match f.pending with
      | none => .null
      | some p => Json.mkObj [("input", .str p.input),
          ("output", .str (if p.output = .json then "json" else "text")),
          ("args", .arr (p.args.map fun n => Lean.toJson n).toArray),
          ("passed", .arr (p.passed.map wObs).toArray)])]

/-- The table as each `$n`'s first entry. -/
def wTable (t : List (Nat × Walk.Position)) : Json :=
  Json.mkObj ((t.map (·.1)).eraseDups.filterMap fun n =>
    (Walk.find n t).map fun p => (toString n, Json.str (wPos p)))

def wEnv (env : List (Walk.Position × Shape.Value)) : Json :=
  Json.mkObj (env.map fun (p, v) => (wPos p, svJson v))

/-- `run $1 $2 … $n`, every number of the scope, and `$9` past its end. -/
def wTemplate (len : Nat) (past : Bool) : List Walk.Tok :=
  [.lit "run"] ++ ((List.range len).map (· + 1) ++ (if past then [9] else [])).flatMap
    fun n => [.lit " ", .ref n]

def wEvent : Walk.Event → Json
  | .invoke c o args => Json.mkObj [("invoke", Json.mkObj [("contract", .str c),
      ("output", .str (if o = .json then "json" else "text")),
      ("args", .arr (args.map fun n => Lean.toJson n).toArray)])]
  | .ret raw _ x err => Json.mkObj [("return", jBytes raw), ("exit", jNat x), ("stderr", jBytes err)]

def wFault : Walk.Fault → Json
  | .busy => Json.mkObj [("fault", "busy")]
  | .idle => Json.mkObj [("fault", "idle")]
  | .unbound n => Json.mkObj [("fault", "unbound"), ("n", (n : Json))]

def wArrival (a : Walk.Arrival) : Json :=
  let tmpl (past : Bool) := wTemplate a.next.scope.length past
  let stored (past : Bool) : Json := match Walk.store a.table (tmpl past) with
    | .error n => Json.mkObj [("unstorable", (n : Json))]
    | .ok ys => Json.mkObj [("stored", .str (String.join (ys.map wStored))),
        ("bound", match Walk.bind a.table ys with
          | some xs => .str (String.join (xs.map wTok))
          | none => .null)]
  Json.mkObj [("dir", .str (if a.dir = .forward then "forward" else "back")),
    ("base", .str a.base), ("env", wEnv a.env), ("table", wTable a.table),
    ("next", wFrame a.next),
    ("template", .str (String.join ((tmpl false).map wTok))), ("learn", stored false),
    ("templatePast", .str (String.join ((tmpl true).map wTok))), ("learnPast", stored true)]

def wRun (f : Walk.Frame) : List Walk.Event → List Json
  | [] => []
  | e :: es => match Walk.sense wHashes f e with
    | .error x => wFault x :: wRun f es
    | .ok a => wArrival a :: wRun a.next es

def wRetV (v : Shape.Value) : Walk.Event := .ret (svText v).toUTF8 (svDecoded v) 0 ByteArray.empty
def wRetText (s : String) : Walk.Event := .ret s.toUTF8 (.error (0, .syntax)) 0 ByteArray.empty
/-- A process that exited non-zero. -/
def wFail (out : String) (x : Nat) (err : String) : Walk.Event :=
  .ret out.toUTF8 (.error (0, .syntax)) x err.toUTF8

def wStart : Walk.Frame :=
  { scope := [⟨.str "list the logs", "goal"⟩,
      ⟨.obj [("dir", .str "/var/log"), ("size", .num "5")], "ret(in(find;goal);{\"dir\":string,\"size\":number})"⟩],
    pending := none }

def wScenarios : List (Walk.Frame × List Walk.Event) :=
  let v := Shape.Value.obj [("n", .num "3"), ("files", .arr [.str "a", .str "b"])]
  [ (wStart, [.invoke "ls" .text [2], wRetText "a.log\nb.log"]),
    (wStart, [.invoke "wc" .json [2, 1, 2], wRetV v, .invoke "cat" .text [3, 1], wRetText "ok"]),
    (wStart, [.invoke "ls" .text [3], .invoke "ls" .text [0], .ret "x".toUTF8 (.error (0, .syntax)) 0 ByteArray.empty]),
    (wStart, [.invoke "ls" .text [1], .invoke "rm" .text [1], wRetText "done"]),
    (wStart, [.invoke "jq" .json [], .ret "{\"a\":".toUTF8 (.error (5, .syntax)) 0 ByteArray.empty, .invoke "fix" .text [3]]),
    (wStart, [.invoke "cat" .text [2], .ret ⟨#[111, 107, 255]⟩ (.error (0, .syntax)) 0 ByteArray.empty]),
    (wStart, [.invoke "jq" .json [1], wRetV (.obj [("a", .num "1"), ("a", .num "2")])]),
    (wStart, [.invoke "echo" .text [1], wRetText "5", .invoke "echo" .json [3], wRetV (.num "5"),
      .invoke "add" .json [3, 4, 3], wRetV (.num "10")]),
    ({ scope := [], pending := none }, [.invoke "date" .text [], wRetText "today"]),
    (wStart, [.invoke "rm" .text [1], wFail "" 1 "rm: list the logs: No such file",
      .invoke "touch" .text [1], wRetText ""]),
    (wStart, [.invoke "jq" .json [2], wFail "{\"partial\":" 2 "", .invoke "jq" .json [2], wFail "" 2 "x"]) ]

def walkFiles : List (String × List Json) :=
  [ ("walk", wScenarios.map fun (f, es) =>
      Json.mkObj [("in", Json.mkObj [("frame", wFrame f), ("events", .arr (es.map wEvent).toArray)]),
        ("out", .arr (wRun f es).toArray)]) ]

/-! ## Proposal 58: dispatch

Each case is a refinement, an arrival's `size`, and what the model writes
each time it is asked: guards as the model writes them (`$1` is what
arrived), or text that is not a guard. Lean gives the transition's states,
the refinement after it, what `trust` runs, whether it is learned, and how
many times the model was asked; then the same arrival again, with nothing
more written. -/

def dState : Refine.State → String
  | .reject => "reject" | .hesitate => "hesitate" | .startle => "startle"
  | .reconcile => "reconcile" | .trust => "trust"

def dWrote : Dispatch.Wrote Atom String → Json
  | .guard c => .str ((guardCel c).replace "arrived" "$1")
  | .unlearned c => .str ((guardCel c).replace "arrived" "$1")
  | .refused => .str "not a guard ("

def dOut (o : Dispatch.Out Atom String) : Json :=
  Json.mkObj [("states", .arr (o.states.map (Json.str ∘ dState)).toArray),
    ("refinement", jRef o.refinement),
    ("instr", match o.instr with | some i => .str i | none => .null),
    ("learned", .bool o.learned), ("asks", jNat o.asks), ("ended", .bool o.ended)]

def dWrites : List (List (Dispatch.Wrote Atom String)) :=
  [ [.guard [(.atom (.gt 2), "C")]],
    [.guard [(.atom (.lt 5), "D"), (.tt, "E")]],
    [.refused, .guard [(.atom .even, "C")]],
    [.guard [(.atom (.gt 100), "C")], .guard [(.tt, "F")]],
    [.refused, .refused],
    [.refused, .refused, .refused, .guard [(.tt, "F")]],
    [] ]

/-- The refinement bound: the most asks in one transition. -/
def dBound : Nat := 3

def dispatchFiles : List (String × List Json) :=
  [ ("dispatch", (refs1 ++ refsPossible).flatMap fun r => values.flatMap fun x => dWrites.map fun ws =>
      let o := Dispatch.decide Atom.sat overlaps x dBound r ws
      let again := Dispatch.decide Atom.sat overlaps x dBound o.refinement []
      Json.mkObj [("in", Json.mkObj [("refinement", jRef r), ("x", jInt x), ("refinementBound", jNat dBound),
          ("writes", .arr (ws.map dWrote).toArray)]),
        ("out", Json.mkObj [("first", dOut o), ("again", dOut again)])]) ]

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
    ("classify", agents.flatMap fun self => ([1, 2, 3] : List Nat).flatMap fun act =>
      (upTo (agents ++ [w.app.exit, w.app.goal, w.app.observe]) 3).map fun reach =>
        case (Json.mkObj [("self", jContract self), ("act", Lean.Json.num (Lean.JsonNumber.fromNat act)),
            ("reach", jList jContract reach)])
          (jOpt jKind (classify act w.app self reach))),
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
        (jInvokeResult (invoke n g))) ] ++ refineFiles ++ shapeFiles ++ walkFiles ++ dispatchFiles

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
