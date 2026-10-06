package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skyra/controller/runtime/replay"
	"testing"
)

type bundleStore map[string]string

func (s bundleStore) Get(kind, hash string) ([]byte, error) {
	b, ok := s[kind+"/"+hash]
	if !ok {
		return nil, fmt.Errorf("missing object")
	}
	return []byte(b), nil
}
func TestPublishedBundlesReplay(t *testing.T) {
	paths, err := filepath.Glob("../runs/*.json")
	if err != nil {
		t.Fatal(err)
	}
	tested := 0
	for _, path := range paths {
		if filepath.Base(path) == "index.json" {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var bundle Bundle
			if err := json.Unmarshal(b, &bundle); err != nil {
				t.Fatal(err)
			}
			n, err := replay.Lineage(bundleStore(bundle.Objects), bundle.Head)
			if err != nil {
				t.Fatal(err)
			}
			if n != bundle.Records {
				t.Fatalf("replayed %d, recorded %d", n, bundle.Records)
			}
		})
		tested++
	}
	if tested == 0 {
		t.Fatal("no bundles verified")
	}
}
