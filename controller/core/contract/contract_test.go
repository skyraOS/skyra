package contract

import (
	"strings"
	"testing"
)

var echoIn = &Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*Schema{
	"text":  {Type: "string"},
	"probe": {Type: "string"},
	"tags":  {Type: "array", Items: &Schema{Type: "string"}},
}}

func TestFits(t *testing.T) {
	for _, c := range []struct {
		s    *Schema
		in   string
		want string // "" fits; otherwise a substring of the error
	}{
		{echoIn, `{"text":"hi"}`, ""},
		{echoIn, `{"text":"hi","probe":"/x","tags":["a","b"]}`, ""},
		{echoIn, `{"probe":"/x"}`, `missing required field "text"`},
		{echoIn, `{"text":"hi","extra":1}`, `unknown field "extra"`},
		{echoIn, `{"text":7}`, "want a string"},
		{echoIn, `{"text":"hi","tags":["a",1]}`, "[1]: want a string"},
		{echoIn, `"hi"`, "want an object"},
		{&Schema{Type: "string"}, `"free text"`, ""},
		{&Schema{Type: "number"}, `12.50`, ""},
		{&Schema{Type: "bool"}, `true`, ""},
		{&Schema{Type: "null"}, `null`, ""},
		{&Schema{Type: "null"}, `0`, "want null"},
		{nil, `"x"`, "no contract"},
	} {
		v, err := Decode([]byte(c.in))
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		err = Fits(c.s, v)
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: want fit, got %v", c.in, err)
		case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
			t.Errorf("%s: want error %q, got %v", c.in, c.want, err)
		}
	}
}

func TestDecodeIsStrict(t *testing.T) {
	for _, in := range []string{``, `{"a":1} {"b":2}`, `{"a":`, `not json`} {
		if _, err := Decode([]byte(in)); err == nil {
			t.Errorf("%q: decoded", in)
		}
	}
}

func TestCanonical(t *testing.T) {
	a, _ := Decode([]byte(`{"b":2,"a":1.0}`))
	b, _ := Decode([]byte(`{ "a": 1.0, "b": 2 }`))
	if string(Canonical(a)) != string(Canonical(b)) || string(Canonical(a)) != `{"a":1.0,"b":2}` {
		t.Fatalf("canonical forms differ: %s %s", Canonical(a), Canonical(b))
	}
}

func TestCheck(t *testing.T) {
	if err := echoIn.Check(); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []*Schema{
		{Type: "object", Required: []string{"x"}},
		{Type: "array"},
		{Type: "text"},
		{Type: "object", Fields: map[string]*Schema{"a": {Type: "?"}}},
	} {
		if bad.Check() == nil {
			t.Errorf("accepted malformed schema %+v", bad)
		}
	}
}
