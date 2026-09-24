package execution

import p "skyra-v1/src/skyra/protocol"

// Views are detached copies for diagnostics and acceptance evidence. Mutating
// one never reaches the context; authority changes only through its methods.

type FrameView struct {
	ID, Parent p.FrameID
	Target     p.ExecutableID
	Phase      string
	Result     p.Observation
}
type ParentView struct {
	Live bool
	Busy p.FrameID
}
type MappingView struct {
	Actual     uint64
	Executable Executable
}

func (e *ExecutionContext) Frames() map[p.FrameID]FrameView {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make(map[p.FrameID]FrameView, len(e.bookkeeper.frames))
	for id, f := range e.bookkeeper.frames {
		result[id] = FrameView{ID: f.id, Parent: f.parent, Target: f.target, Phase: f.phase, Result: p.Observation{Key: f.result.Key, Bytes: append([]byte(nil), f.result.Bytes...)}}
	}
	return result
}
func (e *ExecutionContext) Parents() map[p.FrameID]ParentView {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make(map[p.FrameID]ParentView, len(e.bookkeeper.parents))
	for id, c := range e.bookkeeper.parents {
		result[id] = ParentView{Live: c.live, Busy: c.busy}
	}
	return result
}
func (e *ExecutionContext) Mappings() map[uint64]MappingView {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make(map[uint64]MappingView, len(e.horizon))
	for virtual, m := range e.horizon {
		x := m.executable
		x.Contracts = append([]ArgContract(nil), x.Contracts...)
		result[virtual] = MappingView{Actual: m.actual, Executable: x}
	}
	return result
}

// SinkWired reports that root results have a Bookkeeper sink and a destination.
func (e *ExecutionContext) SinkWired() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	b := e.bookkeeper
	return b.bottom != nil && b.bottom.id == b.rootSink && b.bottom.phase == "sink" && e.rootEgress != nil
}
