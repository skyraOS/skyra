package machine

import (
	"encoding/json"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// returned is a child of a round that invoked to and got payload back.
func returned(to, payload string) Child {
	return Child{Invoke: record.Record{State: record.Invoke, To: to}, Discharge: &record.Record{State: record.Return, Payload: json.RawMessage(payload)}}
}

// The actor is its identity, not its name: re-registering another program as
// echo is a new type for the same payload. One surface and one shape is one
// type, whatever the leaves.
func TestTypeOfAReturn(t *testing.T) {
	other := echoS
	other.SHA256 = "another binary"
	parent := horizon.Horizon{Reach: map[string]horizon.Surface{"echo": echoS}}
	impostor := horizon.Horizon{Reach: map[string]horizon.Surface{"echo": other}}

	a, ok := TypeOf(parent, returned("echo", `{"text":"hi"}`))
	b, _ := TypeOf(impostor, returned("echo", `{"text":"hi"}`))
	if !ok || a.Type == b.Type {
		t.Fatalf("two surfaces named echo must give two types: %+v %+v", a, b)
	}
	c, _ := TypeOf(parent, returned("echo", `{"text":"yo"}`))
	if a.Type != c.Type || a.Presentation != "echo → {text: string}" || a.Binding != "echo" {
		t.Fatalf("one surface and one shape is one type: %+v %+v", a, c)
	}
	if d, _ := TypeOf(parent, returned("echo", `"hi"`)); d.Type == a.Type {
		t.Fatal("another shape is another type")
	}

	// A reference is bound to the actor it stands for, and is that actor's type.
	parent.Refs = map[string]horizon.Grant{"system.ref.1": {Actor: "echo"}}
	r, ok := TypeOf(parent, returned("system.ref.1", `{"text":"hi"}`))
	if !ok || r.Binding != "echo" || r.Type != a.Type {
		t.Fatalf("a reference is typed as its actor: %+v", r)
	}

	// A fault has no type.
	f := returned("echo", `"x"`)
	f.Discharge.State = record.Fault
	if _, ok := TypeOf(parent, f); ok {
		t.Fatal("only a return has a type")
	}
}
