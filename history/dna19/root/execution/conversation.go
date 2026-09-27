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
)

type conversation struct {
	ref           p.ParentRef
	binding       *ports.Binding
	live, started bool
	busy          p.FrameID
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
func (e *ExecutionContext) parentReference(raw string, binding *ports.Binding) (*conversation, error) {
	var r p.ParentRef
	if p.StrictJSON([]byte(raw), &r) != nil {
		return nil, fmt.Errorf("invalid parent")
	}
	c := e.bookkeeper.parents[r.Sequence]
	if c == nil || !c.live || c.ref != r || c.binding != binding || r.Owner != e.identity {
		return nil, fmt.Errorf("unauthorized or retired parent")
	}
	if c.busy != 0 {
		return nil, fmt.Errorf("busy parent")
	}
	return c, nil
}
func (e *ExecutionContext) authorizeConversation(m p.Move, entry *mapping) (p.FrameID, error) {
	if m.Target != p.Codex {
		return 0, nil
	}
	values := map[string]string{}
	for _, a := range m.Invoke.Args {
		values[a.Key] = string(a.Bytes)
	}
	op := values["operation"]
	if op == "connect" || op == "health" {
		return 0, nil
	}
	c, err := e.parentReference(values["parent"], entry.endpoint.Binding())
	if err != nil {
		return 0, err
	}
	fresh := op == "start-thread" || op == "fork-thread"
	if fresh == c.started {
		return 0, fmt.Errorf("parent lifecycle mismatch")
	}
	if op == "fork-thread" {
		source, err := e.parentReference(values["source"], entry.endpoint.Binding())
		if err != nil || source == c || !source.started {
			return 0, fmt.Errorf("unauthorized fork source")
		}
	}
	return c.ref.Sequence, nil
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
