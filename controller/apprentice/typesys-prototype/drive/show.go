package drive

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"skyra/typesys-prototype/model"
	"skyra/typesys-prototype/tree"
)

// A Learned is one key memory has committed something at, as a person
// reads it: an id to name it by, the key, and each implementation the
// model wrote there, its guard chain in rank order.
type Learned struct {
	ID    string
	Key   model.Key
	Impls [][]tree.Branch
}

// LearnedID names a key: the first 12 hex digits of its text's hash,
// the same in every process.
func LearnedID(k model.Key) string {
	h := sha256.Sum256([]byte(k.Text()))
	return hex.EncodeToString(h[:6])
}

// Learn is what memory has learned, by root, kind, then key: each key with
// an implementation. A key seen but never committed at has learned
// nothing, and refutations are no key. Keys are not a program's: the root
// is the version and the operators' forms, and the tools are in the key's
// own positions, so every program that meets a key shares it.
func Learn(mem model.Memory) ([]Learned, error) {
	var ls []Learned
	for text, v := range mem {
		if len(v.Branches) == 0 || !strings.HasPrefix(text, "{") {
			continue
		}
		var k model.Key
		if err := json.Unmarshal([]byte(text), &k); err != nil {
			return nil, err
		}
		byVoter := map[int][]tree.Branch{}
		for _, b := range v.Branches {
			byVoter[b.Voter] = append(byVoter[b.Voter], b)
		}
		var voters []int
		for n := range byVoter {
			voters = append(voters, n)
		}
		slices.Sort(voters)
		l := Learned{ID: LearnedID(k), Key: k}
		for _, n := range voters {
			chain := byVoter[n]
			slices.SortFunc(chain, func(a, b tree.Branch) int { return a.Rank - b.Rank })
			l.Impls = append(l.Impls, chain)
		}
		ls = append(ls, l)
	}
	slices.SortFunc(ls, func(a, b Learned) int {
		return strings.Compare(string(a.Key.Root)+string(a.Key.Kind)+a.Key.Text(), string(b.Key.Root)+string(b.Key.Kind)+b.Key.Text())
	})
	return ls, nil
}

// position is a stored condition's name for a key position, passedN,
// which the model wrote as $(N+1).
var position = regexp.MustCompile(`\bpassed(\d+)\b`)

// asWritten is a stored condition in the model's own $n.
func asWritten(cond string) string {
	return position.ReplaceAllStringFunc(cond, func(s string) string {
		n, _ := strconv.Atoi(s[len("passed"):])
		return "$" + strconv.Itoa(n+1)
	})
}

// Show writes what was learned: each key's id, kind, root and what it has
// seen, $n by $n, then each implementation's chain as the model wrote it,
// a condition then the instruction it selects, in order.
func Show(w io.Writer, ls []Learned) error {
	for i, l := range ls {
		if i > 0 {
			fmt.Fprintln(w)
		}
		root := strings.TrimPrefix(string(l.Key.Root), "sha256:")
		fmt.Fprintf(w, "%s  %s, root %s\n", l.ID, l.Key.Kind, root[:min(8, len(root))])
		for n, s := range l.Key.Seen {
			fmt.Fprintf(w, "  $%d %s\n", n+1, s.Text())
		}
		for n, chain := range l.Impls {
			fmt.Fprintf(w, "  implementation %d:\n", n+1)
			links, ok := tree.Written(chain)
			if !ok {
				return fmt.Errorf("%s: a chain not committed whole", l.ID)
			}
			for _, link := range links {
				t, err := model.ParseTemplate(link.Instr)
				if err != nil {
					return err
				}
				fmt.Fprintf(w, "    %s ? %s\n", asWritten(link.Cond), t.Instr())
			}
		}
	}
	return nil
}
