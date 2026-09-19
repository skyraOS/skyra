```go
// Engram is addressed by EngramID = MoveID = hash(Action + ordered Params). That action+params
// is the ENTIRE hash preimage: same action+params → same EngramID; nothing below enters the hash.
// The fields here are ASSOCIATED DATA keyed by that stable ID, so they may grow/change without
// moving the ID (the git-notes pattern). Where the executable artifact / current compile attaches
// is still open, but it likewise does not change EngramID — the ID is fixed by action+params alone.
type Engram struct {
    ID          EngramID
    Invocations map[InvocationID][]WorkUnit // per-run audit logs, keyed by run; never hashed into ID.
    RememberCommit map[InvocationID]*Possibility // OPEN: parked mapping; Possibility representation and commit semantics remain undefined.
}

type InvocationID struct {
    Value string
}

type WorkUnit struct {
    Timestamp       time.Time
    WorkArtifact    *WorkArtifact
    OperatingEntity *Entity
    Change          OpcodeID
}

type OpcodeID struct {
    Value string
}

type WorkArtifact struct{}
```
