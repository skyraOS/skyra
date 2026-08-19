// Package compiler holds the compiler slot.
//
//	THE CAUSAL SYNTAX LIVES AT AN ADDRESS, AND ALL IT IS IS A PARSER THAT CALLS THE
//	KERNEL'S METHODS.
//
// Which is how every language works — source compiles to syscalls.
//
//	COMPILERS OWN   their syntax. NOTHING ELSE.
//	THE KERNEL OWNS the methods, and scope · topology · causal order · routing · closure.
//
// A Compiler interface carrying Decide/Recognize/Remember was proposed twice and
// reverted twice. All three are KERNEL methods; a compiler holding one would be
// extending causal structure from outside the kernel, which is reserved. The failure
// both times was treating a type check as a predicate rather than as a causal event.
//
// Compilers compose, one per syntax — the runtime's, python's, further syntaxes as
// they earn a place. Each compiles its own syntax down to the same kernel calls, which
// is why there is no bytecode: THE KERNEL PROTOCOL IS THE INSTRUCTION SET.
//
// The selector is DETERMINED, not learned: the frame supplies the syntax, the syntax
// determines the compiler, the creator fills in a body. A parent that could choose it
// could bind a python body to an LLM interpreter, which must stay inexpressible.
package compiler

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/kernel"
)

// Compiler parses one syntax and drives it against the kernel.
//
// IT IS HANDED THE KERNEL. It holds no protocol methods of its own; it calls them.
//
// CALLER-DRIVEN, AND THIS IS THE OPEN PART. A compiler at an address may be a language
// nobody here controls — python's own interpreter drives, python's own stack holds the
// position, and it BLOCKS on a call, because a parent cannot continue without its
// child's return. That is why a kernel-driven Next(State) → Step transition function
// cannot describe the general case, and why the cost lands as held context for the
// life of the chain.
//
// UNRESOLVED, and both live in this signature:
//   - whether Run blocks for the whole life of the possibility, or is re-entered
//   - how a blocking walker emits a CONCURRENT GROUP. A blocking call is sequential by
//     construction, while emitted member order is meaningful and completion order is
//     not. Source layout alone must not make scheduler timing semantic.
type Compiler interface {
	// Run interprets the syntax of one branch, calling k on the possibility's behalf.
	Run(k kernel.Kernel, p ids.PossibilityID, syntax string) error
}

// Registry is what compilers EXIST on this runtime. INSTALLED, not permitted —
// a compiler can be present and unreachable, because which ones a chain MAY USE is
// Horizon's question. Registering one grants it to nobody.
//
// CONSTRUCTOR-KEYED, never instance-keyed: an instance could accumulate state between
// invocations, and then two invocations that look identical would give different
// answers and SAME QUESTION BECOMES A LIE.
type Registry interface {
	Open(ids.Address) (Compiler, error)
}
