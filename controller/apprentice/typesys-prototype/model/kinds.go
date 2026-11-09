package model

import "slices"

// Defs is a scope's definitions: tools and models alike.
func Defs(entries []Iface) []DefLoc {
	var ds []DefLoc
	for _, e := range entries {
		if e.Tag == IDef {
			ds = append(ds, e.Def)
		}
	}
	return ds
}

// ToolTypes is a scope's tool types: its definitions among its program's
// tools, each once, first occurrence first. Models never count.
func ToolTypes(act []DefLoc, entries []Iface) []DefLoc {
	var ts []DefLoc
	for _, d := range Defs(entries) {
		if slices.Contains(act, d) && !slices.Contains(ts, d) {
			ts = append(ts, d)
		}
	}
	return ts
}

// AgentKind is an agent's kind, counted under k: at most k tool types acts,
// more delegates, and none is no agent.
func AgentKind(k int, act []DefLoc, entries []Iface) (Kind, bool) {
	switch t := len(ToolTypes(act, entries)); {
	case t == 0:
		return "", false
	case t <= k:
		return KindAct, true
	}
	return KindDelegate, true
}

// Ops is the operators a kind emits.
func Ops(k Kind) []Op {
	switch k {
	case KindRoot:
		return []Op{OpDelegate, OpBoot, OpReturn}
	case KindDelegate:
		return []Op{OpDelegate, OpReturn}
	case KindAct:
		return []Op{OpAct, OpReturn}
	}
	return []Op{OpReturn}
}

// Pushed is the kind of the frame an operator pushes, given its scope's
// interfaces: an act pushes a leaf, a delegation an agent counted under k, a
// boot a root, and a return pushes nothing.
func Pushed(k int, act []DefLoc, o Op, entries []Iface) (Kind, bool) {
	switch o {
	case OpAct:
		return KindLeaf, true
	case OpDelegate:
		return AgentKind(k, act, entries)
	case OpBoot:
		return KindRoot, true
	}
	return "", false
}
