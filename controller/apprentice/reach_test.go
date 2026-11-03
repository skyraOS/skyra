package main

import "skyra/apprentice/proc"

// heldReach is an envelope's reach: first is each name's registered
// contract, contracts every contract of a name with more.
func heldReach(first map[string]*proc.Schema, contracts map[string][]*proc.Schema) map[string]entry {
	r := map[string]entry{}
	for name, s := range first {
		all := contracts[name]
		if all == nil {
			all = []*proc.Schema{s}
		}
		r[name] = entry{Contracts: all}
	}
	return r
}
