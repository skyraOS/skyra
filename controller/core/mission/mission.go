// Package mission defines the typed Plan/Spec/Launch boundary. It performs
// only pure validation and resolution; it never grants OS access or runs a tree.
package mission

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
)

// Contract is the worker protocol's identity: binary plus canonical input shape.
// Name, resource scope, and the human's authority are not its type.
type Contract struct {
	Binary string `json:"binary"`
	Shape  string `json:"shape"`
}

func validHash(id, prefix string, n int) bool {
	if !strings.HasPrefix(id, prefix) || len(id) != len(prefix)+n {
		return false
	}
	for _, c := range id[len(prefix):] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func Identify(binary string, schema *contract.Schema) (Contract, error) {
	if !(validHash(binary, "sha256:", 64) || validHash(binary, "cdhash:", 40)) {
		return Contract{}, errors.New("contract requires a pinned binary identity")
	}
	if err := schema.Check(); err != nil {
		return Contract{}, err
	}
	return Contract{binary, string(contract.Canonical(contract.Normal(schema)))}, nil
}
func (c Contract) Validate() error {
	var schema contract.Schema
	if err := strict([]byte(c.Shape), &schema); err != nil {
		return err
	}
	normal, err := Identify(c.Binary, &schema)
	if err != nil {
		return err
	}
	if normal != c {
		return errors.New("contract shape is not canonical")
	}
	return nil
}
func hash(value any) string {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(value); err != nil {
		return ""
	}
	sum := sha256.Sum256(bytes.TrimSuffix(b.Bytes(), []byte("\n")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// InvokeType and ReturnType use the tree protocol's existing domain tags and
// canonical encoding. OutputShape is the verified runtime shape, not a claim
// supplied by the model. Known-answer tests pin these to the worker core.
func InvokeType(c Contract, passed []string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	for _, id := range passed {
		if !validHash(id, "sha256:", 64) {
			return "", errors.New("invalid passed return type")
		}
	}
	return hash([]any{"skyra/input/v2", c.Binary, c.Shape, append([]string{}, passed...)}), nil
}
func ReturnType(invocation, outputShape string) (string, error) {
	if !validHash(invocation, "sha256:", 64) || outputShape == "" {
		return "", errors.New("a return type needs its invocation and verified output shape")
	}
	return hash([]any{"skyra/return/v1", invocation, outputShape}), nil
}

type Descriptor struct {
	Ref      root.Capability `json:"ref"`
	Target   string          `json:"target"`
	Contract Contract        `json:"contract"`
	// Pin binds routing, uses, and state to the registered program identity.
	Pin       string             `json:"pin"`
	Authority boundary.Authority `json:"authority"`
}

func copyDescriptor(d Descriptor) Descriptor {
	b, _ := json.Marshal(d)
	var out Descriptor
	_ = json.Unmarshal(b, &out)
	return out
}

type Catalogue struct {
	ceiling boundary.Authority
	entries map[root.Capability]Descriptor
}

func NewCatalogue(ceiling boundary.Authority, descriptors []Descriptor) (Catalogue, error) {
	b, _ := json.Marshal(ceiling)
	var held boundary.Authority
	_ = json.Unmarshal(b, &held)
	out := Catalogue{ceiling: held, entries: map[root.Capability]Descriptor{}}
	for _, d := range descriptors {
		d = copyDescriptor(d)
		if d.Target == "" || d.Pin == "" {
			return Catalogue{}, errors.New("capability needs a target and registered program pin")
		}
		if err := d.Contract.Validate(); err != nil {
			return Catalogue{}, err
		}
		if !held.Allows(d.Authority) {
			return Catalogue{}, errors.New("capability scope exceeds the human's authority ceiling")
		}
		expected := root.Capability(hash([]any{"skyra/mission-capability/v1", d.Target, d.Contract, d.Pin, d.Authority}))
		if d.Ref != "" && d.Ref != expected {
			return Catalogue{}, errors.New("capability reference does not match its contract and scope")
		}
		d.Ref = expected
		if _, ok := out.entries[d.Ref]; ok {
			return Catalogue{}, errors.New("duplicate catalogue capability")
		}
		out.entries[d.Ref] = d
	}
	return out, nil
}
func (c Catalogue) Entries() []Descriptor {
	out := []Descriptor{}
	for _, d := range c.entries {
		out = append(out, copyDescriptor(d))
	}
	slices.SortFunc(out, func(a, b Descriptor) int { return strings.Compare(string(a.Ref), string(b.Ref)) })
	return out
}
func (c Catalogue) Lookup(ref root.Capability) (Descriptor, bool) {
	d, ok := c.entries[ref]
	return copyDescriptor(d), ok
}
func (c Catalogue) Ceiling() boundary.Authority {
	b, _ := json.Marshal(c.ceiling)
	var a boundary.Authority
	_ = json.Unmarshal(b, &a)
	return a
}

// Spec is an immutable proposed task and grant. References are informational
// until the human's committed Invoke conveys the held capabilities to Launch.
//
// Its goal is made of segments, each saying where its words came from: cited,
// quoted verbatim from an observation the root holds; plan, written by Plan;
// or human, written by the human. Coverage counts them. Both are part of the
// Spec, derived from its segments and checked wherever it is decoded, so no
// one states a coverage: it is what the goal is made of.
type Spec struct {
	Version      int               `json:"version"`
	Goal         string            `json:"goal"`
	Segments     []Segment         `json:"segments"`
	Coverage     Coverage          `json:"coverage"`
	Inputs       []root.ID         `json:"inputs"`
	Capabilities []root.Capability `json:"capabilities"`
	// Profile is the constants Launch's tree runs under: its root sets them
	// for every agent in it. Every Spec states them.
	Profile Profile `json:"profile"`
	// Note is Plan's reply to the human with the Spec it returns: what it
	// changed, or what it needs to know. It is not the task, and not counted.
	Note string `json:"note,omitempty"`
}

// SameTask reports whether two Specs ask the same of Launch: the same goal,
// sources, inputs, tools and profile. A note does not change the task.
func (s Spec) SameTask(o Spec) bool {
	return s.Goal == o.Goal && slices.Equal(s.Segments, o.Segments) && slices.Equal(s.Inputs, o.Inputs) && slices.Equal(s.Capabilities, o.Capabilities) && s.Profile == o.Profile
}

// SpecVersion is the version whose goal is made of segments and that states
// the profile its tree runs under.
const SpecVersion = 3

// MinQuote is the fewest words a quote must hold to count as cited. A
// shorter one counts as Plan's words, so scattered words cannot pad coverage.
const MinQuote = 3

const (
	Cited   = "cited"
	ByPlan  = "plan"
	ByHuman = "human"
)

// A Segment is one run of a goal and where its words came from. Of names the
// observation a cited segment quotes.
type Segment struct {
	Text   string  `json:"text"`
	Source string  `json:"source"`
	Of     root.ID `json:"of,omitempty"`
}

// Coverage is the goal's words by source.
type Coverage struct {
	Words int `json:"words"`
	Cited int `json:"cited"`
	Plan  int `json:"plan"`
	Human int `json:"human"`
}

// Assemble is the goal and coverage segments make: their texts in order, a
// space between two that would otherwise run together.
func Assemble(segments []Segment) (string, Coverage) {
	var b strings.Builder
	var c Coverage
	for _, seg := range segments {
		if b.Len() > 0 && seg.Text != "" && !endsSpace(b.String()) && !startsSpace(seg.Text) {
			b.WriteByte(' ')
		}
		b.WriteString(seg.Text)
		n := len(strings.Fields(seg.Text))
		c.Words += n
		switch seg.Source {
		case Cited:
			c.Cited += n
		case ByPlan:
			c.Plan += n
		case ByHuman:
			c.Human += n
		}
	}
	return b.String(), c
}

func endsSpace(s string) bool   { r, _ := utf8.DecodeLastRuneInString(s); return unicode.IsSpace(r) }
func startsSpace(s string) bool { r, _ := utf8.DecodeRuneInString(s); return unicode.IsSpace(r) }

// HumanSpec is a Spec the human wrote: one segment, all theirs.
func HumanSpec(goal string, inputs []root.ID, capabilities []root.Capability, profile Profile) Spec {
	segments := []Segment{{Text: goal, Source: ByHuman}}
	text, coverage := Assemble(segments)
	return Spec{Version: SpecVersion, Goal: text, Segments: segments, Coverage: coverage, Inputs: inputs, Capabilities: capabilities, Profile: profile}
}

// Quotes reports whether value, an observation's JSON value, holds quote
// verbatim in one of its strings, whitespace runs compared as one space.
func Quotes(value json.RawMessage, quote string) bool {
	want := strings.Join(strings.Fields(quote), " ")
	if want == "" {
		return false
	}
	var v any
	if json.Unmarshal(value, &v) != nil {
		return false
	}
	var found func(any) bool
	found = func(v any) bool {
		switch x := v.(type) {
		case string:
			return strings.Contains(strings.Join(strings.Fields(x), " "), want)
		case []any:
			return slices.ContainsFunc(x, found)
		case map[string]any:
			for _, e := range x {
				if found(e) {
					return true
				}
			}
		}
		return false
	}
	return found(v)
}

// CheckCites holds every cited segment to the observation it names, as the
// root holds it. A quote of a Spec holds only within one of its segments that
// a human wrote or cited: Plan's words in a Spec are never anyone's quote.
func (s Spec) CheckCites(lookup func(root.ID) (root.Observation, bool)) error {
	for i, seg := range s.Segments {
		if seg.Source != Cited {
			continue
		}
		o, ok := lookup(seg.Of)
		if !ok {
			return fmt.Errorf("segment %d quotes an observation outside the human's scope", i)
		}
		if quoted, err := DecodeSpec(o.Value()); err == nil {
			if !slices.ContainsFunc(quoted.Segments, func(q Segment) bool {
				return q.Source != ByPlan && Quotes(mustJSON(q.Text), seg.Text)
			}) {
				return fmt.Errorf("segment %d is not words a human wrote or cited in the spec it quotes: %q", i, clip(seg.Text))
			}
			continue
		}
		if !Quotes(o.Value(), seg.Text) {
			return fmt.Errorf("segment %d is not in the observation it quotes: %q", i, clip(seg.Text))
		}
	}
	return nil
}

func mustJSON(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

// Revise is the human's new text as a revision of prev: words they kept keep
// their segment's source, and every word they inserted or changed is theirs.
// A cited run the edit cut below MinQuote words is theirs too, as is one it
// broke, since it no longer quotes its observation verbatim. Inputs carry
// over; the tools and profile are the human's.
func Revise(prev Spec, text string, capabilities []root.Capability, profile Profile) Spec {
	type word struct{ seg, at int }
	var old []word
	var oldText []string
	for i, seg := range prev.Segments {
		for _, w := range strings.Fields(seg.Text) {
			old, oldText = append(old, word{i, len(old)}), append(oldText, w)
		}
	}
	lead := text[:len(text)-len(strings.TrimLeftFunc(text, unicode.IsSpace))]
	tokens := token.FindAllString(text, -1)
	words := make([]string, len(tokens))
	for i, t := range tokens {
		words[i] = strings.TrimRightFunc(t, unicode.IsSpace)
	}
	match := common(oldText, words)
	segments := []Segment{}
	var run []string
	from := -1 // the old word the run's last token kept, or -1
	flush := func() {
		if len(run) == 0 {
			return
		}
		seg := Segment{Text: strings.Join(run, ""), Source: ByHuman}
		if from >= 0 {
			kept := prev.Segments[old[from].seg]
			seg.Source, seg.Of = kept.Source, kept.Of
			if kept.Source == Cited && len(run) < MinQuote {
				seg.Source, seg.Of = ByHuman, ""
			}
		}
		if n := len(segments); n > 0 && seg.Source != Cited && segments[n-1].Source == seg.Source {
			segments[n-1].Text += seg.Text
		} else {
			segments = append(segments, seg)
		}
		run = nil
	}
	for i, t := range tokens {
		m := match[i]
		continues := len(run) > 0 && (m < 0) == (from < 0) && (m < 0 || old[m].seg == old[from].seg && m == from+1)
		if !continues {
			flush()
		}
		run, from = append(run, t), m
	}
	flush()
	if lead != "" {
		if len(segments) > 0 && segments[0].Source != Cited {
			segments[0].Text = lead + segments[0].Text
		} else {
			segments = append([]Segment{{Text: lead, Source: ByHuman}}, segments...)
		}
	}
	goal, coverage := Assemble(segments)
	return Spec{Version: SpecVersion, Goal: goal, Segments: segments, Coverage: coverage, Inputs: slices.Clone(prev.Inputs), Capabilities: capabilities, Profile: profile}
}

// token is a word and the whitespace after it.
var token = regexp.MustCompile(`\S+\s*`)

// common is, for each of b's words, the index of the a word it keeps in a
// longest common subsequence, or -1. Beyond a bound, only the shared prefix
// and suffix are kept, so a huge edit costs no more than a rewrite.
func common(a, b []string) []int {
	out := make([]int, len(b))
	for i := range out {
		out[i] = -1
	}
	p := 0
	for p < len(a) && p < len(b) && a[p] == b[p] {
		out[p] = p
		p++
	}
	sa, sb := len(a), len(b)
	for sa > p && sb > p && a[sa-1] == b[sb-1] {
		sa, sb = sa-1, sb-1
		out[sb] = sa
	}
	ma, mb := a[p:sa], b[p:sb]
	if len(ma)*len(mb) > 4_000_000 || len(ma) == 0 || len(mb) == 0 {
		return out
	}
	// lcs[i][j] is the common length of ma[i:] and mb[j:].
	lcs := make([][]int32, len(ma)+1)
	for i := range lcs {
		lcs[i] = make([]int32, len(mb)+1)
	}
	for i := len(ma) - 1; i >= 0; i-- {
		for j := len(mb) - 1; j >= 0; j-- {
			if ma[i] == mb[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	for i, j := 0, 0; i < len(ma) && j < len(mb); {
		switch {
		case ma[i] == mb[j]:
			out[p+j] = p + i
			i, j = i+1, j+1
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	return out
}

func clip(s string) string {
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}

func SpecSchema() *contract.Schema {
	str := &contract.Schema{Type: "string"}
	num := &contract.Schema{Type: "number"}
	segment := &contract.Schema{Type: "object", Required: []string{"text", "source"}, Fields: map[string]*contract.Schema{"text": str, "source": str, "of": str}}
	coverage := &contract.Schema{Type: "object", Required: []string{"words", "cited", "plan", "human"}, Fields: map[string]*contract.Schema{"words": num, "cited": num, "plan": num, "human": num}}
	return &contract.Schema{Title: "mission-spec/v3", Type: "object", Required: []string{"version", "goal", "segments", "coverage", "inputs", "capabilities", "profile"}, Fields: map[string]*contract.Schema{
		"version": num, "goal": str, "segments": {Type: "array", Items: segment}, "coverage": coverage, "inputs": {Type: "array", Items: str}, "capabilities": {Type: "array", Items: str}, "profile": ProfileSchema(), "note": str,
	}}
}

// ProfileSchema is a profile's shape: all four constants, each required.
func ProfileSchema() *contract.Schema {
	num := &contract.Schema{Type: "number"}
	return &contract.Schema{Type: "object", Required: []string{"act", "turns", "depth", "fanout", "budget"}, Fields: map[string]*contract.Schema{"act": num, "turns": num, "depth": num, "fanout": num, "budget": num}}
}
func endpoint(name string, fields map[string]*contract.Schema, required []string) *contract.Schema {
	return &contract.Schema{Title: name, Type: "object", Required: []string{name}, Fields: map[string]*contract.Schema{name: {Type: "object", Fields: fields, Required: required}}}
}

// ObservationSchema is the adapter's transported observation envelope. Models
// use local names; the adapter supplies the held reference, producer and bytes.
func ObservationSchema() *contract.Schema {
	str := &contract.Schema{Type: "string"}
	return &contract.Schema{Type: "object", Required: []string{"ref", "binary", "contract", "value", "source"}, Fields: map[string]*contract.Schema{"ref": str, "binary": str, "contract": str, "value": {Type: "any"}, "source": {Type: "any"}}}
}
func PlanSchema() *contract.Schema {
	return endpoint("plan", map[string]*contract.Schema{"request": ObservationSchema(), "context": {Type: "array", Items: ObservationSchema()}}, []string{"request", "context"})
}
func LaunchSchema() *contract.Schema {
	return endpoint("launch", map[string]*contract.Schema{"spec": ObservationSchema()}, []string{"spec"})
}

// ProposeSchema is the proposer's input: a Spec whose capabilities are tools
// chosen by name from catalogue, the observation view_plan_contracts returned,
// and whose goal is segments, each a quote from an observation Plan was
// passed, by its id, or text Plan adds. The proposer resolves the names and
// counts the words; a model never writes a capability ref or a coverage.
func ProposeSchema() *contract.Schema {
	str := &contract.Schema{Type: "string"}
	segment := &contract.Schema{Type: "object", Fields: map[string]*contract.Schema{"quote": str, "of": str, "text": str}}
	return endpoint("propose_spec", map[string]*contract.Schema{
		"goal": {Type: "array", Items: segment}, "inputs": {Type: "array", Items: str},
		"catalogue": {Type: "any"}, "capabilities": {Type: "array", Items: str}, "profile": ProfileSchema(), "note": str,
	}, []string{"goal", "inputs", "catalogue", "capabilities", "profile"})
}

// ViewSchema is view_plan_contracts' input: it asks for nothing.
func ViewSchema() *contract.Schema {
	return endpoint("view_plan_contracts", map[string]*contract.Schema{}, []string{})
}

// A View is the capabilities a plan may attach, as view_plan_contracts
// returns them: those the Mission's human holds, each with its reference,
// tool, input contract and the resource scope it runs under.
type View struct {
	Contracts []ViewEntry `json:"contracts"`
}

type ViewEntry struct {
	Ref    root.Capability    `json:"ref"`
	Target string             `json:"target"`
	Input  json.RawMessage    `json:"input"`
	Scope  boundary.Authority `json:"scope"`
}

func (c Catalogue) View() View {
	refs := []root.Capability{}
	for _, d := range c.Entries() {
		refs = append(refs, d.Ref)
	}
	return c.ViewOf(refs)
}

// ViewOf is the View of the capabilities held: those of the catalogue's
// entries the human holds. A held capability the catalogue does not describe
// (Plan's own) is not one a plan attaches.
func (c Catalogue) ViewOf(held []root.Capability) View {
	v := View{Contracts: []ViewEntry{}}
	for _, d := range c.Entries() {
		if slices.Contains(held, d.Ref) {
			v.Contracts = append(v.Contracts, ViewEntry{Ref: d.Ref, Target: d.Target, Input: json.RawMessage(d.Contract.Shape), Scope: d.Authority})
		}
	}
	return v
}

// gutter is the line numbering a passed observation is shown with, which a
// quote copied from it may carry; it is not the observation's text.
var gutter = regexp.MustCompile(`(?m)^[ \t]*\d+\| ?`)

// Propose resolves a proposal against the view it names: every contract of
// each chosen tool, by the reference the view gave it. A tool the view does
// not hold is refused.
func Propose(raw json.RawMessage) (Spec, error) {
	v, err := contract.Decode(raw)
	if err != nil {
		return Spec{}, err
	}
	if err = contract.Fits(ProposeSchema(), v); err != nil {
		return Spec{}, err
	}
	var in struct {
		Propose struct {
			Goal []struct {
				Quote *string `json:"quote"`
				Of    root.ID `json:"of"`
				Text  *string `json:"text"`
			} `json:"goal"`
			Inputs       []root.ID       `json:"inputs"`
			Catalogue    json.RawMessage `json:"catalogue"`
			Capabilities []string        `json:"capabilities"`
			Profile      Profile         `json:"profile"`
			Note         string          `json:"note"`
		} `json:"propose_spec"`
	}
	if err = strict(raw, &in); err != nil {
		return Spec{}, err
	}
	var view View
	if err = strict(in.Propose.Catalogue, &view); err != nil {
		return Spec{}, fmt.Errorf("catalogue is not what view_plan_contracts returned: %w", err)
	}
	refs := []root.Capability{}
	for i, tool := range in.Propose.Capabilities {
		if slices.Contains(in.Propose.Capabilities[:i], tool) {
			return Spec{}, fmt.Errorf("tool %s proposed twice", tool)
		}
		found := false
		for _, e := range view.Contracts {
			if e.Target == tool {
				refs, found = append(refs, e.Ref), true
			}
		}
		if !found {
			return Spec{}, fmt.Errorf("the viewed contracts hold no tool %s", tool)
		}
	}
	segments := []Segment{}
	for i, g := range in.Propose.Goal {
		switch {
		case g.Quote != nil && g.Text == nil:
			if !validHash(string(g.Of), "", 64) {
				return Spec{}, fmt.Errorf("goal segment %d quotes no observation id", i)
			}
			quote := gutter.ReplaceAllString(*g.Quote, "")
			seg := Segment{Text: quote, Source: Cited, Of: g.Of}
			if len(strings.Fields(quote)) < MinQuote {
				seg = Segment{Text: quote, Source: ByPlan}
			}
			segments = append(segments, seg)
		case g.Text != nil && g.Quote == nil && g.Of == "":
			segments = append(segments, Segment{Text: *g.Text, Source: ByPlan})
		default:
			return Spec{}, fmt.Errorf("goal segment %d is neither a quote of an observation nor text", i)
		}
	}
	goal, coverage := Assemble(segments)
	spec, _ := json.Marshal(Spec{Version: SpecVersion, Goal: goal, Segments: segments, Coverage: coverage, Inputs: append([]root.ID{}, in.Propose.Inputs...), Capabilities: refs, Profile: in.Propose.Profile, Note: in.Propose.Note})
	return DecodeSpec(spec)
}
func strict(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if !json.Valid(raw) {
		return errors.New("expected one JSON value")
	}
	return nil
}
func DecodeSpec(raw json.RawMessage) (Spec, error) {
	v, err := contract.Decode(raw)
	if err != nil {
		return Spec{}, err
	}
	if err = contract.Fits(SpecSchema(), v); err != nil {
		return Spec{}, err
	}
	var spec Spec
	if err = strict(raw, &spec); err != nil {
		return Spec{}, err
	}
	// A goal may be empty: Plan's first reply may have no task yet. Only a
	// Spec with a goal resolves for Launch.
	if spec.Version != SpecVersion {
		return Spec{}, fmt.Errorf("spec needs version %d", SpecVersion)
	}
	if why := spec.Profile.Check(); why != "" {
		return Spec{}, errors.New(why)
	}
	if len(spec.Note) > 16*1024 {
		return Spec{}, errors.New("spec note exceeds 16 KiB")
	}
	for i, seg := range spec.Segments {
		switch {
		case seg.Source == Cited && (!validHash(string(seg.Of), "", 64) || len(strings.Fields(seg.Text)) < MinQuote):
			return Spec{}, fmt.Errorf("segment %d is not a quote of at least %d words from an observation", i, MinQuote)
		case seg.Source != Cited && (seg.Source != ByPlan && seg.Source != ByHuman || seg.Of != ""):
			return Spec{}, fmt.Errorf("segment %d has no known source", i)
		}
	}
	if goal, coverage := Assemble(spec.Segments); goal != spec.Goal || coverage != spec.Coverage {
		return Spec{}, errors.New("spec goal and coverage are not what its segments make")
	}
	if len(spec.Goal) > 256*1024 {
		return Spec{}, errors.New("spec goal exceeds 256 KiB")
	}
	for i, id := range spec.Inputs {
		if !validHash(string(id), "", 64) || slices.Contains(spec.Inputs[:i], id) {
			return Spec{}, errors.New("invalid or duplicate input observation reference")
		}
	}
	for i, ref := range spec.Capabilities {
		if !validHash(string(ref), "sha256:", 64) || slices.Contains(spec.Capabilities[:i], ref) {
			return Spec{}, errors.New("invalid or duplicate capability reference")
		}
	}
	spec.Inputs = append([]root.ID{}, spec.Inputs...)
	spec.Capabilities = append([]root.Capability{}, spec.Capabilities...)
	return spec, nil
}

type Resolved struct {
	Spec      Spec
	Contracts []Descriptor
	Inputs    []root.Observation
	Kind      string
	Targets   []string
	Profile   Profile // the constants its tree runs under
}

// A Profile is the constants a tree runs under, as the apprentice's tree
// takes them: its root sets them, in the tree's input, for every agent in it.
type Profile struct {
	Act    int `json:"act"`    // the most contracts one node holds and still invokes itself
	Turns  int `json:"turns"`  // the most turns one agent may take
	Depth  int `json:"depth"`  // the deepest a child may be
	FanOut int `json:"fanout"` // the most instructions one call may carry
	Budget int `json:"budget"` // the most turns the whole tree may take
}

// DefaultProfile is the one-contract tree: an act node holds exactly one
// contract.
var DefaultProfile = Profile{Act: 1, Turns: 8, Depth: 3, FanOut: MaxFanOut, Budget: 64}

// MaxFanOut is the kernel's fan-out: no profile allows more instructions in
// one call.
const MaxFanOut = 32

// Check is why a profile cannot run a tree, or "": the apprentice's rule.
func (p Profile) Check() string {
	switch {
	case p.Act < 1:
		return "a profile's act bound is at least 1: a node acts through something"
	case p.Turns < 1:
		return "a profile allows at least one turn"
	case p.Depth < 0:
		return "a profile's depth is not negative"
	case p.FanOut < 1 || p.FanOut > MaxFanOut:
		return fmt.Sprintf("a profile's fan-out is between 1 and %d", MaxFanOut)
	case p.Budget < 1:
		return "a profile's budget allows at least one turn"
	}
	return ""
}

// KindOf is a root's kind from the distinct contracts it holds: act while
// there are at most the act bound, delegate past it.
func (p Profile) KindOf(contracts int) string {
	if contracts <= p.Act {
		return "act"
	}
	return "delegate"
}

func (c Catalogue) Resolve(spec Spec, state root.State) (Resolved, error) {
	raw, _ := json.Marshal(spec)
	spec, err := DecodeSpec(raw)
	if err != nil {
		return Resolved{}, err
	}
	if strings.TrimSpace(spec.Goal) == "" {
		return Resolved{}, errors.New("a spec needs a goal before it can be launched")
	}
	if err = spec.CheckCites(state.Lookup); err != nil {
		return Resolved{}, err
	}
	out := Resolved{Spec: spec, Contracts: []Descriptor{}, Inputs: []root.Observation{}, Targets: []string{}, Profile: spec.Profile}
	seen := map[Contract]bool{}
	for _, ref := range spec.Capabilities {
		d, ok := c.Lookup(ref)
		if !ok {
			return Resolved{}, errors.New("spec requests a capability absent from the human's catalogue")
		}
		if !root.MayGrant(state.Config().Ceiling, []root.Capability{ref}) {
			return Resolved{}, errors.New("root does not hold the requested capability")
		}
		// Same contract under two scopes is still one callable contract.
		seen[d.Contract] = true
		out.Contracts = append(out.Contracts, d)
		if !slices.Contains(out.Targets, d.Target) {
			out.Targets = append(out.Targets, d.Target)
		}
	}
	for _, id := range spec.Inputs {
		o, ok := state.Lookup(id)
		if !ok {
			return Resolved{}, errors.New("spec cites an observation outside the human's scope")
		}
		out.Inputs = append(out.Inputs, o)
	}
	slices.Sort(out.Targets)
	if len(seen) == 0 {
		return Resolved{}, errors.New("a launch requires at least one callable contract")
	}
	out.Kind = out.Profile.KindOf(len(seen))
	return out, nil
}

// ValidateHandoff checks the exact observations and grant committed by the
// human root. The spec's value and source are retained, not authored anew.
func (c Catalogue) ValidateHandoff(effect root.Effect, state root.State) (Resolved, error) {
	if effect.Kind != root.Dispatch || effect.Invocation == nil || len(effect.Passed) == 0 {
		return Resolved{}, errors.New("Launch requires a committed root dispatch")
	}
	recorded, ok := state.Invocation(effect.Invocation.ID)
	if !ok || recorded.Target != effect.Invocation.Target || !slices.Equal(recorded.Inputs, effect.Invocation.Inputs) || !slices.Equal(recorded.Grant, effect.Invocation.Grant) {
		return Resolved{}, errors.New("Launch invocation is not committed in this root")
	}
	spec, err := DecodeSpec(effect.Passed[0].Value())
	if err != nil {
		return Resolved{}, err
	}
	resolved, err := c.Resolve(spec, state)
	if err != nil {
		return Resolved{}, err
	}
	expected := append([]root.ID{effect.Passed[0].ID()}, spec.Inputs...)
	if !slices.Equal(effect.Invocation.Inputs, expected) || len(effect.Passed) != len(expected) {
		return Resolved{}, errors.New("Launch did not receive precisely the spec and its cited inputs")
	}
	if !sameSet(effect.Invocation.Grant, spec.Capabilities) {
		return Resolved{}, errors.New("Launch grant differs from the approved spec")
	}
	for i, id := range expected {
		original, ok := state.Lookup(id)
		if !ok || effect.Passed[i].ID() != id || original.Source() != effect.Passed[i].Source() || !bytes.Equal(original.Value(), effect.Passed[i].Value()) {
			return Resolved{}, fmt.Errorf("passed observation %d changed identity or provenance", i)
		}
	}
	return resolved, nil
}
func sameSet(a, b []root.Capability) bool {
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
