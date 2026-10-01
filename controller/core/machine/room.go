package machine

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
)

// The room's facets (next/26). The room is the root Horizon; a caller never
// invokes it, only a facet of it, as an ordinary invoke with an ordinary
// frame. The runtime implements the facets itself, because what they do,
// create a Horizon, is a derivation, not a program. No registry may name one.
//
//	{"to": "system.narrow", "payload": {"actor": "<name in reach>", "reach": [...], "authority": {...}}}
//	   → returns {"reference": "<ref>"}: a narrowed Horizon for that actor, now in the caller's reach
//	{"to": "<ref>", "payload": <the actor's input>}
//	   → an ordinary invoke of the actor, under the narrowed Horizon
//
// A narrowing may only drop what is held. "reach" and "authority" are what
// the actor keeps; left out, it keeps all of the caller's.
const (
	SystemPrefix = "system."
	NarrowFacet  = "system.narrow"
	// RouteRunner is the route runner (next/30): a runtime-owned surface with a
	// binary, since it runs Starlark. It sits in the room's reach so every
	// Horizon carries it, but nothing may invoke it and no continuation is
	// shown it: the runtime runs it, under an empty Horizon, when a return is
	// routed.
	RouteRunner = "system.route-runner"
	// RouteFacet records a route (next/30): the caller names a step it was
	// shown, never a hash, and the runtime binds the source to that step's
	// type. It is tested against every recorded run of the type first (F8).
	//
	//	{"to": "system.route", "payload": {"step": <index into steps>, "source": "<starlark>", "accept": [<case>, ...]}}
	//	   → returns {"ok": true, "flips": [...]}, {"ok": false, "flips": [...]} or {"ok": false, "error": "..."}
	RouteFacet = "system.route"
)

// Runner is the route runner's surface, pinned by the runtime: its binary and
// that binary's sha256. No registry may name it.
func Runner(path, sha256 string) horizon.Surface {
	return horizon.Surface{Path: path, SHA256: sha256, System: "runner"}
}

var (
	str      = &contract.Schema{Type: "string"}
	strs     = &contract.Schema{Type: "array", Items: str}
	grants   = &contract.Schema{Type: "array", Items: &contract.Schema{Type: "object", Required: []string{"path", "scope"}, Fields: map[string]*contract.Schema{"path": str, "scope": str}}}
	narrowIn = &contract.Schema{Type: "object", Required: []string{"actor"}, Fields: map[string]*contract.Schema{
		"actor": str,
		"reach": strs,
		"authority": {Type: "object", Fields: map[string]*contract.Schema{
			"read": grants, "write": grants, "mach": strs, "connect": strs,
		}},
	}}
	routeIn = &contract.Schema{Type: "object", Required: []string{"step", "source"}, Fields: map[string]*contract.Schema{
		"step": {Type: "number"}, "source": str, "accept": {Type: "array", Items: &contract.Schema{Type: "number"}},
	}}
)

// RoomFacets are the room's facets, as surfaces genesis puts in the root reach.
func RoomFacets() map[string]horizon.Surface {
	return map[string]horizon.Surface{
		NarrowFacet: {System: "narrow", Input: narrowIn},
		RouteFacet:  {System: "route", Input: routeIn},
	}
}

// narrowing derives the Grant a system.narrow payload asks for under h, or
// says why it reaches past the Horizon. It is checked at intake, so a
// narrowing that asks for anything not held is a horizon fault.
func narrowing(h horizon.Horizon, payload json.RawMessage) (horizon.Grant, error) {
	var ask struct {
		Actor     string             `json:"actor"`
		Reach     *[]string          `json:"reach"`
		Authority *horizon.Authority `json:"authority"`
	}
	if err := json.Unmarshal(payload, &ask); err != nil {
		return horizon.Grant{}, err
	}
	if _, ok := h.Reach[ask.Actor]; !ok || strings.HasPrefix(ask.Actor, SystemPrefix) {
		return horizon.Grant{}, fmt.Errorf("%q is not an actor in reach", ask.Actor)
	}
	g := horizon.Grant{Actor: ask.Actor, Authority: h.Authority}
	if ask.Reach == nil {
		for name := range h.Reach {
			g.Reach = append(g.Reach, name)
		}
	} else {
		for _, name := range *ask.Reach {
			if _, ok := h.Reach[name]; !ok {
				return horizon.Grant{}, fmt.Errorf("narrowing keeps %q, which is not in reach", name)
			}
			if !slices.Contains(g.Reach, name) {
				g.Reach = append(g.Reach, name)
			}
		}
	}
	slices.Sort(g.Reach)
	if ask.Authority != nil {
		if !h.Authority.Allows(*ask.Authority) {
			return horizon.Grant{}, fmt.Errorf("narrowing asks for authority outside the Horizon")
		}
		g.Authority = *ask.Authority
	}
	return g, nil
}

// runFacet is what a facet of the room emits when it runs: nothing executes,
// the result is derived. system.narrow returns the reference: the name the
// waiting caller now holds it under, meaningful only in the reach it was
// recorded into.
func runFacet(s Instance) []byte {
	out, _ := json.Marshal(map[string]any{"to": s.Self.Frame, "payload": map[string]string{"reference": refName(s.Self.Refs)}})
	return out
}

// refName names the next reference a Horizon holding refs receives. It is a
// counter, not the invoke's stamp, because a continuation is shown its
// references and hashes never reach a model (next/30). It is in the runtime's
// namespace, which no registry may name, so it never shadows a surface.
func refName(refs map[string]horizon.Grant) string {
	for n := len(refs) + 1; ; n++ {
		name := fmt.Sprintf("%sref.%d", SystemPrefix, n)
		if _, held := refs[name]; !held {
			return name
		}
	}
}

// kept is a reach restricted to the names a grant keeps. The route runner is
// always kept: it grants nothing, and without it the actor's own children
// could never be routed.
func kept(reach map[string]horizon.Surface, names []string) map[string]horizon.Surface {
	out := map[string]horizon.Surface{}
	if r, ok := reach[RouteRunner]; ok {
		out[RouteRunner] = r
	}
	for _, name := range names {
		if s, ok := reach[name]; ok {
			out[name] = s
		}
	}
	return out
}
