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

/-! ## Proposals 58 and 59: refinements

The atoms compare one field, `arrived.size`, with a constant, or test it
for evenness; the values are that field's. Conditions are written as CEL,
the way the Go receives them. -/

inductive Atom where
  | gt (k : Int) | lt (k : Int) | eq (k : Int) | even
  deriving Repr, DecidableEq

def Atom.sat : Atom → Int → Bool
  | .gt k, x => x > k
  | .lt k, x => x < k
  | .eq k, x => x == k
  | .even, x => x % 2 == 0

/-- Written as a stored guard reads it: the arrival's `size` field, which
the model wrote as `$1.size`. -/
def Atom.cel : Atom → String
  | .gt k => s!"arrived.size > {k}"
  | .lt k => s!"arrived.size < {k}"
  | .eq k => s!"arrived.size == {k}"
  | .even => "int(arrived.size) % 2 == 0"

def condCel : Refine.Cond Atom → String
  | .atom a => a.cel
  | .tt => "true"
  | .and p q => "(" ++ condCel p ++ ") && (" ++ condCel q ++ ")"
  | .not p => "!(" ++ condCel p ++ ")"

abbrev RBranch := Refine.Branch Atom String
abbrev RRef := Refine.Refinement Atom String

def jNat (n : Nat) : Json := Lean.Json.num (Lean.JsonNumber.fromNat n)

def jBranch (b : RBranch) : Json :=
  Json.mkObj [("id", jNat b.id), ("rank", jNat b.rank), ("voter", jNat b.voter),
    ("cond", .str (condCel b.cond)), ("instr", .str b.instr)]

def jRef (r : RRef) : Json :=
  Json.mkObj [("next", jNat r.next), ("branches", jList jBranch r.branches)]

def jOutcome : Refine.Outcome String → Json
  | .null => .null
  | .select i => Json.mkObj [("select", .str i)]
  | .conflict => .str "conflict"

def jAfter : Refine.After String → Json
  | .startleAgain => .str "startle"
  | .trust i l => Json.mkObj [("trust", .str i), ("learned", .bool l)]

/-- A stored instruction's parts, as the Go's `Parts` reads them: its form,
then its head (the tool run, a return's exit status, the delegated goal),
then the rest in order. -/
def parts (i : String) : List String :=
  if i.startsWith "delegate(" then
    match (((i.drop 9).dropEnd 1).toString).splitOn ", " with
    | [] => ["delegate"]
    | g :: rest => "delegate" :: (((g.drop 5).dropEnd 1).toString) :: rest
  else match i.splitOn " " with
    | "return" :: rest => "return" :: rest
    | fs => "run" :: fs

/-- The exact level. -/
def exact : Nat := 4

/-- An instruction's key at level `k`: its first `k` parts, or the whole
instruction at the exact level. -/
def group (k : Nat) (i : String) : List String :=
  if exact ≤ k then [i] else (parts i).take k

/-- The levels from the exact one down to the floor. -/
def levels (floor : Nat) : List Nat := (List.range (exact + 1 - floor)).map (exact - ·)

def policy (inhibition floor : Nat) : Refine.Policy String (List String) :=
  ⟨group, levels floor, inhibition⟩

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

def commitR (r : RRef) (chain : List (Refine.Cond Atom × String)) : RRef := Refine.commit r chain

/-- The empty refinement and every one a single one-link guard makes. -/
def refs1 : List RRef := Refine.Refinement.empty :: chains1.map (commitR .empty)

/-- Every refinement two one-link guards make, one after the other. -/
def refs2 : List RRef := chains1.flatMap fun a => chains1.map fun b => commitR (commitR .empty a) b

/-- Refinements of two voters that disagree: different instructions. -/
def refsSplit : List RRef :=
  (chains1.flatMap fun a => chains1.filterMap fun b =>
    if a.map (·.2) = b.map (·.2) then none else some (commitR (commitR .empty a) b)).take 30

def values : List Int := [0, 2, 3, 5, 6, 8, 9]

def jInt (x : Int) : Json := Lean.Json.num (Lean.JsonNumber.fromInt x)

/-- Instructions as they are stored: canonical, each `$n` a position. -/
def tInstrs : List String :=
  ["return 1 @passed[4] @passed[5]", "return 1 @passed[4]", "@passed[1] @passed[5].url",
    "return 0 @passed[4]", "delegate(goal(@passed[0]), @passed[1])"]

/-- Three voters, each one link from a pool. -/
def tallyRefs : List RRef :=
  let pool := ([Refine.Cond.tt, .atom (.gt 2)].flatMap fun c => (tInstrs.take 4).map fun i => [(c, i)])
  pool.flatMap fun a => pool.flatMap fun b => pool.map fun c => commitR (commitR (commitR .empty a) b) c

/-- The `cel/` experiment's ballots (proposal 59): seven voters, each one
link that holds. -/
def experiment : List RRef :=
  let r := fun (is : List String) => is.foldl (fun r i => commitR r [(.tt, i)]) Refine.Refinement.empty
  let a := "return 1 @passed[4] @passed[5]"
  let b := "return 1 @passed[4]"
  let c := "@passed[1] @passed[5].url"
  [ r (List.replicate 7 "return 0 @passed[4]"),
    r (List.replicate 6 a ++ [b]),
    r (List.replicate 4 a ++ [c, c, b]) ]

def tallyCase (r : RRef) (x : Int) (ι floor : Nat) : Json :=
  Json.mkObj [("in", Json.mkObj [("refinement", jRef r), ("x", jInt x), ("inhibition", jNat ι),
      ("floor", jNat floor)]),
    ("out", jOutcome (Refine.tally (policy ι floor) Atom.sat r.branches x))]

/-- The policy dispatch and `afterStartle` vectors use. -/
def dInhibition : Nat := 0
def dFloor : Nat := 2

def refineFiles : List (String × List Json) :=
  let case (i o : Json) : Json := Json.mkObj [("in", i), ("out", o)]
  [ ("refineCommit", refs1.flatMap fun r => (chains1 ++ chains2).map fun chain =>
      case (Json.mkObj [("refinement", jRef r), ("guard", .str (guardCel chain))])
        (jRef (Refine.commit r chain))),
    ("refineDistinguish", refs2.flatMap fun r => values.map fun v =>
      case (Json.mkObj [("refinement", jRef r), ("x", jInt v)])
        (jOutcome (Refine.distinguish Atom.sat r.branches v))),
    ("refineTally",
      (tallyRefs.flatMap fun r => [0, 3].flatMap fun x => [0, 500, 750, 1000].map fun ι =>
        tallyCase r x ι 2) ++
      (tallyRefs.take 64).flatMap (fun r => [1, 3, 4].map fun f => tallyCase r 3 500 f) ++
      (refsSplit.flatMap fun r => [3, 6].map fun x => tallyCase r x 1000 2) ++
      experiment.flatMap fun r => [0, 250, 500, 750, 1000].flatMap fun ι => [2, 4].map fun f =>
        tallyCase r 0 ι f),
    ("refineAfterStartle", (chains1 ++ chains2.take 10).flatMap fun chain =>
      (refs1 ++ refsSplit.take 6).flatMap fun r => [3, 6, 9].map fun v =>
        case (Json.mkObj [("guard", .str (guardCel chain)), ("committed", jRef r), ("x", jInt v),
            ("inhibition", jNat dInhibition), ("floor", jNat dFloor)])
          (jAfter (Refine.afterStartle (policy dInhibition dFloor) Atom.sat chain r.branches v))) ]

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
  | .exit _ => ""
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
  | .exit n => Json.mkObj [("exitStatus", Lean.Json.num (Lean.JsonNumber.fromNat n))]
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
      svBad.map fun (o, b, d) => case o ⟨b.toArray⟩ d),
    ("returnValue",
      [0, 1].flatMap fun x => ["", "ok\n", "\xff"].flatMap fun out =>
        [none, some "", some "why"].map fun err =>
          let raw (s : String) : ByteArray := if s = "\xff" then ⟨#[0xff]⟩ else s.toUTF8
          let v := Shape.returnValue x (raw out) (err.map raw)
          Json.mkObj [("in", Json.mkObj [("exit", Lean.Json.num (Lean.JsonNumber.fromNat x)),
              ("stdout", jBytes (raw out)), ("stderr", jOpt (fun e => jBytes (raw e)) err)]),
            ("out", Json.mkObj [("value", svJson v),
              ("shape", jOpt (fun s => .str (Shape.Shape.enc s)) (Shape.shapeOf v))])]) ]

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

def wObs (o : Walk.Obs) : Json := Json.mkObj [("value", svJson o.value), ("base", .str o.base),
  ("tool", match o.tool with | some b => .str b | none => .null), ("goal", .bool o.goal)]

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
  | .returned v _ => Json.mkObj [("returned", svJson v)]

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
  { scope := [{ value := .str "list the logs", base := "goal" },
      { value := .obj [("dir", .str "/var/log"), ("size", .num "5")], base := "ret(in(find;goal);{\"dir\":string,\"size\":number})" }],
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

/-! ## Proposal 58: frames

Each case is a scope, tools among it, and an instruction by its first
position; Lean classifies it: a return, a run of a held binary, a child
frame narrowing the tools, or a rejection with its reason. -/

def fTool (b : String) : Walk.Obs := { value := .str b, base := "tool(" ++ b ++ ")", tool := some b }

def fScope : List Walk.Obs :=
  [{ value := .str "/var/log", base := "goal", goal := true }, fTool "ls", fTool "cat", fTool "ls",
   { value := .obj [("n", .num "2"), ("dir", .obj [("path", .str "/tmp")])], base := "ret(in(wc;goal);{\"dir\":{\"path\":string},\"n\":number})" },
   { value := .str "a.log\nb.log\nnotes.txt\n", base := "ret(in(ls;goal);string)" },
   { value := .bytes "one\ntwo".toUTF8, base := "ret(in(cat;goal);bytes)" },
   { value := .bytes (ByteArray.mk #[0xff, 0x0a]), base := "ret(in(cat;goal);bytes)" }]

def fBare : List Walk.Obs := [{ value := .str "/var/log", base := "goal", goal := true }]

def fA (n : Nat) (sel : Frames.Sel := .whole) : Frames.Arg := { n, sel }

def fLines (rs : List (Nat × Nat)) : Frames.Sel := .lines rs

def fInstrs : List Frames.Instr :=
  [ .ret (fA 1), .ret (fA 9), .run 2 [fA 1], .run 3 [fA 1, fA 5], .run 1 [fA 2], .run 9 [], .run 2 [fA 9],
    .delegate (fA 1) [fA 2], .delegate (fA 1) [fA 2, fA 3],
    .delegate (fA 1 (fLines [(1, 1)])) [fA 2, fA 4, fA 6], .delegate (fA 6) [fA 5, fA 2],
    .delegate (fA 2) [fA 2], .delegate (fA 2 (fLines [(1, 1)])) [fA 2], .delegate (fA 9) [fA 2],
    .delegate (fA 1) [], .delegate (fA 1) [fA 1], .delegate (fA 1) [fA 6, fA 2],
    .delegate (fA 5 (.field ["dir", "path"])) [fA 2, fA 5 (.field ["dir", "path"])],
    .delegate (fA 5 (.field ["x"])) [fA 2],
    -- cuts
    .ret (fA 5 (.field ["n"])), .ret (fA 5 (.field ["dir", "path"])), .ret (fA 5 (.field ["missing"])),
    .ret (fA 1 (.field ["n"])), .ret (fA 6 (fLines [(2, 3)])), .ret (fA 6 (fLines [(1, 1), (3, 3)])),
    .ret (fA 6 (fLines [(3, 4)])), .ret (fA 6 (fLines [(0, 1)])), .ret (fA 6 (fLines [(2, 1)])),
    .ret (fA 6 (fLines [])), .ret (fA 7 (fLines [(2, 2)])), .ret (fA 8 (fLines [(1, 1)])),
    .ret (fA 5 (fLines [(1, 1)])), .ret (fA 1 (fLines [(1, 1)])),
    .run 2 [fA 5 (.field ["dir", "path"])], .run 2 [fA 6 (fLines [(9, 9)])],
    .delegate (fA 1) [fA 2 (fLines [(1, 1)]), fA 3 (fLines [(1, 1)]), fA 6 (fLines [(1, 2)])],
    .delegate (fA 1) [fA 2, fA 3 (.field ["x"])] ]

def fArg (a : Frames.Arg) : Json := Json.mkObj [("n", jNat a.n), ("sel", .str a.sel.enc)]

def fInstr : Frames.Instr → Json
  | .ret a => Json.mkObj [("return", fArg a)]
  | .run t args => Json.mkObj [("run", jNat t), ("args", .arr (args.map fArg).toArray)]
  | .delegate g args => Json.mkObj [("delegate", fArg g), ("args", .arr (args.map fArg).toArray)]

def fCutBase (b s : String) : String := "cut(" ++ b ++ ";" ++ s ++ ")"

def fReason : Frames.Reason → Json
  | .unbound n => Json.mkObj [("unbound", jNat n)]
  | .cut n => Json.mkObj [("cut", jNat n)]
  | .noTools => .str "noTools"
  | .actDelegates => .str "actDelegates"
  | .delegateRuns => .str "delegateRuns"
  | .goalTool n => Json.mkObj [("goalTool", jNat n)]
  | .notTool n => Json.mkObj [("notTool", jNat n)]
  | .narrows p h => Json.mkObj [("narrows", .arr #[jNat p, jNat h])]

def fKind : Frames.Kind → Json
  | .ret o => Json.mkObj [("return", wObs o)]
  | .run b ps => Json.mkObj [("run", .str b), ("passed", .arr (ps.map wObs).toArray)]
  | .child ps => Json.mkObj [("child", .arr (ps.map wObs).toArray),
      ("table", wTable (Frames.childTable ps)), ("tools", .arr ((Frames.tools ps).map Json.str).toArray)]
  | .reject r => Json.mkObj [("reject", fReason r)]

/-- An observation as a case's input: as `wObs`, bytes marked, since JSON
has no bytes and an array of numbers is not one. -/
def fInObs (o : Walk.Obs) : Json :=
  match o.value with
  | .bytes _ => (wObs o).setObjVal! "bytes" (.bool true)
  | _ => wObs o

def framesFiles : List (String × List Json) :=
  [ ("frames", [1, 2].flatMap fun k => [fScope, fBare].flatMap fun sc => fInstrs.map fun i =>
      Json.mkObj [("in", Json.mkObj [("act", jNat k), ("scope", .arr (sc.map fInObs).toArray), ("instr", fInstr i)]),
        ("out", fKind (Frames.classify fCutBase k sc i))]) ]

/-! ## Proposal 58: dispatch

Each case is a refinement, an arrival's `size`, and what the model writes
each time it is asked: guards as the model writes them (`$1` is what
arrived), or text that is not a guard. Lean gives the transition's states,
the refinement after it, what `trust` runs, whether it is learned, and how
many times the model was asked; then the same arrival again, with nothing
more written. -/

def dState : Refine.State → String
  | .reject => "reject" | .hesitate => "hesitate" | .startle => "startle"
  | .tally => "tally" | .trust => "trust"

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
  [ ("dispatch", (refs1 ++ refsSplit).flatMap fun r => values.flatMap fun x => dWrites.map fun ws =>
      let p := policy dInhibition dFloor
      let o := Dispatch.decide p Atom.sat x dBound r ws
      let again := Dispatch.decide p Atom.sat x dBound o.refinement []
      Json.mkObj [("in", Json.mkObj [("refinement", jRef r), ("x", jInt x), ("refinementBound", jNat dBound),
          ("inhibition", jNat dInhibition), ("floor", jNat dFloor),
          ("writes", .arr (ws.map dWrote).toArray)]),
        ("out", Json.mkObj [("first", dOut o), ("again", dOut again)])]) ]


/-! ## Proposal 61: definition and instantiation

The model's answers, written under `vectors/model/`: what boot gives, a
frame's key, the kind an operator pushes, a cut and its reduced interface,
reading and inlining, the copy check, and the tree's edges. Hashes are
spelled out as text, as 58's cut bases are, so a replay can form them. -/

namespace M

def jSel (s : Model.Sel) : Json := .str s.toCut.enc

/-- A value as 58's vectors write it, but bytes tagged, `{"bytes": […]}`,
so they are never read back as an array of numbers. -/
partial def jValue : Shape.Value → Json
  | .bytes b => Json.mkObj [("bytes", .arr (b.toList.map fun x => jNat x.toNat).toArray)]
  | .arr xs => .arr (xs.map jValue).toArray
  | .obj kvs => Json.mkObj (kvs.map fun (k, v) => (k, jValue v))
  | v => svJson v

def jOp : Model.Op → Json
  | .delegate => .str "delegate" | .act => .str "act" | .boot => .str "boot" | .ret => .str "return"

def jKind : Model.Kind → Json
  | .root => .str "root" | .delegate => .str "delegate" | .act => .str "act" | .leaf => .str "leaf"

partial def jIface : Model.Iface → Json
  | .shape s => Json.mkObj [("shape", .str s.enc)]
  | .defn b => Json.mkObj [("def", .str b.hash)]
  | .op o ps => Json.mkObj [("op", jOp o), ("parts", .arr (ps.map jIface).toArray)]
  | .sum as => Json.mkObj [("sum", .arr (as.map jIface).toArray)]
  | .cut i s => Json.mkObj [("cut", jIface i), ("sel", jSel s)]

def jScope (s : Model.Scope) : Json :=
  Json.mkObj [("kind", jKind s.kind), ("entries", .arr (s.entries.map jIface).toArray)]

def jLit : Model.Lit → Json
  | .str s => Json.mkObj [("str", .str s)]
  | .exit n => Json.mkObj [("exit", jNat n)]

def jObs (o : Model.Observation) : Json :=
  Json.mkObj [("value", jValue o.value), ("iface", jIface o.iface), ("loc", .str o.loc.hash)]

def jMember : Model.Member → Json
  | .tool b => Json.mkObj [("tool", .str b.hash)]
  | .obs o => Json.mkObj [("obs", jObs o)]

def jSels (ss : List Model.Sel) : Json := .arr (ss.map jSel).toArray

def jPart : Model.Part → Json
  | .str s => Json.mkObj [("str", .str s)]
  | .seen n ss => Json.mkObj [("seen", jNat n), ("sels", jSels ss)]

def jEntry : Model.Entry → Json
  | .arrived o => Json.mkObj [("arrived", jObs o)]
  | .eliminated n ss => Json.mkObj [("eliminate", jNat n), ("sels", jSels ss)]
  | .introduced l => Json.mkObj [("introduce", jLit l)]
  | .joined ps => Json.mkObj [("join", jList jPart ps)]

def jFrame (f : Model.Frame) : Json :=
  Json.mkObj [("id", .str f.id.hash), ("program", .str f.program.hash),
    ("parent", jOpt (fun l => .str l.hash) f.parent), ("scope", jScope f.scope),
    ("bindings", .arr (f.bindings.map jMember).toArray), ("state", .arr (f.state.map jEntry).toArray),
    ("model", jOpt (fun l => .str l.hash) f.model)]

def jArg : Model.Arg → Json
  | .ref (.var n ss) => Json.mkObj [("var", jNat n), ("sels", jSels ss)]
  | .ref (.name s) => Json.mkObj [("name", .str s)]
  | .lit l => Json.mkObj [("lit", jLit l)]
  | .join ps => Json.mkObj [("join", jList jPart ps)]

def ifaceText (i : Model.Iface) : String := (jIface i).compress

/-- Hashes as text: a child by its parent, invoke index and scope; a first
frame by its program and grant's definitions; a slot by its frame and `$n`. -/
def ids : Model.Ids where
  child p n s := ⟨"child(" ++ p.hash ++ "," ++ toString n ++ "," ++ (jScope s).compress ++ ")"⟩
  first d g := ⟨"first(" ++ d.hash ++ ";" ++ ",".intercalate (g.defs.map (·.hash)) ++ ")"⟩
  slot i n := ⟨i.hash ++ "#" ++ toString n⟩
  program p := ⟨"program(" ++ p.version.hash ++ ")"⟩

def ls : Model.DefLoc := ⟨"ls"⟩
def cat : Model.DefLoc := ⟨"cat"⟩
def claude : Model.DefLoc := ⟨"claude"⟩
def gpt : Model.DefLoc := ⟨"gpt"⟩

def consts : Model.Constants := ⟨1, 64, 3, (1, 2), 2, 1⟩

def global (tools models : List Model.DefLoc) : List (String × Model.Iface) :=
  [("act", .op .act [.sum (tools.map .defn)]), ("delegate", .op .delegate [.sum (models.map .defn)])]

/-- The program: `ls` and `cat` its tools, `claude` and `gpt` its models. -/
def prog : Model.Program :=
  ⟨⟨"v1"⟩, consts, global [ls, cat] [claude, gpt], ⟨.root, [.defn ls, .defn cat, .defn claude, .shape .str]⟩⟩

/-- Refused: `cat` is both a tool and a model. -/
def progOverlap : Model.Program := { prog with version := ⟨"v1-overlap"⟩, global := global [ls, cat] [claude, cat] }

def jProgram (p : Model.Program) : Json :=
  Json.mkObj [("version", .str p.version.hash),
    ("constants", Json.mkObj [("act", jNat p.constants.act), ("transitions", jNat p.constants.transitions),
      ("refinements", jNat p.constants.refinements),
      ("inhibition", .arr #[jNat p.constants.inhibition.1, jNat p.constants.inhibition.2]),
      ("floor", jNat p.constants.floor), ("quorum", jNat p.constants.quorum)]),
    ("global", Json.mkObj (p.global.map fun (n, i) => (n, jIface i))), ("entry", jScope p.entry)]

def booter (defs : List Model.DefLoc) : Model.Frame :=
  ⟨⟨"user"⟩, ⟨"program(user)"⟩, none, ⟨.root, defs.map .defn⟩, defs.map .tool, [], none⟩

def bootFile : List Json :=
  let grantDefs : List (List Model.DefLoc) := [[ls, cat, claude], [cat, ls, claude], [ls, cat], [ls, cat, gpt]]
  let contexts : List (List Shape.Value) := [[.str "/tmp/x"], [.num "3"], [], [.str "a", .str "b"]]
  let srcs : List (Option (Model.Frame × Nat)) := [none, some (booter [ls, cat, claude, gpt], 0), some (booter [ls], 2)]
  [prog, progOverlap].flatMap fun p => grantDefs.flatMap fun ds => contexts.flatMap fun cx => srcs.map fun src =>
    Json.mkObj [("in", Json.mkObj [("program", jProgram p),
        ("grant", Json.mkObj [("defs", .arr (ds.map (Json.str ·.hash)).toArray), ("context", .arr (cx.map jValue).toArray)]),
        ("from", jOpt (fun (f, n) => Json.mkObj [("frame", jFrame f), ("invoke", jNat n)]) src)]),
      ("out", jOpt jFrame (Model.boot ids p ⟨ds, cx⟩ src))]

def obs (v : Shape.Value) (n : Nat) : Model.Observation :=
  ⟨v, match Shape.shapeOf v with | some s => .shape s | none => .shape .null, ⟨"f#" ++ toString n⟩⟩

def vDir : Shape.Value := .obj [("dir", .obj [("path", .str "/tmp")]), ("n", .num "1")]
def vText : Shape.Value := .str "a\nb\nc"

/-- A frame: `ls` and a goal bound; an object arrived; cuts, an introduction
copying nothing, a forward elimination, an exit status, text arrived, and
an elimination of itself; a join of a cut and text. -/
def frame : Model.Frame :=
  ⟨⟨"f"⟩, ⟨"program(v1)"⟩, none, ⟨.act, [.defn ls, .shape .str]⟩,
    [.tool ls, .obs (obs (Shape.Value.str "list /tmp") 2)],
    [.arrived (obs vDir 3), .eliminated 3 [.field ["dir"]], .introduced (.str "-la"),
      .eliminated 4 [.field ["path"]], .eliminated 9 [], .introduced (.exit 0), .arrived (obs vText 9),
      .eliminated 10 [.lines [(1, 1)]], .joined [.seen 3 [.field ["dir", "path"]], .str "/x"]],
    none⟩

def keyFile : List Json :=
  (List.range (frame.state.length + 1)).map fun n =>
    let f := { frame with state := frame.state.take n }
    Json.mkObj [("in", Json.mkObj [("frame", jFrame f)]),
      ("out", let k := Model.key (fun d => d) f
        Json.mkObj [("root", .str k.root.hash), ("kind", jKind k.kind), ("seen", .arr (k.seen.map jIface).toArray)])]

def kindFile : List Json :=
  let atoms : List Model.Iface := [.defn ls, .defn cat, .defn claude, .shape .str]
  ([1, 2] : List Nat).flatMap fun k => [Model.Op.delegate, .act, .boot, .ret].flatMap fun o =>
    (upTo atoms 3).map fun es =>
      Json.mkObj [("in", Json.mkObj [("act", jNat k), ("tools", .arr #[.str "ls", .str "cat"]),
          ("op", jOp o), ("entries", .arr (es.map jIface).toArray)]),
        ("out", Json.mkObj [("kind", jOpt jKind (Model.pushed k [ls, cat] o es)),
          ("toolTypes", .arr ((Model.toolTypes [ls, cat] es).map (Json.str ·.hash)).toArray)])]

def cutFile : List Json :=
  let vals : List Shape.Value := [vDir, vText, .bytes "x\ny".toUTF8, .num "2"]
  let sels : List (List Model.Sel) :=
    [[], [.field ["dir"]], [.field ["dir", "path"]], [.field ["dir"], .field ["path"]], [.field ["nope"]],
      [.field []], [.lines [(2, 3)]], [.lines [(1, 1), (3, 3)]], [.lines [(3, 4)]], [.lines []],
      [.field ["dir", "path"], .lines [(1, 1)]]]
  vals.flatMap fun v => sels.map fun ss =>
    let i : Option Model.Iface := (Shape.shapeOf v).map .shape
    Json.mkObj [("in", Json.mkObj [("value", jValue v), ("sels", jSels ss)]),
      ("out", Json.mkObj [("value", jOpt jValue (Model.cutAll v ss)),
        ("iface", jOpt jIface (i.map (ss.foldl .cut))),
        ("reduced", jOpt jIface (i.bind fun i => (ss.foldl .cut i).reduce))])]

def readFile : List Json :=
  let sels : List (List Model.Sel) := [[], [.field ["path"]], [.lines [(1, 1)]]]
  (List.range 12).flatMap fun n => sels.map fun ss =>
    Json.mkObj [("in", Json.mkObj [("frame", jFrame frame), ("n", jNat n), ("sels", jSels ss)]),
      ("out", Json.mkObj [("read", jOpt jValue (Model.read frame n ss)), ("inline", jArg (Model.inline frame n ss))])]

def introduceFile : List Json :=
  let lits : List Model.Lit := [.str "-la", .str "/tmp", .str "list", .str "path", .str "b\nc", .str "zz", .str "",
    .exit 0, .exit 2]
  lits.map fun l =>
    Json.mkObj [("in", Json.mkObj [("frame", jFrame frame), ("lit", jLit l)]),
      ("out", match Model.introduce frame l with
        | .ok e => Json.mkObj [("entry", jEntry e)]
        | .error n => Json.mkObj [("copies", jNat n)])]

def jEdge : Model.Edge → Json
  | .invoke o s b m => Json.mkObj [("invoke", jOp o), ("scope", jScope s),
      ("bindings", .arr (b.map jMember).toArray), ("model", jOpt (fun l => .str l.hash) m)]
  | .boot f => Json.mkObj [("boot", jFrame f)]
  | .ret o => Json.mkObj [("return", jObs o)]

def jTree (t : Model.Tree) : Json :=
  Json.mkObj [("frames", .arr (t.frames.map jFrame).toArray), ("closed", .arr (t.closed.map (Json.str ·.hash)).toArray)]

def reaches (d : Model.DefLoc) (name : String) : List Model.DefLoc :=
  if d = ids.program prog then prog.reach name else []

/-- From genesis: the root delegates an act frame consulting `claude`; the
act frame runs `ls`, whose leaf returns; the act frame returns. Refused on
the way: deciding while pending, an act delegating, a delegation naming
no model or a tool, a narrowing that does not narrow, and returning from
the root. -/
def stepFile : List Json :=
  match Model.boot ids prog ⟨[ls, cat, claude], [.str "/tmp/x"]⟩ none with
  | none => []
  | some r =>
    let t₀ : Model.Tree := ⟨[r], []⟩
    let actScope : Model.Scope := ⟨.act, [.shape .str, .defn ls]⟩
    let a := ids.child r.id 0 actScope
    let leafScope : Model.Scope := ⟨.leaf, [.defn ls, .shape .str]⟩
    let l := ids.child a 0 leafScope
    let out : Model.Observation := obs (Shape.Value.str "x\ny") 1
    let edges : List (Model.InstLoc × Model.Edge) :=
      [ (r.id, .invoke .delegate ⟨.delegate, [.shape .str, .defn ls, .defn cat]⟩ [] (some claude)),
        (r.id, .invoke .delegate actScope [] none),
        (r.id, .invoke .delegate actScope [] (some ls)),
        (r.id, .invoke .delegate actScope [.tool ls] (some claude)),
        (r.id, .ret out),
        (a, .invoke .delegate ⟨.act, [.shape .str, .defn ls]⟩ [] (some claude)),
        (a, .invoke .act leafScope [.tool ls] none),
        (a, .ret out),
        (l, .ret out),
        (l, .ret out),
        (a, .ret out),
        (r.id, .ret out) ]
    let reach := Json.mkObj [((ids.program prog).hash, Json.mkObj [
      ("act", .arr ((prog.reach "act").map (Json.str ·.hash)).toArray),
      ("delegate", .arr ((prog.reach "delegate").map (Json.str ·.hash)).toArray)])]
    let rec go (t : Model.Tree) : List (Model.InstLoc × Model.Edge) → List Json
      | [] => []
      | (fid, e) :: es => match t.step ids 1 reaches fid e with
        | some t' => Json.mkObj [("in", Json.mkObj [("act", jNat 1), ("reaches", reach), ("tree", jTree t),
            ("frame", .str fid.hash), ("edge", jEdge e)]),
            ("out", jTree t')] :: go t' es
        | none => Json.mkObj [("in", Json.mkObj [("act", jNat 1), ("reaches", reach), ("tree", jTree t),
            ("frame", .str fid.hash), ("edge", jEdge e)]),
            ("out", .null)] :: go t es
    go t₀ edges

def jKey (k : Model.Key) : Json :=
  Json.mkObj [("root", .str k.root.hash), ("kind", jKind k.kind), ("seen", .arr (k.seen.map jIface).toArray)]

def jTemplate (t : Model.Template) : Json :=
  Json.mkObj [("op", jOp t.op), ("args", .arr (t.args.map jArg).toArray)]

def jTBranch (b : Refine.Branch Atom Model.Template) : Json :=
  Json.mkObj [("id", jNat b.id), ("rank", jNat b.rank), ("voter", jNat b.voter),
    ("cond", .str (condCel b.cond)), ("instr", jTemplate b.instr)]

def jTRef (r : Refine.Refinement Atom Model.Template) : Json :=
  Json.mkObj [("next", jNat r.next), ("branches", jList jTBranch r.branches)]

def jEvent : Model.Event Atom → Json
  | .hesitate k => Json.mkObj [("hesitate", jKey k)]
  | .startle k i => Json.mkObj [("startle", jKey k), ("impl", .arr (i.chain.map fun (c, t) =>
      Json.mkObj [("cond", .str (condCel c)), ("template", jTemplate t)]).toArray)]

def jTOutcome : Model.Outcome → Json
  | .null => .null
  | .select t => Json.mkObj [("select", jTemplate t)]
  | .conflict => .str "conflict"

/-- A template's group at level `k`, as 59 groups an instruction: its first
`k` parts, its operator then its arguments, or the whole template at the
exact level. Floor 2 never merges two operators, or two heads. -/
def tGroup (k : Nat) (t : Model.Template) : List String :=
  if exact ≤ k then [(jTemplate t).compress] else ((jOp t.op :: t.args.map jArg).take k).map (·.compress)

def tPolicy (ι floor : Nat) : Refine.Policy Model.Template (List String) := ⟨tGroup, levels floor, ι⟩

def k1 : Model.Key := ⟨⟨"user"⟩, .act, [.shape .str, .defn ls]⟩
def k2 : Model.Key := ⟨⟨"user"⟩, .act, [.shape .str, .defn ls, .shape .str]⟩

def var (n : Nat) (ss : List Model.Sel := []) : Model.Arg := .ref (.var n ss)

/-- Templates: running `ls` on a cut, or on an introduction; returning an
exit status, or what arrived; delegating to `claude`. -/
def t1 : Model.Template := ⟨.act, [var 2, var 3 [.field ["path"]]]⟩
def t2 : Model.Template := ⟨.act, [var 2, .lit (.str "-la")]⟩
def t3 : Model.Template := ⟨.ret, [.lit (.exit 0)]⟩
def t4 : Model.Template := ⟨.ret, [var 3]⟩
def t5 : Model.Template := ⟨.delegate, [.ref (.name "claude"), var 1]⟩

def impl (chain : List (Refine.Cond Atom × Model.Template)) : Model.Impl Atom := ⟨chain⟩

/-- Every log of up to three events from a pool at two keys: hesitations,
and startles of one or two links. Each key's refinement is the fold of its
events alone. -/
def memoryFile : List Json :=
  let pool : List (Model.Event Atom) :=
    [.hesitate k1, .hesitate k2, .startle k1 (impl [(.tt, t1)]),
      .startle k1 (impl [(.atom (.gt 2), t2), (.tt, t3)]), .startle k2 (impl [(.atom .even, t4)])]
  let logs := [[]] ++ pool.map ([·]) ++ (pool.flatMap fun a => pool.map fun b => [a, b]) ++
    (pool.flatMap fun a => pool.flatMap fun b => pool.map fun c => [a, b, c])
  logs.map fun log =>
    Json.mkObj [("in", Json.mkObj [("log", .arr (log.map jEvent).toArray), ("keys", .arr #[jKey k1, jKey k2])]),
      ("out", .arr (([k1, k2] : List Model.Key).map fun k => jOpt jTRef (Model.replay log k)).toArray)]

/-- Decisions at a key one, two or three voters committed to, or one
hesitated at, or none: each a tally over templates, under quorums below,
at and above the voters there. At the other key, unseen, `null`. -/
def decideFile : List Json :=
  let pool : List (Model.Impl Atom) :=
    [impl [(.tt, t1)], impl [(.tt, t2)], impl [(.atom (.gt 2), t1)], impl [(.tt, t3)], impl [(.tt, t4)],
      impl [(.atom .even, t5)]]
  let logs : List (List (Model.Event Atom)) := [[], [.hesitate k1]] ++
    pool.map (fun a => [.startle k1 a]) ++
    (pool.flatMap fun a => pool.map fun b => [a, b].map (.startle k1)) ++
    (pool.flatMap fun a => pool.flatMap fun b => pool.map fun c => [a, b, c].map (.startle k1))
  logs.flatMap fun log => ([k1, k2] : List Model.Key).flatMap fun k =>
    (if k = k2 then [(0, 500, 2, 1)] else
      ([0, 3] : List Int).flatMap fun x => ([0, 500, 1000] : List Nat).flatMap fun ι =>
        ([1, 2] : List Nat).flatMap fun f => ([1, 2, 3] : List Nat).map fun q => (x, ι, f, q)).map fun (x, ι, f, q) =>
      Json.mkObj [("in", Json.mkObj [("log", .arr (log.map jEvent).toArray), ("key", jKey k), ("x", jInt x),
          ("inhibition", jNat ι), ("floor", jNat f), ("quorum", jNat q)]),
        ("out", jTOutcome (Model.decideAt (tPolicy ι f) q Atom.sat (Model.replay log) k x))]

/-- Each state of the test frame, and the `$n` it has seen: what the key
holds, never an elimination or an introduction. -/
def positionsFile : List Json :=
  (List.range (frame.state.length + 1)).map fun k =>
    let f := { frame with state := frame.state.take k }
    Json.mkObj [("in", Json.mkObj [("frame", jFrame f)]),
      ("out", Json.mkObj [("seenAt", .arr ((Model.seenAt f).map jNat).toArray)])]

/-- Templates over the test frame's every `$n` and selector, and literals:
each stored by key position, bound back, and bound again after more
eliminations and introductions. -/
def storeFile : List Json :=
  let sels : List (List Model.Sel) := [[], [.field ["path"]], [.lines [(1, 1)]]]
  let args : List Model.Arg := ((List.range 13).flatMap fun n => sels.map fun ss => var n ss) ++
    [.lit (.str "-la"), .lit (.exit 0), .ref (.name "claude"),
      .join [.seen 3 [.field ["dir"]], .str "/"], .join [.seen 6 []], .join []]
  let live := frame.append [.introduced (.str "z"), .eliminated 3 [.field ["dir"]]]
  args.map fun a =>
    let t : Model.Template := ⟨.act, [var 1, a]⟩
    let st := t.store frame
    Json.mkObj [("in", Json.mkObj [("frame", jFrame frame), ("template", jTemplate t), ("live", jFrame live)]),
      ("out", Json.mkObj [("stored", jOpt jTemplate st), ("bound", jOpt jTemplate (st.bind fun st => st.bind frame)),
        ("boundLive", jOpt jTemplate (st.bind fun st => st.bind live))])]

/-- Joins over the test frame: cuts, an elimination and a join inlined,
text written and refused, and parts that read no text. -/
def joinFile : List Json :=
  let cases : List (List Model.Arg) :=
    [ [var 3 [.field ["dir", "path"]], .lit (.str "/"), var 9 [.lines [(1, 1)]]],
      [var 6, .lit (.str "/"), var 9 [.lines [(2, 3)]]],
      [var 11, .lit (.str ".txt")], [var 5, .lit (.str "x")], [],
      [.lit (.str "/tmp")], [.lit (.str "list /tmp")], [.lit (.str "b\nc")], [.lit (.str "a\nb\nc")],
      [.lit (.str "/")], [.lit (.str "/tm")], [.lit (.str "a\nb\n")],
      [var 3], [var 8], [.lit (.exit 0)], [var 1], [var 7], [var 10], [var 12],
      [.ref (.name "claude")], [.join [.str "x"]] ]
  cases.map fun as =>
    Json.mkObj [("in", Json.mkObj [("frame", jFrame frame), ("args", jList jArg as)]),
      ("out", match Model.joinParts frame as with
        | .ok ps => Json.mkObj [("parts", jList jPart ps), ("text", jOpt .str (Model.joinText frame ps))]
        | .error (.copies n) => Json.mkObj [("copies", jNat n)]
        | .error (.untext a) => Json.mkObj [("untext", jArg a)])]

def files : List (String × List Json) :=
  [ ("model/positions", positionsFile), ("model/store", storeFile), ("model/join", joinFile), ("model/memory", memoryFile), ("model/decide", decideFile), ("model/boot", bootFile), ("model/key", keyFile), ("model/kind", kindFile), ("model/cut", cutFile),
    ("model/read", readFile), ("model/introduce", introduceFile), ("model/step", stepFile) ]

end M

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
        (jInvokeResult (invoke n g))) ] ++ refineFiles ++ shapeFiles ++ walkFiles ++ dispatchFiles ++ framesFiles ++ M.files

/-- One case per line, so a change shows as a line in a diff. -/
def render (name : String) (cases : List Json) : String :=
  "{\"function\":\"" ++ name ++ "\",\"cases\":[\n" ++
    ",\n".intercalate (cases.map (·.compress)) ++ "\n]}\n"

def main (args : List String) : IO UInt32 := do
  let dir : System.FilePath := args.headD "vectors"
  let some w := Vectors.world
    | IO.eprintln "vectors: the world did not build"; return 1
  IO.FS.createDirAll dir
  IO.FS.createDirAll (dir / "model")
  for (name, cases) in files w do
    IO.FS.writeFile (dir / (name ++ ".json")) (render name cases)
    IO.println s!"{name}: {cases.length}"
  return 0
