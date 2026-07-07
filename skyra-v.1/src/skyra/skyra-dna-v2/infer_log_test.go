package dnav2

import (
	"errors"
	"strings"
	"testing"
)

// The logger counts calls and errors and passes responses through unchanged. (debug.Being
// is a no-op without debug.Init, so this exercises the timing/rollup path in isolation.)
func TestInferLoggerCountsAndPassesThrough(t *testing.T) {
	calls := 0
	fn := func(system, present string) (string, error) {
		calls++
		if calls == 2 {
			return "", errors.New("boom")
		}
		return "ok:" + firstLine(system, 8), nil
	}
	il := NewInferLogger("skyra", fn)

	if r, err := il.Infer("hello world\nrest", "P"); err != nil || r != "ok:hello wo" {
		t.Fatalf("call 1: r=%q err=%v", r, err)
	}
	if _, err := il.Infer("second", "P"); err == nil {
		t.Fatal("call 2 should surface the underlying error")
	}
	if _, err := il.Infer("third", "P"); err != nil {
		t.Fatalf("call 3: err=%v", err)
	}

	s := il.Summary()
	if !strings.Contains(s, "3 calls") || !strings.Contains(s, "1 errs") {
		t.Fatalf("summary should report 3 calls / 1 err: %q", s)
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("  You are the coherence check.\nmore", 12); got != "You are the " {
		t.Fatalf("firstLine=%q", got)
	}
	if got := firstLine("short", 100); got != "short" {
		t.Fatalf("firstLine=%q", got)
	}
}
