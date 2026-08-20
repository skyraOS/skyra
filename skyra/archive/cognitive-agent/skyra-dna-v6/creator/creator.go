// Package creator holds the creator slot.
//
// Entity dissolved into two slots, and they are reached on different paths:
//
//	RUNTIME PATH   compiler + host. THE CREATOR IS ABSENT. two levels.
//	FAILURE PATH   the creator pops. reached ONLY when compilation fails.
//
// Slots on an engram, not kinds of thing. The distinction is load-bearing: "creator and
// compiler are two types of entity" would be a capability declaration, which the design
// dissolved; "an engram has a creator slot and a compiler slot" is a record of what
// happened. One identity may fill both. Python only ever appears in compiler slots, and
// nobody declared that — it is observed, because python has never been recorded in a
// creator slot.
//
// A LEAF HAS NO CREATION HALF. There is no interior body to author; its observation
// comes from the crossing.
//
// THE CREATOR IS NOT TOLD ITS BEHAVIOUR WILL BE COMPILED. Knowing your output will be
// compiled is knowing something about your reception. And there is no choice to
// withhold: the frame supplies the syntax, the syntax determines the compiler, so
// there was never a decision point. The failure path leaks it, and that is correct —
// ON FAILURE ONLY, and only about the specific thing that failed.
//
//	LEARNING WHAT COMPILES   learning to do the work.        ALLOWED.
//	OBSERVING RECEPTION      learning to flatter the judge.  REFUSED.
//
// THE BLOCKER, RECORDED: the creator appears in no citation. An address gives the
// failure path somewhere to POP TO and leaves authorship UNCITABLE — for an address the
// runtime knows only where it sent, never what produced. Half the problem is closed.
package creator

import "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"

// Creator authors a body. It is the bootstrap — the one place the design requires
// something be handed to it.
//
// OPEN: DOES THE POP RECURSE? If the creator's fix also fails to compile, does the
// creator's creator pop? Everything else in this model got a floor — leaves,
// primitives, hosts, tools. This needs one.
type Creator interface {
	Author(need string) (syntax string, err error)
}

// Registry is what creators EXIST on this runtime. Same relationship as the compiler
// registry: installed is not permitted, and a chain whose Horizon contains no creator
// cannot author new behaviour — it can only compose what already exists.
type Registry interface {
	Open(ids.Address) (Creator, error)
}
