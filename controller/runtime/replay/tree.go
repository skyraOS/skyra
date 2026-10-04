//go:build !js

package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// RecordedTree is bounded recorded work, never a pretend linear strand.
// Order follows predecessor edges and resolve dependencies in emission order.
// Horizons contain only visible metadata; waiting frames outside Root are omitted.
type RecordedTree struct {
	Version  int                        `json:"version"`
	Root     string                     `json:"root"`
	Frontier []string                   `json:"frontier"`
	Records  map[string]record.Record   `json:"records"`
	Opaque   map[string]string          `json:"opaque,omitempty"`
	Order    []string                   `json:"order"`
	Horizons map[string]horizon.Horizon `json:"horizons"`
	Covered  map[string]string          `json:"covered,omitempty"` // supplied endpoint -> retained endpoint covering it
	Inputs   map[string][]byte          `json:"inputs,omitempty"`  // visible refused emissions, keyed by record stamp
}

// ReadTree reconstructs exactly the root/frontier closure. It does not search
// descendants or use file times, and it verifies acceptance independently of
// replay. A valid commit or proposed release is insufficient.
func ReadTree(root, anchor string, leaves []string, target string, schema *contract.Schema, origin string) (RecordedTree, error) {
	return ReadTreeInRoom(root, anchor, leaves, target, schema, origin, "")
}

// ReadTreeInRoom additionally binds runtime calls to their enclosing genesis.
// The room is derived from the requester, never supplied in its payload.
func ReadTreeInRoom(root, anchor string, leaves []string, target string, schema *contract.Schema, origin, room string) (RecordedTree, error) {
	none := RecordedTree{}
	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		return none, err
	}
	if !stampName(anchor) || len(leaves) == 0 {
		return none, refusal("a tree needs an invocation and nonempty observations")
	}
	for _, leaf := range leaves {
		if !stampName(leaf) {
			return none, refusal("invalid observation stamp")
		}
	}
	w := walker{s: dir(root), seen: map[string]horizon.Horizon{}}
	checked := map[string]bool{}
	active := map[string]bool{}
	var accepted func(string) (horizon.Horizon, error)
	accepted = func(s string) (horizon.Horizon, error) {
		if !stampName(s) {
			return horizon.Horizon{}, refusal("invalid observation stamp")
		}
		if active[s] {
			return horizon.Horizon{}, refusal("cyclic acceptance history")
		}
		h, err := w.at(s)
		if err != nil {
			return h, err
		}
		if checked[s] {
			return h, nil
		}
		active[s] = true
		defer delete(active, s)
		r, err := tree.Load(root, s)
		if err != nil {
			return h, err
		}
		if r.Predecessor == "" {
			checked[s] = true
			return h, nil
		}
		under, err := accepted(r.Predecessor)
		if err != nil {
			return h, err
		}
		switch {
		case r.State == record.Resolve:
			winner, err := tree.Resolved(root, r.Frame, r.Round)
			if err != nil || winner != s {
				return h, refusal("resolve %s did not win its round", record.Short(s))
			}
			f, err := os.Open(framePath(root, r.Frame))
			if err != nil {
				return h, err
			}
			children, err := tree.Children(f, r.Round)
			f.Close()
			if err != nil {
				return h, err
			}
			if len(children) != len(r.Deps) {
				return h, refusal("resolve does not cover every slot")
			}
			for i, child := range children {
				d, err := tree.Result(root, child)
				if err != nil || d != r.Deps[i] {
					return h, refusal("resolve discharges do not match its slots")
				}
				if _, err := accepted(d); err != nil {
					return h, err
				}
			}
		case r.Supervisor:
			d, err := tree.Result(root, under.Frame)
			if err != nil || d != s {
				return h, refusal("supervisor record %s did not discharge its frame", record.Short(s))
			}
		default:
			receipt, err := tree.AcceptedHorizon(root, s)
			if err != nil {
				return h, refusal("record %s has no accepted release", record.Short(s))
			}
			a, _ := json.Marshal(h)
			b, _ := json.Marshal(receipt)
			if string(a) != string(b) {
				return h, refusal("accepted release is not what replay derives")
			}
			if h.Frame == s && h.Resume != nil {
				f, err := os.Open(framePath(root, h.Resume.Frame))
				if err != nil {
					return h, err
				}
				children, err := tree.Children(f, h.Resume.Stamp)
				f.Close()
				if err != nil {
					return h, err
				}
				if !slices.Contains(children, s) {
					return h, refusal("accepted child has no slot in its parent's round")
				}
			}
		}
		checked[s] = true
		return h, nil
	}
	inv, err := tree.Load(root, anchor)
	if err != nil {
		return none, err
	}
	if inv.State != record.Invoke || inv.Predecessor == "" || inv.To != target {
		return none, refusal("anchor is not an invoke of %s", target)
	}
	caller, err := accepted(inv.Predecessor)
	if err != nil {
		return none, err
	}
	if caller.Origin != origin {
		return none, refusal("anchor is outside the requester's Horizon")
	}
	if room != "" {
		outer := caller
		for outer.Resume != nil {
			outer = *outer.Resume
		}
		if outer.Frame != room {
			return none, refusal("anchor belongs to another room")
		}
	}
	surface, ok := caller.Reach[target]
	if !ok {
		return none, refusal("anchor target is outside its caller's reach")
	}
	value, err := contract.Decode(inv.Payload)
	if err != nil {
		return none, err
	}
	selected, err := machine.Selected(surface, value)
	if err != nil || string(contract.Canonical(selected)) != string(contract.Canonical(schema)) {
		return none, refusal("anchor was not invoked through the named contract")
	}
	if _, err := accepted(anchor); err != nil {
		return none, err
	}

	out := RecordedTree{Version: 1, Root: anchor, Frontier: slices.Clone(leaves), Records: map[string]record.Record{}, Opaque: map[string]string{}, Horizons: map[string]horizon.Horizon{}}
	slices.Sort(out.Frontier)
	out.Frontier = slices.Compact(out.Frontier)
	// Every branch must reach the anchor by predecessor ancestry. Dependencies
	// contribute joined branches, but cannot authorize a leaf from another root.
	belongs := map[string]bool{anchor: true}
	var within func(string) (bool, error)
	within = func(s string) (bool, error) {
		if yes, ok := belongs[s]; ok {
			return yes, nil
		}
		r, err := tree.Load(root, s)
		if err != nil {
			return false, err
		}
		if r.Predecessor == "" {
			belongs[s] = false
			return false, nil
		}
		yes, err := within(r.Predecessor)
		belongs[s] = yes
		return yes, err
	}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(s string) error {
		if visited[s] {
			return nil
		}
		yes, err := within(s)
		if err != nil {
			return err
		}
		if !yes {
			return refusal("observation %s is outside the retained invocation", record.Short(s))
		}
		h, err := accepted(s)
		if err != nil {
			return err
		}
		r, err := tree.Load(root, s)
		if err != nil {
			return err
		}
		frame := h
		if r.State == record.Return || r.State == record.Fault {
			frame, err = w.at(r.Predecessor)
			if err != nil {
				return err
			}
		}
		if frame.Origin != origin {
			// Walk waiting frames to the outermost call into this narrowed region.
			boundary := frame
			for boundary.Resume != nil && boundary.Resume.Origin != origin {
				boundary = *boundary.Resume
			}
			if boundary.Resume == nil {
				return refusal("observation is outside the requester's Horizon")
			}
			call := boundary.Frame
			if call == anchor {
				return refusal("anchor is a narrowed child")
			}
			c, err := tree.Load(root, call)
			if err != nil {
				return err
			}
			if err := visit(c.Predecessor); err != nil {
				return err
			}
			if _, err := accepted(call); err != nil {
				return err
			}
			if !visited[call] {
				out.Records[call] = c
				out.Order = append(out.Order, call)
				visited[call] = true
			}
			// Internal progress is represented only by its boundary invocation.
			if (r.State == record.Return || r.State == record.Fault) && frame.Frame == call {
				out.Records[s] = r
				out.Opaque[s] = call
				out.Order = append(out.Order, s)
			} else {
				out.Opaque[s] = call
			}
			visited[s] = true
			return nil
		}
		if s != anchor {
			if err := visit(r.Predecessor); err != nil {
				return err
			}
		}
		if r.State == record.Resolve {
			for _, d := range r.Deps {
				if err := visit(d); err != nil {
					return err
				}
			}
		}
		if !visited[s] {
			out.Records[s] = r
			if r.Fault != nil && !r.Supervisor {
				b, err := tree.Get(root, "inputs", r.Input)
				if err != nil {
					return err
				}
				if out.Inputs == nil {
					out.Inputs = map[string][]byte{}
				}
				out.Inputs[s] = b
			}
			out.Order = append(out.Order, s)
			visited[s] = true
		}
		// Metadata is derived, and stripped of ancestors outside the requested tree.
		visible := frame
		visible.Resume = nil
		out.Horizons[s] = visible
		return nil
	}
	for _, leaf := range out.Frontier {
		if err := visit(leaf); err != nil {
			return none, err
		}
	}
	// Normalize even opaque branches without disclosing their internals. Only
	// relationships between stamps the requester supplied are returned.
	out.Covered = map[string]string{}
	for _, leaf := range out.Frontier {
		seen := map[string]bool{}
		var cover func(string) error
		cover = func(s string) error {
			if s == "" || seen[s] {
				return nil
			}
			seen[s] = true
			if s != leaf && slices.Contains(out.Frontier, s) {
				if _, known := out.Covered[s]; !known {
					out.Covered[s] = leaf
				}
			}
			if s == anchor {
				return nil
			}
			r, err := tree.Load(root, s)
			if err != nil {
				return err
			}
			if err := cover(r.Predecessor); err != nil {
				return err
			}
			for _, d := range r.Deps {
				if err := cover(d); err != nil {
					return err
				}
			}
			return nil
		}
		if err := cover(leaf); err != nil {
			return none, err
		}
	}
	maximal := []string{}
	for _, leaf := range out.Frontier {
		if _, covered := out.Covered[leaf]; !covered {
			maximal = append(maximal, leaf)
		}
	}
	out.Frontier = maximal
	if len(out.Covered) == 0 {
		out.Covered = nil
	}
	// Stable topological emission order for open siblings comes from slot positions,
	// never from the frontier's lexicographic serialization. Filled below.
	out.Order, err = causalOrder(root, out.Records, out.Opaque, anchor)
	if err != nil {
		return none, err
	}
	if len(out.Opaque) == 0 {
		out.Opaque = nil
	}
	return out, nil
}

func stampName(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func framePath(root, frame string) string { return root + "/frames/" + frame }

func causalOrder(root string, records map[string]record.Record, opaque map[string]string, anchor string) ([]string, error) {
	children := map[string][]string{}
	for s, r := range records {
		if s != anchor {
			parent := r.Predecessor
			if _, visible := records[parent]; !visible && opaque[s] != "" {
				parent = opaque[s]
			}
			children[parent] = append(children[parent], s)
		}
	}
	positions := map[string]int{}
	for s, r := range records {
		if r.State != record.Invoke && r.State != record.Resolve {
			continue
		}
		frame := s
		if r.State == record.Resolve {
			frame = r.Frame
		}
		f, err := os.Open(framePath(root, frame))
		if err != nil {
			continue
		}
		slots, err := tree.Children(f, s)
		f.Close()
		if err != nil {
			continue
		}
		for i, c := range slots {
			positions[c] = i + 1
		}
	}
	for p, kids := range children {
		slices.SortFunc(kids, func(a, b string) int {
			if positions[a] != positions[b] {
				return positions[a] - positions[b]
			}
			// Only a single accepted non-slot successor is compatible with a serial
			// continuation. Stable tie breaking is not used as execution evidence.
			return 0
		})
		for i := 1; i < len(kids); i++ {
			if positions[kids[i]] == positions[kids[i-1]] {
				return nil, refusal("incompatible accepted successors beneath %s", record.Short(p))
			}
		}
	}
	done := map[string]bool{}
	visiting := map[string]bool{}
	var order []string
	var emit func(string) error
	emit = func(s string) error {
		if done[s] {
			return nil
		}
		if visiting[s] {
			return fmt.Errorf("cyclic recorded tree")
		}
		visiting[s] = true
		r, ok := records[s]
		if !ok {
			return nil
		}
		if s != anchor {
			parent := r.Predecessor
			if _, visible := records[parent]; !visible && opaque[s] != "" {
				parent = opaque[s]
			}
			if _, ok := records[parent]; ok {
				if err := emit(parent); err != nil {
					return err
				}
			}
		}
		for _, d := range r.Deps {
			if _, ok := records[d]; ok {
				if err := emit(d); err != nil {
					return err
				}
			}
		}
		visiting[s] = false
		done[s] = true
		order = append(order, s)
		return nil
	}
	var walk func(string) error
	walk = func(s string) error {
		if err := emit(s); err != nil {
			return err
		}
		for _, kid := range children[s] {
			if err := walk(kid); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(anchor); err != nil {
		return nil, err
	}
	if len(done) != len(records) {
		return nil, refusal("tree has disconnected visible records")
	}
	return order, nil
}
