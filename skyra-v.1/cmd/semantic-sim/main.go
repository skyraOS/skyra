// Command semantic-sim scores pairwise embedding cosine over a set of files.
//
// It exists to test a branch retrieval.md explicitly leans against: whether a
// similarity heuristic can carry the interpret layer, where exact triples cannot
// be posting-list keys. A single cosine is not evidence — two Go files score high
// for being Go. Pass controls alongside the pair under test and read the spread.
//
//	go run ./cmd/semantic-sim a.go b.go control.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"skyra-v1/src/inference"
)

type doc struct {
	path string
	name string
	vec  []float64
}

type pair struct {
	a, b string
	cos  float64
}

func main() {
	paths := os.Args[1:]
	if len(paths) < 2 {
		fmt.Fprintln(os.Stderr, "usage: semantic-sim <file> <file> [file...]")
		os.Exit(2)
	}

	docs := make([]doc, 0, len(paths))
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", p, err)
			os.Exit(1)
		}

		vec, err := inference.Embed(string(body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "embed %s: %v\n", p, err)
			os.Exit(1)
		}

		name := filepath.Join(filepath.Base(filepath.Dir(p)), filepath.Base(p))
		docs = append(docs, doc{path: p, name: name, vec: vec})
		fmt.Fprintf(os.Stderr, "embedded %-28s %d bytes → %d dims\n", name, len(body), len(vec))
	}

	pairs := make([]pair, 0, len(docs)*(len(docs)-1)/2)
	for i := 0; i < len(docs); i++ {
		for j := i + 1; j < len(docs); j++ {
			pairs = append(pairs, pair{
				a:   docs[i].name,
				b:   docs[j].name,
				cos: inference.Cosine(docs[i].vec, docs[j].vec),
			})
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].cos > pairs[j].cos })

	fmt.Printf("\nmodel: %s\n\npairwise cosine, ranked:\n\n", inference.EmbedModel)
	for _, p := range pairs {
		fmt.Printf("  %.4f   %-24s  %s\n", p.cos, p.a, p.b)
	}

	if len(pairs) > 1 {
		hi, lo := pairs[0].cos, pairs[len(pairs)-1].cos
		fmt.Printf("\nspread: %.4f  (high %.4f, low %.4f)\n", hi-lo, hi, lo)
	}
}
