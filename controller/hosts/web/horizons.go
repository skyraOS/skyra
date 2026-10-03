package main

import (
	"encoding/json"
	"slices"
	"sort"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// horizonView is one Horizon of a conversation, as clients draw it (a region
// of the sky, a star per surface). It is the Horizon proper (next/24): what it
// reaches and the OS grants it carries. Every frame released under the same
// reach and authority is the same Horizon, so a conversation that never
// narrows has one. The webhost reads Horizons and decides nothing.
type horizonView struct {
	ID     string        `json:"id"`     // the hash of its reach (by surface identity) and authority
	Root   bool          `json:"root"`   // the room: the Horizon genesis recorded
	Actors []string      `json:"actors"` // the surfaces that ran under it, by name
	Reach  []surfaceView `json:"reach"`  // what emissions under it may invoke, by name
	Grants int           `json:"grants"` // how many OS grants it carries
}

type surfaceView struct {
	Name string `json:"name"`
	ID   string `json:"id"` // the surface's identity: the same surface under any name, in any Horizon
}

// proper is the part of a Horizon that is authority, not position in a stack.
type proper struct {
	reach     map[string]horizon.Surface
	authority horizon.Authority
}

func (p proper) id() string {
	ids := map[string]string{}
	for name, s := range p.reach {
		ids[name] = horizon.Identity(s)
	}
	b, _ := json.Marshal(map[string]any{"reach": ids, "authority": p.authority})
	return record.Sum(b)
}

func (p proper) view() horizonView {
	v := horizonView{ID: p.id(), Actors: []string{},
		Grants: len(p.authority.Read) + len(p.authority.Write) + len(p.authority.Mach) + len(p.authority.Connect)}
	for name, s := range p.reach {
		v.Reach = append(v.Reach, surfaceView{name, horizon.Identity(s)})
	}
	sort.Slice(v.Reach, func(i, j int) bool { return v.Reach[i].Name < v.Reach[j].Name })
	return v
}

// horizons derives a conversation's Horizons from the room genesis recorded
// and every Horizon released in it, including those a reference names before
// anything runs under them. The root comes first, then the rest by id.
func horizons(genesis *record.Record, released []horizon.Horizon) []horizonView {
	byID := map[string]*horizonView{}
	names := map[string]string{} // surface identity → its name in the room
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
		p := proper{genesis.Reach, authorityOf(genesis.Authority)}
		root = p.id()
		for name, s := range genesis.Reach {
			names[horizon.Identity(s)] = name
		}
		add(p, "")
	}
	for _, h := range released {
		for name, s := range h.Reach {
			if _, ok := names[horizon.Identity(s)]; !ok {
				names[horizon.Identity(s)] = name
			}
		}
	}
	for _, h := range released {
		actor := ""
		if h.Surface != nil {
			actor = names[horizon.Identity(*h.Surface)]
		}
		add(proper{h.Reach, h.Authority}, actor)
		for _, g := range h.Refs {
			kept := map[string]horizon.Surface{}
			for _, name := range g.Reach {
				if s, ok := h.Reach[name]; ok {
					kept[name] = s
				}
			}
			add(proper{kept, g.Authority}, "")
		}
	}
	out := []horizonView{}
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

func authorityOf(a *horizon.Authority) horizon.Authority {
	if a == nil {
		return horizon.Authority{}
	}
	return *a
}
