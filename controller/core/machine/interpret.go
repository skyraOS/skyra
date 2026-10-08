package machine

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// interpret derives the state from the emitter's operands (thought 14). An
// emitter supplies only an address and a payload:
//
//	{"to": "<address>", "payload": <JSON>}
//
// It never says whether that is a call or a return. The runtime derives it
// from where the emitter stands (improvements, items 1 and 2):
//
//	addresses the slot its frame discharges                                     → return
//	names a capability in reach, payload fits exactly one of its contracts      → invoke
//	names anything else                                                         → fault (horizon)
//	addresses nothing                                                           → fault (leaf)
//	not readable as operands                                                    → fault (malformed)
//	payload fits none of the contracts of what it addresses, or more than one   → fault (contract)
//
// No name is reserved: the slot is the frame's id, a stamp no capability can
// be named. Every fault carries, as its payload, the observation of why it
// faulted. It is total and pure, so replay derives the same state again.
func interpret(h horizon.Horizon, out []byte) record.Record {
	if len(out) == 0 {
		return faulted(record.Custody, "no output", nil)
	}
	if !utf8.Valid(out) {
		return faulted(record.Malformed, "output is not UTF-8 text", nil)
	}
	printed := contract.Canonical(string(out))
	v, err := contract.Decode(out)
	o, isObject := v.(map[string]any)
	if _, addressed := o["to"]; err != nil || !isObject || !addressed {
		return faulted(record.Leaf, "output addresses nothing", printed)
	}
	to, isName := o["to"].(string)
	payload, hasPayload := o["payload"]
	if !isName || !hasPayload || len(o) != 2 {
		return faulted(record.Malformed, `addressed output must be exactly {"to": <address>, "payload": <JSON>}`, printed)
	}
	attempt := contract.Canonical(o)
	if h.Surface != nil && to == h.Frame {
		// Any JSON value: a return declares no contract.
		return record.Record{State: record.Return, Payload: contract.Canonical(payload)}
	}
	target, ok := h.Reach[to]
	if g, isRef := h.Refs[to]; isRef && !ok {
		target, ok = h.Reach[g.Actor]
	}
	if !ok {
		return faulted(record.Horizon, fmt.Sprintf("%q is not in reach", to), attempt)
	}
	if _, err := Selected(target, payload); err != nil {
		return faulted(record.Contract, to+": "+err.Error(), attempt)
	}
	if target.System == "horizon" {
		if kind, err := checkSyscall(h, contract.Canonical(payload)); err != nil {
			return faulted(kind, to+": "+err.Error(), attempt)
		}
	}
	return record.Record{State: record.Invoke, To: to, Payload: contract.Canonical(payload)}
}

// Selected is the one contract of a target that a payload fits (next/32,
// decisions 20, 27). A contract is (target, schema): the invoke names the
// target, and fit picks the schema. None fits, or more than one does, is a
// contract fault; the kernel never reasons about whether two schemas overlap,
// it refuses when an ambiguity actually occurs.
func Selected(target horizon.Surface, payload any) (*contract.Schema, error) {
	var fit []*contract.Schema
	var first error
	closest := false // the error kept is from a contract whose required fields the payload has
	for _, c := range target.All() {
		if err := contract.Fits(c, payload); err != nil {
			if near := hasRequired(c, payload); first == nil || near && !closest {
				first, closest = err, near
			}
			continue
		}
		fit = append(fit, c)
	}
	switch {
	case len(fit) == 1:
		return fit[0], nil
	case len(fit) > 1:
		return nil, fmt.Errorf("ambiguous: the payload fits %d of its contracts", len(fit))
	case first == nil:
		return nil, errors.New("no contract")
	}
	return nil, first
}

// faulted is an intake fault. Its payload is the observation it was derived
// from, so the parent can pass it on (next/24, "Faults carry their observation").
func faulted(kind, cause string, observation json.RawMessage) record.Record {
	return record.Record{State: record.Fault, Payload: observation, Fault: &record.FaultInfo{Kind: kind, Cause: cause}}
}

// hasRequired reports whether a payload is an object holding every field a
// contract requires: the contract it was most likely meant for.
func hasRequired(c *contract.Schema, payload any) bool {
	o, ok := payload.(map[string]any)
	if !ok || c.Type != "object" {
		return false
	}
	for _, r := range c.Required {
		if _, ok := o[r]; !ok {
			return false
		}
	}
	return true
}
