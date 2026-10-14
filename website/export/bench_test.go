package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/runtime/replay"
)

// BenchmarkPublishedRuns replays each recorded run the website publishes,
// from the bundle in memory, as the browser does.
func BenchmarkPublishedRuns(b *testing.B) {
	paths, err := filepath.Glob("../runs/*.json")
	if err != nil {
		b.Fatal(err)
	}
	for _, path := range paths {
		if filepath.Base(path) == "index.json" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}
		var bundle Bundle
		if err := json.Unmarshal(raw, &bundle); err != nil {
			b.Fatal(err)
		}
		s := bundleStore(bundle.Objects)
		b.Run(strings.TrimSuffix(filepath.Base(path), ".json"), func(b *testing.B) {
			for b.Loop() {
				if n, err := replay.Lineage(s, bundle.Head); err != nil || n != bundle.Records {
					b.Fatalf("replayed %d, %v; recorded %d", n, err, bundle.Records)
				}
			}
			b.ReportMetric(float64(bundle.Records)*float64(b.N)/b.Elapsed().Seconds(), "records/s")
		})
	}
}
