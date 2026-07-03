package userdna

import (
	"fmt"
	"sync"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// UserCell is the user's being in the world. No inference, no cognitive
// pipeline — the human provides those. The only job: record every impulse
// that arrives (inbound, from the world) and every impulse that leaves
// (outbound, authored by the user) with a timestamp so the full interaction
// history is preserved on both sides of the exchange.
//
// Inbound = what the world delivers (skyra's engage, world events).
// Outbound = what the user sends (utterances, commands).
//
// These two logs are the foundation any user-modeling layer reads from later.
type UserCell struct {
	id    string
	slice skyra.Slice

	mu        sync.Mutex
	inbound   []ImpulseRecord
	outbound  []ImpulseRecord
	entities  map[string]skyra.Space // known entities — seeded at boot, grown from inbound senders
	OnInbound func(skyra.Impulse)    // optional hook — called after each inbound impulse is recorded
}

// ImpulseRecord is a timestamped snapshot of one impulse crossing the user's
// boundary. Timestamp is the slice count at receipt/send time (same clock as
// the rest of the system).
type ImpulseRecord struct {
	Timestamp int
	Impulse   skyra.Impulse
}

func NewUserCell(id string, newSlice func(id string) skyra.Slice) *UserCell {
	u := &UserCell{
		id:       id,
		entities: map[string]skyra.Space{},
	}
	if newSlice != nil {
		u.slice = newSlice(id + ":user")
	}
	return u
}

// Seed adds a known entity to michael's world model at boot — before any
// interaction has happened. Use this for beings that share the world with the
// user from the start (e.g. skyra placed at an adjacent point).
func (u *UserCell) Seed(entity skyra.Space) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.entities[entity.ID()] = entity
}

// KnownEntities returns a snapshot of all entities michael is aware of.
// Safe for concurrent reads (TUI polls this to build the engage picker).
func (u *UserCell) KnownEntities() map[string]skyra.Space {
	u.mu.Lock()
	defer u.mu.Unlock()
	out := make(map[string]skyra.Space, len(u.entities))
	for k, v := range u.entities {
		out[k] = v
	}
	return out
}

func (u *UserCell) ID() string                            { return u.id }
func (u *UserCell) Slice() skyra.Slice                    { return u.slice }
func (u *UserCell) NewActivation() skyra.ActivationRecord { return nil }
func (u *UserCell) Content() any                          { return nil }
func (u *UserCell) Children() map[string]skyra.Space      { return nil }
func (u *UserCell) Agents() map[string]skyra.Space        { return nil }
func (u *UserCell) Express(origin skyra.Space, impulse any) any { return nil }

// Observe records every impulse the world delivers to the user.
func (u *UserCell) Observe(origin skyra.Space, impulse any) {
	imp, ok := impulse.(skyra.Impulse)
	if !ok {
		if ptr, ok := impulse.(*skyra.Impulse); ok && ptr != nil {
			imp = *ptr
		} else {
			return
		}
	}
	u.mu.Lock()
	ts := u.count()
	u.inbound = append(u.inbound, ImpulseRecord{Timestamp: ts, Impulse: imp})
	src := ""
	if imp.Source != nil && imp.Source.SpaceID != "" {
		src = imp.Source.SpaceID
		// record the sender as a known entity if we haven't seen them before
		if _, known := u.entities[src]; !known {
			if origin != nil && origin.ID() == src {
				u.entities[src] = origin
			}
		}
	}
	cb := u.OnInbound
	u.mu.Unlock()

	debug.Being("michael", "observe", fmt.Sprintf("[t%d] from=%s primitive=%s payload=%v", ts, src, imp.Primitive, imp.Payload))
	if cb != nil && !imp.Return {
		cb(imp)
	}
}

// Send records an outbound impulse authored by the user (call this from the
// REPL / terminal handler before firing the impulse into the world).
func (u *UserCell) Send(imp skyra.Impulse) {
	u.mu.Lock()
	defer u.mu.Unlock()
	ts := u.count()
	u.outbound = append(u.outbound, ImpulseRecord{Timestamp: ts, Impulse: imp})
	debug.Being("michael", "express", fmt.Sprintf("[t%d] primitive=%s payload=%v", ts, imp.Primitive, imp.Payload))
}

// Inbound returns a snapshot of all impulses received so far.
func (u *UserCell) Inbound() []ImpulseRecord {
	u.mu.Lock()
	defer u.mu.Unlock()
	out := make([]ImpulseRecord, len(u.inbound))
	copy(out, u.inbound)
	return out
}

// Outbound returns a snapshot of all impulses sent so far.
func (u *UserCell) Outbound() []ImpulseRecord {
	u.mu.Lock()
	defer u.mu.Unlock()
	out := make([]ImpulseRecord, len(u.outbound))
	copy(out, u.outbound)
	return out
}

func (u *UserCell) count() int {
	if u.slice != nil {
		return u.slice.Count()
	}
	return len(u.inbound) + len(u.outbound)
}
