package dnav2

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The annotated-sentence parser — the universal querier contract's mouth
// (minimal-loop-uncertainties.md item 1, decided 2026-07-12). A worker emits ONE
// annotated sentence; the subgraph is PARSED from it, so sentence↔subgraph
// correspondence is enforced here by construction — the parser IS the validator
// (query-by-subgraph.md). No dual emission, no second judge.
//
// Syntax:
//
//	nodes:  word[k]            — per-emission ordinal k
//	        {a phrase}[k]      — multi-word referent
//	        word[k=stable:id]  — explicit reuse (the composition hook:
//	                             {the outage}[1=compound:g42])
//	edges:  verb[a→b]          — the annotated word/phrase is the edge label;
//	                             ASCII -> accepted; multi-arc = multiple edges
//	coreference: repeating an ordinal re-references the node
//	             (michael[1] … he[1]); first occurrence's label names it
//	glue:   unannotated text — ignored for the graph, kept in the sentence
//
// Parse failures return a ParseErrors carrying every problem found, so the
// retry-with-correction prompt (bounded, NO AIMD tax — a format error is not an
// epistemic error) can report them all at once.

// ParsedNode is one V element: a material referent.
type ParsedNode struct {
	Ordinal  int
	Label    string // normalized (lowercase, punctuation-trimmed)
	StableID string // from [k=stable:id]; "" = resolve by label post-parse
}

// ParsedEdge is one E element: an asserted relationship between ordinals.
type ParsedEdge struct {
	Label string
	From  int
	To    int
}

// ParsedCompound is one hylomorphic query result: G = (V, E) plus the clean
// sentence (annotations stripped) that becomes the summary node's content.
type ParsedCompound struct {
	Sentence string
	Nodes    []ParsedNode // sorted by ordinal
	Edges    []ParsedEdge // in emission order
}

// ParseErrors collects every problem in one emission — the correction prompt
// reports them all so the retry can fix everything at once.
type ParseErrors struct {
	Problems []string
}

func (e *ParseErrors) Error() string {
	return "annotated sentence: " + strings.Join(e.Problems, "; ")
}

// annotRe matches one annotated element: {phrase}[body] or word[body].
// The word char-class excludes brackets/braces so greedy matching stops at the
// annotation and never swallows a neighbor.
var annotRe = regexp.MustCompile(`(?:\{([^{}]*)\}|([^\s\[\]{}]+))\[([^\[\]]*)\]`)

var (
	edgeBodyRe = regexp.MustCompile(`^(\d+)\s*(?:→|->)\s*(\d+)$`)
	nodeBodyRe = regexp.MustCompile(`^(\d+)(?:=(\S+))?$`)
	wordRe     = regexp.MustCompile(`[\pL\pN_'-]+`)
)

// Unannotated text may only join annotated graph elements. Keeping this list
// deliberately small makes omitted material/formal structure a parse error
// instead of silently treating semantic content as glue.
var glueWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"to": true, "and": true, "or": true, "but": true,
}

// ParseAnnotated parses one annotated sentence into its compound. It returns a
// *ParseErrors listing every problem when the emission is malformed.
func ParseAnnotated(raw string) (ParsedCompound, error) {
	var pe ParseErrors
	out := ParsedCompound{}

	matches := annotRe.FindAllStringSubmatch(raw, -1)
	byOrdinal := map[int]*ParsedNode{}
	for _, m := range matches {
		label := normalizeLabel(firstNonEmpty(m[1], m[2]))
		body := strings.TrimSpace(m[3])

		if em := edgeBodyRe.FindStringSubmatch(body); em != nil {
			from, _ := strconv.Atoi(em[1])
			to, _ := strconv.Atoi(em[2])
			if label == "" {
				pe.Problems = append(pe.Problems, fmt.Sprintf("edge [%s] has no label word", body))
				continue
			}
			out.Edges = append(out.Edges, ParsedEdge{Label: label, From: from, To: to})
			continue
		}

		if nm := nodeBodyRe.FindStringSubmatch(body); nm != nil {
			ord, _ := strconv.Atoi(nm[1])
			stable := nm[2]
			if ord <= 0 {
				pe.Problems = append(pe.Problems, fmt.Sprintf("node ordinal %d must be positive", ord))
				continue
			}
			if label == "" && stable == "" {
				pe.Problems = append(pe.Problems, fmt.Sprintf("node [%d] has neither a label nor a stable id", ord))
				continue
			}
			if existing, ok := byOrdinal[ord]; ok {
				// coreference: same ordinal re-references the node. First
				// occurrence's label names it; a stable id from ANY occurrence
				// binds it; two different stable ids are a contradiction.
				if stable != "" {
					if existing.StableID != "" && existing.StableID != stable {
						pe.Problems = append(pe.Problems,
							fmt.Sprintf("ordinal %d bound to two stable ids: %s and %s", ord, existing.StableID, stable))
					} else {
						existing.StableID = stable
					}
				}
				continue
			}
			byOrdinal[ord] = &ParsedNode{Ordinal: ord, Label: label, StableID: stable}
			continue
		}

		pe.Problems = append(pe.Problems, fmt.Sprintf("unparseable annotation [%s] on %q", body, label))
	}

	// Inspect only text OUTSIDE annotations. Any non-glue word there is semantic
	// content with no node/edge representation, violating sentence↔subgraph
	// correspondence. Report all omissions in one correction pass.
	cursor := 0
	for _, loc := range annotRe.FindAllStringIndex(raw, -1) {
		collectUnannotated(raw[cursor:loc[0]], &pe)
		cursor = loc[1]
	}
	collectUnannotated(raw[cursor:], &pe)

	// every edge endpoint must be a declared node ordinal (whole-sentence scan
	// already done, so forward references are fine).
	for _, e := range out.Edges {
		for _, end := range []int{e.From, e.To} {
			if _, ok := byOrdinal[end]; !ok {
				pe.Problems = append(pe.Problems,
					fmt.Sprintf("edge %q[%d→%d] references undeclared node %d", e.Label, e.From, e.To, end))
			}
		}
	}
	if len(byOrdinal) == 0 {
		pe.Problems = append(pe.Problems, "no annotated nodes — an empty claim")
	}
	if len(pe.Problems) > 0 {
		return ParsedCompound{}, &pe
	}

	for _, n := range byOrdinal {
		out.Nodes = append(out.Nodes, *n)
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].Ordinal < out.Nodes[j].Ordinal })
	out.Sentence = cleanSentence(raw)
	return out, nil
}

func collectUnannotated(gap string, pe *ParseErrors) {
	for _, word := range wordRe.FindAllString(gap, -1) {
		if normalized := strings.ToLower(word); !glueWords[normalized] {
			pe.Problems = append(pe.Problems,
				fmt.Sprintf("unannotated semantic token %q — mark it as a node or relationship", word))
		}
	}
}

// cleanSentence strips the annotations: {phrase}[body] → phrase, word[body] → word.
func cleanSentence(raw string) string {
	s := annotRe.ReplaceAllStringFunc(raw, func(m string) string {
		sub := annotRe.FindStringSubmatch(m)
		return firstNonEmpty(sub[1], sub[2])
	})
	return strings.Join(strings.Fields(s), " ")
}

// normalizeLabel lowercases and trims surrounding punctuation — the exact-match
// reuse key (the token-web discipline).
func normalizeLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Trim(s, `.,!?;:"'()`)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

const parseMaxRetries = 2 // bounded retry-with-correction; NO AIMD tax on format errors

// CorrectionMessage builds the retry prompt for a failed emission — every
// problem reported at once, format restated, same budget (a format error is not
// an epistemic error).
func CorrectionMessage(raw string, err error) string {
	return fmt.Sprintf(
		"Your previous response was:\n%s\n\nIt could not be parsed: %v.\n\n"+
			"Re-emit the SAME claim as one annotated sentence. Format:\n"+
			"- mark each referent:  michael[1]  or  {the skyra project}[2]\n"+
			"- mark each relationship on its verb/word:  asked[1→2]\n"+
			"- reuse an ordinal to re-reference:  michael[1] ... he[1]\n"+
			"- known ids may bind explicitly:  michael[1=entity:michael]\n"+
			"Every referent needs a node annotation; every relationship needs an edge annotation.",
		raw, err)
}
