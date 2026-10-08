package machine

import (
	"encoding/json"
	"errors"
	"fmt"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// accept is the child's half of Release, for every state: the Horizon it
// looked up must be for this commit, must be exactly what this commit derives
// from the Horizon it was emitted under, and must come with its whole frame
// stack, or the instance stops. Authority is derived, never taken on trust
// (next/24): the id is only a handle.
func accept(s Instance, e Released) error {
	if e.Self.Stamp != s.Stamp {
		return errors.New("release is not for this commit")
	}
	want, err := Narrow(s.Under, s.R, s.Stamp)
	if err != nil {
		return fmt.Errorf("release cannot be derived: %w", err)
	}
	if a, b := encode(want), encode(e.Self); a != b {
		// One instruction of several: an invoke is forked, a fault holds a
		// slot of its own.
		if alt, _ := Fanned(s.Under, s.R, s.Stamp); encode(alt) != b {
			return errors.New("release is not what this commit derives")
		}
	}
	if e.Frames != e.Self.Depth() {
		return fmt.Errorf("release carries %d frames, its Horizon needs %d", e.Frames, e.Self.Depth())
	}
	return nil
}

func encode(h horizon.Horizon) string {
	b, _ := json.Marshal(h)
	return string(b)
}

// issueRelease is the parent's half of Release: check the successor's stamp,
// then compute its Horizon.
func issueRelease(parent horizon.Horizon, emitted []byte, parentStamp string, e Stamped, fanned bool) (horizon.Horizon, error) {
	if err := checkSuccessor(e.R, emitted, parentStamp); err != nil {
		return horizon.Horizon{}, fmt.Errorf("stamp failed verification: %w", err)
	}
	if fanned {
		return Fanned(parent, *e.R, e.Stamp)
	}
	return Narrow(parent, *e.R, e.Stamp)
}

// checkSuccessor is the parent's check of a stamp (D3). The shell already
// refused any record that does not hash to its stamp. This checks that the
// successor read exactly the bytes that were emitted and linked to this frame.
func checkSuccessor(r *record.Record, emitted []byte, parentStamp string) error {
	switch {
	case r == nil:
		return errors.New("stamp names no valid record")
	case r.Input != record.Sum(emitted):
		return errors.New("record does not cover the bytes that were emitted")
	case r.Predecessor != parentStamp:
		return errors.New("record does not link to this parent")
	}
	return nil
}

// Narrow computes the child's Horizon from the parent's and the child's
// record. Every call shares its caller's Horizon (next/24): the child runs
// with the same reach, references and grants. Nothing about the callee
// narrows them; a narrowing is requested, through the room (next/26):
//
//	invoke system, narrow        the caller's Horizon, and the waiting caller is
//	                             recorded as holding a new reference: the grant
//	                             its payload derives, named system.ref.<n>
//	invoke system, extend-reach  the caller's Horizon, and the waiting caller's
//	                             reach entry for itself holds one more contract,
//	                             in every frame below that shares its Horizon
//	invoke a reference           the actor, under the narrowed Horizon it names
//
// An invoke opens a new frame for the child, waited on by the parent as it is
// now. A return or fault runs nothing: it discharges the parent's slot.
func Narrow(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	switch r.State {
	case record.Return, record.Fault:
		return horizon.Horizon{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: parent.Frame, Resume: parent.Resume}, nil
	case record.Invoke:
	default:
		return horizon.Horizon{}, fmt.Errorf("no Horizon for a record in a state of %s", r.State)
	}
	waiter := parent
	if g, isRef := parent.Refs[r.To]; isRef {
		if _, named := parent.Reach[r.To]; !named {
			s, ok := parent.Reach[g.Actor]
			if !ok {
				return horizon.Horizon{}, fmt.Errorf("reference %q names %q, which is not in the parent's reach", r.To, g.Actor)
			}
			// A Horizon of its own: created by this invoke, so named by it.
			return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: kept(parent.Reach, g.Reach), Authority: g.Authority,
				Name: g.Actor, Origin: stamp, Frame: stamp, Resume: &waiter}, nil
		}
	}
	s, ok := parent.Reach[r.To]
	if !ok {
		return horizon.Horizon{}, fmt.Errorf("%q is not in the parent's reach", r.To)
	}
	if s.System == "horizon" {
		switch name, args := syscall(r.Payload); name {
		case Narrowing:
			g, err := narrowing(parent, args)
			if err != nil {
				return horizon.Horizon{}, err
			}
			refs := map[string]horizon.Grant{refName(parent.Refs): g}
			for name, held := range parent.Refs {
				refs[name] = held
			}
			waiter.Refs = refs
		case Extending:
			c, err := extension(parent, args)
			if err != nil {
				return horizon.Horizon{}, err
			}
			waiter = extended(parent, parent.Origin, parent.Name, c)
		}
	}
	return horizon.Horizon{Stamp: stamp, Surface: &s, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority,
		Name: r.To, Origin: parent.Origin, Forked: parent.Forked, Frame: stamp, Resume: &waiter}, nil
}

// Fanned is Narrow for a successor that took one instruction of several
// (next/32, H2). An invoke opens its frame as always. A fault holds a slot of
// its own in the emitter's round and discharges it, so its siblings still run
// and the emitter resumes with the fault among its steps; alone, the same
// fault discharges the emitter's own slot.
func Fanned(parent horizon.Horizon, r record.Record, stamp string) (horizon.Horizon, error) {
	if r.State != record.Fault {
		h, err := Narrow(parent, r, stamp)
		if r.State == record.Invoke {
			h.Forked = true // its siblings share this Horizon, each on its own copy
		}
		return h, err
	}
	waiter := parent
	return horizon.Horizon{Stamp: stamp, Reach: parent.Reach, Refs: parent.Refs, Authority: parent.Authority, Frame: stamp, Resume: &waiter}, nil
}
