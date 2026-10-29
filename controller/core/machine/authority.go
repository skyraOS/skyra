package machine

import (
	"encoding/json"
	"errors"
	"fmt"

	"skyra/controller/core/boundary"
	"skyra/controller/core/record"
)

// accept is the child's half of Release, for every state: the Boundary it
// looked up must be for this commit, must be exactly what this commit derives
// from the Boundary it was emitted under, and must come with its whole frame
// stack, or the instance stops. Authority is derived, never taken on trust
// (next/24): the id is only a handle.
func accept(s Instance, released Object) (boundary.Boundary, error) {
	self, err := BoundaryOf(released)
	if err != nil {
		return boundary.Boundary{}, fmt.Errorf("released Boundary: %w", err)
	}
	if self.Stamp != s.Stamp() {
		return boundary.Boundary{}, errors.New("release is not for this commit")
	}
	want, err := Narrow(s.Under, s.R, s.Stamp())
	if err != nil {
		return boundary.Boundary{}, fmt.Errorf("release cannot be derived: %w", err)
	}
	if a, b := encode(want), encode(self); a != b {
		// One instruction of several: an invoke is forked, a fault holds a
		// slot of its own.
		if alt, _ := Fanned(s.Under, s.R, s.Stamp()); encode(alt) != b {
			return boundary.Boundary{}, errors.New("release is not what this commit derives")
		}
	}
	if s.Arrived != self.Depth() {
		return boundary.Boundary{}, fmt.Errorf("release carries %d frames, its Boundary needs %d", s.Arrived, self.Depth())
	}
	return self, nil
}

func encode(h boundary.Boundary) string {
	b, _ := json.Marshal(h)
	return string(b)
}

// issueRelease is the parent's half of Release: check the successor's stamp,
// then compute its Boundary.
func issueRelease(parent boundary.Boundary, emitted []byte, parentStamp string, o Object, fanned bool) (record.Record, boundary.Boundary, error) {
	r, err := checkSuccessor(o, emitted, parentStamp)
	if err != nil {
		return r, boundary.Boundary{}, fmt.Errorf("stamp failed verification: %w", err)
	}
	if fanned {
		h, err := Fanned(parent, r, o.Key)
		return r, h, err
	}
	h, err := Narrow(parent, r, o.Key)
	return r, h, err
}

// checkSuccessor is the parent's check of a stamp (D3): the record read under
// it verifies, the successor read exactly the bytes that were emitted, and it
// linked them to this frame.
func checkSuccessor(o Object, emitted []byte, parentStamp string) (record.Record, error) {
	if !o.Found() {
		return record.Record{}, errors.New("stamp names no valid record")
	}
	r, err := RecordOf(o)
	switch {
	case err != nil:
		return r, err
	case r.Input != record.Sum(emitted):
		return r, errors.New("record does not cover the bytes that were emitted")
	case r.Predecessor != parentStamp:
		return r, errors.New("record does not link to this parent")
	}
	return r, nil
}

// Narrow computes the child's Boundary from the parent's and the child's
// record. Every call shares its caller's Boundary (next/24): the child runs
// with the same reach, references and grants. Nothing about the callee
// narrows them; a narrowing is requested, through system.narrow (next/26):
//
//	invoke system, narrow        the caller's Boundary, and the waiting caller is
//	                             recorded as holding a new reference: the grant
//	                             its payload derives, named system.ref.<n>
//	invoke system, export  the caller's Boundary, and the waiting caller's
//	                             reach entry for itself holds one more contract,
//	                             in every frame below that shares its Boundary
//	invoke a reference           the actor, under the narrowed Boundary it names
//
// An invoke opens a new frame for the child, waited on by the parent as it is
// now. A return or fault runs nothing: it discharges the parent's slot.
func Narrow(parent boundary.Boundary, r record.Record, stamp string) (boundary.Boundary, error) {
	switch r.State {
	case record.Return, record.Fault:
		return boundary.Boundary{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: parent.Frame, Resume: parent.Resume}, nil
	case record.Invoke:
	default:
		return boundary.Boundary{}, fmt.Errorf("no Boundary for a record in a state of %s", r.State)
	}
	waiter := parent
	if g, isRef := parent.Refs[r.To]; isRef {
		if _, named := parent.Reach[r.To]; !named {
			s, ok := parent.Reach[g.Actor]
			if !ok {
				return boundary.Boundary{}, fmt.Errorf("reference %q names %q, which is not in the parent's reach", r.To, g.Actor)
			}
			// A Boundary of its own: created by this invoke, so named by it.
			return boundary.Boundary{Stamp: stamp, Surface: &s, Reach: kept(parent.Reach, g.Reach), Authority: g.Authority, Held: g.Held,
				Name: g.Actor, Origin: stamp, Frame: stamp, Resume: &waiter}, nil
		}
	}
	s, ok := parent.Reach[r.To]
	if !ok {
		return boundary.Boundary{}, fmt.Errorf("%q is not in the parent's reach", r.To)
	}
	if s.System == "boundary" {
		switch name, args := syscall(r.Payload); name {
		case Narrowing:
			g, err := narrowing(parent, args)
			if err != nil {
				return boundary.Boundary{}, err
			}
			refs := map[string]boundary.Grant{refName(parent.Refs): g}
			for name, held := range parent.Refs {
				refs[name] = held
			}
			waiter.Refs = refs
		case Exporting:
			c, err := extension(parent, args)
			if err != nil {
				return boundary.Boundary{}, err
			}
			waiter = extended(parent, parent.Origin, parent.Name, c)
		}
	}
	return boundary.Boundary{Stamp: stamp, Surface: &s, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority,
		Held: parent.Held, Name: r.To, Origin: parent.Origin, Forked: parent.Forked, Frame: stamp, Resume: &waiter}, nil
}

// Fanned is Narrow for a successor that took one instruction of several
// (next/32, H2). An invoke opens its frame as always. A fault holds a slot of
// its own in the emitter's round and discharges it, so its siblings still run
// and the emitter resumes with the fault among its steps; alone, the same
// fault discharges the emitter's own slot.
func Fanned(parent boundary.Boundary, r record.Record, stamp string) (boundary.Boundary, error) {
	if r.State != record.Fault {
		h, err := Narrow(parent, r, stamp)
		if r.State == record.Invoke {
			h.Forked = true // its siblings share this Boundary, each on its own copy
		}
		return h, err
	}
	waiter := parent
	return boundary.Boundary{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: stamp, Resume: &waiter}, nil
}
