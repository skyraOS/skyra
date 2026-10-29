package main

import (
	"slices"
	"sort"

	"skyra/controller/core/boundary"
	"skyra/controller/core/record"
)

// boundaryView is one Boundary of a conversation, as clients draw it (a region
// of the sky, a star per surface). It is the Boundary proper (next/24): what it
// reaches and the OS grants it carries. Every frame released under the same
// reach and authority is the same Boundary, so a conversation that never
// narrows has one. The webhost reads Boundaries and decides nothing.
type boundaryView struct {
	ID     string        `json:"id"`     // boundary.Hash: what it holds, by contract and surface identity, and its authority
	Root   bool          `json:"root"`   // the root Boundary, which genesis recorded
	Actors []string      `json:"actors"` // the surfaces that ran under it, by name
	Reach  []surfaceView `json:"reach"`  // what emissions under it may invoke, by name
	Grants int           `json:"grants"` // how many OS grants it carries
}

type surfaceView struct {
	Name string `json:"name"`
	ID   string `json:"id"` // the surface's identity: the same surface under any name, in any Boundary
}

// proper is the part of a Boundary that is authority, not position in a stack.
type proper struct {
	reach     map[string]boundary.Surface
	held      *boundary.CList
	authority boundary.Authority
}

func (p proper) id() string {
	return boundary.Boundary{Reach: p.reach, Held: p.held, Authority: p.authority}.Hash()
}

func (p proper) view() boundaryView {
	v := boundaryView{ID: p.id(), Actors: []string{},
		Grants: len(p.authority.Read) + len(p.authority.Write) + len(p.authority.Mach) + len(p.authority.Connect)}
	for name, s := range p.reach {
		v.Reach = append(v.Reach, surfaceView{name, boundary.Identity(s)})
	}
	sort.Slice(v.Reach, func(i, j int) bool { return v.Reach[i].Name < v.Reach[j].Name })
	return v
}

// boundaries derives a conversation's Boundaries from the root Boundary genesis recorded
// and every Boundary released in it, including those a reference names before
// anything runs under them. The root comes first, then the rest by id.
func boundaries(genesis *record.Record, released []boundary.Boundary) []boundaryView {
	byID := map[string]*boundaryView{}
	names := map[string]string{} // surface identity → its name in the root reach
	var order []string
	add := func(p proper, actor string) {
		id := p.id()
		v := byID[id]
		if v == nil {
			pv := p.view()
			v = &pv
			byID[id] = v
			order = append(order, id)
		}
		if actor != "" && !slices.Contains(v.Actors, actor) {
			v.Actors = append(v.Actors, actor)
			slices.Sort(v.Actors)
		}
	}
	root := ""
	if genesis != nil && genesis.Reach != nil {
		p := proper{genesis.Reach, nil, authorityOf(genesis.Authority)}
		root = p.id()
		for name, s := range genesis.Reach {
			names[boundary.Identity(s)] = name
		}
		add(p, "")
	}
	for _, h := range released {
		for name, s := range h.Reach {
			if _, ok := names[boundary.Identity(s)]; !ok {
				names[boundary.Identity(s)] = name
			}
		}
	}
	for _, h := range released {
		actor := ""
		if h.Surface != nil {
			actor = names[boundary.Identity(*h.Surface)]
		}
		add(proper{h.Reach, h.Held, h.Authority}, actor)
		for _, g := range h.Refs {
			kept := map[string]boundary.Surface{}
			for _, name := range g.Reach {
				if s, ok := h.Reach[name]; ok {
					kept[name] = s
				}
			}
			add(proper{kept, g.Held, g.Authority}, "")
		}
	}
	out := []boundaryView{}
	if v := byID[root]; v != nil {
		v.Root = true
		out = append(out, *v)
	}
	var rest []string
	for _, id := range order {
		if id != root {
			rest = append(rest, id)
		}
	}
	slices.Sort(rest)
	for _, id := range rest {
		out = append(out, *byID[id])
	}
	return out
}

func authorityOf(a *boundary.Authority) boundary.Authority {
	if a == nil {
		return boundary.Authority{}
	}
	return *a
}
