package main

import "skyra/apprentice/proc"

// heldReach is an envelope's reach that holds everything: first is each name's
// registered contract, contracts every contract of a name with more.
func heldReach(first map[string]*proc.Schema, contracts map[string][]*proc.Schema) map[string]entry {
	r := map[string]entry{}
	for name, s := range first {
		all := contracts[name]
		if all == nil {
			all = []*proc.Schema{s}
		}
		held := make([]int, len(all))
		for i := range held {
			held[i] = i
		}
		r[name] = entry{Contracts: all, Held: held}
	}
	return r
}
