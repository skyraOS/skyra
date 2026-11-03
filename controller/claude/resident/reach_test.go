package main

import (
	"slices"

	"skyra/controller/codex/proceduretools"
)

// reachOf is a continuation's reach: first is each name's registered
// contract, contracts every contract of a name with more.
func reachOf(first map[string]*schema, contracts map[string][]*schema) proceduretools.Reach {
	r := proceduretools.Reach{}
	for name, s := range first {
		all := contracts[name]
		if all == nil {
			all = []*schema{s}
		}
		r[name] = proceduretools.Entry{Contracts: all}
	}
	return r
}

// holding is r with only the named actors kept, all of each; nil names keep
// everything, as r does.
func holding(r proceduretools.Reach, names []string) proceduretools.Reach {
	if names == nil {
		return r
	}
	out := proceduretools.Reach{}
	for name, e := range r {
		if slices.Contains(names, name) {
			out[name] = e
		}
	}
	return out
}
