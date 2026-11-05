import Lean.Data.Json
import Invariants.HumanRoot

open Lean (Json)

def rootSets : List (List String) := [[], ["a"], ["b"], ["a", "b"], ["missing"], ["a", "a"]]
def entrySets : List (List String) := [[], ["think"], ["act"], ["think", "act"]]
def rootTargets : List String := ["think", "act", "unknown"]
def js (xs : List String) : Json := .arr (xs.map Json.str).toArray
def jsets (xs : List (List String)) : Json := .arr (xs.map js).toArray

def jn (n : Nat) : Json := .num n

def rootVectors : Json :=
  let passing := rootSets.zipIdx.flatMap fun (scope, s) =>
    rootSets.zipIdx.map fun (inputs, i) =>
      Json.arr #[jn s, jn i, .bool (HumanRoot.mayPass scope inputs)]
  let grants := rootSets.zipIdx.flatMap fun (ceiling, c) =>
    rootSets.zipIdx.map fun (grant, g) =>
      Json.arr #[jn c, jn g, .bool (HumanRoot.mayGrant ceiling grant)]
  let calls := entrySets.zipIdx.flatMap fun (entries, e) =>
    rootTargets.zipIdx.flatMap fun (target, t) =>
      rootSets.zipIdx.flatMap fun (scope, s) =>
        rootSets.zipIdx.flatMap fun (inputs, i) =>
          rootSets.zipIdx.flatMap fun (ceiling, c) =>
            rootSets.zipIdx.map fun (grant, g) =>
              Json.arr #[jn e, jn t, jn s, jn i, jn c, jn g,
                .bool (HumanRoot.mayInvoke entries target scope inputs ceiling grant)]
  let authoring := ([HumanRoot.Ingress.human, .runtime, .storage, .worker].zipIdx).map fun (ingress, i) =>
    Json.arr #[jn (i+1), .bool (HumanRoot.mayAuthor ingress)]
  Json.mkObj [("sets", jsets rootSets), ("entries", jsets entrySets), ("targets", js rootTargets),
    ("mayAuthor", .arr authoring.toArray), ("mayPass", .arr passing.toArray),
    ("mayGrant", .arr grants.toArray), ("mayInvoke", .arr calls.toArray)]

def main (args : List String) : IO Unit := do
  match args with
  | [path] => IO.FS.writeFile path (rootVectors.compress ++ "\n")
  | _ => throw (IO.userError "usage: rootvectors OUTPUT.json")
