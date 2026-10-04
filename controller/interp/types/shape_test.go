package types

import (
	"reflect"
	"testing"
)

func decoded(t *testing.T, s string) any {
	t.Helper()
	v, err := Decode([]byte(s))
	if err != nil {
		t.Fatalf("%s: %v", s, err)
	}
	return v
}

// The encoding is injective: values that differ in kind, in how keys split,
// or in nesting never share a hash.
func TestShapeAndContentAreInjective(t *testing.T) {
	for _, pair := range [][2]string{
		{`{"ab":1}`, `{"a":1,"b":1}`},
		{`{"ab":"x"}`, `{"a":"x","b":"x"}`},
		{`"1"`, `1`},
		{`[]`, `{}`},
		{`[[]]`, `[]`},
		{`{"":1}`, `{}`},
		{`{"":1}`, `{"a":1}`},
		{`null`, `false`},
		{`[1,"a"]`, `[1]`},
	} {
		a, b := decoded(t, pair[0]), decoded(t, pair[1])
		if Shape(a) == Shape(b) {
			t.Errorf("shape(%s) == shape(%s)", pair[0], pair[1])
		}
		if Content(a) == Content(b) {
			t.Errorf("content(%s) == content(%s)", pair[0], pair[1])
		}
	}
	// Content separates what shape does not.
	for _, pair := range [][2]string{{`"ab"`, `"a"`}, {`1`, `1.0`}, {`true`, `false`}, {`[1,2]`, `[2,1]`}} {
		a, b := decoded(t, pair[0]), decoded(t, pair[1])
		if Shape(a) != Shape(b) || Content(a) == Content(b) {
			t.Errorf("%s vs %s: want one shape and two contents", pair[0], pair[1])
		}
	}
	if Shape(decoded(t, `"x"`)) == Content(decoded(t, `null`)) {
		t.Error("shape and content roots must be domain-separated")
	}
}

func TestKeyOrderChangesNeither(t *testing.T) {
	a, b := decoded(t, `{"a":1,"b":{"c":[true],"d":null}}`), decoded(t, `{"b":{"d":null,"c":[true]},"a":1}`)
	if Shape(a) != Shape(b) || Content(a) != Content(b) {
		t.Fatal("key order must not change either hash")
	}
}

// Keys are shape; values are content. An array's shape does not depend on its
// length or order; its content does.
func TestShapeIgnoresLeafValuesAndLength(t *testing.T) {
	a := decoded(t, `{"text":"hi","media":[{"url":"a"}]}`)
	b := decoded(t, `{"text":"yo","media":[{"url":"b"},{"url":"c"},{"url":"d"}]}`)
	if Shape(a) != Shape(b) {
		t.Fatal("same keys and leaf kinds must be one shape")
	}
	if Content(a) == Content(b) {
		t.Fatal("different leaves must be two contents")
	}
	if Shape(decoded(t, `[1,"a"]`)) != Shape(decoded(t, `["b",2,3]`)) {
		t.Fatal("an array's shape is its distinct element shapes, not their order or count")
	}
	if Shape(decoded(t, `[]`)) == Shape(decoded(t, `[1]`)) {
		t.Fatal("an empty array has no element shapes")
	}
}

// Golden hashes: a change to the encoding is a test failure, not a drift.
// null, "hi" and the object were checked against an independent
// implementation of the encoding in next/30.
func TestGoldenHashes(t *testing.T) {
	for _, c := range []struct{ value, shape, content string }{
		{`null`, "7f1e59f4f2a78633f31a4535695271111bab549efa894c2a5afe22033860b383", "394c54e8f92d4742663f0faef89df53aadb384f9a3307d1350b917a9eba8000b"},
		{`"hi"`, "5c93f3896dcdf4c8f7e75743330f31a01c2dd2ea60effdc6a8636c2ff9dbcb5e", "06da6f69a5d7e31f7b3eee252ff872cf96a38a8d8dffb682bd4ee7ef9a637286"},
		{`1.0`, "03b029301590696c692f851edd21117254b757a6695ea19ae5fff24ca58ffdf5", "43b85f319f88394bef1d3d0c1d7a92d1d8b520020225e1e903eddc9067189d5b"},
		{`true`, "3045e7fa4823547d77e9cdec1a487b06f5ce6370705c11cc2e7f3000e3b1aa95", "1065f2692aa3352c474ab5573930a917a7997b331fab4d401b3c0d14b7e0c583"},
		{`{"text":"hi","media":[{"url":"a"}]}`, "adc8038ac488071740167ceec32889a81c8dbb5cf3d1b95e0e011b6c371ac674", "fc897b4ca01ed654bc7afe7cc1ebb1cda5cfeeb1c0e4546b9f8b0818c443d0c4"},
		{`[1,"a",[]]`, "c22e9752b8e03cd07050d8ace5c70ac07086ab8f6a21ded29021d76e108f4bdb", "cbfa9a9c2b8254440c86b0c221dd3bf82e2150d6044f61eae0a5dccbbcf89d8c"},
	} {
		v := decoded(t, c.value)
		if got := Shape(v); got != c.shape {
			t.Errorf("Shape(%s) = %s, want %s", c.value, got, c.shape)
		}
		if got := Content(v); got != c.content {
			t.Errorf("Content(%s) = %s, want %s", c.value, got, c.content)
		}
	}
}

func TestPresent(t *testing.T) {
	for in, want := range map[string]string{
		`{"text":"hi","media":[{"url":"a"}]}`: `{media: [{url: string}], text: string}`,
		`[1,"a",2]`:                           `[number | string]`,
		`[]`:                                  `[]`,
		`{}`:                                  `{}`,
		`{"a b":true,"":null}`:                `{"": null, "a b": bool}`,
		`"x"`:                                 `string`,
	} {
		if got := Present(decoded(t, in)); got != want {
			t.Errorf("Present(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestDiff(t *testing.T) {
	for _, c := range []struct {
		name, a, b string
		want       Delta
	}{
		{"unchanged", `{"text":"hi","n":[1,2]}`, `{"n":[1,2],"text":"hi"}`,
			Delta{Changed: []string{}, Added: []string{}, Removed: []string{}, Leaves: 3}},
		{"one leaf", `{"text":"hi","n":1}`, `{"text":"yo","n":1}`,
			Delta{Changed: []string{"$.text"}, Added: []string{}, Removed: []string{}, Leaves: 2, Ratio: 0.5}},
		{"a key added and one removed", `{"a":1,"b":2}`, `{"a":1,"c b":{"d":3}}`,
			Delta{Changed: []string{}, Added: []string{`$["c b"]`}, Removed: []string{"$.b"}, Leaves: 2}},
		{"an array element", `{"media":[{"url":"a"},{"url":"b"}]}`, `{"media":[{"url":"a"},{"url":"c"},{"url":"d"}]}`,
			Delta{Changed: []string{"$.media[1].url"}, Added: []string{"$.media[2]"}, Removed: []string{}, Leaves: 3, Ratio: 1.0 / 3}},
		{"a kind changed", `{"a":{"b":1}}`, `{"a":"x"}`,
			Delta{Changed: []string{"$.a"}, Added: []string{}, Removed: []string{}, Leaves: 1, Ratio: 1}},
		{"a root leaf", `"hi"`, `"yo"`,
			Delta{Changed: []string{"$"}, Added: []string{}, Removed: []string{}, Leaves: 1, Ratio: 1}},
		{"no leaves", `{"a":1}`, `{}`,
			Delta{Changed: []string{}, Added: []string{}, Removed: []string{"$.a"}, Leaves: 0}},
	} {
		if got := Diff(decoded(t, c.a), decoded(t, c.b)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}
