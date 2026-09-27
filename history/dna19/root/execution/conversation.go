package execution

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	p "skyra/history/dna19/protocol"
	"skyra/history/dna19/root/ports"
	"strings"
)

// conversation is a minted handle. Only its revocation is stored: whether it
// is busy or started is derived from the frames that used it.
type conversation struct {
	ref     p.ParentRef
	binding *ports.Binding
	live    bool
}

// handleUse records a handle an operation was handed and what it required.
type handleUse struct {
	sequence p.FrameID
	need     string
}

// RandomToken returns 32 random bytes as hex, for parent keys and namespaces.
func RandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func signParent(r p.ParentRef, key []byte) string {
	r.Token = ""
	raw, _ := json.Marshal(r)
	h := hmac.New(sha256.New, key)
	h.Write(raw)
	return hex.EncodeToString(h.Sum(nil))
}

// ProvisionParentAuthority gives trusted host provisioning the key and namespace
// that parent references are signed with. It is set before publication.
func (e *ExecutionContext) ProvisionParentAuthority(key []byte, namespace string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.bookkeeper.key = key
	e.bookkeeper.namespace = namespace
}

// MintParent is trusted host provisioning; caller-supplied bare IDs never mint.
func (e *ExecutionContext) MintParent(target p.ExecutableID) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, err := e.resolve(target)
	if err != nil || !e.live {
		return "", fmt.Errorf("inactive parent authority")
	}
	b := e.bookkeeper
	if len(b.key) == 0 {
		return "", fmt.Errorf("no parent authority provisioned")
	}
	if b.parents == nil {
		b.parents = map[p.FrameID]*conversation{}
	}
	b.next++
	r := p.ParentRef{Namespace: b.namespace, Owner: e.identity, Sequence: b.next, Generation: m.endpoint.Binding().Generation()}
	r.Token = signParent(r, b.key)
	b.parents[r.Sequence] = &conversation{ref: r, binding: m.endpoint.Binding(), live: true}
	raw, _ := json.Marshal(r)
	return string(raw), nil
}
func (c ArgContract) check(args []p.Observation) error {
	if c.Profile != nil {
		return c.Profile.Check(args, false)
	}
	return p.CheckArguments(args, c.Operation)
}

// parentReference is the handshake: the presented bytes must be a handle this
// context minted for this binding, and not revoked.
func (e *ExecutionContext) parentReference(raw string, binding *ports.Binding) (*conversation, error) {
	var r p.ParentRef
	if p.StrictJSON([]byte(raw), &r) != nil {
		return nil, fmt.Errorf("invalid parent")
	}
	c := e.bookkeeper.parents[r.Sequence]
	if c == nil || !c.live || c.ref != r || c.binding != binding || r.Owner != e.identity {
		return nil, fmt.Errorf("unauthorized or retired parent")
	}
	return c, nil
}

// busyFrame derives the open frame holding a handle, or 0.
func (b *Bookkeeper) busyFrame(sequence p.FrameID) p.FrameID {
	for _, f := range b.frames {
		if f.phase == "running" || f.phase == "waiting" {
			for _, u := range f.handles {
				if u.sequence == sequence {
					return f.id
				}
			}
		}
	}
	return 0
}

// started derives whether any operation on a handle has returned.
func (b *Bookkeeper) started(sequence p.FrameID) bool {
	for _, f := range b.frames {
		if f.phase == "returned" {
			for _, u := range f.handles {
				if u.sequence == sequence {
					return true
				}
			}
		}
	}
	return false
}

// authorizeHandles checks every handle an Invoke presents against what its
// profile declares. It knows no target and no operation names: the
// requirements are contract data, and the state they read is derived.
func (e *ExecutionContext) authorizeHandles(profile *p.Profile, args []p.Observation, binding *ports.Binding) ([]handleUse, error) {
	if profile == nil {
		return nil, nil
	}
	values := map[string]string{}
	for _, a := range args {
		values[a.Key] = string(a.Bytes)
	}
	var uses []handleUse
	for _, field := range profile.Fields {
		raw, present := values[field.Key]
		if field.Encoding != "parent" || !present {
			continue
		}
		c, err := e.parentReference(raw, binding)
		if err != nil {
			return nil, err
		}
		for _, u := range uses {
			if u.sequence == c.ref.Sequence {
				return nil, fmt.Errorf("one handle presented twice")
			}
		}
		if e.bookkeeper.busyFrame(c.ref.Sequence) != 0 {
			return nil, fmt.Errorf("busy parent")
		}
		if started := e.bookkeeper.started(c.ref.Sequence); started == (field.Handle == "fresh") {
			return nil, fmt.Errorf("parent lifecycle mismatch")
		}
		uses = append(uses, handleUse{sequence: c.ref.Sequence, need: field.Handle})
	}
	return uses, nil
}

// settleHandles applies what a settled operation does to its handles: a
// Return on an "end" handle, or a failure the profile does not retain,
// revokes it. A borrowed handle is never revoked by the borrower.
func (b *Bookkeeper) settleHandles(f *frame, fault *p.Fault) {
	for _, u := range f.handles {
		if u.need == "borrowed" {
			continue
		}
		retire := fault == nil && u.need == "end"
		if fault != nil {
			retire = true
			for _, code := range f.profile.Retains {
				if strings.HasPrefix(fault.Message, code+":") {
					retire = false
				}
			}
		}
		if retire {
			b.parents[u.sequence].live = false
		}
	}
}
func (b *Bookkeeper) invalidateDescendants(id p.FrameID) {
	f := b.frames[id]
	if f == nil {
		return
	}
	if f.waiting != 0 {
		child := b.frames[f.waiting]
		b.invalidateDescendants(f.waiting)
		if child != nil {
			child.phase = "faulted"
		}
		f.waiting = 0
	}
}

// Capabilities returns detached contracts for this live binding, not Memory's wider catalog.
func (e *ExecutionContext) Capabilities(id p.ExecutableID) ([]p.Profile, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	m, err := e.resolve(id)
	if err != nil {
		return nil, err
	}
	var result []p.Profile
	for _, c := range m.executable.Contracts {
		if c.Profile != nil && !c.Profile.Continuation {
			result = append(result, *c.Profile)
		}
	}
	raw, _ := json.Marshal(result)
	var detached []p.Profile
	_ = json.Unmarshal(raw, &detached)
	return detached, nil
}

// Definition builds the Codex executable from its descriptor, narrowed to grants.
func Definition(d p.Descriptor, grants map[string]bool) Executable {
	e := Executable{ID: p.Codex}
	for _, profile := range d.Profiles {
		if grants == nil || grants[profile.Operation] {
			copy := profile
			e.Contracts = append(e.Contracts, ArgContract{Operation: copy.Operation, Profile: &copy})
		}
	}
	return e
}
