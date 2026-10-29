package proc

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Source is a procedure as it is written: nodes by local name, bindings to
// those names. JSON is the source; remember writes it, compiled from a
// trace, and models never do (32, "The source and the compiled form").
// Compile turns it into nodes by hash, the only form the interpreter runs.
type Source struct {
	Contract   *Schema               `json:"contract"`
	Original   string                `json:"original,omitempty"`
	Refinement *Refinement           `json:"refinement,omitempty"`
	Return     json.RawMessage       `json:"return"`
	Nodes      map[string]SourceNode `json:"nodes"`
	Order      []string              `json:"order,omitempty"` // the nodes' emission order, when several are ready at once; by name otherwise
}

// SourceNode is a node by local name.
type SourceNode struct {
	Call  json.RawMessage `json:"call"`
	In    json.RawMessage `json:"in"`
	Each  json.RawMessage `json:"each,omitempty"`
	After []string        `json:"after,omitempty"`
}

// Compile turns a source into a procedure and its nodes, each named by its
// hash. Every binding must be to a node it runs after, so its value exists
// when it runs; the DAG must have no cycle.
func Compile(src Source) (Procedure, map[string]Node, error) {
	if src.Contract == nil {
		return Procedure{}, nil, fmt.Errorf("a procedure implements a contract")
	}
	order := src.Order
	if order == nil {
		for name := range src.Nodes {
			order = append(order, name)
		}
		slices.Sort(order)
	}
	ids := map[string]string{}
	nodes := map[string]Node{}
	visiting := map[string]bool{}
	var compile func(name string) (string, error)
	compile = func(name string) (string, error) {
		if id, ok := ids[name]; ok {
			return id, nil
		}
		sn, ok := src.Nodes[name]
		if !ok {
			return "", fmt.Errorf("no node %q", name)
		}
		if visiting[name] {
			return "", fmt.Errorf("node %q depends on itself", name)
		}
		visiting[name] = true
		n := Node{After: []string{}}
		for _, a := range sn.After {
			id, err := compile(a)
			if err != nil {
				return "", err
			}
			n.After = append(n.After, id)
		}
		before := closure(name, src.Nodes)
		for _, part := range []json.RawMessage{sn.Call, sn.In, sn.Each} {
			for _, r := range Refs(part) {
				if !before[r] {
					return "", fmt.Errorf("node %q binds %q, which it does not run after", name, r)
				}
			}
		}
		var err error
		if n.Call, err = rename(sn.Call, ids); err != nil {
			return "", err
		}
		if n.In, err = rename(sn.In, ids); err != nil {
			return "", err
		}
		if n.Each, err = rename(sn.Each, ids); err != nil {
			return "", err
		}
		id := n.ID()
		ids[name], nodes[id] = id, n
		visiting[name] = false
		return id, nil
	}
	depended := map[string]bool{}
	for _, sn := range src.Nodes {
		for _, a := range sn.After {
			depended[a] = true
		}
	}
	p := Procedure{Header: Header{Contract: src.Contract, Original: src.Original, Refinement: src.Refinement}, Names: map[string]string{}}
	for _, name := range order {
		if _, err := compile(name); err != nil {
			return Procedure{}, nil, err
		}
	}
	for _, name := range order {
		if !depended[name] {
			p.Body = append(p.Body, ids[name])
		}
	}
	for name, id := range ids {
		p.Names[id] = name
	}
	for _, r := range Refs(src.Return) {
		if _, ok := ids[r]; !ok {
			return Procedure{}, nil, fmt.Errorf("the return binds %q, which is no node", r)
		}
	}
	var err error
	if p.Return, err = rename(src.Return, ids); err != nil {
		return Procedure{}, nil, err
	}
	if r := p.Header.Refinement; r != nil {
		byID := func(m map[string]string) (map[string]string, error) {
			if m == nil {
				return nil, nil
			}
			out := map[string]string{}
			for name, t := range m {
				id, ok := ids[name]
				if !ok {
					return nil, fmt.Errorf("a refinement on %q, which is no node", name)
				}
				out[id] = t
			}
			return out, nil
		}
		edges, err := byID(r.Edges)
		if err != nil {
			return Procedure{}, nil, err
		}
		shown, err := byID(r.EdgesShown)
		if err != nil {
			return Procedure{}, nil, err
		}
		on, err := byID(r.On)
		if err != nil {
			return Procedure{}, nil, err
		}
		p.Header.Refinement = &Refinement{Input: r.Input, Edges: edges, InputShown: r.InputShown, EdgesShown: shown, Header: r.Header, On: on}
	}
	return p, nodes, nil
}

// closure is every node name runs after, transitively.
func closure(name string, nodes map[string]SourceNode) map[string]bool {
	seen := map[string]bool{}
	var walk func(n string)
	walk = func(n string) {
		for _, a := range nodes[n].After {
			if !seen[a] {
				seen[a] = true
				walk(a)
			}
		}
	}
	walk(name)
	return seen
}
