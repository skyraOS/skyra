// Package privilege is the system-administered decision principal.
package privilege

import (
	"fmt"
	c "skyra-v1/src/skyra/skyra-dna-v10/causal"
	"skyra-v1/src/skyra/skyra-dna-v10/kernel"
)

type Decision int

const (
	Auto Decision = iota
	Ask
	Deny
)

func (d Decision) String() string { return [...]string{"Auto", "Ask", "Deny"}[d] }

type Policy interface {
	Decide(c.NoveltyVerdict, c.GrantKind) Decision
}
type NovelPolicy struct{}

func (NovelPolicy) Decide(v c.NoveltyVerdict, _ c.GrantKind) Decision {
	if v.Known {
		return Auto
	}
	return Ask
}

type Registry interface {
	Known(c.Endpoint) c.NoveltyVerdict
}
type Trace interface{ Printf(string, ...any) }
type discardTrace struct{}

func (discardTrace) Printf(string, ...any) {}

type Boundary struct {
	policy    Policy
	registry  Registry
	trap      kernel.Trap
	trace     Trace
	pending   map[c.MoveID]c.FaultValue
	next      int
	wakes     int
	decisions int
}

func New(policy Policy, reg Registry, trap kernel.Trap, trace Trace) *Boundary {
	if trace == nil {
		trace = discardTrace{}
	}
	return &Boundary{policy: policy, registry: reg, trap: trap, trace: trace, pending: map[c.MoveID]c.FaultValue{}}
}

func (b *Boundary) Request(f c.FaultValue) (c.MoveID, *c.Grant, c.Outcome) {
	b.decisions++
	verdict := b.registry.Known(f.Endpoint)
	decision := b.policy.Decide(verdict, f.Kind)
	b.trace.Printf("PRIVILEGE DECIDE kind=%s novelty={known:%t history:%q} policy=%s\n", f.Kind, verdict.Known, verdict.History.Note, decision)
	if decision == Deny {
		return c.MoveID{}, nil, c.Failure
	}
	if decision == Auto {
		g, out := b.authorize(f)
		return c.MoveID{}, &g, out
	}
	b.next++
	b.wakes++
	id := c.MoveID{Value: fmt.Sprintf("permission.%d", b.next)}
	b.pending[id] = f
	return id, nil, c.NotYet
}

// Approve is the later human event. For Register, privilege decides and then
// traps to the kernel; only after that does the caller feed the Grant to Recognize.
func (b *Boundary) Approve(id c.MoveID) (c.Grant, c.Outcome) {
	f, ok := b.pending[id]
	if !ok {
		return c.Grant{}, c.Failure
	}
	delete(b.pending, id)
	b.trace.Printf("HUMAN APPROVE id=%s kind=%s (separate event stack)\n", id.Value, f.Kind)
	return b.authorize(f)
}
func (b *Boundary) authorize(f c.FaultValue) (c.Grant, c.Outcome) {
	g := c.Grant{Kind: f.Kind, Subject: f.Endpoint, Scope: f.Scope}
	if f.Kind == c.RegisterGrant {
		if f.Registration == nil {
			return c.Grant{}, c.Failure
		}
		b.trace.Printf("PRIVILEGE TRAP decision=allow-register -> kernel (privilege holds no Mint)\n")
		cap, out := b.trap.EnterRegister(*f.Registration)
		if out != c.Success {
			return c.Grant{}, out
		}
		g.Minted = &cap
	}
	return g, c.Success
}
func (b *Boundary) WakeCount() int     { return b.wakes }
func (b *Boundary) DecisionCount() int { return b.decisions }
