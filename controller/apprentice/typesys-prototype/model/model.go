// Package model is proposal 61's data model, as Model.lean defines it: six
// pairs, each a definition and its instantiation.
//
//	definition      instantiation
//	interface       observation
//	implementation  decision
//	reference       location
//	scope           frame
//	program         process
//	call graph      tree
//
// A definition never holds a value, and locations point from instantiation
// to definition, never back: no definition type here has a value or an
// InstLoc in it. A frame's key reads its scope and what arrived, never its
// eliminations or introductions.
//
// Shapes and values are 58's, from package tree.
package model

import "skyra/typesys-prototype/tree"

// A DefLoc is a location in definition space: memory and programs, by hash.
type DefLoc string

// An InstLoc is a location in instantiation space: the log, by hash.
type InstLoc string

// An Op is an operator. To invoke is to push, to return is to pop; each
// agent kind's operator is named for the kind.
type Op string

const (
	OpDelegate Op = "delegate"
	OpAct      Op = "act"
	OpBoot     Op = "boot"
	OpReturn   Op = "return"
)

// Pushes is direction as a function of the operator: every operator but
// return pushes.
func (o Op) Pushes() bool { return o != OpReturn }

// A Kind is a frame's kind, a prefix of its scope's type: an agent decides,
// a leaf only returns.
type Kind string

const (
	KindRoot     Kind = "root"
	KindDelegate Kind = "delegate"
	KindAct      Kind = "act"
	KindLeaf     Kind = "leaf"
)

// A Sel is what a cut divides out: a field path, or runs of lines. It is
// fixed in the template, never computed from a value.
type Sel struct {
	Lines bool
	Path  []string // a field path, when not Lines
	Runs  [][2]int // runs of lines, first to last, from 1
}

// An IfaceTag says which of the five ways an interface is built.
type IfaceTag uint8

const (
	IShape IfaceTag = iota // a value's kind
	IDef                   // a definition in memory: a tool or a model
	IOp                    // a composition, a product: an operator over its parts
	ISum                   // a sum: the alternatives one position takes
	ICut                   // a cut, an interface divided by a selector, kept undivided
)

// An Iface is an interface: an observation's type.
type Iface struct {
	Tag   IfaceTag
	Shape tree.Shape // IShape
	Def   DefLoc     // IDef
	Op    Op         // IOp
	Parts []Iface    // IOp's parts, ISum's alternatives
	Of    *Iface     // ICut: what is divided
	Sel   Sel        // ICut: by what
}

func ShapeIface(s tree.Shape) Iface { return Iface{Tag: IShape, Shape: s} }
func DefIface(d DefLoc) Iface       { return Iface{Tag: IDef, Def: d} }

// A Scope is a frame's interface, what it is created with: its kind, and the
// interfaces of what it was passed, in roles' order.
type Scope struct {
	Kind    Kind
	Entries []Iface
}

// A Lit is what an introduction constructs from nothing: a literal, which is
// code, or an exit status.
type Lit struct {
	IsExit bool
	Str    string
	Exit   int
}

// An Observation is a value in a frame, its interface, and where it is in
// the log.
type Observation struct {
	Value any
	Iface Iface
	Loc   InstLoc
}

// A Member is what a frame binds from its scope: a tool, a definition; or an
// observation. Not everything in a frame is an observation.
type Member struct {
	Tool DefLoc // when Obs is nil
	Obs  *Observation
}

// An Entry is what is appended to a frame's state after its creation: what
// arrived, which has been seen; or an elimination or an introduction, which
// are live.
type Entry struct {
	Arrived   *Observation
	Eliminate int   // the $n eliminated, when Arrived and Introduce are nil
	Sels      []Sel // what the elimination cuts
	Introduce *Lit
}

// A Frame is one activation of a scope. Its static link is its program, its
// dynamic link the frame that invoked it, and its model the definition it
// consults at startle, outside its scope and so never in its key.
type Frame struct {
	ID       InstLoc
	Program  DefLoc
	Parent   *InstLoc
	Scope    Scope
	Bindings []Member
	State    []Entry
	Model    *DefLoc
}

// A Key is what has been seen: its program's root, its kind, then the
// scope's interfaces and each arrival's, in order.
type Key struct {
	Root DefLoc
	Kind Kind
	Seen []Iface
}

// Seen is what a frame has seen: its scope, then what arrived. Eliminations
// and introductions are skipped.
func (f Frame) Seen() []Iface {
	seen := append([]Iface(nil), f.Scope.Entries...)
	for _, e := range f.State {
		if e.Arrived != nil {
			seen = append(seen, e.Arrived.Iface)
		}
	}
	return seen
}

// Key is a frame's key; rootOf gives a program's root, H(version, global).
func (f Frame) Key(rootOf func(DefLoc) DefLoc) Key {
	return Key{Root: rootOf(f.Program), Kind: f.Scope.Kind, Seen: f.Seen()}
}
