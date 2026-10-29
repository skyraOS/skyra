package main

import (
	"encoding/json"
	"fmt"
	"os"

	"skyra/apprentice/proc"
)

// installBundle is operator provisioning, before the registered procedures
// enter a Boundary. It validates the entire bundle before writing any node.
func installBundle(dir, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var bundle struct {
		Version    int           `json:"version"`
		Procedures []proc.Source `json:"procedures"`
	}
	if err := json.Unmarshal(b, &bundle); err != nil {
		return err
	}
	if bundle.Version != 1 || len(bundle.Procedures) == 0 {
		return fmt.Errorf("invalid procedure bundle")
	}
	store := proc.Store{Dir: dir}
	keys := map[string]bool{}
	type compiled struct {
		p     proc.Procedure
		nodes map[string]proc.Node
	}
	var all []compiled
	for _, src := range bundle.Procedures {
		p, nodes, err := proc.Compile(src)
		if err != nil {
			return err
		}
		key := proc.Key(src.Contract)
		if keys[key] {
			return fmt.Errorf("bundle repeats a contract")
		}
		keys[key] = true
		if store.Has(src.Contract) {
			held, _, err := store.Procedure(src.Contract)
			if err != nil {
				return err
			}
			a, _ := json.Marshal(held)
			b, _ := json.Marshal(p)
			if string(a) != string(b) {
				return fmt.Errorf("bundle conflicts with an installed representation")
			}
			continue
		}
		all = append(all, compiled{p, nodes})
	}
	for _, c := range all {
		if err := store.Put(c.p, c.nodes); err != nil {
			return err
		}
	}
	return nil
}
