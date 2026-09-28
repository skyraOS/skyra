package tree

import (
	"os"
	"path/filepath"
	"testing"
)

// Write-once files are exclusive and complete: of many claimers exactly one
// wins, the claim holds the whole stamp, and nothing is left behind.
func TestClaimIsExclusiveAndAtomic(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "frames", "f")
	os.MkdirAll(dir, 0o755)
	const stamp = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	var wins int
	for i := 0; i < 8; i++ {
		f, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		// Claim uses Fchdir, which is process-wide: claims run one at a time
		// in a process, and concurrently only across processes.
		won, err := Claim(f, "r", stamp)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if won {
			wins++
		}
	}
	if wins != 1 {
		t.Fatalf("%d claims won, want exactly 1", wins)
	}
	if s, err := Resolved(root, "f", "r"); err != nil || s != stamp {
		t.Fatalf("claim reads %q, %v", s, err)
	}
	left, _ := filepath.Glob(filepath.Join(dir, ".tmp-*"))
	if len(left) != 0 {
		t.Fatalf("temporary files left behind: %v", left)
	}
}
