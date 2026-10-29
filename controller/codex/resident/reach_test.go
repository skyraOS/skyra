package main

import "skyra/controller/codex/proceduretools"

// reachOf is a continuation's reach that holds everything: first is each
// name's registered contract, contracts every contract of a name with more.
func reachOf(first map[string]*schema, contracts map[string][]*schema) proceduretools.Reach {
	r := proceduretools.Reach{}
	for name, s := range first {
		all := contracts[name]
		if all == nil {
			all = []*schema{s}
		}
		held := make([]int, len(all))
		for i := range held {
			held[i] = i
		}
		r[name] = proceduretools.Entry{Contracts: all, Held: held}
	}
	return r
}

// holding is r with only the named actors held, all of each; nil names hold
// everything, as r does.
func holding(r proceduretools.Reach, names []string) proceduretools.Reach {
	if names == nil {
		return r
	}
	out := proceduretools.Reach{}
	for name, e := range r {
		e.Held = []int{}
		for _, n := range names {
			if n == name {
				for i := range e.Contracts {
					e.Held = append(e.Held, i)
				}
			}
		}
		out[name] = e
	}
	return out
}
