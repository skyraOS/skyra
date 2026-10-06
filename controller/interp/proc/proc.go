// Package proc is the interpreter's program (next/32, "Nodes"): procedures,
// their nodes, and the store they are kept in, the interpreter's own state.
//
// A procedure is a header, a return and a body. The body names nodes by
// hash; a node is a call and its immediate dependencies:
//
//	{"call": <target> | <binding to a target>, "in": <template>, "each": <binding to a list>, "after": [<node>, ...]}
//
// A node's id is the hash of its canonical bytes, which cover its after, so a
// procedure's body commits to the whole DAG, and two procedures that share
// calls share their nodes, as git shares trees. There is no code in a
// procedure, only wiring: which calls happen, and whose output feeds whose
// input.
package proc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Schema is a contract, as the kernel encodes it: field for field, so the
// canonical bytes of a contract are the same here as there.
type Schema struct {
	Title    string             `json:"title,omitempty"`
	Type     string             `json:"type"`
	Fields   map[string]*Schema `json:"fields,omitempty"`
	Required []string           `json:"required,omitempty"`
	Items    *Schema            `json:"items,omitempty"`
}

// Canonical is a contract's canonical bytes.
func (s *Schema) Canonical() []byte {
	b, _ := json.Marshal(s)
	return b
}

// Key is a contract's stable identity: a procedure is kept under the
// contract it implements (decision 24).
func Key(s *Schema) string {
	h := sha256.Sum256(s.Canonical())
	return hex.EncodeToString(h[:])
}

// Header is what a procedure is for: the contract it implements, the actor
// it was remembered from (the fallback outside its refinement, H7), and its
// refinement.
type Header struct {
	Contract   *Schema     `json:"contract"`
	Original   string      `json:"original,omitempty"`
	Refinement *Refinement `json:"refinement,omitempty"`
}

// Refinement is the inputs and returns a compiled body is valid for. Until
// the model writes refinements (H9), it is the base types themselves
// (decision 10): the input's, and every resumed edge's.
//
// Where the model wrote a refinement in CEL (H9), it replaces the default:
// Header for the input, On for a node's edge.
type Refinement struct {
	Input      string            `json:"input,omitempty"`       // the input's base type hash
	Edges      map[string]string `json:"edges,omitempty"`       // node → the base type hash of what it returned
	InputShown string            `json:"input_shown,omitempty"` // the same, as a model reads them: never a hash
	EdgesShown map[string]string `json:"edges_shown,omitempty"`
	Header     string            `json:"header,omitempty"` // CEL over the input
	On         map[string]string `json:"on,omitempty"`     // node → CEL over the returns so far
}

// Procedure is a header, a return and a body: the nodes the procedure runs,
// by hash, sinks first. Every node any of them depends on runs too.
type Procedure struct {
	Header Header            `json:"header"`
	Return json.RawMessage   `json:"return"`
	Body   []string          `json:"body"`
	Names  map[string]string `json:"names,omitempty"` // node → the name it was written under: how a model refers to it
}

// Node is a call and its immediate dependencies.
type Node struct {
	Call  json.RawMessage `json:"call"`
	In    json.RawMessage `json:"in"`
	Each  json.RawMessage `json:"each,omitempty"`
	After []string        `json:"after"`
}

// ID is a node's hash: its canonical bytes, its after included.
func (n Node) ID() string {
	b, _ := json.Marshal(canonicalNode(n))
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func canonicalNode(n Node) Node {
	c := Node{Call: canon(n.Call), In: canon(n.In), Each: canon(n.Each), After: slices.Clone(n.After)}
	if c.After == nil {
		c.After = []string{}
	}
	slices.Sort(c.After)
	return c
}

// canon re-encodes JSON with object keys sorted and numbers as written.
func canon(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var v any
	if d.Decode(&v) != nil {
		return b
	}
	out, _ := json.Marshal(v)
	return out
}

// Store is the interpreter's own state, under SKYRA_STATE: procedures by the
// contract they implement, nodes by hash, and where each running frame is.
//
//	procedures/<key>   a procedure, under its contract's key
//	nodes/<id>         a node, under its hash
//	frames/<frame>     a running frame's position: the procedure, and every round it emitted
type Store struct{ Dir string }

// Position is where a running frame is: the procedure it runs, and the
// calls of every round it emitted, in emission order.
type Position struct {
	Procedure string   `json:"procedure"`
	Rounds    [][]Call `json:"rounds"`
	Fallback  bool     `json:"fallback,omitempty"` // a refinement refused: the original actor was invoked, last
}

// Call is one emitted instruction: a node, and for each, its element.
type Call struct {
	Node  string `json:"node"`
	Index int    `json:"index"` // -1: not an each
}

func (s Store) path(kind, name string) (string, error) {
	if !hexName(name) {
		return "", fmt.Errorf("%q is not a name in the store", name)
	}
	return filepath.Join(s.Dir, kind, name), nil
}

func hexName(n string) bool {
	if len(n) != 64 {
		return false
	}
	_, err := hex.DecodeString(n)
	return err == nil
}

func (s Store) write(kind, name string, v any) error {
	return s.publish(kind, name, v, kind == "nodes" || kind == "procedures")
}

// publish flushes bytes before publishing, then flushes the directory. Unique
// temporary names let independent procedures safely share immutable nodes.
func (s Store) publish(kind, name string, v any, immutable bool) error {
	p, err := s.path(kind, name)
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if immutable {
		err = os.Link(tmp, p)
		if errors.Is(err, os.ErrExist) {
			if kind != "nodes" {
				return fmt.Errorf("%s already has a procedure: one representation per contract", name)
			}
			old, readErr := os.ReadFile(p)
			if readErr != nil {
				return readErr
			}
			if !bytes.Equal(old, b) {
				return fmt.Errorf("node %s conflicts with stored bytes", name)
			}
			err = nil
		}
	} else {
		err = os.Rename(tmp, p)
	}
	if err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return err
	}
	// Persist newly created kind directories as well as their entries.
	parent, err := os.Open(s.Dir)
	if err != nil {
		return err
	}
	defer parent.Close()
	return parent.Sync()
}

func (s Store) read(kind, name string, v any) error {
	p, err := s.path(kind, name)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// ErrNone is a store with no procedure for a contract.
var ErrNone = errors.New("no procedure for this contract")

// Procedure reads the procedure for a contract, and every node it runs.
func (s Store) Procedure(c *Schema) (Procedure, map[string]Node, error) {
	var p Procedure
	if err := s.read("procedures", Key(c), &p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return p, nil, ErrNone
		}
		return p, nil, err
	}
	nodes := map[string]Node{}
	var load func(id string) error
	load = func(id string) error {
		if _, ok := nodes[id]; ok {
			return nil
		}
		var n Node
		if err := s.read("nodes", id, &n); err != nil {
			return fmt.Errorf("node %s: %w", id[:8], err)
		}
		if n.ID() != id {
			return fmt.Errorf("node %s does not hash to its name", id[:8])
		}
		nodes[id] = n
		for _, a := range n.After {
			if err := load(a); err != nil {
				return err
			}
		}
		return nil
	}
	for _, id := range p.Body {
		if err := load(id); err != nil {
			return p, nil, err
		}
	}
	return p, nodes, nil
}

// Has reports whether a procedure exists for a contract.
func (s Store) Has(c *Schema) bool {
	p, err := s.path("procedures", Key(c))
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

// Put stores a procedure and its nodes. A contract has one representation
// (decision 28): a second is refused.
func (s Store) Put(p Procedure, nodes map[string]Node) error {
	if s.Has(p.Header.Contract) {
		return fmt.Errorf("%s already has a procedure: one representation per contract", title(p.Header.Contract))
	}
	for id, n := range nodes {
		if n.ID() != id {
			return fmt.Errorf("node %s does not hash to its name", id[:8])
		}
		if err := s.write("nodes", id, n); err != nil {
			return err
		}
	}
	return s.write("procedures", Key(p.Header.Contract), p)
}

// SetRefinement replaces the refinement of a stored procedure: the model
// wrote it at remember, after the compile and before the contract joined any
// reach (H9).
func (s Store) SetRefinement(c *Schema, r *Refinement) error {
	var p Procedure
	if err := s.read("procedures", Key(c), &p); err != nil {
		return err
	}
	p.Header.Refinement = r
	return s.publish("procedures", Key(c), p, false)
}

// Position reads where a frame is; a frame not begun is at no round.
func (s Store) Position(frame string) (Position, error) {
	var p Position
	err := s.read("frames", frame, &p)
	if errors.Is(err, os.ErrNotExist) {
		return Position{}, nil
	}
	return p, err
}

// SetPosition records where a frame is: written before each emission.
func (s Store) SetPosition(frame string, p Position) error { return s.write("frames", frame, p) }

func title(c *Schema) string {
	if c != nil && c.Title != "" {
		return c.Title
	}
	return "the contract"
}

// Remembering is a remember frame's work after it compiled its strand: the
// contracts it stored procedures for, in the order it extends itself with
// them (bottom up, the requested one last).
//
// Before it extends itself, it asks the actor each procedure was remembered
// from to write its refinements (H9), when that actor holds a refine
// contract: Refining is the one being asked about, Attempts how often, and
// Refused why the last answer was refused, shown with the next ask.
type Remembering struct {
	Contracts []*Schema         `json:"contracts"`
	Originals []string          `json:"originals"` // the actor each was remembered from
	Evidence  []json.RawMessage `json:"evidence"`  // what each was compiled from, as names and presentations
	Refining  int               `json:"refining"`
	Attempts  int               `json:"attempts"`
	Refused   string            `json:"refused,omitempty"`
	Asked     string            `json:"asked,omitempty"` // what the last emission asked for: refine or extend
	Extended  int               `json:"extended"`
}

// Remembering reads a remember frame's work, if it has compiled.
func (s Store) Remembering(frame string) (Remembering, bool, error) {
	var r Remembering
	err := s.read("remembering", frame, &r)
	if errors.Is(err, os.ErrNotExist) {
		return r, false, nil
	}
	return r, err == nil, err
}

// SetRemembering records a remember frame's work.
func (s Store) SetRemembering(frame string, r Remembering) error {
	return s.write("remembering", frame, r)
}
