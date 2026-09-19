package scheduler

import (
"context"
"reflect"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"testing"
)

func TestConflictingBufferedEnvelopeHasNoPartialEffects(t *testing.T) {
	dispatches := 0
	s, _ := NewScheduler(&testController{dispatch: func(context.Context, protocol.Move, uint64, func(protocol.OpcodeEvent) error) error { dispatches++; return nil }}, func(protocol.RootResult) error { t.Error("unexpected notification"); return nil })
	m := rootMove()
	id := m.Op.(protocol.Invoke).InvocationID
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	child := support.CloneMove(m)
	inv := child.Op.(protocol.Invoke)
	inv.InvocationID = protocol.InvocationID{Value: "child"}
	inv.ReturnFd = id
	child.Op = inv
	batch := protocol.MoveBatch{Source: &id, Moves: []protocol.Move{child}}
	if err := s.transition(context.Background(), batch); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Snapshot(id)
	conflict := batch
	conflict.Moves = support.CloneMoves(batch.Moves)
	ci := conflict.Moves[0].Op.(protocol.Invoke)
	ci.Params[0].Content.Bytes = []byte("conflict")
	conflict.Moves[0].Op = ci
	if err := s.transition(context.Background(), protocol.Envelope{Runtime: protocol.DispatchResult{Invocation: id}, Continuation: &conflict}); err == nil {
		t.Fatal("conflicting output accepted")
	}
	after, _ := s.Snapshot(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("conflicting envelope partially applied")
	}
	if _, exists := s.Snapshot(inv.InvocationID); exists {
		t.Fatal("conflicting envelope admitted child")
	}
	if err := s.transition(context.Background(), protocol.DispatchResult{Invocation: id}); err != nil {
		t.Fatal(err)
	}
	if dispatches != 2 {
		t.Fatal("original buffered output stranded")
	}
}

