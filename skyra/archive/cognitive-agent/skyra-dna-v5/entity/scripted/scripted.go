// Package scripted is a deterministic interpreter.
//
// This is what questions.md §1 calls for as the FIRST BUILD: the learning
// substrate with everything else stubbed — hand-seeded possibilities,
// deterministic bodies, one entity, no model. It tests the claim everything rests
// on and dodges every other open item.
//
// It is not a mock. It is an entity like any other: it interprets an expectation
// and emits. The only difference is that its judgement is a table, which makes
// the credit mechanism observable without a model's variance sitting on top of
// it.
package scripted

import (
	"fmt"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
)

// Body is one expectation's behaviour: given the frame and how many steps have
// already run, emit.
type Body func(f frame.Frame, step int) (grammar.Emission, error)

// Table maps an expectation's expression to its body.
type Table map[string]Body

type Interpreter struct {
	T    Table
	step int
	Seen []*graph.Observation // what Monitor was handed, in order
}

func New(t Table) func() entity.Interpreter {
	return func() entity.Interpreter { return &Interpreter{T: t} }
}

func (i *Interpreter) Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error) {
	b, ok := i.T[x.Expression]
	if !ok {
		return nil, fmt.Errorf("scripted entity cannot man %q", x.Expression)
	}
	em, err := b(f, i.step)
	i.step++
	return em, err
}

// Monitor records what it was pushed, so a test can assert EMITTED ORDER.
func (i *Interpreter) Monitor(o *graph.Observation) { i.Seen = append(i.Seen, o) }
