package grammar_test

import (
	"strings"
	"testing"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// The grammar's claim is that ONLY TWO FORMS PARSE, and that the incapacity is a
// PARSING property rather than an API property. These are that claim.

func TestTwoFormsAndNoThird(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string // "select" | "commit" | "refused"
	}{
		{"select one member",
			`<select><find cite=obs-0001 entity=codex>obs-0002</engram></select>`, "select"},
		{"select two members is cardinality 2",
			`<select>
			   <a cite=obs-0001 entity=codex>obs-0002</engram>
			   <b cite=obs-0001 entity=codex>obs-0003</engram>
			 </select>`, "select"},
		{"member with no arguments",
			`<select><go cite=obs-0001 entity=codex></engram></select>`, "select"},
		{"orchestrator names",
			`<commit><obs-0007></commit>`, "commit"},
		{"leaf originates",
			`<commit>{"answer": 42}</commit>`, "commit"},
		{"a fenced body is packaging, not a third form",
			"<commit>```json\n{\"answer\": 42}\n```</commit>", "commit"},
		{"narration around an emission is ignored",
			"I'll look this up first.\n\n<select><find cite=obs-0001 entity=codex>obs-0002</engram></select>\n\nDone.", "select"},

		// The refusals. Each is something an entity might plausibly emit, and
		// each must fail rather than be accommodated.
		{"prose only", "I think the answer is 42.", "refused"},
		{"a tool call", `{"tool": "search", "args": {"q": "x"}}`, "refused"},
		{"a third form", `<invoke name="find"/>`, "refused"},
		{"select with no members", `<select></select>`, "refused"},
		{"member missing cite", `<select><find entity=codex>obs-0002</engram></select>`, "refused"},
		{"member missing entity", `<select><find cite=obs-0001>obs-0002</engram></select>`, "refused"},
		{"invented attribute", `<select><find cite=obs-0001 entity=codex async=true>obs-0002</engram></select>`, "refused"},
		{"non-id argument", `<select><find cite=obs-0001 entity=codex>the need</engram></select>`, "refused"},
		{"empty commit", `<commit></commit>`, "refused"},
		{"commit of an empty object", `<commit>{}</commit>`, "refused"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			em, err := grammar.Parse(c.raw)
			switch c.want {
			case "refused":
				if err == nil {
					t.Fatalf("accepted something that is not one of the two forms: %#v", em)
				}
			case "select":
				if err != nil {
					t.Fatalf("refused a valid select: %v", err)
				}
				if _, ok := em.(grammar.Select); !ok {
					t.Fatalf("want Select, got %T", em)
				}
			case "commit":
				if err != nil {
					t.Fatalf("refused a valid commit: %v", err)
				}
				if _, ok := em.(grammar.Commit); !ok {
					t.Fatalf("want Commit, got %T", em)
				}
			}
		})
	}
}

func TestCardinalityIsTheConcurrencyDeclaration(t *testing.T) {
	em, err := grammar.Parse(`<select>
	  <a cite=obs-0001 entity=codex>obs-0002</engram>
	  <b cite=obs-0009 entity=runtime>obs-0003, obs-0004</engram>
	</select>`)
	if err != nil {
		t.Fatal(err)
	}
	s := em.(grammar.Select)
	if len(s.Members) != 2 {
		t.Fatalf("want 2 members, got %d", len(s.Members))
	}
	// cite is PER MEMBER, never per group. A group-level cite would be one cause
	// for N selections, which reintroduces sibling blindness where the design has
	// it solved, and makes only whole groups prunable.
	if s.Members[0].Cite == s.Members[1].Cite {
		t.Error("members share a cite — grouping became a shared cause")
	}
	if s.Members[0].Entity == s.Members[1].Entity {
		t.Error("entity is per member and these should differ")
	}
	if len(s.Members[1].Observations) != 2 {
		t.Errorf("want 2 arguments on the second member, got %d", len(s.Members[1].Observations))
	}
}

func TestCitationIsDerivedFromWhatWasWritten(t *testing.T) {
	// You cannot claim to have used something you did not carry forward.
	em, _ := grammar.Parse(`<commit>{"summary": "it works", "source": "<obs-0004>"}</commit>`)
	c := em.(grammar.Commit)
	cites := c.Content.Cites()
	if len(cites) != 1 || cites[0] != "obs-0004" {
		t.Fatalf("want [obs-0004], got %v", cites)
	}

	// And the case the first real model return produced: a reference sitting
	// inside prose. If this is not derived, a model writing naturally cites
	// nothing and CREDIT FOLLOWS DATA silently stops being true.
	em2, _ := grammar.Parse(
		`<commit>{"summary": "A write-ahead log records changes first. <obs-0001>"}</commit>`)
	c2 := em2.(grammar.Commit)
	if got := c2.Content.Cites(); len(got) != 1 || got[0] != "obs-0001" {
		t.Fatalf("embedded reference not derived: %v", got)
	}
}

func TestNoDeclaredCitationChannelExists(t *testing.T) {
	// Two places for one fact would eventually disagree. There is no Cites field
	// to set, so a commit CANNOT claim to cite A while referencing B.
	em, err := grammar.Parse(`<commit>{"cites": "obs-0099", "body": "<obs-0004>"}</commit>`)
	if err != nil {
		t.Fatal(err)
	}
	c := em.(grammar.Commit)
	cites := c.Content.Cites()
	// "obs-0099" is a bare string, not a reference — it was not written in
	// reference form, so it is not data flow. Only obs-0004 counts.
	if len(cites) != 1 || cites[0] != "obs-0004" {
		t.Fatalf("a declared citation was honoured: %v", cites)
	}
}

func TestShapePartitionsErrorsFromSuccesses(t *testing.T) {
	// Errors partition away from successes WITHOUT ANYONE MODELLING ERRORS.
	ok1 := content.Content{"tests": content.Num(480)}
	ok2 := content.Content{"tests": content.Num(500)}
	bad := content.Content{"error": content.Str("timeout")}

	if ok1.Shape() != ok2.Shape() {
		t.Errorf("a value change altered the shape: %q vs %q", ok1.Shape(), ok2.Shape())
	}
	if ok1.Shape() == bad.Shape() {
		t.Error("an error shares a shape with a success")
	}

	// A list's length is a value change, not a structural one.
	l3 := content.Content{"xs": content.List{content.Num(1), content.Num(2), content.Num(3)}}
	l4 := content.Content{"xs": content.List{content.Num(1), content.Num(2), content.Num(3), content.Num(4)}}
	if l3.Shape() != l4.Shape() {
		t.Errorf("list length changed the shape: %q vs %q", l3.Shape(), l4.Shape())
	}

	// A type change IS structural.
	ls := content.Content{"xs": content.List{content.Str("a")}}
	if l3.Shape() == ls.Shape() {
		t.Error("a list of numbers and a list of strings share a shape")
	}
}

func TestSameQuestionDereferences(t *testing.T) {
	// Comparing Refs by ID makes every question unique again, which is what
	// killed the first InvocationKey.
	store := map[string]content.Content{
		"obs-0001": {"need": content.Str("count the lines")},
		"obs-0002": {"need": content.Str("count the lines")}, // same material, different instance
		"obs-0003": {"need": content.Str("something else")},
	}
	deref := func(id ids.ObservationID) (content.Content, bool) {
		c, ok := store[string(id)]
		return c, ok
	}

	a := content.Content{"arg": content.Ref("obs-0001")}
	b := content.Content{"arg": content.Ref("obs-0002")}
	c := content.Content{"arg": content.Ref("obs-0003")}

	fa := content.Fingerprint(a, deref)
	fb := content.Fingerprint(b, deref)
	fc := content.Fingerprint(c, deref)

	if fa != fb {
		t.Errorf("two instances of the same material are not the same question:\n %s\n %s", fa, fb)
	}
	if fa == fc {
		t.Error("different material compares equal")
	}

	// Order-independent: one universal signature gives position no meaning.
	x := content.FingerprintSet([]content.Content{a, c}, deref)
	y := content.FingerprintSet([]content.Content{c, a}, deref)
	if x != y {
		t.Error("argument order changed the question")
	}

	// An unresolvable reference degrades to identity AND SAYS SO, rather than
	// silently comparing equal to everything.
	orphan := content.Content{"arg": content.Ref("obs-9999")}
	if f := content.Fingerprint(orphan, deref); !strings.Contains(f, "?@") {
		t.Errorf("an unresolvable ref did not mark itself: %s", f)
	}
}
