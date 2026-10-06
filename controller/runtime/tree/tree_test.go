package tree

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// Write-once files are exclusive and complete: of many claimers exactly one
// wins, the claim holds the whole stamp, and nothing is left behind.
func TestClaimIsExclusiveAndAtomic(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "frames", "f")
	os.MkdirAll(dir, 0o755)
	const stamp = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	var wins atomic.Int32
	var group sync.WaitGroup
	before, _ := os.Getwd()
	for i := 0; i < 32; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			f, err := os.Open(dir)
			if err != nil {
				t.Error(err)
				return
			}
			won, err := Claim(f, "r", stamp)
			f.Close()
			if err != nil {
				t.Error(err)
				return
			}
			if won {
				wins.Add(1)
			}
		}()
	}
	group.Wait()
	after, _ := os.Getwd()
	if before != after {
		t.Fatal("claim changed cwd")
	}
	if wins.Load() != 1 {
		t.Fatalf("%d claims won, want exactly 1", wins.Load())
	}
	if s, err := Resolved(root, "f", "r"); err != nil || s != stamp {
		t.Fatalf("claim reads %q, %v", s, err)
	}
	left, _ := filepath.Glob(filepath.Join(dir, ".tmp-*"))
	if len(left) != 0 {
		t.Fatalf("temporary files left behind: %v", left)
	}
}
